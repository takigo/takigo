// Demo: Editable arrowheads on canvas lines.
// Ported from Tk's arrow.tcl demo.
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
	app, err := takigo.NewApp(takigo.Title("Arrowhead Editor Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("arrow"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength(screenunit.In(5)),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("This widget allows you to experiment with different widths and arrowhead shapes for lines in canvases.  To change the line width or the shape of the arrowhead, drag any of the three boxes attached to the oversized arrow.  The arrows on the right give examples at normal scale.  The text at the bottom shows the configuration options as you'd enter them for a canvas line item."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	c := canvas.New(f, "c",
		canvas.Width(screenunit.Pt(375)),
		canvas.Height(screenunit.Pt(262.5)),
		canvas.ReliefOpt(option.ReliefSunken),
		canvas.BorderWidthOpt(2),
	)
	pack.Pack(c, pack.Expand(true), pack.FillOpt(pack.FillBoth))

	// scl scales an integer by ::tk::scalingPct, as the Tcl demo's proc does.
	scl := func(n int) float64 {
		return float64(int(math.Round(float64(n) * float64(screenunit.ScalingPct()) / 100)))
	}
	pt := func(d screenunit.Distance) float64 { return float64(d.Pixels()) }

	a, b, cc, w := scl(8), scl(10), scl(3), scl(2)
	x1, x2, y := scl(40), scl(350), scl(150)
	smallA, smallB, smallC := pt(screenunit.Pt(3.75)), pt(screenunit.Pt(3.75)), pt(screenunit.Pt(1.5))
	boxW := screenunit.Pt(0.75).Pixels()

	activeDragger := 0 // 0=none, 1=box1, 2=box2, 3=box3

	// arrowSetup regenerates everything, as the Tcl proc does.
	redraw := func() {
		c.Delete("all")

		c.CreateLine([]float64{x1, y, x2, y},
			canvas.OutlineColor("LightSeaGreen"),
			canvas.OutlineWidth(int(10*w)),
			canvas.Arrow(canvas.ArrowLast),
			canvas.ArrowShape(10*a, 10*b, 10*cc))
		xtip := x2 - 10*b
		deltaY := 10*cc + 5*w
		c.CreateLine([]float64{x2, y, xtip, y + deltaY, x2 - 10*a, y, xtip, y - deltaY, x2, y},
			canvas.OutlineWidth(screenunit.Pt(1.5).Pixels()),
			canvas.CapStyleOpt(platform.CapRound), canvas.JoinStyleOpt(platform.JoinRound))

		s5 := scl(5)
		box := func(x0, y0, x1, y1 float64, tag string) {
			c.CreateRectangle(x0, y0, x1, y1, canvas.OutlineWidth(boxW), canvas.Tags(tag, "box"))
		}
		box(x2-10*a-s5, y-s5, x2-10*a+s5, y+s5, "box1")
		box(xtip-s5, y-deltaY-s5, xtip+s5, y-deltaY+s5, "box2")
		box(x1-s5, y-5*w-s5, x1+s5, y-5*w+s5, "box3")

		s10, s15, s25, s50, s75, s125 := scl(10), scl(15), scl(25), scl(50), scl(75), scl(125)
		c.CreateLine([]float64{x2 + s50, 0, x2 + s50, pt(screenunit.Pt(750))}, canvas.OutlineWidth(screenunit.Pt(1.5).Pixels()))
		tmp := x2 + scl(100)
		example := func(pts []float64) {
			c.CreateLine(pts, canvas.OutlineWidth(int(w)), canvas.Arrow(canvas.ArrowBoth), canvas.ArrowShape(a, b, cc))
		}
		example([]float64{tmp, y - s125, tmp, y - s75})
		example([]float64{tmp - s25, y, tmp + s25, y})
		example([]float64{tmp - s25, y + s75, tmp + s25, y + s125})

		dim := func(pts []float64) {
			c.CreateLine(pts, canvas.Arrow(canvas.ArrowBoth), canvas.ArrowShape(smallA, smallB, smallC))
		}
		num := func(v float64) string { return fmt.Sprintf("%.0f", v) }
		tmp = x2 + s10
		dim([]float64{tmp, y - 5*w, tmp, y - deltaY})
		c.CreateText(x2+s15, y-deltaY+5*cc, canvas.TextOpt(num(cc)), canvas.AnchorOpt(option.AnchorW))
		tmp = x1 - s10
		dim([]float64{tmp, y - 5*w, tmp, y + 5*w})
		c.CreateText(x1-s15, y, canvas.TextOpt(num(w)), canvas.AnchorOpt(option.AnchorE))
		tmp = y + 5*w + 10*cc + s10
		dim([]float64{x2 - 10*a, tmp, x2, tmp})
		c.CreateText(x2-5*a, tmp+s5, canvas.TextOpt(num(a)), canvas.AnchorOpt(option.AnchorN))
		tmp += s25
		dim([]float64{x2 - 10*b, tmp, x2, tmp})
		c.CreateText(x2-5*b, tmp+s5, canvas.TextOpt(num(b)), canvas.AnchorOpt(option.AnchorN))

		c.CreateText(x1, pt(screenunit.Pt(232.5)), canvas.TextOpt("-width  "+num(w)),
			canvas.AnchorOpt(option.AnchorW), canvas.FontOpt("Helvetica 18"))
		c.CreateText(x1, pt(screenunit.Pt(247.5)),
			canvas.TextOpt(fmt.Sprintf("-arrowshape  {%s  %s  %s}", num(a), num(b), num(cc))),
			canvas.AnchorOpt(option.AnchorW), canvas.FontOpt("Helvetica 18"))
	}

	redraw()

	// Box hover: fill red on Enter, clear fill on Leave.
	c.BindItem("box", event.EnterMask, func(ev *event.Event) {
		c.ItemConfigure("current", canvas.FillColor("red"), canvas.OutlineWidth(boxW))
	})
	c.BindItem("box", event.LeaveMask, func(ev *event.Event) {
		c.ItemConfigure("current", canvas.FillNone())
	})

	// Box press: set which dragger is active.
	c.BindItem("box1", event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button == 1 {
			activeDragger = 1
		}
	})
	c.BindItem("box2", event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button == 1 {
			activeDragger = 2
		}
	})
	c.BindItem("box3", event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button == 1 {
			activeDragger = 3
		}
	})

	// Canvas window motion: drag the active box.
	app.Dispatcher().Bind(c.Win.PlatformID, event.MotionMask, func(ev *event.Event) {
		if ev.State&platform.Button1Mask == 0 || activeDragger == 0 {
			return
		}
		mx := math.Round(c.CanvasX(ev.X))
		my := math.Round(c.CanvasY(ev.Y))
		clamp := func(v, hi float64) float64 { return math.Max(0, math.Min(hi, v)) }
		switch activeDragger {
		case 1: // arrowMove1
			newA := clamp(math.Floor((x2+scl(5)-mx)/10), scl(25))
			if newA != a {
				c.Move("box1", 10*(a-newA), 0)
				a = newA
			}
		case 2: // arrowMove2
			newB := clamp(math.Floor((x2+scl(5)-mx)/10), scl(25))
			newC := clamp(math.Floor((y+scl(5)-my-5*w)/10), scl(20))
			if newB != b || newC != cc {
				c.Move("box2", 10*(b-newB), 10*(cc-newC))
				b, cc = newB, newC
			}
		case 3: // arrowMove3
			newW := clamp(math.Floor((y+scl(2)-my)/5), scl(20))
			if newW != w {
				c.Move("box3", 0, 5*(w-newW))
				w = newW
			}
		}
	})

	// Canvas button release: full redraw.
	app.Dispatcher().Bind(c.Win.PlatformID, event.ButtonReleaseMask, func(ev *event.Event) {
		if ev.Button == 1 && activeDragger != 0 {
			activeDragger = 0
			redraw()
		}
	})

	app.Run()
}
