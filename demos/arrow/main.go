// Demo: Editable arrowheads on canvas lines.
// Ported from Tk's arrow.tcl demo.
package main

import (
	"fmt"
	"os"

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
		label.WrapLength("5i"),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("This widget allows you to experiment with different widths and arrowhead shapes for lines in canvases.  To change the line width or the shape of the arrowhead, drag any of the three boxes attached to the oversized arrow.  The arrows on the right give examples at normal scale.  The text at the bottom shows the configuration options as you'd enter them for a canvas line item."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	c := canvas.New(f, "c",
		canvas.Width("375p"),
		canvas.Height("262.5p"),
		canvas.ReliefOpt(option.ReliefSunken),
		canvas.BorderWidthOpt(2),
	)
	pack.Pack(c, pack.Expand(true), pack.FillOpt(pack.FillBoth))

	// Arrow parameters (mutable state shared by closures).
	// Values are prescaled at 150% (matching Tk's scl() at scalingPct=150).
	a := 12.0   // arrowhead length along shaft (scl(8))
	b := 15.0   // arrowhead total back distance (scl(10))
	cc := 5.0   // arrowhead halfwidth (scl(3))
	w := 3.0    // line width (scl(2))
	x1 := 60.0  // arrow start (scl(40))
	x2 := 525.0 // arrow end (tip) (scl(350))
	y := 225.0  // arrow center Y (scl(150))
	bs := 8.0   // control box half-size (scl(5))

	activeDragger := 0 // 0=none, 1=box1, 2=box2, 3=box3

	var redraw func()
	redraw = func() {
		c.Delete("all")

		// Big arrow (10x scale).
		c.CreateLine([]float64{x1, y, x2, y},
			canvas.OutlineColor("LightSeaGreen"),
			canvas.OutlineWidth(int(10*w)),
			canvas.Arrow(canvas.ArrowLast),
			canvas.ArrowShape(10*a, 10*b, 10*cc))

		// Arrow outline showing arrowhead geometry.
		xtip := x2 - 10*b
		deltaY := 10*cc + 5*w
		c.CreateLine([]float64{
			x2, y,
			xtip, y + deltaY,
			x2 - 10*a, y,
			xtip, y - deltaY,
			x2, y,
		}, canvas.OutlineColor("black"), canvas.OutlineWidth(2),
			canvas.CapStyleOpt(platform.CapRound), canvas.JoinStyleOpt(platform.JoinRound))

		// Control boxes.
		c.CreateRectangle(x2-10*a-bs, y-bs, x2-10*a+bs, y+bs,
			canvas.OutlineColor("black"), canvas.OutlineWidth(1),
			canvas.Tags("box", "box1"))
		c.CreateRectangle(xtip-bs, y-deltaY-bs, xtip+bs, y-deltaY+bs,
			canvas.OutlineColor("black"), canvas.OutlineWidth(1),
			canvas.Tags("box", "box2"))
		c.CreateRectangle(x1-bs, y-5*w-bs, x1+bs, y-5*w+bs,
			canvas.OutlineColor("black"), canvas.OutlineWidth(1),
			canvas.Tags("box", "box3"))

		// Separator line.
		c.CreateLine([]float64{x2 + 75, 0, x2 + 75, 2000},
			canvas.OutlineColor("black"), canvas.OutlineWidth(2))

		// Three small example arrows on the right.
		tmp := x2 + 150
		c.CreateLine([]float64{tmp, y - 188, tmp, y - 113},
			canvas.OutlineWidth(int(w)),
			canvas.Arrow(canvas.ArrowBoth),
			canvas.ArrowShape(a, b, cc))
		c.CreateLine([]float64{tmp - 38, y, tmp + 38, y},
			canvas.OutlineWidth(int(w)),
			canvas.Arrow(canvas.ArrowBoth),
			canvas.ArrowShape(a, b, cc))
		c.CreateLine([]float64{tmp - 38, y + 113, tmp + 38, y + 188},
			canvas.OutlineWidth(int(w)),
			canvas.Arrow(canvas.ArrowBoth),
			canvas.ArrowShape(a, b, cc))

		// Dimension annotation arrows (small arrows showing measured values).
		// smallTips: 3.75p × 1.5 scale ≈ 6px, 6px, 2px
		sa, sb, sc := 8.0, 8.0, 3.0

		// c (halfwidth) annotation.
		tx := x2 + 15
		c.CreateLine([]float64{tx, y - 5*w, tx, y - deltaY},
			canvas.Arrow(canvas.ArrowBoth),
			canvas.ArrowShape(sa, sb, sc))
		c.CreateText(x2+23, y-deltaY+5*cc,
			canvas.TextOpt(fmt.Sprintf("%.0f", cc)),
			canvas.AnchorOpt(option.AnchorW))

		// width annotation.
		tx = x1 - 15
		c.CreateLine([]float64{tx, y - 5*w, tx, y + 5*w},
			canvas.Arrow(canvas.ArrowBoth),
			canvas.ArrowShape(sa, sb, sc))
		c.CreateText(x1-23, y,
			canvas.TextOpt(fmt.Sprintf("%.0f", w)),
			canvas.AnchorOpt(option.AnchorE))

		// a (vertex distance) annotation.
		ty := y + 5*w + 10*cc + 15
		c.CreateLine([]float64{x2 - 10*a, ty, x2, ty},
			canvas.Arrow(canvas.ArrowBoth),
			canvas.ArrowShape(sa, sb, sc))
		c.CreateText(x2-5*a, ty+8,
			canvas.TextOpt(fmt.Sprintf("%.0f", a)),
			canvas.AnchorOpt(option.AnchorN))

		// b (total back distance) annotation.
		ty += 38
		c.CreateLine([]float64{x2 - 10*b, ty, x2, ty},
			canvas.Arrow(canvas.ArrowBoth),
			canvas.ArrowShape(sa, sb, sc))
		c.CreateText(x2-5*b, ty+8,
			canvas.TextOpt(fmt.Sprintf("%.0f", b)),
			canvas.AnchorOpt(option.AnchorN))

		// Parameter text at bottom (232.5p and 247.5p at 144 DPI = 465px, 495px).
		c.CreateText(x1, 465,
			canvas.TextOpt(fmt.Sprintf("-width  %.0f", w)),
			canvas.AnchorOpt(option.AnchorW),
			canvas.FontOpt("Helvetica 18"))
		c.CreateText(x1, 495,
			canvas.TextOpt(fmt.Sprintf("-arrowshape  {%.0f  %.0f  %.0f}", a, b, cc)),
			canvas.AnchorOpt(option.AnchorW),
			canvas.FontOpt("Helvetica 18"))
	}

	redraw()

	// Box hover: fill red on Enter, clear fill on Leave.
	c.BindItem("box", event.EnterMask, func(ev *event.Event) {
		c.ItemConfigure("current", canvas.FillColor("red"))
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
		mx := float64(ev.X)
		my := float64(ev.Y)
		switch activeDragger {
		case 1: // box1 → controls a (arrowhead vertex length)
			newA := (x2 + bs - mx) / 10
			if newA < 0 {
				newA = 0
			}
			if newA > 38 {
				newA = 38
			}
			if newA != a {
				c.Move("box1", 10*(a-newA), 0)
				a = newA
			}
		case 2: // box2 → controls b (back distance) and c (halfwidth)
			newB := (x2 + bs - mx) / 10
			if newB < 0 {
				newB = 0
			}
			if newB > 38 {
				newB = 38
			}
			newC := (y + bs - my - 5*w) / 10
			if newC < 0 {
				newC = 0
			}
			if newC > 30 {
				newC = 30
			}
			if newB != b || newC != cc {
				c.Move("box2", 10*(b-newB), 10*(cc-newC))
				b = newB
				cc = newC
			}
		case 3: // box3 → controls w (line width)
			newW := (y + 3 - my) / 5
			if newW < 0 {
				newW = 0
			}
			if newW > 30 {
				newW = 30
			}
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
