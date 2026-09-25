// Demo: TTK toolbar with styled buttons, checkbutton, menubutton, and combobox.
// Ported from Tk's toolbar.tcl demo.
package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/cursor"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/menu"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Toolbar Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("toolbar"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Toolbar frame (classic frame, matching Tcl's "Must be a frame!").
	toolbar := frame.New(f, "toolbar")
	grid.Grid(toolbar, grid.Sticky(grid.EW))
	tearoff := ttk.NewFrame(toolbar, "tearoff")
	tearoff.Win.SetCursor(uint(cursor.Fleur))
	to := ttk.NewSeparator(tearoff, "to", ttk.SeparatorOrient(ttk.Vertical))
	to2 := ttk.NewSeparator(tearoff, "to2", ttk.SeparatorOrient(ttk.Vertical))
	pack.Pack(to, pack.FillOpt(pack.FillY), pack.Expand(true), pack.PadX("3p"), pack.SideOpt(pack.Left))
	pack.Pack(to2, pack.FillOpt(pack.FillY), pack.Expand(true), pack.SideOpt(pack.Left))
	// The toolbar items are gridded "-in" contents in Tk; takigo's grid has
	// no -in, so they are its children.
	contents := ttk.NewFrame(toolbar, "contents")
	grid.Grid(tearoff, grid.Row(0), grid.Column(0), grid.Sticky(grid.NSEW))
	grid.Grid(contents, grid.Row(0), grid.Column(1), grid.Sticky(grid.NSEW))
	grid.ColumnConfigure(toolbar, 1, grid.Weight(1))
	grid.ColumnConfigure(contents, 1000, grid.Weight(1))

	// Separator below toolbar.
	sep := ttk.NewSeparator(f, "sep")
	grid.Grid(sep, grid.Sticky(grid.EW))

	// Description label.
	msg := ttk.NewLabel(f, "msg",
		ttk.LabelWrapLength("4i"),
		ttk.LabelText("This is a demonstration of how to do a toolbar that is styled correctly and which can be torn off. The buttons are configured to be “toolbar style” buttons by telling them that they are to use the Toolbutton style. At the left end of the toolbar is a simple marker that the cursor changes to a movement icon over; drag that away from the toolbar to tear off the whole toolbar into a separate toplevel widget. When the dragged-off toolbar is no longer needed, just close it like any normal toplevel and it will reattach to the window it was torn off from."),
	)
	grid.Grid(msg, grid.Sticky(grid.EW))

	// Text widget for output messages.
	tw := text.New(f, "txt",
		text.Width(40),
		text.Height(10),
	)
	grid.Grid(tw, grid.Sticky(grid.NSEW))
	grid.RowConfigure(f, 3, grid.Weight(1))
	grid.ColumnConfigure(f, 0, grid.Weight(1))

	// See Code / Dismiss buttons.
	btns := demohelper.AddSeeDismiss(f)
	grid.Grid(btns, grid.Sticky(grid.EW))

	// Helper to append text to the output widget.
	appendMsg := func(msg string) {
		tw.Insert("end", msg+"\n")
		tw.Display()
	}

	// --- Toolbar contents (matches Tk's toolbar.tcl) ---

	// Button (Toolbutton style: flat, raised on hover).
	btnNew := ttk.NewButton(contents, "button",
		ttk.ButtonStyleOpt("Toolbutton"),
		ttk.ButtonText("Button"),
		ttk.ButtonCommand(func() { appendMsg("Button Pressed") }),
	)

	checkVar := widget.NewVariable(false)
	checkBtn := ttk.NewCheckbutton(contents, "check",
		ttk.CheckbuttonStyleOpt("Toolbutton"),
		ttk.CheckbuttonText("Check"),
		ttk.CheckbuttonVar(checkVar),
		ttk.CheckbuttonCommand(func() {
			appendMsg(fmt.Sprintf("check is %v", checkVar.Get()))
		}),
	)

	// Menubutton with example commands.
	exMenu := menu.New(app, "m", menu.TearOffOpt(true))
	exMenu.AddCommand("Just", func() { appendMsg("Just") })
	exMenu.AddCommand("An", func() { appendMsg("An") })
	exMenu.AddCommand("Example", func() { appendMsg("Example") })

	menuBtn := ttk.NewMenubutton(contents, "menu",
		ttk.MenubuttonText("Menu"),
		ttk.MenubuttonMenu(exMenu),
	)

	// Font family combobox.
	families := font.ListFamilies()
	sort.Strings(families)
	combo := ttk.NewCombobox(contents, "combo",
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

	// Grid toolbar items (matching Tcl's single grid line).
	grid.Grid(btnNew, grid.Row(0), grid.Column(0),
		grid.PadX("1.5p"), grid.PadY("3p"), grid.Sticky(grid.NS))
	grid.Grid(checkBtn, grid.Row(0), grid.Column(1),
		grid.PadX("1.5p"), grid.PadY("3p"), grid.Sticky(grid.NS))
	grid.Grid(menuBtn, grid.Row(0), grid.Column(2),
		grid.PadX("1.5p"), grid.PadY("3p"), grid.Sticky(grid.NS))
	grid.Grid(combo, grid.Row(0), grid.Column(3),
		grid.PadX("1.5p"), grid.PadY("3p"), grid.Sticky(grid.NS))

	_ = sep
	app.Run()
}
