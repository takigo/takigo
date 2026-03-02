package canvas

// generateBezierSpline generates a smooth Bezier spline through the given
// control points. Port of Tk's TkMakeBezierCurve.
//
// coords: alternating x,y values (length must be even, >= 4)
// closed: if true, the curve wraps around
// steps: number of intermediate points per segment (default 12)
//
// Returns a new slice of alternating x,y values.
func generateBezierSpline(coords []float64, closed bool, steps int) []float64 {
	n := len(coords) / 2
	if n < 2 {
		return coords
	}
	if steps <= 0 {
		steps = 12
	}

	// For open curves with < 3 points, just return the input.
	if n < 3 && !closed {
		out := make([]float64, len(coords))
		copy(out, coords)
		return out
	}

	var result []float64

	if closed {
		result = generateClosedSpline(coords, n, steps)
	} else {
		result = generateOpenSpline(coords, n, steps)
	}

	return result
}

func generateOpenSpline(coords []float64, n, steps int) []float64 {
	var result []float64

	// First point.
	result = append(result, coords[0], coords[1])

	for i := 0; i <= n-3; i++ {
		// Control points for this segment.
		// p0 = midpoint of segment i to i+1 (or point 0 for first segment)
		// p1 = point i+1
		// p2 = midpoint of segment i+1 to i+2 (or point n-1 for last segment)
		var p0x, p0y, p1x, p1y, p2x, p2y float64

		if i == 0 {
			p0x = coords[0]
			p0y = coords[1]
		} else {
			p0x = (coords[i*2] + coords[(i+1)*2]) / 2
			p0y = (coords[i*2+1] + coords[(i+1)*2+1]) / 2
		}

		p1x = coords[(i+1)*2]
		p1y = coords[(i+1)*2+1]

		if i == n-3 {
			p2x = coords[(n-1)*2]
			p2y = coords[(n-1)*2+1]
		} else {
			p2x = (coords[(i+1)*2] + coords[(i+2)*2]) / 2
			p2y = (coords[(i+1)*2+1] + coords[(i+2)*2+1]) / 2
		}

		// Generate Bezier curve points for this quadratic segment.
		for s := 1; s <= steps; s++ {
			t := float64(s) / float64(steps)
			t2 := t * t
			u := 1 - t
			u2 := u * u

			x := u2*p0x + 2*u*t*p1x + t2*p2x
			y := u2*p0y + 2*u*t*p1y + t2*p2y
			result = append(result, x, y)
		}
	}

	return result
}

func generateClosedSpline(coords []float64, n, steps int) []float64 {
	var result []float64

	for i := 0; i < n; i++ {
		i0 := i
		i1 := (i + 1) % n
		i2 := (i + 2) % n

		p0x := (coords[i0*2] + coords[i1*2]) / 2
		p0y := (coords[i0*2+1] + coords[i1*2+1]) / 2
		p1x := coords[i1*2]
		p1y := coords[i1*2+1]
		p2x := (coords[i1*2] + coords[i2*2]) / 2
		p2y := (coords[i1*2+1] + coords[i2*2+1]) / 2

		if i == 0 {
			result = append(result, p0x, p0y)
		}

		for s := 1; s <= steps; s++ {
			t := float64(s) / float64(steps)
			t2 := t * t
			u := 1 - t
			u2 := u * u

			x := u2*p0x + 2*u*t*p1x + t2*p2x
			y := u2*p0y + 2*u*t*p1y + t2*p2y
			result = append(result, x, y)
		}
	}

	return result
}
