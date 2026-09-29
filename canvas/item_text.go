package canvas

import (
	"math"

	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
)

// TextItem implements a positioned text canvas item.
type TextItem struct {
	ItemBase
	x, y       float64
	text       string
	font       font.Font
	color      *color.ColorRef
	anchor     option.Anchor
	justify    option.Justify
	wrapLength int     // 0 = no wrapping
	angle      float64 // rotation in degrees, clockwise on screen (Tk convention); 0 = normal

	// Cursor state (for focus/edit support).
	cursorPos int  // byte position in text (0 = before first char)
	hasFocus  bool // true when this item has canvas keyboard focus

	// lay is Tk's textPtr->textLayout: computed when the text, font, -width
	// or -justify change (updateBBox), drawn as is by Display.
	lay textLayout
}

// textLayout ports Tk_ComputeTextLayout for the item: lines broken at
// newlines and -width, each split at tabs into segments with their x
// offsets, the layout as wide as its widest line, and each line's x offset
// for -justify.
type textLayout struct {
	lines []string
	segs  [][]font.Segment
	xs    []int
	w, h  int
}

// cursorLine returns the display line holding the insertion cursor and
// the cursor's byte offset in it. It walks the text as font.WrapLines
// consumed it: a newline ends a paragraph, and the blanks at a wrap are
// dropped (a cursor among them sits at the end of the line before).
func (t *TextItem) cursorLine() (line, col int) {
	lines := t.lay.lines
	pos := 0
	for i, l := range lines {
		if i > 0 {
			j := pos
			for j < len(t.text) && (t.text[j] == ' ' || t.text[j] == '\t') {
				j++
			}
			if pos < len(t.text) && t.text[pos] == '\n' {
				pos++
			} else if j < len(t.text) && t.text[j] == '\n' {
				pos = j + 1
			} else {
				pos = j
			}
		}
		if pos+len(l) > len(t.text) || t.text[pos:pos+len(l)] != l {
			return 0, 0 // not the layout of this text; draw at the start
		}
		if next := pos + len(l); t.cursorPos <= next || i == len(lines)-1 {
			if t.cursorPos < pos {
				// Among the blanks dropped at the wrap before this line.
				return max(i-1, 0), len(lines[max(i-1, 0)])
			}
			return i, min(t.cursorPos-pos, len(l))
		}
		pos += len(l)
	}
	return 0, 0
}

// InsertText inserts s at the given byte position and advances the cursor.
func (t *TextItem) InsertText(pos int, s string) {
	if pos < 0 {
		pos = 0
	}
	if pos > len(t.text) {
		pos = len(t.text)
	}
	t.text = t.text[:pos] + s + t.text[pos:]
	t.cursorPos = pos + len(s)
	t.updateBBox()
}

// DeleteChars removes bytes from first to last (exclusive).
func (t *TextItem) DeleteChars(first, last int) {
	n := len(t.text)
	if first < 0 {
		first = 0
	}
	if last > n {
		last = n
	}
	if first >= last {
		return
	}
	t.text = t.text[:first] + t.text[last:]
	if t.cursorPos > first {
		t.cursorPos -= last - first
		if t.cursorPos < first {
			t.cursorPos = first
		}
	}
	t.updateBBox()
}

// SetCursorPos sets the cursor byte position (clamped to text length).
func (t *TextItem) SetCursorPos(pos int) {
	if pos < 0 {
		pos = 0
	}
	if pos > len(t.text) {
		pos = len(t.text)
	}
	t.cursorPos = pos
}

func newTextItem(x, y float64, c *Canvas) *TextItem {
	item := &TextItem{
		x:      x,
		y:      y,
		anchor: option.AnchorCenter,
	}
	item.color = &color.ColorRef{Pixel: 0x000000}
	item.ItemBase.canvas = c

	// Use default font.
	if f, err := c.FontRegistry().Get(font.TkDefaultFont); err == nil {
		item.font = f
	}

	item.updateBBox()
	return item
}

func (t *TextItem) base() *ItemBase { return &t.ItemBase }
func (t *TextItem) Type() string    { return "text" }

func (t *TextItem) BBox() (x1, y1, x2, y2 int) {
	return t.X1, t.Y1, t.X2, t.Y2
}

func (t *TextItem) Coords() []float64 {
	return []float64{t.x, t.y}
}

func (t *TextItem) SetCoords(coords []float64) error {
	if len(coords) >= 2 {
		t.x = coords[0]
		t.y = coords[1]
		t.updateBBox()
	}
	return nil
}

func (t *TextItem) Configure(opts []ItemOption) error {
	c := t.canvas
	for _, opt := range opts {
		if err := opt(c, t); err != nil {
			return err
		}
	}
	t.updateBBox()
	return nil
}

func (t *TextItem) relayout() {
	t.lay = textLayout{}
	if t.font == nil {
		return
	}
	// Like Tk_ComputeTextLayout, an empty text is one empty line high, so
	// the insertion cursor of an empty focused item has a height.
	lines := font.WrapLines(t.font, t.text, t.wrapLength)
	segs := make([][]font.Segment, len(lines))
	widths := make([]int, len(lines))
	w := 0
	for i, l := range lines {
		segs[i] = font.Segments(t.font, l)
		last := segs[i][len(segs[i])-1]
		widths[i] = last.X + t.font.MeasureString(last.Text)
		w = max(w, widths[i])
	}
	xs := make([]int, len(lines))
	for i := range lines {
		switch t.justify {
		case option.JustifyCenter:
			xs[i] = (w - widths[i]) / 2
		case option.JustifyRight:
			xs[i] = w - widths[i]
		}
	}
	t.lay = textLayout{lines: lines, segs: segs, xs: xs, w: w,
		h: len(lines) * t.font.Metrics().Linespace()}
}

// updateBBox recomputes the layout and the bounding box; Translate and
// Scale, which change neither the text nor its font, only call computeBBox.
func (t *TextItem) updateBBox() {
	t.relayout()
	t.computeBBox()
}

func (t *TextItem) computeBBox() {
	if t.font == nil || len(t.text) == 0 {
		t.X1 = int(t.x)
		t.Y1 = int(t.y)
		t.X2 = int(t.x)
		t.Y2 = int(t.y)
		return
	}

	textW, textH := t.lay.w, t.lay.h

	ax, ay := anchorOffset(t.anchor, textW, textH)

	if t.angle == 0 {
		t.X1 = int(math.Floor(t.x + float64(ax) + 0.5))
		t.Y1 = int(math.Floor(t.y + float64(ay) + 0.5))
		t.X2 = t.X1 + textW
		t.Y2 = t.Y1 + textH
		return
	}

	// Rotated bounding box: compute all 4 rotated corners and take the AABB.
	rad := t.angle * math.Pi / 180.0
	cosA, sinA := math.Cos(rad), math.Sin(rad)
	// Corners relative to anchor point (t.x, t.y), before rotation.
	dxs := [4]float64{float64(ax), float64(ax + textW), float64(ax), float64(ax + textW)}
	dys := [4]float64{float64(ay), float64(ay), float64(ay + textH), float64(ay + textH)}
	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64
	for i := range 4 {
		// Clockwise rotation by angle (Tk convention): rotX = dx*c + dy*s, rotY = dy*c - dx*s
		rx := t.x + dxs[i]*cosA + dys[i]*sinA
		ry := t.y + dys[i]*cosA - dxs[i]*sinA
		if rx < minX {
			minX = rx
		}
		if ry < minY {
			minY = ry
		}
		if rx > maxX {
			maxX = rx
		}
		if ry > maxY {
			maxY = ry
		}
	}
	t.X1 = int(math.Floor(minX))
	t.Y1 = int(math.Floor(minY))
	t.X2 = int(math.Ceil(maxX))
	t.Y2 = int(math.Ceil(maxY))
}

func (t *TextItem) Display(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	clipX, clipY, clipW, clipH, originX, originY int) {

	if t.font == nil || t.color == nil {
		return
	}

	m := t.font.Metrics()
	lines, xs := t.lay.lines, t.lay.xs
	ls := m.Linespace()

	ax, ay := anchorOffset(t.anchor, t.lay.w, t.lay.h)

	if t.angle == 0 {
		// DisplayCanvText: drawOrigin = (x, y) + anchor offset, rounded by
		// Tk_CanvasDrawableCoords.
		drawX := drawableCoord(t.x+float64(ax), originX)
		drawY := drawableCoord(t.y+float64(ay), originY)

		if df, ok := t.font.(platform.DrawableFont); ok {
			for i := range lines {
				for _, seg := range t.lay.segs[i] {
					if seg.Text != "" {
						df.DrawString(drawable, drawX+xs[i]+seg.X, drawY+i*ls+m.Ascent,
							seg.Text, t.color.Pixel, t.color.Red, t.color.Green, t.color.Blue)
					}
				}
			}
		}

		// Draw text cursor if focused, on the display line holding it
		// (DisplayCanvText via Tk_CharBbox).
		if t.hasFocus {
			li, col := t.cursorLine()
			cursorX, cursorY := drawX, drawY
			if li < len(lines) {
				cursorX += xs[li] + font.TextWidth(t.font, lines[li][:col])
				cursorY += li * ls
			}
			d.SetForeground(gc, t.color.Pixel)
			d.FillRectangle(drawable, gc, cursorX, cursorY, 2, uint(ls))
		}
		return
	}

	// Rotated text: each line's baseline origin is rotated about the anchor
	// point, as the whole layout is in DisplayCanvText.
	rad := t.angle * math.Pi / 180.0
	cosA, sinA := math.Cos(rad), math.Sin(rad)
	af, ok := t.font.(interface {
		DrawStringAngle(platform.DrawableID, int, int, float64, string, uint64, uint16, uint16, uint16)
	})
	if !ok {
		return
	}
	for i, line := range lines {
		if line == "" {
			continue
		}
		fx := float64(ax + xs[i])
		fy := float64(ay+i*ls) + float64(m.Ascent)
		bx := drawableCoord(t.x+fx*cosA+fy*sinA, originX)
		by := drawableCoord(t.y+fy*cosA-fx*sinA, originY)
		af.DrawStringAngle(drawable, bx, by, t.angle,
			line, t.color.Pixel, t.color.Red, t.color.Green, t.color.Blue)
	}
}

func (t *TextItem) PointDistance(x, y float64) float64 {
	return rectPointDistance(x, y, float64(t.X1), float64(t.Y1), float64(t.X2), float64(t.Y2))
}

func (t *TextItem) AreaOverlap(ax1, ay1, ax2, ay2 float64) int {
	x1, y1, x2, y2 := float64(t.X1), float64(t.Y1), float64(t.X2), float64(t.Y2)
	if ax2 < x1 || ax1 > x2 || ay2 < y1 || ay1 > y2 {
		return -1
	}
	if ax1 <= x1 && ax2 >= x2 && ay1 <= y1 && ay2 >= y2 {
		return 1
	}
	return 0
}

func (t *TextItem) Scale(ox, oy, sx, sy float64) {
	t.x = ox + (t.x-ox)*sx
	t.y = oy + (t.y-oy)*sy
	t.computeBBox()
}

func (t *TextItem) Translate(dx, dy float64) {
	t.x += dx
	t.y += dy
	t.computeBBox()
}

func (t *TextItem) Delete(d platform.DisplayServer) {}

// Postscript emits a PostScript representation of the text item.
//
// Mirrors tk/generic/tkCanvText.c:TextToPostscript.
func (t *TextItem) Postscript(ps *PSContext) error {
	if t.State() == ItemStateHidden {
		return nil
	}
	if t.font == nil || len(t.text) == 0 {
		return nil
	}
	ps.Font(t.font)
	if ps.Prepass {
		return nil
	}
	ps.Color(t.color)

	// Anchor offset in pixels.
	w := t.font.MeasureString(t.text)
	m := t.font.Metrics()
	h := m.Linespace()
	ax, ay := anchorOffset(t.anchor, w, h)
	x := float64(int(t.x) + ax)
	y := float64(int(t.y) + ay)

	// PostScript y is anchored at the baseline. We approximate baseline
	// as y + ascent (Tk does this with tk_anchorY corrections).
	ps.writef("%.15g %.15g moveto\n", x, ps.PsY(int(y)+m.Ascent))
	ps.writef("(%s) show\n", psEscape(t.text))
	ps.write("newpath\n")
	return nil
}

// psEscape escapes a string for inclusion in a PostScript literal string
// between parentheses. Mirrors Tcl_AppendPrintfootToObj conventions.
func psEscape(s string) string {
	var b []byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '(', ')', '\\':
			b = append(b, '\\', c)
		case '\n':
			b = append(b, '\\', 'n')
		case '\r':
			b = append(b, '\\', 'r')
		case '\t':
			b = append(b, '\\', 't')
		default:
			if c < 32 || c > 126 {
				b = append(b, '\\', '0'+(c>>6)&0x7, '0'+(c>>3)&0x7, '0'+c&0x7)
			} else {
				b = append(b, c)
			}
		}
	}
	return string(b)
}

// anchorOffset computes the top-left offset from the anchor point
// for a region of size (w, h).
func anchorOffset(a option.Anchor, w, h int) (int, int) {
	var dx, dy int
	switch a {
	case option.AnchorN:
		dx = -w / 2
	case option.AnchorNE:
		dx = -w
	case option.AnchorE:
		dx = -w
		dy = -h / 2
	case option.AnchorSE:
		dx = -w
		dy = -h
	case option.AnchorS:
		dx = -w / 2
		dy = -h
	case option.AnchorSW:
		dy = -h
	case option.AnchorW:
		dy = -h / 2
	case option.AnchorNW:
		// top-left: no offset
	case option.AnchorCenter:
		dx = -w / 2
		dy = -h / 2
	}
	return dx, dy
}

// Unused import guard.
var _ = math.MaxFloat64
