// Demo: Entry widgets with validation constraints and password mode.
// Ported from Tk's entry3.tcl demo — uses labelframes in a 2x2 grid layout.
package main

import (
	"regexp"
	"strconv"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/focus"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/widget/entry"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/labelframe"
)

var phoneRe = regexp.MustCompile(`^1?-?\(?[0-9]{0,3}\)?-?[0-9]{0,3}-?[0-9]{0,4}$`)

func main() {
	app := demohelper.Setup("Constrained Entry Demonstration", 500, 300,
		"Four different entries are displayed below. You can add characters by pointing, clicking and typing, though each is constrained in what it will accept. The first only accepts integers or the empty string. The second only accepts strings with fewer than ten characters. The third accepts US phone numbers. The fourth is a password field that accepts up to eight characters, displaying them as asterisks.")
	root := app.Window()

	focusMgr := focus.NewManager(app.Dispatcher(), app.Server())
	focusMgr.BindTraversal(root)

	// Middle frame to hold the 2x2 grid of labelframes (matches Tcl's $w.mid).
	mid := frame.New(app, "mid")
	pack.Pack(mid, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Top-left: Integer Entry (only accepts integers or empty string).
	lf1 := labelframe.New(mid, "l1", labelframe.Text("Integer Entry"))
	e1 := entry.New(lf1, "e1",
		entry.Text("12345"),
		entry.ValidateOpt("key"),
		entry.ValidateCmdOpt(func(s string) bool {
			if s == "" {
				return true
			}
			_, err := strconv.Atoi(s)
			return err == nil
		}),
	)
	pack.Pack(e1, pack.FillOpt(pack.FillX), pack.Expand(true),
		pack.PadX("1m"), pack.PadY("1m"))

	// Top-right: Length-Constrained Entry (fewer than 10 characters).
	lf2 := labelframe.New(mid, "l2", labelframe.Text("Length-Constrained Entry"))
	e2 := entry.New(lf2, "e2",
		entry.ValidateOpt("key"),
		entry.ValidateCmdOpt(func(s string) bool {
			return len([]rune(s)) < 10
		}),
	)
	pack.Pack(e2, pack.FillOpt(pack.FillX), pack.Expand(true),
		pack.PadX("1m"), pack.PadY("1m"))

	// Bottom-left: US Phone-Number Entry.
	lf3 := labelframe.New(mid, "l3", labelframe.Text("US Phone-Number Entry"))
	e3 := entry.New(lf3, "e3",
		entry.Text("1-(000)-000-0000"),
		entry.ValidateOpt("key"),
		entry.ValidateCmdOpt(func(s string) bool {
			return phoneRe.MatchString(s)
		}),
	)
	pack.Pack(e3, pack.FillOpt(pack.FillX), pack.Expand(true),
		pack.PadX("1m"), pack.PadY("1m"))

	// Bottom-right: Password Entry (up to 8 characters, displayed as asterisks).
	lf4 := labelframe.New(mid, "l4", labelframe.Text("Password Entry"))
	e4 := entry.New(lf4, "e4",
		entry.Show('*'),
		entry.ValidateOpt("key"),
		entry.ValidateCmdOpt(func(s string) bool {
			return len([]rune(s)) <= 8
		}),
	)
	pack.Pack(e4, pack.FillOpt(pack.FillX), pack.Expand(true),
		pack.PadX("1m"), pack.PadY("1m"))

	// Arrange labelframes in a 2x2 grid (matches Tcl: padx 3m pady 1m).
	grid.Grid(lf1, grid.Row(0), grid.Column(0), grid.Sticky(grid.EW), grid.PadX("3m"), grid.PadY("1m"))
	grid.Grid(lf2, grid.Row(0), grid.Column(1), grid.Sticky(grid.EW), grid.PadX("3m"), grid.PadY("1m"))
	grid.Grid(lf3, grid.Row(1), grid.Column(0), grid.Sticky(grid.EW), grid.PadX("3m"), grid.PadY("1m"))
	grid.Grid(lf4, grid.Row(1), grid.Column(1), grid.Sticky(grid.EW), grid.PadX("3m"), grid.PadY("1m"))

	// Make both columns expand equally.
	grid.ColumnConfigure(mid, 0, grid.Weight(1))
	grid.ColumnConfigure(mid, 1, grid.Weight(1))

	_ = focusMgr
	_ = e1
	_ = e2
	_ = e3
	_ = e4
	app.Run()
}
