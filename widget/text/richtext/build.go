package richtext

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/takigo/takigo/cursor"
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/text"
)

// pos is a position in the built text: line from 0, char a rune offset.
type pos struct{ line, char int }

// charStyle is everything a character tag sets.
type charStyle struct {
	family            string
	size              int
	bold, italic      bool
	underline, strike bool
	fg, bg            string
	offset            int
}

// paraKey is everything a paragraph tag sets.
type paraKey struct {
	lm1, lm2, rm int
	align        string
	nowrap       bool
	tabs         string
	clear        string
}

// lineInfo is the line-level formatting of one built line.
type lineInfo struct {
	line     int
	para     paraKey
	tabs     []text.TabStop
	sp1, sp3 int
	marker   string
	rule     *ruleDef
}

type ruleDef struct {
	thickness int
	color     string
}

// box is a block element with a background, border or padding.
type box struct {
	name               string
	lmargin, rmargin   int
	bg, borderColor    string
	border, padding    [4]int // left, top, right, bottom (the widget's order)
	radius             int
	startLine, endLine int
	open               bool // no line yet
}

// tagDef is a tag the plan creates; kind orders the tags' priorities.
type tagDef struct {
	name string
	kind int
	opts []text.TagOption
}

// Tag kinds, lowest priority first: a paragraph's margins must win over
// the margins of the box around it.
const (
	kindBox = iota
	kindChar
	kindPara
	kindSpacing
	kindLine
	kindLink
	kindHover
)

type tagRange struct {
	name       string
	start, end pos
}

type planImage struct {
	at   pos
	img  widget.WidgetImage
	opts []text.ImageOption
}

// link is one <a href>: its tag and the tag of its hover style, if any.
type link struct {
	tag, hoverTag, href string
}

// plan is what Insert puts in the widget: the text with a U+FFFC for each
// image, the tags to create and the ranges to tag.
type plan struct {
	text    strings.Builder
	tags    []tagDef
	ranges  []tagRange
	images  []planImage
	links   []link
	anchors map[string]pos
}

// inline is what an inline element passes to its content.
type inline struct {
	bg     string
	offset int
	link   string
}

type builder struct {
	cfg  *config
	st   *styler
	p    plan
	pre  string // tag name prefix of this insert
	seq  int
	cur  pos
	tags map[string]bool

	lineOpen, anyLine bool
	atLineStart       bool
	pendingSpace      bool
	pendingBreak      bool // a newline in preformatted text, not yet acted on
	pendingMargin     float64
	pendingMarker     string
	lines             []lineInfo
	lineNL            charStyle // the newline style of the current line
	spaceCS           charStyle // the style of the pending space
	spaceIn           inline

	// The block context of the line being built.
	left, right float64
	align       string
	nowrap      bool
	clear       string
	textIndent  float64
	blockStyle  charStyle

	boxes    []*box
	counters []int
	charTags map[charStyle]string
	paraTags map[string]string
	spTags   map[string]string
}

func newBuilder(cfg *config, st *styler, prefix string) *builder {
	return &builder{
		cfg: cfg, st: st, pre: prefix,
		tags:     map[string]bool{},
		charTags: map[charStyle]string{},
		paraTags: map[string]string{},
		spTags:   map[string]string{},
		p:        plan{anchors: map[string]pos{}},
	}
}

func (b *builder) newName(kind string) string {
	b.seq++
	return b.pre + kind + strconv.Itoa(b.seq)
}

func (b *builder) define(name string, kind int, opts ...text.TagOption) {
	if !b.tags[name] {
		b.tags[name] = true
		b.p.tags = append(b.p.tags, tagDef{name: name, kind: kind, opts: opts})
	}
}

func (b *builder) addRange(name string, start, end pos) {
	if start == end {
		return
	}
	if k := len(b.p.ranges) - 1; k >= 0 && b.p.ranges[k].name == name && b.p.ranges[k].end == start {
		b.p.ranges[k].end = end
		return
	}
	b.p.ranges = append(b.p.ranges, tagRange{name: name, start: start, end: end})
}

func px(f float64) int { return int(math.Round(f)) }

// charStyleOf is the character style of text with style s.
func (b *builder) charStyleOf(s *style, in inline) charStyle {
	return charStyle{
		family: s.family, size: max(px(s.size), 1), bold: s.bold, italic: s.italic,
		underline: s.underline, strike: s.strike, fg: s.color, bg: in.bg, offset: in.offset,
	}
}

func (b *builder) charTag(cs charStyle) string {
	if name, ok := b.charTags[cs]; ok {
		return name
	}
	name := b.newName("c")
	b.charTags[cs] = name
	attrs := font.Attributes{Family: cs.family, Size: -float64(cs.size)}
	if cs.bold {
		attrs.Weight = font.WeightBold
	}
	if cs.italic {
		attrs.Slant = font.SlantItalic
	}
	opts := []text.TagOption{text.TagFont(attrs)}
	if cs.fg != "" {
		opts = append(opts, text.TagForeground(cs.fg))
	}
	if cs.bg != "" {
		opts = append(opts, text.TagBackground(cs.bg))
	}
	if cs.underline {
		opts = append(opts, text.TagUnderline(true))
	}
	if cs.strike {
		opts = append(opts, text.TagOverstrike(true))
	}
	if cs.offset != 0 {
		opts = append(opts, text.TagOffset(cs.offset))
	}
	b.define(name, kindChar, opts...)
	return name
}

// openLine starts a new line with the current block context.
func (b *builder) openLine() {
	if b.anyLine {
		start := b.cur
		b.p.text.WriteByte('\n')
		b.cur = pos{line: b.cur.line + 1}
		b.addRange(b.charTag(b.lineNL), start, b.cur)
	} else {
		b.pendingMargin = 0 // no margin above the first line
	}
	b.anyLine, b.lineOpen, b.atLineStart, b.pendingSpace = true, true, true, false
	b.lines = append(b.lines, lineInfo{
		line: b.cur.line,
		para: paraKey{
			lm1: px(b.left + b.textIndent), lm2: px(b.left), rm: px(b.right),
			align: b.align, nowrap: b.nowrap, clear: b.clear,
		},
		sp1:    px(b.pendingMargin),
		marker: b.pendingMarker,
	})
	b.pendingMargin, b.pendingMarker, b.clear = 0, "", ""
	b.lineNL = b.blockStyle
	for _, bx := range b.boxes {
		if bx.open {
			bx.open, bx.startLine = false, b.cur.line
		}
	}
}

func (b *builder) ensureLine() {
	if !b.lineOpen {
		b.openLine()
	}
}

func (b *builder) closeLine() {
	b.lineOpen, b.pendingSpace = false, false
}

// lineBreak ends the line (a <br> or a newline in preformatted text); an
// empty line is kept.
func (b *builder) lineBreak() {
	b.ensureLine()
	b.closeLine()
}

// emit appends text in one style; it holds no newline.
func (b *builder) emit(s string, cs charStyle, in inline) {
	if s == "" {
		return
	}
	b.ensureLine()
	start := b.cur
	b.p.text.WriteString(s)
	b.cur.char += utf8.RuneCountInString(s)
	b.addRange(b.charTag(cs), start, b.cur)
	if in.link != "" {
		b.addRange(in.link, start, b.cur)
	}
	b.atLineStart = false
}

// carriedSpace adds the space a previous text node ended with, in that
// node's style: a collapsed space belongs to the node its first whitespace
// character is in, so a link's underline stops at the link.
func (b *builder) carriedSpace() {
	if b.pendingSpace {
		b.pendingSpace = false
		b.emit(" ", b.spaceCS, b.spaceIn)
	}
}

// carrySpace keeps a text node's trailing space for the next text.
func (b *builder) carrySpace(space bool, cs charStyle, in inline) {
	if space && b.lineOpen && !b.pendingSpace {
		b.pendingSpace, b.spaceCS, b.spaceIn = true, cs, in
	}
}

// text adds a text node's content as its white-space property says.
func (b *builder) text(t string, s *style, in inline) {
	cs := b.charStyleOf(s, in)
	switch s.whiteSpace {
	case "pre", "pre-wrap", "break-spaces":
		b.preText(t, cs, in, false)
		return
	case "pre-line":
		b.preText(t, cs, in, true)
		return
	}
	var run strings.Builder
	space := false
	for _, r := range t {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\f' {
			if run.Len() > 0 || b.lineOpen && !b.atLineStart {
				space = true
			}
			continue
		}
		if run.Len() == 0 && b.pendingSpace {
			b.carriedSpace()
			space = false
		}
		if space {
			run.WriteByte(' ')
			space = false
		}
		run.WriteRune(r)
	}
	b.emit(run.String(), cs, in)
	b.carrySpace(space, cs, in)
}

// preText keeps spaces (unless collapse) and turns newlines into line
// breaks; a newline just before the end of the block makes no empty line.
func (b *builder) preText(t string, cs charStyle, in inline, collapse bool) {
	t = strings.ReplaceAll(t, "\r\n", "\n")
	var run strings.Builder
	space := false
	flush := func() {
		b.emit(run.String(), cs, in)
		run.Reset()
	}
	for _, r := range t {
		if r == '\n' {
			flush()
			if b.pendingBreak {
				b.lineBreak()
			}
			b.pendingBreak = true
			continue
		}
		if b.pendingBreak {
			flush()
			b.lineBreak()
			b.pendingBreak = false
		}
		if collapse && unicode.IsSpace(r) {
			if b.lineOpen && !b.atLineStart || run.Len() > 0 {
				space = true
			}
			continue
		}
		if run.Len() == 0 && b.pendingSpace {
			b.carriedSpace()
			space = false
		}
		if space {
			run.WriteByte(' ')
			space = false
		}
		if r == '\r' {
			continue
		}
		run.WriteRune(r)
	}
	flush()
	b.carrySpace(space, cs, in)
}

// walk adds the content of element n, whose parent has style parent.
func (b *builder) walk(n *node, parent *style, in inline) {
	if n.kind == textNode {
		b.text(n.text, parent, in)
		return
	}
	s := b.st.compute(n, parent)
	if s.display == "none" {
		return
	}
	if n.id != "" {
		b.anchor(n.id)
	}
	if name, ok := n.attr("name"); ok && n.tag == "a" {
		b.anchor(name)
	}
	switch n.tag {
	case "br":
		b.pendingSpace = false
		b.lineBreak()
		return
	case "img":
		b.image(n, &s, in)
		return
	case "hr":
		b.rule(&s)
		return
	case "wbr":
		return
	}
	switch s.display {
	case "block", "list-item", "flex", "grid", "table-row", "table-cell", "table-caption":
		b.block(n, &s)
	case "table":
		b.table(n, &s)
	default:
		child := in
		if s.background != "" {
			child.bg = s.background
		}
		switch s.verticalAlign {
		case "super":
			child.offset += px(parent.size * 0.35)
		case "sub":
			child.offset -= px(parent.size * 0.2)
		}
		if href, ok := n.attr("href"); ok && n.tag == "a" {
			child.link = b.link(n, href)
		}
		for _, c := range n.children {
			b.walk(c, &s, child)
		}
	}
}

func (b *builder) anchor(name string) {
	if _, ok := b.p.anchors[name]; !ok {
		if b.lineOpen {
			b.p.anchors[name] = b.cur
		} else {
			b.p.anchors[name] = pos{line: b.cur.line + boolInt(b.anyLine)}
		}
	}
}

// link creates the tag of a link and its hover style.
func (b *builder) link(n *node, href string) string {
	name := b.newName("a")
	b.define(name, kindLink, text.TagCursor(cursor.Hand2))
	l := link{tag: name, href: href}
	if decls := b.st.hoverDeclarations(n); len(decls) > 0 {
		var hs style
		for _, d := range decls {
			b.st.apply(&hs, &hs, d)
		}
		var opts []text.TagOption
		if hs.color != "" {
			opts = append(opts, text.TagForeground(hs.color))
		}
		if hs.background != "" {
			opts = append(opts, text.TagBackground(hs.background))
		}
		if hs.underline {
			opts = append(opts, text.TagUnderline(true))
		}
		if len(opts) > 0 {
			l.hoverTag = name + "h"
			b.define(l.hoverTag, kindHover, opts...)
		}
	}
	b.p.links = append(b.p.links, l)
	return name
}

// block adds a block element: a new line, collapsed vertical margins,
// indents, an optional box and a list marker.
func (b *builder) block(n *node, s *style) {
	b.closeLine()
	b.pendingBreak = false
	b.pendingMargin = max(b.pendingMargin, s.margin[top])
	if s.clear == "left" || s.clear == "right" || s.clear == "both" {
		b.clear = s.clear // for the block's first line
	}
	saved := struct {
		left, right, indent float64
		align               string
		nowrap              bool
		bs                  charStyle
	}{b.left, b.right, b.textIndent, b.align, b.nowrap, b.blockStyle}

	b.left += s.margin[left]
	b.right += s.margin[right]
	var bx *box
	if s.background != "" || s.border != [4]float64{} || s.padding != [4]float64{} {
		bx = &box{
			name: b.newName("b"), lmargin: px(b.left), rmargin: px(b.right),
			bg: s.background, borderColor: firstNonEmpty(s.borderColor[:]...),
			border:  [4]int{px(s.border[left]), px(s.border[top]), px(s.border[right]), px(s.border[bottom])},
			padding: [4]int{px(s.padding[left]), px(s.padding[top]), px(s.padding[right]), px(s.padding[bottom])},
			radius:  px(s.radius), open: true,
		}
		b.openBox(bx)
	}
	b.left += s.border[left] + s.padding[left]
	b.right += s.border[right] + s.padding[right]
	b.align, b.textIndent = s.align, s.textIndent
	b.nowrap = s.whiteSpace == "nowrap" || s.whiteSpace == "pre"
	b.blockStyle = b.charStyleOf(s, inline{})

	if s.display == "list-item" {
		b.pendingMarker = b.marker(n, s)
	}
	isList := n.tag == "ol" || n.tag == "ul" || n.tag == "menu"
	if isList {
		start := 1
		if v, err := strconv.Atoi(n.attrs["start"]); err == nil {
			start = v
		}
		b.counters = append(b.counters, start-1)
	}
	children := n.children
	if n.tag == "pre" || n.tag == "listing" {
		// A newline right after the start tag is not content.
		if len(children) > 0 && children[0].kind == textNode && strings.HasPrefix(children[0].text, "\n") {
			first := *children[0]
			first.text = first.text[1:]
			children = append([]*node{&first}, children[1:]...)
		}
	}
	for _, c := range children {
		b.walk(c, s, inline{})
	}
	if isList {
		b.counters = b.counters[:len(b.counters)-1]
	}
	if b.pendingMarker != "" { // an empty item still shows its marker
		b.openLine()
	}
	b.closeLine()
	b.pendingBreak = false
	if bx != nil {
		b.closeBox(bx)
	}
	b.left, b.right, b.textIndent = saved.left, saved.right, saved.indent
	b.align, b.nowrap, b.blockStyle = saved.align, saved.nowrap, saved.bs
	b.pendingMargin = max(b.pendingMargin, s.margin[bottom])
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

// openBox starts a box and defines its tag, so that a box inside another
// has the higher priority.
func (b *builder) openBox(bx *box) {
	b.boxes = append(b.boxes, bx)
	opts := []text.TagOption{
		text.TagLMargin1(bx.lmargin), text.TagRMargin(bx.rmargin),
		text.TagBlockPaddingSides(bx.padding[0], bx.padding[1], bx.padding[2], bx.padding[3]),
		text.TagBlockBorderSides(bx.border[0], bx.border[1], bx.border[2], bx.border[3]),
	}
	if bx.bg != "" {
		opts = append(opts, text.TagBlockBackground(bx.bg))
	}
	if bx.borderColor != "" {
		opts = append(opts, text.TagBlockBorderColor(bx.borderColor))
	}
	if bx.radius > 0 {
		opts = append(opts, text.TagBlockRadius(bx.radius))
	}
	b.define(bx.name, kindBox, opts...)
}

// closeBox tags the lines of a box from its first line through the newline
// of its last one.
func (b *builder) closeBox(bx *box) {
	b.boxes = b.boxes[:len(b.boxes)-1]
	if bx.open {
		return // no lines
	}
	bx.endLine = b.cur.line
	b.p.ranges = append(b.p.ranges, tagRange{name: bx.name,
		start: pos{line: bx.startLine}, end: pos{line: bx.endLine + 1}})
}

// marker returns a list item's marker and counts it.
func (b *builder) marker(n *node, s *style) string {
	k := len(b.counters) - 1
	if k < 0 {
		b.counters = append(b.counters, 0)
		k = 0
	}
	if v, err := strconv.Atoi(n.attrs["value"]); err == nil {
		b.counters[k] = v - 1
	}
	b.counters[k]++
	return listMarker(s.listStyle, b.counters[k])
}

// rule adds a horizontal rule: an empty line with a line drawn across it.
func (b *builder) rule(s *style) {
	b.closeLine()
	b.pendingMargin = max(b.pendingMargin, s.margin[top])
	thick, col := px(s.border[top]), s.borderColor[top]
	if thick == 0 {
		thick = max(px(s.border[bottom]), 1)
		col = cmpOr(s.borderColor[bottom], col)
	}
	saved := b.blockStyle
	b.blockStyle = b.charStyleOf(s, inline{})
	b.blockStyle.size = max(thick, b.blockStyle.size/3)
	b.openLine()
	b.lines[len(b.lines)-1].rule = &ruleDef{thickness: thick, color: cmpOr(col, "#808080")}
	b.closeLine()
	b.blockStyle = saved
	b.pendingMargin = max(b.pendingMargin, s.margin[bottom])
}

func cmpOr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// image adds an <img>, scaled to its width and height and to the space
// available, or its alt text when it cannot be loaded.
func (b *builder) image(n *node, s *style, in inline) {
	img := b.cfg.loadImage(n.attrs["src"])
	if img == nil {
		if alt := n.attrs["alt"]; alt != "" {
			b.text("["+alt+"]", s, in)
		}
		return
	}
	avail := 0
	if b.cfg.width > 0 {
		avail = b.cfg.width - px(b.left) - px(b.right)
	}
	img = fitImage(img, s, avail, b.cfg.fitImages)
	floating := s.float == "left" || s.float == "right"
	if !floating {
		b.carriedSpace()
	}
	atStart := !b.lineOpen || b.atLineStart
	b.placeImage(img, imageOptions(s, true), b.charStyleOf(s, in), in)
	// A float is not on the line: the text after it starts the line.
	b.atLineStart = floating && atStart
}

// imageOptions are the text widget options of an image with style s: its
// vertical-align, its margins as padding and, if floats, its float.
func imageOptions(s *style, floats bool) []text.ImageOption {
	align := text.AlignBaseline
	switch s.verticalAlign {
	case "middle", "center":
		align = text.AlignCenter
	case "top", "text-top":
		align = text.AlignTop
	case "bottom", "text-bottom":
		align = text.AlignBottom
	}
	opts := []text.ImageOption{text.ImageAlignOpt(align)}
	padX, padY := max(s.margin[left], s.margin[right]), max(s.margin[top], s.margin[bottom])
	switch {
	case floats && s.float == "left":
		// A float's gap is on its text side only.
		opts = append(opts, text.ImageFloatOpt(text.FloatLeft))
		padX, padY = s.margin[right], s.margin[bottom]
	case floats && s.float == "right":
		opts = append(opts, text.ImageFloatOpt(text.FloatRight))
		padX, padY = s.margin[left], s.margin[bottom]
	}
	if padX > 0 {
		opts = append(opts, text.ImagePadX(px(padX)))
	}
	if padY > 0 {
		opts = append(opts, text.ImagePadY(px(padY)))
	}
	return opts
}

// placeImage adds an image's placeholder character at the end of the text.
func (b *builder) placeImage(img widget.WidgetImage, opts []text.ImageOption, cs charStyle, in inline) {
	b.ensureLine()
	b.p.images = append(b.p.images, planImage{at: b.cur, img: img, opts: opts})
	start := b.cur
	b.p.text.WriteRune('\uFFFC')
	b.cur.char++
	b.addRange(b.charTag(cs), start, b.cur)
	if in.link != "" {
		b.addRange(in.link, start, b.cur)
	}
	b.atLineStart = false
}

// finish defines the line-level tags and their ranges.
func (b *builder) finish() {
	for _, li := range b.lines {
		start, end := pos{line: li.line}, pos{line: li.line + 1}
		key := fmt.Sprint(li.para)
		name, ok := b.paraTags[key]
		if !ok {
			name = b.newName("p")
			b.paraTags[key] = name
			opts := []text.TagOption{
				text.TagLMargin1(li.para.lm1), text.TagLMargin2(li.para.lm2), text.TagRMargin(li.para.rm),
			}
			switch li.para.align {
			case "center":
				opts = append(opts, text.TagJustify(option.JustifyCenter))
			case "right":
				opts = append(opts, text.TagJustify(option.JustifyRight))
			}
			if li.para.nowrap {
				opts = append(opts, text.TagWrap(text.WrapNone))
			}
			switch li.para.clear {
			case "left":
				opts = append(opts, text.TagClear(text.ClearLeft))
			case "right":
				opts = append(opts, text.TagClear(text.ClearRight))
			case "both":
				opts = append(opts, text.TagClear(text.ClearBoth))
			}
			if len(li.tabs) > 0 {
				opts = append(opts, text.TagTabs(li.tabs...))
			}
			b.define(name, kindPara, opts...)
		}
		b.addRange(name, start, end)
		if li.sp1 > 0 {
			b.addRange(b.spacingTag(1, li.sp1), start, end)
		}
		if li.sp3 > 0 {
			b.addRange(b.spacingTag(3, li.sp3), start, end)
		}
		if li.marker != "" {
			m := b.newName("m")
			b.define(m, kindLine, text.TagMarker(li.marker))
			b.addRange(m, start, end)
		}
		if li.rule != nil {
			r := b.newName("r")
			b.define(r, kindLine, text.TagRule(li.rule.thickness, li.rule.color))
			b.addRange(r, start, end)
		}
	}
}

func (b *builder) spacingTag(which, v int) string {
	key := fmt.Sprintf("%d:%d", which, v)
	if name, ok := b.spTags[key]; ok {
		return name
	}
	name := b.newName("s")
	b.spTags[key] = name
	if which == 1 {
		b.define(name, kindSpacing, text.TagSpacing1(v))
	} else {
		b.define(name, kindSpacing, text.TagSpacing3(v))
	}
	return name
}
