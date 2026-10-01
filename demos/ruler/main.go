// Demo: Ruler with interactive draggable tab stops.
// Ported from Tk's ruler.tcl demo.
package main

import (
	"fmt"
	"github.com/msorc/takigo/screenunit"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

// Ruler geometry: demo_rulerInfo, set from Tk distances once the DPI is known.
var (
	rulerLeft, rulerRight, rulerTop, rulerBottom float64
	tabSize, gridPx                              float64
)

func cm(v float64) float64 { return screenunit.Cm(v).Float() }

func main() {
	app, err := takigo.NewApp(takigo.Title("Ruler Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("ruler"),
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
		label.Text("This canvas widget shows a mock-up of a ruler.  You can create tab stops by dragging them out of the well to the right of the ruler.  You can also drag existing tab stops.  If you drag a tab stop far enough up or down so that it turns dim, it will be deleted when you release the mouse button."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	rulerLeft, rulerRight = cm(1), cm(13)
	rulerTop, rulerBottom = cm(1), cm(1.5)
	tabSize, gridPx = cm(.2), cm(.25)

	c := canvas.New(f, "c",
		canvas.Width(screenunit.Cm(14.8)),
		canvas.Height(screenunit.Cm(2.5)),
	)
	pack.Pack(c, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// mkTab creates a downward-pointing triangle polygon at (x, y) apex.
	mkTab := func(x, y float64, tags ...string) canvas.ItemID {
		return c.CreatePolygon(
			[]float64{x, y, x + tabSize, y + tabSize, x - tabSize, y + tabSize},
			canvas.FillColor("black"), canvas.OutlineNone(),
			canvas.Tags(tags...),
		)
	}

	line := func(pts ...float64) {
		c.CreateLine(pts, canvas.OutlineColor("black"), canvas.OutlineWidth(1))
	}
	line(cm(1), cm(.5), cm(1), cm(1), cm(13), cm(1), cm(13), cm(.5))
	for i := range 12 {
		x := float64(i + 1)
		line(cm(x), cm(1), cm(x), cm(.6))
		line(cm(x+.25), cm(1), cm(x+.25), cm(.8))
		line(cm(x+.5), cm(1), cm(x+.5), cm(.7))
		line(cm(x+.75), cm(1), cm(x+.75), cm(.8))
		c.CreateText(cm(x+.15), cm(.75), canvas.TextOpt(fmt.Sprint(i)),
			canvas.AnchorOpt(option.AnchorSW))
	}

	// The well: a rectangle in the canvas background and a prototype tab
	// at [winfo pixels 13.5c], [winfo pixels .65c].
	c.CreateRectangle(cm(13.2), cm(1), cm(13.8), cm(.5),
		canvas.FillColor(widget.DefBackground), canvas.Tags("well"))
	mkTab(float64(screenunit.Cm(13.5).Pixels()), float64(screenunit.Cm(0.65).Pixels()), "well", "welltab")

	// Drag state.
	activeID := canvas.ItemID(0) // 0 = nothing being dragged
	ax := 0.0                    // active tab apex x
	ay := 0.0                    // active tab apex y
	markedDelete := false

	// snapGrid snaps x to the nearest grid position within [left, right].
	snapGrid := func(x float64) float64 {
		x = rulerLeft + gridPx*float64(int((x-rulerLeft+gridPx/2)/gridPx))
		if x < rulerLeft {
			x = rulerLeft
		}
		if x > rulerRight {
			x = rulerRight
		}
		return x
	}

	// moveActive moves the active tab to (cx, cy) and applies appropriate style.
	moveActive := func(cx, my float64) {
		if activeID == 0 {
			return
		}
		cx = snapGrid(cx)
		var cy float64
		idStr := activeID
		if my >= rulerTop && my <= rulerBottom {
			cy = rulerTop + 2
			c.ItemConfigure(idStr, canvas.FillColor("red"))
			markedDelete = false
		} else {
			cy = my - tabSize - 2
			c.ItemConfigure(idStr, canvas.FillColor("gray"))
			markedDelete = true
		}
		dx := cx - ax
		dy := cy - ay
		c.Move(idStr, dx, dy)
		ax = cx
		ay = cy
	}

	// well click: create new tab from well.
	c.BindItem("well", event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button != 1 {
			return
		}
		mx := float64(ev.X)
		my := float64(ev.Y)
		newID := mkTab(mx, my, "tab")
		activeID = newID
		ax = mx
		ay = my
		markedDelete = false
		// Immediately snap to ruler position.
		moveActive(mx, my)
	})

	// tab click: select existing tab for dragging.
	c.BindItem("tab", event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button != 1 {
			return
		}
		// Find the clicked tab via currentItem.
		ids := c.FindWithTag("current")
		if len(ids) == 0 {
			return
		}
		activeID = ids[0]
		cx := snapGrid(float64(ev.X))
		ay = rulerTop + 2
		ax = cx
		markedDelete = false
		c.ItemConfigure(activeID, canvas.FillColor("red"))
		c.Raise(activeID)
	})

	// Window-level motion: drag the active tab.
	app.Dispatcher().Bind(c.Win.PlatformID, event.MotionMask, func(ev *event.Event) {
		if ev.State&platform.Button1Mask == 0 || activeID == 0 {
			return
		}
		moveActive(float64(ev.X), float64(ev.Y))
	})

	// Window-level button release: drop or delete the active tab.
	app.Dispatcher().Bind(c.Win.PlatformID, event.ButtonReleaseMask, func(ev *event.Event) {
		if ev.Button != 1 || activeID == 0 {
			return
		}
		idStr := activeID
		if markedDelete {
			c.Delete(idStr)
		} else {
			// Restore normal (black) style and snap to ruler.
			c.ItemConfigure(idStr, canvas.FillColor("black"))
		}
		activeID = 0
		markedDelete = false
	})

	app.Run()
}
