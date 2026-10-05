package main

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/takigo/takigo/internal/treedump"
)

// Kind classifies a difference. Lower values are more likely root causes and
// are reported first.
type Kind int

const (
	KindToplevel Kind = iota
	KindFont
	KindClass
	KindMissing
	KindExtra
	KindReqSize
	KindSize
	KindPos
	KindRender
)

var kindNames = [...]string{"TOPLEVEL", "FONT", "CLASS", "MISSING", "EXTRA", "REQSIZE", "SIZE", "POS", "RENDER"}

func (k Kind) String() string { return kindNames[k] }

// Diff is one reported difference.
type Diff struct {
	Kind    Kind     `json:"-"`
	KindStr string   `json:"kind"`
	Class   string   `json:"class,omitempty"`
	Go      string   `json:"go,omitempty"`
	Tcl     string   `json:"tcl,omitempty"`
	Detail  string   `json:"detail"`
	Pixels  *float64 `json:"pixels_pct,omitempty"`
	depth   int
}

// node is a widget with resolved children.
type node struct {
	w        *treedump.Widget
	children []*node
	parent   *node
	depth    int
}

// Pair is a matched Go/Tcl widget.
type Pair struct {
	Go, Tcl *node
}

// Result is the full comparison.
type Result struct {
	Pairs []Pair `json:"-"`
	Diffs []Diff `json:"diffs"`

	// Children left unmatched under their parent, retried across the whole
	// toplevel: Tk widgets can be managed by a geometry master other than
	// their parent (grid/pack -in), so the trees need not line up.
	orphanGo, orphanTcl []*node
}

func buildTrees(d *treedump.Dump) map[string]*node {
	byPath := map[string]*node{}
	for i := range d.Widgets {
		byPath[d.Widgets[i].Path] = &node{w: &d.Widgets[i]}
	}
	roots := map[string]*node{}
	for i := range d.Widgets {
		w := &d.Widgets[i]
		n := byPath[w.Path]
		if w.Path == w.Toplevel {
			roots[w.Toplevel] = n
			continue
		}
		if p := byPath[w.Parent]; p != nil {
			n.parent = p
			p.children = append(p.children, n)
		}
	}
	var setDepth func(n *node, d int)
	setDepth = func(n *node, d int) {
		n.depth = d
		sort.SliceStable(n.children, func(i, j int) bool { return n.children[i].w.Index < n.children[j].w.Index })
		for _, c := range n.children {
			setDepth(c, d+1)
		}
	}
	for _, r := range roots {
		setDepth(r, 0)
	}
	return roots
}

// transparent reports whether n is a plain container that only exists on one
// side for structural reasons (e.g. the Go demos' outer frame): a frame that
// covers its parent exactly.
func transparent(n *node) bool {
	if n.parent == nil || !isFrameClass(n.w.Class) || len(n.children) == 0 {
		return false
	}
	p := n.parent.w
	return n.w.X == p.X && n.w.Y == p.Y && n.w.W == p.W && n.w.H == p.H
}

func isFrameClass(c string) bool { return c == "Frame" || c == "TFrame" }

// effectiveChildren flattens transparent containers into their parent.
func effectiveChildren(n *node) []*node {
	var out []*node
	for _, c := range n.children {
		if transparent(c) {
			out = append(out, effectiveChildren(c)...)
		} else {
			out = append(out, c)
		}
	}
	return out
}

func baseClass(c string) string {
	if len(c) > 1 && c[0] == 'T' && c[1] >= 'A' && c[1] <= 'Z' {
		return strings.ToLower(c[1:])
	}
	return strings.ToLower(c)
}

func lastName(path string) string {
	if _, name, ok := strings.CutLast(path, "."); ok {
		return name
	}
	return path
}

// score rates how likely a and b are the same widget; < 0 means never.
func score(a, b *node, aOrigin, bOrigin [2]int) float64 {
	s := 0.0
	switch {
	case a.w.Class == b.w.Class:
		s += 10
	case baseClass(a.w.Class) == baseClass(b.w.Class):
		s += 4
	default:
		return -1
	}
	if lastName(a.w.Path) == lastName(b.w.Path) {
		s += 5
	}
	dist := math.Abs(float64((a.w.X-aOrigin[0])-(b.w.X-bOrigin[0]))) +
		math.Abs(float64((a.w.Y-aOrigin[1])-(b.w.Y-bOrigin[1]))) +
		math.Abs(float64(a.w.W-b.w.W)) + math.Abs(float64(a.w.H-b.w.H))
	s += 6 * math.Max(0, 1-dist/200)
	if a.w.Index == b.w.Index {
		s += 1
	}
	return s
}

// Compare matches the two dumps and lists differences, most likely root
// causes first.
func Compare(goDump, tclDump *treedump.Dump) *Result {
	res := &Result{}
	goRoots, tclRoots := buildTrees(goDump), buildTrees(tclDump)

	compareFonts(res, goDump, tclDump)

	tclByTitle := map[string]treedump.Toplevel{}
	for _, t := range tclDump.Toplevels {
		tclByTitle[t.Title] = t
	}
	usedTcl := map[string]bool{}
	for i, gt := range goDump.Toplevels {
		tt, ok := tclByTitle[gt.Title]
		if !ok && i < len(tclDump.Toplevels) && !usedTcl[tclDump.Toplevels[i].Path] {
			tt, ok = tclDump.Toplevels[i], true
		}
		if !ok {
			res.add(Diff{Kind: KindToplevel, Go: gt.Path, Detail: fmt.Sprintf("toplevel %q only in Go", gt.Title)})
			continue
		}
		usedTcl[tt.Path] = true
		if gt.W != tt.W || gt.H != tt.H {
			res.add(Diff{Kind: KindToplevel, Go: gt.Path, Tcl: tt.Path,
				Detail: fmt.Sprintf("%q size go %dx%d tcl %dx%d (Δ %+d,%+d)", gt.Title, gt.W, gt.H, tt.W, tt.H, gt.W-tt.W, gt.H-tt.H)})
		}
		g, t := goRoots[gt.Path], tclRoots[tt.Path]
		if g == nil || t == nil {
			continue
		}
		res.Pairs = append(res.Pairs, Pair{g, t})
		res.orphanGo, res.orphanTcl = nil, nil
		res.matchChildren(g, t)
		res.matchOrphans()
	}
	for _, tt := range tclDump.Toplevels {
		if !usedTcl[tt.Path] {
			res.add(Diff{Kind: KindToplevel, Tcl: tt.Path, Detail: fmt.Sprintf("toplevel %q only in Tcl", tt.Title)})
		}
	}

	sort.SliceStable(res.Diffs, func(i, j int) bool {
		if res.Diffs[i].Kind != res.Diffs[j].Kind {
			return res.Diffs[i].Kind < res.Diffs[j].Kind
		}
		// Deeper REQSIZE diffs are closer to the cause of a layout shift.
		return res.Diffs[i].depth > res.Diffs[j].depth
	})
	return res
}

func (r *Result) add(d Diff) {
	d.KindStr = d.Kind.String()
	r.Diffs = append(r.Diffs, d)
}

func (r *Result) matchChildren(g, t *node) {
	gc, tc := effectiveChildren(g), effectiveChildren(t)
	gOrigin, tOrigin := [2]int{g.w.X, g.w.Y}, [2]int{t.w.X, t.w.Y}

	type cand struct {
		i, j int
		s    float64
	}
	var cands []cand
	for i, a := range gc {
		for j, b := range tc {
			if s := score(a, b, gOrigin, tOrigin); s >= 0 {
				cands = append(cands, cand{i, j, s})
			}
		}
	}
	sort.SliceStable(cands, func(x, y int) bool { return cands[x].s > cands[y].s })
	gUsed, tUsed := make([]bool, len(gc)), make([]bool, len(tc))
	type match struct{ a, b *node }
	var matches []match
	for _, c := range cands {
		if gUsed[c.i] || tUsed[c.j] {
			continue
		}
		gUsed[c.i], tUsed[c.j] = true, true
		matches = append(matches, match{gc[c.i], tc[c.j]})
	}
	for i, a := range gc {
		if !gUsed[i] && a.w.Mapped {
			r.orphanGo = append(r.orphanGo, a)
		}
	}
	for j, b := range tc {
		if !tUsed[j] && b.w.Mapped {
			r.orphanTcl = append(r.orphanTcl, b)
		}
	}
	for _, m := range matches {
		r.Pairs = append(r.Pairs, Pair{m.a, m.b})
		r.comparePair(m.a, m.b, gOrigin, tOrigin)
		r.matchChildren(m.a, m.b)
	}
}

// matchOrphans pairs leftover widgets of one toplevel by class and absolute
// position, then reports the rest as EXTRA/MISSING.
func (r *Result) matchOrphans() {
	for len(r.orphanGo) > 0 || len(r.orphanTcl) > 0 {
		gs, ts := r.orphanGo, r.orphanTcl
		r.orphanGo, r.orphanTcl = nil, nil
		type cand struct {
			i, j int
			s    float64
		}
		var cands []cand
		for i, a := range gs {
			for j, b := range ts {
				if a.w.Class != b.w.Class {
					continue
				}
				if s := score(a, b, [2]int{}, [2]int{}); s >= 10 {
					cands = append(cands, cand{i, j, s})
				}
			}
		}
		sort.SliceStable(cands, func(x, y int) bool { return cands[x].s > cands[y].s })
		gUsed, tUsed := make([]bool, len(gs)), make([]bool, len(ts))
		matched := false
		for _, c := range cands {
			if gUsed[c.i] || tUsed[c.j] {
				continue
			}
			gUsed[c.i], tUsed[c.j] = true, true
			matched = true
			a, b := gs[c.i], ts[c.j]
			r.Pairs = append(r.Pairs, Pair{a, b})
			r.comparePair(a, b, [2]int{}, [2]int{})
			r.matchChildren(a, b)
		}
		for i, a := range gs {
			if !gUsed[i] {
				r.add(Diff{Kind: KindExtra, Class: a.w.Class, Go: a.w.Path, depth: a.depth,
					Detail: fmt.Sprintf("only in Go (%dx%d at %d,%d)", a.w.W, a.w.H, a.w.X, a.w.Y)})
			}
		}
		for j, b := range ts {
			if !tUsed[j] {
				r.add(Diff{Kind: KindMissing, Class: b.w.Class, Tcl: b.w.Path, depth: b.depth,
					Detail: fmt.Sprintf("missing in Go (%dx%d at %d,%d)", b.w.W, b.w.H, b.w.X, b.w.Y)})
			}
		}
		if !matched {
			break
		}
	}
}

func (r *Result) comparePair(a, b *node, gOrigin, tOrigin [2]int) {
	base := Diff{Class: b.w.Class, Go: a.w.Path, Tcl: b.w.Path, depth: b.depth}
	if a.w.Class != b.w.Class {
		d := base
		d.Kind, d.Detail = KindClass, fmt.Sprintf("class go %s tcl %s", a.w.Class, b.w.Class)
		r.add(d)
	}
	if a.w.Mapped != b.w.Mapped {
		d := base
		d.Kind, d.Detail = KindMissing, fmt.Sprintf("mapped go %v tcl %v", a.w.Mapped, b.w.Mapped)
		r.add(d)
	}
	reqDiff := a.w.ReqW != b.w.ReqW || a.w.ReqH != b.w.ReqH
	if reqDiff {
		d := base
		d.Kind = KindReqSize
		d.Detail = fmt.Sprintf("req go %dx%d tcl %dx%d (Δ %+d,%+d)", a.w.ReqW, a.w.ReqH, b.w.ReqW, b.w.ReqH, a.w.ReqW-b.w.ReqW, a.w.ReqH-b.w.ReqH)
		r.add(d)
	}
	sizeFollowsReq := a.w.W-b.w.W == a.w.ReqW-b.w.ReqW && a.w.H-b.w.H == a.w.ReqH-b.w.ReqH
	if (a.w.W != b.w.W || a.w.H != b.w.H) && !sizeFollowsReq {
		d := base
		d.Kind = KindSize
		d.Detail = fmt.Sprintf("size go %dx%d tcl %dx%d (Δ %+d,%+d)", a.w.W, a.w.H, b.w.W, b.w.H, a.w.W-b.w.W, a.w.H-b.w.H)
		r.add(d)
	}
	dx := (a.w.X - gOrigin[0]) - (b.w.X - tOrigin[0])
	dy := (a.w.Y - gOrigin[1]) - (b.w.Y - tOrigin[1])
	if dx != 0 || dy != 0 {
		d := base
		d.Kind = KindPos
		d.Detail = fmt.Sprintf("offset in parent go %d,%d tcl %d,%d (Δ %+d,%+d)",
			a.w.X-gOrigin[0], a.w.Y-gOrigin[1], b.w.X-tOrigin[0], b.w.Y-tOrigin[1], dx, dy)
		r.add(d)
	}
}

func compareFonts(r *Result, g, t *treedump.Dump) {
	names := make([]string, 0, len(t.Fonts))
	for n := range t.Fonts {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		tf := t.Fonts[n]
		gf, ok := g.Fonts[n]
		if !ok {
			r.add(Diff{Kind: KindFont, Tcl: n, Detail: "named font missing in Go"})
			continue
		}
		var parts []string
		if gf.Size != tf.Size {
			parts = append(parts, fmt.Sprintf("size %g/%g", gf.Size, tf.Size))
		}
		if gf.Weight != tf.Weight {
			parts = append(parts, fmt.Sprintf("weight %s/%s", gf.Weight, tf.Weight))
		}
		if gf.Slant != tf.Slant {
			parts = append(parts, fmt.Sprintf("slant %s/%s", gf.Slant, tf.Slant))
		}
		if gf.Ascent != tf.Ascent || gf.Descent != tf.Descent || gf.Linespace != tf.Linespace {
			parts = append(parts, fmt.Sprintf("ascent/descent/linespace %d/%d/%d vs %d/%d/%d",
				gf.Ascent, gf.Descent, gf.Linespace, tf.Ascent, tf.Descent, tf.Linespace))
		}
		if gf.Sample != tf.Sample {
			parts = append(parts, fmt.Sprintf("sample width %d/%d", gf.Sample, tf.Sample))
		}
		if len(parts) > 0 {
			r.add(Diff{Kind: KindFont, Go: n, Tcl: n, Detail: "go/tcl " + strings.Join(parts, ", ")})
		}
	}
}
