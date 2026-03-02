package bind

import (
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/window"
)

// NewEngine creates a new binding engine for the given display.
func NewEngine(d *window.Display) *Engine {
	e := &Engine{
		table:         NewBindingTable(),
		display:       d,
		tags:          make(map[xlib.Window]*tagInfo),
		virtualEvents: make(map[string][]Sequence),
	}
	installDefaultVirtualEvents(e)
	return e
}

// Install hooks the binding engine into the event dispatcher as a global
// handler. The engine fires AFTER per-window handlers registered via
// Dispatcher.Bind(), so existing widget bindings are unaffected.
func (e *Engine) Install(dispatcher *event.Dispatcher) {
	dispatcher.BindGlobal(event.AllEventsMask, func(ev *event.Event) {
		e.dispatch(ev)
	})
}

// Bind adds a binding for a tag (widget path, class name, or "all").
// The pattern is a Tk-style event pattern string.
func (e *Engine) Bind(tag, pattern string, handler HandlerFunc) error {
	seq, err := Parse(pattern)
	if err != nil {
		return err
	}
	e.table.Add(tag, seq, handler)
	return nil
}

// Unbind removes all bindings for a tag that match the given pattern.
func (e *Engine) Unbind(tag, pattern string) error {
	seq, err := Parse(pattern)
	if err != nil {
		return err
	}
	e.table.Remove(tag, seq)
	return nil
}

// RegisterWindow registers a window with the binding engine, establishing
// its tag chain for dispatch. className is the widget class (e.g. "Button").
func (e *Engine) RegisterWindow(w *window.Window, className string) {
	if w.XWindow == xlib.Window(0) {
		return
	}
	info := &tagInfo{
		win:       w,
		className: className,
		tags:      buildTagChain(w, className),
	}
	e.tags[w.XWindow] = info
}

// UnregisterWindow removes a window from the binding engine.
func (e *Engine) UnregisterWindow(w *window.Window) {
	delete(e.tags, w.XWindow)
}

// BindTags returns the current tag chain for a window.
func (e *Engine) BindTags(w *window.Window) []string {
	info := e.tags[w.XWindow]
	if info == nil {
		return nil
	}
	result := make([]string, len(info.tags))
	copy(result, info.tags)
	return result
}

// SetBindTags replaces the tag chain for a window.
func (e *Engine) SetBindTags(w *window.Window, tags []string) {
	info := e.tags[w.XWindow]
	if info == nil {
		return
	}
	info.tags = make([]string, len(tags))
	copy(info.tags, tags)
}

// AddVirtualEvent defines a virtual event that maps to one or more physical
// patterns. For example: AddVirtualEvent("Copy", "<Control-c>").
func (e *Engine) AddVirtualEvent(virtual string, patterns ...string) error {
	for _, p := range patterns {
		seq, err := Parse(p)
		if err != nil {
			return err
		}
		e.virtualEvents[virtual] = append(e.virtualEvents[virtual], seq)
	}
	return nil
}

// RemoveVirtualEvent removes a virtual event definition.
func (e *Engine) RemoveVirtualEvent(virtual string) {
	delete(e.virtualEvents, virtual)
}

// GenerateEvent dispatches a virtual event to a window as if it had occurred.
func (e *Engine) GenerateEvent(w *window.Window, virtual string) {
	info := e.tags[w.XWindow]
	if info == nil {
		return
	}

	for _, tag := range info.tags {
		bindings := e.table.Lookup(tag)
		for i := range bindings {
			b := &bindings[i]
			if len(b.seq.Patterns) == 1 && b.seq.Patterns[0].Virtual == virtual {
				ed := &EventData{}
				if b.handler(ed) {
					return // break
				}
			}
		}
	}
}
