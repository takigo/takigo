// Demo: Checkbuttons with tri-state master toggle.
// Ported from Tk's check.tcl demo.
package main

import (
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/checkbutton"
)

func main() {
	app := demohelper.Setup("Checkbutton Demonstration", 400, 350,
		"Four checkbuttons are displayed below. If you click on a "+
			"button, it will toggle the button's selection state and set a "+
			"variable to a value indicating the state of the checkbutton. "+
			"The first button also follows the state of the other three. "+
			"If only some of the three are checked, the first button will "+
			"display the tri-state mode.")

	// Variables for the three sub-checkbuttons.
	wipers := widget.NewVariable(false)
	brakes := widget.NewVariable(false)
	sober := widget.NewVariable(false)

	// Variable for the master "Safety Check" button.
	safety := widget.NewVariable(false)

	// Guard against recursive updates.
	inCheck := false

	// updateMaster sets the master checkbutton based on the sub-checkbutton states.
	// In Tk this would show a tri-state indicator when partially checked;
	// since our checkbutton only supports bool, we show checked only when all
	// three are checked, unchecked otherwise.
	updateMaster := func() {
		if inCheck {
			return
		}
		inCheck = true
		w, b, s := wipers.Get(), brakes.Get(), sober.Get()
		if w && b && s {
			safety.Set(true)
		} else {
			safety.Set(false)
		}
		inCheck = false
	}

	// updateSubs sets all sub-checkbuttons to match the master.
	updateSubs := func() {
		if inCheck {
			return
		}
		inCheck = true
		val := safety.Get()
		wipers.Set(val)
		brakes.Set(val)
		sober.Set(val)
		inCheck = false
	}

	// Master "Safety Check" checkbutton — packed without extra left padding.
	cb0 := checkbutton.New(app, "safety",
		checkbutton.Text("Safety Check"),
		checkbutton.Var(safety),
		checkbutton.Command(updateSubs),
	)
	pack.Pack(cb0, pack.SideOpt(pack.Top), pack.PadY(2), pack.Anchor(option.AnchorW))

	// Sub-checkbuttons — indented with extra left padding like the Tk original.
	cb1 := checkbutton.New(app, "wipers",
		checkbutton.Text("Wipers OK"),
		checkbutton.Var(wipers),
		checkbutton.Command(updateMaster),
	)
	pack.Pack(cb1, pack.SideOpt(pack.Top), pack.PadY(2), pack.Anchor(option.AnchorW), pack.PadX(20))

	cb2 := checkbutton.New(app, "brakes",
		checkbutton.Text("Brakes OK"),
		checkbutton.Var(brakes),
		checkbutton.Command(updateMaster),
	)
	pack.Pack(cb2, pack.SideOpt(pack.Top), pack.PadY(2), pack.Anchor(option.AnchorW), pack.PadX(20))

	cb3 := checkbutton.New(app, "sober",
		checkbutton.Text("Driver Sober"),
		checkbutton.Var(sober),
		checkbutton.Command(updateMaster),
	)
	pack.Pack(cb3, pack.SideOpt(pack.Top), pack.PadY(2), pack.Anchor(option.AnchorW), pack.PadX(20))

	_ = cb0
	_ = cb1
	_ = cb2
	_ = cb3
	app.Run()
}
