package ttk

import "testing"

func TestStateSpecMatches(t *testing.T) {
	tests := []struct {
		name  string
		spec  StateSpec
		state State
		want  bool
	}{
		{"empty spec matches zero", StateSpec{}, 0, true},
		{"empty spec matches any", StateSpec{}, StateActive | StateFocus, true},
		{"on bit set", StateSpec{OnBits: StateActive}, StateActive, true},
		{"on bit set among others", StateSpec{OnBits: StateActive}, StateActive | StateFocus, true},
		{"on bit not set", StateSpec{OnBits: StateActive}, StateFocus, false},
		{"off bit clear", StateSpec{OffBits: StateDisabled}, StateActive, true},
		{"off bit set", StateSpec{OffBits: StateDisabled}, StateDisabled, false},
		{"on+off both satisfied", StateSpec{OnBits: StateActive, OffBits: StateDisabled}, StateActive, true},
		{"on+off conflict", StateSpec{OnBits: StateActive, OffBits: StateDisabled}, StateActive | StateDisabled, false},
		{"multiple on bits all set", StateSpec{OnBits: StateActive | StateFocus}, StateActive | StateFocus | StateHover, true},
		{"multiple on bits partial", StateSpec{OnBits: StateActive | StateFocus}, StateActive, false},
	}
	for _, tt := range tests {
		if got := tt.spec.Matches(tt.state); got != tt.want {
			t.Errorf("%s: Matches(%b) = %v, want %v", tt.name, tt.state, got, tt.want)
		}
	}
}

func TestStateSpecModify(t *testing.T) {
	tests := []struct {
		name  string
		spec  StateSpec
		state State
		want  State
	}{
		{"set on bits", StateSpec{OnBits: StateActive}, 0, StateActive},
		{"clear off bits", StateSpec{OffBits: StateDisabled}, StateDisabled | StateActive, StateActive},
		{"set on clear off", StateSpec{OnBits: StatePressed, OffBits: StateActive}, StateActive, StatePressed},
		{"no change", StateSpec{}, StateActive, StateActive},
	}
	for _, tt := range tests {
		if got := tt.spec.Modify(tt.state); got != tt.want {
			t.Errorf("%s: Modify(%b) = %b, want %b", tt.name, tt.state, got, tt.want)
		}
	}
}

func TestStateMapLookup(t *testing.T) {
	sm := StateMap[string]{
		{Spec: StateSpec{OnBits: StatePressed}, Value: "pressed"},
		{Spec: StateSpec{OnBits: StateActive}, Value: "active"},
		{Spec: StateSpec{}, Value: "default"},
	}

	// Pressed state matches first entry.
	if v, ok := sm.Lookup(StatePressed); !ok || v != "pressed" {
		t.Errorf("Lookup(Pressed) = (%q, %v), want (\"pressed\", true)", v, ok)
	}

	// Active (not pressed) matches second entry.
	if v, ok := sm.Lookup(StateActive); !ok || v != "active" {
		t.Errorf("Lookup(Active) = (%q, %v), want (\"active\", true)", v, ok)
	}

	// Both pressed and active: first match (pressed) wins.
	if v, ok := sm.Lookup(StatePressed | StateActive); !ok || v != "pressed" {
		t.Errorf("Lookup(Pressed|Active) = (%q, %v), want (\"pressed\", true)", v, ok)
	}

	// Zero state matches default (empty spec matches anything).
	if v, ok := sm.Lookup(0); !ok || v != "default" {
		t.Errorf("Lookup(0) = (%q, %v), want (\"default\", true)", v, ok)
	}
}

func TestStateMapLookupNoMatch(t *testing.T) {
	sm := StateMap[int]{
		{Spec: StateSpec{OnBits: StatePressed}, Value: 42},
	}
	v, ok := sm.Lookup(StateActive)
	if ok || v != 0 {
		t.Errorf("Lookup(Active) = (%d, %v), want (0, false)", v, ok)
	}
}
