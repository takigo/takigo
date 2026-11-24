package text

// Block boxes, rules, list markers and margin colours. Margin colours port
// DisplayDLine's -lmargincolor/-rmargincolor fills (tkTextDisp.c); the rest
// are takigo extensions for HTML-like documents, with no Tk counterpart.

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/takigo/takigo/color"
	"github.com/takigo/takigo/internal/nanosvg"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/widget"
)

// blockBox is the visible part of one block: the run of display lines of
// consecutive logical lines whose first characters carry the same block tag.
type blockBox struct {
	tag               *Tag
	y0, y1            int
	hasTop, hasBottom bool         // the box's own top and bottom edges are visible
	depth             int          // boxes around it on its first line
	under             *color.Color // the background of the box around it
}

// drawBlocks paints the block boxes behind the visible lines, the outer
// ones first.
func (t *TextWidget) drawBlocks(d platform.DisplayServer, gc platform.GCID, drawable platform.DrawableID, dlines []displayLine) {
	var done []*blockBox
	open := map[*Tag]*blockBox{}
	closeBox := func(tag *Tag) {
		done = append(done, open[tag])
		delete(open, tag)
	}
	for _, dl := range dlines {
		if dl.height == 0 {
			continue // an elided line
		}
		p := &t.layout.line(dl.logicalLine).props
		for tag := range open {
			if !slices.ContainsFunc(p.blocks, func(lb lineBlock) bool { return lb.tag == tag }) {
				closeBox(tag)
			}
		}
		first := dl.startChar == 0
		last := dl.endChar >= len(t.doc.Lines[dl.logicalLine-1].Text)
		lineTop := t.insetY + dl.y - dl.spacingAbove
		lineBottom := t.insetY + dl.y + dl.height + dl.spacingBelow
		for i, lb := range p.blocks {
			opens := first && lb.start
			top := lineTop
			if opens {
				top += lb.top
			}
			closes := last && lb.end
			bottom := lineBottom
			if closes {
				bottom -= lb.bottom
			}
			if opens && open[lb.tag] != nil {
				closeBox(lb.tag)
			}
			bx := open[lb.tag]
			if bx == nil {
				bx = &blockBox{tag: lb.tag, y0: top, hasTop: opens, depth: i}
				for k := i - 1; k >= 0; k-- {
					if c := p.blocks[k].tag.BlockBackground; c != nil {
						bx.under = c
						break
					}
				}
				open[lb.tag] = bx
			}
			bx.y1 = bottom
			if closes {
				bx.hasBottom = true
				closeBox(lb.tag)
			}
		}
	}
	for tag := range open {
		closeBox(tag)
	}
	slices.SortStableFunc(done, func(a, b *blockBox) int { return cmp.Compare(a.depth, b.depth) })
	for _, bx := range done {
		t.drawBlockBox(d, gc, drawable, bx)
	}
}

// drawBlockBox fills a block's background and draws its border; the
// corners are rounded by BlockRadius outside a Classic App.
func (t *TextWidget) drawBlockBox(d platform.DisplayServer, gc platform.GCID, drawable platform.DrawableID, b *blockBox) {
	tag := b.tag
	x0 := t.insetX + tag.LMargin1 - t.xOffset
	x1 := t.Win.Width - t.insetX - tag.RMargin
	y0, y1 := b.y0, b.y1
	if x1 <= x0 || y1 <= y0 {
		return
	}
	bw := tag.BlockBorder
	if !b.hasTop {
		bw[1] = 0
	}
	if !b.hasBottom {
		bw[3] = 0
	}
	borderCol := tag.BlockBorderColor
	if borderCol == nil {
		borderCol = t.Foreground
	}
	r := tag.BlockRadius
	uniform := tag.BlockBorder[0] == tag.BlockBorder[1] && tag.BlockBorder[1] == tag.BlockBorder[2] &&
		tag.BlockBorder[2] == tag.BlockBorder[3]
	r = min(r, (x1-x0)/2, (y1-y0)/2)
	if r <= 0 || !uniform || widget.Classic(t.App) {
		r = 0
	}
	rt, rb := 0, 0 // the corner height at the top and bottom
	if b.hasTop {
		rt = r
	}
	if b.hasBottom {
		rb = r
	}
	fill := func(c *color.Color, x, y, w, h int) {
		if c == nil || w <= 0 || h <= 0 {
			return
		}
		d.SetForeground(gc, c.Pixel)
		d.FillRectangle(drawable, gc, x, y, uint(w), uint(h))
	}
	bg := tag.BlockBackground
	fill(bg, x0+r, y0, x1-x0-2*r, y1-y0)
	fill(bg, x0, y0+rt, r, y1-y0-rt-rb)
	fill(bg, x1-r, y0+rt, r, y1-y0-rt-rb)
	if borderCol != nil {
		fill(borderCol, x0, y0+rt, bw[0], y1-y0-rt-rb)
		fill(borderCol, x1-bw[2], y0+rt, bw[2], y1-y0-rt-rb)
		if b.hasTop {
			fill(borderCol, x0+rt, y0, x1-x0-2*rt, bw[1])
		}
		if b.hasBottom {
			fill(borderCol, x0+rb, y1-bw[3], x1-x0-2*rb, bw[3])
		}
	}
	if r == 0 {
		return
	}
	// The corners are the quarters of an anti-aliased rounded square,
	// blended over the box around this one or the widget background.
	under := uint64(0xFFFFFF)
	switch {
	case b.under != nil:
		under = b.under.Pixel
	case t.Background != nil:
		under = t.Background.Pixel
	}
	px, w, h := roundedCorner(r, tag.BlockBorder[0], bg, borderCol, under)
	if w != 2*r || h != 2*r {
		return
	}
	put := func(sx, sy, dx, dy int) {
		d.PutImageRGBA(drawable, gc, t.Win.Depth, px, w*4, w, h, sx, sy, dx, dy, r, r, under)
	}
	if b.hasTop {
		put(0, 0, x0, y0)
		put(r, 0, x1-r, y0)
	}
	if b.hasBottom {
		put(0, r, x0, y1-r)
		put(r, r, x1-r, y1-r)
	}
}

// roundedCorner rasterizes a 2r × 2r rounded square with border width bw
// and returns it blended over under, opaque RGBA.
func roundedCorner(r, bw int, fill, border *color.Color, under uint64) ([]uint8, int, int) {
	hex := func(c *color.Color) string {
		if c == nil {
			return "none"
		}
		return fmt.Sprintf("#%02x%02x%02x", c.Red>>8, c.Green>>8, c.Blue>>8)
	}
	size := 2 * r
	inset := float64(bw) / 2
	stroke := ""
	if bw > 0 && border != nil {
		stroke = fmt.Sprintf(` stroke="%s" stroke-width="%d"`, hex(border), bw)
	}
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d">`+
		`<rect x="%g" y="%g" width="%g" height="%g" rx="%g" fill="%s"%s/></svg>`,
		size, size, inset, inset, float64(size)-2*inset, float64(size)-2*inset,
		max(float64(r)-inset, 0), hex(fill), stroke)
	img, err := nanosvg.Parse(svg)
	if err != nil {
		return nil, 0, 0
	}
	w, h := nanosvg.Size(img, 1)
	return nanosvg.BlendOver(nanosvg.Rasterize(img, 1, w, h), under), w, h
}

// drawLineDecorations draws what belongs to a display line rather than to
// its characters: the margin colours, a rule and a list marker.
func (t *TextWidget) drawLineDecorations(d platform.DisplayServer, gc platform.GCID, drawable platform.DrawableID,
	dl displayLine, p *lineProps) {
	top := t.insetY + dl.y - dl.spacingAbove
	full := dl.spacingAbove + dl.height + dl.spacingBelow
	if c := p.lmColor; c != nil && dl.leftMargin > 0 {
		d.SetForeground(gc, c.Pixel)
		d.FillRectangle(drawable, gc, t.insetX, top, uint(dl.leftMargin), uint(full))
	}
	if c := p.rmColor; c != nil && dl.rightMargin > 0 {
		d.SetForeground(gc, c.Pixel)
		d.FillRectangle(drawable, gc, t.Win.Width-t.insetX-dl.rightMargin, top, uint(dl.rightMargin), uint(full))
	}
	if dl.startChar != 0 {
		return
	}
	if rule := p.rule; rule != nil {
		c := rule.RuleColor
		if c == nil {
			c = t.Foreground
		}
		x0 := t.insetX + dl.leftMargin - t.xOffset
		x1 := t.Win.Width - t.insetX - dl.rightMargin
		if c != nil && x1 > x0 {
			d.SetForeground(gc, c.Pixel)
			d.FillRectangle(drawable, gc, x0, t.insetY+dl.y+(dl.height-rule.Rule)/2, uint(x1-x0), uint(rule.Rule))
		}
	}
	if p.marker != "" {
		f, fg := t.Font, t.Foreground
		for _, tag := range t.doc.TagsAt(Index{Line: dl.logicalLine}) {
			if tag.Font != nil {
				f = tag.Font
			}
			if tag.Foreground != nil {
				fg = tag.Foreground
			}
		}
		df, ok := f.(platform.DrawableFont)
		if !ok || fg == nil {
			return
		}
		x := t.insetX + dl.leftMargin - t.xOffset - f.MeasureString(" ") - f.MeasureString(p.marker)
		df.DrawString(drawable, x, t.insetY+dl.y+dl.ascent, p.marker, fg.Pixel, fg.Red, fg.Green, fg.Blue)
	}
}
