package window

import "testing"

func TestValues(t *testing.T) {
	var nilWin *Window
	a, b, c := new(ValueKey), new(ValueKey), new(ValueKey)
	if nilWin.Value(a) != nil {
		t.Fatal("a nil window has a value")
	}
	w := &Window{}
	w.SetValue(a, nil) // removing what is not there is a no-op
	w.SetValue(a, 1)
	w.SetValue(b, 2)
	w.SetValue(c, 3)
	w.SetValue(b, 20)
	if w.Value(a) != 1 || w.Value(b) != 20 || w.Value(c) != 3 {
		t.Fatalf("values = %v %v %v, want 1 20 3", w.Value(a), w.Value(b), w.Value(c))
	}
	w.SetValue(a, nil) // the last entry moves into a's slot
	if w.Value(a) != nil || w.Value(b) != 20 || w.Value(c) != 3 {
		t.Fatalf("after removing a: %v %v %v, want <nil> 20 3", w.Value(a), w.Value(b), w.Value(c))
	}
	w.SetValue(c, nil)
	w.SetValue(b, nil)
	if len(w.values) != 0 {
		t.Fatalf("%d entries left after removing all", len(w.values))
	}
}
