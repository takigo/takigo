//go:build darwin

package cocoa

import (
	"fmt"

	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/font/coretext"
	clib "github.com/msorc/takigo/internal/cocoa"
)

// FontOpener implements font.FontOpener using Core Text.
type FontOpener struct{}

// OpenFont opens a font using Core Text via the cocoa C bindings.
func (o *FontOpener) OpenFont(attrs font.Attributes) (font.Font, error) {
	return openCoreTextFont(attrs)
}

// Families lists the Core Text font families.
func (o *FontOpener) Families() []string {
	return coretext.ListFamilies()
}

// Verify at compile time.
var _ font.FontOpener = (*FontOpener)(nil)

// openCoreTextFont opens a font using Core Text and returns a CoreTextFont.
func openCoreTextFont(attrs font.Attributes) (*CoreTextFont, error) {
	family := attrs.Family
	if family == "" {
		family = "sans-serif"
	}

	size := attrs.Size
	if size == 0 {
		size = 12
	}
	if size < 0 {
		size = -size // pixel size on macOS is same as point size at 72 DPI
	}

	weight := 0
	if attrs.Weight == font.WeightBold {
		weight = 1
	}

	slant := 0
	switch attrs.Slant {
	case font.SlantItalic:
		slant = 1
	case font.SlantOblique:
		slant = 2
	}

	fid := clib.OpenFont(family, size, weight, slant)
	if fid == 0 {
		// Try with fallback family
		fid = clib.OpenFont("Helvetica Neue", size, weight, slant)
		if fid == 0 {
			return nil, fmt.Errorf("font: failed to open %q", family)
		}
	}

	f := &CoreTextFont{
		fid:   fid,
		attrs: attrs,
		metrics: font.Metrics{
			Ascent:   clib.FontAscent(fid),
			Descent:  clib.FontDescent(fid),
			MaxWidth: clib.FontMaxWidth(fid),
			Fixed:    clib.FontIsFixed(fid),
		},
	}
	return f, nil
}
