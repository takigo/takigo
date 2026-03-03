// Demo: Menu buttons with different directions.
// Ported from Tk's menubu.tcl demo.
package main

import (
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/menu"
	"github.com/msorc/takigo/widget/menubutton"
)

func main() {
	app := demohelper.Setup("Menubutton Demonstration", 500, 350,
		"This is a demonstration of menubuttons. The \"Below\" menubutton pops its menu below the button; the \"Right\" button pops to the right, etc.")

	statusLabel := label.New(app, "status",
		label.Text("Status: Ready"),
		label.Anchor(option.AnchorW),
		label.Background("#e8e8e8"),
		label.PadX(5), label.PadY(2),
	)
	pack.Pack(statusLabel, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	setStatus := func(s string) {
		statusLabel.Text = s
		statusLabel.Display()
	}

	// Helper to create a menubutton with a 2-item menu.
	makeMB := func(parent widget.Caregiver, name string, dir menubutton.Direction) *menubutton.Menubutton {
		m := menu.New(app, "menu_"+name)
		m.AddCommand(name+" menu: first item", func() { setStatus(name + " menu: first item") })
		m.AddCommand(name+" menu: second item", func() { setStatus(name + " menu: second item") })

		return menubutton.New(parent, "mb_"+name,
			menubutton.Text(name),
			menubutton.MenuOpt(m),
			menubutton.DirectionOpt(dir),
		)
	}

	// Body frame for the compass layout.
	body := frame.New(app, "body")
	pack.Pack(body, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(20), pack.PadY(20))

	// Top row: "Below" centered.
	topRow := frame.New(body, "top_row")
	pack.Pack(topRow, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))
	mbBelow := makeMB(topRow, "Below", menubutton.Below)
	pack.Pack(mbBelow, pack.PadX(10), pack.PadY(10))

	// Middle row: "Right" on the left side, "Left" on the right side.
	midRow := frame.New(body, "mid_row")
	pack.Pack(midRow, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true))
	mbRight := makeMB(midRow, "Right", menubutton.Right)
	pack.Pack(mbRight, pack.SideOpt(pack.Left), pack.PadX(10), pack.PadY(10))
	mbLeft := makeMB(midRow, "Left", menubutton.Left)
	pack.Pack(mbLeft, pack.SideOpt(pack.Right), pack.PadX(10), pack.PadY(10))

	// Bottom row: "Above" centered.
	botRow := frame.New(body, "bot_row")
	pack.Pack(botRow, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))
	mbAbove := makeMB(botRow, "Above", menubutton.Above)
	pack.Pack(mbAbove, pack.PadX(10), pack.PadY(10))

	_ = statusLabel
	_ = mbBelow
	_ = mbLeft
	_ = mbRight
	_ = mbAbove
	app.Run()
}
