package bind

import (
	"slices"
	"testing"

	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/window"
)

// TestClassHandlersRunAtClassTag checks that a widget's own input
// handlers run where Tk runs its class bindings: after bindings on the
// widget's path, which can suppress them with break, and not at all when
// bindtags leave out the class tag.
func TestClassHandlersRunAtClassTag(t *testing.T) {
	const top, btn platform.WindowID = 1, 2
	press := func() *event.Event { return &event.Event{Type: event.ButtonPressType, Window: btn, Button: 1} }
	expose := func() *event.Event { return &event.Event{Type: event.ExposeType, Window: btn} }

	tests := []struct {
		name      string
		pathBreak bool
		bindtags  []string
		ev        *event.Event
		want      []string
	}{
		{name: "path, class, all", ev: press(), want: []string{"path", "class", "all"}},
		{name: "break on path suppresses class", pathBreak: true, ev: press(), want: []string{"path"}},
		{name: "bindtags without class", bindtags: []string{".b", "all"}, ev: press(), want: []string{"path", "all"}},
		{name: "class tag moved first", bindtags: []string{"Button", ".b", "all"}, ev: press(), want: []string{"class", "path", "all"}},
		{name: "event handlers run before bindings", pathBreak: true, ev: expose(), want: []string{"handler", "path-expose"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &window.Display{Windows: map[platform.WindowID]*window.Window{}}
			root := &window.Window{PlatformID: top, PathName: ".", Display: d, Flags: window.FlagTopLevel, Class: "Toplevel"}
			b := &window.Window{PlatformID: btn, PathName: ".b", Name: "b", Parent: root, Display: d, Class: "Button"}
			d.RegisterWindow(top, root)
			d.RegisterWindow(btn, b)

			disp := event.NewDispatcher()
			e := NewEngine(d)
			e.Install(disp)
			var got []string
			disp.Bind(btn, event.ButtonPressMask, func(*event.Event) { got = append(got, "class") })
			disp.Bind(btn, event.ExposureMask, func(*event.Event) { got = append(got, "handler") })
			on := func(tag, pattern, name string, brk bool) {
				if err := e.Bind(tag, pattern, func(*EventData) bool { got = append(got, name); return brk }); err != nil {
					t.Fatal(err)
				}
			}
			on(".b", "<Button-1>", "path", tt.pathBreak)
			on(".b", "<Expose>", "path-expose", tt.pathBreak)
			on("all", "<Button-1>", "all", false)
			if tt.bindtags != nil {
				e.SetBindTags(b, tt.bindtags)
			}

			disp.Dispatch(tt.ev)
			if !slices.Equal(got, tt.want) {
				t.Errorf("ran %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClassHandlersRunForUnknownWindows(t *testing.T) {
	d := &window.Display{Windows: map[platform.WindowID]*window.Window{}}
	disp := event.NewDispatcher()
	NewEngine(d).Install(disp)
	ran := false
	disp.Bind(9, event.KeyPressMask, func(*event.Event) { ran = true })
	disp.Dispatch(&event.Event{Type: event.KeyPressType, Window: 9})
	if !ran {
		t.Error("handler of a window the display does not know did not run")
	}
}

func TestChainStopsWhenClassHandlerDestroysWindow(t *testing.T) {
	d := &window.Display{Windows: map[platform.WindowID]*window.Window{}}
	b := &window.Window{PlatformID: 2, PathName: ".b", Display: d, Flags: window.FlagTopLevel, Class: "Button"}
	d.RegisterWindow(2, b)
	disp := event.NewDispatcher()
	e := NewEngine(d)
	e.Install(disp)
	disp.Bind(2, event.ButtonPressMask, func(*event.Event) { b.Flags |= window.FlagAlreadyDead })
	ranAll := false
	if err := e.Bind("all", "<Button-1>", func(*EventData) bool { ranAll = true; return false }); err != nil {
		t.Fatal(err)
	}
	disp.Dispatch(&event.Event{Type: event.ButtonPressType, Window: 2, Button: 1})
	if ranAll {
		t.Error(`"all" binding ran after the class handler destroyed the window`)
	}
}
