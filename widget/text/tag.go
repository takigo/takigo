package text

import (
	"slices"
	"sort"

	"github.com/takigo/takigo/color"
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/screenunit"
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

	ranges     []TagRange
	registered bool
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
func TagForeground[C color.Spec](name C) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		if color.IsEmpty(name) {
			tag.Foreground = nil
			return
		}
		if col, err := cache.Resolve(name); err == nil {
			tag.Foreground = col
		}
	}
}

// TagBackground sets the tag's background color. Pass "" to clear.
func TagBackground[C color.Spec](name C) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		if color.IsEmpty(name) {
			tag.Background = nil
			return
		}
		if col, err := cache.Resolve(name); err == nil {
			tag.Background = col
		}
	}
}

// TagFont sets the tag's font.
func TagFont[F font.Spec](name F) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		if f, err := reg.Resolve(name); err == nil {
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
func TagOffset[L screenunit.Length](pixels L) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.Offset = screenunit.ToPixels(pixels)
		tag.OffsetSet = true
	}
}

// TagLMargin1 sets the left margin (pixels) for the first display line of a logical line.
func TagLMargin1[L screenunit.Length](pixels L) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.LMargin1 = screenunit.ToPixels(pixels)
	}
}

// TagLMargin2 sets the left margin (pixels) for wrapped continuation display lines.
func TagLMargin2[L screenunit.Length](pixels L) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.LMargin2 = screenunit.ToPixels(pixels)
	}
}

// TagRMargin sets the right margin (pixels).
func TagRMargin[L screenunit.Length](pixels L) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.RMargin = screenunit.ToPixels(pixels)
	}
}

// TagSpacing1 sets extra space (pixels) above the first display line of a logical line.
func TagSpacing1[L screenunit.Length](pixels L) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.Spacing1 = screenunit.ToPixels(pixels)
	}
}

// TagSpacing2 sets extra space (pixels) between wrapped display lines of the same logical line.
func TagSpacing2[L screenunit.Length](pixels L) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.Spacing2 = screenunit.ToPixels(pixels)
	}
}

// TagSpacing3 sets extra space (pixels) below the last display line of a logical line.
func TagSpacing3[L screenunit.Length](pixels L) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.Spacing3 = screenunit.ToPixels(pixels)
	}
}

// TagRelief sets a 3D border style drawn around text covered by this tag.
func TagRelief(r option.Relief) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.Relief = r
		tag.ReliefSet = true
	}
}

// TagBorderWidth sets the border thickness used with TagRelief.
func TagBorderWidth[L screenunit.Length](n L) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.BorderWidth = screenunit.ToPixels(n)
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

// selPriority keeps the selection above every other tag; Tk raises "sel"
// to the top with a "tag raise" at widget creation.
const selPriority = 1 << 30

// ranges are kept sorted by Start, disjoint, with adjacent ranges merged.

// firstEnding returns the index of the first range whose End is after idx
// (or at idx when inclusive is true).
func (tg *Tag) firstEnding(idx Index, inclusive bool) int {
	return sort.Search(len(tg.ranges), func(i int) bool {
		c := Compare(tg.ranges[i].End, idx)
		return c > 0 || (inclusive && c == 0)
	})
}

// firstStarting returns the index of the first range whose Start is at or
// after idx (or strictly after when exclusive is true).
func (tg *Tag) firstStarting(idx Index, exclusive bool) int {
	return sort.Search(len(tg.ranges), func(i int) bool {
		c := Compare(tg.ranges[i].Start, idx)
		return c > 0 || (!exclusive && c == 0)
	})
}

// rangeAt returns the index of the range covering idx.
func (tg *Tag) rangeAt(idx Index) (int, bool) {
	i := tg.firstEnding(idx, false)
	if i < len(tg.ranges) && Compare(tg.ranges[i].Start, idx) <= 0 {
		return i, true
	}
	return i, false
}

// covers reports whether idx carries the tag.
func (tg *Tag) covers(idx Index) bool {
	_, ok := tg.rangeAt(idx)
	return ok
}

// rangesOnLine returns the ranges that touch the given logical line.
func (tg *Tag) rangesOnLine(line int) []TagRange {
	lo := tg.firstEnding(Index{Line: line}, false)
	hi := tg.firstStarting(Index{Line: line + 1}, false)
	if lo >= hi {
		return nil
	}
	return tg.ranges[lo:hi]
}

// add tags [start, end), merging with overlapping and adjacent ranges.
func (tg *Tag) add(start, end Index, toEnd bool) {
	if Compare(start, end) >= 0 {
		return
	}
	lo := tg.firstEnding(start, true)
	hi := tg.firstStarting(end, true)
	merged := TagRange{TagName: tg.Name, Start: start, End: end, ToEnd: toEnd}
	if lo < hi {
		if Compare(tg.ranges[lo].Start, start) < 0 {
			merged.Start = tg.ranges[lo].Start
		}
		if last := tg.ranges[hi-1]; Compare(last.End, end) > 0 {
			merged.End, merged.ToEnd = last.End, last.ToEnd
		} else if Compare(last.End, end) == 0 {
			merged.ToEnd = merged.ToEnd || last.ToEnd
		}
	}
	tg.ranges = slices.Replace(tg.ranges, lo, hi, merged)
}

// remove untags [start, end), trimming or splitting the ranges it overlaps.
func (tg *Tag) remove(start, end Index) {
	if Compare(start, end) >= 0 {
		return
	}
	lo := tg.firstEnding(start, false)
	hi := tg.firstStarting(end, false)
	if lo >= hi {
		return
	}
	var pieces []TagRange
	if first := tg.ranges[lo]; Compare(first.Start, start) < 0 {
		pieces = append(pieces, TagRange{TagName: tg.Name, Start: first.Start, End: start})
	}
	if last := tg.ranges[hi-1]; Compare(last.End, end) > 0 {
		pieces = append(pieces, TagRange{TagName: tg.Name, Start: end, End: last.End, ToEnd: last.ToEnd})
	}
	tg.ranges = slices.Replace(tg.ranges, lo, hi, pieces...)
}

// adjustInsert shifts ranges for text inserted at idx.
func (tg *Tag) adjustInsert(idx Index, newlines, lastLineLen int) {
	for i := tg.firstEnding(idx, true); i < len(tg.ranges); i++ {
		tr := &tg.ranges[i]
		if newlines == 0 && tr.Start.Line > idx.Line {
			return
		}
		tr.Start = adjustIdxInsert(tr.Start, idx, newlines, lastLineLen, false)
		tr.End = adjustIdxInsert(tr.End, idx, newlines, lastLineLen, !tr.ToEnd)
	}
}

// adjustDelete shifts ranges for the deletion of [start, end), dropping
// those that become empty and merging the pair that becomes adjacent.
func (tg *Tag) adjustDelete(start, end Index) {
	lo := tg.firstEnding(start, false)
	n := lo
	for i := lo; i < len(tg.ranges); i++ {
		tr := tg.ranges[i]
		if start.Line == end.Line && tr.Start.Line > end.Line {
			n += copy(tg.ranges[n:], tg.ranges[i:])
			break
		}
		tr.Start = adjustIdxDelete(tr.Start, start, end)
		tr.End = adjustIdxDelete(tr.End, start, end)
		if Compare(tr.Start, tr.End) < 0 {
			tg.ranges[n] = tr
			n++
		}
	}
	tg.ranges = tg.ranges[:n]
	for k := max(lo-1, 0); k <= lo && k+1 < len(tg.ranges); k++ {
		if Compare(tg.ranges[k].End, tg.ranges[k+1].Start) == 0 {
			tg.ranges[k].End, tg.ranges[k].ToEnd = tg.ranges[k+1].End, tg.ranges[k+1].ToEnd
			tg.ranges = slices.Delete(tg.ranges, k+1, k+2)
			return
		}
	}
}

// affectsLayout reports whether the tag changes line wrapping or heights.
func (tg *Tag) affectsLayout() bool {
	return tg.Font != nil || tg.OffsetSet || tg.JustifySet ||
		tg.LMargin1 != 0 || tg.LMargin2 != 0 || tg.RMargin != 0 ||
		tg.Spacing1 != 0 || tg.Spacing2 != 0 || tg.Spacing3 != 0
}
