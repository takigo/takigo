package event

import (
	"sync"

	"github.com/msorc/takigo/platform"
)

// Handler is a function called when an event is received.
type Handler func(*Event)

// BindingID identifies a specific handler binding for later removal.
type BindingID uint64

// registration represents a single event handler binding. Registrations
// live on the heap so the Dispatcher can index them by pointer in a
// reverse map (byID) and use swap-with-last removal in the per-window
// slice without invalidating other handlers.
type registration struct {
	id      BindingID
	mask    Mask
	handler Handler
	window  platform.WindowID // 0 for global handlers
	idx     int               // position in handlers[window] or global
	global  bool
}

// Dispatcher manages event handler registration and dispatch per window.
type Dispatcher struct {
	mu       sync.RWMutex
	handlers map[platform.WindowID][]*registration
	global   []*registration // handlers for all windows
	byID     map[BindingID]*registration
	nextID   BindingID
}

// NewDispatcher creates a new event dispatcher.
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[platform.WindowID][]*registration),
		byID:     make(map[BindingID]*registration),
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
	reg := &registration{id: id, mask: mask, handler: h, window: w, idx: len(d.handlers[w])}
	d.handlers[w] = append(d.handlers[w], reg)
	d.byID[id] = reg
	return id
}

// BindGlobal registers an event handler for all windows.
func (d *Dispatcher) BindGlobal(mask Mask, h Handler) BindingID {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.nextID++
	id := d.nextID
	reg := &registration{id: id, mask: mask, handler: h, global: true, idx: len(d.global)}
	d.global = append(d.global, reg)
	d.byID[id] = reg
	return id
}

// Unbind removes all handlers for a specific window.
func (d *Dispatcher) Unbind(w platform.WindowID) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, r := range d.handlers[w] {
		delete(d.byID, r.id)
	}
	delete(d.handlers, w)
}

// UnbindID removes a specific handler by its BindingID. The reverse
// map makes this O(1) regardless of how many other handlers exist.
// Returns true if the handler was found and removed.
func (d *Dispatcher) UnbindID(id BindingID) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	reg, ok := d.byID[id]
	if !ok {
		return false
	}
	delete(d.byID, id)
	if reg.global {
		swapRemove(&d.global, reg)
	} else {
		regs := d.handlers[reg.window]
		swapRemove(&regs, reg)
		d.handlers[reg.window] = regs
	}
	return true
}

// swapRemove removes reg from *regs in O(1) by swapping it with the
// last element and shrinking the slice. The displaced element's
// stored idx is updated so future removals stay correct.
//
// Dispatch snapshots the pointer slice before invoking handlers, so
// an in-flight dispatch may still hold the old pointer and call it
// one extra time — acceptable, since the caller has already decided
// to remove the handler.
func swapRemove(regs *[]*registration, reg *registration) {
	list := *regs
	last := len(list) - 1
	if reg.idx != last {
		list[reg.idx] = list[last]
		list[reg.idx].idx = reg.idx
	}
	list[last] = nil
	*regs = list[:last]
}

// Dispatch sends an event to all matching handlers.
func (d *Dispatcher) Dispatch(ev *Event) {
	mask := TypeToMask(ev.Type)
	if mask == 0 {
		return
	}

	d.mu.RLock()
	// Copy the slice headers into local slices. Required because
	// UnbindID swap-removes elements under the write lock, which
	// would otherwise mutate the underlying array while Dispatch
	// iterates a snapshot.
	windowHandlers := append([]*registration(nil), d.handlers[ev.Window]...)
	globalHandlers := append([]*registration(nil), d.global...)
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
