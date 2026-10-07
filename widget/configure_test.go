package widget

import (
	"testing"

	"github.com/takigo/takigo/internal/stubs"
)

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

func newTestWidget() (*testWidget, *stubs.Manager) {
	m := &stubs.Manager{}
	w := &testWidget{Base: *newTestBase("Test")}
	w.Win.GeomManager = m
	return w, m
}

func TestConfigureRequestsOnChange(t *testing.T) {
	w, m := newTestWidget()
	_ = Configure(w, []testOption{text("abc")}, w.computeGeometry)
	if w.Win.ReqWidth != 21 || m.Requests != 1 || w.displays != 1 {
		t.Fatalf("ReqWidth = %d, requests = %d, displays = %d", w.Win.ReqWidth, m.Requests, w.displays)
	}
	_ = Configure(w, []testOption{text("abc")}, w.computeGeometry)
	if m.Requests != 1 || w.displays != 2 {
		t.Fatalf("unchanged request: requests = %d, displays = %d", m.Requests, w.displays)
	}
}

func TestConfigureDirectRequestWrite(t *testing.T) {
	w, m := newTestWidget()
	width := func(n int) testOption { return func(w *testWidget) { w.Win.ReqWidth = n } }
	_ = Configure(w, []testOption{width(300)}, nil)
	if m.Requests != 1 {
		t.Fatalf("requests = %d", m.Requests)
	}
}

func TestConfigureSyncsBackground(t *testing.T) {
	w, _ := newTestWidget()
	bg := func(name string) testOption { return func(w *testWidget) { w.SetBackgroundColor(name) } }
	_ = Configure(w, []testOption{bg("#123456")}, nil)
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
	_ = Configure(w, []testOption{border(3)}, geom)
	_ = Configure(w, []testOption{border(3)}, geom)
	if arranged != 1 || m.Requests != 0 {
		t.Fatalf("arranged = %d, requests = %d", arranged, m.Requests)
	}
}
