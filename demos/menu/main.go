// Demo: Menu bar with cascading submenus.
// Ported from Tk's menu.tcl demo.
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
	d := demohelper.Setup("Menu Demo", 500, 350,
		"Click the menu buttons below to open menus.\nMenus support commands, separators, and cascaded submenus.")
	defer d.App.Destroy()
	root, app := d.Root, d.App

	// Status label.
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

	// Menu bar frame.
	menuBar := frame.New(root, "menubar", app,
		frame.Relief(option.ReliefRaised),
		frame.BorderWidth(1),
	)
	pack.Pack(menuBar.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// File menu.
	fileMenu := menu.New(root, "filemenu", app)
	fileMenu.AddCommand("New", func() { setStatus("File > New") })
	fileMenu.AddCommand("Open...", func() { setStatus("File > Open") })
	fileMenu.AddCommand("Save", func() { setStatus("File > Save") })
	fileMenu.AddCommand("Save As...", func() { setStatus("File > Save As") })
	fileMenu.AddSeparator()
	fileMenu.AddCommand("Quit", func() { app.Quit() })

	fileMb := menubutton.New(menuBar.Window(), "filemb", app,
		menubutton.Text("File"),
		menubutton.MenuOpt(fileMenu),
	)
	pack.Pack(fileMb.Window(), pack.SideOpt(pack.Left))

	// Edit menu with cascade.
	editMenu := menu.New(root, "editmenu", app)
	editMenu.AddCommand("Undo", func() { setStatus("Edit > Undo") })
	editMenu.AddCommand("Redo", func() { setStatus("Edit > Redo") })
	editMenu.AddSeparator()
	editMenu.AddCommand("Cut", func() { setStatus("Edit > Cut") })
	editMenu.AddCommand("Copy", func() { setStatus("Edit > Copy") })
	editMenu.AddCommand("Paste", func() { setStatus("Edit > Paste") })

	editMb := menubutton.New(menuBar.Window(), "editmb", app,
		menubutton.Text("Edit"),
		menubutton.MenuOpt(editMenu),
	)
	pack.Pack(editMb.Window(), pack.SideOpt(pack.Left))

	// Help menu.
	helpMenu := menu.New(root, "helpmenu", app)
	helpMenu.AddCommand("About", func() { setStatus("Help > About: Takigo Menu Demo") })
	helpMenu.AddCommand("Documentation", func() { setStatus("Help > Documentation") })

	helpMb := menubutton.New(menuBar.Window(), "helpmb", app,
		menubutton.Text("Help"),
		menubutton.MenuOpt(helpMenu),
	)
	pack.Pack(helpMb.Window(), pack.SideOpt(pack.Left))

	_ = statusLabel
	_ = fileMb
	_ = editMb
	_ = helpMb
	d.Run()
}
