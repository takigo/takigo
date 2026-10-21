package scrollbar

import (
	"sync"
	"testing"
	"time"

	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/widget"
)

// TestScrollbarAutoRepeat ports the auto-repeat of ::tk::ScrollSelect
// (tk/library/scrlbar.tcl:212): a held arrow repeats the action every
// -repeatinterval ms once -repeatdelay ms have passed, and the release
// stops it.
func TestScrollbarAutoRepeat(t *testing.T) {
	app := testutil.NewTestApp(t)

	var mu sync.Mutex
	var got []widget.ScrollRequest
	s := New(app, "s",
		CommandOpt(func(r widget.ScrollRequest) {
			mu.Lock()
			got = append(got, r)
			mu.Unlock()
		}),
		RepeatDelayOpt(20), RepeatIntervalOpt(20))

	// A scrollbar of a known size, so the hit test finds the arrows.
	s.Win.Width, s.Win.Height = 15, 100
	s.computeGeometry()

	count := func() int {
		mu.Lock()
		defer mu.Unlock()
		return len(got)
	}

	// Press the up arrow.
	app.Dispatcher().Dispatch(&event.Event{
		Type: event.ButtonPressType, Window: s.Win.PlatformID,
		X: 5, Y: 5, Button: 1,
	})
	if n := count(); n != 1 {
		t.Fatalf("the press produced %d commands, want 1", n)
	}

	// The action repeats while the button is held.
	app.After(300*time.Millisecond, func() { app.Quit() })
	app.MainLoop()
	repeated := count()
	if repeated < 3 {
		t.Fatalf("a held arrow produced %d commands, want the action repeated", repeated)
	}

	// Releasing the button stops the repeat.
	app.Dispatcher().Dispatch(&event.Event{
		Type: event.ButtonReleaseType, Window: s.Win.PlatformID, Button: 1,
	})
	app.After(200*time.Millisecond, func() { app.Quit() })
	app.MainLoop()
	if n := count(); n != repeated {
		t.Errorf("the repeat continued after the release: %d -> %d commands", repeated, n)
	}
}

// TestScrollbarRepeatDefaults checks Tk's defaults for -repeatdelay and
// -repeatinterval (tk/unix/tkUnixDefault.h).
func TestScrollbarRepeatDefaults(t *testing.T) {
	app := testutil.NewTestApp(t)
	s := New(app, "s")
	if s.RepeatDelay != 300 {
		t.Errorf("RepeatDelay = %d, want 300", s.RepeatDelay)
	}
	if s.RepeatInterval != 100 {
		t.Errorf("RepeatInterval = %d, want 100", s.RepeatInterval)
	}
}
