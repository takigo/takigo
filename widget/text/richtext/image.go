package richtext

import (
	goimage "image"
	"math"

	"github.com/takigo/takigo/image"
	"github.com/takigo/takigo/widget"
)

// rgbaImage is an image whose pixels can be read for scaling.
type rgbaImage interface {
	widget.WidgetImage
	RGBA() *goimage.RGBA
}

// imageSize works out the size an image is shown at from its natural size,
// the CSS width, height and max-width, and the width available (0 when
// unknown).
func imageSize(w0, h0 int, s *style, avail int, fit bool) (int, int) {
	if w0 <= 0 || h0 <= 0 {
		return w0, h0
	}
	resolve := func(l length) (float64, bool) {
		switch {
		case !l.set:
			return 0, false
		case l.percent > 0 && avail > 0:
			return l.percent * float64(avail) / 100, true
		case l.percent > 0:
			return 0, false
		}
		return l.px, l.px > 0
	}
	w, h := float64(w0), float64(h0)
	cw, wok := resolve(s.width)
	ch, hok := resolve(s.height)
	switch {
	case wok && hok:
		w, h = cw, ch
	case wok:
		w, h = cw, cw*float64(h0)/float64(w0)
	case hok:
		w, h = ch*float64(w0)/float64(h0), ch
	}
	maxW, ok := resolve(s.maxWidth)
	if !ok && fit && avail > 0 {
		maxW, ok = float64(avail), true
	}
	if ok && w > maxW {
		h, w = h*maxW/w, maxW
	}
	return max(int(math.Round(w)), 1), max(int(math.Round(h)), 1)
}

// fitImage returns img at the size the style asks for; an image whose
// pixels cannot be read keeps its size.
func fitImage(img widget.WidgetImage, s *style, avail int, fit bool) widget.WidgetImage {
	w0, h0 := img.Width(), img.Height()
	w, h := imageSize(w0, h0, s, avail, fit)
	if w == w0 && h == h0 {
		return img
	}
	src, ok := img.(rgbaImage)
	if !ok || src.RGBA() == nil {
		return img
	}
	name := ""
	if n, ok := img.(interface{ Name() string }); ok {
		name = n.Name()
	}
	return image.NewPhoto(name, scaleRGBA(src.RGBA(), w, h))
}

// scaleRGBA resamples premultiplied RGBA: each destination pixel averages
// the source area it covers (with fractional edges) when shrinking, and
// interpolates bilinearly when enlarging.
func scaleRGBA(src *goimage.RGBA, w, h int) *goimage.RGBA {
	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()
	dst := goimage.NewRGBA(goimage.Rect(0, 0, w, h))
	if sw == 0 || sh == 0 {
		return dst
	}
	at := func(x, y int) []uint8 {
		o := src.PixOffset(sb.Min.X+x, sb.Min.Y+y)
		return src.Pix[o : o+4]
	}
	fx, fy := float64(sw)/float64(w), float64(sh)/float64(h)
	for y := range h {
		for x := range w {
			var acc [4]float64
			if fx <= 1 && fy <= 1 {
				// Bilinear, sampling at the pixel centre.
				sx := max((float64(x)+0.5)*fx-0.5, 0)
				sy := max((float64(y)+0.5)*fy-0.5, 0)
				x0, y0 := int(sx), int(sy)
				x1, y1 := min(x0+1, sw-1), min(y0+1, sh-1)
				ax, ay := sx-float64(x0), sy-float64(y0)
				for c := range 4 {
					top := float64(at(x0, y0)[c])*(1-ax) + float64(at(x1, y0)[c])*ax
					bot := float64(at(x0, y1)[c])*(1-ax) + float64(at(x1, y1)[c])*ax
					acc[c] = top*(1-ay) + bot*ay
				}
			} else {
				// Area average over [x0, x1) × [y0, y1) in source pixels.
				x0, x1 := float64(x)*fx, float64(x+1)*fx
				y0, y1 := float64(y)*fy, float64(y+1)*fy
				var total float64
				for sy := int(y0); sy < int(math.Ceil(y1)) && sy < sh; sy++ {
					wy := math.Min(y1, float64(sy+1)) - math.Max(y0, float64(sy))
					for sx := int(x0); sx < int(math.Ceil(x1)) && sx < sw; sx++ {
						wx := math.Min(x1, float64(sx+1)) - math.Max(x0, float64(sx))
						wgt := wx * wy
						p := at(sx, sy)
						for c := range 4 {
							acc[c] += float64(p[c]) * wgt
						}
						total += wgt
					}
				}
				if total > 0 {
					for c := range 4 {
						acc[c] /= total
					}
				}
			}
			o := dst.PixOffset(x, y)
			for c := range 4 {
				dst.Pix[o+c] = uint8(math.Round(min(max(acc[c], 0), 255)))
			}
		}
	}
	return dst
}
