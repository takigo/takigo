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
	app := d.App

	// Integer spinbox (0-100).
	intLabel := label.New(app, "intlabel",
		label.Text("Integer (0 to 100):"),
		label.Anchor(option.AnchorW),
		label.PadX(20),
	)
	pack.Pack(intLabel, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	intSpin := spinbox.New(app, "intspin",
		spinbox.FromOpt(0),
		spinbox.ToOpt(100),
		spinbox.IncrementOpt(1),
	)
	intSpin.SetText("0")
	pack.Pack(intSpin, pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(5))

	// Float spinbox (0-10 step 0.5).
	floatLabel := label.New(app, "floatlabel",
		label.Text("Float (0.0 to 10.0, step 0.5):"),
		label.Anchor(option.AnchorW),
		label.PadX(20),
	)
	pack.Pack(floatLabel, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	floatSpin := spinbox.New(app, "floatspin",
		spinbox.FromOpt(0),
		spinbox.ToOpt(10),
		spinbox.IncrementOpt(0.5),
		spinbox.FormatOpt("%.1f"),
	)
	floatSpin.SetText("0.0")
	pack.Pack(floatSpin, pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(5))

	// Values spinbox (days of week).
	valLabel := label.New(app, "vallabel",
		label.Text("Day of week:"),
		label.Anchor(option.AnchorW),
		label.PadX(20),
	)
	pack.Pack(valLabel, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	valSpin := spinbox.New(app, "valspin",
		spinbox.ValuesOpt([]string{
			"Sunday", "Monday", "Tuesday", "Wednesday",
			"Thursday", "Friday", "Saturday",
		}),
	)
	valSpin.SetText("Sunday")
	pack.Pack(valSpin, pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(5))

	_ = intLabel
	_ = intSpin
	_ = floatLabel
	_ = floatSpin
	_ = valLabel
	_ = valSpin
	d.Run()
}
