package canvas

import (
	"math"

	"github.com/msorc/takigo/platform"
)

// LineItem implements a polyline canvas item with optional arrows and smoothing.
type LineItem struct {
	ItemBase
	coords      []float64
	color       *colorRef
	width       int
	dash        []byte
	capStyle    int
	joinStyle   int
	arrow       ArrowMode
	arrowShapeA float64 // distance along shaft
	arrowShapeB float64 // distance along side
	arrowShapeC float64 // half-width of base
	smooth      bool
	splineSteps int
}

func newLineItem(coords []float64, c *Canvas) *LineItem {
	item := &LineItem{
		coords:      append([]float64{}, coords...),
		width:       1,
		capStyle:    platform.CapButt,
		joinStyle:   platform.JoinRound,
		arrowShapeA: 8,
		arrowShapeB: 10,
		arrowShapeC: 3,
		splineSteps: 12,
	}
	// Default color is black.
	item.color = &colorRef{Pixel: 0x000000}
	item.ItemBase.canvas = c
	item.updateBBox()
	return item
}

func (l *LineItem) base() *ItemBase { return &l.ItemBase }
func (l *LineItem) Type() string    { return "line" }

func (l *LineItem) BBox() (x1, y1, x2, y2 int) {
	return l.X1, l.Y1, l.X2, l.Y2
}

func (l *LineItem) Coords() []float64 {
	out := make([]float64, len(l.coords))
	copy(out, l.coords)
	return out
}

func (l *LineItem) SetCoords(coords []float64) error {
	l.coords = append(l.coords[:0], coords...)
	l.updateBBox()
	return nil
}

func (l *LineItem) Configure(opts []ItemOption) error {
	c := l.canvas
	for _, opt := range opts {
		if err := opt(c, l); err != nil {
			return err
		}
	}
	l.updateBBox()
	return nil
}

func (l *LineItem) updateBBox() {
	if len(l.coords) < 2 {
		return
	}
	hw := float64(l.width)/2.0 + 1
	minX, minY := l.coords[0], l.coords[1]
	maxX, maxY := minX, minY
	for i := 2; i < len(l.coords)-1; i += 2 {
		x, y := l.coords[i], l.coords[i+1]
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
	// Expand for arrow heads.
	if l.arrow != ArrowNone {
		hw += l.arrowShapeB
	}
	l.X1 = int(math.Floor(minX - hw))
	l.Y1 = int(math.Floor(minY - hw))
	l.X2 = int(math.Ceil(maxX + hw))
	l.Y2 = int(math.Ceil(maxY + hw))
}

func (l *LineItem) Display(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	clipX, clipY, clipW, clipH, originX, originY int) {

	if len(l.coords) < 4 || l.color == nil {
		return
	}

	d.SetForeground(gc, l.color.Pixel)
	lineStyle := platform.LineSolid
	if len(l.dash) > 0 {
		lineStyle = platform.LineOnOffDash
		d.SetDashes(gc, 0, l.dash)
	}
	d.SetLineAttributes(gc, uint(l.width), lineStyle, l.capStyle, l.joinStyle)

	// Get display coords.
	displayCoords := l.coords
	if l.smooth && len(l.coords) >= 6 {
		displayCoords = generateBezierSpline(l.coords, false, l.splineSteps)
	}

	points := make([]platform.Point, len(displayCoords)/2)
	for i := 0; i < len(displayCoords)-1; i += 2 {
		points[i/2] = platform.Point{
			X: int16(displayCoords[i]) - int16(originX),
			Y: int16(displayCoords[i+1]) - int16(originY),
		}
	}

	if len(points) >= 2 {
		d.DrawLines(drawable, gc, points, platform.CoordModeOrigin)
	}

	// Draw arrows.
	if l.arrow == ArrowFirst || l.arrow == ArrowBoth {
		l.drawArrow(d, drawable, gc, originX, originY, true)
	}
	if l.arrow == ArrowLast || l.arrow == ArrowBoth {
		l.drawArrow(d, drawable, gc, originX, originY, false)
	}

	// Reset line attributes.
	d.SetLineAttributes(gc, 1, platform.LineSolid, platform.CapButt, platform.JoinMiter)
}

// drawArrow draws an arrowhead at one end of the line.
func (l *LineItem) drawArrow(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	originX, originY int, first bool) {

	n := len(l.coords)
	if n < 4 {
		return
	}

	var tipX, tipY, baseX, baseY float64
	if first {
		tipX, tipY = l.coords[0], l.coords[1]
		baseX, baseY = l.coords[2], l.coords[3]
	} else {
		tipX, tipY = l.coords[n-2], l.coords[n-1]
		baseX, baseY = l.coords[n-4], l.coords[n-3]
	}

	dx := tipX - baseX
	dy := tipY - baseY
	length := math.Sqrt(dx*dx + dy*dy)
	if length < 0.001 {
		return
	}

	// Unit vector along the line toward the tip.
	ux := dx / length
	uy := dy / length
	// Perpendicular.
	px := -uy
	py := ux

	a := l.arrowShapeA
	c := l.arrowShapeC

	// Arrow polygon: tip, left, right (3 points).
	arrowPoints := []platform.Point{
		{X: int16(tipX) - int16(originX), Y: int16(tipY) - int16(originY)},
		{X: int16(tipX-ux*a+px*c) - int16(originX), Y: int16(tipY-uy*a+py*c) - int16(originY)},
		{X: int16(tipX-ux*a-px*c) - int16(originX), Y: int16(tipY-uy*a-py*c) - int16(originY)},
	}

	d.FillPolygon(drawable, gc, arrowPoints, platform.PolygonConvex, platform.CoordModeOrigin)
}

func (l *LineItem) PointDistance(x, y float64) float64 {
	if len(l.coords) < 4 {
		return math.MaxFloat64
	}
	minDist := math.MaxFloat64
	for i := 0; i < len(l.coords)-3; i += 2 {
		d := segmentPointDistance(x, y,
			l.coords[i], l.coords[i+1],
			l.coords[i+2], l.coords[i+3])
		if d < minDist {
			minDist = d
		}
	}
	// Subtract half the line width for hit testing.
	minDist -= float64(l.width) / 2
	if minDist < 0 {
		minDist = 0
	}
	return minDist
}

func (l *LineItem) AreaOverlap(ax1, ay1, ax2, ay2 float64) int {
	bx1, by1, bx2, by2 := float64(l.X1), float64(l.Y1), float64(l.X2), float64(l.Y2)
	if ax2 < bx1 || ax1 > bx2 || ay2 < by1 || ay1 > by2 {
		return -1
	}
	if ax1 <= bx1 && ax2 >= bx2 && ay1 <= by1 && ay2 >= by2 {
		return 1
	}
	return 0
}

func (l *LineItem) Scale(ox, oy, sx, sy float64) {
	for i := 0; i < len(l.coords)-1; i += 2 {
		l.coords[i] = ox + (l.coords[i]-ox)*sx
		l.coords[i+1] = oy + (l.coords[i+1]-oy)*sy
	}
	l.updateBBox()
}

func (l *LineItem) Translate(dx, dy float64) {
	for i := 0; i < len(l.coords)-1; i += 2 {
		l.coords[i] += dx
		l.coords[i+1] += dy
	}
	l.updateBBox()
}

func (l *LineItem) Delete(d platform.DisplayServer) {}

// segmentPointDistance computes the distance from point (px,py) to the
// line segment from (x1,y1) to (x2,y2).
func segmentPointDistance(px, py, x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	if dx == 0 && dy == 0 {
		return math.Sqrt((px-x1)*(px-x1) + (py-y1)*(py-y1))
	}
	t := ((px-x1)*dx + (py-y1)*dy) / (dx*dx + dy*dy)
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	closestX := x1 + t*dx
	closestY := y1 + t*dy
	return math.Sqrt((px-closestX)*(px-closestX) + (py-closestY)*(py-closestY))
}
