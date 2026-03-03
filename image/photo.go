package image

import (
	"fmt"
	goimage "image"
	"image/draw"
	_ "image/gif" // register GIF decoder
	_ "image/png" // register PNG decoder
	"io"
	"os"

	"github.com/msorc/takigo/platform"
)

// Photo is an image backed by Go RGBA pixel data. It caches a Pixmap
// for efficient repeated drawing, re-creating it when the background color
// changes or the pixel data is invalidated.
type Photo struct {
	name string
	rgba *goimage.RGBA

	// Pixmap cache.
	server     platform.DisplayServer
	pixmap     platform.PixmapID
	pixmapW    int
	pixmapH    int
	pixmapBg   uint64 // bgPixel used when rendering the cached pixmap
	pixmapDirty bool
}

// NewPhoto creates a photo image from existing RGBA data.
func NewPhoto(name string, rgba *goimage.RGBA) *Photo {
	return &Photo{
		name:        name,
		rgba:        rgba,
		pixmapDirty: true,
	}
}

// NewPhotoFromFile opens a file and decodes it as an image.
// Supports PNG and GIF via Go stdlib registered decoders.
func NewPhotoFromFile(name, path string) (*Photo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("image: open %s: %w", path, err)
	}
	defer f.Close()
	return NewPhotoFromReader(name, f)
}

// NewPhotoFromReader decodes an image from a reader.
func NewPhotoFromReader(name string, r io.Reader) (*Photo, error) {
	img, _, err := goimage.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("image: decode: %w", err)
	}
	rgba := toRGBA(img)
	return NewPhoto(name, rgba), nil
}

// Name returns the image's registered name.
func (p *Photo) Name() string { return p.name }

// Width returns the image width in pixels.
func (p *Photo) Width() int { return p.rgba.Rect.Dx() }

// Height returns the image height in pixels.
func (p *Photo) Height() int { return p.rgba.Rect.Dy() }

// RGBA returns the underlying pixel data for direct manipulation.
// Call Invalidate after modifying pixels to force pixmap re-creation.
func (p *Photo) RGBA() *goimage.RGBA { return p.rgba }

// Invalidate marks the cached pixmap as stale, forcing re-creation on
// the next Draw call.
func (p *Photo) Invalidate() {
	p.pixmapDirty = true
}

// Draw renders a region of the photo onto a drawable.
func (p *Photo) Draw(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	depth int,
	imgX, imgY, w, h, dstX, dstY int,
	bgPixel uint64) {

	if w <= 0 || h <= 0 {
		return
	}

	// Clamp to image bounds.
	imgW := p.Width()
	imgH := p.Height()
	if imgX < 0 {
		dstX -= imgX
		w += imgX
		imgX = 0
	}
	if imgY < 0 {
		dstY -= imgY
		h += imgY
		imgY = 0
	}
	if imgX+w > imgW {
		w = imgW - imgX
	}
	if imgY+h > imgH {
		h = imgH - imgY
	}
	if w <= 0 || h <= 0 {
		return
	}

	// Ensure cached pixmap is up to date.
	p.ensurePixmap(d, drawable, gc, depth, bgPixel)

	if p.pixmap != platform.PixmapID(0) {
		// Copy from cached pixmap.
		d.CopyArea(platform.PixmapDrawable(p.pixmap), drawable, gc,
			imgX, imgY, uint(w), uint(h), dstX, dstY)
	} else {
		// Fallback: direct PutImage (no caching).
		d.PutImageRGBA(drawable, gc, depth,
			p.rgba.Pix, p.rgba.Stride, imgW, imgH,
			imgX, imgY, dstX, dstY, w, h, bgPixel)
	}
}

// ensurePixmap creates or re-creates the cached pixmap if needed.
func (p *Photo) ensurePixmap(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	depth int, bgPixel uint64) {

	imgW := p.Width()
	imgH := p.Height()

	needRecreate := p.pixmapDirty ||
		p.pixmap == platform.PixmapID(0) ||
		p.server != d ||
		p.pixmapW != imgW ||
		p.pixmapH != imgH ||
		p.pixmapBg != bgPixel

	if !needRecreate {
		return
	}

	// Free old pixmap.
	if p.pixmap != platform.PixmapID(0) && p.server != nil {
		p.server.FreePixmap(p.pixmap)
		p.pixmap = platform.PixmapID(0)
	}

	// Create new pixmap.
	pix := d.CreatePixmap(drawable, uint(imgW), uint(imgH), uint(depth))
	if pix == platform.PixmapID(0) {
		return
	}

	// Render RGBA data into the pixmap.
	d.PutImageRGBA(platform.PixmapDrawable(pix), gc, depth,
		p.rgba.Pix, p.rgba.Stride, imgW, imgH,
		0, 0, 0, 0, imgW, imgH, bgPixel)

	p.server = d
	p.pixmap = pix
	p.pixmapW = imgW
	p.pixmapH = imgH
	p.pixmapBg = bgPixel
	p.pixmapDirty = false
}

// Destroy releases the cached pixmap.
func (p *Photo) Destroy() {
	if p.pixmap != platform.PixmapID(0) && p.server != nil {
		p.server.FreePixmap(p.pixmap)
		p.pixmap = platform.PixmapID(0)
	}
}

// toRGBA converts any image.Image to *image.RGBA.
func toRGBA(img goimage.Image) *goimage.RGBA {
	if rgba, ok := img.(*goimage.RGBA); ok {
		return rgba
	}
	bounds := img.Bounds()
	rgba := goimage.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
	return rgba
}
