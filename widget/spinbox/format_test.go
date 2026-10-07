package spinbox

import "testing"

func TestFormatValue(t *testing.T) {
	tests := []struct {
		from, to, incr float64
		format         string
		v              float64
		want           string
	}{
		{0, 100, 1, "", 42, "42"},
		{0, 1, 0.1, "", 0.3, "0.3"},
		{-10, 10, 0.25, "", 2.5, "2.5"},
		{0, 100, 1, "%05.1f", 7, "007.0"},
		{0, 1e7, 1e-3, "", 1, "1.000"},
		{0, 1e-8, 1e-10, "", 5e-9, "5.00e-09"},
	}
	for _, tt := range tests {
		s := &Spinbox{From: tt.from, To: tt.to, Increment: tt.incr, Format: tt.format}
		if got := s.formatValue(tt.v); got != tt.want {
			t.Errorf("from %v to %v incr %v format %q: formatValue(%v) = %q, want %q",
				tt.from, tt.to, tt.incr, tt.format, tt.v, got, tt.want)
		}
	}
}

func TestSyncValuesIndex(t *testing.T) {
	s := &Spinbox{Values: []string{"a", "b", "c"}}
	s.Text = []rune("c")
	s.syncValuesIndex()
	if s.valuesIndex != 2 {
		t.Errorf("valuesIndex = %d after the text became %q, want 2", s.valuesIndex, "c")
	}
	s.Text = []rune("zzz")
	s.syncValuesIndex()
	if s.valuesIndex != 2 {
		t.Errorf("an unknown text moved valuesIndex to %d", s.valuesIndex)
	}
}
