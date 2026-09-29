package focus

import (
	"testing"

	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/window"
)

type focusServer struct {
	platform.DisplayServer
	focused platform.WindowID
}

func (s *focusServer) SetInputFocus(w platform.WindowID, _ int, _ platform.Timestamp) { s.focused = w }

func TestSwallowEchoDropsOnlyTheEchoedPair(t *testing.T) {
	top := &window.Window{PlatformID: 1, Flags: window.FlagTopLevel}
	a := &window.Window{PlatformID: 2, Parent: top}
	b := &window.Window{PlatformID: 3, Parent: top}
	top.Children = []*window.Window{a, b}
	srv := &focusServer{}
	d := event.NewDispatcher()
	var seen []string
	d.BindGlobal(event.FocusChangeMask, func(ev *event.Event) {
		kind := "in"
		if ev.Type == event.FocusOutType {
			kind = "out"
		}
		seen = append(seen, kind)
	})
	m := NewManager(d, srv, nil)
	m.toplevelReady[top] = true

	m.SetFocus(a)
	m.SetFocus(b)
	if srv.focused != b.PlatformID {
		t.Fatalf("X focus on %d, want %d", srv.focused, b.PlatformID)
	}
	real := func(typ event.Type, w *window.Window) *event.Event {
		return &event.Event{Type: typ, Window: w.PlatformID, FocusMode: platform.FocusModeNormal}
	}
	if !m.SwallowEcho(real(event.FocusOutType, a)) || !m.SwallowEcho(real(event.FocusInType, b)) {
		t.Error("the X focus pair for SetFocus(b) was not recognised as an echo")
	}
	if m.SwallowEcho(real(event.FocusInType, b)) {
		t.Error("a second FocusIn on b was swallowed")
	}
	if m.SwallowEcho(real(event.FocusOutType, b)) {
		t.Error("a genuine FocusOut (the application losing focus) was swallowed")
	}
	if want := []string{"in", "out", "in"}; len(seen) != len(want) {
		t.Errorf("dispatched %v, want %v", seen, want)
	}
}
