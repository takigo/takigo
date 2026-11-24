package dialog

import (
	"fmt"
	"strings"

	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/ttk"
	"github.com/takigo/takigo/widget"

	tkimage "github.com/takigo/takigo/image"
)

// fileIcons are the 16px pictograms of the file dialog: SVG rasterized at
// the screen's scaling, as tkfbox.tcl does with its folder, file and updir
// images. The toolbar arrows and the home glyph are line art in the
// theme's foreground (tkfbox's UpdateUpdirImageData tints its updir image
// the same way), so they follow a dark theme.
type fileIcons struct {
	folder, file, up, home, newFolder, drive *tkimage.Photo
	// back and forward, then their variants in the disabled foreground.
	back, forward, backOff, forwardOff *tkimage.Photo
}

func newFileIcons(app widget.AppContext) *fileIcons {
	fg, off := glyphColors(app)
	return &fileIcons{
		folder:     svgIcon(svgFolder, ""),
		file:       svgIcon(svgFile, ""),
		up:         svgIcon(svgUp, fg),
		back:       svgIcon(svgBack, fg),
		forward:    svgIcon(svgForward, fg),
		backOff:    svgIcon(svgBack, off),
		forwardOff: svgIcon(svgForward, off),
		home:       svgIcon(svgHome, fg),
		newFolder:  svgIcon(svgNewFolder, ""),
		drive:      svgIcon(svgDrive, ""),
	}
}

func (ic *fileIcons) destroy() {
	for _, p := range []*tkimage.Photo{ic.folder, ic.file, ic.up, ic.back, ic.forward, ic.backOff, ic.forwardOff, ic.home, ic.newFolder, ic.drive} {
		p.Destroy()
	}
}

// glyphColors returns the normal and disabled foreground of the root ttk
// style as "#rrggbb".
func glyphColors(app widget.AppContext) (fg, off string) {
	st := ttk.ThemeFor(app).ResolveStyle(".")
	hex := func(p uint64) string { return fmt.Sprintf("#%06x", p&0xffffff) }
	return hex(ttk.LookupColor(st, "-foreground", 0, 0x000000)),
		hex(ttk.LookupColor(st, "-foreground", ttk.StateDisabled, 0xa3a3a3))
}

// svgIcon rasterizes src with "currentColor" replaced by ink, at Tk's
// ::tk::svgFmt scale (ScalingPct / 100).
func svgIcon(src, ink string) *tkimage.Photo {
	if ink != "" {
		src = strings.ReplaceAll(src, "currentColor", ink)
	}
	p, err := tkimage.NewPhotoFromSVG("", src, float64(screenunit.ScalingPct())/100)
	if err != nil {
		panic("dialog: built-in icon: " + err.Error())
	}
	return p
}

const svg16 = `<svg width="16" height="16" version="1.1" xmlns="http://www.w3.org/2000/svg">`

// glyph wraps line art drawn with a 1.5px round-capped currentColor stroke.
func glyph(paths string) string {
	return svg16 + `<g fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">` +
		paths + `</g></svg>`
}

var (
	svgBack    = glyph(`<path d="m13 8h-9.5m4-4.5-4.5 4.5 4.5 4.5"/>`)
	svgForward = glyph(`<path d="m3 8h9.5m-4-4.5 4.5 4.5-4.5 4.5"/>`)
	svgUp      = glyph(`<path d="m8 13v-9.5m-4.5 4 4.5-4.5 4.5 4.5"/>`)
	svgHome    = glyph(`<path d="m1.75 7.75 6.25-5.5 6.25 5.5"/><path d="m3.75 6.5v7.25h3v-3.5h2.5v3.5h3v-7.25"/>`)
)

// folderShape is a folder with its tab behind a lighter front panel.
const folderShape = `<linearGradient id="front" x1="0" y1="6" x2="0" y2="14" gradientUnits="userSpaceOnUse">` +
	`<stop offset="0" stop-color="#ffd772"/><stop offset="1" stop-color="#f3b13b"/></linearGradient>` +
	`<path d="m1 3.5c0-0.83 0.67-1.5 1.5-1.5h3.38c0.4 0 0.78 0.16 1.06 0.44l1.12 1.12c0.28 0.28 0.66 0.44 1.06 0.44h4.38c0.83 0 1.5 0.67 1.5 1.5v7c0 0.83-0.67 1.5-1.5 1.5h-11c-0.83 0-1.5-0.67-1.5-1.5z" fill="#d68f22"/>` +
	`<rect x="1" y="5.5" width="14" height="8.5" rx="1.5" fill="url(#front)"/>` +
	`<path d="m2.5 6.4h11" stroke="#fff3cf" stroke-width="0.8" stroke-linecap="round"/>`

var (
	svgFolder = svg16 + folderShape + `</svg>`

	svgNewFolder = svg16 + folderShape +
		`<circle cx="12" cy="11.5" r="3.75" fill="#2f9e44" stroke="#ffffff"/>` +
		`<path d="m12 9.75v3.5m-1.75-1.75h3.5" stroke="#ffffff" stroke-width="1.25" stroke-linecap="round"/></svg>`

	svgFile = svg16 +
		`<path d="m3.5 1.5h6l3 3v10h-9z" fill="#fdfdfd" stroke="#8a94a1" stroke-linejoin="round"/>` +
		`<path d="m9.5 1.5v3h3z" fill="#e2e6eb" stroke="#8a94a1" stroke-linejoin="round"/>` +
		`<path d="m5.5 7.5h5m-5 2h5m-5 2h3" stroke="#b4bcc6"/></svg>`

	svgDrive = svg16 +
		`<linearGradient id="body" x1="0" y1="4" x2="0" y2="12" gradientUnits="userSpaceOnUse">` +
		`<stop offset="0" stop-color="#eef1f4"/><stop offset="1" stop-color="#c2c9d1"/></linearGradient>` +
		`<rect x="1.5" y="4.5" width="13" height="7" rx="1.5" fill="url(#body)" stroke="#7a8592"/>` +
		`<path d="m3.5 9.5h6" stroke="#7a8592" stroke-linecap="round"/>` +
		`<circle cx="12" cy="9.5" r="1" fill="#30c552"/></svg>`
)
