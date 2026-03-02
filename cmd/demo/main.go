// Phase 9 demo: TTK themed widgets alongside classic widgets.
package main

import (
	"fmt"
	goimage "image"
	"image/color"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/focus"
	"github.com/msorc/takigo/geometry/pack"
	tkimage "github.com/msorc/takigo/image"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/listbox"
	"github.com/msorc/takigo/widget/scale"
	"github.com/msorc/takigo/widget/scrollbar"
)

// generateTestImage creates a 64x64 RGBA image with colored quadrants.
func generateTestImage() *goimage.RGBA {
	img := goimage.NewRGBA(goimage.Rect(0, 0, 64, 64))
	colors := [4]color.RGBA{
		{R: 220, G: 50, B: 50, A: 255},  // top-left: red
		{R: 50, G: 150, B: 220, A: 255},  // top-right: blue
		{R: 50, G: 180, B: 80, A: 255},   // bottom-left: green
		{R: 220, G: 180, B: 50, A: 255},  // bottom-right: yellow
	}
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			qi := 0
			if x >= 32 {
				qi++
			}
			if y >= 32 {
				qi += 2
			}
			img.SetRGBA(x, y, colors[qi])
		}
	}
	return img
}

func main() {
	app, err := takigo.NewApp(takigo.Title("Takigo Phase 9 — TTK Themed Widgets"), takigo.Size(700, 550))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Set clam theme as default.
	ttk.SetCurrentTheme("clam")

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	// Focus manager.
	focusMgr := focus.NewManager(app.Dispatcher(), app.DisplayPtr())
	focusMgr.BindTraversal(root)

	// Title.
	titleLabel := label.New(root, "title", app,
		label.Text("Phase 9: TTK Themed Widgets"),
		label.PadX(10),
		label.PadY(5),
	)
	pack.Pack(titleLabel.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	// --- TTK section ---
	ttkFrame := ttk.NewFrame(root, "ttkFrame", app,
		ttk.FrameBorderWidth(2),
		ttk.FrameRelief(option.ReliefGroove),
	)
	pack.Pack(ttkFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadX(10), pack.PadY(5))

	ttkLabel := ttk.NewLabel(ttkFrame.Window(), "ttkLabel", app,
		ttk.LabelText("TTK Label (clam theme)"),
	)
	pack.Pack(ttkLabel.Window(), pack.SideOpt(pack.Left), pack.PadX(8), pack.PadY(4))

	ttkBtn := ttk.NewButton(ttkFrame.Window(), "ttkBtn", app,
		ttk.ButtonText("TTK Button"),
		ttk.ButtonCommand(func() {
			fmt.Println("TTK button clicked!")
		}),
	)
	pack.Pack(ttkBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(8), pack.PadY(4))

	// Separator.
	ttkSep := ttk.NewSeparator(root, "ttkSep", app)
	pack.Pack(ttkSep.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadX(10), pack.PadY(2))

	// --- Image section (from Phase 8) ---
	imgFrame := frame.New(root, "imgFrame", app)
	pack.Pack(imgFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadX(10), pack.PadY(5))

	// Generate a test image at runtime.
	testRGBA := generateTestImage()
	testPhoto := tkimage.NewPhoto("test", testRGBA)
	app.ImageRegistry().Register(testPhoto)

	// Label with image only.
	imgLabel := label.New(imgFrame.Window(), "imgLabel", app,
		label.ImageOpt(testPhoto),
		label.BorderWidth(2),
		label.Relief(option.ReliefGroove),
		label.PadX(4),
		label.PadY(4),
	)
	pack.Pack(imgLabel.Window(), pack.SideOpt(pack.Left), pack.PadX(5))

	// Button with image + text (compound left).
	imgBtn := button.New(imgFrame.Window(), "imgBtn", app,
		button.Text("Click Me"),
		button.ImageOpt(testPhoto),
		button.CompoundOpt(widget.CompoundLeft),
		button.PadX(8),
		button.PadY(4),
		button.Command(func() {
			fmt.Println("Image button clicked!")
		}),
	)
	pack.Pack(imgBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(5))

	// Optionally load a PNG from file if provided as argument.
	var imgPath string
	for _, arg := range os.Args[1:] {
		if arg != "--" {
			imgPath = arg
			break
		}
	}
	if imgPath != "" {
		filePhoto, err := tkimage.NewPhotoFromFile("file", imgPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not load %s: %v\n", imgPath, err)
		} else {
			app.ImageRegistry().Register(filePhoto)
			fileLabel := label.New(imgFrame.Window(), "fileLabel", app,
				label.ImageOpt(filePhoto),
				label.BorderWidth(2),
				label.Relief(option.ReliefSunken),
				label.PadX(4),
				label.PadY(4),
			)
			pack.Pack(fileLabel.Window(), pack.SideOpt(pack.Left), pack.PadX(5))
		}
	}

	// --- Classic widgets to show coexistence ---
	midFrame := frame.New(root, "midFrame", app)
	pack.Pack(midFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadX(10), pack.PadY(5))

	// Scale.
	scaleLabel := label.New(midFrame.Window(), "scaleLabel", app,
		label.Text("Scale: 50"),
		label.Anchor(option.AnchorW),
	)
	sc := scale.New(midFrame.Window(), "scale1", app,
		scale.OrientOpt(scale.Horizontal),
		scale.FromOpt(0),
		scale.ToOpt(100),
		scale.ValueOpt(50),
		scale.ShowValueOpt(true),
		scale.CommandOpt(func(v float64) {
			scaleLabel.Text = fmt.Sprintf("Scale: %.0f", v)
			scaleLabel.Display()
		}),
	)
	pack.Pack(sc.Window(), pack.SideOpt(pack.Left), pack.FillOpt(pack.FillX), pack.Expand(true), pack.PadX(5))
	pack.Pack(scaleLabel.Window(), pack.SideOpt(pack.Left), pack.PadX(5))

	// Listbox with scrollbar.
	lbFrame := frame.New(root, "lbFrame", app)
	pack.Pack(lbFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true), pack.PadX(10), pack.PadY(5))

	items := make([]string, 20)
	for i := range items {
		items[i] = fmt.Sprintf("Item %d", i+1)
	}

	lb := listbox.New(lbFrame.Window(), "listbox", app,
		listbox.Items(items...),
		listbox.Height(8),
		listbox.Width(30),
	)

	yscroll := scrollbar.New(lbFrame.Window(), "yscroll", app,
		scrollbar.OrientOpt(scrollbar.Vertical),
		scrollbar.WidthOpt(14),
		scrollbar.CommandOpt(func(args ...interface{}) {
			if len(args) < 1 {
				return
			}
			switch args[0] {
			case "moveto":
				if len(args) >= 2 {
					if f, ok := args[1].(float64); ok {
						lb.YViewMoveTo(f)
					}
				}
			case "scroll":
				if len(args) >= 3 {
					n, _ := args[1].(int)
					unit, _ := args[2].(string)
					lb.YViewScroll(n, unit == "pages")
				}
			}
		}),
	)
	lb.YScrollCmd = func(first, last float64) {
		yscroll.Set(first, last)
	}
	pack.Pack(yscroll.Window(), pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(lb.Window(), pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	first, last := lb.YVisibleRange()
	yscroll.Set(first, last)

	// Status label.
	statusLabel := label.New(root, "status", app,
		label.Text("Phase 9: TTK + Classic widgets. Esc to quit."),
		label.Background("#e8e8e8"),
		label.Anchor(option.AnchorW),
		label.PadX(5),
		label.PadY(2),
	)
	pack.Pack(statusLabel.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// --- Root event handlers ---
	app.Dispatcher().Bind(root.XWindow, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			root.Width = ev.ConfigWidth
			root.Height = ev.ConfigHeight
			pack.ArrangeContainer(root)
		}
	})

	app.Dispatcher().Bind(root.XWindow, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		d := root.Display.XDisplay
		gc := root.GC
		d.SetForeground(gc, bgColor.Pixel)
		d.FillRectangle(root.Drawable(), gc, 0, 0, uint(root.Width), uint(root.Height))
		d.Flush()
	})

	app.Dispatcher().BindGlobal(event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == xlib.XK_Escape {
			app.Quit()
		}
	})

	fmt.Println("Takigo Phase 9 Demo — TTK Themed Widgets")
	fmt.Println("Esc to quit.")
	app.Run()
	fmt.Println("Goodbye!")

	_ = titleLabel
	_ = statusLabel
	_ = focusMgr
	_ = imgLabel
	_ = imgBtn
	_ = scaleLabel
	_ = ttkLabel
	_ = ttkBtn
	_ = ttkSep
}
