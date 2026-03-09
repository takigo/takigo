// Demo: Simple form with labeled entries using pack layout.
// Ported from Tk's form.tcl demo.
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
	app, err := takigo.NewApp(takigo.Title("Form Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("form"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength("4i"),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("This window contains a simple form where you can type in the various entries and use tabs to move circularly between the entries."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Form rows using pack layout (matching Tcl's per-row frames).
	// Tcl only labels 3 of 5 rows: Name, Address, Phone.
	labels := []string{"Name:", "Address:", "", "", "Phone:"}

	var firstEntry *entry.Entry
	for i, labelText := range labels {
		row := frame.New(f, fmt.Sprintf("f%d", i+1),
			frame.BorderWidth(2),
		)
		e := entry.New(row, "entry",
			entry.Width(40),
		)
		l := label.New(row, "label")
		if labelText != "" {
			l.Text = labelText
			l.Display()
		}
		pack.Pack(e, pack.SideOpt(pack.Right))
		pack.Pack(l, pack.SideOpt(pack.Left))
		pack.Pack(row, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

		if i == 0 {
			firstEntry = e
		}
	}

	// Set focus to first entry.
	if firstEntry != nil {
		app.After(0, func() {
			app.FocusManager().SetFocus(firstEntry.Window())
		})
	}

	app.Run()
}
