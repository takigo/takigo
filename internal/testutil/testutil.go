// Package testutil provides helpers for integration tests requiring an X11 display.
package testutil

import (
	"os"
	"testing"

	takigo "github.com/msorc/takigo"
	"github.com/msorc/takigo/internal/displaylock"
)

// RequireDisplay skips the test if no X11 display is available or the
// build has no display backend (CGO_ENABLED=0 on Unix).
func RequireDisplay(t *testing.T) {
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
