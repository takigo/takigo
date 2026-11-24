package text

import (
	"slices"
	"sort"

	"github.com/takigo/takigo/color"
	"github.com/takigo/takigo/cursor"
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
	// Elide hides the text (-elide); ElideSet tells whether it was configured.
	Elide    bool
	ElideSet bool
	// Wrap is the wrap mode of the lines that start with this tag (-wrap).
	Wrap    WrapMode
	WrapSet bool
	// Tabs are the tab stops of the lines that start with this tag (-tabs).
	Tabs []TabStop
	// UnderlineFg and OverstrikeFg colour the lines -underline and
	// -overstrike draw (-underlinefg, -overstrikefg); nil uses the foreground.
	UnderlineFg  *color.Color
	OverstrikeFg *color.Color
	// LMarginColor and RMarginColor fill the left and right margins of the
	// lines that start with this tag (-lmargincolor, -rmargincolor).
	LMarginColor *color.Color
	RMarginColor *color.Color
	// SelectBackground and SelectForeground colour the tag's text while it
	// is selected (-selectbackground, -selectforeground).
	SelectBackground *color.Color
	SelectForeground *color.Color
	// TabStyle is how the tabs of the lines that start with this tag find
	// their stop (-tabstyle), when TabStyleSet.
	TabStyle    TabStyle
	TabStyleSet bool

	// The options below are takigo extensions that Tk's text tags lack.

	// Cursor is the mouse cursor over text with this tag (CursorSet).
	Cursor    cursor.Shape
	CursorSet bool
	// BlockBackground fills a box behind consecutive lines with this tag,
	// as a CSS block's background: from the tag's LMargin1 to its RMargin,
	// around the lines' spacing inside the box.
	BlockBackground *color.Color
	// BlockBorder is the box's border width on each side (left, top, right,
	// bottom), drawn in BlockBorderColor (the foreground when nil).
	BlockBorder      [4]int
	BlockBorderColor *color.Color
	// BlockPadding is the space between the box's border and the text on
	// each side (left, top, right, bottom).
	BlockPadding [4]int
	// BlockRadius rounds the box's corners; a Classic App draws them square.
	BlockRadius int
	// Marker is drawn in the margin before the first line of each range of
	// the tag that starts a line, right-aligned against LMargin1: a list
	// bullet or number that is not part of the text.
	Marker string
	// Rule draws a horizontal line of this thickness across the lines with
	// this tag, between the margins, in RuleColor (the foreground when nil).
	Rule      int
	RuleColor *color.Color
	// Clear starts the lines with this tag below floating images.
	Clear Clear

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

// resolveColor sets *dst from name, or clears it for "".
func resolveColor[C color.Spec](cache *color.Cache, name C, dst **color.Color) {
	if color.IsEmpty(name) {
		*dst = nil
		return
	}
	if col, err := cache.Resolve(name); err == nil {
		*dst = col
	}
}

// TagElide hides (true) or shows the tag's text (-elide). A logical line
// whose characters and newline are all elided takes no space; eliding only
// a newline does not join its line with the next, unlike Tk.
func TagElide(on bool) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.Elide, tag.ElideSet = on, true
	}
}

// TagWrap sets the wrap mode of the logical lines that start with this tag
// (-wrap).
func TagWrap(mode WrapMode) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.Wrap, tag.WrapSet = mode, true
	}
}

// TagTabs sets the tab stops of the logical lines that start with this tag
// (-tabs). No stops restores the widget's.
func TagTabs(stops ...TabStop) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.Tabs = slices.Clone(stops)
	}
}

// TagUnderlineFg sets the colour of the underline (-underlinefg).
func TagUnderlineFg[C color.Spec](name C) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		resolveColor(cache, name, &tag.UnderlineFg)
	}
}

// TagOverstrikeFg sets the colour of the overstrike line (-overstrikefg).
func TagOverstrikeFg[C color.Spec](name C) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		resolveColor(cache, name, &tag.OverstrikeFg)
	}
}

// TagLMarginColor fills the left margin of the tag's lines (-lmargincolor).
func TagLMarginColor[C color.Spec](name C) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		resolveColor(cache, name, &tag.LMarginColor)
	}
}

// TagRMarginColor fills the right margin of the tag's lines (-rmargincolor).
func TagRMarginColor[C color.Spec](name C) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		resolveColor(cache, name, &tag.RMarginColor)
	}
}

// TagSelectBackground sets the background of the tag's text while it is
// selected (-selectbackground). Pass "" to clear.
func TagSelectBackground[C color.Spec](name C) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		resolveColor(cache, name, &tag.SelectBackground)
	}
}

// TagSelectForeground sets the foreground of the tag's text while it is
// selected (-selectforeground). Pass "" to clear.
func TagSelectForeground[C color.Spec](name C) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		resolveColor(cache, name, &tag.SelectForeground)
	}
}

// TagTabStyle sets how the tabs of the lines that start with this tag find
// their stop (-tabstyle).
func TagTabStyle(style TabStyle) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.TabStyle, tag.TabStyleSet = style, true
	}
}

// TagCursor sets the mouse cursor shown over the tag's text, such as
// cursor.Hand2 for a link. Tk's tags have no -cursor.
func TagCursor(shape cursor.Shape) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.Cursor, tag.CursorSet = shape, true
	}
}

// TagBlockBackground fills a box behind the consecutive lines that carry
// the tag, like a CSS block background. Pass "" to clear. Not in Tk.
func TagBlockBackground[C color.Spec](name C) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		resolveColor(cache, name, &tag.BlockBackground)
	}
}

// TagBlockBorder draws a border of the given width and colour on every side
// of the tag's block box. Not in Tk.
func TagBlockBorder[L screenunit.Length, C color.Spec](width L, c C) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		w := max(screenunit.ToPixels(width), 0)
		tag.BlockBorder = [4]int{w, w, w, w}
		resolveColor(cache, c, &tag.BlockBorderColor)
	}
}

// TagBlockBorderSides sets the block border's width on each side, such as
// a left bar for a quotation. Not in Tk.
func TagBlockBorderSides[L screenunit.Length](left, top, right, bottom L) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.BlockBorder = [4]int{
			max(screenunit.ToPixels(left), 0), max(screenunit.ToPixels(top), 0),
			max(screenunit.ToPixels(right), 0), max(screenunit.ToPixels(bottom), 0),
		}
	}
}

// TagBlockBorderColor sets the block border's colour. Not in Tk.
func TagBlockBorderColor[C color.Spec](c C) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		resolveColor(cache, c, &tag.BlockBorderColor)
	}
}

// TagBlockPadding sets the space between the block border and the text on
// every side. Not in Tk.
func TagBlockPadding[L screenunit.Length](pad L) TagOption {
	return TagBlockPaddingSides(pad, pad, pad, pad)
}

// TagBlockPaddingSides sets the block padding on each side. Not in Tk.
func TagBlockPaddingSides[L screenunit.Length](left, top, right, bottom L) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.BlockPadding = [4]int{
			max(screenunit.ToPixels(left), 0), max(screenunit.ToPixels(top), 0),
			max(screenunit.ToPixels(right), 0), max(screenunit.ToPixels(bottom), 0),
		}
	}
}

// TagBlockRadius rounds the corners of the block box (square in a Classic
// App). Not in Tk.
func TagBlockRadius[L screenunit.Length](r L) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.BlockRadius = max(screenunit.ToPixels(r), 0)
	}
}

// TagMarker draws marker (a bullet such as "•" or a number such as "3.")
// in the left margin of the first line of every range of the tag that
// starts a line. Give each list item its own tag: ranges of one tag that
// touch merge into one item. Not in Tk.
func TagMarker(marker string) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.Marker = marker
	}
}

// TagRule draws a horizontal line of the given thickness and colour across
// the tag's lines, between the margins, like HTML's <hr>. A thickness of 0
// removes it; an empty colour uses the foreground. Not in Tk.
func TagRule[L screenunit.Length, C color.Spec](thickness L, c C) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.Rule = max(screenunit.ToPixels(thickness), 0)
		resolveColor(cache, c, &tag.RuleColor)
	}
}

// hasBlock reports whether the tag draws a block box or moves text for one.
func (tg *Tag) hasBlock() bool {
	return tg.BlockBackground != nil || tg.BlockBorder != [4]int{} || tg.BlockPadding != [4]int{}
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
		tg.Spacing1 != 0 || tg.Spacing2 != 0 || tg.Spacing3 != 0 ||
		tg.ElideSet || tg.WrapSet || len(tg.Tabs) > 0 || tg.TabStyleSet ||
		tg.LMarginColor != nil || tg.RMarginColor != nil ||
		tg.hasBlock() || tg.Marker != "" || tg.Rule > 0 || tg.Clear != ClearNone
}
