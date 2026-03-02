// Demo: Checkbuttons with on/off state and status label.
// Ported from Tk's check.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/checkbutton"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Checkbutton Demonstration"), takigo.Size(400, 350))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	// Description.
	msg := label.New(root, "msg", app,
		label.Text("Three checkbuttons are displayed below. Click on a\nbutton to toggle its state. The current state of each\nbutton is displayed at the bottom."),
		label.Anchor(option.AnchorW),
		label.PadX(10),
		label.PadY(5),
	)
	pack.Pack(msg.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// Dismiss button at bottom.
	btnFrame := frame.New(root, "btnframe", app)
	pack.Pack(btnFrame.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))

	dismissBtn := button.New(btnFrame.Window(), "dismiss", app,
		button.Text("Dismiss"),
		button.Command(func() { app.Quit() }),
		button.PadX(10),
		button.PadY(4),
	)
	pack.Pack(dismissBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(10))

	// Status label.
	statusLabel := label.New(root, "status", app,
		label.Text("Status: all unchecked"),
		label.Anchor(option.AnchorW),
		label.PadX(10),
	)
	pack.Pack(statusLabel.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))

	// Variables.
	wipers := widget.NewVariable(false)
	brakes := widget.NewVariable(false)
	sober := widget.NewVariable(false)

	updateStatus := func() {
		status := "Status:"
		if wipers.Get() {
			status += " Wipers-on"
		}
		if brakes.Get() {
			status += " Brakes-on"
		}
		if sober.Get() {
			status += " Sober"
		}
		if !wipers.Get() && !brakes.Get() && !sober.Get() {
			status += " all unchecked"
		}
		statusLabel.Text = status
		statusLabel.Display()
	}

	// Checkbuttons.
	cb1 := checkbutton.New(root, "wipers", app,
		checkbutton.Text("Safety Check: Wipers OK"),
		checkbutton.Var(wipers),
		checkbutton.Command(updateStatus),
	)
	pack.Pack(cb1.Window(), pack.SideOpt(pack.Top), pack.PadY(4), pack.Anchor(option.AnchorW), pack.PadX(20))

	cb2 := checkbutton.New(root, "brakes", app,
		checkbutton.Text("Safety Check: Brakes OK"),
		checkbutton.Var(brakes),
		checkbutton.Command(updateStatus),
	)
	pack.Pack(cb2.Window(), pack.SideOpt(pack.Top), pack.PadY(4), pack.Anchor(option.AnchorW), pack.PadX(20))

	cb3 := checkbutton.New(root, "sober", app,
		checkbutton.Text("Safety Check: Driver Sober"),
		checkbutton.Var(sober),
		checkbutton.Command(updateStatus),
	)
	pack.Pack(cb3.Window(), pack.SideOpt(pack.Top), pack.PadY(4), pack.Anchor(option.AnchorW), pack.PadX(20))

	// Root event handlers.
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
		gc := root.GC
		d.SetForeground(gc, bgColor.Pixel)
		d.FillRectangle(root.Drawable(), gc, 0, 0, uint(root.Width), uint(root.Height))
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
	_ = cb1
	_ = cb2
	_ = cb3
	app.MainLoop()
}
