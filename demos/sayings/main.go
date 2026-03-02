// Demo: Scrollable sayings listbox with vertical scrollbar.
// Ported from Tk's sayings.tcl demo.
package main

import (
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/listbox"
	"github.com/msorc/takigo/widget/scrollbar"
)

func main() {
	d := demohelper.Setup("Sayings", 500, 350,
		"A listbox with well-known sayings.\nScroll vertically to see them all.")
	root, app := d.Root, d.App

	// Sayings data.
	sayings := []string{
		"Don't speculate, measure",
		"A penny for your thoughts",
		"A stitch in time saves nine",
		"An apple a day keeps the doctor away",
		"Don't put all your eggs in one basket",
		"Early to bed and early to rise makes a man healthy, wealthy, and wise",
		"Every cloud has a silver lining",
		"Fool me once, shame on you; fool me twice, shame on me",
		"Good things come to those who wait",
		"Haste makes waste",
		"If at first you don't succeed, try, try again",
		"Jack of all trades, master of none",
		"Keep your friends close and your enemies closer",
		"Laughter is the best medicine",
		"Make hay while the sun shines",
		"Necessity is the mother of invention",
		"Once bitten, twice shy",
		"People who live in glass houses shouldn't throw stones",
		"The early bird catches the worm",
		"The pen is mightier than the sword",
	}

	// Listbox frame with Y scrollbar.
	lbFrame := frame.New(root, "lbframe", app)
	pack.Pack(lbFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	lb := listbox.New(lbFrame.Window(), "sayings", app,
		listbox.Items(sayings...),
		listbox.Width(20),
		listbox.Height(10),
	)

	yscroll := scrollbar.New(lbFrame.Window(), "yscroll", app,
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
		yscroll.Set(first, last)
	}

	pack.Pack(yscroll, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(lb, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	first, last := lb.YVisibleRange()
	yscroll.Set(first, last)

	d.Run()
}
