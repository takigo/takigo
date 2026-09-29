package canvas

import "slices"

// generateBezierSpline ports TkMakeBezierCurve (tk/generic/tkTrig.c) for
// alternating x,y coords. A curve whose first and last points coincide is
// treated as closed; closed forces that by appending the first point, as
// tkCanvPoly.c does for smoothed polygons. Open curves with fewer than three
// points are returned unchanged, since Tk's line item draws them straight.
func generateBezierSpline(coords []float64, closed bool, steps int) []float64 {
	return appendBezierSpline(nil, coords, closed, steps)
}

// appendBezierSpline appends the spline through coords to dst (see
// generateBezierSpline) without allocating beyond dst's growth.
func appendBezierSpline(dst, coords []float64, closed bool, steps int) []float64 {
	n := len(coords) / 2
	if steps <= 0 {
		steps = 12
	}
	if n < 2 || (n < 3 && !closed) {
		return append(dst, coords...)
	}
	// wrap: the closing point Tk appends is read from coords[0:2] instead
	// of copying coords.
	wrap := closed && (coords[0] != coords[2*n-2] || coords[1] != coords[2*n-1])
	if wrap {
		n++
	}
	pt := func(i int) (float64, float64) {
		if wrap && i == n-1 {
			return coords[0], coords[1]
		}
		return coords[2*i], coords[2*i+1]
	}
	dst = slices.Grow(dst, 2+2*steps*n)

	var ctl [8]float64
	x0, y0 := pt(0)
	xl, yl := pt(n - 1)
	isClosed := x0 == xl && y0 == yl
	if isClosed {
		xp, yp := pt(n - 2)
		x1, y1 := pt(1)
		ctl = [8]float64{
			0.5*xp + 0.5*x0, 0.5*yp + 0.5*y0,
			0.167*xp + 0.833*x0, 0.167*yp + 0.833*y0,
			0.833*x0 + 0.167*x1, 0.833*y0 + 0.167*y1,
			0.5*x0 + 0.5*x1, 0.5*y0 + 0.5*y1,
		}
		dst = append(dst, ctl[0], ctl[1])
		dst = bezierPoints(dst, ctl, steps)
	} else {
		dst = append(dst, x0, y0)
	}
	for i := 2; i < n; i++ {
		ax, ay := pt(i - 2)
		bx, by := pt(i - 1)
		cx, cy := pt(i)
		if i == 2 && !isClosed {
			ctl[0], ctl[1] = ax, ay
			ctl[2], ctl[3] = 0.333*ax+0.667*bx, 0.333*ay+0.667*by
		} else {
			ctl[0], ctl[1] = 0.5*ax+0.5*bx, 0.5*ay+0.5*by
			ctl[2], ctl[3] = 0.167*ax+0.833*bx, 0.167*ay+0.833*by
		}
		if i == n-1 && !isClosed {
			ctl[4], ctl[5] = .667*bx+.333*cx, .667*by+.333*cy
			ctl[6], ctl[7] = cx, cy
		} else {
			ctl[4], ctl[5] = .833*bx+.167*cx, .833*by+.167*cy
			ctl[6], ctl[7] = 0.5*bx+0.5*cx, 0.5*by+0.5*cy
		}
		if (ax == bx && ay == by) || (bx == cx && by == cy) {
			dst = append(dst, ctl[6], ctl[7])
			continue
		}
		dst = bezierPoints(dst, ctl, steps)
	}
	return dst
}

// bezierPoints ports TkBezierPoints: steps points along the cubic Bezier
// with the four control points in ctl (t = 1/steps .. 1).
func bezierPoints(out []float64, ctl [8]float64, steps int) []float64 {
	for i := 1; i <= steps; i++ {
		t := float64(i) / float64(steps)
		t2, t3 := t*t, t*t*t
		u := 1 - t
		u2, u3 := u*u, u*u*u
		out = append(out,
			ctl[0]*u3+3*(ctl[2]*t*u2+ctl[4]*t2*u)+ctl[6]*t3,
			ctl[1]*u3+3*(ctl[3]*t*u2+ctl[5]*t2*u)+ctl[7]*t3)
	}
	return out
}
