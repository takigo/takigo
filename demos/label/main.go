// Demo: Labels with various styles and options.
// Ported from Tk's label.tcl demo.
package main

import (
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	d := demohelper.Setup("Label Demonstration", 450, 350,
		"Five labels are displayed below. They are all the same except\nfor their relief and border width settings.")
	app := d.App

	// Labels with different relief styles.
	reliefs := []struct {
		name   string
		text   string
		relief option.Relief
	}{
		{"flat", "Flat (default)", option.ReliefFlat},
		{"raised", "Raised", option.ReliefRaised},
		{"sunken", "Sunken", option.ReliefSunken},
		{"groove", "Groove", option.ReliefGroove},
		{"ridge", "Ridge", option.ReliefRidge},
	}

	for _, r := range reliefs {
		l := label.New(app, r.name,
			label.Text(r.text),
			label.BorderWidth(3),
			label.Relief(r.relief),
			label.Width(20),
			label.PadX(10),
			label.PadY(5),
		)
		pack.Pack(l, pack.SideOpt(pack.Top), pack.PadY(5), pack.PadX(10))
		_ = l
	}

	d.Run()
}
