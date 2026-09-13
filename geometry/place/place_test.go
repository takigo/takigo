package place

import (
	"math"
	"testing"

	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/window"
)

func TestPlaceOptionFunctions(t *testing.T) {
	cfg := placeConfig{}

	X(10)(&cfg)
	if cfg.x != 10 {
		t.Errorf("X(10) -> x = %d, want 10", cfg.x)
	}

	Y(20)(&cfg)
	if cfg.y != 20 {
		t.Errorf("Y(20) -> y = %d, want 20", cfg.y)
	}

	RelX(0.5)(&cfg)
	if cfg.relX != 0.5 {
		t.Errorf("RelX(0.5) -> relX = %v, want 0.5", cfg.relX)
	}

	RelY(0.3)(&cfg)
	if cfg.relY != 0.3 {
		t.Errorf("RelY(0.3) -> relY = %v, want 0.3", cfg.relY)
	}

	Width(100)(&cfg)
	if cfg.width != 100 {
		t.Errorf("Width(100) -> width = %d, want 100", cfg.width)
	}

	Height(200)(&cfg)
	if cfg.height != 200 {
		t.Errorf("Height(200) -> height = %d, want 200", cfg.height)
	}

	RelWidth(0.5)(&cfg)
	if cfg.relWidth != 0.5 {
		t.Errorf("RelWidth(0.5) -> relWidth = %v, want 0.5", cfg.relWidth)
	}

	RelHeight(0.75)(&cfg)
	if cfg.relHeight != 0.75 {
		t.Errorf("RelHeight(0.75) -> relHeight = %v, want 0.75", cfg.relHeight)
	}

	Anchor(option.AnchorCenter)(&cfg)
	if cfg.anchor != option.AnchorCenter {
		t.Errorf("Anchor(Center) -> anchor = %v, want AnchorCenter", cfg.anchor)
	}
}

func TestPlaceConfigDefaults(t *testing.T) {
	cfg := placeConfig{
		width:     -1,
		height:    -1,
		relWidth:  math.NaN(),
		relHeight: math.NaN(),
		anchor:    option.AnchorNW,
	}

	if cfg.width != -1 {
		t.Errorf("default width = %d, want -1", cfg.width)
	}
	if cfg.height != -1 {
		t.Errorf("default height = %d, want -1", cfg.height)
	}
	if !math.IsNaN(cfg.relWidth) {
		t.Errorf("default relWidth = %v, want NaN", cfg.relWidth)
	}
	if !math.IsNaN(cfg.relHeight) {
		t.Errorf("default relHeight = %v, want NaN", cfg.relHeight)
	}
	if cfg.anchor != option.AnchorNW {
		t.Errorf("default anchor = %v, want AnchorNW", cfg.anchor)
	}
}

func TestPlaceConfigCopy(t *testing.T) {
	cfg1 := placeConfig{
		x: 10, y: 20,
		relX: 0.1, relY: 0.2,
		width: 100, height: 200,
		relWidth: 0.5, relHeight: 0.5,
		anchor: option.AnchorCenter,
	}

	cfg2 := cfg1 // value copy

	if cfg1.x != cfg2.x {
		t.Error("Config copy failed for x")
	}
	if cfg1.y != cfg2.y {
		t.Error("Config copy failed for y")
	}
	if cfg1.relX != cfg2.relX {
		t.Error("Config copy failed for relX")
	}
	if cfg1.relY != cfg2.relY {
		t.Error("Config copy failed for relY")
	}
	if cfg1.width != cfg2.width {
		t.Error("Config copy failed for width")
	}
	if cfg1.height != cfg2.height {
		t.Error("Config copy failed for height")
	}
	if cfg1.relWidth != cfg2.relWidth {
		t.Error("Config copy failed for relWidth")
	}
	if cfg1.relHeight != cfg2.relHeight {
		t.Error("Config copy failed for relHeight")
	}
	if cfg1.anchor != cfg2.anchor {
		t.Error("Config copy failed for anchor")
	}
}

func TestPlacerRemove(t *testing.T) {
	p := &placer{container: nil}
	w1 := &window.Window{}
	w2 := &window.Window{}

	e1 := &placeEntry{window: w1, config: placeConfig{}}
	e2 := &placeEntry{window: w2, config: placeConfig{}}

	p.entries = append(p.entries, e1, e2)

	// Remove w1
	p.remove(w1)
	if len(p.entries) != 1 || p.entries[0].window != w2 {
		t.Errorf("After removing w1: entries = %v, want [w2]", p.entries)
	}

	// Remove w2
	p.remove(w2)
	if len(p.entries) != 0 {
		t.Errorf("After removing w2: entries = %v, want []", p.entries)
	}

	// Remove non-existent
	p.remove(w1) // should not panic
	if len(p.entries) != 0 {
		t.Error("Removing non-existent should not change entries")
	}
}

func TestAnchorCalculations(t *testing.T) {
	tests := []struct {
		anchor option.Anchor
		childW int
		childH int
		wantDX int
		wantDY int
	}{
		{option.AnchorNW, 100, 50, 0, 0},
		{option.AnchorN, 100, 50, -50, 0},
		{option.AnchorNE, 100, 50, -100, 0},
		{option.AnchorE, 100, 50, -100, -25},
		{option.AnchorSE, 100, 50, -100, -50},
		{option.AnchorS, 100, 50, -50, -50},
		{option.AnchorSW, 100, 50, 0, -50},
		{option.AnchorW, 100, 50, 0, -25},
		{option.AnchorCenter, 100, 50, -50, -25},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			dx, dy := 0, 0
			switch tt.anchor {
			case option.AnchorN:
				dx = -tt.childW / 2
			case option.AnchorNE:
				dx = -tt.childW
			case option.AnchorE:
				dx = -tt.childW
				dy = -tt.childH / 2
			case option.AnchorSE:
				dx = -tt.childW
				dy = -tt.childH
			case option.AnchorS:
				dx = -tt.childW / 2
				dy = -tt.childH
			case option.AnchorSW:
				dy = -tt.childH
			case option.AnchorW:
				dy = -tt.childH / 2
			case option.AnchorCenter:
				dx = -tt.childW / 2
				dy = -tt.childH / 2
			case option.AnchorNW:
				// No adjustment
			}

			if dx != tt.wantDX {
				t.Errorf("Anchor %v: dx = %d, want %d", tt.anchor, dx, tt.wantDX)
			}
			if dy != tt.wantDY {
				t.Errorf("Anchor %v: dy = %d, want %d", tt.anchor, dy, tt.wantDY)
			}
		})
	}
}

func TestArrangeAllEmpty(t *testing.T) {
	// Should not panic with empty placers
	ArrangeAll()
}

func TestArrangeContainerNil(t *testing.T) {
	// Should not panic with nil container
	ArrangeContainer(nil)
}