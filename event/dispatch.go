package event

import (
	"slices"
	"sync"
	"sync/atomic"

	"github.com/takigo/takigo/platform"
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
	window  platform.WindowID // for a global handler, its owner (0 for none)
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
	owned    map[platform.WindowID][]*registration
	byID     map[BindingID]*registration
	nextID   BindingID
	chain    ChainFunc
}

// ChainFunc runs the Tk binding-tag chain for ev (bind.Engine). With
// runClass set it runs ev's window's own handlers, via DispatchWindow, at
// the class tag's position, so bindings on the widget's path run first and
// a break there, or bindtags without the class tag, suppresses them. It
// reports whether ev's window has a tag chain.
type ChainFunc func(ev *Event, runClass bool) bool

// SetChain installs the binding-tag chain; call it before Run.
//
// A window's handlers for input events (keys, buttons, motion,
// enter/leave) implement its class bindings, as library/*.tcl does in Tk,
// and run inside the chain. Its handlers for other events are Tk's event
// handlers (Tk_CreateEventHandler): they run before any binding.
func (d *Dispatcher) SetChain(fn ChainFunc) {
	d.mu.Lock()
	d.chain = fn
	d.mu.Unlock()
}

// isBindingEvent reports whether t is delivered to class behaviour through
// the binding chain rather than to event handlers.
func isBindingEvent(t Type) bool {
	switch t {
	case KeyPressType, KeyReleaseType, ButtonPressType, ButtonReleaseType,
		MotionType, EnterType, LeaveType:
		return true
	}
	return false
}

// NewDispatcher creates a new event dispatcher.
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[platform.WindowID][]*registration),
		owned:    make(map[platform.WindowID][]*registration),
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

// BindGlobalFor is BindGlobal for a handler that belongs to window owner:
// Unbind(owner), which runs when owner is destroyed, removes it too, so a
// widget's global handler does not outlive the widget.
func (d *Dispatcher) BindGlobalFor(owner platform.WindowID, mask Mask, h Handler) BindingID {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.nextID++
	reg := &registration{id: d.nextID, mask: mask, handler: h, window: owner, global: true}
	d.global = appendCOW(d.global, reg)
	if owner != 0 {
		d.owned[owner] = append(d.owned[owner], reg)
	}
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
	for _, r := range d.owned[w] {
		r.dead.Store(true)
		delete(d.byID, r.id)
		d.global = removeCOW(d.global, r)
	}
	delete(d.owned, w)
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
		if regs := slices.DeleteFunc(d.owned[reg.window], func(r *registration) bool { return r == reg }); len(regs) > 0 {
			d.owned[reg.window] = regs
		} else {
			delete(d.owned, reg.window)
		}
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
// handlers and the binding chain (see SetChain), then global handlers.
func (d *Dispatcher) Dispatch(ev *Event) {
	mask := TypeToMask(ev.Type)
	if mask == 0 {
		return
	}

	d.mu.RLock()
	windowHandlers := d.handlers[ev.Window]
	globalHandlers := d.global
	chain := d.chain
	d.mu.RUnlock()

	switch {
	case chain == nil:
		run(windowHandlers, ev, mask)
	case isBindingEvent(ev.Type):
		if !chain(ev, true) {
			run(windowHandlers, ev, mask)
		}
	default:
		run(windowHandlers, ev, mask)
		chain(ev, false)
	}
	for _, r := range globalHandlers {
		if r.mask&mask != 0 && !r.dead.Load() {
			r.handler(ev)
		}
	}
}

// DispatchWindow runs only the handlers bound to ev's window.
func (d *Dispatcher) DispatchWindow(ev *Event) {
	d.mu.RLock()
	windowHandlers := d.handlers[ev.Window]
	d.mu.RUnlock()
	run(windowHandlers, ev, TypeToMask(ev.Type))
}

func run(regs []*registration, ev *Event, mask Mask) {
	for _, r := range regs {
		if r.mask&mask != 0 && !r.dead.Load() {
			r.handler(ev)
		}
	}
}
