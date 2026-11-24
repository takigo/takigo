package richtext

import (
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/widget"
)

// testBuild lays out html with a 16px base font, 1 CSS px per pixel and
// text measured at 8px per character.
func testBuild(t *testing.T, html string, opts ...Option) *builder {
	t.Helper()
	cfg := &config{fitImages: true, images: map[string]widget.WidgetImage{}}
	for _, o := range opts {
		o(cfg)
	}
	cfg.measure = func(_ charStyle, s string) int { return 8 * utf8.RuneCountInString(s) }
	return testBuildCfg(html, cfg)
}

func testBuildCfg(html string, cfg *config) *builder {
	tree, docCSS := parseHTML(html)
	st := &styler{rootPx: 16, scale: 1, ptPx: 4.0 / 3}
	st.sheet.parseCSS(DefaultStylesheet, 0)
	for _, css := range append(docCSS, cfg.css...) {
		st.sheet.parseCSS(css, 1)
	}
	root := style{family: "sans-serif", size: 16, whiteSpace: "normal", listStyle: "disc", display: "block"}
	if !hasElement(tree, "body") {
		tree.tag = "body"
	}
	b := newBuilder(cfg, st, "t:")
	b.blockStyle = b.charStyleOf(&root, inline{})
	b.lineNL = b.blockStyle
	if tree.tag == "body" {
		b.walk(tree, &root, inline{})
	} else {
		for _, c := range tree.children {
			b.walk(c, &root, inline{})
		}
	}
	b.finish()
	return b
}

func lines(b *builder) []string { return strings.Split(b.p.text.String(), "\n") }

// styleAt returns the character style of the character at q.
func styleAt(t *testing.T, b *builder, q pos) charStyle {
	t.Helper()
	for cs, name := range b.charTags {
		for _, r := range b.p.ranges {
			if r.name == name && !less(q, r.start) && less(q, r.end) {
				return cs
			}
		}
	}
	t.Fatalf("no character style at %+v", q)
	return charStyle{}
}

func less(a, b pos) bool { return a.line < b.line || a.line == b.line && a.char < b.char }

func TestParseHTMLImpliedEnds(t *testing.T) {
	root, styles := parseHTML(`<!DOCTYPE html><style>p{color:red}</style>` +
		`<p>one<p>two &amp; <b>three</b><ul><li>a<li>b</ul><!-- gone --><br/>x&nbsp;y&#x41;&#65;`)
	if len(styles) != 1 || !strings.Contains(styles[0], "color:red") {
		t.Errorf("styles = %q", styles)
	}
	var tags []string
	var walk func(n *node, depth int)
	walk = func(n *node, depth int) {
		for _, c := range n.children {
			if c.kind == elementNode {
				tags = append(tags, strings.Repeat(".", depth)+c.tag)
				walk(c, depth+1)
			} else {
				tags = append(tags, strings.Repeat(".", depth)+"'"+c.text+"'")
			}
		}
	}
	walk(root, 0)
	want := []string{"style", "p", ".'one'", "p", ".'two & '", ".b", "..'three'",
		"ul", ".li", "..'a'", ".li", "..'b'", "br", "'x\u00a0yAA'"}
	if !slices.Equal(tags, want) {
		t.Errorf("tree =\n%q\nwant\n%q", tags, want)
	}
}

func TestParseHTMLAttributes(t *testing.T) {
	root, _ := parseHTML(`<a href="x.html?a=1&amp;b=2" class='k  j' data-x=3 hidden>t</a>`)
	a := root.children[0]
	if a.attrs["href"] != "x.html?a=1&b=2" || !slices.Equal(a.classes, []string{"k", "j"}) ||
		a.attrs["data-x"] != "3" {
		t.Errorf("attrs = %v classes %v", a.attrs, a.classes)
	}
	if _, ok := a.attrs["hidden"]; !ok {
		t.Error("boolean attribute lost")
	}
}

func TestCSSCascade(t *testing.T) {
	b := testBuild(t, `<style>
		p { color: blue }
		.note { color: green }
		div > p { color: purple }
		#x { color: orange }
		em { color: red !important }
		@media print { p { color: black } }
	</style>
	<p>a</p><p class="note">b</p><div><p>c</p></div><p class="note" id="x">d</p>
	<p style="color: teal">e</p><p class=note><em style="color: gray">f</em></p>`)
	want := []string{"blue", "green", "purple", "orange", "teal", "red"}
	for i, c := range want {
		if got := styleAt(t, b, pos{line: i}).fg; got != c {
			t.Errorf("paragraph %d: color %q, want %q", i, got, c)
		}
	}
}

func TestStyleValues(t *testing.T) {
	st := &styler{rootPx: 16, scale: 1, ptPx: 4.0 / 3}
	st.sheet.parseCSS(`h { font: italic bold 2em serif; margin: 0.5em 1px 2px; color: rgb(255, 0, 128) }
		k { background: hsl(120, 100%, 50%) url(x.png); border: 2px dashed #abc; font-size: 75% }`, 0)
	parent := style{size: 10}
	h := st.compute(&node{kind: elementNode, tag: "h"}, &parent)
	if h.size != 20 || !h.italic || !h.bold || h.family != "serif" {
		t.Errorf("font shorthand: %+v", h)
	}
	if h.margin != [4]float64{10, 1, 2, 1} {
		t.Errorf("margin = %v, want ems of the element's own size", h.margin)
	}
	if h.color != "#ff0080" {
		t.Errorf("color = %q", h.color)
	}
	k := st.compute(&node{kind: elementNode, tag: "k"}, &parent)
	if k.background != "#00ff00" || k.border[left] != 2 || k.borderColor[top] != "#abc" || k.size != 7.5 {
		t.Errorf("k = bg %q border %v %v size %v", k.background, k.border, k.borderColor, k.size)
	}
}

func TestBlocksAndMarginCollapse(t *testing.T) {
	b := testBuild(t, `<h1>Title</h1>
		<p>Some   text
		over lines.</p>
		<p>Second</p>`)
	if got := lines(b); !slices.Equal(got, []string{"Title", "Some text over lines.", "Second"}) {
		t.Fatalf("lines = %q", got)
	}
	// h1: 32px font, bottom margin 0.4em = 12.8; p: top 0. Collapsed: 13.
	if b.lines[0].sp1 != 0 {
		t.Errorf("first line has %d px above, want none", b.lines[0].sp1)
	}
	if b.lines[1].sp1 != 13 {
		t.Errorf("between h1 and p: %d, want max(12.8, 0) = 13", b.lines[1].sp1)
	}
	if b.lines[2].sp1 != 16 {
		t.Errorf("between paragraphs: %d, want 16", b.lines[2].sp1)
	}
	if cs := styleAt(t, b, pos{0, 0}); cs.size != 32 || !cs.bold {
		t.Errorf("h1 style = %+v", cs)
	}
}

func TestLists(t *testing.T) {
	b := testBuild(t, `<ol start="3"><li>three</li><li>four<ul><li>in</li></ul></li></ol>
		<ol type="a"><li>x<li value="5">y</ol>`)
	var markers []string
	var margins []int
	for _, li := range b.lines {
		markers = append(markers, li.marker)
		margins = append(margins, li.para.lm1)
	}
	if want := []string{"3.", "4.", "◦", "a.", "e."}; !slices.Equal(markers, want) {
		t.Errorf("markers = %q, want %q", markers, want)
	}
	if margins[0] != 32 || margins[2] != 64 {
		t.Errorf("indents = %v, want 2em per level", margins)
	}
}

func TestPreformatted(t *testing.T) {
	b := testBuild(t, "<pre>\n  a  b\n\nc\n</pre><p>after</p>")
	if got := lines(b); !slices.Equal(got, []string{"  a  b", "", "c", "after"}) {
		t.Fatalf("lines = %q", got)
	}
	if !b.lines[0].para.nowrap {
		t.Error("pre lines wrap")
	}
	if cs := styleAt(t, b, pos{0, 2}); cs.family != "monospace" {
		t.Errorf("pre font = %q", cs.family)
	}
	var box *tagRange
	for i, r := range b.p.ranges {
		if strings.HasPrefix(r.name, "t:b") {
			box = &b.p.ranges[i]
		}
	}
	if box == nil || box.start != (pos{0, 0}) || box.end != (pos{3, 0}) {
		t.Errorf("pre box = %+v, want lines 0-2 through the last newline", box)
	}
}

func TestBreaksRulesAndInline(t *testing.T) {
	b := testBuild(t, `<p>a<br>b<br><br>c x<sup>2</sup> <a href="#t">go</a> on</p><hr><p id="t">end</p>`)
	if got := lines(b); !slices.Equal(got, []string{"a", "b", "", "c x2 go on", "", "end"}) {
		t.Fatalf("lines = %q", got)
	}
	if b.lines[4].rule == nil || b.lines[4].rule.thickness != 2 {
		t.Errorf("hr line = %+v", b.lines[4])
	}
	if cs := styleAt(t, b, pos{3, 3}); cs.offset <= 0 || cs.size != 12 {
		t.Errorf("sup = %+v", cs)
	}
	if len(b.p.links) != 1 || b.p.links[0].href != "#t" || b.p.links[0].hoverTag == "" {
		t.Errorf("links = %+v", b.p.links)
	}
	if b.p.anchors["t"] != (pos{line: 5}) {
		t.Errorf("anchor = %+v", b.p.anchors["t"])
	}
	if cs := styleAt(t, b, pos{3, 5}); !cs.underline || cs.fg != "#0969da" {
		t.Errorf("link style = %+v", cs)
	}
	if cs := styleAt(t, b, pos{3, 4}); cs.underline {
		t.Error("the space before the link is underlined")
	}
	if cs := styleAt(t, b, pos{3, 7}); cs.underline {
		t.Error("the space after the link is underlined")
	}
	for _, r := range b.p.ranges {
		if r.name == b.p.links[0].tag && r.start != (pos{3, 5}) {
			t.Errorf("link range starts at %+v, want 3.5", r.start)
		}
	}
}

func TestTable(t *testing.T) {
	b := testBuild(t, `<style>td,th{padding:0 8px;border:none}</style><table>
		<tr><th>Name</th><th>N</th></tr>
		<tr><td>apple</td><td style="text-align:right">10</td></tr>
		<tr><td>fig</td><td align=right>200</td></tr></table>`)
	if got := lines(b); !slices.Equal(got, []string{"Name\tN", "apple\t10", "fig\t200"}) {
		t.Fatalf("lines = %q", got)
	}
	// Column 0 is 5×8 + 16 = 56 wide, column 1 3×8 + 16 = 40: the right
	// aligned cells end at 56 + 40 - 8.
	if s := b.lines[1].tabs; len(s) != 1 || s[0].Pos != 88 || s[0].Align != 1 {
		t.Errorf("row stops = %+v", s)
	}
	if s := b.lines[0].tabs; len(s) != 1 || s[0].Pos != 64 {
		t.Errorf("header stops = %+v, want left at 56 + 8", s)
	}
	if cs := styleAt(t, b, pos{0, 0}); !cs.bold {
		t.Error("th is not bold")
	}
}

func TestImageAltAndSize(t *testing.T) {
	b := testBuild(t, `<p>x <img src="missing.png" alt="logo"> y</p>`,
		ImageResolver(func(string) (widget.WidgetImage, error) { return nil, nil }))
	if got := lines(b); !slices.Equal(got, []string{"x [logo] y"}) {
		t.Errorf("lines = %q", got)
	}
	cases := []struct {
		css      string
		avail    int
		wantW    int
		wantH    int
		fit      bool
		w0, h0   int
		describe string
	}{
		{"", 0, 100, 50, true, 100, 50, "natural size"},
		{"width: 50px", 0, 50, 25, true, 100, 50, "width keeps the aspect"},
		{"height: 10px", 0, 20, 10, true, 100, 50, "height keeps the aspect"},
		{"width: 50%", 300, 150, 75, true, 100, 50, "percent of the space"},
		{"", 40, 40, 20, true, 100, 50, "fitted to the space"},
		{"", 40, 100, 50, false, 100, 50, "not fitted"},
		{"max-width: 30px", 0, 30, 15, true, 100, 50, "max-width"},
	}
	st := &styler{rootPx: 16, scale: 1, ptPx: 1}
	for _, c := range cases {
		var s style
		for _, d := range parseDeclarations(c.css) {
			st.apply(&s, &s, d)
		}
		if w, h := imageSize(c.w0, c.h0, &s, c.avail, c.fit); w != c.wantW || h != c.wantH {
			t.Errorf("%s: %dx%d, want %dx%d", c.describe, w, h, c.wantW, c.wantH)
		}
	}
}

func TestListMarkers(t *testing.T) {
	for _, c := range []struct {
		style string
		n     int
		want  string
	}{
		{"decimal", 12, "12."}, {"lower-alpha", 28, "ab."}, {"upper-roman", 1994, "MCMXCIV."},
		{"disc", 3, "•"}, {"none", 1, ""},
	} {
		if got := listMarker(c.style, c.n); got != c.want {
			t.Errorf("%s %d = %q, want %q", c.style, c.n, got, c.want)
		}
	}
}

func TestBaseFontSize(t *testing.T) {
	_, root := newStyler(font.Attributes{Family: "x", Size: -20}, nil, &config{})
	if root.size != 20 || root.family != "x" {
		t.Errorf("root = %+v", root)
	}
}

func FuzzParseHTML(f *testing.F) {
	for _, s := range []string{"<p>a<b>b</p>c", "<table><tr><td>x", "<pre>\n x</pre>", "&#xffffff;<!--", "<a href=x>"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		cfg := &config{fitImages: true, images: map[string]widget.WidgetImage{},
			measure: func(_ charStyle, s string) int { return len(s) }}
		cfg.resolver = func(string) (widget.WidgetImage, error) { return nil, nil }
		b := testBuildCfg(s, cfg)
		n := strings.Count(b.p.text.String(), "\n") + 1
		for _, li := range b.lines {
			if li.line >= n {
				t.Fatalf("line info for line %d of %d", li.line, n)
			}
		}
	})
}

type fakeImage struct{ w, h int }

func (f fakeImage) Width() int  { return f.w }
func (f fakeImage) Height() int { return f.h }
func (fakeImage) Draw(platform.DisplayServer, platform.DrawableID, platform.GCID, int, int, int, int, int, int, int, uint64) {
}

func TestFloatedImage(t *testing.T) {
	b := testBuild(t, `<p><img src="x" style="float: right; margin: 0 0 4px 8px">
		Text beside it.</p>`,
		ImageResolver(func(string) (widget.WidgetImage, error) { return fakeImage{40, 30}, nil }))
	if got := lines(b); !slices.Equal(got, []string{"\uFFFCText beside it."}) {
		t.Errorf("lines = %q, want no space after the float", got)
	}
	b2 := testBuild(t, `<p>a</p><h2 style="clear: both">b</h2><p>c</p>`)
	if b2.lines[1].para.clear != "both" || b2.lines[2].para.clear != "" {
		t.Errorf("clear = %q/%q, want only on the heading", b2.lines[1].para.clear, b2.lines[2].para.clear)
	}
	if len(b.p.images) != 1 || len(b.p.images[0].opts) != 4 {
		t.Errorf("image options = %d, want align, float, padx, pady", len(b.p.images[0].opts))
	}
}

func TestTableWrapsCells(t *testing.T) {
	// 8px per character; the text area is 200px. Column 0 wants "short"
	// (40 + 16 padding), column 1 a 25-word sentence: it wraps.
	cfg := &config{fitImages: true, images: map[string]widget.WidgetImage{}, width: 200,
		measure: func(_ charStyle, s string) int { return 8 * utf8.RuneCountInString(s) }}
	cfg.css = []string{"td { padding: 0 8px; border: none }"}
	b := testBuildCfg(`<table><tr><td>short</td><td>one two three four five six seven eight nine</td></tr>
		<tr><td>a<br>b<br>c</td><td>x</td></tr></table>`, cfg)
	got := lines(b)
	// Column 1 gets the rest, 144px: 128 for text, 16 characters.
	want := []string{"short\tone two three", "\tfour five six", "\tseven eight nine", "a\tx", "b", "c"}
	if !slices.Equal(got, want) {
		t.Fatalf("lines =\n%q\nwant\n%q", got, want)
	}
	// Every line of a row shares its stops; column 1 starts at 56 + 8.
	for _, li := range b.lines[:3] {
		if len(li.tabs) != 1 || li.tabs[0].Pos != 64 {
			t.Errorf("line %d stops %+v", li.line, li.tabs)
		}
	}
	// One box per row, over all its lines.
	var boxes []tagRange
	for _, r := range b.p.ranges {
		if strings.HasPrefix(r.name, "t:b") {
			boxes = append(boxes, r)
		}
	}
	if len(boxes) != 2 || boxes[0].start.line != 0 || boxes[0].end.line != 3 || boxes[1].end.line != 6 {
		t.Errorf("row boxes = %+v", boxes)
	}
}

func TestColumnWidths(t *testing.T) {
	c := func(minW, maxW int) cell { return cell{minW: minW, maxW: maxW, colspan: 1, rowspan: 1} }
	widths := func(rows [][]cell, avail int) []int {
		grid, n := placeCells(rows)
		return columnWidths(grid, n, avail)
	}
	rows := [][]cell{{c(10, 20), c(30, 130)}}
	for _, tc := range []struct {
		avail int
		want  []int
	}{
		{0, []int{20, 130}},   // unknown width: unwrapped
		{500, []int{20, 130}}, // fits
		{30, []int{10, 30}},   // not even the minimum fits
		{100, []int{15, 85}},  // 60 extra shared 10:100, the rounding to the last
	} {
		if got := widths(rows, tc.avail); !slices.Equal(got, tc.want) {
			t.Errorf("avail %d: %v, want %v", tc.avail, got, tc.want)
		}
	}
	fixed := c(5, 50)
	fixed.fixed = 20
	if got := widths([][]cell{{fixed, c(10, 100)}}, 60); got[0] != 20 {
		t.Errorf("fixed column = %d, want 20", got[0])
	}
	// A cell spanning two columns that are narrower than it widens both.
	wide := c(60, 60)
	wide.colspan = 2
	if got := widths([][]cell{{wide}, {c(10, 10), c(20, 20)}}, 0); !slices.Equal(got, []int{25, 35}) {
		t.Errorf("spanned columns = %v, want the 30 they lack shared", got)
	}
}

func TestTableSpans(t *testing.T) {
	cfg := &config{fitImages: true, images: map[string]widget.WidgetImage{},
		measure: func(_ charStyle, s string) int { return 8 * utf8.RuneCountInString(s) }}
	cfg.css = []string{"td { padding: 0; border: none }"}
	b := testBuildCfg(`<table>
		<tr><td rowspan="3">tall<br>cell<br>of<br>four</td><td colspan="2">wide</td></tr>
		<tr><td>b</td><td>c</td></tr>
		<tr><td>d</td><td>e</td></tr></table>`, cfg)
	// The rowspan cell's four lines flow down its three rows, the last row
	// taking two lines; the colspan cell skips a tab stop.
	want := []string{"tall\twide", "cell\tb\tc", "of\td\te", "four"}
	if got := lines(b); !slices.Equal(got, want) {
		t.Fatalf("lines =\n%q\nwant\n%q", got, want)
	}
}

func TestTableImage(t *testing.T) {
	b := testBuild(t, `<table><tr><td><img src="x" style="margin: 0 2px"> pic</td><td>y</td></tr></table>`,
		ImageResolver(func(string) (widget.WidgetImage, error) { return fakeImage{20, 10}, nil }))
	if got := lines(b); !slices.Equal(got, []string{"\uFFFC pic\ty"}) {
		t.Errorf("lines = %q", got)
	}
	if len(b.p.images) != 1 {
		t.Fatalf("%d images placed", len(b.p.images))
	}
	// Column 0: the image (20 + 2×2), " pic" and 13px of padding each side;
	// column 1's text starts 13px into it.
	if s := b.lines[0].tabs; len(s) != 1 || s[0].Pos != 24+4*8+3*13 {
		t.Errorf("stops = %+v", s)
	}
}
