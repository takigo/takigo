package text

import (
	"math/bits"
	"slices"

	"github.com/msorc/takigo/font"
)

// layoutCache keeps each logical line's wrapped display lines and pixel
// height for one TextWidget, as Tk caches pixelHeight per line in the
// B-tree (TkTextUpdateOneLine), with Fenwick trees over the per-line
// totals so scroll arithmetic is O(log lines). Peers sharing a Document
// each have their own cache since width, font and wrap mode differ.
type layoutCache struct {
	t     *TextWidget
	lines []lineLayout
	key   layoutKey

	px, dl    fenwick
	sumsValid bool

	// dirtyLo..dirtyHi bounds the invalid entries (0-based, inclusive).
	dirtyLo, dirtyHi int
}

type lineLayout struct {
	valid  bool
	props  lineProps
	dls    []displayLine // height and ascent set; y, spacing and justify are not
	hs     []int         // display line heights including spacing
	pixels int
	width  int
}

// layoutKey holds the widget state every line's layout depends on; a change
// in any of it invalidates the whole cache.
type layoutKey struct {
	availWidth int
	font       font.Font
	wrapMode   WrapMode
	tagGen     int
	embedded   uint64
}

// Change describes an edit to a Document: lines From..To changed and Delta
// lines were inserted (> 0) or removed (< 0) after From.
type Change struct {
	From, To int
	Delta    int
}

func (c *layoutCache) init(t *TextWidget) {
	c.t = t
	c.dirtyLo, c.dirtyHi = 1, 0
}

func (c *layoutCache) currentKey() layoutKey {
	t := c.t
	h := uint64(len(t.embeddedWindows))<<32 | uint64(len(t.embeddedImages))
	for _, ew := range t.embeddedWindows {
		h = h*1000003 ^ uint64(ew.win.ReqWidth)<<40 ^ uint64(ew.win.ReqHeight)<<20 ^ uint64(ew.padX)<<10 ^ uint64(ew.padY)
	}
	for _, ei := range t.embeddedImages {
		h = h*1000003 ^ uint64(ei.index.Line)<<40 ^ uint64(ei.index.Char)<<20 ^ uint64(ei.img.Height())
	}
	return layoutKey{
		availWidth: t.Win.Width - 2*t.insetX,
		font:       t.Font,
		wrapMode:   t.wrapMode,
		tagGen:     t.doc.layoutGen,
		embedded:   h,
	}
}

// check invalidates everything when the widget state or the line count no
// longer matches what the cache was built for.
func (c *layoutCache) check() {
	n := c.t.doc.LineCount()
	key := c.currentKey()
	if len(c.lines) == n && key == c.key {
		return
	}
	c.key = key
	if len(c.lines) != n {
		c.lines = make([]lineLayout, n)
	}
	c.invalidateAll()
}

func (c *layoutCache) invalidateAll() {
	for i := range c.lines {
		c.lines[i].valid = false
	}
	c.dirtyLo, c.dirtyHi = 0, len(c.lines)-1
	c.sumsValid = false
}

// invalidate marks logical lines from..to (1-based, inclusive) stale.
func (c *layoutCache) invalidate(from, to int) {
	from = max(from, 1)
	to = min(to, len(c.lines))
	for i := from - 1; i < to; i++ {
		c.lines[i].valid = false
	}
	if from <= to {
		if c.dirtyLo > c.dirtyHi {
			c.dirtyLo, c.dirtyHi = from-1, to-1
		} else {
			c.dirtyLo, c.dirtyHi = min(c.dirtyLo, from-1), max(c.dirtyHi, to-1)
		}
	}
}

// apply is the Document change listener.
func (c *layoutCache) apply(ch Change) {
	if len(c.lines) != c.t.doc.LineCount()-ch.Delta {
		return // out of sync; check rebuilds
	}
	switch {
	case ch.Delta > 0:
		c.lines = slices.Insert(c.lines, ch.From, make([]lineLayout, ch.Delta)...)
	case ch.Delta < 0:
		c.lines = slices.Delete(c.lines, ch.From, ch.From-ch.Delta)
	}
	c.invalidate(ch.From, ch.To)
	if ch.Delta != 0 {
		c.sumsValid = false
		c.dirtyHi = len(c.lines) - 1
	}
}

// validate lays out line i (0-based), keeping the prefix sums current.
func (c *layoutCache) validate(i int) {
	t := c.t
	ll := &c.lines[i]
	oldPx, oldDl := ll.pixels, len(ll.dls)
	l := i + 1
	props := t.resolveLineProps(l)
	dls := t.setMetrics(l, t.wrapLine(l, c.key.availWidth, props.lm1, props.lm2, props.rm))
	hs := make([]int, len(dls))
	pixels := 0
	for k, dl := range dls {
		h := dl.height + props.sp2
		if k == 0 {
			h += props.sp1 - props.sp2
		}
		if k == len(dls)-1 {
			h += props.sp3
		}
		hs[k] = h
		pixels += h
	}
	*ll = lineLayout{
		valid:  true,
		props:  props,
		dls:    dls,
		hs:     hs,
		pixels: pixels,
		width:  t.measureRange(l, 0, len(t.doc.Lines[i].Text)),
	}
	if c.sumsValid {
		c.px.add(i, pixels-oldPx)
		c.dl.add(i, len(dls)-oldDl)
	}
}

// line returns the layout of logical line l (1-based).
func (c *layoutCache) line(l int) *lineLayout {
	c.check()
	ll := &c.lines[l-1]
	if !ll.valid {
		c.validate(l - 1)
	}
	// Lines inserted or removed above shift the numbering of cached lines.
	if len(ll.dls) > 0 && ll.dls[0].logicalLine != l {
		for i := range ll.dls {
			ll.dls[i].logicalLine = l
		}
	}
	return ll
}

// ensureAll validates every stale line and rebuilds the prefix sums.
func (c *layoutCache) ensureAll() {
	c.check()
	for i := c.dirtyLo; i <= c.dirtyHi; i++ {
		if !c.lines[i].valid {
			c.validate(i)
		}
	}
	c.dirtyLo, c.dirtyHi = 1, 0
	if !c.sumsValid {
		px := make([]int, len(c.lines))
		dl := make([]int, len(c.lines))
		for i := range c.lines {
			px[i] = c.lines[i].pixels
			dl[i] = len(c.lines[i].dls)
		}
		c.px = newFenwick(px)
		c.dl = newFenwick(dl)
		c.sumsValid = true
	}
}

func (c *layoutCache) totalPixels() int {
	c.ensureAll()
	return c.px.total()
}

func (c *layoutCache) totalDisplayLines() int {
	c.ensureAll()
	return c.dl.total()
}

// pixelsBefore returns the height of lines 1..l-1.
func (c *layoutCache) pixelsBefore(l int) int {
	c.ensureAll()
	return c.px.prefix(l - 1)
}

// displayLinesBefore returns the number of display lines in lines 1..l-1.
func (c *layoutCache) displayLinesBefore(l int) int {
	c.ensureAll()
	return c.dl.prefix(l - 1)
}

// lineAtPixel returns the logical line and display line offset containing
// document pixel y; past the end it returns the last line, offset 0.
func (c *layoutCache) lineAtPixel(y int) (int, int) {
	c.ensureAll()
	i := c.px.find(y)
	if i >= len(c.lines) {
		return len(c.lines), 0
	}
	y -= c.px.prefix(i)
	for k, h := range c.lines[i].hs {
		if y < h {
			return i + 1, k
		}
		y -= h
	}
	return i + 1, len(c.lines[i].hs) - 1
}

// lineAtDisplayLine returns the logical line and offset of display line n
// (0-based); past the end it returns the last line, offset 0.
func (c *layoutCache) lineAtDisplayLine(n int) (int, int) {
	c.ensureAll()
	i := c.dl.find(n)
	if i >= len(c.lines) {
		return len(c.lines), 0
	}
	return i + 1, n - c.dl.prefix(i)
}

// maxWidth returns the pixel width of the widest logical line.
func (c *layoutCache) maxWidth() int {
	c.ensureAll()
	w := 0
	for i := range c.lines {
		w = max(w, c.lines[i].width)
	}
	return w
}

// fenwick is a binary indexed tree over non-negative ints.
type fenwick struct {
	tree []int // 1-based
}

func newFenwick(vals []int) fenwick {
	n := len(vals)
	tree := make([]int, n+1)
	copy(tree[1:], vals)
	for i := 1; i <= n; i++ {
		if j := i + i&-i; j <= n {
			tree[j] += tree[i]
		}
	}
	return fenwick{tree: tree}
}

func (f *fenwick) add(i, delta int) {
	for i++; i < len(f.tree); i += i & -i {
		f.tree[i] += delta
	}
}

// prefix returns the sum of elements [0, i).
func (f *fenwick) prefix(i int) int {
	s := 0
	for i = min(i, len(f.tree)-1); i > 0; i -= i & -i {
		s += f.tree[i]
	}
	return s
}

func (f *fenwick) total() int {
	return f.prefix(len(f.tree) - 1)
}

// find returns the index of the element whose cumulative range contains
// target, that is the largest i with prefix(i) <= target; it is len(vals)
// when target is at or past the total.
func (f *fenwick) find(target int) int {
	n := len(f.tree) - 1
	if n == 0 {
		return 0
	}
	pos := 0
	for step := 1 << (bits.Len(uint(n)) - 1); step > 0; step >>= 1 {
		if next := pos + step; next <= n && f.tree[next] <= target {
			pos = next
			target -= f.tree[next]
		}
	}
	return pos
}
