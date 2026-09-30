// Package screenunit converts Tk-style screen distances to pixels.
//
// Tk supports distance values with unit suffixes:
//   - bare number or no suffix → pixels (passthrough)
//   - "p" suffix → points (1/72 inch)
//   - "m" suffix → millimeters
//   - "c" suffix → centimeters
//   - "i" suffix → inches
//
// Examples: 10, "10", "3p", "2.5m", "1c", "0.5i"
package screenunit

import (
	"fmt"
	"log"
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

// Px converts a Tk-style screen distance to pixels.
// Accepts: int (passthrough), float64 (rounded), string ("3p", "2.5m", "1c", "4i", "10").
// Panics on invalid input for fail-fast behavior during development.
// Use TryPx for untrusted input that may be invalid.
func Px(v any) int {
	px, err := TryPx(v)
	if err != nil {
		panic(err)
	}
	return px
}

// PxOr converts v like Px, but on invalid input logs a warning and returns
// prev, so an option setter keeps its previous value instead of panicking.
func PxOr(v any, prev int) int {
	px, err := TryPx(v)
	if err != nil {
		log.Printf("%v; keeping %d", err, prev)
		return prev
	}
	return px
}

// TryPx converts a Tk-style screen distance to pixels, returning an error
// on invalid input instead of panicking. Use this for user-provided or
// untrusted input (config files, command-line arguments, etc.).
func TryPx(v any) (int, error) {
	switch val := v.(type) {
	case int:
		return val, nil
	case float64:
		return int(math.Round(val)), nil
	case string:
		return tryParseDistance(val)
	default:
		return 0, fmt.Errorf("screenunit: unsupported type %T", v)
	}
}

// Float converts a Tk-style screen distance to unrounded pixels, as
// Tk_GetDoublePixelsFromObj does for canvas coordinates. Panics on invalid
// input like Px.
func Float(v any) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case float64:
		return val
	case string:
		f, err := parseDistance(val)
		if err != nil {
			panic(err)
		}
		return f
	default:
		panic(fmt.Errorf("screenunit: unsupported type %T", v))
	}
}

// tryParseDistance parses a Tk-style distance string into rounded pixels.
func tryParseDistance(s string) (int, error) {
	f, err := parseDistance(s)
	if err != nil {
		return 0, err
	}
	return int(math.Round(f)), nil
}

// parseDistance parses a Tk-style distance string into unrounded pixels.
func parseDistance(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("screenunit: empty string")
	}

	// Check for unit suffix.
	last := s[len(s)-1]
	var numStr string
	var multiplier float64

	switch last {
	case 'p': // points (1/72 inch)
		numStr = s[:len(s)-1]
		multiplier = 25.4 / 72.0
	case 'm': // millimeters
		numStr = s[:len(s)-1]
		multiplier = 1.0
	case 'c': // centimeters
		numStr = s[:len(s)-1]
		multiplier = 10.0
	case 'i': // inches
		numStr = s[:len(s)-1]
		multiplier = 25.4
	default:
		// Bare number — pixels.
		val, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, fmt.Errorf("screenunit: invalid distance %q: %v", s, err)
		}
		return val, nil
	}

	numStr = strings.TrimSpace(numStr)
	if numStr == "" {
		return 0, fmt.Errorf("screenunit: missing number in %q", s)
	}

	val, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("screenunit: invalid number in %q: %v", s, err)
	}

	// Convert: value_in_mm * pixels_per_mm
	// pixels_per_mm = widthPx / widthMM
	m := screen.Load()
	return val * multiplier * float64(m.widthPx) / float64(m.widthMM), nil
}
