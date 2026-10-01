package widget

import (
	"testing"

	"github.com/msorc/takigo/window"
)

type fakeManager struct{ requests int }

func (*fakeManager) Name() string                   { return "fake" }
func (m *fakeManager) RequestProc(*window.Window)   { m.requests++ }
func (*fakeManager) LostContentProc(*window.Window) {}

type testWidget struct {
	Base
	text     string
	displays int
}

func (w *testWidget) Display() { w.displays++ }

func (w *testWidget) computeGeometry() {
	w.Win.ReqWidth = 7 * len(w.text)
	w.Win.ReqHeight = 13
}

type testOption func(*testWidget)

func text(s string) testOption { return func(w *testWidget) { w.text = s } }

func newTestWidget() (*testWidget, *fakeManager) {
	m := &fakeManager{}
	w := &testWidget{Base: *newTestBase("Test")}
	w.Win.GeomManager = m
	return w, m
}

func TestConfigureRequestsOnChange(t *testing.T) {
	w, m := newTestWidget()
	Configure(w, []testOption{text("abc")}, w.computeGeometry)
	if w.Win.ReqWidth != 21 || m.requests != 1 || w.displays != 1 {
		t.Fatalf("ReqWidth = %d, requests = %d, displays = %d", w.Win.ReqWidth, m.requests, w.displays)
	}
	Configure(w, []testOption{text("abc")}, w.computeGeometry)
	if m.requests != 1 || w.displays != 2 {
		t.Fatalf("unchanged request: requests = %d, displays = %d", m.requests, w.displays)
	}
}

func TestConfigureDirectRequestWrite(t *testing.T) {
	w, m := newTestWidget()
	width := func(n int) testOption { return func(w *testWidget) { w.Win.ReqWidth = n } }
	Configure(w, []testOption{width(300)}, nil)
	if m.requests != 1 {
		t.Fatalf("requests = %d", m.requests)
	}
}

func TestConfigureSyncsBackground(t *testing.T) {
	w, _ := newTestWidget()
	bg := func(name string) testOption { return func(w *testWidget) { w.SetBackgroundColor(name) } }
	Configure(w, []testOption{bg("#123456")}, nil)
	if w.Win.BackgroundPixel != w.Background.Pixel || w.Win.BackgroundPixel == 0 {
		t.Fatalf("BackgroundPixel = %x, want %x", w.Win.BackgroundPixel, w.Background.Pixel)
	}
}

func TestConfigureRearrangesContent(t *testing.T) {
	w, m := newTestWidget()
	arranged := 0
	w.Win.OnConfigure(func() { arranged++ })
	border := func(n int) testOption { return func(w *testWidget) { w.BorderWidth = n } }
	geom := func() { w.Win.InternalBorderLeft = w.BorderWidth }
	Configure(w, []testOption{border(3)}, geom)
	Configure(w, []testOption{border(3)}, geom)
	if arranged != 1 || m.requests != 0 {
		t.Fatalf("arranged = %d, requests = %d", arranged, m.requests)
	}
}
