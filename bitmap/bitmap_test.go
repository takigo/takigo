package bitmap

import (
	"image/color"
	"testing"
)

func TestGetKnownBitmaps(t *testing.T) {
	known := []string{"error", "gray12", "gray25", "gray50", "gray75", "hourglass", "info", "questhead", "question", "warning"}

	fg := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	bg := color.RGBA{R: 0, G: 0, B: 255, A: 255}

	for _, name := range known {
		t.Run(name, func(t *testing.T) {
			photo := Get(name, fg, bg)
			if photo == nil {
				t.Errorf("Get(%q) returned nil for known bitmap", name)
				return
			}
			if photo.Width() <= 0 {
				t.Errorf("Get(%q): width = %d, want > 0", name, photo.Width())
			}
			if photo.Height() <= 0 {
				t.Errorf("Get(%q): height = %d, want > 0", name, photo.Height())
			}
			if photo.Name() != "bitmap:"+name {
				t.Errorf("Get(%q): name = %q, want %q", name, photo.Name(), "bitmap:"+name)
			}
		})
	}
}

func TestGetUnknownBitmap(t *testing.T) {
	fg := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	bg := color.RGBA{R: 0, G: 0, B: 255, A: 255}

	photo := Get("nonexistent", fg, bg)
	if photo != nil {
		t.Errorf("Get(\"nonexistent\") = %v, want nil", photo)
	}
}

func TestParseXBM(t *testing.T) {
	src := `#define test_width 8
#define test_height 8
static unsigned char test_bits[] = {
   0xFF, 0x00, 0xFF, 0x00, 0xFF, 0x00, 0xFF, 0x00};`

	w, h, bits := parseXBM(src)
	if w != 8 {
		t.Errorf("width = %d, want 8", w)
	}
	if h != 8 {
		t.Errorf("height = %d, want 8", h)
	}
	if len(bits) != 8 {
		t.Errorf("bits length = %d, want 8", len(bits))
	}
	if bits[0] != 0xFF || bits[1] != 0x00 {
		t.Errorf("bits = %v, want [0xFF, 0x00, ...]", bits[:2])
	}
}

func TestParseXBMInvalid(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		wantW    int
		wantH    int
		wantBits int
	}{
		{"empty", "", 0, 0, 0},
		{"no define", "static unsigned char bits[] = { 0x00 };", 0, 0, 0},
		{"negative width", "#define t_width -8\n#define t_height 1\nstatic char t_bits[] = { 0x00 };", 0, 0, 0},
		{"data too short", "#define t_width 16\n#define t_height 2\nstatic char t_bits[] = { 0x00, 0x01 };", 0, 0, 0},
		{"no braces", "#define test_width 8\n#define test_height 8", 0, 0, 0},
		{"malformed", "#define test_width foo", 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, h, bits := parseXBM(tt.src)
			if w != tt.wantW {
				t.Errorf("width = %d, want %d", w, tt.wantW)
			}
			if h != tt.wantH {
				t.Errorf("height = %d, want %d", h, tt.wantH)
			}
			if len(bits) != tt.wantBits {
				t.Errorf("bits length = %d, want %d", len(bits), tt.wantBits)
			}
		})
	}
}

func TestBuiltinNames(t *testing.T) {
	expected := []string{"error", "gray12", "gray25", "gray50", "gray75", "hourglass", "info", "questhead", "question", "warning"}

	for _, name := range expected {
		if _, ok := builtins[name]; !ok {
			t.Errorf("builtins missing %q", name)
		}
	}

	if len(builtins) != len(expected) {
		t.Errorf("builtins has %d entries, want %d", len(builtins), len(expected))
	}
}

func TestBitmapPhotoColors(t *testing.T) {
	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	blue := color.RGBA{R: 0, G: 0, B: 255, A: 255}

	photo := Get("gray50", red, blue)
	if photo == nil {
		t.Fatal("Get(gray50) returned nil")
	}

	// gray50 is a checkerboard, so we should see both colors
	rgba := photo.RGBA()
	if rgba == nil {
		t.Fatal("photo.RGBA() returned nil")
	}

	// Check a few pixels
	foundRed := false
	foundBlue := false
	for y := 0; y < photo.Height(); y++ {
		for x := 0; x < photo.Width(); x++ {
			c := rgba.At(x, y)
			r, g, b, _ := c.RGBA()
			if r > 0 && g == 0 && b == 0 {
				foundRed = true
			}
			if r == 0 && g == 0 && b > 0 {
				foundBlue = true
			}
		}
	}

	if !foundRed {
		t.Error("Expected red pixels in bitmap")
	}
	if !foundBlue {
		t.Error("Expected blue pixels in bitmap")
	}
}

func TestImagePackageInterface(t *testing.T) {
	// Verify image.Photo interface works
	photo := Get("gray12", color.RGBA{255, 0, 0, 255}, color.RGBA{0, 0, 255, 255})
	if photo == nil {
		t.Fatal("Get returned nil")
	}

	_ = photo.Width()
	_ = photo.Height()
	_ = photo.Name()
	_ = photo.RGBA()
}
