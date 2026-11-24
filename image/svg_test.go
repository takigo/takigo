package image

import (
	"bytes"
	"errors"
	goimage "image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewPhotoFromSVGFileWithoutConverters(t *testing.T) {
	t.Setenv("PATH", "")
	p, err := NewPhotoFromSVGFile("tcl", "../demos/images/Tcl.svg")
	if err != nil {
		t.Fatal(err)
	}
	if p.Width() != 125 || p.Height() != 265 {
		t.Fatalf("size = %dx%d, want 125x265", p.Width(), p.Height())
	}
	// The feather body is filled #c3b15f through a style attribute inside
	// translated groups.
	tan := 0
	for y := range p.Height() {
		for x := range p.Width() {
			if c := p.rgba.RGBAAt(x, y); c.R == 0xc3 && c.G == 0xb1 && c.B == 0x5f && c.A == 0xff {
				tan++
			}
		}
	}
	if tan < 1000 {
		t.Errorf("%d pixels of the #c3b15f body colour, want the feather filled", tan)
	}
}

// stubSVGConverter replaces the external converter and records that it
// was used.
func stubSVGConverter(t *testing.T, called *bool) {
	t.Helper()
	old := svgRasterizer
	svgRasterizer = func(string) ([]byte, error) {
		*called = true
		return nil, errors.New("no converter in this test")
	}
	t.Cleanup(func() { svgRasterizer = old })
}

func writeSVG(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestNewPhotoFromSVGFileMalformed checks that a document nanosvg cannot
// parse is reported and that the external converter takes over.
func TestNewPhotoFromSVGFileMalformed(t *testing.T) {
	for _, tt := range []struct {
		name, content string
	}{
		{"truncated", `<svg width="10" height="10"><rect x="0" y="0" width="5" height="5"`},
		{"empty", ``},
	} {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			stubSVGConverter(t, &called)
			path := writeSVG(t, "bad.svg", tt.content)

			_, err := NewPhotoFromSVGFile("bad", path)
			if err == nil {
				t.Fatal("NewPhotoFromSVGFile succeeded on a malformed svg")
			}
			if !called {
				t.Error("the external converter was not used")
			}
			if !strings.Contains(err.Error(), "no converter in this test") {
				t.Errorf("error = %v, want the converter's error", err)
			}
		})
	}
}

// TestNewPhotoFromSVGFileValid checks that a document nanosvg can parse
// is rasterized without the external converter.
func TestNewPhotoFromSVGFileValid(t *testing.T) {
	called := false
	stubSVGConverter(t, &called)
	path := writeSVG(t, "good.svg",
		`<svg width="20" height="10"><rect x="0" y="0" width="20" height="10" fill="#ff0000"/></svg>`)

	photo, err := NewPhotoFromSVGFile("good", path)
	if err != nil {
		t.Fatalf("NewPhotoFromSVGFile: %v", err)
	}
	if called {
		t.Error("the external converter was used for a valid svg")
	}
	if photo.Width() != 20 || photo.Height() != 10 {
		t.Errorf("size = %dx%d, want 20x10", photo.Width(), photo.Height())
	}
}

// TestNewPhotoFromSVGFileConverter checks the fallback that reads the
// raster the external converter produced.
func TestNewPhotoFromSVGFileConverter(t *testing.T) {
	old := svgRasterizer
	svgRasterizer = func(string) ([]byte, error) {
		var buf bytes.Buffer
		img := goimage.NewRGBA(goimage.Rect(0, 0, 4, 3))
		if err := png.Encode(&buf, img); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
	t.Cleanup(func() { svgRasterizer = old })

	path := writeSVG(t, "bad.svg", `<svg width="10" height="10"><rect`)
	photo, err := NewPhotoFromSVGFile("bad", path)
	if err != nil {
		t.Fatalf("NewPhotoFromSVGFile: %v", err)
	}
	if photo.Width() != 4 || photo.Height() != 3 {
		t.Errorf("size = %dx%d, want 4x3", photo.Width(), photo.Height())
	}
}

func TestNewPhotoFromSVGScale(t *testing.T) {
	const src = `<svg width="16" height="8" xmlns="http://www.w3.org/2000/svg"><rect width="8" height="8" fill="#ff0000"/></svg>`
	p, err := NewPhotoFromSVG("", src, 1.5)
	if err != nil {
		t.Fatal(err)
	}
	if p.Width() != 24 || p.Height() != 12 {
		t.Fatalf("size = %dx%d, want 24x12", p.Width(), p.Height())
	}
	if c := p.rgba.RGBAAt(2, 2); c.R != 0xff || c.A != 0xff {
		t.Errorf("pixel (2,2) = %v, want opaque red", c)
	}
	if c := p.rgba.RGBAAt(20, 2); c.A != 0 {
		t.Errorf("pixel (20,2) = %v, want transparent", c)
	}
	if _, err := NewPhotoFromSVG("", `<svg xmlns="http://www.w3.org/2000/svg"/>`, 1); err == nil {
		t.Error("an svg without a size gave no error")
	}
}
