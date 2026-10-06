// Package draw provides higher-level drawing primitives and 3D border/relief
// rendering. It ports tk/generic/tk3d.c and tk/unix/tkUnix3d.c.
package draw

import (
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/platform"
)

// Point represents an x,y coordinate.
type Point struct {
	X, Y int
}

// Rect represents a rectangle.
type Rect struct {
	X, Y, Width, Height int
}

// FillRect fills a rectangle on a drawable.
func FillRect(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, r Rect) {
	d.FillRectangle(drawable, gc, r.X, r.Y, uint(r.Width), uint(r.Height))
}

// StrokeRect draws a rectangle outline.
func StrokeRect(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, r Rect) {
	d.DrawRectangle(drawable, gc, r.X, r.Y, uint(r.Width), uint(r.Height))
}

// Line draws a line between two points.
func Line(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, p1, p2 Point) {
	d.DrawLine(drawable, gc, p1.X, p1.Y, p2.X, p2.Y)
}

// FillPolygon fills a polygon defined by points using platform.FillPolygon.
func FillPolygon(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, points []Point) {
	if len(points) < 3 {
		return
	}
	ppoints := make([]platform.Point, len(points))
	for i, p := range points {
		ppoints[i] = platform.Point{X: int16(p.X), Y: int16(p.Y)}
	}
	d.FillPolygon(drawable, gc, ppoints, platform.PolygonComplex, platform.CoordModeOrigin)
}

// DrawLines draws connected line segments.
func DrawLines(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, points []Point) {
	if len(points) < 2 {
		return
	}
	ppoints := make([]platform.Point, len(points))
	for i, p := range points {
		ppoints[i] = platform.Point{X: int16(p.X), Y: int16(p.Y)}
	}
	d.DrawLines(drawable, gc, ppoints, platform.CoordModeOrigin)
}

// Border holds the three colors needed for 3D relief drawing:
// the base color, a lighter shade, and a darker shade.
type Border struct {
	// Pixel values for TrueColor displays.
	BgPixel    uint64
	LightPixel uint64
	DarkPixel  uint64
}

// NewBorder creates a border from a base color (16-bit RGB components).
// It computes the light and dark shading per Tk's algorithm
// (tk/unix/tkUnix3d.c TkpGetShadows).
func NewBorder(r, g, b uint16) *Border {
	return &Border{
		BgPixel:    toPixel(r, g, b),
		LightPixel: toPixel(lightColor(r, g, b)),
		DarkPixel:  toPixel(darkColor(r, g, b)),
	}
}

// NewBorderFromPixel creates a border from an 8-bit RGB pixel value.
func NewBorderFromPixel(pixel uint64) *Border {
	r8 := uint8((pixel >> 16) & 0xFF)
	g8 := uint8((pixel >> 8) & 0xFF)
	b8 := uint8(pixel & 0xFF)
	// X widens 8-bit channels as v*257 (0xff -> 0xffff).
	r := uint16(r8) * 257
	g := uint16(g8) * 257
	b := uint16(b8) * 257
	return NewBorder(r, g, b)
}

// Draw3DRectangle draws a 3D border around a rectangle.
func Draw3DRectangle(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	border *Border, x, y, width, height, borderWidth int, relief option.Relief) {

	if borderWidth <= 0 || width <= 0 || height <= 0 {
		return
	}

	// Tk_Draw3DRectangle (tk/generic/tk3d.c): the vertical bevels first,
	// then mitered horizontal bevels on top of them.
	if width < 2*borderWidth {
		borderWidth = width / 2
	}
	if height < 2*borderWidth {
		borderWidth = height / 2
	}
	switch relief {
	case option.ReliefFlat:
		return
	case option.ReliefSolid:
		d.SetForeground(gc, 0) // black
		for i := range borderWidth {
			d.DrawRectangle(drawable, gc, x+i, y+i, uint(width-2*i-1), uint(height-2*i-1))
		}
		return
	default:
	}
	verticalBevel(d, drawable, gc, border, x, y, borderWidth, height, true, relief)
	verticalBevel(d, drawable, gc, border, x+width-borderWidth, y, borderWidth, height, false, relief)
	horizontalBevel(d, drawable, gc, border, x, y, width, borderWidth, true, true, true, relief)
	horizontalBevel(d, drawable, gc, border, x, y+height-borderWidth, width, borderWidth, false, false, false, relief)
}

// verticalBevel ports Tk_3DVerticalBevel (tk/unix/tkUnix3d.c).
func verticalBevel(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	border *Border, x, y, width, height int, leftBevel bool, relief option.Relief) {
	fill := func(pixel uint64, x, y, w, h int) {
		if w <= 0 || h <= 0 {
			return
		}
		d.SetForeground(gc, pixel)
		d.FillRectangle(drawable, gc, x, y, uint(w), uint(h))
	}
	switch relief {
	case option.ReliefRaised, option.ReliefSunken:
		pixel := border.LightPixel
		if leftBevel != (relief == option.ReliefRaised) {
			pixel = border.DarkPixel
		}
		fill(pixel, x, y, width, height)
	case option.ReliefRidge, option.ReliefGroove:
		left, right := border.LightPixel, border.DarkPixel
		if relief == option.ReliefGroove {
			left, right = right, left
		}
		half := width / 2
		if !leftBevel && width&1 != 0 {
			half++
		}
		fill(left, x, y, half, height)
		fill(right, x+half, y, width-half, height)
	default:
	}
}

// horizontalBevel ports Tk_3DHorizontalBevel (tk/unix/tkUnix3d.c): one
// line per row, with ends that slope in or out to miter the corners.
func horizontalBevel(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	border *Border, x, y, width, height int, leftIn, rightIn, topBevel bool, relief option.Relief) {
	var topPixel, bottomPixel uint64
	switch relief {
	case option.ReliefGroove:
		topPixel, bottomPixel = border.DarkPixel, border.LightPixel
	case option.ReliefRidge:
		topPixel, bottomPixel = border.LightPixel, border.DarkPixel
	case option.ReliefRaised, option.ReliefSunken:
		topPixel = border.LightPixel
		if topBevel != (relief == option.ReliefRaised) {
			topPixel = border.DarkPixel
		}
		bottomPixel = topPixel
	default:
		return
	}

	x1, x2 := x, x+width
	x1Delta, x2Delta := -1, 1
	if leftIn {
		x1Delta = 1
	} else {
		x1 += height
	}
	if rightIn {
		x2Delta = -1
	} else {
		x2 -= height
	}
	halfway := y + height/2
	if !topBevel && height&1 != 0 {
		halfway++
	}
	for bottom := y + height; y < bottom; y++ {
		if x1 < x2 {
			pixel := bottomPixel
			if y < halfway {
				pixel = topPixel
			}
			d.SetForeground(gc, pixel)
			d.FillRectangle(drawable, gc, x1, y, uint(x2-x1), 1)
		}
		x1 += x1Delta
		x2 += x2Delta
	}
}

// Fill3DRectangle fills a rectangle with 3D relief.
func Fill3DRectangle(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	border *Border, x, y, width, height, borderWidth int, relief option.Relief) {

	// Fill interior.
	innerX := x + borderWidth
	innerY := y + borderWidth
	innerW := width - 2*borderWidth
	innerH := height - 2*borderWidth
	if innerW > 0 && innerH > 0 {
		d.SetForeground(gc, border.BgPixel)
		d.FillRectangle(drawable, gc, innerX, innerY, uint(innerW), uint(innerH))
	}

	// Draw the 3D border.
	Draw3DRectangle(d, drawable, gc, border, x, y, width, height, borderWidth, relief)
}

// lightColor computes the light shadow from Tk's algorithm.
// Input/output are 16-bit RGB.
func lightColor(r, g, b uint16) (uint16, uint16, uint16) {
	const maxVal = 65535

	// If very bright (green > 95%), darken by 10%.
	if g > maxVal*95/100 {
		return uint16(90 * uint32(r) / 100),
			uint16(90 * uint32(g) / 100),
			uint16(90 * uint32(b) / 100)
	}

	// Normal: maxVal of (boost 40%, halfway to white).
	boost := func(v uint16) uint16 { return uint16(min(14*uint32(v)/10, maxVal)) }
	lr := maxU16(boost(r), uint16((maxVal+uint32(r))/2))
	lg := maxU16(boost(g), uint16((maxVal+uint32(g))/2))
	lb := maxU16(boost(b), uint16((maxVal+uint32(b))/2))
	return lr, lg, lb
}

// darkColor computes the dark shadow from Tk's algorithm.
func darkColor(r, g, b uint16) (uint16, uint16, uint16) {
	const maxVal = 65535

	// Check perceived brightness: 0.5*R^2 + 1.0*G^2 + 0.28*B^2.
	brightness := 0.5*float64(r)*float64(r) +
		1.0*float64(g)*float64(g) +
		0.28*float64(b)*float64(b)
	threshold := 0.05 * float64(maxVal) * float64(maxVal)

	if brightness < threshold {
		// Very dark: lighten by 1/4 toward white.
		return uint16((maxVal + 3*uint32(r)) / 4),
			uint16((maxVal + 3*uint32(g)) / 4),
			uint16((maxVal + 3*uint32(b)) / 4)
	}

	// Normal: reduce by 40%.
	return uint16(60 * uint32(r) / 100),
		uint16(60 * uint32(g) / 100),
		uint16(60 * uint32(b) / 100)
}

func toPixel(r, g, b uint16) uint64 {
	return uint64(r>>8)<<16 | uint64(g>>8)<<8 | uint64(b>>8)
}

func maxU16(a, b uint16) uint16 {
	if a > b {
		return a
	}
	return b
}

func min64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
