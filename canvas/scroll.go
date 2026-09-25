package canvas

// setOrigin ports CanvasSetOrigin (tk/generic/tkCanvas.c) without scroll
// increments. x, y are the canvas coordinates at the inner (inset) edge, i.e.
// Tk's xOrigin + inset. With a scroll region (-confine is always on) the view
// is only shifted when one side is outside the region and the other has room.
func (c *Canvas) setOrigin(x, y int) {
	if c.hasScrollRegion {
		innerW := c.Win.Width - 2*c.inset
		innerH := c.Win.Height - 2*c.inset
		x = confine(x, x+innerW, c.scrollRegion[0], c.scrollRegion[2])
		y = confine(y, y+innerH, c.scrollRegion[1], c.scrollRegion[3])
	}

	if x == c.xOrigin && y == c.yOrigin {
		return
	}
	c.xOrigin = x
	c.yOrigin = y
	c.scheduleRedraw()
	c.notifyScrollbars()
}

// notifyScrollbars calls XScrollCmd/YScrollCmd with the current visible range.
func (c *Canvas) notifyScrollbars() {
	if !c.hasScrollRegion {
		return
	}

	winW := c.Win.Width - 2*c.inset
	winH := c.Win.Height - 2*c.inset
	totalW := c.scrollRegion[2] - c.scrollRegion[0]
	totalH := c.scrollRegion[3] - c.scrollRegion[1]

	if c.XScrollCmd != nil && totalW > 0 {
		first := float64(c.xOrigin-c.scrollRegion[0]) / float64(totalW)
		last := float64(c.xOrigin-c.scrollRegion[0]+winW) / float64(totalW)
		if first < 0 {
			first = 0
		}
		if last > 1 {
			last = 1
		}
		c.XScrollCmd(first, last)
	}

	if c.YScrollCmd != nil && totalH > 0 {
		first := float64(c.yOrigin-c.scrollRegion[1]) / float64(totalH)
		last := float64(c.yOrigin-c.scrollRegion[1]+winH) / float64(totalH)
		if first < 0 {
			first = 0
		}
		if last > 1 {
			last = 1
		}
		c.YScrollCmd(first, last)
	}
}

// XView scrolls horizontally to the given canvas x coordinate.
func (c *Canvas) XView(x int) {
	c.setOrigin(x, c.yOrigin)
}

// XViewScroll scrolls horizontally by count units or pages.
func (c *Canvas) XViewScroll(count int, pages bool) {
	winW := c.Win.Width - 2*c.inset
	if pages {
		dx := int(float64(winW) * 0.9)
		c.setOrigin(c.xOrigin+count*dx, c.yOrigin)
	} else {
		// Units = 10 pixels.
		c.setOrigin(c.xOrigin+count*10, c.yOrigin)
	}
}

// XViewMoveTo scrolls to a fraction (0.0-1.0) of the scroll region.
func (c *Canvas) XViewMoveTo(fraction float64) {
	if !c.hasScrollRegion {
		return
	}
	totalW := c.scrollRegion[2] - c.scrollRegion[0]
	x := c.scrollRegion[0] + int(fraction*float64(totalW))
	c.setOrigin(x, c.yOrigin)
}

// YView scrolls vertically to the given canvas y coordinate.
func (c *Canvas) YView(y int) {
	c.setOrigin(c.xOrigin, y)
}

// YViewScroll scrolls vertically by count units or pages.
func (c *Canvas) YViewScroll(count int, pages bool) {
	winH := c.Win.Height - 2*c.inset
	if pages {
		dy := int(float64(winH) * 0.9)
		c.setOrigin(c.xOrigin, c.yOrigin+count*dy)
	} else {
		c.setOrigin(c.xOrigin, c.yOrigin+count*10)
	}
}

// YViewMoveTo scrolls to a fraction (0.0-1.0) of the scroll region.
func (c *Canvas) YViewMoveTo(fraction float64) {
	if !c.hasScrollRegion {
		return
	}
	totalH := c.scrollRegion[3] - c.scrollRegion[1]
	y := c.scrollRegion[1] + int(fraction*float64(totalH))
	c.setOrigin(c.xOrigin, y)
}

// SetScrollRegion sets the scrollable area of the canvas.
func (c *Canvas) SetScrollRegion(x1, y1, x2, y2 int) {
	c.scrollRegion = [4]int{x1, y1, x2, y2}
	c.hasScrollRegion = true
	c.notifyScrollbars()
}

// XVisibleRange returns the visible X fraction (for scrollbar).
func (c *Canvas) XVisibleRange() (float64, float64) {
	if !c.hasScrollRegion {
		return 0, 1
	}
	winW := c.Win.Width - 2*c.inset
	totalW := c.scrollRegion[2] - c.scrollRegion[0]
	if totalW <= 0 {
		return 0, 1
	}
	first := float64(c.xOrigin-c.scrollRegion[0]) / float64(totalW)
	last := float64(c.xOrigin-c.scrollRegion[0]+winW) / float64(totalW)
	if first < 0 {
		first = 0
	}
	if last > 1 {
		last = 1
	}
	return first, last
}

// YVisibleRange returns the visible Y fraction (for scrollbar).
func (c *Canvas) YVisibleRange() (float64, float64) {
	if !c.hasScrollRegion {
		return 0, 1
	}
	winH := c.Win.Height - 2*c.inset
	totalH := c.scrollRegion[3] - c.scrollRegion[1]
	if totalH <= 0 {
		return 0, 1
	}
	first := float64(c.yOrigin-c.scrollRegion[1]) / float64(totalH)
	last := float64(c.yOrigin-c.scrollRegion[1]+winH) / float64(totalH)
	if first < 0 {
		first = 0
	}
	if last > 1 {
		last = 1
	}
	return first, last
}

// confine shifts the visible span [lo, hi) toward the region [r1, r2] the way
// CanvasSetOrigin does and returns the new lo.
func confine(lo, hi, r1, r2 int) int {
	left := lo - r1
	right := r2 - hi
	switch {
	case left < 0 && right > 0:
		return lo + min(-left, right)
	case right < 0 && left > 0:
		return lo - min(-right, left)
	}
	return lo
}
