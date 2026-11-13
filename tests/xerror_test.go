//go:build linux || freebsd || openbsd || netbsd

package takigo_test

import (
	"testing"
	"time"

	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/internal/xlib"
	"github.com/takigo/takigo/platform"
)

// TestXErrorsDuringRoundTrips: X errors read during a round trip while the
// loop takes events are counted and survived. With a reader goroutine
// blocked in XNextEvent, libX11 aborted the process within a few rounds
// (poll_for_event's assertion on event_waiter).
func TestXErrorsDuringRoundTrips(t *testing.T) {
	app := testutil.NewTestApp(t)
	d := app.Server()
	const rounds, perRound = 100, 20
	before := xlib.ErrorCount()
	app.After(100*time.Millisecond, func() {
		defer app.Quit()
		root := app.Root()
		for i := range rounds {
			for range perRound {
				d.FreePixmap(platform.PixmapID(0x7fffffff)) // BadPixmap
			}
			// ConfigureNotify events keep arriving meanwhile.
			d.MoveResizeWindow(root.PlatformID, 0, 0, uint(200+i%2), 150)
			d.Sync(false)
		}
	})
	app.After(30*time.Second, func() {
		t.Error("the loop did not finish the rounds")
		app.Quit()
	})
	app.MainLoop()
	if n := xlib.ErrorCount() - before; n != rounds*perRound {
		t.Errorf("%d X errors were reported, want %d", n, rounds*perRound)
	}
}
