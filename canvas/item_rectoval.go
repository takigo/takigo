package canvas

import (
	"math"

	"github.com/msorc/takigo/internal/xlib"
)

// RectOvalItem implements rectangle and oval canvas items.
type RectOvalItem struct {
	ItemBase
	typeName     string // "rectangle" or "oval"
	coords       [4]float64
	fill         *colorRef
	outline      *colorRef
	outlineWidth int
	dash         []byte
	fillGC       xlib.GC
	outlineGC    xlib.GC
}

func newRectOvalItem(typeName string, x1, y1, x2, y2 float64, c *Canvas) *RectOvalItem {
	item := &RectOvalItem{
		typeName:     typeName,
		coords:       [4]float64{x1, y1, x2, y2},
		outlineWidth: 1,
	}
	// Default outline is black.
	item.outline = &colorRef{Pixel: 0x000000}
	item.ItemBase.canvas = c
	item.updateBBox()
	return item
}

func (r *RectOvalItem) base() *ItemBase { return &r.ItemBase }

func (r *RectOvalItem) Type() string { return r.typeName }

func (r *RectOvalItem) BBox() (x1, y1, x2, y2 int) {
	return r.X1, r.Y1, r.X2, r.Y2
}

func (r *RectOvalItem) Coords() []float64 {
	return r.coords[:]
}

func (r *RectOvalItem) SetCoords(coords []float64) error {
	if len(coords) >= 4 {
		copy(r.coords[:], coords[:4])
		r.updateBBox()
	}
	return nil
}

func (r *RectOvalItem) Configure(opts []ItemOption) error {
	c := r.canvas
	for _, opt := range opts {
		if err := opt(c, r); err != nil {
			return err
		}
	}
	r.updateBBox()
	return nil
}

func (r *RectOvalItem) updateBBox() {
	hw := float64(r.outlineWidth) / 2.0
	x1, y1, x2, y2 := r.coords[0], r.coords[1], r.coords[2], r.coords[3]
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	r.X1 = int(math.Floor(x1 - hw))
	r.Y1 = int(math.Floor(y1 - hw))
	r.X2 = int(math.Ceil(x2 + hw))
	r.Y2 = int(math.Ceil(y2 + hw))
}

func (r *RectOvalItem) Display(d *xlib.Display, drawable xlib.Drawable, gc xlib.GC,
	clipX, clipY, clipW, clipH, originX, originY int) {

	x1 := int(r.coords[0]) - originX
	y1 := int(r.coords[1]) - originY
	x2 := int(r.coords[2]) - originX
	y2 := int(r.coords[3]) - originY
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	w := x2 - x1
	h := y2 - y1
	if w <= 0 || h <= 0 {
		return
	}

	if r.typeName == "rectangle" {
		if r.fill != nil {
			d.SetForeground(gc, r.fill.Pixel)
			d.FillRectangle(drawable, gc, x1, y1, uint(w), uint(h))
		}
		if r.outline != nil && r.outlineWidth > 0 {
			d.SetForeground(gc, r.outline.Pixel)
			d.SetLineAttributes(gc, uint(r.outlineWidth), xlib.LineSolid, xlib.CapButt, xlib.JoinMiter)
			if len(r.dash) > 0 {
				d.SetLineAttributes(gc, uint(r.outlineWidth), xlib.LineOnOffDash, xlib.CapButt, xlib.JoinMiter)
				d.SetDashes(gc, 0, r.dash)
			}
			d.DrawRectangle(drawable, gc, x1, y1, uint(w), uint(h))
			// Reset line style.
			d.SetLineAttributes(gc, 1, xlib.LineSolid, xlib.CapButt, xlib.JoinMiter)
		}
	} else { // oval
		if r.fill != nil {
			d.SetForeground(gc, r.fill.Pixel)
			d.FillArc(drawable, gc, x1, y1, uint(w), uint(h), 0, 360*64)
		}
		if r.outline != nil && r.outlineWidth > 0 {
			d.SetForeground(gc, r.outline.Pixel)
			d.SetLineAttributes(gc, uint(r.outlineWidth), xlib.LineSolid, xlib.CapButt, xlib.JoinMiter)
			if len(r.dash) > 0 {
				d.SetLineAttributes(gc, uint(r.outlineWidth), xlib.LineOnOffDash, xlib.CapButt, xlib.JoinMiter)
				d.SetDashes(gc, 0, r.dash)
			}
			d.DrawArc(drawable, gc, x1, y1, uint(w), uint(h), 0, 360*64)
			d.SetLineAttributes(gc, 1, xlib.LineSolid, xlib.CapButt, xlib.JoinMiter)
		}
	}
}

func (r *RectOvalItem) PointDistance(x, y float64) float64 {
	x1, y1, x2, y2 := r.coords[0], r.coords[1], r.coords[2], r.coords[3]
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	if r.typeName == "rectangle" {
		return rectPointDistance(x, y, x1, y1, x2, y2)
	}
	// Oval: use ellipse equation.
	return ovalPointDistance(x, y, x1, y1, x2, y2)
}

func (r *RectOvalItem) AreaOverlap(ax1, ay1, ax2, ay2 float64) int {
	x1, y1, x2, y2 := r.coords[0], r.coords[1], r.coords[2], r.coords[3]
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	// Check if completely outside.
	if ax2 < x1 || ax1 > x2 || ay2 < y1 || ay1 > y2 {
		return -1
	}
	// Check if completely enclosed.
	if ax1 <= x1 && ax2 >= x2 && ay1 <= y1 && ay2 >= y2 {
		return 1
	}
	return 0
}

func (r *RectOvalItem) Scale(ox, oy, sx, sy float64) {
	for i := 0; i < 4; i += 2 {
		r.coords[i] = ox + (r.coords[i]-ox)*sx
		r.coords[i+1] = oy + (r.coords[i+1]-oy)*sy
	}
	r.updateBBox()
}

func (r *RectOvalItem) Translate(dx, dy float64) {
	r.coords[0] += dx
	r.coords[1] += dy
	r.coords[2] += dx
	r.coords[3] += dy
	r.updateBBox()
}

func (r *RectOvalItem) Delete(d *xlib.Display) {
	// No per-item GCs allocated (we use shared gc), nothing to free.
}

// --- Geometry helpers ---

func rectPointDistance(px, py, x1, y1, x2, y2 float64) float64 {
	dx := 0.0
	if px < x1 {
		dx = x1 - px
	} else if px > x2 {
		dx = px - x2
	}

	dy := 0.0
	if py < y1 {
		dy = y1 - py
	} else if py > y2 {
		dy = py - y2
	}

	if dx == 0 && dy == 0 {
		return 0 // inside
	}
	return math.Sqrt(dx*dx + dy*dy)
}

func ovalPointDistance(px, py, x1, y1, x2, y2 float64) float64 {
	cx := (x1 + x2) / 2
	cy := (y1 + y2) / 2
	rx := (x2 - x1) / 2
	ry := (y2 - y1) / 2

	if rx <= 0 || ry <= 0 {
		return math.Sqrt((px-cx)*(px-cx) + (py-cy)*(py-cy))
	}

	// Normalized coordinates.
	nx := (px - cx) / rx
	ny := (py - cy) / ry
	dist := math.Sqrt(nx*nx+ny*ny) - 1.0
	if dist <= 0 {
		return 0
	}
	// Scale back to pixel distance (approximate).
	return dist * math.Min(rx, ry)
}
