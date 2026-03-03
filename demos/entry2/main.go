// Demo: Entry widgets with horizontal scrollbars.
// Ported from Tk's entry2.tcl demo.
package main

import (
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/focus"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/widget/entry"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/scrollbar"
)

func main() {
	app := demohelper.Setup("Entry Demonstration (with scrollbars)", 450, 320,
		"Three different entries are displayed below, with a scrollbar for each entry. You can add characters by pointing, clicking and typing. The normal editing characters are supported, along with many Emacs bindings. For entries that are too large to fit in the window all at once, you can scan through the entries with the scrollbars.")
	root := app.Window()

	focusMgr := focus.NewManager(app.Dispatcher(), app.DisplayPtr())
	focusMgr.BindTraversal(root)

	// Helper: create entry+scrollbar pair.
	makeEntryWithScroll := func(parent *frame.Frame, name, text string) *entry.Entry {
		ef := frame.New(parent, name+"_frame")
		pack.Pack(ef, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
			pack.PadX(20), pack.PadY(5))

		e := entry.New(ef, name)
		sb := scrollbar.New(ef, name+"_sb",
			scrollbar.OrientOpt(scrollbar.Horizontal),
			scrollbar.CommandOpt(func(args ...any) {
				if len(args) < 1 {
					return
				}
				switch args[0] {
				case "moveto":
					if len(args) >= 2 {
						if f, ok := args[1].(float64); ok {
							e.XViewMoveTo(f)
						}
					}
				case "scroll":
					if len(args) >= 3 {
						n, _ := args[1].(int)
						unit, _ := args[2].(string)
						e.XViewScroll(n, unit == "pages")
					}
				}
			}),
		)

		e.SetText(text)
		e.ScrollCmd = func(first, last float64) {
			sb.Set(first, last)
		}

		pack.Pack(e, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))
		pack.Pack(sb, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

		return e
	}

	// Container frame.
	container := frame.New(app, "container")
	pack.Pack(container, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Three entry+scrollbar pairs.
	e1 := makeEntryWithScroll(container, "e1", "Initial value")
	e2 := makeEntryWithScroll(container, "e2",
		"This entry contains a long value, much too long to fit in the window at one time, and thus you can use the scrollbar to see the rest.")
	e3 := makeEntryWithScroll(container, "e3", "")

	_ = focusMgr
	_ = e1
	_ = e2
	_ = e3
	app.Run()
}
