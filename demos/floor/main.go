// Demo: Building floorplan drawn on canvas.
// Ported from Tk's floor.tcl demo (simplified — basic floor layout).
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Floor Plan"), takigo.Size(650, 500))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	// Description.
	msg := label.New(root, "msg", app,
		label.Text("A simplified building floor plan drawn with canvas lines\nand polygons. Rooms are labeled with text items."),
		label.Anchor(option.AnchorW),
		label.PadX(10),
		label.PadY(5),
	)
	pack.Pack(msg.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// Dismiss button.
	btnFrame := frame.New(root, "btnframe", app)
	pack.Pack(btnFrame.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))

	dismissBtn := button.New(btnFrame.Window(), "dismiss", app,
		button.Text("Dismiss"),
		button.Command(func() { app.Quit() }),
		button.PadX(10),
		button.PadY(4),
	)
	pack.Pack(dismissBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(10))

	// Canvas.
	c := canvas.New(root, "floor", app,
		canvas.Background("#f5f5dc"),
		canvas.Width(600),
		canvas.Height(380),
	)
	pack.Pack(c.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	wallColor := "#333333"
	wallWidth := 3
	roomFill := "#e8e8d0"
	doorColor := "#8B4513"

	// Outer walls.
	c.CreateRectangle(20, 20, 580, 360,
		canvas.OutlineColor(wallColor), canvas.OutlineWidth(wallWidth),
		canvas.FillColor(roomFill))

	// Room dividers (internal walls).
	// Horizontal walls.
	c.CreateLine([]float64{20, 180, 350, 180},
		canvas.OutlineColor(wallColor), canvas.OutlineWidth(wallWidth))
	c.CreateLine([]float64{400, 180, 580, 180},
		canvas.OutlineColor(wallColor), canvas.OutlineWidth(wallWidth))

	// Vertical walls.
	c.CreateLine([]float64{200, 20, 200, 130},
		canvas.OutlineColor(wallColor), canvas.OutlineWidth(wallWidth))
	c.CreateLine([]float64{200, 180, 200, 360},
		canvas.OutlineColor(wallColor), canvas.OutlineWidth(wallWidth))
	c.CreateLine([]float64{400, 20, 400, 180},
		canvas.OutlineColor(wallColor), canvas.OutlineWidth(wallWidth))
	c.CreateLine([]float64{400, 230, 400, 360},
		canvas.OutlineColor(wallColor), canvas.OutlineWidth(wallWidth))

	// Doors (small gaps represented by colored rectangles).
	c.CreateRectangle(195, 130, 205, 180,
		canvas.FillColor(doorColor), canvas.OutlineColor(doorColor))
	c.CreateRectangle(350, 175, 400, 185,
		canvas.FillColor(doorColor), canvas.OutlineColor(doorColor))
	c.CreateRectangle(395, 230, 405, 270,
		canvas.FillColor(doorColor), canvas.OutlineColor(doorColor))

	// Room labels.
	rooms := []struct {
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

	for _, r := range rooms {
		c.CreateText(r.x, r.y,
			canvas.TextOpt(r.name),
			canvas.FontOpt("Sans 10"),
			canvas.TextColor("#555555"),
			canvas.AnchorOpt(option.AnchorCenter))
	}

	// Furniture (simple rectangles).
	// Conference table.
	c.CreateRectangle(260, 70, 340, 130,
		canvas.FillColor("#b8860b"), canvas.OutlineColor("#8b6914"), canvas.OutlineWidth(1))

	// Desks in offices.
	c.CreateRectangle(50, 50, 100, 80,
		canvas.FillColor("#cd853f"), canvas.OutlineColor("#8b5e3c"), canvas.OutlineWidth(1))
	c.CreateRectangle(460, 50, 540, 80,
		canvas.FillColor("#cd853f"), canvas.OutlineColor("#8b5e3c"), canvas.OutlineWidth(1))
	c.CreateRectangle(460, 250, 540, 280,
		canvas.FillColor("#cd853f"), canvas.OutlineColor("#8b5e3c"), canvas.OutlineWidth(1))

	// Title.
	c.CreateText(300, 375,
		canvas.TextOpt("Floor Plan — Ground Level"),
		canvas.FontOpt("Sans Bold 11"),
		canvas.TextColor("#333333"),
		canvas.AnchorOpt(option.AnchorS))

	// Root event handlers.
	app.Dispatcher().Bind(root.XWindow, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			root.Width = ev.ConfigWidth
			root.Height = ev.ConfigHeight
			pack.ArrangeContainer(root)
		}
	})

	app.Dispatcher().Bind(root.XWindow, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		d := root.Display.XDisplay
		gc := root.GC
		d.SetForeground(gc, bgColor.Pixel)
		d.FillRectangle(root.Drawable(), gc, 0, 0, uint(root.Width), uint(root.Height))
		d.Flush()
	})

	app.Dispatcher().BindGlobal(event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == xlib.XK_Escape {
			app.Quit()
		}
	})

	_ = msg
	_ = dismissBtn
	app.MainLoop()
}
