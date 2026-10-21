package focus

import (
	"slices"
	"testing"

	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/window"
)

// A destroyed focus window must get its FocusOut, so a widget that draws
// a focus ring learns that the focus is gone.
func TestDestroyedFocusWindowGetsFocusOut(t *testing.T) {
	r := newFocusRig(t)
	r.real(event.FocusInType, r.top, platform.FocusDetailNonlinear)
	r.m.SetFocus(r.a)
	r.takeReceived()

	// The toplevel goes with the focused window.
	r.top.Flags |= window.FlagAlreadyDead
	r.m.HandleDestroyWindow(r.a)
	if got := r.takeReceived(); !slices.Equal(got, []string{"out:a"}) {
		t.Errorf("focus events %v, want [out:a]", got)
	}
	if r.m.FocusWindow() != nil {
		t.Errorf("the focus is still %v after the window was destroyed", r.m.FocusWindow())
	}
}

// The same, but the toplevel survives: the focus moves to it, and the
// old focus window still gets its FocusOut.
func TestDestroyedFocusWindowInSurvivingToplevel(t *testing.T) {
	r := newFocusRig(t)
	r.real(event.FocusInType, r.top, platform.FocusDetailNonlinear)
	r.m.SetFocus(r.a)
	r.takeReceived()

	r.m.HandleDestroyWindow(r.a)
	if got := r.takeReceived(); !slices.Equal(got, []string{"out:a", "in:top"}) {
		t.Errorf("focus events %v, want [out:a in:top]", got)
	}
	if r.m.FocusWindow() != r.top {
		t.Errorf("the focus is %v, want the toplevel", r.m.FocusWindow())
	}
}

// The manager tells its listener about the change of focus, as when the
// focus moves to another window.
func TestFocusChangeListenerIsToldOnDestroy(t *testing.T) {
	r := newFocusRig(t)
	var lost, gained *window.Window
	r.m.OnFocusChange = func(l, g *window.Window) { lost, gained = l, g }
	r.real(event.FocusInType, r.top, platform.FocusDetailNonlinear)
	r.m.SetFocus(r.a)
	r.takeReceived()

	r.top.Flags |= window.FlagAlreadyDead
	r.m.HandleDestroyWindow(r.a)
	if lost != r.a || gained != nil {
		t.Errorf("OnFocusChange(%v, %v), want (%v, nil)", lost, gained, r.a)
	}
}
