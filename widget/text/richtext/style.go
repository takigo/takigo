package richtext

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
)

// Sides index the four-sided box properties: top, right, bottom, left as in
// CSS shorthand order.
const (
	top = iota
	right
	bottom
	left
)

// style is an element's computed style, the subset of CSS this package
// draws. Lengths are in pixels.
type style struct {
	// Inherited.
	family     string
	size       float64
	bold       bool
	italic     bool
	color      string
	align      string // left, right, center
	whiteSpace string // normal, pre, nowrap, pre-wrap, pre-line
	listStyle  string
	underline  bool // text-decoration, which descendants draw too
	strike     bool
	textIndent float64

	// Not inherited.
	display       string
	background    string
	margin        [4]float64
	padding       [4]float64
	border        [4]float64
	borderColor   [4]string
	radius        float64
	verticalAlign string
	width, height length
	maxWidth      length
	float         string
	clear         string
}

// length is a CSS length kept for resolving later against a size that is
// not known when the style is computed (an image's width in percent).
type length struct {
	px      float64
	percent float64
	set     bool
}

// defaultDisplay is the user agent's display of each element.
var defaultDisplay = map[string]string{
	"html": "block", "body": "block", "div": "block", "p": "block", "pre": "block",
	"blockquote": "block", "h1": "block", "h2": "block", "h3": "block", "h4": "block",
	"h5": "block", "h6": "block", "ul": "block", "ol": "block", "li": "list-item",
	"dl": "block", "dt": "block", "dd": "block", "hr": "block", "table": "table",
	"thead": "block", "tbody": "block", "tfoot": "block", "tr": "block", "td": "block",
	"th": "block", "caption": "block", "section": "block", "article": "block",
	"header": "block", "footer": "block", "nav": "block", "aside": "block",
	"main": "block", "figure": "block", "figcaption": "block", "address": "block",
	"center": "block", "form": "block", "fieldset": "block", "details": "block",
	"summary": "block", "legend": "block",
	"head": "none", "title": "none", "script": "none", "style": "none", "meta": "none",
	"link": "none", "template": "none", "base": "none",
}

// styler computes styles from the stylesheets of a document.
type styler struct {
	sheet  stylesheet
	rootPx float64 // the root font size, for rem
	scale  float64 // device pixels per CSS pixel
	ptPx   float64 // device pixels per point
}

// compute returns the style of element n given its parent's style.
func (st *styler) compute(n *node, parent *style) style {
	s := style{
		family: parent.family, size: parent.size, bold: parent.bold, italic: parent.italic,
		color: parent.color, align: parent.align, whiteSpace: parent.whiteSpace,
		listStyle: parent.listStyle, underline: parent.underline, strike: parent.strike,
		textIndent: parent.textIndent,
		display:    cmp.Or(defaultDisplay[n.tag], "inline"),
	}
	decls := st.cascade(n, presentational(n), false)
	// Ems in the other properties are of the element's own font size.
	isFont := func(d declaration) bool { return d.prop == "font-size" || d.prop == "font" }
	for _, d := range decls {
		if isFont(d) {
			st.apply(&s, parent, d)
		}
	}
	for _, d := range decls {
		if !isFont(d) {
			st.apply(&s, parent, d)
		}
	}
	return s
}

// hoverDeclarations returns the declarations of the :hover rules matching
// n, for a link's hover style.
func (st *styler) hoverDeclarations(n *node) []declaration {
	return st.cascade(n, nil, true)
}

// cascade collects the declarations that apply to n in cascade order: the
// default stylesheet, the presentational hints of HTML attributes, author
// rules, the style attribute, then !important ones. With hover only those
// of :hover rules.
func (st *styler) cascade(n *node, hints []declaration, hover bool) []declaration {
	type hit struct {
		d      declaration
		weight [5]int
	}
	var hits []hit
	for _, r := range st.sheet.rules {
		if r.sel.hasHover() != hover || !r.sel.matches(n, hover) {
			continue
		}
		for _, d := range r.decls {
			imp := 0
			if d.important {
				imp = 1
			}
			hits = append(hits, hit{d, [5]int{imp, 2 * r.origin, r.sel.spec[0]*1e6 + r.sel.spec[1]*1e3 + r.sel.spec[2], r.order, 0}})
		}
	}
	for _, d := range hints {
		hits = append(hits, hit{d, [5]int{0, 1, 0, 0, 0}})
	}
	if !hover {
		if inline, ok := n.attr("style"); ok {
			for _, d := range parseDeclarations(inline) {
				imp := 0
				if d.important {
					imp = 1
				}
				hits = append(hits, hit{d, [5]int{imp, 3, 0, 0, 0}})
			}
		}
	}
	slices.SortStableFunc(hits, func(a, b hit) int {
		for i := range a.weight {
			if c := cmp.Compare(a.weight[i], b.weight[i]); c != 0 {
				return c
			}
		}
		return 0
	})
	out := make([]declaration, len(hits))
	for i, h := range hits {
		out[i] = h.d
	}
	return out
}

// presentational maps HTML attributes to declarations (align, bgcolor,
// <font>, width and height), which author CSS overrides.
func presentational(n *node) []declaration {
	var ds []declaration
	add := func(p, v string) { ds = append(ds, declaration{prop: p, value: v}) }
	if v, ok := n.attr("align"); ok && n.tag != "img" {
		add("text-align", v)
	}
	if v, ok := n.attr("bgcolor"); ok {
		add("background-color", v)
	}
	if v, ok := n.attr("color"); ok && n.tag == "font" {
		add("color", v)
	}
	if v, ok := n.attr("face"); ok && n.tag == "font" {
		add("font-family", v)
	}
	if v, ok := n.attr("size"); ok && n.tag == "font" {
		sizes := []string{"x-small", "small", "medium", "large", "x-large", "xx-large", "xx-large"}
		if k, err := strconv.Atoi(strings.TrimPrefix(v, "+")); err == nil {
			if strings.HasPrefix(v, "+") || strings.HasPrefix(v, "-") {
				k += 3
			}
			add("font-size", sizes[min(max(k, 1), 7)-1])
		}
	}
	if n.tag == "img" {
		for _, a := range []string{"width", "height"} {
			if v, ok := n.attr(a); ok {
				if _, err := strconv.Atoi(v); err == nil {
					v += "px"
				}
				add(a, v)
			}
		}
		if v, ok := n.attr("align"); ok {
			switch strings.ToLower(v) {
			case "left", "right":
				add("float", v)
			default:
				add("vertical-align", v)
			}
		}
	}
	if n.tag == "table" {
		if v, ok := n.attr("border"); ok && v != "0" {
			add("border", "1px solid #808080")
		}
	}
	if v, ok := n.attr("type"); ok && (n.tag == "ol" || n.tag == "ul" || n.tag == "li") {
		switch v {
		case "1":
			add("list-style-type", "decimal")
		case "a":
			add("list-style-type", "lower-alpha")
		case "A":
			add("list-style-type", "upper-alpha")
		case "i":
			add("list-style-type", "lower-roman")
		case "I":
			add("list-style-type", "upper-roman")
		default:
			add("list-style-type", v)
		}
	}
	return ds
}

// apply sets one declaration on s.
func (st *styler) apply(s *style, parent *style, d declaration) {
	v := strings.TrimSpace(d.value)
	lv := strings.ToLower(v)
	if lv == "inherit" {
		st.inherit(s, parent, d.prop)
		return
	}
	switch d.prop {
	case "color":
		if c, ok := parseColor(v); ok {
			s.color = c
		}
	case "background-color":
		if c, ok := parseColor(v); ok {
			s.background = c
		}
	case "background":
		for _, f := range splitValue(v) {
			if c, ok := parseColor(f); ok {
				s.background = c
			}
		}
	case "font-family":
		s.family = parseFamily(v)
	case "font-size":
		if px, ok := st.fontSize(lv, parent.size); ok {
			s.size = px
		}
	case "font-weight":
		switch lv {
		case "bold", "bolder":
			s.bold = true
		case "normal", "lighter":
			s.bold = false
		default:
			if w, err := strconv.Atoi(lv); err == nil {
				s.bold = w >= 600
			}
		}
	case "font-style":
		s.italic = lv == "italic" || lv == "oblique"
	case "font":
		st.applyFont(s, parent, lv)
	case "text-decoration", "text-decoration-line":
		if lv == "none" {
			s.underline, s.strike = false, false
		}
		for f := range strings.FieldsSeq(lv) {
			switch f {
			case "underline":
				s.underline = true
			case "line-through":
				s.strike = true
			}
		}
	case "text-align":
		switch lv {
		case "left", "start", "justify":
			s.align = "left"
		case "right", "end":
			s.align = "right"
		case "center", "middle":
			s.align = "center"
		}
	case "text-indent":
		if px, ok := st.length(lv, s.size); ok {
			s.textIndent = px
		}
	case "white-space":
		s.whiteSpace = lv
	case "list-style-type":
		s.listStyle = lv
	case "list-style":
		for f := range strings.FieldsSeq(lv) {
			if listMarker(f, 1) != "" || f == "none" {
				s.listStyle = f
			}
		}
	case "display":
		s.display = lv
	case "vertical-align":
		s.verticalAlign = lv
	case "float":
		s.float = lv
	case "clear":
		s.clear = lv
	case "width":
		s.width = st.boxLength(lv, s.size)
	case "height":
		s.height = st.boxLength(lv, s.size)
	case "max-width":
		s.maxWidth = st.boxLength(lv, s.size)
	case "border-radius":
		if f := strings.Fields(lv); len(f) > 0 {
			if px, ok := st.length(f[0], s.size); ok {
				s.radius = max(px, 0)
			}
		}
	default:
		st.applyBox(s, d.prop, v, lv)
	}
}

// applyBox sets the margin, padding and border properties.
func (st *styler) applyBox(s *style, prop, v, lv string) {
	switch prop {
	case "margin":
		st.sides(&s.margin, lv, s.size)
	case "padding":
		st.sides(&s.padding, lv, s.size)
	case "margin-top", "margin-right", "margin-bottom", "margin-left":
		if px, ok := st.length(lv, s.size); ok {
			s.margin[sideOf(prop)] = px
		}
	case "padding-top", "padding-right", "padding-bottom", "padding-left":
		if px, ok := st.length(lv, s.size); ok {
			s.padding[sideOf(prop)] = max(px, 0)
		}
	case "border":
		w, c := st.border(lv, s.size)
		s.border = [4]float64{w, w, w, w}
		s.borderColor = [4]string{c, c, c, c}
	case "border-top", "border-right", "border-bottom", "border-left":
		w, c := st.border(lv, s.size)
		s.border[sideOf(prop)], s.borderColor[sideOf(prop)] = w, c
	case "border-width":
		st.sides(&s.border, lv, s.size)
	case "border-color":
		if c, ok := parseColor(v); ok {
			s.borderColor = [4]string{c, c, c, c}
		}
	case "border-top-width", "border-right-width", "border-bottom-width", "border-left-width":
		if px, ok := st.borderWidth(lv, s.size); ok {
			s.border[sideOf(prop)] = px
		}
	case "border-top-color", "border-right-color", "border-bottom-color", "border-left-color":
		if c, ok := parseColor(v); ok {
			s.borderColor[sideOf(prop)] = c
		}
	}
}

func (st *styler) inherit(s, parent *style, prop string) {
	switch prop {
	case "background-color", "background":
		s.background = parent.background
	case "display":
		s.display = parent.display
	case "margin":
		s.margin = parent.margin
	case "padding":
		s.padding = parent.padding
	}
}

// applyFont reads the font shorthand: style and weight keywords, a size
// (with an optional /line-height) and the family list.
func (st *styler) applyFont(s, parent *style, v string) {
	fields := strings.Fields(v)
	for i, f := range fields {
		switch f {
		case "italic", "oblique":
			s.italic = true
		case "bold", "bolder":
			s.bold = true
		case "normal":
		default:
			size, _, _ := strings.Cut(f, "/")
			if px, ok := st.fontSize(size, parent.size); ok {
				s.size = px
				if i+1 < len(fields) {
					s.family = parseFamily(strings.Join(fields[i+1:], " "))
				}
				return
			}
		}
	}
}

// splitValue splits a property value at spaces outside parentheses.
func splitValue(v string) []string {
	var out []string
	depth, start := 0, -1
	for i, r := range v {
		switch r {
		case '(':
			depth++
		case ')':
			depth = max(depth-1, 0)
		case ' ', '\t', '\n':
			if depth == 0 {
				if start >= 0 {
					out = append(out, v[start:i])
				}
				start = -1
				continue
			}
		}
		if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		out = append(out, v[start:])
	}
	return out
}

func sideOf(prop string) int {
	switch {
	case strings.Contains(prop, "top"):
		return top
	case strings.Contains(prop, "right"):
		return right
	case strings.Contains(prop, "bottom"):
		return bottom
	}
	return left
}

// sides reads a 1-4 value box shorthand.
func (st *styler) sides(dst *[4]float64, v string, fontPx float64) {
	var vals []float64
	for f := range strings.FieldsSeq(v) {
		px, ok := st.length(f, fontPx)
		if !ok {
			return
		}
		vals = append(vals, px)
	}
	switch len(vals) {
	case 1:
		*dst = [4]float64{vals[0], vals[0], vals[0], vals[0]}
	case 2:
		*dst = [4]float64{vals[0], vals[1], vals[0], vals[1]}
	case 3:
		*dst = [4]float64{vals[0], vals[1], vals[2], vals[1]}
	case 4:
		*dst = [4]float64{vals[0], vals[1], vals[2], vals[3]}
	}
}

// border reads "1px solid #ccc" into a width and colour; a style of none
// or hidden is no border, and a missing width is medium.
func (st *styler) border(v string, fontPx float64) (float64, string) {
	w, c, styled := 3*st.scale, "", false
	for _, f := range splitValue(v) {
		switch f {
		case "none", "hidden":
			return 0, ""
		case "solid", "dashed", "dotted", "double", "groove", "ridge", "inset", "outset":
			styled = true
			continue
		}
		if px, ok := st.borderWidth(f, fontPx); ok {
			w = px
		} else if col, ok := parseColor(f); ok {
			c = col
		}
	}
	if !styled {
		return 0, ""
	}
	return w, c
}

func (st *styler) borderWidth(v string, fontPx float64) (float64, bool) {
	switch v {
	case "thin":
		return st.scale, true
	case "medium":
		return 3 * st.scale, true
	case "thick":
		return 5 * st.scale, true
	}
	return st.length(v, fontPx)
}

// length converts a CSS length to device pixels; percentages are not
// lengths here.
func (st *styler) length(v string, fontPx float64) (float64, bool) {
	v = strings.TrimSpace(v)
	if v == "0" || v == "auto" {
		return 0, true
	}
	num, unit := splitUnit(v)
	n, err := strconv.ParseFloat(num, 64)
	if err != nil {
		return 0, false
	}
	switch unit {
	case "px":
		return n * st.scale, true
	case "em":
		return n * fontPx, true
	case "rem":
		return n * st.rootPx, true
	case "ex", "ch":
		return n * fontPx / 2, true
	case "pt":
		return n * st.ptPx, true
	case "pc":
		return n * 12 * st.ptPx, true
	case "in":
		return n * 72 * st.ptPx, true
	case "cm":
		return n * 72 / 2.54 * st.ptPx, true
	case "mm":
		return n * 72 / 25.4 * st.ptPx, true
	case "":
		return n * st.scale, n == 0
	}
	return 0, false
}

// boxLength reads a width or height, keeping percentages.
func (st *styler) boxLength(v string, fontPx float64) length {
	if p, ok := strings.CutSuffix(v, "%"); ok {
		if n, err := strconv.ParseFloat(p, 64); err == nil {
			return length{percent: n, set: true}
		}
		return length{}
	}
	if v == "auto" || v == "none" {
		return length{}
	}
	if px, ok := st.length(v, fontPx); ok {
		return length{px: px, set: true}
	}
	return length{}
}

// fontSize reads font-size, relative to the parent's size.
func (st *styler) fontSize(v string, parentPx float64) (float64, bool) {
	keywords := map[string]float64{
		"xx-small": 9.0 / 16, "x-small": 10.0 / 16, "small": 13.0 / 16, "medium": 1,
		"large": 18.0 / 16, "x-large": 24.0 / 16, "xx-large": 32.0 / 16, "xxx-large": 3,
	}
	if k, ok := keywords[v]; ok {
		return st.rootPx * k, true
	}
	switch v {
	case "smaller":
		return parentPx / 1.2, true
	case "larger":
		return parentPx * 1.2, true
	}
	if p, ok := strings.CutSuffix(v, "%"); ok {
		n, err := strconv.ParseFloat(p, 64)
		return parentPx * n / 100, err == nil
	}
	px, ok := st.length(v, parentPx)
	return px, ok && px > 0
}

func splitUnit(v string) (string, string) {
	i := len(v)
	for i > 0 && (v[i-1] >= 'a' && v[i-1] <= 'z' || v[i-1] == '%') {
		i--
	}
	return v[:i], v[i:]
}

// parseFamily picks the first family of a CSS list, mapping the generic
// families to fontconfig's.
func parseFamily(v string) string {
	first, _, _ := strings.Cut(v, ",")
	first = strings.Trim(strings.TrimSpace(first), `"'`)
	switch strings.ToLower(first) {
	case "monospace", "ui-monospace", "courier", "courier new", "consolas", "menlo", "monaco":
		return "monospace"
	case "sans-serif", "system-ui", "ui-sans-serif", "-apple-system", "helvetica", "arial":
		return "sans-serif"
	case "serif", "ui-serif", "times", "times new roman", "georgia":
		return "serif"
	}
	return first
}

// parseColor reads a CSS colour into a form color.Cache resolves: names
// and #rgb/#rrggbb as given, rgb() and hsl() converted to #rrggbb. ok is
// true with "" for transparent.
func parseColor(v string) (string, bool) {
	lv := strings.ToLower(strings.TrimSpace(v))
	switch {
	case lv == "transparent" || lv == "none":
		return "", true
	case strings.HasPrefix(lv, "#"):
		hex := lv[1:]
		switch len(hex) {
		case 3, 6:
			return lv, isHex(hex)
		case 4, 8: // with alpha, which is dropped
			if !isHex(hex) {
				return "", false
			}
			return "#" + hex[:len(hex)*3/4], true
		}
		return "", false
	case strings.HasPrefix(lv, "rgb"):
		args, ok := funcArgs(lv)
		if !ok || len(args) < 3 {
			return "", false
		}
		if len(args) > 3 && args[3] == 0 {
			return "", true
		}
		return fmt.Sprintf("#%02x%02x%02x", clamp255(args[0]), clamp255(args[1]), clamp255(args[2])), true
	case strings.HasPrefix(lv, "hsl"):
		args, ok := funcArgs(lv)
		if !ok || len(args) < 3 {
			return "", false
		}
		r, g, b := hslToRGB(args[0], args[1]/255, args[2]/255)
		return fmt.Sprintf("#%02x%02x%02x", r, g, b), true
	}
	for _, c := range lv {
		if c < 'a' || c > 'z' {
			return "", false
		}
	}
	return lv, lv != ""
}

func isHex(s string) bool {
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

// funcArgs reads the arguments of rgb(...)/hsl(...): numbers, with
// percentages scaled to 0-255 and alpha left at 0-1.
func funcArgs(v string) ([]float64, bool) {
	lp, rp := strings.IndexByte(v, '('), strings.LastIndexByte(v, ')')
	if lp < 0 || rp < lp {
		return nil, false
	}
	inner := strings.NewReplacer(",", " ", "/", " ").Replace(v[lp+1 : rp])
	var out []float64
	for f := range strings.FieldsSeq(inner) {
		f = strings.TrimSuffix(f, "deg")
		pct := strings.HasSuffix(f, "%")
		n, err := strconv.ParseFloat(strings.TrimSuffix(f, "%"), 64)
		if err != nil {
			return nil, false
		}
		if pct {
			n = n * 255 / 100
			if len(out) == 3 {
				n /= 255
			}
		}
		out = append(out, n)
	}
	return out, true
}

func clamp255(f float64) int { return int(math.Round(min(max(f, 0), 255))) }

func hslToRGB(h, s, l float64) (int, int, int) {
	h = math.Mod(h, 360) / 360
	if h < 0 {
		h++
	}
	if s == 0 {
		v := clamp255(l * 255)
		return v, v, v
	}
	q := l * (1 + s)
	if l >= 0.5 {
		q = l + s - l*s
	}
	p := 2*l - q
	hue := func(t float64) float64 {
		switch {
		case t < 0:
			t++
		case t > 1:
			t--
		}
		switch {
		case t < 1.0/6:
			return p + (q-p)*6*t
		case t < 0.5:
			return q
		case t < 2.0/3:
			return p + (q-p)*(2.0/3-t)*6
		}
		return p
	}
	return clamp255(hue(h+1.0/3) * 255), clamp255(hue(h) * 255), clamp255(hue(h-1.0/3) * 255)
}

// listMarker returns the marker of the n'th item (from 1) of a list style,
// or "" for none or an unknown style.
func listMarker(styleName string, n int) string {
	switch styleName {
	case "disc":
		return "•"
	case "circle":
		return "◦"
	case "square":
		return "▪"
	case "decimal":
		return strconv.Itoa(n) + "."
	case "decimal-leading-zero":
		return fmt.Sprintf("%02d.", n)
	case "lower-alpha", "lower-latin":
		return alpha(n, 'a') + "."
	case "upper-alpha", "upper-latin":
		return alpha(n, 'A') + "."
	case "lower-roman":
		return strings.ToLower(roman(n)) + "."
	case "upper-roman":
		return roman(n) + "."
	case "lower-greek":
		if n >= 1 && n <= 24 {
			return string(rune('α'+n-1+boolInt(n > 17))) + "."
		}
		return strconv.Itoa(n) + "."
	}
	return ""
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func alpha(n int, base rune) string {
	if n <= 0 {
		return strconv.Itoa(n)
	}
	var out []rune
	for n > 0 {
		n--
		out = append([]rune{base + rune(n%26)}, out...)
		n /= 26
	}
	return string(out)
}

func roman(n int) string {
	if n <= 0 || n >= 4000 {
		return strconv.Itoa(n)
	}
	vals := []int{1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1}
	syms := []string{"M", "CM", "D", "CD", "C", "XC", "L", "XL", "X", "IX", "V", "IV", "I"}
	var b strings.Builder
	for i, v := range vals {
		for n >= v {
			b.WriteString(syms[i])
			n -= v
		}
	}
	return b.String()
}
