package color

import (
	"sync"
	"testing"
)

func TestParseHexRGB(t *testing.T) {
	t.Parallel()
	r, g, b, err := Parse("#f00")
	if err != nil {
		t.Fatal(err)
	}
	if r != 0xffff || g != 0 || b != 0 {
		t.Errorf("#f00 = (%d,%d,%d), want (0xffff,0,0)", r, g, b)
	}
}

func TestParseHexRRGGBB(t *testing.T) {
	t.Parallel()
	r, g, b, err := Parse("#ff8800")
	if err != nil {
		t.Fatal(err)
	}
	if r != 0xffff || g != 0x8888 || b != 0 {
		t.Errorf("#ff8800 = (%04x,%04x,%04x), want (ffff,8888,0000)", r, g, b)
	}
}

func TestParseHexRRRRGGGGBBBB(t *testing.T) {
	t.Parallel()
	r, g, b, err := Parse("#ffff00000000")
	if err != nil {
		t.Fatal(err)
	}
	if r != 0xffff || g != 0 || b != 0 {
		t.Errorf("#ffff00000000 = (%04x,%04x,%04x), want (ffff,0000,0000)", r, g, b)
	}
}

func TestParseNamedColor(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		wantR uint16
		wantG uint16
		wantB uint16
	}{
		{"red", 0xffff, 0, 0},
		{"blue", 0, 0, 0xffff},
		{"DarkGreen", 0, 0x6464, 0},
	}
	for _, tt := range tests {
		r, g, b, err := Parse(tt.name)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", tt.name, err)
			continue
		}
		if r != tt.wantR || g != tt.wantG || b != tt.wantB {
			t.Errorf("Parse(%q) = (%04x,%04x,%04x), want (%04x,%04x,%04x)",
				tt.name, r, g, b, tt.wantR, tt.wantG, tt.wantB)
		}
	}
}

func TestParseInvalid(t *testing.T) {
	t.Parallel()
	invalids := []string{"", "#zz", "#12345", "nonexistent", "#zzzzzz", "#12g", "#-12345", "#1234567890123"}
	for _, s := range invalids {
		_, _, _, err := Parse(s)
		if err == nil {
			t.Errorf("Parse(%q) should return error", s)
		}
	}
}

func TestParseHexRRRGGGBBB(t *testing.T) {
	t.Parallel()
	r, g, b, err := Parse("#fff800123")
	if err != nil {
		t.Fatal(err)
	}
	if r != 0xffff || g != 0x8008 || b != 0x1231 {
		t.Errorf("Parse(#fff800123) = %04x %04x %04x, want ffff 8008 1231", r, g, b)
	}
}

func TestColorRGBA(t *testing.T) {
	t.Parallel()
	c := &Color{Red: 0xff00, Green: 0x8000, Blue: 0}
	r, g, b, a := c.RGBA()
	if r != 0xff || g != 0x80 || b != 0 || a != 255 {
		t.Errorf("RGBA() = (%d,%d,%d,%d), want (255,128,0,255)", r, g, b, a)
	}
}

func TestTrueColorPixelRoundTrip(t *testing.T) {
	t.Parallel()
	// Parse a color, compute pixel, verify it's the expected 24-bit value.
	r, g, b, _ := Parse("#ff8800")
	pixel := trueColorPixel(r, g, b)
	wantPixel := uint64(0xff)<<16 | uint64(0x88)<<8 | uint64(0x00)
	if pixel != wantPixel {
		t.Errorf("trueColorPixel(#ff8800) = %06x, want %06x", pixel, wantPixel)
	}
}

func TestParseHexWhite(t *testing.T) {
	t.Parallel()
	r, g, b, err := Parse("#fff")
	if err != nil {
		t.Fatal(err)
	}
	if r != 0xffff || g != 0xffff || b != 0xffff {
		t.Errorf("#fff = (%04x,%04x,%04x), want (ffff,ffff,ffff)", r, g, b)
	}
}

func TestParseHexBlack(t *testing.T) {
	t.Parallel()
	r, g, b, err := Parse("#000000")
	if err != nil {
		t.Fatal(err)
	}
	if r != 0 || g != 0 || b != 0 {
		t.Errorf("#000000 = (%d,%d,%d), want (0,0,0)", r, g, b)
	}
}

// Concurrent GetByValue calls for one value share one Color.
func TestGetByValueConcurrent(t *testing.T) {
	t.Parallel()
	c := NewCache(0)
	for v := range uint16(200) {
		var wg sync.WaitGroup
		got := make([]*Color, 8)
		for i := range got {
			wg.Go(func() { got[i], _ = c.GetByValue(v, v, v) })
		}
		wg.Wait()
		for _, col := range got {
			if col != got[0] {
				t.Fatalf("value %d: GetByValue returned different Colors for the same value", v)
			}
		}
	}
}
