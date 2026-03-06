package event

import (
	"sync"

	"github.com/msorc/takigo/platform"
)

// Handler is a function called when an event is received.
type Handler func(*Event)

// registration represents a single event handler binding.
type registration struct {
	mask    Mask
	handler Handler
}

// Dispatcher manages event handler registration and dispatch per window.
type Dispatcher struct {
	mu       sync.RWMutex
	handlers map[platform.WindowID][]registration
	global   []registration // handlers for all windows
}

// NewDispatcher creates a new event dispatcher.
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[platform.WindowID][]registration),
	}
}

// Bind registers an event handler for a specific window and event mask.
func (d *Dispatcher) Bind(w platform.WindowID, mask Mask, h Handler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers[w] = append(d.handlers[w], registration{mask: mask, handler: h})
}

// BindGlobal registers an event handler for all windows.
func (d *Dispatcher) BindGlobal(mask Mask, h Handler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.global = append(d.global, registration{mask: mask, handler: h})
}

// Unbind removes all handlers for a specific window.
func (d *Dispatcher) Unbind(w platform.WindowID) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.handlers, w)
}

// Dispatch sends an event to all matching handlers.
func (d *Dispatcher) Dispatch(ev *Event) {
	mask := TypeToMask(ev.Type)
	if mask == 0 {
		return
	}

	d.mu.RLock()
	windowHandlers := d.handlers[ev.Window]
	globalHandlers := d.global
	d.mu.RUnlock()

	for _, r := range windowHandlers {
		if r.mask&mask != 0 {
			r.handler(ev)
		}
	}
	for _, r := range globalHandlers {
		if r.mask&mask != 0 {
			r.handler(ev)
		}
	}
}
