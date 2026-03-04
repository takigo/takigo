// Demo: Scale with label feedback.
// Ported from Tk's ttkscale.tcl demo.
package main

import (
	"fmt"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/scale"
)

func main() {
	app := demohelper.Setup("Themed Scale Demonstration", 400, 250,
		"A label tied to a horizontal scale is displayed below. If you click or drag mouse button 1 in the scale, you can change the contents of the label; a callback command is used to couple the slider to both the text and the coloring of the label.")

	// Rainbow color list — use X11 named colors matching Tcl.
	colorList := []string{"Red", "Orange", "Yellow", "Green", "Blue", "Violet"}

	// Inner frame with border (matches Tcl's `ttk::frame $w.frame -borderwidth 7.5p`).
	fr := frame.New(app, "frame",
		frame.BorderWidth(10), // 7.5p ≈ 10px
	)
	pack.Pack(fr, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// Color label display — packed first (label before scale, matching Tcl).
	valueLabel := label.New(fr, "label",
		label.Text("Color: Red"),
		label.Foreground("Red"),
		label.Anchor(option.AnchorW),
	)
	pack.Pack(valueLabel, pack.SideOpt(pack.Top))

	sc := scale.New(fr, "scale",
		scale.OrientOpt(scale.Horizontal),
		scale.FromOpt(0),
		scale.ToOpt(5),
		scale.ValueOpt(0),
		scale.CommandOpt(func(v float64) {
			idx := int(v)
			if idx < 0 {
				idx = 0
			}
			if idx >= len(colorList) {
				idx = len(colorList) - 1
			}
			c := colorList[idx]
			valueLabel.Text = fmt.Sprintf("Color: %s", c)
			col, err := app.ColorCache().Get(c)
			if err == nil {
				valueLabel.Foreground = col
			}
			valueLabel.Display()
		}),
	)
	pack.Pack(sc, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	_ = valueLabel
	_ = sc
	app.Run()
}
