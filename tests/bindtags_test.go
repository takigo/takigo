package takigo_test

import (
	"testing"

	"github.com/msorc/takigo/bind"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/testutil"
	"github.com/msorc/takigo/widget/button"
)

// A binding on a button's path runs before the Button class behaviour and
// can suppress it with break, as "bind .b <1> break" does in Tk.
func TestPathBindingBreakSuppressesButtonClass(t *testing.T) {
	for _, tt := range []struct {
		name        string
		brk         bool
		wantInvoked bool
	}{
		{"no break", false, true},
		{"break", true, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			app := testutil.NewTestApp(t)
			invoked := false
			b := button.New(app, "b", button.Text("x"), button.Command(func() { invoked = true }))
			pack.Pack(geometry.Group{b})
			app.UpdateIdleTasks()
			ranPath := false
			if err := app.Bind().Bind(b.Window().PathName, "<ButtonPress-1>", func(*bind.EventData) bool {
				ranPath = true
				return tt.brk
			}); err != nil {
				t.Fatal(err)
			}
			id := b.Window().PlatformID
			disp := app.Dispatcher()
			disp.Dispatch(&event.Event{Type: event.ButtonPressType, Window: id, Button: 1, X: 1, Y: 1})
			disp.Dispatch(&event.Event{Type: event.ButtonReleaseType, Window: id, Button: 1, X: 1, Y: 1})
			if !ranPath {
				t.Fatal("path binding did not run")
			}
			if invoked != tt.wantInvoked {
				t.Errorf("command invoked = %v, want %v", invoked, tt.wantInvoked)
			}
		})
	}
}
