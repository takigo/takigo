package canvas

import (
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/widget"
)

// ImageItem implements a positioned image canvas item.
type ImageItem struct {
	ItemBase
	x, y   float64
	image  widget.WidgetImage
	anchor option.Anchor
}

func newImageItem(x, y float64, c *Canvas) *ImageItem {
	item := &ImageItem{
		x:      x,
		y:      y,
		anchor: option.AnchorCenter,
	}
	item.canvas = c
	item.updateBBox()
	return item
}

func (im *ImageItem) base() *ItemBase { return &im.ItemBase }
func (im *ImageItem) Type() string    { return "image" }

func (im *ImageItem) BBox() (x1, y1, x2, y2 int) {
	return im.X1, im.Y1, im.X2, im.Y2
}

func (im *ImageItem) Coords() []float64 {
	return []float64{im.x, im.y}
}

func (im *ImageItem) SetCoords(coords []float64) error {
	if len(coords) >= 2 {
		im.x = coords[0]
		im.y = coords[1]
		im.updateBBox()
	}
	return nil
}

func (im *ImageItem) Configure(opts []ItemOption) error {
	c := im.canvas
	for _, opt := range opts {
		if err := opt(c, im); err != nil {
			return err
		}
	}
	im.updateBBox()
	return nil
}

func (im *ImageItem) updateBBox() {
	if im.image == nil {
		im.X1 = int(im.x)
		im.Y1 = int(im.y)
		im.X2 = int(im.x)
		im.Y2 = int(im.y)
		return
	}

	w := im.image.Width()
	h := im.image.Height()
	ax, ay := anchorOffset(im.anchor, w, h)
	im.X1 = int(im.x) + ax
	im.Y1 = int(im.y) + ay
	im.X2 = im.X1 + w
	im.Y2 = im.Y1 + h
}

func (im *ImageItem) Display(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	clipX, clipY, clipW, clipH, originX, originY int) {

	if im.image == nil {
		return
	}

	w := im.image.Width()
	h := im.image.Height()
	ax, ay := anchorOffset(im.anchor, w, h)
	drawX := drawableCoord(im.x, originX) + ax
	drawY := drawableCoord(im.y, originY) + ay

	// Get depth/bgPixel from the canvas window.
	win := im.canvas.Win
	bgPixel := win.BackgroundPixel
	if im.canvas.Background != nil {
		bgPixel = im.canvas.Background.Pixel
	}

	im.image.Draw(d, drawable, gc,
		win.Depth,
		0, 0, w, h, drawX, drawY, bgPixel)
}

func (im *ImageItem) PointDistance(x, y float64) float64 {
	return rectPointDistance(x, y, float64(im.X1), float64(im.Y1), float64(im.X2), float64(im.Y2))
}

func (im *ImageItem) AreaOverlap(ax1, ay1, ax2, ay2 float64) int {
	x1, y1, x2, y2 := float64(im.X1), float64(im.Y1), float64(im.X2), float64(im.Y2)
	if ax2 < x1 || ax1 > x2 || ay2 < y1 || ay1 > y2 {
		return -1
	}
	if ax1 <= x1 && ax2 >= x2 && ay1 <= y1 && ay2 >= y2 {
		return 1
	}
	return 0
}

func (im *ImageItem) Scale(ox, oy, sx, sy float64) {
	im.x = ox + (im.x-ox)*sx
	im.y = oy + (im.y-oy)*sy
	im.updateBBox()
}

func (im *ImageItem) Translate(dx, dy float64) {
	im.x += dx
	im.y += dy
	im.updateBBox()
}

func (im *ImageItem) Delete(d platform.DisplayServer) {}

// Postscript emits a byte-for-byte 24-bit RGB dump of the image as
// PostScript `image` data. Mirrors tk/generic/tkCanvImg.c:ImageToPostscript +
// tk/generic/tkImgBmap.c:Tk_PostscriptImage. Output is large (no compression)
// but matches Tk's wire format exactly.
func (im *ImageItem) Postscript(ps *PSContext) error {
	if im.State() == ItemStateHidden || im.image == nil {
		return nil
	}
	if ps.Prepass {
		return nil
	}
	w := im.image.Width()
	h := im.image.Height()
	if w == 0 || h == 0 {
		return nil
	}
	ax, ay := anchorOffset(im.anchor, w, h)
	x := float64(int(im.x) + ax)
	y := float64(int(im.y) + ay)

	// Dump the image as RGBA bytes via the WidgetImage interface. We then
	// convert to 3-byte RGB for PostScript `<< /BitsPerComponent 8 /Decode [0 1 0 1 0 1] >> image`.
	rgba := im.dumpRGBA()
	if len(rgba) != w*h*4 {
		return nil
	}
	rgb := make([]byte, w*h*3)
	for i := 0; i < w*h; i++ {
		rgb[i*3+0] = rgba[i*4+0]
		rgb[i*3+1] = rgba[i*4+1]
		rgb[i*3+2] = rgba[i*4+2]
	}

	hex := make([]byte, len(rgb)*2)
	for i, b := range rgb {
		hex[i*2+0] = hexDigit[b>>4]
		hex[i*2+1] = hexDigit[b&0xF]
	}

	ps.writef("gsave %.15g %.15g translate %d %d scale\n", x, ps.PsY(int(y)+h), w, h)
	ps.writef("<< /ImageType 1 /Width %d /Height %d /BitsPerComponent 8 /Decode [0 1 0 1 0 1] /DataSource <%s> >> image\n",
		w, h, string(hex))
	ps.write("grestore newpath\n")
	return nil
}

// dumpRGBA returns the image's RGBA pixels as a flat byte slice.
// Returns nil if the image doesn't expose its pixels (e.g. SVG).
func (im *ImageItem) dumpRGBA() []byte {
	type pixelSource interface {
		Pixels() []byte
	}
	if r, ok := im.image.(pixelSource); ok {
		return r.Pixels()
	}
	return nil
}
