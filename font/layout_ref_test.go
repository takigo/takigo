package font

import (
	"math/rand/v2"
	"reflect"
	"strings"
	"testing"
)

// refFitLine is the straightforward fitLine the galloping one replaced.
func refFitLine(f Measurer, s string, wrapLength int) string {
	if TextWidth(f, s) <= wrapLength {
		return s
	}
	best := -1
	for i := 1; i < len(s); i++ {
		if (s[i] == ' ' || s[i] == '\t') && s[i-1] != ' ' && s[i-1] != '\t' {
			if TextWidth(f, s[:i]) > wrapLength {
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

func refWrapLines(f Measurer, text string, wrapLength int) []string {
	var lines []string
	for _, para := range strings.Split(text, "\n") {
		if para == "" {
			lines = append(lines, "")
			continue
		}
		for rest := para; rest != ""; {
			line := refFitLine(f, rest, wrapLength)
			lines = append(lines, line)
			rest = strings.TrimLeft(rest[len(line):], " \t")
		}
	}
	return lines
}

func TestWrapLinesMatchesReference(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	alphabet := []string{"a", "b", "é", "語", " ", " ", "  ", "\t", "\n"}
	for range 20000 {
		var sb strings.Builder
		for range rng.IntN(60) {
			sb.WriteString(alphabet[rng.IntN(len(alphabet))])
		}
		text, wrap := sb.String(), 1+rng.IntN(90)
		got := WrapLines(fixedMeasurer{}, text, wrap)
		want := refWrapLines(fixedMeasurer{}, text, wrap)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("WrapLines(%q, %d) = %q, want %q", text, wrap, got, want)
		}
	}
}

// countingMeasurer counts the bytes it is asked to measure.
type countingMeasurer struct{ bytes int }

func (m *countingMeasurer) MeasureString(s string) int {
	m.bytes += len(s)
	return fixedMeasurer{}.MeasureString(s)
}

func TestWrapLinesCostIsNearLinear(t *testing.T) {
	text := strings.Repeat("lorem ipsum dolor sit amet ", 800) // ~21.6k bytes
	m := &countingMeasurer{}
	WrapLines(m, text, 400)
	if limit := 40 * len(text); m.bytes > limit {
		t.Errorf("measured %d bytes to wrap %d, want at most %d", m.bytes, len(text), limit)
	}
}

func BenchmarkWrapLinesParagraph(b *testing.B) {
	text := strings.Repeat("lorem ipsum dolor sit amet ", 800)
	for b.Loop() {
		WrapLines(fixedMeasurer{}, text, 400)
	}
}
