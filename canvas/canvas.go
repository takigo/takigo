package canvas

import (
	"fmt"
	"iter"
	"slices"

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
	idMap    map[ItemID]*itemEntry
	tagIndex map[string]map[ItemID]*itemEntry // tag name → item IDs for O(1) tag lookup
	nextID   ItemID
	markSeq  uint64 // last stamp handed out by markEntries
	dead     int    // deleted entries still in items (see compact)

	// Items that need placing in positionWindowItems.
	nWindowItems int

	// Scratch buffers reused by the item Display procs within one paint.
	ptBuf    []platform.Point
	coordBuf []float64

	stipples map[string]platform.PixmapID // depth-1 pixmaps by bitmap spec

	// Scroll state.
	xOrigin, yOrigin int
	scrollRegion     [4]int
	hasScrollRegion  bool

	// Display.
	redrawPending    bool
	pixmap           platform.PixmapID
	pixmapW, pixmapH int

	// damage is the area to repaint at the next idle redraw, in canvas
	// coordinates (Tk's redrawX1..redrawY2); damageAll repaints the whole
	// window, border included.
	damage    [4]int
	hasDamage bool
	damageAll bool

	// Item pick / events.
	currentItem *itemEntry
	// pointerState is the modifier/button state of the last pointer event
	// (Tk's canvasPtr->state); while it has a button down, the current item
	// keeps the pointer, as an X grab does for windows.
	pointerState uint
	// leftGrabbed is Tk's LEFT_GRABBED_ITEM: the pointer left the current
	// item with a button down, so it already got its <Leave>.
	leftGrabbed  bool
	repicking    bool
	itemBindings map[string][]itemHandler // by tag
	idBindings   map[ItemID][]itemHandler // by item ID
	closeEnough  float64                  // hit-test tolerance (default 1.0)
	antialias    bool                     // draw shapes with coverage (see smooth.go)

	// Keyboard focus for text items.
	focusItemID ItemID // 0 = none

	// Scrollbar callbacks.
	XScrollCmd func(first, last float64)
	YScrollCmd func(first, last float64)

	// Computed inset (borderWidth + highlightWidth).
	inset int

	// Tk -width/-height: the drawing area, excluding the inset.
	reqW, reqH int
}

// newCanvas returns a Canvas with its item bookkeeping initialised and no
// window; New attaches the window and applies the options.
func newCanvas() *Canvas {
	return &Canvas{
		idMap:        make(map[ItemID]*itemEntry),
		tagIndex:     make(map[string]map[ItemID]*itemEntry),
		nextID:       1,
		itemBindings: make(map[string][]itemHandler),
		idBindings:   make(map[ItemID][]itemHandler),
		closeEnough:  1.0,
	}
}

// markEntries stamps entries with a fresh mark so a pass over the display
// list can test membership with a field compare instead of a set lookup.
// compact drops deleted entries from the display list; every reader of
// c.items calls it first.
func (c *Canvas) compact() {
	if c.dead == 0 {
		return
	}
	c.items = slices.DeleteFunc(c.items, func(e *itemEntry) bool { return e.dead })
	c.dead = 0
}

func (c *Canvas) markEntries(entries []*itemEntry) uint64 {
	c.markSeq++
	for _, e := range entries {
		e.mark = c.markSeq
	}
	return c.markSeq
}

// scratchPoints returns n points for one item's Display proc; the next
// call reuses the slice.
func (c *Canvas) scratchPoints(n int) []platform.Point {
	c.ptBuf = slices.Grow(c.ptBuf[:0], n)[:n]
	return c.ptBuf
}

// CanvasOption configures a Canvas.
type CanvasOption func(*Canvas)

// Width sets -width (a Tk distance: pixels or "10c", "3i", ...).
func Width[L screenunit.Length](w L) CanvasOption {
	return func(c *Canvas) { c.reqW = screenunit.ToPixels(w) }
}

// Height sets -height (a Tk distance).
func Height[L screenunit.Length](h L) CanvasOption {
	return func(c *Canvas) { c.reqH = screenunit.ToPixels(h) }
}

func Background[C color.Spec](name C) CanvasOption {
	return func(c *Canvas) { c.SetBackgroundColor(name) }
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

// Antialias turns anti-aliased drawing of the canvas's shapes on or off.
// It is on by default, and off in an App created with takigo.Classic,
// where items are drawn with the display server's jagged-edged primitives
// exactly as Tk draws them.
func Antialias(on bool) CanvasOption {
	return func(c *Canvas) { c.antialias = on }
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

	c := newCanvas()
	widget.InitBase(&c.Base, w, app)
	w.OnDestroy(c.Destroy)
	w.Class = "Canvas"

	// Canvas defaults (tkUnixDefault.h: -width 10c -height 7c,
	// -highlightthickness 1, normal background).
	c.HighlightWidth = 1
	c.reqW = screenunit.Cm(10).Pixels()
	c.reqH = screenunit.Cm(7).Pixels()
	c.antialias = !widget.Classic(app)

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

	// Like Tk_CreateWindow, the canvas is 1x1 until a geometry manager
	// sizes it; -scrollregion confinement before then depends on that.
	w.Width, w.Height = 1, 1

	// Set window background pixel for child window creation.
	if c.Base.Background != nil {
		w.SetBackgroundPixel(c.Base.Background.Pixel)
	}

	// Set up event handlers.
	bindCanvas(c)

	return c
}

// Display draws the canvas and all its items now.
func (c *Canvas) Display() {
	c.damageAll, c.hasDamage = false, false
	if c.Destroyed {
		return
	}
	w := c.Win
	if w.PlatformID == 0 || !w.IsMapped() {
		return
	}
	winW := w.Width - 2*c.inset
	winH := w.Height - 2*c.inset
	if winW <= 0 || winH <= 0 {
		return
	}
	c.paint(c.xOrigin, c.yOrigin, c.xOrigin+winW, c.yOrigin+winH)

	d := w.Display.Server
	gc := w.GC
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
	c.NeedRedraw = false
}

// redrawDamage repaints what was damaged since the last redraw; it ports
// DisplayCanvas, which redraws only redrawX1..redrawY2 unless the whole
// window (REDRAW_BORDERS) was invalidated.
func (c *Canvas) redrawDamage() {
	c.redrawPending = false
	if c.damageAll {
		c.Display()
		return
	}
	if !c.hasDamage || c.Destroyed {
		return
	}
	c.hasDamage = false
	w := c.Win
	if w.PlatformID == 0 || !w.IsMapped() {
		return
	}
	x1 := max(c.damage[0], c.xOrigin)
	y1 := max(c.damage[1], c.yOrigin)
	x2 := min(c.damage[2], c.xOrigin+w.Width-2*c.inset)
	y2 := min(c.damage[3], c.yOrigin+w.Height-2*c.inset)
	if x1 >= x2 || y1 >= y2 {
		return
	}
	c.paint(x1, y1, x2, y2)
	w.Display.Server.Flush()
}

// paint repaints the visible canvas area x1,y1 .. x2,y2 (canvas
// coordinates): it clears that part of the off-screen pixmap, draws the
// items whose boxes overlap it, and copies it to the window. Items are drawn
// with the same pixmap origin and clip whatever the area, so a partial
// repaint produces the same pixels as a full one.
func (c *Canvas) paint(x1, y1, x2, y2 int) {
	c.compact()
	w := c.Win
	d := w.Display.Server
	winW := w.Width - 2*c.inset
	winH := w.Height - 2*c.inset

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

	// The origin in canvas coordinates that maps to the pixmap origin.
	pixOriginX := c.xOrigin - overdraw
	pixOriginY := c.yOrigin - overdraw

	// Clear the repainted area to the background color.
	bgPixel := uint64(0xFFFFFF)
	if c.Base.Background != nil {
		bgPixel = c.Base.Background.Pixel
	}
	d.SetForeground(gc, bgPixel)
	d.FillRectangle(pxDrawable, gc, x1-pixOriginX, y1-pixOriginY, uint(x2-x1), uint(y2-y1))

	// Clip rectangle in canvas coordinates.
	clipX := pixOriginX
	clipY := pixOriginY
	clipW := pixW
	clipH := pixH

	// With anti-aliasing, shapes are rasterized here and blended into a
	// copy of the repainted area, which goes to the pixmap when an item the
	// display server must draw comes up, and at the end (see smooth.go).
	var smooth *layer
	if c.antialias {
		smooth = newLayer(pixOriginX, pixOriginY, pixW, pixH, x1-pixOriginX, y1-pixOriginY, x2-x1, y2-y1)
		smooth.plain = true
		smooth.bg = bgPixel
	}

	// Draw items bottom to top.
	for _, entry := range c.items {
		item := entry.item
		if item.State() == ItemStateHidden {
			continue
		}

		// Cull by bounding box against the repainted area.
		ix1, iy1, ix2, iy2 := item.BBox()
		if ix2 < x1 || ix1 > x2 || iy2 < y1 || iy1 > y2 {
			continue
		}

		if smooth != nil {
			if si, ok := item.(smoothItem); ok && smooth.paint(d, pxDrawable, si, ix1, iy1, ix2, iy2) {
				continue
			}
			smooth.flush(d, pxDrawable, gc, w.Depth)
			smooth.invalidate()
		}
		item.Display(d, pxDrawable, gc, clipX, clipY, clipW, clipH, pixOriginX, pixOriginY)
	}
	if smooth != nil {
		smooth.flush(d, pxDrawable, gc, w.Depth)
	}

	// Copy the repainted area to the window (accounting for overdraw
	// offset and inset).
	d.CopyArea(pxDrawable, w.Drawable(), gc,
		x1-pixOriginX, y1-pixOriginY, uint(x2-x1), uint(y2-y1),
		x1-c.xOrigin+c.inset, y1-c.yOrigin+c.inset)

	// Position embedded window items.
	if c.nWindowItems > 0 {
		c.positionWindowItems()
	}
}

// positionWindowItems maps and positions all embedded WindowItems relative to
// the canvas window. Items outside the visible area are unmapped.
func (c *Canvas) positionWindowItems() {
	c.compact()
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

// scheduleRedraw schedules a repaint of the whole window at idle time.
func (c *Canvas) scheduleRedraw() {
	c.damageAll = true
	c.scheduleIdle()
}

// eventuallyRedraw ports Tk_CanvasEventuallyRedraw: it adds x1,y1 .. x2,y2
// (canvas coordinates) to the area repainted at idle time.
func (c *Canvas) eventuallyRedraw(x1, y1, x2, y2 int) {
	if x1 >= x2 || y1 >= y2 {
		return
	}
	if c.hasDamage {
		x1 = min(x1, c.damage[0])
		y1 = min(y1, c.damage[1])
		x2 = max(x2, c.damage[2])
		y2 = max(y2, c.damage[3])
	}
	c.damage = [4]int{x1, y1, x2, y2}
	c.hasDamage = true
	c.scheduleIdle()
}

// damageSlop widens an item's box for what it draws just outside it, such
// as a text item's insertion cursor.
const damageSlop = 2

// redrawItems damages the current boxes of entries; callers that change an
// item call it before and after the change so both areas are repainted.
func (c *Canvas) redrawItems(entries ...*itemEntry) {
	for _, e := range entries {
		if e == nil {
			continue
		}
		x1, y1, x2, y2 := e.item.BBox()
		c.eventuallyRedraw(x1-damageSlop, y1-damageSlop, x2+damageSlop, y2+damageSlop)
	}
}

func (c *Canvas) scheduleIdle() {
	if c.redrawPending || c.Destroyed {
		return
	}
	c.redrawPending = true
	c.App.DoWhenIdle(func() {
		if !c.Destroyed {
			c.redrawDamage()
		}
	})
}

// Destroy cleans up the canvas and all its items.
func (c *Canvas) Destroy() {
	c.compact()
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

	for _, p := range c.stipples {
		if p != 0 {
			d.FreePixmap(p)
		}
	}
	c.stipples = nil

	// Free pixmap.
	if c.pixmap != 0 {
		d.FreePixmap(c.pixmap)
		c.pixmap = 0
	}
	window.DestroyWindow(c.Win)
}

// Configure sets canvas options.
func (c *Canvas) Configure(opts ...CanvasOption) error {
	c.BeginOptions()
	for _, opt := range opts {
		opt(c)
	}
	err := c.EndOptions()
	if inset := c.BorderWidth + c.HighlightWidth; inset != c.inset {
		c.xOrigin += inset - c.inset
		c.yOrigin += inset - c.inset
		c.inset = inset
	}
	// Tk relays out at idle, so CanvasSetOrigin still sees the old window
	// size; our geometry managers resize synchronously, so confine first.
	c.setOrigin(c.xOrigin, c.yOrigin)
	if c.Base.Background != nil {
		c.Win.SetBackgroundPixel(c.Base.Background.Pixel)
	}
	if rw, rh := c.reqW+2*c.inset, c.reqH+2*c.inset; rw != c.Win.ReqWidth || rh != c.Win.ReqHeight {
		geometry.GeometryRequest(c.Win, rw, rh)
	}
	c.scheduleRedraw()
	return err
}

// --- Item creation ---

// createItem configures a new item and adds it. Like the option setters,
// a bad option is logged and the options after it are not applied.
func (c *Canvas) createItem(item Item, opts []ItemOption) ItemID {
	c.configureNew(item, opts)
	return c.addItem(item)
}

func (c *Canvas) configureNew(item Item, opts []ItemOption) bool {
	if err := item.Configure(opts); err != nil {
		c.OptionFailed(fmt.Errorf("canvas: create %s: %w", item.Type(), err))
		return false
	}
	return true
}

func (c *Canvas) addItem(item Item) ItemID {
	base := itemBase(item)
	id := c.nextID
	c.nextID++
	base.ID = id
	base.canvas = c

	entry := &itemEntry{id: id, item: item}
	c.items = append(c.items, entry)
	c.idMap[id] = entry
	if _, ok := item.(*WindowItem); ok {
		c.nWindowItems++
	}

	// Register initial tags in the index.
	for _, tag := range base.Tags {
		c.tagIndexAdd(tag, id)
	}

	c.redrawItems(entry)
	return id
}

// tagIndexAdd registers item id under tag in the tag index.
func (c *Canvas) tagIndexAdd(tag string, id ItemID) {
	m := c.tagIndex[tag]
	if m == nil {
		m = make(map[ItemID]*itemEntry)
		c.tagIndex[tag] = m
	}
	m[id] = c.idMap[id]
}

// tagIndexRemove unregisters item id from tag in the tag index.
func (c *Canvas) tagIndexRemove(tag string, id ItemID) {
	if m := c.tagIndex[tag]; m != nil {
		delete(m, id)
		if len(m) == 0 {
			delete(c.tagIndex, tag)
		}
	}
}

// CreateRectangle creates a rectangle item.
func (c *Canvas) CreateRectangle(x1, y1, x2, y2 float64, opts ...ItemOption) ItemID {
	item := newRectOvalItem("rectangle", x1, y1, x2, y2, c)
	return c.createItem(item, opts)
}

// CreateOval creates an oval item.
func (c *Canvas) CreateOval(x1, y1, x2, y2 float64, opts ...ItemOption) ItemID {
	item := newRectOvalItem("oval", x1, y1, x2, y2, c)
	return c.createItem(item, opts)
}

// CreateLine creates a line item.
func (c *Canvas) CreateLine(coords []float64, opts ...ItemOption) ItemID {
	item := newLineItem(coords, c)
	return c.createItem(item, opts)
}

// CreatePolygon creates a polygon item.
func (c *Canvas) CreatePolygon(coords []float64, opts ...ItemOption) ItemID {
	item := newPolygonItem(coords, c)
	return c.createItem(item, opts)
}

// CreateArc creates an arc item.
func (c *Canvas) CreateArc(x1, y1, x2, y2 float64, opts ...ItemOption) ItemID {
	item := newArcItem(x1, y1, x2, y2, c)
	return c.createItem(item, opts)
}

// CreateText creates a text item.
func (c *Canvas) CreateText(x, y float64, opts ...ItemOption) ItemID {
	item := newTextItem(x, y, c)
	return c.createItem(item, opts)
}

// CreateImage creates an image item.
// CreateWindow embeds a child window at canvas position (x, y).
// The window must be a child of the canvas window (created with canvas as parent).
func (c *Canvas) CreateWindow(x, y float64, w *window.Window, opts ...ItemOption) ItemID {
	item := newWindowItem(x, y, w, c)
	if c.configureNew(item, opts) {
		item.updateBBox()
	}
	return c.addItem(item)
}

// CreateBitmap creates a 1-bit XBM bitmap item at (x, y).
func (c *Canvas) CreateBitmap(x, y float64, xbm *XBMData, opts ...ItemOption) ItemID {
	item := newBitmapItem(x, y, xbm, c)
	return c.createItem(item, opts)
}

func (c *Canvas) CreateImage(x, y float64, opts ...ItemOption) ItemID {
	item := newImageItem(x, y, c)
	return c.createItem(item, opts)
}

// --- Item manipulation ---

// Delete removes all items matching tagOrID.
func (c *Canvas) Delete[S Selector](sel S) {
	tagOrID := selectorString(sel)
	entries := c.resolve(tagOrID)
	if len(entries) == 0 {
		return
	}

	c.redrawItems(entries...)
	d := c.Win.Display.Server
	stamp := c.markEntries(entries)
	for _, e := range entries {
		e.item.Delete(d)
		delete(c.idMap, e.id)
		if _, ok := e.item.(*WindowItem); ok {
			c.nWindowItems--
		}
		if base := itemBase(e.item); base != nil {
			for _, tag := range base.Tags {
				c.tagIndexRemove(tag, e.id)
			}
		}
		// DeleteItems drops the item's bindings and its focus.
		delete(c.idBindings, e.id)
		if c.focusItemID == e.id {
			c.focusItemID = 0
		}
	}
	// The entries leave the display list at the next compact, so deleting
	// items one by one costs O(1) each rather than a pass over the list.
	for _, e := range entries {
		e.dead = true
	}
	c.dead += len(entries)
	if c.dead > len(c.items)/2 {
		c.compact() // amortised: a create/delete loop without paints stays bounded
	}

	// Clear current item if deleted.
	if c.currentItem != nil && c.currentItem.mark == stamp {
		c.currentItem = nil
	}
}

// CurrentItem returns the ID of the item under the mouse cursor, or -1 if none.
func (c *Canvas) CurrentItem() ItemID {
	if c.currentItem == nil {
		return -1
	}
	return c.currentItem.id
}

// Move translates all items matching tagOrID by (dx, dy).
func (c *Canvas) Move[S Selector](sel S, dx, dy float64) {
	tagOrID := selectorString(sel)
	for _, entry := range c.resolve(tagOrID) {
		c.redrawItems(entry)
		entry.item.Translate(dx, dy)
		c.redrawItems(entry)
	}
}

// Scale scales all items matching tagOrID about (ox, oy).
func (c *Canvas) Scale[S Selector](sel S, ox, oy, sx, sy float64) {
	tagOrID := selectorString(sel)
	for _, entry := range c.resolve(tagOrID) {
		c.redrawItems(entry)
		entry.item.Scale(ox, oy, sx, sy)
		c.redrawItems(entry)
	}
}

// Raise moves items matching tagOrID to the top of the display list.
func (c *Canvas) Raise[S Selector](sel S) {
	tagOrID := selectorString(sel)
	c.compact()
	entries := c.resolve(tagOrID)
	if len(entries) == 0 {
		return
	}
	// In-place stable partition: kept items first, then entries, which
	// resolve already returns in display order.
	stamp := c.markEntries(entries)
	j := 0
	for _, e := range c.items {
		if e.mark != stamp {
			c.items[j] = e
			j++
		}
	}
	copy(c.items[j:], entries)
	c.redrawItems(entries...)
}

// Lower moves items matching tagOrID to the bottom of the display list.
func (c *Canvas) Lower[S Selector](sel S) {
	tagOrID := selectorString(sel)
	c.compact()
	entries := c.resolve(tagOrID)
	if len(entries) == 0 {
		return
	}
	// In-place stable partition from the back: kept items shifted right,
	// then entries (in display order) at the front.
	stamp := c.markEntries(entries)
	j := len(c.items)
	for _, e := range slices.Backward(c.items) {
		if e.mark != stamp {
			j--
			c.items[j] = e
		}
	}
	copy(c.items[:j], entries)
	c.redrawItems(entries...)
}

// AddTag adds a tag to all items matching tagOrID.
func (c *Canvas) AddTag[S Selector](newTag string, sel S) {
	tagOrID := selectorString(sel)
	for _, entry := range c.resolve(tagOrID) {
		if base := itemBase(entry.item); base != nil {
			base.AddTag(newTag)
		}
	}
}

// DeleteTag removes a tag from all items matching tagOrID.
func (c *Canvas) DeleteTag[S Selector](tag string, sel S) {
	tagOrID := selectorString(sel)
	for _, entry := range c.resolve(tagOrID) {
		if base := itemBase(entry.item); base != nil {
			base.RemoveTag(tag)
		}
	}
}

// ItemConfigure configures items matching tagOrID.
func (c *Canvas) ItemConfigure[S Selector](sel S, opts ...ItemOption) error {
	tagOrID := selectorString(sel)
	for _, entry := range c.resolve(tagOrID) {
		c.redrawItems(entry)
		err := entry.item.Configure(opts)
		c.redrawItems(entry)
		if err != nil {
			return err
		}
	}
	return nil
}

// ItemCoords returns coordinates of the first item matching tagOrID.
func (c *Canvas) ItemCoords[S Selector](sel S) []float64 {
	tagOrID := selectorString(sel)
	entries := c.resolve(tagOrID)
	if len(entries) == 0 {
		return nil
	}
	return entries[0].item.Coords()
}

// SetItemCoords sets coordinates on the first item matching tagOrID.
func (c *Canvas) SetItemCoords[S Selector](sel S, coords []float64) error {
	tagOrID := selectorString(sel)
	entries := c.resolve(tagOrID)
	if len(entries) == 0 {
		return nil
	}
	c.redrawItems(entries[0])
	err := entries[0].item.SetCoords(coords)
	c.redrawItems(entries[0])
	return err
}

// Items iterates over the IDs of the items sel selects, in display order.
// It works on a snapshot, so items may be deleted during the iteration.
func (c *Canvas) Items[S Selector](sel S) iter.Seq[ItemID] {
	entries := c.resolve(selectorString(sel))
	return func(yield func(ItemID) bool) {
		for _, e := range entries {
			if !yield(e.id) {
				return
			}
		}
	}
}

// FindWithTag returns the IDs of the items sel selects, in display order.
func (c *Canvas) FindWithTag[S Selector](sel S) []ItemID {
	tagOrID := selectorString(sel)
	entries := c.resolve(tagOrID)
	ids := make([]ItemID, len(entries))
	for i, e := range entries {
		ids[i] = e.id
	}
	return ids
}

// Find returns item IDs that satisfy the given search mode.
// Supported modes: "all", "closest" (args: x, y), "enclosed" (args: x1,y1,x2,y2),
// "overlapping" (args: x1,y1,x2,y2).
func (c *Canvas) Find(mode string, args ...float64) []ItemID {
	c.compact()
	switch mode {
	case "all":
		ids := make([]ItemID, len(c.items))
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
			return []ItemID{entry.id}
		}
		return nil
	case "enclosed":
		if len(args) < 4 {
			return nil
		}
		var ids []ItemID
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
		var ids []ItemID
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
func (c *Canvas) BBox[S Selector](sel S) (x1, y1, x2, y2 int) {
	tagOrID := selectorString(sel)
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
func (c *Canvas) GetTags[S Selector](sel S) []string {
	tagOrID := selectorString(sel)
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
func (c *Canvas) Focus[S Selector](sel S) {
	tagOrID := selectorString(sel)
	// Clear previous focus.
	if c.focusItemID != 0 {
		if prev, ok := c.idMap[c.focusItemID]; ok {
			if ti, ok := prev.item.(*TextItem); ok {
				ti.hasFocus = false
			}
			c.redrawItems(prev)
		}
		c.focusItemID = 0
	}
	if tagOrID == "" {
		return
	}
	entries := c.resolve(tagOrID)
	for _, e := range entries {
		if ti, ok := e.item.(*TextItem); ok {
			ti.hasFocus = true
			c.focusItemID = e.id
			c.redrawItems(e)
			// Make canvas window focusable and give it X11 focus.
			c.Win.Flags |= window.FlagFocusable
			widget.Focus(c.App, c.Win)
			break
		}
	}
}

// ICursor sets the insertion cursor position in a text item.
// index can be a number or "end".
func (c *Canvas) ICursor[S Selector](sel S, index string) {
	tagOrID := selectorString(sel)
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
			c.redrawItems(e)
			ti.SetCursorPos(pos)
			c.redrawItems(e)
			break
		}
	}
}

// Insert inserts text into a text item at the given index.
// index can be a number or "end".
func (c *Canvas) Insert[S Selector](sel S, index string, text string) {
	tagOrID := selectorString(sel)
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
			c.redrawItems(e)
			ti.InsertText(pos, text)
			c.redrawItems(e)
			break
		}
	}
}

// Dchars deletes characters from a text item between first and last indices.
// Supports numeric indices and "insert"/"end" keywords; "insert-1" subtracts 1.
func (c *Canvas) Dchars[S Selector](sel S, first string, last string) {
	tagOrID := selectorString(sel)
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
			c.redrawItems(e)
			ti.DeleteChars(f, l)
			c.redrawItems(e)
			break
		}
	}
}

// CanvasX ports "$canvas canvasx x": the canvas coordinate of window x.
func (c *Canvas) CanvasX(x int) float64 { return float64(x + c.xOrigin - c.inset) }

// CanvasY ports "$canvas canvasy y": the canvas coordinate of window y.
func (c *Canvas) CanvasY(y int) float64 { return float64(y + c.yOrigin - c.inset) }
