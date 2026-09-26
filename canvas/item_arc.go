package canvas

import (
	"math"

	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/platform"
)

// ArcItem implements an arc/chord/pieslice canvas item.
type ArcItem struct {
	ItemBase

	// Per-Display state for the outline stipple (see setLineAttrs).
	drawable         platform.DrawableID
	originX, originY int
	stippleOff       func()
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
	item.outline = &color.ColorRef{Pixel: 0x000000}
	item.ItemBase.canvas = c
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
			d.SetForeground(gc, a.fill.Pixel)
			d.FillArc(drawable, gc, x1, y1, uint(w), uint(h), angle1, angle2)
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
	a.stippleOff = a.canvas.stippleOn(d, a.drawable, gc, a.outlineStipple, a.originX, a.originY)
}

func (a *ArcItem) resetLineAttrs(d platform.DisplayServer, gc platform.GCID) {
	d.SetLineAttributes(gc, 1, platform.LineSolid, platform.CapButt, platform.JoinMiter)
	if a.stippleOff != nil {
		a.stippleOff()
		a.stippleOff = nil
	}
}

func (a *ArcItem) PointDistance(x, y float64) float64 {
	// Approximate: use distance to bounding box.
	return rectPointDistance(x, y, a.coords[0], a.coords[1], a.coords[2], a.coords[3])
}

func (a *ArcItem) AreaOverlap(ax1, ay1, ax2, ay2 float64) int {
	x1, y1, x2, y2 := a.coords[0], a.coords[1], a.coords[2], a.coords[3]
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	if ax2 < x1 || ax1 > x2 || ay2 < y1 || ay1 > y2 {
		return -1
	}
	if ax1 <= x1 && ax2 >= x2 && ay1 <= y1 && ay2 >= y2 {
		return 1
	}
	return 0
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
