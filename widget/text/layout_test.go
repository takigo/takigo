package text

import (
	"math/rand/v2"
	"reflect"
	"testing"
)

func TestFenwick(t *testing.T) {
	vals := []int{3, 0, 5, 2, 7, 1}
	f := newFenwick(vals)
	sum := 0
	for i, v := range vals {
		if got := f.prefix(i); got != sum {
			t.Errorf("prefix(%d) = %d, want %d", i, got, sum)
		}
		sum += v
	}
	if f.total() != sum {
		t.Errorf("total = %d, want %d", f.total(), sum)
	}
	// find: cumulative ranges are [0,3) 0, [3,3) 1 (empty), [3,8) 2, [8,10) 3, [10,17) 4, [17,18) 5.
	for target, want := range map[int]int{0: 0, 2: 0, 3: 2, 7: 2, 8: 3, 10: 4, 17: 5, 18: 6, 100: 6} {
		if got := f.find(target); got != want {
			t.Errorf("find(%d) = %d, want %d", target, got, want)
		}
	}
	f.add(1, 4)
	if got := f.prefix(2); got != 7 {
		t.Errorf("after add prefix(2) = %d, want 7", got)
	}
	if got := f.find(5); got != 1 {
		t.Errorf("after add find(5) = %d, want 1", got)
	}
	if e := newFenwick(nil); e.total() != 0 || e.find(0) != 0 {
		t.Error("empty fenwick")
	}
}

// freshLayout lays out every line of t from scratch with a cache that
// received no change notifications.
func freshLayout(t *TextWidget) *layoutCache {
	c := &layoutCache{}
	c.init(t)
	c.ensureAll()
	return c
}

func checkLayoutMatches(t *testing.T, step int, tw *TextWidget) {
	t.Helper()
	got := &tw.layout
	want := freshLayout(tw)
	if got.totalPixels() != want.totalPixels() || got.totalDisplayLines() != want.totalDisplayLines() {
		t.Fatalf("step %d: totals %d/%d, want %d/%d", step,
			got.totalPixels(), got.totalDisplayLines(), want.totalPixels(), want.totalDisplayLines())
	}
	for l := 1; l <= tw.doc.LineCount(); l++ {
		g, w := got.line(l), want.line(l)
		if !reflect.DeepEqual(g.dls, w.dls) || !reflect.DeepEqual(g.hs, w.hs) || g.width != w.width || g.props != w.props {
			t.Fatalf("step %d line %d: cached %+v %v %d, fresh %+v %v %d", step, l, g.dls, g.hs, g.width, w.dls, w.hs, w.width)
		}
		if got.pixelsBefore(l) != want.pixelsBefore(l) || got.displayLinesBefore(l) != want.displayLinesBefore(l) {
			t.Fatalf("step %d line %d: prefix sums differ", step, l)
		}
	}
	for y := 0; y < want.totalPixels()+20; y += 7 {
		gl, goPos := got.lineAtPixel(y)
		wl, wo := want.lineAtPixel(y)
		if gl != wl || goPos != wo {
			t.Fatalf("step %d lineAtPixel(%d) = %d/%d, want %d/%d", step, y, gl, goPos, wl, wo)
		}
	}
	if got.maxWidth() != want.maxWidth() {
		t.Fatalf("step %d maxWidth %d, want %d", step, got.maxWidth(), want.maxWidth())
	}
}

func TestLayoutCacheMatchesRecompute(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	tw := newBenchWidget(benchDoc(40, 50), 200, 100)
	doc := tw.doc
	doc.TagConfigure("margin", nil, nil, TagLMargin1(20), TagLMargin2(10), TagSpacing1(3), TagSpacing3(2))
	doc.TagConfigure("plain", nil, nil, TagUnderline(true))
	doc.TagConfigure("sup", nil, nil, TagOffset(4))
	words := []string{"a", "word ", "longer words here ", "\n", "x\ny\nz", "  ", "\n\n"}
	randIdx := func() Index {
		l := rng.IntN(doc.LineCount()) + 1
		return Index{l, rng.IntN(len(doc.Lines[l-1].Text) + 1)}
	}
	for step := range 400 {
		switch rng.IntN(10) {
		case 0, 1, 2:
			doc.Insert(randIdx(), words[rng.IntN(len(words))])
		case 3, 4:
			a, b := randIdx(), randIdx()
			if Compare(a, b) > 0 {
				a, b = b, a
			}
			doc.Delete(a, b)
		case 5:
			a, b := randIdx(), randIdx()
			if Compare(a, b) > 0 {
				a, b = b, a
			}
			doc.TagAdd([]string{"margin", "plain", "sup", "sel"}[rng.IntN(4)], a, b)
		case 6:
			a, b := randIdx(), randIdx()
			if Compare(a, b) > 0 {
				a, b = b, a
			}
			doc.TagRemove([]string{"margin", "plain", "sup", "sel"}[rng.IntN(4)], a, b)
		case 7:
			tw.Win.Width = 100 + rng.IntN(300)
		case 8:
			tw.wrapMode = WrapMode(rng.IntN(3))
		case 9:
			doc.TagConfigure("margin", nil, nil, TagRMargin(rng.IntN(30)))
		}
		if step%3 == 0 || rng.IntN(4) == 0 {
			// Touch the cache in the visible region the way redraws do.
			tw.topLine = rng.IntN(doc.LineCount()) + 1
			tw.topCharOffset = 0
			tw.computeVisibleLines()
			tw.scrollByDisplayLines(rng.IntN(5) - 2)
			tw.clampScrollPosition()
		}
		checkLayoutMatches(t, step, tw)
	}
}

func TestScrollMathOnCache(t *testing.T) {
	tw := newBenchWidget(benchDoc(30, 50), 200, 100)
	total := tw.totalDisplayLines()
	if total < 30 {
		t.Fatalf("expected wrapping, got %d display lines", total)
	}
	tw.topLine, tw.topCharOffset = 1, 0
	tw.scrollByDisplayLines(total + 100)
	if l, o := tw.layout.lineAtDisplayLine(total - 1); tw.topLine != l || tw.topCharOffset != o {
		t.Fatalf("scroll past end: %d/%d, want %d/%d", tw.topLine, tw.topCharOffset, l, o)
	}
	tw.scrollByDisplayLines(-total - 100)
	if tw.topLine != 1 || tw.topCharOffset != 0 {
		t.Fatalf("scroll before start: %d/%d", tw.topLine, tw.topCharOffset)
	}
	tw.YViewMoveTo(0.5)
	first, _ := tw.yviewFractions()
	if first < 0.4 || first > 0.6 {
		t.Fatalf("yview moveto 0.5 gave first=%v", first)
	}
	tw.YViewMoveTo(2)
	tw.clampScrollPosition()
	if tw.computeDisplayLinesBefore(tw.topLine, tw.topCharOffset) != total-(100/13) {
		t.Fatalf("clamp: top display line %d, want %d", tw.computeDisplayLinesBefore(tw.topLine, tw.topCharOffset), total-100/13)
	}
}
