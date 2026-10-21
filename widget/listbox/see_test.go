package listbox

import (
	"testing"

	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/platform"
)

// TestSeeRedraws checks that See scrolls the listbox and repaints it, as
// YView does.
func TestSeeRedraws(t *testing.T) {
	app := testutil.NewTestApp(t)
	lb := New(app, "lb", Items("a", "b", "c", "d", "e", "f"))

	calls := 0
	lb.SetDisplayProc(func() { calls++ })

	lb.See(4)
	app.UpdateIdleTasks()
	if calls == 0 {
		t.Error("See did not redraw the listbox")
	}
}

// TestKeyboardSelectionNotifies checks that a selection made with the
// keyboard fires the selection callback, as Tk fires <<ListboxSelect>> from
// its key bindings (tk::ListboxUpDown, tk/library/listbox.tcl:383).
func TestKeyboardSelectionNotifies(t *testing.T) {
	app := testutil.NewTestApp(t)
	lb := New(app, "lb", Items("a", "b", "c"))

	notified := 0
	lb.SelectCmd = func() { notified++ }

	// The arrow keys move the active item and the selection with it.
	setActive(lb, 1)
	if notified != 1 {
		t.Errorf("moving the active item notified %d times, want 1", notified)
	}
	if sel := lb.Selection(); len(sel) != 1 || sel[0] != 1 {
		t.Errorf("selection = %v, want [1]", sel)
	}

	// The space key toggles the selection of the active item.
	notified = 0
	app.Dispatcher().Dispatch(&event.Event{
		Type: event.KeyPressType, Window: lb.Win.PlatformID,
		KeySym: platform.XK_space,
	})
	if notified != 1 {
		t.Errorf("the space key notified %d times, want 1", notified)
	}
}
