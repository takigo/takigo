// Demo: Vertical paned window with text and listbox.
// Ported from Tk's paned2.tcl demo.
package main

import (
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/listbox"
	"github.com/msorc/takigo/widget/panedwindow"
	"github.com/msorc/takigo/widget/scrollbar"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app := demohelper.Setup("Vertical Paned Window Demonstration", 500, 450,
		"The sash between the two scrolled windows below can be used to divide the area between them. Use the left mouse button to resize by moving the sash.")

	// Vertical paned window.
	pw := panedwindow.New(app, "vpanes",
		panedwindow.OrientOpt(panedwindow.Vertical),
	)
	pack.Pack(pw, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX("2m"), pack.PadY("1.5p"))

	// Top pane: listbox with vertical scrollbar.
	topFrame := frame.New(pw, "top")
	lb := listbox.New(topFrame, "widgetlist",
		listbox.Items(
			"List of Tk Widgets",
			"button",
			"canvas",
			"checkbutton",
			"entry",
			"frame",
			"label",
			"labelframe",
			"listbox",
			"menu",
			"menubutton",
			"message",
			"panedwindow",
			"radiobutton",
			"scale",
			"scrollbar",
			"spinbox",
			"text",
			"toplevel",
		),
	)

	lbScroll := scrollbar.New(topFrame, "lbscroll",
		scrollbar.OrientOpt(scrollbar.Vertical),
		scrollbar.CommandOpt(func(args ...any) {
			if len(args) < 1 {
				return
			}
			switch args[0] {
			case "moveto":
				if len(args) >= 2 {
					if f, ok := args[1].(float64); ok {
						lb.YViewMoveTo(f)
					}
				}
			case "scroll":
				if len(args) >= 3 {
					n, _ := args[1].(int)
					unit, _ := args[2].(string)
					lb.YViewScroll(n, unit == "pages")
				}
			}
		}),
	)
	lb.YScrollCmd = func(first, last float64) {
		lbScroll.Set(first, last)
	}

	pack.Pack(lbScroll, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(lb, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))
	pw.Add(topFrame.Window(), 100)

	// Bottom pane: text widget with x+y scrollbars (grid layout).
	bottomFrame := frame.New(pw, "bottom")

	tw := text.New(bottomFrame, "text",
		text.Width(30),
		text.Height(8),
		text.WrapModeOpt(text.WrapNone),
	)

	yscroll := scrollbar.New(bottomFrame, "yscroll",
		scrollbar.OrientOpt(scrollbar.Vertical),
		scrollbar.CommandOpt(func(args ...any) {
			if len(args) < 1 {
				return
			}
			switch args[0] {
			case "moveto":
				if len(args) >= 2 {
					if f, ok := args[1].(float64); ok {
						tw.YViewMoveTo(f)
					}
				}
			case "scroll":
				if len(args) >= 3 {
					n, _ := args[1].(int)
					unit, _ := args[2].(string)
					tw.YViewScroll(n, unit == "pages")
				}
			}
		}),
	)
	tw.YScrollCmd = func(first, last float64) {
		yscroll.Set(first, last)
	}

	xscroll := scrollbar.New(bottomFrame, "xscroll",
		scrollbar.OrientOpt(scrollbar.Horizontal),
		scrollbar.CommandOpt(func(args ...any) {
			if len(args) < 1 {
				return
			}
			switch args[0] {
			case "moveto":
				if len(args) >= 2 {
					if f, ok := args[1].(float64); ok {
						tw.XViewMoveTo(f)
					}
				}
			case "scroll":
				if len(args) >= 3 {
					n, _ := args[1].(int)
					unit, _ := args[2].(string)
					tw.XViewScroll(n, unit == "pages")
				}
			}
		}),
	)
	tw.XScrollCmd = func(first, last float64) {
		xscroll.Set(first, last)
	}

	// Grid: text at (0,0), yscroll at (0,1), xscroll at (1,0).
	grid.Grid(tw, grid.Row(0), grid.Column(0), grid.Sticky(grid.NSEW))
	grid.Grid(yscroll, grid.Row(0), grid.Column(1), grid.Sticky(grid.NSEW))
	grid.Grid(xscroll, grid.Row(1), grid.Column(0), grid.Sticky(grid.NSEW))
	grid.ColumnConfigure(bottomFrame.Window(), 0, grid.SlotConfig{Weight: 1})
	grid.RowConfigure(bottomFrame.Window(), 0, grid.SlotConfig{Weight: 1})

	tw.Insert("1.0", "This is just a normal text widget")

	pw.Add(bottomFrame.Window(), 100)

	app.Run()
}
