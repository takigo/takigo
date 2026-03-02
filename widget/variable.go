// Package widget - Variable provides a simple observable value for linking
// check/radio button groups. Single-threaded (event loop only).
package widget

// Variable is a generic observable value. Listeners are notified on Set.
// T must be comparable for change detection.
type Variable[T comparable] struct {
	value     T
	listeners []func(old, new T)
}

// NewVariable creates a Variable with the given initial value.
func NewVariable[T comparable](initial T) *Variable[T] {
	return &Variable[T]{value: initial}
}

// Get returns the current value.
func (v *Variable[T]) Get() T {
	return v.value
}

// Set updates the value and notifies listeners if changed.
func (v *Variable[T]) Set(val T) {
	if val == v.value {
		return
	}
	old := v.value
	v.value = val
	for _, fn := range v.listeners {
		fn(old, val)
	}
}

// OnChange registers a listener called when the value changes.
// Returns an unsubscribe function.
func (v *Variable[T]) OnChange(fn func(old, new T)) func() {
	v.listeners = append(v.listeners, fn)
	idx := len(v.listeners) - 1
	return func() {
		if idx < len(v.listeners) {
			v.listeners = append(v.listeners[:idx], v.listeners[idx+1:]...)
		}
	}
}
