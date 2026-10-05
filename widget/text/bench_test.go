package text

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/window"
)

// fixedFont is a display-free font.Font: 7px per rune, ascent 10, descent 3.
type fixedFont struct{}

func (fixedFont) Attrs() font.Attributes     { return font.Attributes{} }
func (fixedFont) Metrics() font.Metrics      { return font.Metrics{Ascent: 10, Descent: 3, MaxWidth: 7} }
func (fixedFont) MeasureString(s string) int { return 7 * utf8.RuneCountInString(s) }
func (fixedFont) Close()                     {}

func newBenchWidget(doc *Document, w, h int) *TextWidget {
	t := &TextWidget{
		doc:      doc,
		topLine:  1,
		wrapMode: WrapWord,
	}
	t.Win = &window.Window{Width: w, Height: h}
	t.Font = fixedFont{}
	t.doc.putTag(&Tag{Name: "sel", Priority: selPriority})
	t.layout.init(t)
	t.doc.Listeners = append(t.doc.Listeners, t.layout.apply)
	t.doc.subscribeObjects(t.objectDeleted)
	return t
}

func benchText(lines, width int) string {
	var b strings.Builder
	word := "lorem "
	for i := range lines {
		fmt.Fprintf(&b, "%05d ", i)
		for b.Len()%(width+1) < width-len(word) {
			b.WriteString(word)
		}
		b.WriteByte('\n')
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func benchDoc(lines, width int) *Document {
	return docWithText(benchText(lines, width))
}

func benchTaggedDoc(lines, ranges int) *Document {
	doc := benchDoc(lines, 60)
	for i := range ranges {
		l := i%lines + 1
		doc.TagAdd(fmt.Sprintf("t%d", i%8), Index{l, 6}, Index{l, 11})
	}
	return doc
}

func BenchmarkInsertCharEnd(b *testing.B) {
	doc := benchDoc(10000, 60)
	b.ResetTimer()
	for b.Loop() {
		doc.Insert(doc.EndIndex(), "x")
	}
}

func BenchmarkInsertCharTop(b *testing.B) {
	doc := benchDoc(10000, 60)
	b.ResetTimer()
	for b.Loop() {
		doc.Insert(Index{1, 0}, "x")
		doc.Delete(Index{1, 0}, Index{1, 1})
	}
}

func BenchmarkInsertNewlineMiddle(b *testing.B) {
	doc := benchDoc(10000, 60)
	b.ResetTimer()
	for b.Loop() {
		doc.Insert(Index{5000, 0}, "\n")
		doc.Delete(Index{5000, 0}, Index{5001, 0})
	}
}

func BenchmarkDeleteChar(b *testing.B) {
	doc := benchDoc(10000, 60)
	n := len(doc.Lines[0].Text)
	b.ResetTimer()
	for b.Loop() {
		doc.Delete(Index{1, n - 1}, Index{1, n})
		doc.Insert(Index{1, n - 1}, "x")
	}
}

func BenchmarkInsertCharManyTags(b *testing.B) {
	doc := benchTaggedDoc(1000, 5000)
	b.ResetTimer()
	for b.Loop() {
		doc.Insert(Index{500, 0}, "x")
		doc.Delete(Index{500, 0}, Index{500, 1})
	}
}

func BenchmarkTagAddMany(b *testing.B) {
	for b.Loop() {
		benchTaggedDoc(1000, 5000)
	}
}

func BenchmarkTagsAt(b *testing.B) {
	doc := benchTaggedDoc(1000, 5000)
	b.ResetTimer()
	for b.Loop() {
		doc.TagsAt(Index{500, 8})
	}
}

func BenchmarkSegmentsForRange(b *testing.B) {
	doc := benchTaggedDoc(1000, 5000)
	tw := newBenchWidget(doc, 600, 400)
	b.ResetTimer()
	for b.Loop() {
		tw.segmentsForRange(500, 0, len(doc.Lines[499].Text))
	}
}

func BenchmarkYviewFractions(b *testing.B) {
	tw := newBenchWidget(benchDoc(2000, 80), 400, 300)
	tw.topLine = 1000
	b.ResetTimer()
	for b.Loop() {
		tw.yviewFractions()
	}
}

func BenchmarkTotalDisplayLines(b *testing.B) {
	tw := newBenchWidget(benchDoc(2000, 80), 400, 300)
	b.ResetTimer()
	for b.Loop() {
		tw.totalDisplayLines()
	}
}

func BenchmarkClampScroll(b *testing.B) {
	tw := newBenchWidget(benchDoc(2000, 80), 400, 300)
	b.ResetTimer()
	for b.Loop() {
		tw.topLine = 1999
		tw.clampScrollPosition()
	}
}

func BenchmarkScrollDown(b *testing.B) {
	tw := newBenchWidget(benchDoc(2000, 80), 400, 300)
	b.ResetTimer()
	for b.Loop() {
		tw.topLine, tw.topCharOffset = 1, 0
		tw.scrollByDisplayLines(3)
		tw.clampScrollPosition()
	}
}

func BenchmarkIndexFromPixel(b *testing.B) {
	tw := newBenchWidget(benchDoc(2000, 80), 400, 300)
	tw.topLine = 1000
	b.ResetTimer()
	for b.Loop() {
		tw.indexFromPixel(200, 150)
	}
}

func BenchmarkSelectionDrag(b *testing.B) {
	tw := newBenchWidget(benchTaggedDoc(2000, 5000), 400, 300)
	tw.topLine = 1000
	tw.selAnchor = Index{1000, 0}
	b.ResetTimer()
	for b.Loop() {
		idx := tw.indexFromPixel(200, 150)
		tw.updateSelection(idx)
	}
}

// BenchmarkWrapLongParagraph lays out one 20k-character logical line.
func BenchmarkWrapLongParagraph(b *testing.B) {
	t := newBenchWidget(docWithText(strings.Repeat("lorem ipsum dolor ", 1111)), 600, 400)
	for b.Loop() {
		t.wrapLine(1, 590, 0, 0, 0)
	}
}
