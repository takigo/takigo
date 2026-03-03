package canvas

import (
	"math"

	"github.com/msorc/takigo/platform"
)

// PolygonItem implements a filled polygon canvas item.
type PolygonItem struct {
	ItemBase
	coords       []float64
	fill         *colorRef
	outline      *colorRef
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
	// Default: filled black, no outline.
	item.fill = &colorRef{Pixel: 0x000000}
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

func (p *PolygonItem) updateBBox() {
	if len(p.coords) < 4 {
		return
	}
	hw := float64(p.outlineWidth)/2.0 + 1
	minX, minY := p.coords[0], p.coords[1]
	maxX, maxY := minX, minY
	for i := 2; i < len(p.coords)-1; i += 2 {
		x, y := p.coords[i], p.coords[i+1]
		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
	}
	p.X1 = int(math.Floor(minX - hw))
	p.Y1 = int(math.Floor(minY - hw))
	p.X2 = int(math.Ceil(maxX + hw))
	p.Y2 = int(math.Ceil(maxY + hw))
}

func (p *PolygonItem) Display(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	clipX, clipY, clipW, clipH, originX, originY int) {

	if len(p.coords) < 6 {
		return
	}

	displayCoords := p.coords
	if p.smooth && len(p.coords) >= 6 {
		displayCoords = generateBezierSpline(p.coords, true, p.splineSteps)
	}

	points := make([]platform.Point, len(displayCoords)/2)
	for i := 0; i < len(displayCoords)-1; i += 2 {
		points[i/2] = platform.Point{
			X: int16(displayCoords[i]) - int16(originX),
			Y: int16(displayCoords[i+1]) - int16(originY),
		}
	}

	if p.fill != nil && len(points) >= 3 {
		d.SetForeground(gc, p.fill.Pixel)
		d.FillPolygon(drawable, gc, points, platform.PolygonComplex, platform.CoordModeOrigin)
	}

	if p.outline != nil && p.outlineWidth > 0 && len(points) >= 2 {
		d.SetForeground(gc, p.outline.Pixel)
		lineStyle := platform.LineSolid
		if len(p.dash) > 0 {
			lineStyle = platform.LineOnOffDash
			d.SetDashes(gc, 0, p.dash)
		}
		d.SetLineAttributes(gc, uint(p.outlineWidth), lineStyle, platform.CapButt, platform.JoinRound)
		// Close the polygon by appending the first point.
		closed := make([]platform.Point, len(points)+1)
		copy(closed, points)
		closed[len(points)] = points[0]
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
	minDist := math.MaxFloat64
	n := len(p.coords) / 2
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		d := segmentPointDistance(x, y,
			p.coords[i*2], p.coords[i*2+1],
			p.coords[j*2], p.coords[j*2+1])
		if d < minDist {
			minDist = d
		}
	}
	return minDist
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

// pointInPolygon tests if (px,py) is inside the polygon using ray casting.
func pointInPolygon(px, py float64, coords []float64) bool {
	n := len(coords) / 2
	inside := false
	j := n - 1
	for i := 0; i < n; i++ {
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
