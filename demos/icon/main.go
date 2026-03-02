// Demo: Iconic buttons with checkbuttons and radiobuttons using images.
// Ported from Tk's icon.tcl demo.
package main

import (
	"fmt"
	goimage "image"
	"image/color"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	tkimage "github.com/msorc/takigo/image"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/checkbutton"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/radiobutton"
)

func makeColorSquare(name string, c color.RGBA, sz int) *tkimage.Photo {
	img := goimage.NewRGBA(goimage.Rect(0, 0, sz, sz))
	border := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	for y := range sz {
		for x := range sz {
			if x < 2 || x >= sz-2 || y < 2 || y >= sz-2 {
				img.SetRGBA(x, y, border)
			} else {
				img.SetRGBA(x, y, c)
			}
		}
	}
	return tkimage.NewPhoto(name, img)
}

func makeFlagImage(name string, up bool) *tkimage.Photo {
	const w, h = 32, 32
	img := goimage.NewRGBA(goimage.Rect(0, 0, w, h))
	bg := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	pole := color.RGBA{R: 80, G: 60, B: 40, A: 255}
	flag := color.RGBA{R: 200, G: 40, B: 40, A: 255}
	// Background.
	for y := range h {
		for x := range w {
			img.SetRGBA(x, y, bg)
		}
	}
	// Pole.
	for y := 4; y < h-2; y++ {
		img.SetRGBA(6, y, pole)
		img.SetRGBA(7, y, pole)
	}
	// Flag.
	fy := 4
	if !up {
		fy = 16
	}
	for y := fy; y < fy+12; y++ {
		for x := 8; x < 24; x++ {
			img.SetRGBA(x, y, flag)
		}
	}
	return tkimage.NewPhoto(name, img)
}

func main() {
	app := demohelper.Setup("Iconic Button Demonstration", 450, 400,
		"This demo shows checkbuttons and radiobuttons with\ncolored icon images instead of text labels.")

	statusLabel := label.New(app, "status",
		label.Text("Flag: down, Color: red"),
		label.Anchor(option.AnchorW),
		label.PadX(10),
	)
	pack.Pack(statusLabel, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))

	// Create images.
	flagUp := makeFlagImage("flag_up", true)
	flagDown := makeFlagImage("flag_down", false)
	app.ImageRegistry().Register(flagUp)
	app.ImageRegistry().Register(flagDown)

	redImg := makeColorSquare("sq_red", color.RGBA{R: 220, G: 40, B: 40, A: 255}, 24)
	greenImg := makeColorSquare("sq_green", color.RGBA{R: 40, G: 180, B: 40, A: 255}, 24)
	blueImg := makeColorSquare("sq_blue", color.RGBA{R: 40, G: 40, B: 220, A: 255}, 24)
	yellowImg := makeColorSquare("sq_yellow", color.RGBA{R: 220, G: 200, B: 40, A: 255}, 24)
	app.ImageRegistry().Register(redImg)
	app.ImageRegistry().Register(greenImg)
	app.ImageRegistry().Register(blueImg)
	app.ImageRegistry().Register(yellowImg)

	// Variables.
	flagVar := widget.NewVariable(false)
	colorVar := widget.NewVariable("red")

	// Flag display label.
	flagLabel := label.New(app, "flagdisp",
		label.ImageOpt(flagDown),
		label.BorderWidth(2), label.Relief(option.ReliefSunken),
		label.PadX(4), label.PadY(4),
	)
	pack.Pack(flagLabel, pack.SideOpt(pack.Top), pack.PadY(10))

	updateStatus := func() {
		state := "down"
		if flagVar.Get() {
			state = "up"
		}
		statusLabel.Text = fmt.Sprintf("Flag: %s, Color: %s", state, colorVar.Get())
		statusLabel.Display()
	}

	// Checkbutton for flag.
	cb := checkbutton.New(app, "flagcb",
		checkbutton.Text("Raise Flag"),
		checkbutton.Var(flagVar),
		checkbutton.Command(func() {
			if flagVar.Get() {
				flagLabel.Img = flagUp
			} else {
				flagLabel.Img = flagDown
			}
			flagLabel.Display()
			updateStatus()
		}),
	)
	pack.Pack(cb, pack.SideOpt(pack.Top), pack.PadY(5))

	// Radiobutton group for colors.
	colorFrame := frame.New(app, "colors",
		frame.BorderWidth(2), frame.Relief(option.ReliefGroove),
	)
	pack.Pack(colorFrame, pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(10))

	colorTitle := label.New(colorFrame, "ctitle", label.Text("Select Color"))
	pack.Pack(colorTitle, pack.SideOpt(pack.Top), pack.PadY(5))

	type colorDef struct {
		name  string
		image *tkimage.Photo
	}
	colorDefs := []colorDef{
		{"red", redImg},
		{"green", greenImg},
		{"blue", blueImg},
		{"yellow", yellowImg},
	}

	for _, cd := range colorDefs {
		rb := radiobutton.New(colorFrame, "rb_"+cd.name,
			radiobutton.Text(cd.name),
			radiobutton.Value(cd.name),
			radiobutton.Var(colorVar),
			radiobutton.Command(updateStatus),
		)
		pack.Pack(rb, pack.SideOpt(pack.Top), pack.PadY(2),
			pack.Anchor(option.AnchorW), pack.PadX(10))
		_ = rb
	}

	_ = statusLabel
	_ = flagLabel
	_ = cb
	_ = colorTitle
	app.Run()
}
