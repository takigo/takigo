// Demo: Image viewer — select from generated images.
// Ported from Tk's image2.tcl demo (simplified — no file browsing).
package main

import (
	goimage "image"
	"image/color"
	"math"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	tkimage "github.com/msorc/takigo/image"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/listbox"
	"github.com/msorc/takigo/widget/scrollbar"
)

func main() {
	d := demohelper.Setup("Image Viewer", 550, 400,
		"Select an image from the list to preview it.")
	root, app := d.Root, d.App

	// Generate several images.
	type imgEntry struct {
		name  string
		photo *tkimage.Photo
	}

	makeCircles := func(name string, size int) *tkimage.Photo {
		img := goimage.NewRGBA(goimage.Rect(0, 0, size, size))
		cx, cy := float64(size)/2, float64(size)/2
		for y := range size {
			for x := range size {
				dx := float64(x) - cx
				dy := float64(y) - cy
				dist := math.Sqrt(dx*dx + dy*dy)
				ring := int(dist/10) % 3
				var c color.RGBA
				switch ring {
				case 0:
					c = color.RGBA{R: 220, G: 60, B: 60, A: 255}
				case 1:
					c = color.RGBA{R: 60, G: 180, B: 60, A: 255}
				default:
					c = color.RGBA{R: 60, G: 60, B: 220, A: 255}
				}
				img.SetRGBA(x, y, c)
			}
		}
		return tkimage.NewPhoto(name, img)
	}

	makeStripes := func(name string, size int) *tkimage.Photo {
		img := goimage.NewRGBA(goimage.Rect(0, 0, size, size))
		for y := range size {
			for x := range size {
				stripe := (x + y) / 10 % 4
				var c color.RGBA
				switch stripe {
				case 0:
					c = color.RGBA{R: 255, G: 200, B: 200, A: 255}
				case 1:
					c = color.RGBA{R: 200, G: 255, B: 200, A: 255}
				case 2:
					c = color.RGBA{R: 200, G: 200, B: 255, A: 255}
				default:
					c = color.RGBA{R: 255, G: 255, B: 200, A: 255}
				}
				img.SetRGBA(x, y, c)
			}
		}
		return tkimage.NewPhoto(name, img)
	}

	images := []imgEntry{
		{"Concentric Circles", makeCircles("circles", 160)},
		{"Diagonal Stripes", makeStripes("stripes", 160)},
	}

	for _, e := range images {
		app.ImageRegistry().Register(e.photo)
	}

	// Layout: listbox on left, preview on right.
	mainFrame := frame.New(root, "mainframe", app)
	pack.Pack(mainFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Listbox.
	lbFrame := frame.New(mainFrame.Window(), "lbframe", app)
	pack.Pack(lbFrame.Window(), pack.SideOpt(pack.Left), pack.FillOpt(pack.FillY), pack.PadX(5))

	names := make([]string, len(images))
	for i, e := range images {
		names[i] = e.name
	}

	lb := listbox.New(lbFrame.Window(), "imglist", app,
		listbox.Items(names...),
		listbox.Height(8),
		listbox.Width(20),
	)
	yscroll := scrollbar.New(lbFrame.Window(), "yscroll", app,
		scrollbar.OrientOpt(scrollbar.Vertical),
		scrollbar.CommandOpt(func(args ...any) {
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
	lb.YScrollCmd = func(first, last float64) { yscroll.Set(first, last) }
	pack.Pack(yscroll.Window(), pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(lb.Window(), pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Preview label.
	previewLabel := label.New(mainFrame.Window(), "preview", app,
		label.Text("(select an image)"),
		label.BorderWidth(2),
		label.Relief(option.ReliefSunken),
		label.PadX(10), label.PadY(10),
	)
	pack.Pack(previewLabel.Window(), pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10))

	// Selection handler — use bind engine for button click.
	app.Dispatcher().Bind(lb.Window().XWindow, event.ButtonPressMask, func(_ *event.Event) {
		app.DoWhenIdle(func() {
			sel := lb.Selection()
			if len(sel) > 0 && sel[0] < len(images) {
				previewLabel.Text = ""
				previewLabel.Img = images[sel[0]].photo
				previewLabel.Compound = widget.CompoundCenter
				previewLabel.Display()
			}
		})
	})

	_ = previewLabel
	d.Run()
}
