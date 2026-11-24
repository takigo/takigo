package text

// Tab stops: -tabs of the widget and of tags, ports of TkTextGetTabs,
// AdjustForTab and the tab parts of SizeOfTab (tkTextDisp.c).

import (
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/takigo/takigo/screenunit"
)

// TabAlign is how text after a tab lines up with its stop.
type TabAlign int

const (
	TabLeft    TabAlign = iota // the text starts at the stop
	TabRight                   // the text ends at the stop
	TabCenter                  // the text is centred on the stop
	TabNumeric                 // the decimal point is at the stop
)

// TabStyle is -tabstyle: how a tab finds its stop.
type TabStyle int

const (
	// TabTabular gives the n'th tab of a line the n'th stop, so tabs line
	// columns up even when a column is too wide (Tk's default).
	TabTabular TabStyle = iota
	// TabWordProcessor moves a tab to the first stop past the text before
	// it, as word processors do.
	TabWordProcessor
)

// TabStyleOpt sets the widget's -tabstyle.
func TabStyleOpt(style TabStyle) TextOption {
	return func(t *TextWidget) {
		t.tabStyle = style
		t.tabsGen++
	}
}

// TabStop is one tab stop: a distance from the left edge of the text area
// and an alignment.
type TabStop struct {
	Pos   int
	Align TabAlign
}

// Tab returns a tab stop at a distance with an alignment.
func Tab[L screenunit.Length](pos L, align TabAlign) TabStop {
	return TabStop{Pos: screenunit.ToPixels(pos), Align: align}
}

// ParseTabs reads Tk's -tabs list, such as "2c left 4c right 6c numeric":
// distances, each optionally followed by its alignment (left by default).
func ParseTabs(s string) ([]TabStop, error) {
	var stops []TabStop
	for f := range strings.FieldsSeq(s) {
		switch strings.ToLower(f) {
		case "left", "right", "center", "numeric":
			if len(stops) == 0 {
				return nil, fmt.Errorf("tab alignment %q must follow a distance", f)
			}
			stops[len(stops)-1].Align = map[string]TabAlign{
				"left": TabLeft, "right": TabRight, "center": TabCenter, "numeric": TabNumeric,
			}[strings.ToLower(f)]
			continue
		}
		d, err := screenunit.Parse(f)
		if err != nil {
			return nil, err
		}
		px := d.Pixels()
		if px <= 0 || (len(stops) > 0 && px <= stops[len(stops)-1].Pos) {
			return nil, fmt.Errorf("tabs must be monotonically increasing: %q", f)
		}
		stops = append(stops, TabStop{Pos: px})
	}
	return stops, nil
}

// Tabs sets the widget's tab stops (-tabs); no stops restores tabs every
// -tabwidth average characters.
func Tabs(stops ...TabStop) TextOption {
	return func(t *TextWidget) {
		t.tabs = slices.Clone(stops)
		t.tabsGen++
	}
}

// stopAt returns the index'th stop (from 0), extrapolating past the last
// from the last two as TkTextGetTabs's tabIncrement does.
func stopAt(stops []TabStop, index int) TabStop {
	if index < len(stops) {
		return stops[index]
	}
	last := stops[len(stops)-1]
	inc := last.Pos
	if n := len(stops); n > 1 {
		inc = last.Pos - stops[n-2].Pos
	}
	return TabStop{Pos: last.Pos + (index+1-len(stops))*inc, Align: last.Align}
}

// tabStyleFor returns the -tabstyle of a logical line: its first
// character's tag's, else the widget's.
func (t *TextWidget) tabStyleFor(lineIdx int) TabStyle {
	for _, tg := range t.doc.tagsByPriority {
		if tg.TabStyleSet {
			return t.resolveLineProps(lineIdx).tabStyle
		}
	}
	return t.tabStyle
}

// tabStopsFor returns the custom tab stops of a logical line (the first
// character's tag, else the widget's) and the left margin of the display
// line starting at startChar, which the stops are measured past: Tk's chunk
// x starts at the margin. nil stops mean the default tabs.
func (t *TextWidget) tabStopsFor(lineIdx, startChar int) ([]TabStop, int) {
	stops := t.tabs
	hasTagTabs := false
	for _, tg := range t.doc.tagsByPriority {
		if len(tg.Tabs) > 0 {
			hasTagTabs = true
			break
		}
	}
	if !hasTagTabs && len(stops) == 0 {
		return nil, 0
	}
	p := t.resolveLineProps(lineIdx)
	if len(p.tabs) > 0 {
		stops = p.tabs
	}
	if len(stops) == 0 {
		return nil, 0
	}
	if startChar == 0 {
		return stops, p.lm1
	}
	return stops, p.lm2
}

// alignTabs ports AdjustForTab for custom stops: each tab segment widens so
// the group of segments after it (up to the next tab) lines up with the
// tab's stop, by at least a space. Segment x positions are set from 0.
func (t *TextWidget) alignTabs(segs []textSegment, stops []TabStop, origin int, style TabStyle) {
	space := t.Font.MeasureString(" ")
	x, tabIdx := 0, -1
	for i := range segs {
		segs[i].x = x
		if !segs[i].isTab {
			x += segs[i].width
			continue
		}
		tabIdx++
		stop := stopAt(stops, tabIdx)
		// SizeOfTab: a word-processor tab skips the stops it is already past.
		for style == TabWordProcessor && stop.Pos <= origin+x {
			tabIdx++
			stop = stopAt(stops, tabIdx)
		}
		end := i + 1
		for end < len(segs) && !segs[end].isTab {
			end++
		}
		group := segs[i+1 : end]
		width := 0
		for _, s := range group {
			width += s.width
		}
		desired := stop.Pos
		switch stop.Align {
		case TabRight:
			desired = stop.Pos - width
		case TabCenter:
			desired = stop.Pos - width/2
		case TabNumeric:
			if dx, ok := decimalOffset(group, t); ok {
				desired = stop.Pos - dx
			} else {
				desired = stop.Pos - width
			}
		default:
		}
		segs[i].width = max(desired-(origin+x), space)
		x += segs[i].width
	}
}

// decimalOffset finds, as AdjustForTab does for numeric tabs, the last '.'
// or ',' before the first character after a digit that is not part of a
// number, and returns the width of the group's text before it.
func decimalOffset(group []textSegment, t *TextWidget) (int, bool) {
	gotDigit, found := false, false
	before, dx := 0, 0
	for _, s := range group {
		f := s.font
		if f == nil {
			f = t.Font
		}
		runes := []rune(s.text)
		for k, r := range runes {
			switch {
			case unicode.IsDigit(r):
				gotDigit = true
			case r == '.' || r == ',':
				dx, found = before+f.MeasureString(string(runes[:k])), true
			case gotDigit:
				if !found {
					dx = before + f.MeasureString(string(runes[:k]))
				}
				return dx, true
			}
		}
		before += s.width
	}
	return dx, found
}
