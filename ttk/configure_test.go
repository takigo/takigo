package ttk

import (
	"testing"

	"github.com/msorc/takigo/window"
)

type fakeManager struct{ requests int }

func (*fakeManager) Name() string                   { return "fake" }
func (m *fakeManager) RequestProc(*window.Window)   { m.requests++ }
func (*fakeManager) LostContentProc(*window.Window) {}

type sizedWidget struct {
	TtkWidget
	width, margin int
	displays      int
}

type sizedOption func(*sizedWidget)

func newSizedWidget() (*sizedWidget, *fakeManager) {
	m := &fakeManager{}
	w := &sizedWidget{}
	w.Win = &window.Window{GeomManager: m}
	w.DisplayFunc = func() { w.displays++ }
	return w, m
}

func (w *sizedWidget) size() {
	w.Win.ReqWidth, w.Win.ReqHeight = w.width, 10
	w.Win.InternalBorderLeft = w.margin
}

func TestResizeRequestsOnChange(t *testing.T) {
	w, m := newSizedWidget()
	w.width = 40
	w.resize(w.size)
	w.resize(w.size)
	if m.requests != 1 || w.Win.ReqWidth != 40 {
		t.Fatalf("requests = %d, ReqWidth = %d", m.requests, w.Win.ReqWidth)
	}
}

func TestConfigure(t *testing.T) {
	w, m := newSizedWidget()
	arranged, synced := 0, 0
	w.Win.OnConfigure(func() { arranged++ })
	width := func(n int) sizedOption { return func(w *sizedWidget) { w.width = n } }
	margin := func(n int) sizedOption { return func(w *sizedWidget) { w.margin = n } }
	sync := func() { synced++ }

	configure(&w.TtkWidget, w, []sizedOption{width(40)}, sync, w.size)
	if m.requests != 1 || arranged != 0 || synced != 1 || w.displays != 1 {
		t.Fatalf("width: requests = %d, arranged = %d, synced = %d, displays = %d", m.requests, arranged, synced, w.displays)
	}
	configure(&w.TtkWidget, w, []sizedOption{margin(5)}, sync, w.size)
	if m.requests != 1 || arranged != 1 || w.displays != 2 {
		t.Fatalf("margin: requests = %d, arranged = %d, displays = %d", m.requests, arranged, w.displays)
	}
	configure(&w.TtkWidget, w, []sizedOption{width(40), margin(5)}, sync, w.size)
	if m.requests != 1 || arranged != 1 || w.displays != 3 {
		t.Fatalf("unchanged: requests = %d, arranged = %d, displays = %d", m.requests, arranged, w.displays)
	}
}
