// Demo: Canvas text items at various positions and anchors.
// Ported from Tk's ctext.tcl demo (simplified — no inline editing).
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Canvas Text Demo"), takigo.Size(550, 400))
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
		label.Text("Canvas text items rendered at various positions\nwith different anchors, fonts, and colors."),
		label.Anchor(option.AnchorW),
		label.PadX(10),
		label.PadY(5),
	)
	pack.Pack(msg.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// Dismiss button.
	btnFrame := frame.New(root, "btnframe", app)
	pack.Pack(btnFrame.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))

	dismissBtn := button.New(btnFrame.Window(), "dismiss", app,
		button.Text("Dismiss"),
		button.Command(func() { app.Quit() }),
		button.PadX(10),
		button.PadY(4),
	)
	pack.Pack(dismissBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(10))

	// Canvas.
	c := canvas.New(root, "canvas", app,
		canvas.Background("white"),
		canvas.Width(500),
		canvas.Height(300),
	)
	pack.Pack(c.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Title text.
	c.CreateText(250, 30,
		canvas.TextOpt("Canvas Text Demo"),
		canvas.FontOpt("Sans Bold 18"),
		canvas.TextColor("navy"),
		canvas.AnchorOpt(option.AnchorCenter),
	)

	// Multi-line centered text.
	c.CreateText(250, 80,
		canvas.TextOpt("This text is rendered directly on the canvas.\nMultiple lines are supported via newlines."),
		canvas.FontOpt("Sans 12"),
		canvas.TextColor("black"),
		canvas.AnchorOpt(option.AnchorCenter),
	)

	// Left-anchored text.
	c.CreateText(50, 150,
		canvas.TextOpt("Left-anchored text\nat position (50, 150)"),
		canvas.FontOpt("Sans Italic 11"),
		canvas.TextColor("darkred"),
		canvas.AnchorOpt(option.AnchorNW),
	)

	// Right-anchored text.
	c.CreateText(450, 150,
		canvas.TextOpt("Right-anchored text\nat position (450, 150)"),
		canvas.FontOpt("Sans 11"),
		canvas.TextColor("darkblue"),
		canvas.AnchorOpt(option.AnchorNE),
	)

	// Center paragraph.
	c.CreateText(250, 230,
		canvas.TextOpt("Center-anchored text with a longer paragraph.\nThis demonstrates how canvas text wraps and displays\nmultiple lines of content."),
		canvas.FontOpt("Serif 11"),
		canvas.TextColor("darkgreen"),
		canvas.AnchorOpt(option.AnchorCenter),
	)

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
	app.MainLoop()
}
