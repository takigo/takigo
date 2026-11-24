package richtext

import (
	"slices"
	"strings"
)

// declaration is one "property: value" pair.
type declaration struct {
	prop, value string
	important   bool
}

// compound is one simple selector sequence such as "a.ext#top".
type compound struct {
	tag     string // "" or "*" matches any element
	id      string
	classes []string
	pseudo  string // "hover" for :hover; other pseudo-classes never match
}

// selector is a chain of compounds joined by combinators, rightmost last.
// combinators[i] joins parts[i] and parts[i+1]: ' ' descendant, '>' child.
type selector struct {
	parts       []compound
	combinators []byte
	spec        [3]int // ids, classes and pseudo-classes, elements
}

// rule is a selector with its declarations; order is its position in the
// cascade.
type rule struct {
	sel    selector
	decls  []declaration
	origin int // 0 the default stylesheet, 1 author styles
	order  int
}

// stylesheet is an ordered list of rules.
type stylesheet struct {
	rules []rule
}

// parseCSS adds the rules of css to the stylesheet. At-rules are skipped
// with their blocks; selectors it cannot read are dropped.
func (ss *stylesheet) parseCSS(css string, origin int) {
	css = stripComments(css)
	for {
		css = strings.TrimSpace(css)
		if css == "" {
			return
		}
		if css[0] == '@' {
			css = skipAtRule(css)
			continue
		}
		open := strings.IndexByte(css, '{')
		if open < 0 {
			return
		}
		closeAt := strings.IndexByte(css[open:], '}')
		if closeAt < 0 {
			closeAt = len(css) - open
		}
		sels, body := css[:open], css[open+1:open+closeAt]
		decls := parseDeclarations(body)
		for s := range strings.SplitSeq(sels, ",") {
			if sel, ok := parseSelector(strings.TrimSpace(s)); ok {
				ss.rules = append(ss.rules, rule{sel: sel, decls: decls, origin: origin, order: len(ss.rules)})
			}
		}
		if open+closeAt+1 > len(css) {
			return
		}
		css = css[open+closeAt+1:]
	}
}

func stripComments(s string) string {
	for {
		a := strings.Index(s, "/*")
		if a < 0 {
			return s
		}
		b := strings.Index(s[a+2:], "*/")
		if b < 0 {
			return s[:a]
		}
		s = s[:a] + " " + s[a+2+b+2:]
	}
}

// skipAtRule drops an at-rule: up to ';' or through its balanced block.
func skipAtRule(s string) string {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case ';':
			return s[i+1:]
		case '{':
			depth := 0
			for j := i; j < len(s); j++ {
				switch s[j] {
				case '{':
					depth++
				case '}':
					depth--
					if depth == 0 {
						return s[j+1:]
					}
				}
			}
			return ""
		}
	}
	return ""
}

// parseDeclarations reads "a: b; c: d" (a style attribute or rule body).
func parseDeclarations(body string) []declaration {
	var decls []declaration
	for part := range strings.SplitSeq(body, ";") {
		prop, value, ok := strings.Cut(part, ":")
		if !ok {
			continue
		}
		prop = strings.ToLower(strings.TrimSpace(prop))
		value = strings.TrimSpace(value)
		important := false
		if v, ok := strings.CutSuffix(value, "!important"); ok {
			value, important = strings.TrimSpace(v), true
		}
		if prop == "" || value == "" {
			continue
		}
		decls = append(decls, declaration{prop: prop, value: value, important: important})
	}
	return decls
}

// parseSelector reads a selector of type, class, id and :hover parts
// joined by descendant and child combinators.
func parseSelector(s string) (selector, bool) {
	var sel selector
	s = strings.ReplaceAll(s, ">", " > ")
	pendingChild := false
	for tok := range strings.FieldsSeq(s) {
		if tok == ">" {
			if len(sel.parts) == 0 || pendingChild {
				return sel, false
			}
			pendingChild = true
			continue
		}
		c, ok := parseCompound(tok)
		if !ok {
			return sel, false
		}
		if len(sel.parts) > 0 {
			comb := byte(' ')
			if pendingChild {
				comb = '>'
			}
			sel.combinators = append(sel.combinators, comb)
		}
		pendingChild = false
		sel.parts = append(sel.parts, c)
		if c.id != "" {
			sel.spec[0]++
		}
		sel.spec[1] += len(c.classes)
		if c.pseudo != "" {
			sel.spec[1]++
		}
		if c.tag != "" && c.tag != "*" {
			sel.spec[2]++
		}
	}
	return sel, len(sel.parts) > 0 && !pendingChild
}

func parseCompound(tok string) (compound, bool) {
	var c compound
	i := 0
	readIdent := func() string {
		start := i
		for i < len(tok) && tok[i] != '.' && tok[i] != '#' && tok[i] != ':' && tok[i] != '[' {
			i++
		}
		return tok[start:i]
	}
	c.tag = strings.ToLower(readIdent())
	for i < len(tok) {
		switch tok[i] {
		case '.':
			i++
			if cl := readIdent(); cl != "" {
				c.classes = append(c.classes, cl)
			} else {
				return c, false
			}
		case '#':
			i++
			c.id = readIdent()
		case ':':
			i++
			if i < len(tok) && tok[i] == ':' {
				return c, false // pseudo-elements
			}
			c.pseudo = strings.ToLower(readIdent())
			if c.pseudo != "hover" && c.pseudo != "link" && c.pseudo != "first-child" {
				return c, false
			}
		default:
			return c, false // attribute selectors
		}
	}
	return c, true
}

// matches reports whether the compound matches the element; hover tells
// whether :hover parts may match.
func (c *compound) matches(n *node, hover bool) bool {
	if n == nil || n.kind != elementNode || n.tag == "" {
		return false
	}
	if c.tag != "" && c.tag != "*" && c.tag != n.tag {
		return false
	}
	if c.id != "" && c.id != n.id {
		return false
	}
	for _, cl := range c.classes {
		if !slices.Contains(n.classes, cl) {
			return false
		}
	}
	switch c.pseudo {
	case "hover":
		return hover
	case "link":
		_, ok := n.attr("href")
		return n.tag == "a" && ok
	case "first-child":
		for _, s := range n.parent.children {
			if s.kind == elementNode {
				return s == n
			}
		}
	}
	return true
}

// matches reports whether the selector matches the element.
func (s *selector) matches(n *node, hover bool) bool {
	last := len(s.parts) - 1
	if !s.parts[last].matches(n, hover) {
		return false
	}
	return s.matchFrom(last-1, n.parent, s.combinators, hover)
}

// matchFrom matches parts[..i] against the ancestors starting at n.
func (s *selector) matchFrom(i int, n *node, combs []byte, hover bool) bool {
	if i < 0 {
		return true
	}
	if combs[i] == '>' {
		return s.parts[i].matches(n, hover) && s.matchFrom(i-1, n.parent, combs, hover)
	}
	for a := n; a != nil; a = a.parent {
		if s.parts[i].matches(a, hover) && s.matchFrom(i-1, a.parent, combs, hover) {
			return true
		}
	}
	return false
}

// hasHover reports whether the selector's last part needs :hover.
func (s *selector) hasHover() bool {
	return s.parts[len(s.parts)-1].pseudo == "hover"
}
