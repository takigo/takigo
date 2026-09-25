// Command demodiff compares the widget trees dumped by a takigo demo and its
// Tk original (TAKIGO_DUMP_TREE, see internal/treedump and
// scripts/tk_dump_tree.tcl) and reports structural differences, most likely
// root causes first. With -go-png/-tcl-png it also diffs each matched leaf
// widget's pixels, flagging RENDER differences where geometry already agrees.
//
//	demodiff [-json] [-go-png go.png -tcl-png tcl.png] go.tree.json tcl.tree.json
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"strings"

	"github.com/msorc/takigo/internal/treedump"
)

func main() {
	jsonOut := flag.Bool("json", false, "print JSON instead of text")
	goPNG := flag.String("go-png", "", "Go screenshot for per-widget pixel diffs")
	tclPNG := flag.String("tcl-png", "", "Tcl screenshot for per-widget pixel diffs")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: demodiff [-json] [-go-png F -tcl-png F] go.tree.json tcl.tree.json\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(2)
	}
	goDump, err := treedump.Load(flag.Arg(0))
	if err != nil {
		fatal(err)
	}
	tclDump, err := treedump.Load(flag.Arg(1))
	if err != nil {
		fatal(err)
	}
	res := Compare(goDump, tclDump)
	if *goPNG != "" && *tclPNG != "" {
		gi, err := loadPNG(*goPNG)
		if err != nil {
			fatal(err)
		}
		ti, err := loadPNG(*tclPNG)
		if err != nil {
			fatal(err)
		}
		addPixelDiffs(res, gi, ti)
	}
	if *jsonOut {
		out := struct {
			Summary map[string]int `json:"summary"`
			Diffs   []Diff         `json:"diffs"`
		}{summary(res), res.Diffs}
		if out.Diffs == nil {
			out.Diffs = []Diff{}
		}
		b, _ := json.MarshalIndent(out, "", " ")
		fmt.Println(string(b))
		return
	}
	printText(res)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "demodiff:", err)
	os.Exit(1)
}

func loadPNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}

func summary(res *Result) map[string]int {
	s := map[string]int{"total": len(res.Diffs), "matched": len(res.Pairs)}
	for _, d := range res.Diffs {
		s[strings.ToLower(d.KindStr)]++
	}
	return s
}

func printText(res *Result) {
	for _, d := range res.Diffs {
		loc := d.Go
		if d.Tcl != "" && d.Tcl != d.Go {
			if loc != "" {
				loc += " ⇄ "
			}
			loc += d.Tcl
		}
		line := fmt.Sprintf("%-8s %-12s %s  %s", d.KindStr, d.Class, loc, d.Detail)
		if d.Pixels != nil {
			line += fmt.Sprintf("  [pixels %.1f%%]", *d.Pixels)
		}
		fmt.Println(strings.TrimRight(line, " "))
	}
	s := summary(res)
	var parts []string
	for _, k := range kindNames {
		if n := s[strings.ToLower(k)]; n > 0 {
			parts = append(parts, fmt.Sprintf("%s %d", strings.ToLower(k), n))
		}
	}
	fmt.Printf("summary: %d diffs, %d matched widgets", s["total"], s["matched"])
	if len(parts) > 0 {
		fmt.Printf(" (%s)", strings.Join(parts, ", "))
	}
	fmt.Println()
}

// addPixelDiffs compares every matched leaf widget's pixels. Pairs with equal
// size get the percentage attached to their existing diffs, or a new RENDER
// diff when geometry matched exactly but pixels did not.
func addPixelDiffs(res *Result, gi, ti image.Image) {
	idx := map[[2]string][]int{}
	for i, d := range res.Diffs {
		k := [2]string{d.Go, d.Tcl}
		idx[k] = append(idx[k], i)
	}
	for _, p := range res.Pairs {
		if p.Go.parent == nil || hasMappedChildren(p.Go) || hasMappedChildren(p.Tcl) {
			continue
		}
		g, t := p.Go.w, p.Tcl.w
		if g.W != t.W || g.H != t.H || g.W <= 0 || g.H <= 0 {
			continue
		}
		pct := cropDiff(gi, ti, g.X, g.Y, t.X, t.Y, g.W, g.H)
		k := [2]string{g.Path, t.Path}
		if ids := idx[k]; len(ids) > 0 {
			for _, i := range ids {
				v := pct
				res.Diffs[i].Pixels = &v
			}
			continue
		}
		if pct > 0 {
			v := pct
			res.Diffs = append(res.Diffs, Diff{Kind: KindRender, KindStr: KindRender.String(), Class: t.Class,
				Go: g.Path, Tcl: t.Path, Detail: fmt.Sprintf("geometry matches, %dx%d", g.W, g.H), Pixels: &v})
		}
	}
}

func hasMappedChildren(n *node) bool {
	for _, c := range n.children {
		if c.w.Mapped {
			return true
		}
	}
	return false
}

// cropDiff returns the percentage of pixels in the w×h crops that differ by
// more than a small per-channel tolerance (absorbs anti-aliasing noise).
func cropDiff(gi, ti image.Image, gx, gy, tx, ty, w, h int) float64 {
	const tol = 24 << 8
	diff, total := 0, 0
	gb, tb := gi.Bounds(), ti.Bounds()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			gp := image.Pt(gb.Min.X+gx+x, gb.Min.Y+gy+y)
			tp := image.Pt(tb.Min.X+tx+x, tb.Min.Y+ty+y)
			if !gp.In(gb) || !tp.In(tb) {
				continue
			}
			total++
			r1, g1, b1, _ := gi.At(gp.X, gp.Y).RGBA()
			r2, g2, b2, _ := ti.At(tp.X, tp.Y).RGBA()
			if absDiff(r1, r2) > tol || absDiff(g1, g2) > tol || absDiff(b1, b2) > tol {
				diff++
			}
		}
	}
	if total == 0 {
		return 0
	}
	return float64(diff) * 100 / float64(total)
}

func absDiff(a, b uint32) uint32 {
	if a > b {
		return a - b
	}
	return b - a
}
