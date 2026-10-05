// Demo: Simple form with labeled entries using pack layout.
// Ported from Tk's form.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/bind"
	"github.com/takigo/takigo/demos/demohelper"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/widget/entry"
	"github.com/takigo/takigo/widget/frame"
	"github.com/takigo/takigo/widget/label"
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
		label.WrapLength(screenunit.In(4)),
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
		l := label.New(row, "label", label.Text(labelText))
		pack.Pack(e, pack.SideOpt(pack.Right))
		pack.Pack(l, pack.SideOpt(pack.Left))
		pack.Pack(row, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

		if i == 0 {
			firstEntry = e
		}
	}

	// Bind Return to dismiss the window (matching Tcl: bind $w <Return> "destroy $w").
	eng := app.Bind()
	eng.BindWindow(app, "<Return>", func(_ *bind.EventData) bool {
		app.Quit()
		return true
	})

	// Set focus to first entry.
	if firstEntry != nil {
		app.After(0, func() {
			app.FocusManager().SetFocus(firstEntry.Window())
		})
	}

	app.Run()
}
