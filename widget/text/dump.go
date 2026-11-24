package text

// The dump command of tk/generic/tkText.c (TextDumpCmd, DumpLine).

import (
	"cmp"
	"slices"
	"strings"
)

// DumpKind is the kind of a dumped item.
type DumpKind int

const (
	DumpText   DumpKind = iota // a run of characters (-text)
	DumpMark                   // a mark (-mark)
	DumpTagOn                  // a tag starts (-tag)
	DumpTagOff                 // a tag ends (-tag)
	DumpImage                  // an embedded image (-image)
	DumpWindow                 // an embedded window (-window)
)

func (k DumpKind) String() string {
	return [...]string{"text", "mark", "tagon", "tagoff", "image", "window"}[k]
}

// DumpItem is one item of a dump: what, its value (the text, or the name
// of the mark, tag, image or window) and where.
type DumpItem struct {
	Kind  DumpKind
	Value string
	Index Index
}

// order sorts the items at one index as the B-tree holds them: a tag that
// ends before the marks there, a tag that starts after them, and the
// character last.
func (k DumpKind) order() int {
	return [...]int{4, 1, 2, 0, 3, 3}[k]
}

// Dump returns the content of [start, end) as Tk's dump does: runs of text
// (split at every tag toggle, mark and embedded object, each line's run
// ending with its newline), marks, tag toggles, images and windows, in
// order. With no kinds it dumps all of them (-all); DumpTagOn or
// DumpTagOff selects both toggles (-tag). Toggles and marks at end are not
// included.
func (t *TextWidget) Dump[A, B IndexSpec](start A, end B, kinds ...DumpKind) ([]DumpItem, error) {
	s, ok1 := indexOf(t, start)
	e, ok2 := indexOf(t, end)
	if !ok1 || !ok2 {
		return nil, ErrBadIndex
	}
	toggle := func(k DumpKind) DumpKind {
		if k == DumpTagOff {
			return DumpTagOn
		}
		return k
	}
	want := func(k DumpKind) bool {
		return len(kinds) == 0 || slices.ContainsFunc(kinds, func(w DumpKind) bool { return toggle(w) == toggle(k) })
	}
	if Compare(s, e) >= 0 {
		return nil, nil
	}
	in := func(p Index) bool { return Compare(p, s) >= 0 && Compare(p, e) < 0 }
	d := t.doc
	var items []DumpItem
	splits := map[Index]bool{}
	add := func(k DumpKind, value string, at Index) {
		splits[at] = true
		if want(k) {
			items = append(items, DumpItem{Kind: k, Value: value, Index: at})
		}
	}
	for _, tg := range d.tagsByPriority {
		for _, r := range tg.ranges {
			if in(r.Start) {
				add(DumpTagOn, tg.Name, r.Start)
			}
			if in(r.End) {
				add(DumpTagOff, tg.Name, r.End)
			}
		}
	}
	objects := map[Index]bool{}
	for name, m := range d.Marks {
		switch {
		case d.objects[name] && d.isPlaceholder(m.Pos):
			if !in(m.Pos) {
				continue
			}
			objects[m.Pos] = true
			if ei := d.images[name]; ei != nil {
				add(DumpImage, ei.instanceName(), m.Pos)
			} else if ew := t.embeddedWinAt(m.Pos.Line, m.Pos.Char); ew != nil {
				add(DumpWindow, ew.win.PathName, m.Pos)
			} else {
				splits[m.Pos] = true // another peer's window
			}
		case in(m.Pos) && !d.objects[name]:
			add(DumpMark, name, m.Pos)
		}
	}
	if want(DumpText) {
		items = append(items, t.dumpText(s, e, splits, objects)...)
	}
	slices.SortStableFunc(items, func(a, b DumpItem) int {
		return cmp.Or(Compare(a.Index, b.Index), cmp.Compare(a.Kind.order(), b.Kind.order()),
			strings.Compare(a.Value, b.Value))
	})
	return items, nil
}

// dumpText returns the runs of text of [s, e), broken at splits and around
// the objects' placeholders, which are left out.
func (t *TextWidget) dumpText(s, e Index, splits, objects map[Index]bool) []DumpItem {
	d := t.doc
	var items []DumpItem
	for l := s.Line; l <= e.Line && l <= d.LineCount(); l++ {
		text := d.Lines[l-1].Text
		from, to := 0, len(text)
		if l == s.Line {
			from = s.Char
		}
		withNL := l < d.LineCount()
		if l == e.Line {
			to, withNL = min(e.Char, len(text)), false
		}
		var run strings.Builder
		runStart := Index{Line: l, Char: from}
		flush := func() {
			if run.Len() > 0 {
				items = append(items, DumpItem{Kind: DumpText, Value: run.String(), Index: runStart})
				run.Reset()
			}
		}
		for c := from; c < to; c++ {
			at := Index{Line: l, Char: c}
			if splits[at] && c > from {
				flush()
			}
			if objects[at] {
				flush()
				runStart = Index{Line: l, Char: c + 1}
				continue
			}
			if run.Len() == 0 {
				runStart = at
			}
			run.WriteRune(text[c])
		}
		if withNL {
			if at := (Index{Line: l, Char: len(text)}); splits[at] && to > from {
				flush()
			}
			if run.Len() == 0 {
				runStart = Index{Line: l, Char: len(text)}
			}
			run.WriteByte('\n')
		}
		flush()
	}
	return items
}
