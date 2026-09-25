package canvas

import (
	"math"

	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/platform"
)

// RectOvalItem implements rectangle and oval canvas items.
type RectOvalItem struct {
	ItemBase
	typeName     string // "rectangle" or "oval"
	coords       [4]float64
	fill         *color.ColorRef
	outline      *color.ColorRef
	outlineWidth int
	dash         []byte
	fillGC       platform.GCID
	outlineGC    platform.GCID
}

func newRectOvalItem(typeName string, x1, y1, x2, y2 float64, c *Canvas) *RectOvalItem {
	item := &RectOvalItem{
		typeName:     typeName,
		coords:       [4]float64{x1, y1, x2, y2},
		outlineWidth: 1,
	}
	// Default outline is black.
	item.outline = &color.ColorRef{Pixel: 0x000000}
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

func (r *RectOvalItem) Display(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	clipX, clipY, clipW, clipH, originX, originY int) {

	x1 := drawableCoord(r.coords[0], originX)
	y1 := drawableCoord(r.coords[1], originY)
	x2 := drawableCoord(r.coords[2], originX)
	y2 := drawableCoord(r.coords[3], originY)
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
		if fill := r.fillFor(r.fill); fill != nil {
			d.SetForeground(gc, fill.Pixel)
			off := r.canvas.stippleOn(d, drawable, gc, r.stipple, originX, originY)
			d.FillRectangle(drawable, gc, x1, y1, uint(w), uint(h))
			off()
		}
		if r.outline != nil && r.outlineWidth > 0 {
			d.SetForeground(gc, r.outline.Pixel)
			d.SetLineAttributes(gc, uint(r.outlineWidth), platform.LineSolid, platform.CapButt, platform.JoinMiter)
			if len(r.dash) > 0 {
				d.SetLineAttributes(gc, uint(r.outlineWidth), platform.LineOnOffDash, platform.CapButt, platform.JoinMiter)
				d.SetDashes(gc, 0, r.dash)
			}
			d.DrawRectangle(drawable, gc, x1, y1, uint(w), uint(h))
			// Reset line style.
			d.SetLineAttributes(gc, 1, platform.LineSolid, platform.CapButt, platform.JoinMiter)
		}
	} else { // oval
		if fill := r.fillFor(r.fill); fill != nil {
			d.SetForeground(gc, fill.Pixel)
			off := r.canvas.stippleOn(d, drawable, gc, r.stipple, originX, originY)
			d.FillArc(drawable, gc, x1, y1, uint(w), uint(h), 0, 360*64)
			off()
		}
		if r.outline != nil && r.outlineWidth > 0 {
			d.SetForeground(gc, r.outline.Pixel)
			d.SetLineAttributes(gc, uint(r.outlineWidth), platform.LineSolid, platform.CapButt, platform.JoinMiter)
			if len(r.dash) > 0 {
				d.SetLineAttributes(gc, uint(r.outlineWidth), platform.LineOnOffDash, platform.CapButt, platform.JoinMiter)
				d.SetDashes(gc, 0, r.dash)
			}
			d.DrawArc(drawable, gc, x1, y1, uint(w), uint(h), 0, 360*64)
			d.SetLineAttributes(gc, 1, platform.LineSolid, platform.CapButt, platform.JoinMiter)
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

func (r *RectOvalItem) Delete(d platform.DisplayServer) {
	// No per-item GCs allocated (we use shared gc), nothing to free.
}

// Postscript emits a PostScript representation of the rectangle or oval.
//
// Mirrors tk/generic/tkCanvRectOval.c:RRectOvalToPostscript / ROvalToPostscript.
// Rectangle uses a closed 4-point path; oval uses a closed ellipse path
// (PostScript ellipse via the prolog's Ellipse / ova macros, but to stay
// portable we emit a circle approximation built from moveto/lineto — Tk's
// ellipse macro requires it to be defined in the prolog. We use moveto+lineto
// along the perimeter with sufficient segments for visual fidelity).
func (r *RectOvalItem) Postscript(ps *PSContext) error {
	if r.State() == ItemStateHidden || ps.Prepass {
		// Even on prepass we should still register fonts; rectoval has none.
		return nil
	}
	switch r.typeName {
	case "rectangle":
		return psRectangle(r, ps)
	case "oval":
		return psOval(r, ps)
	}
	return nil
}

func psRectangle(r *RectOvalItem, ps *PSContext) error {
	ps.Path([]float64{
		float64(r.X1), float64(r.Y1),
		float64(r.X2), float64(r.Y1),
		float64(r.X2), float64(r.Y2),
		float64(r.X1), float64(r.Y2),
	})
	if r.fill != nil {
		ps.Color(r.fill)
		ps.write("gsave fill grestore newpath\n")
		// Re-emit the path for the outline.
		ps.Path([]float64{
			float64(r.X1), float64(r.Y1),
			float64(r.X2), float64(r.Y1),
			float64(r.X2), float64(r.Y2),
			float64(r.X1), float64(r.Y2),
		})
	}
	if r.outline != nil && r.outlineWidth > 0 {
		ps.Outline(r.outlineWidth, dashInts(r.dash), 0, r.outline, nil)
	}
	ps.write("newpath\n")
	return nil
}

func psOval(r *RectOvalItem, ps *PSContext) error {
	// Approximate the oval with a closed polygon (60 segments).
	// Tk's ellipse uses its own prolog macro; we keep takigo self-contained.
	cx := float64(r.X1+r.X2) / 2.0
	cy := float64(r.Y1+r.Y2) / 2.0
	rx := float64(r.X2-r.X1) / 2.0
	ry := float64(r.Y2-r.Y1) / 2.0
	const segs = 60
	pts := make([]float64, 0, segs*2)
	for i := 0; i < segs; i++ {
		theta := 2 * math.Pi * float64(i) / float64(segs)
		pts = append(pts, cx+rx*math.Cos(theta), cy+ry*math.Sin(theta))
	}
	ps.Path(pts)
	if r.fill != nil {
		ps.Color(r.fill)
		ps.write("gsave fill grestore newpath\n")
		// Re-emit for stroke.
		ps.Path(pts)
	}
	if r.outline != nil && r.outlineWidth > 0 {
		ps.Outline(r.outlineWidth, dashInts(r.dash), 0, r.outline, nil)
	}
	ps.write("newpath\n")
	return nil
}

func dashInts(b []byte) []int {
	out := make([]int, len(b))
	for i, v := range b {
		out[i] = int(v)
	}
	return out
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
