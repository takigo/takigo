package richtext

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/text"
)

// atom is a piece of a table cell's content: a word or part of one in one
// style, a space, or a forced line break.
type atom struct {
	text  string
	cs    charStyle
	in    inline
	width int
	space bool
	brk   bool
	img   widget.WidgetImage // an image, laid out like a word
	opts  []text.ImageOption
}

// cell is one table cell's content and its box.
type cell struct {
	atoms       []atom
	style       style
	padL, padR  int
	padT, padB  int
	borderT     int
	borderB     int
	borderColor string
	minW, maxW  int // content widths: the longest word, and unwrapped
	fixed       int // the CSS width, or 0
	colspan     int
	rowspan     int
	lines       [][]atom
}

// placed is a cell at its place in the table's grid.
type placed struct {
	*cell
	row, col int
	next     int // the first of its lines not shown yet
}

// table lays a table out on tab stops. Columns get widths as in CSS's
// automatic table layout: at least their longest word, at most their
// unwrapped content, sharing the width available in proportion. Each
// cell's content is word-wrapped to its column, and a row is as many lines
// as its tallest cell, all in one box that draws the cells' background
// and the borders above and below. Cells hold inline content; blocks and
// <br> inside them break lines.
func (b *builder) table(n *node, s *style) {
	b.closeLine()
	b.pendingMargin = max(b.pendingMargin, s.margin[top])
	savedLeft, savedRight := b.left, b.right
	b.left += s.margin[left] + s.border[left] + s.padding[left]
	b.right += s.margin[right] + s.border[right] + s.padding[right]

	var rows [][]cell
	var rowStyles []style
	var caption *node
	var collect func(parent *node, ps *style)
	collect = func(parent *node, ps *style) {
		for _, c := range parent.children {
			if c.kind != elementNode {
				continue
			}
			switch c.tag {
			case "caption":
				caption = c
			case "thead", "tbody", "tfoot":
				cs := b.st.compute(c, ps)
				collect(c, &cs)
			case "tr":
				rs := b.st.compute(c, ps)
				var row []cell
				for _, td := range c.children {
					if td.kind == elementNode && (td.tag == "td" || td.tag == "th") {
						row = append(row, b.cell(td, &rs))
					}
				}
				rows = append(rows, row)
				rowStyles = append(rowStyles, rs)
			}
		}
	}
	collect(n, s)
	if caption != nil {
		b.block(caption, s)
	}
	grid, ncols := placeCells(rows)
	avail := 0
	if b.cfg.width > 0 {
		avail = b.cfg.width - px(b.left) - px(b.right)
	}
	widths := columnWidths(grid, ncols, avail)
	tableW := 0
	for _, w := range widths {
		tableW += w
	}
	for _, p := range grid {
		w := -p.padL - p.padR
		for c := p.col; c < p.col+p.colspan; c++ {
			w += widths[c]
		}
		p.lines = wrap(p.atoms, w)
	}
	for ri := range rows {
		b.row(grid, ri, &rowStyles[ri], widths, tableW)
	}
	b.left, b.right = savedLeft, savedRight
	b.pendingMargin = max(b.pendingMargin, s.margin[bottom])
}

// placeCells puts the cells on the table's grid as the HTML table model
// does: each row's cells take the first columns not covered by a rowspan
// from above.
func placeCells(rows [][]cell) ([]*placed, int) {
	var grid []*placed
	taken := map[[2]int]bool{}
	ncols := 0
	for r, row := range rows {
		c := 0
		for i := range row {
			cl := &row[i]
			for taken[[2]int{r, c}] {
				c++
			}
			cl.rowspan = min(cl.rowspan, len(rows)-r)
			for dr := range cl.rowspan {
				for dc := range cl.colspan {
					taken[[2]int{r + dr, c + dc}] = true
				}
			}
			grid = append(grid, &placed{cell: cl, row: r, col: c})
			c += cl.colspan
			ncols = max(ncols, c)
		}
	}
	return grid, ncols
}

// columnWidths gives each column its maximum width when the table fits,
// its minimum when not even that fits, and otherwise the minimum plus a
// share of the rest in proportion to how much more the column wants. A
// cell spanning columns spreads what they lack evenly over them.
func columnWidths(grid []*placed, ncols, avail int) []int {
	minW, maxW := make([]int, ncols), make([]int, ncols)
	bounds := func(c *cell) (int, int) {
		pad := c.padL + c.padR
		lo, hi := c.minW+pad, c.maxW+pad
		if c.fixed > 0 {
			lo = max(c.fixed+pad, lo)
			hi = lo
		}
		return lo, max(hi, lo)
	}
	for _, p := range grid {
		if p.colspan == 1 {
			lo, hi := bounds(p.cell)
			minW[p.col], maxW[p.col] = max(minW[p.col], lo), max(maxW[p.col], hi)
		}
	}
	for _, p := range grid {
		if p.colspan == 1 {
			continue
		}
		lo, hi := bounds(p.cell)
		spread := func(ws []int, need int) {
			have := 0
			for c := p.col; c < p.col+p.colspan; c++ {
				have += ws[c]
			}
			for k, c := 0, p.col; c < p.col+p.colspan; k, c = k+1, c+1 {
				if extra := need - have; extra > 0 {
					ws[c] += extra / p.colspan
					if k < extra%p.colspan {
						ws[c]++
					}
				}
			}
		}
		spread(minW, lo)
		spread(maxW, hi)
	}
	for c := range ncols {
		maxW[c] = max(maxW[c], minW[c])
	}
	sumMin, sumMax := 0, 0
	for i := range ncols {
		sumMin += minW[i]
		sumMax += maxW[i]
	}
	switch {
	case avail <= 0 || sumMax <= avail:
		return maxW
	case sumMin >= avail:
		return minW
	}
	widths := make([]int, ncols)
	extra, want := avail-sumMin, sumMax-sumMin
	given := 0
	for i := range ncols {
		widths[i] = minW[i] + extra*(maxW[i]-minW[i])/want
		given += widths[i]
	}
	// Rounding leftovers go to the last column that can take them.
	for i := ncols - 1; i >= 0 && given < avail; i-- {
		if d := min(avail-given, maxW[i]-widths[i]); d > 0 {
			widths[i] += d
			given += d
		}
	}
	return widths
}

// cell collects a table cell's content as atoms and measures it.
func (b *builder) cell(td *node, rowStyle *style) cell {
	cs := b.st.compute(td, rowStyle)
	c := cell{
		style: cs, padL: px(cs.padding[left]), padR: px(cs.padding[right]),
		padT: px(cs.padding[top]), padB: px(cs.padding[bottom]),
		borderT: px(cs.border[top]), borderB: px(cs.border[bottom]),
		borderColor: cmpOr(cs.borderColor[bottom], cs.borderColor[top]),
	}
	c.colspan, c.rowspan = 1, 1
	if v, err := strconv.Atoi(td.attrs["colspan"]); err == nil && v > 1 {
		c.colspan = min(v, 1000)
	}
	if v, err := strconv.Atoi(td.attrs["rowspan"]); err == nil && v > 1 {
		c.rowspan = min(v, 65534)
	}
	if cs.width.set && cs.width.percent == 0 {
		c.fixed = px(cs.width.px)
	} else if v, err := strconv.Atoi(td.attrs["width"]); err == nil {
		c.fixed = v
	}
	content := func() bool {
		k := len(c.atoms)
		return k > 0 && !c.atoms[k-1].brk
	}
	addSpace := func(cs charStyle, in inline) {
		if k := len(c.atoms); content() && !c.atoms[k-1].space {
			c.atoms = append(c.atoms, atom{text: " ", cs: cs, in: in, space: true})
		}
	}
	addBreak := func() {
		if k := len(c.atoms); k > 0 && c.atoms[k-1].space {
			c.atoms = c.atoms[:k-1]
		}
		c.atoms = append(c.atoms, atom{brk: true})
	}
	var walk func(n *node, s *style, in inline)
	walk = func(n *node, s *style, in inline) {
		if n.kind == textNode {
			st := b.charStyleOf(s, in)
			var word strings.Builder
			flush := func() {
				if word.Len() > 0 {
					c.atoms = append(c.atoms, atom{text: word.String(), cs: st, in: in})
					word.Reset()
				}
			}
			for _, r := range n.text {
				if unicode.IsSpace(r) && r != '\u00a0' {
					flush()
					addSpace(st, in)
					continue
				}
				word.WriteRune(r)
			}
			flush()
			return
		}
		es := b.st.compute(n, s)
		if es.display == "none" {
			return
		}
		switch n.tag {
		case "br":
			addBreak()
			return
		case "img":
			if img := b.cfg.loadImage(n.attrs["src"]); img != nil {
				img = fitImage(img, &es, 0, false)
				c.atoms = append(c.atoms, atom{img: img, opts: imageOptions(&es, false),
					width: img.Width() + 2*px(max(es.margin[left], es.margin[right])),
					cs:    b.charStyleOf(&es, in), in: in})
			} else if alt := n.attrs["alt"]; alt != "" {
				c.atoms = append(c.atoms, atom{text: "[" + alt + "]", cs: b.charStyleOf(&es, in), in: in})
			}
			return
		}
		child := in
		if es.background != "" && es.display == "inline" {
			child.bg = es.background
		}
		if href, ok := n.attr("href"); ok && n.tag == "a" {
			child.link = b.link(n, href)
		}
		block := es.display != "inline"
		if block && content() {
			addBreak()
		}
		for _, ch := range n.children {
			walk(ch, &es, child)
		}
		if block && content() {
			addBreak()
		}
	}
	for _, ch := range td.children {
		walk(ch, &cs, inline{})
	}
	for len(c.atoms) > 0 && (c.atoms[len(c.atoms)-1].brk || c.atoms[len(c.atoms)-1].space) {
		c.atoms = c.atoms[:len(c.atoms)-1]
	}
	word, line := 0, 0
	for i := range c.atoms {
		a := &c.atoms[i]
		if !a.brk && a.img == nil {
			a.width = b.cfg.measure(a.cs, a.text)
		}
		switch {
		case a.brk:
			word, line = 0, 0
		case a.space:
			word = 0
			line += a.width
		default:
			word += a.width
			line += a.width
		}
		c.minW, c.maxW = max(c.minW, word), max(c.maxW, line)
	}
	return c
}

// wrap breaks atoms into lines of at most w pixels at spaces; a word wider
// than w gets a line of its own.
func wrap(atoms []atom, w int) [][]atom {
	var lines [][]atom
	var cur []atom
	curW := 0
	for i := 0; i < len(atoms); {
		a := atoms[i]
		if a.brk {
			lines = append(lines, cur)
			cur, curW = nil, 0
			i++
			continue
		}
		if a.space {
			i++
			continue
		}
		// The word: atoms up to the next space or break.
		j, ww := i, 0
		for j < len(atoms) && !atoms[j].space && !atoms[j].brk {
			ww += atoms[j].width
			j++
		}
		var sp []atom
		if i > 0 && atoms[i-1].space && len(cur) > 0 {
			sp = atoms[i-1 : i]
		}
		spW := 0
		if sp != nil {
			spW = sp[0].width
		}
		if len(cur) > 0 && curW+spW+ww > w {
			lines = append(lines, cur)
			cur, curW, sp, spW = nil, 0, nil, 0
		}
		cur = append(cur, sp...)
		cur = append(cur, atoms[i:j]...)
		curW += spW + ww
		i = j
	}
	return append(lines, cur)
}

// row adds row r of the table: the cells that start in it and the lines
// of the rowspan cells from above that are not shown yet, on the row's tab
// stops. The row is as many lines as its own cells need, and its last
// spanned row as many as a rowspan cell still has.
func (b *builder) row(grid []*placed, r int, rs *style, widths []int, tableW int) {
	var cells []*placed
	for _, p := range grid {
		if p.row <= r && r < p.row+p.rowspan {
			cells = append(cells, p)
		}
	}
	if len(cells) == 0 {
		return
	}
	colLeft := make([]int, len(widths)+1)
	colLeft[0] = px(b.left)
	for i, w := range widths {
		colLeft[i+1] = colLeft[i] + w
	}
	padL := cells[0].padL
	stops := make([]text.TabStop, 0, len(widths))
	for c := 1; c < len(widths); c++ {
		stop := text.TabStop{Pos: colLeft[c] + padL}
		for _, p := range cells {
			if p.col != c {
				continue
			}
			end := colLeft[c+p.colspan]
			switch p.style.align {
			case "right":
				stop = text.TabStop{Pos: end - p.padR, Align: text.TabRight}
			case "center":
				stop = text.TabStop{Pos: (colLeft[c] + end) / 2, Align: text.TabCenter}
			default:
				stop = text.TabStop{Pos: colLeft[c] + p.padL}
			}
		}
		stops = append(stops, stop)
	}
	padT, padB, bT, bB, bcol, bg := 0, 0, 0, 0, "", rs.background
	nlines := 1
	for _, p := range cells {
		if p.row == r {
			padT, bT = max(padT, p.padT), max(bT, p.borderT)
		}
		last := r == p.row+p.rowspan-1
		if last {
			padB, bB = max(padB, p.padB), max(bB, p.borderB)
			nlines = max(nlines, len(p.lines)-p.next)
		}
		bcol = cmpOr(bcol, p.borderColor)
		if bg == "" && p.row == r {
			bg = p.style.background
		}
	}
	if r > 0 {
		bT = 0
	}
	rm := px(b.right)
	if b.cfg.width > 0 && colLeft[0]+tableW < b.cfg.width-rm {
		rm = b.cfg.width - colLeft[0] - tableW
	}
	bx := &box{
		name: b.newName("b"), lmargin: colLeft[0], rmargin: rm, bg: bg, borderColor: bcol,
		border:  [4]int{0, bT, 0, bB},
		padding: [4]int{padL, padT, 0, padB},
		open:    true,
	}
	b.openBox(bx)
	savedNowrap, savedBS := b.nowrap, b.blockStyle
	b.nowrap = true
	b.blockStyle = b.charStyleOf(&cells[0].style, inline{})
	for range nlines {
		b.openLine()
		li := &b.lines[len(b.lines)-1]
		li.tabs = stops
		li.para.tabs = fmt.Sprint(stops)
		tabs := 0
		for _, p := range cells {
			if p.next < len(p.lines) && len(p.lines[p.next]) > 0 {
				for ; tabs < p.col; tabs++ {
					b.emit("\t", b.blockStyle, inline{})
				}
			}
			if p.next < len(p.lines) {
				for _, a := range p.lines[p.next] {
					if a.img != nil {
						b.placeImage(a.img, a.opts, a.cs, a.in)
					} else {
						b.emit(a.text, a.cs, a.in)
					}
				}
				p.next++
			}
		}
		b.closeLine()
	}
	b.closeBox(bx)
	b.nowrap, b.blockStyle = savedNowrap, savedBS
}
