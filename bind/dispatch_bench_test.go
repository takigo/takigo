package bind

import (
	"testing"

	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/window"
)

// BenchmarkEngineDispatchKey measures a plain key press through a window's
// four-tag chain with the default virtual events defined and a few bindings
// per tag, none of them virtual matches for the key.
func BenchmarkEngineDispatchKey(b *testing.B) {
	d := &window.Display{Windows: map[platform.WindowID]*window.Window{}}
	root := &window.Window{PlatformID: 1, PathName: ".", Display: d, Flags: window.FlagTopLevel, Class: "Toplevel"}
	w := &window.Window{PlatformID: 2, PathName: ".e", Name: "e", Parent: root, Display: d, Class: "Entry"}
	d.RegisterWindow(1, root)
	d.RegisterWindow(2, w)
	e := NewEngine(d)
	nop := func(*EventData) bool { return false }
	for _, tag := range []string{".e", "Entry", ".", "all"} {
		for _, p := range []string{"<Return>", "<Control-Key-x>", "<<Copy>>", "<<Paste>>", "<Button-1>"} {
			if err := e.Bind(tag, p, nop); err != nil {
				b.Fatal(err)
			}
		}
	}
	ev := &event.Event{Type: event.KeyPressType, Window: 2, KeySym: 'q'}
	b.ReportAllocs()
	for b.Loop() {
		e.dispatch(ev, false)
	}
}

// BenchmarkEngineDispatchSequenceKey is BenchmarkEngineDispatchKey with a
// two-event sequence whose first pattern matches, so every press promotes.
func BenchmarkEngineDispatchSequenceKey(b *testing.B) {
	d := &window.Display{Windows: map[platform.WindowID]*window.Window{}}
	root := &window.Window{PlatformID: 1, PathName: ".", Display: d, Flags: window.FlagTopLevel, Class: "Toplevel"}
	w := &window.Window{PlatformID: 2, PathName: ".e", Name: "e", Parent: root, Display: d, Class: "Entry"}
	d.RegisterWindow(1, root)
	d.RegisterWindow(2, w)
	e := NewEngine(d)
	nop := func(*EventData) bool { return false }
	for _, tag := range []string{".e", "Entry", ".", "all"} {
		for _, p := range []string{"<Return>", "<Key-q><Key-z>", "<<Copy>>", "<Button-1>"} {
			if err := e.Bind(tag, p, nop); err != nil {
				b.Fatal(err)
			}
		}
	}
	ev := &event.Event{Type: event.KeyPressType, Window: 2, KeySym: 'q'}
	b.ReportAllocs()
	for b.Loop() {
		e.dispatch(ev, false)
	}
}
