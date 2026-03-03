// Demo: All canvas item types showcase.
// Ported from Tk's items.tcl demo.
package main

import (
	"fmt"
	"math"

	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
)

func main() {
	app := demohelper.Setup("Canvas Item Demonstration", 700, 550,
		"This window contains a canvas widget with examples of the various kinds of items supported by canvases. The following operations are supported:\n  Left-button drag: moves item under pointer.\n  Middle-button drag: repositions view.\n  Right-button drag: strokes out area.")

	// Canvas.
	c := canvas.New(app, "items",
		canvas.Background("white"),
		canvas.Width(660),
		canvas.Height(420),
	)
	pack.Pack(c, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Track original colors per item ID for hover restore.
	type itemColors struct {
		fill    string // fill color (or text color for text items)
		outline string // outline color
		isText  bool   // true for text items (use TextColor instead of FillColor)
	}
	origColors := map[string]itemColors{}

	// Helper to record original colors for a shape item.
	record := func(id int64, fill, outline string) {
		origColors[fmt.Sprintf("%d", id)] = itemColors{fill: fill, outline: outline}
	}

	// Helper to record original colors for a text item.
	recordText := func(id int64, textColor string) {
		origColors[fmt.Sprintf("%d", id)] = itemColors{fill: textColor, isText: true}
	}

	// Section 1: Rectangles.
	c.CreateText(110, 15, canvas.TextOpt("Rectangles"), canvas.FontOpt("Sans Bold 11"),
		canvas.AnchorOpt(option.AnchorCenter))

	record(c.CreateRectangle(20, 30, 100, 90,
		canvas.FillColor("#4a86c8"), canvas.OutlineColor("black"), canvas.OutlineWidth(2),
		canvas.Tags("item")), "#4a86c8", "black")

	record(c.CreateRectangle(120, 30, 200, 90,
		canvas.FillColor("#e8a835"), canvas.OutlineColor("black"), canvas.OutlineWidth(1),
		canvas.Tags("item")), "#e8a835", "black")

	record(c.CreateRectangle(20, 100, 200, 130,
		canvas.FillColor("#6bb86b"), canvas.OutlineColor("darkgreen"), canvas.OutlineWidth(3),
		canvas.Tags("item")), "#6bb86b", "darkgreen")

	// Section 2: Ovals.
	c.CreateText(330, 15, canvas.TextOpt("Ovals"), canvas.FontOpt("Sans Bold 11"),
		canvas.AnchorOpt(option.AnchorCenter))

	record(c.CreateOval(240, 30, 340, 90,
		canvas.FillColor("#c85a5a"), canvas.OutlineColor("black"), canvas.OutlineWidth(2),
		canvas.Tags("item")), "#c85a5a", "black")

	record(c.CreateOval(360, 30, 420, 130,
		canvas.FillColor("#9b59b6"), canvas.OutlineColor("black"), canvas.OutlineWidth(1),
		canvas.Tags("item")), "#9b59b6", "black")

	// Section 3: Lines.
	c.CreateText(550, 15, canvas.TextOpt("Lines"), canvas.FontOpt("Sans Bold 11"),
		canvas.AnchorOpt(option.AnchorCenter))

	record(c.CreateLine([]float64{460, 30, 640, 30},
		canvas.OutlineColor("red"), canvas.OutlineWidth(3),
		canvas.Tags("item")), "", "red")

	record(c.CreateLine([]float64{460, 50, 550, 90, 640, 50},
		canvas.OutlineColor("blue"), canvas.OutlineWidth(2),
		canvas.Tags("item")), "", "blue")

	record(c.CreateLine([]float64{460, 110, 550, 70, 640, 110},
		canvas.OutlineColor("darkgreen"), canvas.OutlineWidth(2), canvas.Smooth(true),
		canvas.Tags("item")), "", "darkgreen")

	record(c.CreateLine([]float64{460, 130, 640, 130},
		canvas.OutlineColor("black"), canvas.OutlineWidth(2),
		canvas.Arrow(canvas.ArrowBoth),
		canvas.Tags("item")), "", "black")

	// Section 4: Polygons.
	c.CreateText(110, 155, canvas.TextOpt("Polygons"), canvas.FontOpt("Sans Bold 11"),
		canvas.AnchorOpt(option.AnchorCenter))

	// Triangle.
	record(c.CreatePolygon([]float64{60, 170, 20, 260, 100, 260},
		canvas.FillColor("#e74c3c"), canvas.OutlineColor("black"), canvas.OutlineWidth(2),
		canvas.Tags("item")), "#e74c3c", "black")

	// Pentagon.
	cx, cy, r := 160.0, 220.0, 40.0
	penta := make([]float64, 10)
	for i := range 5 {
		angle := math.Pi/2 + float64(i)*2*math.Pi/5
		penta[i*2] = cx + r*math.Cos(angle)
		penta[i*2+1] = cy - r*math.Sin(angle)
	}
	record(c.CreatePolygon(penta,
		canvas.FillColor("#3498db"), canvas.OutlineColor("navy"), canvas.OutlineWidth(2),
		canvas.Tags("item")), "#3498db", "navy")

	// Section 5: Arcs.
	c.CreateText(330, 155, canvas.TextOpt("Arcs"), canvas.FontOpt("Sans Bold 11"),
		canvas.AnchorOpt(option.AnchorCenter))

	record(c.CreateArc(240, 170, 340, 270,
		canvas.FillColor("#f39c12"), canvas.OutlineColor("black"), canvas.OutlineWidth(2),
		canvas.StartAngle(0), canvas.Extent(120), canvas.ArcStyleOpt(canvas.ArcStylePieslice),
		canvas.Tags("item")), "#f39c12", "black")

	record(c.CreateArc(350, 170, 430, 270,
		canvas.OutlineColor("#c0392b"), canvas.OutlineWidth(3),
		canvas.StartAngle(30), canvas.Extent(270), canvas.ArcStyleOpt(canvas.ArcStyleArc),
		canvas.Tags("item")), "", "#c0392b")

	// Section 6: Text.
	c.CreateText(550, 155, canvas.TextOpt("Text Items"), canvas.FontOpt("Sans Bold 11"),
		canvas.AnchorOpt(option.AnchorCenter))

	recordText(c.CreateText(550, 190, canvas.TextOpt("Default font"),
		canvas.TextColor("black"), canvas.AnchorOpt(option.AnchorCenter),
		canvas.Tags("item")), "black")

	recordText(c.CreateText(550, 220, canvas.TextOpt("Bold text"),
		canvas.FontOpt("Sans Bold 14"), canvas.TextColor("navy"),
		canvas.AnchorOpt(option.AnchorCenter),
		canvas.Tags("item")), "navy")

	recordText(c.CreateText(550, 250, canvas.TextOpt("Italic text"),
		canvas.FontOpt("Sans Italic 12"), canvas.TextColor("darkred"),
		canvas.AnchorOpt(option.AnchorCenter),
		canvas.Tags("item")), "darkred")

	// Section 7: Dashed lines.
	c.CreateText(330, 290, canvas.TextOpt("Dashed Lines"), canvas.FontOpt("Sans Bold 11"),
		canvas.AnchorOpt(option.AnchorCenter))

	record(c.CreateLine([]float64{20, 310, 300, 310},
		canvas.OutlineColor("black"), canvas.OutlineWidth(2), canvas.Dash(6, 4),
		canvas.Tags("item")), "", "black")

	record(c.CreateLine([]float64{20, 330, 300, 330},
		canvas.OutlineColor("blue"), canvas.OutlineWidth(2), canvas.Dash(12, 4, 4, 4),
		canvas.Tags("item")), "", "blue")

	record(c.CreateRectangle(340, 300, 640, 380,
		canvas.OutlineColor("gray50"), canvas.OutlineWidth(2), canvas.Dash(8, 4),
		canvas.Tags("item")), "", "gray50")

	record(c.CreateOval(380, 305, 600, 375,
		canvas.FillColor("#eaf2f8"), canvas.OutlineColor("#2980b9"), canvas.OutlineWidth(2),
		canvas.Tags("item")), "#eaf2f8", "#2980b9")

	// --- Event bindings for item interaction (matching Tk items.tcl) ---

	// Track which item is currently highlighted for restore on Leave.
	var highlightedID string

	// Hover: highlight item on Enter, restore original colors on Leave.
	c.BindItem("item", event.EnterMask, func(ev *event.Event) {
		highlightedID = ""
		ids := c.FindWithTag("current")
		if len(ids) == 0 {
			return
		}
		idStr := fmt.Sprintf("%d", ids[0])
		colors, ok := origColors[idStr]
		if !ok {
			return
		}
		highlightedID = idStr

		if colors.isText {
			// Text items: change text color.
			c.ItemConfigure("current", canvas.TextColor("SteelBlue2"))
		} else if colors.fill != "" {
			// Filled shapes: change fill color.
			c.ItemConfigure("current", canvas.FillColor("SteelBlue2"))
		} else if colors.outline != "" {
			// Unfilled shapes / lines: change outline color.
			c.ItemConfigure("current", canvas.OutlineColor("SteelBlue2"))
		}
	})

	c.BindItem("item", event.LeaveMask, func(ev *event.Event) {
		if highlightedID == "" {
			return
		}
		colors, ok := origColors[highlightedID]
		if !ok {
			return
		}
		// Restore original colors.
		if colors.isText {
			c.ItemConfigure(highlightedID, canvas.TextColor(colors.fill))
		} else {
			if colors.fill != "" {
				c.ItemConfigure(highlightedID, canvas.FillColor(colors.fill))
			}
			if colors.outline != "" {
				c.ItemConfigure(highlightedID, canvas.OutlineColor(colors.outline))
			}
		}
		highlightedID = ""
	})

	// Drag state for Button-1 item dragging.
	var lastX, lastY int

	// ButtonPress-1 on item: start drag.
	c.BindItem("item", event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button != 1 {
			return
		}
		lastX = ev.X
		lastY = ev.Y
	})

	// B1-Motion on item: drag the item under pointer.
	c.BindItem("item", event.MotionMask, func(ev *event.Event) {
		if ev.State&xlib.Button1Mask == 0 {
			return
		}
		dx := float64(ev.X - lastX)
		dy := float64(ev.Y - lastY)
		c.Move("current", dx, dy)
		lastX = ev.X
		lastY = ev.Y
	})

	app.Run()
}
