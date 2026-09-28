// Package entry implements the entry widget for single-line text input.
// It ports tk/generic/tkEntry.c.
package entry

import (
	"log"

	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/cursor"
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/entryutil"
	"github.com/msorc/takigo/window"
)

// Entry is a single-line text input widget.
type Entry struct {
	widget.Base

	// Text state.
	text     []rune // the content
	ShowChar rune   // 0 = normal, else password char (e.g. '*')

	// Cursor.
	InsertPos int // rune index of cursor (0..len(text))

	// Selection (half-open interval [SelFirst, SelLast)).
	SelFirst  int // -1 = no selection
	SelLast   int // -1 = no selection
	SelAnchor int // fixed end of selection
	imeMark   int // insert position when the input method began composing

	// Scroll state.
	LeftIndex int // rune index of leftmost visible char

	// Layout.
	layoutX  int // pixel offset: windowX = charX + layoutX
	layoutY  int // pixel Y of text baseline
	inset    int // border + highlight + padding
	avgWidth int // average char width

	// Visual config.
	InsertWidth int
	SelBg       *color.ColorRef
	SelFg       *color.ColorRef
	InsertBg    *color.ColorRef
	Anchor      option.Anchor
	Justify     option.Justify
	PrefWidth   int // preferred width in characters

	// Placeholder.
	Placeholder   string
	PlaceholderFg *color.ColorRef

	// Validation.
	// Validate is when to validate: "", "none", "key", "focus", "focusin", "focusout", "all".
	Validate string
	// ValidateCmd is called with the prospective new value; returns true to allow, false to reject.
	ValidateCmd func(string) bool

	// Blink.
	CursorOn bool
	HasFocus bool

	// Scrollbar callback.
	ScrollCmd func(first, last float64)
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

// PlaceholderForeground sets the placeholder text color.
func PlaceholderForeground(name string) EntryOption {
	return func(e *Entry) {
		col, err := e.App.ColorCache().Get(name)
		if err != nil {
			log.Printf("entry: failed to get color %q: %v", name, err)
			return
		}
		e.PlaceholderFg = col.Ref()
	}
}

// Show sets the password character (0 = normal display).
func Show(ch rune) EntryOption {
	return func(e *Entry) { e.ShowChar = ch }
}

// Background sets the background color.
func Background(name string) EntryOption {
	return func(e *Entry) {
		col, err := e.App.ColorCache().Get(name)
		if err != nil {
			log.Printf("entry: failed to get color %q: %v", name, err)
			return
		}
		e.Background = col
		e.UpdateBorder()
	}
}

// Foreground sets the text color.
func Foreground(name string) EntryOption {
	return func(e *Entry) {
		col, err := e.App.ColorCache().Get(name)
		if err != nil {
			log.Printf("entry: failed to get color %q: %v", name, err)
			return
		}
		e.Foreground = col
	}
}

// FontOpt sets the font.
func FontOpt(name string) EntryOption {
	return func(e *Entry) {
		f, err := e.App.FontRegistry().Get(name)
		if err != nil {
			log.Printf("entry: failed to get font %q: %v", name, err)
			return
		}
		e.Font = f
	}
}

// Width sets the preferred width in characters.
func Width(w int) EntryOption {
	return func(e *Entry) { e.PrefWidth = w }
}

// ValidateOpt sets when to validate: "none", "key", "focus", "focusin", "focusout", "all".
func ValidateOpt(v string) EntryOption {
	return func(e *Entry) { e.Validate = v }
}

// ValidateCmdOpt sets the validation callback called with the prospective value.
// Returns true to allow the change, false to reject it.
func ValidateCmdOpt(fn func(string) bool) EntryOption {
	return func(e *Entry) { e.ValidateCmd = fn }
}

// BorderWidth sets the border width.
func BorderWidth(w int) EntryOption {
	return func(e *Entry) { e.BorderWidth = w }
}

// ScrollCommand sets the callback for scrollbar notification.
func ScrollCommand(fn func(first, last float64)) EntryOption {
	return func(e *Entry) { e.ScrollCmd = fn }
}

// --- Ttk-compatible aliases (prefix with Entry) for consistent naming ---
// These aliases match the naming convention used by ttk widgets (ttk.EntryText, etc.)
// allowing consistent option naming when both classic and ttk widgets are used.

// EntryText is an alias for Text.
var EntryText = Text

// EntryPlaceholder is an alias for Placeholder.
var EntryPlaceholder = Placeholder

// EntryPlaceholderForeground is an alias for PlaceholderForeground.
var EntryPlaceholderForeground = PlaceholderForeground

// EntryShow is an alias for Show.
var EntryShow = Show

// EntryBackground is an alias for Background.
var EntryBackground = Background

// EntryForeground is an alias for Foreground.
var EntryForeground = Foreground

// EntryFontOpt is an alias for FontOpt.
var EntryFontOpt = FontOpt

// EntryWidth is an alias for Width.
var EntryWidth = Width

// EntryValidateOpt is an alias for ValidateOpt.
var EntryValidateOpt = ValidateOpt

// EntryValidateCmdOpt is an alias for ValidateCmdOpt.
var EntryValidateCmdOpt = ValidateCmdOpt

// EntryBorderWidth is an alias for BorderWidth.
var EntryBorderWidth = BorderWidth

// EntryScrollCommand is an alias for ScrollCommand.
var EntryScrollCommand = ScrollCommand

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
	e.SetDisplayProc(e.display)
	w.Class = "Entry"

	// Entry-specific defaults.
	e.BorderWidth = widget.DefBorderWidth
	e.Relief = option.ReliefSunken
	e.HighlightWidth = 1

	// Default colors.
	if bg, err := app.ColorCache().Get("#ffffff"); err == nil {
		e.Background = bg
		e.UpdateBorder()
	}

	if sel, err := app.ColorCache().Get("#3399ff"); err == nil {
		e.SelBg = sel.Ref()
	}
	if selfg, err := app.ColorCache().Get("#ffffff"); err == nil {
		e.SelFg = selfg.Ref()
	}
	if ins, err := app.ColorCache().Get("#000000"); err == nil {
		e.InsertBg = ins.Ref()
	}
	if phfg, err := app.ColorCache().Get("#a3a3a3"); err == nil {
		e.PlaceholderFg = phfg.Ref()
	}

	for _, opt := range opts {
		opt(e)
	}

	e.computeGeometry()

	if e.Background != nil {
		w.BackgroundPixel = e.Background.Pixel
	}

	w.Flags |= window.FlagFocusable
	w.SetCursor(uint(cursor.XTerm))
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
	totalWidth := entryutil.MeasureRunes(e.Font, dt)
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
		maxOff := entryutil.RuneIndexAtPixel(e.Font, dt, totalWidth-availWidth)
		if e.LeftIndex > maxOff {
			e.LeftIndex = maxOff
		}
		if e.LeftIndex < 0 {
			e.LeftIndex = 0
		}
		leftCharX := entryutil.MeasureRunes(e.Font, dt[:e.LeftIndex])
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
		cursorX := entryutil.MeasureRunes(e.Font, dt[:e.InsertPos]) + e.layoutX
		if cursorX >= e.Win.Width-e.inset {
			e.LeftIndex = e.InsertPos - availWidth/e.avgWidth
			if e.LeftIndex < 0 {
				e.LeftIndex = 0
			}
			e.computeGeometry()
		}
	}
}

// closestGap returns the rune index of the nearest inter-character gap.
func (e *Entry) closestGap(x int) int {
	if e.Font == nil || len(e.text) == 0 {
		return 0
	}
	dt := e.displayText()
	xInLayout := x - e.layoutX

	idx := entryutil.RuneIndexAtPixel(e.Font, dt, xInLayout)
	if idx >= len(dt) {
		return len(dt)
	}

	// Check if x is past the midpoint of the character.
	charStart := entryutil.MeasureRunes(e.Font, dt[:idx])
	charEnd := entryutil.MeasureRunes(e.Font, dt[:idx+1])
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

	// EntryVisibleRange: the character under the last pixel inside the
	// border (Tk_PointToChar), counting a partly visible one.
	dt := e.displayText()
	x := e.Win.Width - e.inset - e.layoutX - 1
	chars := 0
	for chars < len(dt) && entryutil.MeasureRunes(e.Font, dt[:chars+1]) <= x {
		chars++
	}
	if chars < n {
		chars++
	}
	visible := max(chars-e.LeftIndex, 1)
	return first, float64(e.LeftIndex+visible) / float64(n)
}

// tryEdit checks whether a proposed edit is valid.
// prospective is the text that would result from the edit.
// Returns true if the edit should be allowed (no validator set, or validator approves).
func (e *Entry) tryEdit(prospective string) bool {
	if e.ValidateCmd == nil {
		return true
	}
	v := e.Validate
	if v != "key" && v != "all" {
		return true
	}
	return e.ValidateCmd(prospective)
}

// tryFocusValidate runs focus-triggered validation. Returns true if valid.
func (e *Entry) tryFocusValidate(trigger string) bool {
	if e.ValidateCmd == nil {
		return true
	}
	v := e.Validate
	if v == "all" || v == trigger || (v == "focus" && (trigger == "focusin" || trigger == "focusout")) {
		return e.ValidateCmd(string(e.text))
	}
	return true
}

// notifyScrollbar calls the scroll command if set.
func (e *Entry) notifyScrollbar() {
	if e.ScrollCmd != nil {
		first, last := e.VisibleRange()
		e.ScrollCmd(first, last)
	}
}

// Display schedules a redraw at idle time; see widget.Base.EventuallyRedraw.
func (e *Entry) Display() {
	e.EventuallyRedraw()
}

// display draws the entry widget.
func (e *Entry) display() {
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
			selStartX := entryutil.MeasureRunes(e.Font, dt[:entryutil.ClampIdx(e.SelFirst, len(dt))]) + e.layoutX
			selEndX := entryutil.MeasureRunes(e.Font, dt[:entryutil.ClampIdx(e.SelLast, len(dt))]) + e.layoutX

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
				segStart := entryutil.ClampIdx(e.SelFirst, len(dt))
				segEnd := entryutil.ClampIdx(e.SelLast, len(dt))
				seg := string(dt[segStart:segEnd])
				segX := entryutil.MeasureRunes(e.Font, dt[:segStart]) + e.layoutX
				xftFont.DrawString(w.Drawable(), segX, e.layoutY, seg,
					e.SelFg.Pixel, e.SelFg.Red, e.SelFg.Green, e.SelFg.Blue)
			}
			// After selection.
			if e.SelLast < len(dt) {
				seg := string(dt[e.SelLast:])
				segX := entryutil.MeasureRunes(e.Font, dt[:e.SelLast]) + e.layoutX
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
		cursorX := entryutil.MeasureRunes(e.Font, dt[:entryutil.ClampIdx(e.InsertPos, len(dt))]) + e.layoutX
		if cursorX >= e.inset && cursorX < w.Width-e.inset {
			m := e.Font.Metrics()
			d.SetForeground(gc, e.InsertBg.Pixel)
			d.FillRectangle(w.Drawable(), gc,
				cursorX-e.InsertWidth/2, e.inset,
				uint(e.InsertWidth), uint(m.Linespace()))
		}
	}

	// DisplayEntry draws the border and focus highlight last, so they cover
	// text that runs past the viewable part of the window.
	if e.Border != nil && e.BorderWidth > 0 {
		hl := e.HighlightWidth
		draw.Draw3DRectangle(d, w.Drawable(), gc, e.Border,
			hl, hl, w.Width-2*hl, w.Height-2*hl, e.BorderWidth, e.Relief)
	}
	// The highlight ring sits outside the border (focus colour or
	// -highlightbackground).
	e.DrawHighlightBorder(e.HasFocus, 0)

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
