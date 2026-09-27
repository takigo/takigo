package cursor

import "testing"

func TestShapeConstants(t *testing.T) {
	tests := []struct {
		name  string
		shape Shape
		want  uint
	}{
		{"Arrow", Arrow, 0},
		{"Crosshair", Crosshair, 1},
		{"Fleur", Fleur, 2},
		{"Hand1", Hand1, 3},
		{"Hand2", Hand2, 4},
		{"LeftPtr", LeftPtr, 5},
		{"Plus", Plus, 6},
		{"QuestionArrow", QuestionArrow, 7},
		{"SBHDoubleArrow", SBHDoubleArrow, 8},
		{"SBVDoubleArrow", SBVDoubleArrow, 9},
		{"SizingAngle", SizingAngle, 10},
		{"TopLeftArrow", TopLeftArrow, 11},
		{"Watch", Watch, 12},
		{"XTerm", XTerm, 13},
		{"BottomRightCorner", BottomRightCorner, 14},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if uint(tt.shape) != tt.want {
				t.Errorf("%s = %d, want %d", tt.name, tt.shape, tt.want)
			}
		})
	}
}

func TestShapeOrdering(t *testing.T) {
	shapes := []Shape{
		Arrow, Crosshair, Fleur, Hand1, Hand2, LeftPtr, Plus,
		QuestionArrow, SBHDoubleArrow, SBVDoubleArrow, SizingAngle,
		TopLeftArrow, Watch, XTerm, BottomRightCorner,
	}

	for i, s := range shapes {
		if uint(s) != uint(i) {
			t.Errorf("Shape %d = %d, want %d", i, s, i)
		}
	}
}

func TestShapeString(t *testing.T) {
	// Shape is just a uint, no String() method, but we can test it converts
	s := Arrow
	if uint(s) != 0 {
		t.Errorf("Arrow = %d, want 0", s)
	}
}
