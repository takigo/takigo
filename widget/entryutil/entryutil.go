// Package entryutil provides shared text-editing utility functions used by
// both the entry and spinbox widgets.
package entryutil

import "github.com/msorc/takigo/font"

// MeasureRunes returns the pixel width of a rune slice using the given font.
func MeasureRunes(f font.Font, runes []rune) int {
	if len(runes) == 0 {
		return 0
	}
	return f.MeasureString(string(runes))
}

// RuneIndexAtPixel finds the rune index at a given pixel offset using binary search.
func RuneIndexAtPixel(f font.Font, runes []rune, targetX int) int {
	if targetX <= 0 || len(runes) == 0 {
		return 0
	}
	lo, hi := 0, len(runes)
	for lo < hi {
		mid := (lo + hi) / 2
		w := MeasureRunes(f, runes[:mid+1])
		if w <= targetX {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo
}

// ClampIdx clamps an index to [0, max].
func ClampIdx(idx, max int) int {
	if idx < 0 {
		return 0
	}
	if idx > max {
		return max
	}
	return idx
}

// WordStart returns the rune index of the start of the word at or before pos.
func WordStart(text []rune, pos int) int {
	if pos <= 0 {
		return 0
	}
	if pos > len(text) {
		pos = len(text)
	}
	// Skip back past non-word chars.
	i := pos - 1
	for i > 0 && !IsWordChar(text[i]) {
		i--
	}
	// Skip back past word chars.
	for i > 0 && IsWordChar(text[i-1]) {
		i--
	}
	return i
}

// WordEnd returns the rune index past the end of the word at or after pos.
func WordEnd(text []rune, pos int) int {
	if pos >= len(text) {
		return len(text)
	}
	i := pos
	// Skip past word chars.
	for i < len(text) && IsWordChar(text[i]) {
		i++
	}
	// Skip past non-word chars.
	for i < len(text) && !IsWordChar(text[i]) {
		i++
	}
	return i
}

// IsWordChar returns true for alphanumeric and underscore characters.
func IsWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') || r == '_'
}
