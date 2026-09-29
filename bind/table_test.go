package bind

import (
	"testing"

	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/window"
)

// --- BindingTable ---

func TestBindingTableAddAndLookup(t *testing.T) {
	bt := NewBindingTable()
	seq, _ := Parse("<Button-1>")
	bt.Add("Button", seq, func(ev *EventData) bool { return false })

	bindings := bt.Lookup("Button")
	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(bindings))
	}
}

func TestBindingTableAddReplacesSameSequence(t *testing.T) {
	bt := NewBindingTable()
	seq, _ := Parse("<Button-1>")
	other, _ := Parse("<Button-3>")
	bt.Add("Button", seq, func(ev *EventData) bool { return false })
	bt.Add("Button", other, func(ev *EventData) bool { return false })
	before := bt.Lookup("Button")
	bt.Add("Button", seq, func(ev *EventData) bool { return true })

	bindings := bt.Lookup("Button")
	if len(bindings) != 2 {
		t.Fatalf("expected 2 bindings after rebinding, got %d", len(bindings))
	}
	if !bindings[0].handler(nil) {
		t.Error("rebinding did not replace the handler")
	}
	if before[0].handler(nil) {
		t.Error("rebinding modified a slice handed out by Lookup")
	}
}

func TestBindingTableLookupEmpty(t *testing.T) {
	bt := NewBindingTable()
	bindings := bt.Lookup("nonexistent")
	if bindings != nil {
		t.Error("expected nil for nonexistent tag")
	}
}

func TestBindingTableMultipleBindings(t *testing.T) {
	bt := NewBindingTable()
	seq1, _ := Parse("<Button-1>")
	seq2, _ := Parse("<Button-3>")
	bt.Add("Button", seq1, func(ev *EventData) bool { return false })
	bt.Add("Button", seq2, func(ev *EventData) bool { return false })

	bindings := bt.Lookup("Button")
	if len(bindings) != 2 {
		t.Fatalf("expected 2 bindings, got %d", len(bindings))
	}
}

func TestBindingTableRemove(t *testing.T) {
	bt := NewBindingTable()
	seq1, _ := Parse("<Button-1>")
	seq2, _ := Parse("<Button-3>")
	bt.Add("Button", seq1, func(ev *EventData) bool { return false })
	bt.Add("Button", seq2, func(ev *EventData) bool { return false })

	bt.Remove("Button", seq1)
	bindings := bt.Lookup("Button")
	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding after remove, got %d", len(bindings))
	}
}

func TestBindingTableRemoveAll(t *testing.T) {
	bt := NewBindingTable()
	seq, _ := Parse("<Button-1>")
	bt.Add("Button", seq, func(ev *EventData) bool { return false })
	bt.Add("Button", seq, func(ev *EventData) bool { return false })

	bt.RemoveAll("Button")
	bindings := bt.Lookup("Button")
	if bindings != nil {
		t.Error("expected nil after RemoveAll")
	}
}

func TestBindingTableRemoveLastCleans(t *testing.T) {
	bt := NewBindingTable()
	seq, _ := Parse("<Button-1>")
	bt.Add("Button", seq, func(ev *EventData) bool { return false })
	bt.Remove("Button", seq)

	bindings := bt.Lookup("Button")
	if bindings != nil {
		t.Error("removing last binding should clean up tag entry")
	}
}

// --- Pattern matching ---

func TestPatternMatchesKeyPress(t *testing.T) {
	seq, _ := Parse("<Key-a>")
	pat := &seq.Patterns[0]

	ev := &event.Event{
		Type:   event.KeyPressType,
		KeySym: platform.KeySym(0x61), // 'a'
	}
	if !pat.matches(ev, 0) {
		t.Error("Key-a should match keypress 'a'")
	}

	ev.KeySym = platform.KeySym(0x62) // 'b'
	if pat.matches(ev, 0) {
		t.Error("Key-a should not match keypress 'b'")
	}
}

func TestPatternMatchesButtonPress(t *testing.T) {
	seq, _ := Parse("<Button-1>")
	pat := &seq.Patterns[0]

	ev := &event.Event{
		Type:   event.ButtonPressType,
		Button: 1,
	}
	if !pat.matches(ev, 0) {
		t.Error("Button-1 should match button 1 press")
	}

	ev.Button = 2
	if pat.matches(ev, 0) {
		t.Error("Button-1 should not match button 2 press")
	}
}

func TestPatternMatchesControlModifier(t *testing.T) {
	seq, _ := Parse("<Control-c>")
	pat := &seq.Patterns[0]

	ev := &event.Event{
		Type:   event.KeyPressType,
		KeySym: platform.KeySym(0x63), // 'c'
		State:  platform.ControlMask,
	}
	if !pat.matches(ev, 0) {
		t.Error("Control-c should match with ControlMask set")
	}

	ev.State = 0
	if pat.matches(ev, 0) {
		t.Error("Control-c should not match without ControlMask")
	}
}

func TestPatternMatchesDoubleClick(t *testing.T) {
	seq, _ := Parse("<Double-Button-1>")
	pat := &seq.Patterns[0]

	ev := &event.Event{
		Type:   event.ButtonPressType,
		Button: 1,
	}
	// Without double-click modifier
	if pat.matches(ev, 0) {
		t.Error("Double-Button-1 should not match single click")
	}
	// With double-click modifier
	if !pat.matches(ev, ModDouble) {
		t.Error("Double-Button-1 should match with ModDouble")
	}
}

func TestPatternMatchesVirtualReturns(t *testing.T) {
	seq, _ := Parse("<<Copy>>")
	pat := &seq.Patterns[0]

	ev := &event.Event{Type: event.KeyPressType}
	if pat.matches(ev, 0) {
		t.Error("virtual pattern should not match physical events directly")
	}
}

func TestPatternMatchesWrongEventType(t *testing.T) {
	seq, _ := Parse("<Button-1>")
	pat := &seq.Patterns[0]

	ev := &event.Event{
		Type:   event.KeyPressType, // wrong type
		Button: 1,
	}
	if pat.matches(ev, 0) {
		t.Error("button pattern should not match key event")
	}
}

// --- Specificity ---

func TestPatternSpecificity(t *testing.T) {
	tests := []struct {
		pattern string
		wantGt  string // this pattern should be more specific than wantGt
	}{
		{"<Control-Button-1>", "<Button-1>"}, // modifier adds specificity
		{"<Button-1>", "<ButtonPress>"},      // detail adds specificity
		{"<Double-Button-1>", "<Button-1>"},  // double adds specificity
		{"<Control-Shift-a>", "<Control-a>"}, // more modifiers = more specific
	}
	for _, tt := range tests {
		seq1, _ := Parse(tt.pattern)
		seq2, _ := Parse(tt.wantGt)
		s1 := seq1.Patterns[0].specificity()
		s2 := seq2.Patterns[0].specificity()
		if s1 <= s2 {
			t.Errorf("%s (score=%d) should be more specific than %s (score=%d)",
				tt.pattern, s1, tt.wantGt, s2)
		}
	}
}

// --- Tag chain ---

func TestBuildTagChain(t *testing.T) {
	root := &window.Window{
		PathName: ".",
		Flags:    window.FlagTopLevel,
	}
	child := &window.Window{
		PathName: ".frame1",
		Parent:   root,
	}
	btn := &window.Window{
		PathName: ".frame1.button1",
		Parent:   child,
	}

	tags := buildTagChain(btn, "Button")
	// Should be: [pathName, className, toplevelPath, "all"]
	if len(tags) != 4 {
		t.Fatalf("expected 4 tags, got %d: %v", len(tags), tags)
	}
	if tags[0] != ".frame1.button1" {
		t.Errorf("tags[0] = %q, want widget path", tags[0])
	}
	if tags[1] != "Button" {
		t.Errorf("tags[1] = %q, want class name", tags[1])
	}
	if tags[2] != "." {
		t.Errorf("tags[2] = %q, want toplevel path", tags[2])
	}
	if tags[3] != "all" {
		t.Errorf("tags[3] = %q, want \"all\"", tags[3])
	}
}

func TestBuildTagChainToplevel(t *testing.T) {
	root := &window.Window{
		PathName: ".",
		Flags:    window.FlagTopLevel,
	}
	tags := buildTagChain(root, "Toplevel")
	// When widget IS the toplevel, don't duplicate: [., Toplevel, all]
	if len(tags) != 3 {
		t.Fatalf("expected 3 tags for toplevel, got %d: %v", len(tags), tags)
	}
}

func TestLookupSnapshotSurvivesRebinding(t *testing.T) {
	bt := NewBindingTable()
	mk := func(p string) Sequence {
		seq, err := Parse(p)
		if err != nil {
			t.Fatal(err)
		}
		return seq
	}
	bt.Add("Button", mk("<Button-1>"), nil)
	bt.Add("Button", mk("<Button-2>"), nil)
	bt.Add("Button", mk("<Button-3>"), nil)
	snap := bt.Lookup("Button")
	want := []string{snap[0].seq.String(), snap[1].seq.String(), snap[2].seq.String()}

	bt.Remove("Button", mk("<Button-1>"))
	bt.Add("Button", mk("<Enter>"), nil)
	bt.Remove("Button", mk("<Button-3>"))

	for i, w := range want {
		if got := snap[i].seq.String(); got != w {
			t.Errorf("snapshot[%d] = %s after rebinding, want %s", i, got, w)
		}
	}
	if got := len(bt.Lookup("Button")); got != 2 {
		t.Errorf("table has %d bindings, want 2", got)
	}
}

func TestUpdateClickState(t *testing.T) {
	e := &Engine{}
	click := func(typ event.Type, ms, x int) Modifier {
		return e.updateClickState(&event.Event{Type: typ, Window: 1, Button: 1,
			Time: platform.Timestamp(ms), RootX: x, RootY: 10})
	}
	press, release := event.ButtonPressType, event.ButtonReleaseType

	if click(press, 1000, 10) != 0 || click(release, 1050, 10) != 0 {
		t.Fatal("first click counted as a repeat")
	}
	if click(press, 1200, 12) != ModDouble {
		t.Error("second nearby press is not a double click")
	}
	if click(release, 1250, 12) != ModDouble {
		t.Error("second release is not a double click (<Double-ButtonRelease-1>)")
	}
	if click(press, 1400, 30) != 0 {
		t.Error("a press 18px away continued the click count")
	}
}
