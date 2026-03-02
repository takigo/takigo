// Demo: Modal dialog examples.
// Ported from Tk's dialog1.tcl + dialog2.tcl demo.
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
	d := demohelper.Setup("Modal Dialogs", 400, 250, "Modal dialog examples. Each dialog blocks\ninteraction with the main window until dismissed.")
	app := d.App

	statusLabel := label.New(app, "status",
		label.Text("Result: —"),
		label.Anchor(option.AnchorW),
		label.Background("#e8e8e8"),
		label.PadX(5), label.PadY(2),
	)
	pack.Pack(statusLabel, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	setStatus := func(s string) {
		statusLabel.Text = s
		statusLabel.Display()
	}

	// Dialog 1: Simple OK message.
	btn1 := button.New(app, "btn1",
		button.Text("Simple Message"),
		button.Command(func() {
			result := dialog.ShowMessage(app,
				dialog.MsgTitle("Greeting"),
				dialog.MsgMessage("Hello! This is a modal dialog."),
				dialog.MsgType(dialog.MsgInfo),
				dialog.MsgButtons(dialog.BtnOK),
			)
			setStatus(fmt.Sprintf("Simple dialog: button %d", result))
		}),
		button.PadX(10), button.PadY(6),
	)
	pack.Pack(btn1, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(30), pack.PadY(5))

	// Dialog 2: Yes/No/Cancel question.
	btn2 := button.New(app, "btn2",
		button.Text("Save Changes?"),
		button.Command(func() {
			result := dialog.ShowMessage(app,
				dialog.MsgTitle("Save"),
				dialog.MsgMessage("Do you want to save your changes?"),
				dialog.MsgDetail("If you don't save, your changes will be lost."),
				dialog.MsgType(dialog.MsgQuestion),
				dialog.MsgButtons(dialog.BtnYesNoCancel),
			)
			labels := []string{"Yes", "No", "Cancel"}
			if int(result) >= 0 && int(result) < len(labels) {
				setStatus(fmt.Sprintf("Save dialog: %s", labels[result]))
			}
		}),
		button.PadX(10), button.PadY(6),
	)
	pack.Pack(btn2, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(30), pack.PadY(5))

	_ = statusLabel
	_ = btn1
	_ = btn2
	d.Run()
}
