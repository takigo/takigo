// Demo: Canvas text items with interactive configuration and text editing.
// Ported from Tk's ctext.tcl demo.
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
	app, err := takigo.NewApp(takigo.Title("Canvas Text Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("Text"),
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
		label.Text("This window displays a string of text to demonstrate the text facilities of canvas widgets. You can click in the colored boxes to adjust the position of the text relative to its positioning point or change its justification."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f, nil)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	c := canvas.New(f, "canvas",
		canvas.Background("white"),
		canvas.Width(500),
		canvas.Height(350),
	)
	pack.Pack(c, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Red anchor-point marker.
	c.CreateRectangle(245, 163, 255, 173,
		canvas.OutlineColor("black"), canvas.FillColor("red"), canvas.OutlineWidth(1))

	// Main text item — initially anchored at N.
	textID := c.CreateText(250, 168,
		canvas.TextOpt("This is just a string of text to demonstrate the text facilities of canvas widgets. Click the colored boxes to change the anchor or justification."),
		canvas.WidthOpt(440),
		canvas.AnchorOpt(option.AnchorN),
		canvas.FontOpt("Helvetica 16"),
		canvas.JustifyOpt(option.JustifyLeft),
		canvas.Tags("text"),
	)
	textIDStr := fmt.Sprintf("%d", textID)

	// Track box fill colors for hover restore.
	origFill := map[string]string{}

	// mkBox creates a clickable config box at pixel position (px, py).
	// On click, it applies opts to the text item.
	mkBox := func(px, py float64, fill string, opts ...canvas.ItemOption) {
		id := c.CreateRectangle(px, py, px+30, py+30,
			canvas.OutlineColor("black"), canvas.FillColor(fill),
			canvas.OutlineWidth(1), canvas.Tags("config"))
		idStr := fmt.Sprintf("%d", id)
		origFill[idStr] = fill
		c.BindItem(idStr, event.ButtonPressMask, func(ev *event.Event) {
			if ev.Button == 1 {
				c.ItemConfigure(textIDStr, opts...)
			}
		})
	}

	// --- Anchor position selector (3x3 grid at top-left) ---
	// Box positions map to anchor directions.
	anchorColor := "LightSkyBlue1"
	bx, by := 50.0, 50.0

	// Row 0: SE, S, SW
	mkBox(bx, by, anchorColor, canvas.AnchorOpt(option.AnchorSE))
	mkBox(bx+32, by, anchorColor, canvas.AnchorOpt(option.AnchorS))
	mkBox(bx+64, by, anchorColor, canvas.AnchorOpt(option.AnchorSW))
	// Row 1: E, center, W
	mkBox(bx, by+32, anchorColor, canvas.AnchorOpt(option.AnchorE))
	mkBox(bx+32, by+32, anchorColor, canvas.AnchorOpt(option.AnchorCenter))
	mkBox(bx+64, by+32, anchorColor, canvas.AnchorOpt(option.AnchorW))
	// Row 2: NE, N, NW
	mkBox(bx, by+64, anchorColor, canvas.AnchorOpt(option.AnchorNE))
	mkBox(bx+32, by+64, anchorColor, canvas.AnchorOpt(option.AnchorN))
	mkBox(bx+64, by+64, anchorColor, canvas.AnchorOpt(option.AnchorNW))

	c.CreateText(bx+47, by-5,
		canvas.TextOpt("Text Position"),
		canvas.AnchorOpt(option.AnchorS),
		canvas.FontOpt("Times 16"),
		canvas.TextColor("brown"))

	// --- Justification selector (3 boxes in a row) ---
	justColor := "SeaGreen2"
	jx, jy := 350.0, 50.0
	mkBox(jx, jy, justColor, canvas.JustifyOpt(option.JustifyLeft))
	mkBox(jx+32, jy, justColor, canvas.JustifyOpt(option.JustifyCenter))
	mkBox(jx+64, jy, justColor, canvas.JustifyOpt(option.JustifyRight))

	c.CreateText(jx+47, jy-5,
		canvas.TextOpt("Justification"),
		canvas.AnchorOpt(option.AnchorS),
		canvas.FontOpt("Times 16"),
		canvas.TextColor("brown"))

	// --- Angle selector (row of boxes at bottom-left) ---
	angleColor := "Orange1"
	angles := []float64{0, 45, 90, 135, 180, 270}
	ax2, ay2 := 50.0, 270.0
	for i, deg := range angles {
		label := fmt.Sprintf("%.0f°", deg)
		bxPos := ax2 + float64(i)*50
		id := c.CreateRectangle(bxPos, ay2, bxPos+44, ay2+30,
			canvas.OutlineColor("black"), canvas.FillColor(angleColor),
			canvas.OutlineWidth(1), canvas.Tags("config"))
		idStr := fmt.Sprintf("%d", id)
		origFill[idStr] = angleColor
		c.CreateText(bxPos+22, ay2+15,
			canvas.TextOpt(label),
			canvas.AnchorOpt(option.AnchorCenter),
			canvas.FontOpt("Helvetica 11"))
		c.BindItem(idStr, event.ButtonPressMask, func(ev *event.Event) {
			if ev.Button == 1 {
				c.ItemConfigure(textIDStr, canvas.TextAngle(deg))
			}
		})
	}
	c.CreateText(ax2+125, ay2-5,
		canvas.TextOpt("Angle"),
		canvas.AnchorOpt(option.AnchorS),
		canvas.FontOpt("Times 16"),
		canvas.TextColor("brown"))

	// --- Config box hover: darken on Enter, restore on Leave ---
	c.BindItem("config", event.EnterMask, func(ev *event.Event) {
		ids := c.FindWithTag("current")
		if len(ids) == 0 {
			return
		}
		idStr := fmt.Sprintf("%d", ids[0])
		if _, ok := origFill[idStr]; ok {
			c.ItemConfigure(idStr, canvas.FillColor("black"))
		}
	})
	c.BindItem("config", event.LeaveMask, func(ev *event.Event) {
		ids := c.FindWithTag("current")
		if len(ids) == 0 {
			return
		}
		idStr := fmt.Sprintf("%d", ids[0])
		if orig, ok := origFill[idStr]; ok {
			c.ItemConfigure(idStr, canvas.FillColor(orig))
		}
	})

	// --- Click on text item to set keyboard focus and cursor ---
	c.BindItem(textIDStr, event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button == 1 {
			c.Focus(textIDStr)
			c.ICursor(textIDStr, "end")
		}
	})

	// Set initial focus to the text item with cursor at end.
	c.Focus(textIDStr)
	c.ICursor(textIDStr, "end")

	app.Run()
}
