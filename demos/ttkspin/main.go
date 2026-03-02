// Demo: Spinboxes with integer, float, and string values.
// Ported from Tk's ttkspin.tcl demo (adapted for classic spinbox).
package main

import (
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/spinbox"
)

func main() {
	d := demohelper.Setup("Spinbox Demonstration", 400, 350,
		"Three spinboxes are shown below. The first uses\nan integer range, the second uses float values,\nand the third uses a list of string values.")
	root, app := d.Root, d.App

	// Integer spinbox (0-100).
	intLabel := label.New(root, "intlabel", app,
		label.Text("Integer (0 to 100):"),
		label.Anchor(option.AnchorW),
		label.PadX(20),
	)
	pack.Pack(intLabel.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	intSpin := spinbox.New(root, "intspin", app,
		spinbox.FromOpt(0),
		spinbox.ToOpt(100),
		spinbox.IncrementOpt(1),
	)
	intSpin.SetText("0")
	pack.Pack(intSpin.Window(), pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(5))

	// Float spinbox (0-10 step 0.5).
	floatLabel := label.New(root, "floatlabel", app,
		label.Text("Float (0.0 to 10.0, step 0.5):"),
		label.Anchor(option.AnchorW),
		label.PadX(20),
	)
	pack.Pack(floatLabel.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	floatSpin := spinbox.New(root, "floatspin", app,
		spinbox.FromOpt(0),
		spinbox.ToOpt(10),
		spinbox.IncrementOpt(0.5),
		spinbox.FormatOpt("%.1f"),
	)
	floatSpin.SetText("0.0")
	pack.Pack(floatSpin.Window(), pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(5))

	// Values spinbox (days of week).
	valLabel := label.New(root, "vallabel", app,
		label.Text("Day of week:"),
		label.Anchor(option.AnchorW),
		label.PadX(20),
	)
	pack.Pack(valLabel.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	valSpin := spinbox.New(root, "valspin", app,
		spinbox.ValuesOpt([]string{
			"Sunday", "Monday", "Tuesday", "Wednesday",
			"Thursday", "Friday", "Saturday",
		}),
	)
	valSpin.SetText("Sunday")
	pack.Pack(valSpin.Window(), pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(5))

	_ = intLabel
	_ = intSpin
	_ = floatLabel
	_ = floatSpin
	_ = valLabel
	_ = valSpin
	d.Run()
}
