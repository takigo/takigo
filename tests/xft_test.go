//go:build linux || freebsd || openbsd || netbsd

package takigo_test

import (
	"sync"
	"testing"
	"time"

	takigo "github.com/takigo/takigo"
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/internal/xlib"
	"github.com/takigo/takigo/platform"
)

// Apps drawing text on their own event loops, then destroyed at the same
// time, must leave libXft usable for the next App. Xft's process-wide state
// has no locks (CI run 36395555374 hung in XftDefaultSubstitute after the
// clipboard test's two Apps), and fonts left in Xft's unreferenced-font
// cache were destroyed after libXrender dropped the display, sending
// requests with a stale opcode.
func TestAppsShareXftSafely(t *testing.T) {
	testutil.RequireDisplay(t)
	for range 10 {
		var wg sync.WaitGroup
		for range 3 {
			a, err := takigo.NewApp(takigo.Size(200, 40))
			if err != nil {
				t.Fatal(err)
			}
			f, err := a.FontRegistry().Get(font.TkDefaultFont)
			if err != nil {
				t.Fatal(err)
			}
			df := f.(platform.DrawableFont)
			d := platform.WindowDrawable(a.Root().PlatformID)
			n := 0
			var draw func()
			draw = func() {
				for range 50 {
					df.DrawString(d, 2, 20, "naïve café 日本", 0, 0, 0, 0)
				}
				if n++; n < 20 {
					a.DoWhenIdle(draw)
				} else {
					a.Quit()
				}
			}
			a.DoWhenIdle(draw)
			wg.Go(a.Run) // Run destroys the App, concurrently with the others
		}
		wg.Wait()

		done := make(chan struct{})
		go func() {
			defer close(done)
			a, err := takigo.NewApp(takigo.Size(20, 20))
			if err != nil {
				t.Error(err)
				return
			}
			if _, err := a.FontRegistry().Get(font.TkFixedFont); err != nil {
				t.Error(err)
			}
			a.Destroy()
		}()
		select {
		case <-done:
		case <-time.After(20 * time.Second):
			t.Fatal("opening a font hung after Apps used Xft")
		}
	}
}

// Drawing text in one App after another must not cause X errors. Fonts that
// XftFontClose parked in Xft's unreferenced-font cache used to be destroyed
// by Xft's XCloseDisplay hook after libXrender had dropped the display, so
// from the second display on every round sent requests with a garbage
// RENDER opcode.
func TestSuccessiveAppsDrawTextWithoutXErrors(t *testing.T) {
	testutil.RequireDisplay(t)
	before := xlib.ErrorCount()
	for range 5 {
		a, err := takigo.NewApp(takigo.Size(200, 40))
		if err != nil {
			t.Fatal(err)
		}
		f, err := a.FontRegistry().Get(font.TkDefaultFont)
		if err != nil {
			t.Fatal(err)
		}
		d := platform.WindowDrawable(a.Root().PlatformID)
		for range 10 {
			f.(platform.DrawableFont).DrawString(d, 2, 20, "naïve café", 0, 0, 0, 0)
		}
		a.Server().Sync(false)
		a.Destroy()
	}
	if n := xlib.ErrorCount() - before; n != 0 {
		t.Errorf("%d X errors while Apps drew text one after another, want 0", n)
	}
}
