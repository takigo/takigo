// Demo: 2D plot with draggable data points.
// Ported from Tk's plot.tcl demo.
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
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
		label.Text("This window displays a canvas widget containing a simple 2-dimensional plot. You can doctor the data by dragging any of the points with mouse button 1."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f, nil)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Canvas.
	c := canvas.New(f, "plot",
		canvas.Background("white"),
		canvas.Width(500),
		canvas.Height(350),
		canvas.ReliefOpt(option.ReliefRaised),
	)
	pack.Pack(c, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// Plot area dimensions.
	const (
		plotLeft   = 60
		plotRight  = 480
		plotTop    = 40
		plotBottom = 310
	)

	// Draw axes.
	c.CreateLine([]float64{plotLeft, plotBottom, plotRight, plotBottom},
		canvas.OutlineColor("black"), canvas.OutlineWidth(2))
	c.CreateLine([]float64{plotLeft, plotBottom, plotLeft, plotTop},
		canvas.OutlineColor("black"), canvas.OutlineWidth(2))

	// Title.
	c.CreateText(float64(plotLeft+plotRight)/2, float64(plotTop-20),
		canvas.TextOpt("A Simple Plot"), canvas.AnchorOpt(option.AnchorCenter),
		canvas.FontOpt("Helvetica 16"), canvas.FillColor("brown"))

	// Tick marks and labels — x: 0-100 every 10, y: 0-250 every 50.
	for i := 0; i <= 10; i++ {
		x := plotLeft + i*(plotRight-plotLeft)/10
		c.CreateLine([]float64{float64(x), float64(plotBottom), float64(x), float64(plotBottom + 5)},
			canvas.OutlineColor("black"))
		c.CreateText(float64(x), float64(plotBottom+12),
			canvas.TextOpt(fmt.Sprintf("%d", i*10)),
			canvas.AnchorOpt(option.AnchorN), canvas.FontOpt("Helvetica 16"))
	}
	for i := 0; i <= 5; i++ {
		y := plotBottom - i*(plotBottom-plotTop)/5
		c.CreateLine([]float64{float64(plotLeft - 5), float64(y), float64(plotLeft), float64(y)},
			canvas.OutlineColor("black"))
		c.CreateText(float64(plotLeft-10), float64(y),
			canvas.TextOpt(fmt.Sprintf("%d.0", i*50)),
			canvas.AnchorOpt(option.AnchorE), canvas.FontOpt("Helvetica 16"))
	}

	// Data points matching Tk's plot.tcl exactly.
	dataX := []float64{9, 15, 24.75, 24, 45.75, 56.25, 73.5}
	dataY := []float64{42, 70.5, 73.5, 90, 135, 120, 167.25}

	// Convert data to pixel coordinates (x: 0-100, y: 0-250).
	toPixelX := func(v float64) float64 {
		return plotLeft + v/100*(plotRight-plotLeft)
	}
	toPixelY := func(v float64) float64 {
		return plotBottom - v/250*(plotBottom-plotTop)
	}

	// Draw connecting line.
	lineCoords := make([]float64, len(dataX)*2)
	for i := range dataX {
		lineCoords[i*2] = toPixelX(dataX[i])
		lineCoords[i*2+1] = toPixelY(dataY[i])
	}
	lineID := c.CreateLine(lineCoords,
		canvas.OutlineColor("#3498db"), canvas.OutlineWidth(2),
		canvas.Tags("dataline"))
	lineIDStr := strconv.FormatInt(lineID, 10)

	// Draw data points as small filled circles.
	ptSize := 5.0
	pointIDs := make([]int64, len(dataX))
	for i := range dataX {
		px := toPixelX(dataX[i])
		py := toPixelY(dataY[i])
		pointIDs[i] = c.CreateOval(px-ptSize, py-ptSize, px+ptSize, py+ptSize,
			canvas.FillColor("SkyBlue2"), canvas.OutlineColor("black"), canvas.OutlineWidth(1),
			canvas.Tags("point"))
	}

	// updateLine rebuilds the connecting line coordinates from current point positions.
	updateLine := func() {
		coords := make([]float64, 0, len(pointIDs)*2)
		for _, pid := range pointIDs {
			pidStr := strconv.FormatInt(pid, 10)
			oc := c.ItemCoords(pidStr)
			if len(oc) >= 4 {
				// Oval coords are x1,y1,x2,y2; center is midpoint.
				cx := (oc[0] + oc[2]) / 2
				cy := (oc[1] + oc[3]) / 2
				coords = append(coords, cx, cy)
			}
		}
		c.SetItemCoords(lineIDStr, coords)
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

	// B1-Motion on point: drag the selected point and update the line.
	c.BindItem("point", event.MotionMask, func(ev *event.Event) {
		if ev.State&platform.Button1Mask == 0 {
			return
		}
		dx := float64(ev.X - lastX)
		dy := float64(ev.Y - lastY)
		c.Move("selected", dx, dy)
		lastX = ev.X
		lastY = ev.Y
		updateLine()
	})

	app.Run()
}
