package ttk

import (
	"fmt"
	"sort"

	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// TreeSelectMode defines how treeview items are selected.
type TreeSelectMode int

const (
	TreeSelectBrowse   TreeSelectMode = iota // single selection follows click
	TreeSelectExtended                       // shift/ctrl extends selection
	TreeSelectNone                           // no selection
)

// TreeItem holds data for a single treeview item.
type TreeItem struct {
	ID       string
	Parent   *TreeItem
	Children []*TreeItem
	Text     string   // tree column (#0) text
	Values   []string // data column values
	Open     bool
	Tags     []string
}

// TreeColumn describes a data column.
type TreeColumn struct {
	ID             string
	Width          int           // pixels (default 200)
	MinWidth       int           // default 20
	Anchor         option.Anchor // cell text alignment (default W)
	Stretch        bool          // auto-resize (default true)
	HeadingText    string
	HeadingAnchor  option.Anchor // default Center
	HeadingCommand func()
}

// sortIndicator tracks which column has a sort arrow.
type sortIndicator struct {
	columnID string
	reverse  bool
}

// Treeview is a themed hierarchical multicolumn list widget.
type Treeview struct {
	TtkWidget

	// Columns.
	columns        []*TreeColumn
	showTree       bool // show tree column (#0)
	showHeadings   bool // show heading row
	treeColumnWidth int
	treeHeadingText string

	// Data.
	root   *TreeItem            // hidden root (ID="")
	items  map[string]*TreeItem // O(1) lookup
	nextID int                  // auto-gen I001, I002...

	// Flattened display list.
	displayList  []*TreeItem
	displayDepth []int

	// Selection.
	selection  map[string]bool
	focus      string
	selectMode TreeSelectMode
	selAnchor  int // display-list index anchor for extended mode

	// Scrolling.
	topIndex   int
	xOffset    int
	YScrollCmd func(first, last float64)
	XScrollCmd func(first, last float64)

	// Display.
	Font          font.Font
	rowHeight     int
	headingHeight int
	indent        int // pixels per depth level (default 20)
	hasFocus      bool

	// Column resize drag.
	resizeCol  int // column index being resized, -1 if none
	resizeDragX int

	// Sort indicator.
	sortInd *sortIndicator

	// Callbacks.
	OnOpen   func(id string)
	OnClose  func(id string)
	OnSelect func()
}

// TreeviewOption configures a Treeview.
type TreeviewOption func(*Treeview)

// TreeviewColumns sets the data column IDs.
func TreeviewColumns(ids ...string) TreeviewOption {
	return func(tv *Treeview) {
		tv.columns = make([]*TreeColumn, len(ids))
		for i, id := range ids {
			tv.columns[i] = &TreeColumn{
				ID:            id,
				Width:         200,
				MinWidth:      20,
				Anchor:        option.AnchorW,
				Stretch:       true,
				HeadingAnchor: option.AnchorCenter,
			}
		}
	}
}

// TreeviewShow sets which parts to display ("tree", "headings").
func TreeviewShow(parts ...string) TreeviewOption {
	return func(tv *Treeview) {
		tv.showTree = false
		tv.showHeadings = false
		for _, p := range parts {
			switch p {
			case "tree":
				tv.showTree = true
			case "headings":
				tv.showHeadings = true
			}
		}
	}
}

// TreeviewHeight sets the number of visible rows.
func TreeviewHeight(rows int) TreeviewOption {
	return func(tv *Treeview) {
		if tv.rowHeight > 0 {
			headH := 0
			if tv.showHeadings {
				headH = tv.headingHeight
			}
			tv.Win.ReqHeight = rows*tv.rowHeight + headH + 4
		}
	}
}

// TreeviewSelectMode sets the selection mode.
func TreeviewSelectMode(mode TreeSelectMode) TreeviewOption {
	return func(tv *Treeview) { tv.selectMode = mode }
}

// TreeviewYScrollCommand sets the Y scrollbar callback.
func TreeviewYScrollCommand(fn func(float64, float64)) TreeviewOption {
	return func(tv *Treeview) { tv.YScrollCmd = fn }
}

// TreeviewXScrollCommand sets the X scrollbar callback.
func TreeviewXScrollCommand(fn func(float64, float64)) TreeviewOption {
	return func(tv *Treeview) { tv.XScrollCmd = fn }
}

// NewTreeview creates a themed treeview widget.
func NewTreeview(parent widget.Caregiver, name string, opts ...TreeviewOption) *Treeview {
	app := parent.AppContext()
	win := window.NewChildWindow(parent.Window(), name, 0, 0, 400, 200)
	window.MakeWindowExist(win)

	tv := &Treeview{
		showTree:        true,
		showHeadings:    true,
		treeColumnWidth: 200,
		root:            &TreeItem{ID: ""},
		items:           map[string]*TreeItem{"": nil}, // root mapped as ""
		selection:       make(map[string]bool),
		selectMode:      TreeSelectExtended,
		selAnchor:       -1,
		indent:          20,
		resizeCol:       -1,
	}

	tv.Font, _ = app.FontRegistry().Get(font.TkDefaultFont)
	if tv.Font != nil {
		m := tv.Font.Metrics()
		tv.rowHeight = m.Linespace() + 6
		tv.headingHeight = m.Linespace() + 8
	} else {
		tv.rowHeight = 20
		tv.headingHeight = 24
	}

	// Map root properly.
	tv.items[""] = tv.root

	InitTtkWidget(&tv.TtkWidget, win, app, "TTreeview")

	for _, opt := range opts {
		opt(tv)
	}

	// Set requested size.
	headH := 0
	if tv.showHeadings {
		headH = tv.headingHeight
	}
	win.ReqHeight = 10*tv.rowHeight + headH + 4
	win.ReqWidth = tv.totalWidth() + 4

	win.Flags |= window.FlagFocusable
	bindTreeview(tv, app)

	return tv
}

// --- Item Options ---

// ItemOption configures items on insert.
type ItemOption func(*TreeItem)

// ItemText sets the tree column text.
func ItemText(text string) ItemOption {
	return func(item *TreeItem) { item.Text = text }
}

// ItemValues sets the data column values.
func ItemValues(vals ...string) ItemOption {
	return func(item *TreeItem) { item.Values = vals }
}

// ItemOpen sets whether the item is expanded.
func ItemOpen(open bool) ItemOption {
	return func(item *TreeItem) { item.Open = open }
}

// ItemID sets a custom item ID.
func ItemID(id string) ItemOption {
	return func(item *TreeItem) { item.ID = id }
}

// ItemTags sets item tags.
func ItemTags(tags ...string) ItemOption {
	return func(item *TreeItem) { item.Tags = tags }
}

// --- Tree Operations ---

// Insert adds a new item under parentID at the given index.
// An index of -1 or end appends. Returns the item ID.
func (tv *Treeview) Insert(parentID string, index int, opts ...ItemOption) string {
	parent := tv.items[parentID]
	if parent == nil {
		parent = tv.root
	}

	item := &TreeItem{Parent: parent}
	for _, opt := range opts {
		opt(item)
	}

	if item.ID == "" {
		tv.nextID++
		item.ID = fmt.Sprintf("I%03d", tv.nextID)
	}

	if index < 0 || index >= len(parent.Children) {
		parent.Children = append(parent.Children, item)
	} else {
		parent.Children = append(parent.Children, nil)
		copy(parent.Children[index+1:], parent.Children[index:])
		parent.Children[index] = item
	}

	tv.items[item.ID] = item
	tv.rebuildDisplayList()
	tv.Display()
	return item.ID
}

// Delete removes items and all their descendants.
func (tv *Treeview) Delete(ids ...string) {
	for _, id := range ids {
		item := tv.items[id]
		if item == nil || id == "" {
			continue
		}
		tv.removeItem(item)
	}
	tv.rebuildDisplayList()
	tv.Display()
}

func (tv *Treeview) removeItem(item *TreeItem) {
	// Remove children recursively.
	for _, child := range item.Children {
		tv.removeItem(child)
	}
	// Remove from parent's children.
	if item.Parent != nil {
		children := item.Parent.Children
		for i, c := range children {
			if c == item {
				item.Parent.Children = append(children[:i], children[i+1:]...)
				break
			}
		}
	}
	delete(tv.items, item.ID)
	delete(tv.selection, item.ID)
	if tv.focus == item.ID {
		tv.focus = ""
	}
}

// Move moves an item to a new parent at the given index.
func (tv *Treeview) Move(id, parentID string, index int) {
	item := tv.items[id]
	if item == nil || id == "" {
		return
	}
	newParent := tv.items[parentID]
	if newParent == nil {
		newParent = tv.root
	}

	// Remove from old parent.
	if item.Parent != nil {
		children := item.Parent.Children
		for i, c := range children {
			if c == item {
				item.Parent.Children = append(children[:i], children[i+1:]...)
				break
			}
		}
	}

	// Add to new parent.
	item.Parent = newParent
	if index < 0 || index >= len(newParent.Children) {
		newParent.Children = append(newParent.Children, item)
	} else {
		newParent.Children = append(newParent.Children, nil)
		copy(newParent.Children[index+1:], newParent.Children[index:])
		newParent.Children[index] = item
	}

	tv.rebuildDisplayList()
	tv.Display()
}

// Item returns the item with the given ID.
func (tv *Treeview) Item(id string) *TreeItem {
	return tv.items[id]
}

// Children returns the IDs of an item's children.
func (tv *Treeview) Children(parentID string) []string {
	parent := tv.items[parentID]
	if parent == nil {
		parent = tv.root
	}
	ids := make([]string, len(parent.Children))
	for i, c := range parent.Children {
		ids[i] = c.ID
	}
	return ids
}

// Parent returns the parent ID of an item.
func (tv *Treeview) Parent(id string) string {
	item := tv.items[id]
	if item == nil || item.Parent == nil {
		return ""
	}
	return item.Parent.ID
}

// Exists returns whether an item ID exists.
func (tv *Treeview) Exists(id string) bool {
	_, ok := tv.items[id]
	return ok
}

// SetItemText sets the tree column text for an item.
func (tv *Treeview) SetItemText(id, text string) {
	if item := tv.items[id]; item != nil {
		item.Text = text
		tv.Display()
	}
}

// SetItemValues sets the data column values for an item.
func (tv *Treeview) SetItemValues(id string, vals ...string) {
	if item := tv.items[id]; item != nil {
		item.Values = vals
		tv.Display()
	}
}

// SetItemOpen sets whether an item is expanded.
func (tv *Treeview) SetItemOpen(id string, open bool) {
	item := tv.items[id]
	if item == nil {
		return
	}
	if item.Open == open {
		return
	}
	item.Open = open
	if open && tv.OnOpen != nil {
		tv.OnOpen(id)
	} else if !open && tv.OnClose != nil {
		tv.OnClose(id)
	}
	tv.rebuildDisplayList()
	tv.notifyYScrollbar()
	tv.Display()
}

// --- Selection ---

// Selection returns the IDs of selected items.
func (tv *Treeview) Selection() []string {
	var result []string
	for _, item := range tv.displayList {
		if tv.selection[item.ID] {
			result = append(result, item.ID)
		}
	}
	return result
}

// SelectionSet sets the selection to the given IDs.
func (tv *Treeview) SelectionSet(ids ...string) {
	tv.selection = make(map[string]bool)
	for _, id := range ids {
		if tv.items[id] != nil {
			tv.selection[id] = true
		}
	}
	if tv.OnSelect != nil {
		tv.OnSelect()
	}
	tv.Display()
}

// SelectionAdd adds IDs to the selection.
func (tv *Treeview) SelectionAdd(ids ...string) {
	for _, id := range ids {
		if tv.items[id] != nil {
			tv.selection[id] = true
		}
	}
	if tv.OnSelect != nil {
		tv.OnSelect()
	}
	tv.Display()
}

// SelectionRemove removes IDs from the selection.
func (tv *Treeview) SelectionRemove(ids ...string) {
	for _, id := range ids {
		delete(tv.selection, id)
	}
	if tv.OnSelect != nil {
		tv.OnSelect()
	}
	tv.Display()
}

// Focus returns the focused item ID.
func (tv *Treeview) Focus() string {
	return tv.focus
}

// SetFocus sets the focused item.
func (tv *Treeview) SetFocus(id string) {
	tv.focus = id
	tv.Display()
}

// --- Column Configuration ---

// ColumnOption configures a data column.
type ColumnOption func(*TreeColumn)

// ColWidth sets column width.
func ColWidth(w int) ColumnOption { return func(c *TreeColumn) { c.Width = w } }

// ColMinWidth sets column minimum width.
func ColMinWidth(w int) ColumnOption { return func(c *TreeColumn) { c.MinWidth = w } }

// ColAnchor sets cell text alignment.
func ColAnchor(a option.Anchor) ColumnOption { return func(c *TreeColumn) { c.Anchor = a } }

// ColStretch sets whether the column stretches.
func ColStretch(b bool) ColumnOption { return func(c *TreeColumn) { c.Stretch = b } }

// ColumnConfigure configures a data column by ID.
func (tv *Treeview) ColumnConfigure(id string, opts ...ColumnOption) {
	col := tv.findColumn(id)
	if col == nil {
		return
	}
	for _, opt := range opts {
		opt(col)
	}
	tv.Display()
}

// HeadingOption configures a column heading.
type HeadingOption func(*TreeColumn)

// HeadText sets heading text.
func HeadText(text string) HeadingOption {
	return func(c *TreeColumn) { c.HeadingText = text }
}

// HeadAnchor sets heading text alignment.
func HeadAnchor(a option.Anchor) HeadingOption {
	return func(c *TreeColumn) { c.HeadingAnchor = a }
}

// HeadCommand sets the heading click callback.
func HeadCommand(fn func()) HeadingOption {
	return func(c *TreeColumn) { c.HeadingCommand = fn }
}

// HeadingConfigure configures a column heading by ID.
// Use "#0" for the tree column heading.
func (tv *Treeview) HeadingConfigure(id string, opts ...HeadingOption) {
	if id == "#0" {
		// Tree column heading.
		col := &TreeColumn{ID: "#0", HeadingText: tv.treeHeadingText}
		for _, opt := range opts {
			opt(col)
		}
		tv.treeHeadingText = col.HeadingText
		tv.Display()
		return
	}
	col := tv.findColumn(id)
	if col == nil {
		return
	}
	for _, opt := range opts {
		opt(col)
	}
	tv.Display()
}

// SetSortIndicator shows a sort arrow on the given column.
func (tv *Treeview) SetSortIndicator(columnID string, reverse bool) {
	tv.sortInd = &sortIndicator{columnID: columnID, reverse: reverse}
	tv.Display()
}

// ClearSortIndicator removes the sort indicator.
func (tv *Treeview) ClearSortIndicator() {
	tv.sortInd = nil
	tv.Display()
}

func (tv *Treeview) findColumn(id string) *TreeColumn {
	for _, col := range tv.columns {
		if col.ID == id {
			return col
		}
	}
	return nil
}

// --- Scrolling ---

// YView scrolls to the given display-list index.
func (tv *Treeview) YView(index int) {
	if index < 0 {
		index = 0
	}
	max := len(tv.displayList) - tv.visibleRows()
	if max < 0 {
		max = 0
	}
	if index > max {
		index = max
	}
	tv.topIndex = index
	tv.notifyYScrollbar()
	tv.Display()
}

// YViewScroll scrolls by count units or pages.
func (tv *Treeview) YViewScroll(count int, pages bool) {
	if pages {
		vis := tv.visibleRows()
		if vis < 1 {
			vis = 1
		}
		count *= vis - 2
		if count == 0 {
			if pages {
				count = 1
			}
		}
	}
	tv.YView(tv.topIndex + count)
}

// YViewMoveTo scrolls to a fraction of the total items.
func (tv *Treeview) YViewMoveTo(fraction float64) {
	index := int(fraction*float64(len(tv.displayList)) + 0.5)
	tv.YView(index)
}

// YVisibleRange returns the visible fraction for the y scrollbar.
func (tv *Treeview) YVisibleRange() (float64, float64) {
	n := len(tv.displayList)
	if n == 0 {
		return 0, 1
	}
	first := float64(tv.topIndex) / float64(n)
	vis := tv.visibleRows()
	last := float64(tv.topIndex+vis) / float64(n)
	if last > 1 {
		last = 1
	}
	return first, last
}

// See scrolls to make the given item visible.
func (tv *Treeview) See(id string) {
	idx := tv.displayIndex(id)
	if idx < 0 {
		return
	}
	vis := tv.visibleRows()
	if idx < tv.topIndex {
		tv.topIndex = idx
	} else if idx >= tv.topIndex+vis {
		tv.topIndex = idx - vis + 1
	}
	tv.notifyYScrollbar()
	tv.Display()
}

func (tv *Treeview) visibleRows() int {
	if tv.rowHeight <= 0 {
		return 1
	}
	avail := tv.Win.Height - tv.headerOffset() - 4
	if avail < 1 {
		return 1
	}
	return avail / tv.rowHeight
}

func (tv *Treeview) headerOffset() int {
	if tv.showHeadings {
		return tv.headingHeight
	}
	return 0
}

func (tv *Treeview) notifyYScrollbar() {
	if tv.YScrollCmd != nil {
		first, last := tv.YVisibleRange()
		tv.YScrollCmd(first, last)
	}
}

func (tv *Treeview) displayIndex(id string) int {
	for i, item := range tv.displayList {
		if item.ID == id {
			return i
		}
	}
	return -1
}

// --- Display List ---

func (tv *Treeview) rebuildDisplayList() {
	tv.displayList = tv.displayList[:0]
	tv.displayDepth = tv.displayDepth[:0]
	tv.walkChildren(tv.root, 0)
}

func (tv *Treeview) walkChildren(parent *TreeItem, depth int) {
	for _, child := range parent.Children {
		tv.displayList = append(tv.displayList, child)
		tv.displayDepth = append(tv.displayDepth, depth)
		if child.Open && len(child.Children) > 0 {
			tv.walkChildren(child, depth+1)
		}
	}
}

func (tv *Treeview) totalWidth() int {
	w := 0
	if tv.showTree {
		w += tv.treeColumnWidth
	}
	for _, col := range tv.columns {
		w += col.Width
	}
	if w < 200 {
		w = 200
	}
	return w
}

// --- Display Engine ---

// Display draws the treeview widget.
func (tv *Treeview) Display() {
	if tv.Destroyed {
		return
	}
	win := tv.Win
	if win.XWindow == xlib.Window(0) {
		return
	}

	d := win.Display.XDisplay
	gc := win.GC
	width := win.Width
	height := win.Height

	if width <= 0 || height <= 0 {
		return
	}

	// Allocate or resize pixmap.
	if tv.pixmap == xlib.Pixmap(0) || tv.pixmapW != width || tv.pixmapH != height {
		if tv.pixmap != xlib.Pixmap(0) {
			d.FreePixmap(tv.pixmap)
		}
		tv.pixmap = d.CreatePixmap(win.Drawable(), uint(width), uint(height), uint(win.Depth))
		tv.pixmapW = width
		tv.pixmapH = height
	}

	pixDrawable := xlib.PixmapDrawable(tv.pixmap)

	// Colors from style.
	bg := LookupColor(tv.Context.Style, "-background", tv.State, 0xd9d9d9)
	fg := LookupColor(tv.Context.Style, "-foreground", tv.State, 0x000000)
	fieldBg := LookupColor(tv.Context.Style, "-fieldbackground", tv.State, 0xffffff)
	selBg := LookupColor(tv.Context.Style, "-selectbackground", tv.State, 0x4a6984)
	selFg := LookupColor(tv.Context.Style, "-selectforeground", tv.State, 0xffffff)

	// Fill background.
	d.SetForeground(gc, bg)
	d.FillRectangle(pixDrawable, gc, 0, 0, uint(width), uint(height))

	// Field area (white background for items).
	itemAreaY := tv.headerOffset()
	itemAreaH := height - itemAreaY
	if itemAreaH > 0 {
		d.SetForeground(gc, fieldBg)
		d.FillRectangle(pixDrawable, gc, 0, itemAreaY, uint(width), uint(itemAreaH))
	}

	xftFont, isXft := tv.Font.(*font.XftFont)
	if !isXft {
		d.CopyArea(pixDrawable, win.Drawable(), gc, 0, 0, uint(width), uint(height), 0, 0)
		d.Flush()
		return
	}

	m := tv.Font.Metrics()
	fgR, fgG, fgB := colorToRGB16(fg)
	selFgR, selFgG, selFgB := colorToRGB16(selFg)

	// --- Draw headings ---
	if tv.showHeadings {
		headBg := LookupColor(tv.Context.Style, "-background", tv.State, 0xd9d9d9)
		border := draw.NewBorderFromPixel(headBg)

		colX := 0
		if tv.showTree {
			// Tree column heading.
			draw.Fill3DRectangle(d, pixDrawable, gc, border,
				colX, 0, tv.treeColumnWidth, tv.headingHeight, 1, option.ReliefRaised)
			if tv.treeHeadingText != "" {
				tv.drawAlignedText(xftFont, pixDrawable, colX+4, 0, tv.treeColumnWidth-8,
					tv.headingHeight, tv.treeHeadingText, option.AnchorW, fg, fgR, fgG, fgB, m)
			}
			colX += tv.treeColumnWidth
		}

		for _, col := range tv.columns {
			draw.Fill3DRectangle(d, pixDrawable, gc, border,
				colX, 0, col.Width, tv.headingHeight, 1, option.ReliefRaised)
			if col.HeadingText != "" {
				textW := col.Width - 8
				textX := colX + 4
				// Reserve space for sort indicator.
				if tv.sortInd != nil && tv.sortInd.columnID == col.ID {
					textW -= 12
				}
				if textW > 0 {
					tv.drawAlignedText(xftFont, pixDrawable, textX, 0, textW,
						tv.headingHeight, col.HeadingText, col.HeadingAnchor, fg, fgR, fgG, fgB, m)
				}
			}
			// Sort indicator.
			if tv.sortInd != nil && tv.sortInd.columnID == col.ID {
				tv.drawSortIndicator(d, pixDrawable, gc, colX+col.Width-14, tv.headingHeight/2, tv.sortInd.reverse, fg)
			}
			colX += col.Width
		}

		// Fill remainder of heading row.
		if colX < width {
			draw.Fill3DRectangle(d, pixDrawable, gc, border,
				colX, 0, width-colX, tv.headingHeight, 1, option.ReliefRaised)
		}
	}

	// --- Draw items ---
	visRows := tv.visibleRows()
	for i := 0; i < visRows; i++ {
		idx := tv.topIndex + i
		if idx >= len(tv.displayList) {
			break
		}

		item := tv.displayList[idx]
		depth := tv.displayDepth[idx]
		rowY := itemAreaY + i*tv.rowHeight

		isSelected := tv.selection[item.ID]
		isFocused := tv.focus == item.ID

		// Selection highlight.
		if isSelected {
			d.SetForeground(gc, selBg)
			d.FillRectangle(pixDrawable, gc, 0, rowY, uint(width), uint(tv.rowHeight))
		}

		textPixel := fg
		textR, textG, textB := fgR, fgG, fgB
		if isSelected {
			textPixel = selFg
			textR, textG, textB = selFgR, selFgG, selFgB
		}

		colX := 0

		// Tree column.
		if tv.showTree {
			indentX := depth * tv.indent
			indicatorX := colX + indentX + 2
			textStartX := colX + indentX + tv.indent + 2

			// Draw expand/collapse indicator if item has children.
			if len(item.Children) > 0 {
				tv.drawIndicator(d, pixDrawable, gc, indicatorX, rowY, item.Open, textPixel)
			}

			// Draw item text.
			if item.Text != "" {
				textY := rowY + (tv.rowHeight-m.Linespace())/2 + m.Ascent
				maxW := tv.treeColumnWidth - (textStartX - colX) - 4
				tv.drawClippedText(xftFont, pixDrawable, textStartX, textY, maxW,
					item.Text, textPixel, textR, textG, textB)
			}

			colX += tv.treeColumnWidth
		}

		// Data columns.
		for ci, col := range tv.columns {
			val := ""
			if ci < len(item.Values) {
				val = item.Values[ci]
			}
			if val != "" {
				tv.drawAlignedText(xftFont, pixDrawable, colX+4, rowY, col.Width-8,
					tv.rowHeight, val, col.Anchor, textPixel, textR, textG, textB, m)
			}
			colX += col.Width
		}

		// Focus dotted rectangle.
		if isFocused && tv.hasFocus {
			d.SetForeground(gc, fg)
			d.DrawRectangle(pixDrawable, gc, 0, rowY, uint(width-1), uint(tv.rowHeight-1))
		}
	}

	// Draw column separator lines in item area.
	if itemAreaH > 0 {
		sepColor := LookupColor(tv.Context.Style, "-bordercolor", tv.State, 0xd9d9d9)
		d.SetForeground(gc, sepColor)
		colX := 0
		if tv.showTree {
			colX += tv.treeColumnWidth
			d.DrawLine(pixDrawable, gc, colX-1, itemAreaY, colX-1, height)
		}
		for _, col := range tv.columns {
			colX += col.Width
			d.DrawLine(pixDrawable, gc, colX-1, itemAreaY, colX-1, height)
		}
	}

	// Sunken border.
	border := draw.NewBorderFromPixel(bg)
	draw.Draw3DRectangle(d, pixDrawable, gc, border, 0, 0, width, height, 1, option.ReliefSunken)

	// Copy to window.
	d.CopyArea(pixDrawable, win.Drawable(), gc, 0, 0, uint(width), uint(height), 0, 0)
	d.Flush()
}

func (tv *Treeview) drawAlignedText(xftFont *font.XftFont, drawable xlib.Drawable,
	x, y, maxW, h int, text string, anchor option.Anchor,
	pixel uint64, r, g, b uint16, m font.Metrics) {

	textW := tv.Font.MeasureString(text)
	textY := y + (h-m.Linespace())/2 + m.Ascent

	textX := x
	switch anchor {
	case option.AnchorCenter, option.AnchorN, option.AnchorS:
		textX = x + (maxW-textW)/2
	case option.AnchorE, option.AnchorNE, option.AnchorSE:
		textX = x + maxW - textW
	}

	tv.drawClippedText(xftFont, drawable, textX, textY, maxW, text, pixel, r, g, b)
}

func (tv *Treeview) drawClippedText(xftFont *font.XftFont, drawable xlib.Drawable,
	x, y, maxW int, text string, pixel uint64, r, g, b uint16) {

	if maxW <= 0 {
		return
	}
	textW := tv.Font.MeasureString(text)
	if textW <= maxW {
		xftFont.DrawString(drawable, x, y, text, pixel, r, g, b)
		return
	}
	// Truncate with ellipsis.
	ellipsis := "..."
	ellW := tv.Font.MeasureString(ellipsis)
	avail := maxW - ellW
	if avail <= 0 {
		return
	}
	for i := len(text); i > 0; i-- {
		if tv.Font.MeasureString(text[:i]) <= avail {
			xftFont.DrawString(drawable, x, y, text[:i]+ellipsis, pixel, r, g, b)
			return
		}
	}
}

func (tv *Treeview) drawIndicator(d *xlib.Display, drawable xlib.Drawable, gc xlib.GC,
	x, y int, open bool, pixel uint64) {

	// Draw a small triangle: right-pointing (closed) or down-pointing (open).
	cx := x + 4
	cy := y + tv.rowHeight/2

	d.SetForeground(gc, pixel)

	if open {
		// Down-pointing triangle.
		points := []draw.Point{
			{X: cx - 4, Y: cy - 2},
			{X: cx + 4, Y: cy - 2},
			{X: cx, Y: cy + 3},
		}
		draw.FillPolygon(d, drawable, gc, points)
	} else {
		// Right-pointing triangle.
		points := []draw.Point{
			{X: cx - 2, Y: cy - 4},
			{X: cx + 3, Y: cy},
			{X: cx - 2, Y: cy + 4},
		}
		draw.FillPolygon(d, drawable, gc, points)
	}
}

func (tv *Treeview) drawSortIndicator(d *xlib.Display, drawable xlib.Drawable, gc xlib.GC,
	x, cy int, reverse bool, pixel uint64) {

	d.SetForeground(gc, pixel)

	if reverse {
		// Down arrow.
		points := []draw.Point{
			{X: x, Y: cy - 3},
			{X: x + 8, Y: cy - 3},
			{X: x + 4, Y: cy + 3},
		}
		draw.FillPolygon(d, drawable, gc, points)
	} else {
		// Up arrow.
		points := []draw.Point{
			{X: x + 4, Y: cy - 3},
			{X: x + 8, Y: cy + 3},
			{X: x, Y: cy + 3},
		}
		draw.FillPolygon(d, drawable, gc, points)
	}
}

func colorToRGB16(pixel uint64) (uint16, uint16, uint16) {
	r := uint16((pixel>>16)&0xFF) << 8
	g := uint16((pixel>>8)&0xFF) << 8
	b := uint16((pixel)&0xFF) << 8
	return r, g, b
}

// --- Hit Testing ---

type hitRegion int

const (
	hitNothing hitRegion = iota
	hitHeading
	hitSeparator
	hitIndicator
	hitTree
	hitCell
)

type hitResult struct {
	region hitRegion
	itemID string
	colIdx int // -1 for tree column
	dispIdx int // display list index
}

func (tv *Treeview) hitTest(x, y int) hitResult {
	// Heading area.
	if tv.showHeadings && y < tv.headingHeight {
		colX := 0
		colIdx := -1
		if tv.showTree {
			if x >= tv.treeColumnWidth-3 && x <= tv.treeColumnWidth+1 {
				return hitResult{region: hitSeparator, colIdx: -1}
			}
			if x < tv.treeColumnWidth {
				return hitResult{region: hitHeading, colIdx: -1}
			}
			colX = tv.treeColumnWidth
		}
		for ci, col := range tv.columns {
			colIdx = ci
			nextX := colX + col.Width
			if x >= nextX-3 && x <= nextX+1 {
				return hitResult{region: hitSeparator, colIdx: colIdx}
			}
			if x >= colX && x < nextX {
				return hitResult{region: hitHeading, colIdx: colIdx}
			}
			colX = nextX
		}
		return hitResult{region: hitHeading, colIdx: colIdx}
	}

	// Item area.
	itemAreaY := tv.headerOffset()
	if y < itemAreaY {
		return hitResult{region: hitNothing}
	}

	row := (y - itemAreaY) / tv.rowHeight
	dispIdx := tv.topIndex + row
	if dispIdx >= len(tv.displayList) {
		return hitResult{region: hitNothing}
	}

	item := tv.displayList[dispIdx]
	depth := tv.displayDepth[dispIdx]

	// Check tree column.
	if tv.showTree && x < tv.treeColumnWidth {
		indentX := depth * tv.indent
		indicatorEnd := indentX + tv.indent
		if x >= indentX && x < indicatorEnd && len(item.Children) > 0 {
			return hitResult{region: hitIndicator, itemID: item.ID, colIdx: -1, dispIdx: dispIdx}
		}
		return hitResult{region: hitTree, itemID: item.ID, colIdx: -1, dispIdx: dispIdx}
	}

	// Check data columns.
	colX := 0
	if tv.showTree {
		colX = tv.treeColumnWidth
	}
	for ci, col := range tv.columns {
		if x >= colX && x < colX+col.Width {
			return hitResult{region: hitCell, itemID: item.ID, colIdx: ci, dispIdx: dispIdx}
		}
		colX += col.Width
	}

	return hitResult{region: hitCell, itemID: item.ID, colIdx: -1, dispIdx: dispIdx}
}

// --- Event Bindings ---

func bindTreeview(tv *Treeview, app widget.AppContext) {
	win := tv.Win

	// Expose.
	app.Dispatcher().Bind(win.XWindow, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		tv.Display()
	})

	// Configure (resize).
	app.Dispatcher().Bind(win.XWindow, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			win.Width = ev.ConfigWidth
			win.Height = ev.ConfigHeight
			tv.notifyYScrollbar()
			tv.Display()
		}
	})

	// Focus.
	app.Dispatcher().Bind(win.XWindow, event.FocusChangeMask, func(ev *event.Event) {
		if ev.Type == event.FocusInType {
			tv.hasFocus = true
			tv.Display()
		} else if ev.Type == event.FocusOutType {
			tv.hasFocus = false
			tv.Display()
		}
	})

	// Button press.
	app.Dispatcher().Bind(win.XWindow, event.ButtonPressMask, func(ev *event.Event) {
		// Take focus.
		app.DisplayPtr().SetInputFocus(win.XWindow, xlib.RevertToParent, xlib.CurrentTime)

		if ev.Button == 1 {
			hit := tv.hitTest(ev.X, ev.Y)
			switch hit.region {
			case hitSeparator:
				// Start column resize.
				tv.resizeCol = hit.colIdx
				tv.resizeDragX = ev.X
			case hitHeading:
				// Heading click.
				if hit.colIdx >= 0 && hit.colIdx < len(tv.columns) {
					col := tv.columns[hit.colIdx]
					if col.HeadingCommand != nil {
						col.HeadingCommand()
					}
				}
			case hitIndicator:
				// Toggle open/close.
				if item := tv.items[hit.itemID]; item != nil {
					tv.SetItemOpen(hit.itemID, !item.Open)
				}
			case hitTree, hitCell:
				// Select item.
				if hit.itemID != "" {
					tv.handleSelect(hit.itemID, hit.dispIdx, ev.State)
					tv.focus = hit.itemID
					tv.Display()
				}
			}
		} else if ev.Button == 4 {
			tv.YView(tv.topIndex - 3)
		} else if ev.Button == 5 {
			tv.YView(tv.topIndex + 3)
		}
	})

	// Motion (column resize drag).
	app.Dispatcher().Bind(win.XWindow, event.MotionMask, func(ev *event.Event) {
		if tv.resizeCol < 0 {
			return
		}
		delta := ev.X - tv.resizeDragX
		if tv.resizeCol == -1 {
			// Tree column resize.
			newW := tv.treeColumnWidth + delta
			if newW < 20 {
				newW = 20
			}
			tv.treeColumnWidth = newW
		} else if tv.resizeCol < len(tv.columns) {
			col := tv.columns[tv.resizeCol]
			newW := col.Width + delta
			if newW < col.MinWidth {
				newW = col.MinWidth
			}
			col.Width = newW
		}
		tv.resizeDragX = ev.X
		tv.Display()
	})

	// Button release.
	app.Dispatcher().Bind(win.XWindow, event.ButtonReleaseMask, func(ev *event.Event) {
		if ev.Button == 1 {
			tv.resizeCol = -1
		}
	})

	// Keyboard.
	app.Dispatcher().Bind(win.XWindow, event.KeyPressMask, func(ev *event.Event) {
		switch ev.KeySym {
		case xlib.XK_Up:
			tv.moveFocus(-1)
		case xlib.XK_Down:
			tv.moveFocus(1)
		case xlib.XK_Left:
			// Collapse current or move to parent.
			if item := tv.items[tv.focus]; item != nil {
				if item.Open && len(item.Children) > 0 {
					tv.SetItemOpen(tv.focus, false)
				} else if item.Parent != nil && item.Parent.ID != "" {
					tv.focus = item.Parent.ID
					tv.SelectionSet(tv.focus)
					tv.See(tv.focus)
				}
			}
		case xlib.XK_Right:
			// Expand current or move to first child.
			if item := tv.items[tv.focus]; item != nil {
				if !item.Open && len(item.Children) > 0 {
					tv.SetItemOpen(tv.focus, true)
				} else if len(item.Children) > 0 {
					tv.focus = item.Children[0].ID
					tv.SelectionSet(tv.focus)
					tv.See(tv.focus)
				}
			}
		case xlib.XK_Return, xlib.XK_space:
			if item := tv.items[tv.focus]; item != nil && len(item.Children) > 0 {
				tv.SetItemOpen(tv.focus, !item.Open)
			}
		case xlib.XK_Home:
			if len(tv.displayList) > 0 {
				tv.focus = tv.displayList[0].ID
				tv.SelectionSet(tv.focus)
				tv.See(tv.focus)
			}
		case xlib.XK_End:
			if len(tv.displayList) > 0 {
				tv.focus = tv.displayList[len(tv.displayList)-1].ID
				tv.SelectionSet(tv.focus)
				tv.See(tv.focus)
			}
		}
	})

	// Enter/Leave for hover state.
	app.Dispatcher().Bind(win.XWindow, event.EnterMask, func(ev *event.Event) {
		tv.ChangeState(StateHover|StateActive, 0)
	})
	app.Dispatcher().Bind(win.XWindow, event.LeaveMask, func(ev *event.Event) {
		tv.ChangeState(0, StateHover|StateActive|StatePressed)
	})
}

func (tv *Treeview) handleSelect(id string, dispIdx int, state uint) {
	switch tv.selectMode {
	case TreeSelectNone:
		return
	case TreeSelectBrowse:
		tv.selection = map[string]bool{id: true}
		tv.selAnchor = dispIdx
	case TreeSelectExtended:
		shift := state&xlib.ShiftMask != 0
		ctrl := state&xlib.ControlMask != 0
		if shift && tv.selAnchor >= 0 {
			tv.selection = make(map[string]bool)
			lo, hi := tv.selAnchor, dispIdx
			if lo > hi {
				lo, hi = hi, lo
			}
			for i := lo; i <= hi; i++ {
				if i < len(tv.displayList) {
					tv.selection[tv.displayList[i].ID] = true
				}
			}
		} else if ctrl {
			if tv.selection[id] {
				delete(tv.selection, id)
			} else {
				tv.selection[id] = true
			}
			tv.selAnchor = dispIdx
		} else {
			tv.selection = map[string]bool{id: true}
			tv.selAnchor = dispIdx
		}
	}
	if tv.OnSelect != nil {
		tv.OnSelect()
	}
}

func (tv *Treeview) moveFocus(delta int) {
	if len(tv.displayList) == 0 {
		return
	}
	idx := tv.displayIndex(tv.focus)
	if idx < 0 {
		idx = 0
	} else {
		idx += delta
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= len(tv.displayList) {
		idx = len(tv.displayList) - 1
	}
	tv.focus = tv.displayList[idx].ID
	tv.SelectionSet(tv.focus)
	tv.See(tv.focus)
}

// --- Sorting helpers ---

// SortChildren sorts the children of parentID by the given comparison function.
func (tv *Treeview) SortChildren(parentID string, less func(a, b *TreeItem) bool) {
	parent := tv.items[parentID]
	if parent == nil {
		parent = tv.root
	}
	sort.SliceStable(parent.Children, func(i, j int) bool {
		return less(parent.Children[i], parent.Children[j])
	})
	tv.rebuildDisplayList()
	tv.Display()
}
