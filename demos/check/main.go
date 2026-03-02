// Demo: Checkbuttons with on/off state and status label.
// Ported from Tk's check.tcl demo.
package main

import (
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/checkbutton"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app := demohelper.Setup("Checkbutton Demonstration", 400, 350, "Three checkbuttons are displayed below. Click on a\nbutton to toggle its state. The current state of each\nbutton is displayed at the bottom.")

	// Status label.
	statusLabel := label.New(app, "status",
		label.Text("Status: all unchecked"),
		label.Anchor(option.AnchorW),
		label.PadX(10),
	)
	pack.Pack(statusLabel, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))

	// Variables.
	wipers := widget.NewVariable(false)
	brakes := widget.NewVariable(false)
	sober := widget.NewVariable(false)

	updateStatus := func() {
		status := "Status:"
		if wipers.Get() {
			status += " Wipers-on"
		}
		if brakes.Get() {
			status += " Brakes-on"
		}
		if sober.Get() {
			status += " Sober"
		}
		if !wipers.Get() && !brakes.Get() && !sober.Get() {
			status += " all unchecked"
		}
		statusLabel.Text = status
		statusLabel.Display()
	}

	// Checkbuttons.
	cb1 := checkbutton.New(app, "wipers",
		checkbutton.Text("Safety Check: Wipers OK"),
		checkbutton.Var(wipers),
		checkbutton.Command(updateStatus),
	)
	pack.Pack(cb1, pack.SideOpt(pack.Top), pack.PadY(4), pack.Anchor(option.AnchorW), pack.PadX(20))

	cb2 := checkbutton.New(app, "brakes",
		checkbutton.Text("Safety Check: Brakes OK"),
		checkbutton.Var(brakes),
		checkbutton.Command(updateStatus),
	)
	pack.Pack(cb2, pack.SideOpt(pack.Top), pack.PadY(4), pack.Anchor(option.AnchorW), pack.PadX(20))

	cb3 := checkbutton.New(app, "sober",
		checkbutton.Text("Safety Check: Driver Sober"),
		checkbutton.Var(sober),
		checkbutton.Command(updateStatus),
	)
	pack.Pack(cb3, pack.SideOpt(pack.Top), pack.PadY(4), pack.Anchor(option.AnchorW), pack.PadX(20))

	_ = statusLabel
	_ = cb1
	_ = cb2
	_ = cb3
	app.Run()
}
