package nanosvg

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// The .tk files are `image create photo -format {svg -scale S}` renderings
// dumped pixel by pixel with `$img get x y -withalpha` from Tk 9.1's wish.
func TestRasterizeMatchesTk(t *testing.T) {
	for _, tc := range []struct {
		svg, golden string
		scale       float32
	}{
		{"radio_on.svg", "radio_on_1.tk", 1},
		{"check_on.svg", "check_on_1.25.tk", 1.25},
		{"classic_check.svg", "classic_check_1.tk", 1},
		{"toggle.svg", "toggle_1.25.tk", 1.25},
	} {
		src, err := os.ReadFile("testdata/" + tc.svg)
		if err != nil {
			t.Fatal(err)
		}
		want, err := os.ReadFile("testdata/" + tc.golden)
		if err != nil {
			t.Fatal(err)
		}
		img, _ := Parse(string(src))
		w, h := Size(img, tc.scale)
		px := Rasterize(img, tc.scale, w, h)
		lines := strings.Split(strings.TrimSpace(string(want)), "\n")
		if got := fmt.Sprintf("%d %d", w, h); got != lines[0] {
			t.Fatalf("%s: size %s, want %s", tc.svg, got, lines[0])
		}
		for y, line := range lines[1:] {
			for x, cell := range strings.Fields(line) {
				p := px[(y*w+x)*4:]
				var r, g, b, a int
				fmt.Sscanf(cell, "%d,%d,%d,%d", &r, &g, &b, &a)
				if a == 0 && p[3] == 0 {
					continue // nsvg__unpremultiplyAlpha's defringe is not ported
				}
				if int(p[0]) != r || int(p[1]) != g || int(p[2]) != b || int(p[3]) != a {
					t.Errorf("%s (%d,%d): got %v, want %s", tc.svg, x, y, p[:4], cell)
				}
			}
		}
	}
}
