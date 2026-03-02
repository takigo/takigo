// Demo: Editable arrowheads on canvas lines.
// Ported from Tk's arrow.tcl demo.
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
	app, err := takigo.NewApp(takigo.Title("Arrow Shapes"), takigo.Size(550, 400))
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
		label.Text("Various arrow shapes on canvas lines.\nArrows can appear at first, last, or both ends."),
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
	c := canvas.New(root, "arrows", app,
		canvas.Background("white"),
		canvas.Width(500),
		canvas.Height(300),
	)
	pack.Pack(c.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Row 1: Arrow at last end with different shapes.
	c.CreateText(250, 15, canvas.TextOpt("Arrow at Last End"),
		canvas.FontOpt("Sans Bold 11"), canvas.AnchorOpt(option.AnchorCenter))

	y := 40.0
	shapes := [][3]float64{
		{8, 10, 3},
		{12, 16, 5},
		{16, 20, 7},
		{20, 28, 10},
	}
	for _, s := range shapes {
		c.CreateLine([]float64{50, y, 250, y},
			canvas.OutlineColor("black"), canvas.OutlineWidth(2),
			canvas.Arrow(canvas.ArrowLast),
			canvas.ArrowShape(s[0], s[1], s[2]))
		c.CreateText(280, y,
			canvas.TextOpt(fmt.Sprintf("%.0f, %.0f, %.0f", s[0], s[1], s[2])),
			canvas.FontOpt("Sans 9"), canvas.AnchorOpt(option.AnchorW))
		y += 30
	}

	// Row 2: Arrow at both ends.
	c.CreateText(250, 175, canvas.TextOpt("Arrows at Both Ends"),
		canvas.FontOpt("Sans Bold 11"), canvas.AnchorOpt(option.AnchorCenter))

	c.CreateLine([]float64{50, 200, 450, 200},
		canvas.OutlineColor("#e74c3c"), canvas.OutlineWidth(3),
		canvas.Arrow(canvas.ArrowBoth), canvas.ArrowShape(12, 16, 5))

	c.CreateLine([]float64{50, 240, 450, 240},
		canvas.OutlineColor("#3498db"), canvas.OutlineWidth(4),
		canvas.Arrow(canvas.ArrowBoth), canvas.ArrowShape(16, 24, 8))

	// Polyline with arrows.
	c.CreateLine([]float64{50, 280, 150, 260, 250, 290, 350, 260, 450, 280},
		canvas.OutlineColor("#27ae60"), canvas.OutlineWidth(2),
		canvas.Arrow(canvas.ArrowBoth), canvas.ArrowShape(10, 14, 5),
		canvas.Smooth(true))

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
