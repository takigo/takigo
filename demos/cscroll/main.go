// Demo: Scrollable canvas grid.
// Ported from Tk's cscroll.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/scrollbar"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Scrollable Canvas"), takigo.Size(550, 450))
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
		label.Text("A scrollable canvas with a grid of rectangles.\nUse scrollbars to navigate the large canvas."),
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

	// Grid frame for canvas + scrollbars.
	gridFrame := frame.New(root, "gridframe", app)
	pack.Pack(gridFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	c := canvas.New(gridFrame.Window(), "cscroll", app,
		canvas.Background("white"),
		canvas.Width(400),
		canvas.Height(300),
		canvas.ScrollRegion(0, 0, 1200, 900),
	)

	yscroll := scrollbar.New(gridFrame.Window(), "yscroll", app,
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

	xscroll := scrollbar.New(gridFrame.Window(), "xscroll", app,
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
			yscroll.Set(first, last)
		}),
		canvas.XScrollCommand(func(first, last float64) {
			xscroll.Set(first, last)
		}),
	)

	grid.Grid(c.Window(), grid.Row(0), grid.Column(0), grid.Sticky(grid.NSEW))
	grid.Grid(yscroll.Window(), grid.Row(0), grid.Column(1), grid.Sticky(grid.NS))
	grid.Grid(xscroll.Window(), grid.Row(1), grid.Column(0), grid.Sticky(grid.EW))

	grid.RowConfigure(gridFrame.Window(), 0, grid.SlotConfig{Weight: 1})
	grid.ColumnConfigure(gridFrame.Window(), 0, grid.SlotConfig{Weight: 1})

	// Create a grid of colored rectangles.
	colors := []string{
		"#e74c3c", "#3498db", "#2ecc71", "#f39c12",
		"#9b59b6", "#1abc9c", "#e67e22", "#34495e",
	}
	cellSize := 80.0
	gap := 10.0

	for row := range 10 {
		for col := range 14 {
			x1 := float64(col)*(cellSize+gap) + gap
			y1 := float64(row)*(cellSize+gap) + gap
			x2 := x1 + cellSize
			y2 := y1 + cellSize
			colorIdx := (row + col) % len(colors)

			c.CreateRectangle(x1, y1, x2, y2,
				canvas.FillColor(colors[colorIdx]),
				canvas.OutlineColor("black"),
				canvas.OutlineWidth(1))

			c.CreateText((x1+x2)/2, (y1+y2)/2,
				canvas.TextOpt(fmt.Sprintf("%d,%d", col, row)),
				canvas.FontOpt("Sans 8"),
				canvas.AnchorOpt(option.AnchorCenter))
		}
	}

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
