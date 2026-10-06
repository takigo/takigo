// Demo: 2D plot with draggable data points.
// Ported from Tk's plot.tcl demo.
package main

import (
	"fmt"
	"math"
	"os"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/canvas"
	"github.com/takigo/takigo/demos/demohelper"
	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/widget/frame"
	"github.com/takigo/takigo/widget/label"
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
		label.WrapLength(screenunit.In(4)),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("This window displays a canvas widget containing a simple 2-dimensional plot.  You can doctor the data by dragging any of the points with mouse button 1."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Canvas.
	c := canvas.New(f, "c",
		canvas.Width(screenunit.Pt(337.5)),
		canvas.Height(screenunit.Pt(225)),
		canvas.ReliefOpt(option.ReliefRaised),
	)
	pack.Pack(c, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// Coordinates are in points, as in plot.tcl.
	pt := func(v float64) float64 { return screenunit.Pt(v).Float() }
	lineW := screenunit.Pt(1.5).Pixels()
	const plotFont = "Helvetica 16"

	c.CreateLine([]float64{pt(75), pt(187.5), pt(300), pt(187.5)}, canvas.OutlineWidth(lineW))
	c.CreateLine([]float64{pt(75), pt(187.5), pt(75), pt(37.5)}, canvas.OutlineWidth(lineW))
	c.CreateText(pt(168.75), pt(15),
		canvas.TextOpt("A Simple Plot"), canvas.FontOpt(plotFont), canvas.FillColor("brown"))

	for i := 0; i <= 10; i++ {
		x := 75 + float64(i)*22.5
		c.CreateLine([]float64{pt(x), pt(187.5), pt(x), pt(183.75)}, canvas.OutlineWidth(lineW))
		c.CreateText(pt(x), pt(190.5),
			canvas.TextOpt(fmt.Sprintf("%d", 10*i)),
			canvas.AnchorOpt(option.AnchorN), canvas.FontOpt(plotFont))
	}
	for i := 0; i <= 5; i++ {
		y := 187.5 - float64(i)*30
		c.CreateLine([]float64{pt(75), pt(y), pt(78.75), pt(y)}, canvas.OutlineWidth(lineW))
		c.CreateText(pt(72), pt(y),
			canvas.TextOpt(fmt.Sprintf("%d.0", i*50)),
			canvas.AnchorOpt(option.AnchorE), canvas.FontOpt(plotFont))
	}

	points := [][2]float64{{9, 42}, {15, 70.5}, {24.75, 73.5}, {24, 90}, {45.75, 135}, {56.25, 120}, {73.5, 167.25}}
	for _, p := range points {
		x := 75 + 2.25*p[0]
		dy := 3 * p[1] / 5
		if p[1] == math.Trunc(p[1]) {
			dy = math.Trunc(dy) // Tcl: (3*42)/5 is integer division
		}
		y := 187.5 - dy
		c.CreateOval(pt(x-4.5), pt(y-4.5), pt(x+4.5), pt(y+4.5),
			canvas.OutlineWidth(screenunit.Pt(0.75).Pixels()), canvas.OutlineColor("black"),
			canvas.FillColor("SkyBlue2"), canvas.Tags("point"))
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
