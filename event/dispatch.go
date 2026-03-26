package event

import (
	"sync"

	"github.com/msorc/takigo/platform"
)

// Handler is a function called when an event is received.
type Handler func(*Event)

// BindingID identifies a specific handler binding for later removal.
type BindingID uint64

// registration represents a single event handler binding.
type registration struct {
	id      BindingID
	mask    Mask
	handler Handler
}

// Dispatcher manages event handler registration and dispatch per window.
type Dispatcher struct {
	mu       sync.RWMutex
	handlers map[platform.WindowID][]registration
	global   []registration // handlers for all windows
	nextID   BindingID
}

// NewDispatcher creates a new event dispatcher.
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[platform.WindowID][]registration),
	}
}

// Bind registers an event handler for a specific window and event mask.
// Returns a BindingID that can be passed to UnbindID to remove this
// specific handler without affecting other handlers on the same window.
func (d *Dispatcher) Bind(w platform.WindowID, mask Mask, h Handler) BindingID {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.nextID++
	id := d.nextID
	d.handlers[w] = append(d.handlers[w], registration{id: id, mask: mask, handler: h})
	return id
}

// BindGlobal registers an event handler for all windows.
func (d *Dispatcher) BindGlobal(mask Mask, h Handler) BindingID {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.nextID++
	id := d.nextID
	d.global = append(d.global, registration{id: id, mask: mask, handler: h})
	return id
}

// Unbind removes all handlers for a specific window.
func (d *Dispatcher) Unbind(w platform.WindowID) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.handlers, w)
}

// UnbindID removes a specific handler by its BindingID.
// Returns true if the handler was found and removed.
func (d *Dispatcher) UnbindID(id BindingID) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Search window handlers.
	for w, regs := range d.handlers {
		for i, r := range regs {
			if r.id == id {
				d.handlers[w] = append(regs[:i], regs[i+1:]...)
				return true
			}
		}
	}

	// Search global handlers.
	for i, r := range d.global {
		if r.id == id {
			d.global = append(d.global[:i], d.global[i+1:]...)
			return true
		}
	}

	return false
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
