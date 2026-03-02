package text

import "testing"

func TestCompare(t *testing.T) {
	tests := []struct {
		a, b Index
		want int
	}{
		{Index{1, 0}, Index{1, 0}, 0},
		{Index{1, 0}, Index{2, 0}, -1},
		{Index{2, 0}, Index{1, 0}, 1},
		{Index{1, 0}, Index{1, 5}, -1},
		{Index{1, 5}, Index{1, 0}, 1},
		{Index{1, 3}, Index{2, 0}, -1},
		{Index{3, 0}, Index{2, 99}, 1},
	}
	for _, tt := range tests {
		got := Compare(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("Compare(%v, %v) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestClamp(t *testing.T) {
	doc := NewDocument()
	doc.Lines[0].Text = []rune("hello")

	// Before start.
	got := Clamp(Index{0, 0}, doc)
	if got != (Index{1, 0}) {
		t.Errorf("Clamp before start = %v, want {1,0}", got)
	}

	// After end.
	got = Clamp(Index{5, 0}, doc)
	if got != (Index{1, 5}) {
		t.Errorf("Clamp after end = %v, want {1,5}", got)
	}

	// Within bounds.
	got = Clamp(Index{1, 3}, doc)
	if got != (Index{1, 3}) {
		t.Errorf("Clamp within = %v, want {1,3}", got)
	}

	// Char past line end.
	got = Clamp(Index{1, 99}, doc)
	if got != (Index{1, 5}) {
		t.Errorf("Clamp char overflow = %v, want {1,5}", got)
	}

	// Negative char.
	got = Clamp(Index{1, -1}, doc)
	if got != (Index{1, 0}) {
		t.Errorf("Clamp negative char = %v, want {1,0}", got)
	}
}

func TestLineStartEnd(t *testing.T) {
	doc := NewDocument()
	doc.Lines[0].Text = []rune("hello")

	if got := LineStart(1); got != (Index{1, 0}) {
		t.Errorf("LineStart(1) = %v", got)
	}
	if got := LineEnd(1, doc); got != (Index{1, 5}) {
		t.Errorf("LineEnd(1) = %v, want {1,5}", got)
	}
}

func TestParseIndex(t *testing.T) {
	doc := NewDocument()
	doc.Lines[0].Text = []rune("hello")
	doc.Insert(Index{1, 5}, "\nworld")

	tests := []struct {
		spec string
		want Index
		ok   bool
	}{
		{"1.0", Index{1, 0}, true},
		{"end", doc.EndIndex(), true},
		{"2.5", Index{2, 5}, true},
		{"1.end", Index{1, 5}, true},
		{"insert", Index{1, 0}, true}, // default insert mark position
		{"bogus", Index{}, false},
	}
	for _, tt := range tests {
		got, ok := ParseIndex(doc, tt.spec)
		if ok != tt.ok {
			t.Errorf("ParseIndex(%q) ok = %v, want %v", tt.spec, ok, tt.ok)
			continue
		}
		if ok && got != tt.want {
			t.Errorf("ParseIndex(%q) = %v, want %v", tt.spec, got, tt.want)
		}
	}
}

func TestForward(t *testing.T) {
	doc := NewDocument()
	doc.Lines[0].Text = []rune("ab")
	doc.Insert(Index{1, 2}, "\ncd")

	// Forward 1 within line.
	got := Forward(Index{1, 0}, 1, doc)
	if got != (Index{1, 1}) {
		t.Errorf("Forward 1 = %v, want {1,1}", got)
	}

	// Forward across line boundary.
	got = Forward(Index{1, 1}, 2, doc) // 'b' + '\n' = 2 runes to get to line 2 char 0
	if got != (Index{2, 0}) {
		t.Errorf("Forward across line = %v, want {2,0}", got)
	}
}

func TestBackward(t *testing.T) {
	doc := NewDocument()
	doc.Lines[0].Text = []rune("ab")
	doc.Insert(Index{1, 2}, "\ncd")

	// Backward within line.
	got := Backward(Index{2, 2}, 1, doc)
	if got != (Index{2, 1}) {
		t.Errorf("Backward 1 = %v, want {2,1}", got)
	}

	// Backward across line boundary.
	got = Backward(Index{2, 0}, 1, doc) // cross newline to end of line 1
	if got != (Index{1, 2}) {
		t.Errorf("Backward across line = %v, want {1,2}", got)
	}
}
