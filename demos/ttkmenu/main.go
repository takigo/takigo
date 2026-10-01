// Demo: Several Ttk menubutton widgets.
// Ported from Tk's ttkmenu.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/menu"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Ttk Menu Buttons"),
		takigo.Geometry("+300+300"),
		takigo.IconName("ttkmenu"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := ttk.NewLabel(f, "msg",
		ttk.LabelWrapLength(screenunit.In(4)),
		ttk.LabelJustify(option.JustifyLeft),
		ttk.LabelText("Ttk is the new Tk themed widget set, and one widget that is available in themed form is the menubutton. Below are some themed menu buttons that allow you to pick the current theme in use. Notice how picking a theme changes the way that the menu buttons themselves look, and that the central menu button is styled differently (in a way that is normally suitable for toolbars). However, there are no themed menus; the standard Tk menus were judged to have a sufficiently good look-and-feel on all platforms, especially as they are implemented as native controls in many places."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	msgSep := ttk.NewSeparator(f, "msgSep")
	pack.Pack(msgSep, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Container frame for the compass grid layout.
	cf := ttk.NewFrame(f, "cf")
	pack.Pack(cf, pack.FillOpt(pack.FillX))
	f1 := ttk.NewFrame(f, "f1")
	pack.Pack(f1, pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Create five menubuttons in compass directions: above, left, center, right, below.
	m1 := ttk.NewMenubutton(cf, "m1",
		ttk.MenubuttonText("Select a theme"),
		ttk.MenubuttonDirection(ttk.DirAbove),
	)
	m2 := ttk.NewMenubutton(cf, "m2",
		ttk.MenubuttonText("Select a theme"),
		ttk.MenubuttonDirection(ttk.DirLeft),
	)
	m3 := ttk.NewMenubutton(cf, "m3",
		ttk.MenubuttonText("Select a theme"),
		ttk.MenubuttonDirection(ttk.DirRight),
	)
	m4 := ttk.NewMenubutton(cf, "m4",
		ttk.MenubuttonText("Select a theme"),
		ttk.MenubuttonStyleOpt("TMenubutton.Toolbutton"),
	)
	m5 := ttk.NewMenubutton(cf, "m5",
		ttk.MenubuttonText("Select a theme"),
		ttk.MenubuttonDirection(ttk.DirBelow),
	)

	// Build menus listing available themes; selecting one switches the theme.
	themes := ttk.ThemeNames()
	buttons := []*ttk.Menubutton{m1, m2, m3, m4, m5}
	for i, mb := range buttons {
		m := menu.New(cf, fmt.Sprintf("menu%d", i+1))
		for _, themeName := range themes {
			name := themeName // capture for closure
			m.AddCommand(name, func() {
				_ = ttk.UseTheme(app, name)
			})
		}
		mb.Menu = m
	}

	// Grid layout: compass arrangement.
	//     .  m1  .
	//    m2  m4  m3
	//     .  m5  .
	grid.SetAnchor(cf, option.AnchorCenter)
	grid.Grid(m1, grid.Row(0), grid.Column(1), grid.PadX(screenunit.Pt(2.25)), grid.PadY(screenunit.Pt(1.5)))
	grid.Grid(m2, grid.Row(1), grid.Column(0), grid.PadX(screenunit.Pt(2.25)), grid.PadY(screenunit.Pt(1.5)))
	grid.Grid(m4, grid.Row(1), grid.Column(1), grid.PadX(screenunit.Pt(2.25)), grid.PadY(screenunit.Pt(1.5)))
	grid.Grid(m3, grid.Row(1), grid.Column(2), grid.PadX(screenunit.Pt(2.25)), grid.PadY(screenunit.Pt(1.5)))
	grid.Grid(m5, grid.Row(2), grid.Column(1), grid.PadX(screenunit.Pt(2.25)), grid.PadY(screenunit.Pt(1.5)))

	app.Run()
}
