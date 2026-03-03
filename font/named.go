package font

import (
	"sync"
)

// Named font constants matching Tk's default named fonts.
const (
	TkDefaultFont      = "TkDefaultFont"
	TkTextFont         = "TkTextFont"
	TkFixedFont        = "TkFixedFont"
	TkMenuFont         = "TkMenuFont"
	TkHeadingFont      = "TkHeadingFont"
	TkCaptionFont      = "TkCaptionFont"
	TkSmallCaptionFont = "TkSmallCaptionFont"
	TkIconFont         = "TkIconFont"
	TkTooltipFont      = "TkTooltipFont"
)

// namedFontDefs maps named font names to their default attributes.
var namedFontDefs = map[string]Attributes{
	TkDefaultFont:      {Family: "sans-serif", Size: 10, Weight: WeightNormal, Slant: SlantRoman},
	TkTextFont:         {Family: "sans-serif", Size: 10, Weight: WeightNormal, Slant: SlantRoman},
	TkFixedFont:        {Family: "monospace", Size: 10, Weight: WeightNormal, Slant: SlantRoman},
	TkMenuFont:         {Family: "sans-serif", Size: 10, Weight: WeightNormal, Slant: SlantRoman},
	TkHeadingFont:      {Family: "sans-serif", Size: 12, Weight: WeightBold, Slant: SlantRoman},
	TkCaptionFont:      {Family: "sans-serif", Size: 10, Weight: WeightBold, Slant: SlantRoman},
	TkSmallCaptionFont: {Family: "sans-serif", Size: 9, Weight: WeightNormal, Slant: SlantRoman},
	TkIconFont:         {Family: "sans-serif", Size: 9, Weight: WeightNormal, Slant: SlantRoman},
	TkTooltipFont:      {Family: "sans-serif", Size: 9, Weight: WeightNormal, Slant: SlantRoman},
}

// FontOpener is the interface for platform-specific font creation.
// The X11 backend implements this using OpenXft.
type FontOpener interface {
	OpenFont(attrs Attributes) (Font, error)
}

// Registry manages named fonts and caches opened font handles.
type Registry struct {
	mu     sync.Mutex
	opener FontOpener
	named  map[string]Attributes
	cache  map[string]Font
}

// NewRegistry creates a new font registry with the given font opener.
func NewRegistry(opener FontOpener) *Registry {
	reg := &Registry{
		opener: opener,
		named:  make(map[string]Attributes),
		cache:  make(map[string]Font),
	}

	// Register default named fonts.
	for name, attrs := range namedFontDefs {
		reg.named[name] = attrs
	}

	return reg
}

// Define registers or updates a named font with the given attributes.
func (r *Registry) Define(name string, attrs Attributes) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.named[name] = attrs
	// Invalidate cached font for this name.
	if f, ok := r.cache[name]; ok {
		f.Close()
		delete(r.cache, name)
	}
}

// Get returns a font by name. If name is a registered named font,
// it returns that. Otherwise it parses the name as a font descriptor.
func (r *Registry) Get(name string) (Font, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check cache first.
	if f, ok := r.cache[name]; ok {
		return f, nil
	}

	// Check named fonts.
	var attrs Attributes
	if a, ok := r.named[name]; ok {
		attrs = a
	} else {
		// Parse as descriptor.
		var err error
		attrs, err = ParseDescriptor(name)
		if err != nil {
			return nil, err
		}
	}

	// Open via platform-specific opener.
	f, err := r.opener.OpenFont(attrs)
	if err != nil {
		return nil, err
	}

	r.cache[name] = f
	return f, nil
}

// Close releases all cached fonts.
func (r *Registry) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, f := range r.cache {
		f.Close()
	}
	r.cache = make(map[string]Font)
}
