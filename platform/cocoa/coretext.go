//go:build darwin

package cocoa

import (
	"github.com/takigo/takigo/font"
	clib "github.com/takigo/takigo/internal/cocoa"
	"github.com/takigo/takigo/platform"
)

// CoreTextFont implements font.Font using Core Text via cgo.
type CoreTextFont struct {
	fid     clib.FontID
	attrs   font.Attributes
	metrics font.Metrics
}

func (f *CoreTextFont) Attrs() font.Attributes { return f.attrs }
func (f *CoreTextFont) Metrics() font.Metrics  { return f.metrics }

func (f *CoreTextFont) MeasureString(s string) int {
	return clib.MeasureString(f.fid, s)
}

func (f *CoreTextFont) Close() {
	if f.fid != 0 {
		clib.CloseFont(f.fid)
		f.fid = 0
	}
}

// DrawString implements platform.DrawableFont.
func (f *CoreTextFont) DrawString(drawable platform.DrawableID, x, y int, s string, pixel uint64, r, g, b uint16) {
	if len(s) == 0 || f.fid == 0 {
		return
	}
	clib.DrawString(clib.Drawable(uintptr(drawable)), f.fid, x, y, s, pixel, r, g, b)
}

// Verify interfaces at compile time.
var _ font.Font = (*CoreTextFont)(nil)
var _ platform.DrawableFont = (*CoreTextFont)(nil)
