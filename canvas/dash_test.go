package canvas

import (
	"math"
	"testing"
)

func pieceLength(xy []float32) float64 {
	var n float64
	for i := 2; i+1 < len(xy); i += 2 {
		n += math.Hypot(float64(xy[i]-xy[i-2]), float64(xy[i+1]-xy[i-1]))
	}
	return n
}

func TestDashPieces(t *testing.T) {
	tests := []struct {
		name    string
		xy      []float32
		pattern []byte
		want    [][2]float32 // start and end x of each piece, for horizontal lines
	}{
		{"even pattern", []float32{0, 0, 20, 0}, []byte{4, 2}, [][2]float32{{0, 4}, {6, 10}, {12, 16}, {18, 20}}},
		{"odd pattern repeats twice", []float32{0, 0, 14, 0}, []byte{3}, [][2]float32{{0, 3}, {6, 9}, {12, 14}}},
		{"ends exactly on an off", []float32{0, 0, 6, 0}, []byte{4, 2}, [][2]float32{{0, 4}}},
		{"zero-length elements are skipped", []float32{0, 0, 10, 0}, []byte{0, 0, 5, 5}, [][2]float32{{0, 5}}},
		{"empty pattern draws the whole line", []float32{0, 0, 10, 0}, nil, [][2]float32{{0, 10}}},
		{"all-zero pattern draws the whole line", []float32{0, 0, 10, 0}, []byte{0, 0}, [][2]float32{{0, 10}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dashPieces(tt.xy, tt.pattern)
			if len(got) != len(tt.want) {
				t.Fatalf("%d pieces %v, want %d", len(got), got, len(tt.want))
			}
			for i, p := range got {
				if p[0] != tt.want[i][0] || p[len(p)-2] != tt.want[i][1] {
					t.Errorf("piece %d runs %v..%v, want %v..%v", i, p[0], p[len(p)-2], tt.want[i][0], tt.want[i][1])
				}
			}
		})
	}
}

// A dash that turns a corner keeps the corner, and the pieces add up to
// the "on" share of the path.
func TestDashPiecesAroundCorners(t *testing.T) {
	square := []float32{0, 0, 10, 0, 10, 10, 0, 10, 0, 0}
	pieces := dashPieces(square, []byte{6, 2})
	var on float64
	corner := false
	for _, p := range pieces {
		on += pieceLength(p)
		if len(p) > 4 {
			corner = true
		}
	}
	if math.Abs(on-30) > 1e-4 {
		t.Errorf("on length = %v of a 40 pixel path with a 6-on 2-off pattern, want 30", on)
	}
	if !corner {
		t.Error("no dash runs around a corner")
	}
}

func TestEllipsePoints(t *testing.T) {
	xy := ellipsePoints(50, 40, 20, 10, 0, 360)
	if xy[0] != 70 || xy[1] != 40 {
		t.Errorf("starts at (%v,%v), want (70,40)", xy[0], xy[1])
	}
	for i := 0; i+1 < len(xy); i += 2 {
		dx, dy := float64(xy[i]-50)/20, float64(xy[i+1]-40)/10
		if e := dx*dx + dy*dy; math.Abs(e-1) > 1e-4 {
			t.Fatalf("point %d (%v,%v) is off the ellipse", i/2, xy[i], xy[i+1])
		}
	}
	// A quarter of the way round (90 degrees) is the top, with y pointing down.
	q := len(xy) / 2 / 4 * 2
	if math.Abs(float64(xy[q]-50)) > 1.5 || xy[q+1] > 31 {
		t.Errorf("quarter point (%v,%v) is not at the top", xy[q], xy[q+1])
	}
}
