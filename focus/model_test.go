package focus

import (
	"slices"
	"testing"

	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/window"
)

type focusServer struct {
	platform.DisplayServer
	focused []platform.WindowID
}

func (s *focusServer) SetInputFocus(w platform.WindowID, _ int, _ platform.Timestamp) {
	s.focused = append(s.focused, w)
}

type focusRig struct {
	m        *Manager
	srv      *focusServer
	top      *window.Window
	a, b     *window.Window
	top2, c  *window.Window
	received []string
}

func newFocusRig(t *testing.T) *focusRig {
	t.Helper()
	d := &window.Display{Windows: map[platform.WindowID]*window.Window{}}
	r := &focusRig{srv: &focusServer{}}
	r.top = &window.Window{PlatformID: 1, Flags: window.FlagTopLevel, Display: d, Name: "top"}
	r.a = &window.Window{PlatformID: 2, Parent: r.top, Display: d, Name: "a", X: 10, Y: 20}
	r.b = &window.Window{PlatformID: 3, Parent: r.top, Display: d, Name: "b", X: 100, Y: 20}
	r.top.Children = []*window.Window{r.a, r.b}
	r.top2 = &window.Window{PlatformID: 4, Flags: window.FlagTopLevel, Display: d, Name: "top2"}
	r.c = &window.Window{PlatformID: 5, Parent: r.top2, Display: d, Name: "c"}
	r.top2.Children = []*window.Window{r.c}
	for _, w := range []*window.Window{r.top, r.a, r.b, r.top2, r.c} {
		d.RegisterWindow(w.PlatformID, w)
	}
	disp := event.NewDispatcher()
	disp.BindGlobal(event.FocusChangeMask, func(ev *event.Event) {
		kind := "in"
		if ev.Type == event.FocusOutType {
			kind = "out"
		}
		r.received = append(r.received, kind+":"+d.LookupWindow(ev.Window).Name)
	})
	r.m = NewManager(disp, r.srv, d)
	return r
}

// real delivers an X event through FilterEvent and reports whether it
// would be dispatched.
func (r *focusRig) real(typ event.Type, w *window.Window, detail int) bool {
	return r.m.FilterEvent(&event.Event{Type: typ, Window: w.PlatformID, FocusDetail: detail})
}

func (r *focusRig) takeReceived() []string {
	got := r.received
	r.received = nil
	return got
}

func TestFocusStaysOnToplevel(t *testing.T) {
	r := newFocusRig(t)
	r.m.SetFocus(r.a) // before X has shown the toplevel viewable
	if len(r.srv.focused) != 0 {
		t.Errorf("SetInputFocus before the toplevel was focused: %v", r.srv.focused)
	}
	if r.real(event.FocusInType, r.top, platform.FocusDetailNonlinear) {
		t.Error("toplevel FocusIn was passed on instead of being handled")
	}
	r.m.SetFocus(r.b)
	if !slices.Equal(r.srv.focused, nil) {
		t.Errorf("X focus moved within a toplevel that has it: %v", r.srv.focused)
	}
	if got, want := r.takeReceived(), []string{"in:a", "out:a", "in:b"}; !slices.Equal(got, want) {
		t.Errorf("focus events %v, want %v", got, want)
	}
	if r.real(event.FocusInType, r.a, platform.FocusDetailAncestor) || r.real(event.FocusOutType, r.b, platform.FocusDetailAncestor) {
		t.Error("focus events on child windows were passed on")
	}
}

func TestKeysGoToFocusWindow(t *testing.T) {
	r := newFocusRig(t)
	r.real(event.FocusInType, r.top, platform.FocusDetailNonlinear)
	r.m.SetFocus(r.b)
	// X delivers the key to a, the child under the pointer.
	ev := &event.Event{Type: event.KeyPressType, Window: r.a.PlatformID, X: 5, Y: 5}
	r.m.FilterEvent(ev)
	if ev.Window != r.b.PlatformID || ev.X != 5+10-100 || ev.Y != 5 {
		t.Errorf("key went to %d at %d,%d; want b (3) at -85,5", ev.Window, ev.X, ev.Y)
	}
}

func TestApplicationLosesAndRegainsFocus(t *testing.T) {
	r := newFocusRig(t)
	r.real(event.FocusInType, r.top, platform.FocusDetailNonlinear)
	r.m.SetFocus(r.a)
	r.takeReceived()

	r.real(event.FocusOutType, r.top, platform.FocusDetailNonlinear) // Alt-Tab away
	if got := r.takeReceived(); !slices.Equal(got, []string{"out:a"}) {
		t.Errorf("on losing the X focus: %v, want [out:a]", got)
	}
	r.m.SetFocus(r.b) // recorded only: focus without -force
	if got := r.takeReceived(); len(got) != 0 {
		t.Errorf("SetFocus while unfocused dispatched %v", got)
	}
	r.real(event.FocusInType, r.top, platform.FocusDetailNonlinear) // back
	if got := r.takeReceived(); !slices.Equal(got, []string{"in:b"}) {
		t.Errorf("on regaining the X focus: %v, want [in:b]", got)
	}
}

func TestFocusMovesAcrossToplevels(t *testing.T) {
	r := newFocusRig(t)
	r.real(event.FocusInType, r.top, platform.FocusDetailNonlinear)
	r.real(event.FocusInType, r.top2, platform.FocusDetailNonlinear)
	r.m.SetFocus(r.a)
	r.srv.focused = nil
	r.takeReceived()

	r.m.SetFocus(r.c)
	if !slices.Equal(r.srv.focused, []platform.WindowID{r.top2.PlatformID}) {
		t.Errorf("X focus set to %v, want top2", r.srv.focused)
	}
	// X then reports the move; it changes nothing further.
	r.real(event.FocusOutType, r.top, platform.FocusDetailNonlinear)
	r.real(event.FocusInType, r.top2, platform.FocusDetailNonlinear)
	if got := r.takeReceived(); !slices.Equal(got, []string{"out:a", "in:c"}) {
		t.Errorf("focus events %v, want [out:a in:c]", got)
	}
}

func TestImplicitFocusFollowsPointer(t *testing.T) {
	r := newFocusRig(t)
	r.m.SetFocus(r.a)
	r.takeReceived()
	r.real(event.FocusOutType, r.top, platform.FocusDetailNonlinear)
	r.takeReceived()

	r.real(event.FocusInType, r.top, platform.FocusDetailPointer) // pointer enters, no WM
	r.m.SetFocus(r.b)
	if len(r.srv.focused) != 0 {
		t.Errorf("implicit focus claimed the X focus: %v", r.srv.focused)
	}
	r.real(event.LeaveType, r.top, platform.FocusDetailNonlinear) // pointer leaves
	if got := r.takeReceived(); !slices.Equal(got, []string{"in:a", "out:a", "in:b", "out:b"}) {
		t.Errorf("focus events %v", got)
	}
}
