// Demo: TTK toolbar with buttons, separator, menubutton, and combobox.
// Ported from Tk's toolbar.tcl demo.
package main

import (
	"fmt"
	"sort"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget/menu"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app := demohelper.Setup("Toolbar Demonstration", 500, 350,
		"This is a demonstration of how to do a toolbar that is styled correctly. The buttons are configured to be \"toolbar style\" buttons by telling them that they are to use the Toolbutton style. Below the toolbar is a text widget that shows messages when toolbar items are activated.")

	ttk.SetCurrentTheme("clam")

	// Inner frame (packs into app alongside demohelper's msg/btnframe).
	main := ttk.NewFrame(app, "main")
	pack.Pack(main, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Toolbar frame (row 0 inside main).
	toolbar := ttk.NewFrame(main, "toolbar",
		ttk.FrameBorderWidth(1),
		ttk.FrameRelief(option.ReliefRaised),
	)
	grid.Grid(toolbar, grid.Row(0), grid.Column(0), grid.Sticky(grid.EW))

	// Separator below toolbar (row 1).
	sep0 := ttk.NewSeparator(main, "toolsep", ttk.SeparatorOrient(ttk.Horizontal))
	grid.Grid(sep0, grid.Row(1), grid.Column(0), grid.Sticky(grid.EW))

	// Text widget for output messages, no scrollbar (row 2, expands).
	tw := text.New(main, "txt",
		text.Width(40),
		text.Height(10),
	)
	grid.Grid(tw, grid.Row(2), grid.Column(0), grid.Sticky(grid.NSEW))
	grid.RowConfigure(main.Window(), 2, grid.SlotConfig{Weight: 1})
	grid.ColumnConfigure(main.Window(), 0, grid.SlotConfig{Weight: 1})

	// Helper to append text to the output widget.
	appendMsg := func(msg string) {
		tw.Insert("end", msg+"\n")
		tw.Display()
	}

	// --- Toolbar contents (matches Tk's toolbar.tcl) ---

	// Button (Toolbutton style: flat, raised on hover).
	btnNew := ttk.NewButton(toolbar, "button",
		ttk.ButtonStyleOpt("Toolbutton"),
		ttk.ButtonText("Button"),
		ttk.ButtonCommand(func() { appendMsg("Button Pressed") }),
	)
	pack.Pack(btnNew, pack.SideOpt(pack.Left), pack.PadX("1.5p"), pack.PadY("3p"))

	// Check button (TTK checkbutton).
	checkVar := widget.NewVariable(false)
	checkBtn := ttk.NewCheckbutton(toolbar, "check",
		ttk.CheckbuttonText("Check"),
		ttk.CheckbuttonVar(checkVar),
		ttk.CheckbuttonCommand(func() {
			appendMsg(fmt.Sprintf("check is %v", checkVar.Get()))
		}),
	)
	pack.Pack(checkBtn, pack.SideOpt(pack.Left), pack.PadX("1.5p"), pack.PadY("3p"))

	// Vertical separator.
	sep := ttk.NewSeparator(toolbar, "sep", ttk.SeparatorOrient(ttk.Vertical))
	pack.Pack(sep, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillY), pack.PadX("1.5p"), pack.PadY("3p"))

	// Menubutton with example commands (Toolbutton style).
	exMenu := menu.New(app, "exmenu", menu.TearOffOpt(true))
	exMenu.AddCommand("Just", func() { appendMsg("Just") })
	exMenu.AddCommand("An", func() { appendMsg("An") })
	exMenu.AddCommand("Example", func() { appendMsg("Example") })

	menuBtn := ttk.NewMenubutton(toolbar, "menu",
		ttk.MenubuttonStyleOpt("TMenubutton.Toolbutton"),
		ttk.MenubuttonText("Menu"),
		ttk.MenubuttonMenu(exMenu),
	)
	pack.Pack(menuBtn, pack.SideOpt(pack.Left), pack.PadX("1.5p"), pack.PadY("3p"))

	// Font family combobox.
	families := font.ListFamilies()
	sort.Strings(families)
	combo := ttk.NewCombobox(toolbar, "combo",
		ttk.ComboboxValues(families),
		ttk.ComboboxCbState(ttk.ComboReadonly),
		ttk.ComboboxCommand(func(val string) {
			f, err := app.FontRegistry().Get(val + " 10")
			if err == nil {
				tw.Font = f
				tw.Display()
			}
		}),
	)
	pack.Pack(combo, pack.SideOpt(pack.Left), pack.PadX("1.5p"), pack.PadY("3p"))

	_ = sep0
	_ = btnNew
	_ = checkBtn
	_ = sep
	_ = menuBtn
	_ = combo
	app.Run()
}
