// Demo: Menus and cascaded menus using menubuttons.
// Ported from Tk's menubu.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/menu"
	"github.com/msorc/takigo/widget/menubutton"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Menu Button Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("menubutton"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Body frame — expands to fill.
	body := frame.New(f, "body")
	pack.Pack(body, pack.Expand(true), pack.FillOpt(pack.FillBoth))

	// Bottom buttons.
	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Helper to create a menubutton with a 2-item menu.
	makeMB := func(parent widget.Caregiver, name, text string, dir menubutton.Direction) *menubutton.Menubutton {
		mb := menubutton.New(parent, name,
			menubutton.Text(text),
			menubutton.UnderlineOpt(0),
			menubutton.DirectionOpt(dir),
		)
		m := menu.New(mb, "m")
		m.AddCommand(text+" menu: first item", func() {
			fmt.Printf("You have selected the first item from the %s menu.\n", text)
		})
		m.AddCommand(text+" menu: second item", func() {
			fmt.Printf("You have selected the second item from the %s menu.\n", text)
		})
		mb.Menu = m
		return mb
	}

	// Compass grid layout:
	//   row 0, col 1: below (sticky n)
	//   row 1, col 0: right (sticky w)
	//   row 1, col 1: center frame
	//   row 1, col 2: left  (sticky e)
	//   row 2, col 1: above (sticky s)

	mbBelow := makeMB(body, "below", "Below", menubutton.Below)
	grid.Grid(mbBelow, grid.Row(0), grid.Column(1), grid.Sticky(grid.StickN))

	mbRight := makeMB(body, "right", "Right", menubutton.Right)
	center := frame.New(body, "center")
	mbLeft := makeMB(body, "left", "Left", menubutton.Left)
	grid.Grid(mbRight, grid.Row(1), grid.Column(0), grid.Sticky(grid.StickW))
	grid.Grid(center, grid.Row(1), grid.Column(1), grid.Sticky(grid.NSEW))
	grid.Grid(mbLeft, grid.Row(1), grid.Column(2), grid.Sticky(grid.StickE))

	mbAbove := makeMB(body, "above", "Above", menubutton.Above)
	grid.Grid(mbAbove, grid.Row(2), grid.Column(1), grid.Sticky(grid.StickS))

	// Center label.
	centerLabel := label.New(center, "label",
		label.WrapLength("225p"),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("This is a demonstration of menubuttons. The \"Below\" menubutton pops its menu below the button; the \"Right\" button pops to the right, etc."),
	)
	pack.Pack(centerLabel, pack.SideOpt(pack.Top), pack.PadX("18p"), pack.PadY("18p"))

	grid.ColumnConfigure(body, 1, grid.Weight(1))
	grid.RowConfigure(body, 1, grid.Weight(1))

	_ = mbBelow
	_ = mbLeft
	_ = mbRight
	_ = mbAbove
	_ = center
	_ = centerLabel
	app.Run()
}
