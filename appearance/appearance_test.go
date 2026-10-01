package appearance

import "testing"

func TestEnvironmentOverride(t *testing.T) {
	t.Setenv("TAKIGO_APPEARANCE", "dark")
	if got := System(); got != Dark {
		t.Errorf("System() = %v with TAKIGO_APPEARANCE=dark", got)
	}
	t.Setenv("TAKIGO_APPEARANCE", "LIGHT")
	if got := System(); got != Light {
		t.Errorf("System() = %v with TAKIGO_APPEARANCE=LIGHT", got)
	}
	if Dark.String() != "dark" || Light.String() != "light" {
		t.Errorf("String() = %q, %q", Dark, Light)
	}
}
