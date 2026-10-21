// Package spinbox implements a spinbox widget — an entry with up/down
// spinner buttons for numeric or list-based value selection.
package spinbox

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"time"

	"github.com/takigo/takigo/color"
	"github.com/takigo/takigo/cursor"
	"github.com/takigo/takigo/draw"
	"github.com/takigo/takigo/internal/textedit"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/internal/entryedit"
	"github.com/takigo/takigo/window"
)

// Spinbox is a single-line text entry with up/down spinner buttons.
type Spinbox struct {
	widget.Base
	entryedit.Buffer

	edit entryedit.Editor

	// Range mode.
	From      float64
	To        float64
	Increment float64
	Format    string
	Wrap      bool

	// Values mode (overrides range if non-empty).
	Values []string

	// Current value index for values mode.
	valuesIndex int

	// Callback.
	Command func(value string)

	// Layout.
	layoutX     int
	layoutY     int
	inset       int
	avgWidth    int
	buttonWidth int
	PrefWidth   int

	// Colors.
	SelBg *color.ColorRef
	// ButtonBg is -buttonbackground (default DEF_BUTTON_BG_COLOR).
	ButtonBg *color.ColorRef
	SelFg    *color.ColorRef
	InsertBg *color.ColorRef

	// Interaction state.
	hasFocus bool
	cursorOn bool

	// blink blinks the insert cursor while the widget has the focus.
	blink         *widget.CursorBlinker
	pressedButton string // "up", "down", or ""

	// Auto-repeat of the spinner buttons and the auto-scan of the text
	// (tk/library/spinbox.tcl).
	RepeatDelay    int // ms before the repeat starts
	RepeatInterval int // ms between repeats
	repeatCancel   func() bool
	scanCancel     func() bool
	lastX          int // last pointer position, for the auto-scan direction

	// Validation.
	Validate    string
	ValidateCmd func(string) bool

	// State is the interaction state (Tk's -state): StateNormal,
	// StateDisabled or StateReadonly.
	State widget.State

	// Scrollbar callback.
	ScrollCmd func(first, last float64)
}

// SpinboxOption configures a Spinbox.
type SpinboxOption func(*Spinbox)

func FromOpt(v float64) SpinboxOption          { return func(s *Spinbox) { s.From = v } }
func ToOpt(v float64) SpinboxOption            { return func(s *Spinbox) { s.To = v } }
func IncrementOpt(v float64) SpinboxOption     { return func(s *Spinbox) { s.Increment = v } }
func FormatOpt(f string) SpinboxOption         { return func(s *Spinbox) { s.Format = f } }
func WrapOpt(b bool) SpinboxOption             { return func(s *Spinbox) { s.Wrap = b } }
func ValuesOpt(v []string) SpinboxOption       { return func(s *Spinbox) { s.Values = v } }
func CommandOpt(fn func(string)) SpinboxOption { return func(s *Spinbox) { s.Command = fn } }
func WidthOpt(w int) SpinboxOption             { return func(s *Spinbox) { s.PrefWidth = w } }
func ValidateOpt(v string) SpinboxOption       { return func(s *Spinbox) { s.Validate = v } }
func ValidateCmdOpt(fn func(string) bool) SpinboxOption {
	return func(s *Spinbox) { s.ValidateCmd = fn }
}

// RepeatDelayOpt sets the delay in milliseconds before the spinner buttons
// start to repeat while held (Tk's -repeatdelay).
func RepeatDelayOpt(ms int) SpinboxOption {
	return func(s *Spinbox) { s.RepeatDelay = ms }
}

// RepeatIntervalOpt sets the interval in milliseconds between the repeats
// of a held spinner button (Tk's -repeatinterval).
func RepeatIntervalOpt(ms int) SpinboxOption {
	return func(s *Spinbox) { s.RepeatInterval = ms }
}

// ButtonBackground sets -buttonbackground.
func ButtonBackground[C color.Spec](name C) SpinboxOption {
	return func(s *Spinbox) {
		if c, err := s.App.ColorCache().Resolve(name); err == nil {
			s.ButtonBg = c.Ref()
		} else {
			s.OptionFailed(err)
		}
	}
}

// Background sets the background colour.
func Background[C color.Spec](name C) SpinboxOption {
	return func(s *Spinbox) { s.SetBackgroundColor(name) }
}

// SelectBackground sets the background colour of the selected text
// (Tk's -selectbackground).
func SelectBackground[C color.Spec](name C) SpinboxOption {
	return func(s *Spinbox) {
		col, err := s.App.ColorCache().Resolve(name)
		if err != nil {
			s.OptionFailed(err)
			return
		}
		s.SelBg = col.Ref()
	}
}

// SelectForeground sets the colour of the selected text
// (Tk's -selectforeground).
func SelectForeground[C color.Spec](name C) SpinboxOption {
	return func(s *Spinbox) {
		col, err := s.App.ColorCache().Resolve(name)
		if err != nil {
			s.OptionFailed(err)
			return
		}
		s.SelFg = col.Ref()
	}
}

// InsertBackground sets the colour of the insertion cursor
// (Tk's -insertbackground).
func InsertBackground[C color.Spec](name C) SpinboxOption {
	return func(s *Spinbox) {
		col, err := s.App.ColorCache().Resolve(name)
		if err != nil {
			s.OptionFailed(err)
			return
		}
		s.InsertBg = col.Ref()
	}
}

// State sets the interaction state (Tk's -state): widget.StateNormal,
// widget.StateDisabled or widget.StateReadonly. A disabled spinbox does not
// respond to the mouse and cannot be edited, a read-only one can be
// selected and copied but not edited.
func State(st widget.State) SpinboxOption {
	return func(s *Spinbox) { s.State = st }
}

// ReadOnly makes the spinbox read-only (Tk's -readonly): its text can be
// selected and copied, but not edited.
func ReadOnly(on bool) SpinboxOption {
	return func(s *Spinbox) {
		if on {
			s.State = widget.StateReadonly
		} else {
			s.State = widget.StateNormal
		}
	}
}

// New creates a new Spinbox widget.
func New(parent widget.Caregiver, name string, opts ...SpinboxOption) *Spinbox {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(w)

	s := &Spinbox{
		PrefWidth:   10,
		cursorOn:    true,
		From:        0,
		To:          100,
		Increment:   1,
		buttonWidth: 16,
		// DEF_SPINBOX_REPEAT_DELAY and DEF_SPINBOX_REPEAT_INTERVAL
		// (tk/generic/tkEntry.c).
		RepeatDelay:    400,
		RepeatInterval: 100,
	}
	s.Init()
	widget.InitBase(&s.Base, w, app)
	s.SetDisplayProc(s.display)
	s.blink = widget.NewCursorBlinker(app, s.toggleCursor)
	w.Class = "Spinbox"
	s.edit = entryedit.Editor{
		Buf: &s.Buffer, App: app, Win: w,
		TryEdit: s.tryEdit, Insert: s.Insert, Delete: func(index, count int) { s.Delete(index, index+count) },
		Moved: func() { s.seeInsert(); s.Display() },
	}

	s.BorderWidth = 1 // DEF_ENTRY_BORDER_WIDTH
	s.Relief = option.ReliefSunken
	s.HighlightWidth = 1

	// Default colors.
	if bg, err := app.ColorCache().Get(widget.PaletteFor(app).FieldBackground); err == nil {
		s.Background = bg
		s.UpdateBorder()
	}
	if sel, err := app.ColorCache().Get(widget.PaletteFor(app).SelectBackground); err == nil {
		s.SelBg = sel.Ref()
	}
	if selfg, err := app.ColorCache().Get(widget.PaletteFor(app).SelectForeground); err == nil {
		s.SelFg = selfg.Ref()
	}
	if ins, err := app.ColorCache().Get(widget.PaletteFor(app).InsertBackground); err == nil {
		s.InsertBg = ins.Ref()
	}

	for _, opt := range opts {
		opt(s)
	}

	// Set initial value.
	if len(s.Values) > 0 {
		s.Text = []rune(s.Values[0])
	} else {
		s.Text = []rune(s.formatValue(s.From))
	}

	s.computeGeometry()

	if s.Background != nil {
		w.SetBackgroundPixel(s.Background.Pixel)
	}

	w.Flags |= window.FlagFocusable
	w.SetCursor(cursor.XTerm)
	bindSpinbox(s, app)

	return s
}

// GetText returns the spinbox text.
func (s *Spinbox) GetText() string {
	return s.Get()
}

// SetText sets the spinbox text.
func (s *Spinbox) SetText(text string) {
	s.Set(text)
	s.computeGeometry()
	s.Display()
}

// SpinUp increments the value.
func (s *Spinbox) SpinUp() {
	if len(s.Values) > 0 {
		s.syncValuesIndex()
		s.valuesIndex++
		if s.valuesIndex >= len(s.Values) {
			if s.Wrap {
				s.valuesIndex = 0
			} else {
				s.valuesIndex = len(s.Values) - 1
			}
		}
		s.SetText(s.Values[s.valuesIndex])
	} else {
		val := s.currentNumericValue()
		val += s.Increment
		lo, hi := s.From, s.To
		if lo > hi {
			lo, hi = hi, lo
		}
		if val > hi {
			if s.Wrap {
				val = lo
			} else {
				val = hi
			}
		}
		s.SetText(s.formatValue(val))
	}
	s.fireCommand()
}

// SpinDown decrements the value.
func (s *Spinbox) SpinDown() {
	if len(s.Values) > 0 {
		s.syncValuesIndex()
		s.valuesIndex--
		if s.valuesIndex < 0 {
			if s.Wrap {
				s.valuesIndex = len(s.Values) - 1
			} else {
				s.valuesIndex = 0
			}
		}
		s.SetText(s.Values[s.valuesIndex])
	} else {
		val := s.currentNumericValue()
		val -= s.Increment
		lo, hi := s.From, s.To
		if lo > hi {
			lo, hi = hi, lo
		}
		if val < lo {
			if s.Wrap {
				val = hi
			} else {
				val = lo
			}
		}
		s.SetText(s.formatValue(val))
	}
	s.fireCommand()
}

func (s *Spinbox) currentNumericValue() float64 {
	val, err := strconv.ParseFloat(string(s.Text), 64)
	if err != nil {
		return s.From
	}
	return val
}

// syncValuesIndex finds the text in -values if it was changed since the
// last spin, as SpinboxInvoke does; an unknown text keeps the index.
func (s *Spinbox) syncValuesIndex() {
	if s.valuesIndex >= 0 && s.valuesIndex < len(s.Values) && s.Values[s.valuesIndex] == string(s.Text) {
		return
	}
	if i := slices.Index(s.Values, string(s.Text)); i >= 0 {
		s.valuesIndex = i
	}
}

func (s *Spinbox) formatValue(v float64) string {
	if s.Format != "" {
		return fmt.Sprintf(s.Format, v)
	}
	return fmt.Sprintf(s.digitFormat(), v)
}

// digitFormat ports ComputeFormat: enough digits for the -from/-to range at
// the -increment's precision, in %f or, when shorter, %e notation.
func (s *Spinbox) digitFormat() string {
	maxValue := max(math.Abs(s.From), math.Abs(s.To))
	if maxValue == 0 {
		maxValue = 1
	}
	mostSig := int(math.Floor(math.Log10(maxValue)))
	leastSig := 0
	if math.Abs(s.Increment) > math.SmallestNonzeroFloat64 {
		leastSig = int(math.Floor(math.Log10(s.Increment)))
	}
	numDigits := max(mostSig-leastSig+1, 1)

	eDigits := numDigits + 4
	if numDigits > 1 {
		eDigits++
	}
	afterDecimal := max(numDigits-mostSig-1, 0)
	fDigits := afterDecimal
	if mostSig >= 0 {
		fDigits = mostSig + afterDecimal
	}
	if afterDecimal > 0 {
		fDigits++
	}
	if mostSig < 0 {
		fDigits++
	}
	if fDigits <= eDigits {
		return fmt.Sprintf("%%.%df", afterDecimal)
	}
	return fmt.Sprintf("%%.%de", numDigits-1)
}

func (s *Spinbox) fireCommand() {
	if s.Command != nil {
		s.Command(string(s.Text))
	}
}

// tryEdit checks whether a proposed edit is valid. Returns true to allow.
// A spinbox that is disabled or read-only cannot be edited.
func (s *Spinbox) tryEdit(prospective string) bool {
	if s.State != widget.StateNormal {
		return false
	}
	if s.ValidateCmd == nil {
		return true
	}
	v := s.Validate
	if v != "key" && v != "all" {
		return true
	}
	return s.ValidateCmd(prospective)
}

// Insert inserts text before character index (Tk: insert INDEX STRING).
// It does nothing when the spinbox is not in the normal state.
func (s *Spinbox) Insert(index int, text string) {
	if s.State != widget.StateNormal {
		return
	}
	if !s.Buffer.Insert(index, []rune(text)) {
		return
	}
	s.computeGeometry()
	s.seeInsert()
	s.Display()
}

// Delete removes the characters from index first to last, exclusive
// (Tk: delete FIRST LAST). It does nothing when the spinbox is not in the
// normal state.
func (s *Spinbox) Delete(first, last int) {
	if s.State != widget.StateNormal {
		return
	}
	if !s.Buffer.Delete(first, last-first) {
		return
	}
	s.computeGeometry()
	s.seeInsert()
	s.Display()
}

// DeleteSelection deletes the selected text.
func (s *Spinbox) DeleteSelection() {
	if s.HasSelection() {
		s.Delete(s.SelFirst, s.SelLast)
	}
}

// XView scrolls the spinbox text so that the character at index is the
// leftmost visible one.
func (s *Spinbox) XView(index int) {
	if index < 0 {
		index = 0
	}
	if index > len(s.Text) {
		index = len(s.Text)
	}
	s.LeftIndex = index
	s.computeGeometry()
	s.Display()
}

// XViewScroll scrolls the text by count units (characters) or pages.
func (s *Spinbox) XViewScroll(count int, pages bool) {
	if pages {
		availWidth := s.Win.Width - 2*s.inset - s.buttonWidth
		charsPerPage := max(availWidth/s.avgWidth-2, 1)
		count *= charsPerPage
	}
	s.XView(s.LeftIndex + count)
}

// XViewMoveTo scrolls to a fraction of the text.
func (s *Spinbox) XViewMoveTo(fraction float64) {
	index := int(fraction*float64(len(s.Text)) + 0.5)
	s.XView(index)
}

// spin repeats the action of the pressed spinner button: ::tk::spinbox::Invoke
// (tk/library/spinbox.tcl:301) re-invokes the element every -repeatinterval
// ms once -repeatdelay ms have passed.
func (s *Spinbox) spin() {
	if s.State == widget.StateDisabled {
		return
	}
	switch s.pressedButton {
	case "up":
		s.SpinUp()
	case "down":
		s.SpinDown()
	default:
		return
	}
	s.Display()
}

// startRepeat schedules the next repeat of the pressed spinner button.
func (s *Spinbox) startRepeat() {
	s.stopRepeat()
	if s.App == nil || s.RepeatDelay <= 0 || s.pressedButton == "" {
		return
	}
	s.repeatCancel = s.App.After(time.Duration(s.RepeatDelay)*time.Millisecond, func() {
		s.repeatCancel = nil
		s.repeatStep()
	})
}

// repeatStep repeats the action and schedules the next repeat.
func (s *Spinbox) repeatStep() {
	s.repeatCancel = nil
	if s.Destroyed() {
		return
	}
	s.spin()
	if s.RepeatInterval <= 0 {
		return
	}
	s.repeatCancel = s.App.After(time.Duration(s.RepeatInterval)*time.Millisecond, func() {
		s.repeatStep()
	})
}

// stopRepeat cancels a pending repeat.
func (s *Spinbox) stopRepeat() {
	if s.repeatCancel != nil {
		s.repeatCancel()
		s.repeatCancel = nil
	}
}

// startAutoScan scrolls the text while the pointer is held outside the
// spinbox: ::tk::spinbox::AutoScan (tk/library/spinbox.tcl:565).
func (s *Spinbox) startAutoScan(dir int) {
	s.stopAutoScan()
	if s.App == nil || dir == 0 {
		return
	}
	s.autoScanStep(dir)
}

// autoScanStep scrolls the text two characters and schedules the next step.
func (s *Spinbox) autoScanStep(dir int) {
	s.scanCancel = nil
	if s.Destroyed() {
		return
	}
	s.XViewScroll(2*dir, false)
	s.ExtendTo(s.closestGap(s.lastX))
	s.Display()
	s.scanCancel = s.App.After(50*time.Millisecond, func() {
		s.autoScanStep(dir)
	})
}

// stopAutoScan cancels a pending auto-scan.
func (s *Spinbox) stopAutoScan() {
	if s.scanCancel != nil {
		s.scanCancel()
		s.scanCancel = nil
	}
}

func (s *Spinbox) computeGeometry() {
	if s.Font == nil {
		return
	}
	m := s.Font.Metrics()
	s.inset = s.BorderWidth + s.HighlightWidth + 1
	s.avgWidth = max(s.Font.MeasureString("0"), 1)

	// EntryWorldChanged: the button column is one "0" plus 2*(1+XPAD).
	s.buttonWidth = max(11, s.avgWidth+4)
	w := s.Win
	w.ReqWidth = s.PrefWidth*s.avgWidth + 2*s.inset + s.buttonWidth
	w.ReqHeight = m.Linespace() + 2*s.inset

	s.layoutY = s.inset + m.Ascent

	totalWidth := textedit.MeasureRunes(s.Font, s.Text)
	availWidth := max(w.Width-2*s.inset-s.buttonWidth, 1)

	if totalWidth <= availWidth {
		s.LeftIndex = 0
		s.layoutX = s.inset
	} else {
		s.LeftIndex = max(0, min(s.LeftIndex, len(s.Text)))
		leftCharX := textedit.MeasureRunes(s.Font, s.Text[:s.LeftIndex])
		s.layoutX = s.inset - leftCharX
	}
}

func (s *Spinbox) seeInsert() {
	if s.Font == nil {
		return
	}
	availWidth := s.Win.Width - 2*s.inset - s.buttonWidth
	if availWidth < 1 {
		return
	}
	if s.InsertPos < s.LeftIndex {
		s.LeftIndex = s.InsertPos
		s.computeGeometry()
	} else {
		cursorX := textedit.MeasureRunes(s.Font, s.Text[:s.InsertPos]) + s.layoutX
		if cursorX >= s.Win.Width-s.inset-s.buttonWidth {
			s.LeftIndex = max(s.InsertPos-availWidth/s.avgWidth, 0)
			s.computeGeometry()
		}
	}
}

func (s *Spinbox) closestGap(x int) int {
	if s.Font == nil || len(s.Text) == 0 {
		return 0
	}
	xInLayout := x - s.layoutX
	idx := textedit.RuneIndexAtPixel(s.Font, s.Text, xInLayout)
	if idx >= len(s.Text) {
		return len(s.Text)
	}
	charStart := textedit.MeasureRunes(s.Font, s.Text[:idx])
	charEnd := textedit.MeasureRunes(s.Font, s.Text[:idx+1])
	mid := (charStart + charEnd) / 2
	if xInLayout >= mid {
		return idx + 1
	}
	return idx
}

// buttonBox returns the button column left edge, top and half height as
// drawn by DisplayEntry.
func (s *Spinbox) buttonBox() (startx, top, height int) {
	in := s.inset - 1
	return s.Win.Width - (s.buttonWidth + in), in, (s.Win.Height - 2*in) / 2
}

// hitButton returns "up", "down", or "" for the button area.
func (s *Spinbox) hitButton(x, y int) string {
	startx, top, h := s.buttonBox()
	if x < startx || x >= startx+s.buttonWidth || y < top {
		return ""
	}
	if y < top+h {
		return "up"
	}
	if y < top+2*h {
		return "down"
	}
	return ""
}

// Display schedules a redraw at idle time; see widget.Base.EventuallyRedraw.
func (s *Spinbox) Display() {
	s.EventuallyRedraw()
}

// display draws the spinbox.
func (s *Spinbox) display() {
	if s.Destroyed() {
		return
	}
	w := s.Win
	if w.PlatformID == platform.WindowID(0) {
		return
	}

	d := w.Display.Server
	gc := w.GC

	// Background.
	if s.Background != nil {
		d.SetForeground(gc, s.Background.Pixel)
	}
	d.FillRectangle(w.Drawable(), gc, 0, 0, uint(w.Width), uint(w.Height))

	// Draw text.
	xftFont, isXft := s.Font.(platform.DrawableFont)
	// A disabled spinbox does not show its selection, as Tk's
	// DisplayEntry does (tk/generic/tkEntry.c).
	showSelection := s.State != widget.StateDisabled
	if isXft && len(s.Text) > 0 {
		// Selection highlight.
		if s.hasFocus && showSelection && s.SelFirst >= 0 && s.SelLast > s.SelFirst && s.SelBg != nil {
			selStartX := textedit.MeasureRunes(s.Font, s.Text[:textedit.ClampIdx(s.SelFirst, len(s.Text))]) + s.layoutX
			selEndX := textedit.MeasureRunes(s.Font, s.Text[:textedit.ClampIdx(s.SelLast, len(s.Text))]) + s.layoutX
			if selStartX < s.inset {
				selStartX = s.inset
			}
			rightEdge := w.Width - s.inset - s.buttonWidth
			if selEndX > rightEdge {
				selEndX = rightEdge
			}
			if selEndX > selStartX {
				m := s.Font.Metrics()
				d.SetForeground(gc, s.SelBg.Pixel)
				d.FillRectangle(w.Drawable(), gc, selStartX, s.inset,
					uint(selEndX-selStartX), uint(m.Linespace()))
			}
		}

		// Draw all text.
		if s.Foreground != nil {
			xftFont.DrawString(w.Drawable(), s.layoutX, s.layoutY, string(s.Text),
				s.Foreground.Pixel, s.Foreground.Red, s.Foreground.Green, s.Foreground.Blue)
		}
	}

	// Cursor. A disabled or read-only spinbox has no cursor, as Tk's
	// DisplayEntry and EntryBlinkProc do (tk/generic/tkEntry.c).
	if s.hasFocus && s.cursorOn && s.State == widget.StateNormal && s.InsertBg != nil && s.Font != nil {
		cursorX := textedit.MeasureRunes(s.Font, s.Text[:textedit.ClampIdx(s.InsertPos, len(s.Text))]) + s.layoutX
		rightEdge := w.Width - s.inset - s.buttonWidth
		if cursorX >= s.inset && cursorX < rightEdge {
			m := s.Font.Metrics()
			d.SetForeground(gc, s.InsertBg.Pixel)
			d.FillRectangle(w.Drawable(), gc, cursorX-1, s.inset, 2, uint(m.Linespace()))
		}
	}

	s.drawButtons(d, gc)

	// DisplayEntry draws the border and focus highlight last.
	if s.Border != nil && s.BorderWidth > 0 {
		hl := s.HighlightWidth
		draw.Draw3DRectangle(d, w.Drawable(), gc, s.Border,
			hl, hl, w.Width-2*hl, w.Height-2*hl, s.BorderWidth, s.Relief)
	}
	s.DrawHighlightBorder(s.hasFocus, 0)

}

// drawButtons ports the spin button drawing in DisplayEntry (tkEntry.c).
func (s *Spinbox) drawButtons(d platform.DisplayServer, gc platform.GCID) {
	w := s.Win
	startx, in, height := s.buttonBox()
	bb := draw.NewBorderFromPixel(0xd9d9d9)
	if c, err := s.App.ColorCache().Get(widget.PaletteFor(s.App).Background); err == nil {
		bb = draw.NewBorderFromPixel(c.Pixel)
	}
	if s.ButtonBg != nil {
		bb = draw.NewBorderFromPixel(s.ButtonBg.Pixel)
	}
	upRelief, downRelief := option.ReliefRaised, option.ReliefRaised
	if s.pressedButton == "up" {
		upRelief = option.ReliefSunken
	}
	if s.pressedButton == "down" {
		downRelief = option.ReliefSunken
	}
	draw.Fill3DRectangle(d, w.Drawable(), gc, bb, startx, in, s.buttonWidth, height, 1, upRelief)
	draw.Fill3DRectangle(d, w.Drawable(), gc, bb, startx, in+height, s.buttonWidth, height, 1, downRelief)
	if s.Foreground == nil {
		return
	}
	const pad = 2 // XPAD + 1
	xw := s.buttonWidth - 2*pad
	if xw <= 1 {
		return
	}
	space := height - 2*pad
	if xw%2 == 0 {
		xw++
	}
	th := min((xw+1)/2, space)
	space = (space - th) / 2
	startx += pad
	d.SetForeground(gc, s.Foreground.Pixel)
	b2i := func(b bool) int {
		if b {
			return 1
		}
		return 0
	}
	off := b2i(s.pressedButton == "up")
	starty := in + height - pad - space
	y0 := starty - 1 + off
	d.FillPolygon(w.Drawable(), gc, []platform.Point{
		{X: int32(startx + off), Y: int32(y0)},
		{X: int32(startx + xw/2 + off), Y: int32(starty - th - 1 + off)},
		{X: int32(startx + xw + off), Y: int32(y0)},
	}, 2, 0)
	off = b2i(s.pressedButton == "down")
	starty = in + height + pad + space
	y0 = starty + off
	d.FillPolygon(w.Drawable(), gc, []platform.Point{
		{X: int32(startx + 1 + off), Y: int32(y0)},
		{X: int32(startx + xw/2 + off), Y: int32(starty + th - 1 + off)},
		{X: int32(startx - 1 + xw + off), Y: int32(y0)},
	}, 2, 0)
}

// Configure applies options.
func (s *Spinbox) Configure(opts ...SpinboxOption) error {
	return widget.Configure(s, opts, s.computeGeometry)
}

// Destroy cleans up the spinbox.
func (s *Spinbox) Destroy() {
	if s.Destroyed() {
		return
	}
	s.stopRepeat()
	s.stopAutoScan()
	s.blink.Stop()
	s.MarkDestroyed()
	window.DestroyWindow(s.Win)
}

// toggleCursor blinks the insert cursor, as Tk's entry does.
func (s *Spinbox) toggleCursor() {
	s.cursorOn = !s.cursorOn
	s.Display()
}
