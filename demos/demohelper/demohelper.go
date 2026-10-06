// Package demohelper provides common boilerplate for Tk demo applications.
package demohelper

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/geometry"
	"github.com/takigo/takigo/geometry/grid"
	"github.com/takigo/takigo/geometry/pack"
	tkimage "github.com/takigo/takigo/image"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/ttk"
	_ "github.com/takigo/takigo/ttk/clamtheme"
	_ "github.com/takigo/takigo/ttk/defaulttheme"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/frame"
	"github.com/takigo/takigo/widget/labelframe"
	"github.com/takigo/takigo/widget/scrollbar"
	"github.com/takigo/takigo/widget/text"
	"github.com/takigo/takigo/widget/toplevel"
	"github.com/takigo/takigo/window"
)

//go:embed icons/*.png
var icons embed.FS

var (
	img map[string]*tkimage.Photo
	// varsWindow is the single reusable "See Variables" toplevel (nil until first use).
	varsWindow *toplevel.Toplevel
	// varsUnsubs holds OnChange unsubscribers from the current vars dialog so
	// they can be released when the dialog closes (or is rebuilt).
	varsUnsubs []func()
	// lastBottomButtons stores buttons created by the most recent AddBottomButtons call.
	lastBottomButtons []*ttk.Button
)

// releaseVarsSubscriptions unsubscribes all OnChange listeners registered by
// the current vars dialog so they don't keep firing after the window is gone.
func releaseVarsSubscriptions() {
	for _, u := range varsUnsubs {
		u()
	}
	varsUnsubs = nil
}

// BottomButtons returns the TTK buttons created by the most recent
// AddSeeDismiss or AddBottomButtons call.
func BottomButtons() []*ttk.Button {
	return lastBottomButtons
}

// Image returns one of the launcher icons ("view", "delete", "refresh",
// "print"), as the Tcl demos' ::img::<name> images.
// Now is time.Now, except under TAKIGO_FREEZE_TIMERS=1, where it returns the
// fixed instant scripts/demo_wrapper.tcl gives [clock seconds].
func Now() time.Time {
	if os.Getenv("TAKIGO_FREEZE_TIMERS") == "1" {
		return time.Unix(1700000000, 0)
	}
	return time.Now()
}

func Image(name string) *tkimage.Photo { return img[name] }

type DemoVars[T comparable] map[string]*widget.Variable[T]

func init() {
	img = make(map[string]*tkimage.Photo)
	for _, name := range []string{"view", "delete", "refresh", "print"} {
		img[name] = loadIcon(name)
	}
}

// DemoDir returns the absolute path to a subdirectory under demos/.
func DemoDir(name string) string {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		return name
	}
	demosRoot := filepath.Dir(filepath.Dir(file))
	return filepath.Join(demosRoot, name)
}

func AddVarsSeeDismiss[T comparable](parent widget.Caregiver, vars *DemoVars[T]) *ttk.Frame {
	varFunc := func(f *ttk.Frame) *ttk.Button {
		return ttk.NewButton(f, "vars",
			ttk.ButtonText("See Variables"),
			ttk.ButtonImage(img["view"]),
			ttk.ButtonCompound(widget.CompoundLeft),
			ttk.ButtonCommand(func() { showVars(parent.AppContext(), vars) }),
		)
	}

	return AddBottomButtons(parent, varFunc)
}

func AddSeeDismiss(parent widget.Caregiver) *ttk.Frame {
	return AddBottomButtons(parent, func(*ttk.Frame) *ttk.Button { return nil })
}

// NamedVar is a name + widget variable pair, used by AddSeeDismissWithVars to
// show the See Variables button.
type NamedVar struct {
	Name string
	Var  any // one of *widget.Variable[T] for any T
}

// AddSeeDismissWithVars creates the See Code / Dismiss button bar with an
// additional See Variables button that displays the given named variables.
// Matches Tcl's: addSeeDismiss $w.buttons $w [list size color align ...].
func AddSeeDismissWithVars(parent widget.Caregiver, vars []NamedVar) *ttk.Frame {
	varsMap := make(map[string]any, len(vars))
	for _, nv := range vars {
		varsMap[nv.Name] = nv.Var
	}
	return AddBottomButtons(parent, func(f *ttk.Frame) *ttk.Button {
		return ttk.NewButton(f, "vars",
			ttk.ButtonText("See Variables"),
			ttk.ButtonImage(img["view"]),
			ttk.ButtonCompound(widget.CompoundLeft),
			ttk.ButtonCommand(func() { showVarsAny(parent.AppContext(), varsMap) }),
		)
	})
}

// showVarsAny is a type-erased showVars for mixed-type variable maps.
// It uses each variable's underlying *widget.Variable via a small adapter.
func showVarsAny(app widget.AppContext, vars map[string]any) {
	releaseVarsSubscriptions()
	if varsWindow != nil && !varsWindow.Destroyed {
		varsWindow.Destroy()
	}
	varsWindow = toplevel.New(app, "vars",
		toplevel.Title("Variable values"),
	)
	varsWindow.Show()

	varsRoot := varsWindow.Window()

	b := ttk.NewFrame(varsWindow, "frame")
	grid.Grid(b, grid.Sticky(grid.NSEW))

	f := labelframe.New(b, "title", labelframe.Text("Variable values:"))

	// Sort names for stable display.
	names := make([]string, 0, len(vars))
	for name := range vars {
		names = append(names, name)
	}
	slices.Sort(names)

	for row, name := range names {
		nameLabel := ttk.NewLabel(f, "n_"+name, ttk.LabelText(name+":"))
		valLabel := ttk.NewLabel(f, "v_"+name,
			ttk.LabelText(fmt.Sprintf("%v", variableGet(vars[name]))),
		)
		if unsub := bindAnyVarToLabel(valLabel, vars[name]); unsub != nil {
			varsUnsubs = append(varsUnsubs, unsub)
		}
		grid.Grid(geometry.Group{nameLabel, valLabel}, grid.Column(0), grid.Row(row),
			grid.PadX(screenunit.Pt(1.5)), grid.PadY(screenunit.Pt(1.5)), grid.Sticky(grid.StickW))
	}

	okBtn := ttk.NewButton(b, "ok",
		ttk.ButtonText("OK"),
		ttk.ButtonCommand(func() {
			releaseVarsSubscriptions()
			varsWindow.Destroy()
		}),
	)

	grid.Grid(f, grid.Sticky(grid.NSEW), grid.PadX(screenunit.Pt(3)))
	grid.Grid(okBtn, grid.Row(1), grid.Sticky(grid.StickE), grid.PadX(screenunit.Pt(3)), grid.PadY(screenunit.Pt(3)))

	grid.ColumnConfigure(f, 1, grid.Weight(1))
	grid.RowConfigure(f, 100, grid.Weight(1))
	grid.ColumnConfigure(b, 0, grid.Weight(1))
	grid.RowConfigure(b, 0, grid.Weight(1))
	grid.ColumnConfigure(varsRoot, 0, grid.Weight(1))
	grid.RowConfigure(varsRoot, 0, grid.Weight(1))

	varsWindow.OnClose(func() {
		releaseVarsSubscriptions()
		varsWindow.Destroy()
	})
}

// bindAnyVarToLabel subscribes valLabel to a variable of any common type so
// the label updates whenever the variable changes. Returns an unsubscribe
// function (or nil if the variable type is not supported).
func bindAnyVarToLabel(valLabel *ttk.Label, v any) func() {
	switch val := v.(type) {
	case *widget.Variable[string]:
		return val.OnChange(func(_, new string) {
			updateValueLabel(valLabel, new)
		})
	case *widget.Variable[bool]:
		return val.OnChange(func(_, new bool) {
			updateValueLabel(valLabel, fmt.Sprintf("%v", new))
		})
	case *widget.Variable[int]:
		return val.OnChange(func(_, new int) {
			updateValueLabel(valLabel, fmt.Sprintf("%v", new))
		})
	case *widget.Variable[float64]:
		return val.OnChange(func(_, new float64) {
			updateValueLabel(valLabel, fmt.Sprintf("%v", new))
		})
	}
	return nil
}

// updateValueLabel sets valLabel's text and lets it grow/shrink to fit by
// recomputing its requested size and propagating the change to the geometry
// manager. This matches Tk's -textvariable behavior where a label re-requests
// its natural size whenever the underlying variable changes.
func updateValueLabel(valLabel *ttk.Label, text string) {
	if valLabel.Destroyed {
		return
	}
	valLabel.Configure(ttk.LabelText(text))
}

// variableGet extracts the current value from any *widget.Variable[T].
// Variables are generic; we use reflection to call Get().
func variableGet(v any) any {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() || rv.Kind() != reflect.Pointer {
		return "<unknown>"
	}
	method := rv.MethodByName("Get")
	if !method.IsValid() {
		return "<unknown>"
	}
	results := method.Call(nil)
	if len(results) == 0 {
		return "<unknown>"
	}
	return results[0].Interface()
}

func AddBottomButtons(parent widget.Caregiver, varsFunc func(*ttk.Frame) *ttk.Button) *ttk.Frame {
	_, callerFile, _, _ := runtime.Caller(2)
	return bottomButtons(parent, callerFile, func(f *ttk.Frame) window.Windower {
		if b := varsFunc(f); b != nil {
			return b
		}
		return nil
	})
}

// AddSeeDismissExtra is addSeeDismiss with an extra widget (the launcher's
// "extra" argument): it is gridded between the empty column and See Code.
func AddSeeDismissExtra(parent widget.Caregiver, extra func(*ttk.Frame) window.Windower) *ttk.Frame {
	_, callerFile, _, _ := runtime.Caller(1)
	return bottomButtons(parent, callerFile, extra)
}

func bottomButtons(parent widget.Caregiver, callerFile string, extra func(*ttk.Frame) window.Windower) *ttk.Frame {
	btnFrame := ttk.NewFrame(parent, "bottom_buttons")

	sep := ttk.NewSeparator(btnFrame, "sep")
	grid.Grid(sep, grid.ColumnSpan(4), grid.Row(0), grid.Sticky(grid.EW), grid.PadY(screenunit.Pt(1.5)))

	dismissBtn := ttk.NewButton(btnFrame, "dismiss",
		ttk.ButtonText("Dismiss"),
		ttk.ButtonImage(img["delete"]),
		ttk.ButtonCompound(widget.CompoundLeft),
		ttk.ButtonCommand(func() { parent.AppContext().Quit() }),
	)

	codeBtn := ttk.NewButton(btnFrame, "code",
		ttk.ButtonText("See Code"),
		ttk.ButtonImage(img["view"]),
		ttk.ButtonCompound(widget.CompoundLeft),
		ttk.ButtonCommand(func() { showCode(parent.AppContext(), callerFile) }),
	)

	buttons := []window.Windower{grid.Relative(grid.RelEmpty), codeBtn, dismissBtn}
	lastBottomButtons = []*ttk.Button{codeBtn, dismissBtn}

	if w := extra(btnFrame); w != nil {
		buttons = []window.Windower{grid.Relative(grid.RelEmpty), w, codeBtn, dismissBtn}
		if b, ok := w.(*ttk.Button); ok {
			lastBottomButtons = []*ttk.Button{b, codeBtn, dismissBtn}
		}
	}

	grid.Grid(geometry.Group(buttons), grid.PadX(screenunit.Pt(3)), grid.PadY(screenunit.Pt(3)))
	grid.ColumnConfigure(btnFrame, 0, grid.Weight(1))

	return btnFrame
}

// makeViewIcon creates a 16x16 magnifying glass icon (matches Tk's ::img::view).
// loadIcon decodes one of the launcher icons that Tk itself rasterized from
// the SVGs in tk/library/demos/widget (scripts/export_launcher_icons.tcl), so
// the See Code / Dismiss buttons are pixel-identical to Tk's.
func loadIcon(name string) *tkimage.Photo {
	f, err := icons.Open("icons/" + name + ".png")
	if err != nil {
		panic(err)
	}
	defer func() { _ = f.Close() }()
	p, err := tkimage.NewPhotoFromReader("::img::"+name, f)
	if err != nil {
		panic(err)
	}
	return p
}

// codeWindow is the single reusable "See Code" toplevel (nil until first use).
var codeWindow *toplevel.Toplevel
var codeText *text.TextWidget

// showCode opens (or raises) a toplevel window displaying the demo source.
// Uses TTK widgets for the button bar (matching Tk's showCode proc).
// ShowCode opens the See Code window for the calling demo's source file,
// like the launcher's showCode proc.
func ShowCode(app widget.AppContext) {
	_, callerFile, _, _ := runtime.Caller(1)
	showCode(app, callerFile)
}

func showCode(app widget.AppContext, srcFile string) {
	source, err := os.ReadFile(srcFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "See Code: %v\n", err)
		return
	}

	if codeWindow != nil && !codeWindow.Destroyed {
		// Reuse: clear old text and insert new source.
		codeText.Delete("1.0", "end")
		codeText.Insert("1.0", string(source))
		codeText.See("1.0")
		codeWindow.Show()
		return
	}

	codeWindow = toplevel.New(app, "code",
		toplevel.Title(fmt.Sprintf("Demo code: %s", srcFile)),
		toplevel.Background("#d9d9d9"),
	)
	codeWindow.Show()

	codeRoot := codeWindow.Window()

	// Button bar at bottom — ttk::frame + ttk::separator + ttk::button.
	codeBtnFrame := ttk.NewFrame(codeWindow, "codebtns")
	pack.Pack(codeBtnFrame, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	codeSep := ttk.NewSeparator(codeBtnFrame, "codesep")
	pack.Pack(codeSep, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(2))

	dismissBtn := ttk.NewButton(codeBtnFrame, "codedismiss",
		ttk.ButtonText("Dismiss"),
		ttk.ButtonCommand(func() { codeWindow.Hide() }),
	)
	pack.Pack(dismissBtn, pack.SideOpt(pack.Right), pack.PadX(4), pack.PadY(4))

	// Text widget with scrollbar (classic text, same as Tk's showCode).
	txtFrame := frame.New(codeWindow, "codetxtframe")
	pack.Pack(txtFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	codeText = text.New(txtFrame, "codetext",
		text.Width(80),
		text.Height(24),
		text.WrapModeOpt(text.WrapNone),
		text.FontOpt("monospace 10"),
		text.ReadOnly(true),
	)

	sb := scrollbar.New(txtFrame, "codesb")
	codeText.YScrollCmd = func(first, last float64) { sb.Set(first, last) }
	sb.Command = widget.ScrollY(codeText)

	pack.Pack(sb, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(codeText, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	codeText.Insert("1.0", string(source))
	codeText.See("1.0")

	// Handle resize.
	app.Dispatcher().Bind(codeRoot.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			codeRoot.Width = ev.ConfigWidth
			codeRoot.Height = ev.ConfigHeight
			pack.ArrangeContainer(codeRoot)
		}
	})

	// Close button hides instead of destroying.
	codeWindow.OnClose(func() { codeWindow.Hide() })
}

// showVars opens a toplevel window displaying the current values of demo variables.
// Matches Tk's showVars proc from tk/library/demos/widget. The values are kept
// live: whenever a linked variable changes, the corresponding label updates
// immediately, matching Tcl's `ttk::label ... -textvariable $var` behavior.
//
//	proc showVars {w args} {
//	    catch {destroy $w}
//	    toplevel $w
//	    wm title $w "Variable values"
//	    ...
//	    ttk::label $f.v$var -textvariable $var -anchor w
//	    ...
//	}
func showVars[T comparable](app widget.AppContext, vars *DemoVars[T]) {
	asAny := make(map[string]any, len(*vars))
	for name, v := range *vars {
		asAny[name] = v
	}
	showVarsAny(app, asAny)
}

var fileIcons = map[string]*tkimage.Photo{}

// FileIcon ports [tk fileicon path size] on X11 (library/fileicon.tcl), for
// size 16 or 48: the icon is chosen by directory, executable bit and
// extension, from Tk's own SVGs rasterized by scripts/export_fileicons.tcl.
func FileIcon(path string, size int) *tkimage.Photo {
	name := "text"
	if st, err := os.Stat(path); err == nil && st.IsDir() {
		name = "folder"
	} else if err == nil && st.Mode()&0o111 != 0 {
		name = "executable"
	} else {
		switch strings.ToLower(filepath.Ext(path)) {
		case ".svg":
			name = "drawing"
		case ".bmp", ".gif", ".ico", ".icns", ".jpg", ".jpeg", ".png", ".tif", ".tiff", ".webp", ".pdf":
			name = "image"
		case ".aac", ".aiff", ".mid", ".midi", ".mp3", ".wav", ".oga", ".opus", ".weba":
			name = "audio"
		case ".ics":
			name = "calendar"
		case ".abw", ".odt", ".doc", ".docx", ".rtf":
			name = "word"
		case ".odp", ".ppt", ".pptx":
			name = "presentation"
		case ".csv", ".ods", ".xls", ".xlsx":
			name = "spreadsheet"
		case ".arc", ".bz", ".bz2", ".jar", ".mpkg", ".rar", ".tar", ".zip", ".tz", ".tgz", ".7z":
			name = "archive"
		case ".css", ".js", ".html", ".htm", ".epub", ".json", ".swf", ".xhtml", ".xml":
			name = "html"
		case ".csh", ".php", ".sh", ".py", ".pl", ".tcl", ".rb":
			name = "script"
		case ".bin", ".db", ".so", ".dll", ".dylib":
			name = "binary"
		case ".eot", ".otf", ".ttf", ".woff", ".woff2":
			name = "font"
		case ".eml", ".msg", ".mbox", ".pst", ".oft":
			name = "mail"
		}
	}
	name = fmt.Sprintf("fileicon-%s-%d", name, size)
	if p, ok := fileIcons[name]; ok {
		return p
	}
	p := loadIcon(name)
	fileIcons[name] = p
	return p
}
