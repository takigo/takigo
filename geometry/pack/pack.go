// Package pack implements the pack geometry manager, which arranges
// children around the edges of a cavity. It ports tk/generic/tkPack.c.
package pack

import (
	"slices"

	"github.com/msorc/takigo/geometry"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/window"
)

// Side specifies which edge of the cavity to pack against.
type Side int

const (
	Top Side = iota
	Bottom
	Left
	Right
)

// Fill specifies how to fill the allocated frame.
type Fill int

const (
	FillNone Fill = iota
	FillX
	FillY
	FillBoth
)

// PackOption configures a Pack call.
type PackOption func(*packConfig)

type packConfig struct {
	side            Side
	fill            Fill
	expand          bool
	anchor          option.Anchor
	padX            int // total horizontal padding (left + right), as in tkPack.c
	padY            int // total vertical padding (top + bottom)
	padLeft, padTop int
	iPadX           int
	iPadY           int
	in              *window.Window // -in: a container other than the parent
}

// Side sets the packing side.
func SideOpt(s Side) PackOption { return func(c *packConfig) { c.side = s } }

// FillOpt sets the fill mode.
func FillOpt(f Fill) PackOption { return func(c *packConfig) { c.fill = f } }

// Expand sets whether the child absorbs extra space.
func Expand(b bool) PackOption { return func(c *packConfig) { c.expand = b } }

// Anchor sets the anchor position within the frame.
func Anchor(a option.Anchor) PackOption { return func(c *packConfig) { c.anchor = a } }

// PadX sets the exterior horizontal padding.
// Accepts int (pixels), float64 (rounded pixels), or string with unit suffix ("3p", "2m", "1c", "0.5i").
func PadX(p any) PackOption {
	return func(c *packConfig) { c.padLeft = screenunit.PxOr(p, c.padLeft); c.padX = 2 * c.padLeft }
}

// PadXPair sets asymmetric exterior horizontal padding (-padx {left right}).
func PadXPair(left, right any) PackOption {
	return func(c *packConfig) {
		right0 := c.padX - c.padLeft
		c.padLeft = screenunit.PxOr(left, c.padLeft)
		c.padX = c.padLeft + screenunit.PxOr(right, right0)
	}
}

// PadY sets the exterior vertical padding.
// Accepts int (pixels), float64 (rounded pixels), or string with unit suffix ("3p", "2m", "1c", "0.5i").
func PadY(p any) PackOption {
	return func(c *packConfig) { c.padTop = screenunit.PxOr(p, c.padTop); c.padY = 2 * c.padTop }
}

// PadYPair sets asymmetric exterior vertical padding (-pady {top bottom}).
func PadYPair(top, bottom any) PackOption {
	return func(c *packConfig) {
		bottom0 := c.padY - c.padTop
		c.padTop = screenunit.PxOr(top, c.padTop)
		c.padY = c.padTop + screenunit.PxOr(bottom, bottom0)
	}
}

// IPadX sets the interior horizontal padding.
// Accepts int (pixels), float64 (rounded pixels), or string with unit suffix ("3p", "2m", "1c", "0.5i").
func IPadX(p any) PackOption { return func(c *packConfig) { c.iPadX = screenunit.PxOr(p, c.iPadX) } }

// IPadY sets the interior vertical padding.
// Accepts int (pixels), float64 (rounded pixels), or string with unit suffix ("3p", "2m", "1c", "0.5i").
func IPadY(p any) PackOption { return func(c *packConfig) { c.iPadY = screenunit.PxOr(p, c.iPadY) } }

// packEntry holds packing configuration for a single child.
type packEntry struct {
	window *window.Window
	config packConfig
}

// packer manages the pack state for a container window.
type packer struct {
	container *window.Window
	pending   bool // an arrange is scheduled for idle time
	entries   []*packEntry
}

// singleton manager instance.
var mgr = &packManager{}

// packers tracks per-container state.
var packers = map[*window.Window]*packer{}

// hooked records the containers whose destroy and configure hooks are
// registered.
var hooked = map[*window.Window]bool{}

type packManager struct{}

func (m *packManager) Name() string { return "pack" }

func (m *packManager) RequestProc(content *window.Window) {
	if p, ok := packers[containerFor(content)]; ok {
		p.scheduleArrange()
	}
}

// LostContentProc drops content from its container, e.g. when content is
// destroyed or taken over by another geometry manager.
func (m *packManager) LostContentProc(content *window.Window) {
	container := containerFor(content)
	delete(containerOf, content)
	if p, ok := packers[container]; ok {
		p.remove(content)
		p.scheduleArrange()
	}
}

// containerFor returns the window content is managed in: its -in
// container if one was given, else its parent.
func containerFor(content *window.Window) *window.Window {
	if c := containerOf[content]; c != nil {
		return c
	}
	return content.Parent
}

// forgetContainer drops a destroyed container's state; content managed
// in it from outside its subtree (via -in) becomes unmanaged, as in
// Tk's DestroyNotify handling in the geometry managers.
func forgetContainer(container *window.Window) {
	p, ok := packers[container]
	if !ok {
		return
	}
	delete(packers, container)
	for _, e := range p.entries {
		if w := e.window; containerOf[w] == container {
			delete(containerOf, w)
			w.GeomManager = nil
			if w.IsMapped() && !w.IsDestroyed() && w.PlatformID != 0 {
				w.Display.Server.UnmapWindow(w.PlatformID)
				window.MarkUnmapped(w)
			}
		}
	}
}

// Pack adds children to their parent's pack layout.
// Accepts a geometry.Elementer (e.g. geometry.Group) containing one or more
// widgets. All widgets receive the same options, matching Tk's
// "pack configure .w1 .w2 .w3 -side left" behavior.
// See tk/generic/tkPack.c ConfigureContent.
// In is pack's -in: manage the content inside container, which must be the
// content's parent or a descendant of it (Tk_MaintainGeometry keeps it there).
func In(container window.Windower) PackOption {
	return func(c *packConfig) { c.in = container.Window() }
}

// containerOf records each content window's container (Tk's containerPtr).
var containerOf = map[*window.Window]*window.Window{}

func Pack(children geometry.Elementer, opts ...PackOption) {
	cfg := packConfig{
		side:   Top,
		fill:   FillNone,
		anchor: option.AnchorCenter,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	elements := children.GeometryElements()

	// Arrange each affected container once, after all elements are added,
	// rather than once per element.
	var touched []*packer
	for _, elem := range elements {
		w := elem.Window()
		parent := w.Parent
		if cfg.in != nil {
			parent = cfg.in
		}
		if parent == nil {
			continue
		}
		if old := containerOf[w]; old != nil && old != parent {
			if op, ok := packers[old]; ok {
				op.remove(w)
				op.scheduleArrange()
			}
		}
		containerOf[w] = parent

		geometry.ManageGeometry(w, mgr)

		p, ok := packers[parent]
		if !ok {
			p = &packer{container: parent}
			packers[parent] = p
		}
		if !hooked[parent] {
			// The packer is dropped when its last content goes but the
			// hooks stay with the window, so register them once.
			hooked[parent] = true
			parent.OnDestroy(func() {
				forgetContainer(parent)
				delete(hooked, parent)
			})
			// Re-layout when resized by external forces (e.g. PanedWindow).
			parent.OnConfigure(func() {
				if pp, ok2 := packers[parent]; ok2 {
					pp.scheduleArrange()
				}
			})
		}

		// Update or add entry.
		found := false
		for _, e := range p.entries {
			if e.window == w {
				e.config = cfg
				found = true
				break
			}
		}
		if !found {
			p.entries = append(p.entries, &packEntry{window: w, config: cfg})
		}
		if !slices.Contains(touched, p) {
			touched = append(touched, p)
		}
	}
	for _, p := range touched {
		p.scheduleArrange()
	}
}

// Forget removes a child from pack management.
func Forget(child window.Windower) {
	w := child.Window()
	parent := containerOf[w]
	if parent == nil {
		parent = w.Parent
	}
	if parent == nil {
		return
	}
	delete(containerOf, w)
	w.GeomManager = nil
	// pack forget unmaps the content and re-packs the rest.
	if w.IsMapped() && w.PlatformID != platform.WindowID(0) {
		w.Display.Server.UnmapWindow(w.PlatformID)
		window.MarkUnmapped(w)
	}
	if p, ok := packers[parent]; ok {
		p.remove(w)
		p.scheduleArrange()
	}
}

// remove removes a child from the packer's entry list.
func (p *packer) remove(child *window.Window) {
	for i, e := range p.entries {
		if e.window == child {
			p.entries = append(p.entries[:i], p.entries[i+1:]...)
			break
		}
	}
	if len(p.entries) == 0 {
		delete(packers, p.container)
	}
}

// arrange performs the two-pass layout algorithm.
func (p *packer) arrange() {
	container := p.container
	// As in tkPack.c ArrangePacking, a container left without content
	// keeps its size, so another geometry manager can take it over.
	if container.PlatformID == platform.WindowID(0) || container.IsDestroyed() || len(p.entries) == 0 {
		return
	}

	// Pass 1: Compute required container size.
	maxWidth, maxHeight := p.computeSize()

	// Update container's requested size (for propagation).
	maxWidth += container.InternalBorderLeft + container.InternalBorderRight
	maxHeight += container.InternalBorderTop + container.InternalBorderBottom

	if container.ReqWidth != maxWidth || container.ReqHeight != maxHeight {
		if container.IsTopLevel() {
			// For toplevel windows, resize the X window to fit content.
			// This matches Tk's Tk_GeometryRequest which calls XResizeWindow
			// for toplevels so the window manager adjusts the window size.
			container.ReqWidth = maxWidth
			container.ReqHeight = maxHeight
			window.ResizeToplevel(container, maxWidth, maxHeight)
		} else {
			geometry.GeometryRequest(container, maxWidth, maxHeight)
		}
	}

	// Pass 2: Allocate positions using actual container size.
	cavityX := container.InternalBorderLeft
	cavityY := container.InternalBorderTop
	cavityW := container.Width - container.InternalBorderLeft - container.InternalBorderRight
	cavityH := container.Height - container.InternalBorderTop - container.InternalBorderBottom

	if cavityW < 0 {
		cavityW = 0
	}
	if cavityH < 0 {
		cavityH = 0
	}

	for _, e := range p.entries {
		child := e.window
		cfg := &e.config
		bw2 := 2 * child.BorderWidth

		childReqW := child.ReqWidth + bw2 + cfg.iPadX*2
		childReqH := child.ReqHeight + bw2 + cfg.iPadY*2

		// Compute frame (space allocated to this child).
		var frameX, frameY, frameW, frameH int

		switch cfg.side {
		case Top:
			frameX = cavityX
			frameY = cavityY
			frameW = cavityW
			frameH = childReqH + cfg.padY
			if cfg.expand {
				frameH += yExpansion(p.entries, e, cavityH)
			}
			if frameH > cavityH {
				frameH = cavityH
			}
			cavityY += frameH
			cavityH -= frameH

		case Bottom:
			frameX = cavityX
			frameW = cavityW
			frameH = childReqH + cfg.padY
			if cfg.expand {
				frameH += yExpansion(p.entries, e, cavityH)
			}
			if frameH > cavityH {
				frameH = cavityH
			}
			frameY = cavityY + cavityH - frameH
			cavityH -= frameH

		case Left:
			frameX = cavityX
			frameY = cavityY
			frameH = cavityH
			frameW = childReqW + cfg.padX
			if cfg.expand {
				frameW += xExpansion(p.entries, e, cavityW)
			}
			if frameW > cavityW {
				frameW = cavityW
			}
			cavityX += frameW
			cavityW -= frameW

		case Right:
			frameY = cavityY
			frameH = cavityH
			frameW = childReqW + cfg.padX
			if cfg.expand {
				frameW += xExpansion(p.entries, e, cavityW)
			}
			if frameW > cavityW {
				frameW = cavityW
			}
			frameX = cavityX + cavityW - frameW
			cavityW -= frameW
		}

		// Position child within frame.
		childW := childReqW
		childH := childReqH

		if cfg.fill == FillX || cfg.fill == FillBoth {
			childW = frameW - cfg.padX
		}
		if cfg.fill == FillY || cfg.fill == FillBoth {
			childH = frameH - cfg.padY
		}

		// Apply anchor.
		childX, childY := anchorPosition(cfg.anchor, frameX+cfg.padLeft, frameY+cfg.padTop,
			frameW-cfg.padX, frameH-cfg.padY, childW, childH)

		// ArrangePacking: content with no room is unmapped and left where
		// it was.
		if childW-bw2 <= 0 || childH-bw2 <= 0 {
			if child.Flags&window.FlagMapped != 0 && child.PlatformID != platform.WindowID(0) {
				container.Display.Server.UnmapWindow(child.PlatformID)
				window.MarkUnmapped(child)
			}
			continue
		}

		// Move and resize the child window; content packed -in another
		// container is offset by that container's position.
		dx, dy := window.ContentOffset(container, child)
		moved := child.X != childX+dx || child.Y != childY+dy
		child.X = childX + dx
		child.Y = childY + dy
		child.Width = childW - bw2
		child.Height = childH - bw2

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

// computeSize calculates the minimum container size needed for all children.
//
// Top/Bottom children accumulate height; Left/Right children accumulate width.
// For the cross-axis, Left/Right children need cavity space AFTER Top/Bottom
// children consume theirs, so their max height is added to the Top/Bottom sum
// (not max'd). This matches how pass 2 allocates cavity space.
func (p *packer) computeSize() (int, int) {
	// ArrangePacking (tk/generic/tkPack.c): walk the content in packing order,
	// tracking the space already used on the other axis, so that e.g. a
	// right-packed scrollbar adds to the width of a later top-packed text.
	var width, height, maxWidth, maxHeight int
	for _, e := range p.entries {
		cfg := &e.config
		child := e.window
		bw2 := 2 * child.BorderWidth
		childW := child.ReqWidth + bw2 + cfg.iPadX*2 + cfg.padX
		childH := child.ReqHeight + bw2 + cfg.iPadY*2 + cfg.padY

		switch cfg.side {
		case Top, Bottom:
			maxWidth = max(maxWidth, childW+width)
			height += childH
		case Left, Right:
			maxHeight = max(maxHeight, childH+height)
			width += childW
		}
	}
	return max(maxWidth, width), max(maxHeight, height)
}

// expansion computes extra space for an expanding child along one axis.
// sideMatch returns true for entries packed on the same axis (Left/Right or Top/Bottom).
// childNeed returns the space an entry requires along that axis.
func expansion(entries []*packEntry, target *packEntry, cavity int,
	sideMatch func(Side) bool, childNeed func(*packEntry) int) int {

	numExpand := 0
	found := false

	for _, e := range entries {
		if e == target {
			found = true
		}
		if !found {
			continue
		}
		if !sideMatch(e.config.side) {
			continue
		}
		cavity -= childNeed(e)
		if e.config.expand {
			numExpand++
		}
	}

	if numExpand > 0 && cavity > 0 {
		return cavity / numExpand
	}
	return 0
}

// xExpansion computes extra horizontal space for an expanding child.
func xExpansion(entries []*packEntry, target *packEntry, cavityW int) int {
	return expansion(entries, target, cavityW,
		func(s Side) bool { return s == Left || s == Right },
		func(e *packEntry) int {
			return e.window.ReqWidth + 2*e.window.BorderWidth + e.config.iPadX*2 + e.config.padX
		})
}

// yExpansion computes extra vertical space for an expanding child.
func yExpansion(entries []*packEntry, target *packEntry, cavityH int) int {
	return expansion(entries, target, cavityH,
		func(s Side) bool { return s == Top || s == Bottom },
		func(e *packEntry) int {
			return e.window.ReqHeight + 2*e.window.BorderWidth + e.config.iPadY*2 + e.config.padY
		})
}

// anchorPosition computes the x,y position for a child within a frame
// based on the anchor setting.
func anchorPosition(a option.Anchor, frameX, frameY, frameW, frameH, childW, childH int) (int, int) {
	var x, y int

	switch a {
	case option.AnchorNW:
		x = frameX
		y = frameY
	case option.AnchorN:
		x = frameX + (frameW-childW)/2
		y = frameY
	case option.AnchorNE:
		x = frameX + frameW - childW
		y = frameY
	case option.AnchorW:
		x = frameX
		y = frameY + (frameH-childH)/2
	case option.AnchorCenter:
		x = frameX + (frameW-childW)/2
		y = frameY + (frameH-childH)/2
	case option.AnchorE:
		x = frameX + frameW - childW
		y = frameY + (frameH-childH)/2
	case option.AnchorSW:
		x = frameX
		y = frameY + frameH - childH
	case option.AnchorS:
		x = frameX + (frameW-childW)/2
		y = frameY + frameH - childH
	case option.AnchorSE:
		x = frameX + frameW - childW
		y = frameY + frameH - childH
	}

	return x, y
}

// ArrangeAll triggers layout for all pack-managed containers.
// Call this after window resize events.
func ArrangeAll() {
	for _, p := range packers {
		p.scheduleArrange()
	}
}

// Arrange (and so map) the content once its container is mapped, as the
// managers' structure procs do on MapNotify; keep -in content with its
// container when it moves or is unmapped (Tk_MaintainGeometry).
func init() {
	window.AddMappedHook(ArrangeContainer)
	window.AddMovedHook(func(w *window.Window) {
		if p, ok := packers[w]; ok && p.hasForeign() {
			p.scheduleArrange()
		}
	})
	window.AddUnmappedHook(func(w *window.Window) {
		if p, ok := packers[w]; ok {
			p.unmapForeign()
		}
	})
}

// hasForeign reports whether some content is packed -in this container
// without being its child.
func (p *packer) hasForeign() bool {
	for _, e := range p.entries {
		if e.window.Parent != p.container {
			return true
		}
	}
	return false
}

// unmapForeign unmaps -in content whose container was unmapped; X does this
// for real children.
func (p *packer) unmapForeign() {
	for _, e := range p.entries {
		w := e.window
		if w.Parent != p.container && w.IsMapped() && w.PlatformID != platform.WindowID(0) {
			w.Display.Server.UnmapWindow(w.PlatformID)
			window.MarkUnmapped(w)
		}
	}
}

// ArrangeContainer triggers layout for a specific container.
func ArrangeContainer(container *window.Window) {
	if p, ok := packers[container]; ok {
		p.arrange()
	}
}

// scheduleArrange re-arranges the container at idle time.
func (p *packer) scheduleArrange() {
	geometry.WhenIdle(p.container, &p.pending, p.arrange)
}
