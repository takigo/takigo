package bind

import (
	"fmt"
	"slices"
	"testing"

	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
)

func TestParseButtonPress(t *testing.T) {
	seq, err := Parse("<Button-1>")
	if err != nil {
		t.Fatal(err)
	}
	if len(seq.Patterns) != 1 {
		t.Fatalf("expected 1 pattern, got %d", len(seq.Patterns))
	}
	p := seq.Patterns[0]
	if p.EventType != event.ButtonPressType {
		t.Errorf("EventType = %d, want ButtonPressType", p.EventType)
	}
	if p.Button != 1 {
		t.Errorf("Button = %d, want 1", p.Button)
	}
}

func TestParseKeyA(t *testing.T) {
	seq, err := Parse("<Key-a>")
	if err != nil {
		t.Fatal(err)
	}
	p := seq.Patterns[0]
	if p.EventType != event.KeyPressType {
		t.Errorf("EventType = %d, want KeyPressType", p.EventType)
	}
	if p.KeySym != platform.KeySym(0x0061) {
		t.Errorf("KeySym = %x, want 0x0061", p.KeySym)
	}
}

func TestParseMotion(t *testing.T) {
	seq, err := Parse("<Motion>")
	if err != nil {
		t.Fatal(err)
	}
	if seq.Patterns[0].EventType != event.MotionType {
		t.Errorf("EventType = %d, want MotionType", seq.Patterns[0].EventType)
	}
}

func TestParseEnter(t *testing.T) {
	seq, err := Parse("<Enter>")
	if err != nil {
		t.Fatal(err)
	}
	if seq.Patterns[0].EventType != event.EnterType {
		t.Error("expected EnterType")
	}
}

func TestParseModifiers(t *testing.T) {
	seq, err := Parse("<Control-a>")
	if err != nil {
		t.Fatal(err)
	}
	p := seq.Patterns[0]
	if p.Modifiers&ModControl == 0 {
		t.Error("expected ModControl")
	}
	if p.KeySym != platform.KeySym(0x0061) {
		t.Errorf("KeySym = %x, want 0x0061", p.KeySym)
	}
}

func TestParseControlShift(t *testing.T) {
	seq, err := Parse("<Control-Shift-x>")
	if err != nil {
		t.Fatal(err)
	}
	p := seq.Patterns[0]
	if p.Modifiers&ModControl == 0 {
		t.Error("expected ModControl")
	}
	if p.Modifiers&ModShift == 0 {
		t.Error("expected ModShift")
	}
}

func TestParseAltF4(t *testing.T) {
	seq, err := Parse("<Alt-F4>")
	if err != nil {
		t.Fatal(err)
	}
	p := seq.Patterns[0]
	if p.Modifiers&ModAlt == 0 {
		t.Error("expected ModAlt")
	}
	if p.KeySym != platform.KeySym(0xffc1) {
		t.Errorf("KeySym = %x, want 0xffc1 (F4)", p.KeySym)
	}
}

func TestParseDoubleButton(t *testing.T) {
	seq, err := Parse("<Double-Button-1>")
	if err != nil {
		t.Fatal(err)
	}
	p := seq.Patterns[0]
	if p.Modifiers&ModDouble == 0 {
		t.Error("expected ModDouble")
	}
	if p.Button != 1 {
		t.Errorf("Button = %d, want 1", p.Button)
	}
}

func TestParseTripleButton(t *testing.T) {
	seq, err := Parse("<Triple-Button-1>")
	if err != nil {
		t.Fatal(err)
	}
	p := seq.Patterns[0]
	if p.Modifiers&ModTriple == 0 {
		t.Error("expected ModTriple")
	}
}

func TestParseVirtualEvent(t *testing.T) {
	seq, err := Parse("<<Copy>>")
	if err != nil {
		t.Fatal(err)
	}
	if seq.Patterns[0].Virtual != "Copy" {
		t.Errorf("Virtual = %q, want \"Copy\"", seq.Patterns[0].Virtual)
	}
}

func TestParseVirtualPaste(t *testing.T) {
	seq, err := Parse("<<Paste>>")
	if err != nil {
		t.Fatal(err)
	}
	if seq.Patterns[0].Virtual != "Paste" {
		t.Errorf("Virtual = %q, want \"Paste\"", seq.Patterns[0].Virtual)
	}
}

func TestParseKeyRelease(t *testing.T) {
	seq, err := Parse("<KeyRelease-Escape>")
	if err != nil {
		t.Fatal(err)
	}
	p := seq.Patterns[0]
	if p.EventType != event.KeyReleaseType {
		t.Errorf("EventType = %d, want KeyReleaseType", p.EventType)
	}
	if p.KeySym != platform.XK_Escape {
		t.Errorf("KeySym = %x, want XK_Escape", p.KeySym)
	}
}

func TestParseSingleChar(t *testing.T) {
	seq, err := Parse("a")
	if err != nil {
		t.Fatal(err)
	}
	p := seq.Patterns[0]
	if p.EventType != event.KeyPressType {
		t.Errorf("EventType = %d, want KeyPressType", p.EventType)
	}
	if p.KeySym != platform.KeySym(0x0061) {
		t.Errorf("KeySym = %x, want 0x0061", p.KeySym)
	}
}

func TestParseSingleCharQ(t *testing.T) {
	seq, err := Parse("q")
	if err != nil {
		t.Fatal(err)
	}
	if seq.Patterns[0].KeySym != platform.KeySym(0x0071) {
		t.Errorf("KeySym = %x, want 0x0071", seq.Patterns[0].KeySym)
	}
}

func TestParseErrors(t *testing.T) {
	errors := []string{
		"",           // empty
		"<Unknown>",  // unknown event type
		"<Button-0>", // invalid button number
		"<Button-6>", // button out of range
		"abc",        // multi-char non-pattern
	}
	for _, s := range errors {
		_, err := Parse(s)
		if err == nil {
			t.Errorf("Parse(%q) should return error", s)
		}
	}
}

func TestSpecificity(t *testing.T) {
	tests := []struct {
		pattern string
		want    int
	}{
		{"<Motion>", 0},
		{"<Button-1>", 4},
		{"<Control-a>", 6},              // 4 (keysym) + 2 (control)
		{"<Control-Shift-x>", 8},        // 4 + 2 + 2
		{"<Double-Button-1>", 5},        // 4 + 1
		{"<Control-Shift-Button-1>", 8}, // 4 + 2 + 2
	}
	for _, tt := range tests {
		seq := MustParse(tt.pattern)
		got := seq.Patterns[0].specificity()
		if got != tt.want {
			t.Errorf("specificity(%q) = %d, want %d", tt.pattern, got, tt.want)
		}
	}
}

func TestPatternMatches(t *testing.T) {
	// Button-1 pattern should match a button press event with button 1.
	p := MustParse("<Button-1>").Patterns[0]
	ev := &event.Event{Type: event.ButtonPressType, Button: 1}
	if !p.matches(ev, 0) {
		t.Error("<Button-1> should match ButtonPress with button=1")
	}

	// Should not match button 2.
	ev2 := &event.Event{Type: event.ButtonPressType, Button: 2}
	if p.matches(ev2, 0) {
		t.Error("<Button-1> should not match button=2")
	}

	// Should not match key press.
	ev3 := &event.Event{Type: event.KeyPressType, KeySym: platform.KeySym(0x0061)}
	if p.matches(ev3, 0) {
		t.Error("<Button-1> should not match KeyPress")
	}
}

func TestPatternMatchesModifiers(t *testing.T) {
	p := MustParse("<Control-a>").Patterns[0]

	// Control+a with Control held.
	ev := &event.Event{
		Type:   event.KeyPressType,
		KeySym: platform.KeySym(0x0061),
		State:  platform.ControlMask,
	}
	if !p.matches(ev, 0) {
		t.Error("<Control-a> should match with Control held")
	}

	// Without Control.
	ev2 := &event.Event{
		Type:   event.KeyPressType,
		KeySym: platform.KeySym(0x0061),
		State:  0,
	}
	if p.matches(ev2, 0) {
		t.Error("<Control-a> should not match without Control")
	}
}

func TestPatternMatchesDouble(t *testing.T) {
	p := MustParse("<Double-Button-1>").Patterns[0]
	ev := &event.Event{Type: event.ButtonPressType, Button: 1}

	// Without double click modifier.
	if p.matches(ev, 0) {
		t.Error("<Double-Button-1> should not match without ModDouble")
	}

	// With double click modifier.
	if !p.matches(ev, ModDouble) {
		t.Error("<Double-Button-1> should match with ModDouble")
	}
}

func TestVirtualPatternDoesNotMatchPhysical(t *testing.T) {
	seq := MustParse("<<Copy>>")
	p := seq.Patterns[0]
	ev := &event.Event{Type: event.KeyPressType, KeySym: platform.KeySym(0x0063), State: platform.ControlMask}
	if p.matches(ev, 0) {
		t.Error("virtual pattern should not match physical events")
	}
}

func TestSpecificityControlShiftButton(t *testing.T) {
	// <Control-Shift-Button-1> = 4 (button detail) + 2 (control) + 2 (shift) = 8
	seq := MustParse("<Control-Shift-Button-1>")
	got := seq.Patterns[0].specificity()
	if got != 8 {
		t.Errorf("specificity(<Control-Shift-Button-1>) = %d, want 8", got)
	}
}

// TestSequenceStringRoundTripsKeys checks that String names the key, so
// patterns for different keys differ and the string parses back to the
// same pattern.
func TestSequenceStringRoundTripsKeys(t *testing.T) {
	inputs := []string{"<Key>", "<KeyRelease-Escape>", "<Control-Shift-z>", "<Alt-F4>", "q", "<Key-at>", "<Prior>", "<Next>"}
	for name := range keysymNames {
		inputs = append(inputs, "<Key-"+name+">")
	}
	seen := map[string]string{}
	for _, in := range inputs {
		seq, err := Parse(in)
		if err != nil {
			t.Fatalf("Parse(%q): %v", in, err)
		}
		s := seq.String()
		back, err := Parse(s)
		if err != nil {
			t.Errorf("Parse(%q) (String of %q): %v", s, in, err)
			continue
		}
		if !slices.Equal(back.Patterns, seq.Patterns) {
			t.Errorf("%q -> %q parses to %+v, want %+v", in, s, back.Patterns, seq.Patterns)
		}
		ks := seq.Patterns[0]
		id := fmt.Sprintf("%v/%v/%v", ks.EventType, ks.Modifiers, ks.KeySym)
		if other, ok := seen[s]; ok && other != id {
			t.Errorf("different patterns %s and %s share the string %q", other, id, s)
		}
		seen[s] = id
	}
}
