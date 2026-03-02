// Demo: Vertical scale controlling a canvas bar height.
// Ported from Tk's vscale.tcl demo.
package main

import (
	"fmt"

	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/scale"
)

func main() {
	app := demohelper.Setup("Vertical Scale", 400, 400,
		"Drag the vertical scale to change the bar height.")

	// Middle frame: scale on left, canvas on right.
	midFrame := frame.New(app, "midframe")
	pack.Pack(midFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Canvas.
	c := canvas.New(midFrame, "canvas",
		canvas.Background("white"),
		canvas.Width(250),
		canvas.Height(250),
	)

	barBottom := 240.0

	drawBar := func(v float64) {
		c.Delete("bar")
		c.Delete("label")
		h := v / 100 * 200
		c.CreateRectangle(80, barBottom-h, 180, barBottom,
			canvas.FillColor("#3498db"), canvas.OutlineColor("#2980b9"), canvas.OutlineWidth(2),
			canvas.Tags("bar"))
		c.CreateText(130, barBottom-h-10,
			canvas.TextOpt(fmt.Sprintf("%.0f%%", v)),
			canvas.FontOpt("Sans Bold 11"), canvas.AnchorOpt(option.AnchorCenter),
			canvas.Tags("label"))
	}
	drawBar(50)

	// Vertical scale.
	sc := scale.New(midFrame, "vscale",
		scale.OrientOpt(scale.Vertical),
		scale.FromOpt(100),
		scale.ToOpt(0),
		scale.ValueOpt(50),
		scale.ShowValueOpt(true),
		scale.CommandOpt(func(v float64) { drawBar(v) }),
	)

	pack.Pack(sc, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillY), pack.PadX(10))
	pack.Pack(c, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	_ = sc
	app.Run()
}
