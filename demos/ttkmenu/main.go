// Demo: TTK Menubuttons arranged in compass directions for theme selection.
// Ported from Tk's ttkmenu.tcl demo.
package main

import (
	"fmt"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget/menu"
)

func main() {
	app := demohelper.Setup("Themed Menu Buttons", 400, 350,
		"Ttk is the new Tk themed widget set, and one widget that is available in themed form is the menubutton. Below are some themed menu buttons that allow you to pick the current theme in use. Notice how picking a theme changes the way that the menu buttons themselves look, and that the central menu button is styled differently (in a way that is normally suitable for toolbars). However, there are no themed menus; the standard Tk menus were judged to have a sufficiently good look-and-feel on all platforms, especially as they are implemented as native controls in many places.")

	ttk.SetCurrentTheme("clam")

	// Container frame for the compass grid layout.
	f := ttk.NewFrame(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true), pack.PadY(10))

	// Create five menubuttons in compass directions: above, left, center, right, below.
	m1 := ttk.NewMenubutton(f, "m1",
		ttk.MenubuttonText("Select a theme"),
		ttk.MenubuttonDirection(ttk.DirAbove),
	)
	m2 := ttk.NewMenubutton(f, "m2",
		ttk.MenubuttonText("Select a theme"),
		ttk.MenubuttonDirection(ttk.DirLeft),
	)
	m3 := ttk.NewMenubutton(f, "m3",
		ttk.MenubuttonText("Select a theme"),
		ttk.MenubuttonDirection(ttk.DirRight),
	)
	m4 := ttk.NewMenubutton(f, "m4",
		ttk.MenubuttonText("Select a theme"),
	)
	m5 := ttk.NewMenubutton(f, "m5",
		ttk.MenubuttonText("Select a theme"),
		ttk.MenubuttonDirection(ttk.DirBelow),
	)

	// Build menus listing available themes; selecting one switches the theme.
	themes := ttk.ThemeNames()
	buttons := []*ttk.Menubutton{m1, m2, m3, m4, m5}
	for i, mb := range buttons {
		m := menu.New(f, fmt.Sprintf("menu%d", i+1))
		for _, themeName := range themes {
			name := themeName // capture for closure
			m.AddCommand(name, func() {
				ttk.SetCurrentTheme(name)
			})
		}
		mb.Menu = m
	}

	// Grid layout: compass arrangement.
	//     .  m1  .
	//    m2  m4  m3
	//     .  m5  .
	grid.Grid(m1, grid.Row(0), grid.Column(1), grid.PadX("2.25p"), grid.PadY("1.5p"))
	grid.Grid(m2, grid.Row(1), grid.Column(0), grid.PadX("2.25p"), grid.PadY("1.5p"))
	grid.Grid(m4, grid.Row(1), grid.Column(1), grid.PadX("2.25p"), grid.PadY("1.5p"))
	grid.Grid(m3, grid.Row(1), grid.Column(2), grid.PadX("2.25p"), grid.PadY("1.5p"))
	grid.Grid(m5, grid.Row(2), grid.Column(1), grid.PadX("2.25p"), grid.PadY("1.5p"))

	app.Run()
}
