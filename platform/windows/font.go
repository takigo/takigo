//go:build windows

package windows

import (
	"github.com/msorc/takigo/font"
	w32 "github.com/msorc/takigo/internal/win32"
)

// FontOpener implements font.FontOpener using GDI on Windows.
type FontOpener struct {
	screenDC  w32.HDC
	resolveDC font.DCResolver
}

// OpenFont opens a font using GDI's CreateFontIndirectW.
func (o *FontOpener) OpenFont(attrs font.Attributes) (font.Font, error) {
	return font.OpenGDI(w32.HDC(o.screenDC), attrs, o.resolveDC)
}

// Verify at compile time.
var _ font.FontOpener = (*FontOpener)(nil)
