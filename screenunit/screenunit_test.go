package screenunit

import (
	"errors"
	"math"
	"sync"
	"testing"
)

func TestToPixelsNumbers(t *testing.T) {
	if got := ToPixels(42); got != 42 {
		t.Errorf("ToPixels(42) = %d, want 42", got)
	}
	if got := ToPixels(3.7); got != 4 {
		t.Errorf("ToPixels(3.7) = %d, want 4", got)
	}
	if got := ToPixels(Px(3.2)); got != 3 {
		t.Errorf("ToPixels(Px(3.2)) = %d, want 3", got)
	}
	if got := ToFloat(2.5); got != 2.5 {
		t.Errorf("ToFloat(2.5) = %v, want 2.5", got)
	}
}

func TestUnits(t *testing.T) {
	// 9600px across 2540mm is exactly 96 DPI.
	SetScreenDPI(9600, 2540, 0)
	defer SetScreenDPI(1920, 508, 0)
	pxPerMM := 9600.0 / 2540.0

	tests := []struct {
		input string
		d     Distance
		want  float64
	}{
		{"1p", Pt(1), 25.4 / 72.0 * pxPerMM},
		{"72p", Pt(72), 25.4 * pxPerMM},
		{"1m", Mm(1), pxPerMM},
		{"10m", Mm(10), 10 * pxPerMM},
		{"1c", Cm(1), 10 * pxPerMM},
		{"2.5c", Cm(2.5), 25 * pxPerMM},
		{"1i", In(1), 25.4 * pxPerMM},
		{"0.5i", In(0.5), 12.7 * pxPerMM},
		{"10", Px(10), 10},
		{"-3.5", Px(-3.5), -3.5},
		{" 2 m ", Mm(2), 2 * pxPerMM},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parsed, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse(%q): %v", tt.input, err)
			}
			if parsed != tt.d {
				t.Errorf("Parse(%q) = %v, want %v", tt.input, parsed, tt.d)
			}
			if got := tt.d.Float(); math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("%v.Float() = %v, want %v", tt.d, got, tt.want)
			}
			if got, want := tt.d.Pixels(), int(math.Round(tt.want)); got != want {
				t.Errorf("%v.Pixels() = %d, want %d", tt.d, got, want)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		d    Distance
		want string
	}{
		{Pt(1.5), "1.5p"}, {Mm(3), "3m"}, {Cm(2), "2c"}, {In(0.5), "0.5i"}, {Px(10), "10"}, {Distance{}, "0"},
	}
	for _, tt := range tests {
		if got := tt.d.String(); got != tt.want {
			t.Errorf("String() = %q, want %q", got, tt.want)
		}
	}
}

func TestParseErrors(t *testing.T) {
	for _, in := range []string{"", "   ", "p", "abc", "3x", "1.2.3m", "3pp"} {
		if _, err := Parse(in); !errors.Is(err, ErrBadDistance) {
			t.Errorf("Parse(%q) error = %v, want ErrBadDistance", in, err)
		}
	}
}

func TestMustParsePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("MustParse of a bad distance did not panic")
		}
	}()
	MustParse("bogus")
}

func TestXftDPI(t *testing.T) {
	defer SetScreenDPI(1920, 508, 0)
	SetScreenDPI(1920, 508, 144)
	if got := In(1).Pixels(); got != 144 {
		t.Errorf("1i at Xft.dpi 144 = %d, want 144", got)
	}
	if got := ScalingPct(); got != 150 {
		t.Errorf("ScalingPct = %d, want 150", got)
	}
}

// Every NewApp sets the metrics while other Apps convert distances.
func TestConcurrentSetAndConvert(t *testing.T) {
	defer SetScreenDPI(1920, 508, 0)
	var wg sync.WaitGroup
	for i := range 4 {
		wg.Go(func() {
			for range 500 {
				SetScreenDPI(1920+i, 508, 0)
				_ = Mm(10).Pixels()
				_ = DPI()
			}
		})
	}
	wg.Wait()
}
