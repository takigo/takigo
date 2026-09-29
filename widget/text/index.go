// Package text implements a multi-line text editor widget.
// It ports a practical subset of tk/generic/tkText*.c.
package text

import (
	"strconv"
	"strings"
	"unicode"
)

// Index represents a position in the text document.
// Line is 1-based (Tk convention). Char is 0-based rune offset within the line.
type Index struct {
	Line int
	Char int
}

// Compare returns -1 if a < b, 0 if a == b, 1 if a > b.
func Compare(a, b Index) int {
	if a.Line < b.Line {
		return -1
	}
	if a.Line > b.Line {
		return 1
	}
	if a.Char < b.Char {
		return -1
	}
	if a.Char > b.Char {
		return 1
	}
	return 0
}

// Clamp adjusts idx to be within the valid range of doc.
func Clamp(idx Index, doc *Document) Index {
	if idx.Line < 1 {
		idx.Line = 1
		idx.Char = 0
		return idx
	}
	if idx.Line > doc.LineCount() {
		idx.Line = doc.LineCount()
		idx.Char = len(doc.Lines[idx.Line-1].Text)
		return idx
	}
	lineLen := len(doc.Lines[idx.Line-1].Text)
	if idx.Char < 0 {
		idx.Char = 0
	}
	if idx.Char > lineLen {
		idx.Char = lineLen
	}
	return idx
}

// LineStart returns the start of the given line.
func LineStart(line int) Index {
	return Index{Line: line, Char: 0}
}

// LineEnd returns the end of the given line in the document.
func LineEnd(line int, doc *Document) Index {
	if line < 1 || line > doc.LineCount() {
		return Clamp(Index{Line: line, Char: 0}, doc)
	}
	return Index{Line: line, Char: len(doc.Lines[line-1].Text)}
}

// Forward moves an index forward by count runes, crossing line boundaries.
func Forward(idx Index, count int, doc *Document) Index {
	for count > 0 && (idx.Line < doc.LineCount() || idx.Char < len(doc.Lines[idx.Line-1].Text)) {
		lineLen := len(doc.Lines[idx.Line-1].Text)
		remaining := lineLen - idx.Char
		if count <= remaining {
			idx.Char += count
			return idx
		}
		count -= remaining + 1 // +1 for the implicit newline
		if idx.Line < doc.LineCount() {
			idx.Line++
			idx.Char = 0
		} else {
			idx.Char = lineLen
			return idx
		}
	}
	return Clamp(idx, doc)
}

// Backward moves an index backward by count runes, crossing line boundaries.
func Backward(idx Index, count int, doc *Document) Index {
	for count > 0 && (idx.Line > 1 || idx.Char > 0) {
		if count <= idx.Char {
			idx.Char -= count
			return idx
		}
		count -= idx.Char + 1 // +1 for the implicit newline
		if idx.Line > 1 {
			idx.Line--
			idx.Char = len(doc.Lines[idx.Line-1].Text)
		} else {
			idx.Char = 0
			return idx
		}
	}
	return Clamp(idx, doc)
}

// WordStart returns the index of the start of the word at or before idx.
func WordStart(idx Index, doc *Document) Index {
	idx = Clamp(idx, doc)
	line := doc.Lines[idx.Line-1].Text
	pos := min(idx.Char, len(line))
	// Skip back past non-word chars.
	for pos > 0 && !isWordRune(line[pos-1]) {
		pos--
	}
	// Skip back past word chars.
	for pos > 0 && isWordRune(line[pos-1]) {
		pos--
	}
	return Index{Line: idx.Line, Char: pos}
}

// WordEnd returns the index past the end of the word at or after idx.
func WordEnd(idx Index, doc *Document) Index {
	idx = Clamp(idx, doc)
	line := doc.Lines[idx.Line-1].Text
	pos := idx.Char
	// Skip past word chars.
	for pos < len(line) && isWordRune(line[pos]) {
		pos++
	}
	// Skip past non-word chars.
	for pos < len(line) && !isWordRune(line[pos]) {
		pos++
	}
	return Index{Line: idx.Line, Char: pos}
}

// UpLine moves the index up by one line, preserving character offset.
func UpLine(idx Index, doc *Document) Index {
	if idx.Line <= 1 {
		return Index{Line: 1, Char: 0}
	}
	idx.Line--
	lineLen := len(doc.Lines[idx.Line-1].Text)
	if idx.Char > lineLen {
		idx.Char = lineLen
	}
	return idx
}

// DownLine moves the index down by one line, preserving character offset.
func DownLine(idx Index, doc *Document) Index {
	if idx.Line >= doc.LineCount() {
		return LineEnd(doc.LineCount(), doc)
	}
	idx.Line++
	lineLen := len(doc.Lines[idx.Line-1].Text)
	if idx.Char > lineLen {
		idx.Char = lineLen
	}
	return idx
}

// ParseIndex parses an index spec string. Supports "line.char", "end",
// "insert", "sel.first", "sel.last", and mark names.
func ParseIndex(doc *Document, spec string) (Index, bool) {
	spec = strings.TrimSpace(spec)

	switch spec {
	case "end":
		return doc.EndIndex(), true
	case "insert":
		if m := doc.MarkPos("insert"); m != nil {
			return *m, true
		}
		return Index{1, 0}, true
	case "sel.first":
		ranges := doc.TagRangesFor("sel")
		if len(ranges) > 0 {
			return ranges[0].Start, true
		}
		return Index{}, false
	case "sel.last":
		ranges := doc.TagRangesFor("sel")
		if len(ranges) > 0 {
			return ranges[len(ranges)-1].End, true
		}
		return Index{}, false
	}

	// Try "line.char" format.
	if dot := strings.IndexByte(spec, '.'); dot >= 0 {
		lineStr := spec[:dot]
		charStr := spec[dot+1:]
		line, err1 := strconv.Atoi(lineStr)
		if err1 != nil {
			return Index{}, false
		}
		if charStr == "end" {
			idx := LineEnd(line, doc)
			return idx, true
		}
		ch, err2 := strconv.Atoi(charStr)
		if err2 != nil {
			return Index{}, false
		}
		return Clamp(Index{Line: line, Char: ch}, doc), true
	}

	// Try mark name.
	if pos := doc.MarkPos(spec); pos != nil {
		return *pos, true
	}

	return Index{}, false
}

func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}
