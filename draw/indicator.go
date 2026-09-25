package draw

import (
	"image"
	"math"
	"sync"

	"github.com/msorc/takigo/platform"
)

// IndicatorKind selects one of Tk 9's check/radio indicator images.
type IndicatorKind int

const (
	CheckIndicator IndicatorKind = iota
	RadioIndicator
)

// IndicatorDim is the indicator size at 100% scaling (CHECK_BUTTON_DIM and
// RADIO_BUTTON_DIM in tk/unix/tkUnixButton.c).
const IndicatorDim = 16

// IndicatorState is Tk's "on" value: 0 off, 1 on, 2 tri-state.
type IndicatorState int

const (
	IndicatorOff IndicatorState = iota
	IndicatorOn
	IndicatorTristate
)

type indicatorKey struct {
	kind                             IndicatorKind
	on                               bool
	dark, light, interior, indicator uint64
}

var (
	indicatorMu    sync.Mutex
	indicatorCache = map[indicatorKey]*image.RGBA{}
)

// DrawCheckIndicator ports TkpDrawCheckIndicator (tk/unix/tkUnixButton.c):
// Tk 9 draws check and radio indicators from small SVG images coloured with
// the border's shadows, the select colour and the indicator colour. It
// draws the image centred on (cx, cy). Colours are 0xRRGGBB pixels.
func DrawCheckIndicator(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	depth int, cx, cy int, kind IndicatorKind, border *Border,
	indicatorColor, selectColor, disabledColor uint64, state IndicatorState, disabled bool) {
	interior, mark := selectColor, indicatorColor
	if state == IndicatorTristate || disabled {
		interior, mark = border.BgPixel, disabledColor
	}
	img := indicatorImage(indicatorKey{
		kind: kind, on: state != IndicatorOff,
		dark: border.DarkPixel, light: border.LightPixel,
		interior: interior, indicator: mark,
	})
	x, y := cx-IndicatorDim/2, cy-IndicatorDim/2
	d.PutImageRGBA(drawable, gc, depth, img.Pix, img.Stride, IndicatorDim, IndicatorDim,
		0, 0, x, y, IndicatorDim, IndicatorDim, border.BgPixel)
}

func indicatorImage(k indicatorKey) *image.RGBA {
	indicatorMu.Lock()
	defer indicatorMu.Unlock()
	if img, ok := indicatorCache[k]; ok {
		return img
	}
	img := rasterizeIndicator(k)
	indicatorCache[k] = img
	return img
}

// rgba is a straight-alpha colour with float channels in [0,1].
type rgba struct{ r, g, b, a float64 }

func pixelColor(p uint64, alpha float64) rgba {
	return rgba{float64(p>>16&0xff) / 255, float64(p>>8&0xff) / 255, float64(p&0xff) / 255, alpha}
}

// shape returns the colour of one layer at a point, or ok=false outside it.
type shape func(x, y float64) (rgba, bool)

// rasterizeIndicator renders the SVG layers of checkbtnOffData,
// checkbtnOnData, radiobtnOffData and radiobtnOnData with 8x8 supersampling.
func rasterizeIndicator(k indicatorKey) *image.RGBA {
	var layers []shape
	solid := func(p uint64, inside func(x, y float64) bool) shape {
		c := pixelColor(p, 1)
		return func(x, y float64) (rgba, bool) { return c, inside(x, y) }
	}
	circle := func(r float64) func(x, y float64) bool {
		return func(x, y float64) bool { return (x-8)*(x-8)+(y-8)*(y-8) <= r*r }
	}
	switch k.kind {
	case CheckIndicator:
		// borderdark: m0 0v16l1-1v-14h14l1-1h-16z
		layers = append(layers, solid(k.dark, polygon([][2]float64{{0, 0}, {0, 16}, {1, 15}, {1, 1}, {15, 1}, {16, 0}})))
		// borderlight: m16 0-1 1v14h-14l-1 1h16v-16z
		layers = append(layers, solid(k.light, polygon([][2]float64{{16, 0}, {15, 1}, {15, 15}, {1, 15}, {0, 16}, {16, 16}})))
		layers = append(layers, solid(k.interior, func(x, y float64) bool { return x >= 2 && x < 14 && y >= 2 && y < 14 }))
		if k.on {
			// indicator: m4.5 8 3 3 4-6, stroke-width 2, round caps/joins.
			layers = append(layers, solid(k.indicator, stroke([][2]float64{{4.5, 8}, {7.5, 11}, {11.5, 5}}, 1)))
		}
	case RadioIndicator:
		dark, light := pixelColor(k.dark, 1), pixelColor(k.light, 0)
		outer := circle(8)
		layers = append(layers, func(x, y float64) (rgba, bool) {
			if !outer(x, y) {
				return rgba{}, false
			}
			// linearGradient (5,5)->(11,11), dark opaque -> light transparent.
			t := math.Max(0, math.Min(1, ((x-5)+(y-5))/12))
			return rgba{
				dark.r + (light.r-dark.r)*t,
				dark.g + (light.g-dark.g)*t,
				dark.b + (light.b-dark.b)*t,
				dark.a + (light.a-dark.a)*t,
			}, true
		})
		r := 6.5
		if k.on {
			r = 7
		}
		layers = append(layers, solid(k.interior, circle(r)))
		if k.on {
			layers = append(layers, solid(k.indicator, circle(4)))
		}
	}

	const ss = 8
	img := image.NewRGBA(image.Rect(0, 0, IndicatorDim, IndicatorDim))
	for py := range IndicatorDim {
		for px := range IndicatorDim {
			var acc rgba // premultiplied sum
			for sy := range ss {
				for sx := range ss {
					x := float64(px) + (float64(sx)+0.5)/ss
					y := float64(py) + (float64(sy)+0.5)/ss
					var c rgba // premultiplied "over" composite
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

// polygon returns an even-odd point-in-polygon test.
func polygon(pts [][2]float64) func(x, y float64) bool {
	return func(x, y float64) bool {
		in := false
		for i, j := 0, len(pts)-1; i < len(pts); j, i = i, i+1 {
			xi, yi, xj, yj := pts[i][0], pts[i][1], pts[j][0], pts[j][1]
			if (yi > y) != (yj > y) && x < (xj-xi)*(y-yi)/(yj-yi)+xi {
				in = !in
			}
		}
		return in
	}
}

// stroke returns a test for points within half-width hw of a polyline with
// round caps and joins.
func stroke(pts [][2]float64, hw float64) func(x, y float64) bool {
	return func(x, y float64) bool {
		for i := 0; i+1 < len(pts); i++ {
			ax, ay, bx, by := pts[i][0], pts[i][1], pts[i+1][0], pts[i+1][1]
			dx, dy := bx-ax, by-ay
			t := ((x-ax)*dx + (y-ay)*dy) / (dx*dx + dy*dy)
			t = math.Max(0, math.Min(1, t))
			ex, ey := x-(ax+t*dx), y-(ay+t*dy)
			if ex*ex+ey*ey <= hw*hw {
				return true
			}
		}
		return false
	}
}
