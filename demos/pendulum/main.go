// Demo: Pendulum animation linked to simulation of physical system.
// Ported from Tk's pendulum.tcl demo.
package main

import (
	"fmt"
	"math"
	"os"
	"time"

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
	"github.com/takigo/takigo/widget/labelframe"
	"github.com/takigo/takigo/widget/panedwindow"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Pendulum Animation Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("pendulum"),
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
		label.Text("This demonstration shows how Tcl/Tk can be used to carry out animations that are linked to simulations of physical systems. In the left canvas is a graphical representation of the physical system itself, a simple pendulum, and in the right canvas is a graph of the phase space of the system, which is a plot of the angle (relative to the vertical) against the angular velocity. The pendulum bob may be repositioned by clicking and dragging anywhere on the left canvas."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Panedwindow with two labelframes (matches Tcl structure).
	pw := panedwindow.New(f, "p")
	pack.Pack(pw, pack.FillOpt(pack.FillBoth), pack.Expand(true))

	l1 := labelframe.New(pw, "l1", labelframe.Text("Pendulum Simulation"))
	l2 := labelframe.New(pw, "l2", labelframe.Text("Phase Space"))
	pw.Add(l1.Window(), 0)
	pw.Add(l2.Window(), 0)
	pw.SetStretch(l1.Window(), panedwindow.StretchAlways)
	pw.SetStretch(l2.Window(), panedwindow.StretchAlways)

	pt := screenunit.Distance.Float
	// tkScl is [tk scaling]: pixels per point.
	tkScl := screenunit.DPI() / 72
	rnd := func(v float64) float64 { return math.Round(v * tkScl) }
	length := rnd(111)
	xHome, yHome := rnd(120), rnd(18)
	rBob, rPivot := rnd(12), rnd(3)

	c := canvas.New(l1, "c", canvas.Width(screenunit.Pt(240)), canvas.Height(screenunit.Pt(150)),
		canvas.Background("white"), canvas.BorderWidthOpt(screenunit.Pt(1.5).Pixels()),
		canvas.ReliefOpt(option.ReliefSunken))
	c.CreateText(pt(screenunit.Pt(3)), pt(screenunit.Pt(3)), canvas.AnchorOpt(option.AnchorNW),
		canvas.TextOpt("Click to Adjust Bob Start Position"))
	c.CreateLine([]float64{0, 25, 320, 25}, canvas.Tags("plate"),
		canvas.OutlineColor("grey50"), canvas.OutlineWidth(screenunit.Pt(1.5).Pixels()))
	c.CreateOval(155, 20, 165, 30, canvas.Tags("pivot"), canvas.FillColor("grey50"), canvas.OutlineNone())
	c.CreateLine([]float64{1, 1, 1, 1}, canvas.Tags("rod"),
		canvas.OutlineColor("black"), canvas.OutlineWidth(screenunit.Pt(2.25).Pixels()))
	c.CreateOval(1, 1, 2, 2, canvas.Tags("bob"), canvas.FillColor("yellow"), canvas.OutlineColor("black"))
	pack.Pack(c, pack.FillOpt(pack.FillBoth), pack.Expand(true))

	k := canvas.New(l2, "k", canvas.Width(screenunit.Pt(240)), canvas.Height(screenunit.Pt(150)),
		canvas.Background("white"), canvas.BorderWidthOpt(screenunit.Pt(1.5).Pixels()),
		canvas.ReliefOpt(option.ReliefSunken))
	k.CreateLine([]float64{pt(screenunit.Pt(120)), pt(screenunit.Pt(150)), pt(screenunit.Pt(120)), 0}, canvas.OutlineColor("grey75"),
		canvas.Arrow(canvas.ArrowLast), canvas.Tags("y_axis"))
	k.CreateLine([]float64{0, pt(screenunit.Pt(75)), pt(screenunit.Pt(240)), pt(screenunit.Pt(75))}, canvas.OutlineColor("grey75"),
		canvas.Arrow(canvas.ArrowLast), canvas.Tags("x_axis"))
	for i := 90; i >= 0; i -= 10 {
		k.CreateLine([]float64{0, 0, 1, 1}, canvas.Smooth(true),
			canvas.Tags(fmt.Sprintf("graph%d", i)), canvas.OutlineColor(fmt.Sprintf("grey%d", i)))
	}
	k.CreateText(0, 0, canvas.AnchorOpt(option.AnchorNE), canvas.TextOpt("θ"), canvas.Tags("label_theta"))
	k.CreateText(0, 0, canvas.AnchorOpt(option.AnchorNE), canvas.TextOpt("δθ"), canvas.Tags("label_dtheta"))
	pack.Pack(k, pack.FillOpt(pack.FillBoth), pack.Expand(true))

	var points []float64
	theta, dTheta := 45.0, 0.0
	var psw, psh float64

	showPendulum := func(at bool, x, y float64) {
		if at && (x != xHome || y != yHome) {
			dTheta = 0
			x2, y2 := x-xHome, y-yHome
			length = math.Hypot(x2, y2)
			theta = math.Atan2(x2, y2) * 180 / math.Pi
		} else {
			angle := theta * math.Pi / 180
			x = xHome + length*math.Sin(angle)
			y = yHome + length*math.Cos(angle)
		}
		c.SetItemCoords("rod", []float64{xHome, yHome, x, y})
		c.SetItemCoords("bob", []float64{x - rBob, y - rBob, x + rBob, y + rBob})
	}
	showPendulum(false, 0, 0)

	showPhase := func() {
		scl := float64(screenunit.ScalingPct()) / 100
		points = append(points, theta+psw, -20*scl*dTheta+psh)
		if len(points) > 100 {
			points = points[len(points)-100:]
		}
		for i := 0; i < 100; i += 10 {
			// lrange $points end-($i-1) end-($i-12)
			n := len(points)
			lo, hi := max(0, n-i), min(n, n+12-i)
			if hi-lo < 4 {
				continue
			}
			tag := fmt.Sprintf("graph%d", i)
			k.SetItemCoords(tag, append([]float64(nil), points[lo:hi]...))
			k.Scale(tag, psw, psh, scl, scl)
		}
	}

	// <Configure> bindings: both canvases lay themselves out from their size.
	c.Win.OnConfigure(func() {
		w := float64(c.Win.Width)
		c.SetItemCoords("plate", []float64{0, pt(screenunit.Pt(18)), w, pt(screenunit.Pt(18))})
		xHome = float64(c.Win.Width / 2)
		c.SetItemCoords("pivot", []float64{xHome - rPivot, pt(screenunit.Pt(15)), xHome + rPivot, pt(screenunit.Pt(21))})
	})
	k.Win.OnConfigure(func() {
		w, h := float64(k.Win.Width), float64(k.Win.Height)
		psh, psw = float64(k.Win.Height/2), float64(k.Win.Width/2)
		k.SetItemCoords("x_axis", []float64{pt(screenunit.Pt(1.5)), psh, w - math.Round(1.5*tkScl), psh})
		k.SetItemCoords("y_axis", []float64{psw, h - math.Round(1.5*tkScl), psw, pt(screenunit.Pt(1.5))})
		k.SetItemCoords("label_dtheta", []float64{psw - math.Round(3*tkScl), pt(screenunit.Pt(4.5))})
		k.SetItemCoords("label_theta", []float64{w - math.Round(4.5*tkScl), psh + math.Round(3*tkScl)})
	})

	// recomputeAngle: the Tcl demo's two-stage estimate of one time step.
	recomputeAngle := func() {
		scaling := 3000.0 / length / length
		dd := func(t float64) float64 { return -math.Sin(t*math.Pi/180) * scaling }
		firstDD := dd(theta)
		midD := dTheta + firstDD
		mid := theta + (dTheta+midD)/2
		midDD := dd(mid)
		midD = dTheta + (firstDD+midDD)/2
		mid = theta + (dTheta+midD)/2
		midDD = dd(mid)
		lastD := midD + midDD
		last := mid + (midD+lastD)/2
		lastDD := dd(last)
		lastD = midD + (midDD+lastDD)/2
		last = mid + (midD+lastD)/2
		dTheta, theta = lastD, last
	}

	running := true
	gen := 0
	var repeat func(g int)
	repeat = func(g int) {
		if !running || g != gen {
			return
		}
		recomputeAngle()
		showPendulum(false, 0, 0)
		showPhase()
		app.After(15*time.Millisecond, func() { repeat(g) })
	}
	app.After(500*time.Millisecond, func() { repeat(gen) })

	disp := app.Dispatcher()
	disp.Bind(c.Win.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button == 1 {
			running = false
			gen++
			showPendulum(true, float64(ev.X), float64(ev.Y))
		}
	})
	disp.Bind(c.Win.PlatformID, event.MotionMask, func(ev *event.Event) {
		if !running && ev.State&platform.Button1Mask != 0 {
			showPendulum(true, float64(ev.X), float64(ev.Y))
		}
	})
	disp.Bind(c.Win.PlatformID, event.ButtonReleaseMask, func(ev *event.Event) {
		if ev.Button == 1 {
			showPendulum(true, float64(ev.X), float64(ev.Y))
			running = true
			gen++
			g := gen
			app.After(15*time.Millisecond, func() { repeat(g) })
		}
	})

	app.Run()
}
