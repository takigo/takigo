// Demo: Iconic buttons with checkbuttons and radiobuttons using images.
// Ported from Tk's icon.tcl demo.
package main

import (
	"fmt"
	goimage "image"
	"image/color"
	"os"

	"github.com/msorc/takigo"
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

func makeFlagImage(name string, up bool) *tkimage.Photo {
	const w, h = 40, 40
	img := goimage.NewRGBA(goimage.Rect(0, 0, w, h))
	bg := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	pole := color.RGBA{R: 80, G: 60, B: 40, A: 255}
	flag := color.RGBA{R: 200, G: 40, B: 40, A: 255}
	for y := range h {
		for x := range w {
			img.SetRGBA(x, y, bg)
		}
	}
	// Pole.
	for y := 4; y < h-2; y++ {
		img.SetRGBA(8, y, pole)
		img.SetRGBA(9, y, pole)
	}
	// Flag (raised: near top; lowered: near bottom).
	fy := 4
	if !up {
		fy = 20
	}
	for y := fy; y < fy+14; y++ {
		for x := 10; x < 30; x++ {
			img.SetRGBA(x, y, flag)
		}
	}
	return tkimage.NewPhoto(name, img)
}

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

func main() {
	app, err := takigo.NewApp(takigo.Title("Iconic Button Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("icon"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength("5i"),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("This window shows three ways of using bitmaps or images in "+
			"radiobuttons and checkbuttons. On the left are two "+
			"radiobuttons, each of which displays a bitmap and an "+
			"indicator. In the middle is a checkbutton that displays a "+
			"different image depending on whether it is selected or not. "+
			"On the right is a checkbutton that displays a single bitmap "+
			"but changes its background color to indicate whether or not "+
			"it is selected."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Create flag images.
	flagUp := makeFlagImage("flag_up", true)
	flagDown := makeFlagImage("flag_down", false)
	app.ImageRegistry().Register(flagUp)
	app.ImageRegistry().Register(flagDown)

	// Create color square images.
	redImg := makeColorSquare("sq_red", color.RGBA{R: 220, G: 40, B: 40, A: 255}, 30)
	greenImg := makeColorSquare("sq_green", color.RGBA{R: 40, G: 180, B: 40, A: 255}, 30)
	app.ImageRegistry().Register(redImg)
	app.ImageRegistry().Register(greenImg)

	// Shared variable.
	flagVar := widget.NewVariable(false)

	// Outer frame to hold three columns.
	outer := frame.New(f, "outer")
	pack.Pack(outer, pack.SideOpt(pack.Top), pack.Expand(true), pack.FillOpt(pack.FillBoth))

	// -- Left column: radiobuttons with images + indicators --
	leftFrame := frame.New(outer, "left",
		frame.BorderWidth(2), frame.Relief(option.ReliefGroove),
	)
	pack.Pack(leftFrame, pack.SideOpt(pack.Left), pack.PadX(10), pack.PadY(10), pack.Expand(true))

	ltitle := label.New(leftFrame, "ltitle",
		label.Text("Radiobuttons\nwith images"),
		label.Anchor(option.AnchorCenter),
	)
	pack.Pack(ltitle, pack.SideOpt(pack.Top), pack.PadY(4))
	// Use a string variable for radiobuttons (separate from flagVar).
	rbVar := widget.NewVariable("down")
	rbUp := radiobutton.New(leftFrame, "rb_up",
		radiobutton.Text("Flag Up"),
		radiobutton.Value("up"),
		radiobutton.Var(rbVar),
	)
	rbDown := radiobutton.New(leftFrame, "rb_down",
		radiobutton.Text("Flag Down"),
		radiobutton.Value("down"),
		radiobutton.Var(rbVar),
	)
	pack.Pack(rbUp, pack.SideOpt(pack.Top), pack.PadY(4), pack.Anchor(option.AnchorW))
	pack.Pack(rbDown, pack.SideOpt(pack.Top), pack.PadY(4), pack.Anchor(option.AnchorW))

	// -- Middle column: checkbutton with selectimage --
	midFrame := frame.New(outer, "mid",
		frame.BorderWidth(2), frame.Relief(option.ReliefGroove),
	)
	pack.Pack(midFrame, pack.SideOpt(pack.Left), pack.PadX(10), pack.PadY(10), pack.Expand(true))

	mtitle := label.New(midFrame, "mtitle",
		label.Text("Checkbutton\nselectimage"),
		label.Anchor(option.AnchorCenter),
	)
	pack.Pack(mtitle, pack.SideOpt(pack.Top), pack.PadY(4))
	// Middle checkbutton: flagDown normally, flagUp when selected; no indicator.
	cbMid := checkbutton.New(midFrame, "cb_mid",
		checkbutton.ImageOpt(flagDown),
		checkbutton.SelectImageOpt(flagUp),
		checkbutton.IndicatorOnOpt(false),
		checkbutton.Var(flagVar),
	)
	pack.Pack(cbMid, pack.SideOpt(pack.Top), pack.PadY(8))

	// -- Right column: checkbutton changing background via selectcolor --
	rightFrame := frame.New(outer, "right",
		frame.BorderWidth(2), frame.Relief(option.ReliefGroove),
	)
	pack.Pack(rightFrame, pack.SideOpt(pack.Left), pack.PadX(10), pack.PadY(10), pack.Expand(true))

	rtitle := label.New(rightFrame, "rtitle",
		label.Text("Checkbutton\ncolor squares"),
		label.Anchor(option.AnchorCenter),
	)
	pack.Pack(rtitle, pack.SideOpt(pack.Top), pack.PadY(4))
	// Right checkbutton: single color-square image, changes background on select.
	colorVar := widget.NewVariable(false)
	cbRight := checkbutton.New(rightFrame, "cb_right",
		checkbutton.ImageOpt(redImg),
		checkbutton.SelectImageOpt(greenImg),
		checkbutton.IndicatorOnOpt(false),
		checkbutton.Var(colorVar),
		checkbutton.Text("Toggle"),
	)
	pack.Pack(cbRight, pack.SideOpt(pack.Top), pack.PadY(8))

	_ = rbVar
	_ = rbUp
	_ = rbDown
	_ = cbMid
	_ = cbRight
	_ = flagVar
	_ = colorVar
	app.Run()
}
