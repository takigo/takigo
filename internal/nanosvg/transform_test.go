package nanosvg

import (
	"math"
	"testing"
)

func TestParseTransform(t *testing.T) {
	apply := func(xf [6]float32, x, y float32) (float32, float32) {
		return x*xf[0] + y*xf[2] + xf[4], x*xf[1] + y*xf[3] + xf[5]
	}
	tests := []struct {
		spec         string
		x, y, wx, wy float32
	}{
		{"translate(10,20)", 1, 1, 11, 21},
		{"translate(5)", 1, 1, 6, 1},
		{"scale(2)", 3, 4, 6, 8},
		{"scale(2, 3)", 1, 1, 2, 3},
		{"translate(10 0) scale(2)", 1, 1, 12, 2},
		{"matrix(1,0,0,1,7,8)", 0, 0, 7, 8},
		{"rotate(90)", 1, 0, 0, 1},
		{"rotate(90 1 1)", 2, 1, 1, 2},
	}
	for _, tt := range tests {
		x, y := apply(parseTransform(tt.spec), tt.x, tt.y)
		if math.Abs(float64(x-tt.wx)) > 1e-4 || math.Abs(float64(y-tt.wy)) > 1e-4 {
			t.Errorf("%s maps (%v,%v) to (%v,%v), want (%v,%v)", tt.spec, tt.x, tt.y, x, y, tt.wx, tt.wy)
		}
	}
}

func TestParseStyleAndGroupTransform(t *testing.T) {
	img, err := Parse(`<svg width="20" height="20">
		<g transform="translate(10,10)">
			<rect x="0" y="0" width="5" height="5" fill="#000000" style="fill:#ff0000;stroke:none"/>
		</g></svg>`)
	if err != nil {
		t.Fatal(err)
	}
	px := Rasterize(img, 1, 20, 20)
	at := func(x, y int) []uint8 { i := (y*20 + x) * 4; return px[i : i+4] }
	if p := at(12, 12); p[0] != 0xff || p[1] != 0 || p[3] != 0xff {
		t.Errorf("pixel inside translated rect = %v, want opaque red from style", p)
	}
	if p := at(2, 2); p[3] != 0 {
		t.Errorf("pixel at untranslated position = %v, want transparent", p)
	}
}
