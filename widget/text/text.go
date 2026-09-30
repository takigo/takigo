// Package text implements a multi-line text editor widget with marks, tags,
// undo/redo, word/char wrapping, selection, and scrollbar integration.
// It ports tk/generic/tkText*.c.
package text

import (
	"fmt"

	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/cursor"
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// runeEmbeddedWindow is the placeholder rune inserted into the document text
// at each embedded window position.  The display engine measures this rune as
// the window's width so that text wraps around the window correctly.
const runeEmbeddedWindow = '\uFFFC' // Unicode Object Replacement Character

// embeddedImage records an image embedded at a text index position.
type embeddedImage struct {
	index Index
	img   widget.WidgetImage
}

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
	xOffset       int // horizontal pixel offset (WrapNone)

	// Display config.
	wrapMode    WrapMode
	tabWidth    int // characters, default 8
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

	// Embedded images drawn inline with text.
	embeddedImages []embeddedImage

	// Embedded windows (inline child windows positioned at placeholder characters).
	embeddedWindows []embeddedWin
	ewSeq           int // sequence counter for embedded window mark names

	// Offscreen pixmap.
	pixmap           platform.PixmapID
	pixmapW, pixmapH int
	redrawPending    bool
	yScrollPending   bool

	// Stipple pixmap cache: name → depth-1 Pixmap.
	stippleCache map[string]platform.PixmapID

	layout layoutCache
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
	if bg, err := app.ColorCache().Get("#ffffff"); err == nil {
		t.Background = bg
		t.UpdateBorder()
	}

	// Selection colors.
	if sel, err := app.ColorCache().Get("#3399ff"); err == nil {
		t.selBg = sel.Ref()
	}
	if selfg, err := app.ColorCache().Get("#ffffff"); err == nil {
		t.selFg = selfg.Ref()
	}
	if ins, err := app.ColorCache().Get("#000000"); err == nil {
		t.insertColor = ins.Ref()
	}

	// Create the "sel" tag with highest priority.
	selTag := &Tag{Name: "sel", Priority: selPriority}
	if t.selFg != nil {
		if fgCol, err := app.ColorCache().Get("#ffffff"); err == nil {
			selTag.Foreground = fgCol
		}
	}
	if t.selBg != nil {
		if bgCol, err := app.ColorCache().Get("#3399ff"); err == nil {
			selTag.Background = bgCol
		}
	}
	t.doc.putTag(selTag)
	t.layout.init(t)
	t.doc.Listeners = append(t.doc.Listeners, t.layout.apply)

	t.undoStack = NewUndoStack(100)

	for _, opt := range opts {
		opt(t)
	}

	t.computeGeometry()

	if t.Background != nil {
		w.SetBackgroundPixel(t.Background.Pixel)
	}

	w.Flags |= window.FlagFocusable
	w.SetCursor(uint(cursor.XTerm)) // XC_xterm — I-beam cursor for text
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
		undoEnabled: false, // peers share history via the primary
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

	if bg, err := app.ColorCache().Get("#ffffff"); err == nil {
		t.Background = bg
		t.UpdateBorder()
	}
	if sel, err := app.ColorCache().Get("#3399ff"); err == nil {
		t.selBg = sel.Ref()
	}
	if selfg, err := app.ColorCache().Get("#ffffff"); err == nil {
		t.selFg = selfg.Ref()
	}
	if ins, err := app.ColorCache().Get("#000000"); err == nil {
		t.insertColor = ins.Ref()
	}

	t.undoStack = NewUndoStack(0)
	t.layout.init(t)
	unsubLayout := doc.Subscribe(t.layout.apply)

	for _, opt := range opts {
		opt(t)
	}

	t.computeGeometry()

	if t.Background != nil {
		w.SetBackgroundPixel(t.Background.Pixel)
	}

	w.Flags |= window.FlagFocusable
	w.SetCursor(uint(cursor.XTerm))
	bindText(t, app)

	// Register as a document listener so edits from other peers trigger a
	// redraw; the peer stops listening when destroyed.
	unsubRedraw := doc.Subscribe(func(Change) {
		t.notifyYScrollbar()
		t.scheduleRedraw()
	})
	w.OnDestroy(func() {
		unsubLayout()
		unsubRedraw()
	})

	return t
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

	if t.setGrid {
		t.applySetGrid(avgWidth, lineHeight)
	}
}

// applySetGrid sets the WM size-increment hints on the nearest toplevel so the
// window resizes in whole character steps (Tk's -setgrid 1 behaviour).
func (t *TextWidget) applySetGrid(charW, lineH int) {
	top := window.Toplevel(t.Win)
	if top == nil || top.PlatformID == 0 {
		return
	}
	inset := 2 * t.inset
	hints := &platform.SizeHints{
		Flags:     platform.PResizeInc | platform.PMinSize,
		WidthInc:  charW,
		HeightInc: lineH,
		MinWidth:  inset + charW,
		MinHeight: inset + lineH,
	}
	t.App.Server().SetWMNormalHints(top.PlatformID, hints)
}

// Display schedules a redraw; like Tk's REDRAW_PENDING, any number of
// edits between two idle points paint once.
func (t *TextWidget) Display() {
	t.scheduleRedraw()
}

// display draws the text widget.
func (t *TextWidget) display() {
	if t.Destroyed {
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

	// Allocate or resize pixmap.
	if t.pixmap == 0 || t.pixmapW != winW || t.pixmapH != winH {
		if t.pixmap != 0 {
			d.FreePixmap(t.pixmap)
		}
		t.pixmap = d.CreatePixmap(w.Drawable(), uint(winW), uint(winH), uint(w.Depth))
		t.pixmapW = winW
		t.pixmapH = winH
	}
	if t.pixmap == 0 {
		return
	}

	dlines := t.renderToPixmap()

	// Copy pixmap to window.
	d.CopyArea(platform.PixmapDrawable(t.pixmap), w.Drawable(), gc,
		0, 0, uint(winW), uint(winH), 0, 0)

	if len(t.embeddedWindows) > 0 {
		t.positionEmbeddedWindows(dlines)
	}

	// Draw border on top.
	if t.Border != nil && t.BorderWidth > 0 {
		hl := t.HighlightWidth
		draw.Draw3DRectangle(d, w.Drawable(), gc, t.Border,
			hl, hl, w.Width-2*hl, w.Height-2*hl, t.BorderWidth, t.Relief)
	}
	// The highlight ring sits outside the border (focus colour or
	// -highlightbackground).
	t.DrawHighlightBorder(t.hasFocus, 0)

}

// scheduleRedraw schedules a redraw via the idle loop.
func (t *TextWidget) scheduleRedraw() {
	if t.redrawPending || t.Destroyed || t.App == nil {
		return
	}
	t.redrawPending = true
	t.App.DoWhenIdle(func() {
		t.redrawPending = false
		if t.Destroyed {
			return
		}
		if t.yScrollPending {
			t.yScrollPending = false
			t.YScrollCmd(t.yviewFractions())
		}
		t.display()
	})
}

// --- Public API ---

// Insert inserts text at the given index string.
func (t *TextWidget) Insert(index, txt string) {
	idx, ok := ParseIndex(t.doc, index)
	if !ok {
		return
	}
	endIdx := t.doc.Insert(idx, txt)
	if t.undoEnabled {
		t.undoStack.RecordInsert(idx, endIdx, txt)
	}
	t.notifyYScrollbar()
	t.Display()
}

// Delete deletes text between two index strings.
func (t *TextWidget) Delete(startIndex, endIndex string) {
	start, ok1 := ParseIndex(t.doc, startIndex)
	end, ok2 := ParseIndex(t.doc, endIndex)
	if !ok1 || !ok2 {
		return
	}
	if Compare(start, end) >= 0 {
		return
	}
	text := t.doc.Get(start, end)
	t.doc.Delete(start, end)
	if t.undoEnabled {
		t.undoStack.RecordDelete(start, end, text)
	}
	t.notifyYScrollbar()
	t.Display()
}

// Get returns the text between two index strings.
func (t *TextWidget) Get(startIndex, endIndex string) string {
	start, ok1 := ParseIndex(t.doc, startIndex)
	end, ok2 := ParseIndex(t.doc, endIndex)
	if !ok1 || !ok2 {
		return ""
	}
	return t.doc.Get(start, end)
}

// See scrolls the view to make the given index visible.
func (t *TextWidget) See(index string) {
	idx, ok := ParseIndex(t.doc, index)
	if !ok {
		return
	}
	t.seeIndex(idx)
	t.notifyYScrollbar()
	t.Display()
}

// SetInsertPos moves the insert cursor to the given index.
func (t *TextWidget) SetInsertPos(index string) {
	idx, ok := ParseIndex(t.doc, index)
	if !ok {
		return
	}
	t.doc.MarkSet("insert", idx)
}

// MarkSet sets a mark at the given index.
func (t *TextWidget) MarkSet(markName, index string) {
	idx, ok := ParseIndex(t.doc, index)
	if !ok {
		return
	}
	t.doc.MarkSet(markName, idx)
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
func (t *TextWidget) SetWrapMode(mode WrapMode) {
	t.Configure(WrapModeOpt(mode))
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
	t.topLine = line
	t.topCharOffset = 0
	t.clampScrollPosition()
	t.notifyYScrollbar()
	t.Display()
}

// YViewMoveTo scrolls to a fraction of the total pixel height (yview
// moveto), topping the display line that holds that pixel; the Go text has
// no partial-line topPixelOffset.
func (t *TextWidget) YViewMoveTo(fraction float64) {
	total := t.layout.totalPixels()
	if total <= 0 {
		return
	}
	target := max(0, int(fraction*float64(total)+0.5))
	t.topLine, t.topCharOffset = t.layout.lineAtPixel(target)
	t.clampScrollPosition()
	t.notifyYScrollbar()
	t.Display()
}

// YViewScroll scrolls by count units or pages.
func (t *TextWidget) YViewScroll(count int, pages bool) {
	if pages {
		visLines := max((t.Win.Height-2*t.insetY)/t.lineHeight(), 1)
		count *= visLines
	}

	t.scrollByDisplayLines(count)
	t.clampScrollPosition()
	t.notifyYScrollbar()
	t.Display()
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
func (t *TextWidget) Configure(opts ...TextOption) {
	widget.Configure(t, opts, t.computeGeometry)
}

// Destroy cleans up the text widget.
func (t *TextWidget) Destroy() {
	if t.Destroyed {
		return
	}
	t.Destroyed = true
	d := t.Win.Display.Server
	if t.pixmap != 0 {
		d.FreePixmap(t.pixmap)
		t.pixmap = 0
	}
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
func (t *TextWidget) EndIndex() string {
	n := t.doc.LineCount()
	if n == 0 {
		return "1.0"
	}
	c := len(t.doc.Lines[n-1].Text)
	return fmt.Sprintf("%d.%d", n, c)
}

// WindowCreate embeds a child window inline at the given text index.
// A placeholder character is inserted into the document so that text wraps
// around the window, and the position is tracked by a mark.
func (t *TextWidget) WindowCreate(indexStr string, w *window.Window) {
	t.WindowCreatePad(indexStr, w, 0, 0)
}

// WindowCreatePad is "window create" with -padx and -pady (Tk distances).
func (t *TextWidget) WindowCreatePad(indexStr string, w *window.Window, padX, padY any) {
	idx, ok := ParseIndex(t.doc, indexStr)
	if !ok {
		return
	}
	// Insert placeholder character into the document.
	t.doc.Insert(idx, string(runeEmbeddedWindow))
	// Create a left-gravity mark to track the placeholder position.
	t.ewSeq++
	markName := fmt.Sprintf("_ew%d", t.ewSeq)
	t.doc.MarkSet(markName, idx)
	if m, mok := t.doc.Marks[markName]; mok {
		m.Gravity = GravityLeft
	}
	t.embeddedWindows = append(t.embeddedWindows, embeddedWin{markName: markName, win: w,
		padX: screenunit.PxOr(padX, 0), padY: screenunit.PxOr(padY, 0)})
}

// ImageCreate embeds an image at the given text index, treating it as an inline element.
func (t *TextWidget) ImageCreate(indexStr string, img widget.WidgetImage) {
	idx, ok := ParseIndex(t.doc, indexStr)
	if !ok {
		return
	}
	t.embeddedImages = append(t.embeddedImages, embeddedImage{index: idx, img: img})
}

// lineHeightFor returns the display line height for the given logical line,
// taking into account any embedded images and windows on that line.
func (t *TextWidget) lineHeightFor(lineIdx int) int {
	h := t.lineHeight()
	for _, ei := range t.embeddedImages {
		if ei.index.Line == lineIdx {
			if imgH := ei.img.Height(); imgH > h {
				h = imgH
			}
		}
	}
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
			// Delete placeholder character and mark.
			if m, ok := t.doc.Marks[ew.markName]; ok {
				pos := m.Pos
				t.doc.Delete(pos, Index{Line: pos.Line, Char: pos.Char + 1})
				t.doc.MarkUnset(ew.markName)
			}
			t.embeddedWindows = append(t.embeddedWindows[:i], t.embeddedWindows[i+1:]...)
			t.Display()
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
func (t *TextWidget) SetPadX(n int) {
	t.Configure(PadXOpt(n))
}

// SetPadY sets vertical padding between the border and the text content.
func (t *TextWidget) SetPadY(n int) {
	t.Configure(PadYOpt(n))
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
			ew.win.Flags &^= window.FlagMapped
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

	// Check if index is above the visible area.
	if idx.Line < t.topLine || (idx.Line == t.topLine && t.topCharOffset > 0) {
		t.topLine = idx.Line
		t.topCharOffset = 0
		return
	}

	// Check if index is below the visible area.
	dlines := t.computeVisibleLines()
	if len(dlines) == 0 {
		t.topLine = idx.Line
		t.topCharOffset = 0
		return
	}

	lastDL := dlines[len(dlines)-1]
	lastIdx := Index{Line: lastDL.logicalLine, Char: lastDL.endChar}
	if Compare(idx, lastIdx) > 0 || (idx.Line == lastDL.logicalLine && idx.Char > lastDL.endChar) {
		// Need to scroll down.
		t.scrollDownToShow(idx)
	}
}

// scrollDownToShow scrolls down until idx is visible.
func (t *TextWidget) scrollDownToShow(idx Index) {
	availHeight := t.Win.Height - 2*t.insetY
	if availHeight <= 0 {
		return
	}

	// Jump close to the answer with the layout cache: the top that puts
	// idx's display line at the bottom of the view. The loop below then
	// settles it by at most a line or two instead of stepping from the
	// current top one display line at a time.
	if idx.Line >= 1 && idx.Line <= t.doc.LineCount() {
		ll := t.layout.line(idx.Line)
		bottom := t.layout.pixelsBefore(idx.Line)
		for k, h := range ll.hs {
			bottom += h
			if k < len(ll.dls) && idx.Char <= ll.dls[k].endChar {
				break
			}
		}
		l, off := t.layout.lineAtPixel(max(bottom-availHeight, 0))
		if t.computeDisplayLinesBefore(l, off) > t.computeDisplayLinesBefore(t.topLine, t.topCharOffset) {
			t.topLine, t.topCharOffset = l, off
		}
	}

	// Scroll down one display line at a time.
	for range 1000 { // safety limit
		dlines := t.computeVisibleLines()
		if len(dlines) == 0 {
			break
		}
		lastDL := dlines[len(dlines)-1]
		if idx.Line < lastDL.logicalLine ||
			(idx.Line == lastDL.logicalLine && idx.Char <= lastDL.endChar) {
			break
		}
		t.scrollByDisplayLines(1)
	}
}

// clampScrollPosition ensures the view doesn't scroll past the end of content.
// The last line of content should not scroll above the bottom of the viewport.
func (t *TextWidget) clampScrollPosition() {
	totalDL := t.totalDisplayLines()
	visLines := max((t.Win.Height-2*t.insetY)/t.lineHeight(), 1)
	if totalDL <= visLines {
		// All content fits — reset to top.
		t.topLine = 1
		t.topCharOffset = 0
		return
	}
	// Maximum top display line: totalDL - visLines.
	maxTopDL := totalDL - visLines
	topDL := t.computeDisplayLinesBefore(t.topLine, t.topCharOffset)
	if topDL > maxTopDL {
		t.topLine, t.topCharOffset = t.layout.lineAtDisplayLine(maxTopDL)
	}
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
}

// notifyYScrollbar ports GetYView (tkTextDisp.c): the fractions are pixel
// counts, the pixels above the top display line and those shown up to the
// bottom of the text area, over the pixel height of the whole text.
func (t *TextWidget) notifyYScrollbar() {
	if t.YScrollCmd == nil {
		return
	}
	t.yScrollPending = true
	t.scheduleRedraw()
}

func (t *TextWidget) yviewFractions() (float64, float64) {
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
