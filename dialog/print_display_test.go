package dialog

import (
	"fmt"
	"testing"

	"github.com/takigo/takigo/geometry/grid"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/ttk"
	"github.com/takigo/takigo/widget"
)

// TestPrinterListScrollbar builds the print dialog's printer list with more
// printers than rows and checks that the thumb follows the view.
func TestPrinterListScrollbar(t *testing.T) {
	app := testutil.NewTestApp(t)
	pf := ttk.NewFrame(app, "printerf")
	pack.Pack(pf)
	tv := ttk.NewTreeview(pf, "prlist", ttk.TreeviewHeight(5), ttk.TreeviewColumns("printer"),
		ttk.TreeviewShow("headings"))
	sy := ttk.NewScrollbar(pf, "sy", ttk.ScrollbarCommand(widget.ScrollY(tv)))
	tv.YScrollCmd = autoscroll(sy, grid.Row(0), grid.Column(1), grid.Sticky(grid.NS))
	for i := range 10 {
		tv.Insert("", -1, ttk.ItemValues(fmt.Sprintf("printer%d", i)))
	}
	grid.Grid(tv, grid.Row(0), grid.Column(0), grid.Sticky(grid.NSEW))
	app.UpdateIdleTasks()
	arrangeNow(app.Root())
	app.UpdateIdleTasks()

	if sy.Window().GeomManager == nil {
		t.Fatal("the scrollbar is not gridded for a partial view")
	}
	first, last := tv.YVisibleRange()
	if first != 0 || last >= 1 {
		t.Fatalf("view = %v..%v, want a partial view at the top", first, last)
	}
	tv.YViewMoveTo(0.5) // what dragging the thumb does
	first, last = tv.YVisibleRange()
	if first == 0 {
		t.Fatalf("YViewMoveTo(0.5) left the view at %v..%v", first, last)
	}
	if sy.First != first || sy.Last != last {
		t.Errorf("thumb = %v..%v after scrolling, want %v..%v", sy.First, sy.Last, first, last)
	}
}
