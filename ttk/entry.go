package ttk

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/takigo/takigo/cursor"
	"github.com/takigo/takigo/draw"
	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/internal/textedit"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/ttk/internal/entrytext"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/window"
)

// FieldState is the -state of an entry, combobox or spinbox: whether
// its text can be edited, only chosen (readonly) or neither.
type FieldState int

const (
	FieldNormal FieldState = iota
	FieldDisabled
	FieldReadonly
)

// ValidateMode mirrors tk/generic/ttk/ttkEntry.c VMODE.
type ValidateMode int

const (
	ValidateNone     ValidateMode = iota
	ValidateKey                   // on each edit
	ValidateFocus                 // on FocusOut
	ValidateFocusIn               // on FocusIn
	ValidateFocusOut              // on FocusOut (synonym of focus)
	ValidateAll                   // on each edit + FocusIn/Out
)

// Entry is a themed single-line text-entry widget.
// It ports tk/generic/ttk/ttkEntry.c and tk/library/ttk/entry.tcl.
type Entry struct {
	TtkWidget

	edit entrytext.Helper
	Font font.Font

	// Behaviour.
	Placeholder string
	Show        rune // 0 = show text; non-zero = show N copies of this char (password)
	Justify     option.Justify
	WidthChars  int
	StateMode   FieldState

	// Validation.
	ValidateMode ValidateMode
	ValidateCmd  func(string) bool
	InvalidCmd   func()
	ExportSelect bool

	// Text variable linkage.
	TextVar *widget.Variable[string]
	unsub   func()
	linked  *widget.Variable[string]

	// Layout metrics.
	insetX int
	insetY int

	// Horizontal scroll (Tk's Scrollable).
	leftIndex int // first visible rune

	// XScroll callback.
	XScrollCmd func(first, last float64)
}

// EntryOption configures an Entry.
type EntryOption func(*Entry)

// EntryText sets the initial text.
func EntryText(s string) EntryOption {
	return func(e *Entry) { e.edit.Text = []rune(s) }
}

// EntryPlaceholder sets placeholder text shown when the field is empty.
func EntryPlaceholder(s string) EntryOption {
	return func(e *Entry) { e.Placeholder = s }
}

// EntryShow sets the password-mask character (e.g. '*'). Pass 0 to disable.
func EntryShow(ch rune) EntryOption {
	return func(e *Entry) { e.Show = ch }
}

// EntryFont sets the font by name.
func EntryFont[F font.Spec](name F) EntryOption {
	return func(e *Entry) {
		f, err := e.App.FontRegistry().Resolve(name)
		if err != nil {
			e.OptionFailed(err)
			return
		}
		e.Font = f
		e.edit.Font = f
	}
}

// EntryTextVariable links the entry text to a string variable.
func EntryTextVariable(v *widget.Variable[string]) EntryOption {
	return func(e *Entry) { e.TextVar = v }
}

// EntryWidth sets the preferred width in characters (Tk: -width).
func EntryWidth(n int) EntryOption {
	return func(e *Entry) { e.WidthChars = n }
}

// EntryJustify sets text justification.
func EntryJustify(j option.Justify) EntryOption {
	return func(e *Entry) { e.Justify = j }
}

// FieldState sets -state.
func EntryState(s FieldState) EntryOption {
	return func(e *Entry) { e.StateMode = s }
}

// EntryValidate sets -validate: when -validatecommand runs.
func EntryValidate(v ValidateMode) EntryOption {
	return func(e *Entry) { e.ValidateMode = v }
}

// EntryValidateCmd sets the validation callback. Return false to reject edit.
func EntryValidateCmd(fn func(string) bool) EntryOption {
	return func(e *Entry) { e.ValidateCmd = fn }
}

// EntryInvalidCmd sets the callback fired when validation rejects a value.
func EntryInvalidCmd(fn func()) EntryOption {
	return func(e *Entry) { e.InvalidCmd = fn }
}

// EntryExportSelection toggles tying internal selection to X selection.
func EntryExportSelection(on bool) EntryOption {
	return func(e *Entry) { e.ExportSelect = on }
}

// EntryXScrollCommand sets the callback invoked with (first, last) when the
// horizontal scroll position changes.
func EntryXScrollCommand(fn func(first, last float64)) EntryOption {
	return func(e *Entry) { e.XScrollCmd = fn }
}

// EntryStyle sets -style.
func EntryStyle(name string) EntryOption {
	return func(e *Entry) { e.StyleName = name }
}

// NewEntry creates a themed entry widget.
func NewEntry(parent widget.Caregiver, name string, opts ...EntryOption) *Entry {
	app := parent.AppContext()
	win := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(win)
	win.Flags |= window.FlagFocusable // ttk::takefocus accepts it

	e := &Entry{
		Justify:      option.JustifyLeft,
		WidthChars:   20,
		StateMode:    FieldNormal,
		ValidateMode: ValidateNone,
		ExportSelect: true,
		insetX:       2,
		insetY:       2,
	}
	e.Font, _ = app.FontRegistry().Get(font.TkDefaultFont)

	e.edit = entrytext.Helper{
		Font:  e.Font,
		TextX: e.insetX + 1,
		App:   app,
		Win:   win,
		Redraw: func() {
			e.Display()
		},
		Editable: func() bool { return e.StateMode == FieldNormal },
		Validate: e.runValidate,
	}

	InitTtkWidget(&e.TtkWidget, win, app, "TEntry")
	e.reconfigure = func() { _ = e.Configure() }
	win.OnDestroy(e.Destroy)
	e.DisplayFunc = e.Display

	for _, opt := range opts {
		opt(e)
	}
	if e.StyleName != "TEntry" {
		e.RefreshTheme()
	}
	e.requestSize()
	e.syncState()
	win.SetCursor(cursor.XTerm)
	e.linkTextVar()

	bindEntry(e, app)

	return e
}

// requestSize requests -width average characters by one line, plus the
// field border (inset) and TEntry's -padding 1 on each side.
func (e *Entry) requestSize() {
	if e.Font == nil {
		return
	}
	m := e.Font.Metrics()
	avgW := e.Font.MeasureString("0")
	if avgW < 1 {
		avgW = 8
	}
	e.Win.ReqWidth = e.WidthChars*avgW + 2*e.insetX + 2
	e.Win.ReqHeight = m.Linespace() + 2*e.insetY + 2
}

// syncState applies -state to the widget state.
func (e *Entry) syncState() {
	switch e.StateMode {
	case FieldDisabled:
		e.ChangeState(StateDisabled, StateReadonly)
	case FieldReadonly:
		e.ChangeState(StateReadonly, StateDisabled)
	default:
		e.ChangeState(0, StateDisabled|StateReadonly)
	}
}

// linkTextVar takes the text from -textvariable and follows its changes.
func (e *Entry) linkTextVar() {
	if e.TextVar == e.linked {
		return
	}
	if e.unsub != nil {
		e.unsub()
		e.unsub = nil
	}
	e.linked = e.TextVar
	if e.TextVar == nil {
		return
	}
	e.edit.Text = []rune(e.TextVar.Get())
	e.edit.InsertPos = len(e.edit.Text)
	e.unsub = e.TextVar.OnChange(func(_, val string) {
		if val == e.edit.Get() {
			return
		}
		e.edit.Set(val)
	})
}

// Window returns the underlying window.
func (e *Entry) Window() *window.Window { return e.Win }

// AppContext satisfies widget.Caregiver.
func (e *Entry) AppContext() widget.AppContext { return e.App }

// GeometryElements makes the entry usable in geometry.Group.
func (e *Entry) GeometryElements() []window.Windower { return []window.Windower{e} }

// Destroy frees resources and unsubscribes from the text variable.
func (e *Entry) Destroy() {
	if e.unsub != nil {
		e.unsub()
		e.unsub = nil
	}
	e.TtkWidget.Destroy()
}

// ---- subcommands (Tk: $entry xxx) ----

// Get returns the current text.
func (e *Entry) Get() string { return e.edit.Get() }

// Set replaces the text.
func (e *Entry) Set(s string) {
	e.edit.Set(s)
	e.notifyTextVar()
}

// Insert inserts text at the given index (Tk: insert INDEX TEXT).
func (e *Entry) Insert(index, text string) error {
	pos, err := e.resolveIndex(index)
	if err != nil {
		return err
	}
	if !e.edit.InsertAt(pos, []rune(text)) {
		return fmt.Errorf("entry: validate rejected insert")
	}
	e.notifyTextVar()
	return nil
}

// Delete removes the [first, last) range (Tk: delete FIRST ?LAST?).
// If last is empty, deletes the single character at first.
func (e *Entry) Delete(first, last string) error {
	firstIdx, err := e.resolveIndex(first)
	if err != nil {
		return err
	}
	lastIdx := firstIdx + 1
	if last != "" {
		lastIdx, err = e.resolveIndex(last)
		if err != nil {
			return err
		}
	}
	if !e.edit.DeleteRange(firstIdx, lastIdx) {
		return fmt.Errorf("entry: validate rejected delete")
	}
	e.notifyTextVar()
	return nil
}

// Index resolves an index string to a rune position.
// Supports "insert", "end", "sel.first", "sel.last", "@N", and integers.
func (e *Entry) Index(idx string) (int, error) { return e.resolveIndex(idx) }

// ICursor moves the insertion cursor (Tk: icursor POS).
func (e *Entry) ICursor(pos string) error {
	p, err := e.resolveIndex(pos)
	if err != nil {
		return err
	}
	e.edit.MoveCursor(p, e.edit.InsertPos, false)
	return nil
}

// BBox returns the bounding box (in widget-window coordinates) of the
// character at idx (Tk: bbox INDEX).
func (e *Entry) BBox(idx string) (x, y, w, h int, err error) {
	p, err := e.resolveIndex(idx)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	if p > len(e.edit.Text) {
		p = len(e.edit.Text)
	}
	if p > 0 {
		p--
	}
	if e.Font == nil || p < 0 || p >= len(e.edit.Text) {
		return 0, 0, 0, 0, nil
	}
	m := e.Font.Metrics()
	x = e.edit.TextX - e.edit.XOffset + e.Font.MeasureString(string(e.edit.Text[:p]))
	y = e.insetY
	w = e.Font.MeasureString(string(e.edit.Text[p : p+1]))
	if w < 1 {
		w = e.Font.MeasureString("0")
	}
	h = m.Linespace()
	return
}

// Validate forces validation. Returns true if the current value passes.
func (e *Entry) Validate() bool {
	if e.ValidateCmd == nil {
		return true
	}
	ok := e.ValidateCmd(e.edit.Get())
	if !ok && e.InvalidCmd != nil {
		e.InvalidCmd()
	}
	return ok
}

// Selection exposes Tk's $entry selection subcommands.
type Selection struct{ e *Entry }

// Selection returns the entry's selection subcommand handle.
func (e *Entry) Selection() *Selection { return &Selection{e: e} }

// Clear removes the current selection.
func (s *Selection) Clear() { s.e.edit.ClearSelection(); s.e.Display() }

// Present reports whether a selection exists.
func (s *Selection) Present() bool { return s.e.edit.HasSelection() }

// Range sets the selection to [first, last).
func (s *Selection) Range(first, last string) error {
	f, err := s.e.resolveIndex(first)
	if err != nil {
		return err
	}
	l, err := s.e.resolveIndex(last)
	if err != nil {
		return err
	}
	if f > l {
		f, l = l, f
	}
	if f < 0 {
		f = 0
	}
	if l > len(s.e.edit.Text) {
		l = len(s.e.edit.Text)
	}
	if f == l {
		s.e.edit.ClearSelection()
	} else {
		s.e.edit.SelAnchor = f
		s.e.edit.SelFirst = f
		s.e.edit.SelLast = l
		s.e.edit.InsertPos = l
	}
	s.e.Display()
	return nil
}

// SetState changes the entry state.
func (e *Entry) SetState(s FieldState) {
	e.StateMode = s
	e.syncState()
	e.Display()
}

// Configure sets options after creation.
func (e *Entry) Configure(opts ...EntryOption) error {
	return configure(&e.TtkWidget, e, opts, func() {
		e.syncState()
		e.linkTextVar()
	}, e.requestSize)
}

// Identify returns the element name under the given point.
func (e *Entry) Identify(x, y int) string {
	if y < 0 || y >= e.Win.Height {
		return ""
	}
	return "textarea"
}

// ---- display ----

// Display draws the entry. Mirrors ttkEntry.c EntryDisplay.
func (e *Entry) Display() {
	if e.Destroyed() {
		return
	}
	win := e.Win
	if win.PlatformID == 0 {
		return
	}
	d := win.Display.Server
	gc := win.GC
	width := win.Width
	height := win.Height
	if width <= 0 || height <= 0 {
		return
	}

	pixDrawable := e.backBuffer(width, height)
	if pixDrawable == 0 {
		return
	}

	fg := LookupColor(e.Context.Style, "-foreground", e.State, 0x000000)
	selBg := LookupColor(e.Context.Style, "-selectbackground", e.State, 0x4a6984)
	selFg := LookupColor(e.Context.Style, "-selectforeground", e.State, 0xffffff)
	insertColor := LookupColor(e.Context.Style, "-insertcolor", e.State, 0x000000)
	insertWidth := LookupInt(e.Context.Style, "-insertwidth", e.State, 1)

	// Entry.field fills the widget (EntryLayout); draw it as the element does.
	(&FieldElement{ctx: e.Context}).Draw(d, pixDrawable, gc, Box{0, 0, width, height}, e.State)

	// Determine display text (mask if -show set).
	display := e.edit.Text
	if e.Show != 0 {
		display = make([]rune, len(e.edit.Text))
		for i := range display {
			display[i] = e.Show
		}
	}
	showPlaceholder := len(e.edit.Text) == 0 && e.Placeholder != "" && e.State&StateFocus == 0

	// Scroll the text horizontally if the insertion cursor has moved off-screen.
	e.maybeScrollIntoView(display)

	ft := fieldText{
		font: e.Font, text: display, x: e.edit.TextX - e.edit.XOffset,
		left: e.edit.TextX, right: width - 2, top: e.insetY, height: height - 2*e.insetY,
		selFirst: -1, selLast: -1, cursor: -1,
		fg: fg, selBg: selBg, selFg: selFg, insertColor: insertColor, insertWidth: insertWidth,
	}
	if showPlaceholder {
		ft.text = []rune(e.Placeholder)
		ft.fg = LookupColor(e.Context.Style, "-placeholderforeground", e.State, 0xb3b3b3)
	} else {
		if e.State&StateFocus != 0 && e.edit.HasSelection() {
			ft.selFirst, ft.selLast = e.edit.SelFirst, e.edit.SelLast
		}
		if e.State&StateFocus != 0 && e.StateMode != FieldDisabled {
			ft.cursor = e.edit.InsertPos
		}
	}
	drawFieldText(d, pixDrawable, gc, ft)

	// Copy to window.
	d.CopyArea(pixDrawable, win.Drawable(), gc, 0, 0, uint(width), uint(height), 0, 0)
	d.Flush()

	if e.XScrollCmd != nil {
		e.XScrollCmd(e.xview(display))
	}
}

// xview returns the visible fraction of the text, as ttkEntry.c's
// EntryDoLayout computes xscroll.first and xscroll.last.
func (e *Entry) xview(display []rune) (first, last float64) {
	if len(display) == 0 || e.Font == nil {
		return 0, 1
	}
	visible := max(e.Win.Width-e.edit.TextX-2, 1)
	right := textedit.RuneIndexAtPixel(e.Font, display, e.edit.XOffset+visible)
	n := float64(len(display))
	return float64(e.leftIndex) / n, float64(max(right, e.leftIndex)) / n
}

// maybeScrollIntoView shifts e.leftIndex so the insertion cursor stays in
// view and, like EntryDoLayout, so no blank space is left at the right while
// text is scrolled off the left.
func (e *Entry) maybeScrollIntoView(display []rune) {
	if e.Font == nil || e.Win.Width <= 0 {
		e.leftIndex, e.edit.XOffset = 0, 0
		return
	}
	visible := max(e.Win.Width-e.edit.TextX-2, 1)
	prefix := func(i int) int { return e.Font.MeasureString(string(display[:i])) }
	curIdx := min(e.edit.InsertPos, len(display))
	left := min(e.leftIndex, len(display))
	cursorX := prefix(curIdx)
	if curIdx < left {
		left = curIdx
	} else if cursorX-prefix(left) > visible {
		left = sort.Search(curIdx, func(i int) bool { return cursorX-prefix(i) <= visible })
	}
	if left > 0 {
		total := prefix(len(display))
		left = min(left, sort.Search(left, func(i int) bool { return total-prefix(i) <= visible }))
	}
	e.leftIndex = left
	e.edit.XOffset = prefix(left)
}

// ---- helpers ----

func (e *Entry) notifyTextVar() {
	if e.TextVar != nil {
		e.TextVar.Set(e.edit.Get())
	}
}

func (e *Entry) resolveIndex(s string) (int, error) {
	switch {
	case s == "end":
		return len(e.edit.Text), nil
	case s == "insert":
		return e.edit.InsertPos, nil
	case strings.HasPrefix(s, "sel.first"):
		if e.edit.SelFirst < 0 {
			return 0, fmt.Errorf("entry: no selection")
		}
		return e.edit.SelFirst, nil
	case strings.HasPrefix(s, "sel.last"):
		if e.edit.SelLast < 0 {
			return 0, fmt.Errorf("entry: no selection")
		}
		return e.edit.SelLast, nil
	case strings.HasPrefix(s, "@"):
		var x int
		if _, err := fmt.Sscanf(s[1:], "%d", &x); err != nil {
			return 0, fmt.Errorf("entry: bad index %q", s)
		}
		return e.edit.ClosestGap(x), nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("entry: bad index %q", s)
	}
	if n < 0 {
		n = 0
	}
	if n > len(e.edit.Text) {
		n = len(e.edit.Text)
	}
	return n, nil
}

func (e *Entry) runValidate(reason entrytext.ValidateReason, newValue string) bool {
	if e.ValidateCmd == nil {
		return true
	}
	var run bool
	switch e.ValidateMode {
	case ValidateKey:
		run = reason == entrytext.ValidateKey
	case ValidateFocus:
		run = reason == entrytext.ValidateFocusIn || reason == entrytext.ValidateFocusOut
	case ValidateFocusIn:
		run = reason == entrytext.ValidateFocusIn
	case ValidateFocusOut:
		run = reason == entrytext.ValidateFocusOut
	case ValidateAll:
		run = true
	default:
		return true
	}
	if !run {
		return true
	}
	ok := e.ValidateCmd(newValue)
	if !ok && e.InvalidCmd != nil {
		e.InvalidCmd()
	}
	return ok
}

// Avoid unused-import errors in builds without the binding file.
var _ = event.ExposureMask
var _ = draw.NewBorderFromPixel
var _ = option.ReliefRaised
