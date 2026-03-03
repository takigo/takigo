// Demo: Entry widgets with validation constraints and password mode.
// Ported from Tk's entry3.tcl demo — uses labelframes in a 2x2 grid layout.
package main

import (
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/focus"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/widget/entry"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/labelframe"
)

func main() {
	app := demohelper.Setup("Constrained Entry Demonstration", 500, 300,
		"Four different entries are displayed below. You can add characters by pointing, clicking and typing, though each is constrained in what it will accept. The first only accepts integers or the empty string (checking is not enforced in this Go port). The second only accepts strings with fewer than ten characters. The third accepts US phone numbers. The fourth is a password field that accepts up to eight characters, displaying them as asterisks.")
	root := app.Window()

	focusMgr := focus.NewManager(app.Dispatcher(), app.Server())
	focusMgr.BindTraversal(root)

	// Middle frame to hold the 2x2 grid of labelframes.
	mid := frame.New(app, "mid")
	pack.Pack(mid, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(10))

	// Top-left: Integer Entry
	lf1 := labelframe.New(mid, "l1", labelframe.Text("Integer Entry"))
	e1 := entry.New(lf1, "e1", entry.Text("12345"))
	pack.Pack(e1, pack.FillOpt(pack.FillX), pack.Expand(true),
		pack.PadX(4), pack.PadY(4))

	// Top-right: Length-Constrained Entry
	lf2 := labelframe.New(mid, "l2", labelframe.Text("Length-Constrained Entry"))
	e2 := entry.New(lf2, "e2")
	pack.Pack(e2, pack.FillOpt(pack.FillX), pack.Expand(true),
		pack.PadX(4), pack.PadY(4))

	// Bottom-left: US Phone-Number Entry
	lf3 := labelframe.New(mid, "l3", labelframe.Text("US Phone-Number Entry"))
	e3 := entry.New(lf3, "e3", entry.Text("1-(000)-000-0000"))
	pack.Pack(e3, pack.FillOpt(pack.FillX), pack.Expand(true),
		pack.PadX(4), pack.PadY(4))

	// Bottom-right: Password Entry
	lf4 := labelframe.New(mid, "l4", labelframe.Text("Password Entry"))
	e4 := entry.New(lf4, "e4", entry.Show('*'))
	pack.Pack(e4, pack.FillOpt(pack.FillX), pack.Expand(true),
		pack.PadX(4), pack.PadY(4))

	// Arrange labelframes in a 2x2 grid.
	grid.Grid(lf1, grid.Row(0), grid.Column(0), grid.Sticky(grid.EW), grid.PadX(4), grid.PadY(4))
	grid.Grid(lf2, grid.Row(0), grid.Column(1), grid.Sticky(grid.EW), grid.PadX(4), grid.PadY(4))
	grid.Grid(lf3, grid.Row(1), grid.Column(0), grid.Sticky(grid.EW), grid.PadX(4), grid.PadY(4))
	grid.Grid(lf4, grid.Row(1), grid.Column(1), grid.Sticky(grid.EW), grid.PadX(4), grid.PadY(4))

	// Make both columns expand equally.
	grid.ColumnConfigure(mid.Window(), 0, grid.SlotConfig{Weight: 1})
	grid.ColumnConfigure(mid.Window(), 1, grid.SlotConfig{Weight: 1})

	_ = focusMgr
	_ = e1
	_ = e2
	_ = e3
	_ = e4
	app.Run()
}
