package ttk

import (
	"strings"
	"sync"

	"github.com/msorc/takigo/option"
)

// Style holds defaults and state maps for a named widget style.
type Style struct {
	Name     string
	Parent   *Style
	Defaults map[string]any
	Maps     map[string]StateMap[any]
}

// Lookup returns the value for optionName at the given state.
// It checks state maps first, then defaults, walking up the parent chain.
func (s *Style) Lookup(optionName string, state State) (any, bool) {
	for cur := s; cur != nil; cur = cur.Parent {
		if m, ok := cur.Maps[optionName]; ok {
			if v, found := m.Lookup(state); found {
				return v, true
			}
		}
		if v, ok := cur.Defaults[optionName]; ok {
			return v, true
		}
	}
	return nil, false
}

// LookupColor returns a color pixel value from the style.
func LookupColor(s *Style, name string, state State, fallback uint64) uint64 {
	v, ok := s.Lookup(name, state)
	if !ok {
		return fallback
	}
	if pixel, ok := v.(uint64); ok {
		return pixel
	}
	return fallback
}

// LookupInt returns an integer value from the style.
func LookupInt(s *Style, name string, state State, fallback int) int {
	v, ok := s.Lookup(name, state)
	if !ok {
		return fallback
	}
	if n, ok := v.(int); ok {
		return n
	}
	return fallback
}

// LookupRelief returns a relief value from the style.
func LookupRelief(s *Style, name string, state State, fallback option.Relief) option.Relief {
	v, ok := s.Lookup(name, state)
	if !ok {
		return fallback
	}
	if r, ok := v.(option.Relief); ok {
		return r
	}
	return fallback
}

// LookupPadding returns a padding value from the style.
func LookupPadding(s *Style, name string, state State, fallback Padding) Padding {
	v, ok := s.Lookup(name, state)
	if !ok {
		return fallback
	}
	if p, ok := v.(Padding); ok {
		return p
	}
	return fallback
}

// Theme holds elements, styles, and layout templates for a visual theme.
type Theme struct {
	Name     string
	Parent   *Theme
	Elements map[string]ElementFactory
	Styles   map[string]*Style
	Layouts  map[string]*LayoutTemplate
}

// NewTheme creates a theme with the given name and optional parent.
func NewTheme(name string, parent *Theme) *Theme {
	return &Theme{
		Name:     name,
		Parent:   parent,
		Elements: make(map[string]ElementFactory),
		Styles:   make(map[string]*Style),
		Layouts:  make(map[string]*LayoutTemplate),
	}
}

// RegisterElement registers an element factory with this theme.
func (t *Theme) RegisterElement(name string, factory ElementFactory) {
	t.Elements[name] = factory
}

// GetElement returns the element factory for name, falling back to parent theme.
func (t *Theme) GetElement(name string) ElementFactory {
	for cur := t; cur != nil; cur = cur.Parent {
		if f, ok := cur.Elements[name]; ok {
			return f
		}
	}
	return func(*DrawContext) Element { return NullElement{} }
}

// GetStyle returns the style for name, creating it on demand.
// Parents to the same-named style in the parent theme if available,
// otherwise auto-parents via dot-separated naming: "TButton" → ".".
func (t *Theme) GetStyle(name string) *Style {
	if s, ok := t.Styles[name]; ok {
		return s
	}
	// Create on demand.
	s := &Style{
		Name:     name,
		Defaults: make(map[string]any),
		Maps:     make(map[string]StateMap[any]),
	}
	if name == "." {
		// Root style: parent to parent theme's root if available.
		if t.Parent != nil {
			if parentRoot, ok := t.Parent.Styles["."]; ok {
				s.Parent = parentRoot
			}
		}
	} else {
		// Try same-named style in parent theme first (cross-theme inheritance).
		if t.Parent != nil {
			if parentStyle, ok := t.Parent.Styles[name]; ok {
				s.Parent = parentStyle
			}
		}
		// If no parent theme style, fall back to local dot-separated parent.
		if s.Parent == nil {
			if dot := strings.LastIndex(name, "."); dot >= 0 {
				s.Parent = t.GetStyle(name[:dot])
			} else {
				s.Parent = t.GetStyle(".")
			}
		}
	}
	t.Styles[name] = s
	return s
}

// RegisterLayout registers a layout template for a widget class.
func (t *Theme) RegisterLayout(name string, tmpl *LayoutTemplate) {
	t.Layouts[name] = tmpl
}

// GetLayout returns the layout template for name, falling back to parent.
func (t *Theme) GetLayout(name string) *LayoutTemplate {
	for cur := t; cur != nil; cur = cur.Parent {
		if tmpl, ok := cur.Layouts[name]; ok {
			return tmpl
		}
	}
	return nil
}

// Package-level theme registry.
var (
	themesMu     sync.RWMutex
	themes       = make(map[string]*Theme)
	currentTheme *Theme
)

// RegisterTheme registers a theme in the global registry.
func RegisterTheme(t *Theme) {
	themesMu.Lock()
	defer themesMu.Unlock()
	themes[t.Name] = t
	if currentTheme == nil {
		currentTheme = t
	}
}

// SetCurrentTheme sets the active theme by name.
func SetCurrentTheme(name string) {
	themesMu.Lock()
	defer themesMu.Unlock()
	if t, ok := themes[name]; ok {
		currentTheme = t
	}
}

// CurrentTheme returns the active theme.
func CurrentTheme() *Theme {
	themesMu.RLock()
	defer themesMu.RUnlock()
	return currentTheme
}
