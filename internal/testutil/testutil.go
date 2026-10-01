// Package testutil provides helpers for integration tests requiring an X11 display.
package testutil

import (
	"image"
	"os"
	"testing"
	"time"

	takigo "github.com/msorc/takigo"
	"github.com/msorc/takigo/internal/displaylock"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/window"
)

// RequireDisplay skips the test if no X11 display is available or the
// build has no display backend (CGO_ENABLED=0 on Unix).
func RequireDisplay(t testing.TB) {
	t.Helper()
	if !haveBackend {
		t.Skip("no display backend in this build (cgo disabled)")
	}
	if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
		t.Skip("no display available (set DISPLAY or use xvfb-run)")
	}
	displaylock.Acquire(t)
}

// NewTestApp creates a takigo App for testing and registers cleanup.
func NewTestApp(t *testing.T) *takigo.App {
	t.Helper()
	RequireDisplay(t)

	app, err := takigo.NewApp(takigo.Title("test"), takigo.Size(200, 150))
	if err != nil {
		t.Fatalf("NewApp: %v", err)
	}
	t.Cleanup(func() {
		app.Destroy()
	})
	return app
}

// Grab maps the App's window, lets it draw, and returns w's pixels as
// non-premultiplied RGBA. It runs the event loop until the grab is done, so
// call it once the widgets are created and managed.
func Grab(t *testing.T, app *takigo.App, w *window.Window) *image.NRGBA {
	t.Helper()
	var img *image.NRGBA
	app.After(200*time.Millisecond, func() {
		defer app.Quit()
		app.UpdateIdleTasks()
		app.Server().Flush()
		pix := app.Server().GetImageRGBA(platform.WindowDrawable(w.PlatformID), 0, 0, w.Width, w.Height)
		if pix == nil {
			return
		}
		img = &image.NRGBA{Pix: pix, Stride: 4 * w.Width, Rect: image.Rect(0, 0, w.Width, w.Height)}
	})
	app.MainLoop()
	if img == nil {
		t.Fatalf("could not read back %s (%dx%d)", w.PathName, w.Width, w.Height)
	}
	return img
}
