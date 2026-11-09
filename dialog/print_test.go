package dialog

import (
	"slices"
	"testing"
)

func TestWrapLines(t *testing.T) {
	tests := []struct {
		in   string
		wl   int
		want []string
	}{
		{"short", 10, []string{"short"}},
		{"aaa bbb ccc", 7, []string{"aaa bbb", "ccc"}},
		{"aaa  bbb", 4, []string{"aaa ", "bbb"}},
		{"  lead", 10, []string{"  lead"}},
		{"one\ntwo", 10, []string{"one", "two"}},
		{"", 10, nil},
		{"x\n", 10, []string{"x", ""}}, // an empty line stays a line
	}
	for _, tt := range tests {
		if got := wrapLines(tt.in, tt.wl); !slices.Equal(got, tt.want) {
			t.Errorf("wrapLines(%q, %d) = %q, want %q", tt.in, tt.wl, got, tt.want)
		}
	}
}

func TestParseLpoptions(t *testing.T) {
	got := parseLpoptions("copies=1 printer-location='Room 2' printer-state=3 sides")
	want := map[string]string{"copies": "1", "printer-location": "Room 2", "printer-state": "3", "sides": ""}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("parseLpoptions[%q] = %q, want %q (all: %v)", k, got[k], v, got)
		}
	}
}

func TestLprArgs(t *testing.T) {
	o := defaultPrintOptions()
	o.copies, o.media = 2, "letter"
	if got, want := lprArgs("p", "t", o, false), []string{"-o", "copies=2", "-o", "media=letter", "-T", "t", "-P", "p"}; !slices.Equal(got, want) {
		t.Errorf("canvas args = %q, want %q", got, want)
	}
	o.tzoom, o.pprint, o.numberUp = 50, true, 4
	got := lprArgs("p", "t", o, true)
	for _, want := range []string{"cpi=20.0", "lpi=12.0", "prettyprint=true", "number-up=4", "page-left=70"} {
		if !slices.Contains(got, want) {
			t.Errorf("text args %q lack %q", got, want)
		}
	}
	if slices.Index(got, "cpi=20.0") != slices.IndexFunc(got, func(s string) bool { return len(s) > 4 && s[:4] == "cpi=" }) ||
		slices.Contains(got, "cpi=10.0") {
		t.Errorf("pretty print overrode the zoom's cpi: %q", got)
	}
}
