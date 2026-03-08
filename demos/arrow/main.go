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
	app, err := takigo.NewApp(takigo.Title("Arrowhead Editor"),
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
		label.Text("This widget allows you to experiment with different widths and arrowhead shapes for lines in canvases. To change the line width or the shape of the arrowhead, drag any of the three boxes attached to the oversized arrow. The arrows on the right give examples at normal scale. The text at the bottom shows the configuration options as you'd enter them for a canvas line item."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	c := canvas.New(f, "arrows",
		canvas.Background("white"),
		canvas.Width(500),
		canvas.Height(350),
	)
	pack.Pack(c, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Arrow parameters (mutable state shared by closures).
	a := 8.0    // arrowhead length along shaft
	b := 10.0   // arrowhead total back distance
	cc := 3.0   // arrowhead halfwidth (c reserved keyword)
	w := 2.0    // line width
	x1 := 40.0  // arrow start
	x2 := 350.0 // arrow end (tip)
	y := 150.0  // arrow center Y
	bs := 5.0   // control box half-size

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
		}, canvas.OutlineColor("black"), canvas.OutlineWidth(1))

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
		c.CreateLine([]float64{x2 + 50, 0, x2 + 50, 350},
			canvas.OutlineColor("black"), canvas.OutlineWidth(1))

		// Three small example arrows on the right.
		tmp := x2 + 100
		c.CreateLine([]float64{tmp, y - 125, tmp, y - 75},
			canvas.OutlineWidth(int(w)),
			canvas.Arrow(canvas.ArrowBoth),
			canvas.ArrowShape(a, b, cc))
		c.CreateLine([]float64{tmp - 25, y, tmp + 25, y},
			canvas.OutlineWidth(int(w)),
			canvas.Arrow(canvas.ArrowBoth),
			canvas.ArrowShape(a, b, cc))
		c.CreateLine([]float64{tmp - 25, y + 75, tmp + 25, y + 125},
			canvas.OutlineWidth(int(w)),
			canvas.Arrow(canvas.ArrowBoth),
			canvas.ArrowShape(a, b, cc))

		// Dimension annotation arrows (small arrows showing measured values).
		sa, sb, sc := 3.75, 3.75, 1.5

		// c (halfwidth) annotation.
		tx := x2 + 10
		c.CreateLine([]float64{tx, y - 5*w, tx, y - deltaY},
			canvas.Arrow(canvas.ArrowBoth),
			canvas.ArrowShape(sa, sb, sc))
		c.CreateText(x2+15, y-deltaY+5*cc,
			canvas.TextOpt(fmt.Sprintf("%.0f", cc)),
			canvas.AnchorOpt(option.AnchorW))

		// width annotation.
		tx = x1 - 10
		c.CreateLine([]float64{tx, y - 5*w, tx, y + 5*w},
			canvas.Arrow(canvas.ArrowBoth),
			canvas.ArrowShape(sa, sb, sc))
		c.CreateText(x1-15, y,
			canvas.TextOpt(fmt.Sprintf("%.0f", w)),
			canvas.AnchorOpt(option.AnchorE))

		// a (vertex distance) annotation.
		ty := y + 5*w + 10*cc + 10
		c.CreateLine([]float64{x2 - 10*a, ty, x2, ty},
			canvas.Arrow(canvas.ArrowBoth),
			canvas.ArrowShape(sa, sb, sc))
		c.CreateText(x2-5*a, ty+5,
			canvas.TextOpt(fmt.Sprintf("%.0f", a)),
			canvas.AnchorOpt(option.AnchorN))

		// b (total back distance) annotation.
		ty += 25
		c.CreateLine([]float64{x2 - 10*b, ty, x2, ty},
			canvas.Arrow(canvas.ArrowBoth),
			canvas.ArrowShape(sa, sb, sc))
		c.CreateText(x2-5*b, ty+5,
			canvas.TextOpt(fmt.Sprintf("%.0f", b)),
			canvas.AnchorOpt(option.AnchorN))

		// Parameter text at bottom.
		c.CreateText(x1, 310,
			canvas.TextOpt(fmt.Sprintf("-width  %.0f", w)),
			canvas.AnchorOpt(option.AnchorW),
			canvas.FontOpt("Helvetica 14"))
		c.CreateText(x1, 330,
			canvas.TextOpt(fmt.Sprintf("-arrowshape  {%.0f  %.0f  %.0f}", a, b, cc)),
			canvas.AnchorOpt(option.AnchorW),
			canvas.FontOpt("Helvetica 14"))
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
			if newA > 25 {
				newA = 25
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
			if newB > 25 {
				newB = 25
			}
			newC := (y + bs - my - 5*w) / 10
			if newC < 0 {
				newC = 0
			}
			if newC > 20 {
				newC = 20
			}
			if newB != b || newC != cc {
				c.Move("box2", 10*(b-newB), 10*(cc-newC))
				b = newB
				cc = newC
			}
		case 3: // box3 → controls w (line width)
			newW := (y + 2 - my) / 5
			if newW < 0 {
				newW = 0
			}
			if newW > 20 {
				newW = 20
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
