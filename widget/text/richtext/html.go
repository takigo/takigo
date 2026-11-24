package richtext

import (
	"slices"
	"strings"
	"unicode"
)

// nodeKind tells element nodes from text nodes.
type nodeKind int

const (
	elementNode nodeKind = iota
	textNode
)

// node is one node of the parsed document tree.
type node struct {
	kind     nodeKind
	tag      string // lower case; "" for the root
	attrs    map[string]string
	text     string // a text node's text, entities decoded
	parent   *node
	children []*node
	classes  []string
	id       string
}

func (n *node) attr(name string) (string, bool) {
	v, ok := n.attrs[name]
	return v, ok
}

// voidElements never have content or an end tag.
var voidElements = map[string]bool{
	"area": true, "base": true, "br": true, "col": true, "embed": true, "hr": true,
	"img": true, "input": true, "link": true, "meta": true, "source": true,
	"track": true, "wbr": true,
}

// rawTextElements hold text up to their end tag, unparsed.
var rawTextElements = map[string]bool{"script": true, "style": true, "textarea": true, "title": true}

// closesP are the elements whose start tag ends an open <p>.
var closesP = map[string]bool{
	"address": true, "article": true, "aside": true, "blockquote": true, "details": true,
	"div": true, "dl": true, "fieldset": true, "figcaption": true, "figure": true,
	"footer": true, "form": true, "h1": true, "h2": true, "h3": true, "h4": true,
	"h5": true, "h6": true, "header": true, "hr": true, "main": true, "nav": true,
	"ol": true, "p": true, "pre": true, "section": true, "table": true, "ul": true,
	"center": true,
}

// implied lists, for an element whose start tag ends an open one, the
// elements it closes and the elements that stop the search (a scope).
var implied = map[string]struct{ closes, scope []string }{
	"li":     {[]string{"li"}, []string{"ul", "ol"}},
	"dt":     {[]string{"dt", "dd"}, []string{"dl"}},
	"dd":     {[]string{"dt", "dd"}, []string{"dl"}},
	"tr":     {[]string{"tr", "td", "th"}, []string{"table", "thead", "tbody", "tfoot"}},
	"td":     {[]string{"td", "th"}, []string{"tr", "table"}},
	"th":     {[]string{"td", "th"}, []string{"tr", "table"}},
	"thead":  {[]string{"thead", "tbody", "tfoot", "tr", "td", "th"}, []string{"table"}},
	"tbody":  {[]string{"thead", "tbody", "tfoot", "tr", "td", "th"}, []string{"table"}},
	"tfoot":  {[]string{"thead", "tbody", "tfoot", "tr", "td", "th"}, []string{"table"}},
	"option": {[]string{"option"}, []string{"select"}},
}

// parseHTML builds a tree from HTML the way a forgiving browser parser
// would for the common cases: unknown or stray end tags are ignored, open
// elements are closed by their parent's end tag, <p>, <li>, <td> and the
// like close implicitly, and void elements need no end tag. The text of
// <style> elements is returned too.
func parseHTML(src string) (root *node, styles []string) {
	root = &node{kind: elementNode}
	cur := root
	appendText := func(s string) {
		if s == "" {
			return
		}
		if k := len(cur.children); k > 0 && cur.children[k-1].kind == textNode {
			cur.children[k-1].text += s
			return
		}
		cur.children = append(cur.children, &node{kind: textNode, text: s, parent: cur})
	}
	inScope := func(tag string, scope []string) *node {
		for n := cur; n != nil && n != root; n = n.parent {
			if n.tag == tag {
				return n
			}
			if slices.Contains(scope, n.tag) {
				return nil
			}
		}
		return nil
	}
	closeTo := func(n *node) { cur = n.parent }

	i := 0
	for i < len(src) {
		lt := strings.IndexByte(src[i:], '<')
		if lt < 0 {
			appendText(decodeEntities(src[i:]))
			break
		}
		appendText(decodeEntities(src[i : i+lt]))
		i += lt
		rest := src[i:]
		switch {
		case strings.HasPrefix(rest, "<!--"):
			end := strings.Index(rest[4:], "-->")
			if end < 0 {
				return root, styles
			}
			i += 4 + end + 3
			continue
		case strings.HasPrefix(rest, "<!") || strings.HasPrefix(rest, "<?"):
			end := strings.IndexByte(rest, '>')
			if end < 0 {
				return root, styles
			}
			i += end + 1
			continue
		case strings.HasPrefix(rest, "</"):
			name, _ := scanName(rest[2:])
			end := strings.IndexByte(rest, '>')
			if name == "" || end < 0 {
				appendText("<")
				i++
				continue
			}
			i += end + 1
			if name == "br" { // </br> is a <br> to browsers
				cur.children = append(cur.children, &node{kind: elementNode, tag: "br", parent: cur})
				continue
			}
			for o := cur; o != nil && o != root; o = o.parent {
				if o.tag == name {
					closeTo(o)
					break
				}
			}
			continue
		}
		name, n := scanName(rest[1:])
		if name == "" {
			appendText("<")
			i++
			continue
		}
		attrs, selfClose, used := scanAttrs(rest[1+n:])
		i += 1 + n + used
		if closesP[name] {
			if p := inScope("p", []string{"button", "table", "td", "th", "li", "blockquote", "div"}); p != nil {
				closeTo(p)
			}
		}
		if rule, ok := implied[name]; ok {
			for _, c := range rule.closes {
				if o := inScope(c, rule.scope); o != nil {
					closeTo(o)
					break
				}
			}
		}
		el := &node{kind: elementNode, tag: name, attrs: attrs, parent: cur}
		if cls, ok := attrs["class"]; ok {
			el.classes = strings.Fields(cls)
		}
		el.id = attrs["id"]
		cur.children = append(cur.children, el)
		if voidElements[name] || selfClose {
			continue
		}
		if rawTextElements[name] {
			end := indexFold(src[i:], "</"+name)
			body := src[i:]
			if end >= 0 {
				body = src[i : i+end]
			}
			if name == "style" {
				styles = append(styles, body)
			} else if name != "script" {
				el.children = append(el.children, &node{kind: textNode, text: decodeEntities(body), parent: el})
			}
			if end < 0 {
				return root, styles
			}
			i += end
			if gt := strings.IndexByte(src[i:], '>'); gt >= 0 {
				i += gt + 1
			} else {
				i = len(src)
			}
			continue
		}
		cur = el
	}
	return root, styles
}

// scanName reads a tag name and returns it in lower case with its length.
func scanName(s string) (string, int) {
	n := 0
	for n < len(s) {
		c := s[n]
		if c == '>' || c == '/' || c == '=' || unicode.IsSpace(rune(c)) {
			break
		}
		n++
	}
	if n == 0 || !isLetter(s[0]) {
		return "", 0
	}
	return strings.ToLower(s[:n]), n
}

func isLetter(c byte) bool { return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' }

// scanAttrs reads attributes up to the end of a start tag, returning them,
// whether the tag ends with "/>" and the bytes used, '>' included.
func scanAttrs(s string) (map[string]string, bool, int) {
	attrs := map[string]string{}
	i := 0
	for i < len(s) {
		for i < len(s) && unicode.IsSpace(rune(s[i])) {
			i++
		}
		if i >= len(s) {
			break
		}
		switch s[i] {
		case '>':
			return attrs, false, i + 1
		case '/':
			if i+1 < len(s) && s[i+1] == '>' {
				return attrs, true, i + 2
			}
			i++
			continue
		}
		start := i
		for i < len(s) && s[i] != '=' && s[i] != '>' && !unicode.IsSpace(rune(s[i])) &&
			(s[i] != '/' || i+1 >= len(s) || s[i+1] != '>') {
			i++
		}
		name := strings.ToLower(s[start:i])
		for i < len(s) && unicode.IsSpace(rune(s[i])) {
			i++
		}
		if i < len(s) && s[i] == '=' {
			i++
			for i < len(s) && unicode.IsSpace(rune(s[i])) {
				i++
			}
			var val string
			if i < len(s) && (s[i] == '"' || s[i] == '\'') {
				q := s[i]
				end := strings.IndexByte(s[i+1:], q)
				if end < 0 {
					val, i = s[i+1:], len(s)
				} else {
					val, i = s[i+1:i+1+end], i+1+end+1
				}
			} else {
				vs := i
				for i < len(s) && s[i] != '>' && !unicode.IsSpace(rune(s[i])) {
					i++
				}
				val = s[vs:i]
			}
			if name != "" {
				attrs[name] = decodeEntities(val)
			}
		} else if name != "" {
			attrs[name] = ""
		}
	}
	return attrs, false, len(s)
}

// indexFold is strings.Index ignoring ASCII case.
func indexFold(s, sub string) int {
	return strings.Index(strings.ToLower(s), strings.ToLower(sub))
}
