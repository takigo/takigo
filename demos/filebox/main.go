// Demo: File open/save dialogs.
// Ported from Tk's filebox.tcl demo.
package main

import (
	"fmt"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/dialog"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	d := demohelper.Setup("File Dialogs", 400, 250,
		"Click a button to open a file dialog.\nThe selected path is shown in the status bar.")
	root, app := d.Root, d.App

	statusLabel := label.New(root, "status", app,
		label.Text("Selected: —"),
		label.Anchor(option.AnchorW),
		label.Background("#e8e8e8"),
		label.PadX(5), label.PadY(2),
	)
	pack.Pack(statusLabel, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	setStatus := func(s string) {
		statusLabel.Text = s
		statusLabel.Display()
	}

	fileTypes := []dialog.FileType{
		{Name: "Go files", Pattern: "*.go"},
		{Name: "Text files", Pattern: "*.txt"},
		{Name: "All files", Pattern: "*"},
	}

	openBtn := button.New(root, "open", app,
		button.Text("Open File..."),
		button.Command(func() {
			path, ok := dialog.OpenFile(app,
				dialog.FileParent(root),
				dialog.FileTitle("Open File"),
				dialog.FileTypes(fileTypes...),
			)
			if ok {
				setStatus(fmt.Sprintf("Open: %s", path))
			} else {
				setStatus("Open cancelled")
			}
		}),
		button.PadX(10), button.PadY(6),
	)
	pack.Pack(openBtn, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(30), pack.PadY(10))

	saveBtn := button.New(root, "save", app,
		button.Text("Save File..."),
		button.Command(func() {
			path, ok := dialog.SaveFile(app,
				dialog.FileParent(root),
				dialog.FileTitle("Save File"),
				dialog.FileTypes(fileTypes...),
			)
			if ok {
				setStatus(fmt.Sprintf("Save: %s", path))
			} else {
				setStatus("Save cancelled")
			}
		}),
		button.PadX(10), button.PadY(6),
	)
	pack.Pack(saveBtn, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(30), pack.PadY(5))

	_ = statusLabel
	_ = openBtn
	_ = saveBtn
	d.Run()
}
