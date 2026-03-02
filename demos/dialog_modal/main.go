// Demo: Modal dialog examples.
// Ported from Tk's dialog1.tcl + dialog2.tcl demo.
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
	app, err := takigo.NewApp(takigo.Title("Modal Dialogs"), takigo.Size(400, 250))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	msg := label.New(root, "msg", app,
		label.Text("Modal dialog examples. Each dialog blocks\ninteraction with the main window until dismissed."),
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

	// Dialog 1: Simple OK message.
	btn1 := button.New(root, "btn1", app,
		button.Text("Simple Message"),
		button.Command(func() {
			result := dialog.ShowMessage(app,
				dialog.MsgParent(root),
				dialog.MsgTitle("Greeting"),
				dialog.MsgMessage("Hello! This is a modal dialog."),
				dialog.MsgType(dialog.MsgInfo),
				dialog.MsgButtons(dialog.BtnOK),
			)
			setStatus(fmt.Sprintf("Simple dialog: button %d", result))
		}),
		button.PadX(10), button.PadY(6),
	)
	pack.Pack(btn1.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(30), pack.PadY(5))

	// Dialog 2: Yes/No/Cancel question.
	btn2 := button.New(root, "btn2", app,
		button.Text("Save Changes?"),
		button.Command(func() {
			result := dialog.ShowMessage(app,
				dialog.MsgParent(root),
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
	pack.Pack(btn2.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(30), pack.PadY(5))

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
	_ = btn1
	_ = btn2
	app.MainLoop()
}
