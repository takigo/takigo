package image

import (
	"bytes"
	goimage "image"
	"image/color"
	"os"
	"testing"
)

func TestPPMDecoderBinaryRGB(t *testing.T) {
	// 2x1 image, P6, maxval 255, pixels: red green.
	header := "P6\n2 1\n255\n"
	data := append([]byte(header), 0xff, 0x00, 0x00, 0x00, 0xff, 0x00)

	p, err := NewPhotoFromPPMReader("test", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if p.Width() != 2 || p.Height() != 1 {
		t.Fatalf("dims: %dx%d, want 2x1", p.Width(), p.Height())
	}
	want := []color.RGBA{{255, 0, 0, 255}, {0, 255, 0, 255}}
	for i, w := range want {
		got := p.RGBA().At(i, 0)
		if !equalRGBA(got, w) {
			t.Errorf("px %d: got %v, want %v", i, got, w)
		}
	}
}

func TestPPMDecoderBinaryGray(t *testing.T) {
	// 2x1 P5, maxval 255, pixels: 32 200.
	header := "P5\n2 1\n255\n"
	data := append([]byte(header), 32, 200)

	p, err := NewPhotoFromPPMReader("test", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if p.Width() != 2 || p.Height() != 1 {
		t.Fatalf("dims: %dx%d", p.Width(), p.Height())
	}
	want := []byte{32, 200}
	for i, w := range want {
		got := p.RGBA().At(i, 0)
		gr := color.GrayModel.Convert(got).(color.Gray)
		if gr.Y != w {
			t.Errorf("px %d: got gray %d, want %d", i, gr.Y, w)
		}
	}
}

func TestPPMDecoderASCIIRGB(t *testing.T) {
	src := "P3\n# comment\n2 2\n255\n255 0 0 0 255 0\n0 0 255 128 128 128\n"
	p, err := NewPhotoFromPPMReader("test", bytes.NewReader([]byte(src)))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if p.Width() != 2 || p.Height() != 2 {
		t.Fatalf("dims: %dx%d", p.Width(), p.Height())
	}
	want := []color.RGBA{
		{255, 0, 0, 255},
		{0, 255, 0, 255},
		{0, 0, 255, 255},
		{128, 128, 128, 255},
	}
	for i, w := range want {
		got := p.RGBA().At(i%2, i/2)
		if !equalRGBA(got, w) {
			t.Errorf("px %d: got %v, want %v", i, got, w)
		}
	}
}

func TestPPMDecoderBinaryPBM(t *testing.T) {
	// 8x2 P4 bitmap, two rows: 0xAA, 0x55.
	header := "P4\n8 2\n"
	data := append([]byte(header), 0xAA, 0x55)
	p, err := NewPhotoFromPPMReader("test", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if p.Width() != 8 || p.Height() != 2 {
		t.Fatalf("dims: %dx%d", p.Width(), p.Height())
	}
	// 0xAA = 10101010 -> 1=black at columns 0,2,4,6
	for x := 0; x < 8; x++ {
		got := p.RGBA().At(x, 0)
		want := color.RGBA{255, 255, 255, 255}
		if x%2 == 0 {
			want = color.RGBA{0, 0, 0, 255}
		}
		if !equalRGBA(got, want) {
			t.Errorf("row0 x=%d: got %v, want %v", x, got, want)
		}
	}
}

func TestPPMDecoderInvalidMagic(t *testing.T) {
	if _, err := NewPhotoFromPPMReader("x", bytes.NewReader([]byte("NOTPPM\n"))); err == nil {
		t.Fatal("expected error on invalid magic")
	}
}

func equalRGBA(a color.Color, b color.RGBA) bool {
	r1, g1, b1, a1 := a.RGBA()
	r2, g2, b2, a2 := b.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}

func TestNewPhotoFromFileDispatch(t *testing.T) {
	// PPM via dispatch.
	ppm := "P6\n1 1\n255\n" + string([]byte{10, 20, 30})
	tmp := t.TempDir()
	path := tmp + "/a.ppm"
	if err := writeFile(path, []byte(ppm)); err != nil {
		t.Fatal(err)
	}
	p, err := NewPhotoFromFile("a", path)
	if err != nil {
		t.Fatalf("dispatch ppm: %v", err)
	}
	if p.Width() != 1 || p.Height() != 1 {
		t.Fatalf("dims: %dx%d", p.Width(), p.Height())
	}
	got := color.RGBA{10, 20, 30, 255}
	if !equalRGBA(p.RGBA().At(0, 0), got) {
		t.Errorf("px: got %v want %v", p.RGBA().At(0, 0), got)
	}
}

func writeFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0o644)
}

func TestDetectImageFormat(t *testing.T) {
	cases := []struct {
		path string
		want string
	}{
		{"a.PPM", "ppm"},
		{"a.svg", "svg"},
		{"a.XBM", "xbm"},
		{"a.png", "png"},
		{"a", ""},
	}
	for _, c := range cases {
		if got := detectImageFormat(c.path); got != c.want {
			t.Errorf("detect(%q)=%q, want %q", c.path, got, c.want)
		}
	}
}

// silence unused warning when not all helpers are exercised.
var _ = goimage.NewRGBA
