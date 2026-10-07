package ttk

import (
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/internal/textedit"
	"github.com/takigo/takigo/platform"
)

// fieldText is what an entry-like widget (entry, combobox, spinbox) shows
// in its field: the text, the selection and the insertion cursor, drawn as
// EntryDisplay in ttkEntry.c does.
type fieldText struct {
	font font.Font
	text []rune
	x    int // where the text starts, scrolled

	left, right int // the visible span of the field
	top, height int // the inner box the text is centred in

	selFirst, selLast int // the selection; selFirst < 0 for none
	cursor            int // rune index of the cursor; -1 for none

	fg, selBg, selFg, insertColor uint64
	insertWidth                   int
}

// drawFieldText draws ft: the selection's background, the text in
// segments around the selection, and the cursor.
func drawFieldText(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, ft fieldText) {
	if ft.font == nil {
		return
	}
	m := ft.font.Metrics()
	baseline := ft.top + (ft.height-m.Linespace())/2 + m.Ascent
	n := len(ft.text)
	at := func(i int) int {
		return ft.x + ft.font.MeasureString(string(ft.text[:textedit.ClampIdx(i, n)]))
	}
	hasSel := ft.selFirst >= 0 && ft.selLast > ft.selFirst && n > 0
	if hasSel {
		x0 := max(at(ft.selFirst), ft.left)
		x1 := min(at(ft.selLast), ft.right)
		if x1 > x0 {
			d.SetForeground(gc, ft.selBg)
			d.FillRectangle(drawable, gc, x0, ft.top, uint(x1-x0), uint(m.Linespace()))
		}
	}
	if df, ok := ft.font.(platform.DrawableFont); ok {
		seg := func(start, end int, pixel uint64) {
			start, end = textedit.ClampIdx(start, n), textedit.ClampIdx(end, n)
			if start >= end {
				return
			}
			s := string(ft.text[start:end])
			x := at(start)
			if x >= ft.right || x+ft.font.MeasureString(s) <= 0 {
				return
			}
			drawString(df, drawable, x, baseline, s, pixel)
		}
		if hasSel {
			seg(0, ft.selFirst, ft.fg)
			seg(ft.selFirst, ft.selLast, ft.selFg)
			seg(ft.selLast, n, ft.fg)
		} else {
			seg(0, n, ft.fg)
		}
	}
	if ft.cursor >= 0 {
		if x := at(ft.cursor); x >= ft.left && x < ft.right {
			d.SetForeground(gc, ft.insertColor)
			d.FillRectangle(drawable, gc, x, ft.top, uint(ft.insertWidth), uint(ft.height))
		}
	}
}

// drawString draws s with its baseline at (x, y) in the colour of pixel.
func drawString(df platform.DrawableFont, drawable platform.DrawableID, x, y int, s string, pixel uint64) {
	df.DrawString(drawable, x, y, s, pixel,
		uint16((pixel>>16)&0xFF)*257, uint16((pixel>>8)&0xFF)*257, uint16(pixel&0xFF)*257)
}
