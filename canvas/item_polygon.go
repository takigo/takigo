package canvas

import (
	"math"

	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/platform"
)

// PolygonItem implements a filled polygon canvas item.
type PolygonItem struct {
	ItemBase
	coords       []float64
	fill         *color.ColorRef
	outline      *color.ColorRef
	outlineWidth int
	dash         []byte
	smooth       bool
	splineSteps  int
}

func newPolygonItem(coords []float64, c *Canvas) *PolygonItem {
	item := &PolygonItem{
		coords:       append([]float64{}, coords...),
		outlineWidth: 1,
		splineSteps:  12,
	}
	// Tk 9 defaults (tkCanvPoly.c): no -fill, -outline DEF_CANVITEM_OUTLINE
	// (black on unix).
	item.outline = c.defaultInk()
	item.ItemBase.canvas = c
	item.updateBBox()
	return item
}

func (p *PolygonItem) base() *ItemBase { return &p.ItemBase }
func (p *PolygonItem) Type() string    { return "polygon" }

func (p *PolygonItem) BBox() (x1, y1, x2, y2 int) {
	return p.X1, p.Y1, p.X2, p.Y2
}

func (p *PolygonItem) Coords() []float64 {
	out := make([]float64, len(p.coords))
	copy(out, p.coords)
	return out
}

func (p *PolygonItem) SetCoords(coords []float64) error {
	p.coords = append(p.coords[:0], coords...)
	p.updateBBox()
	return nil
}

func (p *PolygonItem) Configure(opts []ItemOption) error {
	c := p.canvas
	for _, opt := range opts {
		if err := opt(c, p); err != nil {
			return err
		}
	}
	p.updateBBox()
	return nil
}

// updateBBox ports ComputePolygonBbox (tk/generic/tkCanvPoly.c): the first
// point truncated, the others rounded (TkIncludePoint), grown by half the
// outline width and by Tk's extra pixel of fudge.
func (p *PolygonItem) updateBBox() {
	if len(p.coords) < 2 {
		return
	}
	p.X1, p.X2 = int(p.coords[0]), int(p.coords[0])
	p.Y1, p.Y2 = int(p.coords[1]), int(p.coords[1])
	for i := 2; i+1 < len(p.coords); i += 2 {
		x, y := int(p.coords[i]+0.5), int(p.coords[i+1]+0.5)
		p.X1, p.X2 = min(p.X1, x), max(p.X2, x)
		p.Y1, p.Y2 = min(p.Y1, y), max(p.Y2, y)
	}
	if p.outline != nil && p.outlineWidth > 0 {
		grow := int((float64(p.outlineWidth) + 1.5) / 2)
		p.X1 -= grow
		p.X2 += grow
		p.Y1 -= grow
		p.Y2 += grow
	}
	p.X1--
	p.X2++
	p.Y1--
	p.Y2++
}

func (p *PolygonItem) Display(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	clipX, clipY, clipW, clipH, originX, originY int) {

	if len(p.coords) < 6 {
		return
	}

	c := p.canvas
	displayCoords := p.coords
	if p.smooth && len(p.coords) >= 6 {
		displayCoords = appendBezierSpline(c.coordBuf[:0], p.coords, true, p.splineSteps)
		c.coordBuf = displayCoords
	}

	// One extra point closes the outline below.
	n := len(displayCoords) / 2
	points := c.scratchPoints(n + 1)[:n]
	for i := range points {
		points[i] = drawablePoint(displayCoords[2*i], displayCoords[2*i+1], originX, originY)
	}

	if fill := p.fillFor(p.fill); fill != nil && len(points) >= 3 {
		d.SetForeground(gc, fill.Pixel)
		stippled := c.stippleOn(d, drawable, gc, p.stipple, originX, originY)
		d.FillPolygon(drawable, gc, points, platform.PolygonComplex, platform.CoordModeOrigin)
		if stippled {
			stippleOff(d, gc)
		}
	}

	if p.outline != nil && p.outlineWidth > 0 && len(points) >= 2 {
		d.SetForeground(gc, p.outline.Pixel)
		lineStyle := platform.LineSolid
		if len(p.dash) > 0 {
			lineStyle = platform.LineOnOffDash
			d.SetDashes(gc, 0, p.dash)
		}
		// ConfigurePolygon's outline GC: CapRound and -joinstyle (round).
		d.SetLineAttributes(gc, uint(p.outlineWidth), lineStyle, platform.CapRound, platform.JoinRound)
		// Close the polygon by appending the first point.
		closed := points[:n+1]
		closed[n] = points[0]
		d.DrawLines(drawable, gc, closed, platform.CoordModeOrigin)
		d.SetLineAttributes(gc, 1, platform.LineSolid, platform.CapButt, platform.JoinMiter)
	}
}

func (p *PolygonItem) PointDistance(x, y float64) float64 {
	if len(p.coords) < 6 {
		return math.MaxFloat64
	}
	// Check if point is inside polygon using ray casting.
	if pointInPolygon(x, y, p.coords) {
		return 0
	}
	// Otherwise return distance to nearest edge.
	minSq := math.MaxFloat64
	n := len(p.coords) / 2
	px, py := p.coords[2*n-2], p.coords[2*n-1]
	for i := range n {
		qx, qy := p.coords[2*i], p.coords[2*i+1]
		if d := segmentPointDistanceSq(x, y, px, py, qx, qy); d < minSq {
			minSq = d
		}
		px, py = qx, qy
	}
	return math.Sqrt(minSq)
}

func (p *PolygonItem) AreaOverlap(ax1, ay1, ax2, ay2 float64) int {
	bx1, by1, bx2, by2 := float64(p.X1), float64(p.Y1), float64(p.X2), float64(p.Y2)
	if ax2 < bx1 || ax1 > bx2 || ay2 < by1 || ay1 > by2 {
		return -1
	}
	if ax1 <= bx1 && ax2 >= bx2 && ay1 <= by1 && ay2 >= by2 {
		return 1
	}
	return 0
}

func (p *PolygonItem) Scale(ox, oy, sx, sy float64) {
	for i := 0; i < len(p.coords)-1; i += 2 {
		p.coords[i] = ox + (p.coords[i]-ox)*sx
		p.coords[i+1] = oy + (p.coords[i+1]-oy)*sy
	}
	p.updateBBox()
}

func (p *PolygonItem) Translate(dx, dy float64) {
	for i := 0; i < len(p.coords)-1; i += 2 {
		p.coords[i] += dx
		p.coords[i+1] += dy
	}
	p.updateBBox()
}

func (p *PolygonItem) Delete(d platform.DisplayServer) {}

// Postscript emits a PostScript representation of the polygon.
//
// Mirrors tk/generic/tkCanvPoly.c:PolygonToPostscript. Smooth polygons are
// sampled into polyline approximations.
func (p *PolygonItem) Postscript(ps *PSContext) error {
	if p.State() == ItemStateHidden || len(p.coords) < 4 {
		return nil
	}
	if p.smooth {
		steps := 12
		pts := generateBezierSpline(p.coords, true, steps)
		ps.Path(pts)
	} else {
		ps.Path(p.coords)
	}
	if p.fill != nil {
		ps.Color(p.fill)
		ps.write("gsave fill grestore newpath\n")
		if p.smooth {
			ps.Path(generateBezierSpline(p.coords, true, 12))
		} else {
			ps.Path(p.coords)
		}
	}
	if p.outline != nil && p.outlineWidth > 0 {
		ps.Outline(p.outlineWidth, dashInts(p.dash), 0, p.outline, nil)
	}
	ps.write("newpath\n")
	return nil
}

// pointInPolygon tests if (px,py) is inside the polygon using ray casting.
func pointInPolygon(px, py float64, coords []float64) bool {
	n := len(coords) / 2
	inside := false
	j := n - 1
	for i := range n {
		xi, yi := coords[i*2], coords[i*2+1]
		xj, yj := coords[j*2], coords[j*2+1]
		if ((yi > py) != (yj > py)) &&
			(px < (xj-xi)*(py-yi)/(yj-yi)+xi) {
			inside = !inside
		}
		j = i
	}
	return inside
}
