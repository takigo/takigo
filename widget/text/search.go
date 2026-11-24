package text

// The search command of tk/generic/tkText.c (TextSearchCmd), matching line
// by line as Tk does without -nolinestop.

import (
	"regexp"
	"unicode/utf8"
)

// SearchOption configures Search.
type SearchOption func(*searchConfig)

type searchConfig struct {
	backwards, regexp, nocase bool
	stop                      func(t *TextWidget) (Index, bool)
}

// SearchBackwards searches towards the start of the text (-backwards).
func SearchBackwards() SearchOption { return func(c *searchConfig) { c.backwards = true } }

// SearchRegexp treats the pattern as a Go regular expression (-regexp).
func SearchRegexp() SearchOption { return func(c *searchConfig) { c.regexp = true } }

// SearchNoCase ignores case (-nocase).
func SearchNoCase() SearchOption { return func(c *searchConfig) { c.nocase = true } }

// SearchStop ends the search at index instead of wrapping around the text
// (the stopIndex argument): a forward search takes no match starting at or
// after it, a backward one none starting before it.
func SearchStop[I IndexSpec](index I) SearchOption {
	return func(c *searchConfig) {
		c.stop = func(t *TextWidget) (Index, bool) { return indexOf(t, index) }
	}
}

// Match is a search result: the matched text runs from Start to End.
type Match struct {
	Start, End Index
}

// Count is the number of characters matched (-count).
func (m Match) Count(t *TextWidget) int {
	return len([]rune(t.doc.Get(m.Start, m.End)))
}

// searchSpan is the part of a line a match may start in: [from, to).
type searchSpan struct {
	line, from, to int
}

// Search looks for pattern from index and returns the first match (the
// nearest one before index when searching backwards). Without SearchStop
// the search wraps around the text. Each match lies within one line.
func (t *TextWidget) Search[I IndexSpec](pattern string, index I, opts ...SearchOption) (Match, bool) {
	re, spans, ok := t.searchSetup(pattern, index, opts)
	if !ok {
		return Match{}, false
	}
	backwards := spans.backwards
	for _, sp := range spans.list {
		line := string(t.doc.Lines[sp.line-1].Text)
		var best *Match
		for _, loc := range re.FindAllStringIndex(line, -1) {
			if loc[0] == loc[1] {
				continue
			}
			s := utf8.RuneCountInString(line[:loc[0]])
			if s < sp.from || s >= sp.to {
				continue
			}
			m := Match{
				Start: Index{Line: sp.line, Char: s},
				End:   Index{Line: sp.line, Char: s + utf8.RuneCountInString(line[loc[0]:loc[1]])},
			}
			if !backwards {
				return m, true
			}
			best = &m
		}
		if best != nil {
			return *best, true
		}
	}
	return Match{}, false
}

// SearchAll returns every match from index to the end of the text and then
// from the start back to index (search -all), in that order.
func (t *TextWidget) SearchAll[I IndexSpec](pattern string, index I, opts ...SearchOption) []Match {
	re, spans, ok := t.searchSetup(pattern, index, opts)
	if !ok {
		return nil
	}
	var all []Match
	for _, sp := range spans.list {
		line := string(t.doc.Lines[sp.line-1].Text)
		var found []Match
		for _, loc := range re.FindAllStringIndex(line, -1) {
			if loc[0] == loc[1] {
				continue
			}
			s := utf8.RuneCountInString(line[:loc[0]])
			if s < sp.from || s >= sp.to {
				continue
			}
			found = append(found, Match{
				Start: Index{Line: sp.line, Char: s},
				End:   Index{Line: sp.line, Char: s + utf8.RuneCountInString(line[loc[0]:loc[1]])},
			})
		}
		if spans.backwards {
			for i, j := 0, len(found)-1; i < j; i, j = i+1, j-1 {
				found[i], found[j] = found[j], found[i]
			}
		}
		all = append(all, found...)
	}
	return all
}

type searchSpans struct {
	list      []searchSpan
	backwards bool
}

// searchSetup compiles the pattern and lists, in search order, the spans
// of the lines a match may start in.
func (t *TextWidget) searchSetup[I IndexSpec](pattern string, index I, opts []SearchOption) (*regexp.Regexp, searchSpans, bool) {
	var c searchConfig
	for _, opt := range opts {
		opt(&c)
	}
	if pattern == "" {
		return nil, searchSpans{}, false
	}
	start, ok := indexOf(t, index)
	if !ok {
		return nil, searchSpans{}, false
	}
	expr := pattern
	if !c.regexp {
		expr = regexp.QuoteMeta(pattern)
	}
	if c.nocase {
		expr = "(?i)" + expr
	}
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, searchSpans{}, false
	}
	n := t.doc.LineCount()
	lineLen := func(l int) int { return len(t.doc.Lines[l-1].Text) + 1 }
	var stop *Index
	if c.stop != nil {
		s, ok := c.stop(t)
		if !ok {
			return nil, searchSpans{}, false
		}
		stop = &s
	}
	var list []searchSpan
	if !c.backwards {
		list = append(list, searchSpan{start.Line, start.Char, lineLen(start.Line)})
		for l := start.Line + 1; l <= n; l++ {
			list = append(list, searchSpan{l, 0, lineLen(l)})
		}
		if stop == nil {
			for l := 1; l < start.Line; l++ {
				list = append(list, searchSpan{l, 0, lineLen(l)})
			}
			list = append(list, searchSpan{start.Line, 0, start.Char})
		} else {
			list = clipSpans(list, func(sp *searchSpan) bool {
				if sp.line > stop.Line {
					return false
				}
				if sp.line == stop.Line {
					sp.to = min(sp.to, stop.Char)
				}
				return sp.from < sp.to
			})
		}
	} else {
		list = append(list, searchSpan{start.Line, 0, start.Char})
		for l := start.Line - 1; l >= 1; l-- {
			list = append(list, searchSpan{l, 0, lineLen(l)})
		}
		if stop == nil {
			for l := n; l > start.Line; l-- {
				list = append(list, searchSpan{l, 0, lineLen(l)})
			}
			list = append(list, searchSpan{start.Line, start.Char, lineLen(start.Line)})
		} else {
			list = clipSpans(list, func(sp *searchSpan) bool {
				if sp.line < stop.Line {
					return false
				}
				if sp.line == stop.Line {
					sp.from = max(sp.from, stop.Char)
				}
				return sp.from < sp.to
			})
		}
	}
	return re, searchSpans{list: list, backwards: c.backwards}, true
}

// clipSpans keeps the spans keep accepts; keep may narrow them.
func clipSpans(list []searchSpan, keep func(*searchSpan) bool) []searchSpan {
	out := list[:0]
	for _, sp := range list {
		if keep(&sp) {
			out = append(out, sp)
		}
	}
	return out
}
