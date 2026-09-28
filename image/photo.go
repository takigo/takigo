package image

import (
	"fmt"
	goimage "image"
	"image/color"
	"image/draw"
	_ "image/gif" // register GIF decoder
	_ "image/png" // register PNG decoder
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/msorc/takigo/internal/nanosvg"
	"github.com/msorc/takigo/platform"
)

// Photo is an image backed by Go RGBA pixel data. It caches rendered
// Pixmaps for repeated drawing: one for an opaque image, or one per
// background colour (up to maxPhotoPixmaps) for an image with
// transparency, so a photo shown on several backgrounds (e.g. a button's
// normal and active colours) is not re-uploaded on every redraw.
type Photo struct {
	name string
	rgba *goimage.RGBA
	// straight holds the decoded straight-alpha pixels (w*4 stride) while
	// rgba is unmodified, so drawing can blend exactly like Tk's photos.
	straight []byte

	// Pixmap cache, most recently used last.
	server      platform.DisplayServer
	pixmaps     []photoPixmap
	pixmapW     int
	pixmapH     int
	pixmapDirty bool
	opaque      bool // no transparent pixels: the background is irrelevant
}

// maxPhotoPixmaps bounds the per-background pixmaps kept for one photo.
const maxPhotoPixmaps = 4

// photoPixmap is a Photo rendered over one background pixel.
type photoPixmap struct {
	id platform.PixmapID
	bg uint64
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
// Supported formats are selected by file extension:
//
//   - .png, .gif, .jpg/.jpeg  via Go stdlib registered decoders
//   - .xbm                    via NewPhotoFromXBMFile (black fg, transparent bg)
//   - .ppm, .pgm, .pbm        via NewPhotoFromPPMFile (Netpbm P1..P6)
//   - .svg                    via NewPhotoFromSVGFile (nanosvg)
func NewPhotoFromFile(name, path string) (*Photo, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".xbm":
		return NewPhotoFromXBMFile(name, path, color.RGBA{0, 0, 0, 255}, color.RGBA{0, 0, 0, 0})
	case ".ppm", ".pgm", ".pbm":
		return NewPhotoFromPPMFile(name, path)
	case ".svg":
		return NewPhotoFromSVGFile(name, path)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("image: open %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()
	return NewPhotoFromReader(name, f)
}

// NewPhotoFromXBMFile loads an XBM (X BitMap) file and creates a Photo.
// fg and bg are the colors for bit=1 and bit=0 pixels respectively.
// Use bg.A=0 for transparent backgrounds.
func NewPhotoFromXBMFile(name, path string, fg, bg color.RGBA) (*Photo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("image: open xbm %s: %w", path, err)
	}
	return NewPhotoFromXBM(name, string(data), fg, bg)
}

// NewPhotoFromXBM parses XBM source text and creates a Photo.
func NewPhotoFromXBM(name, src string, fg, bg color.RGBA) (*Photo, error) {
	var w, h int
	for _, line := range strings.Split(src, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "#define") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		val, err := strconv.Atoi(fields[2])
		if err != nil {
			continue
		}
		if strings.HasSuffix(fields[1], "_width") {
			w = val
		} else if strings.HasSuffix(fields[1], "_height") {
			h = val
		}
	}
	if w == 0 || h == 0 {
		return nil, fmt.Errorf("image: xbm: missing width or height")
	}

	start := strings.Index(src, "{")
	end := strings.LastIndex(src, "}")
	if start < 0 || end < 0 || end <= start {
		return nil, fmt.Errorf("image: xbm: could not find data array")
	}
	body := src[start+1 : end]

	var bits []byte
	for _, tok := range strings.FieldsFunc(body, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n' || r == '\r'
	}) {
		tok = strings.TrimSpace(tok)
		tok = strings.TrimPrefix(tok, "0x")
		tok = strings.TrimPrefix(tok, "0X")
		b, err := strconv.ParseUint(tok, 16, 8)
		if err != nil {
			continue
		}
		bits = append(bits, byte(b))
	}

	rowBytes := (w + 7) / 8
	if len(bits) < rowBytes*h {
		return nil, fmt.Errorf("image: xbm: not enough data")
	}

	rgba := goimage.NewRGBA(goimage.Rect(0, 0, w, h))
	for row := range h {
		for col := range w {
			byteIdx := row*rowBytes + col/8
			bit := (bits[byteIdx] >> uint(col%8)) & 1
			if bit == 1 {
				rgba.SetRGBA(col, row, fg)
			} else {
				rgba.SetRGBA(col, row, bg)
			}
		}
	}
	return NewPhoto(name, rgba), nil
}

// NewPhotoFromReader decodes an image from a reader.
func NewPhotoFromReader(name string, r io.Reader) (*Photo, error) {
	img, _, err := goimage.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("image: decode: %w", err)
	}
	rgba := toRGBA(img)
	p := NewPhoto(name, rgba)
	if _, pre := img.(*goimage.RGBA); !pre && img.Bounds().Min == (goimage.Point{}) {
		n := goimage.NewNRGBA(img.Bounds())
		draw.Draw(n, n.Rect, img, n.Rect.Min, draw.Src)
		p.straight = n.Pix
	}
	return p, nil
}

// CopyOption configures a NewPhotoFromPhoto call.
type CopyOption func(*copyConfig)

type copyConfig struct {
	zoomX, zoomY float64
}

// Zoom sets the zoom factor for both axes.
// Matches Tk's "imageName copy srcName -zoom x y".
func Zoom(factor float64) CopyOption {
	return func(c *copyConfig) {
		c.zoomX = factor
		c.zoomY = factor
	}
}

// NewPhotoFromPhoto creates a new photo by copying (and optionally scaling)
// an existing photo. Matches Tk's "destImage copy srcImage ?-zoom x y?".
func NewPhotoFromPhoto(src *Photo, name string, opts ...CopyOption) *Photo {
	cfg := copyConfig{zoomX: 1, zoomY: 1}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.zoomX <= 0 {
		cfg.zoomX = 1
	}
	if cfg.zoomY <= 0 {
		cfg.zoomY = 1
	}

	srcW := src.Width()
	srcH := src.Height()
	dstW := int(float64(srcW)*cfg.zoomX + 0.5)
	dstH := int(float64(srcH)*cfg.zoomY + 0.5)
	if dstW < 1 {
		dstW = 1
	}
	if dstH < 1 {
		dstH = 1
	}

	dst := goimage.NewRGBA(goimage.Rect(0, 0, dstW, dstH))
	srcRGBA := src.rgba

	// Nearest-neighbor scaling.
	for y := range dstH {
		srcY := int(float64(y) / cfg.zoomY)
		if srcY >= srcH {
			srcY = srcH - 1
		}
		for x := range dstW {
			srcX := int(float64(x) / cfg.zoomX)
			if srcX >= srcW {
				srcX = srcW - 1
			}
			off := srcY*srcRGBA.Stride + srcX*4
			dOff := y*dst.Stride + x*4
			copy(dst.Pix[dOff:dOff+4], srcRGBA.Pix[off:off+4])
		}
	}

	return NewPhoto(name, dst)
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

// Pixels returns the underlying RGBA pixel buffer (4 bytes per pixel, RGBA
// order) suitable for byte-for-byte export to PostScript / PNG / etc.
// Returns nil if the photo has no pixels.
func (p *Photo) Pixels() []byte {
	if p.rgba == nil {
		return nil
	}
	return p.rgba.Pix
}

// Invalidate marks the cached pixmap as stale, forcing re-creation on
// the next Draw call.
func (p *Photo) Invalidate() {
	p.pixmapDirty = true
	p.straight = nil
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

	if pix := p.ensurePixmap(d, drawable, gc, depth, bgPixel); pix != platform.PixmapID(0) {
		d.CopyArea(platform.PixmapDrawable(pix), drawable, gc,
			imgX, imgY, uint(w), uint(h), dstX, dstY)
	} else {
		// Fallback: direct PutImage (no caching).
		d.PutImageRGBA(drawable, gc, depth,
			p.rgba.Pix, p.rgba.Stride, imgW, imgH,
			imgX, imgY, dstX, dstY, w, h, bgPixel)
	}
}

// ensurePixmap returns a pixmap holding the image rendered over bgPixel,
// creating it if needed. It returns 0 if no pixmap could be created.
func (p *Photo) ensurePixmap(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	depth int, bgPixel uint64) platform.PixmapID {
	imgW := p.Width()
	imgH := p.Height()

	if p.pixmapDirty || p.server != d || p.pixmapW != imgW || p.pixmapH != imgH {
		p.freePixmaps()
		p.server, p.pixmapW, p.pixmapH = d, imgW, imgH
		p.opaque = p.straight == nil && isOpaque(p.rgba)
		p.pixmapDirty = false
	}
	key := bgPixel
	if p.opaque {
		key = 0
	}
	for i, c := range p.pixmaps {
		if c.bg == key {
			if last := len(p.pixmaps) - 1; i != last {
				copy(p.pixmaps[i:], p.pixmaps[i+1:])
				p.pixmaps[last] = c
			}
			return c.id
		}
	}

	pix := d.CreatePixmap(drawable, uint(imgW), uint(imgH), uint(depth))
	if pix == platform.PixmapID(0) {
		return 0
	}

	// Render RGBA data into the pixmap; straight-alpha pixels are blended
	// over bgPixel as BlendComplexAlpha (tkImgPhInstance.c) does.
	data, stride := p.rgba.Pix, p.rgba.Stride
	if p.straight != nil {
		data, stride = nanosvg.BlendOver(p.straight, bgPixel), imgW*4
	}
	d.PutImageRGBA(platform.PixmapDrawable(pix), gc, depth,
		data, stride, imgW, imgH,
		0, 0, 0, 0, imgW, imgH, bgPixel)

	if len(p.pixmaps) == maxPhotoPixmaps {
		d.FreePixmap(p.pixmaps[0].id)
		p.pixmaps = p.pixmaps[1:]
	}
	p.pixmaps = append(p.pixmaps, photoPixmap{id: pix, bg: key})
	return pix
}

// freePixmaps releases every cached pixmap.
func (p *Photo) freePixmaps() {
	for _, c := range p.pixmaps {
		if p.server != nil {
			p.server.FreePixmap(c.id)
		}
	}
	p.pixmaps = nil
}

// isOpaque reports whether every pixel of img has full alpha.
func isOpaque(img *goimage.RGBA) bool {
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		row := img.Pix[img.PixOffset(b.Min.X, y):img.PixOffset(b.Max.X, y)]
		for i := 3; i < len(row); i += 4 {
			if row[i] != 0xff {
				return false
			}
		}
	}
	return true
}

// Destroy releases the cached pixmap.
func (p *Photo) Destroy() {
	p.freePixmaps()
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
