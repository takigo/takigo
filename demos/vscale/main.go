// Demo: Vertical scale controlling a canvas bar height.
// Ported from Tk's vscale.tcl demo.
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
	"github.com/msorc/takigo/widget/scale"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Vertical Scale"), takigo.Size(400, 400))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	msg := label.New(root, "msg", app,
		label.Text("Drag the vertical scale to change the bar height."),
		label.Anchor(option.AnchorW),
		label.PadX(10), label.PadY(5),
	)
	pack.Pack(msg.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	btnFrame := frame.New(root, "btnframe", app)
	pack.Pack(btnFrame.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))

	dismissBtn := button.New(btnFrame.Window(), "dismiss", app,
		button.Text("Dismiss"),
		button.Command(func() { app.Quit() }),
		button.PadX(10), button.PadY(4),
	)
	pack.Pack(dismissBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(10))

	// Middle frame: scale on left, canvas on right.
	midFrame := frame.New(root, "midframe", app)
	pack.Pack(midFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Canvas.
	c := canvas.New(midFrame.Window(), "canvas", app,
		canvas.Background("white"),
		canvas.Width(250),
		canvas.Height(250),
	)

	barBottom := 240.0

	drawBar := func(v float64) {
		c.Delete("bar")
		c.Delete("label")
		h := v / 100 * 200
		c.CreateRectangle(80, barBottom-h, 180, barBottom,
			canvas.FillColor("#3498db"), canvas.OutlineColor("#2980b9"), canvas.OutlineWidth(2),
			canvas.Tags("bar"))
		c.CreateText(130, barBottom-h-10,
			canvas.TextOpt(fmt.Sprintf("%.0f%%", v)),
			canvas.FontOpt("Sans Bold 11"), canvas.AnchorOpt(option.AnchorCenter),
			canvas.Tags("label"))
	}
	drawBar(50)

	// Vertical scale.
	sc := scale.New(midFrame.Window(), "vscale", app,
		scale.OrientOpt(scale.Vertical),
		scale.FromOpt(100),
		scale.ToOpt(0),
		scale.ValueOpt(50),
		scale.ShowValueOpt(true),
		scale.CommandOpt(func(v float64) { drawBar(v) }),
	)

	pack.Pack(sc.Window(), pack.SideOpt(pack.Left), pack.FillOpt(pack.FillY), pack.PadX(10))
	pack.Pack(c.Window(), pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))

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
	_ = sc
	app.MainLoop()
}
