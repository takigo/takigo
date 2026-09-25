package main

import (
	"image"
	"image/color"
	"testing"

	"github.com/msorc/takigo/internal/treedump"
)

func wd(path, class, parent, top string, index, x, y, w, h int) treedump.Widget {
	return treedump.Widget{Path: path, Class: class, Parent: parent, Toplevel: top, Index: index,
		X: x, Y: y, W: w, H: h, ReqW: w, ReqH: h, Mapped: true}
}

func kinds(res *Result) map[Kind]int {
	m := map[Kind]int{}
	for _, d := range res.Diffs {
		m[d.Kind]++
	}
	return m
}

func TestCompare(t *testing.T) {
	tclTree := &treedump.Dump{
		Toplevels: []treedump.Toplevel{{Path: ".demo", Title: "Demo", W: 200, H: 100}},
		Widgets: []treedump.Widget{
			wd(".demo", "Toplevel", "", ".demo", 0, 0, 0, 200, 100),
			wd(".demo.msg", "Label", ".demo", ".demo", 0, 0, 0, 200, 40),
			wd(".demo.b1", "Button", ".demo", ".demo", 1, 50, 40, 100, 30),
			wd(".demo.b2", "Button", ".demo", ".demo", 2, 50, 70, 100, 30),
		},
	}

	tests := []struct {
		name    string
		widgets []treedump.Widget
		top     treedump.Toplevel
		want    map[Kind]int
	}{
		{
			name: "identical through transparent wrapper frame and renamed widgets",
			top:  treedump.Toplevel{Path: ".", Title: "Demo", W: 200, H: 100},
			widgets: []treedump.Widget{
				wd(".", "", "", ".", 0, 0, 0, 200, 100),
				wd(".f", "Frame", ".", ".", 0, 0, 0, 200, 100),
				wd(".f.msg", "Label", ".f", ".", 0, 0, 0, 200, 40),
				wd(".f.btn_a", "Button", ".f", ".", 1, 50, 40, 100, 30),
				wd(".f.btn_b", "Button", ".f", ".", 2, 50, 70, 100, 30),
			},
			want: map[Kind]int{},
		},
		{
			name: "reqsize diff shifts sibling",
			top:  treedump.Toplevel{Path: ".", Title: "Demo", W: 200, H: 98},
			widgets: []treedump.Widget{
				wd(".", "", "", ".", 0, 0, 0, 200, 98),
				wd(".msg", "Label", ".", ".", 0, 0, 0, 200, 40),
				wd(".b1", "Button", ".", ".", 1, 50, 40, 100, 28),
				wd(".b2", "Button", ".", ".", 2, 50, 68, 100, 30),
			},
			want: map[Kind]int{KindToplevel: 1, KindReqSize: 1, KindPos: 1},
		},
		{
			name: "missing and class mismatch",
			top:  treedump.Toplevel{Path: ".", Title: "Demo", W: 200, H: 100},
			widgets: []treedump.Widget{
				wd(".", "", "", ".", 0, 0, 0, 200, 100),
				wd(".msg", "Label", ".", ".", 0, 0, 0, 200, 40),
				wd(".b1", "TButton", ".", ".", 1, 50, 40, 100, 30),
			},
			want: map[Kind]int{KindClass: 1, KindMissing: 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goTree := &treedump.Dump{Toplevels: []treedump.Toplevel{tt.top}, Widgets: tt.widgets}
			res := Compare(goTree, tclTree)
			got := kinds(res)
			if len(got) != len(tt.want) {
				t.Fatalf("kinds = %v, want %v; diffs: %+v", got, tt.want, res.Diffs)
			}
			for k, n := range tt.want {
				if got[k] != n {
					t.Fatalf("kinds = %v, want %v; diffs: %+v", got, tt.want, res.Diffs)
				}
			}
		})
	}
}

func TestCompareFonts(t *testing.T) {
	f := treedump.Font{Size: 10, Weight: "normal", Slant: "roman", Ascent: 13, Descent: 4, Linespace: 17, Sample: 377}
	g := f
	g.Size, g.Sample = 9, 356
	res := Compare(
		&treedump.Dump{Fonts: map[string]treedump.Font{"TkDefaultFont": g, "TkFixedFont": f}},
		&treedump.Dump{Fonts: map[string]treedump.Font{"TkDefaultFont": f, "TkFixedFont": f, "TkIconFont": f}},
	)
	if got := kinds(res)[KindFont]; got != 2 {
		t.Fatalf("font diffs = %d, want 2 (TkDefaultFont differs, TkIconFont missing): %+v", got, res.Diffs)
	}
}

func TestCropDiff(t *testing.T) {
	a := image.NewRGBA(image.Rect(0, 0, 10, 10))
	b := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			a.Set(x, y, color.White)
			b.Set(x, y, color.White)
		}
	}
	b.Set(1, 1, color.Black)
	b.Set(2, 2, color.RGBA{250, 250, 250, 255})
	if got := cropDiff(a, b, 0, 0, 0, 0, 10, 10); got != 1 {
		t.Fatalf("cropDiff = %v, want 1 (one pixel of 100 beyond tolerance)", got)
	}
	if got := cropDiff(a, b, 5, 5, 5, 5, 5, 5); got != 0 {
		t.Fatalf("cropDiff = %v, want 0", got)
	}
}
