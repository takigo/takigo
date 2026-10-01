package nanosvg

import "math"

// Line caps and joins for StrokePath.
const (
	CapButt   = capButt
	CapRound  = capRound
	CapSquare = capSquare

	JoinMiter = joinMiter
	JoinRound = joinRound
	JoinBevel = joinBevel
)

// NewImage returns an empty image w x h to add shapes to with FillPath and
// StrokePath, for rasterizing geometry built in code rather than parsed
// from SVG. Shapes are painted in the order they are added.
func NewImage(w, h int) *Image {
	return &Image{Width: float32(w), Height: float32(h)}
}

// Empty reports whether the image has no shapes.
func (img *Image) Empty() bool { return len(img.shapes) == 0 }

// Reset removes the image's shapes, keeping its size.
func (img *Image) Reset() { img.shapes = img.shapes[:0] }

// RGBA packs a straight-alpha colour for FillPath and StrokePath.
func RGBA(r, g, b, a uint8) uint32 {
	return uint32(r) | uint32(g)<<8 | uint32(b)<<16 | uint32(a)<<24
}

// FillPath adds a filled shape. pts is a path in the form Polyline,
// Ellipse and Arc return: a start point followed by cubic segments. The
// path is closed for filling.
func (img *Image) FillPath(pts []float32, color uint32, evenOdd bool) {
	if len(pts) < 8 {
		return
	}
	img.shapes = append(img.shapes, &shape{
		fill:    paint{typ: paintColor, color: color},
		opacity: 1,
		evenOdd: evenOdd,
		paths:   []*path{{pts: pts, closed: true}},
	})
}

// StrokePath adds a stroked path of the given width.
func (img *Image) StrokePath(pts []float32, closed bool, color uint32, width float32, lineCap, lineJoin int) {
	if len(pts) < 8 || width <= 0 {
		return
	}
	img.shapes = append(img.shapes, &shape{
		stroke:         paint{typ: paintColor, color: color},
		opacity:        1,
		strokeWidth:    width,
		strokeLineCap:  lineCap,
		strokeLineJoin: lineJoin,
		miterLimit:     10,
		paths:          []*path{{pts: pts, closed: closed}},
	})
}

// Polyline returns the path through the points xy (x0, y0, x1, y1, ...).
func Polyline(xy []float32) []float32 {
	if len(xy) < 4 {
		return nil
	}
	pts := make([]float32, 0, 2+3*(len(xy)-2))
	pts = append(pts, xy[0], xy[1])
	for i := 2; i+1 < len(xy); i += 2 {
		x0, y0, x1, y1 := xy[i-2], xy[i-1], xy[i], xy[i+1]
		pts = append(pts, x0, y0, x1, y1, x1, y1)
	}
	return pts
}

// Ellipse returns the closed path of an ellipse.
func Ellipse(cx, cy, rx, ry float32) []float32 {
	return Arc(cx, cy, rx, ry, 0, 360)
}

// Arc returns the path of an elliptical arc starting at angle start and
// sweeping extent, both in degrees, counter-clockwise with y pointing down
// (X's and Tk's convention: 90 is up).
func Arc(cx, cy, rx, ry, start, extent float32) []float32 {
	n := int(math.Ceil(math.Abs(float64(extent)) / 90))
	if n == 0 {
		return nil
	}
	step := float64(extent) / float64(n) * math.Pi / 180
	// Control-point distance for a cubic approximating an arc of step.
	k := float32(4.0 / 3.0 * math.Tan(step/4))
	a := float64(start) * math.Pi / 180
	point := func(a float64) (x, y, dx, dy float32) {
		s, c := math.Sincos(a)
		return cx + rx*float32(c), cy - ry*float32(s), -rx * float32(s), -ry * float32(c)
	}
	x, y, dx, dy := point(a)
	pts := make([]float32, 0, 2+6*n)
	pts = append(pts, x, y)
	for range n {
		a += step
		x2, y2, dx2, dy2 := point(a)
		pts = append(pts, x+k*dx, y+k*dy, x2-k*dx2, y2-k*dy2, x2, y2)
		x, y, dx, dy = x2, y2, dx2, dy2
	}
	return pts
}
