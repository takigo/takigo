package font

import "strings"

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
// to end at a word boundary.
func fitLine(f Measurer, s string, wrapLength int) string {
	if f.MeasureString(s) <= wrapLength {
		return s
	}
	best := -1
	for i := 1; i < len(s); i++ {
		if (s[i] == ' ' || s[i] == '\t') && s[i-1] != ' ' && s[i-1] != '\t' {
			if f.MeasureString(s[:i]) > wrapLength {
				break
			}
			best = i
		}
	}
	if best > 0 {
		return s[:best]
	}
	runes := []rune(s)
	n := 1
	for n < len(runes) && f.MeasureString(string(runes[:n+1])) <= wrapLength {
		n++
	}
	return string(runes[:n])
}
