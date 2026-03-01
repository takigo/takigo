package config

import (
	"fmt"
	"sync"
)

// Table provides runtime cget/configure access to widget options.
// It maps option names to getter/setter pairs operating on a target struct.
type Table struct {
	mu      sync.RWMutex
	entries map[string]*entry
	aliases map[string]string // alias -> canonical name
}

// Getter retrieves the current value of an option from a target.
type Getter func(target any) any

// Setter sets a value on the target. Returns the changed-option bitmask.
type Setter func(target any, value any) error

type entry struct {
	spec   Spec
	get    Getter
	set    Setter
	mask   int // bitmask indicating which aspect changed
}

// NewTable creates a new configuration table.
func NewTable() *Table {
	return &Table{
		entries: make(map[string]*entry),
		aliases: make(map[string]string),
	}
}

// Register adds an option to the table with its getter and setter.
// mask is a bitmask the widget can use to track which options changed.
func (t *Table) Register(spec Spec, get Getter, set Setter, mask int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.entries[spec.Name] = &entry{spec: spec, get: get, set: set, mask: mask}
	if spec.Alias != "" {
		t.aliases[spec.Alias] = spec.Name
	}
}

// resolve returns the canonical name for an option (resolving aliases).
func (t *Table) resolve(name string) string {
	if canonical, ok := t.aliases[name]; ok {
		return canonical
	}
	return name
}

// Cget returns the current value of a named option.
func (t *Table) Cget(target any, name string) (any, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	name = t.resolve(name)
	e, ok := t.entries[name]
	if !ok {
		return nil, fmt.Errorf("unknown option %q", name)
	}
	return e.get(target), nil
}

// Configure sets one or more options on the target.
// Returns a bitmask of changed option categories.
func (t *Table) Configure(target any, nameValues ...any) (int, error) {
	if len(nameValues)%2 != 0 {
		return 0, fmt.Errorf("configure requires name-value pairs")
	}
	t.mu.RLock()
	defer t.mu.RUnlock()

	changed := 0
	for i := 0; i < len(nameValues); i += 2 {
		name, ok := nameValues[i].(string)
		if !ok {
			return changed, fmt.Errorf("option name must be a string, got %T", nameValues[i])
		}
		name = t.resolve(name)
		e, ok := t.entries[name]
		if !ok {
			return changed, fmt.Errorf("unknown option %q", name)
		}
		if err := e.set(target, nameValues[i+1]); err != nil {
			return changed, fmt.Errorf("option %q: %w", name, err)
		}
		changed |= e.mask
	}
	return changed, nil
}

// Specs returns all registered option specs.
func (t *Table) Specs() []Spec {
	t.mu.RLock()
	defer t.mu.RUnlock()
	specs := make([]Spec, 0, len(t.entries))
	for _, e := range t.entries {
		specs = append(specs, e.spec)
	}
	return specs
}
