package text

import (
	"strings"

	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
)

// tabWidth is the number of spaces per tab stop.
const tabWidth = 4

// expandTabs replaces tab characters with spaces to align to tabWidth stops.
func expandTabs(s string) string {
	if !strings.ContainsRune(s, '\t') {
		return s
	}
	var buf strings.Builder
	col := 0
	for _, r := range s {
		if r == '\t' {
			spaces := tabWidth - (col % tabWidth)
			for i := 0; i < spaces; i++ {
				buf.WriteByte(' ')
			}
			col += spaces
		} else {
			buf.WriteRune(r)
			col++
		}
	}
	return buf.String()
}

// displayLine represents one visual line (possibly a fragment of a logical line).
type displayLine struct {
	logicalLine  int // 1-based
	startChar    int // rune offset where this display line starts
	endChar      int // exclusive
	y            int // pixel Y in content area (relative to visible area)
	height       int // pixel height (excluding spacing)
	spacingAbove int // extra pixels above (Spacing1 for first frag, Spacing2 for rest)
	spacingBelow int // extra pixels below (Spacing3 for last frag of logical line)
	leftMargin   int // left margin pixels (LMargin1 for first frag, LMargin2 for rest)
	rightMargin  int // right margin pixels
	justify      option.Justify
}

// textSegment is a run of text with uniform style.
type textSegment struct {
	text       string
	x          int
	width      int
	fg         *color.Color
	bg         *color.Color
	font       font.Font
	underline  bool
	overstrike bool
	offset     int // vertical pixel offset (positive = up/superscript, negative = down/subscript)
	relief     option.Relief
	reliefBW   int // border-width for 3D relief drawing
	reliefSet  bool
}

// lineProps holds resolved per-logical-line properties from tags.
type lineProps struct {
	lm1, lm2, rm       int
	sp1, sp2, sp3       int
	justify             option.Justify
}

// resolveLineProps resolves tag properties for the first character of a logical line.
func (t *TextWidget) resolveLineProps(lineIdx int) lineProps {
	tags := t.doc.TagsAt(Index{Line: lineIdx, Char: 0})
	var p lineProps
	for _, tag := range tags {
		if tag.LMargin1 > 0 {
			p.lm1 = tag.LMargin1
		}
		if tag.LMargin2 > 0 {
			p.lm2 = tag.LMargin2
		}
		if tag.RMargin > 0 {
			p.rm = tag.RMargin
		}
		if tag.Spacing1 > 0 {
			p.sp1 = tag.Spacing1
		}
		if tag.Spacing2 > 0 {
			p.sp2 = tag.Spacing2
		}
		if tag.Spacing3 > 0 {
			p.sp3 = tag.Spacing3
		}
		if tag.JustifySet {
			p.justify = tag.Justify
		}
	}
	return p
}

// wrapLine wraps a logical line into display lines based on the wrap mode.
// lineIdx is 1-based. availWidth is the pixel width available (before margins).
// lm1 is the left margin for the first fragment, lm2 for continuations, rm is the right margin.
func (t *TextWidget) wrapLine(lineIdx, availWidth, lm1, lm2, rm int) []displayLine {
	if lineIdx < 1 || lineIdx > t.doc.LineCount() {
		return nil
	}

	line := t.doc.Lines[lineIdx-1]
	lineLen := len(line.Text)
	h := t.lineHeight()

	if lineLen == 0 || t.Font == nil {
		return []displayLine{{
			logicalLine: lineIdx,
			startChar:   0,
			endChar:     lineLen,
			height:      h,
			leftMargin:  lm1,
			rightMargin: rm,
		}}
	}

	effectiveWidth1 := availWidth - lm1 - rm
	effectiveWidth2 := availWidth - lm2 - rm
	if effectiveWidth1 <= 0 {
		effectiveWidth1 = 1
	}
	if effectiveWidth2 <= 0 {
		effectiveWidth2 = 1
	}

	if t.wrapMode == WrapNone || availWidth <= 0 {
		return []displayLine{{
			logicalLine: lineIdx,
			startChar:   0,
			endChar:     lineLen,
			height:      h,
			leftMargin:  lm1,
			rightMargin: rm,
		}}
	}

	var result []displayLine
	start := 0

	for start < lineLen {
		isFirst := (start == 0)
		ew := effectiveWidth2
		lm := lm2
		if isFirst {
			ew = effectiveWidth1
			lm = lm1
		}

		end := lineLen
		seg := string(line.Text[start:end])
		w := t.Font.MeasureString(seg)

		if w <= ew {
			result = append(result, displayLine{
				logicalLine: lineIdx,
				startChar:   start,
				endChar:     end,
				height:      h,
				leftMargin:  lm,
				rightMargin: rm,
			})
			break
		}

		// Binary search for break point.
		lo, hi := start+1, end
		for lo < hi {
			mid := (lo + hi) / 2
			mw := t.Font.MeasureString(string(line.Text[start:mid]))
			if mw <= ew {
				lo = mid + 1
			} else {
				hi = mid
			}
		}
		breakAt := lo - 1
		if breakAt <= start {
			breakAt = start + 1
		}

		if t.wrapMode == WrapWord {
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
			leftMargin:  lm,
			rightMargin: rm,
		})
		start = breakAt
	}

	if len(result) == 0 {
		result = append(result, displayLine{
			logicalLine: lineIdx,
			startChar:   0,
			endChar:     0,
			height:      h,
			leftMargin:  lm1,
			rightMargin: rm,
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
		props := t.resolveLineProps(lineIdx)
		dls := t.wrapLine(lineIdx, availWidth, props.lm1, props.lm2, props.rm)
		startDL := 0
		if lineIdx == t.topLine {
			startDL = t.topCharOffset
		}
		for i := startDL; i < len(dls) && y < availHeight; i++ {
			dl := dls[i]
			// Apply spacing.
			if i == 0 {
				dl.spacingAbove = props.sp1
			} else {
				dl.spacingAbove = props.sp2
			}
			if i == len(dls)-1 {
				dl.spacingBelow = props.sp3
			}
			dl.justify = props.justify
			dl.y = y + dl.spacingAbove
			result = append(result, dl)
			y += dl.spacingAbove + dl.height + dl.spacingBelow
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
		overstrike := false
		offset := 0
		relief := option.ReliefFlat
		reliefBW := 1
		reliefSet := false

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
			if tag.Overstrike {
				overstrike = true
			}
			if tag.OffsetSet {
				offset = tag.Offset
			}
			if tag.ReliefSet {
				relief = tag.Relief
				reliefSet = true
				if tag.BorderWidth > 0 {
					reliefBW = tag.BorderWidth
				}
			}
		}

		segments = append(segments, textSegment{
			text:       segText,
			x:          x,
			width:      segWidth,
			fg:         fg,
			bg:         bg,
			font:       f,
			underline:  underline,
			overstrike: overstrike,
			offset:     offset,
			relief:     relief,
			reliefBW:   reliefBW,
			reliefSet:  reliefSet,
		})
		x += segWidth
	}
	return segments
}

// renderToPixmap draws the text widget content to the offscreen pixmap.
func (t *TextWidget) renderToPixmap() {
	w := t.Win
	d := w.Display.Server
	gc := w.GC
	pxDrawable := platform.PixmapDrawable(t.pixmap)

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

	drawableFont, isDrawable := t.Font.(platform.DrawableFont)
	if !isDrawable {
		return
	}

	m := t.Font.Metrics()
	dlines := t.computeVisibleLines()

	for _, dl := range dlines {
		baseY := t.inset + dl.y + m.Ascent
		segments := t.segmentsForRange(dl.logicalLine, dl.startChar, dl.endChar)

		// Compute total segment width for justification.
		totalW := 0
		for _, seg := range segments {
			totalW += seg.width
		}
		availW := w.Width - 2*t.inset - dl.leftMargin - dl.rightMargin
		justifyOffset := 0
		switch dl.justify {
		case option.JustifyCenter:
			justifyOffset = (availW - totalW) / 2
			if justifyOffset < 0 {
				justifyOffset = 0
			}
		case option.JustifyRight:
			justifyOffset = availW - totalW
			if justifyOffset < 0 {
				justifyOffset = 0
			}
		}

		xOffset := t.inset + dl.leftMargin + justifyOffset - t.xOffset
		for _, seg := range segments {
			segX := xOffset + seg.x
			segBaseY := baseY - seg.offset

			// Draw background if set.
			if seg.bg != nil {
				d.SetForeground(gc, seg.bg.Pixel)
				d.FillRectangle(pxDrawable, gc, segX, t.inset+dl.y, uint(seg.width), uint(dl.height))
			}

			// Draw selection highlight.
			t.drawSelectionHighlight(d, gc, pxDrawable, dl, seg, segX, xOffset)

			// Draw text.
			df := drawableFont
			if seg.font != nil {
				if sf, ok := seg.font.(platform.DrawableFont); ok {
					df = sf
				}
			}
			if seg.fg != nil {
				df.DrawString(pxDrawable, segX, segBaseY, seg.text,
					seg.fg.Pixel, seg.fg.Red, seg.fg.Green, seg.fg.Blue)
			}

			// Draw underline.
			if seg.underline && seg.fg != nil {
				underY := segBaseY + 2
				d.SetForeground(gc, seg.fg.Pixel)
				d.DrawLine(pxDrawable, gc, segX, underY, segX+seg.width, underY)
			}

			// Draw overstrike (strikethrough).
			if seg.overstrike && seg.fg != nil {
				strikeY := segBaseY - m.Ascent/2
				d.SetForeground(gc, seg.fg.Pixel)
				d.DrawLine(pxDrawable, gc, segX, strikeY, segX+seg.width, strikeY)
			}

			// Draw 3D relief border around the segment.
			if seg.reliefSet && seg.relief != option.ReliefFlat {
				border := t.Border
				if border == nil {
					bgPixelLocal := uint64(0xD9D9D9) // default gray
					if t.Background != nil {
						bgPixelLocal = t.Background.Pixel
					}
					border = draw.NewBorderFromPixel(bgPixelLocal)
				}
				bw := seg.reliefBW
				segRowY := t.inset + dl.y
				draw.Draw3DRectangle(d, pxDrawable, gc, border,
					segX, segRowY, seg.width, dl.height, bw, seg.relief)
			}
		}
	}

	// Draw cursor.
	if t.hasFocus && t.cursorOn {
		t.drawCursor(d, gc, pxDrawable, dlines)
	}
}

// drawSelectionHighlight draws selection highlight for a segment if applicable.
func (t *TextWidget) drawSelectionHighlight(d platform.DisplayServer, gc platform.GCID, drawable platform.DrawableID,
	dl displayLine, seg textSegment, segX, xOffset int) {
	selRanges := t.doc.TagRangesFor("sel")
	if len(selRanges) == 0 || t.selBg == nil {
		return
	}
	sr := selRanges[0]

	segStartIdx := Index{Line: dl.logicalLine, Char: dl.startChar}
	segEndIdx := Index{Line: dl.logicalLine, Char: dl.endChar}
	if Compare(segEndIdx, sr.Start) <= 0 || Compare(segStartIdx, sr.End) >= 0 {
		return
	}

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
func (t *TextWidget) drawCursor(d platform.DisplayServer, gc platform.GCID, drawable platform.DrawableID, dlines []displayLine) {
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
		lineText := t.doc.Lines[dl.logicalLine-1].Text
		cursorChars := insertPos.Char - dl.startChar
		var cursorX int
		if cursorChars > 0 && dl.startChar+cursorChars <= len(lineText) {
			cursorX = t.Font.MeasureString(string(lineText[dl.startChar : dl.startChar+cursorChars]))
		}
		cursorX += t.inset + dl.leftMargin - t.xOffset

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

	var dl *displayLine
	for i := range dlines {
		dlY := t.inset + dlines[i].y
		if y >= dlY && y < dlY+dlines[i].height {
			dl = &dlines[i]
			break
		}
	}
	if dl == nil {
		dl = &dlines[len(dlines)-1]
	}

	lineText := t.doc.Lines[dl.logicalLine-1].Text
	segText := lineText[dl.startChar:dl.endChar]
	xInContent := x - t.inset - dl.leftMargin + t.xOffset

	if len(segText) == 0 || t.Font == nil {
		return Index{Line: dl.logicalLine, Char: dl.startChar}
	}

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

// computeTotalHeight returns the total pixel height of all lines including spacing.
func (t *TextWidget) computeTotalHeight() int {
	if t.Font == nil {
		return 0
	}
	availWidth := t.Win.Width - 2*t.inset
	total := 0
	for lineIdx := 1; lineIdx <= t.doc.LineCount(); lineIdx++ {
		props := t.resolveLineProps(lineIdx)
		dls := t.wrapLine(lineIdx, availWidth, props.lm1, props.lm2, props.rm)
		for i, dl := range dls {
			sp := props.sp2
			if i == 0 {
				sp = props.sp1
			}
			sb := 0
			if i == len(dls)-1 {
				sb = props.sp3
			}
			total += sp + dl.height + sb
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

// computeDisplayLinesBefore returns the total number of display line slots from
// the top of the document to (but not including) the given logical line and offset.
func (t *TextWidget) computeDisplayLinesBefore(lineIdx, dlOffset int) int {
	availWidth := t.Win.Width - 2*t.inset
	count := 0
	for l := 1; l < lineIdx; l++ {
		props := t.resolveLineProps(l)
		count += len(t.wrapLine(l, availWidth, props.lm1, props.lm2, props.rm))
	}
	count += dlOffset
	return count
}

// totalDisplayLines returns the total number of display lines in the document.
func (t *TextWidget) totalDisplayLines() int {
	availWidth := t.Win.Width - 2*t.inset
	count := 0
	for l := 1; l <= t.doc.LineCount(); l++ {
		props := t.resolveLineProps(l)
		count += len(t.wrapLine(l, availWidth, props.lm1, props.lm2, props.rm))
	}
	return count
}
