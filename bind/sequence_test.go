package bind

import (
	"testing"

	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/window"
)

func newSeqEngine(t *testing.T) *Engine {
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

func key(win platform.WindowID, ks platform.KeySym, state uint) *event.Event {
	return &event.Event{Type: event.KeyPressType, Window: win, KeySym: ks, State: state}
}

func release(win platform.WindowID, ks platform.KeySym) *event.Event {
	return &event.Event{Type: event.KeyReleaseType, Window: win, KeySym: ks}
}

func TestMultiEventSequence(t *testing.T) {
	const escape, controlL = 0xff1b, 0xffe3
	tests := []struct {
		name   string
		events []*event.Event
		want   []string
	}{
		{"sequence completes past release and motion",
			[]*event.Event{key(7, escape, 0), release(7, escape), {Type: event.MotionType, Window: 7}, key(7, 'a', 0)},
			[]string{"seq"}},
		{"single pattern alone", []*event.Event{key(7, 'a', 0)}, []string{"a"}},
		{"another key breaks the sequence",
			[]*event.Event{key(7, escape, 0), key(7, 'b', 0), key(7, 'a', 0)},
			[]string{"a"}},
		{"another window breaks the sequence",
			[]*event.Event{key(7, escape, 0), key(8, 'a', 0)},
			[]string{"a"}},
		{"modifier key between parts",
			[]*event.Event{key(7, 'x', 0), key(7, controlL, 0), key(7, 'c', platform.ControlMask)},
			[]string{"ctrl-seq"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newSeqEngine(t)
			var got []string
			bind := func(pattern, name string) {
				if err := e.Bind("all", pattern, func(*EventData) bool { got = append(got, name); return false }); err != nil {
					t.Fatal(err)
				}
			}
			bind("<Escape><Key-a>", "seq")
			bind("<Key-a>", "a")
			bind("<Key-x><Control-Key-c>", "ctrl-seq")
			for _, ev := range tt.events {
				e.dispatch(ev)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("fired %q, want %q", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("fired %q, want %q", got, tt.want)
				}
			}
		})
	}
}
