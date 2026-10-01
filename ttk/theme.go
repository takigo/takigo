package ttk

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/msorc/takigo/appearance"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// Style holds defaults and state maps for a named widget style.
type Style struct {
	Name     string
	Parent   *Style
	Defaults map[string]any
	Maps     map[string]StateMap[any]
	// fallback is the same-named style of the parent theme, consulted when
	// the chain has no value. Tk keeps such values as element defaults; the
	// port keeps them in the default theme's styles. ResolveStyle links it
	// lazily while other Apps' loops may be in Lookup, hence the atomic.
	fallback atomic.Pointer[Style]
}

// Lookup returns the value for optionName at the given state. Like
// Ttk_QueryStyle it consults the state maps along the whole parent chain
// (the first style that maps the option decides, as in Ttk_StyleMap), then
// the defaults along the chain, then the fallback style.
// Stops after 20 levels to guard against accidental cycles.
func (s *Style) Lookup(optionName string, state State) (any, bool) {
	const maxDepth = 20
	for style, fallbacks := s, 0; style != nil && fallbacks < maxDepth; style, fallbacks = style.fallback.Load(), fallbacks+1 {
		for cur, depth := style, 0; cur != nil && depth < maxDepth; cur, depth = cur.Parent, depth+1 {
			if m, ok := cur.Maps[optionName]; ok {
				if v, found := m.Lookup(state); found {
					return v, true
				}
				break
			}
		}
		for cur, depth := style, 0; cur != nil && depth < maxDepth; cur, depth = cur.Parent, depth+1 {
			if v, ok := cur.Defaults[optionName]; ok {
				return v, true
			}
		}
	}
	return nil, false
}

// LookupAs is Lookup for a value of type T; it reports false when the
// option is unset or holds a value of another type.
func (s *Style) LookupAs[T any](optionName string, state State) (T, bool) {
	v, _ := s.Lookup(optionName, state)
	t, ok := v.(T)
	return t, ok
}

// LookupColor returns a color pixel value from the style.
func LookupColor(s *Style, name string, state State, fallback uint64) uint64 {
	if pixel, ok := s.LookupAs[uint64](name, state); ok {
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
	case screenunit.Distance:
		return n.Pixels()
	}
	return fallback
}

// LookupRelief returns a relief value from the style.
func LookupRelief(s *Style, name string, state State, fallback option.Relief) option.Relief {
	if r, ok := s.LookupAs[option.Relief](name, state); ok {
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
	case screenunit.Distance:
		return UniformPadding(p.Pixels())
	case string:
		return ParsePadding(p)
	}
	return fallback
}

// Theme holds elements, styles, and layout templates for a visual theme.
//
// Themes are process-wide, shared by every App. Use the methods to reach
// Elements, Styles and Layouts: they hold themesMu, since GetStyle creates
// styles on demand from any App's loop goroutine. A style's Defaults and
// Maps are not locked: fill them in before widgets use the style.
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
	themesMu.Lock()
	defer themesMu.Unlock()
	t.Elements[name] = factory
}

// GetElement returns the element factory for name, falling back to parent theme.
func (t *Theme) GetElement(name string) ElementFactory {
	themesMu.RLock()
	defer themesMu.RUnlock()
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
	themesMu.Lock()
	defer themesMu.Unlock()
	return t.getStyle(name)
}

// getStyle is GetStyle for callers holding themesMu.
func (t *Theme) getStyle(name string) *Style {
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
		// Tk's style inheritance (ttkTheme.c Ttk_GetStyle): "a.b.c"
		// inherits from "b.c", the name minus its leading component.
		if _, after, ok := strings.Cut(name, "."); ok {
			s.Parent = t.getStyle(after)
		} else {
			s.Parent = t.getStyle(".")
		}
	}
	t.Styles[name] = s
	return s
}

// ResolveStyle returns the style for name in this theme, as Ttk_GetStyle
// does, so its parent chain ends at this theme's root and the theme's
// colours win. When a parent theme configures the same style, that style
// becomes its fallback.
func (t *Theme) ResolveStyle(name string) *Style {
	themesMu.Lock()
	defer themesMu.Unlock()
	return t.resolveStyle(name)
}

// resolveStyle is ResolveStyle for callers holding themesMu.
func (t *Theme) resolveStyle(name string) *Style {
	s := t.getStyle(name)
	if name != "." && s.fallback.Load() == nil {
		for cur := t.Parent; cur != nil; cur = cur.Parent {
			if _, ok := cur.Styles[name]; ok {
				s.fallback.Store(cur.resolveStyle(name))
				break
			}
		}
	}
	return s
}

// RegisterLayout registers a layout template for a widget class.
func (t *Theme) RegisterLayout(name string, tmpl *LayoutTemplate) {
	themesMu.Lock()
	defer themesMu.Unlock()
	t.Layouts[name] = tmpl
}

// GetLayout returns the layout template for name, falling back to parent.
// Like Ttk_CreateLayout, "a.b.c" falls back to the layout of "b.c", then "c".
func (t *Theme) GetLayout(name string) *LayoutTemplate {
	themesMu.RLock()
	defer themesMu.RUnlock()
	for {
		for cur := t; cur != nil; cur = cur.Parent {
			if tmpl, ok := cur.Layouts[name]; ok {
				return tmpl
			}
		}
		dot := strings.Index(name, ".")
		if dot < 0 {
			return nil
		}
		name = name[dot+1:]
	}
}

// Package-level theme registry. themesMu also guards every Theme's
// Elements, Styles and Layouts.
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

var (
	appThemeKey  = new(window.ValueKey)
	ttkWidgetKey = new(window.ValueKey)
)

// ErrUnknownTheme is returned by UseTheme for a name no theme is registered
// under.
var ErrUnknownTheme = errors.New("ttk: unknown theme")

// ThemeFor returns the theme app's widgets use: the one chosen with
// UseTheme, or the process default (CurrentTheme) when none was.
func ThemeFor(app widget.AppContext) *Theme {
	if app != nil {
		if root := app.Window(); root != nil {
			if t, ok := root.Value(appThemeKey).(*Theme); ok {
				return t
			}
		}
	}
	return CurrentTheme()
}

// UseTheme makes the named theme the one app's themed widgets use and
// re-themes the existing ones, like Tk's "ttk::style theme use". Other
// Apps in the process are not affected.
func UseTheme(app widget.AppContext, name string) error {
	themesMu.RLock()
	t, ok := themes[name]
	themesMu.RUnlock()
	if !ok {
		return fmt.Errorf("%w %q", ErrUnknownTheme, name)
	}
	root := app.Window()
	root.SetValue(appThemeKey, t)
	for win := range root.Descendants() {
		w, ok := win.Value(ttkWidgetKey).(*TtkWidget)
		if !ok || w.Destroyed {
			continue
		}
		w.RefreshTheme()
		if w.reconfigure != nil {
			w.reconfigure()
			continue
		}
		reqW, reqH := win.ReqWidth, win.ReqHeight
		w.updateReqFromLayout()
		if (win.ReqWidth != reqW || win.ReqHeight != reqH) && win.GeomManager != nil {
			win.GeomManager.RequestProc(win)
		}
		w.redisplay()
	}
	return nil
}

// UseSystemTheme gives app the light or the dark theme according to the
// desktop's appearance (see package appearance), and returns the name it
// chose. Both themes must be registered, e.g. "clam" and "dark" by
// importing ttk/clamtheme and ttk/darktheme. It samples the appearance
// once; call it again to follow a change.
func UseSystemTheme(app widget.AppContext, light, dark string) (string, error) {
	name := light
	if appearance.System() == appearance.Dark {
		name = dark
	}
	return name, UseTheme(app, name)
}

// CurrentTheme returns the process-wide default theme, which an App uses
// until UseTheme gives it its own.
func CurrentTheme() *Theme {
	themesMu.RLock()
	defer themesMu.RUnlock()
	return currentTheme
}

// LookupTheme returns the registered theme with the given name, or nil.
func LookupTheme(name string) *Theme {
	themesMu.RLock()
	defer themesMu.RUnlock()
	return themes[name]
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
