// Demo: Buttons that change the window background color.
// Ported from Tk's button.tcl demo.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Button Demonstration"), takigo.Size(400, 350))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()
	defaultBg, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = defaultBg.Pixel

	// Description.
	msg := label.New(root, "msg", app,
		label.Text("Click any button to change the background color.\nThe color resets after 1.5 seconds."),
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

	// Color-changing function.
	changeColor := func(colorName string) {
		c, err := app.ColorCache().Get(colorName)
		if err != nil {
			return
		}
		root.BackgroundPixel = c.Pixel
		d := root.Display.XDisplay
		gc := root.GC
		d.SetForeground(gc, c.Pixel)
		d.FillRectangle(root.Drawable(), gc, 0, 0, uint(root.Width), uint(root.Height))
		d.Flush()

		// Reset after 1.5 seconds.
		app.After(1500*time.Millisecond, func() {
			root.BackgroundPixel = defaultBg.Pixel
			d.SetForeground(gc, defaultBg.Pixel)
			d.FillRectangle(root.Drawable(), gc, 0, 0, uint(root.Width), uint(root.Height))
			d.Flush()
			// Redraw all children.
			pack.ArrangeContainer(root)
		})
	}

	// Color buttons.
	colors := []struct {
		text  string
		color string
	}{
		{"Peach Puff", "#ffdab9"},
		{"Light Blue", "#add8e6"},
		{"Bisque", "#ffe4c4"},
		{"Light Green", "#90ee90"},
	}

	for _, c := range colors {
		colorVal := c.color
		btn := button.New(root, "btn_"+c.text, app,
			button.Text(c.text),
			button.Command(func() { changeColor(colorVal) }),
			button.PadX(10),
			button.PadY(6),
		)
		pack.Pack(btn.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
			pack.Expand(true), pack.PadX(20), pack.PadY(5))
		_ = btn
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
		d.SetForeground(gc, root.BackgroundPixel)
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
	app.MainLoop()
}
