package grid

import (
	"testing"

	"github.com/takigo/takigo/geometry"
	"github.com/takigo/takigo/window"
)

func gridEntryOf(t *testing.T, w *window.Window) gridConfig {
	t.Helper()
	g, ok := gridders.Get(containerFor(w))
	if !ok || g.entry(w) == nil {
		t.Fatalf("%s is not gridded", w.PathName)
	}
	return g.entry(w).config
}

// TestGridMergesOptionsOfGriddedContent checks Tk's "grid configure"
// semantics: options not given keep their values, an omitted row or column
// keeps the widget's cell, and only content gridded anew starts from the
// defaults.
func TestGridMergesOptionsOfGriddedContent(t *testing.T) {
	parent := &window.Window{PathName: ".gm"}
	a := &window.Window{PathName: ".gm.a", Parent: parent}
	b := &window.Window{PathName: ".gm.b", Parent: parent}
	c := &window.Window{PathName: ".gm.c", Parent: parent}
	t.Cleanup(func() {
		for _, w := range []*window.Window{a, b, c} {
			Forget(w)
		}
	})

	Grid(geometry.Group{a}, Row(2), Column(1), ColumnSpan(2), Sticky(EW), PadX(4))
	Grid(geometry.Group{a}, Sticky(NS))
	got := gridEntryOf(t, a)
	if got.row != 2 || got.column != 1 || got.columnSpan != 2 || got.sticky != NS || got.padLeft != 4 || got.padX != 8 {
		t.Errorf("after Sticky(NS): %+v, want row 2 column 1 span 2 sticky NS padx 4", got)
	}

	Grid(geometry.Group{a}, Column(3))
	if got := gridEntryOf(t, a); got.row != 2 || got.column != 3 || got.sticky != NS {
		t.Errorf("after Column(3): %+v, want row 2 column 3 sticky NS", got)
	}
	Grid(geometry.Group{a}, Row(0))
	if got := gridEntryOf(t, a); got.row != 0 || got.column != 3 {
		t.Errorf("after Row(0): %+v, want row 0 column 3", got)
	}

	// A group of gridded and new content: a keeps its cell, the new b gets
	// the next free row and the column after a's position in the group.
	Grid(geometry.Group{a, b})
	if got := gridEntryOf(t, a); got.row != 0 || got.column != 3 {
		t.Errorf("a moved to %d,%d, want 0,3", got.row, got.column)
	}
	if got := gridEntryOf(t, b); got.row != 1 || got.column != 1 {
		t.Errorf("b placed at %d,%d, want 1,1", got.row, got.column)
	}

	// After forget, grid starts over from the defaults.
	Forget(a)
	Grid(geometry.Group{a})
	if got := gridEntryOf(t, a); got.row != 2 || got.column != 0 || got.sticky != 0 || got.columnSpan != 1 || got.padX != 0 {
		t.Errorf("after forget and grid: %+v, want row 2 column 0 and defaults", got)
	}
	_ = c
}

// TestGridInKeepsCellAndContainer checks that In moves content to another
// container with its cell and options, and that re-gridding it afterwards
// without In leaves it there.
func TestGridInKeepsCellAndContainer(t *testing.T) {
	parent := &window.Window{PathName: ".gi"}
	inner := &window.Window{PathName: ".gi.inner", Parent: parent}
	a := &window.Window{PathName: ".gi.a", Parent: parent}
	t.Cleanup(func() { Forget(a); Forget(inner) })

	Grid(geometry.Group{inner}, Row(0), Column(0))
	Grid(geometry.Group{a}, Row(1), Column(2), Sticky(EW))
	Grid(geometry.Group{a}, In(inner))
	if got := containerFor(a); got != inner {
		t.Fatalf("a is in %s, want .gi.inner", got.PathName)
	}
	if got := gridEntryOf(t, a); got.row != 1 || got.column != 2 || got.sticky != EW {
		t.Errorf("after In: %+v, want row 1 column 2 sticky EW", got)
	}
	if g, _ := gridders.Get(parent); g.entry(a) != nil {
		t.Error("a is still in its old container's grid")
	}

	Grid(geometry.Group{a}, Sticky(NSEW))
	if got := containerFor(a); got != inner {
		t.Errorf("re-grid without In moved a to %s, want .gi.inner", got.PathName)
	}
}
