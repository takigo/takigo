package event

import (
	"sync"
	"sync/atomic"

	"github.com/msorc/takigo/platform"
)

// Handler is a function called when an event is received.
type Handler func(*Event)

// BindingID identifies a specific handler binding for later removal.
type BindingID uint64

// registration represents a single event handler binding. Fields other
// than dead are never modified after Bind, so Dispatch can read them
// without holding the lock.
type registration struct {
	id      BindingID
	mask    Mask
	handler Handler
	window  platform.WindowID // 0 for global handlers
	global  bool
	dead    atomic.Bool // set by Unbind/UnbindID; Dispatch skips it
}

// Dispatcher manages event handler registration and dispatch per window.
//
// The handler slices are copy-on-write: Bind and UnbindID publish a new
// slice instead of modifying the old one, so Dispatch iterates a
// consistent snapshot without copying it, and handlers keep the order
// they were bound in.
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
	reg := &registration{id: d.nextID, mask: mask, handler: h, window: w}
	d.handlers[w] = appendCOW(d.handlers[w], reg)
	d.byID[reg.id] = reg
	return reg.id
}

// BindGlobal registers an event handler for all windows.
func (d *Dispatcher) BindGlobal(mask Mask, h Handler) BindingID {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.nextID++
	reg := &registration{id: d.nextID, mask: mask, handler: h, global: true}
	d.global = appendCOW(d.global, reg)
	d.byID[reg.id] = reg
	return reg.id
}

// Unbind removes all handlers for a specific window. Handlers removed
// while an event is being dispatched are not called for it.
func (d *Dispatcher) Unbind(w platform.WindowID) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, r := range d.handlers[w] {
		r.dead.Store(true)
		delete(d.byID, r.id)
	}
	delete(d.handlers, w)
}

// UnbindID removes a specific handler by its BindingID.
// Returns true if the handler was found and removed.
func (d *Dispatcher) UnbindID(id BindingID) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	reg, ok := d.byID[id]
	if !ok {
		return false
	}
	delete(d.byID, id)
	reg.dead.Store(true)
	if reg.global {
		d.global = removeCOW(d.global, reg)
	} else if regs := removeCOW(d.handlers[reg.window], reg); len(regs) > 0 {
		d.handlers[reg.window] = regs
	} else {
		delete(d.handlers, reg.window)
	}
	return true
}

// appendCOW appends reg to regs. Writing past len(regs) is safe even if
// an in-flight Dispatch holds regs: its snapshot's length excludes the
// new slot, and removals always build a fresh backing array.
func appendCOW(regs []*registration, reg *registration) []*registration {
	return append(regs, reg)
}

// removeCOW returns a new slice holding regs without reg, keeping order.
func removeCOW(regs []*registration, reg *registration) []*registration {
	out := make([]*registration, 0, len(regs))
	for _, r := range regs {
		if r != reg {
			out = append(out, r)
		}
	}
	return out
}

// Dispatch sends an event to all matching handlers: the window's own
// handlers first, then global ones.
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
		if r.mask&mask != 0 && !r.dead.Load() {
			r.handler(ev)
		}
	}
	for _, r := range globalHandlers {
		if r.mask&mask != 0 && !r.dead.Load() {
			r.handler(ev)
		}
	}
}
