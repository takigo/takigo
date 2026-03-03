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
	app := demohelper.Setup("Spinbox Demonstration", 400, 300,
		"Three different spin-boxes are displayed below. You can add characters by pointing, clicking and typing. Note that the first spin-box will only permit you to type in integers, and the third selects from a list of Australian cities.")
	root := app.Window()

	focusMgr := focus.NewManager(app.Dispatcher(), app.DisplayPtr())
	focusMgr.BindTraversal(root)

	// Status label.
	statusLabel := label.New(app, "status",
		label.Text("Status: Ready"),
		label.Anchor(option.AnchorW),
		label.PadX(10),
	)
	pack.Pack(statusLabel, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	setStatus := func(msg string) {
		statusLabel.Text = msg
		statusLabel.Display()
	}

	// Spinbox 1: Integer range 1-10.
	spin1Label := label.New(app, "spin1label",
		label.Text("Integer (1-10):"),
		label.Anchor(option.AnchorW),
		label.PadX(20),
	)
	pack.Pack(spin1Label, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(2))

	spin1 := spinbox.New(app, "spin1",
		spinbox.FromOpt(1),
		spinbox.ToOpt(10),
		spinbox.IncrementOpt(1),
		spinbox.WrapOpt(true),
		spinbox.CommandOpt(func(v string) {
			setStatus(fmt.Sprintf("Integer: %s", v))
		}),
	)
	pack.Pack(spin1, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(20), pack.PadY(5))

	// Spinbox 2: Float range 0-3, increment 0.5.
	spin2Label := label.New(app, "spin2label",
		label.Text("Float (0.0-3.0, step 0.5):"),
		label.Anchor(option.AnchorW),
		label.PadX(20),
	)
	pack.Pack(spin2Label, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(2))

	spin2 := spinbox.New(app, "spin2",
		spinbox.FromOpt(0),
		spinbox.ToOpt(3),
		spinbox.IncrementOpt(0.5),
		spinbox.WrapOpt(true),
		spinbox.CommandOpt(func(v string) {
			setStatus(fmt.Sprintf("Float: %s", v))
		}),
	)
	pack.Pack(spin2, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(20), pack.PadY(5))

	// Spinbox 3: Australian cities.
	spin3Label := label.New(app, "spin3label",
		label.Text("City:"),
		label.Anchor(option.AnchorW),
		label.PadX(20),
	)
	pack.Pack(spin3Label, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(2))

	cities := []string{
		"Canberra", "Sydney", "Melbourne", "Perth",
		"Adelaide", "Brisbane", "Hobart", "Darwin", "Alice Springs",
	}
	spin3 := spinbox.New(app, "spin3",
		spinbox.ValuesOpt(cities),
		spinbox.WrapOpt(true),
		spinbox.CommandOpt(func(v string) {
			setStatus(fmt.Sprintf("City: %s", v))
		}),
	)
	pack.Pack(spin3, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(20), pack.PadY(5))

	_ = focusMgr
	_ = spin1Label
	_ = spin2Label
	_ = spin3Label
	_ = statusLabel
	_ = spin1
	_ = spin2
	_ = spin3
	app.Run()
}
