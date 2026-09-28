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
			for range spaces {
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
	ascent       int // baseline offset from the top of the line
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
	bgStipple  platform.PixmapID // depth-1 stipple bitmap for background (0 = none)
	fgStipple  platform.PixmapID // depth-1 stipple bitmap for foreground (0 = none)
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
	lm1, lm2, rm  int
	sp1, sp2, sp3 int
	justify       option.Justify
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

// measureRange measures the pixel width of text from startChar to endChar
// on the given logical line, substituting embedded window widths for their
// placeholder characters.
func (t *TextWidget) measureRange(lineIdx, startChar, endChar int) int {
	if startChar >= endChar || t.Font == nil {
		return 0
	}
	line := t.doc.Lines[lineIdx-1]

	// Collect embedded window positions in this range.
	type winInfo struct {
		char  int
		width int
	}
	var wins []winInfo
	for _, ew := range t.embeddedWindows {
		m, ok := t.doc.Marks[ew.markName]
		if !ok || m.Pos.Line != lineIdx {
			continue
		}
		if m.Pos.Char >= startChar && m.Pos.Char < endChar {
			wins = append(wins, winInfo{char: m.Pos.Char, width: ew.win.ReqWidth + 2*ew.padX})
		}
	}

	if len(wins) == 0 {
		return t.Font.MeasureString(string(line.Text[startChar:endChar]))
	}

	// Sort by position (insertion sort for small N).
	for i := 1; i < len(wins); i++ {
		for j := i; j > 0 && wins[j].char < wins[j-1].char; j-- {
			wins[j], wins[j-1] = wins[j-1], wins[j]
		}
	}

	width := 0
	pos := startChar
	for _, wi := range wins {
		if wi.char > pos {
			width += t.Font.MeasureString(string(line.Text[pos:wi.char]))
		}
		width += wi.width
		pos = wi.char + 1
	}
	if pos < endChar {
		width += t.Font.MeasureString(string(line.Text[pos:endChar]))
	}
	return width
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
	h := t.lineHeightFor(lineIdx)

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
		w := t.measureRange(lineIdx, start, end)

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
			mw := t.measureRange(lineIdx, start, mid)
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
			isSpace := func(i int) bool { return line.Text[i] == ' ' || line.Text[i] == '\t' }
			if breakAt < end && isSpace(breakAt) {
				// TkTextCharLayoutProc: whitespace after the last word
				// that fits may run past the right edge.
				for breakAt < end && isSpace(breakAt) {
					breakAt++
				}
			} else {
				wb := breakAt
				for wb > start && !isSpace(wb-1) {
					wb--
				}
				if wb > start {
					breakAt = wb
				}
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

// lineMetrics ports how TkTextCharLayoutProc/LayoutDLine size a display
// line: the largest ascent and descent over its chunks (a tag -offset
// raises the ascent and lowers the descent), counting the terminating
// newline's tags on the last fragment of a logical line.
func (t *TextWidget) lineMetrics(lineIdx, start, end int, last bool) (ascent, descent int) {
	add := func(f font.Font, offset int) {
		if f == nil {
			return
		}
		m := f.Metrics()
		ascent = max(ascent, m.Ascent+offset)
		descent = max(descent, m.Descent-offset)
	}
	for _, seg := range t.segmentsForRange(lineIdx, start, end) {
		f := seg.font
		if f == nil {
			f = t.Font
		}
		add(f, seg.offset)
	}
	if last || start >= end {
		f := t.Font
		off := 0
		n := len(t.doc.Lines[lineIdx-1].Text)
		for _, tag := range t.doc.TagsAt(Index{Line: lineIdx, Char: n}) {
			if tag.Font != nil {
				f = tag.Font
			}
			if tag.OffsetSet {
				off = tag.Offset
			}
		}
		add(f, off)
	}
	return ascent, descent
}

// setMetrics ports LayoutDLine's height: the chunks' ascent+descent, at
// least the -align center minHeight of embedded windows and images in that
// fragment, with the baseline centred in any extra height.
func (t *TextWidget) setMetrics(lineIdx int, dls []displayLine) []displayLine {
	in := func(i int, char int) bool {
		dl := dls[i]
		return (char >= dl.startChar && char < dl.endChar) || (dl.startChar == dl.endChar && char == dl.startChar)
	}
	for i := range dls {
		a, d := t.lineMetrics(lineIdx, dls[i].startChar, dls[i].endChar, i == len(dls)-1)
		minH := 0
		for _, ei := range t.embeddedImages {
			// Images take no placeholder, so one at the end of the line
			// belongs to its last display line.
			atEnd := i == len(dls)-1 && ei.index.Char >= dls[i].endChar
			if ei.index.Line == lineIdx && (in(i, ei.index.Char) || atEnd) {
				minH = max(minH, ei.img.Height())
			}
		}
		for _, ew := range t.embeddedWindows {
			if m, ok := t.doc.Marks[ew.markName]; ok && m.Pos.Line == lineIdx && in(i, m.Pos.Char) {
				minH = max(minH, ew.win.ReqHeight+2*ew.padY)
			}
		}
		dls[i].height, dls[i].ascent = a+d, a
		if minH > a+d {
			dls[i].height = minH
			dls[i].ascent = a + (minH-a-d)/2
		}
	}
	return dls
}

// computeVisibleLines returns the display lines visible from the current scroll position.
func (t *TextWidget) computeVisibleLines() []displayLine {
	availWidth := t.Win.Width - 2*t.insetX
	availHeight := t.Win.Height - 2*t.insetY
	if availWidth <= 0 || availHeight <= 0 {
		return nil
	}

	var result []displayLine
	y := 0

	for lineIdx := t.topLine; lineIdx <= t.doc.LineCount() && y < availHeight; lineIdx++ {
		props := t.resolveLineProps(lineIdx)
		dls := t.setMetrics(lineIdx, t.wrapLine(lineIdx, availWidth, props.lm1, props.lm2, props.rm))
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

	// Add breakpoints around embedded window placeholders so each
	// placeholder rune becomes its own segment.
	for i := 0; i < endChar-startChar; i++ {
		if text[i] == runeEmbeddedWindow {
			breaks = append(breaks, i, i+1)
		}
	}
	// Re-sort and de-duplicate breaks.
	for i := 1; i < len(breaks); i++ {
		for j := i; j > 0 && breaks[j] < breaks[j-1]; j-- {
			breaks[j], breaks[j-1] = breaks[j-1], breaks[j]
		}
	}
	n := 1
	for i := 1; i < len(breaks); i++ {
		if breaks[i] != breaks[i-1] {
			breaks[n] = breaks[i]
			n++
		}
	}
	breaks = breaks[:n]

	// Build segments between breakpoints.
	var segments []textSegment
	x := 0
	for i := 0; i < len(breaks)-1; i++ {
		segStart := breaks[i]
		segEnd := breaks[i+1]
		segText := string(text[segStart:segEnd])
		segWidth := t.Font.MeasureString(segText)

		// Handle embedded window placeholder: use window width, empty text.
		if segEnd-segStart == 1 && text[segStart] == runeEmbeddedWindow {
			if ew := t.embeddedWinAt(lineIdx, startChar+segStart); ew != nil {
				// EmbWinLayoutProc: the chunk is the window plus -padx each side.
				segWidth = ew.win.ReqWidth + 2*ew.padX
				segText = ""
			}
		}

		// Resolve attributes at this position.
		fg := t.Foreground
		var bg *color.Color
		var bgStipple, fgStipple platform.PixmapID
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
			if tag.BgStipple != "" {
				bgStipple = t.stipplePixmap(tag.BgStipple)
			}
			if tag.FgStipple != "" {
				fgStipple = t.stipplePixmap(tag.FgStipple)
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
			bgStipple:  bgStipple,
			fgStipple:  fgStipple,
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

// renderToPixmap draws the text widget content to the offscreen pixmap and
// returns the display lines it laid out.
func (t *TextWidget) renderToPixmap() []displayLine {
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
		return nil
	}

	drawableFont, isDrawable := t.Font.(platform.DrawableFont)
	if !isDrawable {
		return nil
	}

	dlines := t.computeVisibleLines()

	// DisplayDLine draws each line into a window-wide pixmap whose top is
	// the line's top, so stipples start their pattern there.
	defer d.SetTSOrigin(gc, 0, 0)
	for _, dl := range dlines {
		baseY := t.insetY + dl.y + dl.ascent
		segments := t.segmentsForRange(dl.logicalLine, dl.startChar, dl.endChar)
		d.SetTSOrigin(gc, 0, t.insetY+dl.y-dl.spacingAbove)

		// Compute total segment width for justification.
		totalW := 0
		for _, seg := range segments {
			totalW += seg.width
		}
		availW := w.Width - 2*t.insetX - dl.leftMargin - dl.rightMargin
		justifyOffset := 0
		switch dl.justify {
		case option.JustifyCenter:
			justifyOffset = max((availW-totalW)/2, 0)
		case option.JustifyRight:
			justifyOffset = max(availW-totalW, 0)
		}

		xOffset := t.insetX + dl.leftMargin + justifyOffset - t.xOffset

		// Draw selection highlight once per display line, before any text.
		t.drawSelectionHighlight(d, gc, pxDrawable, dl, xOffset)

		for _, seg := range segments {
			segX := xOffset + seg.x
			segBaseY := baseY - seg.offset

			// Draw background if set.
			if seg.bg != nil {
				d.SetForeground(gc, seg.bg.Pixel)
				if seg.bgStipple != 0 {
					d.SetStipple(gc, seg.bgStipple)
					d.SetFillStyle(gc, platform.FillStippled)
					d.FillRectangle(pxDrawable, gc, segX, t.insetY+dl.y, uint(seg.width), uint(dl.height))
					d.SetFillStyle(gc, platform.FillSolid)
				} else {
					d.FillRectangle(pxDrawable, gc, segX, t.insetY+dl.y, uint(seg.width), uint(dl.height))
				}
			} else if seg.bgStipple != 0 {
				// Stipple with no explicit background color: use black dots on normal bg.
				d.SetForeground(gc, 0)
				d.SetStipple(gc, seg.bgStipple)
				d.SetFillStyle(gc, platform.FillStippled)
				d.FillRectangle(pxDrawable, gc, segX, t.insetY+dl.y, uint(seg.width), uint(dl.height))
				d.SetFillStyle(gc, platform.FillSolid)
			}

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

			// Tk_UnderlineChars for -underline, and for -overstrike raised by
			// descent + 3/10 of the ascent (CharDisplayProc).
			if (seg.underline || seg.overstrike) && seg.fg != nil {
				sf := seg.font
				if sf == nil {
					sf = t.Font
				}
				pos, h := font.Underline(sf)
				sm := sf.Metrics()
				d.SetForeground(gc, seg.fg.Pixel)
				if seg.underline {
					d.FillRectangle(pxDrawable, gc, segX, segBaseY+pos, uint(seg.width), uint(h))
				}
				if seg.overstrike {
					y := segBaseY - sm.Descent - sm.Ascent*3/10
					d.FillRectangle(pxDrawable, gc, segX, y+pos, uint(seg.width), uint(h))
				}
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
				segRowY := t.insetY + dl.y
				draw.Draw3DRectangle(d, pxDrawable, gc, border,
					segX, segRowY, seg.width, dl.height, bw, seg.relief)
			}
		}
	}

	// Draw cursor (not in read-only mode).
	if t.hasFocus && t.cursorOn && !t.readOnly {
		t.drawCursor(d, gc, pxDrawable, dlines)
	}

	// Draw inline images.
	if len(t.embeddedImages) > 0 {
		bgPx := uint64(0xFFFFFF)
		if t.Background != nil {
			bgPx = t.Background.Pixel
		}
		for _, ei := range t.embeddedImages {
			for _, dl := range dlines {
				if dl.logicalLine == ei.index.Line {
					imgX := t.insetX + dl.leftMargin
					imgY := t.insetY + dl.y
					ei.img.Draw(d, pxDrawable, gc, w.Depth,
						0, 0, ei.img.Width(), ei.img.Height(),
						imgX, imgY, bgPx)
					break
				}
			}
		}
	}
	return dlines
}

// drawSelectionHighlight draws the selection highlight for a display line if applicable.
func (t *TextWidget) drawSelectionHighlight(d platform.DisplayServer, gc platform.GCID, drawable platform.DrawableID,
	dl displayLine, xOffset int) {
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
	d.FillRectangle(drawable, gc, hlStartX, t.insetY+dl.y, uint(hlEndX-hlStartX), uint(dl.height))
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
		// Compute justifyOffset the same way renderToPixmap does.
		justifyOffset := 0
		if dl.justify != option.JustifyLeft {
			lineSegs := t.segmentsForRange(dl.logicalLine, dl.startChar, dl.endChar)
			totalW := 0
			for _, seg := range lineSegs {
				totalW += seg.width
			}
			availW := t.Win.Width - 2*t.insetX - dl.leftMargin - dl.rightMargin
			switch dl.justify {
			case option.JustifyCenter:
				justifyOffset = max((availW-totalW)/2, 0)
			case option.JustifyRight:
				justifyOffset = max(availW-totalW, 0)
			}
		}

		cursorX := t.insetX + dl.leftMargin + justifyOffset - t.xOffset
		if insertPos.Char > dl.startChar {
			segs := t.segmentsForRange(dl.logicalLine, dl.startChar, insertPos.Char)
			for _, seg := range segs {
				cursorX += seg.width
			}
		}

		// DisplayDLine renders into a line pixmap spanning only the text
		// area (inset + padx), which clips the cursor at its edges.
		x0 := max(cursorX-t.insertWidth/2, t.insetX)
		x1 := min(cursorX-t.insertWidth/2+t.insertWidth, t.Win.Width-t.insetX)
		if x1 > x0 {
			d.SetForeground(gc, t.insertColor.Pixel)
			d.FillRectangle(drawable, gc, x0, t.insetY+dl.y, uint(x1-x0), uint(dl.height))
		}
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
		dlY := t.insetY + dlines[i].y
		if y >= dlY && y < dlY+dlines[i].height {
			dl = &dlines[i]
			break
		}
	}
	if dl == nil {
		dl = &dlines[len(dlines)-1]
	}

	if t.Font == nil {
		return Index{Line: dl.logicalLine, Char: dl.startChar}
	}

	segments := t.segmentsForRange(dl.logicalLine, dl.startChar, dl.endChar)

	// Compute justifyOffset the same way renderToPixmap does.
	justifyOffset := 0
	if dl.justify != option.JustifyLeft {
		totalW := 0
		for _, seg := range segments {
			totalW += seg.width
		}
		availW := t.Win.Width - 2*t.insetX - dl.leftMargin - dl.rightMargin
		switch dl.justify {
		case option.JustifyCenter:
			justifyOffset = max((availW-totalW)/2, 0)
		case option.JustifyRight:
			justifyOffset = max(availW-totalW, 0)
		}
	}

	// x position relative to the start of the line's text content.
	xInContent := x - t.insetX - dl.leftMargin - justifyOffset + t.xOffset

	if len(segments) == 0 {
		return Index{Line: dl.logicalLine, Char: dl.startChar}
	}

	// Walk segments, finding the one that contains the click, then binary-search
	// within that segment using its actual font.
	runeOffset := 0
	for si, seg := range segments {
		segRunes := []rune(seg.text)
		isLast := si == len(segments)-1

		if xInContent < seg.x+seg.width || isLast {
			if xInContent <= seg.x {
				return Index{Line: dl.logicalLine, Char: dl.startChar + runeOffset}
			}
			xInSeg := xInContent - seg.x
			f := seg.font
			if f == nil {
				f = t.Font
			}
			lo, hi := 0, len(segRunes)
			for lo < hi {
				mid := (lo + hi) / 2
				if f.MeasureString(string(segRunes[:mid+1])) <= xInSeg {
					lo = mid + 1
				} else {
					hi = mid
				}
			}
			charIdx := lo
			if charIdx < len(segRunes) {
				var charStart int
				if charIdx > 0 {
					charStart = f.MeasureString(string(segRunes[:charIdx]))
				}
				charEnd := f.MeasureString(string(segRunes[:charIdx+1]))
				if xInSeg >= (charStart+charEnd)/2 {
					charIdx++
				}
			}
			return Index{Line: dl.logicalLine, Char: dl.startChar + runeOffset + charIdx}
		}
		runeOffset += len(segRunes)
	}

	return Index{Line: dl.logicalLine, Char: dl.endChar}
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
	availWidth := t.Win.Width - 2*t.insetX
	count := 0
	for l := 1; l < lineIdx; l++ {
		props := t.resolveLineProps(l)
		count += len(t.wrapLine(l, availWidth, props.lm1, props.lm2, props.rm))
	}
	count += dlOffset
	return count
}

// displayLinePixels returns the heights of lineIdx's display lines including
// their -spacing1/2/3, as computeVisibleLines stacks them.
func (t *TextWidget) displayLinePixels(lineIdx int) []int {
	availWidth := t.Win.Width - 2*t.insetX
	props := t.resolveLineProps(lineIdx)
	dls := t.setMetrics(lineIdx, t.wrapLine(lineIdx, availWidth, props.lm1, props.lm2, props.rm))
	hs := make([]int, len(dls))
	for i, dl := range dls {
		h := dl.height + props.sp2
		if i == 0 {
			h += props.sp1 - props.sp2
		}
		if i == len(dls)-1 {
			h += props.sp3
		}
		hs[i] = h
	}
	return hs
}

// totalDisplayLines returns the total number of display lines in the document.
func (t *TextWidget) totalDisplayLines() int {
	availWidth := t.Win.Width - 2*t.insetX
	count := 0
	for l := 1; l <= t.doc.LineCount(); l++ {
		props := t.resolveLineProps(l)
		count += len(t.wrapLine(l, availWidth, props.lm1, props.lm2, props.rm))
	}
	return count
}
