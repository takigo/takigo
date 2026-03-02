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
	d := demohelper.Setup("Scale with Label", 400, 250,
		"Drag the scale to update the label value.")
	app := d.App

	// Value display.
	valueLabel := label.New(app, "value",
		label.Text("Value: 50"),
		label.Anchor(option.AnchorCenter),
		label.PadX(10), label.PadY(10),
	)
	pack.Pack(valueLabel, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(10))

	sc := scale.New(app, "ttkscale",
		scale.OrientOpt(scale.Horizontal),
		scale.FromOpt(0),
		scale.ToOpt(100),
		scale.ValueOpt(50),
		scale.ShowValueOpt(true),
		scale.CommandOpt(func(v float64) {
			valueLabel.Text = fmt.Sprintf("Value: %.0f", v)
			valueLabel.Display()
		}),
	)
	pack.Pack(sc, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(30), pack.PadY(10))

	_ = valueLabel
	_ = sc
	d.Run()
}
