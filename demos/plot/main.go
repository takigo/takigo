// Demo: 2D plot with draggable data points.
// Ported from Tk's plot.tcl demo.
package main

import (
	"fmt"
	"github.com/msorc/takigo"
	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"os"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Plot Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("Plot"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength("4i"),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("This window displays a canvas widget containing a simple 2-dimensional plot.  You can doctor the data by dragging any of the points with mouse button 1."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Canvas.
	c := canvas.New(f, "c",
		canvas.Background("white"),
		canvas.Width(450),  // 337.5p
		canvas.Height(300), // 225p
		canvas.ReliefOpt(option.ReliefRaised),
	)
	pack.Pack(c, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// Plot area dimensions (in pixels at 96 DPI, matching Tcl point values).
	const (
		plotLeft   = 100 // 75p
		plotRight  = 400 // 300p
		plotTop    = 50  // 37.5p
		plotBottom = 250 // 187.5p
	)

	// Draw axes.
	c.CreateLine([]float64{plotLeft, plotBottom, plotRight, plotBottom},
		canvas.OutlineColor("black"), canvas.OutlineWidth(2)) // 1.5p
	c.CreateLine([]float64{plotLeft, plotBottom, plotLeft, plotTop},
		canvas.OutlineColor("black"), canvas.OutlineWidth(2)) // 1.5p

	// Title (168.75p = 225px, 15p = 20px).
	c.CreateText(225, 20,
		canvas.TextOpt("A Simple Plot"), canvas.AnchorOpt(option.AnchorCenter),
		canvas.FontOpt("Helvetica 16"), canvas.FillColor("brown"))

	// Tick marks and labels — x: 0-100 every 10, y: 0-250 every 50.
	for i := 0; i <= 10; i++ {
		x := plotLeft + i*30 // 22.5p step = 30px
		c.CreateLine([]float64{float64(x), plotBottom, float64(x), plotBottom - 5},
			canvas.OutlineColor("black"), canvas.OutlineWidth(2)) // 1.5p
		c.CreateText(float64(x), plotBottom+4, // 190.5p = 254px
			canvas.TextOpt(fmt.Sprintf("%d", i*10)),
			canvas.AnchorOpt(option.AnchorN), canvas.FontOpt("Helvetica 16"))
	}
	for i := 0; i <= 5; i++ {
		y := plotBottom - i*(plotBottom-plotTop)/5
		c.CreateLine([]float64{plotLeft, float64(y), plotLeft + 5, float64(y)},
			canvas.OutlineColor("black"), canvas.OutlineWidth(2)) // 1.5p
		c.CreateText(plotLeft-4, float64(y), // 72p = 96px
			canvas.TextOpt(fmt.Sprintf("%d.0", i*50)),
			canvas.AnchorOpt(option.AnchorE), canvas.FontOpt("Helvetica 16"))
	}

	// Data points matching Tk's plot.tcl exactly.
	dataX := []float64{9, 15, 24.75, 24, 45.75, 56.25, 73.5}
	dataY := []float64{42, 70.5, 73.5, 90, 135, 120, 167.25}

	// Convert data to pixel coordinates.
	// Tcl: x_pt = 75 + 2.25*dataX, y_pt = 187.5 - 0.6*dataY (in points).
	// At 96 DPI (1p = 4/3 px): x_px = 100 + 3*dataX, y_px = 250 - 0.8*dataY.
	toPixelX := func(v float64) float64 {
		return plotLeft + 3*v
	}
	toPixelY := func(v float64) float64 {
		return plotBottom - 0.8*v
	}

	// Draw data points as small filled circles (4.5p radius = 6px).
	ptSize := 6.0
	for i := range dataX {
		px := toPixelX(dataX[i])
		py := toPixelY(dataY[i])
		c.CreateOval(px-ptSize, py-ptSize, px+ptSize, py+ptSize,
			canvas.FillColor("SkyBlue2"), canvas.OutlineColor("black"), canvas.OutlineWidth(1),
			canvas.Tags("point"))
	}

	// Drag state.
	var lastX, lastY int

	// Hover: change color on enter/leave.
	c.BindItem("point", event.EnterMask, func(ev *event.Event) {
		c.ItemConfigure("current", canvas.FillColor("red"))
	})
	c.BindItem("point", event.LeaveMask, func(ev *event.Event) {
		c.ItemConfigure("current", canvas.FillColor("SkyBlue2"))
	})

	// ButtonPress-1 on point: start drag.
	c.BindItem("point", event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button != 1 {
			return
		}
		c.DeleteTag("selected", "selected")
		c.AddTag("selected", "current")
		c.Raise("current")
		lastX = ev.X
		lastY = ev.Y
	})

	// ButtonRelease-1 on point: end drag.
	c.BindItem("point", event.ButtonReleaseMask, func(ev *event.Event) {
		if ev.Button != 1 {
			return
		}
		c.DeleteTag("selected", "selected")
	})

	// B1-Motion on canvas: drag the selected point (matches Tcl's bind $c <B1-Motion>).
	app.Dispatcher().Bind(c.Win.PlatformID, event.MotionMask, func(ev *event.Event) {
		if ev.State&platform.Button1Mask == 0 {
			return
		}
		dx := float64(ev.X - lastX)
		dy := float64(ev.Y - lastY)
		c.Move("selected", dx, dy)
		lastX = ev.X
		lastY = ev.Y
	})

	app.Run()
}
