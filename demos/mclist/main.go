// Demo: Multi-column listbox using a Ttk tree widget.
// Ported from Tk's mclist.tcl demo.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/window"
)

type countryData struct {
	Country  string
	Capital  string
	Currency string
}

var countries = []countryData{
	{"Argentina", "Buenos Aires", "ARS"},
	{"Australia", "Canberra", "AUD"},
	{"Brazil", "Brazilia", "BRL"},
	{"Canada", "Ottawa", "CAD"},
	{"China", "Beijing", "CNY"},
	{"France", "Paris", "EUR"},
	{"Germany", "Berlin", "EUR"},
	{"India", "New Delhi", "INR"},
	{"Italy", "Rome", "EUR"},
	{"Japan", "Tokyo", "JPY"},
	{"Mexico", "Mexico City", "MXN"},
	{"Russia", "Moscow", "RUB"},
	{"South Africa", "Pretoria", "ZAR"},
	{"United Kingdom", "London", "GBP"},
	{"United States", "Washington, D.C.", "USD"},
}

func main() {
	app, err := takigo.NewApp(takigo.Title("Multi-Column List"),
		takigo.Geometry("+300+300"),
		takigo.IconName("mclist"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := ttk.NewLabel(f, "msg",
		ttk.LabelWrapLength(screenunit.In(4)),
		ttk.LabelJustify(option.JustifyLeft),
		ttk.LabelAnchor(option.AnchorN),
		ttk.LabelPadding("10 2 10 6"),
		ttk.LabelText("Ttk is the new Tk themed widget set. One of the widgets it includes is a tree widget, which can be configured to display multiple columns of informational data without displaying the tree itself. This is a simple way to build a listbox that has multiple columns. Clicking on the heading for a column will sort the data by that column. You can also change the width of the columns by dragging the boundary between them."),
	)
	pack.Pack(msg, pack.FillOpt(pack.FillX))

	// Grid checkbutton: the "extra" widget of addSeeDismiss. Toggling it
	// enables row stripes and column separators (tglGrid in mclist.tcl).
	gridVar := widget.NewVariable(false)
	var gridCb *ttk.Checkbutton
	btns := demohelper.AddSeeDismissExtra(f, func(bf *ttk.Frame) window.Windower {
		gridCb = ttk.NewCheckbutton(bf, "cb1",
			ttk.CheckbuttonText("Grid"),
			ttk.CheckbuttonVar(gridVar),
		)
		return gridCb
	})
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Container frame (grid layout for treeview + scrollbars).
	tvFrame := ttk.NewFrame(f, "container")
	pack.Pack(tvFrame, pack.FillOpt(pack.FillBoth), pack.Expand(true))

	tv := ttk.NewTreeview(tvFrame, "tree",
		ttk.TreeviewColumns("country", "capital", "currency"),
		ttk.TreeviewShow("headings"),
	)

	// Column widths as in mclist.tcl: the heading text in the heading font
	// plus the 16px noArrow image and 4px, widened to fit "value  ".
	headFont, _ := app.FontRegistry().Get(font.TkHeadingFont)
	rowFont, _ := app.FontRegistry().Get(font.TkDefaultFont)
	morePx := 16 + 4*screenunit.ScalingPct()/100
	colWidth := map[string]int{}
	cols := []string{"country", "capital", "currency"}
	for i, name := range []string{"Country", "Capital", "Currency"} {
		colWidth[cols[i]] = headFont.MeasureString(name) + morePx
	}
	for _, c := range countries {
		for col, v := range map[string]string{"country": c.Country, "capital": c.Capital, "currency": c.Currency} {
			colWidth[col] = max(colWidth[col], rowFont.MeasureString(v+"  "))
		}
	}
	for _, col := range cols {
		tv.ColumnConfigure(col, ttk.ColWidth(colWidth[col]))
	}

	// Wire Grid checkbox command now that tv exists.
	gridCb.Command = func() {
		enabled := gridVar.Get()
		tv.SetStripe(enabled)
		for _, col := range []string{"country", "capital", "currency"} {
			tv.SetColSeparator(col, enabled)
		}
	}

	// Sort state tracker.
	sortReverse := map[string]bool{}

	makeSortCmd := func(colID string) func() {
		return func() {
			reverse := sortReverse[colID]
			colIdx := -1
			switch colID {
			case "country":
				colIdx = 0
			case "capital":
				colIdx = 1
			case "currency":
				colIdx = 2
			}
			if colIdx < 0 {
				return
			}
			tv.SortChildren("", func(a, b *ttk.TreeItem) bool {
				va, vb := "", ""
				if colIdx < len(a.Values) {
					va = a.Values[colIdx]
				}
				if colIdx < len(b.Values) {
					vb = b.Values[colIdx]
				}
				cmp := strings.Compare(strings.ToLower(va), strings.ToLower(vb))
				if reverse {
					return cmp > 0
				}
				return cmp < 0
			})
			tv.SetSortIndicator(colID, reverse)
			sortReverse[colID] = !reverse
		}
	}

	// Configure headings.
	tv.HeadingConfigure("country", ttk.HeadText("Country"), ttk.HeadAnchor(option.AnchorW), ttk.HeadCommand(makeSortCmd("country")))
	tv.HeadingConfigure("capital", ttk.HeadText("Capital"), ttk.HeadAnchor(option.AnchorW), ttk.HeadCommand(makeSortCmd("capital")))
	tv.HeadingConfigure("currency", ttk.HeadText("Currency"), ttk.HeadAnchor(option.AnchorW), ttk.HeadCommand(makeSortCmd("currency")))

	// Insert data.
	for _, c := range countries {
		tv.Insert("", -1, ttk.ItemValues(c.Country, c.Capital, c.Currency))
	}

	// Scrollbars — grid layout: tree(0,0), yscroll(0,1), xscroll(1,0).
	yscroll := ttk.NewScrollbar(tvFrame, "vsb",
		ttk.ScrollbarOrientOpt(ttk.Vertical),
		ttk.ScrollbarCommandOpt(widget.ScrollY(tv)),
	)
	xscroll := ttk.NewScrollbar(tvFrame, "hsb",
		ttk.ScrollbarOrientOpt(ttk.Horizontal),
	)

	tv.YScrollCmd = func(first, last float64) {
		yscroll.Set(first, last)
	}

	grid.Grid(tv, grid.Row(0), grid.Column(0), grid.Sticky(grid.NSEW))
	grid.Grid(yscroll, grid.Row(0), grid.Column(1), grid.Sticky(grid.NSEW))
	grid.Grid(xscroll, grid.Row(1), grid.Column(0), grid.Sticky(grid.NSEW))
	grid.ColumnConfigure(tvFrame, 0, grid.Weight(1))
	grid.RowConfigure(tvFrame, 0, grid.Weight(1))

	first, last := tv.YVisibleRange()
	yscroll.Set(first, last)
	_ = xscroll

	app.Run()
}
