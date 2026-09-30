// Package panedwindow implements a paned window widget that manages
// resizable child panes separated by draggable sashes.
// It ports tk/generic/tkPanedWindow.c.
package panedwindow

import (
	"slices"

	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// pwGeomMgr implements window.GeomManager for panedwindow pane children.
// It receives RequestProc calls from child panes when they resize,
// and propagates the aggregated natural size upward.
type pwGeomMgr struct {
	pw *PanedWindow
}

func (m *pwGeomMgr) Name() string { return "panedwindow" }

func (m *pwGeomMgr) RequestProc(content *window.Window) {
	pw := m.pw
	if pw.Weighted {
		pw.ttkSchedule(true, false)
		return
	}
	// PanedWindowReqProc: an unmapped panedwindow adopts the request as
	// the pane size; once mapped, requests only re-arrange.
	if pw.Win.IsMapped() {
		pw.scheduleArrange()
		return
	}
	for i := range pw.panes {
		if p := &pw.panes[i]; p.win == content && !p.userSet {
			p.size = pw.reqSize(content)
		}
	}
	pw.propagateReqSize()
}

// LostContentProc drops a pane that was destroyed or taken over by
// another manager, as PanedWindowLostPaneProc does.
func (m *pwGeomMgr) LostContentProc(content *window.Window) {
	pw := m.pw
	i := slices.IndexFunc(pw.panes, func(p pane) bool { return p.win == content })
	if i < 0 {
		return
	}
	pw.panes = slices.Delete(pw.panes, i, i+1)
	if !pw.Win.IsDestroyed() {
		pw.contentChanged()
	}
}

// propagateReqSize computes the panedwindow's natural size from its panes
// and calls geometry.GeometryRequest so the pack manager above can resize.
func (pw *PanedWindow) propagateReqSize() {
	if len(pw.panes) == 0 {
		return
	}
	totalSashSpace := (len(pw.panes) - 1) * pw.SashWidth
	var reqW, reqH int
	if pw.Orient == Horizontal {
		for _, p := range pw.panes {
			reqW += p.win.ReqWidth
			if p.win.ReqHeight > reqH {
				reqH = p.win.ReqHeight
			}
		}
		reqW += totalSashSpace
	} else {
		for _, p := range pw.panes {
			reqH += p.win.ReqHeight
			if p.win.ReqWidth > reqW {
				reqW = p.win.ReqWidth
			}
		}
		reqH += totalSashSpace
	}
	reqW += 2 * pw.BorderWidth
	reqH += 2 * pw.BorderWidth
	if reqW < 1 {
		reqW = 1
	}
	if reqH < 1 {
		reqH = 1
	}
	geometry.GeometryRequest(pw.Win, reqW, reqH)
}

// Orient specifies the paned window orientation.
type Orient int

const (
	Horizontal Orient = iota // panes side by side
	Vertical                 // panes stacked
)

// pane holds information about a managed child pane.
type pane struct {
	win     *window.Window
	minSize int
	size    int // Tk paneWidth/paneHeight: requested or sash-set size, 0 = unset
	disp    int // size after the last ArrangePanes (stretch applied)
	stretch Stretch
	weight  int  // ttk pane -weight (Weighted mode only)
	userSet bool // size fixed by a sash move (Tk's -width/-height)
}

// Stretch is Tk's pane -stretch: which panes absorb extra or missing space.
type Stretch int

const (
	StretchLast Stretch = iota
	StretchFirst
	StretchMiddle
	StretchAlways
	StretchNever
)

func isStretchable(s Stretch, i, first, last int) bool {
	switch s {
	case StretchAlways:
		return true
	case StretchFirst:
		return i == first
	case StretchLast:
		return i == last
	case StretchMiddle:
		return i != first && i != last
	}
	return false
}

// PanedWindow manages resizable child panes separated by sashes.
type PanedWindow struct {
	widget.Base

	Orient     Orient
	panes      []pane
	SashWidth  int
	HandleSize int
	// FlatSash renders a thin flat sash instead of a 3D raised sash.
	FlatSash bool
	// Weighted selects ttk::panedwindow's layout (ttkPanedwindow.c
	// PlaceSashes): panes grow by -weight and the last pane absorbs the
	// rest, instead of the classic -stretch rules.
	Weighted bool
	GripSize int // grip length for FlatSash

	// mapped mirrors Tk_IsMapped for ArrangePanes: set on the first
	// MapNotify, which the event loop delivers after setup code has run.
	mapped bool

	arrangePending bool // an ArrangePanes is scheduled for idle time

	// Weighted mode's ttkManager.c update state.
	ttkPending, ttkResize, ttkRelayout bool

	// ShowHandle is Tk's -showhandle (default off): draw a raised sash and
	// handle instead of the default flat, invisible sash.
	ShowHandle bool

	// Sash interaction.
	dragSash      int // index of sash being dragged, -1 = none
	dragStartPos  int
	dragStartSize int

	// geomMgr is the geometry manager instance for this panedwindow's panes.
	geomMgr *pwGeomMgr
}

// PanedWindowOption configures a PanedWindow.
type PanedWindowOption func(*PanedWindow)

func OrientOpt(o Orient) PanedWindowOption {
	return func(pw *PanedWindow) { pw.Orient = o }
}

func SashWidthOpt(w int) PanedWindowOption {
	return func(pw *PanedWindow) { pw.SashWidth = w }
}

func HandleSizeOpt(s int) PanedWindowOption {
	return func(pw *PanedWindow) { pw.HandleSize = s }
}

func Background(name string) PanedWindowOption {
	return func(pw *PanedWindow) { pw.SetBackgroundName(name) }
}

// --- Ttk-compatible aliases (prefix with Panedwindow) for consistent naming ---
// These aliases match the naming convention used by ttk widgets
// allowing consistent option naming when both classic and ttk widgets are used.

// PanedwindowOrientOpt is an alias for OrientOpt.
var PanedwindowOrientOpt = OrientOpt

// PanedwindowSashWidthOpt is an alias for SashWidthOpt.
var PanedwindowSashWidthOpt = SashWidthOpt

// PanedwindowHandleSizeOpt is an alias for HandleSizeOpt.
var PanedwindowHandleSizeOpt = HandleSizeOpt

// PanedwindowBackground is an alias for Background.
var PanedwindowBackground = Background

// New creates a new PanedWindow widget.
func New(parent widget.Caregiver, name string, opts ...PanedWindowOption) *PanedWindow {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 200, 200)
	window.MakeWindowExist(w)

	pw := &PanedWindow{
		Orient:     Horizontal,
		SashWidth:  3, // DEF_PANEDWINDOW_SASHWIDTH
		HandleSize: 8,
		dragSash:   -1,
	}
	pw.geomMgr = &pwGeomMgr{pw: pw}
	widget.InitBase(&pw.Base, w, app)
	pw.SetDisplayProc(pw.display)
	w.Class = "Panedwindow"
	pw.BorderWidth = widget.DefBorderWidth
	pw.Relief = option.ReliefFlat

	for _, opt := range opts {
		opt(pw)
	}

	if pw.Background != nil {
		w.SetBackgroundPixel(pw.Background.Pixel)
	}

	bindPanedWindow(pw, app)
	return pw
}

// Add adds a child window as a pane.
func (pw *PanedWindow) Add(child *window.Window, minSize int) {
	if minSize < 0 {
		minSize = 0
	}
	geometry.ManageGeometry(child, pw.geomMgr)
	pw.panes = append(pw.panes, pane{
		win:     child,
		minSize: minSize,
		size:    pw.reqSize(child),
	})
	pw.contentChanged()
}

// SetWeight sets the ttk -weight of the pane holding child (Weighted mode).
func (pw *PanedWindow) SetWeight(child *window.Window, weight int) {
	for i := range pw.panes {
		if pw.panes[i].win == child {
			pw.panes[i].weight = max(weight, 0)
		}
	}
	pw.scheduleArrange()
}

// SetStretch sets the -stretch mode of the pane holding child.
func (pw *PanedWindow) SetStretch(child *window.Window, s Stretch) {
	for i := range pw.panes {
		if pw.panes[i].win == child {
			pw.panes[i].stretch = s
		}
	}
	pw.scheduleArrange()
}

// Remove removes a pane by its window.
func (pw *PanedWindow) Remove(child *window.Window) {
	for i, p := range pw.panes {
		if p.win == child {
			child.GeomManager = nil
			pw.panes = append(pw.panes[:i], pw.panes[i+1:]...)
			pw.contentChanged()
			return
		}
	}
}

// Panes returns the managed windows.
func (pw *PanedWindow) Panes() []*window.Window {
	result := make([]*window.Window, len(pw.panes))
	for i, p := range pw.panes {
		result[i] = p.win
	}
	return result
}

// reqSize is w's requested size along the panedwindow's orientation.
func (pw *PanedWindow) reqSize(w *window.Window) int {
	if pw.Orient == Horizontal {
		return w.ReqWidth
	}
	return w.ReqHeight
}

// contentChanged recomputes the requested size and re-arranges after panes
// were added or removed.
func (pw *PanedWindow) contentChanged() {
	if pw.Weighted {
		pw.ttkSchedule(true, true)
		return
	}
	pw.propagateReqSize()
	pw.scheduleArrange()
}

// ttkSchedule ports ttkManager.c ScheduleUpdate/ManagerIdleProc for
// Weighted mode: a size recomputation and a relayout run at idle time,
// and after a size request the relayout waits for a further idle round,
// so requests from nested panes settle before any pane is placed.
func (pw *PanedWindow) ttkSchedule(resize, relayout bool) {
	pw.ttkResize = pw.ttkResize || resize
	pw.ttkRelayout = pw.ttkRelayout || relayout
	if pw.ttkPending {
		return
	}
	d := pw.Win.Display
	if d == nil || d.DoWhenIdle == nil {
		pw.ttkIdle()
		return
	}
	pw.ttkPending = true
	d.DoWhenIdle(pw.ttkIdle)
}

func (pw *PanedWindow) ttkIdle() {
	pw.ttkPending = false
	if pw.Destroyed {
		return
	}
	if pw.ttkResize {
		pw.ttkResize = false
		pw.propagateReqSize()
		pw.ttkSchedule(false, true)
	}
	if pw.ttkRelayout && !pw.ttkPending {
		pw.ttkRelayout = false
		pw.arrangePanes()
		pw.Display()
	}
}

// scheduleArrange arranges and redraws the panes at idle time, as
// tkPanedWindow.c schedules ArrangePanes with Tcl_DoWhenIdle; the
// requested size (propagateReqSize) is still recomputed at once, like
// ComputeGeometry.
func (pw *PanedWindow) scheduleArrange() {
	geometry.WhenIdle(pw.Win, &pw.arrangePending, func() {
		if pw.Destroyed {
			return
		}
		pw.arrangePanes()
		pw.Display()
	})
}

// arrangePanes distributes available space among panes and positions them.
func (pw *PanedWindow) arrangePanes() {
	if len(pw.panes) == 0 {
		return
	}

	w := pw.Win

	// ArrangePanes (tkPanedWindow.c): panes keep their base size and the
	// stretchable ones share what is left over, in proportion to size.
	bw := pw.BorderWidth
	n := len(pw.panes)
	pwSize := w.Height - 2*bw
	if pw.Orient == Horizontal {
		pwSize = w.Width - 2*bw
	}
	// Tk_IsMapped is true once Tk_MapWindow was called (when the geometry
	// manager maps the panedwindow), before X confirms it with MapNotify.
	mapped := pw.Win.IsMapped()
	for i := range pw.panes {
		p := &pw.panes[i]
		// Classic sizes come from Add and PanedWindowReqProc (RequestProc);
		// ttk's PaneRequest tracks requests until the pane is placed.
		track := p.size == 0
		if pw.Weighted {
			track = !p.win.IsMapped()
		}
		if !p.userSet && track {
			if pw.Orient == Horizontal {
				p.size = p.win.ReqWidth
			} else {
				p.size = p.win.ReqHeight
			}
		}
	}
	if pw.Weighted {
		pw.placeSashes(pwSize)
		pw.placePanes()
		return
	}
	reserve := pwSize - (n-1)*pw.SashWidth
	dynSize, dynMin := 0, 0
	for i, p := range pw.panes {
		reserve -= p.size
		if isStretchable(p.stretch, i, 0, n-1) && mapped {
			dynSize += p.size
			dynMin += p.minSize
		}
	}
	for i := range pw.panes {
		p := &pw.panes[i]
		size := p.size
		if isStretchable(p.stretch, i, 0, n-1) {
			var frac float64
			if dynSize > 0 {
				frac = float64(size) / float64(dynSize)
			} else if pwSize > 0 {
				frac = float64(size) / float64(pwSize)
			}
			dynSize -= size
			dynMin -= p.minSize
			amount := int(frac * float64(reserve))
			if size+amount >= p.minSize {
				reserve -= amount
				size += amount
			} else {
				reserve += size - p.minSize
				size = p.minSize
			}
			if i == n-1 && reserve > 0 {
				size += reserve
				reserve = 0
			}
		} else if dynSize-dynMin+reserve < 0 {
			if size+dynSize-dynMin+reserve <= p.minSize {
				reserve += size - p.minSize
				size = p.minSize
			} else {
				size += dynSize - dynMin + reserve
				reserve = dynMin - dynSize
			}
		}
		p.disp = size
	}

	pw.placePanes()
}

// placeSashes ports ttkPanedwindow.c PlaceSashes: every pane gets its
// request size, the difference to the available space is shared by
// -weight, and the last sash is pinned to the end (ShoveUp), so the last
// pane absorbs what the weights leave over.
func (pw *PanedWindow) placeSashes(available int) {
	n := len(pw.panes)
	sash := pw.SashWidth
	reqSize, totalWeight := 0, 0
	for _, p := range pw.panes {
		reqSize += p.size
		if p.size != 0 {
			totalWeight += p.weight
		}
	}
	difference := available - reqSize - sash*(n-1)
	delta, remainder := 0, 0
	if totalWeight != 0 {
		delta, remainder = difference/totalWeight, difference%totalWeight
		if remainder < 0 {
			delta--
			remainder += totalWeight
		}
	}
	sashPos := make([]int, n)
	pos := 0
	for i, p := range pw.panes {
		weight := 0
		if p.size != 0 {
			weight = p.weight
		}
		size := p.size + delta*weight
		weight = min(weight, remainder)
		remainder -= weight
		size = max(size+weight, 0)
		pos += size
		sashPos[i] = pos
		pos += sash
	}
	var shoveUp func(i, pos int) int
	shoveUp = func(i, pos int) int {
		if i == 0 {
			pos = max(pos, 0)
		} else if pos < sashPos[i-1]+sash {
			pos = shoveUp(i-1, pos-sash) + sash
		}
		sashPos[i] = pos
		return pos
	}
	shoveUp(n-1, available)
	start := 0
	for i := range pw.panes {
		pw.panes[i].disp = max(sashPos[i]-start, 0)
		start = sashPos[i] + sash
	}
}

// placePanes positions each pane at its computed size (PlacePanes).
func (pw *PanedWindow) placePanes() {
	w := pw.Win
	d := w.Display.Server
	bw := pw.BorderWidth
	pos := bw
	for i := range pw.panes {
		p := &pw.panes[i]
		paneW, paneH := p.disp, w.Height-2*bw
		paneX, paneY := pos, bw
		if pw.Orient == Vertical {
			paneW, paneH = w.Width-2*bw, p.disp
			paneX, paneY = bw, pos
		}
		if paneW < 1 {
			paneW = 1
		}
		if paneH < 1 {
			paneH = 1
		}

		d.MoveResizeWindow(p.win.PlatformID, paneX, paneY, uint(paneW), uint(paneH))
		p.win.X = paneX
		p.win.Y = paneY
		p.win.Width = paneW
		p.win.Height = paneH
		window.SyncBackground(p.win)
		d.MapWindow(p.win.PlatformID)
		window.MarkMapped(p.win)

		pos += p.disp + pw.SashWidth
	}
}

// hitSash returns the sash index under a pointer position, or -1.
func (pw *PanedWindow) hitSash(x, y int) int {
	if len(pw.panes) <= 1 {
		return -1
	}

	var pos int
	if pw.Orient == Horizontal {
		pos = x
	} else {
		pos = y
	}

	offset := pw.BorderWidth
	for i := 0; i < len(pw.panes)-1; i++ {
		offset += pw.panes[i].disp
		if pos >= offset && pos < offset+pw.SashWidth {
			return i
		}
		offset += pw.SashWidth
	}
	return -1
}

// moveSash moves sash by delta pixels, adjusting adjacent panes.
func (pw *PanedWindow) moveSash(sashIdx, newPaneSize int) {
	if sashIdx < 0 || sashIdx >= len(pw.panes)-1 {
		return
	}

	p1 := &pw.panes[sashIdx]
	p2 := &pw.panes[sashIdx+1]

	oldSize := p1.disp
	if newPaneSize < p1.minSize {
		newPaneSize = p1.minSize
	}

	maxSize := p1.disp + p2.disp - p2.minSize
	if newPaneSize > maxSize {
		newPaneSize = maxSize
	}

	delta := newPaneSize - oldSize
	// MoveSash/PlaceSash: every pane's size becomes its current size.
	for i := range pw.panes {
		pw.panes[i].size = pw.panes[i].disp
		pw.panes[i].userSet = true
	}
	p1.size = newPaneSize
	p2.size = p2.disp - delta

	pw.arrangePanes()
	pw.Display()
}

// Display schedules a redraw at idle time; see widget.Base.EventuallyRedraw.
func (pw *PanedWindow) Display() {
	pw.EventuallyRedraw()
}

// display draws the paned window (background and sashes).
func (pw *PanedWindow) display() {
	if pw.Destroyed {
		return
	}
	w := pw.Win
	if w.PlatformID == platform.WindowID(0) {
		return
	}

	d := w.Display.Server
	gc := w.GC

	// Background.
	if pw.Background != nil {
		d.SetForeground(gc, pw.Background.Pixel)
	}
	d.FillRectangle(w.Drawable(), gc, 0, 0, uint(w.Width), uint(w.Height))

	// Draw sashes.
	if pw.FlatSash {
		// ttk::panedwindow Sash layout (ttkPanedwindow.c): an invisible
		// sash with a GripSize-long grip in the border's dark colour.
		if pw.Border != nil {
			d.SetForeground(gc, pw.Border.DarkPixel)
		}
		offset := pw.BorderWidth
		for i := 0; i < len(pw.panes)-1; i++ {
			offset += pw.panes[i].disp
			if pw.SashWidth > 2 {
				if pw.Orient == Horizontal {
					d.FillRectangle(w.Drawable(), gc, offset+1, (w.Height-pw.GripSize)/2,
						uint(pw.SashWidth-2), uint(pw.GripSize))
				} else {
					d.FillRectangle(w.Drawable(), gc, (w.Width-pw.GripSize)/2, offset+1,
						uint(pw.GripSize), uint(pw.SashWidth-2))
				}
			}
			offset += pw.SashWidth
		}
	} else if pw.Border != nil && pw.ShowHandle {
		// Tk's defaults (-sashrelief flat, -showhandle 0) draw nothing
		// over the background; this raised sash is the -showhandle look.
		offset := pw.BorderWidth
		for i := 0; i < len(pw.panes)-1; i++ {
			offset += pw.panes[i].disp
			if pw.Orient == Horizontal {
				draw.Draw3DRectangle(d, w.Drawable(), gc, pw.Border,
					offset, 0, pw.SashWidth, w.Height, 1, option.ReliefRaised)
				// Handle.
				handleY := (w.Height - pw.HandleSize) / 2
				draw.Draw3DRectangle(d, w.Drawable(), gc, pw.Border,
					offset+1, handleY, pw.SashWidth-2, pw.HandleSize, 1, option.ReliefRaised)
			} else {
				draw.Draw3DRectangle(d, w.Drawable(), gc, pw.Border,
					0, offset, w.Width, pw.SashWidth, 1, option.ReliefRaised)
				// Handle.
				handleX := (w.Width - pw.HandleSize) / 2
				draw.Draw3DRectangle(d, w.Drawable(), gc, pw.Border,
					handleX, offset+1, pw.HandleSize, pw.SashWidth-2, 1, option.ReliefRaised)
			}
			offset += pw.SashWidth
		}
	}

}

// Configure applies options.
func (pw *PanedWindow) Configure(opts ...PanedWindowOption) {
	widget.Configure(pw, opts, pw.scheduleArrange)
}

// Destroy cleans up.
func (pw *PanedWindow) Destroy() {
	if pw.Destroyed {
		return
	}
	pw.Destroyed = true
	window.DestroyWindow(pw.Win)
}

func bindPanedWindow(pw *PanedWindow, app widget.AppContext) {
	w := pw.Win
	w.Flags |= window.FlagFocusable

	// Expose.
	app.Dispatcher().Bind(w.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		pw.Display()
	})

	// Configure (resize).
	app.Dispatcher().Bind(w.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.MapType && !pw.mapped {
			pw.mapped = true
			return
		}
		if ev.Type == event.ConfigureType {
			w.Width = ev.ConfigWidth
			w.Height = ev.ConfigHeight
			if pw.Weighted {
				// ttkManager.c ManagerEventHandler relayouts at once.
				pw.arrangePanes()
				pw.Display()
			} else {
				pw.scheduleArrange()
			}
		}
	})

	// Button press on sash.
	app.Dispatcher().Bind(w.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button != 1 {
			return
		}
		sash := pw.hitSash(ev.X, ev.Y)
		if sash >= 0 {
			pw.dragSash = sash
			if pw.Orient == Horizontal {
				pw.dragStartPos = ev.X
			} else {
				pw.dragStartPos = ev.Y
			}
			pw.dragStartSize = pw.panes[sash].disp
		}
	})

	// Button release.
	app.Dispatcher().Bind(w.PlatformID, event.ButtonReleaseMask, func(ev *event.Event) {
		pw.dragSash = -1
	})

	// Motion (drag sash).
	app.Dispatcher().Bind(w.PlatformID, event.MotionMask, func(ev *event.Event) {
		if pw.dragSash < 0 {
			return
		}
		var pos int
		if pw.Orient == Horizontal {
			pos = ev.X
		} else {
			pos = ev.Y
		}
		delta := pos - pw.dragStartPos
		newSize := pw.dragStartSize + delta
		pw.moveSash(pw.dragSash, newSize)
	})

	// Keyboard sash movement.
	app.Dispatcher().Bind(w.PlatformID, event.KeyPressMask, func(ev *event.Event) {
		if len(pw.panes) <= 1 {
			return
		}
		step := 10
		sashIdx := 0 // move first sash by default

		switch ev.KeySym {
		case platform.XK_Left:
			if pw.Orient == Horizontal {
				pw.moveSash(sashIdx, pw.panes[sashIdx].disp-step)
			}
		case platform.XK_Right:
			if pw.Orient == Horizontal {
				pw.moveSash(sashIdx, pw.panes[sashIdx].disp+step)
			}
		case platform.XK_Up:
			if pw.Orient == Vertical {
				pw.moveSash(sashIdx, pw.panes[sashIdx].disp-step)
			}
		case platform.XK_Down:
			if pw.Orient == Vertical {
				pw.moveSash(sashIdx, pw.panes[sashIdx].disp+step)
			}
		}
	})
}
