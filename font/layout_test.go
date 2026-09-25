package font

import (
	"reflect"
	"testing"
)

// fixedMeasurer measures every character as 8px wide, spaces as 4px.
type fixedMeasurer struct{}

func (fixedMeasurer) MeasureString(s string) int {
	w := 0
	for _, r := range s {
		if r == ' ' {
			w += 4
		} else {
			w += 8
		}
	}
	return w
}

func TestWrapLines(t *testing.T) {
	tests := []struct {
		text string
		wrap int
		want []string
	}{
		{"aaa bbb ccc", 0, []string{"aaa bbb ccc"}},
		{"aaa bbb ccc", 30, []string{"aaa", "bbb", "ccc"}},
		{"aaa bbb ccc", 60, []string{"aaa bbb", "ccc"}},
		{"aaa bbb ccc", 52, []string{"aaa bbb", "ccc"}},
		{"a.  bb cc", 40, []string{"a.  bb", "cc"}},
		{"a.  bb", 40, []string{"a.  bb"}},
		{"abc", 10, []string{"a", "b", "c"}},
		{"one\n\ntwo", 100, []string{"one", "", "two"}},
		{"xx   yy", 20, []string{"xx", "yy"}},
	}
	for _, tt := range tests {
		got := WrapLines(fixedMeasurer{}, tt.text, tt.wrap)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("WrapLines(%q, %d) = %q, want %q", tt.text, tt.wrap, got, tt.want)
		}
	}
}
