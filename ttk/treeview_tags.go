// Treeview tags: tk/generic/ttk/ttkTreeview.c (tag commands,
// TreeviewBindEventProc) over ttkTagSet.c.
package ttk

import (
	"slices"

	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/widget"
)

// treeTag holds a tag's display options (TagOptionSpecs) and bindings;
// nil options are unset.
type treeTag struct {
	foreground, background, stripedBackground *uint64
	font                                      font.Font
	image                                     widget.WidgetImage
	bindings                                  []tagHandler
}

type tagHandler struct {
	mask    event.Mask
	handler func(*event.Event)
}

// TagOption sets a tag's display option.
type TagOption func(*treeTag)

// TagForeground sets the text colour of the tag's items.
func TagForeground(pixel uint64) TagOption {
	return func(t *treeTag) { t.foreground = &pixel }
}

// TagBackground sets the row background of the tag's items.
func TagBackground(pixel uint64) TagOption {
	return func(t *treeTag) { t.background = &pixel }
}

// TagStripedBackground sets the background of the tag's striped rows.
func TagStripedBackground(pixel uint64) TagOption {
	return func(t *treeTag) { t.stripedBackground = &pixel }
}

// TagFont sets the font of the tag's items.
func TagFont(f font.Font) TagOption {
	return func(t *treeTag) { t.font = f }
}

// TagImage sets the tree-column image of the tag's items that have none.
func TagImage(img widget.WidgetImage) TagOption {
	return func(t *treeTag) { t.image = img }
}

// tag returns the named tag, creating it (Ttk_GetTag): tags created
// earlier take priority over later ones.
func (tv *Treeview) tag(name string) *treeTag {
	if t := tv.tags[name]; t != nil {
		return t
	}
	t := &treeTag{}
	tv.tags[name] = t
	tv.tagOrder = append(tv.tagOrder, name)
	return t
}

// TagConfigure sets display options of a tag.
func (tv *Treeview) TagConfigure(name string, opts ...TagOption) {
	t := tv.tag(name)
	for _, opt := range opts {
		opt(t)
	}
	tv.scheduleRedisplay()
}

// TagBind runs handler for the events in mask that reach an item with the
// tag: key and virtual events go to the focus item, button and motion
// events to the item under the pointer. Tag bindings run before the
// treeview's own bindings.
func (tv *Treeview) TagBind(name string, mask event.Mask, handler func(*event.Event)) {
	t := tv.tag(name)
	t.bindings = append(t.bindings, tagHandler{mask, handler})
}

// TagUnbind removes the tag's bindings.
func (tv *Treeview) TagUnbind(name string) {
	if t := tv.tags[name]; t != nil {
		t.bindings = nil
	}
}

// TagDelete removes the tags from every item and forgets their options
// and bindings.
func (tv *Treeview) TagDelete(names ...string) {
	for _, name := range names {
		if tv.tags[name] == nil {
			continue
		}
		delete(tv.tags, name)
		tv.tagOrder = slices.DeleteFunc(tv.tagOrder, func(n string) bool { return n == name })
		for _, item := range tv.items {
			item.Tags = slices.DeleteFunc(item.Tags, func(n string) bool { return n == name })
		}
	}
	tv.scheduleRedisplay()
}

// TagNames returns the tags that exist, in creation order.
func (tv *Treeview) TagNames() []string { return slices.Clone(tv.tagOrder) }

// TagAdd adds the tag to the items.
func (tv *Treeview) TagAdd(name string, ids ...string) {
	tv.tag(name)
	for _, id := range ids {
		if item := tv.items[id]; item != nil && item != tv.root && !slices.Contains(item.Tags, name) {
			item.Tags = append(item.Tags, name)
		}
	}
	tv.scheduleRedisplay()
}

// TagRemove removes the tag from the items, or from every item when no
// IDs are given.
func (tv *Treeview) TagRemove(name string, ids ...string) {
	drop := func(item *TreeItem) {
		item.Tags = slices.DeleteFunc(item.Tags, func(n string) bool { return n == name })
	}
	if len(ids) == 0 {
		for _, item := range tv.items {
			drop(item)
		}
	}
	for _, id := range ids {
		if item := tv.items[id]; item != nil {
			drop(item)
		}
	}
	tv.scheduleRedisplay()
}

// TagHas returns the items with the tag, in tree order.
func (tv *Treeview) TagHas(name string) []string {
	var ids []string
	tv.preorder(func(item *TreeItem) {
		if slices.Contains(item.Tags, name) {
			ids = append(ids, item.ID)
		}
	})
	return ids
}

// HasTag reports whether the item has the tag.
func (tv *Treeview) HasTag(id, name string) bool {
	item := tv.items[id]
	return item != nil && slices.Contains(item.Tags, name)
}

// itemLook is a DisplayItem: what tags make of an item's row.
type itemLook struct {
	foreground, background *uint64
	font                   font.Font
	image                  widget.WidgetImage
}

// itemLook ports Ttk_TagSetValues with OverrideStriped: per option the
// highest-priority tag that sets it wins; on a striped row a striped
// background replaces the background.
func (tv *Treeview) itemLook(item *TreeItem, striped bool) itemLook {
	var look itemLook
	best := len(tv.tagOrder)
	bgPrio, fgPrio, fontPrio, imgPrio := best, best, best, best
	var stripedBg *uint64
	stripedPrio := best
	for _, name := range item.Tags {
		t := tv.tags[name]
		if t == nil {
			continue
		}
		prio := slices.Index(tv.tagOrder, name)
		if t.foreground != nil && prio < fgPrio {
			look.foreground, fgPrio = t.foreground, prio
		}
		if t.background != nil && prio < bgPrio {
			look.background, bgPrio = t.background, prio
		}
		if t.stripedBackground != nil && prio < stripedPrio {
			stripedBg, stripedPrio = t.stripedBackground, prio
		}
		if t.font != nil && prio < fontPrio {
			look.font, fontPrio = t.font, prio
		}
		if t.image != nil && prio < imgPrio {
			look.image, imgPrio = t.image, prio
		}
	}
	if striped && stripedBg != nil {
		look.background = stripedBg
	}
	return look
}

// dispatchTagEvent ports TreeviewBindEventProc: deliver ev to the
// bindings of the target item's tags, in the item's tag order.
func (tv *Treeview) dispatchTagEvent(ev *event.Event) {
	if len(tv.tags) == 0 {
		return
	}
	var item *TreeItem
	switch ev.Type {
	case event.KeyPressType, event.KeyReleaseType, event.VirtualType:
		item = tv.items[tv.focus]
	case event.ButtonPressType, event.ButtonReleaseType, event.MotionType:
		if id := tv.IdentifyItem(ev.Y); id != "" {
			item = tv.items[id]
		}
	default:
	}
	if item == nil || item == tv.root {
		return
	}
	mask := event.TypeToMask(ev.Type)
	for _, name := range slices.Clone(item.Tags) {
		t := tv.tags[name]
		if t == nil {
			continue
		}
		for _, h := range t.bindings {
			if h.mask&mask != 0 {
				h.handler(ev)
			}
		}
	}
}
