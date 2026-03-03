// Demo: Simple form with labeled entries using grid layout.
// Ported from Tk's form.tcl demo.
package main

import (
	"fmt"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/focus"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/entry"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app := demohelper.Setup("Form Demonstration", 450, 280,
		"This window contains a simple form where you can type in the various entries and use tabs to move circularly between the entries.")
	root := app.Window()

	focusMgr := focus.NewManager(app.Dispatcher(), app.DisplayPtr())
	focusMgr.BindTraversal(root)

	// Form grid.
	formFrame := frame.New(app, "form")
	pack.Pack(formFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(20), pack.PadY(10))

	fields := []string{"Name:", "Address:", "City:", "State:", "Phone:"}
	entries := make([]*entry.Entry, len(fields))

	for i, fieldName := range fields {
		l := label.New(formFrame, fmt.Sprintf("l%d", i),
			label.Text(fieldName),
			label.Anchor(option.AnchorE),
		)
		e := entry.New(formFrame, fmt.Sprintf("e%d", i),
			entry.Width(30),
		)
		grid.Grid(l, grid.Row(i), grid.Column(0), grid.Sticky(grid.StickE), grid.PadX(5), grid.PadY(4))
		grid.Grid(e, grid.Row(i), grid.Column(1), grid.Sticky(grid.EW), grid.PadX(5), grid.PadY(4))
		entries[i] = e
		_ = l
	}

	grid.ColumnConfigure(formFrame.Window(), 1, grid.SlotConfig{Weight: 1})

	_ = focusMgr
	_ = entries
	app.Run()
}
