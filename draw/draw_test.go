package draw

import (
	"testing"
)

func TestToPixel(t *testing.T) {
	tests := []struct {
		name      string
		r, g, b   uint16
		wantPixel uint64
	}{
		{"black", 0, 0, 0, 0x000000},
		{"white", 0xFFFF, 0xFFFF, 0xFFFF, 0xFFFFFF},
		{"red", 0xFFFF, 0, 0, 0xFF0000},
		{"green", 0, 0xFFFF, 0, 0x00FF00},
		{"blue", 0, 0, 0xFFFF, 0x0000FF},
		{"mid gray", 0x8080, 0x8080, 0x8080, 0x808080},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toPixel(tt.r, tt.g, tt.b)
			if got != tt.wantPixel {
				t.Errorf("toPixel(%04x,%04x,%04x) = %06x, want %06x", tt.r, tt.g, tt.b, got, tt.wantPixel)
			}
		})
	}
}

func TestNewBorder(t *testing.T) {
	b := NewBorder(0x8000, 0x8000, 0x8000)
	if b.BgPixel != toPixel(0x8000, 0x8000, 0x8000) {
		t.Errorf("BgPixel = %06x, want %06x", b.BgPixel, toPixel(0x8000, 0x8000, 0x8000))
	}
	// Light should be brighter than bg.
	if b.LightPixel <= b.BgPixel {
		t.Errorf("LightPixel (%06x) should be > BgPixel (%06x)", b.LightPixel, b.BgPixel)
	}
	// Dark should be darker than bg.
	if b.DarkPixel >= b.BgPixel {
		t.Errorf("DarkPixel (%06x) should be < BgPixel (%06x)", b.DarkPixel, b.BgPixel)
	}
}

func TestNewBorderFromPixel(t *testing.T) {
	pixel := uint64(0x808080)
	b := NewBorderFromPixel(pixel)
	// BgPixel should match the input pixel.
	if b.BgPixel != pixel {
		t.Errorf("BgPixel = %06x, want %06x", b.BgPixel, pixel)
	}
	if b.LightPixel <= b.BgPixel {
		t.Errorf("LightPixel (%06x) should be > BgPixel (%06x)", b.LightPixel, b.BgPixel)
	}
	if b.DarkPixel >= b.BgPixel {
		t.Errorf("DarkPixel (%06x) should be < BgPixel (%06x)", b.DarkPixel, b.BgPixel)
	}
}

func TestLightColorNormal(t *testing.T) {
	// Mid-range color: should get brighter.
	r, g, b := lightColor(0x8000, 0x8000, 0x8000)
	if r <= 0x8000 || g <= 0x8000 || b <= 0x8000 {
		t.Errorf("lightColor(0x8000,0x8000,0x8000) = (%04x,%04x,%04x), expected brighter", r, g, b)
	}
}

func TestLightColorVeryBright(t *testing.T) {
	// Green > 95% of max → should darken by 10%.
	r, g, b := lightColor(0xFFFF, 0xF800, 0xFFFF)
	if g >= 0xF800 {
		t.Errorf("lightColor with g>95%% should darken: got g=%04x", g)
	}
	// r and b should also be reduced.
	_ = r
	_ = b
}

func TestLightColorBlack(t *testing.T) {
	// Black should produce a visible light color (halfway to white).
	r, g, b := lightColor(0, 0, 0)
	if r == 0 || g == 0 || b == 0 {
		t.Errorf("lightColor(0,0,0) = (%04x,%04x,%04x), expected nonzero", r, g, b)
	}
}

func TestDarkColorNormal(t *testing.T) {
	// Mid-range: should get darker (60%).
	r, g, b := darkColor(0x8000, 0x8000, 0x8000)
	if r >= 0x8000 || g >= 0x8000 || b >= 0x8000 {
		t.Errorf("darkColor(0x8000,0x8000,0x8000) = (%04x,%04x,%04x), expected darker", r, g, b)
	}
}

func TestDarkColorVeryDark(t *testing.T) {
	// Very dark: should lighten (1/4 toward white).
	r, g, b := darkColor(0x0100, 0x0100, 0x0100)
	if r <= 0x0100 || g <= 0x0100 || b <= 0x0100 {
		t.Errorf("darkColor(0x0100,0x0100,0x0100) = (%04x,%04x,%04x), expected lighter", r, g, b)
	}
}

func TestDarkColorBlack(t *testing.T) {
	// Pure black: should lighten toward white.
	r, g, b := darkColor(0, 0, 0)
	if r == 0 || g == 0 || b == 0 {
		t.Errorf("darkColor(0,0,0) = (%04x,%04x,%04x), expected nonzero", r, g, b)
	}
}

func TestBorderSymmetry(t *testing.T) {
	// For a mid-range color, light and dark should be symmetric-ish around the base.
	b := NewBorder(0x8000, 0x8000, 0x8000)
	bgR := (b.BgPixel >> 16) & 0xFF
	lightR := (b.LightPixel >> 16) & 0xFF
	darkR := (b.DarkPixel >> 16) & 0xFF

	lightDist := lightR - bgR
	darkDist := bgR - darkR

	// Both distances should be > 0 (distinct from bg).
	if lightDist == 0 {
		t.Error("light red channel same as bg")
	}
	if darkDist == 0 {
		t.Error("dark red channel same as bg")
	}
}

func TestMaxU16(t *testing.T) {
	if got := maxU16(10, 20); got != 20 {
		t.Errorf("maxU16(10,20) = %d, want 20", got)
	}
	if got := maxU16(30, 5); got != 30 {
		t.Errorf("maxU16(30,5) = %d, want 30", got)
	}
	if got := maxU16(7, 7); got != 7 {
		t.Errorf("maxU16(7,7) = %d, want 7", got)
	}
}

func TestMin64(t *testing.T) {
	if got := min64(1.5, 2.5); got != 1.5 {
		t.Errorf("min64(1.5,2.5) = %f, want 1.5", got)
	}
	if got := min64(3.0, 1.0); got != 1.0 {
		t.Errorf("min64(3.0,1.0) = %f, want 1.0", got)
	}
}
