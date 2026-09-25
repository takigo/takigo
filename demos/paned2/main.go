// Demo: Paned window that separates two windows vertically.
// Ported from Tk's paned2.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/listbox"
	"github.com/msorc/takigo/widget/panedwindow"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Vertical Paned Window Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("paned2"),
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
		label.Text("The sash between the two scrolled windows below can be used to divide the area between them.  Use the left mouse button to resize without redrawing by just moving the sash, and use the middle mouse button to resize opaquely (always redrawing the windows in each position.)"),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Vertical paned window.
	pw := panedwindow.New(f, "pane",
		panedwindow.OrientOpt(panedwindow.Vertical),
	)
	pack.Pack(pw, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX("2m"), pack.PadY("1.5p"))

	// Top pane: listbox with vertical scrollbar.
	topFrame := frame.New(pw, "top")
	lb := listbox.New(topFrame, "list",
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

	lbScroll := ttk.NewScrollbar(topFrame, "scr",
		ttk.ScrollbarOrientOpt(ttk.Vertical),
		ttk.ScrollbarCommandOpt(func(args ...any) {
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

	// Invert first item to highlight it (matches Tcl's itemconfigure 0 -bg fg -fg bg).
	lb.ItemConfigure(0, "#d9d9d9", "#000000")

	pack.Pack(lbScroll, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(lb, pack.FillOpt(pack.FillBoth), pack.Expand(true))
	pw.Add(topFrame.Window(), 0)
	pw.SetStretch(topFrame.Window(), panedwindow.StretchAlways)

	// Bottom pane: text widget with x+y scrollbars (grid layout).
	bottomFrame := frame.New(pw, "bottom")

	tw := text.New(bottomFrame, "text",
		text.Width(30),
		text.Height(8),
		text.WrapModeOpt(text.WrapNone),
	)

	yscroll := ttk.NewScrollbar(bottomFrame, "yscr",
		ttk.ScrollbarOrientOpt(ttk.Vertical),
		ttk.ScrollbarCommandOpt(func(args ...any) {
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

	xscroll := ttk.NewScrollbar(bottomFrame, "xscr",
		ttk.ScrollbarOrientOpt(ttk.Horizontal),
		ttk.ScrollbarCommandOpt(func(args ...any) {
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
	grid.ColumnConfigure(bottomFrame, 0, grid.Weight(1))
	grid.RowConfigure(bottomFrame, 0, grid.Weight(1))

	tw.Insert("1.0", "This is just a normal text widget")

	pw.Add(bottomFrame.Window(), 0)
	pw.SetStretch(bottomFrame.Window(), panedwindow.StretchAlways)

	app.Run()
}
