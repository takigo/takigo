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

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/event"
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
	"github.com/msorc/takigo/widget/scrollbar"
	"github.com/msorc/takigo/widget/text"
	"github.com/msorc/takigo/widget/toplevel"
)

// Setup creates a standard demo window with a description label, and
// "See Code" / "Dismiss" buttons at the bottom (matching Tk's addSeeDismiss).
// Calls os.Exit(1) on failure.
func Setup(title string, width, height int, description string) *takigo.App {
	// Capture the caller's source file for "See Code".
	_, callerFile, _, _ := runtime.Caller(1)

	app, err := takigo.NewApp(takigo.Title(title), takigo.Size(width, height))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	// Description label (classic label, same as Tk's "label $w.msg").
	msg := label.New(app, "msg",
		label.Text(description),
		label.Anchor(option.AnchorW),
		label.PadX(10),
		label.PadY(5),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// Button bar at bottom — uses TTK widgets matching Tk's addSeeDismiss:
	// ttk::frame, ttk::separator, ttk::button.
	btnFrame := ttk.NewFrame(app, "btnframe")
	pack.Pack(btnFrame, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// ttk::separator.
	sep := ttk.NewSeparator(btnFrame, "sep")
	pack.Pack(sep, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(2))

	// Create button icons.
	viewIcon := makeViewIcon()
	deleteIcon := makeDeleteIcon()
	app.ImageRegistry().Register(viewIcon)
	app.ImageRegistry().Register(deleteIcon)

	// Pack buttons right-to-left so they appear right-aligned:
	// Dismiss (rightmost), then See Code.
	dismissBtn := ttk.NewButton(btnFrame, "dismiss",
		ttk.ButtonText("Dismiss"),
		ttk.ButtonImage(deleteIcon),
		ttk.ButtonCompound(widget.CompoundLeft),
		ttk.ButtonCommand(func() { app.Quit() }),
	)
	pack.Pack(dismissBtn, pack.SideOpt(pack.Right), pack.PadX(4), pack.PadY(4))

	seeCodeBtn := ttk.NewButton(btnFrame, "seecode",
		ttk.ButtonText("See Code"),
		ttk.ButtonImage(viewIcon),
		ttk.ButtonCompound(widget.CompoundLeft),
		ttk.ButtonCommand(func() { showCode(app, callerFile) }),
	)
	pack.Pack(seeCodeBtn, pack.SideOpt(pack.Right), pack.PadX(4), pack.PadY(4))

	// Configure handler.
	app.Dispatcher().Bind(root.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			root.Width = ev.ConfigWidth
			root.Height = ev.ConfigHeight
			pack.ArrangeContainer(root)
		}
	})

	// Expose handler — reads root.BackgroundPixel so dynamic bg changes work.
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
func showCode(app *takigo.App, srcFile string) {
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

// NewFrame creates a plain frame as a child of the given parent.
func NewFrame(parent widget.Caregiver, name string) *frame.Frame {
	return frame.New(parent, name)
}

// DemoDir returns the absolute path to a demo directory by name,
// relative to the demos/ root found via the caller's source file location.
func DemoDir(name string) string {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		return name
	}
	// Walk up from callers' source file to find the demos/ root.
	// demos/demohelper/demohelper.go → demos/ is one level up.
	demosRoot := filepath.Dir(filepath.Dir(file))
	return filepath.Join(demosRoot, name)
}
