package window

import (
	"slices"
	"testing"
)

type recordingManager struct{ lost []string }

func (m *recordingManager) Name() string        { return "test" }
func (m *recordingManager) RequestProc(*Window) {}
func (m *recordingManager) LostContentProc(w *Window) {
	m.lost = append(m.lost, w.PathName)
}

func TestDestroyWindowRunsHooks(t *testing.T) {
	d := &Display{}
	root := &Window{PathName: ".", Display: d}
	frame := NewChildWindow(root, "f", 0, 0, 10, 10)
	button := NewChildWindow(frame, "b", 0, 0, 10, 10)
	mgr := &recordingManager{}
	frame.GeomManager = mgr
	button.GeomManager = mgr

	var order []string
	d.OnWindowDestroy(func(w *Window) { order = append(order, "display "+w.PathName) })
	button.OnDestroy(func() { order = append(order, "button base") })
	button.OnDestroy(func() {
		order = append(order, "button widget")
		DestroyWindow(button) // a widget's Destroy calling back in must not recurse
	})
	frame.OnDestroy(func() { order = append(order, "frame") })

	DestroyWindow(frame)
	DestroyWindow(frame)

	want := []string{
		"display .f.b", "button widget", "button base",
		"display .f", "frame",
	}
	if !slices.Equal(order, want) {
		t.Errorf("hook order = %q\nwant %q", order, want)
	}
	if !slices.Equal(mgr.lost, []string{".f.b", ".f"}) {
		t.Errorf("LostContentProc called for %q, want [.f.b .f]", mgr.lost)
	}
	if len(root.Children) != 0 || len(frame.Children) != 0 {
		t.Errorf("destroyed windows still linked: root %d, frame %d children", len(root.Children), len(frame.Children))
	}
	if !button.IsDestroyed() || frame.GeomManager != nil {
		t.Error("destroyed window not marked dead or still managed")
	}
}

func TestOnConfigureHooksCoexist(t *testing.T) {
	w := &Window{}
	var calls []string
	w.OnConfigure(func() { calls = append(calls, "pack") })
	w.OnConfigure(func() { calls = append(calls, "user") })
	w.NotifyConfigure()
	if !slices.Equal(calls, []string{"pack", "user"}) {
		t.Errorf("configure callbacks = %q", calls)
	}
}
