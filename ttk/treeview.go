package ttk

import (
	"fmt"
	"sort"

	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
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

// displayEntry pairs an item with its depth in the flattened display list.
type displayEntry struct {
	item  *TreeItem
	depth int
}

// TreeItem holds data for a single treeview item.
type TreeItem struct {
	ID       string
	Parent   *TreeItem
	Children []*TreeItem
	Text     string   // tree column (#0) text
	Values   []string // data column values
	Open     bool
	Tags     []string
	Image    widget.WidgetImage // optional icon shown in tree column
}

// TreeColumn describes a data column.
type TreeColumn struct {
	ID             string
	Width          int           // pixels (default 200)
	MinWidth       int           // default 20
	Anchor         option.Anchor // cell text alignment (default W)
	Stretch        bool          // auto-resize (default true)
	Separator      bool          // draw vertical separator line after this column
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
	columns         []*TreeColumn
	showTree        bool // show tree column (#0)
	slack           int  // tree area width minus the columns (ttkTreeview.c)
	showHeadings    bool // show heading row
	treeColumnWidth int
	treeHeadingText string

	// Data.
	root   *TreeItem            // hidden root (ID="")
	items  map[string]*TreeItem // O(1) lookup
	nextID int                  // auto-gen I001, I002...

	// Flattened display list (rebuilt on structural changes).
	displayList []displayEntry

	// Deferred redisplay.
	redisplayPending bool

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
	heightRows    int // -height: rows requested (default 10)
	indent        int // pixels per depth level (default 20)
	hasFocus      bool

	// Column resize drag.
	resizeCol   int // column index being resized, -1 if none
	resizeDragX int

	// Sort indicator.
	sortInd *sortIndicator

	// Grid mode.
	Stripe bool // draw alternating row stripe

	// Double-click tracking.
	lastClickTime platform.Timestamp
	lastClickItem string

	// Callbacks.
	OnOpen        func(id string)
	OnClose       func(id string)
	OnSelect      func()
	OnDoubleClick func(id string)
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
	return func(tv *Treeview) { tv.heightRows = rows }
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

// treeviewFieldBorder is the default theme's Treeview.field border width.
const treeviewFieldBorder = 1

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
		heightRows:      10,
		resizeCol:       -1,
	}

	tv.Font, _ = app.FontRegistry().Get(font.TkDefaultFont)
	if tv.Font != nil {
		m := tv.Font.Metrics()
		// ttk::setTreeviewRowHeight: linespace + 2; the heading layout is
		// the text plus the Treeheading.border's 1px on each side.
		tv.rowHeight = m.Linespace() + 2
		tv.headingHeight = m.Linespace() + 2
	} else {
		tv.rowHeight = 20
		tv.headingHeight = 24
	}

	// Map root properly.
	tv.items[""] = tv.root

	InitTtkWidget(&tv.TtkWidget, win, app, "TTreeview")
	tv.DisplayFunc = tv.Display

	for _, opt := range opts {
		opt(tv)
	}
	// RecomputeSlack at creation: the tree area is still empty.
	tv.slack = -tv.totalWidth()

	tv.requestSize()

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
// ItemImage sets an icon image displayed in the tree column before the text.
func ItemImage(img widget.WidgetImage) ItemOption {
	return func(item *TreeItem) { item.Image = img }
}

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
	tv.scheduleRedisplay()
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
	tv.scheduleRedisplay()
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

	tv.scheduleRedisplay()
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
	tv.scheduleRedisplay()
}

// --- Selection ---

// Selection returns the IDs of selected items.
func (tv *Treeview) Selection() []string {
	var result []string
	for _, e := range tv.displayList {
		if tv.selection[e.item.ID] {
			result = append(result, e.item.ID)
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

// ColSeparatorOpt sets whether a vertical separator line is drawn after this column.
func ColSeparatorOpt(b bool) ColumnOption { return func(c *TreeColumn) { c.Separator = b } }

// ColumnConfigure configures a data column by ID.
func (tv *Treeview) ColumnConfigure(id string, opts ...ColumnOption) {
	col := tv.findColumn(id)
	if col == nil {
		return
	}
	for _, opt := range opts {
		opt(col)
	}
	if tv.Win.IsViewable() {
		// Mapped: keep the request, re-fit the columns (ttkTreeview.c).
		tv.slack = tv.treeAreaWidth() - tv.totalWidth()
		tv.resizeColumns(tv.totalWidth())
	} else {
		tv.requestSize()
	}
	tv.Display()
}

// requestSize ports TreeviewSize: the displayed columns by -height rows plus
// the heading, inside the Treeview.field border; a change re-runs the
// parent's geometry manager like Tk_GeometryRequest.
func (tv *Treeview) requestSize() {
	headH := 0
	if tv.showHeadings {
		headH = tv.headingHeight
	}
	w := tv.totalWidth() + 2*treeviewFieldBorder
	h := tv.heightRows*tv.rowHeight + headH + 2*treeviewFieldBorder
	win := tv.Win
	if w == win.ReqWidth && h == win.ReqHeight {
		return
	}
	win.ReqWidth, win.ReqHeight = w, h
	if win.GeomManager != nil {
		win.GeomManager.RequestProc(win)
	}
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

// SetStripe enables or disables alternating row stripe shading.
func (tv *Treeview) SetStripe(enabled bool) {
	tv.Stripe = enabled
	tv.Display()
}

// SetColSeparator enables or disables the vertical separator line for a column.
func (tv *Treeview) SetColSeparator(id string, enabled bool) {
	col := tv.findColumn(id)
	if col != nil {
		col.Separator = enabled
	}
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
	avail := tv.Win.Height - tv.headerOffset() - 2*treeviewFieldBorder
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
	for i, e := range tv.displayList {
		if e.item.ID == id {
			return i
		}
	}
	return -1
}

// scheduleRedisplay batches rebuildDisplayList + Display via the idle loop.
// Multiple calls before the next idle phase are coalesced into one redisplay.
func (tv *Treeview) scheduleRedisplay() {
	if tv.redisplayPending || tv.Destroyed {
		return
	}
	tv.redisplayPending = true
	tv.App.DoWhenIdle(func() {
		tv.redisplayPending = false
		if !tv.Destroyed {
			tv.rebuildDisplayList()
			tv.notifyYScrollbar()
			tv.Display()
		}
	})
}

// --- Display List ---

func (tv *Treeview) rebuildDisplayList() {
	tv.displayList = tv.displayList[:0]
	tv.walkChildren(tv.root, 0)
}

func (tv *Treeview) walkChildren(parent *TreeItem, depth int) {
	for _, child := range parent.Children {
		tv.displayList = append(tv.displayList, displayEntry{item: child, depth: depth})
		if child.Open && len(child.Children) > 0 {
			tv.walkChildren(child, depth+1)
		}
	}
}

// colRef is a display column for the slack logic: the tree column (#0) and
// the data columns, all stretchable unless configured otherwise.
type colRef struct {
	width   *int
	min     int
	stretch bool
}

func (tv *Treeview) displayColumnRefs() []colRef {
	var refs []colRef
	if tv.showTree {
		refs = append(refs, colRef{&tv.treeColumnWidth, 20, true})
	}
	for _, c := range tv.columns {
		refs = append(refs, colRef{&c.Width, c.MinWidth, c.Stretch})
	}
	return refs
}

// headingFont is the Heading style's -font (TkHeadingFont in defaults.tcl).
func (tv *Treeview) headingFont() font.Font {
	if f, err := tv.App.FontRegistry().Get(font.TkHeadingFont); err == nil {
		return f
	}
	return tv.Font
}

func (tv *Treeview) treeAreaWidth() int { return tv.Win.Width - 2*treeviewFieldBorder }

// stretchCol ports Stretch: move a column edge by n down to its minimum.
func stretchCol(c colRef, n int) int {
	if nw := *c.width + n; nw < c.min {
		n = c.min - *c.width
		*c.width = c.min
	} else {
		*c.width = nw
	}
	return n
}

// resizeColumns ports ResizeColumns (ttkTreeview.c): take up the width
// change in the slack first, spread what is left evenly over stretchable
// columns (remainder round-robin), then shove the leftovers left.
func (tv *Treeview) resizeColumns(newWidth int) {
	cols := tv.displayColumnRefs()
	delta := newWidth - (tv.totalWidth() + tv.slack)
	// PickupSlack.
	extra := 0
	if ns := tv.slack + delta; (ns < 0 && tv.slack >= 0) || (ns > 0 && tv.slack <= 0) {
		tv.slack, extra = 0, ns
	} else {
		tv.slack = ns
	}
	// DistributeWidth.
	m := 0
	for _, c := range cols {
		if c.stretch {
			m++
		}
	}
	if m > 0 {
		w := tv.totalWidth()
		d, r := extra/m, extra%m
		if r < 0 {
			r += m
			d--
		}
		for _, c := range cols {
			if c.stretch {
				w++
				add := d
				if w%m < r {
					add++
				}
				extra -= stretchCol(c, add)
			}
		}
	}
	// ShoveLeft from the last column, then DepositSlack.
	for i := len(cols) - 1; extra != 0 && i >= 0; i-- {
		if cols[i].stretch {
			extra -= stretchCol(cols[i], extra)
		}
	}
	tv.slack += extra
}

func (tv *Treeview) totalWidth() int {
	w := 0
	if tv.showTree {
		w += tv.treeColumnWidth
	}
	for _, col := range tv.columns {
		w += col.Width
	}
	return w
}

// SortChildren sorts the children of parentID by the given comparison function.
func (tv *Treeview) SortChildren(parentID string, less func(a, b *TreeItem) bool) {
	parent := tv.items[parentID]
	if parent == nil {
		parent = tv.root
	}
	sort.SliceStable(parent.Children, func(i, j int) bool {
		return less(parent.Children[i], parent.Children[j])
	})
	tv.scheduleRedisplay()
}
