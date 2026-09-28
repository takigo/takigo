package bind

import (
	"testing"

	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/window"
)

func TestVirtualEventChoiceIsDeterministic(t *testing.T) {
	for range 50 {
		d := &window.Display{Windows: map[platform.WindowID]*window.Window{}}
		w := &window.Window{PlatformID: 7, PathName: ".e", Name: "e", Display: d, Flags: window.FlagTopLevel}
		d.RegisterWindow(w.PlatformID, w)

		e := NewEngine(d)
		e.RegisterWindow(w, "Entry")
		// <Control-a> now defines both <<SelectAll>> (a default, defined
		// first) and <<LineStart>>, as in Tk's text and entry bindings.
		if err := e.AddVirtualEvent("LineStart", "<Control-a>"); err != nil {
			t.Fatal(err)
		}
		var got string
		for _, name := range []string{"LineStart", "SelectAll"} {
			if err := e.Bind(".e", "<<"+name+">>", func(*EventData) bool { got = name; return true }); err != nil {
				t.Fatal(err)
			}
		}

		e.dispatch(&event.Event{Type: event.KeyPressType, Window: 7, KeySym: 'a', State: platform.ControlMask}, false)
		if got != "SelectAll" {
			t.Fatalf("<Control-a> dispatched <<%s>>, want <<SelectAll>>", got)
		}
	}
}
