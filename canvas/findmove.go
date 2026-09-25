package canvas

import "math"

// FindClosest ports "find closest x y ?halo? ?start?" (tkCanvas.c): the
// search runs circularly from the first item matching start (or the bottom
// of the display list) and a later item wins ties, so with start it yields
// the topmost closest item below start. Hidden items are skipped; 0 means
// none.
func (c *Canvas) FindClosest(x, y, halo float64, start string) int64 {
	if len(c.items) == 0 {
		return 0
	}
	startIdx := 0
	if start != "" {
		if es := c.resolve(start); len(es) > 0 {
			for i, e := range c.items {
				if e == es[0] {
					startIdx = i
					break
				}
			}
		}
	}
	dist := func(e *itemEntry) float64 {
		return math.Max(0, e.item.PointDistance(x, y)-halo)
	}
	hidden := func(e *itemEntry) bool { return e.item.State() == ItemStateHidden }
	n := len(c.items)
	idx := startIdx
	for k := 0; k < n && hidden(c.items[idx]); k++ {
		idx = (idx + 1) % n
	}
	if hidden(c.items[idx]) {
		return 0
	}
	closest, best := idx, dist(c.items[idx])
	for i := (idx + 1) % n; i != startIdx; i = (i + 1) % n {
		e := c.items[i]
		if hidden(e) {
			continue
		}
		if d := dist(e); d <= best {
			closest, best = i, d
		}
	}
	return c.items[closest].id
}

// MoveTo ports "moveto tagOrId x y": translate every matching item so the
// bounding box origin of the first one lands on (x, y).
func (c *Canvas) MoveTo(tagOrID string, x, y float64) {
	es := c.resolve(tagOrID)
	if len(es) == 0 {
		return
	}
	x1, y1, _, _ := es[0].item.BBox()
	c.Move(tagOrID, x-float64(x1), y-float64(y1))
}
