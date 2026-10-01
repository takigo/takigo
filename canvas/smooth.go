package canvas

import (
	"math"

	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/internal/nanosvg"
	"github.com/msorc/takigo/platform"
)

// Anti-aliased drawing. Tk draws canvas items with the X core protocol
// (XDrawLine, XFillPolygon, XFillArc), whose edges are jagged; here the
// shapes of consecutive items are collected in a layer and rasterized with
// coverage, then blended into the pixmap. Items that need the display
// server (text, images, embedded windows, stipples, dashes) are drawn by
// it as before, in order, between layers. Antialias(false) and a Classic
// App keep Tk's drawing.

// smoothItem is implemented by the items that can be drawn anti-aliased.
type smoothItem interface {
	// paintSmooth adds the item's shapes to l and reports true, or reports
	// false when this item needs the display server's drawing.
	paintSmooth(l *layer) bool
}

// layer is the anti-aliased drawing of one repaint. Shapes are given in
// the frame of the canvas pixmap, whose pixel (0, 0) is the canvas
// coordinate (x0, y0), whatever part is repainted: an item then has the
// same coordinates, and so exactly the same pixels, in every repaint.
//
// Each item is rasterized by itself and blended over what is below it.
// Blending item by item, rather than collecting shapes and blending them
// together, makes the result independent of which other items take part
// in a repaint, so a repaint of a damaged area gives the pixels a full
// repaint would.
//
// While the repainted window holds only the background, items are blended
// into under, a copy of it that needs no read-back. Once the display server
// has drawn there (text, an image), the items of a run are held back and
// blended, when the run ends, into the part of the pixmap they cover, read
// back once.
type layer struct {
	img    *nanosvg.Image
	raster nanosvg.Rasterizer
	x0, y0 int
	w, h   int

	rx, ry, rw, rh int // the repainted window, in frame pixels

	plain bool    // the window holds only the background colour
	under []uint8 // the window's pixels while plain, opaque RGBA
	dirty bool    // under has changes the pixmap does not
	bg    uint64

	run            []coverage // items waiting to be blended (not plain)
	bx1, by1       int        // the box the run covers, in frame pixels
	bx2, by2       int
	readBack       bool // GetImageRGBA works on the pixmap
	readBackTested bool
}

// coverage is one rasterized item: straight-alpha pixels and their place
// in the frame.
type coverage struct {
	px         []uint8
	x, y, w, h int
}

func newLayer(x0, y0, w, h, rx, ry, rw, rh int) *layer {
	return &layer{img: nanosvg.NewImage(w, h), x0: x0, y0: y0, w: w, h: h, rx: rx, ry: ry, rw: rw, rh: rh}
}

// paint draws one item anti-aliased; ix1 .. iy2 is its bounding box in
// canvas coordinates. It reports false when the item must be drawn by the
// display server instead.
func (l *layer) paint(d platform.DisplayServer, drawable platform.DrawableID, item smoothItem, ix1, iy1, ix2, iy2 int) bool {
	if !l.plain {
		if !l.readBackTested {
			l.readBack = d.GetImageRGBA(drawable, l.rx, l.ry, 1, 1) != nil
			l.readBackTested = true
		}
		if !l.readBack {
			return false // nothing to blend with
		}
	}
	l.img.Reset()
	if !item.paintSmooth(l) {
		return false
	}
	if l.img.Empty() {
		return true
	}
	// Rasterize the item's box only, with room for the soft edge.
	const pad = 2
	wx := max(ix1-l.x0-pad, l.rx)
	wy := max(iy1-l.y0-pad, l.ry)
	ww := min(ix2-l.x0+pad, l.rx+l.rw) - wx
	wh := min(iy2-l.y0+pad, l.ry+l.rh) - wy
	if ww <= 0 || wh <= 0 {
		return true
	}
	px := l.raster.Region(l.img, 1, l.w, l.h, wx, wy, ww, wh)
	if l.plain {
		if l.under == nil {
			l.under = make([]uint8, l.rw*l.rh*4)
			r, g, b := uint8(l.bg>>16), uint8(l.bg>>8), uint8(l.bg)
			for i := 0; i < len(l.under); i += 4 {
				l.under[i], l.under[i+1], l.under[i+2], l.under[i+3] = r, g, b, 255
			}
		}
		nanosvg.Blend(l.under, l.rw, wx-l.rx, wy-l.ry, px, ww, wh)
		l.dirty = true
		return true
	}
	if len(l.run) == 0 {
		l.bx1, l.by1, l.bx2, l.by2 = wx, wy, wx+ww, wy+wh
	} else {
		l.bx1, l.by1 = min(l.bx1, wx), min(l.by1, wy)
		l.bx2, l.by2 = max(l.bx2, wx+ww), max(l.by2, wy+wh)
	}
	l.run = append(l.run, coverage{px: append([]uint8(nil), px...), x: wx, y: wy, w: ww, h: wh})
	return true
}

// flush puts what has been blended into the pixmap.
func (l *layer) flush(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, depth int) {
	if l.dirty {
		d.PutImageRGBA(drawable, gc, depth, l.under, l.rw*4, l.rw, l.rh, 0, 0, l.rx, l.ry, l.rw, l.rh, l.bg)
		l.dirty = false
	}
	if len(l.run) == 0 {
		return
	}
	bw, bh := l.bx2-l.bx1, l.by2-l.by1
	if under := d.GetImageRGBA(drawable, l.bx1, l.by1, bw, bh); under != nil {
		for _, c := range l.run {
			nanosvg.Blend(under, bw, c.x-l.bx1, c.y-l.by1, c.px, c.w, c.h)
		}
		d.PutImageRGBA(drawable, gc, depth, under, bw*4, bw, bh, 0, 0, l.bx1, l.by1, bw, bh, l.bg)
	}
	l.run = l.run[:0]
}

// invalidate records that the display server is about to draw into the
// window: from then on blending needs the pixmap's own pixels.
func (l *layer) invalidate() {
	l.under = nil
	l.plain = false
}

// x and y convert a canvas coordinate to the layer's, rounded to a whole
// pixel exactly as the X drawing rounds it (drawableCoord), so that
// axis-aligned edges stay sharp and both ways of drawing agree on where an
// item is.
func (l *layer) x(v float64) float32 { return float32(drawableCoord(v, l.x0)) }
func (l *layer) y(v float64) float32 { return float32(drawableCoord(v, l.y0)) }

func paintColor(c *color.ColorRef) uint32 {
	return nanosvg.RGBA(uint8(c.Pixel>>16), uint8(c.Pixel>>8), uint8(c.Pixel), c.Alpha())
}

// fill adds a filled path. X fills the pixels whose centres lie inside a
// shape whose vertices are pixel centres, with ties going up and left: for
// integer vertices that is the half-open area the vertices bound, which is
// what the path covers here without any offset.
func (l *layer) fill(pts []float32, c *color.ColorRef, evenOdd bool) {
	l.img.FillPath(pts, paintColor(c), evenOdd)
}

// stroke adds a stroked path of width pixels. X centres a line on the
// pixel centres of its end points: a line of odd width is therefore moved
// half a pixel so it covers whole pixels, and one of even width is not.
func (l *layer) stroke(pts []float32, closed bool, c *color.ColorRef, width int, lineCap, lineJoin int) {
	width = max(width, 1)
	if width%2 == 1 {
		shifted := make([]float32, len(pts))
		for i, v := range pts {
			shifted[i] = v + 0.5
		}
		pts = shifted
	}
	l.img.StrokePath(pts, closed, paintColor(c), float32(width), lineCap, lineJoin)
}

func (r *RectOvalItem) paintSmooth(l *layer) bool {
	if r.stipple != "" || r.outlineStipple != "" {
		return false
	}
	x1, y1 := l.x(r.coords[0]), l.y(r.coords[1])
	x2, y2 := l.x(r.coords[2]), l.y(r.coords[3])
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	if x2-x1 <= 0 || y2-y1 <= 0 {
		return true
	}
	var path, outline []float32
	join := nanosvg.JoinMiter
	if r.typeName == "rectangle" {
		outline = []float32{x1, y1, x2, y1, x2, y2, x1, y2, x1, y1}
		path = nanosvg.Polyline(outline)
	} else {
		cx, cy, rx, ry := (x1+x2)/2, (y1+y2)/2, (x2-x1)/2, (y2-y1)/2
		path = nanosvg.Ellipse(cx, cy, rx, ry)
		join = nanosvg.JoinRound
		if len(r.dash) > 0 {
			outline = ellipsePoints(cx, cy, rx, ry, 0, 360)
		}
	}
	if fill := r.fillFor(r.fill); fill != nil {
		l.fill(path, fill, false)
	}
	if r.outline != nil && r.outlineWidth > 0 {
		if len(r.dash) > 0 {
			l.strokeDashed(outline, r.dash, r.outline, r.outlineWidth, nanosvg.CapButt, join)
		} else {
			l.stroke(path, true, r.outline, r.outlineWidth, nanosvg.CapButt, join)
		}
	}
	return true
}

// points converts canvas coordinates to layer coordinates. Vertices the
// user gave are rounded to whole pixels as X does; the points of a spline
// are not, or the rounding would put the steps back into the curve.
func (l *layer) points(coords []float64, exact bool) []float32 {
	xy := make([]float32, len(coords))
	for i := 0; i+1 < len(coords); i += 2 {
		if exact {
			xy[i], xy[i+1] = float32(coords[i]-float64(l.x0)), float32(coords[i+1]-float64(l.y0))
		} else {
			xy[i], xy[i+1] = l.x(coords[i]), l.y(coords[i+1])
		}
	}
	return xy
}

// strokeOffset is how far stroke moves a path of the given width; shapes
// that must line up with a stroke (arrowheads) move with it.
func strokeOffset(width int) float32 {
	if max(width, 1)%2 == 1 {
		return 0.5
	}
	return 0
}

func smoothCap(capStyle int) int {
	switch capStyle {
	case platform.CapRound:
		return nanosvg.CapRound
	case platform.CapProjecting:
		return nanosvg.CapSquare
	}
	return nanosvg.CapButt
}

func smoothJoin(joinStyle int) int {
	switch joinStyle {
	case platform.JoinRound:
		return nanosvg.JoinRound
	case platform.JoinBevel:
		return nanosvg.JoinBevel
	}
	return nanosvg.JoinMiter
}

func (p *PolygonItem) paintSmooth(l *layer) bool {
	if p.stipple != "" || p.outlineStipple != "" {
		return false
	}
	if len(p.coords) < 6 {
		return true
	}
	coords := p.coords
	if p.smooth {
		c := p.canvas
		coords = appendBezierSpline(c.coordBuf[:0], p.coords, true, p.splineSteps)
		c.coordBuf = coords
	}
	xy := l.points(coords, p.smooth)
	xy = append(xy, xy[0], xy[1])
	path := nanosvg.Polyline(xy)
	if fill := p.fillFor(p.fill); fill != nil {
		// X's default fill rule is EvenOdd.
		l.fill(path, fill, true)
	}
	if p.outline != nil && p.outlineWidth > 0 {
		if len(p.dash) > 0 {
			l.strokeDashed(xy, p.dash, p.outline, p.outlineWidth, nanosvg.CapRound, nanosvg.JoinRound)
		} else {
			l.stroke(path, true, p.outline, p.outlineWidth, nanosvg.CapRound, nanosvg.JoinRound)
		}
	}
	return true
}

func (ln *LineItem) paintSmooth(l *layer) bool {
	if ln.stipple != "" {
		return false
	}
	if len(ln.coords) < 4 || ln.color == nil {
		return true
	}
	xy := l.points(ln.shaftCoords(), ln.smooth && len(ln.coords) >= 6)
	if len(ln.dash) > 0 {
		l.strokeDashed(xy, ln.dash, ln.color, ln.width, smoothCap(ln.capStyle), smoothJoin(ln.joinStyle))
	} else {
		l.strokeLine(xy, ln.color, ln.width, smoothCap(ln.capStyle), smoothJoin(ln.joinStyle))
	}

	off := strokeOffset(ln.width)
	arrow := func(first bool) {
		p, ok := ln.arrowPolygon(first)
		if !ok {
			return
		}
		head := l.points(p[:], true)
		for i := range head {
			head[i] += off
		}
		l.fill(nanosvg.Polyline(head), ln.color, false)
	}
	if ln.arrow == ArrowFirst || ln.arrow == ArrowBoth {
		arrow(true)
	}
	if ln.arrow == ArrowLast || ln.arrow == ArrowBoth {
		arrow(false)
	}
	return true
}

func (a *ArcItem) paintSmooth(l *layer) bool {
	if a.stipple != "" || a.outlineStipple != "" {
		return false
	}
	x1, y1 := l.x(a.coords[0]), l.y(a.coords[1])
	x2, y2 := l.x(a.coords[2]), l.y(a.coords[3])
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	if x2-x1 <= 0 || y2-y1 <= 0 || a.extent == 0 {
		return true
	}
	cx, cy, rx, ry := (x1+x2)/2, (y1+y2)/2, (x2-x1)/2, (y2-y1)/2
	path := nanosvg.Arc(cx, cy, rx, ry, float32(a.start), float32(a.extent))
	if len(path) == 0 {
		return true
	}
	n := len(path)
	sx, sy, ex, ey := path[0], path[1], path[n-2], path[n-1]
	// The straight sides that close the shape, as X draws them: separate
	// lines, each starting the dash pattern afresh.
	var sides [][]float32
	switch a.style {
	case ArcStylePieslice:
		// The wedge: the arc, a line to the centre, and back to the start.
		path = append(path, ex, ey, cx, cy, cx, cy)
		path = append(path, cx, cy, sx, sy, sx, sy)
		sides = [][]float32{{cx, cy, sx, sy}, {cx, cy, ex, ey}}
	case ArcStyleChord:
		path = append(path, ex, ey, sx, sy, sx, sy)
		sides = [][]float32{{sx, sy, ex, ey}}
	}
	if a.fill != nil && a.style != ArcStyleArc {
		l.fill(path, a.fill, false)
	}
	if a.outline == nil || a.outlineWidth <= 0 {
		return true
	}
	if len(a.dash) > 0 {
		curve := ellipsePoints(cx, cy, rx, ry, float32(a.start), float32(a.extent))
		l.strokeDashed(curve, a.dash, a.outline, a.outlineWidth, nanosvg.CapButt, nanosvg.JoinMiter)
		for _, side := range sides {
			l.strokeDashed(side, a.dash, a.outline, a.outlineWidth, nanosvg.CapButt, nanosvg.JoinMiter)
		}
		return true
	}
	l.stroke(path, a.style != ArcStyleArc, a.outline, a.outlineWidth, nanosvg.CapButt, nanosvg.JoinMiter)
	return true
}

// strokeLine strokes the open polyline xy. With butt caps, X draws a line
// from pixel centre a to pixel centre b over the pixels [a, b): after the
// half-pixel move of an odd width, both ends are therefore pulled back half
// a pixel along the line, which keeps the ends of horizontal and vertical
// lines, and of their dashes, on pixel edges.
func (l *layer) strokeLine(xy []float32, c *color.ColorRef, width int, lineCap, lineJoin int) {
	if len(xy) < 4 {
		return
	}
	if lineCap == nanosvg.CapButt && strokeOffset(width) != 0 {
		xy = append([]float32{}, xy...)
		back := func(i, j int) { // move point i half a pixel away from where j lies ahead of it
			dx, dy := xy[j]-xy[i], xy[j+1]-xy[i+1]
			if n := float32(math.Hypot(float64(dx), float64(dy))); n > 0 {
				xy[i] -= dx / n / 2
				xy[i+1] -= dy / n / 2
			}
		}
		n := len(xy)
		ex, ey := xy[n-2]-xy[n-4], xy[n-1]-xy[n-3]
		back(0, 2)
		if d := float32(math.Hypot(float64(ex), float64(ey))); d > 0 {
			xy[n-2] -= ex / d / 2
			xy[n-1] -= ey / d / 2
		}
	}
	l.stroke(nanosvg.Polyline(xy), false, c, width, lineCap, lineJoin)
}

// strokeDashed strokes the "on" pieces of an X dash pattern along the
// polyline xy; see dashPieces.
func (l *layer) strokeDashed(xy []float32, pattern []byte, c *color.ColorRef, width int, lineCap, lineJoin int) {
	for _, piece := range dashPieces(xy, pattern) {
		l.strokeLine(piece, c, width, lineCap, lineJoin)
	}
}

// dashPieces cuts the polyline xy (x0, y0, x1, y1, ...) into the pieces an
// X LineOnOffDash line draws: pattern holds alternating on and off lengths
// in pixels, starting on, and repeats along the line. A pattern of odd
// length is used twice over, as XSetDashes does.
func dashPieces(xy []float32, pattern []byte) [][]float32 {
	total := 0
	for _, n := range pattern {
		total += int(n)
	}
	if len(xy) < 4 || total == 0 {
		return [][]float32{xy}
	}
	if len(pattern)%2 == 1 {
		pattern = append(append([]byte{}, pattern...), pattern...)
	}
	var pieces [][]float32
	var cur []float32
	on := true
	i := 0
	left := float32(pattern[0])
	advance := func() {
		i = (i + 1) % len(pattern)
		left = float32(pattern[i])
		on = i%2 == 0
	}
	for left == 0 {
		advance()
	}
	x, y := xy[0], xy[1]
	cur = append(cur, x, y)
	for k := 2; k+1 < len(xy); k += 2 {
		x2, y2 := xy[k], xy[k+1]
		seg := float32(math.Hypot(float64(x2-x), float64(y2-y)))
		for seg > left {
			// The pattern element ends inside this segment.
			t := left / seg
			x, y = x+(x2-x)*t, y+(y2-y)*t
			seg -= left
			if on {
				pieces = append(pieces, append(cur, x, y))
			}
			advance()
			for left == 0 {
				advance()
			}
			cur = nil
			if on {
				cur = []float32{x, y}
			}
		}
		left -= seg
		x, y = x2, y2
		if on {
			cur = append(cur, x, y)
		}
	}
	if on && len(cur) >= 4 {
		pieces = append(pieces, cur)
	}
	return pieces
}

// ellipsePoints returns a polyline along an elliptical arc, close enough to
// it for cutting into dashes (X's angle convention: 90 degrees is up).
func ellipsePoints(cx, cy, rx, ry, start, extent float32) []float32 {
	n := max(int(math.Abs(float64(extent))/360*float64(max(rx, ry))*math.Pi), 16)
	xy := make([]float32, 0, 2*(n+1))
	for i := range n + 1 {
		a := (float64(start) + float64(extent)*float64(i)/float64(n)) * math.Pi / 180
		xy = append(xy, cx+rx*float32(math.Cos(a)), cy-ry*float32(math.Sin(a)))
	}
	return xy
}
