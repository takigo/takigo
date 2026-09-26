package text

import (
	"github.com/msorc/takigo/bitmap"
	"github.com/msorc/takigo/platform"
)

// stipplePixmap returns a cached depth-1 Pixmap for the named stipple pattern,
// creating it on first use. Returns 0 if the name is unknown.
func (t *TextWidget) stipplePixmap(name string) platform.PixmapID {
	if t.stippleCache == nil {
		t.stippleCache = make(map[string]platform.PixmapID)
	}
	if pm, ok := t.stippleCache[name]; ok {
		return pm
	}
	// Tk's own bitmaps (gray12 is 16x16, etc.), as -bgstipple resolves them.
	w, h, bits, ok := bitmap.Data(name)
	if !ok {
		return 0
	}
	d := t.Win.Display.Server
	drawable := platform.WindowDrawable(t.Win.PlatformID)
	pm := d.CreateBitmapFromData(drawable, bits, uint(w), uint(h))
	t.stippleCache[name] = pm
	return pm
}
