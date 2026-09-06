// Package panedwindow implements a paned window widget that manages
// resizable child panes separated by draggable sashes.
// It ports tk/generic/tkPanedWindow.c.
package panedwindow

import (
	"log"

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
	m.pw.propagateReqSize()
}

func (m *pwGeomMgr) LostContentProc(content *window.Window) {}

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
	size    int // allocated size along orient axis
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
	return func(pw *PanedWindow) {
		col, err := pw.App.ColorCache().Get(name)
		if err != nil {
			log.Printf("panedwindow: failed to get color %q: %v", name, err)
			return
		}
		pw.Background = col
		pw.UpdateBorder()
	}
}

// New creates a new PanedWindow widget.
func New(parent widget.Caregiver, name string, opts ...PanedWindowOption) *PanedWindow {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 200, 200)
	window.MakeWindowExist(w)

	pw := &PanedWindow{
		Orient:     Horizontal,
		SashWidth:  8,
		HandleSize: 8,
		dragSash:   -1,
	}
	pw.geomMgr = &pwGeomMgr{pw: pw}
	widget.InitBase(&pw.Base, w, app)
	pw.BorderWidth = 0
	pw.Relief = option.ReliefFlat

	for _, opt := range opts {
		opt(pw)
	}

	if pw.Background != nil {
		w.BackgroundPixel = pw.Background.Pixel
	}

	bindPanedWindow(pw, app)
	return pw
}

// Add adds a child window as a pane.
func (pw *PanedWindow) Add(child *window.Window, minSize int) {
	if minSize < 1 {
		minSize = 1
	}
	geometry.ManageGeometry(child, pw.geomMgr)
	pw.panes = append(pw.panes, pane{
		win:     child,
		minSize: minSize,
		size:    0,
	})
	pw.propagateReqSize()
	pw.arrangePanes()
}

// Remove removes a pane by its window.
func (pw *PanedWindow) Remove(child *window.Window) {
	for i, p := range pw.panes {
		if p.win == child {
			child.GeomManager = nil
			pw.panes = append(pw.panes[:i], pw.panes[i+1:]...)
			pw.propagateReqSize()
			pw.arrangePanes()
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

// arrangePanes distributes available space among panes and positions them.
func (pw *PanedWindow) arrangePanes() {
	if len(pw.panes) == 0 {
		return
	}

	w := pw.Win
	d := w.Display.Server

	totalSashSpace := (len(pw.panes) - 1) * pw.SashWidth
	var totalAvail int
	if pw.Orient == Horizontal {
		totalAvail = w.Width - totalSashSpace
	} else {
		totalAvail = w.Height - totalSashSpace
	}
	if totalAvail < 0 {
		totalAvail = 0
	}

	// If any pane has size=0 (initial), distribute equally.
	needInit := false
	for _, p := range pw.panes {
		if p.size == 0 {
			needInit = true
			break
		}
	}
	if needInit {
		each := totalAvail / len(pw.panes)
		if each < 1 {
			each = 1
		}
		for i := range pw.panes {
			pw.panes[i].size = each
			if pw.panes[i].size < pw.panes[i].minSize {
				pw.panes[i].size = pw.panes[i].minSize
			}
		}
	}

	// Ensure total matches available. Last pane gets remainder.
	usedByFixed := 0
	for i := 0; i < len(pw.panes)-1; i++ {
		if pw.panes[i].size < pw.panes[i].minSize {
			pw.panes[i].size = pw.panes[i].minSize
		}
		usedByFixed += pw.panes[i].size
	}
	lastSize := totalAvail - usedByFixed
	lastIdx := len(pw.panes) - 1
	if lastSize < pw.panes[lastIdx].minSize {
		lastSize = pw.panes[lastIdx].minSize
	}
	pw.panes[lastIdx].size = lastSize

	// Position each pane.
	pos := 0
	for i := range pw.panes {
		p := &pw.panes[i]
		paneW, paneH := p.size, w.Height
		paneX, paneY := pos, 0
		if pw.Orient == Vertical {
			paneW, paneH = w.Width, p.size
			paneX, paneY = 0, pos
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
		d.MapWindow(p.win.PlatformID)

		if pw.Orient == Horizontal {
			pos += p.size + pw.SashWidth
		} else {
			pos += p.size + pw.SashWidth
		}
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

	offset := 0
	for i := 0; i < len(pw.panes)-1; i++ {
		offset += pw.panes[i].size
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

	oldSize := p1.size
	if newPaneSize < p1.minSize {
		newPaneSize = p1.minSize
	}

	maxSize := p1.size + p2.size - p2.minSize
	if newPaneSize > maxSize {
		newPaneSize = maxSize
	}

	delta := newPaneSize - oldSize
	p1.size = newPaneSize
	p2.size -= delta

	pw.arrangePanes()
	pw.Display()
}

// Display draws the paned window (background and sashes).
func (pw *PanedWindow) Display() {
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
		// Flat TTK-style sash: a single thin line with a small grip dot.
		d.SetForeground(gc, uint64(0x9e9e9e))
		offset := 0
		for i := 0; i < len(pw.panes)-1; i++ {
			offset += pw.panes[i].size
			if pw.Orient == Horizontal {
				midX := offset + pw.SashWidth/2
				d.DrawLine(w.Drawable(), gc, midX, 0, midX, w.Height)
			} else {
				midY := offset + pw.SashWidth/2
				d.DrawLine(w.Drawable(), gc, 0, midY, w.Width, midY)
			}
			offset += pw.SashWidth
		}
	} else if pw.Border != nil {
		offset := 0
		for i := 0; i < len(pw.panes)-1; i++ {
			offset += pw.panes[i].size
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

	d.Flush()
}

// Configure applies options.
func (pw *PanedWindow) Configure(opts ...option.Option) {
	option.Apply(pw, opts)
	pw.UpdateBorder()
	if pw.Background != nil {
		pw.Win.BackgroundPixel = pw.Background.Pixel
	}
	pw.arrangePanes()
	pw.Display()
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
		if ev.Type == event.ConfigureType {
			w.Width = ev.ConfigWidth
			w.Height = ev.ConfigHeight
			pw.arrangePanes()
			pw.Display()
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
			pw.dragStartSize = pw.panes[sash].size
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
				pw.moveSash(sashIdx, pw.panes[sashIdx].size-step)
			}
		case platform.XK_Right:
			if pw.Orient == Horizontal {
				pw.moveSash(sashIdx, pw.panes[sashIdx].size+step)
			}
		case platform.XK_Up:
			if pw.Orient == Vertical {
				pw.moveSash(sashIdx, pw.panes[sashIdx].size-step)
			}
		case platform.XK_Down:
			if pw.Orient == Vertical {
				pw.moveSash(sashIdx, pw.panes[sashIdx].size+step)
			}
		}
	})
}
