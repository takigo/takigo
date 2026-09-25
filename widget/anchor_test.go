package widget

import (
	"testing"

	"github.com/msorc/takigo/option"
)

func TestComputeAnchor(t *testing.T) {
	tests := []struct {
		name                          string
		a                             option.Anchor
		winW, winH, inset, padX, padY int
		innerW, innerH                int
		wantX, wantY                  int
	}{
		// puzzle.tcl button: centring ignores padx even when it does not fit.
		{"center ignores pad", option.AnchorCenter, 29, 29, 0, 11, 4, 8, 17, 10, 6},
		{"west uses pad", option.AnchorW, 100, 30, 2, 5, 1, 20, 10, 7, 10},
		{"south-east", option.AnchorSE, 100, 30, 2, 5, 1, 20, 10, 73, 17},
		{"north", option.AnchorN, 100, 30, 1, 5, 3, 20, 10, 40, 4},
		{"negative truncates to zero", option.AnchorCenter, 10, 10, 0, 0, 0, 13, 13, -1, -1},
	}
	for _, tt := range tests {
		x, y := ComputeAnchor(tt.a, tt.winW, tt.winH, tt.inset, tt.padX, tt.padY, tt.innerW, tt.innerH)
		if x != tt.wantX || y != tt.wantY {
			t.Errorf("%s: got (%d,%d), want (%d,%d)", tt.name, x, y, tt.wantX, tt.wantY)
		}
	}
}
