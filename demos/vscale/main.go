// Demo: Vertical scale controlling a canvas arrow size.
// Ported from Tk's vscale.tcl demo.
package main

import (
	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/scale"
)

// setHeight mirrors the Tk setHeight proc.
func setHeight(c *canvas.Canvas, value float64) {
	height := value + 21
	y2 := height - 30
	if y2 < 21 {
		y2 = 21
	}
	// Polygon: shaft rectangle + triangular arrowhead pointing down, closed.
	coords := []float64{
		15, 20, 35, 20, 35, y2, 45, y2, 25, height, 5, y2, 15, y2, 15, 20,
	}
	c.SetItemCoords("poly", coords)
	c.SetItemCoords("line", coords)
}

func main() {
	app := demohelper.Setup("Vertical Scale Demonstration", 400, 400,
		"An arrow and a vertical scale are displayed below. If you click or drag mouse button 1 in the scale, you can change the size of the arrow.")

	// Middle frame with border (matches Tcl's `frame -borderwidth 7.5p`).
	fr := frame.New(app, "frame",
		frame.BorderWidth(10), // 7.5p ≈ 10px
	)
	pack.Pack(fr, pack.SideOpt(pack.Top))

	// Canvas for arrow display.
	c := canvas.New(fr, "canvas",
		canvas.Width(50),
		canvas.Height(300),
		canvas.BorderWidthOpt(0),
		canvas.HighlightWidthOpt(0),
	)

	// Create initial polygon and line items with dummy coords.
	c.CreatePolygon([]float64{0, 0, 1, 1, 2, 2},
		canvas.FillColor("SeaGreen3"), canvas.Tags("poly"))
	c.CreateLine([]float64{0, 0, 1, 1, 2, 2, 0, 0},
		canvas.OutlineColor("black"), canvas.Tags("line"))

	// Vertical scale: 0-250, matching Tcl's -tickinterval 50.
	sc := scale.New(fr, "vscale",
		scale.OrientOpt(scale.Vertical),
		scale.FromOpt(0),
		scale.ToOpt(250),
		scale.ValueOpt(75),
		scale.TickIntervalOpt(50),
		scale.CommandOpt(func(v float64) {
			setHeight(c, v)
		}),
	)

	pack.Pack(sc, pack.SideOpt(pack.Left), pack.Anchor(option.AnchorNE))
	pack.Pack(c, pack.SideOpt(pack.Left), pack.Anchor(option.AnchorNW), pack.FillOpt(pack.FillY))

	// Set initial arrow.
	setHeight(c, 75)

	app.Run()
}
