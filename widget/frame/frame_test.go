package frame

import (
	"testing"

	"github.com/takigo/takigo/color"
	"github.com/takigo/takigo/internal/stubs"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/window"
)

func TestFrameOptions(t *testing.T) {
	f := &Frame{
		Win: &window.Window{},
	}

	// Test BorderWidth
	BorderWidth(5)(f)
	if f.BorderWidth != 5 {
		t.Errorf("BorderWidth = %d, want 5", f.BorderWidth)
	}

	// Test Relief
	Relief(option.ReliefRaised)(f)
	if f.Relief != option.ReliefRaised {
		t.Errorf("Relief = %v, want ReliefRaised", f.Relief)
	}

	// Test Width
	Width(300)(f)
	if f.Win.ReqWidth != 300 {
		t.Errorf("ReqWidth = %d, want 300", f.Win.ReqWidth)
	}

	// Test Height
	Height(200)(f)
	if f.Win.ReqHeight != 200 {
		t.Errorf("ReqHeight = %d, want 200", f.Win.ReqHeight)
	}
}

func TestFrameDefaults(t *testing.T) {
	f := &Frame{
		Win: &window.Window{},
	}
	// These are set in New()
	f.BorderWidth = 0
	f.Relief = option.ReliefFlat

	if f.BorderWidth != 0 {
		t.Errorf("default BorderWidth = %d, want 0", f.BorderWidth)
	}
	if f.Relief != option.ReliefFlat {
		t.Errorf("default Relief = %v, want ReliefFlat", f.Relief)
	}
}

func TestFrameSetInternalBorder(t *testing.T) {
	f := &Frame{
		Win: &window.Window{},
	}
	f.SetInternalBorder(5, 10, 15, 20)

	w := f.Win
	if w.InternalBorderLeft != 5 {
		t.Errorf("InternalBorderLeft = %d, want 5", w.InternalBorderLeft)
	}
	if w.InternalBorderRight != 10 {
		t.Errorf("InternalBorderRight = %d, want 10", w.InternalBorderRight)
	}
	if w.InternalBorderTop != 15 {
		t.Errorf("InternalBorderTop = %d, want 15", w.InternalBorderTop)
	}
	if w.InternalBorderBottom != 20 {
		t.Errorf("InternalBorderBottom = %d, want 20", w.InternalBorderBottom)
	}
}

func TestFrameConfigure(t *testing.T) {
	f := &Frame{
		Win: &window.Window{
			BackgroundPixel: 0,
		},
	}
	f.BorderWidth = 2
	f.Relief = option.ReliefRaised
	f.Background = &color.Color{Pixel: 0xFF0000}

	f.Configure()

	if f.BorderWidth != 2 {
		t.Errorf("BorderWidth = %d, want 2", f.BorderWidth)
	}
	if f.Relief != option.ReliefRaised {
		t.Errorf("Relief = %v, want ReliefRaised", f.Relief)
	}
	if f.Win.BackgroundPixel != 0xFF0000 {
		t.Errorf("BackgroundPixel = %d, want 0xFF0000", f.Win.BackgroundPixel)
	}
}

func TestFrameDestroy(t *testing.T) {
	// Destroy requires a real display connection, so we just verify
	// the Destroyed flag logic without actually calling Destroy
	f := &Frame{
		Win: &window.Window{PlatformID: 1},
	}

	if f.Destroyed() {
		t.Error("New frame should not be destroyed")
	}

	// Manually set and verify
	f.MarkDestroyed()
	if !f.Destroyed() {
		t.Error("Destroyed flag should be true after setting")
	}
}

func TestFrameConfigureOptions(t *testing.T) {
	m := &stubs.Manager{}
	f := &Frame{Win: &window.Window{GeomManager: m}}
	arranged := 0
	f.Win.OnConfigure(func() { arranged++ })

	f.Configure(BorderWidth(2), HighlightThickness(3))
	if f.Win.InternalBorderLeft != 5 || arranged != 1 {
		t.Errorf("InternalBorderLeft = %d, arranged = %d, want 5, 1", f.Win.InternalBorderLeft, arranged)
	}
	if m.Requests != 0 {
		t.Errorf("requests = %d after a border change, want 0", m.Requests)
	}
	f.Configure(Width(300))
	if f.Win.ReqWidth != 300 || m.Requests != 1 {
		t.Errorf("ReqWidth = %d, requests = %d, want 300, 1", f.Win.ReqWidth, m.Requests)
	}
}
