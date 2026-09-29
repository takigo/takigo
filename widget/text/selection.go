package text

// HasSelection returns true if there is an active selection.
func (t *TextWidget) HasSelection() bool {
	ranges := t.doc.TagRangesFor("sel")
	return len(ranges) > 0
}

// GetSelection returns the currently selected text, or "" if none.
func (t *TextWidget) GetSelection() string {
	ranges := t.doc.TagRangesFor("sel")
	if len(ranges) == 0 {
		return ""
	}
	return t.doc.Get(ranges[0].Start, ranges[0].End)
}

// SelectAll selects all text.
func (t *TextWidget) SelectAll() {
	t.clearSelection()
	start := Index{1, 0}
	end := t.doc.EndIndex()
	if Compare(start, end) < 0 {
		t.doc.TagAdd("sel", start, end)
		t.selAnchor = start
	}
}

// clearSelection removes the sel tag from the entire document.
func (t *TextWidget) clearSelection() {
	t.doc.TagRemove("sel", Index{1, 0}, t.doc.EndIndex())
}

// setSelection sets the selection between anchor and pos.
func (t *TextWidget) setSelection(anchor, pos Index) {
	lo, hi := anchor, pos
	if Compare(pos, anchor) < 0 {
		lo, hi = pos, anchor
	}
	if cur := t.doc.TagRangesFor("sel"); len(cur) == 1 && cur[0].Start == lo && cur[0].End == hi {
		return
	}
	t.clearSelection()
	if Compare(lo, hi) < 0 {
		t.doc.TagAdd("sel", lo, hi)
	}
}

// updateSelection extends or shrinks the selection from selAnchor to pos.
func (t *TextWidget) updateSelection(pos Index) {
	t.setSelection(t.selAnchor, pos)
}

// deleteSelection deletes the selected text and returns true,
// or returns false if no selection.
func (t *TextWidget) deleteSelection() bool {
	ranges := t.doc.TagRangesFor("sel")
	if len(ranges) == 0 {
		return false
	}
	sr := ranges[0]
	text := t.doc.Get(sr.Start, sr.End)
	t.clearSelection()
	t.doc.Delete(sr.Start, sr.End)
	if t.undoEnabled {
		t.undoStack.RecordDelete(sr.Start, sr.End, text)
	}
	t.doc.MarkSet("insert", sr.Start)
	return true
}
