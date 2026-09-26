package canvas

import (
	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/geometry"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// Canvas is a 2D drawing surface that supports arbitrary graphical items.
type Canvas struct {
	widget.Base

	items    []*itemEntry
	idMap    map[int64]*itemEntry
	tagIndex map[string]map[int64]*itemEntry // tag name → item IDs for O(1) tag lookup
	nextID   int64

	stipples map[string]platform.PixmapID // depth-1 pixmaps by bitmap spec

	// Scroll state.
	xOrigin, yOrigin int
	scrollRegion     [4]int
	hasScrollRegion  bool

	// Display.
	redrawPending    bool
	pixmap           platform.PixmapID
	pixmapW, pixmapH int

	// Item pick / events.
	currentItem  *itemEntry
	itemBindings map[string][]itemHandler
	closeEnough  float64 // hit-test tolerance (default 1.0)

	// Keyboard focus for text items.
	focusItemID int64 // 0 = none

	// Scrollbar callbacks.
	XScrollCmd func(first, last float64)
	YScrollCmd func(first, last float64)

	// Computed inset (borderWidth + highlightWidth).
	inset int

	// Tk -width/-height: the drawing area, excluding the inset.
	reqW, reqH int

	// displayFunc is stored so ScheduleRedraw can call it.
	displayFunc func()
}

// CanvasOption configures a Canvas.
type CanvasOption func(*Canvas)

// Width sets -width (a Tk distance: pixels or "10c", "3i", ...).
func Width(w any) CanvasOption {
	return func(c *Canvas) { c.reqW = screenunit.Px(w) }
}

// Height sets -height (a Tk distance).
func Height(h any) CanvasOption {
	return func(c *Canvas) { c.reqH = screenunit.Px(h) }
}

func Background(name string) CanvasOption {
	return func(c *Canvas) {
		col, err := c.App.ColorCache().Get(name)
		if err == nil {
			c.Base.Background = col
			c.UpdateBorder()
		}
	}
}

func BorderWidthOpt(w int) CanvasOption {
	return func(c *Canvas) {
		c.BorderWidth = w
		c.inset = w + c.HighlightWidth
	}
}

func ReliefOpt(r option.Relief) CanvasOption {
	return func(c *Canvas) { c.Relief = r }
}

func HighlightWidthOpt(w int) CanvasOption {
	return func(c *Canvas) {
		c.HighlightWidth = w
		c.inset = c.BorderWidth + w
	}
}

func CloseEnough(d float64) CanvasOption {
	return func(c *Canvas) { c.closeEnough = d }
}

func XScrollCommand(fn func(first, last float64)) CanvasOption {
	return func(c *Canvas) { c.XScrollCmd = fn }
}

func YScrollCommand(fn func(first, last float64)) CanvasOption {
	return func(c *Canvas) { c.YScrollCmd = fn }
}

func ScrollRegion(x1, y1, x2, y2 int) CanvasOption {
	return func(c *Canvas) {
		c.scrollRegion = [4]int{x1, y1, x2, y2}
		c.hasScrollRegion = true
	}
}

// New creates a new Canvas widget.
func New(parent widget.Caregiver, name string, opts ...CanvasOption) *Canvas {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(w)

	c := &Canvas{
		idMap:        make(map[int64]*itemEntry),
		tagIndex:     make(map[string]map[int64]*itemEntry),
		nextID:       1,
		itemBindings: make(map[string][]itemHandler),
		closeEnough:  1.0,
	}
	widget.InitBase(&c.Base, w, app)
	w.Class = "Canvas"

	// Canvas defaults (tkUnixDefault.h: -width 10c -height 7c,
	// -highlightthickness 1, normal background).
	c.HighlightWidth = 1
	c.reqW = screenunit.Px("10c")
	c.reqH = screenunit.Px("7c")

	// Apply options.
	for _, opt := range opts {
		opt(c)
	}

	c.inset = c.BorderWidth + c.HighlightWidth
	// xOrigin is the canvas coordinate at the inner (inset) edge; Tk starts
	// with canvas 0 at window pixel 0, underneath the border.
	c.xOrigin, c.yOrigin = c.inset, c.inset
	// The requested size includes the inset (CanvasWorldChanged in tkCanvas.c).
	w.ReqWidth = c.reqW + 2*c.inset
	w.ReqHeight = c.reqH + 2*c.inset

	// Store display function reference for idle callback.
	c.displayFunc = c.Display

	// Like Tk_CreateWindow, the canvas is 1x1 until a geometry manager
	// sizes it; -scrollregion confinement before then depends on that.
	w.Width, w.Height = 1, 1

	// Set window background pixel for child window creation.
	if c.Base.Background != nil {
		w.BackgroundPixel = c.Base.Background.Pixel
	}

	// Set up event handlers.
	bindCanvas(c)

	return c
}

// Display draws the canvas and all its items.
func (c *Canvas) Display() {
	if c.Destroyed {
		return
	}
	w := c.Win
	if w.PlatformID == 0 || !w.IsMapped() {
		return
	}

	d := w.Display.Server

	// Compute visible canvas rectangle.
	winW := w.Width - 2*c.inset
	winH := w.Height - 2*c.inset
	if winW <= 0 || winH <= 0 {
		return
	}

	// Overdraw padding (30px like Tk).
	const overdraw = 30
	pixW := winW + 2*overdraw
	pixH := winH + 2*overdraw

	// Allocate or resize pixmap.
	if c.pixmap == 0 || c.pixmapW != pixW || c.pixmapH != pixH {
		if c.pixmap != 0 {
			d.FreePixmap(c.pixmap)
		}
		c.pixmap = d.CreatePixmap(w.Drawable(), uint(pixW), uint(pixH), uint(w.Depth))
		c.pixmapW = pixW
		c.pixmapH = pixH
	}
	if c.pixmap == 0 {
		return
	}

	pxDrawable := platform.PixmapDrawable(c.pixmap)
	gc := w.GC

	// Clear pixmap to background color.
	bgPixel := uint64(0xFFFFFF)
	if c.Base.Background != nil {
		bgPixel = c.Base.Background.Pixel
	}
	d.SetForeground(gc, bgPixel)
	d.FillRectangle(pxDrawable, gc, 0, 0, uint(pixW), uint(pixH))

	// The origin in canvas coordinates that maps to the pixmap origin.
	pixOriginX := c.xOrigin - overdraw
	pixOriginY := c.yOrigin - overdraw

	// Clip rectangle in canvas coordinates.
	clipX := pixOriginX
	clipY := pixOriginY
	clipW := pixW
	clipH := pixH

	// Draw items bottom to top.
	for _, entry := range c.items {
		item := entry.item
		if base := itemBase(item); base != nil && base.State() == ItemStateHidden {
			continue
		}

		// Cull by bounding box.
		ix1, iy1, ix2, iy2 := item.BBox()
		if ix2 < clipX || ix1 > clipX+clipW || iy2 < clipY || iy1 > clipY+clipH {
			continue
		}

		item.Display(d, pxDrawable, gc, clipX, clipY, clipW, clipH, pixOriginX, pixOriginY)
	}

	// Copy pixmap to window (accounting for overdraw offset and inset).
	d.CopyArea(pxDrawable, w.Drawable(), gc,
		overdraw, overdraw, uint(winW), uint(winH),
		c.inset, c.inset)

	// Position embedded window items.
	c.positionWindowItems()

	// Draw 3D border if configured.
	if c.Border != nil && c.BorderWidth > 0 && c.Relief != option.ReliefFlat {
		hl := c.HighlightWidth
		draw.Draw3DRectangle(d, w.Drawable(), gc, c.Border,
			hl, hl, w.Width-2*hl, w.Height-2*hl, c.BorderWidth, c.Relief)
	}
	// The highlight ring sits outside the border (focus colour or
	// -highlightbackground).
	c.DrawHighlightBorder(false, 0)

	d.Flush()
	c.redrawPending = false
	c.NeedRedraw = false
}

// positionWindowItems maps and positions all embedded WindowItems relative to
// the canvas window. Items outside the visible area are unmapped.
func (c *Canvas) positionWindowItems() {
	d := c.App.Server()
	winW := c.Win.Width - 2*c.inset
	winH := c.Win.Height - 2*c.inset
	for _, entry := range c.items {
		wi, ok := entry.item.(*WindowItem)
		if !ok || wi.win == nil {
			continue
		}
		w := wi.win.ReqWidth
		h := wi.win.ReqHeight
		if w == 0 {
			w = wi.win.Width
		}
		if h == 0 {
			h = wi.win.Height
		}
		if w == 0 || h == 0 {
			continue
		}
		// Convert canvas coords to window coords.
		screenX := wi.X1 - c.xOrigin + c.inset
		screenY := wi.Y1 - c.yOrigin + c.inset
		// Check visibility.
		if screenX+w > 0 && screenX < winW && screenY+h > 0 && screenY < winH {
			wi.win.X, wi.win.Y, wi.win.Width, wi.win.Height = screenX, screenY, w, h
			d.MoveResizeWindow(wi.win.PlatformID, screenX, screenY, uint(w), uint(h))
			d.MapWindow(wi.win.PlatformID)
			window.MarkMapped(wi.win)
		} else {
			d.UnmapWindow(wi.win.PlatformID)
			wi.win.Flags &^= window.FlagMapped
		}
	}
}

// scheduleRedraw schedules a redraw via the idle loop.
func (c *Canvas) scheduleRedraw() {
	if c.redrawPending || c.Destroyed {
		return
	}
	c.redrawPending = true
	c.App.DoWhenIdle(func() {
		if !c.Destroyed {
			c.Display()
		}
	})
}

// Destroy cleans up the canvas and all its items.
func (c *Canvas) Destroy() {
	if c.Destroyed {
		return
	}
	c.Destroyed = true

	d := c.Win.Display.Server

	// Free all items.
	for _, entry := range c.items {
		entry.item.Delete(d)
	}
	c.items = nil
	c.idMap = nil
	c.tagIndex = nil

	// Free pixmap.
	if c.pixmap != 0 {
		d.FreePixmap(c.pixmap)
		c.pixmap = 0
	}
}

// Configure sets canvas options.
func (c *Canvas) Configure(opts ...CanvasOption) {
	for _, opt := range opts {
		opt(c)
	}
	if inset := c.BorderWidth + c.HighlightWidth; inset != c.inset {
		c.xOrigin += inset - c.inset
		c.yOrigin += inset - c.inset
		c.inset = inset
	}
	// Tk relays out at idle, so CanvasSetOrigin still sees the old window
	// size; our geometry managers resize synchronously, so confine first.
	c.setOrigin(c.xOrigin, c.yOrigin)
	if rw, rh := c.reqW+2*c.inset, c.reqH+2*c.inset; rw != c.Win.ReqWidth || rh != c.Win.ReqHeight {
		geometry.GeometryRequest(c.Win, rw, rh)
	}
	c.scheduleRedraw()
}

// --- Item creation ---

func (c *Canvas) addItem(item Item) int64 {
	base := itemBase(item)
	id := c.nextID
	c.nextID++
	base.ID = id
	base.canvas = c

	entry := &itemEntry{id: id, item: item}
	c.items = append(c.items, entry)
	c.idMap[id] = entry

	// Register initial tags in the index.
	for _, tag := range base.Tags {
		c.tagIndexAdd(tag, id)
	}

	c.scheduleRedraw()
	return id
}

// tagIndexAdd registers item id under tag in the tag index.
func (c *Canvas) tagIndexAdd(tag string, id int64) {
	m := c.tagIndex[tag]
	if m == nil {
		m = make(map[int64]*itemEntry)
		c.tagIndex[tag] = m
	}
	m[id] = c.idMap[id]
}

// tagIndexRemove unregisters item id from tag in the tag index.
func (c *Canvas) tagIndexRemove(tag string, id int64) {
	if m := c.tagIndex[tag]; m != nil {
		delete(m, id)
		if len(m) == 0 {
			delete(c.tagIndex, tag)
		}
	}
}

// CreateRectangle creates a rectangle item.
func (c *Canvas) CreateRectangle(x1, y1, x2, y2 float64, opts ...ItemOption) int64 {
	item := newRectOvalItem("rectangle", x1, y1, x2, y2, c)
	item.Configure(opts)
	return c.addItem(item)
}

// CreateOval creates an oval item.
func (c *Canvas) CreateOval(x1, y1, x2, y2 float64, opts ...ItemOption) int64 {
	item := newRectOvalItem("oval", x1, y1, x2, y2, c)
	item.Configure(opts)
	return c.addItem(item)
}

// CreateLine creates a line item.
func (c *Canvas) CreateLine(coords []float64, opts ...ItemOption) int64 {
	item := newLineItem(coords, c)
	item.Configure(opts)
	return c.addItem(item)
}

// CreatePolygon creates a polygon item.
func (c *Canvas) CreatePolygon(coords []float64, opts ...ItemOption) int64 {
	item := newPolygonItem(coords, c)
	item.Configure(opts)
	return c.addItem(item)
}

// CreateArc creates an arc item.
func (c *Canvas) CreateArc(x1, y1, x2, y2 float64, opts ...ItemOption) int64 {
	item := newArcItem(x1, y1, x2, y2, c)
	item.Configure(opts)
	return c.addItem(item)
}

// CreateText creates a text item.
func (c *Canvas) CreateText(x, y float64, opts ...ItemOption) int64 {
	item := newTextItem(x, y, c)
	item.Configure(opts)
	return c.addItem(item)
}

// CreateImage creates an image item.
// CreateWindow embeds a child window at canvas position (x, y).
// The window must be a child of the canvas window (created with canvas as parent).
func (c *Canvas) CreateWindow(x, y float64, w *window.Window, opts ...ItemOption) int64 {
	item := newWindowItem(x, y, w, c)
	if err := item.Configure(opts); err == nil {
		item.updateBBox()
	}
	return c.addItem(item)
}

// CreateBitmap creates a 1-bit XBM bitmap item at (x, y).
func (c *Canvas) CreateBitmap(x, y float64, xbm *XBMData, opts ...ItemOption) int64 {
	item := newBitmapItem(x, y, xbm, c)
	item.Configure(opts)
	return c.addItem(item)
}

func (c *Canvas) CreateImage(x, y float64, opts ...ItemOption) int64 {
	item := newImageItem(x, y, c)
	item.Configure(opts)
	return c.addItem(item)
}

// --- Item manipulation ---

// Delete removes all items matching tagOrID.
func (c *Canvas) Delete(tagOrID string) {
	entries := c.resolve(tagOrID)
	if len(entries) == 0 {
		return
	}

	d := c.Win.Display.Server
	deleteSet := make(map[int64]bool, len(entries))
	for _, e := range entries {
		deleteSet[e.id] = true
		e.item.Delete(d)
		delete(c.idMap, e.id)
	}

	// Remove from display list.
	filtered := c.items[:0]
	for _, e := range c.items {
		if !deleteSet[e.id] {
			filtered = append(filtered, e)
		}
	}
	c.items = filtered

	// Remove deleted items from tag index.
	for _, e := range entries {
		if base := itemBase(e.item); base != nil {
			for _, tag := range base.Tags {
				c.tagIndexRemove(tag, e.id)
			}
		}
	}

	// Clear current item if deleted.
	if c.currentItem != nil && deleteSet[c.currentItem.id] {
		c.currentItem = nil
	}

	c.scheduleRedraw()
}

// CurrentItem returns the ID of the item under the mouse cursor, or -1 if none.
func (c *Canvas) CurrentItem() int64 {
	if c.currentItem == nil {
		return -1
	}
	return c.currentItem.id
}

// Move translates all items matching tagOrID by (dx, dy).
func (c *Canvas) Move(tagOrID string, dx, dy float64) {
	for _, entry := range c.resolve(tagOrID) {
		entry.item.Translate(dx, dy)
	}
	c.scheduleRedraw()
}

// Scale scales all items matching tagOrID about (ox, oy).
func (c *Canvas) Scale(tagOrID string, ox, oy, sx, sy float64) {
	for _, entry := range c.resolve(tagOrID) {
		entry.item.Scale(ox, oy, sx, sy)
	}
	c.scheduleRedraw()
}

// Raise moves items matching tagOrID to the top of the display list.
func (c *Canvas) Raise(tagOrID string) {
	entries := c.resolve(tagOrID)
	if len(entries) == 0 {
		return
	}
	moveSet := make(map[int64]bool, len(entries))
	for _, e := range entries {
		moveSet[e.id] = true
	}

	// In-place partition: kept items first, moved items appended at end.
	n := len(c.items)
	j := 0
	moved := make([]*itemEntry, 0, len(entries))
	for i := 0; i < n; i++ {
		if moveSet[c.items[i].id] {
			moved = append(moved, c.items[i])
		} else {
			c.items[j] = c.items[i]
			j++
		}
	}
	copy(c.items[j:], moved)
	c.scheduleRedraw()
}

// Lower moves items matching tagOrID to the bottom of the display list.
func (c *Canvas) Lower(tagOrID string) {
	entries := c.resolve(tagOrID)
	if len(entries) == 0 {
		return
	}
	moveSet := make(map[int64]bool, len(entries))
	for _, e := range entries {
		moveSet[e.id] = true
	}

	// In-place partition: moved items first, kept items shifted right.
	n := len(c.items)
	j := n - 1
	kept := make([]*itemEntry, 0, n-len(entries))
	for i := n - 1; i >= 0; i-- {
		if moveSet[c.items[i].id] {
			c.items[j] = c.items[i]
			j--
		} else {
			kept = append(kept, c.items[i])
		}
	}
	// kept is in reverse order; copy reversed into the front.
	for i, k := 0, len(kept)-1; k >= 0; i, k = i+1, k-1 {
		c.items[i] = kept[k]
	}
	c.scheduleRedraw()
}

// AddTag adds a tag to all items matching tagOrID.
func (c *Canvas) AddTag(newTag, tagOrID string) {
	for _, entry := range c.resolve(tagOrID) {
		if base := itemBase(entry.item); base != nil {
			base.AddTag(newTag)
		}
	}
}

// DeleteTag removes a tag from all items matching tagOrID.
func (c *Canvas) DeleteTag(tag, tagOrID string) {
	for _, entry := range c.resolve(tagOrID) {
		if base := itemBase(entry.item); base != nil {
			base.RemoveTag(tag)
		}
	}
}

// ItemConfigure configures items matching tagOrID.
func (c *Canvas) ItemConfigure(tagOrID string, opts ...ItemOption) error {
	for _, entry := range c.resolve(tagOrID) {
		if err := entry.item.Configure(opts); err != nil {
			return err
		}
	}
	c.scheduleRedraw()
	return nil
}

// ItemCoords returns coordinates of the first item matching tagOrID.
func (c *Canvas) ItemCoords(tagOrID string) []float64 {
	entries := c.resolve(tagOrID)
	if len(entries) == 0 {
		return nil
	}
	return entries[0].item.Coords()
}

// SetItemCoords sets coordinates on the first item matching tagOrID.
func (c *Canvas) SetItemCoords(tagOrID string, coords []float64) error {
	entries := c.resolve(tagOrID)
	if len(entries) == 0 {
		return nil
	}
	err := entries[0].item.SetCoords(coords)
	if err == nil {
		c.scheduleRedraw()
	}
	return err
}

// FindWithTag returns item IDs matching a tag-or-ID string.
// Supports: numeric IDs, "all", "current", and tag name strings.
func (c *Canvas) FindWithTag(tagOrID string) []int64 {
	entries := c.resolve(tagOrID)
	ids := make([]int64, len(entries))
	for i, e := range entries {
		ids[i] = e.id
	}
	return ids
}

// Find returns item IDs that satisfy the given search mode.
// Supported modes: "all", "closest" (args: x, y), "enclosed" (args: x1,y1,x2,y2),
// "overlapping" (args: x1,y1,x2,y2).
func (c *Canvas) Find(mode string, args ...float64) []int64 {
	switch mode {
	case "all":
		ids := make([]int64, len(c.items))
		for i, e := range c.items {
			ids[i] = e.id
		}
		return ids
	case "closest":
		if len(args) < 2 {
			return nil
		}
		halo := c.closeEnough
		if len(args) >= 3 {
			halo = args[2]
		}
		entry := c.findClosest(args[0], args[1], halo)
		if entry != nil {
			return []int64{entry.id}
		}
		return nil
	case "enclosed":
		if len(args) < 4 {
			return nil
		}
		var ids []int64
		for _, e := range c.items {
			if e.item.AreaOverlap(args[0], args[1], args[2], args[3]) == 1 {
				ids = append(ids, e.id)
			}
		}
		return ids
	case "overlapping":
		if len(args) < 4 {
			return nil
		}
		var ids []int64
		for _, e := range c.items {
			if e.item.AreaOverlap(args[0], args[1], args[2], args[3]) >= 0 {
				ids = append(ids, e.id)
			}
		}
		return ids
	}
	return nil
}

// BBox returns the bounding box of the first item matching tagOrID.
func (c *Canvas) BBox(tagOrID string) (x1, y1, x2, y2 int) {
	// Tk's "bbox" is the union over all matching items.
	found := false
	for _, e := range c.resolve(tagOrID) {
		a1, b1, a2, b2 := e.item.BBox()
		if a2 < a1 || b2 < b1 {
			continue
		}
		if !found {
			x1, y1, x2, y2 = a1, b1, a2, b2
			found = true
			continue
		}
		x1, y1 = min(x1, a1), min(y1, b1)
		x2, y2 = max(x2, a2), max(y2, b2)
	}
	return x1, y1, x2, y2
}

// GetTags returns the tags of the first item matching tagOrID.
func (c *Canvas) GetTags(tagOrID string) []string {
	entries := c.resolve(tagOrID)
	if len(entries) == 0 {
		return nil
	}
	if base := itemBase(entries[0].item); base != nil {
		tags := make([]string, len(base.Tags))
		copy(tags, base.Tags)
		return tags
	}
	return nil
}

// ColorCache returns the application's color cache (for item option resolution).
func (c *Canvas) ColorCache() *color.Cache {
	return c.App.ColorCache()
}

// FontRegistry returns the application's font registry (for text items).
func (c *Canvas) FontRegistry() *font.Registry {
	return c.App.FontRegistry()
}

// DisplayServer returns the platform display server (for items needing display access).
func (c *Canvas) DisplayServer() platform.DisplayServer {
	return c.App.Server()
}

// Focus sets keyboard focus to the given text item. Pass "" to clear focus.
func (c *Canvas) Focus(tagOrID string) {
	// Clear previous focus.
	if c.focusItemID != 0 {
		if prev, ok := c.idMap[c.focusItemID]; ok {
			if ti, ok := prev.item.(*TextItem); ok {
				ti.hasFocus = false
			}
		}
		c.focusItemID = 0
	}
	if tagOrID == "" {
		c.scheduleRedraw()
		return
	}
	entries := c.resolve(tagOrID)
	for _, e := range entries {
		if ti, ok := e.item.(*TextItem); ok {
			ti.hasFocus = true
			c.focusItemID = e.id
			// Make canvas window focusable and give it X11 focus.
			c.Win.Flags |= window.FlagFocusable
			c.App.Server().SetInputFocus(c.Win.PlatformID, platform.RevertToParent, platform.CurrentTime)
			break
		}
	}
	c.scheduleRedraw()
}

// ICursor sets the insertion cursor position in a text item.
// index can be a number or "end".
func (c *Canvas) ICursor(tagOrID string, index string) {
	entries := c.resolve(tagOrID)
	for _, e := range entries {
		if ti, ok := e.item.(*TextItem); ok {
			pos := 0
			if index == "end" {
				pos = len(ti.text)
			} else {
				for _, ch := range index {
					if ch >= '0' && ch <= '9' {
						pos = pos*10 + int(ch-'0')
					}
				}
			}
			ti.SetCursorPos(pos)
			break
		}
	}
	c.scheduleRedraw()
}

// Insert inserts text into a text item at the given index.
// index can be a number or "end".
func (c *Canvas) Insert(tagOrID string, index string, text string) {
	entries := c.resolve(tagOrID)
	for _, e := range entries {
		if ti, ok := e.item.(*TextItem); ok {
			pos := 0
			if index == "end" || index == "insert" {
				pos = ti.cursorPos
			} else {
				for _, ch := range index {
					if ch >= '0' && ch <= '9' {
						pos = pos*10 + int(ch-'0')
					}
				}
			}
			ti.InsertText(pos, text)
			break
		}
	}
	c.scheduleRedraw()
}

// Dchars deletes characters from a text item between first and last indices.
// Supports numeric indices and "insert"/"end" keywords; "insert-1" subtracts 1.
func (c *Canvas) Dchars(tagOrID string, first string, last string) {
	entries := c.resolve(tagOrID)
	for _, e := range entries {
		if ti, ok := e.item.(*TextItem); ok {
			parseIdx := func(s string) int {
				if s == "end" {
					return len(ti.text)
				}
				if s == "insert" {
					return ti.cursorPos
				}
				if s == "insert-1" {
					if ti.cursorPos > 0 {
						return ti.cursorPos - 1
					}
					return 0
				}
				n := 0
				for _, ch := range s {
					if ch >= '0' && ch <= '9' {
						n = n*10 + int(ch-'0')
					}
				}
				return n
			}
			f := parseIdx(first)
			l := parseIdx(last)
			ti.DeleteChars(f, l)
			break
		}
	}
	c.scheduleRedraw()
}

// CanvasX ports "$canvas canvasx x": the canvas coordinate of window x.
func (c *Canvas) CanvasX(x int) float64 { return float64(x + c.xOrigin - c.inset) }

// CanvasY ports "$canvas canvasy y": the canvas coordinate of window y.
func (c *Canvas) CanvasY(y int) float64 { return float64(y + c.yOrigin - c.inset) }
