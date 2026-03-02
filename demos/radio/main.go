// Demo: Radiobutton groups for selecting point size and color.
// Ported from Tk's radio.tcl demo.
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
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/radiobutton"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Radiobutton Demonstration"), takigo.Size(500, 400))
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
		label.Text("Two groups of radiobuttons are displayed below. Click on\na button to select it. The current selection is shown at\nthe bottom."),
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
		label.Text("Point size: 10, Color: red"),
		label.Anchor(option.AnchorW),
		label.PadX(10),
	)
	pack.Pack(statusLabel.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))

	// Variables.
	sizeVar := widget.NewVariable("10")
	colorVar := widget.NewVariable("red")

	updateStatus := func() {
		statusLabel.Text = fmt.Sprintf("Point size: %s, Color: %s", sizeVar.Get(), colorVar.Get())
		statusLabel.Display()
	}

	// Container for two groups side by side.
	groupFrame := frame.New(root, "groups", app)
	pack.Pack(groupFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Point size group.
	sizeFrame := frame.New(groupFrame.Window(), "sizes", app,
		frame.BorderWidth(2), frame.Relief(option.ReliefGroove))
	pack.Pack(sizeFrame.Window(), pack.SideOpt(pack.Left), pack.PadX(10), pack.PadY(5))

	sizeTitle := label.New(sizeFrame.Window(), "sizetitle", app,
		label.Text("Point Size"),
		label.PadX(5),
	)
	pack.Pack(sizeTitle.Window(), pack.SideOpt(pack.Top), pack.PadY(5))

	sizes := []string{"10", "12", "14", "18", "24"}
	for _, s := range sizes {
		rb := radiobutton.New(sizeFrame.Window(), "size_"+s, app,
			radiobutton.Text(s+" point"),
			radiobutton.Value(s),
			radiobutton.Var(sizeVar),
			radiobutton.Command(updateStatus),
		)
		pack.Pack(rb.Window(), pack.SideOpt(pack.Top), pack.PadY(2), pack.Anchor(option.AnchorW), pack.PadX(10))
		_ = rb
	}

	// Color group.
	colorFrame := frame.New(groupFrame.Window(), "colors", app,
		frame.BorderWidth(2), frame.Relief(option.ReliefGroove))
	pack.Pack(colorFrame.Window(), pack.SideOpt(pack.Left), pack.PadX(10), pack.PadY(5))

	colorTitle := label.New(colorFrame.Window(), "colortitle", app,
		label.Text("Color"),
		label.PadX(5),
	)
	pack.Pack(colorTitle.Window(), pack.SideOpt(pack.Top), pack.PadY(5))

	colors := []string{"Red", "Orange", "Yellow", "Green", "Blue"}
	for _, c := range colors {
		val := c
		rb := radiobutton.New(colorFrame.Window(), "color_"+c, app,
			radiobutton.Text(c),
			radiobutton.Value(c),
			radiobutton.Var(colorVar),
			radiobutton.Command(updateStatus),
		)
		pack.Pack(rb.Window(), pack.SideOpt(pack.Top), pack.PadY(2), pack.Anchor(option.AnchorW), pack.PadX(10))
		_ = rb
		_ = val
	}

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
	_ = sizeTitle
	_ = colorTitle
	app.MainLoop()
}
