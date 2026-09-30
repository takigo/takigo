package geometry

import "github.com/msorc/takigo/window"

// Table is a geometry manager's table of per-window state, used like a
// map[*window.Window]V. The entries live on the windows (Window.Value), as
// Tk's live in per-display hash tables, so Apps running on different
// goroutines do not share a map. Declare one with new(Table[V]).
// Loop-only, like the windows.
type Table[V any] struct{ key window.ValueKey }

// Get returns w's entry and whether it has one.
func (t *Table[V]) Get(w *window.Window) (V, bool) {
	v, ok := w.Value(&t.key).(V)
	return v, ok
}

// Of returns w's entry, or the zero V.
func (t *Table[V]) Of(w *window.Window) V {
	v, _ := w.Value(&t.key).(V)
	return v
}

// Set stores w's entry.
func (t *Table[V]) Set(w *window.Window, v V) { w.SetValue(&t.key, v) }

// Delete removes w's entry.
func (t *Table[V]) Delete(w *window.Window) {
	if w != nil {
		w.SetValue(&t.key, nil)
	}
}
