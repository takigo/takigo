package color

import "testing"

func TestParseHexRGB(t *testing.T) {
	r, g, b, err := Parse("#f00")
	if err != nil {
		t.Fatal(err)
	}
	if r != 0xff00 || g != 0 || b != 0 {
		t.Errorf("#f00 = (%d,%d,%d), want (0xff00,0,0)", r, g, b)
	}
}

func TestParseHexRRGGBB(t *testing.T) {
	r, g, b, err := Parse("#ff8800")
	if err != nil {
		t.Fatal(err)
	}
	if r != 0xff00 || g != 0x8800 || b != 0 {
		t.Errorf("#ff8800 = (%04x,%04x,%04x), want (ff00,8800,0000)", r, g, b)
	}
}

func TestParseHexRRRRGGGGBBBB(t *testing.T) {
	r, g, b, err := Parse("#ffff00000000")
	if err != nil {
		t.Fatal(err)
	}
	if r != 0xffff || g != 0 || b != 0 {
		t.Errorf("#ffff00000000 = (%04x,%04x,%04x), want (ffff,0000,0000)", r, g, b)
	}
}

func TestParseNamedColor(t *testing.T) {
	tests := []struct {
		name  string
		wantR uint16
		wantG uint16
		wantB uint16
	}{
		{"red", 0xff00, 0, 0},
		{"blue", 0, 0, 0xff00},
		{"DarkGreen", 0, 0x6400, 0},
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
	invalids := []string{"", "#zz", "#12345", "nonexistent"}
	for _, s := range invalids {
		_, _, _, err := Parse(s)
		if err == nil {
			t.Errorf("Parse(%q) should return error", s)
		}
	}
}

func TestColorRGBA(t *testing.T) {
	c := &Color{Red: 0xff00, Green: 0x8000, Blue: 0}
	r, g, b, a := c.RGBA()
	if r != 0xff || g != 0x80 || b != 0 || a != 255 {
		t.Errorf("RGBA() = (%d,%d,%d,%d), want (255,128,0,255)", r, g, b, a)
	}
}

func TestTrueColorPixelRoundTrip(t *testing.T) {
	// Parse a color, compute pixel, verify it's the expected 24-bit value.
	r, g, b, _ := Parse("#ff8800")
	pixel := trueColorPixel(r, g, b)
	wantPixel := uint64(0xff)<<16 | uint64(0x88)<<8 | uint64(0x00)
	if pixel != wantPixel {
		t.Errorf("trueColorPixel(#ff8800) = %06x, want %06x", pixel, wantPixel)
	}
}

func TestParseHexWhite(t *testing.T) {
	r, g, b, err := Parse("#fff")
	if err != nil {
		t.Fatal(err)
	}
	if r != 0xff00 || g != 0xff00 || b != 0xff00 {
		t.Errorf("#fff = (%04x,%04x,%04x), want (ff00,ff00,ff00)", r, g, b)
	}
}

func TestParseHexBlack(t *testing.T) {
	r, g, b, err := Parse("#000000")
	if err != nil {
		t.Fatal(err)
	}
	if r != 0 || g != 0 || b != 0 {
		t.Errorf("#000000 = (%d,%d,%d), want (0,0,0)", r, g, b)
	}
}
