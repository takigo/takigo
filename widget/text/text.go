// Package text implements a multi-line text editor widget with marks, tags,
// undo/redo, word/char wrapping, selection, and scrollbar integration.
// It ports tk/generic/tkText*.c.
package text

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/takigo/takigo/color"
	"github.com/takigo/takigo/cursor"
	"github.com/takigo/takigo/draw"
	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/window"
)

// runeEmbeddedWindow is the placeholder rune inserted into the document text
// at each embedded window position.  The display engine measures this rune as
// the window's width so that text wraps around the window correctly.
const runeEmbeddedWindow = '\uFFFC' // Unicode Object Replacement Character

// embeddedWin records a window embedded at a text index position.
// The window's position is tracked by a mark in the document so it
// adjusts automatically as surrounding text is inserted or deleted.
type embeddedWin struct {
	markName   string
	win        *window.Window
	padX, padY int // -padx / -pady
}

// TextWidget is a multi-line text editor widget.
type TextWidget struct {
	widget.Base

	doc *Document

	// Scroll state.
	topLine       int // 1-based, first visible logical line
	topCharOffset int // display-line offset within topLine (for wrapping)
	// topPixelOffset is how many pixels of the top display line are above
	// the view (tkTextDisp.c's topPixelOffset); the view's first pixel is
	// viewTop().
	topPixelOffset int
	wheelStep      int // pixels a wheel notch scrolls; 0 is Tk's 30 scaled
	wheel          event.WheelAccumulator
	xOffset        int // horizontal pixel offset (WrapNone)

	// Display config.
	wrapMode    WrapMode
	tabWidth    int // characters, default 8
	tabs        []TabStop
	tabStyle    TabStyle
	tabsGen     int // counts Tabs and TabStyleOpt changes, for the layout cache key
	insertWidth int // cursor width in pixels
	cursorOn    bool
	hasFocus    bool
	prefWidth   int // characters
	prefHeight  int // lines
	inset       int
	insetX      int // inset + PadX
	insetY      int // inset + PadY

	// Colors.
	selBg       *color.ColorRef
	selFg       *color.ColorRef
	insertColor *color.ColorRef

	// Selection.
	selAnchor   Index // fixed end during selection drag
	lastDragIdx Index // index under the pointer at the last drag motion

	// Scrollbar callbacks.
	YScrollCmd func(first, last float64)
	XScrollCmd func(first, last float64)

	// Undo.
	undoEnabled bool
	undoStack   *UndoStack

	// Read-only mode: navigation/selection work, editing blocked.
	readOnly bool
	imeMark  Index // insert position when the input method began composing

	// setGrid: if true, the toplevel window's resize increment is kept at the
	// character cell size so the window resizes in whole-character steps.
	setGrid bool

	// Tag event bindings: tagName → eventName → handlers.
	tagBindings map[string]map[string][]func()
	hoverTags   map[string]bool
	clicks      clickCounter
	// tagCursor is the cursor a tag under the pointer set (tagCursorSet).
	tagCursor    cursor.Shape
	tagCursorSet bool
	// mouseCursor is the widget's -cursor, shown where no tag sets one.
	mouseCursor cursor.Shape

	// Embedded images drawn inline with text.

	// Embedded windows (inline child windows positioned at placeholder characters).
	embeddedWindows []embeddedWin
	ewSeq           int // sequence counter for embedded window mark names

	// yScrollPending means the scrollbar still has to be told the new
	// view; redrawYScroll reports it at idle time.
	yScrollPending  bool
	yScrollReported bool

	// Stipple pixmap cache: name → depth-1 Pixmap.
	stippleCache map[string]platform.PixmapID

	layout layoutCache

	// blink blinks the insert cursor while the widget has the focus.
	blink *widget.CursorBlinker
}

// New creates a new TextWidget.
func New(parent widget.Caregiver, name string, opts ...TextOption) *TextWidget {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(w)

	t := &TextWidget{
		doc:         NewDocument(),
		topLine:     1,
		wrapMode:    WrapNone,
		tabWidth:    8,
		insertWidth: 2,
		cursorOn:    true,
		prefWidth:   80,
		prefHeight:  24,
		undoEnabled: true,
		mouseCursor: cursor.XTerm,
	}
	widget.InitBase(&t.Base, w, app)
	w.OnDestroy(t.Destroy)
	w.Class = "Text"
	if f, err := app.FontRegistry().Get(font.TkFixedFont); err == nil {
		t.Font = f // DEF_TEXT_FONT
	}

	// Text widget defaults.
	t.BorderWidth = widget.DefBorderWidth
	t.Relief = option.ReliefSunken
	t.HighlightWidth = 1
	t.PadX, t.PadY = 1, 1 // DEF_TEXT_PADX, DEF_TEXT_PADY

	// White background.
	if bg, err := app.ColorCache().Get(widget.PaletteFor(app).FieldBackground); err == nil {
		t.Background = bg
		t.UpdateBorder()
	}

	// Selection colors.
	if sel, err := app.ColorCache().Get(widget.PaletteFor(app).SelectBackground); err == nil {
		t.selBg = sel.Ref()
	}
	if selfg, err := app.ColorCache().Get(widget.PaletteFor(app).SelectForeground); err == nil {
		t.selFg = selfg.Ref()
	}
	if ins, err := app.ColorCache().Get(widget.PaletteFor(app).InsertBackground); err == nil {
		t.insertColor = ins.Ref()
	}

	// Create the "sel" tag with highest priority.
	selTag := &Tag{Name: "sel", Priority: selPriority}
	if t.selFg != nil {
		if fgCol, err := app.ColorCache().Get(widget.PaletteFor(app).SelectForeground); err == nil {
			selTag.Foreground = fgCol
		}
	}
	if t.selBg != nil {
		if bgCol, err := app.ColorCache().Get(widget.PaletteFor(app).SelectBackground); err == nil {
			selTag.Background = bgCol
		}
	}
	t.doc.putTag(selTag)
	unsubDoc := t.shareDocument(t.doc)
	unsubObjects := t.doc.subscribeObjects(t.objectDeleted)
	w.OnDestroy(func() {
		unsubDoc()
		unsubObjects()
	})

	t.SetDisplayProc(t.display)
	t.blink = widget.NewCursorBlinker(app, t.toggleCursor)

	for _, opt := range opts {
		opt(t)
	}

	t.computeGeometry()

	if t.Background != nil {
		w.SetBackgroundPixel(t.Background.Pixel)
	}

	w.Flags |= window.FlagFocusable
	w.SetCursor(t.mouseCursor) // XC_xterm — I-beam cursor for text
	bindText(t, app)

	return t
}

// NewPeer creates a new TextWidget that shares the given document.
// The new widget registers itself as a change listener so that edits in
// any peer are reflected in all peers.
func NewPeer(doc *Document, parent widget.Caregiver, name string, opts ...TextOption) *TextWidget {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(w)

	t := &TextWidget{
		doc:         doc,
		topLine:     1,
		wrapMode:    WrapNone,
		tabWidth:    8,
		insertWidth: 2,
		cursorOn:    true,
		prefWidth:   80,
		prefHeight:  24,
		// The undo history belongs to the shared document, as in Tk, where
		// the undo stack is part of the shared text: an undo in the
		// primary undoes what a peer did, and an undo in a peer undoes
		// the shared history.
		undoEnabled: true,
		mouseCursor: cursor.XTerm,
	}
	widget.InitBase(&t.Base, w, app)
	w.OnDestroy(t.Destroy)
	w.Class = "Text"
	if f, err := app.FontRegistry().Get(font.TkFixedFont); err == nil {
		t.Font = f // DEF_TEXT_FONT
	}

	t.BorderWidth = widget.DefBorderWidth
	t.Relief = option.ReliefSunken
	t.HighlightWidth = 1
	t.PadX, t.PadY = 1, 1 // DEF_TEXT_PADX, DEF_TEXT_PADY

	if bg, err := app.ColorCache().Get(widget.PaletteFor(app).FieldBackground); err == nil {
		t.Background = bg
		t.UpdateBorder()
	}
	if sel, err := app.ColorCache().Get(widget.PaletteFor(app).SelectBackground); err == nil {
		t.selBg = sel.Ref()
	}
	if selfg, err := app.ColorCache().Get(widget.PaletteFor(app).SelectForeground); err == nil {
		t.selFg = selfg.Ref()
	}
	if ins, err := app.ColorCache().Get(widget.PaletteFor(app).InsertBackground); err == nil {
		t.insertColor = ins.Ref()
	}

	unsubDoc := t.shareDocument(doc)
	unsubObjects := t.doc.subscribeObjects(t.objectDeleted)
	t.SetDisplayProc(t.display)
	t.blink = widget.NewCursorBlinker(app, t.toggleCursor)

	for _, opt := range opts {
		opt(t)
	}

	t.computeGeometry()

	if t.Background != nil {
		w.SetBackgroundPixel(t.Background.Pixel)
	}

	w.Flags |= window.FlagFocusable
	w.SetCursor(t.mouseCursor)
	bindText(t, app)

	// The document no longer refers to the widget once it is gone.
	w.OnDestroy(func() {
		unsubDoc()
		unsubObjects()
	})

	return t
}

// shareDocument makes the widget a view of the document: it takes the
// document's undo history, as Tk's peers share the undo stack of the
// shared text, and subscribes to the document's changes. It returns a
// function that unsubscribes the widget again, so a destroyed widget is
// not kept alive by the document.
func (t *TextWidget) shareDocument(doc *Document) (unsubscribe func()) {
	t.layout.init(t)
	t.undoStack = doc.UndoStack()
	t.undoEnabled = true
	return t.subscribeDocument()
}

// subscribeDocument registers the widget with the document: every change
// invalidates the widget's layout cache and asks for a repaint, as Tk
// redraws all the peers of a shared text. It returns a function that
// unsubscribes the widget again, so a destroyed widget is not kept alive
// by the document.
func (t *TextWidget) subscribeDocument() (unsubscribe func()) {
	unsubLayout := t.doc.Subscribe(t.layout.apply)
	unsubRedraw := t.doc.Subscribe(func(ch Change) {
		t.keepViewInDocument(ch)
		t.notifyYScrollbar()
		t.scheduleRedraw()
	})
	return func() {
		unsubLayout()
		unsubRedraw()
	}
}

// keepViewInDocument keeps topLine/topCharOffset on existing text after a
// change: lines removed above the view move topLine up with them, as Tk's
// start-of-view index follows the text, and a view that lost its line falls
// back to the line the removal joined it to. It runs inside the document's
// change notification, before any idle callback can read the view.
func (t *TextWidget) keepViewInDocument(ch Change) {
	if n := -ch.Delta; n > 0 && t.topLine > ch.From {
		if t.topLine > ch.From+n {
			t.topLine -= n
		} else {
			t.topLine, t.topCharOffset, t.topPixelOffset = ch.From, 0, 0
		}
	}
	t.clampView()
}

// clampView pulls the view back inside the document: topLine into
// [1, LineCount] (offsets 0 when the line had to move), topCharOffset
// into the display lines of topLine and topPixelOffset into that display
// line.
func (t *TextWidget) clampView() {
	n := t.doc.LineCount()
	if t.topLine > n {
		t.topLine, t.topCharOffset, t.topPixelOffset = max(n, 1), 0, 0
	}
	if t.topLine < 1 {
		t.topLine, t.topCharOffset, t.topPixelOffset = 1, 0, 0
	}
	if (t.topCharOffset > 0 || t.topPixelOffset > 0) && n > 0 {
		hs := t.displayLinePixels(t.topLine)
		if c := min(t.topCharOffset, max(len(hs)-1, 0)); c != t.topCharOffset {
			t.topCharOffset, t.topPixelOffset = c, 0
		}
		h := 0
		if t.topCharOffset < len(hs) {
			h = hs[t.topCharOffset]
		}
		t.topPixelOffset = min(t.topPixelOffset, max(h-1, 0))
	}
}

// computeGeometry calculates the requested window size.
func (t *TextWidget) computeGeometry() {
	t.inset = t.BorderWidth + t.HighlightWidth
	t.insetX = t.inset + t.PadX
	t.insetY = t.inset + t.PadY
	if t.Font == nil {
		return
	}
	m := t.Font.Metrics()
	avgWidth := max(t.Font.MeasureString("0"), 1)
	lineHeight := m.Linespace()

	w := t.Win
	w.ReqWidth = t.prefWidth*avgWidth + 2*t.insetX
	w.ReqHeight = t.prefHeight*lineHeight + 2*t.insetY

	// -setgrid: the toplevel resizes in whole character cells
	// (Tk_SetGrid in TextWorldChanged; Tk_UnsetGrid when it goes off).
	if top := window.Toplevel(t.Win); top != nil && top.WmData != nil {
		if t.setGrid {
			top.WmData.SetGrid(t.Win, t.prefWidth, t.prefHeight, avgWidth, lineHeight)
		} else {
			top.WmData.UnsetGrid(t.Win)
		}
	}
}

// Display schedules a redraw; like Tk's REDRAW_PENDING, any number of
// edits between two idle points paint once.
func (t *TextWidget) Display() {
	t.scheduleRedraw()
}

// display draws the text widget. It runs from the base redraw, which has
// redirected the window's drawable to an off-screen pixmap, so the widget
// paints into that pixmap and the base copies it to the window in one
// request (Tk's Tk_GetPixmap in TkTextDisplayProc).
func (t *TextWidget) display() {
	if t.Destroyed() {
		return
	}
	w := t.Win
	if w.PlatformID == 0 {
		return
	}

	d := w.Display.Server
	gc := w.GC

	winW := w.Width
	winH := w.Height
	if winW <= 0 || winH <= 0 {
		return
	}

	drawable := w.Drawable()
	dlines := t.renderTo(drawable)

	if len(t.embeddedWindows) > 0 {
		t.positionEmbeddedWindows(dlines)
	}

	// Draw border on top.
	if t.Border != nil && t.BorderWidth > 0 {
		hl := t.HighlightWidth
		draw.Draw3DRectangle(d, drawable, gc, t.Border,
			hl, hl, w.Width-2*hl, w.Height-2*hl, t.BorderWidth, t.Relief)
	}
	// The highlight ring sits outside the border (focus colour or
	// -highlightbackground).
	t.DrawHighlightBorder(t.hasFocus, 0)

}

// scheduleRedraw schedules a redraw via the idle loop. The base redraw
// paints into an off-screen pixmap and copies it to the window, as Tk's
// display procs do.
func (t *TextWidget) scheduleRedraw() {
	if t.Destroyed() || t.App == nil {
		return
	}
	t.EventuallyRedraw()
}

// reportYScroll tells the scrollbar the new view at idle time.
func (t *TextWidget) reportYScroll() {
	if t.yScrollReported || t.App == nil {
		return
	}
	t.yScrollReported = true
	t.App.DoWhenIdle(func() {
		t.yScrollReported = false
		if t.Destroyed() || !t.yScrollPending {
			return
		}
		t.yScrollPending = false
		t.YScrollCmd(t.yviewFractions())
	})
}

// --- Public API ---

// Insert inserts text at the given index string.
func (t *TextWidget) Insert[I IndexSpec](index I, txt string) error {
	idx, ok := indexOf(t, index)
	if !ok {
		return ErrBadIndex
	}
	endIdx := t.doc.Insert(idx, txt)
	if t.undoEnabled {
		t.undoStack.RecordInsert(idx, endIdx, txt)
	}
	t.notifyYScrollbar()
	t.Display()
	return nil
}

// Delete deletes text between two index strings.
func (t *TextWidget) Delete[A, B IndexSpec](startIndex A, endIndex B) error {
	start, ok1 := indexOf(t, startIndex)
	end, ok2 := indexOf(t, endIndex)
	if !ok1 || !ok2 {
		return ErrBadIndex
	}
	if Compare(start, end) >= 0 {
		return nil
	}
	text := t.doc.Get(start, end)
	t.doc.Delete(start, end)
	if t.undoEnabled {
		t.undoStack.RecordDelete(start, end, text)
	}
	t.notifyYScrollbar()
	t.Display()
	return nil
}

// Get returns the text between two index strings.
func (t *TextWidget) Get[A, B IndexSpec](startIndex A, endIndex B) string {
	start, ok1 := indexOf(t, startIndex)
	end, ok2 := indexOf(t, endIndex)
	if !ok1 || !ok2 {
		return ""
	}
	return withoutPlaceholders(t.doc.Get(start, end))
}

// withoutPlaceholders drops embedded windows' and images' placeholder
// characters: as in Tk, text retrieval returns only the characters.
func withoutPlaceholders(s string) string {
	return strings.ReplaceAll(s, string(runeEmbeddedWindow), "")
}

// See scrolls the view to make the given index visible.
func (t *TextWidget) See[I IndexSpec](index I) error {
	idx, ok := indexOf(t, index)
	if !ok {
		return ErrBadIndex
	}
	t.seeIndex(idx)
	t.notifyYScrollbar()
	t.Display()
	return nil
}

// SetInsertPos moves the insert cursor to the given index.
func (t *TextWidget) SetInsertPos[I IndexSpec](index I) error {
	idx, ok := indexOf(t, index)
	if !ok {
		return ErrBadIndex
	}
	t.doc.MarkSet("insert", idx)
	return nil
}

// MarkSet sets a mark at the given index.
func (t *TextWidget) MarkSet[I IndexSpec](markName string, index I) error {
	idx, ok := indexOf(t, index)
	if !ok {
		return ErrBadIndex
	}
	t.doc.MarkSet(markName, idx)
	return nil
}

// MarkNames returns all mark names.
func (t *TextWidget) MarkNames() []string {
	return t.doc.MarkNames()
}

// Edit performs undo/redo operations.
func (t *TextWidget) Edit(mode string) {
	switch mode {
	case "undo":
		if t.undoStack.Undo(t.doc) {
			t.seeInsert()
			t.notifyYScrollbar()
			t.Display()
		}
	case "redo":
		if t.undoStack.Redo(t.doc) {
			t.seeInsert()
			t.notifyYScrollbar()
			t.Display()
		}
	case "reset":
		t.undoStack.Reset()
	case "separator":
		t.undoStack.Separator()
	}
}

// SetWrapMode changes the wrap mode at runtime.
func (t *TextWidget) SetWrapMode(mode WrapMode) error {
	return t.Configure(WrapModeOpt(mode))
}

// --- Scroll API ---

// YView scrolls to make the given line the top visible line.
func (t *TextWidget) YView(line int) {
	if line < 1 {
		line = 1
	}
	if line > t.doc.LineCount() {
		line = t.doc.LineCount()
	}
	t.topLine, t.topCharOffset, t.topPixelOffset = line, 0, 0
	t.clampScrollPosition()
	t.notifyYScrollbar()
	t.Display()
}

// YViewMoveTo scrolls to a fraction of the total pixel height (yview
// moveto): that pixel becomes the first one shown, part way into its
// display line.
func (t *TextWidget) YViewMoveTo(fraction float64) {
	total := t.layout.totalPixels()
	if total <= 0 {
		return
	}
	fraction = min(max(fraction, 0), 1)
	t.setViewTop(min(int(0.5+fraction*float64(total)), total-1))
	t.clampScrollPosition()
	t.notifyYScrollbar()
	t.Display()
}

// YViewScroll scrolls by count units (display lines) or pages (yview
// scroll N units|pages). A page is the view's height less two lines, or
// 3/4 of it when a line is over a quarter of the view, in pixels as
// TextYviewCmd does.
func (t *TextWidget) YViewScroll(count int, pages bool) {
	if pages {
		height := t.Win.Height - 2*t.insetY
		lh := t.lineHeight()
		px := height - 2*lh
		if lh*4 >= height {
			if px = 3 * height / 4; px < lh {
				px = min(lh, height)
			}
		}
		t.YViewScrollPixels(px * count)
		return
	}
	t.scrollByDisplayLines(count)
	t.clampScrollPosition()
	t.notifyYScrollbar()
	t.Display()
}

// YViewScrollPixels scrolls by n pixels (yview scroll N pixels); positive
// shows text further down. The top display line may be left partly above
// the view, as YScrollByPixels does.
func (t *TextWidget) YViewScrollPixels(n int) {
	if n == 0 {
		return
	}
	t.setViewTop(t.viewTop() + n)
	t.clampScrollPosition()
	t.notifyYScrollbar()
	t.Display()
}

// wheelPixels is how far one notch of the wheel scrolls: WheelStep, or
// tk::ScaleNum of 120 over text.tcl's factor of 4 (30 pixels at 96 DPI).
func (t *TextWidget) wheelPixels() int {
	if t.wheelStep > 0 {
		return t.wheelStep
	}
	return max(int(math.Round(30*screenunit.ScalingFactor())), 1)
}

// XView scrolls to a horizontal pixel offset.
func (t *TextWidget) XView(offset int) {
	if offset < 0 {
		offset = 0
	}
	t.xOffset = offset
	t.notifyXScrollbar()
	t.Display()
}

// XViewMoveTo scrolls horizontally to a fraction.
func (t *TextWidget) XViewMoveTo(fraction float64) {
	// Approximate total width from longest visible line.
	maxW := t.estimateMaxLineWidth()
	t.xOffset = max(int(fraction*float64(maxW)), 0)
	t.notifyXScrollbar()
	t.Display()
}

// XViewScroll scrolls horizontally.
func (t *TextWidget) XViewScroll(count int, pages bool) {
	if pages {
		availW := max(0, t.Win.Width-2*t.insetX)
		count *= availW
	} else {
		count *= t.Font.MeasureString("0")
	}
	t.XView(t.xOffset + count)
}

// Configure applies options.
func (t *TextWidget) Configure(opts ...TextOption) error {
	return widget.Configure(t, opts, t.computeGeometry)
}

// Destroy cleans up the text widget.
func (t *TextWidget) Destroy() {
	if t.Destroyed() {
		return
	}
	t.MarkDestroyed()
	t.blink.Stop()
	if top := window.Toplevel(t.Win); top != nil && top.WmData != nil {
		top.WmData.UnsetGrid(t.Win)
	}
	d := t.Win.Display.Server
	for _, pm := range t.stippleCache {
		if pm != 0 {
			d.FreePixmap(pm)
		}
	}
	t.stippleCache = nil
	window.DestroyWindow(t.Win)
}

// Doc returns the underlying document (for advanced use).
func (t *TextWidget) Doc() *Document {
	return t.doc
}

// EndIndex returns the current end position as a "line.char" string.
func (t *TextWidget) EndIndex() Index {
	n := t.doc.LineCount()
	if n == 0 {
		return Index{Line: 1}
	}
	return Index{Line: n, Char: len(t.doc.Lines[n-1].Text)}
}

// WindowCreate embeds a child window inline at the given text index.
// A placeholder character is inserted into the document so that text wraps
// around the window, and the position is tracked by a mark.
func (t *TextWidget) WindowCreate[I IndexSpec](indexStr I, w *window.Window) error {
	t.WindowCreatePad(indexStr, w, 0, 0)
	return nil
}

// WindowCreatePad is "window create" with -padx and -pady (Tk distances).
func (t *TextWidget) WindowCreatePad[I IndexSpec, X, Y screenunit.Length](indexStr I, w *window.Window, padX X, padY Y) error {
	idx, ok := indexOf(t, indexStr)
	if !ok {
		return ErrBadIndex
	}
	// Insert placeholder character into the document.
	t.doc.Insert(idx, string(runeEmbeddedWindow))
	// Track the placeholder with a mark just before it. The mark has right
	// gravity so that text inserted at the window's position (before the
	// placeholder) moves the mark along with the placeholder.
	t.ewSeq++
	markName := fmt.Sprintf("_ew%d", t.ewSeq)
	t.doc.MarkSet(markName, idx)
	t.doc.addObject(markName)
	t.embeddedWindows = append(t.embeddedWindows, embeddedWin{markName: markName, win: w,
		padX: screenunit.ToPixels(padX), padY: screenunit.ToPixels(padY)})
	return nil
}

// lineHeightFor returns the display line height for the given logical line,
// taking into account any embedded images and windows on that line.
func (t *TextWidget) lineHeightFor(lineIdx int) int {
	h := t.lineHeight()
	t.doc.eachImage(func(pos Index, ei *embImage) {
		if pos.Line == lineIdx && ei.float == FloatNone {
			h = max(h, ei.height())
		}
	})
	for _, ew := range t.embeddedWindows {
		if m, ok := t.doc.Marks[ew.markName]; ok && m.Pos.Line == lineIdx {
			if wh := ew.win.ReqHeight; wh > h {
				h = wh
			}
		}
	}
	return h
}

// RemoveWindow removes an embedded window from the text widget and unmaps it.
// The placeholder character and tracking mark are also removed.
func (t *TextWidget) RemoveWindow(win *window.Window) {
	d := t.Win.Display.Server
	for i, ew := range t.embeddedWindows {
		if ew.win == win {
			d.UnmapWindow(win.PlatformID)
			// Forget the window first so deleting its placeholder does not
			// destroy it (objectDeleted), then delete the placeholder.
			t.embeddedWindows = append(t.embeddedWindows[:i], t.embeddedWindows[i+1:]...)
			if m, ok := t.doc.Marks[ew.markName]; ok {
				pos := m.Pos
				t.doc.Delete(pos, Index{Line: pos.Line, Char: pos.Char + 1})
				t.doc.MarkUnset(ew.markName)
			}
			t.Display()
			return
		}
	}
}

// objectDeleted destroys an embedded window of this widget whose position
// was deleted, as EmbWinDeleteProc does.
func (t *TextWidget) objectDeleted(name string) {
	for i, ew := range t.embeddedWindows {
		if ew.markName == name {
			t.embeddedWindows = slices.Delete(t.embeddedWindows, i, i+1)
			window.DestroyWindow(ew.win)
			return
		}
	}
}

// MarkGravity sets the gravity of the named mark.
func (t *TextWidget) MarkGravity(markName string, gravity MarkGravity) {
	m, ok := t.doc.Marks[markName]
	if !ok {
		return
	}
	m.Gravity = gravity
}

// SetPadX sets horizontal padding between the border and the text content.
func (t *TextWidget) SetPadX(n int) error {
	return t.Configure(PadXOpt(n))
}

// SetPadY sets vertical padding between the border and the text content.
func (t *TextWidget) SetPadY(n int) error {
	return t.Configure(PadYOpt(n))
}

// positionEmbeddedWindows moves embedded windows to their correct inline
// positions within the text flow.
func (t *TextWidget) positionEmbeddedWindows(dlines []displayLine) {
	d := t.Win.Display.Server
	for _, ew := range t.embeddedWindows {
		m, ok := t.doc.Marks[ew.markName]
		if !ok {
			continue
		}
		pos := m.Pos
		visible := false
		for _, dl := range dlines {
			if dl.logicalLine != pos.Line {
				continue
			}
			if pos.Char < dl.startChar || pos.Char >= dl.endChar {
				continue
			}
			// Compute x by measuring content before the window on this display line.
			xBefore := t.measureRange(pos.Line, dl.startChar, pos.Char)

			// Apply justification offset (same logic as renderToPixmap).
			totalW := t.measureRange(pos.Line, dl.startChar, dl.endChar)
			availW := t.Win.Width - 2*t.insetX - dl.leftMargin - dl.rightMargin
			justifyOffset := 0
			switch dl.justify {
			case option.JustifyCenter:
				justifyOffset = max((availW-totalW)/2, 0)
			case option.JustifyRight:
				justifyOffset = max(availW-totalW, 0)
			default:
			}

			wx := t.insetX + dl.leftMargin + justifyOffset - t.xOffset + xBefore + ew.padX
			ww := ew.win.ReqWidth
			wh := ew.win.ReqHeight
			if wh == 0 {
				wh = dl.height
			}
			// EmbWinDisplayProc, -align center: centred in the line.
			wy := t.insetY + dl.y + (dl.height-wh)/2
			d.MoveResizeWindow(ew.win.PlatformID, wx, wy, uint(ww), uint(wh))
			d.MapWindow(ew.win.PlatformID)
			ew.win.X, ew.win.Y, ew.win.Width, ew.win.Height = wx, wy, ww, wh
			window.MarkMapped(ew.win)
			visible = true
			break
		}
		if !visible {
			d.UnmapWindow(ew.win.PlatformID)
			window.MarkUnmapped(ew.win)
		}
	}
}

func (t *TextWidget) embeddedWinAt(lineIdx, charIdx int) *embeddedWin {
	for i := range t.embeddedWindows {
		ew := &t.embeddedWindows[i]
		if m, ok := t.doc.Marks[ew.markName]; ok {
			if m.Pos.Line == lineIdx && m.Pos.Char == charIdx {
				return ew
			}
		}
	}
	return nil
}

// --- Internal helpers ---

// seeInsert scrolls to make the insert cursor visible.
func (t *TextWidget) seeInsert() {
	insertPos := t.doc.Marks["insert"].Pos
	t.seeIndex(insertPos)
}

// seeIndex scrolls to make a given index visible.
func (t *TextWidget) seeIndex(idx Index) {
	idx = Clamp(idx, t.doc)

	if idx.Line < 1 || idx.Line > t.doc.LineCount() {
		return
	}
	ll := t.layout.line(idx.Line)
	k := 0
	for k < len(ll.dls)-1 && idx.Char >= ll.dls[k].endChar {
		k++
	}
	top := t.displayLineTop(idx.Line, k)
	h := 0
	if k < len(ll.hs) {
		h = ll.hs[k]
	}
	view := t.Win.Height - 2*t.insetY
	switch vt := t.viewTop(); {
	case top < vt:
		// Above the view, or the top line partly above it
		// (TkTextSetYView sets newTopPixelOffset to 0).
		t.topLine, t.topCharOffset, t.topPixelOffset = idx.Line, k, 0
	case view > 0 && top+h > vt+view:
		// Below the view or hanging off its bottom: the line's bottom goes
		// to the bottom of the view, its top when it is taller than the view.
		if h > view {
			t.topLine, t.topCharOffset, t.topPixelOffset = idx.Line, k, 0
		} else {
			t.setViewTop(top + h - view)
		}
	}
}

// clampScrollPosition ensures the view doesn't scroll past the end of content.
// The last line of content should not scroll above the bottom of the viewport.
// As MeasureUp in tkTextDisp.c, the highest top is found in pixels, up from
// the end: display lines differ in height (images, fonts, spacing), so a
// count of lines of the default height would stop short of the end.
func (t *TextWidget) clampScrollPosition() {
	total := t.layout.totalPixels()
	view := max(t.Win.Height-2*t.insetY, 1)
	if total <= view {
		// All content fits — reset to top.
		t.topLine, t.topCharOffset, t.topPixelOffset = 1, 0, 0
		return
	}
	// The space left below the end is filled by showing more of the top
	// line and the lines above it (UpdateDisplayInfo), so the highest top
	// is the pixel total-view, part way into its display line.
	if t.viewTop() > total-view {
		t.setViewTop(total - view)
	}
}

// viewTop is the document pixel shown at the top of the text area.
func (t *TextWidget) viewTop() int {
	t.clampView()
	return t.displayLineTop(t.topLine, t.topCharOffset) + t.topPixelOffset
}

// setViewTop makes document pixel y the first one shown: the display line
// holding it becomes the top, with the pixels above y hidden.
func (t *TextWidget) setViewTop(y int) {
	total := t.layout.totalPixels()
	if total <= 0 || y <= 0 {
		t.topLine, t.topCharOffset, t.topPixelOffset = 1, 0, 0
		return
	}
	y = min(y, total-1)
	t.topLine, t.topCharOffset = t.layout.lineAtPixel(y)
	t.topPixelOffset = max(y-t.displayLineTop(t.topLine, t.topCharOffset), 0)
}

// displayLineTop is the document pixel at which display line off of a
// logical line starts.
func (t *TextWidget) displayLineTop(line, off int) int {
	y := t.layout.pixelsBefore(line)
	hs := t.displayLinePixels(line)
	for _, h := range hs[:min(off, len(hs))] {
		y += h
	}
	return y
}

// scrollByDisplayLines scrolls by n display lines (positive = down, negative = up).
func (t *TextWidget) scrollByDisplayLines(n int) {
	total := t.totalDisplayLines()
	if total == 0 {
		return
	}
	cur := t.computeDisplayLinesBefore(t.topLine, t.topCharOffset)
	target := min(max(cur+n, 0), total-1)
	t.topLine, t.topCharOffset = t.layout.lineAtDisplayLine(target)
	t.topPixelOffset = 0
}

// notifyYScrollbar ports GetYView (tkTextDisp.c): the fractions are pixel
// counts, the pixels above the top display line and those shown up to the
// bottom of the text area, over the pixel height of the whole text.
func (t *TextWidget) notifyYScrollbar() {
	if t.YScrollCmd == nil {
		return
	}
	t.yScrollPending = true
	t.reportYScroll()
	t.scheduleRedraw()
}

// YViewFractions returns the fractions of the total pixel height above the
// view and at its bottom (yview with no arguments); YViewMoveTo(first)
// shows the same place again.
func (t *TextWidget) YViewFractions() (first, last float64) { return t.yviewFractions() }

func (t *TextWidget) yviewFractions() (float64, float64) {
	t.clampView()
	total := t.layout.totalPixels()
	if total == 0 {
		return 0, 1
	}
	above := t.layout.pixelsBefore(t.topLine)
	hs := t.displayLinePixels(t.topLine)
	for _, h := range hs[:min(t.topCharOffset, len(hs))] {
		above += h
	}
	count := above
	above += t.topPixelOffset
	maxY := t.Win.Height - 2*t.insetY
	for _, dl := range t.computeVisibleLines() {
		count += dl.spacingAbove + dl.height + dl.spacingBelow
		if extra := dl.y + dl.height + dl.spacingBelow - maxY; extra > 0 {
			count -= extra
			break
		}
	}
	count = min(count, total)
	return float64(above) / float64(total), float64(count) / float64(total)
}

// notifyXScrollbar calls the X scroll callback.
func (t *TextWidget) notifyXScrollbar() {
	if t.XScrollCmd == nil {
		return
	}
	maxW := t.estimateMaxLineWidth()
	availW := t.Win.Width - 2*t.insetX
	if maxW <= 0 {
		t.XScrollCmd(0, 1)
		return
	}
	first := float64(t.xOffset) / float64(maxW)
	last := float64(t.xOffset+availW) / float64(maxW)
	if first < 0 {
		first = 0
	}
	if last > 1 {
		last = 1
	}
	t.XScrollCmd(first, last)
}

// estimateMaxLineWidth returns an estimate of the widest line's pixel width.
func (t *TextWidget) estimateMaxLineWidth() int {
	if t.Font == nil {
		return 0
	}
	return t.layout.maxWidth()
}

// Suppress unused import warnings.
var _ = (*color.Color)(nil)
var _ = (*font.Registry)(nil)

// toggleCursor blinks the insert cursor, as Tk's text widget does.
func (t *TextWidget) toggleCursor() {
	if t.readOnly {
		return
	}
	t.cursorOn = !t.cursorOn
	t.Display()
}
