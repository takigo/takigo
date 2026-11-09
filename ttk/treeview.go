// Ports tk/generic/ttk/ttkTreeview.c and tk/library/ttk/treeview.tcl.
package ttk

import (
	"fmt"
	"slices"
	"sort"
	"strconv"

	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/window"
)

// TreeSelectMode defines how treeview items are selected.
type TreeSelectMode int

const (
	TreeSelectBrowse   TreeSelectMode = iota // single selection follows click
	TreeSelectExtended                       // shift/ctrl extends selection
	TreeSelectNone                           // no selection
)

// The virtual events a Treeview sends, like Tk's.
const (
	TreeviewSelectEvent = "TreeviewSelect" // the selection changed
	TreeviewOpenEvent   = "TreeviewOpen"   // the user is opening the focus item
	TreeviewCloseEvent  = "TreeviewClose"  // the user closed the focus item
)

// displayEntry pairs an item with its depth in the flattened display list.
type displayEntry struct {
	item  *TreeItem
	depth int
}

// TreeItem holds data for a single treeview item.
type TreeItem struct {
	ID       string
	Parent   *TreeItem // nil for a detached item
	Children []*TreeItem
	Text     string   // tree column (#0) text
	Values   []string // data column values
	Open     bool
	Tags     []string
	Image    widget.WidgetImage // optional icon shown in tree column
}

// TreeColumn describes a column: the tree column #0 or a data column.
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

	index int // position in -columns, -1 for #0
}

func newTreeColumn(id string, index int) *TreeColumn {
	return &TreeColumn{
		ID:            id,
		Width:         200,
		MinWidth:      20,
		Anchor:        option.AnchorW,
		Stretch:       true,
		HeadingAnchor: option.AnchorCenter,
		index:         index,
	}
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
	column0        *TreeColumn   // the tree column #0
	columns        []*TreeColumn // -columns
	displayColumns []string      // -displaycolumns; nil is #all
	showTree       bool          // show tree column (#0)
	slack          int           // tree area width minus the columns (ttkTreeview.c)
	showHeadings   bool          // show heading row

	// Data.
	root   *TreeItem            // hidden root (ID="")
	items  map[string]*TreeItem // O(1) lookup
	nextID int                  // auto-gen I001, I002...

	// Tags, in creation order: the first created has the highest priority.
	tags     map[string]*treeTag
	tagOrder []string

	// Flattened display list (rebuilt on structural changes).
	displayList []displayEntry

	// Deferred redisplay.
	redisplayPending bool

	// Selection.
	selection  map[string]bool
	focus      string
	selectMode TreeSelectMode

	// Scrolling.
	topIndex   int
	xFirst     int // pixel offset of the view (tree.xscroll.first)
	xSent      [2]float64
	xSentValid bool
	YScrollCmd func(first, last float64)
	XScrollCmd func(first, last float64)

	// Display.
	Font          font.Font
	rowHeight     int
	headingHeight int
	heightRows    int // -height: rows requested (default 10)
	indent        int // pixels per depth level (default 20)
	hasFocus      bool

	// Button-1 press state (treeview.tcl State).
	press treeviewPress

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
			tv.columns[i] = newTreeColumn(id, i)
		}
		tv.displayColumns = nil
	}
}

// TreeviewDisplayColumns sets -displaycolumns: the data columns shown, in
// order, by ID or by index in -columns. No IDs, or "#all", shows them all.
func TreeviewDisplayColumns(ids ...string) TreeviewOption {
	return func(tv *Treeview) {
		if len(ids) == 0 || (len(ids) == 1 && ids[0] == "#all") {
			tv.displayColumns = nil
			return
		}
		tv.displayColumns = slices.Clone(ids)
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

// TreeviewStyle sets -style.
func TreeviewStyle(name string) TreeviewOption {
	return func(tv *Treeview) { tv.StyleName = name }
}

// NewTreeview creates a themed treeview widget.
func NewTreeview(parent widget.Caregiver, name string, opts ...TreeviewOption) *Treeview {
	app := parent.AppContext()
	win := window.NewChildWindow(parent.Window(), name, 0, 0, 400, 200)
	window.MakeWindowExist(win)

	tv := &Treeview{
		column0:      newTreeColumn("#0", -1),
		showTree:     true,
		showHeadings: true,
		root:         &TreeItem{ID: ""},
		tags:         make(map[string]*treeTag),
		selection:    make(map[string]bool),
		selectMode:   TreeSelectExtended,
		indent:       20,
		heightRows:   10,
	}
	tv.items = map[string]*TreeItem{"": tv.root}

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

	InitTtkWidget(&tv.TtkWidget, win, app, "TTreeview")
	tv.reconfigure = func() { _ = tv.Configure() }
	tv.DisplayFunc = tv.Display

	for _, opt := range opts {
		opt(tv)
	}
	if tv.StyleName != "TTreeview" {
		tv.RefreshTheme()
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

// ItemImage sets an icon image displayed in the tree column before the text.
func ItemImage(img widget.WidgetImage) ItemOption {
	return func(item *TreeItem) { item.Image = img }
}

// ItemTags sets the item's tags.
func ItemTags(tags ...string) ItemOption {
	return func(item *TreeItem) { item.Tags = tags }
}

// --- Tree Operations ---

// insertChild links item into parent at index (-1 or past the end appends).
func insertChild(parent, item *TreeItem, index int) {
	item.Parent = parent
	if index < 0 || index >= len(parent.Children) {
		parent.Children = append(parent.Children, item)
		return
	}
	parent.Children = slices.Insert(parent.Children, index, item)
}

// detachItem unlinks item from its parent (DetachItem).
func detachItem(item *TreeItem) {
	if p := item.Parent; p != nil {
		p.Children = slices.DeleteFunc(p.Children, func(c *TreeItem) bool { return c == item })
	}
	item.Parent = nil
}

// Insert adds a new item under parentID at the given index.
// An index of -1 or end appends. Returns the item ID.
func (tv *Treeview) Insert(parentID string, index int, opts ...ItemOption) string {
	parent := tv.items[parentID]
	if parent == nil {
		parent = tv.root
	}

	item := &TreeItem{}
	for _, opt := range opts {
		opt(item)
	}

	if item.ID == "" {
		tv.nextID++
		item.ID = fmt.Sprintf("I%03d", tv.nextID)
	}
	for _, t := range item.Tags {
		tv.tag(t)
	}

	insertChild(parent, item, index)
	tv.items[item.ID] = item
	tv.scheduleRedisplay()
	return item.ID
}

// Delete removes items and all their descendants.
func (tv *Treeview) Delete(ids ...string) {
	selChange := false
	for _, id := range ids {
		item := tv.items[id]
		if item == nil || id == "" {
			continue
		}
		detachItem(item)
		selChange = tv.removeItem(item) || selChange
	}
	if selChange {
		tv.selectionChanged()
	}
	tv.scheduleRedisplay()
}

// removeItem forgets item and its descendants and reports whether one of
// them was selected; the caller unlinks item from its parent.
func (tv *Treeview) removeItem(item *TreeItem) bool {
	sel := tv.selection[item.ID]
	for _, child := range item.Children {
		sel = tv.removeItem(child) || sel
	}
	item.Children = nil
	delete(tv.items, item.ID)
	delete(tv.selection, item.ID)
	if tv.focus == item.ID {
		tv.focus = ""
	}
	return sel
}

// Move moves an item to a new parent at the given index; it also
// reattaches a detached item. Moving an item under itself or one of its
// descendants does nothing (AncestryCheck).
func (tv *Treeview) Move(id, parentID string, index int) {
	item := tv.items[id]
	if item == nil || id == "" {
		return
	}
	newParent := tv.items[parentID]
	if newParent == nil {
		newParent = tv.root
	}
	for p := newParent; p != nil; p = p.Parent {
		if p == item {
			return
		}
	}
	detachItem(item)
	insertChild(newParent, item, index)
	tv.scheduleRedisplay()
}

// Detach unlinks items from the tree; they keep their descendants and can
// be put back with Move.
func (tv *Treeview) Detach(ids ...string) {
	for _, id := range ids {
		if item := tv.items[id]; item != nil && item != tv.root {
			detachItem(item)
		}
	}
	tv.scheduleRedisplay()
}

// IsDetached reports whether the item is detached; the root never is.
func (tv *Treeview) IsDetached(id string) bool {
	item := tv.items[id]
	return item != nil && item != tv.root && item.Parent == nil
}

// Detached returns the detached items, sorted by ID.
func (tv *Treeview) Detached() []string {
	var ids []string
	for id, item := range tv.items {
		if item != tv.root && item.Parent == nil {
			ids = append(ids, id)
		}
	}
	slices.Sort(ids)
	return ids
}

// detachedOrAncestor is IsItemOrAncestorDetached.
func (tv *Treeview) detachedOrAncestor(item *TreeItem) bool {
	for ; item != tv.root; item = item.Parent {
		if item.Parent == nil {
			return true
		}
	}
	return false
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

// sibling returns the item delta places from id among its siblings.
func (tv *Treeview) sibling(id string, delta int) string {
	item := tv.items[id]
	if item == nil || item.Parent == nil {
		return ""
	}
	sibs := item.Parent.Children
	if i := slices.Index(sibs, item) + delta; i >= 0 && i < len(sibs) {
		return sibs[i].ID
	}
	return ""
}

// Next returns the ID of the item's next sibling, or "".
func (tv *Treeview) Next(id string) string { return tv.sibling(id, 1) }

// Prev returns the ID of the item's previous sibling, or "".
func (tv *Treeview) Prev(id string) string { return tv.sibling(id, -1) }

// Index returns the item's position among its siblings (0 when detached).
func (tv *Treeview) Index(id string) int {
	item := tv.items[id]
	if item == nil || item.Parent == nil {
		return 0
	}
	return slices.Index(item.Parent.Children, item)
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

// CellValue returns the item's value in a data column ("set $item
// $column"); the column is an ID, a "#n" display number or an index.
func (tv *Treeview) CellValue(id, column string) string {
	item, col := tv.items[id], tv.lookupColumn(column)
	if item == nil || col == nil || col.index < 0 || col.index >= len(item.Values) {
		return ""
	}
	return item.Values[col.index]
}

// SetCellValue sets the item's value in a data column ("set $item $column
// $value"), padding the values with empty strings. Column #0 has no value:
// use SetItemText.
func (tv *Treeview) SetCellValue(id, column, value string) {
	item, col := tv.items[id], tv.lookupColumn(column)
	if item == nil || col == nil || col.index < 0 {
		return
	}
	for len(item.Values) < len(tv.columns) {
		item.Values = append(item.Values, "")
	}
	item.Values[col.index] = value
	tv.Display()
}

// CellValues returns the item's values by column ID ("set $item").
func (tv *Treeview) CellValues(id string) map[string]string {
	item := tv.items[id]
	if item == nil {
		return nil
	}
	m := make(map[string]string)
	for i, c := range tv.columns {
		if i < len(item.Values) {
			m[c.ID] = item.Values[i]
		}
	}
	return m
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

// --- Virtual events ---

// sendVirtual queues <<name>> for the treeview like Tk_SendVirtualEvent.
func (tv *Treeview) sendVirtual(name string) {
	tv.App.DoWhenIdle(func() {
		if !tv.Destroyed() {
			tv.generate(name)
		}
	})
}

// generate dispatches <<name>> now, like "event generate".
func (tv *Treeview) generate(name string) {
	tv.App.Dispatcher().Dispatch(&event.Event{Type: event.VirtualType, Window: tv.Win.PlatformID, Name: name})
}

// --- Selection ---

// preorder calls fn for every attached item below the root, in order.
func (tv *Treeview) preorder(fn func(*TreeItem)) {
	var walk func(*TreeItem)
	walk = func(p *TreeItem) {
		for _, c := range p.Children {
			fn(c)
			walk(c)
		}
	}
	walk(tv.root)
}

// Selection returns the IDs of selected items in tree order.
func (tv *Treeview) Selection() []string {
	var result []string
	tv.preorder(func(item *TreeItem) {
		if tv.selection[item.ID] {
			result = append(result, item.ID)
		}
	})
	return result
}

func (tv *Treeview) selectionChanged() {
	if tv.OnSelect != nil {
		tv.OnSelect()
	}
	tv.sendVirtual(TreeviewSelectEvent)
}

// SelectionSet sets the selection to the given IDs.
func (tv *Treeview) SelectionSet(ids ...string) {
	changed := len(tv.selection) > 0
	tv.selection = make(map[string]bool)
	for _, id := range ids {
		if tv.items[id] != nil {
			tv.selection[id] = true
			changed = true
		}
	}
	if changed {
		tv.selectionChanged()
	}
	tv.Display()
}

// SelectionAdd adds IDs to the selection.
func (tv *Treeview) SelectionAdd(ids ...string) {
	changed := false
	for _, id := range ids {
		if tv.items[id] != nil && !tv.selection[id] {
			tv.selection[id] = true
			changed = true
		}
	}
	if changed {
		tv.selectionChanged()
	}
	tv.Display()
}

// SelectionRemove removes IDs from the selection.
func (tv *Treeview) SelectionRemove(ids ...string) {
	changed := false
	for _, id := range ids {
		if tv.selection[id] {
			delete(tv.selection, id)
			changed = true
		}
	}
	if changed {
		tv.selectionChanged()
	}
	tv.Display()
}

// SelectionToggle toggles the selection state of each ID.
func (tv *Treeview) SelectionToggle(ids ...string) {
	changed := false
	for _, id := range ids {
		if tv.items[id] == nil {
			continue
		}
		if tv.selection[id] {
			delete(tv.selection, id)
		} else {
			tv.selection[id] = true
		}
		changed = true
	}
	if changed {
		tv.selectionChanged()
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
func ColWidth[L screenunit.Length](w L) ColumnOption {
	return func(c *TreeColumn) { c.Width = screenunit.ToPixels(w) }
}

// ColMinWidth sets column minimum width.
func ColMinWidth[L screenunit.Length](w L) ColumnOption {
	return func(c *TreeColumn) { c.MinWidth = screenunit.ToPixels(w) }
}

// ColAnchor sets cell text alignment.
func ColAnchor(a option.Anchor) ColumnOption { return func(c *TreeColumn) { c.Anchor = a } }

// ColStretch sets whether column auto-resizes.
func ColStretch(b bool) ColumnOption { return func(c *TreeColumn) { c.Stretch = b } }

// ColSeparator sets whether a vertical separator line is drawn after this column.
func ColSeparator(b bool) ColumnOption { return func(c *TreeColumn) { c.Separator = b } }

// ColumnConfigure configures a column: "#0" is the tree column, "#n" the
// n-th displayed data column, otherwise a column ID or index.
func (tv *Treeview) ColumnConfigure(id string, opts ...ColumnOption) {
	col := tv.lookupColumn(id)
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

// Column returns the column's settings (see ColumnConfigure for id), or
// nil when there is no such column.
func (tv *Treeview) Column(id string) *TreeColumn { return tv.lookupColumn(id) }

// Columns returns the IDs of -columns.
func (tv *Treeview) Columns() []string {
	ids := make([]string, len(tv.columns))
	for i, c := range tv.columns {
		ids[i] = c.ID
	}
	return ids
}

// lookupColumn ports FindColumn: a column ID, "#n" for display column n
// (#0 is the tree column) or a decimal index in -columns.
func (tv *Treeview) lookupColumn(id string) *TreeColumn {
	if id == "#0" {
		return tv.column0
	}
	for _, col := range tv.columns {
		if col.ID == id {
			return col
		}
	}
	if len(id) > 1 && id[0] == '#' {
		n, err := strconv.Atoi(id[1:])
		if dc := tv.dataDisplayColumns(); err == nil && n >= 1 && n <= len(dc) {
			return dc[n-1]
		}
		return nil
	}
	if n, err := strconv.Atoi(id); err == nil && n >= 0 && n < len(tv.columns) {
		return tv.columns[n]
	}
	return nil
}

// dataDisplayColumns returns the data columns -displaycolumns shows.
func (tv *Treeview) dataDisplayColumns() []*TreeColumn {
	if tv.displayColumns == nil {
		return tv.columns
	}
	var cols []*TreeColumn
	for _, id := range tv.displayColumns {
		var col *TreeColumn
		for _, c := range tv.columns {
			if c.ID == id {
				col = c
				break
			}
		}
		if col == nil {
			if n, err := strconv.Atoi(id); err == nil && n >= 0 && n < len(tv.columns) {
				col = tv.columns[n]
			}
		}
		if col != nil {
			cols = append(cols, col)
		}
	}
	return cols
}

// shownColumns returns the columns drawn left to right: the tree column
// when -show has tree, then the displayed data columns.
func (tv *Treeview) shownColumns() []*TreeColumn {
	dc := tv.dataDisplayColumns()
	if !tv.showTree {
		return dc
	}
	return append([]*TreeColumn{tv.column0}, dc...)
}

// displayNumber is the "#n" number of a shown column.
func (tv *Treeview) displayNumber(col *TreeColumn) int {
	if col == tv.column0 {
		return 0
	}
	return slices.Index(tv.dataDisplayColumns(), col) + 1
}

// computeSize ports TreeviewSize: the displayed columns by -height rows plus
// the heading, inside the Treeview.field border.
func (tv *Treeview) computeSize() {
	headH := 0
	if tv.showHeadings {
		headH = tv.headingHeight
	}
	tv.Win.ReqWidth = tv.totalWidth() + 2*treeviewFieldBorder
	tv.Win.ReqHeight = tv.heightRows*tv.rowHeight + headH + 2*treeviewFieldBorder
}

// requestSize recomputes the request; a change re-runs the parent's
// geometry manager like Tk_GeometryRequest.
func (tv *Treeview) requestSize() {
	tv.resize(tv.computeSize)
}

// fitColumns keeps the request and re-fits the columns once the tree is
// mapped (ttkTreeview.c); before that the request follows the columns.
func (tv *Treeview) fitColumns() {
	if tv.Win.IsViewable() {
		tv.slack = tv.treeAreaWidth() - tv.totalWidth()
		tv.resizeColumns(tv.totalWidth())
	} else {
		tv.slack = -tv.totalWidth()
	}
}

// Configure sets options after creation.
func (tv *Treeview) Configure(opts ...TreeviewOption) error {
	return configure(&tv.TtkWidget, tv, opts, tv.fitColumns, func() {
		if !tv.Win.IsViewable() {
			tv.computeSize()
		}
	})
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

// HeadingConfigure configures a column heading; id is as for
// ColumnConfigure ("#0" is the tree column heading).
func (tv *Treeview) HeadingConfigure(id string, opts ...HeadingOption) {
	col := tv.lookupColumn(id)
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
	if col := tv.lookupColumn(id); col != nil {
		col.Separator = enabled
	}
	tv.Display()
}

// --- Scrolling ---

// YView scrolls to the given display-list index.
func (tv *Treeview) YView(index int) {
	if index < 0 {
		index = 0
	}
	maxTop := max(len(tv.displayList)-tv.visibleRows(), 0)
	if index > maxTop {
		index = maxTop
	}
	tv.topIndex = index
	tv.notifyYScrollbar()
	tv.Display()
}

// YViewScroll scrolls by count units or pages.
func (tv *Treeview) YViewScroll(count int, pages bool) {
	// TtkScrollviewCommand: a page is the number of rows in view.
	if pages {
		count *= max(tv.visibleRows(), 1)
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

// See opens the item's ancestors and scrolls to make it visible; a
// detached item (or one under a detached item) is left alone.
func (tv *Treeview) See(id string) {
	item := tv.items[id]
	if item == nil || item == tv.root || tv.detachedOrAncestor(item) {
		return
	}
	opened := false
	for p := item.Parent; p != tv.root; p = p.Parent {
		if !p.Open {
			p.Open = true
			opened = true
		}
	}
	if opened {
		tv.rebuildDisplayList()
	}
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

// scrolled ports TtkScrolled's sanity checks: a view past the end slides
// back so that it ends at the end.
func scrolled(first, last, total int) (int, int, int) {
	if total <= 0 {
		return 0, 1, 1
	}
	if last > total {
		first = max(first-(last-total), 0)
		last = total
	}
	return first, last, total
}

// scrollTo ports TtkScrollTo's clamping of a new first position.
func scrollTo(newFirst, first, last, total int) int {
	if newFirst >= total {
		newFirst = total - 1
	}
	if newFirst > first && last >= total {
		newFirst = first
	}
	return max(newFirst, 0)
}

// xscrollInfo is the horizontal view in pixels as TreeviewDoLayout reports
// it; it also pulls xFirst back when the view got wider than the columns.
func (tv *Treeview) xscrollInfo() (first, last, total int) {
	first, last, total = scrolled(tv.xFirst, tv.xFirst+max(tv.treeAreaWidth(), 0), tv.totalWidth())
	tv.xFirst = first
	return first, last, total
}

// XView scrolls so that pixel offset px of the columns is at the left edge.
func (tv *Treeview) XView(px int) {
	first, last, total := tv.xscrollInfo()
	if nf := scrollTo(px, first, last, total); nf != first {
		tv.xFirst = nf
		tv.Display()
	}
}

// XViewScroll scrolls by count pixels or pages (TtkScrollviewCommand).
func (tv *Treeview) XViewScroll(count int, pages bool) {
	first, last, _ := tv.xscrollInfo()
	if pages {
		count *= last - first
	}
	tv.XView(first + count)
}

// XViewMoveTo scrolls to a fraction of the total column width.
func (tv *Treeview) XViewMoveTo(fraction float64) {
	_, _, total := tv.xscrollInfo()
	tv.XView(int(fraction*float64(total) + 0.5))
}

// XVisibleRange returns the visible fraction for the x scrollbar.
func (tv *Treeview) XVisibleRange() (float64, float64) {
	first, last, total := tv.xscrollInfo()
	return float64(first) / float64(total), float64(last) / float64(total)
}

// notifyXScrollbar calls XScrollCmd when the view changed since the last
// call, like the idle UpdateScrollbarBG.
func (tv *Treeview) notifyXScrollbar() {
	if tv.XScrollCmd == nil {
		return
	}
	first, last := tv.XVisibleRange()
	if tv.xSentValid && tv.xSent == [2]float64{first, last} {
		return
	}
	tv.xSent, tv.xSentValid = [2]float64{first, last}, true
	tv.XScrollCmd(first, last)
}

func (tv *Treeview) displayIndex(id string) int {
	tv.syncDisplayList()
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
	if tv.redisplayPending || tv.Destroyed() {
		return
	}
	tv.redisplayPending = true
	tv.App.DoWhenIdle(func() {
		tv.redisplayPending = false
		if !tv.Destroyed() {
			tv.rebuildDisplayList()
			tv.notifyYScrollbar()
			tv.Display()
		}
	})
}

// --- Display List ---

// syncDisplayList brings the display list up to date before the idle
// redisplay, so Selection and See work right after Insert or Delete.
func (tv *Treeview) syncDisplayList() {
	if tv.redisplayPending {
		tv.rebuildDisplayList()
	}
}

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

// headingFont is the Heading style's -font (TkHeadingFont in defaults.tcl).
func (tv *Treeview) headingFont() font.Font {
	if f, err := tv.App.FontRegistry().Get(font.TkHeadingFont); err == nil {
		return f
	}
	return tv.Font
}

func (tv *Treeview) treeAreaWidth() int { return tv.Win.Width - 2*treeviewFieldBorder }

// stretchCol ports Stretch: move a column edge by n down to its minimum.
func stretchCol(c *TreeColumn, n int) int {
	if nw := c.Width + n; nw < c.MinWidth {
		n = c.MinWidth - c.Width
		c.Width = c.MinWidth
	} else {
		c.Width = nw
	}
	return n
}

// pickupSlack ports PickupSlack: a width change goes to the slack first;
// columns change only once the slack crosses zero.
func (tv *Treeview) pickupSlack(extra int) int {
	ns := tv.slack + extra
	if (ns < 0 && tv.slack >= 0) || (ns > 0 && tv.slack <= 0) {
		tv.slack = 0
		return ns
	}
	tv.slack = ns
	return 0
}

// shove ports ShoveLeft (step -1) and ShoveRight (step 1): move n pixels
// into the stretchable columns from cols[i] on; returns the leftover.
func shove(cols []*TreeColumn, i, step, n int) int {
	for ; n != 0 && i >= 0 && i < len(cols); i += step {
		if cols[i].Stretch {
			n -= stretchCol(cols[i], n)
		}
	}
	return n
}

// distributeWidth ports DistributeWidth: n pixels spread evenly over the
// stretchable columns, the remainder round-robin.
func (tv *Treeview) distributeWidth(cols []*TreeColumn, n int) int {
	m := 0
	for _, c := range cols {
		if c.Stretch {
			m++
		}
	}
	if m == 0 {
		return n
	}
	w := tv.totalWidth()
	d, r := n/m, n%m
	if r < 0 {
		r += m
		d--
	}
	for _, c := range cols {
		if c.Stretch {
			w++
			add := d
			if w%m < r {
				add++
			}
			n -= stretchCol(c, add)
		}
	}
	return n
}

// resizeColumns ports ResizeColumns (ttkTreeview.c): take up the width
// change in the slack first, spread what is left evenly over stretchable
// columns, then shove the leftovers left.
func (tv *Treeview) resizeColumns(newWidth int) {
	cols := tv.shownColumns()
	delta := newWidth - (tv.totalWidth() + tv.slack)
	tv.slack += shove(cols, len(cols)-1, -1, tv.distributeWidth(cols, tv.pickupSlack(delta)))
}

// dragColumn ports DragColumn: move the right edge of cols[i] by delta,
// taking the change from the columns to its left and right.
func (tv *Treeview) dragColumn(i, delta int) {
	cols := tv.shownColumns()
	dl := delta - shove(cols, i-1, -1, delta-stretchCol(cols[i], delta))
	tv.slack += shove(cols, i+1, 1, tv.pickupSlack(-dl))
}

// DragColumn moves the right edge of a shown column to x, in widget
// coordinates ("drag"); Drop re-fits the columns afterwards.
func (tv *Treeview) DragColumn(column string, x int) {
	col := tv.lookupColumn(column)
	left := treeviewFieldBorder - tv.xFirst
	for i, c := range tv.shownColumns() {
		right := left + c.Width
		if c == col {
			tv.dragColumn(i, x-right)
			tv.Display()
			return
		}
		left = right
	}
}

// Drop ends a column drag ("drop").
func (tv *Treeview) Drop() {
	tv.resizeColumns(tv.totalWidth())
	tv.Display()
}

// totalWidth is TreeWidth: the width of the shown columns.
func (tv *Treeview) totalWidth() int {
	w := 0
	for _, col := range tv.shownColumns() {
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
