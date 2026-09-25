// Package spinbox implements a spinbox widget — an entry with up/down
// spinner buttons for numeric or list-based value selection.
package spinbox

import (
	"fmt"
	"log"
	"math"
	"strconv"

	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/cursor"
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/entryutil"
	"github.com/msorc/takigo/window"
)

// Spinbox is a single-line text entry with up/down spinner buttons.
type Spinbox struct {
	widget.Base

	// Text state (reuses entry patterns).
	text      []rune
	InsertPos int
	SelFirst  int
	SelLast   int
	SelAnchor int
	LeftIndex int

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
	SelBg    *color.ColorRef
	SelFg    *color.ColorRef
	InsertBg *color.ColorRef

	// Interaction state.
	HasFocus      bool
	CursorOn      bool
	pressedButton string // "up", "down", or ""

	// Validation.
	Validate    string
	ValidateCmd func(string) bool

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

func Background(name string) SpinboxOption {
	return func(s *Spinbox) {
		col, err := s.App.ColorCache().Get(name)
		if err != nil {
			log.Printf("spinbox: failed to get color %q: %v", name, err)
			return
		}
		s.Background = col
		s.UpdateBorder()
	}
}

// --- Ttk-compatible aliases (prefix with Spinbox) for consistent naming ---
// These aliases match the naming convention used by ttk widgets
// allowing consistent option naming when both classic and ttk widgets are used.

// SpinboxFromOpt is an alias for FromOpt.
var SpinboxFromOpt = FromOpt

// SpinboxToOpt is an alias for ToOpt.
var SpinboxToOpt = ToOpt

// SpinboxIncrementOpt is an alias for IncrementOpt.
var SpinboxIncrementOpt = IncrementOpt

// SpinboxFormatOpt is an alias for FormatOpt.
var SpinboxFormatOpt = FormatOpt

// SpinboxWrapOpt is an alias for WrapOpt.
var SpinboxWrapOpt = WrapOpt

// SpinboxValuesOpt is an alias for ValuesOpt.
var SpinboxValuesOpt = ValuesOpt

// SpinboxCommandOpt is an alias for CommandOpt.
var SpinboxCommandOpt = CommandOpt

// SpinboxWidthOpt is an alias for WidthOpt.
var SpinboxWidthOpt = WidthOpt

// SpinboxValidateOpt is an alias for ValidateOpt.
var SpinboxValidateOpt = ValidateOpt

// SpinboxValidateCmdOpt is an alias for ValidateCmdOpt.
var SpinboxValidateCmdOpt = ValidateCmdOpt

// SpinboxBackground is an alias for Background.
var SpinboxBackground = Background

// New creates a new Spinbox widget.
func New(parent widget.Caregiver, name string, opts ...SpinboxOption) *Spinbox {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(w)

	s := &Spinbox{
		SelFirst:    -1,
		SelLast:     -1,
		PrefWidth:   10,
		CursorOn:    true,
		From:        0,
		To:          100,
		Increment:   1,
		Format:      "%.0f",
		buttonWidth: 16,
	}
	widget.InitBase(&s.Base, w, app)
	w.Class = "Spinbox"

	s.BorderWidth = 2
	s.Relief = option.ReliefSunken
	s.HighlightWidth = 1

	// Default colors.
	if bg, err := app.ColorCache().Get("#ffffff"); err == nil {
		s.Background = bg
		s.UpdateBorder()
	}
	if sel, err := app.ColorCache().Get("#3399ff"); err == nil {
		s.SelBg = sel.Ref()
	}
	if selfg, err := app.ColorCache().Get("#ffffff"); err == nil {
		s.SelFg = selfg.Ref()
	}
	if ins, err := app.ColorCache().Get("#000000"); err == nil {
		s.InsertBg = ins.Ref()
	}

	for _, opt := range opts {
		opt(s)
	}

	// Set initial value.
	if len(s.Values) > 0 {
		s.text = []rune(s.Values[0])
	} else {
		s.text = []rune(s.formatValue(s.From))
	}

	s.computeGeometry()

	if s.Background != nil {
		w.BackgroundPixel = s.Background.Pixel
	}

	w.Flags |= window.FlagFocusable
	w.SetCursor(uint(cursor.XTerm))
	bindSpinbox(s, app)

	return s
}

// GetText returns the spinbox text.
func (s *Spinbox) GetText() string {
	return string(s.text)
}

// SetText sets the spinbox text.
func (s *Spinbox) SetText(text string) {
	s.text = []rune(text)
	if s.InsertPos > len(s.text) {
		s.InsertPos = len(s.text)
	}
	s.ClearSelection()
	s.computeGeometry()
	s.Display()
}

// SpinUp increments the value.
func (s *Spinbox) SpinUp() {
	if len(s.Values) > 0 {
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
	val, err := strconv.ParseFloat(string(s.text), 64)
	if err != nil {
		return s.From
	}
	return val
}

func (s *Spinbox) formatValue(v float64) string {
	if s.Increment >= 1 {
		return fmt.Sprintf("%.0f", math.Round(v))
	}
	return fmt.Sprintf(s.Format, v)
}

func (s *Spinbox) fireCommand() {
	if s.Command != nil {
		s.Command(string(s.text))
	}
}

// tryEdit checks whether a proposed edit is valid. Returns true to allow.
func (s *Spinbox) tryEdit(prospective string) bool {
	if s.ValidateCmd == nil {
		return true
	}
	v := s.Validate
	if v != "key" && v != "all" {
		return true
	}
	return s.ValidateCmd(prospective)
}

// InsertChars inserts text at the given rune index.
func (s *Spinbox) InsertChars(index int, text string) {
	if len(text) == 0 {
		return
	}
	runes := []rune(text)
	count := len(runes)

	if index < 0 {
		index = 0
	}
	if index > len(s.text) {
		index = len(s.text)
	}

	newText := make([]rune, 0, len(s.text)+count)
	newText = append(newText, s.text[:index]...)
	newText = append(newText, runes...)
	newText = append(newText, s.text[index:]...)
	s.text = newText

	if s.InsertPos >= index {
		s.InsertPos += count
	}
	if s.SelFirst >= index {
		s.SelFirst += count
	}
	if s.SelLast > index {
		s.SelLast += count
	}
	if s.SelAnchor >= index {
		s.SelAnchor += count
	}
	if s.LeftIndex > index {
		s.LeftIndex += count
	}

	s.computeGeometry()
	s.seeInsert()
	s.Display()
}

// DeleteChars deletes count runes starting at index.
func (s *Spinbox) DeleteChars(index, count int) {
	if count <= 0 || len(s.text) == 0 {
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= len(s.text) {
		return
	}
	if index+count > len(s.text) {
		count = len(s.text) - index
	}

	s.text = append(s.text[:index], s.text[index+count:]...)

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
	adjustIndex(&s.InsertPos)
	adjustIndex(&s.SelFirst)
	adjustIndex(&s.SelLast)
	adjustIndex(&s.SelAnchor)
	adjustIndex(&s.LeftIndex)

	if s.SelFirst >= 0 && s.SelLast <= s.SelFirst {
		s.SelFirst = -1
		s.SelLast = -1
	}

	s.computeGeometry()
	s.seeInsert()
	s.Display()
}

// DeleteSelection deletes the selected text.
func (s *Spinbox) DeleteSelection() {
	if s.SelFirst < 0 {
		return
	}
	s.DeleteChars(s.SelFirst, s.SelLast-s.SelFirst)
}

// ClearSelection clears the selection.
func (s *Spinbox) ClearSelection() {
	s.SelFirst = -1
	s.SelLast = -1
}

// SelectAll selects all text.
func (s *Spinbox) SelectAll() {
	if len(s.text) > 0 {
		s.SelFirst = 0
		s.SelLast = len(s.text)
	}
}

func (s *Spinbox) computeGeometry() {
	if s.Font == nil {
		return
	}
	m := s.Font.Metrics()
	s.inset = s.BorderWidth + s.HighlightWidth + 1
	s.avgWidth = s.Font.MeasureString("0")
	if s.avgWidth < 1 {
		s.avgWidth = 1
	}

	w := s.Win
	w.ReqWidth = s.PrefWidth*s.avgWidth + 2*s.inset + s.buttonWidth
	w.ReqHeight = m.Linespace() + 2*s.inset

	s.layoutY = s.inset + m.Ascent

	totalWidth := entryutil.MeasureRunes(s.Font, s.text)
	availWidth := w.Width - 2*s.inset - s.buttonWidth
	if availWidth < 1 {
		availWidth = 1
	}

	if totalWidth <= availWidth {
		s.LeftIndex = 0
		s.layoutX = s.inset
	} else {
		if s.LeftIndex < 0 {
			s.LeftIndex = 0
		}
		leftCharX := entryutil.MeasureRunes(s.Font, s.text[:s.LeftIndex])
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
		cursorX := entryutil.MeasureRunes(s.Font, s.text[:s.InsertPos]) + s.layoutX
		if cursorX >= s.Win.Width-s.inset-s.buttonWidth {
			s.LeftIndex = s.InsertPos - availWidth/s.avgWidth
			if s.LeftIndex < 0 {
				s.LeftIndex = 0
			}
			s.computeGeometry()
		}
	}
}

func (s *Spinbox) closestGap(x int) int {
	if s.Font == nil || len(s.text) == 0 {
		return 0
	}
	xInLayout := x - s.layoutX
	idx := entryutil.RuneIndexAtPixel(s.Font, s.text, xInLayout)
	if idx >= len(s.text) {
		return len(s.text)
	}
	charStart := entryutil.MeasureRunes(s.Font, s.text[:idx])
	charEnd := entryutil.MeasureRunes(s.Font, s.text[:idx+1])
	mid := (charStart + charEnd) / 2
	if xInLayout >= mid {
		return idx + 1
	}
	return idx
}

// hitButton returns "up", "down", or "" for the button area.
func (s *Spinbox) hitButton(x, y int) string {
	btnLeft := s.Win.Width - s.buttonWidth
	if x < btnLeft {
		return ""
	}
	midY := s.Win.Height / 2
	if y < midY {
		return "up"
	}
	return "down"
}

// Display draws the spinbox.
func (s *Spinbox) Display() {
	if s.Destroyed {
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

	// Border.
	if s.Border != nil && s.BorderWidth > 0 {
		hl := s.HighlightWidth
		draw.Draw3DRectangle(d, w.Drawable(), gc, s.Border,
			hl, hl, w.Width-2*hl, w.Height-2*hl, s.BorderWidth, s.Relief)
	}
	// The highlight ring sits outside the border (focus colour or
	// -highlightbackground).
	s.DrawHighlightBorder(s.HasFocus, 0)

	// Draw text.
	xftFont, isXft := s.Font.(platform.DrawableFont)
	if isXft && len(s.text) > 0 {
		// Selection highlight.
		if s.HasFocus && s.SelFirst >= 0 && s.SelLast > s.SelFirst && s.SelBg != nil {
			selStartX := entryutil.MeasureRunes(s.Font, s.text[:entryutil.ClampIdx(s.SelFirst, len(s.text))]) + s.layoutX
			selEndX := entryutil.MeasureRunes(s.Font, s.text[:entryutil.ClampIdx(s.SelLast, len(s.text))]) + s.layoutX
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
			xftFont.DrawString(w.Drawable(), s.layoutX, s.layoutY, string(s.text),
				s.Foreground.Pixel, s.Foreground.Red, s.Foreground.Green, s.Foreground.Blue)
		}
	}

	// Cursor.
	if s.HasFocus && s.CursorOn && s.InsertBg != nil && s.Font != nil {
		cursorX := entryutil.MeasureRunes(s.Font, s.text[:entryutil.ClampIdx(s.InsertPos, len(s.text))]) + s.layoutX
		rightEdge := w.Width - s.inset - s.buttonWidth
		if cursorX >= s.inset && cursorX < rightEdge {
			m := s.Font.Metrics()
			d.SetForeground(gc, s.InsertBg.Pixel)
			d.FillRectangle(w.Drawable(), gc, cursorX-1, s.inset, 2, uint(m.Linespace()))
		}
	}

	// Draw buttons.
	s.drawButtons(d, gc)

	d.Flush()
}

func (s *Spinbox) drawButtons(d platform.DisplayServer, gc platform.GCID) {
	w := s.Win
	btnLeft := w.Width - s.buttonWidth
	midY := w.Height / 2

	// Button background.
	if s.Background != nil {
		d.SetForeground(gc, s.Background.Pixel)
		d.FillRectangle(w.Drawable(), gc, btnLeft, 0, uint(s.buttonWidth), uint(w.Height))
	}

	// Button borders.
	if s.Border != nil {
		upRelief := option.ReliefRaised
		downRelief := option.ReliefRaised
		if s.pressedButton == "up" {
			upRelief = option.ReliefSunken
		}
		if s.pressedButton == "down" {
			downRelief = option.ReliefSunken
		}
		draw.Draw3DRectangle(d, w.Drawable(), gc, s.Border,
			btnLeft, 0, s.buttonWidth, midY, 1, upRelief)
		draw.Draw3DRectangle(d, w.Drawable(), gc, s.Border,
			btnLeft, midY, s.buttonWidth, w.Height-midY, 1, downRelief)
	}

	// Draw triangles.
	if s.Foreground != nil {
		d.SetForeground(gc, s.Foreground.Pixel)

		// Up triangle.
		cx := btnLeft + s.buttonWidth/2
		upCy := midY / 2
		sz := 3
		for row := 0; row < sz; row++ {
			x1 := cx - row
			x2 := cx + row
			y := upCy - sz/2 + row
			d.DrawLine(w.Drawable(), gc, x1, y, x2, y)
		}

		// Down triangle (inverted).
		downCy := midY + (w.Height-midY)/2
		for row := 0; row < sz; row++ {
			x1 := cx - row
			x2 := cx + row
			y := downCy + sz/2 - row
			d.DrawLine(w.Drawable(), gc, x1, y, x2, y)
		}
	}
}

// Configure applies options.
func (s *Spinbox) Configure(opts ...option.Option) {
	option.Apply(s, opts)
	s.UpdateBorder()
	s.computeGeometry()
	if s.Background != nil {
		s.Win.BackgroundPixel = s.Background.Pixel
	}
	s.Display()
}

// Destroy cleans up the spinbox.
func (s *Spinbox) Destroy() {
	if s.Destroyed {
		return
	}
	s.Destroyed = true
	window.DestroyWindow(s.Win)
}
