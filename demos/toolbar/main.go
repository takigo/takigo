// Demo: TTK toolbar with buttons, separator, menubutton, and combobox.
// Ported from Tk's toolbar.tcl demo.
package main

import (
	"fmt"
	"sort"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget/menu"
	"github.com/msorc/takigo/widget/scrollbar"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app := demohelper.Setup("Toolbar Demonstration", 500, 350,
		"This is a demonstration of how to do a toolbar that is styled correctly. The buttons are configured to be \"toolbar style\" buttons by telling them that they are to use the Toolbutton style. Below the toolbar is a text widget that shows messages when toolbar items are activated.")

	ttk.SetCurrentTheme("clam")

	// Toolbar frame.
	toolbar := ttk.NewFrame(app, "toolbar",
		ttk.FrameBorderWidth(1),
		ttk.FrameRelief(option.ReliefRaised),
	)
	pack.Pack(toolbar, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// Separator below toolbar.
	sep0 := ttk.NewSeparator(app, "toolsep", ttk.SeparatorOrient(ttk.Horizontal))
	pack.Pack(sep0, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// Text widget frame (with scrollbar) for output messages.
	txtFrame := demohelper.NewFrame(app, "txtframe")
	pack.Pack(txtFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(4), pack.PadY(4))

	tw := text.New(txtFrame, "output",
		text.Width(40),
		text.Height(10),
		text.WrapModeOpt(text.WrapWord),
	)

	yscroll := scrollbar.New(txtFrame, "yscroll",
		scrollbar.OrientOpt(scrollbar.Vertical),
		scrollbar.CommandOpt(func(args ...any) {
			if len(args) < 1 {
				return
			}
			switch args[0] {
			case "moveto":
				if f, ok := args[1].(float64); ok {
					tw.YViewMoveTo(f)
				}
			case "scroll":
				n, _ := args[1].(int)
				unit, _ := args[2].(string)
				tw.YViewScroll(n, unit == "pages")
			}
		}),
	)

	tw.YScrollCmd = func(first, last float64) {
		yscroll.Set(first, last)
	}

	pack.Pack(yscroll, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(tw, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Helper to append text to the output widget.
	appendMsg := func(msg string) {
		tw.Insert("end", msg+"\n")
		tw.Display()
	}

	// --- Toolbar contents (matches Tk's toolbar.tcl) ---

	// Button.
	btnNew := ttk.NewButton(toolbar, "button",
		ttk.ButtonText("Button"),
		ttk.ButtonCommand(func() { appendMsg("Button Pressed") }),
	)
	pack.Pack(btnNew, pack.SideOpt(pack.Left), pack.PadX(2), pack.PadY(2))

	// Check button (simulated with a regular button toggling state).
	checkState := false
	checkBtn := ttk.NewButton(toolbar, "check",
		ttk.ButtonText("Check"),
		ttk.ButtonCommand(func() {
			checkState = !checkState
			appendMsg(fmt.Sprintf("check is %v", checkState))
		}),
	)
	pack.Pack(checkBtn, pack.SideOpt(pack.Left), pack.PadX(2), pack.PadY(2))

	// Vertical separator.
	sep := ttk.NewSeparator(toolbar, "sep", ttk.SeparatorOrient(ttk.Vertical))
	pack.Pack(sep, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillY), pack.PadX(4), pack.PadY(2))

	// Menubutton with example commands.
	exMenu := menu.New(app, "exmenu")
	exMenu.AddCommand("Just", func() { appendMsg("Just") })
	exMenu.AddCommand("An", func() { appendMsg("An") })
	exMenu.AddCommand("Example", func() { appendMsg("Example") })

	menuBtn := ttk.NewMenubutton(toolbar, "menu",
		ttk.MenubuttonText("Menu"),
		ttk.MenubuttonMenu(exMenu),
	)
	pack.Pack(menuBtn, pack.SideOpt(pack.Left), pack.PadX(2), pack.PadY(2))

	// Font family combobox.
	families := font.ListFamilies()
	sort.Strings(families)
	combo := ttk.NewCombobox(toolbar, "combo",
		ttk.ComboboxValues(families),
		ttk.ComboboxCbState(ttk.ComboReadonly),
		ttk.ComboboxCommand(func(val string) {
			appendMsg(fmt.Sprintf("Font: %s", val))
		}),
	)
	pack.Pack(combo, pack.SideOpt(pack.Left), pack.PadX(2), pack.PadY(2))

	_ = sep0
	_ = btnNew
	_ = checkBtn
	_ = sep
	_ = menuBtn
	_ = combo
	app.Run()
}
