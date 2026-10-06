// Package widget - Variable provides a simple observable value for linking
// check/radio button groups. Single-threaded (event loop only).
package widget

import "slices"

// Variable is a generic observable value. Listeners are notified on Set.
// T must be comparable for change detection.
type Variable[T comparable] struct {
	value     T
	unset     bool
	listeners []*func(old, cur T)
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
		(*fn)(old, val)
	}
}

// OnChange registers a listener called when the value changes.
// Returns an unsubscribe function, which removes exactly this listener
// whatever order listeners are removed in and is safe to call from a
// listener or more than once.
func (v *Variable[T]) OnChange(fn func(old, cur T)) func() {
	p := &fn
	v.listeners = append(slices.Clip(v.listeners), p)
	return func() {
		if i := slices.Index(v.listeners, p); i >= 0 {
			v.listeners = slices.Delete(slices.Clone(v.listeners), i, i+1)
		}
	}
}
