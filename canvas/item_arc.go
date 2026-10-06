package canvas

import (
	"math"

	"github.com/takigo/takigo/color"
	"github.com/takigo/takigo/platform"
)

// ArcItem implements an arc/chord/pieslice canvas item.
type ArcItem struct {
	ItemBase

	// Per-Display state for the outline stipple (see setLineAttrs).
	drawable         platform.DrawableID
	originX, originY int
	stippled         bool
	coords           [4]float64 // bounding box of the ellipse
	start            float64    // start angle in degrees
	extent           float64    // angular extent in degrees
	style            ArcStyle
	fill             *color.ColorRef
	outline          *color.ColorRef
	outlineWidth     int
	dash             []byte
}

func newArcItem(x1, y1, x2, y2 float64, c *Canvas) *ArcItem {
	item := &ArcItem{
		coords:       [4]float64{x1, y1, x2, y2},
		start:        0,
		extent:       90,
		style:        ArcStylePieslice,
		outlineWidth: 1,
	}
	item.outline = c.defaultInk()
	item.canvas = c
	item.updateBBox()
	return item
}

func (a *ArcItem) base() *ItemBase { return &a.ItemBase }
func (a *ArcItem) Type() string    { return "arc" }

func (a *ArcItem) BBox() (x1, y1, x2, y2 int) {
	return a.X1, a.Y1, a.X2, a.Y2
}

func (a *ArcItem) Coords() []float64 {
	return a.coords[:]
}

func (a *ArcItem) SetCoords(coords []float64) error {
	if len(coords) >= 4 {
		copy(a.coords[:], coords[:4])
		a.updateBBox()
	}
	return nil
}

func (a *ArcItem) Configure(opts []ItemOption) error {
	c := a.canvas
	for _, opt := range opts {
		if err := opt(c, a); err != nil {
			return err
		}
	}
	a.updateBBox()
	return nil
}

func (a *ArcItem) updateBBox() {
	hw := float64(a.outlineWidth)/2.0 + 1
	x1, y1, x2, y2 := a.coords[0], a.coords[1], a.coords[2], a.coords[3]
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	a.X1 = int(math.Floor(x1 - hw))
	a.Y1 = int(math.Floor(y1 - hw))
	a.X2 = int(math.Ceil(x2 + hw))
	a.Y2 = int(math.Ceil(y2 + hw))
}

func (a *ArcItem) Display(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	clipX, clipY, clipW, clipH, originX, originY int) {
	a.drawable, a.originX, a.originY = drawable, originX, originY

	x1 := drawableCoord(a.coords[0], originX)
	y1 := drawableCoord(a.coords[1], originY)
	x2 := drawableCoord(a.coords[2], originX)
	y2 := drawableCoord(a.coords[3], originY)
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

	// DisplayArc: X11 angles in 64ths of a degree, rounded.
	angle1 := int(a.start*64 + 0.5)
	angle2 := int(a.extent*64 + 0.5)

	switch a.style {
	case ArcStylePieslice:
		if a.fill != nil {
			d.SetForeground(gc, a.fill.Pixel)
			d.FillArc(drawable, gc, x1, y1, uint(w), uint(h), angle1, angle2)
		}
		if a.outline != nil && a.outlineWidth > 0 {
			d.SetForeground(gc, a.outline.Pixel)
			a.setLineAttrs(d, gc)
			d.DrawArc(drawable, gc, x1, y1, uint(w), uint(h), angle1, angle2)
			cxi, cyi, sx, sy, ex, ey := a.outlineEnds(originX, originY)
			d.DrawLine(drawable, gc, cxi, cyi, sx, sy)
			d.DrawLine(drawable, gc, cxi, cyi, ex, ey)
			a.resetLineAttrs(d, gc)
		}

	case ArcStyleChord:
		if a.fill != nil {
			// ConfigureArc gives the fill GC arc_mode ArcChord for chords.
			d.SetForeground(gc, a.fill.Pixel)
			d.SetArcMode(gc, platform.ArcChord)
			d.FillArc(drawable, gc, x1, y1, uint(w), uint(h), angle1, angle2)
			d.SetArcMode(gc, platform.ArcPieSlice)
		}
		if a.outline != nil && a.outlineWidth > 0 {
			d.SetForeground(gc, a.outline.Pixel)
			a.setLineAttrs(d, gc)
			d.DrawArc(drawable, gc, x1, y1, uint(w), uint(h), angle1, angle2)
			_, _, sx, sy, ex, ey := a.outlineEnds(originX, originY)
			d.DrawLine(drawable, gc, sx, sy, ex, ey)
			a.resetLineAttrs(d, gc)
		}

	case ArcStyleArc:
		if a.outline != nil && a.outlineWidth > 0 {
			d.SetForeground(gc, a.outline.Pixel)
			a.setLineAttrs(d, gc)
			d.DrawArc(drawable, gc, x1, y1, uint(w), uint(h), angle1, angle2)
			a.resetLineAttrs(d, gc)
		}
	}
}

// outlineEnds ports ComputeArcOutline's center1/center2 and DisplayArc's
// vertex: computed in canvas coordinates from the bbox, then each rounded
// with Tk_CanvasDrawableCoords.
func (a *ArcItem) outlineEnds(originX, originY int) (cx, cy, sx, sy, ex, ey int) {
	bx1, bx2 := min(a.coords[0], a.coords[2]), max(a.coords[0], a.coords[2])
	by1, by2 := min(a.coords[1], a.coords[3]), max(a.coords[1], a.coords[3])
	bw, bh := bx2-bx1, by2-by1
	vx, vy := (bx1+bx2)/2, (by1+by2)/2
	angle := -a.start * math.Pi / 180
	sin1, cos1 := math.Sin(angle), math.Cos(angle)
	angle -= a.extent * math.Pi / 180
	sin2, cos2 := math.Sin(angle), math.Cos(angle)
	return drawableCoord(vx, originX), drawableCoord(vy, originY),
		drawableCoord(vx+cos1*bw/2, originX), drawableCoord(vy+sin1*bh/2, originY),
		drawableCoord(vx+cos2*bw/2, originX), drawableCoord(vy+sin2*bh/2, originY)
}

func (a *ArcItem) setLineAttrs(d platform.DisplayServer, gc platform.GCID) {
	lineStyle := platform.LineSolid
	if len(a.dash) > 0 {
		lineStyle = platform.LineOnOffDash
		d.SetDashes(gc, 0, a.dash)
	}
	d.SetLineAttributes(gc, uint(a.outlineWidth), lineStyle, platform.CapButt, platform.JoinMiter)
	a.stippled = a.canvas.stippleOn(d, a.drawable, gc, a.outlineStipple, a.originX, a.originY)
}

func (a *ArcItem) resetLineAttrs(d platform.DisplayServer, gc platform.GCID) {
	d.SetLineAttributes(gc, 1, platform.LineSolid, platform.CapButt, platform.JoinMiter)
	if a.stippled {
		stippleOff(d, gc)
		a.stippled = false
	}
}

// shape returns the arc's outline as a polyline in canvas coordinates:
// the curve, closed through its chord or through the centre for chord and
// pieslice styles. Angles run counterclockwise from 3 o'clock.
func (a *ArcItem) shape() []float64 {
	x1, y1, x2, y2 := a.coords[0], a.coords[1], a.coords[2], a.coords[3]
	cx, cy := (x1+x2)/2, (y1+y2)/2
	rx, ry := math.Abs(x2-x1)/2, math.Abs(y2-y1)/2
	const segs = 60
	pts := make([]float64, 0, 2*segs+6)
	start, extent := a.start*math.Pi/180, a.extent*math.Pi/180
	for i := 0; i <= segs; i++ {
		t := start + extent*float64(i)/segs
		pts = append(pts, cx+rx*math.Cos(t), cy-ry*math.Sin(t))
	}
	switch a.style {
	case ArcStylePieslice:
		pts = append(pts, cx, cy, pts[0], pts[1])
	case ArcStyleChord:
		pts = append(pts, pts[0], pts[1])
	}
	return pts
}

// PointDistance ports ArcToPoint: zero inside a filled chord or pieslice,
// otherwise the distance to the outline less half its width.
func (a *ArcItem) PointDistance(x, y float64) float64 {
	pts := a.shape()
	if a.style != ArcStyleArc && a.fill != nil && pointInPolygon(x, y, pts) {
		return 0
	}
	best := math.MaxFloat64
	for i := 0; i+3 < len(pts); i += 2 {
		best = min(best, segmentPointDistance(x, y, pts[i], pts[i+1], pts[i+2], pts[i+3]))
	}
	if a.outline != nil {
		best -= float64(a.outlineWidth) / 2
	}
	return max(best, 0)
}

// AreaOverlap ports ArcToArea: 1 when the arc lies inside the rectangle,
// -1 when it misses it, 0 when they overlap.
func (a *ArcItem) AreaOverlap(ax1, ay1, ax2, ay2 float64) int {
	pts := a.shape()
	hw := 0.0
	if a.outline != nil {
		hw = float64(a.outlineWidth) / 2
	}
	inside, anyIn := true, false
	for i := 0; i+1 < len(pts); i += 2 {
		px, py := pts[i], pts[i+1]
		in := px-hw >= ax1 && px+hw <= ax2 && py-hw >= ay1 && py+hw <= ay2
		inside = inside && in
		anyIn = anyIn || (px >= ax1-hw && px <= ax2+hw && py >= ay1-hw && py <= ay2+hw)
	}
	switch {
	case inside:
		return 1
	case anyIn:
		return 0
	}
	// No outline point is near the rectangle: they still overlap when an
	// outline segment crosses it or a filled shape contains it.
	for i := 0; i+3 < len(pts); i += 2 {
		if segmentHitsRect(pts[i], pts[i+1], pts[i+2], pts[i+3], ax1-hw, ay1-hw, ax2+hw, ay2+hw) {
			return 0
		}
	}
	if a.style != ArcStyleArc && a.fill != nil && pointInPolygon((ax1+ax2)/2, (ay1+ay2)/2, pts) {
		return 0
	}
	return -1
}

// segmentHitsRect reports whether segment (x1,y1)-(x2,y2) meets the
// rectangle, by Liang-Barsky clipping.
func segmentHitsRect(x1, y1, x2, y2, rx1, ry1, rx2, ry2 float64) bool {
	t0, t1 := 0.0, 1.0
	dx, dy := x2-x1, y2-y1
	for _, e := range [4][2]float64{{-dx, x1 - rx1}, {dx, rx2 - x1}, {-dy, y1 - ry1}, {dy, ry2 - y1}} {
		p, q := e[0], e[1]
		if p == 0 {
			if q < 0 {
				return false
			}
			continue
		}
		r := q / p
		if p < 0 {
			t0 = max(t0, r)
		} else {
			t1 = min(t1, r)
		}
		if t0 > t1 {
			return false
		}
	}
	return true
}

func (a *ArcItem) Scale(ox, oy, sx, sy float64) {
	for i := 0; i < 4; i += 2 {
		a.coords[i] = ox + (a.coords[i]-ox)*sx
		a.coords[i+1] = oy + (a.coords[i+1]-oy)*sy
	}
	a.updateBBox()
}

func (a *ArcItem) Translate(dx, dy float64) {
	a.coords[0] += dx
	a.coords[1] += dy
	a.coords[2] += dx
	a.coords[3] += dy
	a.updateBBox()
}

func (a *ArcItem) Delete(d platform.DisplayServer) {}

// Postscript emits a PostScript representation of the arc/chord/pieslice.
//
// Mirrors tk/generic/tkCanvArc.c:ArcToPostscript. We approximate the ellipse
// arc as a series of line segments (Tk relies on a prolog "Ellipse" macro;
// emitting sampled segments is visually equivalent and keeps our prolog
// minimal).
func (a *ArcItem) Postscript(ps *PSContext) error {
	if a.State() == ItemStateHidden {
		return nil
	}
	x1, y1, x2, y2 := a.coords[0], a.coords[1], a.coords[2], a.coords[3]
	cx := (x1 + x2) / 2
	cy := (y1 + y2) / 2
	rx := (x2 - x1) / 2
	ry := (y2 - y1) / 2
	if rx == 0 || ry == 0 {
		return nil
	}
	startRad := a.start * math.Pi / 180
	extentRad := a.extent * math.Pi / 180

	const segs = 60
	// Build the arc polyline.
	pts := make([]float64, 0, segs*2)
	for i := 0; i <= segs; i++ {
		t := startRad + extentRad*float64(i)/float64(segs)
		pts = append(pts, cx+rx*math.Cos(t), cy+ry*math.Sin(t))
	}

	// Pieslice and chord close the path back to the center.
	if a.style == ArcStylePieslice || a.style == ArcStyleChord {
		pts = append(pts, cx, cy)
	}

	ps.Path(pts)

	switch a.style {
	case ArcStylePieslice, ArcStyleChord:
		if a.fill != nil {
			ps.Color(a.fill)
			ps.write("gsave fill grestore newpath\n")
			ps.Path(pts)
		}
	}

	if a.outline != nil && a.outlineWidth > 0 {
		ps.Outline(a.outlineWidth, dashInts(a.dash), 0, a.outline, nil)
	}
	ps.write("newpath\n")
	return nil
}
