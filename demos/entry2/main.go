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
	d := demohelper.Setup("Entry Demonstration (with scrollbars)", 450, 320,
		"Three entry widgets with scrollbars are displayed below.\nYou can add characters by pointing, clicking and typing.")
	root, app := d.Root, d.App

	focusMgr := focus.NewManager(app.Dispatcher(), app.DisplayPtr())
	focusMgr.BindTraversal(root)

	// Helper: create entry+scrollbar pair.
	makeEntryWithScroll := func(parent *frame.Frame, name, text string) *entry.Entry {
		ef := frame.New(parent.Window(), name+"_frame", app)
		pack.Pack(ef, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
			pack.PadX(20), pack.PadY(5))

		e := entry.New(ef.Window(), name, app)
		sb := scrollbar.New(ef.Window(), name+"_sb", app,
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
	container := frame.New(root, "container", app)
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
	d.Run()
}
