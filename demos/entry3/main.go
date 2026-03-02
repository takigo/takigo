// Demo: Entry widgets with password mode and various constraints.
// Ported from Tk's entry3.tcl demo (simplified — no validation callbacks).
package main

import (
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
	app := demohelper.Setup("Validated Entries", 500, 300,
		"Four different entries are shown, each with\ndifferent constraints. Use Tab to move between them.")
	root := app.Window()

	focusMgr := focus.NewManager(app.Dispatcher(), app.DisplayPtr())
	focusMgr.BindTraversal(root)

	// Form grid.
	formFrame := frame.New(app, "form")
	pack.Pack(formFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(20), pack.PadY(10))

	// Row 0: Integer entry.
	l1 := label.New(formFrame, "l1",
		label.Text("Integer:"),
		label.Anchor(option.AnchorE),
	)
	e1 := entry.New(formFrame, "e1",
		entry.Text("12345"),
		entry.Width(20),
	)
	grid.Grid(l1, grid.Row(0), grid.Column(0), grid.Sticky(grid.StickE), grid.PadX(5), grid.PadY(5))
	grid.Grid(e1, grid.Row(0), grid.Column(1), grid.Sticky(grid.EW), grid.PadX(5), grid.PadY(5))

	// Row 1: Short entry (max ~10 chars concept).
	l2 := label.New(formFrame, "l2",
		label.Text("Short text:"),
		label.Anchor(option.AnchorE),
	)
	e2 := entry.New(formFrame, "e2",
		entry.Width(10),
		entry.Placeholder("Max 10 chars"),
	)
	grid.Grid(l2, grid.Row(1), grid.Column(0), grid.Sticky(grid.StickE), grid.PadX(5), grid.PadY(5))
	grid.Grid(e2, grid.Row(1), grid.Column(1), grid.Sticky(grid.EW), grid.PadX(5), grid.PadY(5))

	// Row 2: Phone number.
	l3 := label.New(formFrame, "l3",
		label.Text("Phone:"),
		label.Anchor(option.AnchorE),
	)
	e3 := entry.New(formFrame, "e3",
		entry.Text("1-(555)-123-4567"),
		entry.Width(20),
	)
	grid.Grid(l3, grid.Row(2), grid.Column(0), grid.Sticky(grid.StickE), grid.PadX(5), grid.PadY(5))
	grid.Grid(e3, grid.Row(2), grid.Column(1), grid.Sticky(grid.EW), grid.PadX(5), grid.PadY(5))

	// Row 3: Password entry.
	l4 := label.New(formFrame, "l4",
		label.Text("Password:"),
		label.Anchor(option.AnchorE),
	)
	e4 := entry.New(formFrame, "e4",
		entry.Show('*'),
		entry.Width(20),
		entry.Placeholder("Enter password"),
	)
	grid.Grid(l4, grid.Row(3), grid.Column(0), grid.Sticky(grid.StickE), grid.PadX(5), grid.PadY(5))
	grid.Grid(e4, grid.Row(3), grid.Column(1), grid.Sticky(grid.EW), grid.PadX(5), grid.PadY(5))

	grid.ColumnConfigure(formFrame.Window(), 1, grid.SlotConfig{Weight: 1})

	_ = focusMgr
	_ = l1
	_ = l2
	_ = l3
	_ = l4
	_ = e1
	_ = e2
	_ = e3
	_ = e4
	app.Run()
}
