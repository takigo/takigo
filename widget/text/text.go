package text

import (
	"fmt"

	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// color.ColorRef stores pre-resolved color components.

// embeddedWin records a window embedded at a text index position.
type embeddedWin struct {
	index Index
	win   *window.Window
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

	// Colors.
	selBg       *color.ColorRef
	selFg       *color.ColorRef
	insertColor *color.ColorRef

	// Selection.
	selAnchor Index // fixed end during selection drag

	// Scrollbar callbacks.
	YScrollCmd func(first, last float64)
	XScrollCmd func(first, last float64)

	// Undo.
	undoEnabled bool
	undoStack   *UndoStack

	// Read-only mode: navigation/selection work, editing blocked.
	readOnly bool

	// setGrid: if true, the toplevel window's resize increment is kept at the
	// character cell size so the window resizes in whole-character steps.
	setGrid bool

	// Tag event bindings: tagName → eventName → handlers.
	tagBindings map[string]map[string][]func()
	hoverTags   map[string]bool

	// Embedded windows (overlay-positioned child windows).
	embeddedWindows []embeddedWin

	// Offscreen pixmap.
	pixmap           platform.PixmapID
	pixmapW, pixmapH int
	displayValid     bool
	redrawPending    bool
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

	// Text widget defaults.
	t.BorderWidth = 2
	t.Relief = option.ReliefSunken
	t.HighlightWidth = 1

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
	selTag := &Tag{Name: "sel", Priority: 1000}
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
	t.doc.Tags["sel"] = selTag

	t.undoStack = NewUndoStack(100)

	for _, opt := range opts {
		opt(t)
	}

	t.inset = t.BorderWidth + t.HighlightWidth + 1
	t.computeGeometry()

	if t.Background != nil {
		w.BackgroundPixel = t.Background.Pixel
	}

	w.Flags |= window.FlagFocusable
	w.SetCursor(152) // XC_xterm — I-beam cursor for text
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

	t.BorderWidth = 2
	t.Relief = option.ReliefSunken
	t.HighlightWidth = 1

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

	for _, opt := range opts {
		opt(t)
	}

	t.inset = t.BorderWidth + t.HighlightWidth + 1
	t.computeGeometry()

	if t.Background != nil {
		w.BackgroundPixel = t.Background.Pixel
	}

	w.Flags |= window.FlagFocusable
	w.SetCursor(152)
	bindText(t, app)

	// Register as a document listener so edits from other peers trigger a redraw.
	doc.Listeners = append(doc.Listeners, func() {
		t.notifyYScrollbar()
		t.scheduleRedraw()
	})

	return t
}

// computeGeometry calculates the requested window size.
func (t *TextWidget) computeGeometry() {
	if t.Font == nil {
		return
	}
	m := t.Font.Metrics()
	avgWidth := t.Font.MeasureString("0")
	if avgWidth < 1 {
		avgWidth = 1
	}
	lineHeight := m.Linespace()

	w := t.Win
	w.ReqWidth = t.prefWidth*avgWidth + 2*t.inset
	w.ReqHeight = t.prefHeight*lineHeight + 2*t.inset

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

// Display draws the text widget.
func (t *TextWidget) Display() {
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

	// Render content to pixmap.
	t.renderToPixmap()

	// Copy pixmap to window.
	d.CopyArea(platform.PixmapDrawable(t.pixmap), w.Drawable(), gc,
		0, 0, uint(winW), uint(winH), 0, 0)

	// Position any embedded windows.
	if len(t.embeddedWindows) > 0 {
		dlines := t.computeVisibleLines()
		t.positionEmbeddedWindows(dlines)
	}

	// Draw border on top.
	if t.Border != nil && t.BorderWidth > 0 {
		draw.Draw3DRectangle(d, w.Drawable(), gc, t.Border,
			0, 0, w.Width, w.Height, t.BorderWidth, t.Relief)
	}

	d.Flush()
	t.redrawPending = false
}

// scheduleRedraw schedules a redraw via the idle loop.
func (t *TextWidget) scheduleRedraw() {
	if t.redrawPending || t.Destroyed {
		return
	}
	t.redrawPending = true
	t.App.DoWhenIdle(func() {
		if !t.Destroyed {
			t.Display()
		}
	})
}

// --- Public API ---

// Insert inserts text at the given index string.
func (t *TextWidget) Insert(index, txt string) {
	idx, ok := ParseIndex(t.doc, index)
	if !ok {
		return
	}
	txt = expandTabs(txt)
	endIdx := t.doc.Insert(idx, txt)
	if t.undoEnabled {
		t.undoStack.RecordInsert(idx, endIdx, txt)
	}
	t.seeInsert()
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
	t.seeInsert()
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

// TagAdd adds a tag to the given range.
func (t *TextWidget) TagAdd(tagName, startIndex, endIndex string) {
	start, ok1 := ParseIndex(t.doc, startIndex)
	end, ok2 := ParseIndex(t.doc, endIndex)
	if !ok1 || !ok2 {
		return
	}
	t.doc.TagAdd(tagName, start, end)
}

// TagRemove removes a tag from the given range.
func (t *TextWidget) TagRemove(tagName, startIndex, endIndex string) {
	start, ok1 := ParseIndex(t.doc, startIndex)
	end, ok2 := ParseIndex(t.doc, endIndex)
	if !ok1 || !ok2 {
		return
	}
	t.doc.TagRemove(tagName, start, end)
}

// TagConfigure configures a tag's display attributes.
func (t *TextWidget) TagConfigure(tagName string, opts ...TagOption) {
	t.doc.TagConfigure(tagName, t.App.ColorCache(), t.App.FontRegistry(), opts...)
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
	t.wrapMode = mode
	t.Display()
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
	t.notifyYScrollbar()
	t.Display()
}

// YViewMoveTo scrolls to a fraction of the total content height.
func (t *TextWidget) YViewMoveTo(fraction float64) {
	totalDL := t.totalDisplayLines()
	if totalDL <= 0 {
		return
	}
	targetDL := int(fraction*float64(totalDL) + 0.5)
	if targetDL < 0 {
		targetDL = 0
	}

	// Walk through lines to find the logical line and display-line offset.
	availWidth := t.Win.Width - 2*t.inset
	dlCount := 0
	for l := 1; l <= t.doc.LineCount(); l++ {
		p := t.resolveLineProps(l)
		dls := t.wrapLine(l, availWidth, p.lm1, p.lm2, p.rm)
		if dlCount+len(dls) > targetDL {
			t.topLine = l
			t.topCharOffset = targetDL - dlCount
			t.notifyYScrollbar()
			t.Display()
			return
		}
		dlCount += len(dls)
	}
	// Past end.
	t.topLine = t.doc.LineCount()
	t.topCharOffset = 0
	t.notifyYScrollbar()
	t.Display()
}

// YViewScroll scrolls by count units or pages.
func (t *TextWidget) YViewScroll(count int, pages bool) {
	if pages {
		visLines := (t.Win.Height - 2*t.inset) / t.lineHeight()
		if visLines < 1 {
			visLines = 1
		}
		count *= visLines
	}

	t.scrollByDisplayLines(count)
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
	t.xOffset = int(fraction * float64(maxW))
	if t.xOffset < 0 {
		t.xOffset = 0
	}
	t.notifyXScrollbar()
	t.Display()
}

// XViewScroll scrolls horizontally.
func (t *TextWidget) XViewScroll(count int, pages bool) {
	if pages {
		availW := t.Win.Width - 2*t.inset
		count *= availW
	} else {
		count *= t.Font.MeasureString("0")
	}
	t.XView(t.xOffset + count)
}

// TagBind binds an event handler to a text tag. Supported events:
// "<Enter>", "<Leave>", "<Button-1>".
func (t *TextWidget) TagBind(tagName, eventName string, handler func()) {
	if t.tagBindings == nil {
		t.tagBindings = make(map[string]map[string][]func())
	}
	if t.tagBindings[tagName] == nil {
		t.tagBindings[tagName] = make(map[string][]func())
	}
	t.tagBindings[tagName][eventName] = append(t.tagBindings[tagName][eventName], handler)
}

// tagsAtIndex returns the set of tag names that cover the given index.
func (t *TextWidget) tagsAtIndex(idx Index) map[string]bool {
	result := make(map[string]bool)
	for _, tr := range t.doc.TagRanges {
		if Compare(idx, tr.Start) >= 0 && Compare(idx, tr.End) < 0 {
			result[tr.TagName] = true
		}
	}
	return result
}

// fireTagHandlers fires all handlers for (tagName, eventName).
func (t *TextWidget) fireTagHandlers(tagName, eventName string) {
	if t.tagBindings == nil {
		return
	}
	if handlers, ok := t.tagBindings[tagName]; ok {
		for _, h := range handlers[eventName] {
			h()
		}
	}
}

// updateTagHover computes Enter/Leave events as the mouse moves over tags.
func (t *TextWidget) updateTagHover(newTags map[string]bool) {
	if t.hoverTags == nil {
		t.hoverTags = make(map[string]bool)
	}
	for tag := range t.hoverTags {
		if !newTags[tag] {
			t.fireTagHandlers(tag, "<Leave>")
		}
	}
	for tag := range newTags {
		if !t.hoverTags[tag] {
			t.fireTagHandlers(tag, "<Enter>")
		}
	}
	t.hoverTags = newTags
}

// Configure applies options.
func (t *TextWidget) Configure(opts ...option.Option) {
	option.Apply(t, opts)
	t.UpdateBorder()
	t.inset = t.BorderWidth + t.HighlightWidth + 1
	t.computeGeometry()
	if t.Background != nil {
		t.Win.BackgroundPixel = t.Background.Pixel
	}
	t.Display()
}

// Destroy cleans up the text widget.
func (t *TextWidget) Destroy() {
	if t.Destroyed {
		return
	}
	t.Destroyed = true
	if t.pixmap != 0 {
		t.Win.Display.Server.FreePixmap(t.pixmap)
		t.pixmap = 0
	}
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

// WindowCreate registers a child window to be positioned at the given text index.
// The window is placed as an overlay at the Y of the line containing that index.
func (t *TextWidget) WindowCreate(indexStr string, w *window.Window) {
	idx, ok := ParseIndex(t.doc, indexStr)
	if !ok {
		return
	}
	t.embeddedWindows = append(t.embeddedWindows, embeddedWin{index: idx, win: w})
}

// positionEmbeddedWindows moves embedded windows to their text positions.
func (t *TextWidget) positionEmbeddedWindows(dlines []displayLine) {
	d := t.Win.Display.Server
	for _, ew := range t.embeddedWindows {
		visible := false
		for _, dl := range dlines {
			if dl.logicalLine != ew.index.Line {
				continue
			}
			pixelY := t.inset + dl.y
			wx := t.inset + dl.leftMargin
			wy := pixelY
			ww := ew.win.ReqWidth
			wh := ew.win.ReqHeight
			if wh == 0 {
				wh = dl.height
			}
			d.MoveResizeWindow(ew.win.PlatformID, wx, wy, uint(ww), uint(wh))
			d.MapWindow(ew.win.PlatformID)
			visible = true
			break
		}
		if !visible {
			d.UnmapWindow(ew.win.PlatformID)
		}
	}
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
	availHeight := t.Win.Height - 2*t.inset
	if availHeight <= 0 {
		return
	}

	// Scroll down one display line at a time.
	for i := 0; i < 1000; i++ { // safety limit
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

// scrollByDisplayLines scrolls by n display lines (positive = down, negative = up).
func (t *TextWidget) scrollByDisplayLines(n int) {
	availWidth := t.Win.Width - 2*t.inset

	if n > 0 {
		// Scroll down.
		for i := 0; i < n; i++ {
			p := t.resolveLineProps(t.topLine)
			dls := t.wrapLine(t.topLine, availWidth, p.lm1, p.lm2, p.rm)
			if t.topCharOffset+1 < len(dls) {
				t.topCharOffset++
			} else if t.topLine < t.doc.LineCount() {
				t.topLine++
				t.topCharOffset = 0
			} else {
				break
			}
		}
	} else {
		// Scroll up.
		for i := 0; i < -n; i++ {
			if t.topCharOffset > 0 {
				t.topCharOffset--
			} else if t.topLine > 1 {
				t.topLine--
				p := t.resolveLineProps(t.topLine)
				dls := t.wrapLine(t.topLine, availWidth, p.lm1, p.lm2, p.rm)
				t.topCharOffset = len(dls) - 1
			} else {
				break
			}
		}
	}
}

// notifyYScrollbar calls the Y scroll callback.
func (t *TextWidget) notifyYScrollbar() {
	if t.YScrollCmd == nil {
		return
	}
	totalDL := t.totalDisplayLines()
	if totalDL <= 0 {
		t.YScrollCmd(0, 1)
		return
	}

	topDL := t.computeDisplayLinesBefore(t.topLine, t.topCharOffset)
	visLines := (t.Win.Height - 2*t.inset) / t.lineHeight()
	if visLines < 1 {
		visLines = 1
	}

	first := float64(topDL) / float64(totalDL)
	last := float64(topDL+visLines) / float64(totalDL)
	if first < 0 {
		first = 0
	}
	if last > 1 {
		last = 1
	}
	t.YScrollCmd(first, last)
}

// notifyXScrollbar calls the X scroll callback.
func (t *TextWidget) notifyXScrollbar() {
	if t.XScrollCmd == nil {
		return
	}
	maxW := t.estimateMaxLineWidth()
	availW := t.Win.Width - 2*t.inset
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
	maxW := 0
	for _, line := range t.doc.Lines {
		if len(line.Text) > 0 {
			w := t.Font.MeasureString(string(line.Text))
			if w > maxW {
				maxW = w
			}
		}
	}
	return maxW
}

// Suppress unused import warnings.
var _ = (*color.Color)(nil)
var _ = (*font.Registry)(nil)
