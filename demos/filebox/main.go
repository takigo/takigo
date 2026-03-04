// Demo: File open/save dialogs.
// Ported from Tk's filebox.tcl demo.
package main

import (
	"fmt"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/dialog"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/ttk"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/entry"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app := demohelper.Setup("File Selection Dialogs", 500, 200,
		"Enter a file name in the entry box or click on the \"Browse\" buttons to select a file name using the file selection dialog.")

	fileTypes := []dialog.FileType{
		{Name: "Text files", Pattern: "*.txt"},
		{Name: "Tcl Scripts", Pattern: "*.tcl"},
		{Name: "C Source Files", Pattern: "*.c"},
		{Name: "All Source Files", Pattern: "*.go"},
		{Name: "Image Files", Pattern: "*.gif"},
		{Name: "All files", Pattern: "*"},
	}

	// Grid frame for label + entry + browse button rows.
	f := frame.New(app, "form")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX("1c"))

	rows := []struct {
		label string
		op    string
	}{
		{"Select a file to open:", "open"},
		{"Select a file to save:", "save"},
	}

	for i, r := range rows {
		row := i
		op := r.op

		l := label.New(f, fmt.Sprintf("lab_%s", op),
			label.Text(r.label),
		)
		e := entry.New(f, fmt.Sprintf("ent_%s", op),
			entry.Width(20),
		)
		b := button.New(f, fmt.Sprintf("but_%s", op),
			button.Text("Browse ..."),
		)

		// Wire up the browse button command to open the appropriate dialog
		// and fill the entry with the selected path.
		ent := e
		b.Command = func() {
			var path string
			var ok bool
			if op == "open" {
				path, ok = dialog.OpenFile(app,
					dialog.FileTitle("Open File"),
					dialog.FileTypes(fileTypes...),
				)
			} else {
				path, ok = dialog.SaveFile(app,
					dialog.FileTitle("Save File"),
					dialog.FileTypes(fileTypes...),
				)
			}
			if ok {
				ent.SetText(path)
				ent.XView(len([]rune(path)))
				ent.Display()
			}
		}

		grid.Grid(l, grid.Row(row), grid.Column(0), grid.Sticky(grid.StickW), grid.PadY("3p"))
		grid.Grid(e, grid.Row(row), grid.Column(1), grid.Sticky(grid.EW), grid.PadX("3p"), grid.PadY("3p"))
		grid.Grid(b, grid.Row(row), grid.Column(2), grid.PadY("3p"))
	}

	grid.ColumnConfigure(f.Window(), 1, grid.SlotConfig{Weight: 1})

	// X11: "Use Motif Style Dialog" checkbutton (matches Tcl's x11 windowingsystem check).
	strictMotif := widget.NewVariable(false)
	strictCb := ttk.NewCheckbutton(app, "strict",
		ttk.CheckbuttonText("Use Motif Style Dialog"),
		ttk.CheckbuttonVar(strictMotif),
	)
	pack.Pack(strictCb, pack.SideOpt(pack.Top))
	_ = strictCb

	app.Run()
}
