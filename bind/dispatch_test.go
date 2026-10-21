package bind

import (
	"testing"

	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/window"
)

// newPairEngine returns an engine with two toplevel windows, ids 7 and 8.
func newPairEngine(t *testing.T) *Engine {
	t.Helper()
	d := &window.Display{Windows: map[platform.WindowID]*window.Window{}}
	for id, name := range map[platform.WindowID]string{7: "e", 8: "f"} {
		w := &window.Window{PlatformID: id, PathName: "." + name, Name: name, Display: d, Flags: window.FlagTopLevel}
		d.RegisterWindow(id, w)
	}
	e := NewEngine(d)
	for _, w := range d.Windows {
		e.RegisterWindow(w, "Entry")
	}
	return e
}

// TestPromotionSurvivesEventForOtherWindow checks that a multi-event
// binding waits for the window it started in: an event delivered to
// another window leaves the pending pattern alone, as tkBind.c keeps the
// promotion list per window.
func TestPromotionSurvivesEventForOtherWindow(t *testing.T) {
	e := newPairEngine(t)
	var got []string
	if err := e.Bind("all", "<Key-a><Key-b>", func(*EventData) bool {
		got = append(got, "seq")
		return false
	}); err != nil {
		t.Fatal(err)
	}
	e.dispatch(&event.Event{Type: event.KeyPressType, Window: 7, KeySym: 'a'}, false)
	e.dispatch(&event.Event{Type: event.KeyPressType, Window: 8, KeySym: 'x'}, false)
	e.dispatch(&event.Event{Type: event.KeyPressType, Window: 7, KeySym: 'b'}, false)
	if len(got) != 1 || got[0] != "seq" {
		t.Errorf("fired %q, want [\"seq\"]", got)
	}
}

// TestVirtualBindingCompetesWithPhysical checks that a virtual event
// binding competes with a physical binding of the same tag: the virtual
// event's physical definition is compared with the physical binding, and
// the more specific one wins.
func TestVirtualBindingCompetesWithPhysical(t *testing.T) {
	tests := []struct {
		name     string
		physical string
		want     string
	}{
		// The virtual event names <Control-a>, which is more specific
		// than a physical binding without a detail.
		{"virtual beats a detail-less physical", "<Key>", "virtual"},
		// A physical binding for the very same event keeps the match.
		{"physical wins for the same event", "<Control-a>", "physical"},
		// A less specific physical binding does not win.
		{"physical with another detail loses", "<Key-b>", "virtual"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newPairEngine(t)
			var got []string
			bind := func(pattern, name string) {
				if err := e.Bind("all", pattern, func(*EventData) bool {
					got = append(got, name)
					return false
				}); err != nil {
					t.Fatal(err)
				}
			}
			bind(tt.physical, "physical")
			bind("<<SelectAll>>", "virtual")
			e.dispatch(&event.Event{Type: event.KeyPressType, Window: 7, KeySym: 'a', State: platform.ControlMask}, false)
			if len(got) != 1 || got[0] != tt.want {
				t.Errorf("fired %q, want [%q]", got, tt.want)
			}
		})
	}
}

// TestBindingTiePrefersNewest checks that a tie between two equally
// specific bindings of a tag goes to the binding defined last, as
// tkBind.c's MatchPatterns does with the pattern sequence number.
func TestBindingTiePrefersNewest(t *testing.T) {
	e := newPairEngine(t)
	var got []string
	bind := func(pattern, name string) {
		if err := e.Bind("all", pattern, func(*EventData) bool {
			got = append(got, name)
			return false
		}); err != nil {
			t.Fatal(err)
		}
	}
	bind("<Control-Key-a>", "control")
	bind("<Alt-Key-a>", "alt")
	// Both bindings match when Control and Alt are held.
	e.dispatch(&event.Event{Type: event.KeyPressType, Window: 7, KeySym: 'a', State: platform.ControlMask | platform.Mod1Mask}, false)
	if len(got) != 1 || got[0] != "alt" {
		t.Errorf("fired %q, want [\"alt\"]", got)
	}
}

// TestDoubleClickBindingBeatsControlButton checks the specificity order of
// tkBind.c's CountSpecialized: a double click counts twice, so
// <Double-Button-1> is more specific than <Control-Button-1>.
func TestDoubleClickBindingBeatsControlButton(t *testing.T) {
	e := newPairEngine(t)
	var got []string
	bind := func(pattern, name string) {
		if err := e.Bind("all", pattern, func(*EventData) bool {
			got = append(got, name)
			return false
		}); err != nil {
			t.Fatal(err)
		}
	}
	bind("<Control-Button-1>", "control")
	bind("<Double-Button-1>", "double")
	press := func(ms platform.Timestamp) *event.Event {
		return &event.Event{Type: event.ButtonPressType, Window: 7, Button: 1, State: platform.ControlMask, Time: ms, RootX: 10, RootY: 10}
	}
	e.dispatch(press(1000), false)
	e.dispatch(press(1100), false)
	want := []string{"control", "double"}
	if len(got) != len(want) {
		t.Fatalf("fired %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("fired %q, want %q", got, want)
		}
	}
}

// TestCompareSeqFollowsTk checks the comparison of two matching patterns
// against tkBind.c's CountSpecialized and CompareModMasks.
func TestCompareSeqFollowsTk(t *testing.T) {
	// better reports whether snd is the better match than fst, in the
	// order the dispatcher uses.
	greater := func(fst, snd Sequence) bool {
		if cmp := compareSeq(fst, snd); cmp != 0 {
			return cmp > 0
		}
		return compareModMasks(fst, snd) > 0
	}
	tests := []struct {
		name string
		fst  string
		snd  string
		want bool // snd is the better match
	}{
		{"a detail beats no detail", "<Key>", "<Key-a>", true},
		{"no detail loses to a detail", "<Key-a>", "<Key>", false},
		{"a double click beats a plain click", "<Button-1>", "<Double-Button-1>", true},
		{"a triple click beats a double click", "<Double-Button-1>", "<Triple-Button-1>", true},
		{"more modifiers beat fewer", "<Key-a>", "<Control-Key-a>", true},
		{"fewer modifiers lose", "<Control-Key-a>", "<Key-a>", false},
		{"equal patterns tie", "<Control-Key-a>", "<Control-Key-a>", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fst := MustParse(tt.fst)
			snd := MustParse(tt.snd)
			if got := greater(fst, snd); got != tt.want {
				t.Errorf("greater(%q, %q) = %v, want %v", tt.fst, tt.snd, got, tt.want)
			}
		})
	}
}
