// Demo: Modal dialog with local grab.
// Ported from Tk's dialog1.tcl demo.
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
	app := demohelper.Setup("Dialog with local grab", 400, 200,
		"This is a modal dialog box. It uses a \"local grab\" on the dialog box. The grab prevents any mouse or keyboard events from getting to any other windows in the application until you have answered the dialog by invoking one of the buttons below. However, you can still interact with other applications.")

	statusLabel := label.New(app, "status",
		label.Text("Result: —"),
		label.Anchor(option.AnchorW),
		label.Background("#e8e8e8"),
		label.PadX(5), label.PadY(2),
	)
	pack.Pack(statusLabel, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	btn := button.New(app, "btn",
		button.Text("Show Dialog"),
		button.Command(func() {
			result := dialog.ShowMessage(app,
				dialog.MsgTitle("Dialog with local grab"),
				dialog.MsgMessage("This is a modal dialog box. It uses a \"local grab\" on the dialog\n"+
					"box. The grab prevents any mouse or keyboard events from getting\n"+
					"to any other windows in the application until you have answered\n"+
					"the dialog by invoking one of the buttons below. However, you can\n"+
					"still interact with other applications."),
				dialog.MsgType(dialog.MsgInfo),
				dialog.MsgButtons(dialog.BtnOKCancel),
			)
			labels := []string{"OK", "Cancel"}
			if int(result) >= 0 && int(result) < len(labels) {
				statusLabel.Text = fmt.Sprintf("You pressed %s", labels[result])
				statusLabel.Display()
			}
		}),
		button.PadX(10), button.PadY(6),
	)
	pack.Pack(btn, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(30), pack.PadY(10))

	_ = btn
	_ = statusLabel
	app.Run()
}
