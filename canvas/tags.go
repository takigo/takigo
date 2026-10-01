package canvas

import (
	"slices"
	"strconv"
)

// ItemID identifies a canvas item; the Create methods return one.
type ItemID int64

// Selector is what the item methods take to pick their items: one ItemID,
// or a string holding a tag, "all" or "current" (or an ID in decimal).
type Selector interface {
	ItemID | string
}

func selectorString[S Selector](sel S) string {
	if id, ok := any(sel).(ItemID); ok {
		return strconv.FormatInt(int64(id), 10)
	}
	return any(sel).(string)
}

// resolve converts a tag-or-ID string to a list of matching item entries.
// Supports: integer IDs, "all", "current", and tag name strings.
func (c *Canvas) resolve[S Selector](sel S) []*itemEntry {
	tagOrID := selectorString(sel)
	// Try numeric ID first.
	if id, err := strconv.ParseInt(tagOrID, 10, 64); err == nil {
		if entry, ok := c.idMap[ItemID(id)]; ok {
			return []*itemEntry{entry}
		}
		return nil
	}

	switch tagOrID {
	case "all":
		c.compact()
		result := make([]*itemEntry, len(c.items))
		copy(result, c.items)
		return result
	case "current":
		if c.currentItem != nil {
			return []*itemEntry{c.currentItem}
		}
		return nil
	default:
		// Tag name: mark the indexed entries, then collect them in display
		// order.
		indexed := c.tagIndex[tagOrID]
		if len(indexed) == 0 {
			return nil
		}
		c.markSeq++
		stamp := c.markSeq
		for _, entry := range indexed {
			entry.mark = stamp
		}
		c.compact()
		result := make([]*itemEntry, 0, len(indexed))
		for _, entry := range c.items {
			if entry.mark == stamp {
				result = append(result, entry)
			}
		}
		return result
	}
}

// itemBase extracts the ItemBase from an Item if it embeds one.
// All our concrete item types embed ItemBase, so we use a type assertion
// on a common interface.
type hasBase interface {
	base() *ItemBase
}

func itemBase(item Item) *ItemBase {
	if hb, ok := item.(hasBase); ok {
		return hb.base()
	}
	return nil
}

// findClosest ports CanvasFindClosest (tk/generic/tkCanvas.c): the topmost
// item within halo of (x, y), skipping hidden items and, before the distance
// test, items whose bounding box is not within halo of the point.
func (c *Canvas) findClosest(x, y float64, halo float64) *itemEntry {
	c.compact()
	for _, entry := range slices.Backward(c.items) {
		item := entry.item
		if item.State() == ItemStateHidden {
			continue
		}
		if !bboxNear(item, x, y, halo) {
			continue
		}
		if item.PointDistance(x, y) <= halo {
			return entry
		}
	}
	return nil
}

// bboxNear reports whether item's bounding box comes within dist of (x, y).
// Every item's box contains what its PointDistance measures, so a false
// result rules the item out without calling it.
func bboxNear(item Item, x, y, dist float64) bool {
	x1, y1, x2, y2 := item.BBox()
	return float64(x1) <= x+dist && float64(x2) >= x-dist &&
		float64(y1) <= y+dist && float64(y2) >= y-dist
}
