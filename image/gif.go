package image

import (
	"fmt"
	goimage "image"
	"image/draw"
	"image/gif"
	"io"
	"time"
)

// Frame is one frame of an animated GIF, composited onto the frames before
// it, with the time it stays on screen.
type Frame struct {
	Photo *Photo
	Delay time.Duration
}

// NewPhotoFramesFromGIF decodes every frame of a GIF. Tk's GIF reader
// returns the raw sub-image selected by "-format {gif -index N}"; the frames
// here are already composited with the GIF disposal methods applied, so each
// can be shown as is. Frame i is named name#i.
func NewPhotoFramesFromGIF(name string, r io.Reader) ([]Frame, error) {
	g, err := gif.DecodeAll(r)
	if err != nil {
		return nil, fmt.Errorf("image: decode gif: %w", err)
	}
	frames := make([]Frame, len(g.Image))
	for i, rgba := range compositeGIF(g) {
		frames[i].Photo = NewPhoto(fmt.Sprintf("%s#%d", name, i), rgba)
		if i < len(g.Delay) {
			frames[i].Delay = time.Duration(g.Delay[i]) * 10 * time.Millisecond
		}
	}
	return frames, nil
}

func compositeGIF(g *gif.GIF) []*goimage.RGBA {
	if len(g.Image) == 0 {
		return nil
	}
	bounds := goimage.Rect(0, 0, g.Config.Width, g.Config.Height)
	if bounds.Empty() {
		bounds = g.Image[0].Bounds()
	}
	screen := goimage.NewRGBA(bounds)
	out := make([]*goimage.RGBA, len(g.Image))
	for i, frame := range g.Image {
		var disposal byte
		if i < len(g.Disposal) {
			disposal = g.Disposal[i]
		}
		var saved *goimage.RGBA
		if disposal == gif.DisposalPrevious {
			saved = cloneRGBA(screen)
		}
		draw.Draw(screen, frame.Bounds(), frame, frame.Bounds().Min, draw.Over)
		out[i] = cloneRGBA(screen)
		switch disposal {
		case gif.DisposalBackground:
			draw.Draw(screen, frame.Bounds(), goimage.Transparent, goimage.Point{}, draw.Src)
		case gif.DisposalPrevious:
			screen = saved
		}
	}
	return out
}

func cloneRGBA(src *goimage.RGBA) *goimage.RGBA {
	dst := goimage.NewRGBA(src.Rect)
	copy(dst.Pix, src.Pix)
	return dst
}
