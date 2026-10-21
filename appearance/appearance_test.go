package appearance

import (
	"testing"
	"time"
)

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

// The desktop is asked at most once per lookupTTL, so that a call in an
// event handler does not block on the bus.
func TestLookupIsCached(t *testing.T) {
	var cache lookupCache
	calls := 0
	lookup := func() (Mode, bool) {
		calls++
		return Dark, true
	}
	for range 3 {
		if mode, ok := cache.get(lookup); mode != Dark || !ok {
			t.Fatalf("cache.get = %v, %v; want dark, true", mode, ok)
		}
	}
	if calls != 1 {
		t.Errorf("the desktop was asked %d times, want 1", calls)
	}
	cache.reset()
	if mode, _ := cache.get(lookup); mode != Dark {
		t.Errorf("cache.get = %v after the cache was cleared", mode)
	}
	if calls != 2 {
		t.Errorf("the desktop was asked %d times after the cache was cleared, want 2", calls)
	}
	if lookupTTL < time.Second {
		t.Errorf("lookupTTL = %v, too short to spare the desktop", lookupTTL)
	}
}
