// Demo: File selection dialogs.
// Ported from Tk's filebox.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/dialog"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/geometry/place"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/frame"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("File Selection Dialogs"),
		takigo.Geometry("+300+300"),
		takigo.IconName("filebox"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// ttk::frame $w._bg: a themed background under everything else.
	bg := ttk.NewFrame(f, "_bg")
	place.Place(bg, place.X(0), place.Y(0), place.RelWidth(1), place.RelHeight(1))

	msg := ttk.NewLabel(f, "msg",
		ttk.LabelWrapLength(screenunit.In(4)),
		ttk.LabelJustify(option.JustifyLeft),
		ttk.LabelText("Enter a file name in the entry box or click on the \"Browse\" buttons to select a file name using the file selection dialog."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	fileTypes := []dialog.FileType{
		{Name: "Text files", Pattern: "*.txt"},
		{Name: "Tcl Scripts", Pattern: "*.tcl"},
		{Name: "C Source Files", Pattern: "*.c"},
		{Name: "All Source Files", Pattern: "*.go"},
		{Name: "Image Files", Pattern: "*.gif"},
		{Name: "All files", Pattern: "*"},
	}

	// Grid frame for label + entry + browse button rows.
	form := ttk.NewFrame(f, "f")
	pack.Pack(form, pack.FillOpt(pack.FillX), pack.PadX(screenunit.Cm(1)))

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

		l := ttk.NewLabel(form, fmt.Sprintf("lab_%s", op),
			ttk.LabelText(r.label),
		)
		e := ttk.NewEntry(form, fmt.Sprintf("ent_%s", op),
			ttk.EntryWidth(20),
		)
		b := ttk.NewButton(form, fmt.Sprintf("but_%s", op),
			ttk.ButtonText("Browse ..."),
		)

		// Wire up the browse button command to open the appropriate dialog
		// and fill the entry with the selected path.
		ent := e
		b.Command = func() {
			var path string
			var ok bool
			if op == "open" {
				path, ok = dialog.OpenFile(app,
					dialog.FileTypes(fileTypes...),
				)
			} else {
				path, ok = dialog.SaveFile(app,
					dialog.FileTypes(fileTypes...),
				)
			}
			if ok {
				_ = ent.Delete("0", "end")
				_ = ent.Insert("0", path)
				_ = ent.ICursor("end")
			}
		}

		grid.Grid(l, grid.Row(row), grid.Column(0), grid.Sticky(grid.StickW), grid.PadY(screenunit.Pt(3)))
		grid.Grid(e, grid.Row(row), grid.Column(1), grid.Sticky(grid.EW), grid.PadX(screenunit.Pt(3)), grid.PadY(screenunit.Pt(3)))
		grid.Grid(b, grid.Row(row), grid.Column(2), grid.Sticky(grid.StickW), grid.PadY(screenunit.Pt(3)))
	}

	grid.ColumnConfigure(form, 1, grid.Weight(1))

	// X11: "Use Motif Style Dialog" checkbutton (matches Tcl's x11 windowingsystem check).
	strictMotif := widget.NewVariable(false)
	strictCb := ttk.NewCheckbutton(f, "strict",
		ttk.CheckbuttonText("Use Motif Style Dialog"),
		ttk.CheckbuttonVar(strictMotif),
	)
	pack.Pack(strictCb)

	app.Run()
}
