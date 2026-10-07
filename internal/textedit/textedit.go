// Package textedit holds the text measuring and word-boundary helpers of
// tk/generic/tkEntry.c that the classic and themed entry-like widgets
// share.
package textedit

import "github.com/takigo/takigo/font"

// MeasureRunes returns the pixel width of runes in f.
func MeasureRunes(f font.Font, runes []rune) int {
	if len(runes) == 0 {
		return 0
	}
	return f.MeasureString(string(runes))
}

// RuneIndexAtPixel returns the index of the rune under pixel x from the
// start of runes: 0 at or before the start, len(runes) past the end.
func RuneIndexAtPixel(f font.Font, runes []rune, x int) int {
	if x <= 0 || len(runes) == 0 {
		return 0
	}
	lo, hi := 0, len(runes)
	for lo < hi {
		mid := (lo + hi) / 2
		if MeasureRunes(f, runes[:mid+1]) <= x {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo
}

// ClampIdx clamps idx to [0, maxVal].
func ClampIdx(idx, maxVal int) int {
	return max(0, min(idx, maxVal))
}

// WordStart returns the index of the start of the word at or before pos.
func WordStart(text []rune, pos int) int {
	if pos <= 0 {
		return 0
	}
	pos = min(pos, len(text))
	i := pos - 1
	for i > 0 && !IsWordChar(text[i]) {
		i--
	}
	for i > 0 && IsWordChar(text[i-1]) {
		i--
	}
	return i
}

// WordEnd returns the index past the end of the word at or after pos.
func WordEnd(text []rune, pos int) int {
	if pos >= len(text) {
		return len(text)
	}
	i := max(pos, 0)
	for i < len(text) && IsWordChar(text[i]) {
		i++
	}
	for i < len(text) && !IsWordChar(text[i]) {
		i++
	}
	return i
}

// IsWordChar reports whether r is a letter, a digit or an underscore.
func IsWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') || r == '_'
}
