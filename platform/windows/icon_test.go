//go:build windows

package windows

import (
	"testing"

	"github.com/takigo/takigo/platform"
)

func TestClosestIcon(t *testing.T) {
	mk := func(n int) platform.IconImage {
		return platform.IconImage{Width: n, Height: n, Pix: make([]byte, 4*n*n)}
	}
	icons := []platform.IconImage{mk(48), mk(16), mk(24), {Width: 32, Height: 32}} // the last has no pixels
	if got := closestIcon(icons, 16); got == nil || got.Width != 16 {
		t.Errorf("closest to 16 = %+v", got)
	}
	// 24 and 48 are both candidates for 32; 24 is nearer, the 32 is invalid.
	if got := closestIcon(icons, 32); got == nil || got.Width != 24 {
		t.Errorf("closest to 32 = %+v, want the 24", got)
	}
	// Equally near: the larger wins.
	if got := closestIcon([]platform.IconImage{mk(24), mk(40)}, 32); got.Width != 40 {
		t.Errorf("tie went to %d, want 40", got.Width)
	}
	if closestIcon(nil, 16) != nil {
		t.Error("an icon was chosen from nothing")
	}
}

func TestIconBits(t *testing.T) {
	ic := &platform.IconImage{Width: 17, Height: 2, Pix: make([]byte, 4*17*2)}
	copy(ic.Pix, []byte{1, 2, 3, 4}) // first pixel: R G B A
	and, xor := iconBits(ic)
	if got := xor[:4]; got[0] != 3 || got[1] != 2 || got[2] != 1 || got[3] != 4 {
		t.Errorf("first pixel = %v, want B G R A = [3 2 1 4]", got)
	}
	// 17 pixels need 3 bytes, padded to 4 per row.
	if len(and) != 4*2 || len(xor) != 4*17*2 {
		t.Errorf("mask %d bytes, colour %d bytes; want 8 and 136", len(and), len(xor))
	}
}
