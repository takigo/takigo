//go:build darwin

package cocoa

import "testing"

// The X11 backend refuses empty slices; the cocoa one used to index
// element 0 and crash the process.
func TestEmptySlicesAreRefused(t *testing.T) {
	SetDashes(GC(0), 0, nil)
	SetDashes(GC(0), 0, []byte{})
	DrawLines(Drawable(0), GC(0), nil, 0)
	FillPolygon(Drawable(0), GC(0), nil, 0, 0)
	PutImageRGBA(Drawable(0), GC(0), 24, nil, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
}

// A bitmap with fewer bytes than its rows need is refused, so that the
// C side never reads past the Go slice.
func TestCreateBitmapFromDataChecksLength(t *testing.T) {
	if got := CreateBitmapFromData(nil, 4, 2); got != 0 {
		t.Errorf("CreateBitmapFromData(nil, 4, 2) = %d, want 0", got)
	}
	if got := CreateBitmapFromData([]byte{0xFF}, 4, 2); got != 0 {
		t.Errorf("CreateBitmapFromData(short, 4, 2) = %d, want 0", got)
	}
}

func TestXBMBytesNeeded(t *testing.T) {
	cases := []struct {
		width, height uint
		want          uint64
	}{
		{4, 2, 2},
		{8, 1, 1},
		{9, 2, 4},
		{0, 0, 0},
	}
	for _, c := range cases {
		if got := xbmBytesNeeded(c.width, c.height); got != c.want {
			t.Errorf("xbmBytesNeeded(%d, %d) = %d, want %d", c.width, c.height, got, c.want)
		}
	}
}
