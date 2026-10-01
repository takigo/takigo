// Demo: Buttons that change the window background color.
// Ported from Tk's button.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Button Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("button"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength(screenunit.In(4)),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("If you click on any of the four buttons below, the background of the button area will change to the color indicated in the button.  You can press Tab to move among the buttons, then press Space to invoke the current button."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Color-changing function: change the frame and label backgrounds.
	// Buttons keep their own default background (matching Tk's behavior).
	changeColor := func(colorName string) {
		f.Configure(frame.Background(colorName))
		msg.Configure(label.Background(colorName))
	}

	// Color buttons — match Tk's button.tcl (X11 named colors, width 10).
	colors := []struct {
		text  string
		color string
	}{
		{"Peach Puff", "PeachPuff1"},
		{"Light Blue", "LightBlue1"},
		{"Sea Green", "SeaGreen2"},
		{"Yellow", "Yellow1"},
	}

	for _, c := range colors {
		btn := button.New(f, "btn_"+c.text,
			button.Text(c.text),
			button.Width(10),
			button.Command(func() { changeColor(c.color) }),
		)
		pack.Pack(btn, pack.SideOpt(pack.Top), pack.Expand(true), pack.PadY(screenunit.Pt(1.5)))
	}

	app.Run()
}
