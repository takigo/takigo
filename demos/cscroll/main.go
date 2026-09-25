// Demo: Simple canvas that can be scrolled in two dimensions.
// Ported from Tk's cscroll.tcl demo.
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/scrollbar"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Scrollable Canvas Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("cscroll"),
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
		label.Text("This window displays a canvas widget that can be scrolled by using the scrollbars, by dragging with button 2 in the canvas, by using a mouse wheel, or with the two-finger gesture on a touchpad.  If you click button 1 on one of the rectangles, its indices will be printed on stdout."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Grid frame for canvas + scrollbars.
	gf := frame.New(f, "grid")
	pack.Pack(gf, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(1), pack.PadY(1))

	// Tk uses centimeter coordinates: scrollregion {-11c -11c 50c 20c}
	// 1c ~ 37.8px. We approximate: -416 -416 1890 756.
	c := canvas.New(gf, "c",
		canvas.ReliefOpt(option.ReliefSunken),
		canvas.BorderWidthOpt(2),
		canvas.ScrollRegion(screenunit.Px("-11c"), screenunit.Px("-11c"), screenunit.Px("50c"), screenunit.Px("20c")),
	)

	vscroll := scrollbar.New(gf, "vscroll",
		scrollbar.OrientOpt(scrollbar.Vertical),
		scrollbar.CommandOpt(func(args ...any) {
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

	hscroll := scrollbar.New(gf, "hscroll",
		scrollbar.OrientOpt(scrollbar.Horizontal),
		scrollbar.CommandOpt(func(args ...any) {
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
			vscroll.Set(first, last)
		}),
		canvas.XScrollCommand(func(first, last float64) {
			hscroll.Set(first, last)
		}),
	)

	grid.Grid(c, grid.Row(0), grid.Column(0), grid.Sticky(grid.NSEW), grid.PadX(1), grid.PadY(1))
	grid.Grid(vscroll, grid.Row(0), grid.Column(1), grid.Sticky(grid.NSEW), grid.PadX(1), grid.PadY(1))
	grid.Grid(hscroll, grid.Row(1), grid.Column(0), grid.Sticky(grid.NSEW), grid.PadX(1), grid.PadY(1))

	grid.RowConfigure(gf, 0, grid.Weight(1), grid.MinSize(0))
	grid.ColumnConfigure(gf, 0, grid.Weight(1), grid.MinSize(0))

	// 20x10 grid of 2c rectangles every 3c, filled with the canvas background
	// ([$c config -bg]); coordinates in cm, converted like Tk canvas coords.
	cm := func(v int) float64 { return screenunit.Float(fmt.Sprintf("%dc", v)) }
	bg := widget.DefBackground

	// Track old fill for enter/leave highlighting.
	var oldFill string

	for i := 0; i < 20; i++ {
		xc := -10 + 3*i
		for j := 0; j < 10; j++ {
			yc := -10 + 3*j
			label := fmt.Sprintf("%d,%d", i, j)

			rectID := c.CreateRectangle(cm(xc), cm(yc), cm(xc+2), cm(yc+2),
				canvas.FillColor(bg),
				canvas.Tags("rect"))

			c.CreateText(cm(xc+1), cm(yc+1),
				canvas.TextOpt(label),
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
