package bind

import (
	"slices"

	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/window"
)

// NewEngine creates a new binding engine for the given display.
func NewEngine(d *window.Display) *Engine {
	e := &Engine{
		table:         NewBindingTable(),
		display:       d,
		tags:          make(map[platform.WindowID]*tagInfo),
		virtualEvents: make(map[string][]Sequence),
	}
	installDefaultVirtualEvents(e)
	return e
}

// Install makes the engine the dispatcher's binding-tag chain. Every
// window gets Tk's default bindtags (path, class, toplevel, "all") unless
// RegisterWindow or SetBindTags gave it others. A widget's own input
// handlers (Dispatcher.Bind) run at its class tag, so a binding on its path
// runs first and can suppress them by returning true (break), as a Tcl
// script does against library/*.tcl class bindings; see Dispatcher.SetChain.
func (e *Engine) Install(dispatcher *event.Dispatcher) {
	e.dispatcher = dispatcher
	dispatcher.SetChain(e.dispatch)
}

// SequenceSpec is what Bind takes as the event to bind: a Tk pattern string
// ("<Control-s>", "<Double-Button-1>", "<<Copy>>") or a Sequence built
// with Key, Button, Virtual and friends.
type SequenceSpec interface {
	string | Sequence
}

func sequenceOf[S SequenceSpec](spec S) (Sequence, error) {
	if seq, ok := any(spec).(Sequence); ok {
		return seq, nil
	}
	return Parse(any(spec).(string))
}

// Bind adds a binding for a tag (widget path, class name, or "all").
func (e *Engine) Bind[S SequenceSpec](tag string, spec S, handler HandlerFunc) error {
	seq, err := sequenceOf(spec)
	if err != nil {
		return err
	}
	e.table.Add(tag, seq, handler)
	return nil
}

// BindWindow adds a binding for one widget: it binds the widget's path
// tag, the first in its tag chain, so the handler runs before the class
// bindings and can stop them by returning true.
func (e *Engine) BindWindow[S SequenceSpec](w window.Windower, spec S, handler HandlerFunc) error {
	return e.Bind(w.Window().PathName, spec, handler)
}

// Unbind removes all bindings for a tag that match the given sequence.
func (e *Engine) Unbind[S SequenceSpec](tag string, spec S) error {
	seq, err := sequenceOf(spec)
	if err != nil {
		return err
	}
	e.table.Remove(tag, seq)
	return nil
}

// UnbindWindow removes a widget's bindings for the given sequence.
func (e *Engine) UnbindWindow[S SequenceSpec](w window.Windower, spec S) error {
	return e.Unbind(w.Window().PathName, spec)
}

// RegisterWindow registers a window with the binding engine, establishing
// its tag chain for dispatch. className is the widget class (e.g. "Button").
func (e *Engine) RegisterWindow(w *window.Window, className string) {
	if w.PlatformID == platform.WindowID(0) {
		return
	}
	info := &tagInfo{
		win:       w,
		className: className,
		tags:      buildTagChain(w, className),
	}
	e.tags[w.PlatformID] = info
}

// UnregisterWindow removes a window from the binding engine along with
// the bindings on its path name, as Tk_DestroyWindow does via
// Tk_DeleteAllBindings, so a later window with the same path starts clean.
func (e *Engine) UnregisterWindow(w *window.Window) {
	delete(e.tags, w.PlatformID)
	if w.PathName != "" {
		e.table.RemoveAll(w.PathName)
	}
}

// BindTags returns the current tag chain for a window.
func (e *Engine) BindTags(w *window.Window) []string {
	info := e.tagInfoFor(w)
	if info == nil {
		return nil
	}
	result := make([]string, len(info.tags))
	copy(result, info.tags)
	return result
}

// SetBindTags replaces the tag chain for a window, like "bindtags w tags".
// Leaving out the window's class tag disables its class behaviour.
func (e *Engine) SetBindTags(w *window.Window, tags []string) {
	info := e.tagInfoFor(w)
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
		e.addVirtual(virtual, seq)
	}
	return nil
}

// addVirtual appends seq to the definition of virtual.
func (e *Engine) addVirtual(virtual string, seq Sequence) {
	if _, ok := e.virtualEvents[virtual]; !ok {
		e.virtualOrder = append(e.virtualOrder, virtual)
	}
	e.virtualEvents[virtual] = append(e.virtualEvents[virtual], seq)
	e.indexVirtuals()
}

// RemoveVirtualEvent removes a virtual event definition.
func (e *Engine) RemoveVirtualEvent(virtual string) {
	delete(e.virtualEvents, virtual)
	e.virtualOrder = slices.DeleteFunc(e.virtualOrder, func(n string) bool { return n == virtual })
	e.indexVirtuals()
}

// GenerateEvent dispatches a virtual event to a window as if it had occurred.
func (e *Engine) GenerateEvent(w *window.Window, virtual string) {
	info := e.tagInfoFor(w)
	if info == nil {
		return
	}

	for _, tag := range info.tags {
		bindings := e.table.Lookup(tag)
		for i := range bindings {
			b := &bindings[i]
			if len(b.seq.Patterns) == 1 && b.seq.Patterns[0].Virtual == virtual {
				ed := &EventData{
					Event:  &event.Event{Type: event.VirtualType, Window: w.PlatformID},
					Window: w,
				}
				if b.handler(ed) {
					return // break
				}
			}
		}
	}
}
