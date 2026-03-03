// Package entry implements the entry widget for single-line text input.
// It ports tk/generic/tkEntry.c.
package entry

import (
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// Entry is a single-line text input widget.
type Entry struct {
	widget.Base

	// Text state.
	text    []rune // the content
	ShowChar rune  // 0 = normal, else password char (e.g. '*')

	// Cursor.
	InsertPos int // rune index of cursor (0..len(text))

	// Selection (half-open interval [SelFirst, SelLast)).
	SelFirst  int // -1 = no selection
	SelLast   int // -1 = no selection
	SelAnchor int // fixed end of selection

	// Scroll state.
	LeftIndex int // rune index of leftmost visible char

	// Layout.
	layoutX  int // pixel offset: windowX = charX + layoutX
	layoutY  int // pixel Y of text baseline
	inset    int // border + highlight + padding
	avgWidth int // average char width

	// Visual config.
	InsertWidth int
	SelBg       *colorRef
	SelFg       *colorRef
	InsertBg    *colorRef
	Anchor      option.Anchor
	Justify     option.Justify
	PrefWidth   int // preferred width in characters

	// Placeholder.
	Placeholder    string
	PlaceholderFg  *colorRef

	// Blink.
	CursorOn bool
	HasFocus bool

	// Scrollbar callback.
	ScrollCmd func(first, last float64)

	// Scan state (middle-button pan).
	scanMarkX     int
	scanMarkIndex int
}

type colorRef struct {
	Pixel uint64
	Red   uint16
	Green uint16
	Blue  uint16
}

// EntryOption configures an Entry.
type EntryOption func(*Entry)

// Text sets the initial text.
func Text(s string) EntryOption {
	return func(e *Entry) { e.text = []rune(s) }
}

// Placeholder sets the placeholder text.
func Placeholder(s string) EntryOption {
	return func(e *Entry) { e.Placeholder = s }
}

// Show sets the password character (0 = normal display).
func Show(ch rune) EntryOption {
	return func(e *Entry) { e.ShowChar = ch }
}

// Background sets the background color.
func Background(name string) EntryOption {
	return func(e *Entry) {
		col, err := e.App.ColorCache().Get(name)
		if err == nil {
			e.Background = col
			e.UpdateBorder()
		}
	}
}

// Foreground sets the text color.
func Foreground(name string) EntryOption {
	return func(e *Entry) {
		col, err := e.App.ColorCache().Get(name)
		if err == nil {
			e.Foreground = col
		}
	}
}

// FontOpt sets the font.
func FontOpt(name string) EntryOption {
	return func(e *Entry) {
		f, err := e.App.FontRegistry().Get(name)
		if err == nil {
			e.Font = f
		}
	}
}

// Width sets the preferred width in characters.
func Width(w int) EntryOption {
	return func(e *Entry) { e.PrefWidth = w }
}

// BorderWidth sets the border width.
func BorderWidth(w int) EntryOption {
	return func(e *Entry) { e.BorderWidth = w }
}

// ScrollCommand sets the callback for scrollbar notification.
func ScrollCommand(fn func(first, last float64)) EntryOption {
	return func(e *Entry) { e.ScrollCmd = fn }
}

// New creates a new Entry widget.
func New(parent widget.Caregiver, name string, opts ...EntryOption) *Entry {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(w)

	e := &Entry{
		SelFirst:    -1,
		SelLast:     -1,
		InsertWidth: 2,
		PrefWidth:   20,
		CursorOn:    true,
		Anchor:      option.AnchorCenter,
		Justify:     option.JustifyLeft,
	}
	widget.InitBase(&e.Base, w, app)

	// Entry-specific defaults.
	e.BorderWidth = 2
	e.Relief = option.ReliefSunken
	e.HighlightWidth = 1

	// Default colors.
	if bg, err := app.ColorCache().Get("#ffffff"); err == nil {
		e.Background = bg
		e.UpdateBorder()
	}

	if sel, err := app.ColorCache().Get("#3399ff"); err == nil {
		e.SelBg = &colorRef{sel.Pixel, sel.Red, sel.Green, sel.Blue}
	}
	if selfg, err := app.ColorCache().Get("#ffffff"); err == nil {
		e.SelFg = &colorRef{selfg.Pixel, selfg.Red, selfg.Green, selfg.Blue}
	}
	if ins, err := app.ColorCache().Get("#000000"); err == nil {
		e.InsertBg = &colorRef{ins.Pixel, ins.Red, ins.Green, ins.Blue}
	}
	if phfg, err := app.ColorCache().Get("#a3a3a3"); err == nil {
		e.PlaceholderFg = &colorRef{phfg.Pixel, phfg.Red, phfg.Green, phfg.Blue}
	}

	for _, opt := range opts {
		opt(e)
	}

	e.computeGeometry()

	if e.Background != nil {
		w.BackgroundPixel = e.Background.Pixel
	}

	w.Flags |= window.FlagFocusable
	bindEntry(e, app)

	return e
}

// GetText returns the entry's text as a string.
func (e *Entry) GetText() string {
	return string(e.text)
}

// SetText sets the entry's text.
func (e *Entry) SetText(s string) {
	e.text = []rune(s)
	if e.InsertPos > len(e.text) {
		e.InsertPos = len(e.text)
	}
	e.ClearSelection()
	e.computeGeometry()
	e.Display()
}

// InsertChars inserts text at the given rune index.
func (e *Entry) InsertChars(index int, s string) {
	if len(s) == 0 {
		return
	}
	runes := []rune(s)
	count := len(runes)

	if index < 0 {
		index = 0
	}
	if index > len(e.text) {
		index = len(e.text)
	}

	// Splice in new text.
	newText := make([]rune, 0, len(e.text)+count)
	newText = append(newText, e.text[:index]...)
	newText = append(newText, runes...)
	newText = append(newText, e.text[index:]...)
	e.text = newText

	// Adjust indexes.
	if e.InsertPos >= index {
		e.InsertPos += count
	}
	if e.SelFirst >= index {
		e.SelFirst += count
	} else if e.SelFirst >= 0 && e.SelFirst < index {
		// no change
	}
	if e.SelLast > index {
		e.SelLast += count
	}
	if e.SelAnchor >= index {
		e.SelAnchor += count
	}
	if e.LeftIndex > index {
		e.LeftIndex += count
	}

	e.computeGeometry()
	e.seeInsert()
	e.notifyScrollbar()
	e.Display()
}

// DeleteChars deletes count runes starting at index.
func (e *Entry) DeleteChars(index, count int) {
	if count <= 0 || len(e.text) == 0 {
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= len(e.text) {
		return
	}
	if index+count > len(e.text) {
		count = len(e.text) - index
	}

	// Remove range.
	e.text = append(e.text[:index], e.text[index+count:]...)

	// Adjust indexes.
	adjustIndex := func(idx *int) {
		if *idx < 0 {
			return
		}
		if *idx >= index+count {
			*idx -= count
		} else if *idx >= index {
			*idx = index
		}
	}
	adjustIndex(&e.InsertPos)
	adjustIndex(&e.SelFirst)
	adjustIndex(&e.SelLast)
	adjustIndex(&e.SelAnchor)
	adjustIndex(&e.LeftIndex)

	// Clear selection if empty.
	if e.SelFirst >= 0 && e.SelLast <= e.SelFirst {
		e.SelFirst = -1
		e.SelLast = -1
	}

	e.computeGeometry()
	e.seeInsert()
	e.notifyScrollbar()
	e.Display()
}

// DeleteSelection deletes the selected text.
func (e *Entry) DeleteSelection() {
	if e.SelFirst < 0 {
		return
	}
	e.DeleteChars(e.SelFirst, e.SelLast-e.SelFirst)
}

// ClearSelection clears the selection.
func (e *Entry) ClearSelection() {
	e.SelFirst = -1
	e.SelLast = -1
}

// SelectRange sets the selection range.
func (e *Entry) SelectRange(first, last int) {
	if first < 0 {
		first = 0
	}
	if last > len(e.text) {
		last = len(e.text)
	}
	if first >= last {
		e.ClearSelection()
		return
	}
	e.SelFirst = first
	e.SelLast = last
}

// SelectAll selects all text.
func (e *Entry) SelectAll() {
	if len(e.text) > 0 {
		e.SelectRange(0, len(e.text))
	}
}

// SelectedText returns the currently selected text.
func (e *Entry) SelectedText() string {
	if e.SelFirst < 0 || e.SelFirst >= e.SelLast {
		return ""
	}
	return string(e.text[e.SelFirst:e.SelLast])
}

// displayText returns the text to display (handles password mode).
func (e *Entry) displayText() []rune {
	if e.ShowChar != 0 {
		dt := make([]rune, len(e.text))
		for i := range dt {
			dt[i] = e.ShowChar
		}
		return dt
	}
	return e.text
}

// computeGeometry recalculates layout after text or size changes.
func (e *Entry) computeGeometry() {
	if e.Font == nil {
		return
	}

	m := e.Font.Metrics()
	e.inset = e.BorderWidth + e.HighlightWidth + 1 // +1 for XPAD
	e.avgWidth = e.Font.MeasureString("0")
	if e.avgWidth < 1 {
		e.avgWidth = 1
	}

	// Set requested size.
	w := e.Win
	w.ReqWidth = e.PrefWidth*e.avgWidth + 2*e.inset
	w.ReqHeight = m.Linespace() + 2*e.inset

	// Layout Y: center text vertically.
	e.layoutY = e.inset + m.Ascent

	dt := e.displayText()
	totalWidth := measureRunes(e.Font, dt)
	availWidth := w.Width - 2*e.inset
	if availWidth < 1 {
		availWidth = 1
	}

	if totalWidth <= availWidth {
		// Text fits — no scrolling needed.
		e.LeftIndex = 0
		e.layoutX = e.inset
	} else {
		// Clamp leftIndex.
		maxOff := runeIndexAtPixel(e.Font, dt, totalWidth-availWidth)
		if e.LeftIndex > maxOff {
			e.LeftIndex = maxOff
		}
		if e.LeftIndex < 0 {
			e.LeftIndex = 0
		}
		leftCharX := measureRunes(e.Font, dt[:e.LeftIndex])
		e.layoutX = e.inset - leftCharX
	}
}

// seeInsert scrolls to make the cursor visible.
func (e *Entry) seeInsert() {
	if e.Font == nil {
		return
	}

	dt := e.displayText()
	availWidth := e.Win.Width - 2*e.inset
	if availWidth < 1 {
		return
	}

	if e.InsertPos < e.LeftIndex {
		e.LeftIndex = e.InsertPos
		e.computeGeometry()
	} else {
		cursorX := measureRunes(e.Font, dt[:e.InsertPos]) + e.layoutX
		if cursorX >= e.Win.Width-e.inset {
			e.LeftIndex = e.InsertPos - availWidth/e.avgWidth
			if e.LeftIndex < 0 {
				e.LeftIndex = 0
			}
			e.computeGeometry()
		}
	}
}

// charAtPixel returns the rune index closest to pixel x.
func (e *Entry) charAtPixel(x int) int {
	if e.Font == nil {
		return 0
	}
	dt := e.displayText()
	xInLayout := x - e.layoutX
	return runeIndexAtPixel(e.Font, dt, xInLayout)
}

// closestGap returns the rune index of the nearest inter-character gap.
func (e *Entry) closestGap(x int) int {
	if e.Font == nil || len(e.text) == 0 {
		return 0
	}
	dt := e.displayText()
	xInLayout := x - e.layoutX

	idx := runeIndexAtPixel(e.Font, dt, xInLayout)
	if idx >= len(dt) {
		return len(dt)
	}

	// Check if x is past the midpoint of the character.
	charStart := measureRunes(e.Font, dt[:idx])
	charEnd := measureRunes(e.Font, dt[:idx+1])
	mid := (charStart + charEnd) / 2
	if xInLayout >= mid {
		return idx + 1
	}
	return idx
}

// XView scrolls the entry.
func (e *Entry) XView(index int) {
	if index < 0 {
		index = 0
	}
	if index > len(e.text) {
		index = len(e.text)
	}
	e.LeftIndex = index
	e.computeGeometry()
	e.notifyScrollbar()
	e.Display()
}

// XViewScroll scrolls by count units or pages.
func (e *Entry) XViewScroll(count int, pages bool) {
	if pages {
		availWidth := e.Win.Width - 2*e.inset
		charsPerPage := availWidth/e.avgWidth - 2
		if charsPerPage < 1 {
			charsPerPage = 1
		}
		count *= charsPerPage
	}
	e.XView(e.LeftIndex + count)
}

// XViewMoveTo scrolls to a fraction of the total text.
func (e *Entry) XViewMoveTo(fraction float64) {
	index := int(fraction*float64(len(e.text)) + 0.5)
	e.XView(index)
}

// VisibleRange returns the fraction of text currently visible.
func (e *Entry) VisibleRange() (float64, float64) {
	n := len(e.text)
	if n == 0 {
		return 0, 1
	}

	first := float64(e.LeftIndex) / float64(n)

	// Approximate chars visible.
	availWidth := e.Win.Width - 2*e.inset
	dt := e.displayText()
	charsVisible := runeIndexAtPixel(e.Font, dt[e.LeftIndex:], availWidth)
	last := float64(e.LeftIndex+charsVisible) / float64(n)
	if last > 1 {
		last = 1
	}
	return first, last
}

// notifyScrollbar calls the scroll command if set.
func (e *Entry) notifyScrollbar() {
	if e.ScrollCmd != nil {
		first, last := e.VisibleRange()
		e.ScrollCmd(first, last)
	}
}

// Display draws the entry widget.
func (e *Entry) Display() {
	if e.Destroyed {
		return
	}
	w := e.Win
	if w.PlatformID == platform.WindowID(0) {
		return
	}

	d := w.Display.Server
	gc := w.GC

	// Background.
	if e.Background != nil {
		d.SetForeground(gc, e.Background.Pixel)
	}
	d.FillRectangle(w.Drawable(), gc, 0, 0, uint(w.Width), uint(w.Height))

	// Border.
	if e.Border != nil && e.BorderWidth > 0 {
		draw.Draw3DRectangle(d, w.Drawable(), gc, e.Border,
			0, 0, w.Width, w.Height, e.BorderWidth, e.Relief)
	}

	dt := e.displayText()
	xftFont, isXft := e.Font.(platform.DrawableFont)

	if len(e.text) == 0 && e.Placeholder != "" && !e.HasFocus {
		// Draw placeholder.
		if isXft && e.PlaceholderFg != nil {
			xftFont.DrawString(w.Drawable(), e.inset, e.layoutY, e.Placeholder,
				e.PlaceholderFg.Pixel, e.PlaceholderFg.Red, e.PlaceholderFg.Green, e.PlaceholderFg.Blue)
		}
	} else if isXft && len(dt) > 0 {
		// Draw selection highlight.
		if e.HasFocus && e.SelFirst >= 0 && e.SelLast > e.SelFirst && e.SelBg != nil {
			selStartX := measureRunes(e.Font, dt[:clampIdx(e.SelFirst, len(dt))]) + e.layoutX
			selEndX := measureRunes(e.Font, dt[:clampIdx(e.SelLast, len(dt))]) + e.layoutX

			if selStartX < e.inset {
				selStartX = e.inset
			}
			rightEdge := w.Width - e.inset
			if selEndX > rightEdge {
				selEndX = rightEdge
			}

			if selEndX > selStartX {
				m := e.Font.Metrics()
				d.SetForeground(gc, e.SelBg.Pixel)
				d.FillRectangle(w.Drawable(), gc, selStartX, e.inset,
					uint(selEndX-selStartX), uint(m.Linespace()))
			}
		}

		// Draw text in segments: before selection, selection, after selection.
		visibleStr := string(dt)
		if e.HasFocus && e.SelFirst >= 0 && e.SelLast > e.SelFirst && e.SelFg != nil && e.Foreground != nil {
			// Before selection.
			if e.SelFirst > 0 {
				seg := string(dt[:e.SelFirst])
				segX := e.layoutX
				xftFont.DrawString(w.Drawable(), segX, e.layoutY, seg,
					e.Foreground.Pixel, e.Foreground.Red, e.Foreground.Green, e.Foreground.Blue)
			}
			// Selection.
			if e.SelFirst < len(dt) {
				segStart := clampIdx(e.SelFirst, len(dt))
				segEnd := clampIdx(e.SelLast, len(dt))
				seg := string(dt[segStart:segEnd])
				segX := measureRunes(e.Font, dt[:segStart]) + e.layoutX
				xftFont.DrawString(w.Drawable(), segX, e.layoutY, seg,
					e.SelFg.Pixel, e.SelFg.Red, e.SelFg.Green, e.SelFg.Blue)
			}
			// After selection.
			if e.SelLast < len(dt) {
				seg := string(dt[e.SelLast:])
				segX := measureRunes(e.Font, dt[:e.SelLast]) + e.layoutX
				xftFont.DrawString(w.Drawable(), segX, e.layoutY, seg,
					e.Foreground.Pixel, e.Foreground.Red, e.Foreground.Green, e.Foreground.Blue)
			}
		} else if e.Foreground != nil {
			// No selection — draw all text in one go.
			xftFont.DrawString(w.Drawable(), e.layoutX, e.layoutY, visibleStr,
				e.Foreground.Pixel, e.Foreground.Red, e.Foreground.Green, e.Foreground.Blue)
		}
	}

	// Draw cursor (outside text block so it works for empty entries too).
	if e.HasFocus && e.CursorOn && e.InsertBg != nil && e.Font != nil {
		cursorX := measureRunes(e.Font, dt[:clampIdx(e.InsertPos, len(dt))]) + e.layoutX
		if cursorX >= e.inset && cursorX < w.Width-e.inset {
			m := e.Font.Metrics()
			d.SetForeground(gc, e.InsertBg.Pixel)
			d.FillRectangle(w.Drawable(), gc,
				cursorX-e.InsertWidth/2, e.inset,
				uint(e.InsertWidth), uint(m.Linespace()))
		}
	}

	d.Flush()
}

// Configure applies options.
func (e *Entry) Configure(opts ...option.Option) {
	option.Apply(e, opts)
	e.UpdateBorder()
	e.computeGeometry()
	if e.Background != nil {
		e.Win.BackgroundPixel = e.Background.Pixel
	}
	e.Display()
}

// Destroy cleans up the entry.
func (e *Entry) Destroy() {
	if e.Destroyed {
		return
	}
	e.Destroyed = true
	window.DestroyWindow(e.Win)
}

// Helper: measure pixel width of a rune slice.
func measureRunes(f font.Font, runes []rune) int {
	if len(runes) == 0 {
		return 0
	}
	return f.MeasureString(string(runes))
}

// Helper: find rune index at a given pixel offset.
func runeIndexAtPixel(f font.Font, runes []rune, targetX int) int {
	if targetX <= 0 || len(runes) == 0 {
		return 0
	}
	// Binary search for the rune index.
	lo, hi := 0, len(runes)
	for lo < hi {
		mid := (lo + hi) / 2
		w := measureRunes(f, runes[:mid+1])
		if w <= targetX {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo
}

// clampIdx clamps an index to [0, max].
func clampIdx(idx, max int) int {
	if idx < 0 {
		return 0
	}
	if idx > max {
		return max
	}
	return idx
}

// wordStart returns the rune index of the start of the word at or before pos.
func wordStart(text []rune, pos int) int {
	if pos <= 0 {
		return 0
	}
	if pos > len(text) {
		pos = len(text)
	}
	// Skip back past non-word chars.
	i := pos - 1
	for i > 0 && !isWordChar(text[i]) {
		i--
	}
	// Skip back past word chars.
	for i > 0 && isWordChar(text[i-1]) {
		i--
	}
	return i
}

// wordEnd returns the rune index past the end of the word at or after pos.
func wordEnd(text []rune, pos int) int {
	if pos >= len(text) {
		return len(text)
	}
	i := pos
	// Skip past word chars.
	for i < len(text) && isWordChar(text[i]) {
		i++
	}
	// Skip past non-word chars.
	for i < len(text) && !isWordChar(text[i]) {
		i++
	}
	return i
}

func isWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') || r == '_'
}

