package image

import "testing"

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
