package platform

import (
	"testing"
)

func TestAtoms(t *testing.T) {
	a := &Atoms{
		WMName:        AtomID(1),
		String:        AtomID(2),
		WMNormalHints: AtomID(3),
		Primary:       AtomID(4),
		Secondary:     AtomID(5),
		Atom:          AtomID(6),
		Cardinal:      AtomID(7),
		Window:        AtomID(8),
	}

	if a.WMName != 1 {
		t.Errorf("WMName = %d, want 1", a.WMName)
	}
	if a.String != 2 {
		t.Errorf("String = %d, want 2", a.String)
	}
	if a.WMNormalHints != 3 {
		t.Errorf("WMNormalHints = %d, want 3", a.WMNormalHints)
	}
	if a.Primary != 4 {
		t.Errorf("Primary = %d, want 4", a.Primary)
	}
	if a.Secondary != 5 {
		t.Errorf("Secondary = %d, want 5", a.Secondary)
	}
	if a.Atom != 6 {
		t.Errorf("Atom = %d, want 6", a.Atom)
	}
	if a.Cardinal != 7 {
		t.Errorf("Cardinal = %d, want 7", a.Cardinal)
	}
	if a.Window != 8 {
		t.Errorf("Window = %d, want 8", a.Window)
	}
}

func TestWindowAttrs(t *testing.T) {
	attrs := &WindowAttrs{
		BackgroundPixel:  0xFFFFFF,
		BorderPixel:      0x000000,
		BitGravity:       0,
		EventMask:        0xFFFFFFFF,
		OverrideRedirect: true,
	}

	if attrs.BackgroundPixel != 0xFFFFFF {
		t.Errorf("BackgroundPixel = %d, want 0xFFFFFF", attrs.BackgroundPixel)
	}
	if attrs.BorderPixel != 0x000000 {
		t.Errorf("BorderPixel = %d, want 0x000000", attrs.BorderPixel)
	}
	if attrs.BitGravity != 0 {
		t.Errorf("BitGravity = %d, want 0", attrs.BitGravity)
	}
	if attrs.EventMask != 0xFFFFFFFF {
		t.Errorf("EventMask = %d, want 0xFFFFFFFF", attrs.EventMask)
	}
	if !attrs.OverrideRedirect {
		t.Error("OverrideRedirect should be true")
	}
}

func TestGCValues(t *testing.T) {
	vals := &GCValues{
		Foreground: 0x000000,
		Background: 0xFFFFFF,
		LineWidth:  2,
		Function:   3,
	}

	if vals.Foreground != 0x000000 {
		t.Errorf("Foreground = %d, want 0x000000", vals.Foreground)
	}
	if vals.Background != 0xFFFFFF {
		t.Errorf("Background = %d, want 0xFFFFFF", vals.Background)
	}
	if vals.LineWidth != 2 {
		t.Errorf("LineWidth = %d, want 2", vals.LineWidth)
	}
	if vals.Function != 3 {
		t.Errorf("Function = %d, want 3", vals.Function)
	}
}

func TestPoint(t *testing.T) {
	p := Point{X: 10, Y: 20}
	if p.X != 10 {
		t.Errorf("X = %d, want 10", p.X)
	}
	if p.Y != 20 {
		t.Errorf("Y = %d, want 20", p.Y)
	}
}

func TestSizeHints(t *testing.T) {
	hints := &SizeHints{
		Flags:      0x1F,
		X:          100,
		Y:          200,
		Width:      800,
		Height:     600,
		MinWidth:   400,
		MinHeight:  300,
		MaxWidth:   1600,
		MaxHeight:  1200,
		WidthInc:   10,
		HeightInc:  10,
		WinGravity: 1,
	}

	if hints.Flags != 0x1F {
		t.Errorf("Flags = %d, want 0x1F", hints.Flags)
	}
	if hints.X != 100 {
		t.Errorf("X = %d, want 100", hints.X)
	}
	if hints.Y != 200 {
		t.Errorf("Y = %d, want 200", hints.Y)
	}
	if hints.Width != 800 {
		t.Errorf("Width = %d, want 800", hints.Width)
	}
	if hints.Height != 600 {
		t.Errorf("Height = %d, want 600", hints.Height)
	}
	if hints.MinWidth != 400 {
		t.Errorf("MinWidth = %d, want 400", hints.MinWidth)
	}
	if hints.MinHeight != 300 {
		t.Errorf("MinHeight = %d, want 300", hints.MinHeight)
	}
	if hints.MaxWidth != 1600 {
		t.Errorf("MaxWidth = %d, want 1600", hints.MaxWidth)
	}
	if hints.MaxHeight != 1200 {
		t.Errorf("MaxHeight = %d, want 1200", hints.MaxHeight)
	}
	if hints.WidthInc != 10 {
		t.Errorf("WidthInc = %d, want 10", hints.WidthInc)
	}
	if hints.HeightInc != 10 {
		t.Errorf("HeightInc = %d, want 10", hints.HeightInc)
	}
	if hints.WinGravity != 1 {
		t.Errorf("WinGravity = %d, want 1", hints.WinGravity)
	}
}

func TestWMHints(t *testing.T) {
	hints := &WMHints{
		Flags:        0x03,
		Input:        true,
		InitialState: 1,
	}

	if hints.Flags != 0x03 {
		t.Errorf("Flags = %d, want 0x03", hints.Flags)
	}
	if !hints.Input {
		t.Error("Input should be true")
	}
	if hints.InitialState != 1 {
		t.Errorf("InitialState = %d, want 1", hints.InitialState)
	}
}

func TestRawEvent(t *testing.T) {
	ev := &RawEvent{
		Data:        nil,
		EventType:   2,
		EventWindow: WindowID(42),
	}

	if ev.EventType != 2 {
		t.Errorf("EventType = %d, want 2", ev.EventType)
	}
	if ev.EventWindow != 42 {
		t.Errorf("EventWindow = %d, want 42", ev.EventWindow)
	}
}

func TestEventParser(_ *testing.T) {
	var _ EventParser
}

func TestDisplayCoreInterface(_ *testing.T) {
	// Verify DisplayCore interface exists and has expected methods
	var dc DisplayCore
	_ = dc
}

func TestDisplayServerInterface(_ *testing.T) {
	// Verify DisplayServer interface exists and embeds DisplayCore
	var ds DisplayServer
	_ = ds
}

func TestDisplayServerComposition(_ *testing.T) {
	// Test that displayServer struct composes all interfaces
	var s displayServer
	_ = s
}

func TestNewDisplayServer(t *testing.T) {
	// Test that NewDisplayServer returns a DisplayServer
	var core DisplayCore
	var wm WindowManager
	var drawer Drawer
	var gc GCManager
	var pixmap PixmapManager
	var eventSrc EventSource
	var grab GrabManager
	var sel SelectionManager
	var cursor CursorManager
	var prop PropertyManager
	var im InputMethodManager

	ds := NewDisplayServer(core, wm, drawer, gc, pixmap, eventSrc, grab, sel, cursor, prop, im)
	if ds == nil {
		t.Error("NewDisplayServer should return non-nil DisplayServer")
	}
}
func TestRuneToKeySym(t *testing.T) {
	tests := []struct {
		r    rune
		want KeySym
	}{
		{'a', 0x61}, {'A', 0x41}, {'é', 0xe9}, {'€', 0x010020ac},
		{'😀', 0x0101f600}, {'\x01', 0}, {0x7f, 0}, {0xd83d, 0},
	}
	for _, tt := range tests {
		if got := RuneToKeySym(tt.r); got != tt.want {
			t.Errorf("RuneToKeySym(%U) = %#x, want %#x", tt.r, got, tt.want)
		}
		if tt.want != 0 && KeySymToRune(tt.want) != tt.r {
			t.Errorf("KeySymToRune(%#x) does not round-trip to %U", tt.want, tt.r)
		}
	}
}

type pumpingCore struct {
	DisplayCore
	pumped int
}

func (c *pumpingCore) PumpEvents() bool { c.pumped++; return c.pumped > 1 }

// The event loop finds a backend's PumpEvents by type assertion on the
// composed server, so the wrapper must expose it.
func TestNewDisplayServerExposesPumpEvents(t *testing.T) {
	core := &pumpingCore{}
	ds := NewDisplayServer(core, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	p, ok := ds.(interface{ PumpEvents() bool })
	if !ok {
		t.Fatal("composed DisplayServer hides the backend's PumpEvents")
	}
	if p.PumpEvents() || !p.PumpEvents() {
		t.Error("PumpEvents' quit result was not forwarded")
	}
	if core.pumped != 2 {
		t.Errorf("PumpEvents reached the backend %d times, want 2", core.pumped)
	}

	var plain struct{ DisplayCore }
	if _, ok := NewDisplayServer(plain, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil).(interface{ PumpEvents() bool }); ok {
		t.Error("PumpEvents exposed for a backend without one")
	}
}
