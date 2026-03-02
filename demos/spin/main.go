// Demo: Spinbox widgets with different value modes.
// Ported from Tk's spin.tcl demo.
package main

import (
	"fmt"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/focus"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/spinbox"
)

func main() {
	d := demohelper.Setup("Spinbox Demonstration", 400, 300,
		"Three spinboxes are shown: an integer range,\na float range, and a list of city names.")
	root, app := d.Root, d.App

	focusMgr := focus.NewManager(app.Dispatcher(), app.DisplayPtr())
	focusMgr.BindTraversal(root)

	// Status label.
	statusLabel := label.New(root, "status", app,
		label.Text("Status: Ready"),
		label.Anchor(option.AnchorW),
		label.PadX(10),
	)
	pack.Pack(statusLabel.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	setStatus := func(msg string) {
		statusLabel.Text = msg
		statusLabel.Display()
	}

	// Spinbox 1: Integer range 1-10.
	spin1Label := label.New(root, "spin1label", app,
		label.Text("Integer (1-10):"),
		label.Anchor(option.AnchorW),
		label.PadX(20),
	)
	pack.Pack(spin1Label.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(2))

	spin1 := spinbox.New(root, "spin1", app,
		spinbox.FromOpt(1),
		spinbox.ToOpt(10),
		spinbox.IncrementOpt(1),
		spinbox.WrapOpt(true),
		spinbox.CommandOpt(func(v string) {
			setStatus(fmt.Sprintf("Integer: %s", v))
		}),
	)
	pack.Pack(spin1.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(20), pack.PadY(5))

	// Spinbox 2: Float range 0-3, increment 0.5.
	spin2Label := label.New(root, "spin2label", app,
		label.Text("Float (0.0-3.0, step 0.5):"),
		label.Anchor(option.AnchorW),
		label.PadX(20),
	)
	pack.Pack(spin2Label.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(2))

	spin2 := spinbox.New(root, "spin2", app,
		spinbox.FromOpt(0),
		spinbox.ToOpt(3),
		spinbox.IncrementOpt(0.5),
		spinbox.WrapOpt(true),
		spinbox.CommandOpt(func(v string) {
			setStatus(fmt.Sprintf("Float: %s", v))
		}),
	)
	pack.Pack(spin2.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(20), pack.PadY(5))

	// Spinbox 3: Australian cities.
	spin3Label := label.New(root, "spin3label", app,
		label.Text("City:"),
		label.Anchor(option.AnchorW),
		label.PadX(20),
	)
	pack.Pack(spin3Label.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(2))

	cities := []string{
		"Canberra", "Sydney", "Melbourne", "Perth",
		"Adelaide", "Brisbane", "Hobart", "Darwin", "Alice Springs",
	}
	spin3 := spinbox.New(root, "spin3", app,
		spinbox.ValuesOpt(cities),
		spinbox.WrapOpt(true),
		spinbox.CommandOpt(func(v string) {
			setStatus(fmt.Sprintf("City: %s", v))
		}),
	)
	pack.Pack(spin3.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(20), pack.PadY(5))

	_ = focusMgr
	_ = spin1Label
	_ = spin2Label
	_ = spin3Label
	_ = statusLabel
	_ = spin1
	_ = spin2
	_ = spin3
	d.Run()
}
