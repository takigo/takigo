// Package listbox implements a listbox widget for displaying a list of
// selectable text items. It ports tk/generic/tkListbox.c.
package listbox

import (
	"github.com/takigo/takigo/color"
	"github.com/takigo/takigo/draw"
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/window"
)

// SelectMode defines how items are selected.
type SelectMode int

const (
	SelectSingle   SelectMode = iota // click selects one item
	SelectBrowse                     // click selects, drag tracks
	SelectMultiple                   // click toggles individual items
	SelectExtended                   // click+shift/ctrl extends
)

// Listbox displays a scrollable list of text items.
type Listbox struct {
	widget.Base

	items       []string
	topIndex    int // first visible item index
	xOffset     int // horizontal scroll offset in pixels
	activeIndex int // active (focused) item, -1 = none

	// Selection.
	selected   map[int]bool
	selectMode SelectMode
	selAnchor  int // selection anchor for extended mode

	// Layout.
	lineHeight int
	inset      int

	// Widest item in pixels (Tk's listPtr->maxWidth), valid while
	// maxWidthOK is set and the font is maxWidthFont.
	maxWidthPx   int
	maxWidthOK   bool
	maxWidthFont font.Font

	// Preferred dimensions.
	PrefWidth  int // in characters
	PrefHeight int // in lines

	// Colors.
	SelBg *color.ColorRef
	SelFg *color.ColorRef

	// Per-item colors (set via ItemConfigure).
	itemFg map[int]*color.ColorRef
	itemBg map[int]*color.ColorRef

	// Scrollbar callbacks.
	YScrollCmd func(first, last float64)
	XScrollCmd func(first, last float64)

	// SelectCmd is called whenever the selection changes.
	SelectCmd func()

	HasFocus bool

	// Text justification within items (left/center/right).
	Justify option.Justify
}

// ListboxOption configures a Listbox.
type ListboxOption func(*Listbox)

func Items(items ...string) ListboxOption {
	return func(lb *Listbox) {
		lb.items = append([]string{}, items...)
		lb.maxWidthOK = false
	}
}
func SelectModeOpt(m SelectMode) ListboxOption {
	return func(lb *Listbox) { lb.selectMode = m }
}
func Width(w int) ListboxOption  { return func(lb *Listbox) { lb.PrefWidth = w } }
func Height(h int) ListboxOption { return func(lb *Listbox) { lb.PrefHeight = h } }
func JustifyOpt(j option.Justify) ListboxOption {
	return func(lb *Listbox) { lb.Justify = j }
}

// Background sets the background colour.
func Background[C color.Spec](name C) ListboxOption {
	return func(lb *Listbox) { lb.SetBackgroundColor(name) }
}

// Foreground sets the text colour.
func Foreground[C color.Spec](name C) ListboxOption {
	return func(lb *Listbox) { lb.SetForegroundColor(name) }
}

func YScrollCommand(fn func(float64, float64)) ListboxOption {
	return func(lb *Listbox) { lb.YScrollCmd = fn }
}

func XScrollCommand(fn func(float64, float64)) ListboxOption {
	return func(lb *Listbox) { lb.XScrollCmd = fn }
}

// New creates a new Listbox widget.
func New(parent widget.Caregiver, name string, opts ...ListboxOption) *Listbox {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(w)

	lb := &Listbox{
		selected:    make(map[int]bool),
		itemFg:      make(map[int]*color.ColorRef),
		itemBg:      make(map[int]*color.ColorRef),
		selectMode:  SelectBrowse,
		activeIndex: -1,
		selAnchor:   -1,
		PrefWidth:   20,
		PrefHeight:  10,
	}
	widget.InitBase(&lb.Base, w, app)
	lb.SetDisplayProc(lb.display)
	w.Class = "Listbox"

	lb.BorderWidth = widget.DefBorderWidth
	lb.Relief = option.ReliefSunken
	lb.HighlightWidth = 1

	// White background for listbox.
	if bg, err := app.ColorCache().Get(widget.PaletteFor(app).FieldBackground); err == nil {
		lb.Background = bg
		lb.UpdateBorder()
	}

	// Selection colors.
	if sel, err := app.ColorCache().Get(widget.PaletteFor(app).SelectBackground); err == nil {
		lb.SelBg = sel.Ref()
	}
	if selfg, err := app.ColorCache().Get(widget.PaletteFor(app).SelectForeground); err == nil {
		lb.SelFg = selfg.Ref()
	}

	for _, opt := range opts {
		opt(lb)
	}

	lb.computeGeometry()

	if lb.Background != nil {
		w.SetBackgroundPixel(lb.Background.Pixel)
	}

	w.Flags |= window.FlagFocusable
	bindListbox(lb, app)
	return lb
}

// ItemCount returns the number of items.
func (lb *Listbox) ItemCount() int {
	return len(lb.items)
}

// GetItems returns a copy of all items.
func (lb *Listbox) GetItems() []string {
	out := make([]string, len(lb.items))
	copy(out, lb.items)
	return out
}

// Insert inserts items at the given index.
func (lb *Listbox) Insert(index int, items ...string) {
	if index < 0 {
		index = 0
	}
	if index > len(lb.items) {
		index = len(lb.items)
	}

	newItems := make([]string, 0, len(lb.items)+len(items))
	newItems = append(newItems, lb.items[:index]...)
	newItems = append(newItems, items...)
	newItems = append(newItems, lb.items[index:]...)
	lb.items = newItems
	if lb.maxWidthOK && lb.Font == lb.maxWidthFont {
		for _, it := range items {
			lb.maxWidthPx = max(lb.maxWidthPx, lb.Font.MeasureString(it))
		}
	}

	// InsertEls: indices at or after index move down.
	n := len(items)
	shift := func(i int) (int, bool) {
		if i >= index {
			return i + n, true
		}
		return i, true
	}
	lb.selected = remapKeys(lb.selected, shift)
	lb.itemFg = remapKeys(lb.itemFg, shift)
	lb.itemBg = remapKeys(lb.itemBg, shift)
	if index <= lb.selAnchor {
		lb.selAnchor += n
	}
	if index < lb.topIndex {
		lb.topIndex += n
	}
	if lb.activeIndex >= 0 && index <= lb.activeIndex {
		lb.activeIndex += n
	}

	lb.notifyYScrollbar()
	lb.Display()
}

// Delete deletes items in the range [first, last].
func (lb *Listbox) Delete(first, last int) {
	if first < 0 {
		first = 0
	}
	if last >= len(lb.items) {
		last = len(lb.items) - 1
	}
	if first > last {
		return
	}

	count := last - first + 1
	if lb.maxWidthOK && lb.Font == lb.maxWidthFont {
		for _, it := range lb.items[first : last+1] {
			if lb.Font.MeasureString(it) >= lb.maxWidthPx {
				lb.maxWidthOK = false
				break
			}
		}
	}
	lb.items = append(lb.items[:first], lb.items[last+1:]...)

	// DeleteEls: deleted indices go, later ones move up.
	shift := func(i int) (int, bool) {
		switch {
		case i > last:
			return i - count, true
		case i >= first:
			return 0, false
		}
		return i, true
	}
	lb.selected = remapKeys(lb.selected, shift)
	lb.itemFg = remapKeys(lb.itemFg, shift)
	lb.itemBg = remapKeys(lb.itemBg, shift)
	switch {
	case last < lb.selAnchor:
		lb.selAnchor -= count
	case first <= lb.selAnchor:
		lb.selAnchor = first
	}
	switch {
	case last < lb.topIndex:
		lb.topIndex -= count
	case first < lb.topIndex:
		lb.topIndex = first
	}
	switch {
	case last < lb.activeIndex:
		lb.activeIndex -= count
	case first <= lb.activeIndex:
		lb.activeIndex = first
	}
	if lb.activeIndex >= len(lb.items) {
		lb.activeIndex = len(lb.items) - 1
	}
	if lb.topIndex > 0 && lb.topIndex >= len(lb.items) {
		lb.topIndex = max(len(lb.items)-1, 0)
	}

	lb.notifyYScrollbar()
	lb.Display()
}

// Selection returns the indices of selected items.
func (lb *Listbox) Selection() []int {
	var result []int
	for i := 0; i < len(lb.items); i++ {
		if lb.selected[i] {
			result = append(result, i)
		}
	}
	return result
}

// SelectionSet selects items in the range [first, last].
func (lb *Listbox) SelectionSet(first, last int) {
	if first < 0 {
		first = 0
	}
	if last >= len(lb.items) {
		last = len(lb.items) - 1
	}
	for i := first; i <= last; i++ {
		lb.selected[i] = true
	}
}

// SelectionClear clears selection in range [first, last].
func (lb *Listbox) SelectionClear(first, last int) {
	for i := first; i <= last; i++ {
		delete(lb.selected, i)
	}
}

// ItemConfigure sets per-item foreground and/or background colors.
// Pass empty string to clear a per-item color (revert to default).
func (lb *Listbox) ItemConfigure(idx int, fg, bg string) {
	if fg == "" {
		delete(lb.itemFg, idx)
	} else if col, err := lb.App.ColorCache().Get(fg); err == nil {
		lb.itemFg[idx] = col.Ref()
	}
	if bg == "" {
		delete(lb.itemBg, idx)
	} else if col, err := lb.App.ColorCache().Get(bg); err == nil {
		lb.itemBg[idx] = col.Ref()
	}
}

// SetJustify changes the text justification and redraws.
func (lb *Listbox) SetJustify(j option.Justify) {
	lb.Configure(JustifyOpt(j))
}

// See scrolls the listbox so that the item at index is visible.
func (lb *Listbox) See(index int) {
	if index < 0 {
		index = 0
	}
	if index >= len(lb.items) {
		index = len(lb.items) - 1
	}
	if index < 0 {
		return
	}

	visLines := lb.visibleLines()
	if index < lb.topIndex {
		lb.topIndex = index
	} else if index >= lb.topIndex+visLines {
		lb.topIndex = index - visLines + 1
	}
	lb.notifyYScrollbar()
}

// YView scrolls to the given index.
func (lb *Listbox) YView(index int) {
	if index < 0 {
		index = 0
	}
	max := len(lb.items) - lb.visibleLines()
	if max < 0 {
		max = 0
	}
	if index > max {
		index = max
	}
	lb.topIndex = index
	lb.notifyYScrollbar()
	lb.Display()
}

// YViewScroll scrolls by count units or pages.
func (lb *Listbox) YViewScroll(count int, pages bool) {
	// ListboxYviewSubCmd: a page is fullLines-2 rows, or one row when the
	// list shows two or fewer.
	if vis := lb.visibleLines(); pages && vis > 2 {
		count *= vis - 2
	}
	lb.YView(lb.topIndex + count)
}

// YViewMoveTo scrolls to a fraction of the list.
func (lb *Listbox) YViewMoveTo(fraction float64) {
	index := int(fraction*float64(len(lb.items)) + 0.5)
	lb.YView(index)
}

// YVisibleRange returns the visible fraction for the y scrollbar.
func (lb *Listbox) YVisibleRange() (float64, float64) {
	n := len(lb.items)
	if n == 0 {
		return 0, 1
	}
	first := float64(lb.topIndex) / float64(n)
	vis := lb.visibleLines()
	last := float64(lb.topIndex+vis) / float64(n)
	if last > 1 {
		last = 1
	}
	return first, last
}

func (lb *Listbox) visibleLines() int {
	if lb.lineHeight <= 0 {
		return 1
	}
	avail := lb.Win.Height - 2*lb.inset
	if avail < 1 {
		return 1
	}
	return avail / lb.lineHeight
}

func (lb *Listbox) indexAtY(y int) int {
	if lb.lineHeight <= 0 {
		return lb.topIndex
	}
	line := max((y-lb.inset)/lb.lineHeight, 0)
	idx := lb.topIndex + line
	if idx >= len(lb.items) {
		idx = len(lb.items) - 1
	}
	if idx < 0 {
		idx = 0
	}
	return idx
}

func (lb *Listbox) computeGeometry() {
	if lb.Font == nil {
		return
	}
	m := lb.Font.Metrics()
	// ListboxComputeGeometry (tkListbox.c): -selectborderwidth is 0 here.
	lb.lineHeight = m.Linespace() + 1
	lb.inset = lb.BorderWidth + lb.HighlightWidth

	avgW := max(lb.Font.MeasureString("0"), 1)
	w := lb.Win
	w.ReqWidth = lb.PrefWidth*avgW + 2*lb.inset
	w.ReqHeight = lb.PrefHeight*lb.lineHeight + 2*lb.inset
}

func (lb *Listbox) notifyYScrollbar() {
	if lb.YScrollCmd != nil {
		first, last := lb.YVisibleRange()
		lb.YScrollCmd(first, last)
	}
	lb.notifyXScrollbar()
}

// maxWidth is Tk's listPtr->maxWidth: the widest element in pixels. It is
// kept up to date by Insert and recomputed only after the widest element
// is deleted or the font changes.
func (lb *Listbox) maxWidth() int {
	if lb.Font == nil {
		return 0
	}
	if !lb.maxWidthOK || lb.Font != lb.maxWidthFont {
		mw := 0
		for _, it := range lb.items {
			mw = max(mw, lb.Font.MeasureString(it))
		}
		lb.maxWidthPx, lb.maxWidthOK, lb.maxWidthFont = mw, true, lb.Font
	}
	return lb.maxWidthPx
}

// remapKeys moves the entries of an index-keyed map by fn, dropping those
// for which it reports false.
func remapKeys[V any](m map[int]V, fn func(int) (int, bool)) map[int]V {
	if len(m) == 0 {
		return m
	}
	out := make(map[int]V, len(m))
	for i, v := range m {
		if j, ok := fn(i); ok {
			out[j] = v
		}
	}
	return out
}

func (lb *Listbox) xScrollUnit() int {
	if lb.Font == nil {
		return 1
	}
	return max(1, lb.Font.MeasureString("0"))
}

// maxOffset ports GetMaxOffset (-selectborderwidth is 0 here).
func (lb *Listbox) maxOffset() int {
	u := lb.xScrollUnit()
	m := max(0, lb.maxWidth()-(lb.Win.Width-2*lb.inset)+u-1)
	return m - m%u
}

// XVisibleRange ports "xview" with no arguments.
func (lb *Listbox) XVisibleRange() (float64, float64) {
	mw := lb.maxWidth()
	if mw == 0 {
		return 0, 1
	}
	ww := lb.Win.Width - 2*lb.inset
	return float64(lb.xOffset) / float64(mw), min(1, float64(lb.xOffset+ww)/float64(mw))
}

// changeOffset ports ChangeListboxOffset: round to whole units and clamp.
func (lb *Listbox) changeOffset(offset int) {
	u := lb.xScrollUnit()
	offset += u / 2
	offset = max(0, min(lb.maxOffset(), offset))
	offset -= offset % u
	if offset != lb.xOffset {
		lb.xOffset = offset
		lb.notifyXScrollbar()
		lb.Display()
	}
}

// XView scrolls so that character column index is at the left edge.
func (lb *Listbox) XView(index int) { lb.changeOffset(index * lb.xScrollUnit()) }

// XViewMoveTo ports "xview moveto".
func (lb *Listbox) XViewMoveTo(fraction float64) {
	lb.changeOffset(int(fraction*float64(lb.maxWidth()) + 0.5))
}

// XViewScroll ports "xview scroll n units|pages".
func (lb *Listbox) XViewScroll(n int, pages bool) {
	u := lb.xScrollUnit()
	if pages {
		if wu := (lb.Win.Width - 2*lb.inset) / u; wu > 2 {
			lb.changeOffset(lb.xOffset + n*u*(wu-2))
			return
		}
	}
	lb.changeOffset(lb.xOffset + n*u)
}

func (lb *Listbox) notifyXScrollbar() {
	if lb.XScrollCmd != nil {
		lb.XScrollCmd(lb.XVisibleRange())
	}
}

// Display schedules a redraw at idle time; see widget.Base.EventuallyRedraw.
func (lb *Listbox) Display() {
	lb.EventuallyRedraw()
}

// display draws the listbox.
func (lb *Listbox) display() {
	if lb.Destroyed {
		return
	}
	w := lb.Win
	if w.PlatformID == platform.WindowID(0) {
		return
	}

	d := w.Display.Server
	gc := w.GC

	// Background.
	if lb.Background != nil {
		d.SetForeground(gc, lb.Background.Pixel)
	}
	d.FillRectangle(w.Drawable(), gc, 0, 0, uint(w.Width), uint(w.Height))

	// DisplayListbox draws the border and highlight ring last, over any
	// text that runs past the viewable area.
	defer func() {
		if lb.Border != nil && lb.BorderWidth > 0 {
			hl := lb.HighlightWidth
			draw.Draw3DRectangle(d, w.Drawable(), gc, lb.Border,
				hl, hl, w.Width-2*hl, w.Height-2*hl, lb.BorderWidth, lb.Relief)
		}
		lb.DrawHighlightBorder(lb.HasFocus, 0)
	}()

	if lb.Font == nil || lb.lineHeight <= 0 {
		return
	}

	df, isDF := lb.Font.(platform.DrawableFont)
	if !isDF {
		return
	}

	m := lb.Font.Metrics()
	visLines := lb.visibleLines()
	clipRight := w.Width - lb.inset

	for i := range visLines {
		itemIdx := lb.topIndex + i
		if itemIdx >= len(lb.items) {
			break
		}

		rowY := lb.inset + i*lb.lineHeight
		text := lb.items[itemIdx]
		textW := lb.Font.MeasureString(text)
		// DisplayListbox text placement.
		var textX int
		switch lb.Justify {
		case option.JustifyCenter:
			textX = (w.Width-textW)/2 - lb.xOffset + lb.maxOffset()/2
		case option.JustifyRight:
			textX = clipRight - textW - lb.xOffset + lb.maxOffset()
		default: // JustifyLeft
			textX = lb.inset - lb.xOffset
		}
		textY := rowY + m.Ascent

		isSelected := lb.selected[itemIdx]

		// Determine effective item colors (per-item overrides default).
		itemBg := lb.itemBg[itemIdx]
		itemFg := lb.itemFg[itemIdx]

		// Selection highlight (selection overrides per-item bg).
		if isSelected && lb.SelBg != nil {
			d.SetForeground(gc, lb.SelBg.Pixel)
			d.FillRectangle(w.Drawable(), gc, lb.inset, rowY,
				uint(clipRight-lb.inset), uint(lb.lineHeight))
		} else if itemBg != nil {
			d.SetForeground(gc, itemBg.Pixel)
			d.FillRectangle(w.Drawable(), gc, lb.inset, rowY,
				uint(clipRight-lb.inset), uint(lb.lineHeight))
		}

		// Text.
		if isSelected && lb.SelFg != nil {
			df.DrawString(w.Drawable(), textX, textY, text,
				lb.SelFg.Pixel, lb.SelFg.Red, lb.SelFg.Green, lb.SelFg.Blue)
		} else if itemFg != nil {
			df.DrawString(w.Drawable(), textX, textY, text,
				itemFg.Pixel, itemFg.Red, itemFg.Green, itemFg.Blue)
		} else if lb.Foreground != nil {
			df.DrawString(w.Drawable(), textX, textY, text,
				lb.Foreground.Pixel, lb.Foreground.Red, lb.Foreground.Green, lb.Foreground.Blue)
		}

		// Active item indicator (underline when focused).
		if lb.HasFocus && itemIdx == lb.activeIndex && lb.Foreground != nil {
			lineY := textY + m.Descent - 1
			d.SetForeground(gc, lb.Foreground.Pixel)
			d.DrawLine(w.Drawable(), gc, textX, lineY, textX+textW, lineY)
		}
	}
}

// Configure applies options.
func (lb *Listbox) Configure(opts ...ListboxOption) error {
	return widget.Configure(lb, opts, lb.computeGeometry)
}

// Destroy cleans up the listbox.
func (lb *Listbox) Destroy() {
	if lb.Destroyed {
		return
	}
	lb.Destroyed = true
	window.DestroyWindow(lb.Win)
}
