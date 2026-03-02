// Demo: Pendulum physics simulation on canvas.
// Ported from Tk's pendulum.tcl demo.
package main

import (
	"fmt"
	"math"
	"os"
	"time"

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
	app, err := takigo.NewApp(takigo.Title("Pendulum"), takigo.Size(400, 450))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	msg := label.New(root, "msg", app,
		label.Text("A simple pendulum physics simulation."),
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

	c := canvas.New(root, "pendulum", app,
		canvas.Background("#1a1a2e"),
		canvas.Width(350),
		canvas.Height(350),
	)
	pack.Pack(c.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Pendulum parameters.
	pivotX, pivotY := 175.0, 50.0
	length := 200.0
	bobRadius := 15.0
	theta := math.Pi / 4 // Initial angle.
	omega := 0.0         // Angular velocity.
	gravity := 9.8
	dt := 0.05
	damping := 0.998

	// Pivot point.
	c.CreateOval(pivotX-4, pivotY-4, pivotX+4, pivotY+4,
		canvas.FillColor("#aaaaaa"), canvas.OutlineColor("#888888"))

	// Animation loop.
	var animate func()
	animate = func() {
		// Physics update (simple pendulum).
		alpha := -gravity / length * math.Sin(theta)
		omega += alpha * dt
		omega *= damping
		theta += omega * dt

		bobX := pivotX + length*math.Sin(theta)
		bobY := pivotY + length*math.Cos(theta)

		// Redraw.
		c.Delete("rod")
		c.Delete("bob")
		c.Delete("trail")

		// Trail shadow.
		for i := 1; i <= 5; i++ {
			t := theta - omega*dt*float64(i)
			tx := pivotX + length*math.Sin(t)
			ty := pivotY + length*math.Cos(t)
			alpha := 50 + i*20
			_ = alpha
			c.CreateOval(tx-bobRadius*0.5, ty-bobRadius*0.5, tx+bobRadius*0.5, ty+bobRadius*0.5,
				canvas.FillColor("#333355"), canvas.Tags("trail"))
		}

		// Rod.
		c.CreateLine([]float64{pivotX, pivotY, bobX, bobY},
			canvas.OutlineColor("#cccccc"), canvas.OutlineWidth(2), canvas.Tags("rod"))

		// Bob.
		c.CreateOval(bobX-bobRadius, bobY-bobRadius, bobX+bobRadius, bobY+bobRadius,
			canvas.FillColor("#e74c3c"), canvas.OutlineColor("#c0392b"), canvas.OutlineWidth(2),
			canvas.Tags("bob"))

		app.After(20*time.Millisecond, animate)
	}
	app.After(20*time.Millisecond, animate)

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
