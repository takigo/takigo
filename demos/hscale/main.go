// Demo: Horizontal scale controlling a canvas arrow.
// Ported from Tk's hscale.tcl demo.
package main

import (
	"fmt"
	"math"

	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/scale"
)

func main() {
	app := demohelper.Setup("Horizontal Scale", 500, 350,
		"Drag the scale to change the arrow angle on the canvas.")

	// Canvas for arrow display.
	c := canvas.New(app, "canvas",
		canvas.Background("white"),
		canvas.Width(300),
		canvas.Height(200),
	)
	pack.Pack(c, pack.SideOpt(pack.Top), pack.PadX(10), pack.PadY(5))

	// Draw initial arrow.
	cx, cy := 150.0, 100.0
	arrowLen := 80.0

	drawArrow := func(angleDeg float64) {
		c.Delete("arrow")
		rad := angleDeg * math.Pi / 180
		ex := cx + arrowLen*math.Cos(rad)
		ey := cy - arrowLen*math.Sin(rad)
		c.CreateLine([]float64{cx, cy, ex, ey},
			canvas.OutlineColor("#e74c3c"), canvas.OutlineWidth(3),
			canvas.Arrow(canvas.ArrowLast), canvas.ArrowShape(12, 16, 5),
			canvas.Tags("arrow"))
		c.CreateText(cx, cy+arrowLen+20,
			canvas.TextOpt(fmt.Sprintf("%.0f°", angleDeg)),
			canvas.FontOpt("Sans 10"), canvas.AnchorOpt(option.AnchorCenter),
			canvas.Tags("arrow"))
	}
	drawArrow(45)

	// Horizontal scale.
	sc := scale.New(app, "hscale",
		scale.OrientOpt(scale.Horizontal),
		scale.FromOpt(0),
		scale.ToOpt(360),
		scale.ValueOpt(45),
		scale.ShowValueOpt(true),
		scale.CommandOpt(func(v float64) {
			drawArrow(v)
		}),
	)
	pack.Pack(sc, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(30), pack.PadY(10))

	_ = sc
	app.Run()
}
