// Package font provides font management with Xft/fontconfig bindings.
// It ports tk/generic/tkFont.c and tk/unix/tkUnixRFont.c.
package font

import (
	"fmt"
	"strconv"
	"strings"
)

// Weight constants for font weight.
type Weight int

const (
	WeightNormal Weight = iota
	WeightBold
)

// Slant constants for font slant.
type Slant int

const (
	SlantRoman Slant = iota
	SlantItalic
	SlantOblique
)

// Attributes describes a font's properties.
type Attributes struct {
	Family     string  // font family name (empty = system default)
	Size       float64 // point size (0 = default, negative = pixel size)
	Weight     Weight
	Slant      Slant
	Underline  bool
	Overstrike bool
}

// Descriptor returns a font descriptor string like "Helvetica Neue Bold 13".
func (a Attributes) Descriptor() string {
	family := a.Family
	if family == "" {
		family = "sans-serif"
	}
	w := ""
	if a.Weight == WeightBold {
		w = " Bold"
	}
	s := ""
	if a.Slant == SlantItalic {
		s = " Italic"
	} else if a.Slant == SlantOblique {
		s = " Oblique"
	}
	size := a.Size
	if size <= 0 {
		size = 10
	}
	return fmt.Sprintf("%s%s%s %.0f", family, w, s, size)
}

// Metrics holds font measurement data.
type Metrics struct {
	Ascent   int  // pixels from baseline to top
	Descent  int  // pixels from baseline to bottom
	MaxWidth int  // width of widest character
	Fixed    bool // true for monospace
}

// Linespace returns the total line height.
func (m Metrics) Linespace() int {
	return m.Ascent + m.Descent
}

// Font is the interface for a loaded, usable font.
type Font interface {
	// Attrs returns the font's actual attributes.
	Attrs() Attributes

	// Metrics returns the font metrics.
	Metrics() Metrics

	// MeasureString returns the pixel width of a string.
	MeasureString(s string) int

	// Close releases the font resources.
	Close()
}

// ParseDescriptor parses a font descriptor string.
// Supports three formats:
//   - XLFD: "-foundry-family-weight-slant-..."
//   - Option-value: "-family Times -size 12 -weight bold"
//   - Simple: "Times 12 bold italic"
func ParseDescriptor(desc string) (Attributes, error) {
	desc = strings.TrimSpace(desc)
	if desc == "" {
		return DefaultAttributes(), nil
	}

	// XLFD format starts with "-" followed by foundry.
	if strings.HasPrefix(desc, "-") && strings.Count(desc, "-") >= 13 {
		return parseXLFD(desc)
	}

	// Option-value format has "-family", "-size", etc.
	if strings.HasPrefix(desc, "-family") || strings.HasPrefix(desc, "-size") ||
		strings.HasPrefix(desc, "-weight") || strings.HasPrefix(desc, "-slant") {
		return parseOptionValue(desc)
	}

	// Simple format: "Family ?size? ?style...?"
	return parseSimple(desc)
}

// DefaultAttributes returns the default font attributes.
func DefaultAttributes() Attributes {
	return Attributes{
		Family: "sans-serif",
		Size:   12,
		Weight: WeightNormal,
		Slant:  SlantRoman,
	}
}

// parseXLFD parses an X Logical Font Description.
func parseXLFD(xlfd string) (Attributes, error) {
	fields := strings.Split(xlfd, "-")
	if len(fields) < 14 {
		return Attributes{}, fmt.Errorf("invalid XLFD: %q", xlfd)
	}

	// Fields: -foundry-family-weight-slant-setwidth-addstyle-pixelsize-pointsize-...
	attrs := DefaultAttributes()
	if fields[2] != "*" && fields[2] != "" {
		attrs.Family = fields[2]
	}

	// Weight (field 3).
	switch strings.ToLower(fields[3]) {
	case "bold", "demi", "demibold":
		attrs.Weight = WeightBold
	}

	// Slant (field 4).
	switch strings.ToLower(fields[4]) {
	case "i":
		attrs.Slant = SlantItalic
	case "o":
		attrs.Slant = SlantOblique
	}

	// Point size (field 8) in tenths of a point.
	if fields[8] != "*" && fields[8] != "" && fields[8] != "0" {
		if pts, err := strconv.ParseFloat(fields[8], 64); err == nil {
			attrs.Size = pts / 10.0
		}
	}

	// Pixel size (field 7) — overrides point size.
	if fields[7] != "*" && fields[7] != "" && fields[7] != "0" {
		if px, err := strconv.ParseFloat(fields[7], 64); err == nil {
			attrs.Size = -px // negative = pixel size
		}
	}

	return attrs, nil
}

// parseOptionValue parses "-family Times -size 12 -weight bold" format.
func parseOptionValue(desc string) (Attributes, error) {
	attrs := DefaultAttributes()
	parts := splitList(desc)

	for i := 0; i < len(parts)-1; i += 2 {
		key := parts[i]
		val := parts[i+1]

		switch key {
		case "-family":
			attrs.Family = val
		case "-size":
			if s, err := strconv.ParseFloat(val, 64); err == nil {
				attrs.Size = s
			}
		case "-weight":
			if val == "bold" {
				attrs.Weight = WeightBold
			}
		case "-slant":
			switch val {
			case "italic":
				attrs.Slant = SlantItalic
			case "oblique":
				attrs.Slant = SlantOblique
			}
		case "-underline":
			attrs.Underline = val == "1" || val == "true"
		case "-overstrike":
			attrs.Overstrike = val == "1" || val == "true"
		}
	}

	return attrs, nil
}

// parseSimple parses "Family ?size? ?style...?" format.
func parseSimple(desc string) (Attributes, error) {
	attrs := DefaultAttributes()
	parts := splitList(desc)
	if len(parts) == 0 {
		return attrs, nil
	}

	// First element: family, braced when it has spaces ({DejaVu Sans} 12).
	attrs.Family = parts[0]

	for _, p := range parts[1:] {
		switch strings.ToLower(p) {
		case "bold":
			attrs.Weight = WeightBold
		case "italic":
			attrs.Slant = SlantItalic
		case "oblique":
			attrs.Slant = SlantOblique
		case "underline":
			attrs.Underline = true
		case "overstrike":
			attrs.Overstrike = true
		default:
			// Try as size.
			if s, err := strconv.ParseFloat(p, 64); err == nil {
				attrs.Size = s
			}
		}
	}

	return attrs, nil
}

// Underline ports the Xft font's underlinePos/underlineHeight
// (tk/unix/tkUnixRFont.c InitFont): half the descent below the baseline,
// a third of the width of "I" thick, kept inside the descent.
func Underline(f Font) (pos, height int) {
	m := f.Metrics()
	pos = m.Descent / 2
	height = max(1, f.MeasureString("I")/3)
	if height+pos > m.Descent {
		height = m.Descent - pos
		if height == 0 {
			pos--
			height = 1
		}
	}
	return pos, height
}

// splitList splits a font description into Tcl list elements: words
// separated by whitespace, where {...} (nesting) or "..." groups words into
// one element, as Tk_GetFontFromObj reads "{DejaVu Sans} 12 bold".
func splitList(s string) []string {
	var out []string
	for i := 0; i < len(s); {
		for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n') {
			i++
		}
		if i == len(s) {
			break
		}
		switch s[i] {
		case '{':
			depth, j := 1, i+1
			for ; j < len(s) && depth > 0; j++ {
				switch s[j] {
				case '{':
					depth++
				case '}':
					depth--
				}
			}
			end := j
			if depth == 0 {
				end = j - 1
			}
			out = append(out, s[i+1:end])
			i = j
		case '"':
			j := strings.IndexByte(s[i+1:], '"')
			if j < 0 {
				out = append(out, s[i+1:])
				i = len(s)
			} else {
				out = append(out, s[i+1:i+1+j])
				i += j + 2
			}
		default:
			j := i
			for j < len(s) && s[j] != ' ' && s[j] != '\t' && s[j] != '\n' {
				j++
			}
			out = append(out, s[i:j])
			i = j
		}
	}
	return out
}
