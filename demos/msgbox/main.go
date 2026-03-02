// Demo: Message box varieties.
// Ported from Tk's msgbox.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/dialog"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Message Boxes"), takigo.Size(400, 350))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	msg := label.New(root, "msg", app,
		label.Text("Click any button to show a message dialog\nwith different types and button combinations."),
		label.Anchor(option.AnchorW),
		label.PadX(10), label.PadY(5),
	)
	pack.Pack(msg.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	btnFrame := frame.New(root, "btnframe", app)
	pack.Pack(btnFrame.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))
	dismissBtn := button.New(btnFrame.Window(), "dismiss", app,
		button.Text("Dismiss"), button.Command(func() { app.Quit() }),
		button.PadX(10), button.PadY(4),
	)
	pack.Pack(dismissBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(10))

	statusLabel := label.New(root, "status", app,
		label.Text("Result: —"),
		label.Anchor(option.AnchorW),
		label.Background("#e8e8e8"),
		label.PadX(5), label.PadY(2),
	)
	pack.Pack(statusLabel.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

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
		pack.Pack(btn.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
			pack.PadX(20), pack.PadY(5))
		_ = btn
	}

	// Root events.
	app.Dispatcher().Bind(root.XWindow, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			root.Width = ev.ConfigWidth
			root.Height = ev.ConfigHeight
			pack.ArrangeContainer(root)
		}
	})
	app.Dispatcher().Bind(root.XWindow, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		d := root.Display.XDisplay
		d.SetForeground(root.GC, bgColor.Pixel)
		d.FillRectangle(root.Drawable(), root.GC, 0, 0, uint(root.Width), uint(root.Height))
		d.Flush()
	})
	app.Dispatcher().BindGlobal(event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == xlib.XK_Escape {
			app.Quit()
		}
	})

	_ = msg
	_ = dismissBtn
	_ = statusLabel
	app.MainLoop()
}
