// Demo: Canvas text items that can be edited and reconfigured.
// Ported from Tk's ctext.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/canvas"
	"github.com/takigo/takigo/demos/demohelper"
	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/widget/frame"
	"github.com/takigo/takigo/widget/label"
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
		label.WrapLength(screenunit.In(5)),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("This window displays a string of text to demonstrate the text facilities of canvas widgets.  You can click in the boxes to adjust the position of the text relative to its positioning point or change its justification, and on a pie slice to change its angle.  The text also supports the following simple bindings for editing:\n  1. You can point, click, and type.\n  2. You can also select with button 1.\n  3. You can copy the selection to the mouse position with button 2.\n  4. Backspace and Control+h delete the selection if there is one;\n     otherwise they delete the character just before the insertion cursor.\n  5. Delete deletes the selection if there is one; otherwise it deletes\n     the character just after the insertion cursor."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	c := canvas.New(f, "c",
		canvas.ReliefOpt(option.ReliefFlat),
		canvas.BorderWidthOpt(0),
		canvas.Width(500),
		canvas.Height(350),
	)
	pack.Pack(c, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true))

	// Red anchor-point marker.
	c.CreateRectangle(245, 163, 255, 173,
		canvas.OutlineColor("black"), canvas.FillColor("red"))

	// Main text item — initially anchored at N.
	textID := c.CreateText(250, 168,
		canvas.TextOpt("This is just a string of text to demonstrate the text facilities of canvas widgets. Bindings have been defined to support editing (see above)."),
		canvas.WidthOpt(440),
		canvas.AnchorOpt(option.AnchorN),
		canvas.FontOpt("Helvetica 24"),
		canvas.JustifyOpt(option.JustifyLeft),
		canvas.Tags("text"),
	)
	textIDStr := textID

	// Track box fill colors for hover restore.
	origFill := map[canvas.ItemID]string{}

	// mkTextConfigBox creates a clickable config box at pixel position (px, py).
	mkTextConfigBox := func(px, py float64, fill string, opts ...canvas.ItemOption) {
		id := c.CreateRectangle(px, py, px+30, py+30,
			canvas.OutlineColor("black"), canvas.FillColor(fill),
			canvas.OutlineWidth(1), canvas.Tags("config"))
		idStr := id
		origFill[idStr] = fill
		c.BindItem(idStr, event.ButtonPressMask, func(ev *event.Event) {
			if ev.Button == 1 {
				c.ItemConfigure(textIDStr, opts...)
			}
		})
	}

	// mkTextConfigPie creates a clickable pie slice at pixel position (px, py).
	mkTextConfigPie := func(px, py, a float64, fill string, opts ...canvas.ItemOption) {
		id := c.CreateArc(px, py, px+90, py+90,
			canvas.StartAngle(a-15), canvas.Extent(30),
			canvas.OutlineColor("black"), canvas.FillColor(fill),
			canvas.OutlineWidth(1), canvas.Tags("config"))
		idStr := id
		origFill[idStr] = fill
		c.BindItem(idStr, event.ButtonPressMask, func(ev *event.Event) {
			if ev.Button == 1 {
				c.ItemConfigure(textIDStr, opts...)
			}
		})
	}

	// --- Anchor position selector (3x3 grid at top-left) ---
	anchorColor := "LightSkyBlue1"
	bx, by := 50.0, 50.0

	// Row 0: SE, S, SW
	mkTextConfigBox(bx, by, anchorColor, canvas.AnchorOpt(option.AnchorSE))
	mkTextConfigBox(bx+30, by, anchorColor, canvas.AnchorOpt(option.AnchorS))
	mkTextConfigBox(bx+60, by, anchorColor, canvas.AnchorOpt(option.AnchorSW))
	// Row 1: E, center, W
	mkTextConfigBox(bx, by+30, anchorColor, canvas.AnchorOpt(option.AnchorE))
	mkTextConfigBox(bx+30, by+30, anchorColor, canvas.AnchorOpt(option.AnchorCenter))
	mkTextConfigBox(bx+60, by+30, anchorColor, canvas.AnchorOpt(option.AnchorW))
	// Row 2: NE, N, NW
	mkTextConfigBox(bx, by+60, anchorColor, canvas.AnchorOpt(option.AnchorNE))
	mkTextConfigBox(bx+30, by+60, anchorColor, canvas.AnchorOpt(option.AnchorN))
	mkTextConfigBox(bx+60, by+60, anchorColor, canvas.AnchorOpt(option.AnchorNW))

	// Small red center box in anchor grid.
	centerID := c.CreateRectangle(bx+40, by+40, bx+50, by+50,
		canvas.OutlineColor("black"), canvas.FillColor("red"))
	centerIDStr := centerID
	c.BindItem(centerIDStr, event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button == 1 {
			c.ItemConfigure(textIDStr, canvas.AnchorOpt(option.AnchorCenter))
		}
	})

	c.CreateText(bx+45, by-5,
		canvas.TextOpt("Text Position"),
		canvas.AnchorOpt(option.AnchorS),
		canvas.FontOpt("Times 20"),
		canvas.TextColor("brown"))

	// --- Angle selector (pie slices in a circle) ---
	angleColor := "Yellow"
	ax, ay := 205.0, 50.0
	mkTextConfigPie(ax, ay, 0, angleColor, canvas.TextAngle(90))
	mkTextConfigPie(ax, ay, 30, angleColor, canvas.TextAngle(120))
	mkTextConfigPie(ax, ay, 60, angleColor, canvas.TextAngle(150))
	mkTextConfigPie(ax, ay, 90, angleColor, canvas.TextAngle(180))
	mkTextConfigPie(ax, ay, 120, angleColor, canvas.TextAngle(210))
	mkTextConfigPie(ax, ay, 150, angleColor, canvas.TextAngle(240))
	mkTextConfigPie(ax, ay, 180, angleColor, canvas.TextAngle(270))
	mkTextConfigPie(ax, ay, 210, angleColor, canvas.TextAngle(300))
	mkTextConfigPie(ax, ay, 240, angleColor, canvas.TextAngle(330))
	mkTextConfigPie(ax, ay, 270, angleColor, canvas.TextAngle(0))
	mkTextConfigPie(ax, ay, 300, angleColor, canvas.TextAngle(30))
	mkTextConfigPie(ax, ay, 330, angleColor, canvas.TextAngle(60))

	c.CreateText(ax+45, ay-5,
		canvas.TextOpt("Text Angle"),
		canvas.AnchorOpt(option.AnchorS),
		canvas.FontOpt("Times 20"),
		canvas.TextColor("brown"))

	// --- Justification selector (3 boxes in a row) ---
	justColor := "SeaGreen2"
	jx, jy := 350.0, 50.0
	mkTextConfigBox(jx, jy, justColor, canvas.JustifyOpt(option.JustifyLeft))
	mkTextConfigBox(jx+30, jy, justColor, canvas.JustifyOpt(option.JustifyCenter))
	mkTextConfigBox(jx+60, jy, justColor, canvas.JustifyOpt(option.JustifyRight))

	c.CreateText(jx+45, jy-5,
		canvas.TextOpt("Justification"),
		canvas.AnchorOpt(option.AnchorS),
		canvas.FontOpt("Times 20"),
		canvas.TextColor("brown"))

	// --- Config box/pie hover: darken on Enter, restore on Leave ---
	c.BindItem("config", event.EnterMask, func(ev *event.Event) {
		ids := c.FindWithTag("current")
		if len(ids) == 0 {
			return
		}
		idStr := ids[0]
		if _, ok := origFill[idStr]; ok {
			c.ItemConfigure(idStr, canvas.FillColor("black"))
		}
	})
	c.BindItem("config", event.LeaveMask, func(ev *event.Event) {
		ids := c.FindWithTag("current")
		if len(ids) == 0 {
			return
		}
		idStr := ids[0]
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

	app.Run()
}
