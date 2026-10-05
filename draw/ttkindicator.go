package draw

import (
	"strings"

	"github.com/takigo/takigo/platform"
)

// SVG data of ttkElements.c (checkbutton_spec, radiobutton_spec, sliderData).
const (
	ttkCheckOffData = `<svg width='16' height='16' version='1.1' xmlns='http://www.w3.org/2000/svg'>
 <rect x='.5' y='.5' width='15' height='15' rx='3.5' fill='#ffffff' stroke='#888888'/>
</svg>`
	ttkCheckOnData = `<svg width='16' height='16' version='1.1' xmlns='http://www.w3.org/2000/svg'>
 <rect x='0' y='0' width='16' height='16' fill='#4a6984' rx='4'/>
 <path d='m4.5 8 3 3 4-6' fill='none' stroke='#ffffff' stroke-linecap='round' stroke-linejoin='round' stroke-width='2'/>
</svg>`
	ttkCheckTriData = `<svg width='16' height='16' version='1.1' xmlns='http://www.w3.org/2000/svg'>
 <rect x='0' y='0' width='16' height='16' fill='#4a6984' rx='4'/>
 <path d='m4 8h8' fill='none' stroke='#ffffff' stroke-width='2'/>
</svg>`
	ttkRadioOffData = `<svg width='16' height='16' version='1.1' xmlns='http://www.w3.org/2000/svg'>
 <circle cx='8' cy='8' r='7.5' fill='#ffffff' stroke='#888888'/>
</svg>`
	ttkRadioOnData = `<svg width='16' height='16' version='1.1' xmlns='http://www.w3.org/2000/svg'>
 <circle cx='8' cy='8' r='8' fill='#4a6984'/>
 <circle cx='8' cy='8' r='3' fill='#ffffff'/>
</svg>`
	ttkRadioTriData = `<svg width='16' height='16' version='1.1' xmlns='http://www.w3.org/2000/svg'>
 <circle cx='8' cy='8' r='8' fill='#4a6984'/>
 <path d='m4 8h8' fill='none' stroke='#ffffff' stroke-width='2'/>
</svg>`
	ttkSliderData = `<svg width='16' height='16' version='1.1' xmlns='http://www.w3.org/2000/svg'>
 <circle cx='8' cy='8' r='7.5' fill='#ffffff' stroke='#c3c3c3'/>
 <circle cx='8' cy='8' r='4' fill='#4a6984'/>
</svg>`
)

// replaceFirst substitutes the first occurrence of each old string, as the
// strstr/memcpy colour patching in ttkElements.c does.
func replaceFirst(s string, pairs ...string) string {
	for i := 0; i+1 < len(pairs); i += 2 {
		s = strings.Replace(s, pairs[i], pairs[i+1], 1)
	}
	return s
}

// DrawTtkIndicator ports IndicatorElementDraw (tk/generic/ttk/ttkElements.c),
// the default theme's check/radio indicator: the checkbtn*/radiobtn* SVGs
// rendered at size x size pixels (16 * ::tk::scalingPct / 100) with bg as
// -indicatorbackground, fg as -indicatorforeground and borderColor as
// -bordercolor, drawn with its top-left corner at (x, y) over bgPixel.
func DrawTtkIndicator(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	depth, x, y, size int, radio bool, state IndicatorState, bg, fg, borderColor, bgPixel uint64) {
	var data string
	switch {
	case !radio && state == IndicatorOff:
		data = ttkCheckOffData
	case !radio && state == IndicatorOn:
		data = ttkCheckOnData
	case !radio:
		data = ttkCheckTriData
	case state == IndicatorOff:
		data = ttkRadioOffData
	case state == IndicatorOn:
		data = ttkRadioOnData
	default:
		data = ttkRadioTriData
	}
	if state == IndicatorOff {
		data = replaceFirst(data, "ffffff", colorStr(bg), "888888", colorStr(borderColor))
	} else {
		data = replaceFirst(data, "4a6984", colorStr(bg), "ffffff", colorStr(fg))
	}
	putSVG(d, drawable, gc, depth, x, y, size, data, bgPixel, false)
}

// DrawTtkSlider ports SliderElementDraw (ttkElements.c): the sliderData SVG
// in -innercolor, -outercolor and -bordercolor at size x size pixels with its
// top-left corner at (x, y), blended over the trough already drawn in the
// off-screen drawable.
func DrawTtkSlider(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	depth, x, y, size int, inner, outer, border, bgPixel uint64) {
	data := replaceFirst(ttkSliderData, "4a6984", colorStr(inner),
		"ffffff", colorStr(outer), "c3c3c3", colorStr(border))
	putSVG(d, drawable, gc, depth, x, y, size, data, bgPixel, true)
}
