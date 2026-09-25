package ttk

import (
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
)

// Display renders the treeview to its window via double-buffered pixmap.
func (tv *Treeview) Display() {
	if tv.Destroyed {
		return
	}
	win := tv.Win
	if win.PlatformID == 0 {
		return
	}

	d := win.Display.Server
	gc := win.GC
	// Content is drawn in the area inside the Treeview.field border.
	fb := treeviewFieldBorder
	width := win.Width - 2*fb
	height := win.Height - 2*fb

	if width <= 0 || height <= 0 {
		return
	}

	// Allocate or resize pixmap.
	if tv.pixmap == 0 || tv.pixmapW != width || tv.pixmapH != height {
		if tv.pixmap != 0 {
			d.FreePixmap(tv.pixmap)
		}
		tv.pixmap = d.CreatePixmap(win.Drawable(), uint(width), uint(height), uint(win.Depth))
		tv.pixmapW = width
		tv.pixmapH = height
	}
	if tv.pixmap == 0 {
		return
	}

	pixDrawable := platform.PixmapDrawable(tv.pixmap)

	// Colors from style.
	bg := LookupColor(tv.Context.Style, "-background", tv.State, 0xd9d9d9)
	fg := LookupColor(tv.Context.Style, "-foreground", tv.State, 0x000000)
	fieldBg := LookupColor(tv.Context.Style, "-fieldbackground", tv.State, 0xffffff)
	selBg := LookupColor(tv.Context.Style, "-selectbackground", tv.State, 0x4a6984)
	selFg := LookupColor(tv.Context.Style, "-selectforeground", tv.State, 0xffffff)

	// Fill background.
	d.SetForeground(gc, bg)
	d.FillRectangle(pixDrawable, gc, 0, 0, uint(width), uint(height))

	// Field area (white background for items).
	itemAreaY := tv.headerOffset()
	itemAreaH := height - itemAreaY
	if itemAreaH > 0 {
		d.SetForeground(gc, fieldBg)
		d.FillRectangle(pixDrawable, gc, 0, itemAreaY, uint(width), uint(itemAreaH))
	}

	df, isDF := tv.Font.(platform.DrawableFont)
	if !isDF {
		d.CopyArea(pixDrawable, win.Drawable(), gc, 0, 0, uint(width), uint(height), fb, fb)
		d.Flush()
		return
	}

	m := tv.Font.Metrics()
	fgR, fgG, fgB := colorToRGB16(fg)
	selFgR, selFgG, selFgB := colorToRGB16(selFg)

	// --- Draw headings ---
	if tv.showHeadings {
		headBg := LookupColor(tv.Context.Style, "-background", tv.State, 0xd9d9d9)
		border := draw.NewBorderFromPixel(headBg)

		colX := 0
		if tv.showTree {
			// Tree column heading.
			draw.Fill3DRectangle(d, pixDrawable, gc, border,
				colX, 0, tv.treeColumnWidth, tv.headingHeight, 1, option.ReliefRaised)
			if tv.treeHeadingText != "" {
				tv.drawAlignedText(df, pixDrawable, colX+4, 0, tv.treeColumnWidth-8,
					tv.headingHeight, tv.treeHeadingText, option.AnchorW, fg, fgR, fgG, fgB, m)
			}
			colX += tv.treeColumnWidth
		}

		for _, col := range tv.columns {
			draw.Fill3DRectangle(d, pixDrawable, gc, border,
				colX, 0, col.Width, tv.headingHeight, 1, option.ReliefRaised)
			if col.HeadingText != "" {
				textW := col.Width - 8
				textX := colX + 4
				// Reserve space for sort indicator.
				if tv.sortInd != nil && tv.sortInd.columnID == col.ID {
					textW -= 12
				}
				if textW > 0 {
					tv.drawAlignedText(df, pixDrawable, textX, 0, textW,
						tv.headingHeight, col.HeadingText, col.HeadingAnchor, fg, fgR, fgG, fgB, m)
				}
			}
			// Sort indicator.
			if tv.sortInd != nil && tv.sortInd.columnID == col.ID {
				tv.drawSortIndicator(d, pixDrawable, gc, colX+col.Width-14, tv.headingHeight/2, tv.sortInd.reverse, fg)
			}
			colX += col.Width
		}

		// Fill remainder of heading row.
		if colX < width {
			draw.Fill3DRectangle(d, pixDrawable, gc, border,
				colX, 0, width-colX, tv.headingHeight, 1, option.ReliefRaised)
		}
	}

	// Stripe color: slightly darker than field background.
	stripeBg := darkenFieldColor(fieldBg, 13)

	// --- Draw items ---
	visRows := tv.visibleRows()
	for i := 0; i < visRows; i++ {
		idx := tv.topIndex + i
		if idx >= len(tv.displayList) {
			break
		}

		entry := tv.displayList[idx]
		item := entry.item
		depth := entry.depth
		rowY := itemAreaY + i*tv.rowHeight

		isSelected := tv.selection[item.ID]
		isFocused := tv.focus == item.ID

		// Selection highlight or alternating stripe.
		if isSelected {
			d.SetForeground(gc, selBg)
			d.FillRectangle(pixDrawable, gc, 0, rowY, uint(width), uint(tv.rowHeight))
		} else if tv.Stripe && idx%2 == 1 {
			d.SetForeground(gc, stripeBg)
			d.FillRectangle(pixDrawable, gc, 0, rowY, uint(width), uint(tv.rowHeight))
		}

		textPixel := fg
		textR, textG, textB := fgR, fgG, fgB
		if isSelected {
			textPixel = selFg
			textR, textG, textB = selFgR, selFgG, selFgB
		}

		colX := 0

		// Tree column.
		if tv.showTree {
			indentX := depth * tv.indent
			indicatorX := colX + indentX + 2
			textStartX := colX + indentX + tv.indent + 2

			// Draw expand/collapse indicator if item has children.
			if len(item.Children) > 0 {
				tv.drawIndicator(d, pixDrawable, gc, indicatorX, rowY, item.Open, textPixel)
			}

			// Draw item icon image (if any).
			if item.Image != nil {
				imgW := item.Image.Width()
				imgH := item.Image.Height()
				imgY := rowY + (tv.rowHeight-imgH)/2
				item.Image.Draw(d, pixDrawable, gc, tv.Win.Depth,
					0, 0, imgW, imgH, textStartX, imgY, tv.Win.BackgroundPixel)
				textStartX += imgW + 3
			}

			// Draw item text.
			if item.Text != "" {
				textY := rowY + (tv.rowHeight-m.Linespace())/2 + m.Ascent
				maxW := tv.treeColumnWidth - (textStartX - colX) - 4
				tv.drawClippedText(df, pixDrawable, textStartX, textY, maxW,
					item.Text, textPixel, textR, textG, textB)
			}

			colX += tv.treeColumnWidth
		}

		// Data columns.
		for ci, col := range tv.columns {
			val := ""
			if ci < len(item.Values) {
				val = item.Values[ci]
			}
			if val != "" {
				tv.drawAlignedText(df, pixDrawable, colX+4, rowY, col.Width-8,
					tv.rowHeight, val, col.Anchor, textPixel, textR, textG, textB, m)
			}
			colX += col.Width
		}

		// Focus dotted rectangle.
		if isFocused && tv.hasFocus {
			d.SetForeground(gc, fg)
			d.DrawRectangle(pixDrawable, gc, 0, rowY, uint(width-1), uint(tv.rowHeight-1))
		}
	}

	// Draw column separator lines in item area (only for columns with Separator set).
	hasSep := false
	for _, col := range tv.columns {
		if col.Separator {
			hasSep = true
			break
		}
	}
	if itemAreaH > 0 && hasSep {
		sepColor := LookupColor(tv.Context.Style, "-bordercolor", tv.State, 0xa0a0a0)
		d.SetForeground(gc, sepColor)
		colX := 0
		if tv.showTree {
			colX += tv.treeColumnWidth
		}
		for _, col := range tv.columns {
			colX += col.Width
			if col.Separator {
				d.DrawLine(pixDrawable, gc, colX-1, itemAreaY, colX-1, height)
			}
		}
	}

	// Copy inside the field border, then draw the border around it.
	d.CopyArea(pixDrawable, win.Drawable(), gc, 0, 0, uint(width), uint(height), fb, fb)
	border := draw.NewBorderFromPixel(bg)
	draw.Draw3DRectangle(d, win.Drawable(), gc, border, 0, 0, win.Width, win.Height, fb, option.ReliefSunken)
	d.Flush()
}

func (tv *Treeview) drawAlignedText(df platform.DrawableFont, drawable platform.DrawableID,
	x, y, maxW, h int, text string, anchor option.Anchor,
	pixel uint64, r, g, b uint16, m font.Metrics) {

	textW := tv.Font.MeasureString(text)
	textY := y + (h-m.Linespace())/2 + m.Ascent

	textX := x
	switch anchor {
	case option.AnchorCenter, option.AnchorN, option.AnchorS:
		textX = x + (maxW-textW)/2
	case option.AnchorE, option.AnchorNE, option.AnchorSE:
		textX = x + maxW - textW
	}

	tv.drawClippedText(df, drawable, textX, textY, maxW, text, pixel, r, g, b)
}

func (tv *Treeview) drawClippedText(df platform.DrawableFont, drawable platform.DrawableID,
	x, y, maxW int, text string, pixel uint64, r, g, b uint16) {

	if maxW <= 0 {
		return
	}
	textW := tv.Font.MeasureString(text)
	if textW <= maxW {
		df.DrawString(drawable, x, y, text, pixel, r, g, b)
		return
	}
	// Truncate with ellipsis.
	ellipsis := "..."
	ellW := tv.Font.MeasureString(ellipsis)
	avail := maxW - ellW
	if avail <= 0 {
		return
	}
	for i := len(text); i > 0; i-- {
		if tv.Font.MeasureString(text[:i]) <= avail {
			df.DrawString(drawable, x, y, text[:i]+ellipsis, pixel, r, g, b)
			return
		}
	}
}

func (tv *Treeview) drawIndicator(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	x, y int, open bool, pixel uint64) {

	// Draw a small triangle: right-pointing (closed) or down-pointing (open).
	cx := x + 4
	cy := y + tv.rowHeight/2

	d.SetForeground(gc, pixel)

	if open {
		// Down-pointing triangle.
		points := []draw.Point{
			{X: cx - 4, Y: cy - 2},
			{X: cx + 4, Y: cy - 2},
			{X: cx, Y: cy + 3},
		}
		draw.FillPolygon(d, drawable, gc, points)
	} else {
		// Right-pointing triangle.
		points := []draw.Point{
			{X: cx - 2, Y: cy - 4},
			{X: cx + 3, Y: cy},
			{X: cx - 2, Y: cy + 4},
		}
		draw.FillPolygon(d, drawable, gc, points)
	}
}

func (tv *Treeview) drawSortIndicator(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	x, cy int, reverse bool, pixel uint64) {

	d.SetForeground(gc, pixel)

	if reverse {
		// Down arrow.
		points := []draw.Point{
			{X: x, Y: cy - 3},
			{X: x + 8, Y: cy - 3},
			{X: x + 4, Y: cy + 3},
		}
		draw.FillPolygon(d, drawable, gc, points)
	} else {
		// Up arrow.
		points := []draw.Point{
			{X: x + 4, Y: cy - 3},
			{X: x + 8, Y: cy + 3},
			{X: x, Y: cy + 3},
		}
		draw.FillPolygon(d, drawable, gc, points)
	}
}

// darkenFieldColor reduces each RGB channel of a pixel color by amount, clamping to 0.
func darkenFieldColor(pixel uint64, amount uint64) uint64 {
	r := (pixel >> 16) & 0xFF
	g := (pixel >> 8) & 0xFF
	b := pixel & 0xFF
	if r > amount {
		r -= amount
	} else {
		r = 0
	}
	if g > amount {
		g -= amount
	} else {
		g = 0
	}
	if b > amount {
		b -= amount
	} else {
		b = 0
	}
	return (r << 16) | (g << 8) | b
}

func colorToRGB16(pixel uint64) (uint16, uint16, uint16) {
	r := uint16((pixel>>16)&0xFF) << 8
	g := uint16((pixel>>8)&0xFF) << 8
	b := uint16((pixel)&0xFF) << 8
	return r, g, b
}
