// Demo: Simple form with labeled entries using grid layout.
// Ported from Tk's form.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/focus"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/entry"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Form Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("form"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	root := app.Window()

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength("4i"),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("This window contains a simple form where you can type in the various entries and use tabs to move circularly between the entries."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	focusMgr := focus.NewManager(app.Dispatcher(), app.Server())
	focusMgr.BindTraversal(root)

	// Form grid.
	formFrame := frame.New(f, "form")
	pack.Pack(formFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(20), pack.PadY(10))

	fields := []string{"Name:", "Address:", "City:", "State:", "Phone:"}
	entries := make([]*entry.Entry, len(fields))

	for i, fieldName := range fields {
		l := label.New(formFrame, fmt.Sprintf("l%d", i),
			label.Text(fieldName),
			label.Anchor(option.AnchorE),
		)
		e := entry.New(formFrame, fmt.Sprintf("e%d", i),
			entry.Width(30),
		)
		grid.Grid(l, grid.Row(i), grid.Column(0), grid.Sticky(grid.StickE), grid.PadX(5), grid.PadY(4))
		grid.Grid(e, grid.Row(i), grid.Column(1), grid.Sticky(grid.EW), grid.PadX(5), grid.PadY(4))
		entries[i] = e
		_ = l
	}

	grid.ColumnConfigure(formFrame, 1, grid.Weight(1))

	_ = focusMgr
	_ = entries
	app.Run()
}
