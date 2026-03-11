package text

import (
	"github.com/msorc/takigo/platform"
)

// Named stipple bitmaps (8×8, XBM LSB-first byte order).
// Each byte encodes 8 pixels; bit 0 = leftmost pixel.
var stippleBitmaps = map[string]struct {
	bits          []byte
	width, height uint
}{
	// gray12: ~12.5% fill — sparse dot pattern (2 pixels per 8-row cycle).
	"gray12": {
		bits: []byte{
			0x22, 0x00, 0x00, 0x00,
			0x88, 0x00, 0x00, 0x00,
		},
		width: 8, height: 8,
	},
	// gray25: ~25% fill — checkerboard at 1-in-4 density.
	"gray25": {
		bits: []byte{
			0x55, 0x00, 0xAA, 0x00,
			0x55, 0x00, 0xAA, 0x00,
		},
		width: 8, height: 8,
	},
	// gray50: 50% fill — standard checkerboard.
	"gray50": {
		bits: []byte{
			0x55, 0xAA, 0x55, 0xAA,
			0x55, 0xAA, 0x55, 0xAA,
		},
		width: 8, height: 8,
	},
	// gray75: ~75% fill — inverse of gray25.
	"gray75": {
		bits: []byte{
			0xAA, 0xFF, 0x55, 0xFF,
			0xAA, 0xFF, 0x55, 0xFF,
		},
		width: 8, height: 8,
	},
}

// stipplePixmap returns a cached depth-1 Pixmap for the named stipple pattern,
// creating it on first use. Returns 0 if the name is unknown.
func (t *TextWidget) stipplePixmap(name string) platform.PixmapID {
	if t.stippleCache == nil {
		t.stippleCache = make(map[string]platform.PixmapID)
	}
	if pm, ok := t.stippleCache[name]; ok {
		return pm
	}
	def, ok := stippleBitmaps[name]
	if !ok {
		return 0
	}
	d := t.Win.Display.Server
	drawable := platform.WindowDrawable(t.Win.PlatformID)
	pm := d.CreateBitmapFromData(drawable, def.bits, def.width, def.height)
	t.stippleCache[name] = pm
	return pm
}
