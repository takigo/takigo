// Package canvas implements a 2D drawing surface widget with arbitrary
// graphical items. It ports tk/generic/tkCanvas.c.
package canvas

import (
	"slices"

	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/platform"
)

// ItemState controls per-item visibility.
type ItemState int

const (
	ItemStateNormal ItemState = iota
	ItemStateDisabled
	ItemStateHidden
)

// Item is the interface implemented by all canvas item types.
type Item interface {
	// Type returns the item type name (e.g. "rectangle", "oval", "line").
	Type() string

	// BBox returns the bounding box in canvas coordinates.
	BBox() (x1, y1, x2, y2 int)

	// Coords returns the item's coordinates as float64 pairs.
	Coords() []float64

	// SetCoords replaces the item's coordinates.
	SetCoords(coords []float64) error

	// Configure applies item-specific options.
	Configure(opts []ItemOption) error

	// Display draws the item onto the given drawable.
	// clipX/clipY/clipW/clipH define the visible area; originX/originY
	// are the canvas coordinates at the drawable origin.
	Display(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
		clipX, clipY, clipW, clipH, originX, originY int)

	// PointDistance returns the distance from (x,y) in canvas coords
	// to the nearest edge of the item. Returns 0 if inside.
	PointDistance(x, y float64) float64

	// AreaOverlap tests overlap with the given rectangle.
	// Returns -1 (outside), 0 (partial overlap), or 1 (fully enclosed).
	AreaOverlap(x1, y1, x2, y2 float64) int

	// Scale scales the item about the given origin.
	Scale(originX, originY, scaleX, scaleY float64)

	// Translate moves the item by (dx, dy).
	Translate(dx, dy float64)

	// Delete frees platform resources held by the item.
	Delete(d platform.DisplayServer)

	// Postscript emits a PostScript representation of the item into ps.
	// Returns nil error; PS emission is best-effort and never aborts the
	// output. Items in state ItemStateHidden should emit nothing.
	Postscript(ps *PSContext) error

	// State returns the item's per-item state (normal/disabled/hidden).
	State() ItemState
}

// ItemBase holds fields common to all item types.
type ItemBase struct {
	ID             int64
	Tags           []string
	X1, Y1, X2, Y2 int // integer bounding box in canvas coords
	state          ItemState
	canvas         *Canvas // back-pointer for color/font/visual resolution

	stipple, outlineStipple string // -stipple / -outlinestipple bitmap specs

	activeFill, disabledFill *color.ColorRef // -activefill / -disabledfill
}

// fillFor picks the fill for the item's state, as the item display procs
// do: -disabledfill when disabled, -activefill when it is the current item.
func (b *ItemBase) fillFor(fill *color.ColorRef) *color.ColorRef {
	if b.state == ItemStateDisabled && b.disabledFill != nil {
		return b.disabledFill
	}
	if b.state != ItemStateDisabled && b.activeFill != nil && b.canvas != nil &&
		b.canvas.currentItem != nil && b.canvas.currentItem.id == b.ID {
		return b.activeFill
	}
	return fill
}

// State returns the item's per-item state.
func (b *ItemBase) State() ItemState { return b.state }

// HasTag returns true if the item has the given tag.
func (b *ItemBase) HasTag(tag string) bool {
	return slices.Contains(b.Tags, tag)
}

// AddTag adds a tag if not already present.
func (b *ItemBase) AddTag(tag string) {
	if !b.HasTag(tag) {
		b.Tags = append(b.Tags, tag)
		if b.canvas != nil {
			b.canvas.tagIndexAdd(tag, b.ID)
		}
	}
}

// RemoveTag removes a tag if present.
func (b *ItemBase) RemoveTag(tag string) {
	for i, t := range b.Tags {
		if t == tag {
			b.Tags = append(b.Tags[:i], b.Tags[i+1:]...)
			if b.canvas != nil {
				b.canvas.tagIndexRemove(tag, b.ID)
			}
			return
		}
	}
}

// itemEntry wraps an Item with its ID for the display list.
type itemEntry struct {
	id   int64
	item Item
	mark uint64 // last markEntries stamp (membership during list passes)
}

// drawableCoord ports Tk_CanvasDrawableCoords: canvas coordinate v minus the
// drawable origin, rounded half away from zero.
func drawableCoord(v float64, origin int) int {
	t := v - float64(origin)
	if t > 0 {
		t += 0.5
	} else {
		t -= 0.5
	}
	return int(t)
}

// drawablePoint converts a canvas point to an X point like drawableCoord.
func drawablePoint(x, y float64, originX, originY int) platform.Point {
	return platform.Point{X: int16(drawableCoord(x, originX)), Y: int16(drawableCoord(y, originY))}
}
