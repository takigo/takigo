// Demo: Basic entry widgets without scrollbars.
// Ported from Tk's entry1.tcl demo.
package main

import (
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/focus"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/widget/entry"
)

func main() {
	d := demohelper.Setup("Entry Demonstration (no scrollbars)", 450, 250,
		"Three entry widgets are displayed below. You can click in\nan entry and type text. Use Tab to move between entries.")
	root, app := d.Root, d.App

	focusMgr := focus.NewManager(app.Dispatcher(), app.DisplayPtr())
	focusMgr.BindTraversal(root)

	// Entry 1: pre-populated.
	e1 := entry.New(root, "e1", app,
		entry.Text("Initial value"),
	)
	pack.Pack(e1, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(20), pack.PadY(5))

	// Entry 2: long text that requires scrolling.
	e2 := entry.New(root, "e2", app,
		entry.Text("This entry contains a long value, much too long to fit in the window at one time, and thus you can use the scrollbar to see the rest."),
	)
	pack.Pack(e2, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(20), pack.PadY(5))

	// Entry 3: placeholder text.
	e3 := entry.New(root, "e3", app,
		entry.Placeholder("Enter text here..."),
	)
	pack.Pack(e3, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(20), pack.PadY(5))

	_ = focusMgr
	_ = e1
	_ = e2
	_ = e3
	d.Run()
}
