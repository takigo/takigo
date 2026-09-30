package takigo

import (
	"os"
	"testing"

	"github.com/msorc/takigo/internal/displaylock"
)

func requireDisplay(t *testing.T) {
	t.Helper()
	if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
		t.Skip("no display available (set DISPLAY or use xvfb-run)")
	}
	displaylock.Acquire(t)
}

func TestNewApp(t *testing.T) {
	requireDisplay(t)

	app, err := NewApp(Title("test"), Size(200, 150))
	if err != nil {
		t.Fatalf("NewApp: %v", err)
	}
	defer app.Destroy()

	if app.Root() == nil {
		t.Error("Root() returned nil")
	}
	if app.Display() == nil {
		t.Error("Display() returned nil")
	}
	if app.ColorCache() == nil {
		t.Error("ColorCache() returned nil")
	}
	if app.FontRegistry() == nil {
		t.Error("FontRegistry() returned nil")
	}
	if app.ImageRegistry() == nil {
		t.Error("ImageRegistry() returned nil")
	}
}

func TestNewAppDestroyCycle(t *testing.T) {
	requireDisplay(t)

	// Create and destroy multiple times to check for leaks/panics.
	for range 3 {
		app, err := NewApp(Title("cycle-test"), Size(100, 100))
		if err != nil {
			t.Fatalf("NewApp: %v", err)
		}
		app.Destroy()
	}
}
