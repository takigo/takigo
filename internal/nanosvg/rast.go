package nanosvg

import (
	"math"
	"sort"
)

const (
	subsamples = 5
	fixShift   = 10
	fix        = 1 << fixShift
	fixMask    = fix - 1
	tessTol    = float32(0.25)
	distTol    = float32(0.01)
)

const (
	ptCorner = 0x01
	ptBevel  = 0x02
	ptLeft   = 0x04
)

type edge struct {
	x0, y0, x1, y1 float32
	dir            int
}

type point struct {
	x, y, dx, dy, len, dmx, dmy float32
	flags                       uint8
}

type activeEdge struct {
	x, dx int
	ey    float32
	dir   int
	next  *activeEdge
}

type rasterizer struct {
	edges    []edge
	points   []point
	scanline []uint8
	bitmap   []uint8
	width    int
	height   int
	stride   int
	// The bitmap holds the window ox, oy, bw x bh of the width x height
	// frame the edges are in (the whole frame for Rasterize).
	ox, oy, bw, bh int
}

func absf(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

func roundf(x float32) float32 {
	if x >= 0 {
		return float32(math.Floor(float64(x + 0.5)))
	}
	return float32(math.Ceil(float64(x - 0.5)))
}

func clampf(a, mn, mx float32) float32 {
	if a != a {
		return mn
	}
	if a < mn {
		return mn
	}
	if a > mx {
		return mx
	}
	return a
}

func sqrtf(x float32) float32 { return float32(math.Sqrt(float64(x))) }

func ptEquals(x1, y1, x2, y2, tol float32) bool {
	dx, dy := x2-x1, y2-y1
	return dx*dx+dy*dy < tol*tol
}

func (r *rasterizer) addPathPoint(x, y float32, flags uint8) {
	if n := len(r.points); n > 0 {
		pt := &r.points[n-1]
		if ptEquals(pt.x, pt.y, x, y, distTol) {
			pt.flags |= flags
			return
		}
	}
	r.points = append(r.points, point{x: x, y: y, flags: flags})
}

func (r *rasterizer) addEdge(x0, y0, x1, y1 float32) {
	if y0 == y1 {
		return
	}
	if y0 < y1 {
		r.edges = append(r.edges, edge{x0, y0, x1, y1, 1})
	} else {
		r.edges = append(r.edges, edge{x1, y1, x0, y0, -1})
	}
}

func normalize(x, y *float32) float32 {
	d := sqrtf(*x**x + *y**y)
	if d > 1e-6 {
		id := 1 / d
		*x *= id
		*y *= id
	}
	return d
}

func (r *rasterizer) flattenCubicBez(x1, y1, x2, y2, x3, y3, x4, y4 float32, level int, typ uint8) {
	if level > 10 {
		return
	}
	x12, y12 := (x1+x2)*0.5, (y1+y2)*0.5
	x23, y23 := (x2+x3)*0.5, (y2+y3)*0.5
	x34, y34 := (x3+x4)*0.5, (y3+y4)*0.5
	x123, y123 := (x12+x23)*0.5, (y12+y23)*0.5
	dx, dy := x4-x1, y4-y1
	d2 := absf((x2-x4)*dy - (y2-y4)*dx)
	d3 := absf((x3-x4)*dy - (y3-y4)*dx)
	if (d2+d3)*(d2+d3) < tessTol*(dx*dx+dy*dy) {
		r.addPathPoint(x4, y4, typ)
		return
	}
	x234, y234 := (x23+x34)*0.5, (y23+y34)*0.5
	x1234, y1234 := (x123+x234)*0.5, (y123+y234)*0.5
	r.flattenCubicBez(x1, y1, x12, y12, x123, y123, x1234, y1234, level+1, 0)
	r.flattenCubicBez(x1234, y1234, x234, y234, x34, y34, x4, y4, level+1, typ)
}

func (r *rasterizer) flattenShape(s *shape, scale float32) {
	for _, p := range s.paths {
		r.points = r.points[:0]
		r.addPathPoint(p.pts[0]*scale, p.pts[1]*scale, 0)
		n := len(p.pts) / 2
		for i := 0; i < n-1; i += 3 {
			q := p.pts[i*2:]
			r.flattenCubicBez(q[0]*scale, q[1]*scale, q[2]*scale, q[3]*scale,
				q[4]*scale, q[5]*scale, q[6]*scale, q[7]*scale, 0, 0)
		}
		r.addPathPoint(p.pts[0]*scale, p.pts[1]*scale, 0)
		for i, j := 0, len(r.points)-1; i < len(r.points); j, i = i, i+1 {
			r.addEdge(r.points[j].x, r.points[j].y, r.points[i].x, r.points[i].y)
		}
	}
}

func initClosed(left, right, p0, p1 *point, lineWidth float32) {
	w := lineWidth * 0.5
	dx, dy := p1.x-p0.x, p1.y-p0.y
	l := normalize(&dx, &dy)
	px, py := p0.x+dx*l*0.5, p0.y+dy*l*0.5
	dlx, dly := dy, -dx
	left.x, left.y = px-dlx*w, py-dly*w
	right.x, right.y = px+dlx*w, py+dly*w
}

func (r *rasterizer) buttCap(left, right, p *point, dx, dy, lineWidth float32, connect bool) {
	w := lineWidth * 0.5
	px, py := p.x, p.y
	dlx, dly := dy, -dx
	lx, ly := px-dlx*w, py-dly*w
	rx, ry := px+dlx*w, py+dly*w
	r.addEdge(lx, ly, rx, ry)
	if connect {
		r.addEdge(left.x, left.y, lx, ly)
		r.addEdge(rx, ry, right.x, right.y)
	}
	left.x, left.y = lx, ly
	right.x, right.y = rx, ry
}

func (r *rasterizer) squareCap(left, right, p *point, dx, dy, lineWidth float32, connect bool) {
	w := lineWidth * 0.5
	px, py := p.x-dx*w, p.y-dy*w
	dlx, dly := dy, -dx
	lx, ly := px-dlx*w, py-dly*w
	rx, ry := px+dlx*w, py+dly*w
	r.addEdge(lx, ly, rx, ry)
	if connect {
		r.addEdge(left.x, left.y, lx, ly)
		r.addEdge(rx, ry, right.x, right.y)
	}
	left.x, left.y = lx, ly
	right.x, right.y = rx, ry
}

const pi32 = float32(3.14159265358979323846264338327)

func cosf(a float32) float32 { return float32(math.Cos(float64(a))) }
func sinf(a float32) float32 { return float32(math.Sin(float64(a))) }

func (r *rasterizer) roundCap(left, right, p *point, dx, dy, lineWidth float32, ncap int, connect bool) {
	w := lineWidth * 0.5
	px, py := p.x, p.y
	dlx, dly := dy, -dx
	var lx, ly, rx, ry, prevx, prevy float32
	for i := range ncap {
		a := float32(i) / float32(ncap-1) * pi32
		ax, ay := cosf(a)*w, sinf(a)*w
		x := px - dlx*ax - dx*ay
		y := py - dly*ax - dy*ay
		if i > 0 {
			r.addEdge(prevx, prevy, x, y)
		}
		prevx, prevy = x, y
		switch i {
		case 0:
			lx, ly = x, y
		case ncap - 1:
			rx, ry = x, y
		}
	}
	if connect {
		r.addEdge(left.x, left.y, lx, ly)
		r.addEdge(rx, ry, right.x, right.y)
	}
	left.x, left.y = lx, ly
	right.x, right.y = rx, ry
}

func (r *rasterizer) bevelJoin(left, right, p0, p1 *point, lineWidth float32) {
	w := lineWidth * 0.5
	dlx0, dly0 := p0.dy, -p0.dx
	dlx1, dly1 := p1.dy, -p1.dx
	lx0, ly0 := p1.x-dlx0*w, p1.y-dly0*w
	rx0, ry0 := p1.x+dlx0*w, p1.y+dly0*w
	lx1, ly1 := p1.x-dlx1*w, p1.y-dly1*w
	rx1, ry1 := p1.x+dlx1*w, p1.y+dly1*w
	r.addEdge(lx0, ly0, left.x, left.y)
	r.addEdge(lx1, ly1, lx0, ly0)
	r.addEdge(right.x, right.y, rx0, ry0)
	r.addEdge(rx0, ry0, rx1, ry1)
	left.x, left.y = lx1, ly1
	right.x, right.y = rx1, ry1
}

func (r *rasterizer) miterJoin(left, right, p0, p1 *point, lineWidth float32) {
	w := lineWidth * 0.5
	dlx0, dly0 := p0.dy, -p0.dx
	dlx1, dly1 := p1.dy, -p1.dx
	var lx1, ly1, rx1, ry1 float32
	if p1.flags&ptLeft != 0 {
		lx1, ly1 = p1.x-p1.dmx*w, p1.y-p1.dmy*w
		r.addEdge(lx1, ly1, left.x, left.y)
		rx0, ry0 := p1.x+dlx0*w, p1.y+dly0*w
		rx1, ry1 = p1.x+dlx1*w, p1.y+dly1*w
		r.addEdge(right.x, right.y, rx0, ry0)
		r.addEdge(rx0, ry0, rx1, ry1)
	} else {
		lx0, ly0 := p1.x-dlx0*w, p1.y-dly0*w
		lx1, ly1 = p1.x-dlx1*w, p1.y-dly1*w
		r.addEdge(lx0, ly0, left.x, left.y)
		r.addEdge(lx1, ly1, lx0, ly0)
		rx1, ry1 = p1.x+p1.dmx*w, p1.y+p1.dmy*w
		r.addEdge(right.x, right.y, rx1, ry1)
	}
	left.x, left.y = lx1, ly1
	right.x, right.y = rx1, ry1
}

func (r *rasterizer) roundJoin(left, right, p0, p1 *point, lineWidth float32, ncap int) {
	w := lineWidth * 0.5
	dlx0, dly0 := p0.dy, -p0.dx
	dlx1, dly1 := p1.dy, -p1.dx
	a0 := float32(math.Atan2(float64(dly0), float64(dlx0)))
	a1 := float32(math.Atan2(float64(dly1), float64(dlx1)))
	da := a1 - a0
	if da < pi32 {
		da += pi32 * 2
	}
	if da > pi32 {
		da -= pi32 * 2
	}
	n := int(float32(math.Ceil(float64((absf(da) / pi32) * float32(ncap)))))
	n = max(n, 2)
	n = min(n, ncap)
	lx, ly := left.x, left.y
	rx, ry := right.x, right.y
	for i := 0; i < n; i++ {
		u := float32(i) / float32(n-1)
		a := a0 + u*da
		ax, ay := cosf(a)*w, sinf(a)*w
		lx1, ly1 := p1.x-ax, p1.y-ay
		rx1, ry1 := p1.x+ax, p1.y+ay
		r.addEdge(lx1, ly1, lx, ly)
		r.addEdge(rx, ry, rx1, ry1)
		lx, ly = lx1, ly1
		rx, ry = rx1, ry1
	}
	left.x, left.y = lx, ly
	right.x, right.y = rx, ry
}

func (r *rasterizer) straightJoin(left, right, p1 *point, lineWidth float32) {
	w := lineWidth * 0.5
	lx, ly := p1.x-p1.dmx*w, p1.y-p1.dmy*w
	rx, ry := p1.x+p1.dmx*w, p1.y+p1.dmy*w
	r.addEdge(lx, ly, left.x, left.y)
	r.addEdge(right.x, right.y, rx, ry)
	left.x, left.y = lx, ly
	right.x, right.y = rx, ry
}

func curveDivs(r, arc, tol float32) int {
	da := float32(math.Acos(float64(r/(r+tol)))) * 2
	divs := int(float32(math.Ceil(float64(arc / da))))
	return max(divs, 2)
}

func (r *rasterizer) expandStroke(points []point, closed bool, lineJoin, lineCap int, lineWidth float32) {
	ncap := curveDivs(lineWidth*0.5, pi32, tessTol)
	var left, right, firstLeft, firstRight point
	var i0, i1, s, e int
	npoints := len(points)
	if closed {
		i0, i1, s, e = npoints-1, 0, 0, npoints
	} else {
		i0, i1, s, e = 0, 1, 1, npoints-1
	}
	if closed {
		initClosed(&left, &right, &points[i0], &points[i1], lineWidth)
		firstLeft, firstRight = left, right
	} else {
		dx, dy := points[i1].x-points[i0].x, points[i1].y-points[i0].y
		normalize(&dx, &dy)
		switch lineCap {
		case capButt:
			r.buttCap(&left, &right, &points[i0], dx, dy, lineWidth, false)
		case capSquare:
			r.squareCap(&left, &right, &points[i0], dx, dy, lineWidth, false)
		case capRound:
			r.roundCap(&left, &right, &points[i0], dx, dy, lineWidth, ncap, false)
		}
	}
	for j := s; j < e; j++ {
		p0, p1 := &points[i0], &points[i1]
		if p1.flags&ptCorner != 0 {
			switch {
			case lineJoin == joinRound:
				r.roundJoin(&left, &right, p0, p1, lineWidth, ncap)
			case lineJoin == joinBevel || p1.flags&ptBevel != 0:
				r.bevelJoin(&left, &right, p0, p1, lineWidth)
			default:
				r.miterJoin(&left, &right, p0, p1, lineWidth)
			}
		} else {
			r.straightJoin(&left, &right, p1, lineWidth)
		}
		i0 = i1
		i1++
	}
	if closed {
		r.addEdge(firstLeft.x, firstLeft.y, left.x, left.y)
		r.addEdge(right.x, right.y, firstRight.x, firstRight.y)
	} else {
		p0, p1 := &points[i0], &points[i1]
		dx, dy := p1.x-p0.x, p1.y-p0.y
		normalize(&dx, &dy)
		switch lineCap {
		case capButt:
			r.buttCap(&right, &left, p1, -dx, -dy, lineWidth, true)
		case capSquare:
			r.squareCap(&right, &left, p1, -dx, -dy, lineWidth, true)
		case capRound:
			r.roundCap(&right, &left, p1, -dx, -dy, lineWidth, ncap, true)
		}
	}
}

func (r *rasterizer) prepareStroke(miterLimit float32, lineJoin int) {
	n := len(r.points)
	for i := range n {
		p0, p1 := &r.points[(i+n-1)%n], &r.points[i]
		p0.dx, p0.dy = p1.x-p0.x, p1.y-p0.y
		p0.len = normalize(&p0.dx, &p0.dy)
	}
	for j := range n {
		p0, p1 := &r.points[(j+n-1)%n], &r.points[j]
		dlx0, dly0 := p0.dy, -p0.dx
		dlx1, dly1 := p1.dy, -p1.dx
		p1.dmx = (dlx0 + dlx1) * 0.5
		p1.dmy = (dly0 + dly1) * 0.5
		dmr2 := p1.dmx*p1.dmx + p1.dmy*p1.dmy
		if dmr2 > 0.000001 {
			s2 := 1 / dmr2
			if s2 > 600 {
				s2 = 600
			}
			p1.dmx *= s2
			p1.dmy *= s2
		}
		if p1.flags&ptCorner != 0 {
			p1.flags = ptCorner
		} else {
			p1.flags = 0
		}
		if cross := p1.dx*p0.dy - p0.dx*p1.dy; cross > 0 {
			p1.flags |= ptLeft
		}
		if p1.flags&ptCorner != 0 {
			if dmr2*miterLimit*miterLimit < 1 || lineJoin == joinBevel || lineJoin == joinRound {
				p1.flags |= ptBevel
			}
		}
	}
}

func (r *rasterizer) flattenShapeStroke(s *shape, scale float32) {
	lineWidth := s.strokeWidth * scale
	for _, p := range s.paths {
		r.points = r.points[:0]
		r.addPathPoint(p.pts[0]*scale, p.pts[1]*scale, ptCorner)
		n := len(p.pts) / 2
		for i := 0; i < n-1; i += 3 {
			q := p.pts[i*2:]
			r.flattenCubicBez(q[0]*scale, q[1]*scale, q[2]*scale, q[3]*scale,
				q[4]*scale, q[5]*scale, q[6]*scale, q[7]*scale, 0, ptCorner)
		}
		if len(r.points) < 2 {
			continue
		}
		closed := p.closed
		last, first := r.points[len(r.points)-1], r.points[0]
		if ptEquals(last.x, last.y, first.x, first.y, distTol) {
			r.points = r.points[:len(r.points)-1]
			closed = true
		}
		r.prepareStroke(s.miterLimit, s.strokeLineJoin)
		r.expandStroke(r.points, closed, s.strokeLineJoin, s.strokeLineCap, lineWidth)
	}
}

func (r *rasterizer) addActive(e *edge, startPoint float32) *activeEdge {
	z := &activeEdge{}
	dxdy := (e.x1 - e.x0) / (e.y1 - e.y0)
	if dxdy < 0 {
		z.dx = int(-roundf(fix * -dxdy))
	} else {
		z.dx = int(roundf(fix * dxdy))
	}
	z.x = int(roundf(fix * (e.x0 + dxdy*(startPoint-e.y0))))
	z.ey = e.y1
	z.dir = e.dir
	return z
}

func fillScanline(scanline []uint8, x0, x1, maxWeight int, xmin, xmax *int) {
	l := len(scanline)
	i, j := x0>>fixShift, x1>>fixShift
	if i < *xmin {
		*xmin = i
	}
	if j > *xmax {
		*xmax = j
	}
	if i < l && j >= 0 {
		if i == j {
			scanline[i] = uint8(int(scanline[i]) + ((x1 - x0) * maxWeight >> fixShift))
		} else {
			if i >= 0 {
				scanline[i] = uint8(int(scanline[i]) + (((fix - (x0 & fixMask)) * maxWeight) >> fixShift))
			} else {
				i = -1
			}
			if j < l {
				scanline[j] = uint8(int(scanline[j]) + (((x1 & fixMask) * maxWeight) >> fixShift))
			} else {
				j = l
			}
			for i++; i < j; i++ {
				scanline[i] = uint8(int(scanline[i]) + maxWeight)
			}
		}
	}
}

func fillActiveEdges(scanline []uint8, e *activeEdge, maxWeight int, xmin, xmax *int, evenOdd bool) {
	x0, w := 0, 0
	if !evenOdd {
		for ; e != nil; e = e.next {
			if w == 0 {
				x0 = e.x
				w += e.dir
			} else {
				x1 := e.x
				w += e.dir
				if w == 0 {
					fillScanline(scanline, x0, x1, maxWeight, xmin, xmax)
				}
			}
		}
		return
	}
	for ; e != nil; e = e.next {
		if w == 0 {
			x0 = e.x
			w = 1
		} else {
			x1 := e.x
			w = 0
			fillScanline(scanline, x0, x1, maxWeight, xmin, xmax)
		}
	}
}

func div255(x int) int { return ((x + 1) * 257) >> 16 }

type cachedPaint struct {
	typ    paintType
	xform  [6]float32
	colors [256]uint32
}

func (r *rasterizer) scanlineSolid(dst []uint8, count int, cover []uint8, x, y int, tx, ty, scale float32, cache *cachedPaint) {
	blend := func(i int, c uint32) {
		cr, cg, cb, ca := int(c&0xff), int(c>>8&0xff), int(c>>16&0xff), int(c>>24&0xff)
		a := div255(int(cover[i]) * ca)
		ia := 255 - a
		rr := div255(cr*a) + div255(ia*int(dst[i*4+0]))
		gg := div255(cg*a) + div255(ia*int(dst[i*4+1]))
		bb := div255(cb*a) + div255(ia*int(dst[i*4+2]))
		a += div255(ia * int(dst[i*4+3]))
		dst[i*4+0], dst[i*4+1], dst[i*4+2], dst[i*4+3] = uint8(rr), uint8(gg), uint8(bb), uint8(a)
	}
	switch cache.typ {
	case paintColor:
		c := cache.colors[0]
		opaque := c>>24 == 0xff
		for i := range count {
			switch {
			case cover[i] == 0:
				// blend would leave the pixel as it is.
			case opaque && cover[i] == 255:
				// What blend computes for full coverage of an opaque colour.
				dst[i*4+0], dst[i*4+1], dst[i*4+2], dst[i*4+3] = uint8(c), uint8(c>>8), uint8(c>>16), 255
			default:
				blend(i, c)
			}
		}
	case paintLinearGradient:
		t := cache.xform
		fx := (float32(x) - tx) / scale
		fy := (float32(y) - ty) / scale
		dx := 1 / scale
		for i := range count {
			gy := fx*t[1] + fy*t[3] + t[5]
			blend(i, cache.colors[int(clampf(gy*255, 0, 255))])
			fx += dx
		}
	default:
	}
}

func (r *rasterizer) rasterizeSortedEdges(tx, ty, scale float32, cache *cachedPaint, evenOdd bool) {
	var active *activeEdge
	e := 0
	maxWeight := 255 / subsamples
	for y := 0; y < min(r.height, r.oy+r.bh); y++ {
		// Rows above the window are still stepped through, so that an edge
		// reaches the window with exactly the x it has in a full frame.
		inWindow := y >= r.oy
		if inWindow {
			clear(r.scanline)
		}
		xmin, xmax := r.width, 0
		for s := range subsamples {
			scany := float32(y*subsamples+s) + 0.5
			step := &active
			for *step != nil {
				z := *step
				if z.ey <= scany {
					*step = z.next
				} else {
					z.x += z.dx
					step = &(*step).next
				}
			}
			for {
				changed := false
				step = &active
				for *step != nil && (*step).next != nil {
					if (*step).x > (*step).next.x {
						t := *step
						q := t.next
						t.next = q.next
						q.next = t
						*step = q
						changed = true
					}
					step = &(*step).next
				}
				if !changed {
					break
				}
			}
			for e < len(r.edges) && r.edges[e].y0 <= scany {
				if r.edges[e].y1 > scany {
					z := r.addActive(&r.edges[e], scany)
					switch {
					case active == nil:
						active = z
					case z.x < active.x:
						z.next = active
						active = z
					default:
						p := active
						for p.next != nil && p.next.x < z.x {
							p = p.next
						}
						z.next = p.next
						p.next = z
					}
				}
				e++
			}
			if active != nil && inWindow {
				fillActiveEdges(r.scanline, active, maxWeight, &xmin, &xmax, evenOdd)
			}
		}
		xmin = max(xmin, r.ox)
		xmax = min(xmax, r.ox+r.bw-1)
		if inWindow && xmin <= xmax {
			off := (y-r.oy)*r.stride + (xmin-r.ox)*4
			r.scanlineSolid(r.bitmap[off:], xmax-xmin+1, r.scanline[xmin:], xmin, y, tx, ty, scale, cache)
		}
	}
}

func unpremultiplyAlpha(img []uint8, w, h, stride int) {
	for y := range h {
		row := img[y*stride:]
		for x := range w {
			p := row[x*4:]
			if a := int(p[3]); a != 0 && a != 255 {
				p[0] = uint8(int(p[0]) * 255 / a)
				p[1] = uint8(int(p[1]) * 255 / a)
				p[2] = uint8(int(p[2]) * 255 / a)
			}
		}
	}
}

func applyOpacity(c uint32, u float32) uint32 {
	iu := int(clampf(u, 0, 1) * 256)
	a := (int(c>>24&0xff) * iu) >> 8
	return c&0xffffff | uint32(a)<<24
}

func lerpRGBA(c0, c1 uint32, u float32) uint32 {
	iu := int(clampf(u, 0, 1) * 256)
	ch := func(s uint) uint32 {
		return uint32(((int(c0>>s&0xff)*(256-iu) + int(c1>>s&0xff)*iu) >> 8) & 0xff)
	}
	return ch(0) | ch(8)<<8 | ch(16)<<16 | ch(24)<<24
}

func initPaint(cache *cachedPaint, p *paint, opacity float32) {
	cache.typ = p.typ
	if p.typ == paintColor {
		cache.colors[0] = applyOpacity(p.color, opacity)
		return
	}
	g := p.gradient
	cache.xform = g.xform
	switch len(g.stops) {
	case 0:
		cache.colors = [256]uint32{}
	case 1:
		c := applyOpacity(g.stops[0].color, opacity)
		for i := range cache.colors {
			cache.colors[i] = c
		}
	default:
		ca := applyOpacity(g.stops[0].color, opacity)
		ua := clampf(g.stops[0].offset, 0, 1)
		ub := clampf(g.stops[len(g.stops)-1].offset, ua, 1)
		ia, ib := int(ua*255), int(ub*255)
		for i := 0; i < ia; i++ {
			cache.colors[i] = ca
		}
		var cb uint32
		for i := 0; i < len(g.stops)-1; i++ {
			ca = applyOpacity(g.stops[i].color, opacity)
			cb = applyOpacity(g.stops[i+1].color, opacity)
			ua = clampf(g.stops[i].offset, 0, 1)
			ub = clampf(g.stops[i+1].offset, 0, 1)
			ia, ib = int(ua*255), int(ub*255)
			count := ib - ia
			if count <= 0 {
				continue
			}
			u := float32(0)
			du := 1 / float32(count)
			for j := range count {
				cache.colors[ia+j] = lerpRGBA(ca, cb, u)
				u += du
			}
		}
		for i := ib; i < 256; i++ {
			cache.colors[i] = cb
		}
	}
}

// Rasterize ports nsvgRasterize with tx = ty = 0: it renders img scaled by
// scale into a w x h straight-alpha RGBA buffer.
func Rasterize(img *Image, scale float32, w, h int) []uint8 {
	return RasterizeRegion(img, scale, w, h, 0, 0, w, h)
}

// RasterizeRegion rasterizes the window x, y, rw x rh of the w x h image
// and returns its rw x rh straight-alpha RGBA pixels. They are exactly the
// pixels Rasterize gives for that window, which a repaint of part of a
// picture relies on.
func RasterizeRegion(img *Image, scale float32, w, h, x, y, rw, rh int) []uint8 {
	return new(Rasterizer).Region(img, scale, w, h, x, y, rw, rh)
}

// Rasterizer is RasterizeRegion with its buffers kept between calls, for
// rasterizing many images in a row.
type Rasterizer struct {
	r     rasterizer
	cache cachedPaint
}

// Region is RasterizeRegion. The pixels it returns are only valid until
// the next call.
func (z *Rasterizer) Region(img *Image, scale float32, w, h, x, y, rw, rh int) []uint8 {
	x, y = max(x, 0), max(y, 0)
	rw, rh = max(min(rw, w-x), 0), max(min(rh, h-y), 0)
	r := &z.r
	r.width, r.height, r.stride = w, h, rw*4
	r.ox, r.oy, r.bw, r.bh = x, y, rw, rh
	if n := rw * rh * 4; cap(r.bitmap) < n {
		r.bitmap = make([]uint8, n)
	} else {
		r.bitmap = r.bitmap[:n]
		clear(r.bitmap)
	}
	if cap(r.scanline) < w {
		r.scanline = make([]uint8, w)
	} else {
		r.scanline = r.scanline[:w]
	}
	cache := &z.cache
	run := func(evenOdd bool, p *paint, opacity float32) {
		for i := range r.edges {
			e := &r.edges[i]
			e.y0 *= subsamples
			e.y1 *= subsamples
		}
		sort.SliceStable(r.edges, func(i, j int) bool { return r.edges[i].y0 < r.edges[j].y0 })
		initPaint(cache, p, opacity)
		r.rasterizeSortedEdges(0, 0, scale, cache, evenOdd)
	}
	for _, s := range img.shapes {
		if s.fill.typ != paintNone {
			r.edges = r.edges[:0]
			r.flattenShape(s, scale)
			run(s.evenOdd, &s.fill, s.opacity)
		}
		if s.stroke.typ != paintNone && s.strokeWidth*scale > 0.01 {
			r.edges = r.edges[:0]
			r.flattenShapeStroke(s, scale)
			run(false, &s.stroke, s.opacity)
		}
	}
	unpremultiplyAlpha(r.bitmap, rw, rh, r.stride)
	return r.bitmap
}

// Size returns the photo size Tk gives img at scale (GetScaleFromParameters).
func Size(img *Image, scale float32) (int, int) {
	return int(math.Ceil(float64(img.Width * scale))), int(math.Ceil(float64(img.Height * scale)))
}

// BlendOver composites straight-alpha rgba over the colour bg (0xRRGGBB) the
// way Tk's photo instances do and returns fully opaque RGBA.
func BlendOver(rgba []uint8, bg uint64) []uint8 {
	out := make([]uint8, len(rgba))
	for i := 0; i < len(out); i += 4 {
		out[i], out[i+1], out[i+2], out[i+3] = uint8(bg>>16), uint8(bg>>8), uint8(bg), 255
	}
	Blend(out, len(rgba)/4, 0, 0, rgba, len(rgba)/4, 1)
	return out
}

// Blend composites the straight-alpha srcW x srcH image src onto the opaque
// RGBA image dst (dstW pixels wide) at (x, y), as BlendComplexAlpha in
// tkImgPhInstance.c does when a photo is drawn over existing pixels.
func Blend(dst []uint8, dstW, x, y int, src []uint8, srcW, srcH int) {
	if dstW <= 0 || len(src) < srcW*srcH*4 {
		return
	}
	dstH := len(dst) / 4 / dstW
	for sy := range srcH {
		for sx := range srcW {
			dx, dy := x+sx, y+sy
			if dx < 0 || dy < 0 || dx >= dstW || dy >= dstH {
				continue
			}
			s := src[(sy*srcW+sx)*4:]
			d := dst[(dy*dstW+dx)*4:]
			a := int(s[3])
			switch a {
			case 0:
			case 255:
				d[0], d[1], d[2] = s[0], s[1], s[2]
			default:
				ua := 255 - a
				d[0] = uint8((int(d[0])*ua + int(s[0])*a) / 255)
				d[1] = uint8((int(d[1])*ua + int(s[1])*a) / 255)
				d[2] = uint8((int(d[2])*ua + int(s[2])*a) / 255)
			}
		}
	}
}
