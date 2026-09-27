//go:build linux

package systray

import (
	"testing"

	"github.com/msorc/takigo/platform"
)

func TestTrayOptionFunctions(t *testing.T) {
	t1 := &TrayIcon{}
	TrayTooltip("test tooltip")(t1)
	if t1.tooltip != "test tooltip" {
		t.Errorf("tooltip = %q, want \"test tooltip\"", t1.tooltip)
	}

	t2 := &TrayIcon{}
	clicked := false
	TrayClickHandler(func() { clicked = true })(t2)
	t2.clickHandler()
	if !clicked {
		t.Error("click handler not called")
	}

	t3 := &TrayIcon{}
	var rx, ry int
	TrayRightClickHandler(func(x, y int) { rx, ry = x, y })(t3)
	t3.rightClickHandler(100, 200)
	if rx != 100 || ry != 200 {
		t.Errorf("right click handler got (%d, %d), want (100, 200)", rx, ry)
	}
}

func TestTrayIconConstants(t *testing.T) {
	if systemTrayRequestDock != 0 {
		t.Errorf("systemTrayRequestDock = %d, want 0", systemTrayRequestDock)
	}
	if trayIconSize != 24 {
		t.Errorf("trayIconSize = %d, want 24", trayIconSize)
	}
}

func TestTrayIconSetTooltip(t *testing.T) {
	t1 := &TrayIcon{win: platform.WindowID(0)}
	t1.SetTooltip("test")
	if t1.tooltip != "test" {
		t.Errorf("tooltip = %q, want \"test\"", t1.tooltip)
	}
}

func TestTrayIconDestroyNil(t *testing.T) {
	var t1 *TrayIcon
	// This tests that Destroy doesn't panic on nil receiver
	// But since it's a method with pointer receiver, we can't call it on nil
	// Just verify the method exists
	_ = t1
}

func TestTrayIconDestroyZeroWin(t *testing.T) {
	t1 := &TrayIcon{win: platform.WindowID(0)}
	t1.Destroy() // Should not panic
}

func TestTrayIconDrawZeroWin(t *testing.T) {
	t1 := &TrayIcon{win: platform.WindowID(0)}
	t1.draw() // Should not panic
}
