package text

import (
	"strconv"
	"strings"
	"unicode"
)

// This file ports the index grammar of tk/generic/tkTextIndex.c (GetIndex,
// ForwBack, StartEnd): a base followed by any number of modifiers.
//
//	base:     line.char | line.end | end | @x,y | mark | tag.first |
//	          tag.last | embedded window path | embedded image name
//	modifier: +/- count ?any|display? chars|indices|lines
//	          ?display? linestart|lineend|wordstart|wordend
//
// Tk's text always ends with a newline and its "end" index lies after it,
// so "end -1c" is the end of the content. A Document has no final newline;
// while modifiers are applied "end" is the position after an implicit one,
// Index{LineCount()+1, 0}, and the result is clamped into the document, so
// "end" and "end -1c" both give the end of the text, as they address the
// same content in Tk.

// index resolves an index string for this widget.
func (t *TextWidget) index(spec string) (Index, bool) {
	return parseIndex(t.doc, t, spec)
}

// parseIndex resolves spec in doc. t supplies what needs a widget: "@x,y",
// embedded window names and the display modifiers; it may be nil.
func parseIndex(doc *Document, t *TextWidget, spec string) (Index, bool) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return Index{}, false
	}
	// A mark name on its own may contain spaces, "+" or "-".
	if m := doc.MarkPos(spec); m != nil {
		return *m, true
	}
	idx, rest, ok := parseBase(doc, t, spec)
	if !ok {
		return Index{}, false
	}
	for {
		rest = strings.TrimLeft(rest, " \t\n")
		if rest == "" {
			break
		}
		if rest[0] == '+' || rest[0] == '-' {
			idx, rest, ok = forwBack(doc, t, idx, rest)
		} else {
			idx, rest, ok = startEnd(doc, t, idx, rest)
		}
		if !ok {
			return Index{}, false
		}
	}
	return Clamp(idx, doc), true
}

// virtualEnd is Tk's "end": the start of the line after the last one.
func virtualEnd(doc *Document) Index { return Index{Line: doc.LineCount() + 1} }

func parseBase(doc *Document, t *TextWidget, spec string) (Index, string, bool) {
	// tag.first / tag.last, taking the last "." so tag names may contain
	// dots and characters such as "+1c".
	if dot := strings.LastIndexByte(spec, '.'); dot > 0 {
		after := spec[dot+1:]
		wantLast, n := false, 0
		switch {
		case strings.HasPrefix(after, "first"):
			n = len("first")
		case strings.HasPrefix(after, "last"):
			wantLast, n = true, len("last")
		}
		if n > 0 {
			if _, ok := doc.Tags[spec[:dot]]; ok {
				ranges := doc.TagRangesFor(spec[:dot])
				if len(ranges) == 0 {
					return Index{}, "", false
				}
				idx := ranges[0].Start
				if wantLast {
					idx = ranges[0].End
					for _, r := range ranges[1:] {
						if Compare(r.End, idx) > 0 {
							idx = r.End
						}
					}
				} else {
					for _, r := range ranges[1:] {
						if Compare(r.Start, idx) < 0 {
							idx = r.Start
						}
					}
				}
				return idx, spec[dot+1+n:], true
			}
		}
	}

	if spec[0] == '@' {
		if t == nil {
			return Index{}, "", false
		}
		x, rest, ok := leadingInt(spec[1:])
		if !ok || rest == "" || rest[0] != ',' {
			return Index{}, "", false
		}
		y, rest, ok := leadingInt(rest[1:])
		if !ok {
			return Index{}, "", false
		}
		return t.indexFromPixel(x, y), rest, true
	}

	if spec[0] >= '0' && spec[0] <= '9' {
		line, rest, ok := leadingInt(spec)
		if !ok || rest == "" || rest[0] != '.' {
			return Index{}, "", false
		}
		rest = rest[1:]
		if line < 1 {
			return Index{Line: 1}, skipCharPart(rest), true
		}
		if line > doc.LineCount() {
			return virtualEnd(doc), skipCharPart(rest), true
		}
		if strings.HasPrefix(rest, "end") {
			return LineEnd(line, doc), rest[len("end"):], true
		}
		ch, rest, ok := leadingInt(rest)
		if !ok {
			return Index{}, "", false
		}
		return Clamp(Index{Line: line, Char: max(ch, 0)}, doc), rest, true
	}

	// Any other base runs up to white space, "+" or "-".
	end := strings.IndexAny(spec, " \t\n+-")
	if end < 0 {
		end = len(spec)
	}
	name, rest := spec[:end], spec[end:]
	if name == "" {
		return Index{}, "", false
	}
	if name[0] == '.' && t != nil {
		if idx, ok := t.windowIndex(name); ok {
			return idx, rest, true
		}
	}
	if strings.HasPrefix("end", name) && name[0] == 'e' {
		return virtualEnd(doc), rest, true
	}
	if m := doc.MarkPos(name); m != nil {
		return *m, rest, true
	}
	if idx, ok := doc.imageIndex(name); ok {
		return idx, rest, true
	}
	return Index{}, "", false
}

// skipCharPart skips the character part of a line.char base whose line is
// out of range, so the modifiers after it still parse.
func skipCharPart(s string) string {
	if strings.HasPrefix(s, "end") {
		return s[len("end"):]
	}
	if _, rest, ok := leadingInt(s); ok {
		return rest
	}
	return s
}

// leadingInt parses an optionally signed decimal integer at the start of s,
// like strtol, and returns the rest of s.
func leadingInt(s string) (int, string, bool) {
	i := 0
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		i++
	}
	j := i
	for j < len(s) && s[j] >= '0' && s[j] <= '9' {
		j++
	}
	if j == i {
		return 0, s, false
	}
	n, err := strconv.Atoi(s[:j])
	if err != nil {
		return 0, s, false
	}
	return n, s[j:], true
}

// modifierWord splits off a leading "display" or "any" qualifier, which may
// be abbreviated as a word of its own ("d lines") or run into the unit
// ("displaylines"). It returns the qualifier and the unit word.
func modifierWord(s string) (qual, unit, rest string) {
	word, rest := nextWord(s)
	for _, q := range []string{"display", "any"} {
		if word == "" || word[0] != q[0] {
			continue
		}
		if len(word) > len(q) {
			if strings.HasPrefix(word, q) {
				return q, word[len(q):], rest
			}
			continue
		}
		if strings.HasPrefix(q, word) {
			unit, rest = nextWord(strings.TrimLeft(rest, " \t\n"))
			return q, unit, rest
		}
	}
	return "", word, rest
}

// nextWord splits s at the first white space, "+" or "-".
func nextWord(s string) (string, string) {
	end := strings.IndexAny(s, " \t\n+-")
	if end < 0 {
		return s, ""
	}
	return s[:end], s[end:]
}

// abbrev reports whether word is a non-empty prefix of full of at least
// minLen bytes.
func abbrev(word, full string, minLen int) bool {
	return len(word) >= max(minLen, 1) && strings.HasPrefix(full, word)
}

// forwBack applies a "+/- count units" modifier (Tk's ForwBack).
func forwBack(doc *Document, t *TextWidget, idx Index, s string) (Index, string, bool) {
	forward := s[0] == '+'
	count, rest, ok := leadingInt(strings.TrimLeft(s[1:], " \t\n"))
	if !ok {
		return idx, "", false
	}
	qual, unit, rest := modifierWord(strings.TrimLeft(rest, " \t\n"))
	if !forward {
		count = -count
	}
	switch {
	case abbrev(unit, "chars", 1), abbrev(unit, "indices", 1):
		// No elided text here: display, any and plain chars coincide.
		return moveChars(doc, idx, count), rest, true
	case abbrev(unit, "lines", 1):
		if qual == "display" {
			if t == nil {
				return idx, "", false
			}
			return t.moveDisplayLines(idx, count), rest, true
		}
		line := max(idx.Line+count, 1)
		if line > doc.LineCount() {
			return virtualEnd(doc), rest, true
		}
		return Clamp(Index{Line: line, Char: idx.Char}, doc), rest, true
	}
	return idx, "", false
}

// moveChars moves idx by count characters (backward when negative),
// counting each line break as one. Moving past the last character reaches
// the virtual end, the implicit final newline.
func moveChars(doc *Document, idx Index, count int) Index {
	n := doc.LineCount()
	if count < 0 {
		if idx.Line > n {
			idx, count = LineEnd(n, doc), count+1 // back over the implicit newline
		}
		return Backward(idx, -count, doc)
	}
	if idx.Line > n {
		return idx
	}
	for {
		left := len(doc.Lines[idx.Line-1].Text) - idx.Char
		if count <= left {
			idx.Char += count
			return idx
		}
		count -= left + 1 // the rest of the line and its line break
		if idx.Line == n {
			return virtualEnd(doc)
		}
		idx = Index{Line: idx.Line + 1}
	}
}

// startEnd applies a linestart/lineend/wordstart/wordend modifier (Tk's
// StartEnd).
func startEnd(doc *Document, t *TextWidget, idx Index, s string) (Index, string, bool) {
	end := 0
	for end < len(s) && (unicode.IsLetter(rune(s[end])) || unicode.IsDigit(rune(s[end]))) {
		end++
	}
	qual, unit, rest := modifierWord(s)
	if qual == "" {
		unit, rest = s[:end], s[end:]
	}
	if idx.Line > doc.LineCount() {
		// The virtual end is the start of an empty line after the text.
		if abbrev(unit, "lineend", 5) || abbrev(unit, "linestart", 5) ||
			abbrev(unit, "wordend", 5) || abbrev(unit, "wordstart", 5) {
			return idx, rest, true
		}
		return idx, "", false
	}
	switch {
	case abbrev(unit, "lineend", 5):
		if qual == "display" && t != nil {
			return t.displayLineEnd(idx), rest, true
		}
		return LineEnd(idx.Line, doc), rest, true
	case abbrev(unit, "linestart", 5):
		if qual == "display" && t != nil {
			return t.displayLineStart(idx), rest, true
		}
		return LineStart(idx.Line), rest, true
	case abbrev(unit, "wordend", 5):
		return tkWordEnd(doc, idx), rest, true
	case abbrev(unit, "wordstart", 5):
		return tkWordStart(doc, idx), rest, true
	}
	return idx, "", false
}

// tkWordChar is Tcl_UniCharIsWordChar: a letter, digit or connector
// punctuation such as "_".
func tkWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.Is(unicode.Pc, r)
}

// tkWordEnd is "wordend": past the word at idx, or one character on when
// idx is not in a word.
func tkWordEnd(doc *Document, idx Index) Index {
	line := doc.Lines[idx.Line-1].Text
	pos := idx.Char
	for pos < len(line) && tkWordChar(line[pos]) {
		pos++
	}
	if pos == idx.Char {
		return moveChars(doc, idx, 1)
	}
	return Index{Line: idx.Line, Char: pos}
}

// tkWordStart is "wordstart": the start of the word at idx, or idx itself
// when it is not in a word.
func tkWordStart(doc *Document, idx Index) Index {
	line := doc.Lines[idx.Line-1].Text
	pos := idx.Char
	if pos >= len(line) || !tkWordChar(line[pos]) {
		return idx
	}
	for pos > 0 && tkWordChar(line[pos-1]) {
		pos--
	}
	return Index{Line: idx.Line, Char: pos}
}

// windowIndex returns the position of the embedded window named path.
func (t *TextWidget) windowIndex(path string) (Index, bool) {
	for _, ew := range t.embeddedWindows {
		if ew.win != nil && ew.win.PathName == path {
			if m := t.doc.MarkPos(ew.markName); m != nil {
				return *m, true
			}
		}
	}
	return Index{}, false
}

// displayLineAt returns the display lines of idx's logical line and the
// number of the one holding idx.
func (t *TextWidget) displayLineAt(idx Index) ([]displayLine, int) {
	dls := t.layout.line(idx.Line).dls
	for k := len(dls) - 1; k > 0; k-- {
		if idx.Char >= dls[k].startChar {
			return dls, k
		}
	}
	return dls, 0
}

// displayLineStart is "display linestart".
func (t *TextWidget) displayLineStart(idx Index) Index {
	dls, k := t.displayLineAt(idx)
	if len(dls) == 0 {
		return LineStart(idx.Line)
	}
	return Index{Line: idx.Line, Char: dls[k].startChar}
}

// displayLineEnd is "display lineend": the last character of a wrapped
// display line, or the end of the logical line on its last display line.
func (t *TextWidget) displayLineEnd(idx Index) Index {
	dls, k := t.displayLineAt(idx)
	if len(dls) == 0 || k == len(dls)-1 {
		return LineEnd(idx.Line, t.doc)
	}
	return Index{Line: idx.Line, Char: max(dls[k].endChar-1, dls[k].startChar)}
}

// moveDisplayLines is "+/- count display lines": the character at the same
// x position count display lines away.
func (t *TextWidget) moveDisplayLines(idx Index, count int) Index {
	if idx.Line > t.doc.LineCount() {
		idx = LineEnd(t.doc.LineCount(), t.doc)
	}
	dls, k := t.displayLineAt(idx)
	if len(dls) == 0 {
		return idx
	}
	x := t.measureRange(idx.Line, dls[k].startChar, idx.Char)
	n := t.layout.displayLinesBefore(idx.Line) + k + count
	if n < 0 {
		n, x = 0, 0 // before the text: the first display line's start (Tk)
	}
	n = min(n, t.layout.totalDisplayLines()-1)
	line, off := t.layout.lineAtDisplayLine(n)
	target := t.layout.line(line).dls
	if off >= len(target) {
		return LineEnd(line, t.doc)
	}
	dl := target[off]
	last := dl.endChar
	if off < len(target)-1 {
		last = max(dl.endChar-1, dl.startChar)
	}
	for c := dl.startChar; c < last; c++ {
		if t.measureRange(line, dl.startChar, c+1) > x {
			return Index{Line: line, Char: c}
		}
	}
	return Index{Line: line, Char: last}
}
