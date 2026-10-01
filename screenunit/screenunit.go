// Package screenunit holds Tk's screen distances and converts them to
// pixels.
//
// A distance option takes a [Length]: a plain number of pixels, or a
// [Distance] built with [Pt], [Mm], [Cm] or [In]:
//
//	pack.PadX(4)                    // pixels
//	pack.PadX(screenunit.Pt(1.5))   // Tk's "1.5p"
//
// [Parse] reads Tk's string form ("3p", "2.5m", "1c", "0.5i", "10") for
// distances that come from a file or the user.
package screenunit

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync/atomic"
)

// metrics are the screen dimensions used for unit conversion.
type metrics struct {
	widthPx, widthMM int
}

// screen holds the current metrics. They are process-wide, set by each
// NewApp and read by every App's loop goroutine, so they are swapped
// atomically; the default assumes 96 DPI (1920 / (25.4 * 96) * 1000 ≈ 508mm).
var screen atomic.Pointer[metrics]

func init() {
	screen.Store(&metrics{widthPx: 1920, widthMM: 508})
}

// SetScreenDPI configures the screen dimensions used for unit conversion.
// widthPx and widthMM are the raw X11 screen dimensions.
// xftDPI is the Xft.dpi value from X resources (0 if not set).
//
// When xftDPI is set, widthMM is rewritten to match the configured DPI,
// mirroring Tk's ScalingCmd (tkCmds.c:1316) which does:
//
//	scalingFactor = xftDPI / 72
//	WidthMMOfScreen = (25.4/72) / scalingFactor * WidthOfScreen
//	                = WidthOfScreen * 25.4 / xftDPI
//
// This ensures that "4i" converts to exactly 4 * xftDPI pixels.
func SetScreenDPI(widthPx, widthMM int, xftDPI float64) {
	m := *screen.Load()
	if widthPx > 0 && widthMM > 0 {
		m.widthPx = widthPx
		m.widthMM = widthMM
	}
	if xftDPI > 0 && m.widthPx > 0 {
		m.widthMM = max(int(math.Round(float64(m.widthPx)*25.4/xftDPI)), 1)
	}
	screen.Store(&m)
}

// DPI returns the current screen DPI (dots per inch).
// Standard desktop DPI is 96; HiDPI displays may be 144, 192, etc.
func DPI() float64 {
	m := screen.Load()
	return float64(m.widthPx) * 25.4 / float64(m.widthMM)
}

// ScalingFactor returns the ratio of actual DPI to the standard 96 DPI baseline.
// Returns 1.0 at 96 DPI, 1.5 at 144 DPI, 2.0 at 192 DPI, etc.
func ScalingFactor() float64 {
	return DPI() / 96.0
}

// ScalingPct returns the scaling percentage rounded to the nearest multiple
// of 25 that is at least 100, matching Tk's ::tk::scalingPct (scaling.tcl).
// Returns 100 at 96 DPI, 150 at 144 DPI, 200 at 192 DPI, etc.
func ScalingPct() int {
	pct := DPI() / 96.0 * 100.0
	scalingPct := 100
	for pct >= float64(scalingPct)+12.5 {
		scalingPct += 25
	}
	return scalingPct
}

// Distance is a screen distance in pixels, points, millimetres,
// centimetres or inches. The zero value is zero pixels.
type Distance struct {
	n float64
	// mm is the size of one unit in millimetres; 0 means n is in pixels.
	mm   float64
	unit byte
}

// Length is what a distance option accepts: a number of pixels or a Distance.
type Length interface {
	int | float64 | Distance
}

// ErrBadDistance is wrapped by the errors Parse returns.
var ErrBadDistance = errors.New("screenunit: bad distance")

// Px returns a distance of n pixels.
func Px(n float64) Distance { return Distance{n: n} }

// Pt returns a distance of n points (1/72 inch), Tk's "p" suffix.
func Pt(n float64) Distance { return Distance{n: n, mm: 25.4 / 72.0, unit: 'p'} }

// Mm returns a distance of n millimetres, Tk's "m" suffix.
func Mm(n float64) Distance { return Distance{n: n, mm: 1, unit: 'm'} }

// Cm returns a distance of n centimetres, Tk's "c" suffix.
func Cm(n float64) Distance { return Distance{n: n, mm: 10, unit: 'c'} }

// In returns a distance of n inches, Tk's "i" suffix.
func In(n float64) Distance { return Distance{n: n, mm: 25.4, unit: 'i'} }

// Float returns the distance in unrounded pixels on the current screen, as
// Tk_GetDoublePixelsFromObj does for canvas coordinates.
func (d Distance) Float() float64 {
	if d.mm == 0 {
		return d.n
	}
	m := screen.Load()
	return d.n * d.mm * float64(m.widthPx) / float64(m.widthMM)
}

// Pixels returns the distance rounded to whole pixels on the current screen.
func (d Distance) Pixels() int { return int(math.Round(d.Float())) }

// String returns Tk's form of the distance, e.g. "1.5p" or "10".
func (d Distance) String() string {
	s := strconv.FormatFloat(d.n, 'g', -1, 64)
	if d.mm == 0 {
		return s
	}
	return s + string(d.unit)
}

// ToFloat converts a Length to unrounded pixels.
func ToFloat[L Length](l L) float64 {
	switch v := any(l).(type) {
	case Distance:
		return v.Float()
	case int:
		return float64(v)
	case float64:
		return v
	}
	return 0
}

// ToPixels converts a Length to whole pixels.
func ToPixels[L Length](l L) int {
	if n, ok := any(l).(int); ok {
		return n
	}
	return int(math.Round(ToFloat(l)))
}

// MustParse is Parse for constant strings; it panics on a bad distance.
func MustParse(s string) Distance {
	d, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return d
}

// Parse reads a Tk distance: a number, optionally followed by p (points),
// m (millimetres), c (centimetres) or i (inches). A bare number is pixels.
// Its errors wrap ErrBadDistance.
func Parse(s string) (Distance, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Distance{}, fmt.Errorf("%w: empty string", ErrBadDistance)
	}
	unit := Px
	num := s
	switch s[len(s)-1] {
	case 'p':
		unit, num = Pt, s[:len(s)-1]
	case 'm':
		unit, num = Mm, s[:len(s)-1]
	case 'c':
		unit, num = Cm, s[:len(s)-1]
	case 'i':
		unit, num = In, s[:len(s)-1]
	}
	n, err := strconv.ParseFloat(strings.TrimSpace(num), 64)
	if err != nil {
		return Distance{}, fmt.Errorf("%w: %q", ErrBadDistance, s)
	}
	return unit(n), nil
}
