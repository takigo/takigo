package widget

import "testing"

func TestVariableUnsubscribeOutOfOrder(t *testing.T) {
	v := NewVariable(0)
	var got []string
	unsubA := v.OnChange(func(_, _ int) { got = append(got, "a") })
	unsubB := v.OnChange(func(_, _ int) { got = append(got, "b") })
	v.OnChange(func(_, _ int) { got = append(got, "c") })

	unsubA()
	unsubB()
	unsubA()
	v.Set(1)
	if len(got) != 1 || got[0] != "c" {
		t.Errorf("listeners run after unsubscribing a and b: %v, want [c]", got)
	}
}

func TestVariableUnsubscribeDuringNotify(t *testing.T) {
	v := NewVariable(0)
	calls := 0
	var unsub func()
	unsub = v.OnChange(func(_, _ int) { calls++; unsub() })
	v.OnChange(func(_, _ int) { calls++ })
	v.Set(1)
	v.Set(2)
	if calls != 3 {
		t.Errorf("calls = %d, want 3", calls)
	}
}
