// Demo: Horizontal scale controlling a canvas arrow.
// Ported from Tk's hscale.tcl demo.
package main

import (
	"fmt"
	"math"
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
	app, err := takigo.NewApp(takigo.Title("Horizontal Scale"), takigo.Size(500, 350))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	msg := label.New(root, "msg", app,
		label.Text("Drag the scale to change the arrow angle on the canvas."),
		label.Anchor(option.AnchorW),
		label.PadX(10),
		label.PadY(5),
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

	// Canvas for arrow display.
	c := canvas.New(root, "canvas", app,
		canvas.Background("white"),
		canvas.Width(300),
		canvas.Height(200),
	)
	pack.Pack(c.Window(), pack.SideOpt(pack.Top), pack.PadX(10), pack.PadY(5))

	// Draw initial arrow.
	cx, cy := 150.0, 100.0
	arrowLen := 80.0

	drawArrow := func(angleDeg float64) {
		c.Delete("arrow")
		rad := angleDeg * math.Pi / 180
		ex := cx + arrowLen*math.Cos(rad)
		ey := cy - arrowLen*math.Sin(rad)
		c.CreateLine([]float64{cx, cy, ex, ey},
			canvas.OutlineColor("#e74c3c"), canvas.OutlineWidth(3),
			canvas.Arrow(canvas.ArrowLast), canvas.ArrowShape(12, 16, 5),
			canvas.Tags("arrow"))
		c.CreateText(cx, cy+arrowLen+20,
			canvas.TextOpt(fmt.Sprintf("%.0f°", angleDeg)),
			canvas.FontOpt("Sans 10"), canvas.AnchorOpt(option.AnchorCenter),
			canvas.Tags("arrow"))
	}
	drawArrow(45)

	// Horizontal scale.
	sc := scale.New(root, "hscale", app,
		scale.OrientOpt(scale.Horizontal),
		scale.FromOpt(0),
		scale.ToOpt(360),
		scale.ValueOpt(45),
		scale.ShowValueOpt(true),
		scale.CommandOpt(func(v float64) {
			drawArrow(v)
		}),
	)
	pack.Pack(sc.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(30), pack.PadY(10))

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
