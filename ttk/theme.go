package ttk

import (
	"sort"
	"strings"
	"sync"

	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/screenunit"
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
// Stops after 20 levels to guard against accidental cycles.
func (s *Style) Lookup(optionName string, state State) (any, bool) {
	const maxDepth = 20
	for cur, depth := s, 0; cur != nil && depth < maxDepth; cur, depth = cur.Parent, depth+1 {
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
	switch n := v.(type) {
	case int:
		return n
	case string:
		return screenunit.Px(n)
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
	switch p := v.(type) {
	case Padding:
		return p
	case string:
		return ParsePadding(p)
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
// Named styles parent to the local root "."; the root "." parents
// to the parent theme's "." for cross-theme default inheritance.
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
		// Root style: parent to parent theme's root for default inheritance.
		if t.Parent != nil {
			if parentRoot, ok := t.Parent.Styles["."]; ok {
				s.Parent = parentRoot
			}
		}
	} else {
		// Named styles parent to local root via dot-separated naming.
		if dot := strings.LastIndex(name, "."); dot >= 0 {
			s.Parent = t.GetStyle(name[:dot])
		} else {
			s.Parent = t.GetStyle(".")
		}
	}
	t.Styles[name] = s
	return s
}

// ResolveStyle finds the best style for name by walking the theme chain.
// If the current theme has the style, use it. Otherwise check parent themes.
// If no theme has it, auto-create in the current theme.
func (t *Theme) ResolveStyle(name string) *Style {
	for cur := t; cur != nil; cur = cur.Parent {
		if s, ok := cur.Styles[name]; ok {
			return s
		}
	}
	return t.GetStyle(name)
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

// ThemeNames returns the sorted names of all registered themes.
func ThemeNames() []string {
	themesMu.RLock()
	defer themesMu.RUnlock()
	names := make([]string, 0, len(themes))
	for name := range themes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
