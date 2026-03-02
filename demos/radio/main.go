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
	d := demohelper.Setup("Radiobutton Demonstration", 500, 400, "Two groups of radiobuttons are displayed below. Click on\na button to select it. The current selection is shown at\nthe bottom.")
	defer d.App.Destroy()
	root, app := d.Root, d.App

	// Status label.
	statusLabel := label.New(root, "status", app,
		label.Text("Point size: 10, Color: red"),
		label.Anchor(option.AnchorW),
		label.PadX(10),
	)
	pack.Pack(statusLabel.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))

	// Variables.
	sizeVar := widget.NewVariable("10")
	colorVar := widget.NewVariable("red")

	updateStatus := func() {
		statusLabel.Text = fmt.Sprintf("Point size: %s, Color: %s", sizeVar.Get(), colorVar.Get())
		statusLabel.Display()
	}

	// Container for two groups side by side.
	groupFrame := frame.New(root, "groups", app)
	pack.Pack(groupFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Point size group.
	sizeFrame := frame.New(groupFrame.Window(), "sizes", app,
		frame.BorderWidth(2), frame.Relief(option.ReliefGroove))
	pack.Pack(sizeFrame.Window(), pack.SideOpt(pack.Left), pack.PadX(10), pack.PadY(5))

	sizeTitle := label.New(sizeFrame.Window(), "sizetitle", app,
		label.Text("Point Size"),
		label.PadX(5),
	)
	pack.Pack(sizeTitle.Window(), pack.SideOpt(pack.Top), pack.PadY(5))

	sizes := []string{"10", "12", "14", "18", "24"}
	for _, s := range sizes {
		rb := radiobutton.New(sizeFrame.Window(), "size_"+s, app,
			radiobutton.Text(s+" point"),
			radiobutton.Value(s),
			radiobutton.Var(sizeVar),
			radiobutton.Command(updateStatus),
		)
		pack.Pack(rb.Window(), pack.SideOpt(pack.Top), pack.PadY(2), pack.Anchor(option.AnchorW), pack.PadX(10))
		_ = rb
	}

	// Color group.
	colorFrame := frame.New(groupFrame.Window(), "colors", app,
		frame.BorderWidth(2), frame.Relief(option.ReliefGroove))
	pack.Pack(colorFrame.Window(), pack.SideOpt(pack.Left), pack.PadX(10), pack.PadY(5))

	colorTitle := label.New(colorFrame.Window(), "colortitle", app,
		label.Text("Color"),
		label.PadX(5),
	)
	pack.Pack(colorTitle.Window(), pack.SideOpt(pack.Top), pack.PadY(5))

	colors := []string{"Red", "Orange", "Yellow", "Green", "Blue"}
	for _, c := range colors {
		val := c
		rb := radiobutton.New(colorFrame.Window(), "color_"+c, app,
			radiobutton.Text(c),
			radiobutton.Value(c),
			radiobutton.Var(colorVar),
			radiobutton.Command(updateStatus),
		)
		pack.Pack(rb.Window(), pack.SideOpt(pack.Top), pack.PadY(2), pack.Anchor(option.AnchorW), pack.PadX(10))
		_ = rb
		_ = val
	}

	_ = statusLabel
	_ = sizeTitle
	_ = colorTitle
	d.Run()
}
