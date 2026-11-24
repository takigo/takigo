package text

// Floating images, a takigo extension with no Tk counterpart: an image
// with ImageFloatOpt leaves the text flow and sits against the left or
// right margin, and the display lines beside it are narrowed until its
// bottom, across logical lines, like a CSS float.

import (
	"slices"

	"github.com/takigo/takigo/color"
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/platform"
)

// ImageFloat is an embedded image's float: whether it stays in the line
// or moves to a margin with the text flowing round it.
type ImageFloat int

const (
	FloatNone  ImageFloat = iota // in the line, like a character (Tk)
	FloatLeft                    // against the left margin
	FloatRight                   // against the right margin
)

// ImageFloatOpt floats the image to a margin. The float's top is the top
// of the display line its index starts, or of the next display line when
// the index is inside one; its -padx separates it from the text only, and
// its -pady is added below it. Not in Tk.
func ImageFloatOpt(f ImageFloat) ImageOption {
	return func(ei *embImage) { ei.float = f }
}

// flowWidth is the width the image takes in its line: none when floated.
func (ei *embImage) flowWidth() int {
	if ei.float != FloatNone {
		return 0
	}
	return ei.width()
}

// floatCarry is a float still beside the text at the top of a logical
// line: bottom is how far below the line's top its bottom is.
type floatCarry struct {
	ei     *embImage
	side   ImageFloat
	x, w   int // left edge and width in the text area, padding included
	h      int
	bottom int
}

// placedFloat is a float whose index is in the logical line: y is how far
// below the line's top its top is.
type placedFloat struct {
	ei   *embImage
	x, y int
}

// floatState tracks the floats beside a logical line while it is wrapped.
type floatState struct {
	active     []floatCarry
	placed     []placedFloat
	availWidth int
}

// margins returns the left and right margins the floats beside y (from
// the line's top) leave: the inner edges of the floats still below y.
func (fs *floatState) margins(y, lm, rm int) (int, int) {
	for _, f := range fs.active {
		if f.bottom <= y {
			continue
		}
		if f.side == FloatLeft {
			lm = max(lm, f.x+f.w)
		} else {
			rm = max(rm, fs.availWidth-f.x)
		}
	}
	return lm, rm
}

// place puts a float at y next to the floats already there.
func (fs *floatState) place(ei *embImage, y, lm, rm int) {
	lm, rm = fs.margins(y, lm, rm)
	w, h := ei.img.Width()+ei.padX, ei.img.Height()+ei.padY
	x := lm
	if ei.float == FloatRight {
		x = fs.availWidth - rm - w
	}
	fs.active = append(fs.active, floatCarry{ei: ei, side: ei.float, x: x, w: w, h: h, bottom: y + h})
	fs.placed = append(fs.placed, placedFloat{ei: ei, x: x, y: y})
}

// carryOut returns the floats that reach below a line pixels tall, with
// their bottoms from the next line's top.
func (fs *floatState) carryOut(pixels int) []floatCarry {
	var out []floatCarry
	for _, f := range fs.active {
		if f.bottom > pixels {
			f.bottom -= pixels
			out = append(out, f)
		}
	}
	return out
}

// floatAt returns the floating image at pos, or nil.
func (t *TextWidget) floatAt(lineIdx, char int) *embImage {
	if ei := t.doc.imageAt(Index{Line: lineIdx, Char: char}); ei != nil && ei.float != FloatNone {
		return ei
	}
	return nil
}

// wrapLineFloats wraps a logical line beside the floats of fs: each
// display line gets the margins the floats leave at its top, so the lines'
// heights are worked out one at a time. Floats whose index starts a
// display line are placed at its top, others below it.
func (t *TextWidget) wrapLineFloats(lineIdx, availWidth int, p lineProps, mode WrapMode, fs *floatState) []displayLine {
	line := t.doc.Lines[lineIdx-1]
	lineLen := len(line.Text)
	h := t.lineHeightFor(lineIdx)
	var dls []displayLine
	y := p.sp1
	placeIn := func(from, to int, y int, lm, rm int) {
		for c := from; c < to; c++ {
			if ei := t.floatAt(lineIdx, c); ei != nil {
				fs.place(ei, y, lm, rm)
			}
		}
	}
	for start := 0; ; {
		lm := p.lm2
		if start == 0 {
			lm = p.lm1
		}
		// Floats at the start of the display line go at its top.
		first := start
		for start < lineLen && t.floatAt(lineIdx, start) != nil {
			placeIn(start, start+1, y, lm, p.rm)
			start++
		}
		lmF, rmF := fs.margins(y, lm, p.rm)
		end := lineLen
		if start < lineLen && mode != WrapNone && availWidth > 0 {
			end = t.nextBreak(lineIdx, start, max(availWidth-lmF-rmF, 1), mode)
		}
		dl := displayLine{logicalLine: lineIdx, startChar: first, endChar: end, height: h,
			leftMargin: lmF, rightMargin: rmF}
		dl = t.setMetrics(lineIdx, []displayLine{dl})[0]
		dls = append(dls, dl)
		next := y + dl.height
		placeIn(start, end, next, lm, p.rm) // floats inside the line go below it
		if end >= lineLen {
			break
		}
		y = next + p.sp2
		start = end
	}
	return dls
}

// floatsVisible returns the floats to draw over the visible lines with
// their window positions: those placed in visible lines and those carried
// into the first one from above.
func (t *TextWidget) floatsVisible(dlines []displayLine) []placedFloat {
	var out []placedFloat
	seen := map[int]bool{}
	for _, dl := range dlines {
		l := dl.logicalLine
		if seen[l] {
			continue
		}
		seen[l] = true
		ll := t.layout.line(l)
		k := slices.IndexFunc(ll.dls, func(d displayLine) bool { return d.startChar == dl.startChar })
		top := t.insetY + dl.y - dl.spacingAbove
		for i := range max(k, 0) {
			top -= ll.hs[i]
		}
		if len(out) == 0 {
			for _, f := range ll.carryIn {
				out = append(out, placedFloat{ei: f.ei, x: f.x, y: top + f.bottom - f.h})
			}
		}
		for _, f := range ll.floats {
			out = append(out, placedFloat{ei: f.ei, x: f.x, y: top + f.y})
		}
	}
	return out
}

// drawFloats draws the visible floating images.
func (t *TextWidget) drawFloats(d platform.DisplayServer, gc platform.GCID, drawable platform.DrawableID, dlines []displayLine) {
	if !t.layout.key.floats {
		return
	}
	bg := uint64(0xFFFFFF)
	if t.Background != nil {
		bg = t.Background.Pixel
	}
	for _, f := range t.floatsVisible(dlines) {
		img := f.ei.img
		x := f.x
		if f.ei.float == FloatRight {
			x += f.ei.padX
		}
		img.Draw(d, drawable, gc, t.Win.Depth, 0, 0, img.Width(), img.Height(),
			t.insetX-t.xOffset+x, f.y, bg)
	}
}

// Clear is a tag's clear: the lines starting with the tag start below the
// floats on those sides.
type Clear int

const (
	ClearNone Clear = iota
	ClearLeft
	ClearRight
	ClearBoth
)

// TagClear makes the logical lines that start with the tag begin below the
// floats on the given sides, like CSS clear. Not in Tk.
func TagClear(c Clear) TagOption {
	return func(cache *color.Cache, reg *font.Registry, tag *Tag) {
		tag.Clear = c
	}
}

// clearance returns how much lower than y a line has to start to be below
// the floats it clears.
func (fs *floatState) clearance(y int, c Clear) int {
	extra := 0
	for _, f := range fs.active {
		if c == ClearBoth || c == ClearLeft && f.side == FloatLeft || c == ClearRight && f.side == FloatRight {
			extra = max(extra, f.bottom-y)
		}
	}
	return extra
}
