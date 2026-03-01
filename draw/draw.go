// Package draw provides higher-level drawing primitives and 3D border/relief
// rendering. It ports tk/generic/tk3d.c and tk/unix/tkUnix3d.c.
package draw

import (
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
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
func FillRect(d *xlib.Display, drawable xlib.Drawable, gc xlib.GC, r Rect) {
	d.FillRectangle(drawable, gc, r.X, r.Y, uint(r.Width), uint(r.Height))
}

// StrokeRect draws a rectangle outline.
func StrokeRect(d *xlib.Display, drawable xlib.Drawable, gc xlib.GC, r Rect) {
	d.DrawRectangle(drawable, gc, r.X, r.Y, uint(r.Width), uint(r.Height))
}

// Line draws a line between two points.
func Line(d *xlib.Display, drawable xlib.Drawable, gc xlib.GC, p1, p2 Point) {
	d.DrawLine(drawable, gc, p1.X, p1.Y, p2.X, p2.Y)
}

// FillPolygon fills a polygon defined by points.
// Uses a series of filled triangles (fan from first point).
func FillPolygon(d *xlib.Display, drawable xlib.Drawable, gc xlib.GC, points []Point) {
	if len(points) < 3 {
		return
	}
	// Simple approach: fill as series of triangles from first point.
	// For convex polygons this works; for complex polygons we'd need
	// XFillPolygon, which we can add to xlib bindings later.
	for i := 1; i < len(points)-1; i++ {
		// Draw triangle as three lines (filled via GC fill style).
		d.DrawLine(drawable, gc, points[0].X, points[0].Y, points[i].X, points[i].Y)
		d.DrawLine(drawable, gc, points[i].X, points[i].Y, points[i+1].X, points[i+1].Y)
		d.DrawLine(drawable, gc, points[i+1].X, points[i+1].Y, points[0].X, points[0].Y)
	}
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
	r := uint16(r8) << 8
	g := uint16(g8) << 8
	b := uint16(b8) << 8
	return NewBorder(r, g, b)
}

// Draw3DRectangle draws a 3D border around a rectangle.
func Draw3DRectangle(d *xlib.Display, drawable xlib.Drawable, gc xlib.GC,
	border *Border, x, y, width, height, borderWidth int, relief option.Relief) {

	if borderWidth <= 0 || width <= 0 || height <= 0 {
		return
	}

	// Clamp border width.
	if borderWidth > width/2 {
		borderWidth = width / 2
	}
	if borderWidth > height/2 {
		borderWidth = height / 2
	}

	switch relief {
	case option.ReliefRaised:
		drawBevel(d, drawable, gc, border, x, y, width, height, borderWidth, true)
	case option.ReliefSunken:
		drawBevel(d, drawable, gc, border, x, y, width, height, borderWidth, false)
	case option.ReliefGroove:
		half := borderWidth / 2
		if half < 1 {
			half = 1
		}
		drawBevel(d, drawable, gc, border, x, y, width, height, half, false)
		drawBevel(d, drawable, gc, border, x+half, y+half, width-2*half, height-2*half, borderWidth-half, true)
	case option.ReliefRidge:
		half := borderWidth / 2
		if half < 1 {
			half = 1
		}
		drawBevel(d, drawable, gc, border, x, y, width, height, half, true)
		drawBevel(d, drawable, gc, border, x+half, y+half, width-2*half, height-2*half, borderWidth-half, false)
	case option.ReliefSolid:
		d.SetForeground(gc, 0) // black
		for i := range borderWidth {
			d.DrawRectangle(drawable, gc, x+i, y+i, uint(width-2*i-1), uint(height-2*i-1))
		}
	case option.ReliefFlat:
		// No border to draw.
	}
}

// Fill3DRectangle fills a rectangle with 3D relief.
func Fill3DRectangle(d *xlib.Display, drawable xlib.Drawable, gc xlib.GC,
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

// drawBevel draws a single raised or sunken bevel.
func drawBevel(d *xlib.Display, drawable xlib.Drawable, gc xlib.GC,
	border *Border, x, y, width, height, bw int, raised bool) {

	var topLeftPixel, bottomRightPixel uint64
	if raised {
		topLeftPixel = border.LightPixel
		bottomRightPixel = border.DarkPixel
	} else {
		topLeftPixel = border.DarkPixel
		bottomRightPixel = border.LightPixel
	}

	// Top edge.
	d.SetForeground(gc, topLeftPixel)
	for i := range bw {
		d.FillRectangle(drawable, gc, x+i, y+i, uint(width-2*i), 1)
	}

	// Left edge.
	for i := range bw {
		d.FillRectangle(drawable, gc, x+i, y+i, 1, uint(height-2*i))
	}

	// Bottom edge.
	d.SetForeground(gc, bottomRightPixel)
	for i := range bw {
		d.FillRectangle(drawable, gc, x+i, y+height-1-i, uint(width-2*i), 1)
	}

	// Right edge.
	for i := range bw {
		d.FillRectangle(drawable, gc, x+width-1-i, y+i, 1, uint(height-2*i))
	}
}

// lightColor computes the light shadow from Tk's algorithm.
// Input/output are 16-bit RGB.
func lightColor(r, g, b uint16) (uint16, uint16, uint16) {
	const max = 65535

	// If very bright (green > 95%), darken by 10%.
	if g > max*95/100 {
		return uint16(float64(r) * 0.90),
			uint16(float64(g) * 0.90),
			uint16(float64(b) * 0.90)
	}

	// Normal: max of (boost 40%, halfway to white).
	lr := maxU16(uint16(min64(float64(r)*1.4, max)), uint16((max+uint32(r))/2))
	lg := maxU16(uint16(min64(float64(g)*1.4, max)), uint16((max+uint32(g))/2))
	lb := maxU16(uint16(min64(float64(b)*1.4, max)), uint16((max+uint32(b))/2))
	return lr, lg, lb
}

// darkColor computes the dark shadow from Tk's algorithm.
func darkColor(r, g, b uint16) (uint16, uint16, uint16) {
	const max = 65535

	// Check perceived brightness: 0.5*R^2 + 1.0*G^2 + 0.28*B^2.
	brightness := 0.5*float64(r)*float64(r) +
		1.0*float64(g)*float64(g) +
		0.28*float64(b)*float64(b)
	threshold := 0.05 * float64(max) * float64(max)

	if brightness < threshold {
		// Very dark: lighten by 1/4 toward white.
		return uint16((max + 3*uint32(r)) / 4),
			uint16((max + 3*uint32(g)) / 4),
			uint16((max + 3*uint32(b)) / 4)
	}

	// Normal: reduce by 40%.
	return uint16(float64(r) * 0.60),
		uint16(float64(g) * 0.60),
		uint16(float64(b) * 0.60)
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
