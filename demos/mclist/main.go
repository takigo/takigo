// Demo: Multi-column sortable list using TTK Treeview.
// Ported from Tk's mclist.tcl demo.
package main

import (
	"strings"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/scrollbar"
)

type countryData struct {
	Country  string
	Capital  string
	Currency string
}

var countries = []countryData{
	{"Argentina", "Buenos Aires", "ARS"},
	{"Australia", "Canberra", "AUD"},
	{"Brazil", "Brasilia", "BRL"},
	{"Canada", "Ottawa", "CAD"},
	{"China", "Beijing", "CNY"},
	{"France", "Paris", "EUR"},
	{"Germany", "Berlin", "EUR"},
	{"India", "New Delhi", "INR"},
	{"Italy", "Rome", "EUR"},
	{"Japan", "Tokyo", "JPY"},
	{"Mexico", "Mexico City", "MXN"},
	{"Russia", "Moscow", "RUB"},
	{"South Korea", "Seoul", "KRW"},
	{"United Kingdom", "London", "GBP"},
	{"United States", "Washington, D.C.", "USD"},
}

func main() {
	d := demohelper.Setup("Multi-Column List", 500, 400,
		"A sortable multi-column list.\nClick column headings to sort.")
	root, app := d.Root, d.App

	ttk.SetCurrentTheme("clam")

	// Treeview frame with scrollbar.
	tvFrame := frame.New(root, "tvframe", app)
	pack.Pack(tvFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	tv := ttk.NewTreeview(tvFrame.Window(), "mclist", app,
		ttk.TreeviewColumns("country", "capital", "currency"),
		ttk.TreeviewShow("headings"),
	)

	// Configure columns.
	tv.ColumnConfigure("country", ttk.ColWidth(180))
	tv.ColumnConfigure("capital", ttk.ColWidth(180))
	tv.ColumnConfigure("currency", ttk.ColWidth(80), ttk.ColAnchor(option.AnchorCenter))

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
	tv.HeadingConfigure("country", ttk.HeadText("Country"), ttk.HeadCommand(makeSortCmd("country")))
	tv.HeadingConfigure("capital", ttk.HeadText("Capital"), ttk.HeadCommand(makeSortCmd("capital")))
	tv.HeadingConfigure("currency", ttk.HeadText("Currency"), ttk.HeadCommand(makeSortCmd("currency")))

	// Insert data.
	for _, c := range countries {
		tv.Insert("", -1, ttk.ItemValues(c.Country, c.Capital, c.Currency))
	}

	// Scrollbar.
	yscroll := scrollbar.New(tvFrame.Window(), "yscroll", app,
		scrollbar.OrientOpt(scrollbar.Vertical),
		scrollbar.CommandOpt(func(args ...any) {
			if len(args) < 1 {
				return
			}
			switch args[0] {
			case "moveto":
				if len(args) >= 2 {
					if f, ok := args[1].(float64); ok {
						tv.YViewMoveTo(f)
					}
				}
			case "scroll":
				if len(args) >= 3 {
					n, _ := args[1].(int)
					unit, _ := args[2].(string)
					tv.YViewScroll(n, unit == "pages")
				}
			}
		}),
	)

	tv.YScrollCmd = func(first, last float64) {
		yscroll.Set(first, last)
	}

	pack.Pack(yscroll, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(tv, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	first, last := tv.YVisibleRange()
	yscroll.Set(first, last)

	d.Run()
}
