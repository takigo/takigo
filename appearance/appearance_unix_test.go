//go:build !windows && !darwin

package appearance

import (
	"testing"

	"github.com/takigo/takigo/internal/dbus"
)

func TestColorSchemeMode(t *testing.T) {
	tests := []struct {
		name   string
		reply  any
		want   Mode
		wantOK bool
	}{
		{"Read: dark", dbus.Variant{Sig: "v", Value: dbus.Variant{Sig: "u", Value: uint32(1)}}, Dark, true},
		{"Read: light", dbus.Variant{Sig: "v", Value: dbus.Variant{Sig: "u", Value: uint32(2)}}, Light, true},
		{"ReadOne: no preference", dbus.Variant{Sig: "u", Value: uint32(0)}, Light, true},
		{"not a number", dbus.Variant{Sig: "s", Value: "dark"}, Light, false},
		{"nothing", nil, Light, false},
	}
	for _, tt := range tests {
		if got, ok := colorSchemeMode(tt.reply); got != tt.want || ok != tt.wantOK {
			t.Errorf("%s: %v, %v; want %v, %v", tt.name, got, ok, tt.want, tt.wantOK)
		}
	}
}

func TestGTKThemeFallback(t *testing.T) {
	t.Setenv("TAKIGO_APPEARANCE", "")
	t.Setenv("DBUS_SESSION_BUS_ADDRESS", "unix:path=/nonexistent/takigo-test-bus")
	t.Setenv("XDG_RUNTIME_DIR", "")
	t.Setenv("GTK_THEME", "Adwaita:dark")
	if got := System(); got != Dark {
		t.Errorf("System() = %v with GTK_THEME=Adwaita:dark and no bus", got)
	}
	t.Setenv("GTK_THEME", "Adwaita")
	if got := System(); got != Light {
		t.Errorf("System() = %v with GTK_THEME=Adwaita and no bus", got)
	}
}
