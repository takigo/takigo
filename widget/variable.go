// Package widget - Variable provides a simple observable value for linking
// check/radio button groups. Single-threaded (event loop only).
package widget

// Variable is a generic observable value. Listeners are notified on Set.
// T must be comparable for change detection.
type Variable[T comparable] struct {
	value     T
	unset     bool
	listeners []func(old, new T)
}

// NewVariable creates a Variable with the given initial value.
func NewVariable[T comparable](initial T) *Variable[T] {
	return &Variable[T]{value: initial}
}

// NewUnsetVariable creates a Variable that, like a Tcl variable that does not
// exist yet, holds no value until the first Set. Widgets linked to it treat
// that as Tk does (e.g. a radiobutton creates it as "").
func NewUnsetVariable[T comparable]() *Variable[T] {
	return &Variable[T]{unset: true}
}

// IsSet reports whether the variable has been given a value.
func (v *Variable[T]) IsSet() bool {
	return !v.unset
}

// Get returns the current value.
func (v *Variable[T]) Get() T {
	return v.value
}

// Set updates the value and notifies listeners if changed.
func (v *Variable[T]) Set(val T) {
	if val == v.value && !v.unset {
		return
	}
	v.unset = false
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
