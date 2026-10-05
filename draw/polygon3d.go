package draw

import (
	"math"

	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/platform"
)

// X11 XFillPolygon shape/mode values.
const (
	shapeComplex    = 0
	shapeConvex     = 2
	coordModeOrigin = 0
)

// Fill3DPolygon ports Tk_Fill3DPolygon (tk/generic/tk3d.c): fill the polygon
// with the border's background, then draw a 3D border along its edges.
func Fill3DPolygon(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	border *Border, pts []platform.Point, borderWidth int, relief option.Relief) {
	d.SetForeground(gc, border.BgPixel)
	d.FillPolygon(drawable, gc, pts, shapeComplex, coordModeOrigin)
	if relief != option.ReliefFlat {
		Draw3DPolygon(d, drawable, gc, border, pts, borderWidth, relief)
	}
}

// Draw3DPolygon ports Tk_Draw3DPolygon: the border lies to the left of the
// path (looking along it) and each edge is light or dark depending on its
// direction, so a clockwise raised polygon is lit from the top-left.
func Draw3DPolygon(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	border *Border, pts []platform.Point, borderWidth int, relief option.Relief) {
	if relief == option.ReliefGroove || relief == option.ReliefRidge {
		half := borderWidth / 2
		first, second := option.ReliefSunken, option.ReliefRaised
		if relief == option.ReliefGroove {
			first, second = second, first
		}
		Draw3DPolygon(d, drawable, gc, border, pts, half, first)
		Draw3DPolygon(d, drawable, gc, border, pts, -half, second)
		return
	}
	n := len(pts)
	if n > 1 && pts[n-1] == pts[0] {
		n--
	}
	if n < 2 {
		return
	}

	var poly [4]point
	var b1, b2, c point
	pointsSeen := 0
	for i := 0; i < n+2; i++ {
		// Edges p1->p2 starting with (n-2, n-1), then (n-1, 0), (0, 1), ...
		p1 := toPoint(pts[(i+n-2)%n])
		p2 := toPoint(pts[(i+n-1)%n])
		if p1 == p2 {
			continue
		}
		newB1 := shiftLine(p1, p2, borderWidth)
		newB2 := point{newB1.x + p2.x - p1.x, newB1.y + p2.y - p1.y}
		poly[3] = p1
		parallel := false
		if pointsSeen >= 1 {
			var ok bool
			poly[2], ok = intersect(newB1, newB2, b1, b2)
			parallel = !ok
			if parallel {
				perp := point{p1.x + (p2.y - p1.y), p1.y - (p2.x - p1.x)}
				poly[2], _ = intersect(p1, perp, b1, b2)
				c, _ = intersect(p1, perp, newB1, newB2)
				shift1 := shiftLine(p1, perp, borderWidth)
				shift2 := point{shift1.x + perp.x - p1.x, shift1.y + perp.y - p1.y}
				poly[3], _ = intersect(p1, p2, shift1, shift2)
			}
		}
		if pointsSeen >= 2 {
			dx, dy := poly[3].x-poly[0].x, poly[3].y-poly[0].y
			var lightOnLeft bool
			if dx > 0 {
				lightOnLeft = dy <= dx
			} else {
				lightOnLeft = dy < dx
			}
			pixel := border.DarkPixel
			if lightOnLeft != (relief == option.ReliefRaised) {
				pixel = border.LightPixel
			}
			d.SetForeground(gc, pixel)
			xp := make([]platform.Point, 4)
			for k, p := range poly {
				xp[k] = platform.Point{X: int16(p.x), Y: int16(p.y)}
			}
			d.FillPolygon(drawable, gc, xp, shapeConvex, coordModeOrigin)
		}
		b1, b2 = newB1, newB2
		poly[0] = poly[3]
		if parallel {
			poly[1] = c
		} else if pointsSeen >= 1 {
			poly[1] = poly[2]
		}
		pointsSeen++
	}
}

type point struct{ x, y int }

func toPoint(p platform.Point) point { return point{int(p.X), int(p.Y)} }

var shiftTable = func() (t [129]int) {
	for i := range t {
		t[i] = int(128/math.Cos(math.Atan(float64(i)/128)) + .5)
	}
	return t
}()

// shiftLine ports ShiftLine: a point on the line parallel to p1->p2, distance
// units to its left, using Tk's integer slope table.
func shiftLine(p1, p2 point, distance int) point {
	p3 := p1
	dx, dy := p2.x-p1.x, p2.y-p1.y
	dyNeg, dxNeg := dy < 0, dx < 0
	if dyNeg {
		dy = -dy
	}
	if dxNeg {
		dx = -dx
	}
	if dy <= dx {
		dy = (distance*shiftTable[(dy<<7)/dx] + 64) >> 7
		if !dxNeg {
			dy = -dy
		}
		p3.y += dy
	} else {
		dx = (distance*shiftTable[(dx<<7)/dy] + 64) >> 7
		if dyNeg {
			dx = -dx
		}
		p3.x += dx
	}
	return p3
}

// intersect ports Intersect: the rounded intersection of lines a1a2 and
// b1b2; ok is false when they are parallel.
func intersect(a1, a2, b1, b2 point) (point, bool) {
	dxadyb := (a2.x - a1.x) * (b2.y - b1.y)
	dxbdya := (b2.x - b1.x) * (a2.y - a1.y)
	dxadxb := (a2.x - a1.x) * (b2.x - b1.x)
	dyadyb := (a2.y - a1.y) * (b2.y - b1.y)
	if dxadyb == dxbdya {
		return point{}, false
	}
	var r point
	p := a1.x*dxbdya - b1.x*dxadyb + (b1.y-a1.y)*dxadxb
	q := dxbdya - dxadyb
	r.x = roundDiv(p, q)
	p = a1.y*dxadyb - b1.y*dxbdya + (b1.x-a1.x)*dyadyb
	q = dxadyb - dxbdya
	r.y = roundDiv(p, q)
	return r, true
}

func roundDiv(p, q int) int {
	if q < 0 {
		p, q = -p, -q
	}
	if p < 0 {
		return -((-p + q/2) / q)
	}
	return (p + q/2) / q
}
