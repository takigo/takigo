package canvas

import (
	"math"

	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/platform"
)

// LineItem implements a polyline canvas item with optional arrows and smoothing.
type LineItem struct {
	ItemBase
	coords      []float64
	color       *color.ColorRef
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
	item.color = &color.ColorRef{Pixel: 0x000000}
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

	// Get display coords (copy so we can shorten endpoints for arrowheads).
	displayCoords := append([]float64{}, l.coords...)
	if l.smooth && len(l.coords) >= 6 {
		displayCoords = generateBezierSpline(l.coords, false, l.splineSteps)
	}

	// Shorten line endpoints so the thick shaft meets the arrowhead polygon
	// edge seamlessly. Tk computes a backup distance by interpolating between
	// arrowShapeA and arrowShapeB based on where the line edge intersects
	// the arrow polygon edge.
	backup := l.arrowBackup()
	if (l.arrow == ArrowFirst || l.arrow == ArrowBoth) && len(displayCoords) >= 4 {
		dx0 := displayCoords[2] - displayCoords[0]
		dy0 := displayCoords[3] - displayCoords[1]
		seg := math.Sqrt(dx0*dx0 + dy0*dy0)
		if seg > 0.001 {
			displayCoords[0] += (dx0 / seg) * backup
			displayCoords[1] += (dy0 / seg) * backup
		}
	}
	if (l.arrow == ArrowLast || l.arrow == ArrowBoth) && len(displayCoords) >= 4 {
		n := len(displayCoords)
		dx0 := displayCoords[n-2] - displayCoords[n-4]
		dy0 := displayCoords[n-1] - displayCoords[n-3]
		seg := math.Sqrt(dx0*dx0 + dy0*dy0)
		if seg > 0.001 {
			displayCoords[n-2] -= (dx0 / seg) * backup
			displayCoords[n-1] -= (dy0 / seg) * backup
		}
	}

	points := make([]platform.Point, len(displayCoords)/2)
	for i := 0; i < len(displayCoords)-1; i += 2 {
		points[i/2] = platform.Point{
			X: int16(drawableCoord(displayCoords[i], originX)),
			Y: int16(drawableCoord(displayCoords[i+1], originY)),
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

// arrowBackup returns the distance to shorten the line at an arrowhead
// endpoint. It interpolates between arrowShapeA and arrowShapeB so the
// thick line edge meets the arrow polygon edge exactly (matching Tk's
// ConfigureArrows computation).
func (l *LineItem) arrowBackup() float64 {
	shapeC := l.arrowShapeC + float64(l.width)/2
	if shapeC < 0.001 {
		return l.arrowShapeA
	}
	fracHeight := (float64(l.width) / 2) / shapeC
	return fracHeight*l.arrowShapeB + (1-fracHeight)*l.arrowShapeA
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

	b := l.arrowShapeB                      // wing distance (further from tip)
	c := l.arrowShapeC + float64(l.width)/2 // halfwidth + half line width (matches Tk)
	backup := l.arrowBackup()               // shaft-edge junction distance
	hw := float64(l.width) / 2              // half line width

	// Tk-style 6-point arrowhead: tip → left_wing → left_shaft_edge →
	// right_shaft_edge → right_wing → tip. The shaft edge points at
	// the backup distance ensure seamless connection with the thick line.
	arrowPoints := []platform.Point{
		drawablePoint(tipX, tipY, originX, originY),
		drawablePoint(tipX-ux*b+px*c, tipY-uy*b+py*c, originX, originY),
		drawablePoint(tipX-ux*backup+px*hw, tipY-uy*backup+py*hw, originX, originY),
		drawablePoint(tipX-ux*backup-px*hw, tipY-uy*backup-py*hw, originX, originY),
		drawablePoint(tipX-ux*b-px*c, tipY-uy*b-py*c, originX, originY),
		drawablePoint(tipX, tipY, originX, originY),
	}

	d.FillPolygon(drawable, gc, arrowPoints, platform.PolygonNonconvex, platform.CoordModeOrigin)
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

// Postscript emits a PostScript representation of the line.
//
// Mirrors tk/generic/tkCanvLine.c:LineToPostscript + ArrowheadPostscript.
// Single-point lines become filled circles; multi-point lines draw a path
// then stroke; arrowheads emit closed polygons at each end.
func (l *LineItem) Postscript(ps *PSContext) error {
	if l.State() == ItemStateHidden {
		return nil
	}
	npts := len(l.coords) / 2
	if npts == 0 {
		return nil
	}
	if npts == 1 {
		// Single point → filled disc.
		ps.writef("matrix currentmatrix\n%.15g %.15g translate %.15g %.15g scale 1 0 moveto 0 0 1 0 360 arc\nsetmatrix\n",
			l.coords[0], ps.PsY(int(l.coords[1])), float64(l.width)/2, float64(l.width)/2)
		if l.color != nil {
			ps.Color(l.color)
			ps.write("fill\n")
		}
		return nil
	}

	// Decide whether to use spline path (smooth) or polyline.
	if l.smooth && len(l.coords) >= 6 {
		// Linear approximation; the prolog has a real bezier curve helper but
		// we sample for portability. SplineSteps controls sampling density.
		steps := l.splineSteps
		if steps < 1 {
			steps = 12
		}
		pts := sampleSpline(l.coords, steps)
		ps.Path(pts)
	} else {
		ps.Path(l.coords)
	}
	ps.Outline(l.width, dashInts(l.dash), 0, l.color, nil)
	ps.write("newpath\n")

	// Arrowheads.
	if l.arrow == ArrowFirst || l.arrow == ArrowBoth {
		psArrow(l, ps, true /*first*/)
	}
	if l.arrow == ArrowLast || l.arrow == ArrowBoth {
		psArrow(l, ps, false /*last*/)
	}
	return nil
}

// sampleSpline returns a sampled polyline approximating the open Bezier
// spline through coords with `steps` segments between each control pair.
func sampleSpline(coords []float64, steps int) []float64 {
	return generateBezierSpline(coords, false, steps)
}

// psArrow emits an arrowhead polygon. Mirrors ArrowheadPostscript.
func psArrow(l *LineItem, ps *PSContext, first bool) {
	n := len(l.coords) / 2
	if n < 2 {
		return
	}
	var tipX, tipY, sideX, sideY float64
	if first {
		tipX, tipY = l.coords[0], l.coords[1]
		sideX, sideY = l.coords[2], l.coords[3]
	} else {
		tipX, tipY = l.coords[n*2-2], l.coords[n*2-1]
		sideX, sideY = l.coords[n*2-4], l.coords[n*2-3]
	}
	dx := sideX - tipX
	dy := sideY - tipY
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist == 0 {
		return
	}
	ux, uy := dx/dist, dy/dist
	px, py := -uy, ux

	shapeC := l.arrowShapeC + float64(l.width)/2
	shapeA := l.arrowBackup()
	shapeB := l.arrowShapeB

	p1x, p1y := tipX, tipY
	p2x, p2y := tipX+shapeB*ux+shapeC*px, tipY+shapeB*uy+shapeC*py
	p3x, p3y := tipX+shapeA*ux+(float64(l.width)/2)*px, tipY+shapeA*uy+(float64(l.width)/2)*py
	p4x, p4y := tipX+shapeA*ux-(float64(l.width)/2)*px, tipY+shapeA*uy-(float64(l.width)/2)*py
	p5x, p5y := tipX+shapeB*ux-shapeC*px, tipY+shapeB*uy-shapeC*py

	ps.Path([]float64{p1x, p1y, p2x, p2y, p3x, p3y, p4x, p4y, p5x, p5y})
	if l.color != nil {
		ps.Color(l.color)
		ps.write("gsave fill grestore newpath\n")
	}
}

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
