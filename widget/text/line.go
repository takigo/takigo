package text

import (
	"slices"
	"sort"
	"strings"

	"github.com/takigo/takigo/color"
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/widget"
)

// Line holds one logical line of text as a rune slice.
type Line struct {
	Text []rune
}

// Document is the in-memory text model: a slice of lines plus marks and tags.
// Multiple TextWidget instances may share one Document (peering); each
// registers a change listener that is called after every insert or delete.
type Document struct {
	Lines []*Line
	Marks map[string]*Mark
	Tags  map[string]*Tag

	// tagsByPriority orders Tags lowest priority first; a tag's priority is
	// its creation order, as in Tk, with "sel" fixed at selPriority.
	tagsByPriority []*Tag
	nextPriority   int

	// layoutGen counts tag configuration changes; a widget's layout cache
	// is keyed on it.
	layoutGen int

	// Listeners are called (in registration order) after every Insert,
	// Delete or layout-affecting tag change.
	Listeners []func(Change)

	// subs are the listeners added with Subscribe, called after Listeners.
	subs []*func(Change)

	// images are the embedded images, by the name of the mark just before
	// each one's placeholder character. They live in the document, so every
	// peer shows them, as in Tk.
	images   map[string]widget.WidgetImage
	imageSeq int

	// objects are the marks of embedded windows and images, each just
	// before its placeholder character; objectSubs are told when a Delete
	// removes one.
	objects    map[string]bool
	objectSubs []*func(name string)
}

// addObject records the embedded window or image whose placeholder follows
// the mark name.
func (d *Document) addObject(name string) {
	if d.objects == nil {
		d.objects = map[string]bool{}
	}
	d.objects[name] = true
}

// subscribeObjects registers fn to be called with the mark name of each
// embedded window or image a Delete removes.
func (d *Document) subscribeObjects(fn func(name string)) (unsubscribe func()) {
	p := &fn
	d.objectSubs = append(slices.Clip(d.objectSubs), p)
	return func() {
		if i := slices.Index(d.objectSubs, p); i >= 0 {
			d.objectSubs = slices.Delete(slices.Clone(d.objectSubs), i, i+1)
		}
	}
}

// objectsIn returns the embedded objects whose placeholders lie in
// [start, end).
func (d *Document) objectsIn(start, end Index) []string {
	var names []string
	for name := range d.objects {
		if m := d.Marks[name]; m != nil && Compare(m.Pos, start) >= 0 && Compare(m.Pos, end) < 0 {
			names = append(names, name)
		}
	}
	return names
}

// dropObjects forgets deleted embedded objects and tells the subscribers,
// as the B-tree calls a segment's deleteProc (EmbWinDeleteProc,
// EmbImageDeleteProc).
func (d *Document) dropObjects(names []string) {
	for _, name := range names {
		delete(d.objects, name)
		delete(d.images, name)
		delete(d.Marks, name)
		for _, fn := range d.objectSubs {
			(*fn)(name)
		}
	}
}

// isPlaceholder reports whether the character at pos is an embedded
// window's or image's placeholder.
func (d *Document) isPlaceholder(pos Index) bool {
	if pos.Line < 1 || pos.Line > len(d.Lines) {
		return false
	}
	text := d.Lines[pos.Line-1].Text
	return pos.Char >= 0 && pos.Char < len(text) && text[pos.Char] == runeEmbeddedWindow
}

// eachImage calls fn for every embedded image still in the text (its
// placeholder not deleted), with its position.
func (d *Document) eachImage(fn func(pos Index, img widget.WidgetImage)) {
	for name, img := range d.images {
		if m := d.MarkPos(name); m != nil && d.isPlaceholder(*m) {
			fn(*m, img)
		}
	}
}

// imageAt returns the embedded image at pos, or nil.
func (d *Document) imageAt(pos Index) widget.WidgetImage {
	for name, img := range d.images {
		if m := d.MarkPos(name); m != nil && *m == pos && d.isPlaceholder(pos) {
			return img
		}
	}
	return nil
}

// imageIndex returns the position of the embedded image named name, as
// the image's Name method reports it (Tk's "image name" index).
func (d *Document) imageIndex(name string) (Index, bool) {
	var found *Index
	d.eachImage(func(pos Index, img widget.WidgetImage) {
		if n, ok := img.(interface{ Name() string }); ok && n.Name() == name {
			if found == nil || Compare(pos, *found) < 0 {
				found = &pos
			}
		}
	})
	if found == nil {
		return Index{}, false
	}
	return *found, true
}

// notifyListeners calls all registered change listeners.
func (d *Document) notifyListeners(ch Change) {
	for _, fn := range d.Listeners {
		fn(ch)
	}
	for _, fn := range d.subs {
		(*fn)(ch)
	}
}

// Subscribe adds a change listener like Listeners and returns a function
// that removes it, for widgets (peers) that go away before the document.
func (d *Document) Subscribe(fn func(Change)) (unsubscribe func()) {
	p := &fn
	d.subs = append(slices.Clip(d.subs), p)
	return func() {
		if i := slices.Index(d.subs, p); i >= 0 {
			d.subs = slices.Delete(slices.Clone(d.subs), i, i+1)
		}
	}
}

// NewDocument creates a new empty document with one empty line and the
// standard built-in marks.
func NewDocument() *Document {
	doc := &Document{
		Lines: []*Line{{}},
		Marks: map[string]*Mark{
			"insert":  {Name: "insert", Pos: Index{1, 0}, Gravity: GravityRight},
			"current": {Name: "current", Pos: Index{1, 0}, Gravity: GravityLeft},
		},
		Tags: make(map[string]*Tag),
	}
	return doc
}

// LineCount returns the number of logical lines.
func (d *Document) LineCount() int {
	return len(d.Lines)
}

// EndIndex returns the index past the last character of the document.
func (d *Document) EndIndex() Index {
	n := len(d.Lines)
	return Index{Line: n, Char: len(d.Lines[n-1].Text)}
}

// Get returns the text between two indexes as a string, with newlines
// between logical lines.
func (d *Document) Get(start, end Index) string {
	start = Clamp(start, d)
	end = Clamp(end, d)
	if Compare(start, end) >= 0 {
		return ""
	}

	if start.Line == end.Line {
		line := d.Lines[start.Line-1].Text
		return string(line[start.Char:end.Char])
	}

	var b strings.Builder
	// First line.
	b.WriteString(string(d.Lines[start.Line-1].Text[start.Char:]))
	b.WriteByte('\n')
	// Middle lines.
	for l := start.Line + 1; l < end.Line; l++ {
		b.WriteString(string(d.Lines[l-1].Text))
		b.WriteByte('\n')
	}
	// Last line.
	b.WriteString(string(d.Lines[end.Line-1].Text[:end.Char]))
	return b.String()
}

// Insert inserts text at idx. The text may contain newlines, which split
// lines. Returns the index after the inserted text.
func (d *Document) Insert(idx Index, text string) Index {
	idx = Clamp(idx, d)
	if len(text) == 0 {
		return idx
	}

	runes := []rune(text)
	insertedLines := splitRunes(runes)
	newlines, lastLineLen := countNewlines(runes)

	// Adjust marks before the structural change.
	for _, m := range d.Marks {
		d.adjustMarkInsert(m, idx, newlines, lastLineLen)
	}
	for _, tg := range d.tagsByPriority {
		tg.adjustInsert(idx, newlines, lastLineLen)
	}

	line := d.Lines[idx.Line-1]

	if newlines == 0 {
		line.Text = slices.Insert(line.Text, idx.Char, insertedLines[0]...)
		result := Index{Line: idx.Line, Char: idx.Char + len(insertedLines[0])}
		d.notifyListeners(Change{From: idx.Line, To: idx.Line})
		return result
	}

	after := slices.Clone(line.Text[idx.Char:])
	line.Text = append(line.Text[:idx.Char], insertedLines[0]...)

	newLines := make([]*Line, newlines)
	for i := 1; i < len(insertedLines); i++ {
		newLines[i-1] = &Line{Text: insertedLines[i]}
	}
	lastNew := newLines[len(newLines)-1]
	endChar := len(lastNew.Text)
	lastNew.Text = append(lastNew.Text, after...)

	d.Lines = slices.Insert(d.Lines, idx.Line, newLines...)

	result := Index{Line: idx.Line + newlines, Char: endChar}
	d.notifyListeners(Change{From: idx.Line, To: idx.Line, Delta: newlines})
	return result
}

// Delete removes text between start and end.
func (d *Document) Delete(start, end Index) {
	start = Clamp(start, d)
	end = Clamp(end, d)
	if Compare(start, end) >= 0 {
		return
	}

	gone := d.objectsIn(start, end)

	// Adjust marks and tags.
	for _, m := range d.Marks {
		d.adjustMarkDelete(m, start, end)
	}
	for _, tg := range d.tagsByPriority {
		tg.adjustDelete(start, end)
	}

	if start.Line == end.Line {
		line := d.Lines[start.Line-1]
		line.Text = slices.Delete(line.Text, start.Char, end.Char)
		d.dropObjects(gone)
		d.notifyListeners(Change{From: start.Line, To: start.Line})
		return
	}

	// Multi-line delete: merge first and last lines, remove those between.
	firstLine := d.Lines[start.Line-1]
	lastLine := d.Lines[end.Line-1]
	firstLine.Text = append(firstLine.Text[:start.Char], lastLine.Text[end.Char:]...)
	d.Lines = slices.Delete(d.Lines, start.Line, end.Line)
	d.dropObjects(gone)
	d.notifyListeners(Change{From: start.Line, To: start.Line, Delta: start.Line - end.Line})
}

// --- Mark methods ---

// MarkSet sets a mark at the given position.
func (d *Document) MarkSet(name string, pos Index) {
	pos = Clamp(pos, d)
	if m, ok := d.Marks[name]; ok {
		m.Pos = pos
	} else {
		d.Marks[name] = &Mark{Name: name, Pos: pos, Gravity: GravityRight}
	}
}

// MarkUnset removes a mark. Built-in marks ("insert", "current") cannot be removed.
func (d *Document) MarkUnset(name string) {
	if name == "insert" || name == "current" {
		return
	}
	delete(d.Marks, name)
}

// MarkPos returns the position of a mark, or nil if not found.
func (d *Document) MarkPos(name string) *Index {
	if m, ok := d.Marks[name]; ok {
		pos := m.Pos
		return &pos
	}
	return nil
}

// MarkNames returns the names of all marks.
func (d *Document) MarkNames() []string {
	names := make([]string, 0, len(d.Marks))
	for name := range d.Marks {
		names = append(names, name)
	}
	return names
}

// --- Tag methods ---

// tag returns the named tag, creating it with the next priority.
func (d *Document) tag(name string) *Tag {
	tg, ok := d.Tags[name]
	if !ok {
		tg = &Tag{Name: name, Priority: d.nextPriority}
		d.Tags[name] = tg
	}
	if !tg.registered {
		d.putTag(tg)
	}
	return tg
}

// putTag installs tg in the document at its Priority.
func (d *Document) putTag(tg *Tag) {
	if old, ok := d.Tags[tg.Name]; ok && old.registered {
		old.registered = false
		d.tagsByPriority = slices.DeleteFunc(d.tagsByPriority, func(t *Tag) bool { return t == old })
	}
	d.Tags[tg.Name] = tg
	tg.registered = true
	if tg.Priority >= d.nextPriority && tg.Priority < selPriority {
		d.nextPriority = tg.Priority + 1
	}
	i := sort.Search(len(d.tagsByPriority), func(i int) bool {
		return d.tagsByPriority[i].Priority > tg.Priority
	})
	d.tagsByPriority = slices.Insert(d.tagsByPriority, i, tg)
}

// TagAdd adds a tag to the given range.
func (d *Document) TagAdd(tagName string, start, end Index) {
	d.tagAdd(tagName, start, end, false)
}

func (d *Document) tagAdd(tagName string, start, end Index, toEnd bool) {
	start = Clamp(start, d)
	end = Clamp(end, d)
	if Compare(start, end) >= 0 {
		return
	}
	tg := d.tag(tagName)
	tg.add(start, end, toEnd)
	if tg.affectsLayout() {
		d.notifyListeners(Change{From: start.Line, To: end.Line})
	}
}

// TagRemove removes a tag from the given range, trimming or splitting the
// ranges that overlap [start, end).
func (d *Document) TagRemove(tagName string, start, end Index) {
	tg, ok := d.Tags[tagName]
	if !ok {
		return
	}
	start, end = Clamp(start, d), Clamp(end, d)
	tg.remove(start, end)
	if tg.affectsLayout() {
		d.notifyListeners(Change{From: start.Line, To: end.Line})
	}
}

// TagConfigure configures a tag's display attributes.
func (d *Document) TagConfigure(tagName string, cache *color.Cache, reg *font.Registry, opts ...TagOption) {
	tag := d.tag(tagName)
	for _, opt := range opts {
		opt(cache, reg, tag)
	}
	d.layoutGen++
}

// TagRangesFor returns the ranges of a tag, sorted and disjoint. The
// result must not be modified.
func (d *Document) TagRangesFor(tagName string) []TagRange {
	if tg, ok := d.Tags[tagName]; ok {
		return tg.ranges
	}
	return nil
}

// TagsAt returns the tags active at a given index, sorted by priority (lowest first).
func (d *Document) TagsAt(idx Index) []*Tag {
	idx = Clamp(idx, d)
	var result []*Tag
	for _, tg := range d.tagsByPriority {
		if tg.covers(idx) {
			result = append(result, tg)
		}
	}
	return result
}

// tagsOnLine calls fn for every tag range touching the line, lowest
// priority tag first.
func (d *Document) tagsOnLine(line int, fn func(*Tag, TagRange)) {
	for _, tg := range d.tagsByPriority {
		for _, tr := range tg.rangesOnLine(line) {
			fn(tg, tr)
		}
	}
}

// --- Internal helpers ---

// countNewlines returns the number of newlines in runes and the number of
// runes after the last one.
func countNewlines(runes []rune) (newlines, lastLineLen int) {
	for _, r := range runes {
		if r == '\n' {
			newlines++
			lastLineLen = 0
		} else {
			lastLineLen++
		}
	}
	return newlines, lastLineLen
}

// adjustMarkInsert adjusts a mark's position after an insert at idx.
func (d *Document) adjustMarkInsert(m *Mark, idx Index, newlineCount, lastLineLen int) {
	cmp := Compare(m.Pos, idx)
	if cmp < 0 {
		return // mark is before insert point
	}
	if cmp == 0 && m.Gravity == GravityLeft {
		return // left-gravity mark stays
	}

	if newlineCount == 0 {
		if m.Pos.Line == idx.Line {
			m.Pos.Char += lastLineLen
		}
	} else {
		if m.Pos.Line == idx.Line {
			m.Pos.Char = m.Pos.Char - idx.Char + lastLineLen
		}
		m.Pos.Line += newlineCount
	}
}

// adjustMarkDelete adjusts a mark's position after a delete from start to end.
func (d *Document) adjustMarkDelete(m *Mark, start, end Index) {
	if Compare(m.Pos, start) <= 0 {
		return // mark is at or before delete start
	}
	if Compare(m.Pos, end) <= 0 {
		// Mark is within deleted range — move to start.
		m.Pos = start
		return
	}
	// Mark is after deleted range.
	if m.Pos.Line == end.Line {
		m.Pos.Char = start.Char + (m.Pos.Char - end.Char)
		m.Pos.Line = start.Line
	} else {
		m.Pos.Line -= end.Line - start.Line
	}
}

func adjustIdxInsert(pos, insertAt Index, newlines, lastLineLen int, leftGravity bool) Index {
	cmp := Compare(pos, insertAt)
	if leftGravity {
		// Left gravity: don't shift when insert is at exact position.
		if cmp <= 0 {
			return pos
		}
	} else {
		// Right gravity: shift when insert is at or after position.
		if cmp < 0 {
			return pos
		}
	}
	if newlines == 0 {
		if pos.Line == insertAt.Line {
			pos.Char += lastLineLen
		}
	} else {
		if pos.Line == insertAt.Line {
			pos.Char = pos.Char - insertAt.Char + lastLineLen
		}
		pos.Line += newlines
	}
	return pos
}

func adjustIdxDelete(pos, start, end Index) Index {
	if Compare(pos, start) <= 0 {
		return pos
	}
	if Compare(pos, end) <= 0 {
		return start
	}
	// pos is after end.
	if pos.Line == end.Line {
		pos.Char = start.Char + (pos.Char - end.Char)
		pos.Line = start.Line
	} else {
		pos.Line -= end.Line - start.Line
	}
	return pos
}

// splitRunes splits runes by newline into a slice of rune slices.
// "abc\ndef" → [['a','b','c'], ['d','e','f']].
func splitRunes(runes []rune) [][]rune {
	var result [][]rune
	start := 0
	for i, r := range runes {
		if r == '\n' {
			result = append(result, runes[start:i:i])
			start = i + 1
		}
	}
	result = append(result, runes[start:len(runes):len(runes)])
	return result
}
