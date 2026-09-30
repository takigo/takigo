package takigo_test

import (
	"testing"

	takigo "github.com/msorc/takigo"
	"github.com/msorc/takigo/internal/testutil"
)

func TestNewApp(t *testing.T) {
	testutil.RequireDisplay(t)

	app, err := takigo.NewApp(takigo.Title("test"), takigo.Size(200, 150))
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
	testutil.RequireDisplay(t)

	// Create and destroy multiple times to check for leaks/panics.
	for range 3 {
		app, err := takigo.NewApp(takigo.Title("cycle-test"), takigo.Size(100, 100))
		if err != nil {
			t.Fatalf("NewApp: %v", err)
		}
		app.Destroy()
	}
}
