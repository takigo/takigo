package text

import (
	"strings"
	"testing"
)

func scrolledWidget(t *testing.T, doc *Document, lines int) *TextWidget {
	t.Helper()
	w, _ := newTestPeer(doc, 400, 100)
	w.YScrollCmd = func(float64, float64) {}
	var sb strings.Builder
	for range lines {
		sb.WriteString("line\n")
	}
	w.Insert("1.0", sb.String())
	w.See("end")
	if w.topLine < 10 {
		t.Fatalf("not scrolled: topLine=%d", w.topLine)
	}
	return w
}

func checkView(t *testing.T, w *TextWidget) {
	t.Helper()
	if n := w.doc.LineCount(); w.topLine < 1 || w.topLine > n {
		t.Fatalf("topLine %d outside 1..%d", w.topLine, n)
	}
	w.yviewFractions()
	w.computeVisibleLines()
}

func TestDeleteAllResetsView(t *testing.T) {
	w := scrolledWidget(t, NewDocument(), 300)
	w.Delete("1.0", "end")
	checkView(t, w)
	if w.topLine != 1 || w.topCharOffset != 0 {
		t.Errorf("view = %d,%d, want 1,0", w.topLine, w.topCharOffset)
	}
	if f, l := w.yviewFractions(); f != 0 || l != 1 {
		t.Errorf("fractions = %v,%v, want 0,1", f, l)
	}
}

func TestDeleteAboveViewMovesTopLine(t *testing.T) {
	w := scrolledWidget(t, NewDocument(), 300)
	top := w.topLine
	w.Delete("1.0", "11.0")
	checkView(t, w)
	if w.topLine != top-10 {
		t.Errorf("topLine = %d, want %d", w.topLine, top-10)
	}
}

func TestSelectionDeleteKeepsView(t *testing.T) {
	w := scrolledWidget(t, NewDocument(), 300)
	w.SelectAll()
	w.deleteSelection()
	checkView(t, w)
}

func TestUndoLargeInsertKeepsView(t *testing.T) {
	doc := NewDocument()
	w, _ := newTestPeer(doc, 400, 100)
	w.YScrollCmd = func(float64, float64) {}
	w.Insert("1.0", strings.Repeat("line\n", 300))
	w.See("end")
	w.Edit("undo")
	checkView(t, w)
}

func TestPeerDeleteKeepsOtherPeerView(t *testing.T) {
	doc := NewDocument()
	a := scrolledWidget(t, doc, 300)
	b, _ := newTestPeer(doc, 400, 100)
	b.Delete("1.0", "end")
	checkView(t, a)
	checkView(t, b)
}
