// Demo: Message box varieties.
// Ported from Tk's msgbox.tcl demo.
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
	d := demohelper.Setup("Message Boxes", 400, 350,
		"Click any button to show a message dialog\nwith different types and button combinations.")
	root, app := d.Root, d.App

	statusLabel := label.New(root, "status", app,
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

	// Message box buttons.
	dialogs := []struct {
		text    string
		msgType dialog.MessageType
		buttons dialog.ButtonSet
		title   string
		message string
	}{
		{"Info (OK)", dialog.MsgInfo, dialog.BtnOK, "Information", "This is an informational message."},
		{"Warning (OK)", dialog.MsgWarning, dialog.BtnOK, "Warning", "This is a warning message!"},
		{"Error (OK)", dialog.MsgError, dialog.BtnOK, "Error", "An error has occurred."},
		{"Question (Yes/No)", dialog.MsgQuestion, dialog.BtnYesNo, "Confirm", "Do you want to proceed?"},
		{"Question (Yes/No/Cancel)", dialog.MsgQuestion, dialog.BtnYesNoCancel, "Save?", "Save changes before closing?"},
	}

	for _, d := range dialogs {
		dlg := d
		btn := button.New(root, "btn_"+dlg.text, app,
			button.Text(dlg.text),
			button.Command(func() {
				result := dialog.ShowMessage(app,
					dialog.MsgParent(root),
					dialog.MsgTitle(dlg.title),
					dialog.MsgMessage(dlg.message),
					dialog.MsgType(dlg.msgType),
					dialog.MsgButtons(dlg.buttons),
				)
				setStatus(fmt.Sprintf("Result: button %d", result))
			}),
			button.PadX(10), button.PadY(4),
		)
		pack.Pack(btn, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
			pack.PadX(20), pack.PadY(5))
		_ = btn
	}

	_ = statusLabel
	d.Run()
}
