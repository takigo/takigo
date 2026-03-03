// Demo: Radiobutton groups for selecting point size, color, and alignment.
// Ported from Tk's radio.tcl demo.
package main

import (
	"strings"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/labelframe"
	"github.com/msorc/takigo/widget/radiobutton"
)

func main() {
	app := demohelper.Setup("Radiobutton Demonstration", 580, 400,
		"Three groups of radiobuttons are displayed below. If you click "+
			"on a button then the button will become selected exclusively "+
			"among all the buttons in its group. A variable is associated "+
			"with each group to indicate which of the group's buttons is "+
			"selected.")

	// Variables.
	sizeVar := widget.NewVariable("12")
	colorVar := widget.NewVariable("red")
	alignVar := widget.NewVariable("left")

	// Container for three groups side by side.
	groupFrame := frame.New(app, "groups")
	pack.Pack(groupFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Point Size group (labelframe).
	sizeFrame := labelframe.New(groupFrame, "sizes",
		labelframe.Text("Point Size"),
	)
	pack.Pack(sizeFrame, pack.SideOpt(pack.Left), pack.PadX(10), pack.PadY(5),
		pack.FillOpt(pack.FillY))

	sizes := []string{"10", "12", "14", "18", "24"}
	for _, s := range sizes {
		rb := radiobutton.New(sizeFrame, "size_"+s,
			radiobutton.Text("Point Size "+s),
			radiobutton.Value(s),
			radiobutton.Var(sizeVar),
		)
		pack.Pack(rb, pack.SideOpt(pack.Top), pack.PadY(2),
			pack.Anchor(option.AnchorW), pack.PadX(10),
			pack.FillOpt(pack.FillX))
	}

	// Color group (labelframe).
	colorFrame := labelframe.New(groupFrame, "colors",
		labelframe.Text("Color"),
	)
	pack.Pack(colorFrame, pack.SideOpt(pack.Left), pack.PadX(10), pack.PadY(5),
		pack.FillOpt(pack.FillY))

	colors := []string{"Red", "Green", "Blue", "Yellow", "Orange", "Purple"}
	for _, c := range colors {
		rb := radiobutton.New(colorFrame, "color_"+strings.ToLower(c),
			radiobutton.Text(c),
			radiobutton.Value(strings.ToLower(c)),
			radiobutton.Var(colorVar),
		)
		pack.Pack(rb, pack.SideOpt(pack.Top), pack.PadY(2),
			pack.Anchor(option.AnchorW), pack.PadX(10),
			pack.FillOpt(pack.FillX))
	}

	// Alignment group (labelframe).
	alignFrame := labelframe.New(groupFrame, "align",
		labelframe.Text("Alignment"),
	)
	pack.Pack(alignFrame, pack.SideOpt(pack.Left), pack.PadX(10), pack.PadY(5),
		pack.FillOpt(pack.FillY))

	alignments := []struct{ text, value string }{
		{"Top", "top"},
		{"Left", "left"},
		{"Right", "right"},
		{"Bottom", "bottom"},
	}
	for _, a := range alignments {
		rb := radiobutton.New(alignFrame, "align_"+a.value,
			radiobutton.Text(a.text),
			radiobutton.Value(a.value),
			radiobutton.Var(alignVar),
		)
		pack.Pack(rb, pack.SideOpt(pack.Top), pack.PadY(2),
			pack.Anchor(option.AnchorW), pack.PadX(10),
			pack.FillOpt(pack.FillX))
	}

	_ = sizeVar
	_ = colorVar
	_ = alignVar
	app.Run()
}
