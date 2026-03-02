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

func TestApply(t *testing.T) {
	type target struct{ val int }
	tgt := &target{}
	opts := []Option{
		func(t any) { t.(*target).val = 42 },
	}
	Apply(tgt, opts)
	if tgt.val != 42 {
		t.Errorf("Apply: val = %d, want 42", tgt.val)
	}
}

func TestApplyMultiple(t *testing.T) {
	type target struct{ a, b int }
	tgt := &target{}
	opts := []Option{
		func(t any) { t.(*target).a = 1 },
		func(t any) { t.(*target).b = 2 },
	}
	Apply(tgt, opts)
	if tgt.a != 1 || tgt.b != 2 {
		t.Errorf("Apply: a=%d b=%d, want a=1 b=2", tgt.a, tgt.b)
	}
}

func TestApplyEmpty(t *testing.T) {
	type target struct{ val int }
	tgt := &target{val: 10}
	Apply(tgt, nil)
	if tgt.val != 10 {
		t.Errorf("Apply(nil): val = %d, want 10", tgt.val)
	}
}
