package font

import (
	"sort"
	"strings"
)

// Measurer is the part of Font needed for text layout.
type Measurer interface {
	MeasureString(s string) int
}

// WrapLines breaks text into display lines the way Tk_ComputeTextLayout
// (tk/generic/tkFont.c) does: newlines always break; with wrapLength > 0 a
// line is broken after the last whole word that still fits (a word ending
// exactly at wrapLength fits), whitespace inside a line is kept verbatim, and
// whitespace at a break is consumed. A single word longer than wrapLength is
// split at the last character that fits (at least one character per line).
func WrapLines(f Measurer, text string, wrapLength int) []string {
	paragraphs := strings.Split(text, "\n")
	if wrapLength <= 0 {
		return paragraphs
	}
	var lines []string
	for _, para := range paragraphs {
		if para == "" {
			lines = append(lines, "")
			continue
		}
		rest := para
		for rest != "" {
			line := fitLine(f, rest, wrapLength)
			lines = append(lines, line)
			rest = strings.TrimLeft(rest[len(line):], " \t")
		}
	}
	return lines
}

// fitLine returns the longest prefix of s that fits in wrapLength, preferring
// to end at a word boundary. Widths grow with the prefix, so it gallops and
// then bisects over the word ends it has scanned, measuring O(log n)
// prefixes no longer than about one line instead of every word end and the
// whole remainder.
func fitLine(f Measurer, s string, wrapLength int) string {
	fits := func(end int) bool { return TextWidth(f, s[:end]) <= wrapLength }

	// ends holds the word ends found so far; len(s) counts as one.
	var ends []int
	scanned := 0
	wordEnd := func(k int) (int, bool) {
		for len(ends) <= k && scanned < len(s) {
			scanned++
			if scanned == len(s) || (isBlank(s[scanned]) && !isBlank(s[scanned-1])) {
				ends = append(ends, scanned)
			}
		}
		if k < len(ends) {
			return ends[k], true
		}
		return 0, false
	}

	lo, hi := -1, 0 // ends[lo] fits (or lo < 0); ends[hi] does not, once found
	for {
		e, ok := wordEnd(hi)
		if !ok {
			break
		}
		if !fits(e) {
			break
		}
		if e == len(s) {
			return s
		}
		lo, hi = hi, 2*hi+1
	}
	hi = min(hi, len(ends))
	for hi-lo > 1 {
		mid := (lo + hi) / 2
		if fits(ends[mid]) {
			lo = mid
		} else {
			hi = mid
		}
	}
	if lo >= 0 {
		return s[:ends[lo]]
	}

	// A single word wider than the line: split it at the last character
	// that fits, keeping at least one.
	var cuts []int // cuts[k] ends the prefix of k+1 runes
	for i := range s {
		if i > 0 {
			cuts = append(cuts, i)
		}
	}
	cuts = append(cuts, len(s))
	n := sort.Search(len(cuts), func(k int) bool { return f.MeasureString(s[:cuts[k]]) > wrapLength })
	return s[:cuts[max(n-1, 0)]]
}

func isBlank(c byte) bool { return c == ' ' || c == '\t' }

// Segment is a run of text without tabs and its x offset within its line.
type Segment struct {
	Text string
	X    int
}

// TabWidth ports fontPtr->tabWidth (tkFont.c): 8 widths of "0".
func TabWidth(f Measurer) int {
	return max(1, 8*f.MeasureString("0"))
}

// Segments splits a line at tabs; like Tk_ComputeTextLayout, a tab moves
// to the next multiple of TabWidth from the line start.
func Segments(f Measurer, line string) []Segment {
	parts := strings.Split(line, "\t")
	segs := make([]Segment, 0, len(parts))
	x, tw := 0, 0
	for i, p := range parts {
		if i > 0 {
			if tw == 0 {
				tw = TabWidth(f)
			}
			x += tw
			x -= x % tw
		}
		segs = append(segs, Segment{p, x})
		x += f.MeasureString(p)
	}
	return segs
}

// TextWidth measures a line with tabs expanded.
func TextWidth(f Measurer, line string) int {
	if !strings.Contains(line, "\t") {
		return f.MeasureString(line)
	}
	segs := Segments(f, line)
	last := segs[len(segs)-1]
	return last.X + f.MeasureString(last.Text)
}
