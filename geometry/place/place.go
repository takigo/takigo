// Package place implements the place geometry manager, which positions
// children using absolute and/or relative coordinates.
// It ports tk/generic/tkPlace.c.
package place

import (
	"math"

	"github.com/msorc/takigo/geometry"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/window"
)

// PlaceOption configures a Place call.
type PlaceOption func(*placeConfig)

type placeConfig struct {
	x         int
	y         int
	relX      float64
	relY      float64
	width     int     // absolute width (-1 = use requested)
	height    int     // absolute height (-1 = use requested)
	relWidth  float64 // relative width (NaN = not set)
	relHeight float64 // relative height (NaN = not set)
	anchor    option.Anchor
	in        *window.Window // -in: a container other than the parent
}

// In is place's -in: place the content relative to container, which must be
// the content's parent or a descendant of it.
func In(container window.Windower) PlaceOption {
	return func(c *placeConfig) { c.in = container.Window() }
}

// containerOf records each content window's container.
var containerOf = new(geometry.Table[*window.Window])

// X sets the absolute x position.
func X(v int) PlaceOption { return func(c *placeConfig) { c.x = v } }

// Y sets the absolute y position.
func Y(v int) PlaceOption { return func(c *placeConfig) { c.y = v } }

// RelX sets the relative x position (0.0-1.0).
func RelX(v float64) PlaceOption { return func(c *placeConfig) { c.relX = v } }

// RelY sets the relative y position (0.0-1.0).
func RelY(v float64) PlaceOption { return func(c *placeConfig) { c.relY = v } }

// Width sets the absolute width.
func Width(v int) PlaceOption { return func(c *placeConfig) { c.width = v } }

// Height sets the absolute height.
func Height(v int) PlaceOption { return func(c *placeConfig) { c.height = v } }

// RelWidth sets the relative width (0.0-1.0, fraction of container).
func RelWidth(v float64) PlaceOption { return func(c *placeConfig) { c.relWidth = v } }

// RelHeight sets the relative height (0.0-1.0, fraction of container).
func RelHeight(v float64) PlaceOption { return func(c *placeConfig) { c.relHeight = v } }

// Anchor sets which point of the child is placed at the position.
func Anchor(a option.Anchor) PlaceOption { return func(c *placeConfig) { c.anchor = a } }

// placeEntry holds placement configuration for a single child.
type placeEntry struct {
	window *window.Window
	config placeConfig
}

// placer manages place state for a container.
type placer struct {
	container *window.Window
	pending   bool // an arrange is scheduled for idle time
	entries   []*placeEntry
}

// singleton manager instance.
var mgr = &placeManager{}

// placers tracks per-container state.
var placers = new(geometry.Table[*placer])

// hooked records the containers whose destroy and configure hooks are
// registered.
var hooked = new(geometry.Table[bool])

type placeManager struct{}

func (m *placeManager) Name() string { return "place" }

func (m *placeManager) RequestProc(content *window.Window) {
	if p, ok := placers.Get(containerFor(content)); ok {
		p.scheduleArrange()
	}
}

// LostContentProc drops content from its container, e.g. when content is
// destroyed or taken over by another geometry manager.
func (m *placeManager) LostContentProc(content *window.Window) {
	container := containerFor(content)
	containerOf.Delete(content)
	if p, ok := placers.Get(container); ok {
		p.remove(content)
		p.scheduleArrange()
	}
}

// containerFor returns the window content is managed in: its -in
// container if one was given, else its parent.
func containerFor(content *window.Window) *window.Window {
	if c := containerOf.Of(content); c != nil {
		return c
	}
	return content.Parent
}

// forgetContainer drops a destroyed container's state; content managed
// in it from outside its subtree (via -in) becomes unmanaged, as in
// Tk's DestroyNotify handling in the geometry managers.
func forgetContainer(container *window.Window) {
	p, ok := placers.Get(container)
	if !ok {
		return
	}
	placers.Delete(container)
	for _, e := range p.entries {
		if w := e.window; containerOf.Of(w) == container {
			containerOf.Delete(w)
			w.GeomManager = nil
			if w.IsMapped() && !w.IsDestroyed() && w.PlatformID != 0 {
				w.Display.Server.UnmapWindow(w.PlatformID)
				window.MarkUnmapped(w)
			}
		}
	}
}

// Place positions a child within its parent using absolute/relative coords.
func Place(child window.Windower, opts ...PlaceOption) {
	w := child.Window()
	parent := w.Parent
	if parent == nil {
		return
	}

	cfg := placeConfig{
		width:     -1,
		height:    -1,
		relWidth:  math.NaN(),
		relHeight: math.NaN(),
		anchor:    option.AnchorNW,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.in != nil {
		parent = cfg.in
	}
	if old := containerOf.Of(w); old != nil && old != parent {
		if op, ok := placers.Get(old); ok {
			op.remove(w)
		}
	}
	containerOf.Set(w, parent)

	geometry.ManageGeometry(w, mgr)

	p, ok := placers.Get(parent)
	if !ok {
		p = &placer{container: parent}
		placers.Set(parent, p)
	}
	if !hooked.Of(parent) {
		// The placer is dropped when its last content goes but the hooks
		// stay with the window, so register them once.
		hooked.Set(parent, true)
		parent.OnDestroy(func() {
			forgetContainer(parent)
			hooked.Delete(parent)
		})
		// Relative placement follows the container's size, whatever
		// resizes it (PlaceStructureProc).
		parent.OnConfigure(func() {
			if pp, ok := placers.Get(parent); ok {
				pp.scheduleArrange()
			}
		})
	}

	// Update or add entry.
	for _, e := range p.entries {
		if e.window == w {
			e.config = cfg
			p.scheduleArrange()
			return
		}
	}

	p.entries = append(p.entries, &placeEntry{window: w, config: cfg})
	p.scheduleArrange()
}

// Forget removes a child from place management.
func Forget(child window.Windower) {
	w := child.Window()
	parent := containerOf.Of(w)
	if parent == nil {
		parent = w.Parent
	}
	if parent == nil {
		return
	}
	containerOf.Delete(w)
	if p, ok := placers.Get(parent); ok {
		p.remove(w)
	}
	w.GeomManager = nil
	// place forget unmaps the content.
	if w.IsMapped() && w.PlatformID != platform.WindowID(0) {
		w.Display.Server.UnmapWindow(w.PlatformID)
		window.MarkUnmapped(w)
	}
}

func (p *placer) remove(child *window.Window) {
	for i, e := range p.entries {
		if e.window == child {
			p.entries = append(p.entries[:i], p.entries[i+1:]...)
			break
		}
	}
	if len(p.entries) == 0 {
		placers.Delete(p.container)
	}
}

// arrange positions all placed children.
func (p *placer) arrange() {
	container := p.container
	if container.PlatformID == platform.WindowID(0) || container.IsDestroyed() {
		return
	}

	// Container usable area.
	containerW := container.Width - container.InternalBorderLeft - container.InternalBorderRight
	containerH := container.Height - container.InternalBorderTop - container.InternalBorderBottom

	for _, e := range p.entries {
		child := e.window
		cfg := &e.config
		bw2 := 2 * child.BorderWidth

		// Compute position.
		// RecomputePlacement rounds the anchor point, and relative sizes
		// via the rounded far edge.
		round := func(v float64) int {
			if v > 0 {
				return int(v + 0.5)
			}
			return int(v - 0.5)
		}
		x1 := float64(cfg.x+container.InternalBorderLeft) + cfg.relX*float64(containerW)
		y1 := float64(cfg.y+container.InternalBorderTop) + cfg.relY*float64(containerH)
		x, y := round(x1), round(y1)

		// Compute size.
		var childW, childH int

		if cfg.width >= 0 || !math.IsNaN(cfg.relWidth) {
			childW = 0
			if cfg.width >= 0 {
				childW += cfg.width
			}
			if !math.IsNaN(cfg.relWidth) {
				childW += round(x1+cfg.relWidth*float64(containerW)) - x
			}
		} else {
			childW = child.ReqWidth + bw2
		}

		if cfg.height >= 0 || !math.IsNaN(cfg.relHeight) {
			childH = 0
			if cfg.height >= 0 {
				childH += cfg.height
			}
			if !math.IsNaN(cfg.relHeight) {
				childH += round(y1+cfg.relHeight*float64(containerH)) - y
			}
		} else {
			childH = child.ReqHeight + bw2
		}

		// Apply anchor — shift position so anchor point lands on (x, y).
		switch cfg.anchor {
		case option.AnchorN:
			x -= childW / 2
		case option.AnchorNE:
			x -= childW
		case option.AnchorE:
			x -= childW
			y -= childH / 2
		case option.AnchorSE:
			x -= childW
			y -= childH
		case option.AnchorS:
			x -= childW / 2
			y -= childH
		case option.AnchorSW:
			y -= childH
		case option.AnchorW:
			y -= childH / 2
		case option.AnchorCenter:
			x -= childW / 2
			y -= childH / 2
		case option.AnchorNW:
			// No adjustment.
		}

		// Clamp minimum size.
		if childW < 1 {
			childW = 1
		}
		if childH < 1 {
			childH = 1
		}

		// Content placed -in another container is offset by its position.
		dx, dy := window.ContentOffset(container, child)
		moved := child.X != x+dx || child.Y != y+dy
		child.X = x + dx
		child.Y = y + dy
		child.Width = childW - bw2
		child.Height = childH - bw2
		if child.Width < 1 {
			child.Width = 1
		}
		if child.Height < 1 {
			child.Height = 1
		}

		if child.PlatformID != platform.WindowID(0) {
			container.Display.Server.MoveResizeWindow(child.PlatformID,
				child.X, child.Y, uint(child.Width), uint(child.Height))
			// Tk maps content only once its container is mapped; the
			// container's MarkMapped re-arranges and maps it then.
			if child.Flags&window.FlagMapped == 0 && window.ContainerViewable(container, child) {
				window.SyncBackground(child)
				container.Display.Server.MapWindow(child.PlatformID)
				window.MarkMapped(child)
			}
		}
		if moved {
			window.NotifyMoved(child)
		}
	}
}

// ArrangeAll triggers layout for the place-managed containers in root's
// subtree, root included.
func ArrangeAll(root *window.Window) {
	if root == nil {
		return
	}
	if p, ok := placers.Get(root); ok {
		p.scheduleArrange()
	}
	for _, c := range root.Children {
		ArrangeAll(c)
	}
}

// Arrange (and so map) the content once its container is mapped, as the
// managers' structure procs do on MapNotify; keep -in content with its
// container when it moves or is unmapped (Tk_MaintainGeometry).
func init() {
	window.AddMappedHook(ArrangeContainer)
	window.AddMovedHook(func(w *window.Window) {
		if p, ok := placers.Get(w); ok && p.hasForeign() {
			p.scheduleArrange()
		}
	})
	window.AddUnmappedHook(func(w *window.Window) {
		p, ok := placers.Get(w)
		if !ok {
			return
		}
		for _, e := range p.entries {
			c := e.window
			if c.Parent != w && c.IsMapped() && c.PlatformID != platform.WindowID(0) {
				c.Display.Server.UnmapWindow(c.PlatformID)
				window.MarkUnmapped(c)
			}
		}
	})
}

func (p *placer) hasForeign() bool {
	for _, e := range p.entries {
		if e.window.Parent != p.container {
			return true
		}
	}
	return false
}

// ArrangeContainer triggers layout for a specific container.
func ArrangeContainer(container *window.Window) {
	if p, ok := placers.Get(container); ok {
		p.arrange()
	}
}

// scheduleArrange re-arranges the container at idle time.
func (p *placer) scheduleArrange() {
	geometry.WhenIdle(p.container, &p.pending, p.arrange)
}
