package image

import (
	"bytes"
	goimage "image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"testing"
	"time"
)

func TestCompositeGIF(t *testing.T) {
	red := color.RGBA{255, 0, 0, 255}
	blue := color.RGBA{0, 0, 255, 255}
	pal := color.Palette{color.RGBA{}, red, blue}

	full := goimage.NewPaletted(goimage.Rect(0, 0, 4, 4), pal)
	for i := range full.Pix {
		full.Pix[i] = 1
	}
	patch := goimage.NewPaletted(goimage.Rect(1, 1, 3, 3), pal)
	for i := range patch.Pix {
		patch.Pix[i] = 2
	}
	dot := goimage.NewPaletted(goimage.Rect(0, 0, 1, 1), pal)

	tests := []struct {
		name     string
		disposal byte
		want     color.RGBA // pixel (1,1) of the third frame
	}{
		{"none keeps the patch", gif.DisposalNone, blue},
		{"background clears the patch", gif.DisposalBackground, color.RGBA{}},
		{"previous restores the first frame", gif.DisposalPrevious, red},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &gif.GIF{
				Image:    []*goimage.Paletted{full, patch, dot},
				Delay:    []int{10, 20, 30},
				Disposal: []byte{gif.DisposalNone, tt.disposal, gif.DisposalNone},
				Config:   goimage.Config{Width: 4, Height: 4},
			}
			frames := compositeGIF(g)
			if len(frames) != 3 {
				t.Fatalf("got %d frames, want 3", len(frames))
			}
			if got := frames[1].RGBAAt(1, 1); got != blue {
				t.Errorf("frame 1 (1,1) = %v, want %v", got, blue)
			}
			if got := frames[1].RGBAAt(0, 0); got != red {
				t.Errorf("frame 1 (0,0) = %v, want %v", got, red)
			}
			if got := frames[2].RGBAAt(1, 1); got != tt.want {
				t.Errorf("frame 2 (1,1) = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewPhotoFramesFromGIF(t *testing.T) {
	pal := color.Palette{color.RGBA{}, color.RGBA{255, 0, 0, 255}}
	g := &gif.GIF{
		Image: []*goimage.Paletted{
			goimage.NewPaletted(goimage.Rect(0, 0, 2, 2), pal),
			goimage.NewPaletted(goimage.Rect(0, 0, 2, 2), pal),
		},
		Delay: []int{5, 7},
	}
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, g); err != nil {
		t.Fatal(err)
	}
	frames, err := NewPhotoFramesFromGIF("anim", &buf)
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 2 {
		t.Fatalf("got %d frames, want 2", len(frames))
	}
	if frames[1].Delay != 70*time.Millisecond {
		t.Errorf("delay = %v, want 70ms", frames[1].Delay)
	}
	if got := frames[1].Photo.Name(); got != "anim#1" {
		t.Errorf("name = %q, want anim#1", got)
	}
}

func TestNewPhotoFromReaderJPEG(t *testing.T) {
	src := goimage.NewRGBA(goimage.Rect(0, 0, 8, 6))
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, src, nil); err != nil {
		t.Fatal(err)
	}
	p, err := NewPhotoFromReader("j", &buf)
	if err != nil {
		t.Fatalf("decode jpeg: %v", err)
	}
	if p.Width() != 8 || p.Height() != 6 {
		t.Errorf("size = %dx%d, want 8x6", p.Width(), p.Height())
	}
}
