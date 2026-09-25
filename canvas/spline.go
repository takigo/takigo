package canvas

// generateBezierSpline ports TkMakeBezierCurve (tk/generic/tkTrig.c) for
// alternating x,y coords. A curve whose first and last points coincide is
// treated as closed; closed forces that by appending the first point, as
// tkCanvPoly.c does for smoothed polygons. Open curves with fewer than three
// points are returned unchanged, since Tk's line item draws them straight.
func generateBezierSpline(coords []float64, closed bool, steps int) []float64 {
	n := len(coords) / 2
	if steps <= 0 {
		steps = 12
	}
	if n < 2 || (n < 3 && !closed) {
		return append([]float64(nil), coords...)
	}
	pts := coords
	if closed && (coords[0] != coords[2*n-2] || coords[1] != coords[2*n-1]) {
		pts = append(append([]float64(nil), coords...), coords[0], coords[1])
		n++
	}

	var out []float64
	var ctl [8]float64
	nc := 2 * n
	isClosed := pts[0] == pts[nc-2] && pts[1] == pts[nc-1]
	if isClosed {
		ctl = [8]float64{
			0.5*pts[nc-4] + 0.5*pts[0], 0.5*pts[nc-3] + 0.5*pts[1],
			0.167*pts[nc-4] + 0.833*pts[0], 0.167*pts[nc-3] + 0.833*pts[1],
			0.833*pts[0] + 0.167*pts[2], 0.833*pts[1] + 0.167*pts[3],
			0.5*pts[0] + 0.5*pts[2], 0.5*pts[1] + 0.5*pts[3],
		}
		out = append(out, ctl[0], ctl[1])
		out = bezierPoints(out, ctl, steps)
	} else {
		out = append(out, pts[0], pts[1])
	}
	for i := 2; i < n; i++ {
		p := pts[2*(i-2):]
		if i == 2 && !isClosed {
			ctl[0], ctl[1] = p[0], p[1]
			ctl[2], ctl[3] = 0.333*p[0]+0.667*p[2], 0.333*p[1]+0.667*p[3]
		} else {
			ctl[0], ctl[1] = 0.5*p[0]+0.5*p[2], 0.5*p[1]+0.5*p[3]
			ctl[2], ctl[3] = 0.167*p[0]+0.833*p[2], 0.167*p[1]+0.833*p[3]
		}
		if i == n-1 && !isClosed {
			ctl[4], ctl[5] = .667*p[2]+.333*p[4], .667*p[3]+.333*p[5]
			ctl[6], ctl[7] = p[4], p[5]
		} else {
			ctl[4], ctl[5] = .833*p[2]+.167*p[4], .833*p[3]+.167*p[5]
			ctl[6], ctl[7] = 0.5*p[2]+0.5*p[4], 0.5*p[3]+0.5*p[5]
		}
		if (p[0] == p[2] && p[1] == p[3]) || (p[2] == p[4] && p[3] == p[5]) {
			out = append(out, ctl[6], ctl[7])
			continue
		}
		out = bezierPoints(out, ctl, steps)
	}
	return out
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
