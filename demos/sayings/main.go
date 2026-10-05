// Demo: Listbox that can be scrolled both horizontally and vertically,
// displaying a collection of well-known sayings.
// Ported from Tk's sayings.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/demos/demohelper"
	"github.com/takigo/takigo/geometry/grid"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/ttk"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/frame"
	"github.com/takigo/takigo/widget/label"
	"github.com/takigo/takigo/widget/listbox"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Listbox Demonstration (well-known sayings)"),
		takigo.Geometry("+300+300"),
		takigo.IconName("sayings"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength(screenunit.In(4)),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("The listbox below contains a collection of well-known sayings.  You can scan the list using either of the scrollbars or by dragging in the listbox window with button 2 pressed."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

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
	lbFrame := frame.New(f, "frame",
		frame.BorderWidth(10), // 7.5p ≈ 10px
	)
	pack.Pack(lbFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(screenunit.Cm(1)))

	lb := listbox.New(lbFrame, "list",
		listbox.Items(sayings...),
		listbox.Width(20),
		listbox.Height(10),
	)

	yscroll := ttk.NewScrollbar(lbFrame, "yscroll",
		ttk.ScrollbarOrientOpt(ttk.Vertical),
		ttk.ScrollbarCommandOpt(widget.ScrollY(lb)),
	)
	lb.YScrollCmd = func(first, last float64) {
		yscroll.Set(first, last)
	}

	xscroll := ttk.NewScrollbar(lbFrame, "xscroll",
		ttk.ScrollbarOrientOpt(ttk.Horizontal),
		ttk.ScrollbarCommandOpt(widget.ScrollX(lb)),
	)
	lb.XScrollCmd = func(first, last float64) {
		xscroll.Set(first, last)
	}

	grid.Grid(lb, grid.Row(0), grid.Column(0), grid.Sticky(grid.NSEW))
	grid.Grid(yscroll, grid.Row(0), grid.Column(1), grid.Sticky(grid.NSEW))
	grid.Grid(xscroll, grid.Row(1), grid.Column(0), grid.Sticky(grid.NSEW))
	grid.RowConfigure(lbFrame, 0, grid.Weight(1))
	grid.ColumnConfigure(lbFrame, 0, grid.Weight(1))

	first, last := lb.YVisibleRange()
	yscroll.Set(first, last)

	app.Run()
}
