package widget

import (
	"log"
	"strings"

	"github.com/msorc/takigo/color"
)

// WidgetBase returns the embedded Base; Configure reaches the common
// fields of any widget through it.
func (b *Base) WidgetBase() *Base { return b }

func (b *Base) logPrefix() string {
	if b.Win != nil && b.Win.Class != "" {
		return strings.ToLower(b.Win.Class)
	}
	return "widget"
}

// LookupColor resolves a colour name; on failure it logs and reports false
// so the caller keeps its previous value, as every option setter does.
func (b *Base) LookupColor(name string) (*color.Color, bool) {
	col, err := b.App.ColorCache().Get(name)
	if err != nil {
		log.Printf("%s: failed to get color %q: %v", b.logPrefix(), name, err)
		return nil, false
	}
	return col, true
}

// SetBackgroundName sets -background from a colour name and rebuilds the
// 3D border. It reports whether the colour was found.
func (b *Base) SetBackgroundName(name string) bool {
	col, ok := b.LookupColor(name)
	if !ok {
		return false
	}
	b.Background = col
	b.UpdateBorder()
	return true
}

// SetForegroundName sets -foreground from a colour name.
func (b *Base) SetForegroundName(name string) bool {
	col, ok := b.LookupColor(name)
	if ok {
		b.Foreground = col
	}
	return ok
}

// SetHighlightBackgroundName sets -highlightbackground from a colour name.
func (b *Base) SetHighlightBackgroundName(name string) bool {
	col, ok := b.LookupColor(name)
	if ok {
		b.HighlightBackground = col
	}
	return ok
}

// SetHighlightColorName sets -highlightcolor from a colour name.
func (b *Base) SetHighlightColorName(name string) bool {
	col, ok := b.LookupColor(name)
	if ok {
		b.HighlightColor = col
	}
	return ok
}

// SetFontName sets -font from a font name or descriptor. It reports
// whether the font was found.
func (b *Base) SetFontName(name string) bool {
	f, err := b.App.FontRegistry().Get(name)
	if err != nil {
		log.Printf("%s: failed to get font %q: %v", b.logPrefix(), name, err)
		return false
	}
	b.Font = f
	return true
}
