package event

import (
	"fmt"
	"sync"
	"testing"

	"github.com/msorc/takigo/platform"
)

func TestDispatcherBindAndDispatch(t *testing.T) {
	d := NewDispatcher()
	var called bool
	d.Bind(platform.WindowID(1), KeyPressMask, func(ev *Event) {
		called = true
	})

	d.Dispatch(&Event{Type: KeyPressType, Window: platform.WindowID(1)})
	if !called {
		t.Error("handler should have been called")
	}
}

func TestDispatcherNoMatchMask(t *testing.T) {
	d := NewDispatcher()
	var called bool
	d.Bind(platform.WindowID(1), KeyPressMask, func(ev *Event) {
		called = true
	})

	// Dispatch a button event — should not trigger key handler.
	d.Dispatch(&Event{Type: ButtonPressType, Window: platform.WindowID(1)})
	if called {
		t.Error("handler should not be called for non-matching mask")
	}
}

func TestDispatcherNoMatchWindow(t *testing.T) {
	d := NewDispatcher()
	var called bool
	d.Bind(platform.WindowID(1), KeyPressMask, func(ev *Event) {
		called = true
	})

	// Dispatch to a different window.
	d.Dispatch(&Event{Type: KeyPressType, Window: platform.WindowID(2)})
	if called {
		t.Error("handler should not be called for different window")
	}
}

func TestDispatcherMultipleHandlers(t *testing.T) {
	d := NewDispatcher()
	count := 0
	d.Bind(platform.WindowID(1), ExposureMask, func(ev *Event) { count++ })
	d.Bind(platform.WindowID(1), ExposureMask, func(ev *Event) { count++ })

	d.Dispatch(&Event{Type: ExposeType, Window: platform.WindowID(1)})
	if count != 2 {
		t.Errorf("expected 2 handlers called, got %d", count)
	}
}

func TestDispatcherGlobal(t *testing.T) {
	d := NewDispatcher()
	var globalCalled bool
	d.BindGlobal(ButtonPressMask, func(ev *Event) {
		globalCalled = true
	})

	d.Dispatch(&Event{Type: ButtonPressType, Window: platform.WindowID(99)})
	if !globalCalled {
		t.Error("global handler should be called for any window")
	}
}

func TestDispatcherUnbind(t *testing.T) {
	d := NewDispatcher()
	var called bool
	d.Bind(platform.WindowID(1), KeyPressMask, func(ev *Event) {
		called = true
	})

	d.Unbind(platform.WindowID(1))
	d.Dispatch(&Event{Type: KeyPressType, Window: platform.WindowID(1)})
	if called {
		t.Error("handler should not be called after Unbind")
	}
}

func TestDispatcherOrderWindowThenGlobal(t *testing.T) {
	d := NewDispatcher()
	var order []string
	d.Bind(platform.WindowID(1), KeyPressMask, func(ev *Event) {
		order = append(order, "window")
	})
	d.BindGlobal(KeyPressMask, func(ev *Event) {
		order = append(order, "global")
	})

	d.Dispatch(&Event{Type: KeyPressType, Window: platform.WindowID(1)})
	if len(order) != 2 || order[0] != "window" || order[1] != "global" {
		t.Errorf("expected [window global], got %v", order)
	}
}

func TestDispatcherUnknownType(t *testing.T) {
	d := NewDispatcher()
	var called bool
	d.BindGlobal(AllEventsMask, func(ev *Event) {
		called = true
	})

	// Unknown type → TypeToMask returns 0 → no dispatch.
	d.Dispatch(&Event{Type: Type(999), Window: platform.WindowID(1)})
	if called {
		t.Error("should not dispatch unknown event type")
	}
}

func TestDispatcherUnbindID(t *testing.T) {
	d := NewDispatcher()
	var aCount, bCount int
	idA := d.Bind(platform.WindowID(1), KeyPressMask, func(ev *Event) { aCount++ })
	idB := d.Bind(platform.WindowID(1), KeyPressMask, func(ev *Event) { bCount++ })
	idG := d.BindGlobal(KeyPressMask, func(ev *Event) { aCount++ })

	// Sanity: all three fire.
	d.Dispatch(&Event{Type: KeyPressType, Window: platform.WindowID(1)})
	if aCount != 2 || bCount != 1 {
		t.Fatalf("baseline: aCount=%d bCount=%d, want 2/1", aCount, bCount)
	}

	// Remove the per-window handler A.
	if !d.UnbindID(idA) {
		t.Fatal("UnbindID(A) returned false")
	}
	d.Dispatch(&Event{Type: KeyPressType, Window: platform.WindowID(1)})
	if aCount != 3 || bCount != 2 {
		t.Errorf("after UnbindID(A): aCount=%d bCount=%d, want 3/2", aCount, bCount)
	}

	// Remove the global handler G; only B should remain.
	if !d.UnbindID(idG) {
		t.Fatal("UnbindID(G) returned false")
	}
	d.Dispatch(&Event{Type: KeyPressType, Window: platform.WindowID(1)})
	if aCount != 3 || bCount != 3 {
		t.Errorf("after UnbindID(G): aCount=%d bCount=%d, want 3/3", aCount, bCount)
	}

	// Removing an unknown ID is a no-op.
	if d.UnbindID(BindingID(99999)) {
		t.Error("UnbindID on unknown ID should return false")
	}

	// After removing A (idx 0) from [A, B], the swap-with-last removal
	// should leave B at idx 0; removing B should now find it there.
	if !d.UnbindID(idB) {
		t.Fatal("UnbindID(B) returned false after swap-with-last")
	}
	if d.UnbindID(idB) {
		t.Error("UnbindID on already-removed B should return false")
	}
}

func TestDispatcherRebindDuringDispatch(t *testing.T) {
	d := NewDispatcher()
	var calls []string
	var laterID BindingID
	d.Bind(1, ExposureMask, func(*Event) {
		calls = append(calls, "first")
		// Unbind a handler that has not run yet, then bind a handler on
		// another window; the freed registration used to be recycled,
		// making this dispatch call window 2's handler.
		d.UnbindID(laterID)
		d.Bind(2, ExposureMask, func(*Event) { calls = append(calls, "window2") })
	})
	laterID = d.Bind(1, ExposureMask, func(*Event) { calls = append(calls, "later") })

	d.Dispatch(&Event{Type: ExposeType, Window: 1})
	if len(calls) != 1 || calls[0] != "first" {
		t.Fatalf("calls = %v, want [first]", calls)
	}
}

func TestDispatcherUnbindIDKeepsOrder(t *testing.T) {
	d := NewDispatcher()
	var got []int
	ids := make([]BindingID, 5)
	for i := range ids {
		ids[i] = d.Bind(1, ExposureMask, func(*Event) { got = append(got, i) })
	}
	d.UnbindID(ids[1])
	d.Dispatch(&Event{Type: ExposeType, Window: 1})
	want := []int{0, 2, 3, 4}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("order = %v, want %v", got, want)
	}
}

func TestDispatcherConcurrentBind(t *testing.T) {
	d := NewDispatcher()
	d.Bind(1, ExposureMask, func(*Event) {})
	stop := make(chan struct{})
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			for {
				select {
				case <-stop:
					return
				default:
				}
				id := d.Bind(1, ExposureMask, func(*Event) {})
				gid := d.BindGlobal(ExposureMask, func(*Event) {})
				d.UnbindID(id)
				d.UnbindID(gid)
			}
		})
	}
	for range 10000 {
		d.Dispatch(&Event{Type: ExposeType, Window: 1})
	}
	close(stop)
	wg.Wait()
}

func TestDispatcherBindGlobalForUnbindsWithOwner(t *testing.T) {
	d := NewDispatcher()
	var owned, other int
	d.BindGlobalFor(platform.WindowID(5), ButtonPressMask, func(ev *Event) { owned++ })
	id := d.BindGlobalFor(platform.WindowID(6), ButtonPressMask, func(ev *Event) { other++ })

	d.Dispatch(&Event{Type: ButtonPressType, Window: platform.WindowID(99)})
	d.Unbind(platform.WindowID(5))
	d.Dispatch(&Event{Type: ButtonPressType, Window: platform.WindowID(99)})
	if owned != 1 || other != 2 {
		t.Fatalf("owned=%d other=%d, want 1 and 2", owned, other)
	}
	if !d.UnbindID(id) {
		t.Fatal("UnbindID of an owned global handler failed")
	}
	if len(d.global) != 0 || len(d.owned) != 0 || len(d.byID) != 0 {
		t.Errorf("leftover registrations: global=%d owned=%d byID=%d", len(d.global), len(d.owned), len(d.byID))
	}
}
