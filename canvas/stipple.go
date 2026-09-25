package canvas

import (
	"github.com/msorc/takigo/bitmap"
	"github.com/msorc/takigo/platform"
)

// stippleOn switches gc to FillStippled with the bitmap spec (a built-in
// name or "@file.xbm") anchored at the canvas origin, as
// Tk_CanvasSetStippleOrigin does. It returns a func that restores solid
// fills; with an empty or unknown spec it does nothing.
func (c *Canvas) stippleOn(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	spec string, originX, originY int) func() {
	if spec == "" || c == nil {
		return func() {}
	}
	pm, ok := c.stipples[spec]
	if !ok {
		if w, h, bits, ok := bitmap.Data(spec); ok {
			pm = d.CreateBitmapFromData(drawable, bits, uint(w), uint(h))
		}
		if c.stipples == nil {
			c.stipples = map[string]platform.PixmapID{}
		}
		c.stipples[spec] = pm
	}
	if pm == 0 {
		return func() {}
	}
	d.SetStipple(gc, pm)
	d.SetFillStyle(gc, platform.FillStippled)
	d.SetTSOrigin(gc, -originX, -originY)
	return func() { d.SetFillStyle(gc, platform.FillSolid) }
}

// Stipple sets -stipple (fill stipple; the line stipple for line items).
func Stipple(spec string) ItemOption {
	return func(_ *Canvas, item Item) error {
		if b := itemBase(item); b != nil {
			b.stipple = spec
		}
		return nil
	}
}

// OutlineStipple sets -outlinestipple.
func OutlineStipple(spec string) ItemOption {
	return func(_ *Canvas, item Item) error {
		if b := itemBase(item); b != nil {
			b.outlineStipple = spec
		}
		return nil
	}
}
