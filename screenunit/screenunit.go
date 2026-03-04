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

// SetScreenDPI configures the DPI used for unit conversion.
// Called once during App initialization from X11 screen metrics.
func SetScreenDPI(widthPx, widthMM int) {
	if widthPx > 0 && widthMM > 0 {
		screenWidthPx = widthPx
		screenWidthMM = widthMM
	}
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
