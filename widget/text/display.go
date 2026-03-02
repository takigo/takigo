package text

import (
	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/internal/xlib"
)

// displayLine represents one visual line (possibly a fragment of a logical line).
type displayLine struct {
	logicalLine int // 1-based
	startChar   int // rune offset where this display line starts
	endChar     int // exclusive
	y           int // pixel Y in content area (relative to visible area)
	height      int // pixel height
}

// textSegment is a run of text with uniform style.
type textSegment struct {
	text      string
	x         int
	width     int
	fg        *color.Color
	bg        *color.Color
	font      font.Font
	underline bool
}

// wrapLine wraps a logical line into display lines based on the wrap mode.
// lineIdx is 1-based. availWidth is the pixel width available for text.
func (t *TextWidget) wrapLine(lineIdx int, availWidth int) []displayLine {
	if lineIdx < 1 || lineIdx > t.doc.LineCount() {
		return nil
	}

	line := t.doc.Lines[lineIdx-1]
	lineLen := len(line.Text)

	if lineLen == 0 || t.Font == nil {
		return []displayLine{{
			logicalLine: lineIdx,
			startChar:   0,
			endChar:     lineLen,
			height:      t.lineHeight(),
		}}
	}

	if t.wrapMode == WrapNone || availWidth <= 0 {
		return []displayLine{{
			logicalLine: lineIdx,
			startChar:   0,
			endChar:     lineLen,
			height:      t.lineHeight(),
		}}
	}

	var result []displayLine
	start := 0
	h := t.lineHeight()

	for start < lineLen {
		end := lineLen
		seg := string(line.Text[start:end])
		w := t.Font.MeasureString(seg)

		if w <= availWidth {
			result = append(result, displayLine{
				logicalLine: lineIdx,
				startChar:   start,
				endChar:     end,
				height:      h,
			})
			break
		}

		// Binary search for break point.
		lo, hi := start+1, end
		for lo < hi {
			mid := (lo + hi) / 2
			mw := t.Font.MeasureString(string(line.Text[start:mid]))
			if mw <= availWidth {
				lo = mid + 1
			} else {
				hi = mid
			}
		}
		breakAt := lo - 1
		if breakAt <= start {
			breakAt = start + 1 // at least one char per display line
		}

		if t.wrapMode == WrapWord {
			// Try to find a word boundary.
			wb := breakAt
			for wb > start && line.Text[wb-1] != ' ' && line.Text[wb-1] != '\t' {
				wb--
			}
			if wb > start {
				breakAt = wb
			}
		}

		result = append(result, displayLine{
			logicalLine: lineIdx,
			startChar:   start,
			endChar:     breakAt,
			height:      h,
		})
		start = breakAt
	}

	if len(result) == 0 {
		result = append(result, displayLine{
			logicalLine: lineIdx,
			startChar:   0,
			endChar:     0,
			height:      h,
		})
	}
	return result
}

// computeVisibleLines returns the display lines visible from the current scroll position.
func (t *TextWidget) computeVisibleLines() []displayLine {
	availWidth := t.Win.Width - 2*t.inset
	availHeight := t.Win.Height - 2*t.inset
	if availWidth <= 0 || availHeight <= 0 {
		return nil
	}

	var result []displayLine
	y := 0

	for lineIdx := t.topLine; lineIdx <= t.doc.LineCount() && y < availHeight; lineIdx++ {
		dls := t.wrapLine(lineIdx, availWidth)
		startDL := 0
		if lineIdx == t.topLine {
			startDL = t.topCharOffset
		}
		for i := startDL; i < len(dls) && y < availHeight; i++ {
			dl := dls[i]
			dl.y = y
			result = append(result, dl)
			y += dl.height
		}
	}
	return result
}

// segmentsForRange computes styled text segments for a portion of a logical line.
func (t *TextWidget) segmentsForRange(lineIdx, startChar, endChar int) []textSegment {
	if startChar >= endChar || t.Font == nil {
		return nil
	}
	line := t.doc.Lines[lineIdx-1]
	text := line.Text[startChar:endChar]

	// Collect tag boundaries within this range.
	type boundary struct {
		pos    int // relative to startChar
		tagOn  string
		tagOff string
	}

	// Build a simple approach: for each char position, resolve the active tags.
	// For efficiency, split at tag range boundaries.
	type tagEvent struct {
		pos   int
		name  string
		start bool
	}
	var events []tagEvent
	for _, tr := range t.doc.TagRanges {
		tag, ok := t.doc.Tags[tr.TagName]
		if !ok || tag == nil {
			continue
		}
		// Convert tag range to char offsets within this display line's line.
		if tr.Start.Line > lineIdx || tr.End.Line < lineIdx {
			continue
		}
		tStart := 0
		if tr.Start.Line == lineIdx && tr.Start.Char > startChar {
			tStart = tr.Start.Char - startChar
		}
		tEnd := endChar - startChar
		if tr.End.Line == lineIdx && tr.End.Char-startChar < tEnd {
			tEnd = tr.End.Char - startChar
		}
		if tStart < 0 {
			tStart = 0
		}
		if tEnd > endChar-startChar {
			tEnd = endChar - startChar
		}
		if tStart >= tEnd {
			continue
		}
		events = append(events, tagEvent{pos: tStart, name: tr.TagName, start: true})
		events = append(events, tagEvent{pos: tEnd, name: tr.TagName, start: false})
	}

	// Sort events by position.
	for i := 1; i < len(events); i++ {
		for j := i; j > 0 && events[j].pos < events[j-1].pos; j-- {
			events[j], events[j-1] = events[j-1], events[j]
		}
	}

	// Collect unique breakpoints.
	breaks := []int{0}
	for _, ev := range events {
		if ev.pos > 0 && ev.pos < endChar-startChar && (len(breaks) == 0 || breaks[len(breaks)-1] != ev.pos) {
			breaks = append(breaks, ev.pos)
		}
	}
	breaks = append(breaks, endChar-startChar)

	// Build segments between breakpoints.
	var segments []textSegment
	x := 0
	for i := 0; i < len(breaks)-1; i++ {
		segStart := breaks[i]
		segEnd := breaks[i+1]
		segText := string(text[segStart:segEnd])
		segWidth := t.Font.MeasureString(segText)

		// Resolve attributes at this position.
		fg := t.Foreground
		var bg *color.Color
		f := t.Font
		underline := false

		tags := t.doc.TagsAt(Index{Line: lineIdx, Char: startChar + segStart})
		for _, tag := range tags {
			if tag.Foreground != nil {
				fg = tag.Foreground
			}
			if tag.Background != nil {
				bg = tag.Background
			}
			if tag.Font != nil {
				f = tag.Font
				segWidth = f.MeasureString(segText)
			}
			if tag.Underline {
				underline = true
			}
		}

		segments = append(segments, textSegment{
			text:      segText,
			x:         x,
			width:     segWidth,
			fg:        fg,
			bg:        bg,
			font:      f,
			underline: underline,
		})
		x += segWidth
	}
	return segments
}

// renderToPixmap draws the text widget content to the offscreen pixmap.
func (t *TextWidget) renderToPixmap() {
	w := t.Win
	d := w.Display.XDisplay
	gc := w.GC
	pxDrawable := xlib.PixmapDrawable(t.pixmap)

	winW := w.Width
	winH := w.Height

	// Clear to background.
	bgPixel := uint64(0xFFFFFF)
	if t.Background != nil {
		bgPixel = t.Background.Pixel
	}
	d.SetForeground(gc, bgPixel)
	d.FillRectangle(pxDrawable, gc, 0, 0, uint(winW), uint(winH))

	if t.Font == nil {
		return
	}

	xftFont, isXft := t.Font.(*font.XftFont)
	if !isXft {
		return
	}

	m := t.Font.Metrics()
	dlines := t.computeVisibleLines()

	for _, dl := range dlines {
		baseY := t.inset + dl.y + m.Ascent
		segments := t.segmentsForRange(dl.logicalLine, dl.startChar, dl.endChar)

		xOffset := t.inset - t.xOffset
		for _, seg := range segments {
			segX := xOffset + seg.x

			// Draw background if set.
			if seg.bg != nil {
				d.SetForeground(gc, seg.bg.Pixel)
				d.FillRectangle(pxDrawable, gc, segX, t.inset+dl.y, uint(seg.width), uint(dl.height))
			}

			// Draw selection highlight.
			t.drawSelectionHighlight(d, gc, pxDrawable, dl, seg, segX, xOffset)

			// Draw text.
			drawFont := xftFont
			if seg.font != nil {
				if sf, ok := seg.font.(*font.XftFont); ok {
					drawFont = sf
				}
			}
			if seg.fg != nil {
				drawFont.DrawString(pxDrawable, segX, baseY, seg.text,
					seg.fg.Pixel, seg.fg.Red, seg.fg.Green, seg.fg.Blue)
			}

			// Draw underline.
			if seg.underline && seg.fg != nil {
				underY := baseY + 2
				d.SetForeground(gc, seg.fg.Pixel)
				d.DrawLine(pxDrawable, gc, segX, underY, segX+seg.width, underY)
			}
		}
	}

	// Draw cursor.
	if t.hasFocus && t.cursorOn {
		t.drawCursor(d, gc, pxDrawable, dlines)
	}
}

// drawSelectionHighlight draws selection highlight for a segment if applicable.
func (t *TextWidget) drawSelectionHighlight(d *xlib.Display, gc xlib.GC, drawable xlib.Drawable,
	dl displayLine, seg textSegment, segX, xOffset int) {
	selRanges := t.doc.TagRangesFor("sel")
	if len(selRanges) == 0 || t.selBg == nil {
		return
	}
	sr := selRanges[0]

	// Check if this segment's line is within selection.
	segStartIdx := Index{Line: dl.logicalLine, Char: dl.startChar}
	segEndIdx := Index{Line: dl.logicalLine, Char: dl.endChar}
	if Compare(segEndIdx, sr.Start) <= 0 || Compare(segStartIdx, sr.End) >= 0 {
		return
	}

	// Compute highlight X range within this segment.
	lineText := t.doc.Lines[dl.logicalLine-1].Text

	hlStart := dl.startChar
	if sr.Start.Line == dl.logicalLine && sr.Start.Char > hlStart {
		hlStart = sr.Start.Char
	}
	hlEnd := dl.endChar
	if sr.End.Line == dl.logicalLine && sr.End.Char < hlEnd {
		hlEnd = sr.End.Char
	}
	if hlStart >= hlEnd {
		return
	}

	hlStartX := xOffset + t.Font.MeasureString(string(lineText[dl.startChar:hlStart]))
	hlEndX := xOffset + t.Font.MeasureString(string(lineText[dl.startChar:hlEnd]))

	d.SetForeground(gc, t.selBg.Pixel)
	d.FillRectangle(drawable, gc, hlStartX, t.inset+dl.y, uint(hlEndX-hlStartX), uint(dl.height))
}

// drawCursor draws the text insertion cursor.
func (t *TextWidget) drawCursor(d *xlib.Display, gc xlib.GC, drawable xlib.Drawable, dlines []displayLine) {
	if t.insertColor == nil {
		return
	}
	insertPos := t.doc.Marks["insert"].Pos

	for _, dl := range dlines {
		if dl.logicalLine != insertPos.Line {
			continue
		}
		if insertPos.Char < dl.startChar || insertPos.Char > dl.endChar {
			continue
		}
		// Found the display line containing the cursor.
		lineText := t.doc.Lines[dl.logicalLine-1].Text
		cursorChars := insertPos.Char - dl.startChar
		var cursorX int
		if cursorChars > 0 && dl.startChar+cursorChars <= len(lineText) {
			cursorX = t.Font.MeasureString(string(lineText[dl.startChar : dl.startChar+cursorChars]))
		}
		cursorX += t.inset - t.xOffset

		d.SetForeground(gc, t.insertColor.Pixel)
		d.FillRectangle(drawable, gc,
			cursorX-t.insertWidth/2, t.inset+dl.y,
			uint(t.insertWidth), uint(dl.height))
		return
	}
}

// indexFromPixel returns the document index at the given pixel coordinates.
func (t *TextWidget) indexFromPixel(x, y int) Index {
	dlines := t.computeVisibleLines()
	if len(dlines) == 0 {
		return Index{1, 0}
	}

	// Find the display line at y.
	var dl *displayLine
	for i := range dlines {
		if y >= t.inset+dlines[i].y && y < t.inset+dlines[i].y+dlines[i].height {
			dl = &dlines[i]
			break
		}
	}
	if dl == nil {
		// Past last line — use last display line.
		dl = &dlines[len(dlines)-1]
	}

	// Find character position at x.
	lineText := t.doc.Lines[dl.logicalLine-1].Text
	segText := lineText[dl.startChar:dl.endChar]
	xInContent := x - t.inset + t.xOffset

	if len(segText) == 0 || t.Font == nil {
		return Index{Line: dl.logicalLine, Char: dl.startChar}
	}

	// Binary search for the closest character gap.
	lo, hi := 0, len(segText)
	for lo < hi {
		mid := (lo + hi) / 2
		w := t.Font.MeasureString(string(segText[:mid+1]))
		if w <= xInContent {
			lo = mid + 1
		} else {
			hi = mid
		}
	}

	// Check if x is past the midpoint of the character for snapping.
	charIdx := lo
	if charIdx < len(segText) {
		var charStart int
		if charIdx > 0 {
			charStart = t.Font.MeasureString(string(segText[:charIdx]))
		}
		charEnd := t.Font.MeasureString(string(segText[:charIdx+1]))
		mid := (charStart + charEnd) / 2
		if xInContent >= mid {
			charIdx++
		}
	}

	return Index{Line: dl.logicalLine, Char: dl.startChar + charIdx}
}

// computeTotalHeight returns the total pixel height of all lines.
func (t *TextWidget) computeTotalHeight() int {
	if t.Font == nil {
		return 0
	}
	availWidth := t.Win.Width - 2*t.inset
	total := 0
	for lineIdx := 1; lineIdx <= t.doc.LineCount(); lineIdx++ {
		dls := t.wrapLine(lineIdx, availWidth)
		for _, dl := range dls {
			total += dl.height
		}
	}
	return total
}

// lineHeight returns the pixel height of one line.
func (t *TextWidget) lineHeight() int {
	if t.Font == nil {
		return 16
	}
	return t.Font.Metrics().Linespace()
}

// computeDisplayLineCount returns the total number of display lines from
// the top of the document to (but not including) the given logical line
// and display-line offset within it.
func (t *TextWidget) computeDisplayLinesBefore(lineIdx, dlOffset int) int {
	availWidth := t.Win.Width - 2*t.inset
	count := 0
	for l := 1; l < lineIdx; l++ {
		count += len(t.wrapLine(l, availWidth))
	}
	count += dlOffset
	return count
}

// totalDisplayLines returns the total number of display lines in the document.
func (t *TextWidget) totalDisplayLines() int {
	availWidth := t.Win.Width - 2*t.inset
	count := 0
	for l := 1; l <= t.doc.LineCount(); l++ {
		count += len(t.wrapLine(l, availWidth))
	}
	return count
}
