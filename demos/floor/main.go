// Demo: Floorplan drawn on canvas.
// Ported from Tk's floor.tcl demo — DEC Western Research Laboratory floorplan
// with three levels, room hover highlighting, and clickable floor switching.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/ttk"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/entry"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

// Global state.
var (
	floorLabels map[canvas.ItemID]string
	floorItems  map[string]canvas.ItemID
	activeFloor int
)

// Colors for the floorplan.
var clr = struct {
	bg1, outline1 string
	bg2, outline2 string
	bg3, outline3 string
	offices       string
	active        string
}{
	bg1: "#a9c1da", outline1: "#77889a",
	bg2: "#9ab0c6", outline2: "#687786",
	bg3: "#8ba0b3", outline3: "#596673",
	offices: "Black",
	active:  "#c4d1df",
}

// Helper functions to reduce verbosity.
func poly(c *canvas.Canvas, coords []float64, fill, outline string, tags ...string) canvas.ItemID {
	opts := []canvas.ItemOption{canvas.Tags(tags...)}
	if fill == "" {
		opts = append(opts, canvas.FillNone())
	} else {
		opts = append(opts, canvas.FillColor(fill))
	}
	if outline == "" {
		opts = append(opts, canvas.OutlineNone())
	} else {
		opts = append(opts, canvas.OutlineColor(outline))
	}
	return c.CreatePolygon(coords, opts...)
}

func ln(c *canvas.Canvas, x1, y1, x2, y2 float64, color string, tags ...string) {
	c.CreateLine([]float64{x1, y1, x2, y2}, canvas.OutlineColor(color), canvas.Tags(tags...))
}

func txt(c *canvas.Canvas, x, y float64, s, color string, tags ...string) {
	c.CreateText(x, y, canvas.TextOpt(s), canvas.TextColor(color),
		canvas.AnchorOpt(option.AnchorCenter), canvas.Tags(tags...))
}

func room(c *canvas.Canvas, coords []float64, name string, tx, ty float64, color string, floorTag string) {
	id := poly(c, coords, "", "", floorTag, "room")
	floorLabels[id] = name
	floorItems[name] = id
	txt(c, tx, ty, name, color, floorTag, "label")
}

func main() {
	app, err := takigo.NewApp(takigo.Title("Floorplan Canvas Demonstration"),
		takigo.Geometry("+20+20"),
		takigo.IconName("Floorplan"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength(screenunit.In(8)),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("This window contains a canvas widget showing the floorplan of Digital Equipment Corporation's Western Research Laboratory.  It has three levels.  At any given time one of the levels is active, meaning that you can see its room structure.  To activate a level, click the left mouse button anywhere on it.  As the mouse moves over the active level, the room under the mouse lights up and its room number appears in the \"Room:\" entry.  You can also type a room number in the entry and the room will light up."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	contentFrame := frame.New(f, "frame")

	// Canvas in sunken frame.
	f1 := frame.New(contentFrame, "f1", frame.BorderWidth(2), frame.Relief(option.ReliefSunken))
	c := canvas.New(f1, "c", canvas.HighlightWidthOpt(0))
	pack.Pack(c, pack.Expand(true), pack.FillOpt(pack.FillBoth))

	// Scrollbars.
	v := ttk.NewScrollbar(contentFrame, "vscroll",
		ttk.ScrollbarOrientOpt(ttk.Vertical),
		ttk.ScrollbarCommandOpt(widget.ScrollY(c)),
	)
	h := ttk.NewScrollbar(contentFrame, "hscroll",
		ttk.ScrollbarOrientOpt(ttk.Horizontal),
		ttk.ScrollbarCommandOpt(widget.ScrollX(c)),
	)
	c.Configure(
		canvas.YScrollCommand(func(first, last float64) { v.Set(first, last) }),
		canvas.XScrollCommand(func(first, last float64) { h.Set(first, last) }),
	)

	grid.Grid(f1, grid.PadX(1), grid.PadY(1), grid.Row(0), grid.Column(0), grid.RowSpan(1), grid.ColumnSpan(1), grid.Sticky(grid.NSEW))
	grid.Grid(v, grid.PadX(1), grid.PadY(1), grid.Row(0), grid.Column(1), grid.RowSpan(1), grid.ColumnSpan(1), grid.Sticky(grid.NSEW))
	grid.Grid(h, grid.PadX(1), grid.PadY(1), grid.Row(1), grid.Column(0), grid.RowSpan(1), grid.ColumnSpan(1), grid.Sticky(grid.NSEW))
	grid.RowConfigure(contentFrame, 0, grid.Weight(1), grid.MinSize(0))
	grid.ColumnConfigure(contentFrame, 0, grid.Weight(1), grid.MinSize(0))
	pack.Pack(contentFrame, pack.Expand(true), pack.FillOpt(pack.FillBoth), pack.PadX(1), pack.PadY(1))

	// Entry widget for room display/input.
	var highlightedID canvas.ItemID

	// setHighlight changes which room is highlighted. Pass "" to clear.
	setHighlight := func(roomName string) {
		// Unhighlight previous room.
		if highlightedID != 0 {
			c.ItemConfigure(highlightedID, canvas.FillNone())
			highlightedID = 0
		}
		// Highlight new room.
		if roomName != "" {
			if itemID, ok := floorItems[roomName]; ok {
				highlightedID = itemID
				c.ItemConfigure(itemID, canvas.FillColor(clr.active))
			}
		}
	}

	ent := entry.New(c, "entry", entry.Width(10),
		entry.ValidateOpt("key"),
		entry.ValidateCmdOpt(func(prospective string) bool {
			setHighlight(prospective)
			return true
		}),
	)

	// floorDisplay recreates the floorplan for the given active floor.
	floorDisplay := func(active int) {
		if activeFloor == active {
			return
		}
		c.Delete("all")
		activeFloor = active

		bg1(c, clr.bg1, clr.outline1)
		bg2(c, clr.bg2, clr.outline2)
		bg3(c, clr.bg3, clr.outline3)

		c.Raise(fmt.Sprintf("floor%d", active))

		// Marker for Z-order reference.
		c.CreateRectangle(0, 100, 1, 101,
			canvas.FillNone(), canvas.OutlineColor(""), canvas.Tags("marker"))

		floorLabels = map[canvas.ItemID]string{}
		floorItems = map[string]canvas.ItemID{}
		switch active {
		case 1:
			fg1(c, clr.offices)
		case 2:
			fg2(c, clr.offices)
		case 3:
			fg3(c, clr.offices)
		}
		c.Raise("room")

		// Offset floors diagonally from each other.
		c.Move("floor1", screenunit.Cm(2).Float(), screenunit.Cm(2).Float())
		c.Move("floor2", screenunit.Cm(1).Float(), screenunit.Cm(1).Float())

		// Room entry and label embedded in canvas.
		c.CreateWindow(screenunit.Pt(450).Float(), screenunit.Pt(75).Float(), ent.Win, canvas.AnchorOpt(option.AnchorW))
		c.CreateText(screenunit.Pt(450).Float(), screenunit.Pt(75).Float(),
			canvas.TextOpt("Room: "),
			canvas.AnchorOpt(option.AnchorE))

		// Scroll region = bbox all; size = bbox + 20px (floor.tcl).
		x1, y1, x2, y2 := c.BBox("all")
		morePx := 20 * screenunit.ScalingPct() / 100
		c.Configure(
			canvas.ScrollRegion(x1, y1, x2, y2),
			canvas.Width(x2-x1+morePx),
			canvas.Height(y2-y1+morePx),
		)
		c.Display()
	}

	// Initial display — show floor 3.
	activeFloor = 0
	floorDisplay(3)

	// Floor clicking.
	c.BindItem("floor1", event.ButtonPressMask, func(ev *event.Event) { floorDisplay(1) })
	c.BindItem("floor2", event.ButtonPressMask, func(ev *event.Event) { floorDisplay(2) })
	c.BindItem("floor3", event.ButtonPressMask, func(ev *event.Event) { floorDisplay(3) })

	// Room hover.
	c.BindItem("room", event.EnterMask, func(ev *event.Event) {
		ids := c.FindWithTag("current")
		if len(ids) > 0 {
			if name, ok := floorLabels[ids[0]]; ok {
				setHighlight(name)
				ent.SetText(name)
			}
		}
	})
	c.BindItem("room", event.LeaveMask, func(ev *event.Event) {
		setHighlight("")
		ent.SetText("")
	})

	app.Run()
}

// ============================================================
// Floor background data (building outlines for each floor).
// ============================================================

func bg1(c *canvas.Canvas, fill, outline string) {
	poly(c, []float64{
		347, 80, 349, 82, 351, 84, 353, 85, 363, 92, 375, 99, 386, 104,
		386, 129, 398, 129, 398, 162, 484, 162, 484, 129, 559, 129, 559, 133, 725,
		133, 725, 129, 802, 129, 802, 389, 644, 389, 644, 391, 559, 391, 559, 327,
		508, 327, 508, 311, 484, 311, 484, 278, 395, 278, 395, 288, 400, 288, 404,
		288, 409, 290, 413, 292, 418, 297, 421, 302, 422, 309, 421, 318, 417, 325,
		411, 330, 405, 332, 397, 333, 344, 333, 340, 334, 336, 336, 335, 338, 332,
		342, 331, 347, 332, 351, 334, 354, 336, 357, 341, 359, 340, 360, 335, 363,
		331, 365, 326, 366, 304, 366, 304, 355, 258, 355, 258, 387, 60, 387, 60, 391,
		0, 391, 0, 337, 3, 337, 3, 114, 8, 114, 8, 25, 30, 25, 30, 5, 93, 5, 98, 5, 104, 7,
		110, 10, 116, 16, 119, 20, 122, 28, 123, 32, 123, 68, 220, 68, 220, 34, 221,
		22, 223, 17, 227, 13, 231, 8, 236, 4, 242, 2, 246, 0, 260, 0, 283, 1, 300, 5,
		321, 14, 335, 22, 348, 25, 365, 29, 363, 39, 358, 48, 352, 56, 337, 70,
		344, 76, 347, 80,
	}, fill, "", "floor1", "bg")
	ln(c, 386, 129, 398, 129, outline, "floor1", "bg")
	ln(c, 258, 355, 258, 387, outline, "floor1", "bg")
	ln(c, 60, 387, 60, 391, outline, "floor1", "bg")
	ln(c, 0, 337, 0, 391, outline, "floor1", "bg")
	ln(c, 60, 391, 0, 391, outline, "floor1", "bg")
	ln(c, 3, 114, 3, 337, outline, "floor1", "bg")
	ln(c, 258, 387, 60, 387, outline, "floor1", "bg")
	ln(c, 484, 162, 398, 162, outline, "floor1", "bg")
	ln(c, 398, 162, 398, 129, outline, "floor1", "bg")
	ln(c, 484, 278, 484, 311, outline, "floor1", "bg")
	ln(c, 484, 311, 508, 311, outline, "floor1", "bg")
	ln(c, 508, 327, 508, 311, outline, "floor1", "bg")
	ln(c, 559, 327, 508, 327, outline, "floor1", "bg")
	ln(c, 644, 391, 559, 391, outline, "floor1", "bg")
	ln(c, 644, 389, 644, 391, outline, "floor1", "bg")
	ln(c, 559, 129, 484, 129, outline, "floor1", "bg")
	ln(c, 484, 162, 484, 129, outline, "floor1", "bg")
	ln(c, 725, 133, 559, 133, outline, "floor1", "bg")
	ln(c, 559, 129, 559, 133, outline, "floor1", "bg")
	ln(c, 725, 129, 802, 129, outline, "floor1", "bg")
	ln(c, 802, 389, 802, 129, outline, "floor1", "bg")
	ln(c, 3, 337, 0, 337, outline, "floor1", "bg")
	ln(c, 559, 391, 559, 327, outline, "floor1", "bg")
	ln(c, 802, 389, 644, 389, outline, "floor1", "bg")
	ln(c, 725, 133, 725, 129, outline, "floor1", "bg")
	ln(c, 8, 25, 8, 114, outline, "floor1", "bg")
	ln(c, 8, 114, 3, 114, outline, "floor1", "bg")
	ln(c, 30, 25, 8, 25, outline, "floor1", "bg")
	ln(c, 484, 278, 395, 278, outline, "floor1", "bg")
	ln(c, 30, 25, 30, 5, outline, "floor1", "bg")
	ln(c, 93, 5, 30, 5, outline, "floor1", "bg")
	ln(c, 98, 5, 93, 5, outline, "floor1", "bg")
	ln(c, 104, 7, 98, 5, outline, "floor1", "bg")
	ln(c, 110, 10, 104, 7, outline, "floor1", "bg")
	ln(c, 116, 16, 110, 10, outline, "floor1", "bg")
	ln(c, 119, 20, 116, 16, outline, "floor1", "bg")
	ln(c, 122, 28, 119, 20, outline, "floor1", "bg")
	ln(c, 123, 32, 122, 28, outline, "floor1", "bg")
	ln(c, 123, 68, 123, 32, outline, "floor1", "bg")
	ln(c, 220, 68, 123, 68, outline, "floor1", "bg")
	ln(c, 386, 129, 386, 104, outline, "floor1", "bg")
	ln(c, 386, 104, 375, 99, outline, "floor1", "bg")
	ln(c, 375, 99, 363, 92, outline, "floor1", "bg")
	ln(c, 353, 85, 363, 92, outline, "floor1", "bg")
	ln(c, 220, 68, 220, 34, outline, "floor1", "bg")
	ln(c, 337, 70, 352, 56, outline, "floor1", "bg")
	ln(c, 352, 56, 358, 48, outline, "floor1", "bg")
	ln(c, 358, 48, 363, 39, outline, "floor1", "bg")
	ln(c, 363, 39, 365, 29, outline, "floor1", "bg")
	ln(c, 365, 29, 348, 25, outline, "floor1", "bg")
	ln(c, 348, 25, 335, 22, outline, "floor1", "bg")
	ln(c, 335, 22, 321, 14, outline, "floor1", "bg")
	ln(c, 321, 14, 300, 5, outline, "floor1", "bg")
	ln(c, 300, 5, 283, 1, outline, "floor1", "bg")
	ln(c, 283, 1, 260, 0, outline, "floor1", "bg")
	ln(c, 260, 0, 246, 0, outline, "floor1", "bg")
	ln(c, 246, 0, 242, 2, outline, "floor1", "bg")
	ln(c, 242, 2, 236, 4, outline, "floor1", "bg")
	ln(c, 236, 4, 231, 8, outline, "floor1", "bg")
	ln(c, 231, 8, 227, 13, outline, "floor1", "bg")
	ln(c, 223, 17, 227, 13, outline, "floor1", "bg")
	ln(c, 221, 22, 223, 17, outline, "floor1", "bg")
	ln(c, 220, 34, 221, 22, outline, "floor1", "bg")
	ln(c, 340, 360, 335, 363, outline, "floor1", "bg")
	ln(c, 335, 363, 331, 365, outline, "floor1", "bg")
	ln(c, 331, 365, 326, 366, outline, "floor1", "bg")
	ln(c, 326, 366, 304, 366, outline, "floor1", "bg")
	ln(c, 304, 355, 304, 366, outline, "floor1", "bg")
	ln(c, 395, 288, 400, 288, outline, "floor1", "bg")
	ln(c, 404, 288, 400, 288, outline, "floor1", "bg")
	ln(c, 409, 290, 404, 288, outline, "floor1", "bg")
	ln(c, 413, 292, 409, 290, outline, "floor1", "bg")
	ln(c, 418, 297, 413, 292, outline, "floor1", "bg")
	ln(c, 421, 302, 418, 297, outline, "floor1", "bg")
	ln(c, 422, 309, 421, 302, outline, "floor1", "bg")
	ln(c, 421, 318, 422, 309, outline, "floor1", "bg")
	ln(c, 421, 318, 417, 325, outline, "floor1", "bg")
	ln(c, 417, 325, 411, 330, outline, "floor1", "bg")
	ln(c, 411, 330, 405, 332, outline, "floor1", "bg")
	ln(c, 405, 332, 397, 333, outline, "floor1", "bg")
	ln(c, 397, 333, 344, 333, outline, "floor1", "bg")
	ln(c, 344, 333, 340, 334, outline, "floor1", "bg")
	ln(c, 340, 334, 336, 336, outline, "floor1", "bg")
	ln(c, 336, 336, 335, 338, outline, "floor1", "bg")
	ln(c, 335, 338, 332, 342, outline, "floor1", "bg")
	ln(c, 331, 347, 332, 342, outline, "floor1", "bg")
	ln(c, 332, 351, 331, 347, outline, "floor1", "bg")
	ln(c, 334, 354, 332, 351, outline, "floor1", "bg")
	ln(c, 336, 357, 334, 354, outline, "floor1", "bg")
	ln(c, 341, 359, 336, 357, outline, "floor1", "bg")
	ln(c, 341, 359, 340, 360, outline, "floor1", "bg")
	ln(c, 395, 288, 395, 278, outline, "floor1", "bg")
	ln(c, 304, 355, 258, 355, outline, "floor1", "bg")
	ln(c, 347, 80, 344, 76, outline, "floor1", "bg")
	ln(c, 344, 76, 337, 70, outline, "floor1", "bg")
	ln(c, 349, 82, 347, 80, outline, "floor1", "bg")
	ln(c, 351, 84, 349, 82, outline, "floor1", "bg")
	ln(c, 353, 85, 351, 84, outline, "floor1", "bg")
}

func bg2(c *canvas.Canvas, fill, outline string) {
	poly(c, []float64{
		559, 129, 484, 129, 484, 162, 398, 162, 398, 129, 315, 129,
		315, 133, 176, 133, 176, 129, 96, 129, 96, 133, 3, 133, 3, 339, 0, 339, 0, 391,
		60, 391, 60, 387, 258, 387, 258, 329, 350, 329, 350, 311, 395, 311, 395, 280,
		484, 280, 484, 311, 508, 311, 508, 327, 558, 327, 558, 391, 644, 391, 644,
		367, 802, 367, 802, 129, 725, 129, 725, 133, 559, 133, 559, 129,
	}, fill, "", "floor2", "bg")
	ln(c, 350, 311, 350, 329, outline, "floor2", "bg")
	ln(c, 398, 129, 398, 162, outline, "floor2", "bg")
	ln(c, 802, 367, 802, 129, outline, "floor2", "bg")
	ln(c, 802, 129, 725, 129, outline, "floor2", "bg")
	ln(c, 725, 133, 725, 129, outline, "floor2", "bg")
	ln(c, 559, 129, 559, 133, outline, "floor2", "bg")
	ln(c, 559, 133, 725, 133, outline, "floor2", "bg")
	ln(c, 484, 162, 484, 129, outline, "floor2", "bg")
	ln(c, 559, 129, 484, 129, outline, "floor2", "bg")
	ln(c, 802, 367, 644, 367, outline, "floor2", "bg")
	ln(c, 644, 367, 644, 391, outline, "floor2", "bg")
	ln(c, 644, 391, 558, 391, outline, "floor2", "bg")
	ln(c, 558, 327, 558, 391, outline, "floor2", "bg")
	ln(c, 558, 327, 508, 327, outline, "floor2", "bg")
	ln(c, 508, 327, 508, 311, outline, "floor2", "bg")
	ln(c, 484, 311, 508, 311, outline, "floor2", "bg")
	ln(c, 484, 280, 484, 311, outline, "floor2", "bg")
	ln(c, 398, 162, 484, 162, outline, "floor2", "bg")
	ln(c, 484, 280, 395, 280, outline, "floor2", "bg")
	ln(c, 395, 280, 395, 311, outline, "floor2", "bg")
	ln(c, 258, 387, 60, 387, outline, "floor2", "bg")
	ln(c, 3, 133, 3, 339, outline, "floor2", "bg")
	ln(c, 3, 339, 0, 339, outline, "floor2", "bg")
	ln(c, 60, 391, 0, 391, outline, "floor2", "bg")
	ln(c, 0, 339, 0, 391, outline, "floor2", "bg")
	ln(c, 60, 387, 60, 391, outline, "floor2", "bg")
	ln(c, 258, 329, 258, 387, outline, "floor2", "bg")
	ln(c, 350, 329, 258, 329, outline, "floor2", "bg")
	ln(c, 395, 311, 350, 311, outline, "floor2", "bg")
	ln(c, 398, 129, 315, 129, outline, "floor2", "bg")
	ln(c, 176, 133, 315, 133, outline, "floor2", "bg")
	ln(c, 176, 129, 96, 129, outline, "floor2", "bg")
	ln(c, 3, 133, 96, 133, outline, "floor2", "bg")
	ln(c, 315, 133, 315, 129, outline, "floor2", "bg")
	ln(c, 176, 133, 176, 129, outline, "floor2", "bg")
	ln(c, 96, 133, 96, 129, outline, "floor2", "bg")
}

func bg3(c *canvas.Canvas, fill, outline string) {
	poly(c, []float64{
		159, 300, 107, 300, 107, 248, 159, 248, 159, 129, 96, 129, 96,
		133, 21, 133, 21, 331, 0, 331, 0, 391, 60, 391, 60, 370, 159, 370, 159, 300,
	}, fill, "", "floor3", "bg")
	poly(c, []float64{
		258, 370, 258, 329, 350, 329, 350, 311, 399, 311, 399, 129,
		315, 129, 315, 133, 176, 133, 176, 129, 159, 129, 159, 370, 258, 370,
	}, fill, "", "floor3", "bg")
	ln(c, 96, 133, 96, 129, outline, "floor3", "bg")
	ln(c, 176, 129, 96, 129, outline, "floor3", "bg")
	ln(c, 176, 129, 176, 133, outline, "floor3", "bg")
	ln(c, 315, 133, 176, 133, outline, "floor3", "bg")
	ln(c, 315, 133, 315, 129, outline, "floor3", "bg")
	ln(c, 399, 129, 315, 129, outline, "floor3", "bg")
	ln(c, 399, 311, 399, 129, outline, "floor3", "bg")
	ln(c, 399, 311, 350, 311, outline, "floor3", "bg")
	ln(c, 350, 329, 350, 311, outline, "floor3", "bg")
	ln(c, 350, 329, 258, 329, outline, "floor3", "bg")
	ln(c, 258, 370, 258, 329, outline, "floor3", "bg")
	ln(c, 60, 370, 258, 370, outline, "floor3", "bg")
	ln(c, 60, 370, 60, 391, outline, "floor3", "bg")
	ln(c, 60, 391, 0, 391, outline, "floor3", "bg")
	ln(c, 0, 391, 0, 331, outline, "floor3", "bg")
	ln(c, 21, 331, 0, 331, outline, "floor3", "bg")
	ln(c, 21, 331, 21, 133, outline, "floor3", "bg")
	ln(c, 96, 133, 21, 133, outline, "floor3", "bg")
	ln(c, 107, 300, 159, 300, outline, "floor3", "bg")
	ln(c, 159, 300, 159, 248, outline, "floor3", "bg")
	ln(c, 159, 248, 107, 248, outline, "floor3", "bg")
	ln(c, 107, 248, 107, 300, outline, "floor3", "bg")
}

// ============================================================
// Floor 1 foreground: rooms, walls, labels.
// ============================================================

func fg1(c *canvas.Canvas, color string) {
	room(c, []float64{375, 246, 375, 172, 341, 172, 341, 246}, "101", 358, 209, color, "floor1")
	room(c, []float64{307, 240, 339, 240, 339, 206, 307, 206}, "Pub Lift1", 323, 223, color, "floor1")
	room(c, []float64{339, 205, 307, 205, 307, 171, 339, 171}, "Priv Lift1", 323, 188, color, "floor1")
	room(c, []float64{42, 389, 42, 337, 1, 337, 1, 389}, "110", 21.5, 363, color, "floor1")
	room(c, []float64{59, 389, 59, 385, 90, 385, 90, 337, 44, 337, 44, 389}, "109", 67, 363, color, "floor1")
	room(c, []float64{51, 300, 51, 253, 6, 253, 6, 300}, "111", 28.5, 276.5, color, "floor1")
	room(c, []float64{98, 248, 98, 309, 79, 309, 79, 248}, "117B", 88.5, 278.5, color, "floor1")
	room(c, []float64{51, 251, 51, 204, 6, 204, 6, 251}, "112", 28.5, 227.5, color, "floor1")
	room(c, []float64{6, 156, 51, 156, 51, 203, 6, 203}, "113", 28.5, 179.5, color, "floor1")
	room(c, []float64{85, 169, 79, 169, 79, 192, 85, 192}, "117A", 82, 180.5, color, "floor1")
	room(c, []float64{77, 302, 77, 168, 53, 168, 53, 302}, "117", 65, 235, color, "floor1")
	room(c, []float64{51, 155, 51, 115, 6, 115, 6, 155}, "114", 28.5, 135, color, "floor1")
	room(c, []float64{95, 115, 53, 115, 53, 168, 95, 168}, "115", 74, 141.5, color, "floor1")
	room(c, []float64{87, 113, 87, 27, 10, 27, 10, 113}, "116", 48.5, 70, color, "floor1")
	room(c, []float64{89, 91, 128, 91, 128, 113, 89, 113}, "118", 108.5, 102, color, "floor1")
	room(c, []float64{178, 128, 178, 132, 216, 132, 216, 91, 163, 91, 163, 112, 149, 112, 149, 128}, "120", 189.5, 111.5, color, "floor1")
	room(c, []float64{79, 193, 87, 193, 87, 169, 136, 169, 136, 192, 156, 192, 156, 169, 175, 169, 175, 246, 79, 246}, "122", 131, 207.5, color, "floor1")
	room(c, []float64{138, 169, 154, 169, 154, 191, 138, 191}, "121", 146, 180, color, "floor1")
	room(c, []float64{99, 300, 126, 300, 126, 309, 99, 309}, "106A", 112.5, 304.5, color, "floor1")
	room(c, []float64{128, 299, 128, 309, 150, 309, 150, 248, 99, 248, 99, 299}, "105", 124.5, 278.5, color, "floor1")
	room(c, []float64{174, 309, 174, 300, 152, 300, 152, 309}, "106B", 163, 304.5, color, "floor1")
	room(c, []float64{176, 299, 176, 309, 216, 309, 216, 248, 152, 248, 152, 299}, "104", 184, 278.5, color, "floor1")
	room(c, []float64{138, 385, 138, 337, 91, 337, 91, 385}, "108", 114.5, 361, color, "floor1")
	room(c, []float64{256, 337, 140, 337, 140, 385, 256, 385}, "107", 198, 361, color, "floor1")
	room(c, []float64{300, 353, 300, 329, 260, 329, 260, 353}, "Smoking", 280, 341, color, "floor1")
	room(c, []float64{314, 135, 314, 170, 306, 170, 306, 246, 177, 246, 177, 135}, "123", 245.5, 190.5, color, "floor1")
	room(c, []float64{217, 248, 301, 248, 301, 326, 257, 326, 257, 310, 217, 310}, "103", 259, 287, color, "floor1")
	room(c, []float64{396, 188, 377, 188, 377, 169, 316, 169, 316, 131, 396, 131}, "124", 356, 150, color, "floor1")
	room(c, []float64{397, 226, 407, 226, 407, 189, 377, 189, 377, 246, 397, 246}, "125", 392, 217.5, color, "floor1")
	room(c, []float64{399, 187, 409, 187, 409, 207, 474, 207, 474, 164, 399, 164}, "126", 436.5, 185.5, color, "floor1")
	room(c, []float64{409, 209, 409, 229, 399, 229, 399, 253, 486, 253, 486, 239, 474, 239, 474, 209}, "127", 436.5, 231, color, "floor1")
	room(c, []float64{501, 164, 501, 174, 495, 174, 495, 188, 490, 188, 490, 204, 476, 204, 476, 164}, "MShower", 488.5, 184, color, "floor1")
	room(c, []float64{497, 176, 513, 176, 513, 204, 492, 204, 492, 190, 497, 190}, "Closet", 502.5, 190, color, "floor1")
	room(c, []float64{476, 237, 476, 206, 513, 206, 513, 254, 488, 254, 488, 237}, "WShower", 494.5, 230, color, "floor1")
	room(c, []float64{486, 131, 558, 131, 558, 135, 724, 135, 724, 166, 697, 166, 697, 275, 553, 275, 531, 254, 515, 254, 515, 174, 503, 174, 503, 161, 486, 161}, "130", 638.5, 205, color, "floor1")
	room(c, []float64{308, 242, 339, 242, 339, 248, 342, 248, 342, 246, 397, 246, 397, 276, 393, 276, 393, 309, 300, 309, 300, 248, 308, 248}, "102", 367.5, 278.5, color, "floor1")
	room(c, []float64{397, 255, 486, 255, 486, 276, 397, 276}, "128", 441.5, 265.5, color, "floor1")
	room(c, []float64{510, 309, 486, 309, 486, 255, 530, 255, 552, 277, 561, 277, 561, 325, 510, 325}, "129", 535.5, 293, color, "floor1")
	room(c, []float64{696, 281, 740, 281, 740, 387, 642, 387, 642, 389, 561, 389, 561, 277, 696, 277}, "133", 628.5, 335, color, "floor1")
	room(c, []float64{742, 387, 742, 281, 800, 281, 800, 387}, "132", 771, 334, color, "floor1")
	room(c, []float64{800, 168, 800, 280, 699, 280, 699, 168}, "134", 749.5, 224, color, "floor1")
	room(c, []float64{726, 131, 726, 166, 800, 166, 800, 131}, "135", 763, 148.5, color, "floor1")
	room(c, []float64{340, 360, 335, 363, 331, 365, 326, 366, 304, 366, 304, 312, 396, 312, 396, 288, 400, 288, 404, 288, 409, 290, 413, 292, 418, 297, 421, 302, 422, 309, 421, 318, 417, 325, 411, 330, 405, 332, 397, 333, 344, 333, 340, 334, 336, 336, 335, 338, 332, 342, 331, 347, 332, 351, 334, 354, 336, 357, 341, 359}, "Ramona Stair", 368, 323, color, "floor1")
	room(c, []float64{30, 23, 30, 5, 93, 5, 98, 5, 104, 7, 110, 10, 116, 16, 119, 20, 122, 28, 123, 32, 123, 68, 220, 68, 220, 87, 90, 87, 90, 23}, "University Stair", 155, 77.5, color, "floor1")
	room(c, []float64{282, 37, 295, 40, 312, 49, 323, 56, 337, 70, 352, 56, 358, 48, 363, 39, 365, 29, 348, 25, 335, 22, 321, 14, 300, 5, 283, 1, 260, 0, 246, 0, 242, 2, 236, 4, 231, 8, 227, 13, 223, 17, 221, 22, 220, 34, 260, 34}, "Plaza Stair", 317.5, 28.5, color, "floor1")
	room(c, []float64{220, 34, 260, 34, 282, 37, 295, 40, 312, 49, 323, 56, 337, 70, 350, 83, 365, 94, 377, 100, 386, 104, 386, 128, 220, 128}, "Plaza Deck", 303, 81, color, "floor1")
	room(c, []float64{257, 336, 77, 336, 6, 336, 6, 301, 77, 301, 77, 310, 257, 310}, "106", 131.5, 318.5, color, "floor1")
	room(c, []float64{146, 110, 162, 110, 162, 91, 130, 91, 130, 115, 95, 115, 95, 128, 114, 128, 114, 151, 157, 151, 157, 153, 112, 153, 112, 130, 97, 130, 97, 168, 175, 168, 175, 131, 146, 131}, "119", 143.5, 133, color, "floor1")

	// Walls.
	ln(c, 155, 191, 155, 189, color, "floor1", "wall")
	ln(c, 155, 177, 155, 169, color, "floor1", "wall")
	ln(c, 96, 129, 96, 169, color, "floor1", "wall")
	ln(c, 78, 169, 176, 169, color, "floor1", "wall")
	ln(c, 176, 247, 176, 129, color, "floor1", "wall")
	ln(c, 340, 206, 307, 206, color, "floor1", "wall")
	ln(c, 340, 187, 340, 170, color, "floor1", "wall")
	ln(c, 340, 210, 340, 201, color, "floor1", "wall")
	ln(c, 340, 247, 340, 224, color, "floor1", "wall")
	ln(c, 340, 241, 307, 241, color, "floor1", "wall")
	ln(c, 376, 246, 376, 170, color, "floor1", "wall")
	ln(c, 307, 247, 307, 170, color, "floor1", "wall")
	ln(c, 376, 170, 307, 170, color, "floor1", "wall")
	ln(c, 315, 129, 315, 170, color, "floor1", "wall")
	ln(c, 147, 129, 176, 129, color, "floor1", "wall")
	ln(c, 202, 133, 176, 133, color, "floor1", "wall")
	ln(c, 398, 129, 315, 129, color, "floor1", "wall")
	ln(c, 258, 352, 258, 387, color, "floor1", "wall")
	ln(c, 60, 387, 60, 391, color, "floor1", "wall")
	ln(c, 0, 337, 0, 391, color, "floor1", "wall")
	ln(c, 60, 391, 0, 391, color, "floor1", "wall")
	ln(c, 3, 114, 3, 337, color, "floor1", "wall")
	ln(c, 258, 387, 60, 387, color, "floor1", "wall")
	ln(c, 52, 237, 52, 273, color, "floor1", "wall")
	ln(c, 52, 189, 52, 225, color, "floor1", "wall")
	ln(c, 52, 140, 52, 177, color, "floor1", "wall")
	ln(c, 395, 306, 395, 311, color, "floor1", "wall")
	ln(c, 531, 254, 398, 254, color, "floor1", "wall")
	ln(c, 475, 178, 475, 238, color, "floor1", "wall")
	ln(c, 502, 162, 398, 162, color, "floor1", "wall")
	ln(c, 398, 129, 398, 188, color, "floor1", "wall")
	ln(c, 383, 188, 376, 188, color, "floor1", "wall")
	ln(c, 408, 188, 408, 194, color, "floor1", "wall")
	ln(c, 398, 227, 398, 254, color, "floor1", "wall")
	ln(c, 408, 227, 398, 227, color, "floor1", "wall")
	ln(c, 408, 222, 408, 227, color, "floor1", "wall")
	ln(c, 408, 206, 408, 210, color, "floor1", "wall")
	ln(c, 408, 208, 475, 208, color, "floor1", "wall")
	ln(c, 484, 278, 484, 311, color, "floor1", "wall")
	ln(c, 484, 311, 508, 311, color, "floor1", "wall")
	ln(c, 508, 327, 508, 311, color, "floor1", "wall")
	ln(c, 559, 327, 508, 327, color, "floor1", "wall")
	ln(c, 644, 391, 559, 391, color, "floor1", "wall")
	ln(c, 644, 389, 644, 391, color, "floor1", "wall")
	ln(c, 514, 205, 475, 205, color, "floor1", "wall")
	ln(c, 496, 189, 496, 187, color, "floor1", "wall")
	ln(c, 559, 129, 484, 129, color, "floor1", "wall")
	ln(c, 484, 162, 484, 129, color, "floor1", "wall")
	ln(c, 725, 133, 559, 133, color, "floor1", "wall")
	ln(c, 559, 129, 559, 133, color, "floor1", "wall")
	ln(c, 725, 149, 725, 167, color, "floor1", "wall")
	ln(c, 725, 129, 802, 129, color, "floor1", "wall")
	ln(c, 802, 389, 802, 129, color, "floor1", "wall")
	ln(c, 739, 167, 802, 167, color, "floor1", "wall")
	ln(c, 396, 188, 408, 188, color, "floor1", "wall")
	ln(c, 0, 337, 9, 337, color, "floor1", "wall")
	ln(c, 58, 337, 21, 337, color, "floor1", "wall")
	ln(c, 43, 391, 43, 337, color, "floor1", "wall")
	ln(c, 105, 337, 75, 337, color, "floor1", "wall")
	ln(c, 91, 387, 91, 337, color, "floor1", "wall")
	ln(c, 154, 337, 117, 337, color, "floor1", "wall")
	ln(c, 139, 387, 139, 337, color, "floor1", "wall")
	ln(c, 227, 337, 166, 337, color, "floor1", "wall")
	ln(c, 258, 337, 251, 337, color, "floor1", "wall")
	ln(c, 258, 328, 302, 328, color, "floor1", "wall")
	ln(c, 302, 355, 302, 311, color, "floor1", "wall")
	ln(c, 395, 311, 302, 311, color, "floor1", "wall")
	ln(c, 484, 278, 395, 278, color, "floor1", "wall")
	ln(c, 395, 294, 395, 278, color, "floor1", "wall")
	ln(c, 473, 278, 473, 275, color, "floor1", "wall")
	ln(c, 473, 256, 473, 254, color, "floor1", "wall")
	ln(c, 533, 257, 531, 254, color, "floor1", "wall")
	ln(c, 553, 276, 551, 274, color, "floor1", "wall")
	ln(c, 698, 276, 553, 276, color, "floor1", "wall")
	ln(c, 559, 391, 559, 327, color, "floor1", "wall")
	ln(c, 802, 389, 644, 389, color, "floor1", "wall")
	ln(c, 741, 314, 741, 389, color, "floor1", "wall")
	ln(c, 698, 280, 698, 167, color, "floor1", "wall")
	ln(c, 707, 280, 698, 280, color, "floor1", "wall")
	ln(c, 802, 280, 731, 280, color, "floor1", "wall")
	ln(c, 741, 280, 741, 302, color, "floor1", "wall")
	ln(c, 698, 167, 727, 167, color, "floor1", "wall")
	ln(c, 725, 137, 725, 129, color, "floor1", "wall")
	ln(c, 514, 254, 514, 175, color, "floor1", "wall")
	ln(c, 496, 175, 514, 175, color, "floor1", "wall")
	ln(c, 502, 175, 502, 162, color, "floor1", "wall")
	ln(c, 475, 166, 475, 162, color, "floor1", "wall")
	ln(c, 496, 176, 496, 175, color, "floor1", "wall")
	ln(c, 491, 189, 496, 189, color, "floor1", "wall")
	ln(c, 491, 205, 491, 189, color, "floor1", "wall")
	ln(c, 487, 238, 475, 238, color, "floor1", "wall")
	ln(c, 487, 240, 487, 238, color, "floor1", "wall")
	ln(c, 487, 252, 487, 254, color, "floor1", "wall")
	ln(c, 315, 133, 304, 133, color, "floor1", "wall")
	ln(c, 256, 133, 280, 133, color, "floor1", "wall")
	ln(c, 78, 247, 270, 247, color, "floor1", "wall")
	ln(c, 307, 247, 294, 247, color, "floor1", "wall")
	ln(c, 214, 133, 232, 133, color, "floor1", "wall")
	ln(c, 217, 247, 217, 266, color, "floor1", "wall")
	ln(c, 217, 309, 217, 291, color, "floor1", "wall")
	ln(c, 217, 309, 172, 309, color, "floor1", "wall")
	ln(c, 154, 309, 148, 309, color, "floor1", "wall")
	ln(c, 175, 300, 175, 309, color, "floor1", "wall")
	ln(c, 151, 300, 175, 300, color, "floor1", "wall")
	ln(c, 151, 247, 151, 309, color, "floor1", "wall")
	ln(c, 78, 237, 78, 265, color, "floor1", "wall")
	ln(c, 78, 286, 78, 309, color, "floor1", "wall")
	ln(c, 106, 309, 78, 309, color, "floor1", "wall")
	ln(c, 130, 309, 125, 309, color, "floor1", "wall")
	ln(c, 99, 309, 99, 247, color, "floor1", "wall")
	ln(c, 127, 299, 99, 299, color, "floor1", "wall")
	ln(c, 127, 309, 127, 299, color, "floor1", "wall")
	ln(c, 155, 191, 137, 191, color, "floor1", "wall")
	ln(c, 137, 169, 137, 191, color, "floor1", "wall")
	ln(c, 78, 171, 78, 169, color, "floor1", "wall")
	ln(c, 78, 190, 78, 218, color, "floor1", "wall")
	ln(c, 86, 192, 86, 169, color, "floor1", "wall")
	ln(c, 86, 192, 78, 192, color, "floor1", "wall")
	ln(c, 52, 301, 3, 301, color, "floor1", "wall")
	ln(c, 52, 286, 52, 301, color, "floor1", "wall")
	ln(c, 52, 252, 3, 252, color, "floor1", "wall")
	ln(c, 52, 203, 3, 203, color, "floor1", "wall")
	ln(c, 3, 156, 52, 156, color, "floor1", "wall")
	ln(c, 8, 25, 8, 114, color, "floor1", "wall")
	ln(c, 63, 114, 3, 114, color, "floor1", "wall")
	ln(c, 75, 114, 97, 114, color, "floor1", "wall")
	ln(c, 108, 114, 129, 114, color, "floor1", "wall")
	ln(c, 129, 114, 129, 89, color, "floor1", "wall")
	ln(c, 52, 114, 52, 128, color, "floor1", "wall")
	ln(c, 132, 89, 88, 89, color, "floor1", "wall")
	ln(c, 88, 25, 88, 89, color, "floor1", "wall")
	ln(c, 88, 114, 88, 89, color, "floor1", "wall")
	ln(c, 218, 89, 144, 89, color, "floor1", "wall")
	ln(c, 147, 111, 147, 129, color, "floor1", "wall")
	ln(c, 162, 111, 147, 111, color, "floor1", "wall")
	ln(c, 162, 109, 162, 111, color, "floor1", "wall")
	ln(c, 162, 96, 162, 89, color, "floor1", "wall")
	ln(c, 218, 89, 218, 94, color, "floor1", "wall")
	ln(c, 218, 89, 218, 119, color, "floor1", "wall")
	ln(c, 8, 25, 88, 25, color, "floor1", "wall")
	ln(c, 258, 337, 258, 328, color, "floor1", "wall")
	ln(c, 113, 129, 96, 129, color, "floor1", "wall")
	ln(c, 302, 355, 258, 355, color, "floor1", "wall")
	ln(c, 386, 104, 386, 129, color, "floor1", "wall")
	ln(c, 377, 100, 386, 104, color, "floor1", "wall")
	ln(c, 365, 94, 377, 100, color, "floor1", "wall")
	ln(c, 350, 83, 365, 94, color, "floor1", "wall")
	ln(c, 337, 70, 350, 83, color, "floor1", "wall")
	ln(c, 337, 70, 323, 56, color, "floor1", "wall")
	ln(c, 312, 49, 323, 56, color, "floor1", "wall")
	ln(c, 295, 40, 312, 49, color, "floor1", "wall")
	ln(c, 282, 37, 295, 40, color, "floor1", "wall")
	ln(c, 260, 34, 282, 37, color, "floor1", "wall")
	ln(c, 253, 34, 260, 34, color, "floor1", "wall")
	ln(c, 386, 128, 386, 104, color, "floor1", "wall")
	ln(c, 113, 152, 156, 152, color, "floor1", "wall")
	ln(c, 113, 152, 113, 129, color, "floor1", "wall")
}

// ============================================================
// Floor 2 foreground: rooms, walls, labels.
// ============================================================

func fg2(c *canvas.Canvas, color string) {
	room(c, []float64{748, 188, 755, 188, 755, 205, 758, 205, 758, 222, 800, 222, 800, 168, 748, 168}, "238", 774, 195, color, "floor2")
	room(c, []float64{726, 188, 746, 188, 746, 166, 800, 166, 800, 131, 726, 131}, "237", 763, 148.5, color, "floor2")
	room(c, []float64{497, 187, 497, 204, 559, 204, 559, 324, 641, 324, 643, 324, 643, 291, 641, 291, 641, 205, 696, 205, 696, 291, 694, 291, 694, 314, 715, 314, 715, 291, 715, 205, 755, 205, 755, 190, 724, 190, 724, 187}, "246", 600, 264, color, "floor2")
	room(c, []float64{694, 279, 643, 279, 643, 314, 694, 314}, "247", 668.5, 296.5, color, "floor2")
	room(c, []float64{232, 250, 308, 250, 308, 242, 339, 242, 339, 246, 397, 246, 397, 255, 476, 255, 476, 250, 482, 250, 559, 250, 559, 274, 482, 274, 482, 278, 396, 278, 396, 274, 232, 274}, "202", 285.5, 260, color, "floor2")
	room(c, []float64{53, 228, 53, 338, 176, 338, 233, 338, 233, 196, 306, 196, 306, 180, 175, 180, 175, 169, 156, 169, 156, 196, 176, 196, 176, 228}, "206", 143, 267, color, "floor2")
	room(c, []float64{51, 277, 6, 277, 6, 338, 51, 338}, "212", 28.5, 307.5, color, "floor2")
	room(c, []float64{557, 276, 486, 276, 486, 309, 510, 309, 510, 325, 557, 325}, "245", 521.5, 300.5, color, "floor2")
	room(c, []float64{560, 389, 599, 389, 599, 326, 560, 326}, "244", 579.5, 357.5, color, "floor2")
	room(c, []float64{601, 389, 601, 326, 643, 326, 643, 389}, "243", 622, 357.5, color, "floor2")
	room(c, []float64{688, 316, 645, 316, 645, 365, 688, 365}, "242", 666.5, 340.5, color, "floor2")
	room(c, []float64{802, 367, 759, 367, 759, 226, 802, 226}, "Barbecue Deck", 780.5, 296.5, color, "floor2")
	room(c, []float64{755, 262, 755, 314, 717, 314, 717, 262}, "240", 736, 288, color, "floor2")
	room(c, []float64{755, 316, 689, 316, 689, 365, 755, 365}, "241", 722, 340.5, color, "floor2")
	room(c, []float64{755, 206, 717, 206, 717, 261, 755, 261}, "239", 736, 233.5, color, "floor2")
	room(c, []float64{695, 277, 643, 277, 643, 206, 695, 206}, "248", 669, 241.5, color, "floor2")
	room(c, []float64{676, 135, 676, 185, 724, 185, 724, 135}, "236", 700, 160, color, "floor2")
	room(c, []float64{675, 135, 635, 135, 635, 145, 628, 145, 628, 185, 675, 185}, "235", 651.5, 160, color, "floor2")
	room(c, []float64{626, 143, 633, 143, 633, 135, 572, 135, 572, 143, 579, 143, 579, 185, 626, 185}, "234", 606, 160, color, "floor2")
	room(c, []float64{557, 135, 571, 135, 571, 145, 578, 145, 578, 185, 527, 185, 527, 131, 557, 131}, "233", 552.5, 158, color, "floor2")
	room(c, []float64{476, 249, 557, 249, 557, 205, 476, 205}, "230", 516.5, 227, color, "floor2")
	room(c, []float64{476, 164, 486, 164, 486, 131, 525, 131, 525, 185, 476, 185}, "232", 500.5, 158, color, "floor2")
	room(c, []float64{476, 186, 495, 186, 495, 204, 476, 204}, "229", 485.5, 195, color, "floor2")
	room(c, []float64{474, 207, 409, 207, 409, 187, 399, 187, 399, 164, 474, 164}, "227", 436.5, 185.5, color, "floor2")
	room(c, []float64{399, 228, 399, 253, 474, 253, 474, 209, 409, 209, 409, 228}, "228", 436.5, 231, color, "floor2")
	room(c, []float64{397, 246, 397, 226, 407, 226, 407, 189, 377, 189, 377, 246}, "226", 392, 217.5, color, "floor2")
	room(c, []float64{377, 169, 316, 169, 316, 131, 397, 131, 397, 188, 377, 188}, "225", 356.5, 150, color, "floor2")
	room(c, []float64{234, 198, 306, 198, 306, 249, 234, 249}, "224", 270, 223.5, color, "floor2")
	room(c, []float64{270, 179, 306, 179, 306, 170, 314, 170, 314, 135, 270, 135}, "223", 292, 157, color, "floor2")
	room(c, []float64{268, 179, 221, 179, 221, 135, 268, 135}, "222", 244.5, 157, color, "floor2")
	room(c, []float64{177, 179, 219, 179, 219, 135, 177, 135}, "221", 198, 157, color, "floor2")
	room(c, []float64{299, 327, 349, 327, 349, 284, 341, 284, 341, 276, 299, 276}, "204", 324, 301.5, color, "floor2")
	room(c, []float64{234, 276, 297, 276, 297, 327, 257, 327, 257, 338, 234, 338}, "205", 265.5, 307, color, "floor2")
	room(c, []float64{256, 385, 256, 340, 212, 340, 212, 385}, "207", 234, 362.5, color, "floor2")
	room(c, []float64{210, 340, 164, 340, 164, 385, 210, 385}, "208", 187, 362.5, color, "floor2")
	room(c, []float64{115, 340, 162, 340, 162, 385, 115, 385}, "209", 138.5, 362.5, color, "floor2")
	room(c, []float64{89, 228, 89, 156, 53, 156, 53, 228}, "217", 71, 192, color, "floor2")
	room(c, []float64{89, 169, 97, 169, 97, 190, 89, 190}, "217A", 93, 179.5, color, "floor2")
	room(c, []float64{89, 156, 89, 168, 95, 168, 95, 135, 53, 135, 53, 156}, "216", 71, 145.5, color, "floor2")
	room(c, []float64{51, 179, 51, 135, 6, 135, 6, 179}, "215", 28.5, 157, color, "floor2")
	room(c, []float64{51, 227, 6, 227, 6, 180, 51, 180}, "214", 28.5, 203.5, color, "floor2")
	room(c, []float64{51, 275, 6, 275, 6, 229, 51, 229}, "213", 28.5, 252, color, "floor2")
	room(c, []float64{114, 340, 67, 340, 67, 385, 114, 385}, "210", 90.5, 362.5, color, "floor2")
	room(c, []float64{59, 389, 59, 385, 65, 385, 65, 340, 1, 340, 1, 389}, "211", 33, 364.5, color, "floor2")
	room(c, []float64{393, 309, 350, 309, 350, 282, 342, 282, 342, 276, 393, 276}, "203", 367.5, 292.5, color, "floor2")
	room(c, []float64{99, 191, 91, 191, 91, 226, 174, 226, 174, 198, 154, 198, 154, 192, 109, 192, 109, 169, 99, 169}, "220", 132.5, 208.5, color, "floor2")
	room(c, []float64{339, 205, 307, 205, 307, 171, 339, 171}, "Priv Lift2", 323, 188, color, "floor2")
	room(c, []float64{307, 240, 339, 240, 339, 206, 307, 206}, "Pub Lift 2", 323, 223, color, "floor2")
	room(c, []float64{175, 168, 97, 168, 97, 131, 175, 131}, "218", 136, 149.5, color, "floor2")
	room(c, []float64{154, 191, 111, 191, 111, 169, 154, 169}, "219", 132.5, 180, color, "floor2")
	room(c, []float64{375, 246, 375, 172, 341, 172, 341, 246}, "201", 358, 209, color, "floor2")

	// Walls.
	ln(c, 641, 186, 678, 186, color, "floor2", "wall")
	ln(c, 757, 350, 757, 367, color, "floor2", "wall")
	ln(c, 634, 133, 634, 144, color, "floor2", "wall")
	ln(c, 634, 144, 627, 144, color, "floor2", "wall")
	ln(c, 572, 133, 572, 144, color, "floor2", "wall")
	ln(c, 572, 144, 579, 144, color, "floor2", "wall")
	ln(c, 398, 129, 398, 162, color, "floor2", "wall")
	ln(c, 174, 197, 175, 197, color, "floor2", "wall")
	ln(c, 175, 197, 175, 227, color, "floor2", "wall")
	ln(c, 757, 206, 757, 221, color, "floor2", "wall")
	ln(c, 396, 188, 408, 188, color, "floor2", "wall")
	ln(c, 727, 189, 725, 189, color, "floor2", "wall")
	ln(c, 747, 167, 802, 167, color, "floor2", "wall")
	ln(c, 747, 167, 747, 189, color, "floor2", "wall")
	ln(c, 755, 189, 739, 189, color, "floor2", "wall")
	ln(c, 769, 224, 757, 224, color, "floor2", "wall")
	ln(c, 802, 224, 802, 129, color, "floor2", "wall")
	ln(c, 802, 129, 725, 129, color, "floor2", "wall")
	ln(c, 725, 189, 725, 129, color, "floor2", "wall")
	ln(c, 725, 186, 690, 186, color, "floor2", "wall")
	ln(c, 676, 133, 676, 186, color, "floor2", "wall")
	ln(c, 627, 144, 627, 186, color, "floor2", "wall")
	ln(c, 629, 186, 593, 186, color, "floor2", "wall")
	ln(c, 579, 144, 579, 186, color, "floor2", "wall")
	ln(c, 559, 129, 559, 133, color, "floor2", "wall")
	ln(c, 725, 133, 559, 133, color, "floor2", "wall")
	ln(c, 484, 162, 484, 129, color, "floor2", "wall")
	ln(c, 559, 129, 484, 129, color, "floor2", "wall")
	ln(c, 526, 129, 526, 186, color, "floor2", "wall")
	ln(c, 540, 186, 581, 186, color, "floor2", "wall")
	ln(c, 528, 186, 523, 186, color, "floor2", "wall")
	ln(c, 511, 186, 475, 186, color, "floor2", "wall")
	ln(c, 496, 190, 496, 186, color, "floor2", "wall")
	ln(c, 496, 205, 496, 202, color, "floor2", "wall")
	ln(c, 475, 205, 527, 205, color, "floor2", "wall")
	ln(c, 558, 205, 539, 205, color, "floor2", "wall")
	ln(c, 558, 205, 558, 249, color, "floor2", "wall")
	ln(c, 558, 249, 475, 249, color, "floor2", "wall")
	ln(c, 662, 206, 642, 206, color, "floor2", "wall")
	ln(c, 695, 206, 675, 206, color, "floor2", "wall")
	ln(c, 695, 278, 642, 278, color, "floor2", "wall")
	ln(c, 642, 291, 642, 206, color, "floor2", "wall")
	ln(c, 695, 291, 695, 206, color, "floor2", "wall")
	ln(c, 716, 208, 716, 206, color, "floor2", "wall")
	ln(c, 757, 206, 716, 206, color, "floor2", "wall")
	ln(c, 757, 221, 757, 224, color, "floor2", "wall")
	ln(c, 793, 224, 802, 224, color, "floor2", "wall")
	ln(c, 757, 262, 716, 262, color, "floor2", "wall")
	ln(c, 716, 220, 716, 264, color, "floor2", "wall")
	ln(c, 716, 315, 716, 276, color, "floor2", "wall")
	ln(c, 757, 315, 703, 315, color, "floor2", "wall")
	ln(c, 757, 325, 757, 224, color, "floor2", "wall")
	ln(c, 757, 367, 644, 367, color, "floor2", "wall")
	ln(c, 689, 367, 689, 315, color, "floor2", "wall")
	ln(c, 647, 315, 644, 315, color, "floor2", "wall")
	ln(c, 659, 315, 691, 315, color, "floor2", "wall")
	ln(c, 600, 325, 600, 391, color, "floor2", "wall")
	ln(c, 627, 325, 644, 325, color, "floor2", "wall")
	ln(c, 644, 391, 644, 315, color, "floor2", "wall")
	ln(c, 615, 325, 575, 325, color, "floor2", "wall")
	ln(c, 644, 391, 558, 391, color, "floor2", "wall")
	ln(c, 563, 325, 558, 325, color, "floor2", "wall")
	ln(c, 558, 391, 558, 314, color, "floor2", "wall")
	ln(c, 558, 327, 508, 327, color, "floor2", "wall")
	ln(c, 558, 275, 484, 275, color, "floor2", "wall")
	ln(c, 558, 302, 558, 275, color, "floor2", "wall")
	ln(c, 508, 327, 508, 311, color, "floor2", "wall")
	ln(c, 484, 311, 508, 311, color, "floor2", "wall")
	ln(c, 484, 275, 484, 311, color, "floor2", "wall")
	ln(c, 475, 208, 408, 208, color, "floor2", "wall")
	ln(c, 408, 206, 408, 210, color, "floor2", "wall")
	ln(c, 408, 222, 408, 227, color, "floor2", "wall")
	ln(c, 408, 227, 398, 227, color, "floor2", "wall")
	ln(c, 398, 227, 398, 254, color, "floor2", "wall")
	ln(c, 408, 188, 408, 194, color, "floor2", "wall")
	ln(c, 383, 188, 376, 188, color, "floor2", "wall")
	ln(c, 398, 188, 398, 162, color, "floor2", "wall")
	ln(c, 398, 162, 484, 162, color, "floor2", "wall")
	ln(c, 475, 162, 475, 254, color, "floor2", "wall")
	ln(c, 398, 254, 475, 254, color, "floor2", "wall")
	ln(c, 484, 280, 395, 280, color, "floor2", "wall")
	ln(c, 395, 311, 395, 275, color, "floor2", "wall")
	ln(c, 307, 197, 293, 197, color, "floor2", "wall")
	ln(c, 278, 197, 233, 197, color, "floor2", "wall")
	ln(c, 233, 197, 233, 249, color, "floor2", "wall")
	ln(c, 307, 179, 284, 179, color, "floor2", "wall")
	ln(c, 233, 249, 278, 249, color, "floor2", "wall")
	ln(c, 269, 179, 269, 133, color, "floor2", "wall")
	ln(c, 220, 179, 220, 133, color, "floor2", "wall")
	ln(c, 155, 191, 110, 191, color, "floor2", "wall")
	ln(c, 90, 190, 98, 190, color, "floor2", "wall")
	ln(c, 98, 169, 98, 190, color, "floor2", "wall")
	ln(c, 52, 133, 52, 165, color, "floor2", "wall")
	ln(c, 52, 214, 52, 177, color, "floor2", "wall")
	ln(c, 52, 226, 52, 262, color, "floor2", "wall")
	ln(c, 52, 274, 52, 276, color, "floor2", "wall")
	ln(c, 234, 275, 234, 339, color, "floor2", "wall")
	ln(c, 226, 339, 258, 339, color, "floor2", "wall")
	ln(c, 211, 387, 211, 339, color, "floor2", "wall")
	ln(c, 214, 339, 177, 339, color, "floor2", "wall")
	ln(c, 258, 387, 60, 387, color, "floor2", "wall")
	ln(c, 3, 133, 3, 339, color, "floor2", "wall")
	ln(c, 165, 339, 129, 339, color, "floor2", "wall")
	ln(c, 117, 339, 80, 339, color, "floor2", "wall")
	ln(c, 68, 339, 59, 339, color, "floor2", "wall")
	ln(c, 0, 339, 46, 339, color, "floor2", "wall")
	ln(c, 60, 391, 0, 391, color, "floor2", "wall")
	ln(c, 0, 339, 0, 391, color, "floor2", "wall")
	ln(c, 60, 387, 60, 391, color, "floor2", "wall")
	ln(c, 258, 329, 258, 387, color, "floor2", "wall")
	ln(c, 350, 329, 258, 329, color, "floor2", "wall")
	ln(c, 395, 311, 350, 311, color, "floor2", "wall")
	ln(c, 398, 129, 315, 129, color, "floor2", "wall")
	ln(c, 176, 133, 315, 133, color, "floor2", "wall")
	ln(c, 176, 129, 96, 129, color, "floor2", "wall")
	ln(c, 3, 133, 96, 133, color, "floor2", "wall")
	ln(c, 66, 387, 66, 339, color, "floor2", "wall")
	ln(c, 115, 387, 115, 339, color, "floor2", "wall")
	ln(c, 163, 387, 163, 339, color, "floor2", "wall")
	ln(c, 234, 275, 276, 275, color, "floor2", "wall")
	ln(c, 288, 275, 309, 275, color, "floor2", "wall")
	ln(c, 298, 275, 298, 329, color, "floor2", "wall")
	ln(c, 341, 283, 350, 283, color, "floor2", "wall")
	ln(c, 321, 275, 341, 275, color, "floor2", "wall")
	ln(c, 375, 275, 395, 275, color, "floor2", "wall")
	ln(c, 315, 129, 315, 170, color, "floor2", "wall")
	ln(c, 376, 170, 307, 170, color, "floor2", "wall")
	ln(c, 307, 250, 307, 170, color, "floor2", "wall")
	ln(c, 376, 245, 376, 170, color, "floor2", "wall")
	ln(c, 340, 241, 307, 241, color, "floor2", "wall")
	ln(c, 340, 245, 340, 224, color, "floor2", "wall")
	ln(c, 340, 210, 340, 201, color, "floor2", "wall")
	ln(c, 340, 187, 340, 170, color, "floor2", "wall")
	ln(c, 340, 206, 307, 206, color, "floor2", "wall")
	ln(c, 293, 250, 307, 250, color, "floor2", "wall")
	ln(c, 271, 179, 238, 179, color, "floor2", "wall")
	ln(c, 226, 179, 195, 179, color, "floor2", "wall")
	ln(c, 176, 129, 176, 179, color, "floor2", "wall")
	ln(c, 182, 179, 176, 179, color, "floor2", "wall")
	ln(c, 174, 169, 176, 169, color, "floor2", "wall")
	ln(c, 162, 169, 90, 169, color, "floor2", "wall")
	ln(c, 96, 169, 96, 129, color, "floor2", "wall")
	ln(c, 175, 227, 90, 227, color, "floor2", "wall")
	ln(c, 90, 190, 90, 227, color, "floor2", "wall")
	ln(c, 52, 179, 3, 179, color, "floor2", "wall")
	ln(c, 52, 228, 3, 228, color, "floor2", "wall")
	ln(c, 52, 276, 3, 276, color, "floor2", "wall")
	ln(c, 155, 177, 155, 169, color, "floor2", "wall")
	ln(c, 110, 191, 110, 169, color, "floor2", "wall")
	ln(c, 155, 189, 155, 197, color, "floor2", "wall")
	ln(c, 350, 283, 350, 329, color, "floor2", "wall")
	ln(c, 162, 197, 155, 197, color, "floor2", "wall")
	ln(c, 341, 275, 341, 283, color, "floor2", "wall")
}

// ============================================================
// Floor 3 foreground: rooms, walls, labels.
// ============================================================

func fg3(c *canvas.Canvas, color string) {
	room(c, []float64{89, 228, 89, 180, 70, 180, 70, 228}, "316", 79.5, 204, color, "floor3")
	room(c, []float64{115, 368, 162, 368, 162, 323, 115, 323}, "309", 138.5, 345.5, color, "floor3")
	room(c, []float64{164, 323, 164, 368, 211, 368, 211, 323}, "308", 187.5, 345.5, color, "floor3")
	room(c, []float64{256, 368, 212, 368, 212, 323, 256, 323}, "307", 234, 345.5, color, "floor3")
	room(c, []float64{244, 276, 297, 276, 297, 327, 260, 327, 260, 321, 244, 321}, "305", 270.5, 301.5, color, "floor3")
	room(c, []float64{251, 219, 251, 203, 244, 203, 244, 219}, "324B", 247.5, 211, color, "floor3")
	room(c, []float64{251, 249, 244, 249, 244, 232, 251, 232}, "324A", 247.5, 240.5, color, "floor3")
	room(c, []float64{223, 135, 223, 179, 177, 179, 177, 135}, "320", 200, 157, color, "floor3")
	room(c, []float64{114, 368, 114, 323, 67, 323, 67, 368}, "310", 90.5, 345.5, color, "floor3")
	room(c, []float64{23, 277, 23, 321, 68, 321, 68, 277}, "312", 45.5, 299, color, "floor3")
	room(c, []float64{23, 229, 68, 229, 68, 275, 23, 275}, "313", 45.5, 252, color, "floor3")
	room(c, []float64{68, 227, 23, 227, 23, 180, 68, 180}, "314", 45.5, 203.5, color, "floor3")
	room(c, []float64{95, 179, 95, 135, 23, 135, 23, 179}, "315", 59, 157, color, "floor3")
	room(c, []float64{99, 226, 99, 204, 91, 204, 91, 226}, "316B", 95, 215, color, "floor3")
	room(c, []float64{91, 202, 99, 202, 99, 180, 91, 180}, "316A", 95, 191, color, "floor3")
	room(c, []float64{97, 169, 109, 169, 109, 192, 154, 192, 154, 198, 174, 198, 174, 226, 101, 226, 101, 179, 97, 179}, "319", 141.5, 209, color, "floor3")
	room(c, []float64{65, 368, 58, 368, 58, 389, 1, 389, 1, 333, 23, 333, 23, 323, 65, 323}, "311", 29.5, 361, color, "floor3")
	room(c, []float64{154, 191, 111, 191, 111, 169, 154, 169}, "318", 132.5, 180, color, "floor3")
	room(c, []float64{175, 168, 97, 168, 97, 131, 175, 131}, "317", 136, 149.5, color, "floor3")
	room(c, []float64{274, 194, 274, 221, 306, 221, 306, 194}, "323", 290, 207.5, color, "floor3")
	room(c, []float64{306, 222, 274, 222, 274, 249, 306, 249}, "325", 290, 235.5, color, "floor3")
	room(c, []float64{263, 179, 224, 179, 224, 135, 263, 135}, "321", 243.5, 157, color, "floor3")
	room(c, []float64{314, 169, 306, 169, 306, 192, 273, 192, 264, 181, 264, 135, 314, 135}, "322", 293.5, 163.5, color, "floor3")
	room(c, []float64{307, 240, 339, 240, 339, 206, 307, 206}, "Pub Lift3", 323, 223, color, "floor3")
	room(c, []float64{339, 205, 307, 205, 307, 171, 339, 171}, "Priv Lift3", 323, 188, color, "floor3")
	room(c, []float64{350, 284, 376, 284, 376, 276, 397, 276, 397, 309, 350, 309}, "303", 373.5, 292.5, color, "floor3")
	room(c, []float64{272, 203, 272, 249, 252, 249, 252, 230, 244, 230, 244, 221, 252, 221, 252, 203}, "324", 262, 226, color, "floor3")
	room(c, []float64{299, 276, 299, 327, 349, 327, 349, 284, 341, 284, 341, 276}, "304", 324, 301.5, color, "floor3")
	room(c, []float64{375, 246, 375, 172, 341, 172, 341, 246}, "301", 358, 209, color, "floor3")
	room(c, []float64{397, 246, 377, 246, 377, 185, 397, 185}, "327", 387, 215.5, color, "floor3")
	room(c, []float64{316, 131, 316, 169, 377, 169, 377, 185, 397, 185, 397, 131}, "326", 356.5, 150, color, "floor3")
	room(c, []float64{308, 251, 242, 251, 242, 274, 342, 274, 342, 282, 375, 282, 375, 274, 397, 274, 397, 248, 339, 248, 339, 242, 308, 242}, "302", 319.5, 261, color, "floor3")
	room(c, []float64{70, 321, 242, 321, 242, 200, 259, 200, 259, 203, 272, 203, 272, 193, 263, 180, 242, 180, 175, 180, 175, 169, 156, 169, 156, 196, 177, 196, 177, 228, 107, 228, 70, 228, 70, 275, 107, 275, 107, 248, 160, 248, 160, 301, 107, 301, 107, 275, 70, 275}, "306", 200.5, 284.5, color, "floor3")

	// Walls.
	ln(c, 341, 275, 341, 283, color, "floor3", "wall")
	ln(c, 162, 197, 155, 197, color, "floor3", "wall")
	ln(c, 396, 247, 399, 247, color, "floor3", "wall")
	ln(c, 399, 129, 399, 311, color, "floor3", "wall")
	ln(c, 258, 202, 243, 202, color, "floor3", "wall")
	ln(c, 350, 283, 350, 329, color, "floor3", "wall")
	ln(c, 251, 231, 243, 231, color, "floor3", "wall")
	ln(c, 243, 220, 251, 220, color, "floor3", "wall")
	ln(c, 243, 250, 243, 202, color, "floor3", "wall")
	ln(c, 155, 197, 155, 190, color, "floor3", "wall")
	ln(c, 110, 192, 110, 169, color, "floor3", "wall")
	ln(c, 155, 192, 110, 192, color, "floor3", "wall")
	ln(c, 155, 177, 155, 169, color, "floor3", "wall")
	ln(c, 176, 197, 176, 227, color, "floor3", "wall")
	ln(c, 69, 280, 69, 274, color, "floor3", "wall")
	ln(c, 21, 276, 69, 276, color, "floor3", "wall")
	ln(c, 69, 262, 69, 226, color, "floor3", "wall")
	ln(c, 21, 228, 69, 228, color, "floor3", "wall")
	ln(c, 21, 179, 75, 179, color, "floor3", "wall")
	ln(c, 69, 179, 69, 214, color, "floor3", "wall")
	ln(c, 90, 220, 90, 227, color, "floor3", "wall")
	ln(c, 90, 204, 90, 202, color, "floor3", "wall")
	ln(c, 90, 203, 100, 203, color, "floor3", "wall")
	ln(c, 90, 187, 90, 179, color, "floor3", "wall")
	ln(c, 90, 227, 176, 227, color, "floor3", "wall")
	ln(c, 100, 179, 100, 227, color, "floor3", "wall")
	ln(c, 100, 179, 87, 179, color, "floor3", "wall")
	ln(c, 96, 179, 96, 129, color, "floor3", "wall")
	ln(c, 162, 169, 96, 169, color, "floor3", "wall")
	ln(c, 173, 169, 176, 169, color, "floor3", "wall")
	ln(c, 182, 179, 176, 179, color, "floor3", "wall")
	ln(c, 176, 129, 176, 179, color, "floor3", "wall")
	ln(c, 195, 179, 226, 179, color, "floor3", "wall")
	ln(c, 224, 133, 224, 179, color, "floor3", "wall")
	ln(c, 264, 179, 264, 133, color, "floor3", "wall")
	ln(c, 238, 179, 264, 179, color, "floor3", "wall")
	ln(c, 273, 207, 273, 193, color, "floor3", "wall")
	ln(c, 273, 235, 273, 250, color, "floor3", "wall")
	ln(c, 273, 224, 273, 219, color, "floor3", "wall")
	ln(c, 273, 193, 307, 193, color, "floor3", "wall")
	ln(c, 273, 222, 307, 222, color, "floor3", "wall")
	ln(c, 273, 250, 307, 250, color, "floor3", "wall")
	ln(c, 384, 247, 376, 247, color, "floor3", "wall")
	ln(c, 340, 206, 307, 206, color, "floor3", "wall")
	ln(c, 340, 187, 340, 170, color, "floor3", "wall")
	ln(c, 340, 210, 340, 201, color, "floor3", "wall")
	ln(c, 340, 247, 340, 224, color, "floor3", "wall")
	ln(c, 340, 241, 307, 241, color, "floor3", "wall")
	ln(c, 376, 247, 376, 170, color, "floor3", "wall")
	ln(c, 307, 250, 307, 170, color, "floor3", "wall")
	ln(c, 376, 170, 307, 170, color, "floor3", "wall")
	ln(c, 315, 129, 315, 170, color, "floor3", "wall")
	ln(c, 376, 283, 366, 283, color, "floor3", "wall")
	ln(c, 376, 283, 376, 275, color, "floor3", "wall")
	ln(c, 399, 275, 376, 275, color, "floor3", "wall")
	ln(c, 341, 275, 320, 275, color, "floor3", "wall")
	ln(c, 341, 283, 350, 283, color, "floor3", "wall")
	ln(c, 298, 275, 298, 329, color, "floor3", "wall")
	ln(c, 308, 275, 298, 275, color, "floor3", "wall")
	ln(c, 243, 322, 243, 275, color, "floor3", "wall")
	ln(c, 243, 275, 284, 275, color, "floor3", "wall")
	ln(c, 258, 322, 226, 322, color, "floor3", "wall")
	ln(c, 212, 370, 212, 322, color, "floor3", "wall")
	ln(c, 214, 322, 177, 322, color, "floor3", "wall")
	ln(c, 163, 370, 163, 322, color, "floor3", "wall")
	ln(c, 165, 322, 129, 322, color, "floor3", "wall")
	ln(c, 84, 322, 117, 322, color, "floor3", "wall")
	ln(c, 71, 322, 64, 322, color, "floor3", "wall")
	ln(c, 115, 322, 115, 370, color, "floor3", "wall")
	ln(c, 66, 322, 66, 370, color, "floor3", "wall")
	ln(c, 52, 322, 21, 322, color, "floor3", "wall")
	ln(c, 21, 331, 0, 331, color, "floor3", "wall")
	ln(c, 21, 331, 21, 133, color, "floor3", "wall")
	ln(c, 96, 133, 21, 133, color, "floor3", "wall")
	ln(c, 176, 129, 96, 129, color, "floor3", "wall")
	ln(c, 315, 133, 176, 133, color, "floor3", "wall")
	ln(c, 315, 129, 399, 129, color, "floor3", "wall")
	ln(c, 399, 311, 350, 311, color, "floor3", "wall")
	ln(c, 350, 329, 258, 329, color, "floor3", "wall")
	ln(c, 258, 322, 258, 370, color, "floor3", "wall")
	ln(c, 60, 370, 258, 370, color, "floor3", "wall")
	ln(c, 60, 370, 60, 391, color, "floor3", "wall")
	ln(c, 0, 391, 0, 331, color, "floor3", "wall")
	ln(c, 60, 391, 0, 391, color, "floor3", "wall")
	ln(c, 307, 250, 307, 242, color, "floor3", "wall")
	ln(c, 273, 250, 307, 250, color, "floor3", "wall")
	ln(c, 258, 250, 243, 250, color, "floor3", "wall")
}
