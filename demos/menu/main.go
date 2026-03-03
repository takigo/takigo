// Demo: Menu bar with cascading submenus.
// Ported from Tk's menu.tcl demo.
package main

import (
	"fmt"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/menu"
	"github.com/msorc/takigo/widget/menubutton"
)

func main() {
	app := demohelper.Setup("Menu Demonstration", 500, 400,
		"This window contains a menubar with cascaded menus. You can post a menu from the keyboard by typing Alt+x, where \"x\" is the character underlined on the menu. You can then traverse among the menus using the arrow keys. When a menu is posted, you can invoke the current entry by typing space, or you can invoke any entry by typing its underlined character. If a menu entry has an accelerator, you can invoke the entry without posting the menu just by typing the accelerator.")

	// Status bar at the bottom.
	statusBar := frame.New(app, "statusbar")
	statusLabel := label.New(statusBar, "status",
		label.Text("    "),
		label.Anchor(option.AnchorW),
		label.Relief(option.ReliefSunken),
		label.Background("#e8e8e8"),
		label.PadX(5), label.PadY(2),
	)
	pack.Pack(statusLabel, pack.SideOpt(pack.Left), pack.Expand(true), pack.FillOpt(pack.FillBoth), pack.PadX(2))
	pack.Pack(statusBar, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(2))

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

	// ── File menu ──
	fileMenu := menu.New(app, "filemenu")
	fileMenu.AddCommand("Open...", func() { setStatus("File > Open") })
	fileMenu.AddCommand("New", func() { setStatus("File > New") })
	fileMenu.AddCommand("Save", func() { setStatus("File > Save") })
	fileMenu.AddCommand("Save As...", func() { setStatus("File > Save As") })
	fileMenu.AddSeparator()
	fileMenu.AddCommand("Print Setup...", func() { setStatus("File > Print Setup") })
	fileMenu.AddCommand("Print...", func() { setStatus("File > Print") })
	fileMenu.AddSeparator()
	fileMenu.AddCommand("Dismiss Menus Demo", func() { app.Quit() })

	fileMb := menubutton.New(menuBar, "filemb",
		menubutton.Text("File"),
		menubutton.MenuOpt(fileMenu),
	)
	pack.Pack(fileMb, pack.SideOpt(pack.Left))

	// ── Basic menu ──
	basicMenu := menu.New(app, "basicmenu")
	basicMenu.AddCommand("Long entry that does nothing", nil)
	for _, letter := range []string{"A", "B", "C", "D", "E", "F"} {
		l := letter
		basicMenu.AddCommandAccel(fmt.Sprintf("Print letter \"%s\"", l), "Meta+"+l, func() {
			setStatus(fmt.Sprintf("Basic > Print letter \"%s\"", l))
		})
	}

	basicMb := menubutton.New(menuBar, "basicmb",
		menubutton.Text("Basic"),
		menubutton.MenuOpt(basicMenu),
	)
	pack.Pack(basicMb, pack.SideOpt(pack.Left))

	// ── Cascades menu ──
	cascadeMenu := menu.New(app, "cascademenu")

	cascadeMenu.AddCommandAccel("Print hello", "Meta+H", func() {
		setStatus("Cascades > Print hello")
	})
	cascadeMenu.AddCommandAccel("Print goodbye", "Meta+G", func() {
		setStatus("Cascades > Print goodbye")
	})

	// Check buttons submenu.
	checkMenu := menu.New(app, "checkmenu")
	checkMenu.AddCheckbutton("Oil checked", false, func() { setStatus("Cascades > Check > Oil") })
	checkMenu.AddCheckbutton("Transmission checked", true, func() { setStatus("Cascades > Check > Transmission") })
	checkMenu.AddCheckbutton("Brakes checked", false, func() { setStatus("Cascades > Check > Brakes") })
	checkMenu.AddCheckbutton("Lights checked", true, func() { setStatus("Cascades > Check > Lights") })
	checkMenu.AddSeparator()
	checkMenu.AddCommand("Show current values", func() {
		entries := checkMenu.Entries()
		msg := "Check values:"
		for _, e := range entries {
			if e.Type == menu.Checkbutton {
				val := "off"
				if e.Checked {
					val = "on"
				}
				msg += fmt.Sprintf("  %s=%s", e.Label, val)
			}
		}
		setStatus(msg)
	})

	cascadeMenu.AddCascade("Check buttons", checkMenu)

	// Radio buttons submenu.
	radioMenu := menu.New(app, "radiomenu")
	radioMenu.AddRadiobutton("10 point", true, func() { setStatus("Cascades > Radio > 10 point") })
	radioMenu.AddRadiobutton("14 point", false, func() { setStatus("Cascades > Radio > 14 point") })
	radioMenu.AddRadiobutton("18 point", false, func() { setStatus("Cascades > Radio > 18 point") })
	radioMenu.AddRadiobutton("24 point", false, func() { setStatus("Cascades > Radio > 24 point") })
	radioMenu.AddRadiobutton("32 point", false, func() { setStatus("Cascades > Radio > 32 point") })
	radioMenu.AddSeparator()
	radioMenu.AddCommand("Show current values", func() {
		entries := radioMenu.Entries()
		msg := "Radio values:"
		for _, e := range entries {
			if e.Type == menu.Radiobutton && e.Checked {
				msg += fmt.Sprintf("  selected=%s", e.Label)
			}
		}
		setStatus(msg)
	})

	cascadeMenu.AddCascade("Radio buttons", radioMenu)

	cascadeMb := menubutton.New(menuBar, "cascademb",
		menubutton.Text("Cascades"),
		menubutton.MenuOpt(cascadeMenu),
	)
	pack.Pack(cascadeMb, pack.SideOpt(pack.Left))

	// ── More menu ──
	moreMenu := menu.New(app, "moremenu")
	for _, item := range []string{
		"An entry",
		"Another entry",
		"Does nothing",
		"Does almost nothing",
		"Make life meaningful",
	} {
		lbl := item
		moreMenu.AddCommand(lbl, func() {
			setStatus(fmt.Sprintf("More > %s", lbl))
		})
	}

	moreMb := menubutton.New(menuBar, "moremb",
		menubutton.Text("More"),
		menubutton.MenuOpt(moreMenu),
	)
	pack.Pack(moreMb, pack.SideOpt(pack.Left))

	// ── Colors menu ──
	colorsMenu := menu.New(app, "colorsmenu")
	for _, color := range []string{"red", "orange", "yellow", "green", "blue"} {
		c := color
		colorsMenu.AddCommand(c, func() {
			setStatus(fmt.Sprintf("Colors > %s", c))
		})
	}

	colorsMb := menubutton.New(menuBar, "colorsmb",
		menubutton.Text("Colors"),
		menubutton.MenuOpt(colorsMenu),
	)
	pack.Pack(colorsMb, pack.SideOpt(pack.Left))

	_ = statusLabel
	_ = fileMb
	_ = basicMb
	_ = cascadeMb
	_ = moreMb
	_ = colorsMb
	app.Run()
}
