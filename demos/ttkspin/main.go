// Demo: Spinboxes with integer, float, and string values.
// Ported from Tk's ttkspin.tcl demo.
package main

import (
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/widget/spinbox"
)

func main() {
	app := demohelper.Setup("Themed Spinbox Demonstration", 400, 350,
		"Three different themed spin-boxes are displayed below. You can add characters by pointing, clicking and typing. The normal Motif editing characters are supported, along with many Emacs bindings. For example, Backspace and Control-h delete the character to the left of the insertion cursor and Delete and Control-d delete the chararacter to the right of the insertion cursor. For values that are too large to fit in the window all at once, you can scan through the value by dragging with mouse button2 pressed. Note that the first spin-box will only permit you to type in integers, and the third selects from a list of Australian cities.")

	padX := screenunit.Px("7.5p")
	padY := screenunit.Px("3p")

	// Integer spinbox (1-10).
	s1 := spinbox.New(app, "s1",
		spinbox.FromOpt(1),
		spinbox.ToOpt(10),
		spinbox.IncrementOpt(1),
		spinbox.WidthOpt(10),
	)
	s1.SetText("1")
	pack.Pack(s1, pack.SideOpt(pack.Top), pack.PadX(padX), pack.PadY(padY))

	// Float spinbox (0-3 step 0.5).
	s2 := spinbox.New(app, "s2",
		spinbox.FromOpt(0),
		spinbox.ToOpt(3),
		spinbox.IncrementOpt(0.5),
		spinbox.FormatOpt("%05.2f"),
		spinbox.WidthOpt(10),
	)
	s2.SetText("00.00")
	pack.Pack(s2, pack.SideOpt(pack.Top), pack.PadX(padX), pack.PadY(padY))

	// Values spinbox (Australian cities).
	s3 := spinbox.New(app, "s3",
		spinbox.ValuesOpt([]string{
			"Canberra", "Sydney", "Melbourne", "Perth",
			"Adelaide", "Brisbane", "Hobart", "Darwin", "Alice Springs",
		}),
		spinbox.WidthOpt(10),
	)
	s3.SetText("Canberra")
	pack.Pack(s3, pack.SideOpt(pack.Top), pack.PadX(padX), pack.PadY(padY))

	_ = s1
	_ = s2
	_ = s3
	app.Run()
}
