package event

import (
	"testing"

	"github.com/msorc/takigo/platform"
)

type wheelParser struct {
	platform.EventParser
	ev platform.ButtonEvent
}

func (p wheelParser) ParseButtonEvent(*platform.RawEvent) platform.ButtonEvent { return p.ev }

// Buttons 4-7 are the wheels (TkQueueWindowEvent in tkEvent.c): a press
// becomes a MouseWheel event and the release is dropped.
func TestWheelButtonsBecomeMouseWheel(t *testing.T) {
	tests := []struct {
		button     uint
		wheelDelta int
		wantDelta  int
		wantShift  bool
	}{
		{4, 0, 120, false},
		{5, 0, -120, false},
		{6, 0, 120, true},
		{7, 0, -120, true},
		{4, 30, 30, false},     // a high-resolution step keeps its size
		{5, -240, -240, false}, // two notches at once
		{7, -15, -15, true},
	}
	for _, tt := range tests {
		p := wheelParser{ev: platform.ButtonEvent{Button: tt.button, WheelDelta: tt.wheelDelta, X: 3, Y: 4}}
		press := FromRawEvent(&platform.RawEvent{EventType: platform.ButtonPressEvent}, p)
		if press.Type != MouseWheelType || press.Delta != tt.wantDelta || press.Button != 0 {
			t.Errorf("button %d: type %v delta %d button %d; want MouseWheel, %d, 0",
				tt.button, press.Type, press.Delta, press.Button, tt.wantDelta)
		}
		if got := press.State&platform.ShiftMask != 0; got != tt.wantShift {
			t.Errorf("button %d: shift = %v, want %v", tt.button, got, tt.wantShift)
		}
		if press.X != 3 || press.Y != 4 {
			t.Errorf("button %d: position lost: %d,%d", tt.button, press.X, press.Y)
		}
		release := FromRawEvent(&platform.RawEvent{EventType: platform.ButtonReleaseEvent}, p)
		if release.Type != 0 {
			t.Errorf("button %d: release was not dropped (type %v)", tt.button, release.Type)
		}
	}

	p := wheelParser{ev: platform.ButtonEvent{Button: 1}}
	if ev := FromRawEvent(&platform.RawEvent{EventType: platform.ButtonPressEvent}, p); ev.Type != ButtonPressType || ev.Button != 1 {
		t.Errorf("button 1 changed: type %v button %d", ev.Type, ev.Button)
	}
	if TypeToMask(MouseWheelType) != MouseWheelMask || AllEventsMask&MouseWheelMask == 0 {
		t.Error("MouseWheelMask is not wired into TypeToMask / AllEventsMask")
	}
}

func TestWheelAccumulator(t *testing.T) {
	var a WheelAccumulator
	// One notch up at 3 units per notch scrolls 3 units up (negative).
	if n := a.Units(120, 3); n != -3 {
		t.Errorf("Units(120, 3) = %d, want -3", n)
	}
	if n := a.Units(-240, 3); n != 6 {
		t.Errorf("Units(-240, 3) = %d, want 6", n)
	}
	// Four quarter-notch steps of a high-resolution wheel add up to one
	// notch instead of four zeros.
	total := 0
	for range 4 {
		total += a.Units(-30, 1)
	}
	if total != 1 {
		t.Errorf("four steps of -30 scrolled %d units, want 1", total)
	}
	// Reversing direction uses up the remainder first.
	var b WheelAccumulator
	if n := b.Units(-100, 1); n != 0 {
		t.Errorf("Units(-100, 1) = %d, want 0", n)
	}
	if n := b.Units(100, 1); n != 0 {
		t.Errorf("after reversing: %d, want 0", n)
	}
	if n := b.Units(-120, 1); n != 1 {
		t.Errorf("then a full notch: %d, want 1", n)
	}
}
