// Demo: Checkbuttons with tri-state master toggle.
// Ported from Tk's check.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/checkbutton"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Checkbutton Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("check"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength("4i"),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("Four checkbuttons are displayed below. If you click on a "+
			"button, it will toggle the button's selection state and set a "+
			"variable to a value indicating the state of the checkbutton. "+
			"The first button also follows the state of the other three. "+
			"If only some of the three are checked, the first button will "+
			"display the tri-state mode. Click the \"See Variables\" button "+
			"to see the current values of the variables."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	vars := make(demohelper.DemoVars[bool])

	btns := demohelper.AddVarsSeeDismiss(f, &vars)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Variables for the three sub-checkbuttons.
	wipers := widget.NewVariable(false)
	brakes := widget.NewVariable(false)
	sober := widget.NewVariable(false)
	// Variable for the master "Safety Check" button.
	safety := widget.NewVariable(false)

	vars["wipers"] = wipers
	vars["brakes"] = brakes
	vars["sober"] = sober
	vars["safety"] = safety

	// Guard against recursive updates.
	inCheck := false

	// masterCb is set after creation; used to control the indeterminate dash display.
	var masterCb *checkbutton.Checkbutton

	// updateMaster sets the master checkbutton based on the sub-checkbutton states.
	// All three checked → checked; none → unchecked; partial → indeterminate (dash).
	updateMaster := func() {
		if inCheck {
			return
		}
		inCheck = true
		w, b, s := wipers.Get(), brakes.Get(), sober.Get()
		count := 0
		for _, v := range []bool{w, b, s} {
			if v {
				count++
			}
		}
		if count == 3 {
			safety.Set(true)
			if masterCb != nil {
				masterCb.SetIndeterminate(false)
			}
		} else if count == 0 {
			safety.Set(false)
			if masterCb != nil {
				masterCb.SetIndeterminate(false)
			}
		} else {
			// Partial: show indeterminate dash.
			safety.Set(false)
			if masterCb != nil {
				masterCb.SetIndeterminate(true)
			}
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
	cb0 := checkbutton.New(f, "safety",
		checkbutton.Text("Safety Check"),
		checkbutton.Var(safety),
		checkbutton.Command(updateSubs),
	)
	masterCb = cb0
	pack.Pack(cb0, pack.SideOpt(pack.Top), pack.PadY("1.5p"), pack.Anchor(option.AnchorW))

	// Sub-checkbuttons — indented with extra left padding like the Tk original.
	cb1 := checkbutton.New(f, "wipers",
		checkbutton.Text("Wipers OK"),
		checkbutton.Var(wipers),
		checkbutton.Command(updateMaster),
	)
	cb2 := checkbutton.New(f, "brakes",
		checkbutton.Text("Brakes OK"),
		checkbutton.Var(brakes),
		checkbutton.Command(updateMaster),
	)
	cb3 := checkbutton.New(f, "sober",
		checkbutton.Text("Driver Sober"),
		checkbutton.Var(sober),
		checkbutton.Command(updateMaster),
	)
	pack.Pack(geometry.Group{cb1, cb2, cb3}, pack.SideOpt(pack.Top), pack.PadY("1.5p"), pack.Anchor(option.AnchorW), pack.PadX("12p"))

	app.Run()
}
