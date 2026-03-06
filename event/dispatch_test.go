package event

import (
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
