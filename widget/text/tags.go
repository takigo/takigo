package text

import (
	"fmt"
	"strings"
)

// Tag methods on *TextWidget. Tag ranges are stored on each *Tag in t.doc;
// per-tag bindings and hover state live on the widget itself
// (t.tagBindings, t.hoverTags).

// TagAdd adds a tag to the given range.
func (t *TextWidget) TagAdd[A, B IndexSpec](tagName string, startIndex A, endIndex B) error {
	start, ok1 := indexOf(t, startIndex)
	end, ok2 := indexOf(t, endIndex)
	if !ok1 || !ok2 {
		return ErrBadIndex
	}
	t.doc.tagAdd(tagName, start, end, isEnd(endIndex))
	t.Display()
	return nil
}

// TagRemove removes a tag from the given range.
func (t *TextWidget) TagRemove[A, B IndexSpec](tagName string, startIndex A, endIndex B) error {
	start, ok1 := indexOf(t, startIndex)
	end, ok2 := indexOf(t, endIndex)
	if !ok1 || !ok2 {
		return ErrBadIndex
	}
	t.doc.TagRemove(tagName, start, end)
	t.Display()
	return nil
}

// TagConfigure configures a tag's display attributes.
func (t *TextWidget) TagConfigure(tagName string, opts ...TagOption) {
	t.doc.TagConfigure(tagName, t.App.ColorCache(), t.App.FontRegistry(), opts...)
	t.Display()
}

// TagDelete deletes the tags, their ranges and their bindings (tag delete).
func (t *TextWidget) TagDelete(tagNames ...string) {
	t.doc.TagDelete(tagNames...)
	for _, name := range tagNames {
		delete(t.tagBindings, name)
		delete(t.hoverTags, name)
	}
	t.Display()
}

// TagRaise raises the tag's priority just above aboveThis, or above all
// other tags when aboveThis is "" (tag raise).
func (t *TextWidget) TagRaise(tagName, aboveThis string) {
	t.doc.TagRaise(tagName, aboveThis)
	t.Display()
}

// TagLower lowers the tag's priority just below belowThis, or below all
// other tags when belowThis is "" (tag lower).
func (t *TextWidget) TagLower(tagName, belowThis string) {
	t.doc.TagLower(tagName, belowThis)
	t.Display()
}

// TagNames returns every tag, lowest priority first (tag names).
func (t *TextWidget) TagNames() []string {
	return t.doc.TagNames()
}

// TagNamesAt returns the tags on the character at index, lowest priority
// first (tag names index).
func (t *TextWidget) TagNamesAt[I IndexSpec](index I) []string {
	idx, ok := indexOf(t, index)
	if !ok {
		return nil
	}
	var names []string
	for _, tg := range t.doc.TagsAt(idx) {
		names = append(names, tg.Name)
	}
	return names
}

// TagRanges returns the ranges the tag covers, in order (tag ranges).
func (t *TextWidget) TagRanges(tagName string) []TagRange {
	return append([]TagRange(nil), t.doc.TagRangesFor(tagName)...)
}

// TagNextRange returns the first range of the tag that starts at or after
// start and before end (tag nextrange). Pass "end" for the rest of the text.
func (t *TextWidget) TagNextRange[A, B IndexSpec](tagName string, start A, end B) (TagRange, bool) {
	s, ok1 := indexOf(t, start)
	e, ok2 := indexOf(t, end)
	if !ok1 || !ok2 {
		return TagRange{}, false
	}
	return t.doc.TagNextRange(tagName, s, e)
}

// TagPrevRange returns the last range of the tag that starts before start
// and at or after end (tag prevrange). Pass "1.0" for the whole text.
func (t *TextWidget) TagPrevRange[A, B IndexSpec](tagName string, start A, end B) (TagRange, bool) {
	s, ok1 := indexOf(t, start)
	e, ok2 := indexOf(t, end)
	if !ok1 || !ok2 {
		return TagRange{}, false
	}
	return t.doc.TagPrevRange(tagName, s, e)
}

// TagBind binds an event handler to a text tag. The events are "<Enter>",
// "<Leave>", "<Motion>", "<Button-N>" (also "<N>" and "<ButtonPress-N>"),
// "<Double-Button-N>", "<Triple-Button-N>" and "<ButtonRelease-N>"; a
// double click runs a tag's <Double-Button-N> binding instead of its
// <Button-N> one when it has both, as in Tk. As in Tk the "current" mark is at the character
// under the pointer while a handler runs, so
//
//	tw.TagBind("link", "<Button-1>", func() {
//		idx, _ := tw.Index("current")
//		follow(idx)
//	})
//
// finds the character that was clicked.
func (t *TextWidget) TagBind(tagName, eventName string, handler func()) {
	if t.tagBindings == nil {
		t.tagBindings = make(map[string]map[string][]func())
	}
	if t.tagBindings[tagName] == nil {
		t.tagBindings[tagName] = make(map[string][]func())
	}
	ev := normalizeTagEvent(eventName)
	t.tagBindings[tagName][ev] = append(t.tagBindings[tagName][ev], handler)
}

// normalizeTagEvent maps the spellings of one button event to one key.
func normalizeTagEvent(ev string) string {
	inner, ok := strings.CutPrefix(ev, "<")
	if !ok {
		return ev
	}
	inner = strings.TrimSuffix(inner, ">")
	if len(inner) == 1 && inner[0] >= '1' && inner[0] <= '9' {
		return "<Button-" + inner + ">"
	}
	if n, ok := strings.CutPrefix(inner, "ButtonPress-"); ok {
		return "<Button-" + n + ">"
	}
	for _, rep := range []string{"Double-", "Triple-"} {
		if rest, ok := strings.CutPrefix(inner, rep); ok {
			if b := normalizeTagEvent("<" + rest + ">"); strings.HasPrefix(b, "<Button-") {
				return "<" + rep + b[1:]
			}
		}
	}
	return ev
}

// fireTagClick fires a button press on the tags at idx the way Tk picks a
// binding: for each tag the most specific one bound, a triple click
// falling back to <Double-Button-N> and then <Button-N>.
func (t *TextWidget) fireTagClick(idx Index, button, count int) {
	if len(t.tagBindings) == 0 {
		return
	}
	single := fmt.Sprintf("<Button-%d>", button)
	names := []string{single}
	if count >= 2 {
		names = append([]string{"<Double-" + single[1:]}, names...)
	}
	if count >= 3 {
		names = append([]string{"<Triple-" + single[1:]}, names...)
	}
	for _, tg := range t.doc.TagsAt(idx) {
		for _, ev := range names {
			if len(t.tagBindings[tg.Name][ev]) > 0 {
				t.fireTagHandlers(tg.Name, ev)
				break
			}
		}
	}
}

// tagsAtIndex returns the set of tag names that cover the given index.
func (t *TextWidget) tagsAtIndex(idx Index) map[string]bool {
	result := make(map[string]bool)
	for _, tg := range t.doc.TagsAt(idx) {
		result[tg.Name] = true
	}
	return result
}

// fireTagHandlers fires all handlers for (tagName, eventName).
func (t *TextWidget) fireTagHandlers(tagName, eventName string) {
	if t.tagBindings == nil {
		return
	}
	if handlers, ok := t.tagBindings[tagName]; ok {
		for _, h := range handlers[eventName] {
			h()
		}
	}
}

// fireTagEvent fires eventName on the tags at idx, lowest priority first
// as Tk's TkTextBindProc runs a character's tags.
func (t *TextWidget) fireTagEvent(idx Index, eventName string) {
	if len(t.tagBindings) == 0 {
		return
	}
	for _, tg := range t.doc.TagsAt(idx) {
		t.fireTagHandlers(tg.Name, eventName)
	}
}

// updateTagHover computes Enter/Leave events as the mouse moves over tags
// and shows the cursor of the highest-priority tag under the pointer that
// has one.
func (t *TextWidget) updateTagHover(newTags map[string]bool) {
	if t.hoverTags == nil {
		t.hoverTags = make(map[string]bool)
	}
	for tag := range t.hoverTags {
		if !newTags[tag] {
			t.fireTagHandlers(tag, "<Leave>")
		}
	}
	for tag := range newTags {
		if !t.hoverTags[tag] {
			t.fireTagHandlers(tag, "<Enter>")
		}
	}
	t.hoverTags = newTags
	t.updateTagCursor(newTags)
}

// updateTagCursor sets the window cursor from the tags under the pointer.
func (t *TextWidget) updateTagCursor(tags map[string]bool) {
	shape, set := t.mouseCursor, false
	for _, tg := range t.doc.tagsByPriority {
		if tg.CursorSet && tags[tg.Name] {
			shape, set = tg.Cursor, true
		}
	}
	if set == t.tagCursorSet && (!set || shape == t.tagCursor) {
		return
	}
	t.tagCursor, t.tagCursorSet = shape, set
	t.Win.SetCursor(shape)
}

// hasTagCursors reports whether any tag sets a cursor.
func (t *TextWidget) hasTagCursors() bool {
	for _, tg := range t.doc.tagsByPriority {
		if tg.CursorSet {
			return true
		}
	}
	return false
}
