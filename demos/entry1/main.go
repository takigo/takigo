// Demo: Several entry widgets without scrollbars.
// Ported from Tk's entry1.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/entry"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Entry Demonstration (no scrollbars)"),
		takigo.Geometry("+300+300"),
		takigo.IconName("entry1"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength("5i"),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("Three different entries are displayed below.  You can add characters by pointing, clicking and typing.  The normal Motif editing characters are supported, along with many Emacs bindings.  For example, Backspace and Control-h delete the character to the left of the insertion cursor and Delete and Control-d delete the chararacter to the right of the insertion cursor.  For entries that are too large to fit in the window all at once, you can scan through the entries by dragging with mouse the middle mouse button pressed."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Entry 1: pre-populated.
	e1 := entry.New(f, "e1",
		entry.Text("Initial value"),
	)
	pack.Pack(e1, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX("7.5p"), pack.PadY("3p"))

	// Entry 2: long text that requires scrolling.
	e2 := entry.New(f, "e2",
		entry.Text("This entry contains a long value, much too long "+
			"to fit in the window at one time, so long in fact "+
			"that you'll have to scan or scroll to see the end."),
	)
	pack.Pack(e2, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX("7.5p"), pack.PadY("3p"))

	// Entry 3: placeholder text.
	e3 := entry.New(f, "e3",
		entry.Placeholder("Enter text here"),
		entry.PlaceholderForeground("gray75"),
	)
	pack.Pack(e3, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX("7.5p"), pack.PadY("3p"))

	_ = e1
	_ = e2
	_ = e3
	app.Run()
}
