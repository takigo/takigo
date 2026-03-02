package text

import (
	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/font"
)

// Tag defines display attributes that can be applied to text ranges.
type Tag struct {
	Name       string
	Priority   int
	Foreground *color.Color
	Background *color.Color
	Font       font.Font
	Underline  bool
	Overstrike bool
}

// TagRange associates a tag name with a half-open index range [Start, End).
type TagRange struct {
	TagName string
	Start   Index
	End     Index
}

// TagOption configures a Tag.
type TagOption func(cache *color.Cache, reg *font.Registry, tag *Tag)

// TagForeground sets the tag's foreground color.
func TagForeground(name string) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		if col, err := cache.Get(name); err == nil {
			tag.Foreground = col
		}
	}
}

// TagBackground sets the tag's background color.
func TagBackground(name string) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		if col, err := cache.Get(name); err == nil {
			tag.Background = col
		}
	}
}

// TagFont sets the tag's font.
func TagFont(name string) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		if f, err := reg.Get(name); err == nil {
			tag.Font = f
		}
	}
}

// TagUnderline sets the tag's underline attribute.
func TagUnderline(on bool) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.Underline = on
	}
}

// TagOverstrike sets the tag's overstrike attribute.
func TagOverstrike(on bool) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.Overstrike = on
	}
}
