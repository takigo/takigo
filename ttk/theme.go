package ttk

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/takigo/takigo/appearance"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/window"
)

// Style holds defaults and state maps for a named widget style.
type Style struct {
	Name     string
	Parent   *Style
	Defaults map[string]any
	Maps     map[string]StateMap[any]

	// theme is the theme that owns the style; it is used to find the
	// cross-theme fallback (the style of the same name in the parent
	// theme).
	theme *Theme

	// parentTheme is set on a theme's root style ("."): the root style
	// inherits the parent theme's root style. The parent theme may create
	// its root style later, so the link is resolved on demand.
	parentTheme *Theme
	rootParent  atomic.Pointer[Style]

	// widgetLevel marks the widget-level style (ttk/widget.go widgetOpts):
	// its values are the widget's own options, which take precedence over
	// the style's state maps, as Tk's InitializeElementRecord does.
	widgetLevel bool

	// fallback caches the parent theme's styles that stand in for this
	// one. It is resolved lazily, and re-resolved when an ancestor theme
	// gains a style, so a style registered after the child theme still
	// becomes the fallback.
	fallback atomic.Pointer[fallbackSet]
}

// fallbackSet is a cached list of fallback styles together with the
// style counters of the themes they come from.
type fallbackSet struct {
	styles []*Style
	gens   []uint64
}

// Lookup returns the value for optionName at the given state. Like
// Ttk_QueryStyle it consults the widget's own option first (only for the
// widget-level style), then the state maps along the whole parent chain
// (the first style that maps the option decides, as in Ttk_StyleMap), then
// the defaults along the chain, then the fallback styles of the parent
// themes.
// Stops after 20 levels to guard against accidental cycles.
func (s *Style) Lookup(optionName string, state State) (any, bool) {
	if s == nil {
		return nil, false
	}
	// The widget's own option value wins over the style's state maps.
	if s.widgetLevel {
		if v, ok := s.Defaults[optionName]; ok {
			return v, true
		}
		return s.Parent.Lookup(optionName, state)
	}
	if v, ok := s.lookupChain(optionName, state); ok {
		return v, true
	}
	for _, fb := range s.fallbackStyles() {
		if v, ok := fb.lookupChain(optionName, state); ok {
			return v, true
		}
	}
	return nil, false
}

// lookupChain looks the option up in the style and its parent chain: the
// state maps first, then the defaults.
func (s *Style) lookupChain(optionName string, state State) (any, bool) {
	const maxDepth = 20
	for cur, depth := s, 0; cur != nil && depth < maxDepth; cur, depth = cur.parentStyle(), depth+1 {
		if m, ok := cur.Maps[optionName]; ok {
			if v, found := m.Lookup(state); found {
				return v, true
			}
			break
		}
	}
	for cur, depth := s, 0; cur != nil && depth < maxDepth; cur, depth = cur.parentStyle(), depth+1 {
		if v, ok := cur.Defaults[optionName]; ok {
			return v, true
		}
	}
	return nil, false
}

// parentStyle returns the style this one inherits from: the explicit
// Parent, or the parent theme's root style for a theme's root style.
func (s *Style) parentStyle() *Style {
	if s.Parent != nil {
		return s.Parent
	}
	if s.parentTheme == nil {
		return nil
	}
	if p := s.rootParent.Load(); p != nil {
		return p
	}
	themesMu.RLock()
	p := s.parentTheme.Styles["."]
	themesMu.RUnlock()
	if p != nil {
		s.rootParent.Store(p)
	}
	return p
}

// fallbackStyles returns the styles of the ancestor themes that stand in
// for this one: the same name first, then the shorter suffixes of the
// name, as Tk's style naming does ("My.TButton" falls back to "TButton").
// They are resolved on every lookup, so a style that the parent theme
// registers after the child theme still becomes the fallback.
func (s *Style) fallbackStyles() []*Style {
	if s.theme == nil {
		return nil
	}
	if s.theme.Parent == nil {
		return nil
	}
	// The cached set is checked while walking the chain, so that a lookup
	// that hits it allocates nothing.
	if c := s.fallback.Load(); c != nil {
		i := 0
		fresh := true
		for th := s.theme.Parent; th != nil; th = th.Parent {
			if i >= len(c.gens) || th.stylesGen.Load() != c.gens[i] {
				fresh = false
				break
			}
			i++
		}
		if fresh && i == len(c.gens) {
			return c.styles
		}
	}
	ancestors := s.theme.ancestorThemes()
	themesMu.RLock()
	defer themesMu.RUnlock()
	c := &fallbackSet{gens: make([]uint64, 0, len(ancestors))}
	for _, th := range ancestors {
		c.gens = append(c.gens, th.stylesGen.Load())
		for _, name := range styleNameSuffixes(s.Name) {
			if st, ok := th.Styles[name]; ok {
				c.styles = append(c.styles, st)
			}
		}
	}
	s.fallback.Store(c)
	return c.styles
}

// styleNameSuffixes returns the name and its shorter suffixes, the way
// Tk's style naming splits "a.b.c" into "b.c" and "c".
func styleNameSuffixes(name string) []string {
	out := []string{name}
	for range 8 {
		_, after, ok := strings.Cut(out[len(out)-1], ".")
		if !ok || after == "" {
			break
		}
		out = append(out, after)
	}
	return out
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
		if pad, err := ParsePadding(p); err == nil {
			return pad
		}
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

	// stylesGen counts the styles created in this theme; a style's cached
	// fallback list is stale once an ancestor theme's counter changed.
	stylesGen atomic.Uint64
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
		theme:    t,
	}
	if name == "." {
		// Root style: it inherits the parent theme's root style, which is
		// resolved on demand because the parent theme may create its root
		// style after this theme exists.
		s.parentTheme = t.Parent
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
	t.stylesGen.Add(1)
	return s
}

// ancestorThemes returns the parent themes of this theme, nearest first.
func (t *Theme) ancestorThemes() []*Theme {
	var out []*Theme
	for cur := t.Parent; cur != nil; cur = cur.Parent {
		out = append(out, cur)
	}
	return out
}

// ResolveStyle returns the style for name in this theme, as Ttk_GetStyle
// does, so its parent chain ends at this theme's root and the theme's
// colours win. When a parent theme configures the same style, that style
// is the fallback, which the style resolves when it is looked up.
func (t *Theme) ResolveStyle(name string) *Style {
	themesMu.Lock()
	defer themesMu.Unlock()
	return t.resolveStyle(name)
}

// resolveStyle is ResolveStyle for callers holding themesMu.
func (t *Theme) resolveStyle(name string) *Style {
	return t.getStyle(name)
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
		if root := app.Root(); root != nil {
			if t, ok := root.Value(appThemeKey).(*Theme); ok {
				return t
			}
		}
		// A dark App (takigo.UseAppearance) gets the dark theme without
		// asking, when one is registered.
		if widget.PaletteFor(app).Dark {
			if t := LookupTheme("dark"); t != nil {
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
	root := app.Root()
	root.SetValue(appThemeKey, t)
	for win := range root.Descendants() {
		w, ok := win.Value(ttkWidgetKey).(*TtkWidget)
		if !ok || w.Destroyed() {
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
