package text

// Tag methods on *TextWidget. Tags are stored in t.doc.TagRanges;
// per-tag bindings and hover state live on the widget itself
// (t.tagBindings, t.hoverTags).

// TagAdd adds a tag to the given range.
func (t *TextWidget) TagAdd(tagName, startIndex, endIndex string) {
	start, ok1 := ParseIndex(t.doc, startIndex)
	end, ok2 := ParseIndex(t.doc, endIndex)
	if !ok1 || !ok2 {
		return
	}
	t.doc.TagAdd(tagName, start, end)
}

// TagRemove removes a tag from the given range.
func (t *TextWidget) TagRemove(tagName, startIndex, endIndex string) {
	start, ok1 := ParseIndex(t.doc, startIndex)
	end, ok2 := ParseIndex(t.doc, endIndex)
	if !ok1 || !ok2 {
		return
	}
	t.doc.TagRemove(tagName, start, end)
}

// TagConfigure configures a tag's display attributes.
func (t *TextWidget) TagConfigure(tagName string, opts ...TagOption) {
	t.doc.TagConfigure(tagName, t.App.ColorCache(), t.App.FontRegistry(), opts...)
}

// TagBind binds an event handler to a text tag. Supported events:
// "<Enter>", "<Leave>", "<Button-1>".
func (t *TextWidget) TagBind(tagName, eventName string, handler func()) {
	if t.tagBindings == nil {
		t.tagBindings = make(map[string]map[string][]func())
	}
	if t.tagBindings[tagName] == nil {
		t.tagBindings[tagName] = make(map[string][]func())
	}
	t.tagBindings[tagName][eventName] = append(t.tagBindings[tagName][eventName], handler)
}

// tagsAtIndex returns the set of tag names that cover the given index.
func (t *TextWidget) tagsAtIndex(idx Index) map[string]bool {
	result := make(map[string]bool)
	for _, tr := range t.doc.TagRanges {
		if Compare(idx, tr.Start) >= 0 && Compare(idx, tr.End) < 0 {
			result[tr.TagName] = true
		}
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

// updateTagHover computes Enter/Leave events as the mouse moves over tags.
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
}
