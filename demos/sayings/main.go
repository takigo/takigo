// Demo: Scrollable sayings listbox with vertical scrollbar.
// Ported from Tk's sayings.tcl demo.
package main

import (
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/listbox"
	"github.com/msorc/takigo/ttk"
)

func main() {
	app := demohelper.Setup("Well-Known Sayings", 500, 350,
		"The listbox below contains a collection of well-known sayings. You can scan the list using either of the scrollbars or by dragging in the listbox window with button 2 pressed.")

	// Sayings data — matches Tk's sayings.tcl exactly.
	sayings := []string{
		"Don't speculate, measure",
		"Waste not, want not",
		"Early to bed and early to rise makes a man healthy, wealthy, and wise",
		"Ask not what your country can do for you, ask what you can do for your country",
		"I shall return",
		"NOT",
		"A picture is worth a thousand words",
		"User interfaces are hard to build",
		"Thou shalt not steal",
		"A penny for your thoughts",
		"Fool me once, shame on you;  fool me twice, shame on me",
		"Every cloud has a silver lining",
		"Where there's smoke there's fire",
		"It takes one to know one",
		"Curiosity killed the cat",
		"Take this job and shove it",
		"Up a creek without a paddle",
		"I'm mad as hell and I'm not going to take it any more",
		"An apple a day keeps the doctor away",
		"Don't look a gift horse in the mouth",
		"Measure twice, cut once",
	}

	// Frame using grid for listbox + scrollbar.
	lbFrame := frame.New(app, "lbframe",
		frame.BorderWidth(10),
	)
	pack.Pack(lbFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX("1c"))

	lb := listbox.New(lbFrame, "sayings",
		listbox.Items(sayings...),
		listbox.Width(20),
		listbox.Height(10),
	)

	yscroll := ttk.NewScrollbar(lbFrame, "yscroll",
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
		yscroll.Set(first, last)
	}

	// Grid layout: listbox row 0 col 0, yscroll row 0 col 1.
	grid.Grid(lb, grid.Row(0), grid.Column(0), grid.Sticky(grid.NSEW))
	grid.Grid(yscroll, grid.Row(0), grid.Column(1), grid.Sticky(grid.NS))
	grid.RowConfigure(lbFrame, 0, grid.Weight(1))
	grid.ColumnConfigure(lbFrame, 0, grid.Weight(1))

	first, last := lb.YVisibleRange()
	yscroll.Set(first, last)

	app.Run()
}
