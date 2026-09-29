package nanosvg

import "testing"

func TestBlendOverEmpty(t *testing.T) {
	if out := BlendOver(nil, 0xffffff); len(out) != 0 {
		t.Errorf("BlendOver(nil) = %d bytes, want 0", len(out))
	}
	Blend(nil, 0, 0, 0, []uint8{1, 2, 3, 4}, 1, 1)
}
