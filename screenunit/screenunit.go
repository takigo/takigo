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
	"math"
	"strconv"
	"strings"
)

// screenWidthPx and screenWidthMM store the screen dimensions
// used for unit conversion. Default assumes 96 DPI.
var (
	screenWidthPx = 1920
	screenWidthMM = 508 // ~96 DPI: 1920 / (25.4 * 96) * 1000 ≈ 508mm
)

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
	if widthPx > 0 && widthMM > 0 {
		screenWidthPx = widthPx
		screenWidthMM = widthMM
	}
	if xftDPI > 0 && screenWidthPx > 0 {
		screenWidthMM = int(math.Round(float64(screenWidthPx) * 25.4 / xftDPI))
		if screenWidthMM <= 0 {
			screenWidthMM = 1
		}
	}
}

// DPI returns the current screen DPI (dots per inch).
// Standard desktop DPI is 96; HiDPI displays may be 144, 192, etc.
func DPI() float64 {
	return float64(screenWidthPx) * 25.4 / float64(screenWidthMM)
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
func Px(v any) int {
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(math.Round(val))
	case string:
		return parseDistance(val)
	default:
		panic(fmt.Sprintf("screenunit.Px: unsupported type %T", v))
	}
}

// parseDistance parses a Tk-style distance string into pixels.
func parseDistance(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		panic("screenunit.Px: empty string")
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
			panic(fmt.Sprintf("screenunit.Px: invalid distance %q: %v", s, err))
		}
		return int(math.Round(val))
	}

	numStr = strings.TrimSpace(numStr)
	if numStr == "" {
		panic(fmt.Sprintf("screenunit.Px: missing number in %q", s))
	}

	val, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		panic(fmt.Sprintf("screenunit.Px: invalid number in %q: %v", s, err))
	}

	// Convert: value_in_mm * pixels_per_mm
	// pixels_per_mm = screenWidthPx / screenWidthMM
	pixels := val * multiplier * float64(screenWidthPx) / float64(screenWidthMM)
	return int(math.Round(pixels))
}
