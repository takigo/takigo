package geometry

import (
	"testing"

	"github.com/takigo/takigo/internal/stubs"
	"github.com/takigo/takigo/window"
)

func TestGeometryRequestClampsAndNotifies(t *testing.T) {
	m := &stubs.Manager{}
	w := &window.Window{GeomManager: m, MinReqWidth: 30, MinReqHeight: 20}
	GeometryRequest(w, 10, 50)
	if w.ReqWidth != 30 || w.ReqHeight != 50 || m.Requests != 1 {
		t.Errorf("request %dx%d after %d RequestProc calls, want 30x50 after 1", w.ReqWidth, w.ReqHeight, m.Requests)
	}
	GeometryRequest(w, 30, 50)
	if m.Requests != 1 {
		t.Error("an unchanged request reached the manager")
	}
	GeometryRequest(w, 0, 0)
	if w.ReqWidth != 30 || w.ReqHeight != 20 || m.Requests != 2 {
		t.Errorf("request %dx%d after %d calls, want the minimum 30x20 after 2", w.ReqWidth, w.ReqHeight, m.Requests)
	}
}

// losingManager counts the content it loses to another manager.
type losingManager struct {
	stubs.Manager
	lost int
}

func (m *losingManager) LostContentProc(*window.Window) { m.lost++ }

func TestManageGeometryTellsTheOldManager(t *testing.T) {
	old := &losingManager{}
	w := &window.Window{}
	ManageGeometry(w, old)
	ManageGeometry(w, old)
	if w.GeomManager != old || old.lost != 0 {
		t.Fatalf("re-managing by the same manager: manager %v, lost %d", w.GeomManager, old.lost)
	}
	next := &stubs.Manager{}
	ManageGeometry(w, next)
	if w.GeomManager != next || old.lost != 1 {
		t.Errorf("after a takeover: manager %v, old lost %d (want 1)", w.GeomManager, old.lost)
	}
}

func TestInternalBorder(t *testing.T) {
	w := &window.Window{Width: 100, Height: 80}
	SetInternalBorder(w, 1, 2, 3, 4)
	if in := w.ContentInsets(); in[0] != 1 || in[1] != 2 || in[2] != 3 || in[3] != 4 {
		t.Errorf("insets = %v", in)
	}
}

func TestWhenIdleCoalesces(t *testing.T) {
	d := stubs.Display()
	w := &window.Window{Display: d}
	runs := 0
	pending := false
	WhenIdle(w, &pending, func() { runs++ })
	if runs != 1 {
		t.Errorf("without an idle scheduler the arrange ran %d times, want at once", runs)
	}
	var queued []func()
	d.DoWhenIdle = func(fn func()) { queued = append(queued, fn) }
	WhenIdle(w, &pending, func() { runs++ })
	WhenIdle(w, &pending, func() { runs++ })
	if len(queued) != 1 || !pending {
		t.Fatalf("%d idle callbacks queued for two requests, want 1", len(queued))
	}
	queued[0]()
	if runs != 2 || pending {
		t.Errorf("after the idle callback: %d runs, pending %v", runs, pending)
	}
}

func TestTable(t *testing.T) {
	var a, b Table[int]
	w := &window.Window{}
	a.Set(w, 1)
	b.Set(w, 2)
	if v, ok := a.Get(w); !ok || v != 1 || b.Of(w) != 2 {
		t.Errorf("two tables on one window: a=%v b=%v", v, b.Of(w))
	}
	a.Delete(w)
	if _, ok := a.Get(w); ok || a.Of(w) != 0 || b.Of(w) != 2 {
		t.Error("Delete touched the other table or left the entry")
	}
	a.Delete(nil)
}
