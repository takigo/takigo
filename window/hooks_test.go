package window

import (
	"testing"

	"github.com/takigo/takigo/platform"
)

func newTestDisplay() *Display {
	return &Display{Server: nil, Windows: map[platform.WindowID]*Window{}}
}

// A display's hooks belong to that display alone: another display's
// windows do not run them.
func TestDisplayHooksArePerDisplay(t *testing.T) {
	d1, d2 := newTestDisplay(), newTestDisplay()
	var mapped, unmapped, moved int
	d1.OnMapped(func(*Window) { mapped++ })
	d1.OnUnmapped(func(*Window) { unmapped++ })
	d1.OnMoved(func(*Window) { moved++ })

	w1 := &Window{Display: d1, PlatformID: 1}
	MarkMapped(w1)
	MarkUnmapped(w1)
	NotifyMoved(w1)
	if mapped != 1 || unmapped != 1 || moved != 1 {
		t.Errorf("hooks ran %d, %d, %d times, want 1 each", mapped, unmapped, moved)
	}

	w2 := &Window{Display: d2, PlatformID: 2}
	MarkMapped(w2)
	MarkUnmapped(w2)
	NotifyMoved(w2)
	if mapped != 1 || unmapped != 1 || moved != 1 {
		t.Errorf("another display's window ran the hooks: %d, %d, %d", mapped, unmapped, moved)
	}
}

// A hook can be unregistered again.
func TestDisplayHookUnregister(t *testing.T) {
	d := newTestDisplay()
	count := 0
	remove := d.OnMapped(func(*Window) { count++ })
	MarkMapped(&Window{Display: d, PlatformID: 1})
	remove()
	MarkMapped(&Window{Display: d, PlatformID: 2})
	if count != 1 {
		t.Errorf("the hook ran %d times, want 1", count)
	}
}
