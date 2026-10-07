package textedit

import (
	"testing"

	"github.com/takigo/takigo/font"
)

// fixedFont is 10px per rune.
type fixedFont struct{}

func (fixedFont) Attrs() font.Attributes     { return font.Attributes{} }
func (fixedFont) Metrics() font.Metrics      { return font.Metrics{} }
func (fixedFont) MeasureString(s string) int { return 10 * len([]rune(s)) }
func (fixedFont) Close()                     {}

func TestRuneIndexAtPixel(t *testing.T) {
	t.Parallel()
	text := []rune("héllo")
	for _, tt := range []struct{ x, want int }{
		{-5, 0}, {0, 0}, {9, 0}, {10, 1}, {25, 2}, {49, 4}, {50, 5}, {500, 5},
	} {
		if got := RuneIndexAtPixel(fixedFont{}, text, tt.x); got != tt.want {
			t.Errorf("RuneIndexAtPixel(%d) = %d, want %d", tt.x, got, tt.want)
		}
	}
	if got := RuneIndexAtPixel(fixedFont{}, nil, 30); got != 0 {
		t.Errorf("empty text: %d, want 0", got)
	}
}

func TestWordBoundaries(t *testing.T) {
	t.Parallel()
	text := []rune("foo_1 bar,  baz")
	// WordStart at a word's first character goes to the previous word, as
	// tk::EntryPreviousWord does; WordEnd skips the word and the gap after it.
	for _, tt := range []struct{ pos, start, end int }{
		{0, 0, 6}, {3, 0, 6}, {5, 0, 6}, {6, 0, 12}, {9, 6, 12},
		{12, 6, 15}, {15, 12, 15}, {99, 12, 15}, {-1, 0, 6},
	} {
		if got := WordStart(text, tt.pos); got != tt.start {
			t.Errorf("WordStart(%d) = %d, want %d", tt.pos, got, tt.start)
		}
		if got := WordEnd(text, tt.pos); got != tt.end {
			t.Errorf("WordEnd(%d) = %d, want %d", tt.pos, got, tt.end)
		}
	}
}

func TestClampIdx(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ idx, max, want int }{{-3, 5, 0}, {2, 5, 2}, {9, 5, 5}, {0, 0, 0}} {
		if got := ClampIdx(tt.idx, tt.max); got != tt.want {
			t.Errorf("ClampIdx(%d, %d) = %d, want %d", tt.idx, tt.max, got, tt.want)
		}
	}
}
