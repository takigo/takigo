// Demo: Editable arrowheads on canvas lines.
// Ported from Tk's arrow.tcl demo.
package main

import (
	"fmt"

	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
)

func main() {
	d := demohelper.Setup("Arrow Shapes", 550, 400,
		"Various arrow shapes on canvas lines.\nArrows can appear at first, last, or both ends.")
	defer d.App.Destroy()
	root, app := d.Root, d.App

	// Canvas.
	c := canvas.New(root, "arrows", app,
		canvas.Background("white"),
		canvas.Width(500),
		canvas.Height(300),
	)
	pack.Pack(c.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Row 1: Arrow at last end with different shapes.
	c.CreateText(250, 15, canvas.TextOpt("Arrow at Last End"),
		canvas.FontOpt("Sans Bold 11"), canvas.AnchorOpt(option.AnchorCenter))

	y := 40.0
	shapes := [][3]float64{
		{8, 10, 3},
		{12, 16, 5},
		{16, 20, 7},
		{20, 28, 10},
	}
	for _, s := range shapes {
		c.CreateLine([]float64{50, y, 250, y},
			canvas.OutlineColor("black"), canvas.OutlineWidth(2),
			canvas.Arrow(canvas.ArrowLast),
			canvas.ArrowShape(s[0], s[1], s[2]))
		c.CreateText(280, y,
			canvas.TextOpt(fmt.Sprintf("%.0f, %.0f, %.0f", s[0], s[1], s[2])),
			canvas.FontOpt("Sans 9"), canvas.AnchorOpt(option.AnchorW))
		y += 30
	}

	// Row 2: Arrow at both ends.
	c.CreateText(250, 175, canvas.TextOpt("Arrows at Both Ends"),
		canvas.FontOpt("Sans Bold 11"), canvas.AnchorOpt(option.AnchorCenter))

	c.CreateLine([]float64{50, 200, 450, 200},
		canvas.OutlineColor("#e74c3c"), canvas.OutlineWidth(3),
		canvas.Arrow(canvas.ArrowBoth), canvas.ArrowShape(12, 16, 5))

	c.CreateLine([]float64{50, 240, 450, 240},
		canvas.OutlineColor("#3498db"), canvas.OutlineWidth(4),
		canvas.Arrow(canvas.ArrowBoth), canvas.ArrowShape(16, 24, 8))

	// Polyline with arrows.
	c.CreateLine([]float64{50, 280, 150, 260, 250, 290, 350, 260, 450, 280},
		canvas.OutlineColor("#27ae60"), canvas.OutlineWidth(2),
		canvas.Arrow(canvas.ArrowBoth), canvas.ArrowShape(10, 14, 5),
		canvas.Smooth(true))

	d.Run()
}
