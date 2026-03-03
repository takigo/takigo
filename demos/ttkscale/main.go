// Demo: Scale with label feedback.
// Ported from Tk's ttkscale.tcl demo.
package main

import (
	"fmt"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/scale"
)

func main() {
	app := demohelper.Setup("Themed Scale Demonstration", 400, 250,
		"A label tied to a horizontal scale is displayed below. If you click or drag mouse button 1 in the scale, you can change the contents of the label; a callback command is used to couple the slider to both the text and the coloring of the label.")

	// Rainbow color list (match Tk's ttkscale.tcl).
	colorList := []struct {
		name string
		hex  string
	}{
		{"Red", "#ff0000"},
		{"Orange", "#ffa500"},
		{"Yellow", "#ffff00"},
		{"Green", "#008000"},
		{"Blue", "#0000ff"},
		{"Violet", "#ee82ee"},
	}

	// Color label display.
	valueLabel := label.New(app, "value",
		label.Text("Color: Red"),
		label.Foreground("#ff0000"),
		label.Anchor(option.AnchorCenter),
		label.PadX(10), label.PadY(10),
	)
	pack.Pack(valueLabel, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(10))

	sc := scale.New(app, "ttkscale",
		scale.OrientOpt(scale.Horizontal),
		scale.FromOpt(0),
		scale.ToOpt(5),
		scale.ValueOpt(0),
		scale.ShowValueOpt(false),
		scale.CommandOpt(func(v float64) {
			idx := int(v)
			if idx < 0 {
				idx = 0
			}
			if idx >= len(colorList) {
				idx = len(colorList) - 1
			}
			c := colorList[idx]
			valueLabel.Text = fmt.Sprintf("Color: %s", c.name)
			col, err := app.ColorCache().Get(c.hex)
			if err == nil {
				valueLabel.Foreground = col
			}
			valueLabel.Display()
		}),
	)
	pack.Pack(sc, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(30), pack.PadY(10))

	_ = valueLabel
	_ = sc
	app.Run()
}
