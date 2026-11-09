// Ports the X11 part of tk/library/print.tcl: "tk print" for canvas and
// text widgets through CUPS (lpstat, lpoptions, lpr).
package dialog

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/takigo/takigo/canvas"
	"github.com/takigo/takigo/geometry"
	"github.com/takigo/takigo/geometry/grid"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/ttk"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/menu"
	"github.com/takigo/takigo/widget/text"
	"github.com/takigo/takigo/window"
)

// ErrPrintUnsupported is returned by Print where only the X11 (CUPS)
// printing of print.tcl is ported.
var ErrPrintUnsupported = errors.New("dialog: printing is only supported with CUPS")

// printOptions are print.tcl's option array; they persist per App, so a
// second Print starts from the choices of the first.
type printOptions struct {
	printer                                          string
	copies                                           int
	media, orient, color                             string // a4/legal/letter, portrait/landscape, color/gray
	czoom, numberUp, tzoom                           int
	pprint                                           bool
	marginTop, marginLeft, marginBottom, marginRight int // points
}

func defaultPrintOptions() *printOptions {
	mm := func(v float64) int { return int(v * 72 / 25.4) }
	return &printOptions{copies: 1, media: "a4", orient: "portrait", color: "color",
		czoom: 100, numberUp: 1, tzoom: 100,
		marginTop: mm(15), marginLeft: mm(25), marginRight: mm(15), marginBottom: mm(15)}
}

var printOptionsKey = new(window.ValueKey)

// --- cups ensemble ---

// cupsDefaultPrinter is "cups defaultprinter": lpstat -d.
func cupsDefaultPrinter() string {
	out, err := exec.Command("lpstat", "-d").Output()
	if err != nil {
		return ""
	}
	if m := regexp.MustCompile(`: ([^[:space:]]+)$`).FindSubmatch(out[:len(strings.TrimRight(string(out), "\n"))]); m != nil {
		return string(m[1])
	}
	return ""
}

// cupsPrinters is "cups getprinters": the printers of lpstat -e with
// their lpoptions.
func cupsPrinters() (map[string]map[string]string, error) {
	out, err := exec.Command("lpstat", "-e").Output()
	var ee *exec.Error
	switch {
	case errors.As(err, &ee):
		return nil, fmt.Errorf("unable to obtain the list of printers. Command \"lpstat\" not found. Please install the CUPS package for your system")
	case err != nil:
		return nil, fmt.Errorf("command \"lpstat\" return with errors")
	}
	res := map[string]map[string]string{}
	for p := range strings.SplitSeq(strings.TrimSpace(string(out)), "\n") {
		if p == "" || res[p] != nil {
			continue
		}
		opts, err := exec.Command("lpoptions", "-p", p).Output() //nolint:gosec // G204: p is a printer lpstat listed
		if err != nil {
			return nil, fmt.Errorf("command \"lpoptions\" return with errors")
		}
		res[p] = parseLpoptions(string(opts))
	}
	return res, nil
}

// parseLpoptions ports Parseoptions: key, key=value, key='value with spaces'.
func parseLpoptions(data string) map[string]string {
	res := map[string]string{}
	// Tcl regexps match leftmost-longest, as POSIX ones do.
	re := regexp.MustCompilePOSIX(`[^ =]+|[^ ]+='[^']+'|[^ ]+=[^ ']+`)
	for _, tok := range re.FindAllString(data, -1) {
		k, v, _ := strings.Cut(tok, "=")
		res[strings.TrimSpace(k)] = strings.Trim(strings.TrimSpace(v), "'")
	}
	return res
}

// lprArgs ports the option handling of "cups print".
func lprArgs(printer, title string, o *printOptions, text bool) []string {
	var a []string
	opt := func(s string) { a = append(a, "-o", s) }
	opt("copies=" + strconv.Itoa(o.copies))
	opt("media=" + o.media)
	if text {
		if o.tzoom != 100 {
			z := float64(o.tzoom) / 100
			opt("cpi=" + tclDouble(10/z))
			opt("lpi=" + tclDouble(6/z))
		}
		if o.pprint {
			opt("prettyprint=true")
			if !slices.ContainsFunc(a, func(s string) bool { return strings.HasPrefix(s, "cpi=") }) {
				opt("cpi=10.0")
				opt("lpi=6.0")
			}
		}
		if o.numberUp != 1 {
			opt("number-up=" + strconv.Itoa(o.numberUp))
		}
		opt("page-top=" + strconv.Itoa(o.marginTop))
		opt("page-left=" + strconv.Itoa(o.marginLeft))
		opt("page-bottom=" + strconv.Itoa(o.marginBottom))
		opt("page-right=" + strconv.Itoa(o.marginRight))
	}
	return append(a, "-T", title, "-P", printer)
}

// tclDouble formats f as Tcl prints a double: 20.0, not 20.
func tclDouble(f float64) string {
	s := strconv.FormatFloat(f, 'g', -1, 64)
	if !strings.ContainsAny(s, ".eEn") {
		s += ".0"
	}
	return s
}

// cupsPrint writes data to a temporary file and runs lpr on it in the
// background; -r has lpr delete the file.
func cupsPrint(args []string, data []byte) error {
	f, err := os.CreateTemp("", "tk_print")
	if err != nil {
		return err
	}
	_, err = f.Write(append(data, '\n'))
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		_ = os.Remove(f.Name())
		return err
	}
	cmd := exec.Command("lpr", append(args, "-r", f.Name())...) //nolint:gosec // G204: lpr options built by lprArgs
	if err := cmd.Start(); err != nil {
		_ = os.Remove(f.Name())
		return err
	}
	go cmd.Wait() //nolint:errcheck // a background job, like "exec lpr ... &"
	return nil
}

// wrapLines ports _wrapLines: break lines longer than wl at a space or
// tab, collapsing the blanks only where it breaks.
func wrapLines(s string, wl int) []string {
	re := regexp.MustCompilePOSIX(fmt.Sprintf(`((^|[^[:blank:]]).{0,%d})([[:blank:]]|$)`, max(wl-1, 0)))
	var res []string
	if s == "" {
		return nil // Tcl's split of "" has no lines
	}
	for line := range strings.SplitSeq(s, "\n") {
		for _, m := range re.FindAllStringSubmatch(line, -1) {
			res = append(res, m[1])
		}
	}
	return res
}

// latin1 is "encoding convertto iso8859-1": unrepresentable characters
// become "?".
func latin1(s string) []byte {
	b := make([]byte, 0, len(s))
	for _, r := range s {
		if r > 0xff {
			r = '?'
		}
		b = append(b, byte(r))
	}
	return b
}

// --- the dialog ---

// Print shows Tk's print dialog for a canvas or text widget ("tk print")
// and sends the job to CUPS with lpr. It returns an error when w is
// neither, when no printer is installed, and on Windows and macOS.
func Print(parent widget.Caregiver, w widget.Widget) error {
	_, isCanvas := w.(*canvas.Canvas)
	if !isCanvas && !isText(w) {
		return fmt.Errorf("printing windows of class %q is not supported", w.Window().Class)
	}
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		return ErrPrintUnsupported
	}
	app := parent.AppContext()
	root := app.Root()
	opts, _ := root.Value(printOptionsKey).(*printOptions)
	if opts == nil {
		opts = defaultPrintOptions()
		root.SetValue(printOptionsKey, opts)
	}

	// _setprintenv
	pinfo, err := cupsPrinters()
	if err != nil {
		return err
	}
	if len(pinfo) == 0 {
		return errors.New("no installed printers found. Please check or update your CUPS installation")
	}
	printers := slices.Sorted(maps.Keys(pinfo))
	if opts.printer == "" {
		if opts.printer = cupsDefaultPrinter(); opts.printer == "" {
			opts.printer = printers[0]
		}
	}

	dlg := *opts // the dialog's copy, kept only on Print
	d := New(parent, "Print", 300, 200)
	pad := screenunit.Pt(9).Pixels()
	sp := screenunit.Pt(2).Pixels()

	// The printer to use.
	pf := ttk.NewFrame(d.Content, "printerf")
	pack.Pack(pf, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadX(pad), pack.PadY(pad))
	grid.Grid(ttk.NewLabel(pf, "printerl", ttk.LabelText("Printer")), grid.Row(0), grid.Column(0), grid.Sticky(grid.StickW))
	tv := ttk.NewTreeview(pf, "prlist", ttk.TreeviewHeight(5), ttk.TreeviewColumns("printer", "location", "state"),
		ttk.TreeviewShow("headings"), ttk.TreeviewSelectMode(ttk.TreeSelectBrowse))
	sy := ttk.NewScrollbar(pf, "sy", ttk.ScrollbarCommand(widget.ScrollY(tv)))
	sx := ttk.NewScrollbar(pf, "sx", ttk.ScrollbarOrient(ttk.Horizontal), ttk.ScrollbarCommand(widget.ScrollX(tv)))
	// _scroll: a scrollbar is gridded only while the view is partial.
	autoscroll := func(sb *ttk.Scrollbar, opts ...grid.GridOption) func(float64, float64) {
		shown := false
		return func(first, last float64) {
			if full := first == 0 && last == 1; full == !shown {
				return
			} else if full {
				grid.Forget(sb)
			} else {
				_ = grid.Grid(sb, opts...)
			}
			shown = !shown
			sb.Set(first, last)
		}
	}
	tv.YScrollCmd = autoscroll(sy, grid.Row(1), grid.Column(1), grid.Sticky(grid.NS))
	tv.XScrollCmd = autoscroll(sx, grid.Row(2), grid.Column(0), grid.Sticky(grid.EW))
	tv.HeadingConfigure("printer", ttk.HeadText("Printer"))
	tv.HeadingConfigure("location", ttk.HeadText("Location"))
	tv.HeadingConfigure("state", ttk.HeadText("State"))
	tv.ColumnConfigure("printer", ttk.ColWidth(200), ttk.ColStretch(false))
	tv.ColumnConfigure("location", ttk.ColWidth(100), ttk.ColStretch(false))
	tv.ColumnConfigure("state", ttk.ColWidth(250), ttk.ColStretch(false))
	statemap := map[string]string{"3": "Idle", "4": "Printing", "5": "Printer stopped"}
	for _, p := range printers {
		info := pinfo[p]
		nstate := info["printer-state"]
		state := statemap[nstate]
		switch nstate {
		case "3", "4":
			if info["printer-is-accepting-jobs"] != "" {
				state += ". Printer is accepting jobs"
			}
		case "5":
			if r := info["printer-state-reasons"]; r != "" {
				state += ". (" + r + ")"
			}
		}
		id := tv.Insert("", -1, ttk.ItemValues(p, info["printer-location"], state))
		if p == dlg.printer {
			tv.SelectionSet(id)
		}
	}
	grid.Grid(tv, grid.Row(1), grid.Column(0), grid.Sticky(grid.NSEW))
	tv.OnSelect = func() { // _onselect
		dlg.printer = ""
		if sel := tv.Selection(); len(sel) > 0 {
			dlg.printer = tv.CellValue(sel[0], "printer")
		}
	}

	// Printing options.
	of := ttk.NewLabelframe(d.Content, "optionsframe", ttk.LabelframeText("Options"),
		ttk.LabelframePadding(strconv.Itoa(screenunit.Pt(2).Pixels())))
	pack.Pack(of, pack.FillOpt(pack.FillX), pack.PadX(pad), pack.PadYPair(0, pad))
	row := 0
	addRow := func(name, label string, w geometry.Elementer) {
		l := ttk.NewLabel(of, name+"l", ttk.LabelText(label+" :"), ttk.LabelAnchor(option.AnchorE))
		grid.Grid(l, grid.Row(row), grid.Column(0), grid.Sticky(grid.EW), grid.PadX(sp), grid.PadY(sp))
		grid.Grid(w, grid.Row(row), grid.Column(1), grid.Sticky(grid.EW), grid.PadX(sp), grid.PadY(sp))
		row++
	}
	spin := func(name string, from, to, incr int, v *int) *ttk.Spinbox {
		sb := ttk.NewSpinbox(of, name, ttk.SpinboxFrom(float64(from)), ttk.SpinboxTo(float64(to)),
			ttk.SpinboxIncrement(float64(incr)), ttk.SpinboxWidth(12),
			ttk.SpinboxCommand(func(val string) { *v, _ = strconv.Atoi(val) }))
		sb.Set(strconv.Itoa(*v))
		sb.ChangeState(ttk.StateReadonly, 0)
		return sb
	}
	// optionMenu is print.tcl's ttk_optionMenu over (label, value) pairs.
	optionMenu := func(name string, v *string, pairs ...string) *ttk.Menubutton {
		text := *v
		for i := 0; i < len(pairs); i += 2 {
			if pairs[i+1] == *v {
				text = pairs[i]
			}
		}
		m := menu.New(of, name+"menu", menu.TearOffOpt(false))
		mb := ttk.NewMenubutton(of, name, ttk.MenubuttonText(text), ttk.MenubuttonMenu(m))
		for i := 0; i < len(pairs); i += 2 {
			label, val := pairs[i], pairs[i+1]
			m.AddCommand(label, func() {
				*v = val
				_ = mb.Configure(ttk.MenubuttonText(label))
			})
		}
		return mb
	}
	addRow("copies", "Copies", spin("copies", 1, 1000, 1, &dlg.copies))
	addRow("media", "Paper", optionMenu("media", &dlg.media, "Letter", "letter", "Legal", "legal", "A4", "a4"))
	nup := strconv.Itoa(dlg.numberUp)
	pprint := widget.NewVariable(dlg.pprint)
	if isCanvas {
		addRow("percent", "Scale", spin("percent", 5, 500, 5, &dlg.czoom))
		addRow("orient", "Orientation", optionMenu("orient", &dlg.orient, "Portrait", "portrait", "Landscape", "landscape"))
		addRow("color", "Output", optionMenu("color", &dlg.color, "RGB", "color", "Grayscale", "gray"))
	} else {
		var pairs []string
		for _, n := range []string{"1", "2", "4", "6", "9", "16"} {
			pairs = append(pairs, n, n)
		}
		addRow("nup", "Pages per sheet", optionMenu("nup", &nup, pairs...))
		addRow("tzoom", "Text scale", spin("tzoom", 50, 200, 5, &dlg.tzoom))
		cb := ttk.NewCheckbutton(of, "pprint", ttk.CheckbuttonText("Pretty print"), ttk.CheckbuttonVar(pprint))
		grid.Grid(cb, grid.Row(row), grid.Column(0), grid.ColumnSpan(2), grid.Sticky(grid.EW), grid.PadX(sp), grid.PadY(sp))
	}

	// The buttons frame: Print right, Cancel left of it.
	bf := ttk.NewFrame(d.Content, "buttonf")
	pack.Pack(bf, pack.FillOpt(pack.FillX), pack.SideOpt(pack.Bottom), pack.PadX(pad), pack.PadYPair(0, pad))
	printBtn := ttk.NewButton(bf, "print", ttk.ButtonText("Print"), ttk.ButtonCommand(func() { d.Close(ResultOK) }))
	cancel := ttk.NewButton(bf, "cancel", ttk.ButtonText("Cancel"), ttk.ButtonCommand(func() { d.Close(ResultCancel) }))
	pack.Pack(printBtn, pack.SideOpt(pack.Right))
	pack.Pack(cancel, pack.SideOpt(pack.Right), pack.PadXPair(0, screenunit.Pt(4.5)))
	pack.Forget(d.BtnFrame)
	if d.Run() != ResultOK {
		return nil
	}
	dlg.numberUp, _ = strconv.Atoi(nup)
	dlg.pprint = pprint.Get()
	*opts = dlg
	if pinfo[opts.printer] == nil {
		return fmt.Errorf("unknown printer or class %q", opts.printer)
	}
	if opts.copies < 0 || opts.copies > 100 {
		return errors.New("copies must be an integer between 0 and 100")
	}
	return runPrint(app.Root().PathName, w, opts)
}

// runPrint ports _runprint: canvas PostScript, or the text wrapped to the
// page, to "cups print" in the background.
func runPrint(appName string, w widget.Widget, o *printOptions) error {
	title := fmt.Sprintf("%s: Tk window %s", appName, w.Window().PathName)
	var data []byte
	switch w := w.(type) {
	case *canvas.Canvas:
		psOpts := []canvas.PostscriptOption{
			canvas.PSColorMode(o.color), canvas.PSRotate(o.orient == "landscape"),
			// Scale based on size of widget, not size of paper.
			canvas.PSPageWidth(float64(o.czoom) / 100 * float64(w.Window().Width)),
		}
		if x1, y1, x2, y2, ok := w.GetScrollRegion(); ok {
			// print.tcl passes the scroll region as -x -y -width -height.
			psOpts = append(psOpts, canvas.PSRegion(x1, y1, x2, y2))
		}
		ps, err := w.Postscript(psOpts...)
		if err != nil {
			return err
		}
		data = latin1(ps)
	case *text.TextWidget:
		pw := map[string]float64{"a4": 8.27, "legal": 8.5, "letter": 8.5}[o.media]
		pw -= float64(o.marginLeft+o.marginRight) / 72
		// The wrap length is 98% of the page width at the default 10 cpi.
		wl := int(9.8 * pw / (float64(o.tzoom) / 100))
		data = []byte(strings.Join(wrapLines(w.Get("1.0", "end"), wl), "\n"))
	}
	return cupsPrint(lprArgs(o.printer, title, o, isText(w)), data)
}

func isText(w widget.Widget) bool {
	_, ok := w.(*text.TextWidget)
	return ok
}
