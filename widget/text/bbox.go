package text

// The bbox and dlineinfo widget commands: TkTextIndexBbox and
// TkTextDLineInfo in tkTextDisp.c, measured on the display lines that
// renderTo draws, so a top line partly above the view (topPixelOffset)
// gives a y above the text area.

import "github.com/takigo/takigo/option"

// visibleDLine returns the display line on screen that holds idx; the
// index after a logical line's last character (its newline) is on the
// line's last display line.
func (t *TextWidget) visibleDLine(idx Index) (displayLine, bool) {
	for _, dl := range t.computeVisibleLines() {
		if dl.logicalLine != idx.Line || idx.Char < dl.startChar || dl.height == 0 {
			continue
		}
		if idx.Char < dl.endChar || (idx.Char == dl.endChar && dl.endChar == len(t.doc.Lines[idx.Line-1].Text)) {
			return dl, true
		}
	}
	return displayLine{}, false
}

// lineX is the window x of a display line's first segment: the left inset
// and margin, the -justify offset and the horizontal scroll.
func (t *TextWidget) lineX(dl displayLine, segs []textSegment) int {
	j := 0
	if dl.justify != option.JustifyLeft {
		totalW := 0
		for _, seg := range segs {
			totalW += seg.width
		}
		availW := t.Win.Width - 2*t.insetX - dl.leftMargin - dl.rightMargin
		switch dl.justify {
		case option.JustifyCenter:
			j = max((availW-totalW)/2, 0)
		case option.JustifyRight:
			j = max(availW-totalW, 0)
		default:
		}
	}
	return t.insetX + dl.leftMargin + j - t.xOffset
}

// BBox returns the box, in window coordinates, of the character at index
// (bbox): a character's box spans its font's ascent and descent, an
// image's is the image as drawn, and the newline at the end of a line
// reaches the right edge of the text area. ok is false when the character
// is not on screen. Like Tk the box is cut at the right and bottom of the
// text area but not at its top.
func (t *TextWidget) BBox[I IndexSpec](index I) (x, y, width, height int, ok bool) {
	idx, valid := indexOf(t, index)
	if !valid || t.Font == nil {
		return 0, 0, 0, 0, false
	}
	dl, found := t.visibleDLine(Clamp(idx, t.doc))
	if !found {
		return 0, 0, 0, 0, false
	}
	idx = Clamp(idx, t.doc)
	segs := t.segmentsForRange(dl.logicalLine, dl.startChar, dl.endChar)
	x = t.lineX(dl, segs)
	top := t.insetY + dl.y
	fm := t.Font.Metrics()
	y, height = top+dl.ascent-fm.Ascent, fm.Ascent+fm.Descent

	n, end := idx.Char-dl.startChar, 0
	placed := false
	for _, seg := range segs {
		end += seg.runes
		if n >= end {
			continue
		}
		k := n - (end - seg.runes)
		runes := []rune(seg.text)
		switch {
		case seg.img != nil:
			ei := seg.img
			x += seg.x + ei.padX
			width, height = ei.img.Width(), ei.img.Height()
			switch ei.align {
			case AlignTop:
				y = top + ei.padY
			case AlignBottom:
				y = top + dl.height - height - ei.padY
			case AlignBaseline:
				y = top + dl.ascent - height
			default:
				y = top + (dl.height-height)/2
			}
		case len(runes) == seg.runes && !seg.elided:
			f := seg.font
			if f == nil {
				f = t.Font
			}
			x += seg.x + f.MeasureString(string(runes[:k]))
			width = f.MeasureString(string(runes[k]))
			sm := f.Metrics()
			y, height = top+dl.ascent-seg.offset-sm.Ascent, sm.Ascent+sm.Descent
		default: // a tab, an embedded window or elided text
			x += seg.x
			width = seg.width
			if seg.elided {
				width = 0
			}
			y, height = top, dl.height
		}
		placed = true
		break
	}
	maxX := t.Win.Width - t.insetX
	if !placed {
		for _, seg := range segs {
			x += seg.width
		}
		x = min(x, maxX)
		width = maxX - x
	}
	if width == 0 {
		if x < t.insetX {
			return 0, 0, 0, 0, false
		}
	} else if x+width <= t.insetX {
		return 0, 0, 0, 0, false
	}
	if x+width > maxX {
		if width = maxX - x; width <= 0 {
			return 0, 0, 0, 0, false
		}
	}
	if maxY := t.Win.Height - t.insetY; y+height > maxY {
		if height = maxY - y; height <= 0 {
			return 0, 0, 0, 0, false
		}
	}
	return x, y, width, height, true
}

// DLineInfo returns the box of the display line holding index, its
// -spacing included, and the baseline's offset from the box's top
// (dlineinfo). ok is false when the line is not on screen. The height is
// cut at the bottom of the text area; a line partly above the view has a y
// above it.
func (t *TextWidget) DLineInfo[I IndexSpec](index I) (x, y, width, height, baseline int, ok bool) {
	idx, valid := indexOf(t, index)
	if !valid {
		return 0, 0, 0, 0, 0, false
	}
	dl, found := t.visibleDLine(Clamp(idx, t.doc))
	if !found {
		return 0, 0, 0, 0, 0, false
	}
	segs := t.segmentsForRange(dl.logicalLine, dl.startChar, dl.endChar)
	x = t.lineX(dl, segs)
	for _, seg := range segs {
		width += seg.width
	}
	y = t.insetY + dl.y - dl.spacingAbove
	height = dl.spacingAbove + dl.height + dl.spacingBelow
	if maxY := t.Win.Height - t.insetY; y+height > maxY {
		height = maxY - y
	}
	return x, y, width, height, dl.spacingAbove + dl.ascent, true
}
