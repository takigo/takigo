// Demo: Horizontal scale controlling a canvas arrow length.
// Ported from Tk's hscale.tcl demo.
package main

import (
	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/scale"
)

func main() {
	app := demohelper.Setup("Horizontal Scale Demonstration", 500, 350,
		"An arrow and a horizontal scale are displayed below. If you click or drag mouse button 1 in the scale, you can change the length of the arrow.")

	// Inner frame with border (matches Tcl's `frame -borderwidth 7.5p`).
	fr := frame.New(app, "frame",
		frame.BorderWidth(10), // 7.5p ≈ 10px
	)
	pack.Pack(fr, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// Canvas for arrow display.
	c := canvas.New(fr, "canvas",
		canvas.Width(300),
		canvas.Height(50),
		canvas.BorderWidthOpt(0),
		canvas.HighlightWidthOpt(0),
	)

	// Create initial polygon and line items with dummy coords.
	c.CreatePolygon([]float64{0, 0, 1, 1, 2, 2},
		canvas.FillColor("DeepSkyBlue3"), canvas.Tags("poly"))
	c.CreateLine([]float64{0, 0, 1, 1, 2, 2, 0, 0},
		canvas.OutlineColor("black"), canvas.Tags("line"))

	// setWidth mirrors the Tk setWidth proc.
	setWidth := func(value float64) {
		width := value + 21
		x2 := width - 30
		if x2 < 21 {
			x2 = 21
		}
		// Polygon: shaft rectangle + triangular arrowhead, closed.
		coords := []float64{
			20, 15, 20, 35, x2, 35, x2, 45, width, 25, x2, 5, x2, 15, 20, 15,
		}
		c.SetItemCoords("poly", coords)
		c.SetItemCoords("line", coords)
	}

	// Horizontal scale: 0-250, matching Tcl's -tickinterval 50.
	sc := scale.New(fr, "hscale",
		scale.OrientOpt(scale.Horizontal),
		scale.FromOpt(0),
		scale.ToOpt(250),
		scale.ValueOpt(75),
		scale.TickIntervalOpt(50),
		scale.CommandOpt(func(v float64) {
			setWidth(v)
		}),
	)

	pack.Pack(c, pack.SideOpt(pack.Top), pack.Expand(true), pack.Anchor(option.AnchorS), pack.FillOpt(pack.FillX), pack.PadX("12p"))
	pack.Pack(sc, pack.SideOpt(pack.Bottom), pack.Expand(true), pack.Anchor(option.AnchorN))

	// Set initial arrow.
	setWidth(75)

	app.Run()
}
