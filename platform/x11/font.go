//go:build linux || freebsd || openbsd || netbsd

package x11

import (
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/font/xft"
	"github.com/msorc/takigo/internal/xlib"
)

// X11FontOpener implements font.FontOpener using Xft/fontconfig.
type X11FontOpener struct {
	display  *xlib.Display
	screen   int
	visual   *xlib.Visual
	colormap xlib.Colormap
}

// NewFontOpener creates a new X11 font opener.
func NewFontOpener(display *xlib.Display, screen int, visual *xlib.Visual, colormap xlib.Colormap) *X11FontOpener {
	return &X11FontOpener{
		display:  display,
		screen:   screen,
		visual:   visual,
		colormap: colormap,
	}
}

// OpenFont opens a font via Xft/fontconfig.
func (o *X11FontOpener) OpenFont(attrs font.Attributes) (font.Font, error) {
	return xft.OpenXft(o.display, o.screen, o.visual, o.colormap, attrs)
}

// Families lists the fontconfig font families.
func (o *X11FontOpener) Families() []string {
	return xft.ListFamilies()
}

// Verify at compile time.
var _ font.FontOpener = (*X11FontOpener)(nil)
