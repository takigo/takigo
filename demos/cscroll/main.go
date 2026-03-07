// Demo: Scrollable canvas grid.
// Ported from Tk's cscroll.tcl demo.
package main

import (
	"fmt"
	"strconv"

	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/ttk"
)

func main() {
	app := demohelper.Setup("Scrollable Canvas Demonstration", 550, 450,
		"This window displays a canvas widget that can be scrolled by using the scrollbars or by dragging with button 2 in the canvas. If you click button 1 on one of the rectangles, its indices will be printed on stdout.")

	// Grid frame for canvas + scrollbars.
	gridFrame := frame.New(app, "gridframe")
	pack.Pack(gridFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true))

	// Tk uses centimeter coordinates: scrollregion {-11c -11c 50c 20c}
	// 1c ~ 37.8px. We approximate: -416 -416 1890 756.
	c := canvas.New(gridFrame, "cscroll",
		canvas.Background("white"),
		canvas.Width(400),
		canvas.Height(300),
		canvas.ScrollRegion(-416, -416, 1890, 756),
	)

	yscroll := ttk.NewScrollbar(gridFrame, "yscroll",
		ttk.ScrollbarOrientOpt(ttk.Vertical),
		ttk.ScrollbarCommandOpt(func(args ...any) {
			if len(args) < 1 {
				return
			}
			switch args[0] {
			case "moveto":
				if len(args) >= 2 {
					if f, ok := args[1].(float64); ok {
						c.YViewMoveTo(f)
					}
				}
			case "scroll":
				if len(args) >= 3 {
					n, _ := args[1].(int)
					unit, _ := args[2].(string)
					c.YViewScroll(n, unit == "pages")
				}
			}
		}),
	)

	xscroll := ttk.NewScrollbar(gridFrame, "xscroll",
		ttk.ScrollbarOrientOpt(ttk.Horizontal),
		ttk.ScrollbarCommandOpt(func(args ...any) {
			if len(args) < 1 {
				return
			}
			switch args[0] {
			case "moveto":
				if len(args) >= 2 {
					if f, ok := args[1].(float64); ok {
						c.XViewMoveTo(f)
					}
				}
			case "scroll":
				if len(args) >= 3 {
					n, _ := args[1].(int)
					unit, _ := args[2].(string)
					c.XViewScroll(n, unit == "pages")
				}
			}
		}),
	)

	c.Configure(
		canvas.YScrollCommand(func(first, last float64) {
			yscroll.Set(first, last)
		}),
		canvas.XScrollCommand(func(first, last float64) {
			xscroll.Set(first, last)
		}),
	)

	grid.Grid(c, grid.Row(0), grid.Column(0), grid.Sticky(grid.NSEW))
	grid.Grid(yscroll, grid.Row(0), grid.Column(1), grid.Sticky(grid.NS))
	grid.Grid(xscroll, grid.Row(1), grid.Column(0), grid.Sticky(grid.EW))

	grid.RowConfigure(gridFrame, 0, grid.Weight(1))
	grid.ColumnConfigure(gridFrame, 0, grid.Weight(1))

	// Create a 20x10 grid of rectangles, matching Tk's cscroll.tcl.
	// Tk uses centimeter units: each cell is 2c x 2c with 3c spacing.
	// 1c ~ 37.8px, so 2c ~ 75.6px, 3c ~ 113.4px.
	cellPx := 75.6     // 2c in pixels
	spacingPx := 113.4  // 3c in pixels
	startX := -378.0    // -10c in pixels
	startY := -378.0    // -10c in pixels
	bg := "white"       // canvas background color, used as default fill

	// Track old fill for enter/leave highlighting.
	var oldFill string

	for i := 0; i < 20; i++ {
		x := startX + spacingPx*float64(i)
		for j := 0; j < 10; j++ {
			y := startY + spacingPx*float64(j)
			label := fmt.Sprintf("%d,%d", i, j)

			rectID := c.CreateRectangle(x, y, x+cellPx, y+cellPx,
				canvas.FillColor(bg),
				canvas.OutlineColor("black"),
				canvas.OutlineWidth(1),
				canvas.Tags("rect"))

			c.CreateText(x+cellPx/2, y+cellPx/2,
				canvas.TextOpt(label),
				canvas.FontOpt("Sans 9"),
				canvas.AnchorOpt(option.AnchorCenter),
				canvas.Tags("text"))

			// Bind Enter: highlight rectangle with LightSeaGreen.
			rectIDStr := strconv.FormatInt(rectID, 10)
			c.BindItem(rectIDStr, event.EnterMask, func(ev *event.Event) {
				oldFill = bg
				c.ItemConfigure(rectIDStr, canvas.FillColor("LightSeaGreen"))
			})

			// Bind Leave: restore original fill.
			c.BindItem(rectIDStr, event.LeaveMask, func(ev *event.Event) {
				c.ItemConfigure(rectIDStr, canvas.FillColor(oldFill))
			})

			// Bind Button-1: print the cell index.
			c.BindItem(rectIDStr, event.ButtonPressMask, func(ev *event.Event) {
				fmt.Printf("You buttoned at %s\n", label)
			})

			// Also bind the text item so clicking on text works too.
			textIDStr := strconv.FormatInt(rectID+1, 10)
			c.BindItem(textIDStr, event.ButtonPressMask, func(ev *event.Event) {
				fmt.Printf("You buttoned at %s\n", label)
			})
			c.BindItem(textIDStr, event.EnterMask, func(ev *event.Event) {
				oldFill = bg
				c.ItemConfigure(rectIDStr, canvas.FillColor("LightSeaGreen"))
			})
			c.BindItem(textIDStr, event.LeaveMask, func(ev *event.Event) {
				c.ItemConfigure(rectIDStr, canvas.FillColor(oldFill))
			})
		}
	}

	app.Run()
}
