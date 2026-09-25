package draw

import (
	"image"
	"math"

	"github.com/msorc/takigo/platform"
)

type ttkIndicatorKey struct {
	radio              bool
	state              IndicatorState
	size               int
	bg, fg, borderColr uint64
}

var ttkIndicatorCache = map[ttkIndicatorKey]*image.RGBA{}

// DrawTtkIndicator ports IndicatorElementDraw (tk/generic/ttk/ttkElements.c),
// the default theme's check/radio indicator: the checkbtn*/radiobtn* SVGs
// rendered at size x size pixels (16 * ::tk::scalingPct / 100) with bg as
// -indicatorbackground, fg as -indicatorforeground and borderColor as
// -bordercolor, drawn with its top-left corner at (x, y).
func DrawTtkIndicator(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	depth, x, y, size int, radio bool, state IndicatorState, bg, fg, borderColor, bgPixel uint64) {
	k := ttkIndicatorKey{radio, state, size, bg, fg, borderColor}
	indicatorMu.Lock()
	img, ok := ttkIndicatorCache[k]
	if !ok {
		img = rasterizeTtkIndicator(k)
		ttkIndicatorCache[k] = img
	}
	indicatorMu.Unlock()
	d.PutImageRGBA(drawable, gc, depth, img.Pix, img.Stride, size, size,
		0, 0, x, y, size, size, bgPixel)
}

func rasterizeTtkIndicator(k ttkIndicatorKey) *image.RGBA {
	solid := func(p uint64, inside func(x, y float64) bool) shape {
		c := pixelColor(p, 1)
		return func(x, y float64) (rgba, bool) { return c, inside(x, y) }
	}
	circle := func(r float64) func(x, y float64) bool {
		return func(x, y float64) bool { return (x-8)*(x-8)+(y-8)*(y-8) <= r*r }
	}
	// dash: path m4 8h8, stroke-width 2, butt caps.
	dash := func(x, y float64) bool { return x >= 4 && x <= 12 && y >= 7 && y <= 9 }
	var layers []shape
	switch {
	case !k.radio && k.state == IndicatorOff:
		// rect .5,.5 15x15 rx 3.5, stroke 1: the stroke spans the rounded
		// rects 0..16 (r 4) and 1..15 (r 3).
		layers = append(layers,
			solid(k.borderColr, roundRect(0, 0, 16, 16, 4)),
			solid(k.bg, roundRect(1, 1, 15, 15, 3)))
	case !k.radio:
		layers = append(layers, solid(k.bg, roundRect(0, 0, 16, 16, 4)))
		if k.state == IndicatorOn {
			layers = append(layers, solid(k.fg, stroke([][2]float64{{4.5, 8}, {7.5, 11}, {11.5, 5}}, 1)))
		} else {
			layers = append(layers, solid(k.fg, dash))
		}
	case k.state == IndicatorOff:
		layers = append(layers, solid(k.borderColr, circle(8)), solid(k.bg, circle(7)))
	default:
		layers = append(layers, solid(k.bg, circle(8)))
		if k.state == IndicatorOn {
			layers = append(layers, solid(k.fg, circle(3)))
		} else {
			layers = append(layers, solid(k.fg, dash))
		}
	}
	return rasterizeLayers(layers, k.size, 16/float64(k.size))
}

// roundRect returns a test for the rounded rectangle x0..x1, y0..y1 with
// corner radius r.
func roundRect(x0, y0, x1, y1, r float64) func(x, y float64) bool {
	return func(x, y float64) bool {
		if x < x0 || x > x1 || y < y0 || y > y1 {
			return false
		}
		cx := math.Max(x0+r, math.Min(x1-r, x))
		cy := math.Max(y0+r, math.Min(y1-r, y))
		return (x-cx)*(x-cx)+(y-cy)*(y-cy) <= r*r
	}
}

// rasterizeLayers composites layers (in 16-unit SVG space, unit = scale
// pixels^-1) into a size x size premultiplied image with 8x8 supersampling.
func rasterizeLayers(layers []shape, size int, unit float64) *image.RGBA {
	const ss = 8
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for py := range size {
		for px := range size {
			var acc rgba
			for sy := range ss {
				for sx := range ss {
					x := (float64(px) + (float64(sx)+0.5)/ss) * unit
					y := (float64(py) + (float64(sy)+0.5)/ss) * unit
					var c rgba
					for _, l := range layers {
						if s, ok := l(x, y); ok {
							c.r = s.r*s.a + c.r*(1-s.a)
							c.g = s.g*s.a + c.g*(1-s.a)
							c.b = s.b*s.a + c.b*(1-s.a)
							c.a = s.a + c.a*(1-s.a)
						}
					}
					acc.r += c.r
					acc.g += c.g
					acc.b += c.b
					acc.a += c.a
				}
			}
			n := float64(ss * ss)
			i := img.PixOffset(px, py)
			img.Pix[i+0] = uint8(math.Round(acc.r / n * 255))
			img.Pix[i+1] = uint8(math.Round(acc.g / n * 255))
			img.Pix[i+2] = uint8(math.Round(acc.b / n * 255))
			img.Pix[i+3] = uint8(math.Round(acc.a / n * 255))
		}
	}
	return img
}
