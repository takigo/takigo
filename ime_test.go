package takigo_test

import (
	"strings"
	"testing"

	"github.com/msorc/takigo/bind"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/internal/testutil"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget/entry"
	"github.com/msorc/takigo/widget/spinbox"
	"github.com/msorc/takigo/widget/text"
	"github.com/msorc/takigo/window"
)

// The events a backend sends while an input method composes "かな", as the
// Cocoa view does (tkMacOSXKeyEvent.c): each composition is inserted as key
// presses between IMEStart and IMEEnd and cleared before the next one.
func TestIMEMarkedText(t *testing.T) {
	app := testutil.NewTestApp(t)
	type target struct {
		w   *window.Window
		get func() string
	}
	e := entry.New(app, "e")
	sb := spinbox.New(app, "sb")
	tx := text.New(app, "tx")
	te := ttk.NewEntry(app, "te")
	tc := ttk.NewCombobox(app, "tc")
	ts := ttk.NewSpinbox(app, "ts")
	targets := map[string]target{
		"entry":         {e.Window(), e.GetText},
		"spinbox":       {sb.Window(), sb.GetText},
		"text":          {tx.Window(), func() string { return strings.TrimSuffix(tx.Get("1.0", "end"), "\n") }},
		"ttk::entry":    {te.Window(), te.Get},
		"ttk::combobox": {tc.Window(), tc.Get},
		"ttk::spinbox":  {ts.Window(), ts.Get},
	}
	sb.SetText("")
	ts.Set("")
	app.UpdateIdleTasks()
	disp := app.Dispatcher()

	for name, tg := range targets {
		t.Run(name, func(t *testing.T) {
			id := tg.w.PlatformID
			keys := func(s string) {
				for _, r := range s {
					ks := platform.KeySym(r)
					if r >= 0x100 {
						ks |= 0x01000000
					}
					disp.Dispatch(&event.Event{Type: event.KeyPressType, Window: id, KeySym: ks, Str: string(r)})
				}
			}
			virtual := func(n string) { disp.Dispatch(&event.Event{Type: event.VirtualType, Window: id, Name: n}) }
			mark := func(s string) {
				virtual(event.IMEStart)
				keys(s)
				virtual(event.IMEEnd)
			}
			expect := func(step, want string) {
				t.Helper()
				if got := tg.get(); got != want {
					t.Fatalf("%s: text = %q, want %q", step, got, want)
				}
			}

			keys("ab")
			mark("k")
			expect("composing k", "abk")
			virtual(event.IMEClear)
			mark("か")
			virtual(event.IMEClear)
			mark("かn")
			expect("composing かn", "abかn")
			virtual(event.IMEClear)
			keys("かな")
			expect("committed", "abかな")

			keys("e")
			virtual(event.AccentBackspace)
			keys("é")
			expect("accent", "abかなé")
		})
	}

	if e.SelFirst >= 0 {
		t.Errorf("entry selection [%d,%d) left after commit", e.SelFirst, e.SelLast)
	}
}

func TestIMEMarkedTextSelected(t *testing.T) {
	app := testutil.NewTestApp(t)
	e := entry.New(app, "e")
	app.UpdateIdleTasks()
	id := e.Window().PlatformID
	disp := app.Dispatcher()
	started := false
	if err := app.BindEng().Bind(e.Window().PathName, "<<TkStartIMEMarkedText>>", func(*bind.EventData) bool {
		started = true
		return false
	}); err != nil {
		t.Fatal(err)
	}
	disp.Dispatch(&event.Event{Type: event.KeyPressType, Window: id, KeySym: 'a', Str: "a"})
	disp.Dispatch(&event.Event{Type: event.VirtualType, Window: id, Name: event.IMEStart})
	disp.Dispatch(&event.Event{Type: event.KeyPressType, Window: id, KeySym: 'k', Str: "k"})
	disp.Dispatch(&event.Event{Type: event.KeyPressType, Window: id, KeySym: 'a', Str: "a"})
	disp.Dispatch(&event.Event{Type: event.VirtualType, Window: id, Name: event.IMEEnd})
	if !started {
		t.Error("<<TkStartIMEMarkedText>> binding did not run")
	}
	if e.SelFirst != 1 || e.SelLast != 3 {
		t.Errorf("selection = [%d,%d), want the marked text [1,3)", e.SelFirst, e.SelLast)
	}
}
