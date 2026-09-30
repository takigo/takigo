package option

import "testing"

func TestReliefStringRoundTrip(t *testing.T) {
	reliefs := []struct {
		r    Relief
		want string
	}{
		{ReliefFlat, "flat"},
		{ReliefRaised, "raised"},
		{ReliefSunken, "sunken"},
		{ReliefGroove, "groove"},
		{ReliefRidge, "ridge"},
		{ReliefSolid, "solid"},
	}
	for _, tt := range reliefs {
		s := tt.r.String()
		if s != tt.want {
			t.Errorf("Relief(%d).String() = %q, want %q", tt.r, s, tt.want)
		}
		parsed := ParseRelief(s)
		if parsed != tt.r {
			t.Errorf("ParseRelief(%q) = %d, want %d", s, parsed, tt.r)
		}
	}
}

func TestReliefUnknown(t *testing.T) {
	if s := Relief(99).String(); s != "flat" {
		t.Errorf("Relief(99).String() = %q, want %q", s, "flat")
	}
	if r := ParseRelief("bogus"); r != ReliefFlat {
		t.Errorf("ParseRelief(\"bogus\") = %d, want ReliefFlat", r)
	}
}

func TestAnchorConstants(t *testing.T) {
	// Verify all anchor constants are distinct.
	anchors := []Anchor{
		AnchorCenter, AnchorN, AnchorNE, AnchorE,
		AnchorSE, AnchorS, AnchorSW, AnchorW, AnchorNW,
	}
	seen := make(map[Anchor]bool)
	for _, a := range anchors {
		if seen[a] {
			t.Errorf("duplicate Anchor value %d", a)
		}
		seen[a] = true
	}
}

func TestJustifyConstants(t *testing.T) {
	// Verify justify constants are distinct.
	justifies := []Justify{JustifyLeft, JustifyCenter, JustifyRight}
	seen := make(map[Justify]bool)
	for _, j := range justifies {
		if seen[j] {
			t.Errorf("duplicate Justify value %d", j)
		}
		seen[j] = true
	}
}
