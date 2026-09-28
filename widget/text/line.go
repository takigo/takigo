package text

import (
	"slices"
	"strings"

	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/font"
)

// Line holds one logical line of text as a rune slice.
type Line struct {
	Text []rune
}

// Document is the in-memory text model: a slice of lines plus marks and tags.
// Multiple TextWidget instances may share one Document (peering); each
// registers a change listener that is called after every insert or delete.
type Document struct {
	Lines     []*Line
	Marks     map[string]*Mark
	Tags      map[string]*Tag
	TagRanges []TagRange

	// Listeners are called (in registration order) after every Insert or Delete.
	Listeners []func()
}

// notifyListeners calls all registered change listeners.
func (d *Document) notifyListeners() {
	for _, fn := range d.Listeners {
		fn()
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
		Tags:      make(map[string]*Tag),
		TagRanges: nil,
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
	d.adjustTagRangesInsert(idx, newlines, lastLineLen)

	line := d.Lines[idx.Line-1]

	if newlines == 0 {
		line.Text = slices.Insert(line.Text, idx.Char, insertedLines[0]...)
		result := Index{Line: idx.Line, Char: idx.Char + len(insertedLines[0])}
		d.notifyListeners()
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
	d.notifyListeners()
	return result
}

// Delete removes text between start and end.
func (d *Document) Delete(start, end Index) {
	start = Clamp(start, d)
	end = Clamp(end, d)
	if Compare(start, end) >= 0 {
		return
	}

	// Adjust marks and tags.
	for _, m := range d.Marks {
		d.adjustMarkDelete(m, start, end)
	}
	d.adjustTagRangesDelete(start, end)

	if start.Line == end.Line {
		line := d.Lines[start.Line-1]
		line.Text = slices.Delete(line.Text, start.Char, end.Char)
		d.notifyListeners()
		return
	}

	// Multi-line delete: merge first and last lines, remove those between.
	firstLine := d.Lines[start.Line-1]
	lastLine := d.Lines[end.Line-1]
	firstLine.Text = append(firstLine.Text[:start.Char], lastLine.Text[end.Char:]...)
	d.Lines = slices.Delete(d.Lines, start.Line, end.Line)
	d.notifyListeners()
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

// TagAdd adds a tag to the given range.
func (d *Document) TagAdd(tagName string, start, end Index) {
	start = Clamp(start, d)
	end = Clamp(end, d)
	if Compare(start, end) >= 0 {
		return
	}
	// Ensure the tag exists.
	if _, ok := d.Tags[tagName]; !ok {
		d.Tags[tagName] = &Tag{Name: tagName}
	}
	d.TagRanges = append(d.TagRanges, TagRange{TagName: tagName, Start: start, End: end})
}

// TagRemove removes a tag from the given range. This removes any tag ranges
// that overlap with [start, end).
func (d *Document) TagRemove(tagName string, start, end Index) {
	start = Clamp(start, d)
	end = Clamp(end, d)
	n := 0
	for _, tr := range d.TagRanges {
		if tr.TagName == tagName && Compare(tr.Start, end) < 0 && Compare(tr.End, start) > 0 {
			continue // remove this range
		}
		d.TagRanges[n] = tr
		n++
	}
	d.TagRanges = d.TagRanges[:n]
}

// TagConfigure configures a tag's display attributes.
func (d *Document) TagConfigure(tagName string, cache *color.Cache, reg *font.Registry, opts ...TagOption) {
	tag, ok := d.Tags[tagName]
	if !ok {
		tag = &Tag{Name: tagName}
		d.Tags[tagName] = tag
	}
	for _, opt := range opts {
		opt(cache, reg, tag)
	}
}

// TagRangesFor returns all ranges for a given tag name.
func (d *Document) TagRangesFor(tagName string) []TagRange {
	var result []TagRange
	for _, tr := range d.TagRanges {
		if tr.TagName == tagName {
			result = append(result, tr)
		}
	}
	return result
}

// TagsAt returns the tags active at a given index, sorted by priority (lowest first).
func (d *Document) TagsAt(idx Index) []*Tag {
	idx = Clamp(idx, d)
	var result []*Tag
	seen := make(map[string]bool)
	for _, tr := range d.TagRanges {
		if seen[tr.TagName] {
			continue
		}
		if Compare(idx, tr.Start) >= 0 && Compare(idx, tr.End) < 0 {
			if tag, ok := d.Tags[tr.TagName]; ok {
				result = append(result, tag)
				seen[tr.TagName] = true
			}
		}
	}
	// Sort by priority (stable, ascending).
	for i := 1; i < len(result); i++ {
		for j := i; j > 0 && result[j].Priority < result[j-1].Priority; j-- {
			result[j], result[j-1] = result[j-1], result[j]
		}
	}
	return result
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

// adjustTagRangesInsert adjusts all tag ranges after an insert.
func (d *Document) adjustTagRangesInsert(idx Index, newlineCount, lastLineLen int) {
	for i := range d.TagRanges {
		tr := &d.TagRanges[i]
		// Tag Start has right gravity: shifts when insert is at or after Start.
		// Tag End has left gravity: shifts only when insert is strictly before End.
		// This matches Tk's behavior where inserting at a tag's End does NOT
		// expand the tag to cover the newly inserted text.
		tr.Start = adjustIdxInsert(tr.Start, idx, newlineCount, lastLineLen, false)
		tr.End = adjustIdxInsert(tr.End, idx, newlineCount, lastLineLen, !tr.ToEnd)
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

// adjustTagRangesDelete adjusts all tag ranges after a delete.
func (d *Document) adjustTagRangesDelete(start, end Index) {
	n := 0
	for _, tr := range d.TagRanges {
		tr.Start = adjustIdxDelete(tr.Start, start, end)
		tr.End = adjustIdxDelete(tr.End, start, end)
		if Compare(tr.Start, tr.End) < 0 {
			d.TagRanges[n] = tr
			n++
		}
	}
	d.TagRanges = d.TagRanges[:n]
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
// "abc\ndef" → [['a','b','c'], ['d','e','f']]
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
