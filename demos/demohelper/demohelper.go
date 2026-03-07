// Package demohelper provides common boilerplate for Tk demo applications.
package demohelper

import (
	"fmt"
	goimage "image"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	tkimage "github.com/msorc/takigo/image"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/labelframe"
	"github.com/msorc/takigo/widget/scrollbar"
	"github.com/msorc/takigo/widget/text"
	"github.com/msorc/takigo/widget/toplevel"
	"github.com/msorc/takigo/window"
)

var (
	img map[string]*tkimage.Photo
	// varsWindow is the single reusable "See Variables" toplevel (nil until first use).
	varsWindow *toplevel.Toplevel
)

type DemoVars map[string]*widget.Variable[any]

func init() {
	img = make(map[string]*tkimage.Photo)
	img["view"] = makeViewIcon()
	img["delete"] = makeDeleteIcon()
}

// Setup creates a standard demo window with a description label, and
// "See Code" / "Dismiss" buttons at the bottom (matching Tk's addSeeDismiss).
// Calls os.Exit(1) on failure.
func Setup(title string, width, height int, description string) *takigo.App {
	app, err := takigo.NewApp(takigo.Title(title), takigo.Size(width, height))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	// Description label.
	msg := label.New(app, "msg",
		label.Text(description),
		label.Anchor(option.AnchorW),
		label.WrapLength("4i"),
		label.PadX(10),
		label.PadY(5),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// Button bar at bottom via AddSeeDismiss.
	btnFrame := AddSeeDismiss(app, nil)
	pack.Pack(btnFrame, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Configure handler.
	app.Dispatcher().Bind(root.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			root.Width = ev.ConfigWidth
			root.Height = ev.ConfigHeight
			pack.ArrangeContainer(root)
		}
	})

	// Expose handler.
	app.Dispatcher().Bind(root.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		d := root.Display.Server
		gc := root.GC
		d.SetForeground(gc, root.BackgroundPixel)
		d.FillRectangle(root.Drawable(), gc, 0, 0, uint(root.Width), uint(root.Height))
		d.Flush()
	})

	// Escape to quit.
	app.Dispatcher().BindGlobal(event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == platform.XK_Escape {
			app.Quit()
		}
	})

	return app
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

// PositionWindow sets the window position, matching Tk's positionWindow proc.
// positionWindow $w → wm geometry $w +300+300
func PositionWindow(t *toplevel.Toplevel) {
	t.WmInfo.SetGeometry("+300+300")
}

func AddSeeDismiss(parent widget.Caregiver, vars *DemoVars) *ttk.Frame {
	_, callerFile, _, _ := runtime.Caller(1)

	btnFrame := ttk.NewFrame(parent, "bottom_buttons")

	sep := ttk.NewSeparator(btnFrame, "sep")
	grid.Grid(sep, grid.ColumnSpan(4), grid.Row(0), grid.Sticky(grid.EW), grid.PadY("1.5p"))

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

	if vars != nil {
		varBtn := ttk.NewButton(btnFrame, "vars",
			ttk.ButtonText("See Variables"),
			ttk.ButtonImage(img["view"]),
			ttk.ButtonCompound(widget.CompoundLeft),
			ttk.ButtonCommand(func() { showVars(parent.AppContext(), vars) }),
		)
		buttons = append(buttons, varBtn)
		buttons[1], buttons[len(buttons)-1] = buttons[len(buttons)-1], buttons[1]
	}

	grid.Grid(geometry.Group(buttons), grid.PadX("3p"), grid.PadY("3p"))
	grid.ColumnConfigure(btnFrame, 0, grid.Weight(1))

	return btnFrame
}

// makeViewIcon creates a 16x16 magnifying glass icon (matches Tk's ::img::view).
func makeViewIcon() *tkimage.Photo {
	const sz = 16
	img := goimage.NewRGBA(goimage.Rect(0, 0, sz, sz))
	black := color.RGBA{R: 0, G: 0, B: 0, A: 255}

	// Draw circle: center (6,6), radius 5.
	cx, cy, r := 6.0, 6.0, 5.0
	for y := range sz {
		for x := range sz {
			dx := float64(x) + 0.5 - cx
			dy := float64(y) + 0.5 - cy
			dist := math.Sqrt(dx*dx + dy*dy)
			d := math.Abs(dist - r)
			if d < 1.0 {
				a := uint8((1.0 - d) * 255)
				img.SetRGBA(x, y, color.RGBA{0, 0, 0, a})
			}
		}
	}
	// Draw handle: thick diagonal from (10,10) to (14,14).
	for i := 0; i < 5; i++ {
		fi := float64(i)
		px, py := 10+int(fi), 10+int(fi)
		for dx := -1; dx <= 0; dx++ {
			for dy := -1; dy <= 0; dy++ {
				nx, ny := px+dx, py+dy
				if nx >= 0 && nx < sz && ny >= 0 && ny < sz {
					img.SetRGBA(nx, ny, black)
				}
			}
		}
	}
	return tkimage.NewPhoto("_dh_view", img)
}

// makeDeleteIcon creates a 16x16 red X icon (matches Tk's ::img::delete).
func makeDeleteIcon() *tkimage.Photo {
	const sz = 16
	img := goimage.NewRGBA(goimage.Rect(0, 0, sz, sz))
	red := color.RGBA{R: 208, G: 0, B: 0, A: 255}

	// Draw two diagonal lines forming an X.
	// Line 1: top-left to bottom-right (3,3)→(12,12)
	// Line 2: top-right to bottom-left (12,3)→(3,12)
	for i := 0; i < 10; i++ {
		for t := 0; t <= 1; t++ {
			// Line 1.
			img.SetRGBA(3+i, 3+i+t, red)
			img.SetRGBA(3+i+t, 3+i, red)
			// Line 2.
			img.SetRGBA(12-i, 3+i+t, red)
			img.SetRGBA(12-i-t, 3+i, red)
		}
	}
	return tkimage.NewPhoto("_dh_delete", img)
}

// codeWindow is the single reusable "See Code" toplevel (nil until first use).
var codeWindow *toplevel.Toplevel
var codeText *text.TextWidget

// showCode opens (or raises) a toplevel window displaying the demo source.
// Uses TTK widgets for the button bar (matching Tk's showCode proc).
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
	sb.Command = func(args ...any) {
		if len(args) < 2 {
			return
		}
		action, _ := args[0].(string)
		switch action {
		case "moveto":
			if f, ok := args[1].(float64); ok {
				codeText.YViewMoveTo(f)
			}
		case "scroll":
			if len(args) >= 3 {
				n, _ := args[1].(int)
				unit, _ := args[2].(string)
				codeText.YViewScroll(n, unit == "pages")
			}
		}
	}

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
// Matches Tk's showVars proc from tk/library/demos/widget.
//
//	proc showVars {w args} {
//	    catch {destroy $w}
//	    toplevel $w
//	    wm title $w "Variable values"
//	    ...
//	}
func showVars(app widget.AppContext, vars *DemoVars) {
	// catch {destroy $w}
	if varsWindow != nil && !varsWindow.Destroyed {
		varsWindow.Destroy()
	}

	// toplevel $w
	// wm title $w "Variable values"
	varsWindow = toplevel.New(app, "vars",
		toplevel.Title("Variable values"),
		toplevel.Background("#d9d9d9"),
	)
	varsWindow.Show()

	varsRoot := varsWindow.Window()

	b := ttk.NewFrame(varsWindow, "frame")
	grid.Grid(b, grid.Sticky(grid.NSEW))

	f := labelframe.New(b, "title", labelframe.Text("Variable values:"))

	names := make([]string, 0, len(*vars))
	for name := range *vars {
		names = append(names, name)
	}
	sort.Strings(names)

	for row, name := range names {
		v := (*vars)[name]
		nameLabel := ttk.NewLabel(f, "n_"+name, ttk.LabelText(name+":"))
		// TODO: textvariable support for live updates
		valLabel := ttk.NewLabel(f, "v_"+name, ttk.LabelText(fmt.Sprintf("%v", v.Get())))
		grid.Grid(nameLabel, grid.Column(0), grid.Row(row),
			grid.PadX("1.5p"), grid.PadY("1.5p"), grid.Sticky(grid.StickW))
		grid.Grid(valLabel, grid.Column(1), grid.Row(row),
			grid.PadX("1.5p"), grid.PadY("1.5p"), grid.Sticky(grid.StickW))
	}

	okBtn := ttk.NewButton(b, "ok",
		ttk.ButtonText("OK"),
		ttk.ButtonCommand(func() { varsWindow.Destroy() }),
	)

	// TODO: bind $w <Return> [list $b.ok invoke]
	// TODO: bind $w <Escape> [list $b.ok invoke]

	grid.Grid(f, grid.Sticky(grid.NSEW), grid.PadX("3p"))
	grid.Grid(okBtn, grid.Row(1), grid.Sticky(grid.StickE), grid.PadX("3p"), grid.PadY("3p"))

	grid.ColumnConfigure(f.Window(), 1, grid.Weight(1))
	grid.RowConfigure(f.Window(), 100, grid.Weight(1))
	grid.ColumnConfigure(b.Window(), 0, grid.Weight(1))
	grid.RowConfigure(b.Window(), 0, grid.Weight(1))
	grid.ColumnConfigure(varsRoot, 0, grid.Weight(1))
	grid.RowConfigure(varsRoot, 0, grid.Weight(1))

	app.Dispatcher().Bind(varsRoot.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			varsRoot.Width = ev.ConfigWidth
			varsRoot.Height = ev.ConfigHeight
			grid.ArrangeContainer(varsRoot)
		}
	})

	varsWindow.OnClose(func() { varsWindow.Destroy() })
}
