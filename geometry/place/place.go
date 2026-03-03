// Package place implements the place geometry manager, which positions
// children using absolute and/or relative coordinates.
// It ports tk/generic/tkPlace.c.
package place

import (
	"math"

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
}

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
	entries   []*placeEntry
}

// singleton manager instance.
var mgr = &placeManager{}

// placers tracks per-container state.
var placers = map[*window.Window]*placer{}

type placeManager struct{}

func (m *placeManager) Name() string { return "place" }

func (m *placeManager) RequestProc(content *window.Window) {
	parent := content.Parent
	if parent == nil {
		return
	}
	if p, ok := placers[parent]; ok {
		p.arrange()
	}
}

func (m *placeManager) LostContentProc(content *window.Window) {
	parent := content.Parent
	if parent == nil {
		return
	}
	if p, ok := placers[parent]; ok {
		p.remove(content)
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

	w.GeomManager = mgr

	p, ok := placers[parent]
	if !ok {
		p = &placer{container: parent}
		placers[parent] = p
	}

	// Update or add entry.
	for _, e := range p.entries {
		if e.window == w {
			e.config = cfg
			p.arrange()
			return
		}
	}

	p.entries = append(p.entries, &placeEntry{window: w, config: cfg})
	p.arrange()
}

// Forget removes a child from place management.
func Forget(child window.Windower) {
	w := child.Window()
	parent := w.Parent
	if parent == nil {
		return
	}
	if p, ok := placers[parent]; ok {
		p.remove(w)
	}
	w.GeomManager = nil
}

func (p *placer) remove(child *window.Window) {
	for i, e := range p.entries {
		if e.window == child {
			p.entries = append(p.entries[:i], p.entries[i+1:]...)
			break
		}
	}
	if len(p.entries) == 0 {
		delete(placers, p.container)
	}
}

// arrange positions all placed children.
func (p *placer) arrange() {
	container := p.container
	if container.PlatformID == platform.WindowID(0) {
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
		x := cfg.x + int(cfg.relX*float64(containerW)) + container.InternalBorderLeft
		y := cfg.y + int(cfg.relY*float64(containerH)) + container.InternalBorderTop

		// Compute size.
		var childW, childH int

		if cfg.width >= 0 || !math.IsNaN(cfg.relWidth) {
			childW = 0
			if cfg.width >= 0 {
				childW += cfg.width
			}
			if !math.IsNaN(cfg.relWidth) {
				childW += int(cfg.relWidth * float64(containerW))
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
				childH += int(cfg.relHeight * float64(containerH))
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

		child.X = x
		child.Y = y
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
			if child.Flags&window.FlagMapped == 0 {
				container.Display.Server.MapWindow(child.PlatformID)
				child.Flags |= window.FlagMapped
			}
		}
	}
}

// ArrangeAll triggers layout for all place-managed containers.
func ArrangeAll() {
	for _, p := range placers {
		p.arrange()
	}
}

// ArrangeContainer triggers layout for a specific container.
func ArrangeContainer(container *window.Window) {
	if p, ok := placers[container]; ok {
		p.arrange()
	}
}
