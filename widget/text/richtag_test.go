package text

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/platform"

	"github.com/takigo/takigo/color"
	"github.com/takigo/takigo/font"
)

// wideFont measures 14px per rune, twice fixedFont.
type wideFont struct{}

func (wideFont) Attrs() font.Attributes     { return font.Attributes{} }
func (wideFont) Metrics() font.Metrics      { return font.Metrics{Ascent: 20, Descent: 6, MaxWidth: 14} }
func (wideFont) MeasureString(s string) int { return 14 * utf8.RuneCountInString(s) }
func (wideFont) Close()                     {}

func TestTagRaiseLower(t *testing.T) {
	tw := newBenchWidget(docWithText("abc"), 400, 100)
	for _, n := range []string{"a", "b", "c"} {
		tw.doc.tag(n)
	}
	if got, want := tw.TagNames(), []string{"a", "b", "c", "sel"}; !slices.Equal(got, want) {
		t.Fatalf("names = %v, want %v", got, want)
	}
	tw.TagRaise("a", "")
	if got, want := tw.TagNames(), []string{"b", "c", "sel", "a"}; !slices.Equal(got, want) {
		t.Errorf("after raise a: %v, want %v (above sel, as in Tk)", got, want)
	}
	tw.TagLower("a", "b")
	if got, want := tw.TagNames(), []string{"a", "b", "c", "sel"}; !slices.Equal(got, want) {
		t.Errorf("after lower a below b: %v, want %v", got, want)
	}
	tw.TagRaise("b", "sel")
	if got, want := tw.TagNames(), []string{"a", "c", "sel", "b"}; !slices.Equal(got, want) {
		t.Errorf("after raise b above sel: %v, want %v", got, want)
	}
	tw.TagLower("b", "sel")
	if got, want := tw.TagNames(), []string{"a", "c", "b", "sel"}; !slices.Equal(got, want) {
		t.Errorf("after lower b below sel: %v, want %v", got, want)
	}
	tw.TagRaise("b", "sel")
	tw.doc.tag("d")
	if got, want := tw.TagNames(), []string{"a", "c", "d", "sel", "b"}; !slices.Equal(got, want) {
		t.Errorf("a new tag: %v, want %v (below sel)", got, want)
	}
	for i := 1; i < len(tw.doc.tagsByPriority); i++ {
		if tw.doc.tagsByPriority[i-1].Priority >= tw.doc.tagsByPriority[i].Priority {
			t.Fatalf("priorities not increasing: %v", tw.TagNames())
		}
	}
}

func TestTagDeleteAndRanges(t *testing.T) {
	tw := newBenchWidget(docWithText("one two three\nfour"), 400, 100)
	tw.TagAdd("x", "1.0", "1.3")
	tw.TagAdd("x", "1.8", "2.2")
	r, ok := tw.TagNextRange("x", "1.1", "end")
	if !ok || r.Start != (Index{1, 8}) || r.End != (Index{2, 2}) {
		t.Errorf("nextrange from 1.1 = %+v %v, want 1.8-2.2", r, ok)
	}
	r, ok = tw.TagPrevRange("x", "1.9", "1.0")
	if !ok || r.Start != (Index{1, 8}) {
		t.Errorf("prevrange from 1.9 = %+v %v, want the range at 1.8", r, ok)
	}
	if _, ok := tw.TagNextRange("x", "2.0", "end"); ok {
		t.Error("nextrange found a range starting before index1")
	}
	if got := tw.TagNamesAt("1.1"); !slices.Equal(got, []string{"x"}) {
		t.Errorf("names at 1.1 = %v", got)
	}
	tw.TagDelete("x")
	if len(tw.TagRanges("x")) != 0 || slices.Contains(tw.TagNames(), "x") {
		t.Error("tag delete left the tag")
	}
	tw.TagAdd("sel", "1.0", "1.2")
	tw.TagDelete("sel")
	if len(tw.TagRanges("sel")) != 0 || !slices.Contains(tw.TagNames(), "sel") {
		t.Error("deleting sel must clear it and keep it")
	}
}

func TestTagFontWrapsWithItsWidth(t *testing.T) {
	// 10 characters of a 14px font are 140px; the line is 100px wide.
	tw := newBenchWidget(docWithText("aaaa bbbb cc"), 100+2*0, 200)
	tw.doc.tag("big").Font = wideFont{}
	tw.TagAdd("big", "1.0", "1.end")
	dls := tw.layout.line(1).dls
	if len(dls) < 2 {
		t.Fatalf("got %d display lines, want the big font to wrap", len(dls))
	}
	if w := tw.measureRange(1, 0, 4); w != 56 {
		t.Errorf("measure = %d, want 4×14", w)
	}
}

func TestElide(t *testing.T) {
	tw := newBenchWidget(docWithText("abcdef\nhidden\nlast"), 400, 200)
	hide := tw.doc.tag("hide")
	hide.Elide, hide.ElideSet = true, true
	tw.TagAdd("hide", "1.2", "1.4")
	if w := tw.measureRange(1, 0, 6); w != 4*7 {
		t.Errorf("width with 2 elided = %d, want 28", w)
	}
	before := tw.layout.totalPixels()
	tw.TagAdd("hide", "2.0", "3.0")
	if got := tw.layout.line(2).pixels; got != 0 {
		t.Errorf("elided line takes %d px, want 0", got)
	}
	if after := tw.layout.totalPixels(); after != before-13 {
		t.Errorf("total = %d, want %d", after, before-13)
	}
	if got := tw.Get("1.0", "end"); got != "abcdef\nhidden\nlast" {
		t.Errorf("Get = %q: eliding must keep the text", got)
	}
}

func TestTagWrap(t *testing.T) {
	tw := newBenchWidget(docWithText("aaaa bbbb cccc dddd\naaaa bbbb cccc dddd"), 70, 200)
	tw.doc.tag("nowrap").Wrap, tw.doc.Tags["nowrap"].WrapSet = WrapNone, true
	tw.TagAdd("nowrap", "2.0", "2.end")
	if n := len(tw.layout.line(1).dls); n < 2 {
		t.Errorf("line 1 has %d display lines, want wrapped", n)
	}
	if n := len(tw.layout.line(2).dls); n != 1 {
		t.Errorf("line 2 has %d display lines, want 1 (-wrap none)", n)
	}
}

func TestCustomTabs(t *testing.T) {
	tw := newBenchWidget(docWithText("a\tbb\tcc\tx\t1.25"), 600, 100)
	tw.tabs = []TabStop{{Pos: 70}, {Pos: 210, Align: TabRight}, {Pos: 280, Align: TabCenter}, {Pos: 400, Align: TabNumeric}}
	tw.tabsGen++
	segs := tw.segmentsForRange(1, 0, len(tw.doc.Lines[0].Text))
	at := func(text string) int {
		for _, s := range segs {
			if s.text == text {
				return s.x
			}
		}
		t.Fatalf("no segment %q in %+v", text, segs)
		return 0
	}
	// Left stop: "bb" starts at 70; right: "cc" ends at 210; center: "x"
	// is centred on 280; numeric: the "." of "1.25" is at 400 ("1" before
	// it, 7px).
	if x := at("bb"); x != 70 {
		t.Errorf("bb at %d, want 70", x)
	}
	if x := at("cc"); x != 210-14 {
		t.Errorf("cc at %d, want 196", x)
	}
	if x := at("x"); x != 280-3 {
		t.Errorf("x at %d, want 277", x)
	}
	if x := at("1.25"); x != 400-7 {
		t.Errorf("1.25 at %d, want 393", x)
	}
	stops, err := ParseTabs("70 left 210 right 280 center")
	if err != nil || len(stops) != 3 || stops[1] != (TabStop{210, TabRight}) {
		t.Errorf("ParseTabs = %+v, %v", stops, err)
	}
	if _, err := ParseTabs("20 10"); err == nil {
		t.Error("decreasing tabs accepted")
	}
	if s := stopAt([]TabStop{{Pos: 10}, {Pos: 30}}, 3); s.Pos != 70 {
		t.Errorf("extrapolated stop = %d, want 70", s.Pos)
	}
}

func TestImageAlignAndPad(t *testing.T) {
	tw := newBenchWidget(docWithText("ab"), 400, 200)
	tw.ImageCreate("1.1", testImage{"pic", 20, 30}, ImagePadX(3), ImagePadY(2))
	dl := tw.layout.line(1).dls[0]
	if dl.height != 34 {
		t.Errorf("height = %d, want image + 2×pady = 34", dl.height)
	}
	if w := tw.measureRange(1, 0, 3); w != 14+26 {
		t.Errorf("width = %d, want text + image + 2×padx", w)
	}
	tw.ImageConfigure("1.1", ImageAlignOpt(AlignBaseline))
	dl = tw.layout.line(1).dls[0]
	// The image rises 30+2 above the baseline; the font descends 3.
	if dl.ascent != 32 || dl.height != 35 {
		t.Errorf("baseline image: ascent %d height %d, want 32/35", dl.ascent, dl.height)
	}
	tw.ImageCreate("1.0", testImage{"pic", 5, 5})
	if got := tw.ImageNames(); !slices.Equal(got, []string{"pic#1", "pic"}) {
		t.Errorf("names = %v, want a #1 suffix for the second pic", got)
	}
	if idx, err := tw.Index("pic#1"); err != nil || idx != (Index{1, 0}) {
		t.Errorf("index pic#1 = %v %v", idx, err)
	}
}

func TestBlockInsets(t *testing.T) {
	tw := newBenchWidget(docWithText("before\nq1\nq2\nafter"), 400, 300)
	q := tw.doc.tag("quote")
	q.BlockBackground = &color.Color{Pixel: 0xeeeeee}
	q.LMargin1 = 10
	q.BlockBorder = [4]int{4, 1, 0, 2}
	q.BlockPadding = [4]int{6, 3, 0, 5}
	q.Spacing1 = 7
	tw.TagAdd("quote", "2.0", "4.0")
	p2, p3 := tw.layout.line(2).props, tw.layout.line(3).props
	if p2.block != q || len(p2.blocks) != 1 || !p2.blocks[0].start || p2.blocks[0].end {
		t.Errorf("line 2: blocks %+v, want the opening line", p2.blocks)
	}
	if !p3.blocks[0].end || p3.blocks[0].start {
		t.Errorf("line 3: %+v, want the closing line", p3.blocks)
	}
	if p2.lm1 != 10+4+6 {
		t.Errorf("text margin = %d, want margin + border + padding", p2.lm1)
	}
	if p2.sp1 != 7+1+3 || p2.blocks[0].top != 7 {
		t.Errorf("line 2 spacing %d, box top %d, want 11 inside and 7 outside", p2.sp1, p2.blocks[0].top)
	}
	if p3.sp3 != 2+5 {
		t.Errorf("line 3 spacing below = %d, want 7", p3.sp3)
	}
	if tw.layout.line(4).props.block != nil {
		t.Error("the line after the block is in it")
	}
	// Untagging line 3's newline makes line 3 close the box at once: the
	// next line (outside the change) must be relaid.
	tw.TagRemove("quote", "2.2", "3.0")
	if !tw.layout.line(2).props.blocks[0].end || !tw.layout.line(3).props.blocks[0].start {
		t.Error("splitting the block did not close and reopen it")
	}
}

func TestMarkerOnRangeStart(t *testing.T) {
	tw := newBenchWidget(docWithText("item one\ncontinued\nitem two"), 400, 300)
	tw.doc.tag("li1").Marker = "1."
	tw.doc.tag("li2").Marker = "2."
	tw.TagAdd("li1", "1.0", "3.0")
	tw.TagAdd("li2", "3.0", "3.end")
	got := []string{tw.layout.line(1).props.marker, tw.layout.line(2).props.marker, tw.layout.line(3).props.marker}
	if !slices.Equal(got, []string{"1.", "", "2."}) {
		t.Errorf("markers = %q, want on the first line of each item only", got)
	}
}

func TestSearch(t *testing.T) {
	tw := newBenchWidget(docWithText("Hello world\nhello again\nbye"), 400, 100)
	m, ok := tw.Search("hello", "1.1")
	if !ok || m.Start != (Index{2, 0}) || m.End != (Index{2, 5}) {
		t.Errorf("forward = %+v %v, want 2.0-2.5", m, ok)
	}
	m, ok = tw.Search("Hello", "1.1")
	if !ok || m.Start != (Index{1, 0}) {
		t.Errorf("case-sensitive search = %+v %v, want it to wrap round to 1.0", m, ok)
	}
	m, ok = tw.Search("hello", "2.1", SearchNoCase())
	if !ok || m.Start != (Index{1, 0}) {
		t.Errorf("wrapping nocase = %+v %v, want 1.0", m, ok)
	}
	m, ok = tw.Search("o", "2.0", SearchBackwards())
	if !ok || m.Start != (Index{1, 7}) {
		t.Errorf("backwards = %+v %v, want the last o before 2.0 (1.7)", m, ok)
	}
	if _, ok = tw.Search("bye", "1.0", SearchStop("2.end")); ok {
		t.Error("search went past its stop index")
	}
	m, ok = tw.Search(`a\w+n`, "1.0", SearchRegexp())
	if !ok || m.Start != (Index{2, 6}) || m.Count(tw) != 5 {
		t.Errorf("regexp = %+v %v", m, ok)
	}
	all := tw.SearchAll("l", "2.0")
	if len(all) != 5 || all[0].Start != (Index{2, 2}) || all[len(all)-1].Start != (Index{1, 9}) {
		t.Errorf("all = %+v, want 5 matches from 2.0 round to 1.9", all)
	}
}

func TestNormalizeTagEvent(t *testing.T) {
	for in, want := range map[string]string{
		"<1>": "<Button-1>", "<ButtonPress-3>": "<Button-3>", "<Button-1>": "<Button-1>",
		"<ButtonRelease-1>": "<ButtonRelease-1>", "<Enter>": "<Enter>",
	} {
		if got := normalizeTagEvent(in); got != want {
			t.Errorf("%s → %s, want %s", in, got, want)
		}
	}
}

func TestNestedBlocks(t *testing.T) {
	tw := newBenchWidget(docWithText("quote\ncode\nquote"), 400, 300)
	outer, inner := tw.doc.tag("outer"), tw.doc.tag("inner")
	outer.BlockBorder = [4]int{3, 0, 0, 0}
	outer.BlockPadding = [4]int{5, 2, 0, 2}
	inner.LMargin1 = 8
	inner.BlockBackground = &color.Color{Pixel: 0xeeeeee}
	inner.BlockPadding = [4]int{4, 6, 4, 6}
	tw.TagAdd("outer", "1.0", "end")
	tw.TagAdd("inner", "2.0", "3.0")
	p := tw.layout.line(2).props
	if len(p.blocks) != 2 || p.blocks[0].tag != outer || p.block != inner {
		t.Fatalf("blocks = %+v, want outer then inner", p.blocks)
	}
	if p.lm1 != 8+4 {
		t.Errorf("margin = %d, want the inner box's", p.lm1)
	}
	// Line 2 opens and closes only the inner box: 6 above and below.
	if !p.blocks[1].start || !p.blocks[1].end || p.blocks[0].start || p.sp1 != 6 || p.sp3 != 6 {
		t.Errorf("line 2 = %+v sp %d/%d", p.blocks, p.sp1, p.sp3)
	}
	if p1 := tw.layout.line(1).props; p1.sp1 != 2 || !p1.blocks[0].start {
		t.Errorf("line 1 = %+v sp1 %d, want the outer box opened", p1.blocks, p1.sp1)
	}
}

func TestFloatingImage(t *testing.T) {
	// 13px lines; a 40px float covers the first four display lines' tops
	// (0, 13, 26, 39) and so narrows lines across logical lines.
	tw := newBenchWidget(docWithText("aaaa bbbb cccc dddd eeee\nffff gggg\nhhhh\niiii\njjjj"), 200, 400)
	tw.ImageCreate("1.0", testImage{"pic", 50, 40}, ImageFloatOpt(FloatLeft))
	dls := tw.layout.line(1).dls
	if dls[0].leftMargin != 50 {
		t.Errorf("first display line margin %d, want the float's width", dls[0].leftMargin)
	}
	if dls[0].height != 13 {
		t.Errorf("the float makes its line %d tall, want 13", dls[0].height)
	}
	// 150px beside the float: "aaaa bbbb cccc dddd " is 20 chars (140px).
	if dls[0].endChar != 21 {
		t.Errorf("first line ends at %d, want 21 (the float character and 20 chars)", dls[0].endChar)
	}
	margins := func(l int) (out []int) {
		for _, d := range tw.layout.line(l).dls {
			out = append(out, d.leftMargin)
		}
		return
	}
	// Line 1: 2 display lines (0, 13); line 2 at 26, line 3 at 39, line 4 at 52.
	if got := [][]int{margins(1), margins(2), margins(3), margins(4)}; !slices.Equal(got[0], []int{50, 50}) ||
		!slices.Equal(got[1], []int{50}) || !slices.Equal(got[2], []int{50}) || !slices.Equal(got[3], []int{0}) {
		t.Errorf("margins = %v, want the float beside the first four display lines", got)
	}
	if f := tw.layout.line(1).floats; len(f) != 1 || f[0].x != 0 || f[0].y != 0 {
		t.Errorf("placed floats = %+v", f)
	}
	if c := tw.layout.line(3).carryIn; len(c) != 1 || c[0].bottom != 40-39 {
		t.Errorf("carry into line 3 = %+v, want 1px left", c)
	}
	// Deleting the image frees the lines below, which no edit touched.
	tw.Delete("1.0", "1.1")
	if got := margins(3); !slices.Equal(got, []int{0}) {
		t.Errorf("after deleting the float line 3 margins = %v", got)
	}
}

func TestFloatRightAndInside(t *testing.T) {
	tw := newBenchWidget(docWithText("ab cd ef gh ij kl mn op"), 120, 400)
	tw.ImageCreate("1.3", testImage{"pic", 30, 20}, ImageFloatOpt(FloatRight))
	dls := tw.layout.line(1).dls
	// The float's index is inside the first display line: it starts below.
	if dls[0].rightMargin != 0 || len(dls) < 2 || dls[1].rightMargin != 30 {
		t.Errorf("right margins %d/%d, want the float from the second line", dls[0].rightMargin, dls[1].rightMargin)
	}
	f := tw.layout.line(1).floats
	if len(f) != 1 || f[0].x != 120-30 || f[0].y != dls[0].height {
		t.Errorf("placed = %+v", f)
	}
	if w := tw.measureRange(1, 0, 6); w != 5*7 {
		t.Errorf("width with the float = %d, want the float to take none", w)
	}
}

func TestTagClear(t *testing.T) {
	tw := newBenchWidget(docWithText("a\nb\nc"), 200, 400)
	tw.ImageCreate("1.0", testImage{"pic", 50, 40}, ImageFloatOpt(FloatLeft))
	tw.doc.tag("clear").Clear = ClearBoth
	tw.TagAdd("clear", "2.0", "2.end")
	// Line 1 is 13px tall: line 2 starts 27px lower, below the float.
	if p := tw.layout.line(2); p.props.sp1 != 27 || p.dls[0].leftMargin != 0 {
		t.Errorf("cleared line: sp1 %d margin %d, want 27 and 0", p.props.sp1, p.dls[0].leftMargin)
	}
	if c := tw.layout.line(3).carryIn; len(c) != 0 {
		t.Errorf("floats carried past the cleared line: %+v", c)
	}
}

func TestDump(t *testing.T) {
	tw := newBenchWidget(docWithText("abcd\nef"), 400, 100)
	tw.TagAdd("x", "1.1", "1.3")
	tw.MarkSet("m", "1.2")
	tw.ImageCreate("2.1", testImage{"pic", 5, 5})
	got, err := tw.Dump("1.0", "end")
	if err != nil {
		t.Fatal(err)
	}
	var s []string
	for _, it := range got {
		s = append(s, fmt.Sprintf("%s %q %d.%d", it.Kind, it.Value, it.Index.Line, it.Index.Char))
	}
	want := []string{
		`mark "current" 1.0`, `text "a" 1.0`,
		`tagon "x" 1.1`, `text "b" 1.1`,
		`mark "m" 1.2`, `text "c" 1.2`,
		`tagoff "x" 1.3`, `text "d\n" 1.3`,
		`text "e" 2.0`, `image "pic" 2.1`, `text "f" 2.2`,
	}
	if !slices.Equal(s, want) {
		t.Errorf("dump =\n%s\nwant\n%s", strings.Join(s, "\n"), strings.Join(want, "\n"))
	}
	tags, _ := tw.Dump("1.0", "end", DumpTagOn)
	if len(tags) != 2 || tags[0].Kind != DumpTagOn || tags[1].Kind != DumpTagOff {
		t.Errorf("-tag dump = %+v", tags)
	}
	part, _ := tw.Dump("1.2", "1.3", DumpText)
	if len(part) != 1 || part[0].Value != "c" {
		t.Errorf("one character = %+v", part)
	}
}

func TestTabStyleWordProcessor(t *testing.T) {
	// Default stops every 8×7 = 56px. "abcdefghij" is 70px, past the
	// first stop: a tabular tab takes a space, a word-processor one goes
	// to the next stop, 112.
	tw := newBenchWidget(docWithText("abcdefghij\tx"), 400, 100)
	tw.tabWidth = 8
	if w := tw.measureRange(1, 0, 12); w != 70+7+7 {
		t.Errorf("tabular width = %d, want 84", w)
	}
	tw.tabStyle = TabWordProcessor
	tw.tabsGen++
	if w := tw.measureRange(1, 0, 12); w != 112+7 {
		t.Errorf("word-processor width = %d, want 119", w)
	}
	// Custom stops: past the first, the tab goes to the second.
	tw.tabs = []TabStop{{Pos: 40}, {Pos: 100}}
	if segs := tw.segmentsForRange(1, 0, 12); segs[len(segs)-1].x != 100 {
		t.Errorf("x after the tab = %d, want 100", segs[len(segs)-1].x)
	}
	tw.tabStyle = TabTabular
	if segs := tw.segmentsForRange(1, 0, 12); segs[len(segs)-1].x != 70+7 {
		t.Errorf("tabular x after the tab = %d, want 77 (one space past the text)", segs[len(segs)-1].x)
	}
}

func TestTagSelectColors(t *testing.T) {
	tw := newBenchWidget(docWithText("abcdef"), 400, 100)
	red, blue := &color.Color{Pixel: 0xff0000}, &color.Color{Pixel: 0x0000ff}
	tw.doc.Tags["sel"].Background = blue
	tw.doc.tag("hot").SelectBackground = red
	tw.doc.Tags["hot"].SelectForeground = blue
	tw.TagAdd("hot", "1.0", "1.3")
	tw.TagAdd("sel", "1.2", "1.5")
	segs := tw.segmentsForRange(1, 0, 6)
	bg := func(i int) uint64 {
		if segs[i].bg == nil {
			return 0
		}
		return segs[i].bg.Pixel
	}
	// Segments: "ab" hot, "c" hot+sel, "de" sel, "f".
	if len(segs) != 4 || bg(0) != 0 || bg(1) != 0xff0000 || segs[1].fg != blue || bg(2) != 0x0000ff {
		t.Errorf("backgrounds %x %x %x, fg %v", bg(0), bg(1), bg(2), segs[1].fg)
	}
}

func TestTagDoubleClick(t *testing.T) {
	tw := newBenchWidget(docWithText("link"), 400, 100)
	tw.TagAdd("a", "1.0", "1.end")
	var got []string
	tw.TagBind("a", "<1>", func() { got = append(got, "single") })
	tw.TagBind("a", "<Double-1>", func() { got = append(got, "double") })
	for range 3 {
		tw.fireTagClick(Index{1, 1}, 1, len(got)+1)
	}
	if want := []string{"single", "double", "double"}; !slices.Equal(got, want) {
		t.Errorf("clicks fired %v, want %v (a triple click falls back to the double binding)", got, want)
	}
	var c clickCounter
	ev := func(ms uint32, x int) *event.Event {
		return &event.Event{Type: event.ButtonPressType, Button: 1, Time: platform.Timestamp(ms), RootX: x}
	}
	if n := []int{c.count(ev(1000, 10)), c.count(ev(1200, 12)), c.count(ev(1300, 12)), c.count(ev(2000, 12)), c.count(ev(2100, 40))}; !slices.Equal(n, []int{1, 2, 3, 1, 1}) {
		t.Errorf("click counts = %v", n)
	}
}
