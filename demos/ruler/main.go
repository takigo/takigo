// Demo: Ruler with interactive draggable tab stops.
// Ported from Tk's ruler.tcl demo.
package main

import (
	"fmt"
	"strconv"

	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
)

// Ruler geometry (pixels at ~38px/cm).
const (
	rulerLeft   = 38.0  // 1c
	rulerRight  = 494.0 // 13c
	rulerTop    = 38.0  // 1c  — ruler bottom edge / tab zone top
	rulerBottom = 57.0  // 1.5c — tab zone bottom (below = delete zone)
	tabSize     = 8.0   // 0.2c — half-size of tab triangle
	gridPx      = 9.5   // 0.25c — grid snap increment
	wellLeft    = 502.0 // 13.2c
	wellRight   = 524.0 // 13.8c
	wellTop     = 19.0  // 0.5c
)

func main() {
	app := demohelper.Setup("Ruler Demonstration", 600, 220,
		"This canvas widget shows a mock-up of a ruler. You can create tab stops by dragging them out of the well to the right of the ruler. You can also drag existing tab stops. If you drag a tab stop far enough up or down so that it turns dim, it will be deleted when you release the mouse button.")

	c := canvas.New(app, "ruler",
		canvas.Background("white"),
		canvas.Width(560),
		canvas.Height(100),
	)
	pack.Pack(c, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(10), pack.PadY(10))

	// mkTab creates a downward-pointing triangle polygon at (x, y) apex.
	mkTab := func(x, y float64, tags ...string) int64 {
		return c.CreatePolygon(
			[]float64{x, y, x + tabSize, y + tabSize, x - tabSize, y + tabSize},
			canvas.FillColor("black"),
			canvas.Tags(tags...),
		)
	}

	// Draw ruler border: two vertical sides + bottom line.
	c.CreateLine([]float64{rulerLeft, wellTop, rulerLeft, rulerTop, rulerRight, rulerTop, rulerRight, wellTop},
		canvas.OutlineColor("black"), canvas.OutlineWidth(1))

	// Draw tick marks and labels (12 centimeter marks, 0-11).
	for i := 0; i < 12; i++ {
		xi := float64(i + 1) // cm position 1-12
		x := rulerLeft + xi*38

		// Full cm tick (full height).
		c.CreateLine([]float64{x, rulerTop, x, rulerTop - 15.2},
			canvas.OutlineColor("black"), canvas.OutlineWidth(1))
		// Half-cm tick.
		c.CreateLine([]float64{x + 19, rulerTop, x + 19, rulerTop - 11.4},
			canvas.OutlineColor("black"), canvas.OutlineWidth(1))
		// Quarter-cm ticks.
		c.CreateLine([]float64{x + 9.5, rulerTop, x + 9.5, rulerTop - 7.6},
			canvas.OutlineColor("black"), canvas.OutlineWidth(1))
		c.CreateLine([]float64{x + 28.5, rulerTop, x + 28.5, rulerTop - 7.6},
			canvas.OutlineColor("black"), canvas.OutlineWidth(1))

		// Label i at just past cm mark.
		c.CreateText(x+5.7, rulerTop-9.5,
			canvas.TextOpt(fmt.Sprintf("%d", i)),
			canvas.FontOpt("Sans 7"),
			canvas.AnchorOpt(option.AnchorSW))
	}

	// Well rectangle (source of new tabs).
	c.CreateRectangle(wellLeft, wellTop, wellRight, rulerTop,
		canvas.FillColor("lightgray"), canvas.OutlineColor("gray"),
		canvas.OutlineWidth(1), canvas.Tags("well"))

	// Well tab (the prototype tab in the well).
	wellTabID := mkTab((wellLeft+wellRight)/2, wellTop+3, "well", "welltab")
	_ = wellTabID

	// Drag state.
	activeID := int64(0) // 0 = nothing being dragged
	ax := 0.0            // active tab apex x
	ay := 0.0            // active tab apex y
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
		idStr := strconv.FormatInt(activeID, 10)
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
		c.ItemConfigure(strconv.FormatInt(activeID, 10), canvas.FillColor("red"))
		c.Raise(strconv.FormatInt(activeID, 10))
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
		idStr := strconv.FormatInt(activeID, 10)
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
