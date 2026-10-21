package takigo

import (
	"strings"
	"testing"

	"github.com/takigo/takigo/appearance"
	"github.com/takigo/takigo/internal/displaylock"
)

func TestResolveAppearance(t *testing.T) {
	if got := resolveAppearance(appConfig{appearance: appearance.Dark}); got != appearance.Dark {
		t.Errorf("an explicit dark App resolved to %v", got)
	}
	if got := resolveAppearance(appConfig{appearance: appearance.Light}); got != appearance.Light {
		t.Errorf("an explicit light App resolved to %v", got)
	}
	t.Setenv("TAKIGO_APPEARANCE", "dark")
	if got := resolveAppearance(appConfig{followSystem: true}); got != appearance.Dark {
		t.Errorf("an App following a dark desktop resolved to %v", got)
	}
}

func TestAppAppearanceIsItsOwnMode(t *testing.T) {
	displaylock.Require(t)
	app, err := NewApp(UseAppearance(appearance.Dark))
	if err != nil {
		if strings.Contains(err.Error(), "needs cgo") {
			t.Skip("the build has no backend:", err)
		}
		t.Fatal(err)
	}
	defer app.Destroy()
	if got := app.Appearance(); got != appearance.Dark {
		t.Errorf("Appearance() = %v, want dark", got)
	}
}

func TestAppHasDragAndDropManager(t *testing.T) {
	displaylock.Require(t)
	app, err := NewApp()
	if err != nil {
		if strings.Contains(err.Error(), "needs cgo") {
			t.Skip("the build has no backend:", err)
		}
		t.Fatal(err)
	}
	defer app.Destroy()
	if app.dnd == nil {
		t.Error("the App has no drag-and-drop manager")
	}
}
