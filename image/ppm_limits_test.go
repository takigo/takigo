package image

import (
	"image/color"
	"strings"
	"testing"
)

func TestDecodePPMRejectsHugeHeader(t *testing.T) {
	for _, src := range []string{"P6 100000 100000 255\n", "P5 70000 1 255\n", "P4 -3 2\n"} {
		if _, err := decodePPM(strings.NewReader(src)); err == nil {
			t.Errorf("decodePPM accepted %q", src)
		}
	}
}

func TestNewPhotoFromXBMRejectsNegativeSize(t *testing.T) {
	src := "#define t_width -8\n#define t_height 1\nstatic char t_bits[] = { 0x00 };"
	if _, err := NewPhotoFromXBM("t", src, color.RGBA{}, color.RGBA{}); err == nil {
		t.Error("NewPhotoFromXBM accepted a negative width")
	}
}
