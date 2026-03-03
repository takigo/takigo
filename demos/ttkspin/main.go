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
	app := demohelper.Setup("Themed Spinbox Demonstration", 400, 350,
		"Three different themed spin-boxes are displayed below. You can add characters by pointing, clicking and typing. Note that the first spin-box will only permit you to type in integers, and the third selects from a list of Australian cities.")

	// Integer spinbox (1-10).
	intLabel := label.New(app, "intlabel",
		label.Text("Integer (1 to 10):"),
		label.Anchor(option.AnchorW),
		label.PadX(20),
	)
	pack.Pack(intLabel, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	intSpin := spinbox.New(app, "intspin",
		spinbox.FromOpt(1),
		spinbox.ToOpt(10),
		spinbox.IncrementOpt(1),
	)
	intSpin.SetText("1")
	pack.Pack(intSpin, pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(5))

	// Float spinbox (0-3 step 0.5).
	floatLabel := label.New(app, "floatlabel",
		label.Text("Float (0.0 to 3.0, step 0.5):"),
		label.Anchor(option.AnchorW),
		label.PadX(20),
	)
	pack.Pack(floatLabel, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	floatSpin := spinbox.New(app, "floatspin",
		spinbox.FromOpt(0),
		spinbox.ToOpt(3),
		spinbox.IncrementOpt(0.5),
		spinbox.FormatOpt("%05.2f"),
	)
	floatSpin.SetText("00.00")
	pack.Pack(floatSpin, pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(5))

	// Values spinbox (Australian cities).
	valLabel := label.New(app, "vallabel",
		label.Text("Australian city:"),
		label.Anchor(option.AnchorW),
		label.PadX(20),
	)
	pack.Pack(valLabel, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	valSpin := spinbox.New(app, "valspin",
		spinbox.ValuesOpt([]string{
			"Canberra", "Sydney", "Melbourne", "Perth",
			"Adelaide", "Brisbane", "Hobart", "Darwin", "Alice Springs",
		}),
	)
	valSpin.SetText("Canberra")
	pack.Pack(valSpin, pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(5))

	_ = intLabel
	_ = intSpin
	_ = floatLabel
	_ = floatSpin
	_ = valLabel
	_ = valSpin
	app.Run()
}
