// Demo: Horizontal paned window with colored panes.
// Ported from Tk's paned1.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/panedwindow"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Horizontal Panes"), takigo.Size(500, 300))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	msg := label.New(root, "msg", app,
		label.Text("A horizontal paned window. Drag the sash to resize panes."),
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

	// Paned window.
	pw := panedwindow.New(root, "panes", app,
		panedwindow.OrientOpt(panedwindow.Horizontal),
	)
	pack.Pack(pw.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Three colored panes.
	colors := []struct {
		name, color, text string
	}{
		{"left", "#e74c3c", "Left\n(Red)"},
		{"center", "#3498db", "Center\n(Blue)"},
		{"right", "#2ecc71", "Right\n(Green)"},
	}

	for _, c := range colors {
		f := frame.New(pw.Window(), c.name, app, frame.Background(c.color))
		l := label.New(f.Window(), c.name+"_label", app,
			label.Text(c.text),
			label.Background(c.color),
			label.Foreground("white"),
		)
		pack.Pack(l.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
			pack.Expand(true), pack.PadX(10), pack.PadY(10))
		pw.Add(f.Window(), 80)
		_ = l
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
	app.MainLoop()
}
