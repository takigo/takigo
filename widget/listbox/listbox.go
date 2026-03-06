// Package listbox implements a listbox widget for displaying a list of
// selectable text items. It ports tk/generic/tkListbox.c.
package listbox

import (
	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
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
	return func(lb *Listbox) { lb.items = append([]string{}, items...) }
}
func SelectModeOpt(m SelectMode) ListboxOption {
	return func(lb *Listbox) { lb.selectMode = m }
}
func Width(w int) ListboxOption  { return func(lb *Listbox) { lb.PrefWidth = w } }
func Height(h int) ListboxOption { return func(lb *Listbox) { lb.PrefHeight = h } }
func JustifyOpt(j option.Justify) ListboxOption {
	return func(lb *Listbox) { lb.Justify = j }
}

func Background(name string) ListboxOption {
	return func(lb *Listbox) {
		col, err := lb.App.ColorCache().Get(name)
		if err == nil {
			lb.Background = col
			lb.UpdateBorder()
		}
	}
}

func Foreground(name string) ListboxOption {
	return func(lb *Listbox) {
		col, err := lb.App.ColorCache().Get(name)
		if err == nil {
			lb.Foreground = col
		}
	}
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

	lb.BorderWidth = 2
	lb.Relief = option.ReliefSunken
	lb.HighlightWidth = 1

	// White background for listbox.
	if bg, err := app.ColorCache().Get("#ffffff"); err == nil {
		lb.Background = bg
		lb.UpdateBorder()
	}

	// Selection colors.
	if sel, err := app.ColorCache().Get("#3399ff"); err == nil {
		lb.SelBg = sel.Ref()
	}
	if selfg, err := app.ColorCache().Get("#ffffff"); err == nil {
		lb.SelFg = selfg.Ref()
	}

	for _, opt := range opts {
		opt(lb)
	}

	lb.computeGeometry()

	if lb.Background != nil {
		w.BackgroundPixel = lb.Background.Pixel
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

	// Adjust selection indices.
	newSel := make(map[int]bool)
	for idx := range lb.selected {
		if idx >= index {
			newSel[idx+len(items)] = true
		} else {
			newSel[idx] = true
		}
	}
	lb.selected = newSel

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
	lb.items = append(lb.items[:first], lb.items[last+1:]...)

	// Adjust selection.
	newSel := make(map[int]bool)
	for idx := range lb.selected {
		if idx >= first && idx <= last {
			continue
		}
		if idx > last {
			newSel[idx-count] = true
		} else {
			newSel[idx] = true
		}
	}
	lb.selected = newSel

	if lb.activeIndex >= len(lb.items) {
		lb.activeIndex = len(lb.items) - 1
	}
	if lb.topIndex > 0 && lb.topIndex >= len(lb.items) {
		lb.topIndex = len(lb.items) - 1
		if lb.topIndex < 0 {
			lb.topIndex = 0
		}
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
	lb.Justify = j
	lb.Display()
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
	if pages {
		vis := lb.visibleLines()
		if vis < 1 {
			vis = 1
		}
		count *= vis - 2
		if count == 0 {
			if count > 0 {
				count = 1
			} else {
				count = -1
			}
		}
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
	line := (y - lb.inset) / lb.lineHeight
	if line < 0 {
		line = 0
	}
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
	lb.lineHeight = m.Linespace() + 2
	lb.inset = lb.BorderWidth + lb.HighlightWidth + 1

	avgW := lb.Font.MeasureString("0")
	if avgW < 1 {
		avgW = 1
	}
	w := lb.Win
	w.ReqWidth = lb.PrefWidth*avgW + 2*lb.inset
	w.ReqHeight = lb.PrefHeight*lb.lineHeight + 2*lb.inset
}

func (lb *Listbox) notifyYScrollbar() {
	if lb.YScrollCmd != nil {
		first, last := lb.YVisibleRange()
		lb.YScrollCmd(first, last)
	}
}

// Display draws the listbox.
func (lb *Listbox) Display() {
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

	// Border.
	if lb.Border != nil && lb.BorderWidth > 0 {
		draw.Draw3DRectangle(d, w.Drawable(), gc, lb.Border,
			0, 0, w.Width, w.Height, lb.BorderWidth, lb.Relief)
	}

	if lb.Font == nil || lb.lineHeight <= 0 {
		d.Flush()
		return
	}

	df, isDF := lb.Font.(platform.DrawableFont)
	if !isDF {
		d.Flush()
		return
	}

	m := lb.Font.Metrics()
	visLines := lb.visibleLines()
	clipRight := w.Width - lb.inset

	for i := 0; i < visLines; i++ {
		itemIdx := lb.topIndex + i
		if itemIdx >= len(lb.items) {
			break
		}

		rowY := lb.inset + i*lb.lineHeight
		text := lb.items[itemIdx]
		textW := lb.Font.MeasureString(text)
		availW := clipRight - lb.inset - 4
		var textX int
		switch lb.Justify {
		case option.JustifyCenter:
			textX = lb.inset + 2 + (availW-textW)/2
		case option.JustifyRight:
			textX = clipRight - 2 - textW
		default: // JustifyLeft
			textX = lb.inset + 2 - lb.xOffset
		}
		textY := rowY + m.Ascent + 1

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

	d.Flush()
}

// Configure applies options.
func (lb *Listbox) Configure(opts ...option.Option) {
	option.Apply(lb, opts)
	lb.UpdateBorder()
	lb.computeGeometry()
	if lb.Background != nil {
		lb.Win.BackgroundPixel = lb.Background.Pixel
	}
	lb.Display()
}

// Destroy cleans up the listbox.
func (lb *Listbox) Destroy() {
	if lb.Destroyed {
		return
	}
	lb.Destroyed = true
	window.DestroyWindow(lb.Win)
}
