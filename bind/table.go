package bind

import (
	"slices"
	"sync"

	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/window"
)

// HandlerFunc is a binding callback. It receives the event and returns true
// to stop further dispatch along the tag chain (break).
type HandlerFunc func(ev *EventData) bool

// EventData is what a binding handler receives.
type EventData struct {
	// Event is the event that triggered the binding. For a virtual event
	// sent with GenerateEvent it is a VirtualType event carrying only the
	// window.
	Event *event.Event
	// Window is the window the binding fired for.
	Window *window.Window
}

// binding associates a parsed pattern sequence with its handler.
type binding struct {
	seq     Sequence
	key     string // seq.String(), computed once for matching on the hot path
	handler HandlerFunc
}

// BindingTable stores tag-based bindings. A tag is a string that identifies
// a binding scope: widget path (".frame1.button1"), class name ("Button"),
// toplevel path ("."), or "all".
//
// Each tag's list is copy-on-write: Add and Remove publish a new slice, so
// Lookup can hand out the current one without copying and a dispatch keeps
// a stable snapshot while its handlers rebind.
type BindingTable struct {
	mu       sync.RWMutex
	bindings map[string][]binding // tag → bindings list
}

// NewBindingTable creates an empty binding table.
func NewBindingTable() *BindingTable {
	return &BindingTable{
		bindings: make(map[string][]binding),
	}
}

// Add registers a binding for a tag with the given pattern sequence and
// handler. Like Tk_CreateBinding, a binding for the same sequence on the
// same tag is replaced; bindings for other sequences keep their order.
func (bt *BindingTable) Add(tag string, seq Sequence, handler HandlerFunc) {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	key := seq.String()
	nb := binding{seq: seq, key: key, handler: handler}
	list := bt.bindings[tag]
	if i := slices.IndexFunc(list, func(b binding) bool { return b.key == key }); i >= 0 {
		list = slices.Clone(list)
		list[i] = nb
		bt.bindings[tag] = list
		return
	}
	bt.bindings[tag] = append(slices.Clip(list), nb)
}

// Remove removes all bindings for a tag that match the given sequence.
// Matching is by pattern string equality.
func (bt *BindingTable) Remove(tag string, seq Sequence) {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	target := seq.String()
	var kept []binding
	for _, b := range bt.bindings[tag] {
		if b.key != target {
			kept = append(kept, b)
		}
	}
	if len(kept) == 0 {
		delete(bt.bindings, tag)
	} else {
		bt.bindings[tag] = kept
	}
}

// Lookup returns all bindings for a given tag. The slice is shared and
// must not be modified.
func (bt *BindingTable) Lookup(tag string) []binding {
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	return bt.bindings[tag]
}

// RemoveAll removes all bindings for a tag.
func (bt *BindingTable) RemoveAll(tag string) {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	delete(bt.bindings, tag)
}
