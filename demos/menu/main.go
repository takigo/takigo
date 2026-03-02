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
	app := demohelper.Setup("Menu Demo", 500, 350,
		"Click the menu buttons below to open menus.\nMenus support commands, separators, and cascaded submenus.")

	// Status label.
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

	// Menu bar frame.
	menuBar := frame.New(app, "menubar",
		frame.Relief(option.ReliefRaised),
		frame.BorderWidth(1),
	)
	pack.Pack(menuBar, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// File menu.
	fileMenu := menu.New(app, "filemenu")
	fileMenu.AddCommand("New", func() { setStatus("File > New") })
	fileMenu.AddCommand("Open...", func() { setStatus("File > Open") })
	fileMenu.AddCommand("Save", func() { setStatus("File > Save") })
	fileMenu.AddCommand("Save As...", func() { setStatus("File > Save As") })
	fileMenu.AddSeparator()
	fileMenu.AddCommand("Quit", func() { app.Quit() })

	fileMb := menubutton.New(menuBar, "filemb",
		menubutton.Text("File"),
		menubutton.MenuOpt(fileMenu),
	)
	pack.Pack(fileMb, pack.SideOpt(pack.Left))

	// Edit menu with cascade.
	editMenu := menu.New(app, "editmenu")
	editMenu.AddCommand("Undo", func() { setStatus("Edit > Undo") })
	editMenu.AddCommand("Redo", func() { setStatus("Edit > Redo") })
	editMenu.AddSeparator()
	editMenu.AddCommand("Cut", func() { setStatus("Edit > Cut") })
	editMenu.AddCommand("Copy", func() { setStatus("Edit > Copy") })
	editMenu.AddCommand("Paste", func() { setStatus("Edit > Paste") })

	editMb := menubutton.New(menuBar, "editmb",
		menubutton.Text("Edit"),
		menubutton.MenuOpt(editMenu),
	)
	pack.Pack(editMb, pack.SideOpt(pack.Left))

	// Help menu.
	helpMenu := menu.New(app, "helpmenu")
	helpMenu.AddCommand("About", func() { setStatus("Help > About: Takigo Menu Demo") })
	helpMenu.AddCommand("Documentation", func() { setStatus("Help > Documentation") })

	helpMb := menubutton.New(menuBar, "helpmb",
		menubutton.Text("Help"),
		menubutton.MenuOpt(helpMenu),
	)
	pack.Pack(helpMb, pack.SideOpt(pack.Left))

	_ = statusLabel
	_ = fileMb
	_ = editMb
	_ = helpMb
	app.Run()
}
