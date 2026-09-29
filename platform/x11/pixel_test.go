//go:build linux || freebsd || openbsd || netbsd

package x11

import "testing"

func maskedFormat(r, g, b uint64) *pixelFormat {
	pf := &pixelFormat{masked: true}
	for i, m := range [3]uint64{r, g, b} {
		for m>>pf.shift[i]&1 == 0 {
			pf.shift[i]++
		}
		pf.max[i] = m >> pf.shift[i]
	}
	return pf
}

func TestPixelFormatMasks(t *testing.T) {
	rgb565 := maskedFormat(0xf800, 0x07e0, 0x001f)
	rgb101010 := maskedFormat(0x3ff00000, 0xffc00, 0x3ff)
	tests := []struct {
		pf       *pixelFormat
		in, want uint64
	}{
		{rgb565, 0xffffff, 0xffff},
		{rgb565, 0xff0000, 0xf800},
		{rgb565, 0x00ff00, 0x07e0},
		{rgb565, 0x808080, 16<<11 | 32<<5 | 16},
		{rgb101010, 0xffffff, 0x3fffffff},
		{rgb101010, 0x0000ff, 0x3ff},
		{&pixelFormat{identity: true}, 0x123456, 0x123456},
	}
	for _, tt := range tests {
		if got := tt.pf.pixel(tt.in); got != tt.want {
			t.Errorf("pixel(%06x) = %x, want %x", tt.in, got, tt.want)
		}
	}
}
