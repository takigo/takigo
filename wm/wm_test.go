package wm

import "testing"

func TestParseGeometry(t *testing.T) {
	tests := []struct {
		input   string
		w, h    int
		x, y    int
		hasSize bool
		hasPos  bool
		negX    bool
		negY    bool
		wantErr bool
	}{
		// Size and position.
		{"800x600+10+20", 800, 600, 10, 20, true, true, false, false, false},
		{"800x600-10-20", 800, 600, 10, 20, true, true, true, true, false},
		{"800x600+0+0", 800, 600, 0, 0, true, true, false, false, false},
		{"800x600-0+0", 800, 600, 0, 0, true, true, true, false, false},
		{"800x600+10-20", 800, 600, 10, 20, true, true, false, true, false},

		// Position only.
		{"+100+200", 0, 0, 100, 200, false, true, false, false, false},
		{"-50+100", 0, 0, 50, 100, false, true, true, false, false},
		{"-50-100", 0, 0, 50, 100, false, true, true, true, false},

		// With = prefix.
		{"=800x600+10+20", 800, 600, 10, 20, true, true, false, false, false},

		// Size only — no position.
		{"800x600", 800, 600, 0, 0, true, false, false, false, false},

		// Empty — returns zero values, no error.
		{"", 0, 0, 0, 0, false, false, false, false, false},

		// Invalid.
		{"abc", 0, 0, 0, 0, false, false, false, false, true},
		{"800xabc+10+20", 0, 0, 0, 0, false, false, false, false, true},
		{"abcx600+10+20", 0, 0, 0, 0, false, false, false, false, true},
		{"+10", 0, 0, 0, 0, false, false, false, false, true},
	}

	for _, tt := range tests {
		w, h, x, y, hasSize, hasPos, negX, negY, err := ParseGeometry(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseGeometry(%q): err=%v, wantErr=%v", tt.input, err, tt.wantErr)
			continue
		}
		if tt.wantErr {
			continue
		}
		if w != tt.w || h != tt.h {
			t.Errorf("ParseGeometry(%q): size=%dx%d, want %dx%d", tt.input, w, h, tt.w, tt.h)
		}
		if x != tt.x || y != tt.y {
			t.Errorf("ParseGeometry(%q): pos=%d,%d, want %d,%d", tt.input, x, y, tt.x, tt.y)
		}
		if hasSize != tt.hasSize {
			t.Errorf("ParseGeometry(%q): hasSize=%v, want %v", tt.input, hasSize, tt.hasSize)
		}
		if hasPos != tt.hasPos {
			t.Errorf("ParseGeometry(%q): hasPos=%v, want %v", tt.input, hasPos, tt.hasPos)
		}
		if negX != tt.negX || negY != tt.negY {
			t.Errorf("ParseGeometry(%q): neg=%v,%v, want %v,%v", tt.input, negX, negY, tt.negX, tt.negY)
		}
	}
}
