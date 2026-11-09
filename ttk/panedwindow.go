// Ports tk/generic/ttk/ttkPanedwindow.c (with the update scheduling of
// ttkManager.c) and tk/library/ttk/panedwindow.tcl.
package ttk

import (
	"errors"
	"fmt"
	"slices"

	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/geometry"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/window"
)

// paneRecord is ttkPanedwindow.c's Pane: the request size, the position
// of the following sash and the -weight.
type paneRecord struct {
	win     *window.Window
	reqSize int
	sashPos int
	weight  int
}

// Panedwindow is a themed panedwindow (ttk::panedwindow): its panes are
// separated by sashes the user drags, and extra space goes to the panes
// by -weight.
type Panedwindow struct {
	TtkWidget

	Orient        Orient
	width, height int // -width, -height; 0 is the panes' size

	panes         []*paneRecord
	sashLayout    *Layout
	sashThickness int

	// ttkManager.c's update state.
	pending, resizeReq, relayoutReq bool

	// panedwindow.tcl's State.
	pressed      bool
	pressX       int
	pressY       int
	pressSash    int
	pressSashPos int
}

// PanedwindowOption configures a Panedwindow.
type PanedwindowOption func(*Panedwindow)

// PanedwindowOrient sets -orient (Tk's default is vertical). Like Tk, it
// can only be set when the widget is created.
func PanedwindowOrient(o Orient) PanedwindowOption {
	return func(pw *Panedwindow) {
		if pw.Win.Value(paneCreatedKey) == nil {
			pw.Orient = o
		}
	}
}

// PanedwindowWidth sets -width in pixels; 0 uses the panes' width.
func PanedwindowWidth(px int) PanedwindowOption { return func(pw *Panedwindow) { pw.width = px } }

// PanedwindowHeight sets -height in pixels; 0 uses the panes' height.
func PanedwindowHeight(px int) PanedwindowOption { return func(pw *Panedwindow) { pw.height = px } }

// PanedwindowStyle sets -style.
func PanedwindowStyle(name string) PanedwindowOption {
	return func(pw *Panedwindow) {
		pw.StyleName = name
		pw.styleSet = true
	}
}

// paneCreatedKey marks a panedwindow whose -orient became read-only.
var paneCreatedKey = new(window.ValueKey)

// PaneOption sets a pane option.
type PaneOption func(*paneRecord)

// PaneWeight sets the pane's -weight: how much of the extra or missing
// space it takes, relative to the other panes.
func PaneWeight(w int) PaneOption { return func(p *paneRecord) { p.weight = w } }

// NewPanedwindow creates a themed panedwindow.
func NewPanedwindow(parent widget.Caregiver, name string, opts ...PanedwindowOption) *Panedwindow {
	app := parent.AppContext()
	win := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(win)

	pw := &Panedwindow{Orient: Vertical, sashThickness: 1}
	InitTtkWidget(&pw.TtkWidget, win, app, "TPanedwindow")
	pw.reconfigure = func() { _ = pw.Configure() }
	pw.DisplayFunc = pw.Display
	for _, opt := range opts {
		opt(pw)
	}
	win.SetValue(paneCreatedKey, true)
	pw.RefreshTheme()
	pw.requestSize()
	bindPanedwindow(pw, app)
	return pw
}

// RefreshTheme rebuilds the layout and the sash sublayout (PanedGetLayout).
func (pw *Panedwindow) RefreshTheme() {
	pw.TtkWidget.RefreshTheme()
	pw.buildSashLayout()
}

// buildSashLayout makes the .Vertical.Sash or .Horizontal.Sash sublayout of
// the widget's style and takes the sash thickness from its size.
func (pw *Panedwindow) buildSashLayout() {
	if pw.Theme == nil {
		return
	}
	// A horizontal panedwindow has vertical sashes.
	name := pw.StyleName + ".Vertical.Sash"
	if pw.Orient == Vertical {
		name = pw.StyleName + ".Horizontal.Sash"
	}
	tmpl := pw.Theme.GetLayout(name)
	if tmpl == nil {
		pw.sashLayout = nil
		return
	}
	style := pw.Theme.ResolveStyle(name)
	ctx := &DrawContext{Display: pw.Context.Display, Depth: pw.Context.Depth, Style: style}
	pw.sashLayout = NewLayout(tmpl, pw.Theme, ctx, style)
	w, h := pw.sashLayout.Size(pw.State)
	if pw.Orient == Horizontal {
		pw.sashThickness = w
	} else {
		pw.sashThickness = h
	}
}

// Configure sets options after creation.
func (pw *Panedwindow) Configure(opts ...PanedwindowOption) error {
	width, height := pw.width, pw.height
	return configure(&pw.TtkWidget, pw, opts, nil, func() {
		// UseTheme re-themes the TtkWidget, then reconfigures.
		pw.buildSashLayout()
		pw.computeSize()
		// PanedPostConfigure: a new -width/-height places the sashes for it.
		if pw.width != width || pw.height != height {
			w, h := pw.Win.Width, pw.Win.Height
			if pw.width > 0 {
				w = pw.width
			}
			if pw.height > 0 {
				h = pw.height
			}
			pw.placeSashes(w, h)
		}
	})
}

// --- Panes ---

func (pw *Panedwindow) along(w *window.Window) int {
	if pw.Orient == Horizontal {
		return w.ReqWidth
	}
	return w.ReqHeight
}

func (pw *Panedwindow) index(child *window.Window) int {
	return slices.IndexFunc(pw.panes, func(p *paneRecord) bool { return p.win == child })
}

// Add adds child as the last pane ("add").
func (pw *Panedwindow) Add(child window.Windower, opts ...PaneOption) error {
	return pw.Insert(len(pw.panes), child, opts...)
}

// Insert puts child at index, moving it there when it already is a pane
// ("insert"); an index past the end appends.
func (pw *Panedwindow) Insert(index int, child window.Windower, opts ...PaneOption) error {
	w := child.Window()
	if src := pw.index(w); src >= 0 {
		index = min(max(index, 0), len(pw.panes)-1)
		p := pw.panes[src]
		pw.panes = slices.Insert(slices.Delete(pw.panes, src, src+1), index, p)
		for _, opt := range opts {
			opt(p)
		}
		pw.schedule(true, true)
		return nil
	}
	if w.Parent != pw.Win {
		return fmt.Errorf("can't add %s as content of %s", w.PathName, pw.Win.PathName)
	}
	p := &paneRecord{win: w, reqSize: pw.along(w)}
	for _, opt := range opts {
		opt(p)
	}
	if p.weight < 0 {
		return errors.New("-weight must be non-negative")
	}
	index = min(max(index, 0), len(pw.panes))
	pw.panes = slices.Insert(pw.panes, index, p)
	geometry.ManageGeometry(w, &panedGeomMgr{pw})
	pw.schedule(true, true)
	return nil
}

// Forget removes child from the panes ("forget") and unmaps it.
func (pw *Panedwindow) Forget(child window.Windower) {
	w := child.Window()
	i := pw.index(w)
	if i < 0 {
		return
	}
	pw.panes = slices.Delete(pw.panes, i, i+1)
	w.GeomManager = nil
	if w.IsMapped() {
		w.Display.Server.UnmapWindow(w.PlatformID)
		window.MarkUnmapped(w)
	}
	pw.schedule(true, true)
}

// Pane sets options of the pane holding child ("pane").
func (pw *Panedwindow) Pane(child window.Windower, opts ...PaneOption) error {
	i := pw.index(child.Window())
	if i < 0 {
		return fmt.Errorf("%s is not managed by %s", child.Window().PathName, pw.Win.PathName)
	}
	p := *pw.panes[i]
	for _, opt := range opts {
		opt(&p)
	}
	if p.weight < 0 {
		return errors.New("-weight must be non-negative")
	}
	*pw.panes[i] = p
	pw.schedule(true, false)
	return nil
}

// PaneWeightOf returns the -weight of the pane holding child.
func (pw *Panedwindow) PaneWeightOf(child window.Windower) int {
	if i := pw.index(child.Window()); i >= 0 {
		return pw.panes[i].weight
	}
	return 0
}

// Panes returns the managed windows ("panes").
func (pw *Panedwindow) Panes() []*window.Window {
	out := make([]*window.Window, len(pw.panes))
	for i, p := range pw.panes {
		out[i] = p.win
	}
	return out
}

// SashPos returns the position of sash i ("sashpos $i").
func (pw *Panedwindow) SashPos(i int) (int, error) {
	if i < 0 || i >= len(pw.panes)-1 {
		return 0, fmt.Errorf("sash index %d out of range", i)
	}
	return pw.panes[i].sashPos, nil
}

// SetSashPos moves sash i to pos, shoving the neighbouring sashes as
// needed, and returns where it ended up ("sashpos $i $pos").
func (pw *Panedwindow) SetSashPos(i, pos int) (int, error) {
	if i < 0 || i >= len(pw.panes)-1 {
		return 0, fmt.Errorf("sash index %d out of range", i)
	}
	if pos < pw.panes[i].sashPos {
		pw.shoveUp(i, pos)
	} else {
		pw.shoveDown(i, pos)
	}
	pw.adjustPanes()
	pw.schedule(false, true)
	return pw.panes[i].sashPos, nil
}

// Identify returns the index of the sash at x, y, or -1 ("identify").
func (pw *Panedwindow) Identify(x, y int) int {
	pos := y
	if pw.Orient == Horizontal {
		pos = x
	}
	for i := range len(pw.panes) - 1 {
		if s := pw.panes[i].sashPos; s <= pos && pos <= s+pw.sashThickness {
			return i
		}
	}
	return -1
}

// --- Layout ---

// shoveUp ports ShoveUp: place sash i at pos, pushing earlier sashes up.
func (pw *Panedwindow) shoveUp(i, pos int) int {
	if i == 0 {
		pos = max(pos, 0)
	} else if prev := pw.panes[i-1]; pos < prev.sashPos+pw.sashThickness {
		pos = pw.shoveUp(i-1, pos-pw.sashThickness) + pw.sashThickness
	}
	pw.panes[i].sashPos = pos
	return pos
}

// shoveDown ports ShoveDown, stopping at the sentinel sash.
func (pw *Panedwindow) shoveDown(i, pos int) int {
	if i == len(pw.panes)-1 {
		pos = pw.panes[i].sashPos
	} else if next := pw.panes[i+1]; pos+pw.sashThickness > next.sashPos {
		pos = pw.shoveDown(i+1, pos+pw.sashThickness) - pw.sashThickness
	}
	pw.panes[i].sashPos = pos
	return pos
}

// panedSize ports PanedSize: the panes' request sizes plus the sashes.
func (pw *Panedwindow) panedSize() (int, int) {
	width, height := 0, 0
	for _, p := range pw.panes {
		if pw.Orient == Horizontal {
			height = max(height, p.win.ReqHeight)
			width += p.reqSize
		} else {
			width = max(width, p.win.ReqWidth)
			height += p.reqSize
		}
	}
	if n := len(pw.panes); n > 1 {
		if pw.Orient == Horizontal {
			width += (n - 1) * pw.sashThickness
		} else {
			height += (n - 1) * pw.sashThickness
		}
	}
	if pw.width > 0 {
		width = pw.width
	}
	if pw.height > 0 {
		height = pw.height
	}
	return width, height
}

func (pw *Panedwindow) computeSize() {
	w, h := pw.panedSize()
	pw.Win.ReqWidth, pw.Win.ReqHeight = max(w, 1), max(h, 1)
}

func (pw *Panedwindow) requestSize() { pw.resize(pw.computeSize) }

// adjustPanes ports AdjustPanes: request sizes from the sash positions.
func (pw *Panedwindow) adjustPanes() {
	pos := 0
	for _, p := range pw.panes {
		p.reqSize = max(p.sashPos-pos, 0)
		pos = p.sashPos + pw.sashThickness
	}
}

// placeSashes ports PlaceSashes: every pane gets its request size, the
// difference to the available space is shared by -weight (a collapsed
// pane takes none), and the last sash is pinned to the end.
func (pw *Panedwindow) placeSashes(width, height int) {
	n := len(pw.panes)
	if n == 0 {
		return
	}
	available := height
	if pw.Orient == Horizontal {
		available = width
	}
	reqSize, totalWeight := 0, 0
	for _, p := range pw.panes {
		reqSize += p.reqSize
		if p.reqSize != 0 {
			totalWeight += p.weight
		}
	}
	difference := available - reqSize - pw.sashThickness*(n-1)
	delta, remainder := 0, 0
	if totalWeight != 0 {
		delta, remainder = difference/totalWeight, difference%totalWeight
		if remainder < 0 {
			delta--
			remainder += totalWeight
		}
	}
	pos := 0
	for _, p := range pw.panes {
		weight := 0
		if p.reqSize != 0 {
			weight = p.weight
		}
		size := p.reqSize + delta*weight
		weight = min(weight, remainder)
		remainder -= weight
		size = max(size+weight, 0)
		pos += size
		p.sashPos = pos
		pos += pw.sashThickness
	}
	pw.shoveUp(n-1, available)
}

// placePanes ports PlacePanes: a pane with no room is unmapped.
func (pw *Panedwindow) placePanes() {
	width, height := pw.Win.Width, pw.Win.Height
	pos := 0
	for _, p := range pw.panes {
		size := p.sashPos - pos
		switch {
		case size <= 0:
			if p.win.IsMapped() {
				p.win.Display.Server.UnmapWindow(p.win.PlatformID)
				window.MarkUnmapped(p.win)
			}
		case pw.Orient == Horizontal:
			geometry.PlaceContent(pw.Win, p.win, pos, 0, size, height)
		default:
			geometry.PlaceContent(pw.Win, p.win, 0, pos, width, size)
		}
		pos = p.sashPos + pw.sashThickness
	}
}

// relayout is PanedPlaceContent.
func (pw *Panedwindow) relayout() {
	pw.placeSashes(pw.Win.Width, pw.Win.Height)
	pw.placePanes()
	pw.redisplay()
}

// schedule ports ttkManager.c's ScheduleUpdate and ManagerIdleProc: a
// size recomputation and a relayout run at idle time, and after a size
// request the relayout waits a further idle round, so requests from
// nested panes settle before any pane is placed.
func (pw *Panedwindow) schedule(resize, relayout bool) {
	pw.resizeReq = pw.resizeReq || resize
	pw.relayoutReq = pw.relayoutReq || relayout
	if pw.pending {
		return
	}
	pw.pending = true
	pw.App.DoWhenIdle(pw.idle)
}

func (pw *Panedwindow) idle() {
	pw.pending = false
	if pw.Destroyed() {
		return
	}
	if pw.resizeReq {
		pw.resizeReq = false
		pw.requestSize()
		pw.schedule(false, true)
	}
	if pw.relayoutReq && !pw.pending {
		pw.relayoutReq = false
		pw.relayout()
	}
}

// panedGeomMgr is the panedwindow as its panes' geometry manager.
type panedGeomMgr struct{ pw *Panedwindow }

func (m *panedGeomMgr) Name() string { return "panedwindow" }

// RequestProc ports PaneRequest: only an unmapped pane's request counts,
// so a mapped pane does not jump while the user drags a sash.
func (m *panedGeomMgr) RequestProc(content *window.Window) {
	pw := m.pw
	if i := pw.index(content); i >= 0 && !content.IsMapped() {
		pw.panes[i].reqSize = pw.along(content)
	}
	pw.schedule(true, false)
}

// LostContentProc drops a pane that was destroyed or taken over.
func (m *panedGeomMgr) LostContentProc(content *window.Window) {
	pw := m.pw
	if i := pw.index(content); i >= 0 {
		pw.panes = slices.Delete(pw.panes, i, i+1)
		if !pw.Destroyed() {
			pw.schedule(true, true)
		}
	}
}

// --- Display ---

// Display draws the background layout, then each sash (PanedDisplay).
func (pw *Panedwindow) Display() {
	if pw.Destroyed() || pw.Layout == nil {
		return
	}
	win := pw.Win
	width, height := win.Width, win.Height
	if win.PlatformID == 0 || width <= 0 || height <= 0 {
		return
	}
	d := win.Display.Server
	gc := win.GC
	drawable := pw.backBuffer(width, height)
	if drawable == 0 {
		return
	}
	d.SetForeground(gc, LookupColor(pw.Context.Style, "-background", pw.State, 0xd9d9d9))
	d.FillRectangle(drawable, gc, 0, 0, uint(width), uint(height))
	args := DrawArgs{Display: d, Drawable: drawable, GC: gc}
	pw.Layout.Place(pw.State, Box{0, 0, width, height})
	pw.Layout.Draw(pw.State, args)
	if pw.sashLayout != nil {
		for i := range len(pw.panes) - 1 {
			pw.sashLayout.Place(pw.State, pw.sashBox(i))
			pw.sashLayout.Draw(pw.State, args)
		}
	}
	d.CopyArea(drawable, win.Drawable(), gc, 0, 0, uint(width), uint(height), 0, 0)
	d.Flush()
}

// sashBox is SashLayout's parcel for the sash after pane i.
func (pw *Panedwindow) sashBox(i int) Box {
	pos := pw.panes[i].sashPos
	if pw.Orient == Horizontal {
		return Box{pos, 0, pw.sashThickness, pw.Win.Height}
	}
	return Box{0, pos, pw.Win.Width, pw.sashThickness}
}

// --- Bindings (panedwindow.tcl) ---

func bindPanedwindow(pw *Panedwindow, app widget.AppContext) {
	d := app.Dispatcher()
	win := pw.Win
	// ttkManager.c's ManagerEventHandler relayouts on a size change.
	d.Bind(win.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			pw.relayout()
		}
	})
	d.Bind(win.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button != 1 {
			return
		}
		sash := pw.Identify(ev.X, ev.Y)
		if pw.pressed = sash >= 0; !pw.pressed {
			return
		}
		pw.pressX, pw.pressY, pw.pressSash = ev.X, ev.Y, sash
		pw.pressSashPos = pw.panes[sash].sashPos
	})
	d.Bind(win.PlatformID, event.MotionMask, func(ev *event.Event) {
		if !pw.pressed || pw.pressSash >= len(pw.panes)-1 {
			return
		}
		delta := ev.Y - pw.pressY
		if pw.Orient == Horizontal {
			delta = ev.X - pw.pressX
		}
		_, _ = pw.SetSashPos(pw.pressSash, pw.pressSashPos+delta)
	})
	d.Bind(win.PlatformID, event.ButtonReleaseMask, func(ev *event.Event) {
		if ev.Button == 1 {
			pw.pressed = false
		}
	})
}
