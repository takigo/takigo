// Demo: Menu buttons with different directions.
// Ported from Tk's menubu.tcl demo.
package main

import (
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/menu"
	"github.com/msorc/takigo/widget/menubutton"
)

func main() {
	d := demohelper.Setup("Menubutton Directions", 500, 350,
		"Four menu buttons showing menus in different directions.\nClick each to see the menu position.")
	defer d.App.Destroy()
	root, app := d.Root, d.App

	statusLabel := label.New(root, "status", app,
		label.Text("Status: Ready"),
		label.Anchor(option.AnchorW),
		label.Background("#e8e8e8"),
		label.PadX(5), label.PadY(2),
	)
	pack.Pack(statusLabel.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	setStatus := func(s string) {
		statusLabel.Text = s
		statusLabel.Display()
	}

	// Create menu buttons with different directions.
	mbFrame := frame.New(root, "mbframe", app)
	pack.Pack(mbFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(20), pack.PadY(20))

	directions := []struct {
		name string
		dir  menubutton.Direction
	}{
		{"Below", menubutton.Below},
		{"Above", menubutton.Above},
		{"Left", menubutton.Left},
		{"Right", menubutton.Right},
	}

	for _, d := range directions {
		m := menu.New(root, "menu_"+d.name, app)
		dirName := d.name
		m.AddCommand("Item 1", func() { setStatus(dirName + " > Item 1") })
		m.AddCommand("Item 2", func() { setStatus(dirName + " > Item 2") })
		m.AddCommand("Item 3", func() { setStatus(dirName + " > Item 3") })

		mb := menubutton.New(mbFrame.Window(), "mb_"+d.name, app,
			menubutton.Text(d.name),
			menubutton.MenuOpt(m),
			menubutton.DirectionOpt(d.dir),
		)
		pack.Pack(mb.Window(), pack.SideOpt(pack.Left), pack.PadX(10), pack.PadY(10))
		_ = mb
	}

	_ = statusLabel
	d.Run()
}
