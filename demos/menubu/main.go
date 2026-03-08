// Demo: Menu buttons with different directions.
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

	msg := label.New(f, "msg",
		label.WrapLength("4i"),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("This is a demonstration of menubuttons. The \"Below\" menubutton pops its menu below the button; the \"Right\" button pops to the right, etc."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Body frame — expands to fill.
	body := frame.New(f, "body")
	pack.Pack(body, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Helper to create a menubutton with a 2-item menu.
	makeMB := func(parent widget.Caregiver, name string, dir menubutton.Direction) *menubutton.Menubutton {
		m := menu.New(app, "menu_"+name)
		m.AddCommand(name+" menu: first item", func() {})
		m.AddCommand(name+" menu: second item", func() {})

		return menubutton.New(parent, "mb_"+name,
			menubutton.Text(name),
			menubutton.MenuOpt(m),
			menubutton.DirectionOpt(dir),
		)
	}

	// Compass grid layout matching Tcl's menubu.tcl:
	//   row 0, col 1: Below (sticky n)
	//   row 1, col 0: Right (sticky w)
	//   row 1, col 1: center frame
	//   row 1, col 2: Left  (sticky e)
	//   row 2, col 1: Above (sticky s)

	mbBelow := makeMB(body, "Below", menubutton.Below)
	grid.Grid(mbBelow, grid.Row(0), grid.Column(1), grid.Sticky(grid.StickN))

	mbRight := makeMB(body, "Right", menubutton.Right)
	grid.Grid(mbRight, grid.Row(1), grid.Column(0), grid.Sticky(grid.StickW))

	center := frame.New(body, "center")
	grid.Grid(center, grid.Row(1), grid.Column(1), grid.Sticky(grid.NSEW))

	mbLeft := makeMB(body, "Left", menubutton.Left)
	grid.Grid(mbLeft, grid.Row(1), grid.Column(2), grid.Sticky(grid.StickE))

	mbAbove := makeMB(body, "Above", menubutton.Above)
	grid.Grid(mbAbove, grid.Row(2), grid.Column(1), grid.Sticky(grid.StickS))

	// Center label.
	centerLabel := label.New(center, "lbl",
		label.Text("This is a demonstration of menubuttons."),
		label.Anchor(option.AnchorCenter),
	)
	pack.Pack(centerLabel, pack.PadX("18p"), pack.PadY("18p"))

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
