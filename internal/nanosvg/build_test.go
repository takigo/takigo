package nanosvg

import (
	"math"
	"testing"
)

func pixel(px []uint8, w, x, y int) [4]uint8 {
	i := (y*w + x) * 4
	return [4]uint8{px[i], px[i+1], px[i+2], px[i+3]}
}

func TestFillPathIsCrispOnPixelEdges(t *testing.T) {
	img := NewImage(10, 10)
	img.FillPath(Polyline([]float32{2, 2, 6, 2, 6, 5, 2, 5}), RGBA(255, 0, 0, 255), false)
	px := Rasterize(img, 1, 10, 10)

	red := [4]uint8{255, 0, 0, 255}
	for y := range 10 {
		for x := range 10 {
			inside := x >= 2 && x < 6 && y >= 2 && y < 5
			got := pixel(px, 10, x, y)
			if inside && got != red {
				t.Errorf("(%d,%d) = %v, want opaque red", x, y, got)
			}
			if !inside && got[3] != 0 {
				t.Errorf("(%d,%d) = %v, want transparent", x, y, got)
			}
		}
	}
}

func TestStrokePathWidthAndColour(t *testing.T) {
	img := NewImage(12, 8)
	// A 1px horizontal line through the centre of row 3.
	img.StrokePath(Polyline([]float32{2, 3.5, 10, 3.5}), false, RGBA(0, 0, 255, 255), 1, CapButt, JoinMiter)
	px := Rasterize(img, 1, 12, 8)
	if got := pixel(px, 12, 6, 3); got != ([4]uint8{0, 0, 255, 255}) {
		t.Errorf("on the line: %v, want opaque blue", got)
	}
	for _, y := range []int{2, 4} {
		if got := pixel(px, 12, 6, y); got[3] != 0 {
			t.Errorf("row %d next to the line: %v, want transparent", y, got)
		}
	}
	if got := pixel(px, 12, 1, 3); got[3] != 0 {
		t.Errorf("before a butt cap: %v, want transparent", got)
	}
}

func TestDiagonalIsAntialiased(t *testing.T) {
	img := NewImage(20, 20)
	img.StrokePath(Polyline([]float32{2, 3, 17, 15}), false, RGBA(0, 0, 0, 255), 1, CapButt, JoinMiter)
	px := Rasterize(img, 1, 20, 20)
	partial := 0
	for i := 3; i < len(px); i += 4 {
		if px[i] > 0 && px[i] < 255 {
			partial++
		}
	}
	if partial < 10 {
		t.Errorf("only %d partly covered pixels on a diagonal: it is not anti-aliased", partial)
	}
}

func TestArcGeometry(t *testing.T) {
	// A quarter arc from 0 to 90 degrees: starts at the right, ends at the top.
	pts := Arc(10, 10, 4, 2, 0, 90)
	if len(pts) != 8 {
		t.Fatalf("a 90 degree arc has %d coordinates, want 8", len(pts))
	}
	near := func(a, b float32) bool { return math.Abs(float64(a-b)) < 1e-4 }
	if !near(pts[0], 14) || !near(pts[1], 10) || !near(pts[6], 10) || !near(pts[7], 8) {
		t.Errorf("arc runs from (%v,%v) to (%v,%v), want (14,10) to (10,8)", pts[0], pts[1], pts[6], pts[7])
	}
	// The midpoint of the cubic lies on the ellipse.
	mx := (pts[0] + 3*pts[2] + 3*pts[4] + pts[6]) / 8
	my := (pts[1] + 3*pts[3] + 3*pts[5] + pts[7]) / 8
	if e := (mx-10)*(mx-10)/16 + (my-10)*(my-10)/4; math.Abs(float64(e-1)) > 0.01 {
		t.Errorf("arc midpoint (%v,%v) is off the ellipse (%v)", mx, my, e)
	}
	if got := len(Ellipse(0, 0, 1, 1)); got != 2+6*4 {
		t.Errorf("an ellipse has %d coordinates, want 26", got)
	}
	if Arc(0, 0, 1, 1, 0, 0) != nil {
		t.Error("a zero-extent arc is not empty")
	}
}

// A region must hold exactly the pixels of the full picture: the canvas
// repaints damaged parts and they have to match a full repaint bit for bit.
func TestRasterizeRegionMatchesFull(t *testing.T) {
	const w, h = 64, 48
	img := NewImage(w, h)
	img.FillPath(Ellipse(30.3, 22.7, 21.1, 13.9), RGBA(200, 30, 30, 255), false)
	img.StrokePath(Polyline([]float32{3.2, 40.1, 20.5, 5.7, 44.9, 39.3, 60.2, 8.8}), false, RGBA(10, 10, 200, 255), 3, CapRound, JoinRound)
	img.FillPath(Polyline([]float32{5, 5, 50, 12, 17, 44}), RGBA(0, 160, 0, 128), true)
	full := Rasterize(img, 1, w, h)

	for _, r := range [][4]int{{0, 0, w, h}, {7, 11, 30, 20}, {33, 0, 31, 48}, {20, 30, 1, 1}, {0, 47, 64, 1}, {60, 40, 20, 20}} {
		x0, y0, rw, rh := r[0], r[1], min(r[2], w-r[0]), min(r[3], h-r[1])
		got := RasterizeRegion(img, 1, w, h, r[0], r[1], r[2], r[3])
		if len(got) != rw*rh*4 {
			t.Errorf("region %v: %d bytes, want %d", r, len(got), rw*rh*4)
			continue
		}
		for y := range rh {
			for x := range rw {
				if a, b := pixel(got, rw, x, y), pixel(full, w, x0+x, y0+y); a != b {
					t.Fatalf("region %v: pixel (%d,%d) = %v, full picture has %v", r, x, y, a, b)
				}
			}
		}
	}
}
