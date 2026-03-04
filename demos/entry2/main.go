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

	focusMgr := focus.NewManager(app.Dispatcher(), app.Server())
	focusMgr.BindTraversal(root)

	// Container frame with border (matches Tcl's `frame -borderwidth 7.5p`).
	fr := frame.New(app, "frame",
		frame.BorderWidth(10), // 7.5p ≈ 10px
	)
	pack.Pack(fr, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.Expand(true))

	// Helper: add an entry+scrollbar pair to the frame.
	makeEntry := func(name, text, placeholder string) *entry.Entry {
		e := entry.New(fr, name)
		if placeholder != "" {
			e.Placeholder = placeholder
		}
		sb := scrollbar.New(fr, name+"_sb",
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
		e.ScrollCmd = func(first, last float64) {
			sb.Set(first, last)
		}
		if text != "" {
			e.SetText(text)
		}
		pack.Pack(e, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))
		pack.Pack(sb, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))
		return e
	}

	e1 := makeEntry("e1", "Initial value", "")

	// Spacer between e1 and e2 (matches Tcl's spacer frame height 7.5p).
	sp1 := frame.New(fr, "spacer1", frame.Height(10))
	pack.Pack(sp1, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	e2 := makeEntry("e2",
		"This entry contains a long value, much too long "+
			"to fit in the window at one time, so long in fact "+
			"that you'll have to scan or scroll to see the end.",
		"")

	// Spacer between e2 and e3.
	sp2 := frame.New(fr, "spacer2", frame.Height(10))
	pack.Pack(sp2, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	e3 := makeEntry("e3", "", "Enter text here")

	_ = focusMgr
	_ = e1
	_ = e2
	_ = e3
	_ = sp1
	_ = sp2
	app.Run()
}
