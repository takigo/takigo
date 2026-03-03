// Demo: Modal dialog with global grab.
// Ported from Tk's dialog2.tcl demo.
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
	app := demohelper.Setup("Dialog with global grab", 400, 200,
		"This dialog box uses a global grab. You will be prevented from interacting with anything on your display until you invoke one of the buttons below. This is almost always a bad idea; don't use global grabs unless you're truly desperate.")

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
				dialog.MsgTitle("Dialog with global grab"),
				dialog.MsgMessage("This dialog box uses a global grab. If you are using an X11\n"+
					"window manager you will be prevented from interacting with\n"+
					"anything on your display until you invoke one of the buttons\n"+
					"below. This is almost always a bad idea; don't use global grabs\n"+
					"with X11 unless you're truly desperate."),
				dialog.MsgType(dialog.MsgWarning),
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
