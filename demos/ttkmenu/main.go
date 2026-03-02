// Demo: TTK Menubuttons with different menu configurations.
// Ported from Tk's ttkmenu.tcl demo (simplified — no tearoff).
package main

import (
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/menu"
)

func main() {
	d := demohelper.Setup("TTK Menubutton Demonstration", 400, 350,
		"Below are themed menubuttons. Click each to\nopen a dropdown menu.")
	app := d.App

	ttk.SetCurrentTheme("clam")

	// Status label.
	statusLabel := label.New(app, "status",
		label.Text("Selected: (none)"),
		label.Anchor(option.AnchorW),
		label.PadX(10),
	)
	pack.Pack(statusLabel, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))

	setStatus := func(s string) {
		statusLabel.Text = "Selected: " + s
		statusLabel.Display()
	}

	// File menubutton.
	fileMenu := menu.New(app, "filemenu")
	fileMenu.AddCommand("New", func() { setStatus("File > New") })
	fileMenu.AddCommand("Open", func() { setStatus("File > Open") })
	fileMenu.AddSeparator()
	fileMenu.AddCommand("Save", func() { setStatus("File > Save") })
	fileMenu.AddCommand("Close", func() { setStatus("File > Close") })

	fileMB := ttk.NewMenubutton(app, "filemb",
		ttk.MenubuttonText("File"),
		ttk.MenubuttonMenu(fileMenu),
	)
	pack.Pack(fileMB, pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(10))

	// Edit menubutton.
	editMenu := menu.New(app, "editmenu")
	editMenu.AddCommand("Cut", func() { setStatus("Edit > Cut") })
	editMenu.AddCommand("Copy", func() { setStatus("Edit > Copy") })
	editMenu.AddCommand("Paste", func() { setStatus("Edit > Paste") })

	editMB := ttk.NewMenubutton(app, "editmb",
		ttk.MenubuttonText("Edit"),
		ttk.MenubuttonMenu(editMenu),
	)
	pack.Pack(editMB, pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(10))

	// Help menubutton.
	helpMenu := menu.New(app, "helpmenu")
	helpMenu.AddCommand("About", func() { setStatus("Help > About") })
	helpMenu.AddCommand("Documentation", func() { setStatus("Help > Documentation") })

	helpMB := ttk.NewMenubutton(app, "helpmb",
		ttk.MenubuttonText("Help"),
		ttk.MenubuttonMenu(helpMenu),
	)
	pack.Pack(helpMB, pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(10))

	_ = statusLabel
	_ = fileMB
	_ = editMB
	_ = helpMB
	d.Run()
}
