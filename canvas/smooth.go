package canvas

import (
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
// Each item is rasterized by itself and blended into under, a copy of the
// repainted window of the pixmap. Blending item by item, rather than
// collecting shapes and blending them together, makes the result
// independent of which other items take part in a repaint, so a repaint of
// a damaged area gives the pixels a full repaint would.
type layer struct {
	img    *nanosvg.Image
	x0, y0 int
	w, h   int

	rx, ry, rw, rh int // the repainted window, in frame pixels

	under []uint8 // the window's pixels, opaque RGBA; nil until needed
	plain bool    // the window holds only the background colour
	dirty bool    // under has changes the pixmap does not
	bg    uint64
}

func newLayer(x0, y0, w, h, rx, ry, rw, rh int) *layer {
	return &layer{img: nanosvg.NewImage(w, h), x0: x0, y0: y0, w: w, h: h, rx: rx, ry: ry, rw: rw, rh: rh}
}

// paint draws one item anti-aliased; ix1 .. iy2 is its bounding box in
// canvas coordinates. It reports false when the item must be drawn by the
// display server instead.
func (l *layer) paint(d platform.DisplayServer, drawable platform.DrawableID, item smoothItem, ix1, iy1, ix2, iy2 int) bool {
	l.img.Reset()
	if !item.paintSmooth(l) {
		return false
	}
	if l.img.Empty() {
		return true
	}
	if l.under == nil {
		if l.plain {
			l.under = make([]uint8, l.rw*l.rh*4)
			r, g, b := uint8(l.bg>>16), uint8(l.bg>>8), uint8(l.bg)
			for i := 0; i < len(l.under); i += 4 {
				l.under[i], l.under[i+1], l.under[i+2], l.under[i+3] = r, g, b, 255
			}
		} else if l.under = d.GetImageRGBA(drawable, l.rx, l.ry, l.rw, l.rh); l.under == nil {
			return false // cannot read the pixmap back to blend with it
		}
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
	px := nanosvg.RasterizeRegion(l.img, 1, l.w, l.h, wx, wy, ww, wh)
	nanosvg.Blend(l.under, l.rw, wx-l.rx, wy-l.ry, px, ww, wh)
	l.dirty = true
	return true
}

// flush puts the blended pixels into the pixmap.
func (l *layer) flush(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, depth int) {
	if !l.dirty {
		return
	}
	d.PutImageRGBA(drawable, gc, depth, l.under, l.rw*4, l.rw, l.rh, 0, 0, l.rx, l.ry, l.rw, l.rh, l.bg)
	l.dirty = false
}

// invalidate records that the display server is about to draw into the
// window, after which under no longer mirrors it.
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
	return nanosvg.RGBA(uint8(c.Pixel>>16), uint8(c.Pixel>>8), uint8(c.Pixel), 255)
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
	if r.stipple != "" || r.outlineStipple != "" || len(r.dash) > 0 {
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
	var path []float32
	join := nanosvg.JoinMiter
	if r.typeName == "rectangle" {
		path = nanosvg.Polyline([]float32{x1, y1, x2, y1, x2, y2, x1, y2, x1, y1})
	} else {
		path = nanosvg.Ellipse((x1+x2)/2, (y1+y2)/2, (x2-x1)/2, (y2-y1)/2)
		join = nanosvg.JoinRound
	}
	if fill := r.fillFor(r.fill); fill != nil {
		l.fill(path, fill, false)
	}
	if r.outline != nil && r.outlineWidth > 0 {
		l.stroke(path, true, r.outline, r.outlineWidth, nanosvg.CapButt, join)
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
	if p.stipple != "" || p.outlineStipple != "" || len(p.dash) > 0 {
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
	path := nanosvg.Polyline(append(xy, xy[0], xy[1]))
	if fill := p.fillFor(p.fill); fill != nil {
		// X's default fill rule is EvenOdd.
		l.fill(path, fill, true)
	}
	if p.outline != nil && p.outlineWidth > 0 {
		l.stroke(path, true, p.outline, p.outlineWidth, nanosvg.CapRound, nanosvg.JoinRound)
	}
	return true
}

func (ln *LineItem) paintSmooth(l *layer) bool {
	if ln.stipple != "" || len(ln.dash) > 0 {
		return false
	}
	if len(ln.coords) < 4 || ln.color == nil {
		return true
	}
	xy := l.points(ln.shaftCoords(), ln.smooth && len(ln.coords) >= 6)
	l.stroke(nanosvg.Polyline(xy), false, ln.color, int(ln.width), smoothCap(ln.capStyle), smoothJoin(ln.joinStyle))

	off := strokeOffset(int(ln.width))
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
	if a.stipple != "" || a.outlineStipple != "" || len(a.dash) > 0 {
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
	cx, cy := (x1+x2)/2, (y1+y2)/2
	path := nanosvg.Arc(cx, cy, (x2-x1)/2, (y2-y1)/2, float32(a.start), float32(a.extent))
	if len(path) == 0 {
		return true
	}
	outlined := a.outline != nil && a.outlineWidth > 0
	switch a.style {
	case ArcStylePieslice:
		// The wedge: the arc, a line to the centre, and back to the start.
		n := len(path)
		path = append(path, path[n-2], path[n-1], cx, cy, cx, cy)
		path = append(path, cx, cy, path[0], path[1], path[0], path[1])
		if a.fill != nil {
			l.fill(path, a.fill, false)
		}
		if outlined {
			l.stroke(path, true, a.outline, a.outlineWidth, nanosvg.CapButt, nanosvg.JoinMiter)
		}
	case ArcStyleChord:
		n := len(path)
		path = append(path, path[n-2], path[n-1], path[0], path[1], path[0], path[1])
		if a.fill != nil {
			l.fill(path, a.fill, false)
		}
		if outlined {
			l.stroke(path, true, a.outline, a.outlineWidth, nanosvg.CapButt, nanosvg.JoinMiter)
		}
	case ArcStyleArc:
		if outlined {
			l.stroke(path, false, a.outline, a.outlineWidth, nanosvg.CapButt, nanosvg.JoinMiter)
		}
	}
	return true
}
