package draw

import (
	"fmt"
	"strings"
	"sync"

	"github.com/msorc/takigo/internal/nanosvg"
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

// SVG data of tkUnixButton.c; DARKKK, LIGHTT, INTROR and INDCTR are replaced
// by colours.
const (
	checkbtnOffData = `<svg id='checkbutton' width='16' height='16' version='1.1' xmlns='http://www.w3.org/2000/svg'>
 <path id='borderdark' d='m0 0v16l1-1v-14h14l1-1h-16z' fill='#DARKKK'/>
 <path id='borderlight' d='m16 0-1 1v14h-14l-1 1h16v-16z' fill='#LIGHTT'/>
 <rect id='rectbackdrop' x='2' y='2' width='12' height='12' fill='#INTROR'/>
</svg>`
	checkbtnOnData = `<svg id='checkbutton' width='16' height='16' version='1.1' xmlns='http://www.w3.org/2000/svg'>
 <path id='borderdark' d='m0 0v16l1-1v-14h14l1-1h-16z' fill='#DARKKK'/>
 <path id='borderlight' d='m16 0-1 1v14h-14l-1 1h16v-16z' fill='#LIGHTT'/>
 <rect id='rectbackdrop' x='2' y='2' width='12' height='12' fill='#INTROR'/>
 <path id='indicator' d='m4.5 8 3 3 4-6' fill='none' stroke='#INDCTR' stroke-linecap='round' stroke-linejoin='round' stroke-width='2'/>
</svg>`
	radiobtnOffData = `<svg id='radiobutton' width='16' height='16' version='1.1' xmlns='http://www.w3.org/2000/svg'>
 <defs>
  <linearGradient id='gradient' x1='5' y1='5' x2='11' y2='11' gradientUnits='userSpaceOnUse'>
   <stop stop-color='#DARKKK' offset='0'/>
   <stop stop-color='#LIGHTT' offset='1' stop-opacity='0'/>
  </linearGradient>
 </defs>
 <circle cx='8' cy='8' r='8' fill='url(#gradient)'/>
 <circle cx='8' cy='8' r='6.5' fill='#INTROR'/>
</svg>`
	radiobtnOnData = `<svg id='radiobutton' width='16' height='16' version='1.1' xmlns='http://www.w3.org/2000/svg'>
 <defs>
  <linearGradient id='gradient' x1='5' y1='5' x2='11' y2='11' gradientUnits='userSpaceOnUse'>
   <stop stop-color='#DARKKK' offset='0'/>
   <stop stop-color='#LIGHTT' offset='1' stop-opacity='0'/>
  </linearGradient>
 </defs>
 <circle cx='8' cy='8' r='8' fill='url(#gradient)'/>
 <circle cx='8' cy='8' r='7' fill='#INTROR'/>
 <circle cx='8' cy='8' r='4' fill='#INDCTR'/>
</svg>`
)

type svgKey struct {
	data  string
	scale float32
}

type svgEntry struct {
	px   []uint8
	w, h int
}

var (
	indicatorMu sync.Mutex
	svgCache    = map[svgKey]svgEntry{}
)

func colorStr(p uint64) string { return fmt.Sprintf("%06x", p&0xffffff) }

// SVGImage renders SVG data like `image create photo -format {svg -scale
// scale}`: straight-alpha RGBA and its size. Results are cached.
func SVGImage(data string, scale float32) ([]uint8, int, int) {
	k := svgKey{data: data, scale: scale}
	indicatorMu.Lock()
	defer indicatorMu.Unlock()
	if e, ok := svgCache[k]; ok {
		return e.px, e.w, e.h
	}
	img, _ := nanosvg.Parse(data)
	w, h := nanosvg.Size(img, scale)
	e := svgEntry{nanosvg.Rasterize(img, scale, w, h), w, h}
	svgCache[k] = e
	return e.px, e.w, e.h
}

// PutPhotoRGBA draws straight-alpha RGBA at (x, y) the way a Tk photo
// instance is drawn: partly transparent pixels are blended over what the
// drawable already holds when readBack is set (use it for off-screen
// pixmaps only: obscured window contents are undefined), else over bg.
func PutPhotoRGBA(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	depth, x, y int, px []uint8, w, h int, bg uint64, readBack bool) {
	var under []uint8
	if readBack {
		under = d.GetImageRGBA(drawable, x, y, w, h)
	}
	if under == nil {
		under = nanosvg.BlendOver(px, bg)
	} else {
		nanosvg.Blend(under, w, 0, 0, px, w, h)
	}
	d.PutImageRGBA(drawable, gc, depth, under, w*4, w, h, 0, 0, x, y, w, h, bg)
}

// putSVG draws one of the 16x16 indicator SVGs scaled to dim x dim pixels
// at (x, y).
func putSVG(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	depth, x, y, dim int, data string, bg uint64, readBack bool) {
	px, w, h := SVGImage(data, float32(dim)/16)
	PutPhotoRGBA(d, drawable, gc, depth, x, y, px, w, h, bg, readBack)
}

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
	data := checkbtnOffData
	switch {
	case kind == CheckIndicator && state != IndicatorOff:
		data = checkbtnOnData
	case kind == RadioIndicator && state == IndicatorOff:
		data = radiobtnOffData
	case kind == RadioIndicator:
		data = radiobtnOnData
	}
	data = strings.NewReplacer(
		"DARKKK", colorStr(border.DarkPixel), "LIGHTT", colorStr(border.LightPixel),
		"INTROR", colorStr(interior), "INDCTR", colorStr(mark)).Replace(data)
	putSVG(d, drawable, gc, depth, cx-IndicatorDim/2, cy-IndicatorDim/2, IndicatorDim, data, border.BgPixel, false)
}
