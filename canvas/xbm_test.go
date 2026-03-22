package canvas

import (
	"image/color"
	"testing"
)

const testXBM = `#define test_width 4
#define test_height 2
static unsigned char test_bits[] = {
   0x05, 0x0a};
`

func TestParseXBMValid(t *testing.T) {
	xbm, err := ParseXBM(testXBM)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if xbm.Width != 4 || xbm.Height != 2 {
		t.Errorf("size: got %dx%d, want 4x2", xbm.Width, xbm.Height)
	}
	if len(xbm.Bits) != 2 {
		t.Errorf("bits: got %d bytes, want 2", len(xbm.Bits))
	}
	if xbm.Bits[0] != 0x05 || xbm.Bits[1] != 0x0a {
		t.Errorf("bits: got %v, want [0x05, 0x0a]", xbm.Bits)
	}
}

func TestParseXBMMissingDimensions(t *testing.T) {
	_, err := ParseXBM(`static unsigned char test_bits[] = { 0x00 };`)
	if err == nil {
		t.Error("expected error for missing dimensions")
	}
}

func TestParseXBMMissingData(t *testing.T) {
	_, err := ParseXBM(`#define test_width 4
#define test_height 2`)
	if err == nil {
		t.Error("expected error for missing data array")
	}
}

func TestParseXBMInsufficientData(t *testing.T) {
	_, err := ParseXBM(`#define test_width 8
#define test_height 4
static unsigned char test_bits[] = {
   0x00, 0x00};
`)
	if err == nil {
		t.Error("expected error for insufficient data bytes")
	}
}

func TestXBMToRGBA(t *testing.T) {
	xbm, err := ParseXBM(testXBM)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	fg := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	bg := color.RGBA{R: 0, G: 0, B: 0, A: 0}
	rgba := xbm.ToRGBA(fg, bg)

	// 4x2 image = 32 bytes (4 channels per pixel)
	if len(rgba) != 4*2*4 {
		t.Fatalf("rgba length: got %d, want 32", len(rgba))
	}

	// Row 0: bits = 0x05 = 0101 → pixels: fg, bg, fg, bg
	// Pixel 0 (fg)
	if rgba[0] != 255 || rgba[1] != 0 || rgba[2] != 0 || rgba[3] != 255 {
		t.Errorf("pixel (0,0) should be fg, got %v", rgba[0:4])
	}
	// Pixel 1 (bg)
	if rgba[4] != 0 || rgba[5] != 0 || rgba[6] != 0 || rgba[7] != 0 {
		t.Errorf("pixel (1,0) should be bg, got %v", rgba[4:8])
	}
	// Pixel 2 (fg)
	if rgba[8] != 255 || rgba[9] != 0 || rgba[10] != 0 || rgba[11] != 255 {
		t.Errorf("pixel (2,0) should be fg, got %v", rgba[8:12])
	}
}
