//go:build windows

package windows

import (
	"bytes"
	"testing"
)

func TestXBMRowToDIB(t *testing.T) {
	tests := []struct {
		name  string
		bits  []byte
		width int
		want  []byte
	}{
		{
			name:  "one pixel",
			bits:  []byte{0x01},
			width: 8,
			want:  []byte{0x80, 0x00, 0x00, 0x00},
		},
		{
			name:  "pixels 8 and 9 of a row of ten",
			bits:  []byte{0x00, 0x03},
			width: 10,
			want:  []byte{0x00, 0xC0, 0x00, 0x00},
		},
		{
			name:  "a full word",
			bits:  []byte{0xff, 0xff, 0xff, 0xff},
			width: 32,
			want:  []byte{0xff, 0xff, 0xff, 0xff},
		},
		{
			name:  "the last pixel of a word",
			bits:  []byte{0x00, 0x00, 0x00, 0x80},
			width: 32,
			want:  []byte{0x00, 0x00, 0x00, 0x01},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := make([]byte, ((tt.width+31)/32)*4)
			xbmRowToDIB(dst, tt.bits, 0, tt.width)
			if !bytes.Equal(dst, tt.want) {
				t.Errorf("row = % X, want % X", dst, tt.want)
			}
		})
	}
}
