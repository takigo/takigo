package screenunit

import (
	"math"
	"testing"
)

func TestPxInt(t *testing.T) {
	if got := Px(42); got != 42 {
		t.Errorf("Px(42) = %d, want 42", got)
	}
	if got := Px(0); got != 0 {
		t.Errorf("Px(0) = %d, want 0", got)
	}
}

func TestPxFloat64(t *testing.T) {
	if got := Px(3.7); got != 4 {
		t.Errorf("Px(3.7) = %d, want 4", got)
	}
	if got := Px(3.2); got != 3 {
		t.Errorf("Px(3.2) = %d, want 3", got)
	}
}

func TestPxBareString(t *testing.T) {
	if got := Px("10"); got != 10 {
		t.Errorf("Px(\"10\") = %d, want 10", got)
	}
	if got := Px("3.5"); got != 4 {
		t.Errorf("Px(\"3.5\") = %d, want 4", got)
	}
}

func TestPxWithUnits(t *testing.T) {
	// Set known DPI: 96 DPI = 2540mm wide at 9600px
	// This gives us exactly 9600/2540 ≈ 3.7795 pixels per mm
	SetScreenDPI(9600, 2540)
	defer SetScreenDPI(1920, 508) // restore

	pxPerMM := 9600.0 / 2540.0

	tests := []struct {
		input string
		want  int
	}{
		// Points: 1pt = 25.4/72 mm
		{"1p", int(math.Round(25.4 / 72.0 * pxPerMM))},
		{"72p", int(math.Round(72.0 * 25.4 / 72.0 * pxPerMM))}, // 72pt = 1 inch
		// Millimeters
		{"1m", int(math.Round(pxPerMM))},
		{"10m", int(math.Round(10.0 * pxPerMM))},
		// Centimeters
		{"1c", int(math.Round(10.0 * pxPerMM))},
		{"2.5c", int(math.Round(25.0 * pxPerMM))},
		// Inches
		{"1i", int(math.Round(25.4 * pxPerMM))},
		{"0.5i", int(math.Round(12.7 * pxPerMM))},
	}

	for _, tt := range tests {
		got := Px(tt.input)
		if got != tt.want {
			t.Errorf("Px(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestPx72PointsEqualsOneInch(t *testing.T) {
	SetScreenDPI(9600, 2540)
	defer SetScreenDPI(1920, 508)

	pts := Px("72p")
	inch := Px("1i")
	if pts != inch {
		t.Errorf("72p=%d should equal 1i=%d", pts, inch)
	}
}

func TestPx1cEquals10m(t *testing.T) {
	SetScreenDPI(9600, 2540)
	defer SetScreenDPI(1920, 508)

	cm := Px("1c")
	mm := Px("10m")
	if cm != mm {
		t.Errorf("1c=%d should equal 10m=%d", cm, mm)
	}
}

func TestPxPanicsOnInvalid(t *testing.T) {
	tests := []struct {
		name  string
		input any
	}{
		{"empty string", ""},
		{"invalid string", "abc"},
		{"bool type", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("Px(%v) should have panicked", tt.input)
				}
			}()
			Px(tt.input)
		})
	}
}

func TestSetScreenDPIIgnoresInvalid(t *testing.T) {
	old := screenWidthPx
	SetScreenDPI(0, 500)
	if screenWidthPx != old {
		t.Error("SetScreenDPI should ignore zero widthPx")
	}
	SetScreenDPI(1920, 0)
	if screenWidthPx != old {
		t.Error("SetScreenDPI should ignore zero widthMM")
	}
}
