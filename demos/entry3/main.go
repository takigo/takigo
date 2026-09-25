// Demo: Entry widgets with constrained input and password mode.
// Ported from Tk's entry3.tcl demo.
package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/entry"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/labelframe"
)

var phoneRe = regexp.MustCompile(`^1?-?\(?[0-9]{0,3}\)?-?[0-9]{0,3}-?[0-9]{0,4}$`)

func main() {
	app, err := takigo.NewApp(takigo.Title("Constrained Entry Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("entry3"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength("5i"),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("Four different entries are displayed below.  You can add characters by pointing, clicking and typing, though each is constrained in what it will accept.  The first only accepts 32-bit integers or the empty string (checking when focus leaves it) and will flash to indicate any problem.  The second only accepts strings with fewer than ten characters and sounds the bell when an attempt to go over the limit is made.  The third accepts US phone numbers, mapping letters to their digit equivalent and sounding the bell on encountering an illegal character or if trying to type over a character that is not a digit.  The fourth is a password field that accepts up to eight characters (silently ignoring further ones), and displaying them as asterisk characters."),
	)

	// Bottom buttons (packed before msg/mid to match Tcl pack order).
	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	pack.Pack(msg, pack.SideOpt(pack.Top))

	// Middle frame to hold the 2x2 grid of labelframes.
	mid := frame.New(f, "mid")
	pack.Pack(mid, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Top-left: Integer Entry (only accepts integers or empty string).
	lf1 := labelframe.New(mid, "l1", labelframe.Text("Integer Entry"))
	e1 := entry.New(lf1, "e",
		entry.ValidateOpt("focus"),
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
	e2 := entry.New(lf2, "e",
		entry.ValidateOpt("key"),
		entry.ValidateCmdOpt(func(s string) bool {
			return len([]rune(s)) < 10
		}),
	)
	pack.Pack(e2, pack.FillOpt(pack.FillX), pack.Expand(true),
		pack.PadX("1m"), pack.PadY("1m"))

	// Bottom-left: US Phone-Number Entry.
	lf3 := labelframe.New(mid, "l3", labelframe.Text("US Phone-Number Entry"))
	e3 := entry.New(lf3, "e",
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
	e4 := entry.New(lf4, "e",
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

	// Make both columns equal width.
	grid.ColumnConfigure(mid, 0, grid.Uniform("1"))
	grid.ColumnConfigure(mid, 1, grid.Uniform("1"))

	_ = e1
	_ = e2
	_ = e3
	_ = e4
	app.Run()
}
