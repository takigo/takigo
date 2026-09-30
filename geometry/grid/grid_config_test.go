package grid

import (
	"testing"

	"github.com/msorc/takigo/geometry"
	"github.com/msorc/takigo/window"
)

func gridSlot(t *testing.T, parent, w *window.Window) (row, col int, ok bool) {
	t.Helper()
	for _, e := range gridders.Of(parent).entries {
		if e.window == w {
			return e.config.row, e.config.column, true
		}
	}
	return 0, 0, false
}

func TestGridRowColumn(t *testing.T) {
	parent := &window.Window{PathName: ".gridrc"}
	a := &window.Window{PathName: ".gridrc.a", Parent: parent}
	b := &window.Window{PathName: ".gridrc.b", Parent: parent}
	c := &window.Window{PathName: ".gridrc.c", Parent: parent}
	neg := &window.Window{PathName: ".gridrc.neg", Parent: parent}
	t.Cleanup(func() {
		for _, w := range []*window.Window{a, b, c, neg} {
			Forget(w)
		}
	})

	Grid(geometry.Group{a}, Row(0), Column(0))
	Grid(geometry.Group{b}, Row(0), Column(0))
	if r, col, _ := gridSlot(t, parent, b); r != 0 || col != 0 {
		t.Errorf("explicit Row(0), Column(0) placed at %d,%d, want 0,0", r, col)
	}
	Grid(geometry.Group{c})
	if r, _, _ := gridSlot(t, parent, c); r != 1 {
		t.Errorf("unspecified row = %d, want next row 1", r)
	}

	Grid(geometry.Group{neg}, Row(-1))
	Grid(geometry.Group{neg}, Column(-2))
	if _, _, ok := gridSlot(t, parent, neg); ok {
		t.Error("negative row/column was accepted")
	}
}

func TestGridKeepsContainerConfigWhenEmpty(t *testing.T) {
	parent := &window.Window{PathName: ".gridkeep"}
	a := &window.Window{PathName: ".gridkeep.a", Parent: parent}
	ColumnConfigure(parent, 0, Weight(1))
	Grid(geometry.Group{a}, Row(0), Column(0))
	Forget(a)
	Grid(geometry.Group{a}, Row(0), Column(0))
	t.Cleanup(func() { Forget(a) })
	if conf := gridders.Of(parent).colConf[0]; conf == nil || conf.Weight != 1 {
		t.Errorf("column 0 config after regridding = %+v, want weight 1", conf)
	}
}
