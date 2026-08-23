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

const (
	safetyAll     = "all"
	safetyNone    = "none"
	safetyPartial = "partial"
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
		label.Text("Four checkbuttons are displayed below.  If you click on a "+
			"button, it will toggle the button's selection state and set a "+
			"Tcl variable to a value indicating the state of the checkbutton.  "+
			"The first button also follows the state of the other three.  "+
			"If only some of the three are checked, the first button will "+
			"display the tri-state mode. Click the \"See Variables\" button to "+
			"see the current values of the variables."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	vars := make(demohelper.DemoVars[string])

	btns := demohelper.AddVarsSeeDismiss(f, &vars)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	wipers := widget.NewVariable("0")
	brakes := widget.NewVariable("0")
	sober := widget.NewVariable("0")
	safety := widget.NewVariable(safetyNone)

	vars["wipers"] = wipers
	vars["brakes"] = brakes
	vars["sober"] = sober
	vars["safety"] = safety

	inCheck := false

	updateMaster := func() {
		if inCheck {
			return
		}
		inCheck = true
		w := wipers.Get() == "1"
		b := brakes.Get() == "1"
		s := sober.Get() == "1"
		count := 0
		for _, v := range []bool{w, b, s} {
			if v {
				count++
			}
		}
		switch count {
		case 3:
			safety.Set(safetyAll)
		case 0:
			safety.Set(safetyNone)
		default:
			safety.Set(safetyPartial)
		}
		inCheck = false
	}

	updateSubs := func() {
		if inCheck {
			return
		}
		inCheck = true
		val := "0"
		if safety.Get() == safetyAll {
			val = "1"
		}
		wipers.Set(val)
		brakes.Set(val)
		sober.Set(val)
		inCheck = false
	}

	cb0 := checkbutton.New(f, "safety",
		checkbutton.Text("Safety Check"),
		checkbutton.Var(safety),
		checkbutton.OnValueOpt(safetyAll),
		checkbutton.OffValueOpt(safetyNone),
		checkbutton.TristateValueOpt(safetyPartial),
		checkbutton.Command(updateSubs),
	)
	pack.Pack(cb0, pack.SideOpt(pack.Top), pack.PadY("1.5p"), pack.Anchor(option.AnchorW))

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
