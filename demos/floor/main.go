// Demo: Building floorplan drawn on canvas.
// Ported from Tk's floor.tcl demo (simplified layout with room hover highlighting).
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
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Floor Plan Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("Floorplan"),
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
		label.Text("This window contains a canvas widget showing a floorplan. As the mouse moves over the active level, the room under the mouse lights up and its room number appears in the entry."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f, nil)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Status label (bottom) — shows room under cursor.
	statusLabel := label.New(f, "status",
		label.Text(""),
		label.Anchor(option.AnchorW),
		label.PadX(10),
	)
	pack.Pack(statusLabel, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Canvas.
	c := canvas.New(f, "floor",
		canvas.Background("#f5f5dc"),
		canvas.Width(600),
		canvas.Height(380),
	)
	pack.Pack(c, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	wallColor := "#333333"
	roomFill := "#e8e8d0"
	roomHighlight := "#ffffaa"
	doorColor := "#8B4513"

	// Outer walls.
	c.CreateRectangle(20, 20, 580, 360,
		canvas.OutlineColor(wallColor), canvas.OutlineWidth(3),
		canvas.FillColor(roomFill))

	// Room dividers (internal walls).
	c.CreateLine([]float64{20, 180, 350, 180}, canvas.OutlineColor(wallColor), canvas.OutlineWidth(3))
	c.CreateLine([]float64{400, 180, 580, 180}, canvas.OutlineColor(wallColor), canvas.OutlineWidth(3))
	c.CreateLine([]float64{200, 20, 200, 130}, canvas.OutlineColor(wallColor), canvas.OutlineWidth(3))
	c.CreateLine([]float64{200, 180, 200, 360}, canvas.OutlineColor(wallColor), canvas.OutlineWidth(3))
	c.CreateLine([]float64{400, 20, 400, 180}, canvas.OutlineColor(wallColor), canvas.OutlineWidth(3))
	c.CreateLine([]float64{400, 230, 400, 360}, canvas.OutlineColor(wallColor), canvas.OutlineWidth(3))

	// Doors.
	c.CreateRectangle(195, 130, 205, 180, canvas.FillColor(doorColor), canvas.OutlineColor(doorColor))
	c.CreateRectangle(350, 175, 400, 185, canvas.FillColor(doorColor), canvas.OutlineColor(doorColor))
	c.CreateRectangle(395, 230, 405, 270, canvas.FillColor(doorColor), canvas.OutlineColor(doorColor))

	// Furniture.
	c.CreateRectangle(260, 70, 340, 130,
		canvas.FillColor("#b8860b"), canvas.OutlineColor("#8b6914"), canvas.OutlineWidth(1))
	c.CreateRectangle(50, 50, 100, 80,
		canvas.FillColor("#cd853f"), canvas.OutlineColor("#8b5e3c"), canvas.OutlineWidth(1))
	c.CreateRectangle(460, 50, 540, 80,
		canvas.FillColor("#cd853f"), canvas.OutlineColor("#8b5e3c"), canvas.OutlineWidth(1))
	c.CreateRectangle(460, 250, 540, 280,
		canvas.FillColor("#cd853f"), canvas.OutlineColor("#8b5e3c"), canvas.OutlineWidth(1))

	// Define rooms: transparent overlay rectangles with "room" tag for hover detection.
	// These sit on top and catch mouse events; FillNone makes them transparent.
	type roomDef struct {
		x1, y1, x2, y2 float64
		name            string
	}
	rooms := []roomDef{
		{22, 22, 198, 178, "Office A"},
		{202, 22, 398, 178, "Conference Room"},
		{402, 22, 578, 178, "Office B"},
		{22, 182, 198, 358, "Kitchen"},
		{202, 182, 398, 358, "Lobby"},
		{402, 182, 578, 358, "Office C"},
	}

	// origFill stores the fill color before highlighting.
	origFill := map[string]string{}

	for _, r := range rooms {
		id := c.CreateRectangle(r.x1, r.y1, r.x2, r.y2,
			canvas.FillColor(roomFill),
			canvas.OutlineColor(""),
			canvas.OutlineWidth(0),
			canvas.Tags("room"),
		)
		idStr := fmt.Sprintf("%d", id)
		origFill[idStr] = roomFill
		name := r.name // capture for closure
		c.BindItem(idStr, event.EnterMask, func(ev *event.Event) {
			c.ItemConfigure(idStr, canvas.FillColor(roomHighlight))
			statusLabel.Text = name
			statusLabel.Display()
		})
		c.BindItem(idStr, event.LeaveMask, func(ev *event.Event) {
			c.ItemConfigure(idStr, canvas.FillColor(origFill[idStr]))
			statusLabel.Text = ""
			statusLabel.Display()
		})
	}

	// Room labels (on top of room overlays).
	labelDefs := []struct {
		x, y float64
		name string
	}{
		{110, 100, "Office A"},
		{300, 100, "Conference\nRoom"},
		{490, 100, "Office B"},
		{110, 270, "Kitchen"},
		{300, 270, "Lobby"},
		{490, 270, "Office C"},
	}
	for _, r := range labelDefs {
		c.CreateText(r.x, r.y,
			canvas.TextOpt(r.name),
			canvas.FontOpt("Sans 10"),
			canvas.TextColor("#555555"),
			canvas.AnchorOpt(option.AnchorCenter))
	}

	// Title.
	c.CreateText(300, 375,
		canvas.TextOpt("Floor Plan — Ground Level"),
		canvas.FontOpt("Sans Bold 11"),
		canvas.TextColor("#333333"),
		canvas.AnchorOpt(option.AnchorS))

	app.Run()
}
