package bind

import "sync"

// HandlerFunc is a binding callback. It receives the event and returns true
// to stop further dispatch along the tag chain (break).
type HandlerFunc func(ev *EventData) bool

// EventData wraps the raw event with additional binding-specific context.
type EventData struct {
	Type     int // reserved for future use
	RawEvent any // *event.Event
}

// binding associates a parsed pattern sequence with its handler.
type binding struct {
	seq     Sequence
	handler HandlerFunc
}

// BindingTable stores tag-based bindings. A tag is a string that identifies
// a binding scope: widget path (".frame1.button1"), class name ("Button"),
// toplevel path ("."), or "all".
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

// Add registers a binding for a tag with the given pattern sequence and handler.
// Multiple bindings per tag are supported; they are appended in order.
func (bt *BindingTable) Add(tag string, seq Sequence, handler HandlerFunc) {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	bt.bindings[tag] = append(bt.bindings[tag], binding{seq: seq, handler: handler})
}

// Remove removes all bindings for a tag that match the given sequence.
// Matching is by pattern string equality.
func (bt *BindingTable) Remove(tag string, seq Sequence) {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	target := seq.String()
	list := bt.bindings[tag]
	n := 0
	for _, b := range list {
		if b.seq.String() != target {
			list[n] = b
			n++
		}
	}
	if n == 0 {
		delete(bt.bindings, tag)
	} else {
		bt.bindings[tag] = list[:n]
	}
}

// Lookup returns all bindings for a given tag.
func (bt *BindingTable) Lookup(tag string) []binding {
	bt.mu.RLock()
	defer bt.mu.RUnlock()

	src := bt.bindings[tag]
	if len(src) == 0 {
		return nil
	}
	result := make([]binding, len(src))
	copy(result, src)
	return result
}

// RemoveAll removes all bindings for a tag.
func (bt *BindingTable) RemoveAll(tag string) {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	delete(bt.bindings, tag)
}
