// Demo: Photo images in labels.
// Ported from Tk's image1.tcl demo.
package main

import (
	"fmt"
	goimage "image"
	"image/color"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	tkimage "github.com/msorc/takigo/image"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Image Demo"), takigo.Size(500, 400))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	msg := label.New(root, "msg", app,
		label.Text("Photo images displayed in labels.\nImages are generated at runtime."),
		label.Anchor(option.AnchorW),
		label.PadX(10), label.PadY(5),
	)
	pack.Pack(msg.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	btnFrame := frame.New(root, "btnframe", app)
	pack.Pack(btnFrame.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))
	dismissBtn := button.New(btnFrame.Window(), "dismiss", app,
		button.Text("Dismiss"), button.Command(func() { app.Quit() }),
		button.PadX(10), button.PadY(4),
	)
	pack.Pack(dismissBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(10))

	// Generate test images.
	makeGradient := func(name string, c1, c2 color.RGBA, w, h int) *tkimage.Photo {
		img := goimage.NewRGBA(goimage.Rect(0, 0, w, h))
		for y := range h {
			t := float64(y) / float64(h)
			for x := range w {
				r := uint8(float64(c1.R)*(1-t) + float64(c2.R)*t)
				g := uint8(float64(c1.G)*(1-t) + float64(c2.G)*t)
				b := uint8(float64(c1.B)*(1-t) + float64(c2.B)*t)
				img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
		return tkimage.NewPhoto(name, img)
	}

	makeCheckerboard := func(name string, size, squares int) *tkimage.Photo {
		img := goimage.NewRGBA(goimage.Rect(0, 0, size, size))
		sq := size / squares
		for y := range size {
			for x := range size {
				if ((x/sq)+(y/sq))%2 == 0 {
					img.SetRGBA(x, y, color.RGBA{R: 200, G: 200, B: 200, A: 255})
				} else {
					img.SetRGBA(x, y, color.RGBA{R: 100, G: 100, B: 100, A: 255})
				}
			}
		}
		return tkimage.NewPhoto(name, img)
	}

	// Create images.
	gradient1 := makeGradient("grad1",
		color.RGBA{R: 255, G: 100, B: 100, A: 255},
		color.RGBA{R: 100, G: 100, B: 255, A: 255}, 128, 128)
	app.ImageRegistry().Register(gradient1)

	gradient2 := makeGradient("grad2",
		color.RGBA{R: 50, G: 200, B: 50, A: 255},
		color.RGBA{R: 200, G: 200, B: 50, A: 255}, 128, 128)
	app.ImageRegistry().Register(gradient2)

	checker := makeCheckerboard("checker", 128, 8)
	app.ImageRegistry().Register(checker)

	// Display images in labels.
	imgFrame := frame.New(root, "imgframe", app)
	pack.Pack(imgFrame.Window(), pack.SideOpt(pack.Top), pack.PadX(10), pack.PadY(10))

	l1 := label.New(imgFrame.Window(), "img1", app,
		label.ImageOpt(gradient1),
		label.BorderWidth(2),
		label.Relief(option.ReliefGroove),
		label.PadX(4), label.PadY(4),
	)
	pack.Pack(l1.Window(), pack.SideOpt(pack.Left), pack.PadX(10))

	l2 := label.New(imgFrame.Window(), "img2", app,
		label.ImageOpt(gradient2),
		label.BorderWidth(2),
		label.Relief(option.ReliefGroove),
		label.PadX(4), label.PadY(4),
	)
	pack.Pack(l2.Window(), pack.SideOpt(pack.Left), pack.PadX(10))

	l3 := label.New(imgFrame.Window(), "img3", app,
		label.ImageOpt(checker),
		label.BorderWidth(2),
		label.Relief(option.ReliefGroove),
		label.PadX(4), label.PadY(4),
	)
	pack.Pack(l3.Window(), pack.SideOpt(pack.Left), pack.PadX(10))

	// Labels.
	labFrame := frame.New(root, "labframe", app)
	pack.Pack(labFrame.Window(), pack.SideOpt(pack.Top), pack.PadX(10), pack.PadY(5))

	for _, name := range []string{"Red→Blue Gradient", "Green→Yellow Gradient", "Checkerboard"} {
		ll := label.New(labFrame.Window(), "lab_"+name, app,
			label.Text(name),
		)
		pack.Pack(ll.Window(), pack.SideOpt(pack.Left), pack.PadX(20))
		_ = ll
	}

	// Root events.
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
		d.SetForeground(root.GC, bgColor.Pixel)
		d.FillRectangle(root.Drawable(), root.GC, 0, 0, uint(root.Width), uint(root.Height))
		d.Flush()
	})
	app.Dispatcher().BindGlobal(event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == xlib.XK_Escape {
			app.Quit()
		}
	})

	_ = msg
	_ = dismissBtn
	_ = l1
	_ = l2
	_ = l3
	app.MainLoop()
}
