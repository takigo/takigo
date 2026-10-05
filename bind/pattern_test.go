package bind

import (
	"fmt"
	"slices"
	"testing"

	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/platform"
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
		"",            // empty
		"<Unknown>",   // unknown event type
		"<Button-0>",  // invalid button number
		"<Button-10>", // button out of range (Tk takes 1-9)
		"<Motion-6>",  // only buttons 1-5 can be held modifiers
		"<Enter-1>",   // button detail on a non-button event
		"abc",         // multi-char non-pattern
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

// TestParseButtonShorthands checks Tk's button forms: a lone digit is a
// button press (<1>, <Double-1>), B1..B5 and <Motion-N> require a held
// button, and <Key-1> is still the digit key.
func TestParseButtonShorthands(t *testing.T) {
	tests := []struct {
		in     string
		typ    event.Type
		button uint
		key    platform.KeySym
		mods   Modifier
	}{
		{in: "<1>", typ: event.ButtonPressType, button: 1},
		{in: "<3>", typ: event.ButtonPressType, button: 3},
		{in: "<Double-1>", typ: event.ButtonPressType, button: 1, mods: ModDouble},
		{in: "<Control-2>", typ: event.ButtonPressType, button: 2, mods: ModControl},
		{in: "<Button-9>", typ: event.ButtonPressType, button: 9},
		{in: "<Key-1>", typ: event.KeyPressType, key: platform.KeySym('1')},
		{in: "1", typ: event.KeyPressType, key: platform.KeySym('1')},
		{in: "<B1-Motion>", typ: event.MotionType, mods: ModButton1},
		{in: "<Button1-Motion>", typ: event.MotionType, mods: ModButton1},
		{in: "<Motion-1>", typ: event.MotionType, mods: ModButton1},
		{in: "<Motion-1-3>", typ: event.MotionType, mods: ModButton1 | ModButton3},
		{in: "<Shift-B2-Motion>", typ: event.MotionType, mods: ModShift | ModButton2},
		{in: "<B1-ButtonRelease-3>", typ: event.ButtonReleaseType, button: 3, mods: ModButton1},
	}
	for _, tt := range tests {
		seq, err := Parse(tt.in)
		if err != nil {
			t.Errorf("Parse(%q): %v", tt.in, err)
			continue
		}
		p := seq.Patterns[0]
		if p.EventType != tt.typ || p.Button != tt.button || p.KeySym != tt.key || p.Modifiers != tt.mods {
			t.Errorf("Parse(%q) = %+v, want type %v button %d key %#x mods %#x",
				tt.in, p, tt.typ, tt.button, tt.key, tt.mods)
		}
		back, err := Parse(seq.String())
		if err != nil || !slices.Equal(back.Patterns, seq.Patterns) {
			t.Errorf("Parse(%q) = %+v, %v; want %+v", seq.String(), back.Patterns, err, seq.Patterns)
		}
	}
}

func TestPatternMatchesHeldButton(t *testing.T) {
	drag := MustParse("<B1-Motion>").Patterns[0]
	plain := MustParse("<Motion>").Patterns[0]
	tests := []struct {
		state       uint
		drag, plain bool
	}{
		{state: 0, drag: false, plain: true},
		{state: platform.Button1Mask, drag: true, plain: true},
		{state: platform.Button3Mask, drag: false, plain: true},
		{state: platform.Button1Mask | platform.ShiftMask, drag: true, plain: true},
	}
	for _, tt := range tests {
		ev := &event.Event{Type: event.MotionType, State: tt.state}
		if got := drag.matches(ev, 0); got != tt.drag {
			t.Errorf("<B1-Motion> with state %#x: matches = %v, want %v", tt.state, got, tt.drag)
		}
		if got := plain.matches(ev, 0); got != tt.plain {
			t.Errorf("<Motion> with state %#x: matches = %v, want %v", tt.state, got, tt.plain)
		}
	}
	if drag.specificity() <= plain.specificity() {
		t.Error("<B1-Motion> should be more specific than <Motion>")
	}
}
