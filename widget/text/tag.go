package text

import (
	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/screenunit"
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
	// Justify controls horizontal alignment of lines that include this tag.
	Justify option.Justify
	// JustifySet is true when Justify has been explicitly configured.
	JustifySet bool
	// Offset is a vertical pixel displacement: positive = superscript, negative = subscript.
	Offset int
	// OffsetSet is true when Offset has been explicitly configured.
	OffsetSet bool
	// LMargin1 is the left margin (px) for the first display line of a logical line.
	LMargin1 int
	// LMargin2 is the left margin (px) for wrapped continuation display lines.
	LMargin2 int
	// RMargin is the right margin (px).
	RMargin int
	// Spacing1 is extra space (px) above the first display line of a logical line.
	Spacing1 int
	// Spacing2 is extra space (px) between wrapped display lines of the same logical line.
	Spacing2 int
	// Spacing3 is extra space (px) below the last display line of a logical line.
	Spacing3 int
	// Relief specifies a 3D border style drawn around text with this tag.
	Relief option.Relief
	// ReliefSet is true when Relief has been explicitly configured.
	ReliefSet bool
	// BorderWidth is the border thickness in pixels used with Relief.
	BorderWidth int
	// BgStipple is the name of a stipple pattern for the background (e.g. "gray12", "gray50").
	BgStipple string
	// FgStipple is the name of a stipple pattern for the foreground (e.g. "gray50").
	FgStipple string
}

// TagRange associates a tag name with a half-open index range [Start, End).
type TagRange struct {
	TagName string
	Start   Index
	End     Index
	// ToEnd marks a range added up to "end", which in Tk covers the final
	// newline: text inserted at the end has tagged characters on both sides
	// and so inherits the tag.
	ToEnd bool
}

// TagOption configures a Tag.
type TagOption func(cache *color.Cache, reg *font.Registry, tag *Tag)

// TagForeground sets the tag's foreground color. Pass "" to clear.
func TagForeground(name string) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		if name == "" {
			tag.Foreground = nil
			return
		}
		if col, err := cache.Get(name); err == nil {
			tag.Foreground = col
		}
	}
}

// TagBackground sets the tag's background color. Pass "" to clear.
func TagBackground(name string) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		if name == "" {
			tag.Background = nil
			return
		}
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

// TagOverstrike sets the tag's overstrike (strikethrough) attribute.
func TagOverstrike(on bool) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.Overstrike = on
	}
}

// TagJustify sets the horizontal justification for lines covered by this tag.
func TagJustify(j option.Justify) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.Justify = j
		tag.JustifySet = true
	}
}

// TagOffset sets the vertical pixel displacement for text in this tag.
// Positive values move text up (superscript); negative values move down (subscript).
func TagOffset(pixels int) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.Offset = pixels
		tag.OffsetSet = true
	}
}

// TagOffsetStr sets TagOffset from a Tk-style distance string (e.g. "4p").
func TagOffsetStr(dist string) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.Offset = screenunit.Px(dist)
		tag.OffsetSet = true
	}
}

// TagLMargin1 sets the left margin (pixels) for the first display line of a logical line.
func TagLMargin1(pixels int) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.LMargin1 = pixels
	}
}

// TagLMargin1Str sets TagLMargin1 from a Tk-style distance string.
func TagLMargin1Str(dist string) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.LMargin1 = screenunit.Px(dist)
	}
}

// TagLMargin2 sets the left margin (pixels) for wrapped continuation display lines.
func TagLMargin2(pixels int) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.LMargin2 = pixels
	}
}

// TagLMargin2Str sets TagLMargin2 from a Tk-style distance string.
func TagLMargin2Str(dist string) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.LMargin2 = screenunit.Px(dist)
	}
}

// TagRMargin sets the right margin (pixels).
func TagRMargin(pixels int) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.RMargin = pixels
	}
}

// TagRMarginStr sets TagRMargin from a Tk-style distance string.
func TagRMarginStr(dist string) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.RMargin = screenunit.Px(dist)
	}
}

// TagSpacing1 sets extra space (pixels) above the first display line of a logical line.
func TagSpacing1(pixels int) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.Spacing1 = pixels
	}
}

// TagSpacing1Str sets TagSpacing1 from a Tk-style distance string.
func TagSpacing1Str(dist string) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.Spacing1 = screenunit.Px(dist)
	}
}

// TagSpacing2 sets extra space (pixels) between wrapped display lines of the same logical line.
func TagSpacing2(pixels int) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.Spacing2 = pixels
	}
}

// TagSpacing2Str sets TagSpacing2 from a Tk-style distance string.
func TagSpacing2Str(dist string) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.Spacing2 = screenunit.Px(dist)
	}
}

// TagSpacing3 sets extra space (pixels) below the last display line of a logical line.
func TagSpacing3(pixels int) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.Spacing3 = pixels
	}
}

// TagSpacing3Str sets TagSpacing3 from a Tk-style distance string.
func TagSpacing3Str(s string) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.Spacing3 = screenunit.Px(s)
	}
}

// TagRelief sets a 3D border style drawn around text covered by this tag.
func TagRelief(r option.Relief) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.Relief = r
		tag.ReliefSet = true
	}
}

// TagBorderWidth sets the border thickness (pixels) used with TagRelief.
func TagBorderWidth(n int) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.BorderWidth = n
	}
}

// TagBgStipple sets the background stipple pattern name (e.g. "gray12", "gray50").
func TagBgStipple(name string) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.BgStipple = name
	}
}

// TagFgStipple sets the foreground stipple pattern name (e.g. "gray50").
func TagFgStipple(name string) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.FgStipple = name
	}
}
