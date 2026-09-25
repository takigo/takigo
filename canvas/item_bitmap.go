package canvas

import (
	"image/color"

	colortakigo "github.com/msorc/takigo/color"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
)

// BitmapItem displays a 1-bit XBM bitmap on the canvas.
// Pixels with bit=1 are drawn with Foreground; bit=0 pixels use Background
// (transparent by default).
type BitmapItem struct {
	ItemBase
	x, y       float64
	anchor     option.Anchor
	xbm        *XBMData
	Foreground color.RGBA // default: black
	Background color.RGBA // default: transparent (A=0)
}

func newBitmapItem(x, y float64, xbm *XBMData, c *Canvas) *BitmapItem {
	item := &BitmapItem{
		x:          x,
		y:          y,
		xbm:        xbm,
		anchor:     option.AnchorCenter,
		Foreground: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		Background: color.RGBA{A: 0}, // transparent
	}
	item.ItemBase.canvas = c
	item.updateBBox()
	return item
}

func (bi *BitmapItem) base() *ItemBase { return &bi.ItemBase }
func (bi *BitmapItem) Type() string    { return "bitmap" }

func (bi *BitmapItem) BBox() (x1, y1, x2, y2 int) {
	return bi.X1, bi.Y1, bi.X2, bi.Y2
}

func (bi *BitmapItem) Coords() []float64 {
	return []float64{bi.x, bi.y}
}

func (bi *BitmapItem) SetCoords(coords []float64) error {
	if len(coords) >= 2 {
		bi.x = coords[0]
		bi.y = coords[1]
		bi.updateBBox()
	}
	return nil
}

func (bi *BitmapItem) Configure(opts []ItemOption) error {
	c := bi.canvas
	for _, opt := range opts {
		if err := opt(c, bi); err != nil {
			return err
		}
	}
	bi.updateBBox()
	return nil
}

func (bi *BitmapItem) updateBBox() {
	if bi.xbm == nil {
		bi.X1 = int(bi.x)
		bi.Y1 = int(bi.y)
		bi.X2 = int(bi.x)
		bi.Y2 = int(bi.y)
		return
	}
	w := bi.xbm.Width
	h := bi.xbm.Height
	ax, ay := anchorOffset(bi.anchor, w, h)
	bi.X1 = int(bi.x) + ax
	bi.Y1 = int(bi.y) + ay
	bi.X2 = bi.X1 + w
	bi.Y2 = bi.Y1 + h
}

func (bi *BitmapItem) Display(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	clipX, clipY, clipW, clipH, originX, originY int) {

	if bi.xbm == nil {
		return
	}

	w := bi.xbm.Width
	h := bi.xbm.Height
	ax, ay := anchorOffset(bi.anchor, w, h)
	drawX := drawableCoord(bi.x, originX) + ax
	drawY := drawableCoord(bi.y, originY) + ay

	rgba := bi.xbm.ToRGBA(bi.Foreground, bi.Background)

	win := bi.canvas.Win
	bgPixel := win.BackgroundPixel
	if bi.canvas.Base.Background != nil {
		bgPixel = bi.canvas.Base.Background.Pixel
	}

	d.PutImageRGBA(drawable, gc, win.Depth, rgba, w, w, h, 0, 0, drawX, drawY, w, h, bgPixel)
}

func (bi *BitmapItem) PointDistance(x, y float64) float64 {
	return rectPointDistance(x, y, float64(bi.X1), float64(bi.Y1), float64(bi.X2), float64(bi.Y2))
}

func (bi *BitmapItem) AreaOverlap(ax1, ay1, ax2, ay2 float64) int {
	x1, y1, x2, y2 := float64(bi.X1), float64(bi.Y1), float64(bi.X2), float64(bi.Y2)
	if ax2 < x1 || ax1 > x2 || ay2 < y1 || ay1 > y2 {
		return -1
	}
	if ax1 <= x1 && ax2 >= x2 && ay1 <= y1 && ay2 >= y2 {
		return 1
	}
	return 0
}

func (bi *BitmapItem) Scale(ox, oy, sx, sy float64) {
	bi.x = ox + (bi.x-ox)*sx
	bi.y = oy + (bi.y-oy)*sy
	bi.updateBBox()
}

func (bi *BitmapItem) Translate(dx, dy float64) {
	bi.x += dx
	bi.y += dy
	bi.updateBBox()
}

func (bi *BitmapItem) Delete(d platform.DisplayServer) {}

// Postscript emits a PostScript representation of the bitmap item.
//
// Mirrors tk/generic/tkCanvBmap.c:BitmapToPostscript.
func (bi *BitmapItem) Postscript(ps *PSContext) error {
	if bi.State() == ItemStateHidden || bi.xbm == nil {
		return nil
	}
	if ps.Prepass {
		return nil
	}
	w, h := bi.xbm.Width, bi.xbm.Height
	ax, ay := anchorOffset(bi.anchor, w, h)
	x := float64(int(bi.x) + ax)
	y := float64(int(bi.y) + ay)
	ps.writef("gsave %.15g %.15g translate %d %d scale\n", x, ps.PsY(int(y)+h), w, h)
	// Foreground is goimage/color.RGBA; convert to takigo ColorRef for the emitter.
	fg := &colortakigo.ColorRef{Pixel: 0, Red: uint16(bi.Foreground.R) << 8, Green: uint16(bi.Foreground.G) << 8, Blue: uint16(bi.Foreground.B) << 8}
	if bi.Foreground.A != 0 {
		ps.Color(fg)
	}
	ps.Bitmap(bi.xbm.Bits, w, h)
	ps.write("grestore newpath\n")
	return nil
}
