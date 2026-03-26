package canvas

import (
	"strconv"
)

// resolve converts a tag-or-ID string to a list of matching item entries.
// Supports: integer IDs, "all", "current", and tag name strings.
func (c *Canvas) resolve(tagOrID string) []*itemEntry {
	// Try numeric ID first.
	if id, err := strconv.ParseInt(tagOrID, 10, 64); err == nil {
		if entry, ok := c.idMap[id]; ok {
			return []*itemEntry{entry}
		}
		return nil
	}

	switch tagOrID {
	case "all":
		result := make([]*itemEntry, len(c.items))
		copy(result, c.items)
		return result
	case "current":
		if c.currentItem != nil {
			return []*itemEntry{c.currentItem}
		}
		return nil
	default:
		// Tag name: use tag index for O(1) lookup per tag.
		indexed := c.tagIndex[tagOrID]
		if len(indexed) == 0 {
			return nil
		}
		// Return entries in display order (iterate items, filter by index).
		result := make([]*itemEntry, 0, len(indexed))
		for _, entry := range c.items {
			if indexed[entry.id] != nil {
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

// findByID returns the item entry with the given ID, or nil.
func (c *Canvas) findByID(id int64) *itemEntry {
	return c.idMap[id]
}

// findClosest returns the topmost item within halo distance of (x, y).
// Items are searched from top (last) to bottom (first) in display order.
func (c *Canvas) findClosest(x, y float64, halo float64) *itemEntry {
	for i := len(c.items) - 1; i >= 0; i-- {
		entry := c.items[i]
		if base := itemBase(entry.item); base != nil && base.State == ItemStateHidden {
			continue
		}
		dist := entry.item.PointDistance(x, y)
		if dist <= halo {
			return entry
		}
	}
	return nil
}
