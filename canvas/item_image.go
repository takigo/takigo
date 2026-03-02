package canvas

import (
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
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
	item.ItemBase.canvas = c
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

func (im *ImageItem) Display(d *xlib.Display, drawable xlib.Drawable, gc xlib.GC,
	clipX, clipY, clipW, clipH, originX, originY int) {

	if im.image == nil {
		return
	}

	w := im.image.Width()
	h := im.image.Height()
	ax, ay := anchorOffset(im.anchor, w, h)
	drawX := int(im.x) + ax - originX
	drawY := int(im.y) + ay - originY

	// Get visual/depth/bgPixel from the canvas window.
	win := im.canvas.Win
	bgPixel := win.BackgroundPixel
	if im.canvas.Base.Background != nil {
		bgPixel = im.canvas.Base.Background.Pixel
	}

	im.image.Draw(d, drawable, gc,
		win.Visual, win.Depth,
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

func (im *ImageItem) Delete(d *xlib.Display) {}
