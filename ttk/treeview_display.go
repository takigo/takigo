package ttk

import (
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/screenunit"
)

// Display renders the treeview to its window via double-buffered pixmap.
func (tv *Treeview) Display() {
	if tv.Destroyed {
		return
	}
	win := tv.Win
	if win.PlatformID == 0 || !win.IsViewable() {
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
	// TreeviewDoLayout: fit the columns to the tree area.
	tv.resizeColumns(width)

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

	_, isDF := tv.Font.(platform.DrawableFont)
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
				tv.drawAlignedText(tv.headingFont(), pixDrawable, colX+1, 0, tv.treeColumnWidth-2,
					tv.headingHeight, tv.treeHeadingText, option.AnchorCenter, fg, fgR, fgG, fgB)
			}
			colX += tv.treeColumnWidth
		}

		for _, col := range tv.columns {
			draw.Fill3DRectangle(d, pixDrawable, gc, border,
				colX, 0, col.Width, tv.headingHeight, 1, option.ReliefRaised)
			if col.HeadingText != "" {
				// Treeheading.border is 1px and the Heading style has no padding.
				textW := col.Width - 2
				textX := colX + 1
				// Reserve space for sort indicator.
				if tv.sortInd != nil && tv.sortInd.columnID == col.ID {
					textW -= 12
				}
				if textW > 0 {
					tv.drawAlignedText(tv.headingFont(), pixDrawable, textX, 0, textW,
						tv.headingHeight, col.HeadingText, col.HeadingAnchor, fg, fgR, fgG, fgB)
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
	for i := range visRows {
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

		// Tree column: the Item layout (ttkTreeview.c) in a parcel indented
		// by depth*indent: Treeitem.indicator (-indicatorsize 9p, odd, plus
		// -indicatormargins {1.5p 1.5p 3p 1.5p}) packed left even for leaves,
		// then Treeitem.image packed left (anchor w), then Treeitem.text.
		if tv.showTree {
			rowBg := fieldBg
			if isSelected {
				rowBg = selBg
			} else if tv.Stripe && idx%2 == 1 {
				rowBg = stripeBg
			}
			parcelX := colX + depth*tv.indent
			indSize := screenunit.Px("9p")
			if indSize%2 == 0 {
				indSize--
			}
			ml, mt, mr, mb := screenunit.Px("1.5p"), screenunit.Px("1.5p"), screenunit.Px("3p"), screenunit.Px("1.5p")
			if len(item.Children) > 0 {
				// Unfilled node: its requested size, centred in the row.
				boxY := rowY + (tv.rowHeight-(indSize+mt+mb))/2
				tv.drawIndicator(d, pixDrawable, gc, Box{parcelX + ml, boxY + mt,
					indSize, indSize}, item.Open, textPixel)
			}
			textStartX := parcelX + indSize + ml + mr

			if item.Image != nil {
				imgW := item.Image.Width()
				imgH := item.Image.Height()
				imgY := rowY + (tv.rowHeight-imgH)/2
				item.Image.Draw(d, pixDrawable, gc, tv.Win.Depth,
					0, 0, imgW, imgH, textStartX, imgY, rowBg)
				textStartX += imgW
			}

			// Draw item text.
			if item.Text != "" {
				textY := rowY + (tv.rowHeight-m.Linespace())/2 + m.Ascent
				maxW := tv.treeColumnWidth - (textStartX - colX) - 4
				tv.drawClippedText(tv.Font, pixDrawable, textStartX, textY, maxW,
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
				tv.drawAlignedText(tv.Font, pixDrawable, colX+4, rowY, col.Width-8,
					tv.rowHeight, val, col.Anchor, textPixel, textR, textG, textB)
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
	// Treeview.field: FieldElement's border is -fieldbackground's 3D border.
	border := draw.NewBorderFromPixel(fieldBg)
	draw.Draw3DRectangle(d, win.Drawable(), gc, border, 0, 0, win.Width, win.Height, fb, option.ReliefSunken)
	d.Flush()
}

func (tv *Treeview) drawAlignedText(f font.Font, drawable platform.DrawableID,
	x, y, maxW, h int, text string, anchor option.Anchor,
	pixel uint64, r, g, b uint16) {

	m := f.Metrics()
	textW := f.MeasureString(text)
	textY := y + (h-m.Linespace())/2 + m.Ascent

	textX := x
	switch anchor {
	case option.AnchorCenter, option.AnchorN, option.AnchorS:
		textX = x + (maxW-textW)/2
	case option.AnchorE, option.AnchorNE, option.AnchorSE:
		textX = x + maxW - textW
	}

	tv.drawClippedText(f, drawable, textX, textY, maxW, text, pixel, r, g, b)
}

func (tv *Treeview) drawClippedText(f font.Font, drawable platform.DrawableID,
	x, y, maxW int, text string, pixel uint64, r, g, b uint16) {

	df, ok := f.(platform.DrawableFont)
	if maxW <= 0 || !ok {
		return
	}
	textW := f.MeasureString(text)
	if textW <= maxW {
		df.DrawString(drawable, x, y, text, pixel, r, g, b)
		return
	}
	// Truncate with ellipsis.
	ellipsis := "..."
	ellW := f.MeasureString(ellipsis)
	avail := maxW - ellW
	if avail <= 0 {
		return
	}
	for i := len(text); i > 0; i-- {
		if f.MeasureString(text[:i]) <= avail {
			df.DrawString(drawable, x, y, text[:i]+ellipsis, pixel, r, g, b)
			return
		}
	}
}

// drawIndicator ports TreeitemIndicatorDraw: an arrow outline sized by
// TtkArrowSize, centred in the (margin-padded) box b, drawn by TtkDrawArrow.
func (tv *Treeview) drawIndicator(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	b Box, open bool, pixel uint64) {
	var cx, cy int
	if open {
		h := b.Width / 2
		cx, cy = 2*h+1, h+1
		if (b.Height-cy)%2 == 1 {
			cy++
		}
	} else {
		h := b.Height / 2
		cx, cy = h+1, 2*h+1
		if (b.Width-cx)%2 == 1 {
			cx++
		}
	}
	b = StickBox(b, cx, cy, 0)
	var pts [4]draw.Point
	if open { // ArrowPoints, ARROW_DOWN
		h := (b.Width - 1) / 2
		x, y := b.X+h, b.Y+b.Height-1
		if b.Height <= h {
			h = b.Height - 1
		}
		pts = [4]draw.Point{{X: x, Y: y}, {X: x - h, Y: y - h}, {X: x + h, Y: y - h}, {X: x, Y: y}}
	} else { // ARROW_RIGHT
		h := (b.Height - 1) / 2
		x, y := b.X+b.Width-1, b.Y+h
		if b.Width <= h {
			h = b.Width - 1
		}
		pts = [4]draw.Point{{X: x, Y: y}, {X: x - h, Y: y - h}, {X: x - h, Y: y + h}, {X: x, Y: y}}
	}
	d.SetForeground(gc, pixel)
	for i := range 3 {
		d.DrawLine(drawable, gc, pts[i].X, pts[i].Y, pts[i+1].X, pts[i+1].Y)
	}
	d.DrawLine(drawable, gc, pts[2].X, pts[2].Y, pts[2].X, pts[2].Y)
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
	r := uint16((pixel>>16)&0xFF) * 257
	g := uint16((pixel>>8)&0xFF) * 257
	b := uint16((pixel)&0xFF) * 257
	return r, g, b
}
