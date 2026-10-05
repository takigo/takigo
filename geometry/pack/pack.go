// Package pack implements the pack geometry manager, which arranges
// children around the edges of a cavity. It ports tk/generic/tkPack.c.
package pack

import (
	"errors"
	"fmt"
	"slices"

	"github.com/takigo/takigo/geometry"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/window"
)

// Side specifies which edge of the cavity to pack against.
type Side = option.Side

// The sides of -side.
const (
	Top    = option.SideTop
	Bottom = option.SideBottom
	Left   = option.SideLeft
	Right  = option.SideRight
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

	// Per-call placement, not kept with the content: -in given in this
	// call, and -before/-after.
	inGiven       bool
	before, after *window.Window
}

func defaultConfig() packConfig {
	return packConfig{side: Top, fill: FillNone, anchor: option.AnchorCenter}
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
// Accepts a number of pixels or a screenunit.Distance such as screenunit.Pt(3).
func PadX[L screenunit.Length](p L) PackOption {
	return func(c *packConfig) { c.padLeft = screenunit.ToPixels(p); c.padX = 2 * c.padLeft }
}

// PadXPair sets asymmetric exterior horizontal padding (-padx {left right}).
func PadXPair[A, B screenunit.Length](left A, right B) PackOption {
	return func(c *packConfig) {
		c.padLeft = screenunit.ToPixels(left)
		c.padX = c.padLeft + screenunit.ToPixels(right)
	}
}

// PadY sets the exterior vertical padding.
// Accepts a number of pixels or a screenunit.Distance such as screenunit.Pt(3).
func PadY[L screenunit.Length](p L) PackOption {
	return func(c *packConfig) { c.padTop = screenunit.ToPixels(p); c.padY = 2 * c.padTop }
}

// PadYPair sets asymmetric exterior vertical padding (-pady {top bottom}).
func PadYPair[A, B screenunit.Length](top A, bottom B) PackOption {
	return func(c *packConfig) {
		c.padTop = screenunit.ToPixels(top)
		c.padY = c.padTop + screenunit.ToPixels(bottom)
	}
}

// IPadX sets the interior horizontal padding.
// Accepts a number of pixels or a screenunit.Distance such as screenunit.Pt(3).
func IPadX[L screenunit.Length](p L) PackOption {
	return func(c *packConfig) { c.iPadX = screenunit.ToPixels(p) }
}

// IPadY sets the interior vertical padding.
// Accepts a number of pixels or a screenunit.Distance such as screenunit.Pt(3).
func IPadY[L screenunit.Length](p L) PackOption {
	return func(c *packConfig) { c.iPadY = screenunit.ToPixels(p) }
}

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
var packers = new(geometry.Table[*packer])

// hooked records the containers whose destroy and configure hooks are
// registered.
var hooked = new(geometry.Table[bool])

type packManager struct{}

func (m *packManager) Name() string { return "pack" }

func (m *packManager) RequestProc(content *window.Window) {
	if p, ok := packers.Get(containerFor(content)); ok {
		p.scheduleArrange()
	}
}

// LostContentProc drops content from its container, e.g. when content is
// destroyed or taken over by another geometry manager.
func (m *packManager) LostContentProc(content *window.Window) {
	container := containerFor(content)
	containerOf.Delete(content)
	if p, ok := packers.Get(container); ok {
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
	p, ok := packers.Get(container)
	if !ok {
		return
	}
	packers.Delete(container)
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

// In is pack's -in: manage the content inside container, which must be the
// content's parent or a descendant of it (Tk_MaintainGeometry keeps it there).
// The content goes at the end of container's packing order.
func In(container window.Windower) PackOption {
	return func(c *packConfig) { c.in, c.inGiven = container.Window(), true }
}

// Before is pack's -before: put the content in sibling's container, just
// before sibling in its packing order. sibling must be packed.
func Before(sibling window.Windower) PackOption {
	return func(c *packConfig) { c.before, c.after = sibling.Window(), nil }
}

// After is pack's -after: put the content in sibling's container, just after
// sibling in its packing order. sibling must be packed.
func After(sibling window.Windower) PackOption {
	return func(c *packConfig) { c.after, c.before = sibling.Window(), nil }
}

// containerOf records each content window's container (Tk's containerPtr).
var containerOf = new(geometry.Table[*window.Window])

// ErrNotPacked is wrapped by the error Pack returns when the Before or
// After sibling is not packed.
var ErrNotPacked = errors.New("pack: window is not packed")

// Pack packs children, like Tk's "pack configure .w1 .w2 ... options".
// It accepts a geometry.Elementer (e.g. geometry.Group) of one or more
// widgets, which all receive the same options.
//
// As in tkPack.c ConfigureContent, content that is already packed keeps the
// options not given here, and its place in the packing order unless In,
// Before or After is given; new content starts from the defaults and goes
// at the end. With Before or After, the first widget goes next to the
// sibling and the rest follow it in order.
func Pack(children geometry.Elementer, opts ...PackOption) error {
	elements := children.GeometryElements()

	// Arrange each affected container once, after all elements are added,
	// rather than once per element.
	var touched []*packer
	touch := func(p *packer) {
		if !slices.Contains(touched, p) {
			touched = append(touched, p)
		}
	}
	// With a position (In, Before, After) the first content goes there and
	// each later one right after the one before it (Tk's prevPtr).
	var prev *window.Window
	positionGiven := false
	for i, elem := range elements {
		w := elem.Window()
		cur := containerOf.Of(w)
		var entry *packEntry
		if cur != nil {
			if p, ok := packers.Get(cur); ok {
				entry = p.entry(w)
			}
		}
		cfg := defaultConfig()
		if entry != nil {
			cfg = entry.config
		}
		for _, opt := range opts {
			opt(&cfg)
		}

		// Where the content goes: container and position in its order.
		container := cur
		var at *window.Window // insert next to this sibling
		afterAt := true
		inGiven := cfg.inGiven
		switch {
		case i > 0 && positionGiven:
			if prev == nil {
				continue
			}
			container, at = containerFor(prev), prev
		case cfg.before != nil || cfg.after != nil:
			positionGiven = true
			sib := cfg.after
			if sib == nil {
				sib, afterAt = cfg.before, false
			}
			if containerOf.Of(sib) == nil {
				return fmt.Errorf("%w: %s", ErrNotPacked, sib.PathName)
			}
			if sib != w { // next to itself: stay in place (tkPack.c)
				container, at = containerOf.Of(sib), sib
			}
		case inGiven:
			positionGiven = true
			container = cfg.in
		case entry == nil:
			container = w.Parent
		}
		if container == nil {
			continue
		}
		cfg.in = nil
		if container != w.Parent {
			cfg.in = container
		}
		cfg.inGiven, cfg.before, cfg.after = false, nil, nil

		if entry != nil && at == nil && container == cur && !inGiven {
			entry.config = cfg
			if p, ok := packers.Get(cur); ok {
				touch(p)
			}
			prev = w
			continue
		}

		if entry != nil && container == cur {
			// Move within its packing order.
			p, _ := packers.Get(cur)
			p.entries = slices.DeleteFunc(p.entries, func(x *packEntry) bool { return x == entry })
			entry.config = cfg
			p.insert(entry, at, afterAt)
			touch(p)
			prev = w
			continue
		}
		if entry != nil {
			if op, ok := packers.Get(cur); ok {
				op.remove(w)
				touch(op)
			}
		}
		containerOf.Set(w, container)
		geometry.ManageGeometry(w, mgr)
		p := packerFor(container)
		e := &packEntry{window: w, config: cfg}
		p.insert(e, at, afterAt)
		touch(p)
		prev = w
	}
	for _, p := range touched {
		p.scheduleArrange()
	}
	return nil
}

// packerFor returns container's packer, creating it and its hooks.
func packerFor(container *window.Window) *packer {
	p, ok := packers.Get(container)
	if !ok {
		p = &packer{container: container}
		packers.Set(container, p)
	}
	if !hooked.Of(container) {
		// The packer is dropped when its last content goes but the
		// hooks stay with the window, so register them once.
		hooked.Set(container, true)
		container.OnDestroy(func() {
			forgetContainer(container)
			hooked.Delete(container)
		})
		// Re-layout when resized by external forces (e.g. PanedWindow).
		container.OnConfigure(func() {
			if pp, ok2 := packers.Get(container); ok2 {
				pp.scheduleArrange()
			}
		})
	}
	return p
}

// entry returns child's entry, or nil.
func (p *packer) entry(child *window.Window) *packEntry {
	for _, e := range p.entries {
		if e.window == child {
			return e
		}
	}
	return nil
}

// insert adds e next to sibling (after it when after is set), or at the
// end when sibling is nil or not in this packer.
func (p *packer) insert(e *packEntry, sibling *window.Window, after bool) {
	i := len(p.entries)
	if sibling != nil {
		if j := slices.IndexFunc(p.entries, func(x *packEntry) bool { return x.window == sibling }); j >= 0 {
			i = j
			if after {
				i++
			}
		}
	}
	p.entries = slices.Insert(p.entries, i, e)
}

// Forget removes a child from pack management.
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
	w.GeomManager = nil
	// pack forget unmaps the content and re-packs the rest.
	if w.IsMapped() && w.PlatformID != platform.WindowID(0) {
		w.Display.Server.UnmapWindow(w.PlatformID)
		window.MarkUnmapped(w)
	}
	if p, ok := packers.Get(parent); ok {
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
		packers.Delete(p.container)
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

// ArrangeAll triggers layout for the pack-managed containers in root's
// subtree, root included.
func ArrangeAll(root *window.Window) {
	if root == nil {
		return
	}
	if p, ok := packers.Get(root); ok {
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
		if p, ok := packers.Get(w); ok && p.hasForeign() {
			p.scheduleArrange()
		}
	})
	window.AddUnmappedHook(func(w *window.Window) {
		if p, ok := packers.Get(w); ok {
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
	if p, ok := packers.Get(container); ok {
		p.arrange()
	}
}

// scheduleArrange re-arranges the container at idle time.
func (p *packer) scheduleArrange() {
	geometry.WhenIdle(p.container, &p.pending, p.arrange)
}
