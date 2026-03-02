// Demo: Radiobutton groups for selecting point size and color.
// Ported from Tk's radio.tcl demo.
package main

import (
	"fmt"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/radiobutton"
)

func main() {
	app := demohelper.Setup("Radiobutton Demonstration", 500, 400, "Two groups of radiobuttons are displayed below. Click on\na button to select it. The current selection is shown at\nthe bottom.")

	// Status label.
	statusLabel := label.New(app, "status",
		label.Text("Point size: 10, Color: red"),
		label.Anchor(option.AnchorW),
		label.PadX(10),
	)
	pack.Pack(statusLabel, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))

	// Variables.
	sizeVar := widget.NewVariable("10")
	colorVar := widget.NewVariable("red")

	updateStatus := func() {
		statusLabel.Text = fmt.Sprintf("Point size: %s, Color: %s", sizeVar.Get(), colorVar.Get())
		statusLabel.Display()
	}

	// Container for two groups side by side.
	groupFrame := frame.New(app, "groups")
	pack.Pack(groupFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Point size group.
	sizeFrame := frame.New(groupFrame, "sizes",
		frame.BorderWidth(2), frame.Relief(option.ReliefGroove))
	pack.Pack(sizeFrame, pack.SideOpt(pack.Left), pack.PadX(10), pack.PadY(5))

	sizeTitle := label.New(sizeFrame, "sizetitle",
		label.Text("Point Size"),
		label.PadX(5),
	)
	pack.Pack(sizeTitle, pack.SideOpt(pack.Top), pack.PadY(5))

	sizes := []string{"10", "12", "14", "18", "24"}
	for _, s := range sizes {
		rb := radiobutton.New(sizeFrame, "size_"+s,
			radiobutton.Text(s+" point"),
			radiobutton.Value(s),
			radiobutton.Var(sizeVar),
			radiobutton.Command(updateStatus),
		)
		pack.Pack(rb, pack.SideOpt(pack.Top), pack.PadY(2), pack.Anchor(option.AnchorW), pack.PadX(10))
		_ = rb
	}

	// Color group.
	colorFrame := frame.New(groupFrame, "colors",
		frame.BorderWidth(2), frame.Relief(option.ReliefGroove))
	pack.Pack(colorFrame, pack.SideOpt(pack.Left), pack.PadX(10), pack.PadY(5))

	colorTitle := label.New(colorFrame, "colortitle",
		label.Text("Color"),
		label.PadX(5),
	)
	pack.Pack(colorTitle, pack.SideOpt(pack.Top), pack.PadY(5))

	colors := []string{"Red", "Orange", "Yellow", "Green", "Blue"}
	for _, c := range colors {
		val := c
		rb := radiobutton.New(colorFrame, "color_"+c,
			radiobutton.Text(c),
			radiobutton.Value(c),
			radiobutton.Var(colorVar),
			radiobutton.Command(updateStatus),
		)
		pack.Pack(rb, pack.SideOpt(pack.Top), pack.PadY(2), pack.Anchor(option.AnchorW), pack.PadX(10))
		_ = rb
		_ = val
	}

	_ = statusLabel
	_ = sizeTitle
	_ = colorTitle
	app.Run()
}
