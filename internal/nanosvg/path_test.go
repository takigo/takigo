package nanosvg

import "testing"

// coverage rasterizes src at scale 1 and returns its alpha channel.
func coverage(t *testing.T, src string) []uint8 {
	t.Helper()
	img, err := Parse(src)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	w, h := Size(img, 1)
	px := Rasterize(img, 1, w, h)
	alpha := make([]uint8, w*h)
	for i := range alpha {
		alpha[i] = px[i*4+3]
	}
	return alpha
}

// dome renders a path in a 16x16 document, as reference for the shorthands.
func dome(t *testing.T, d string) []uint8 {
	return coverage(t, `<svg width="16" height="16" xmlns="http://www.w3.org/2000/svg"><path d="`+d+`"/></svg>`)
}

func maxDiff(a, b []uint8) int {
	d := 0
	for i := range a {
		d = max(d, abs(int(a[i])-int(b[i])))
	}
	return d
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// TestPathCurveCommands checks that the A, Q, T and S path commands trace
// the same outline as an equivalent primitive or longhand path.
func TestPathCurveCommands(t *testing.T) {
	const head = `<svg width="16" height="16" xmlns="http://www.w3.org/2000/svg">`
	circle := coverage(t, head+`<circle cx="8" cy="8" r="6"/></svg>`)
	square := coverage(t, head+`<rect x="2" y="2" width="12" height="12"/></svg>`)
	tests := []struct {
		name string
		path string
		want []uint8
	}{
		{"absolute arcs", "M2 8A6 6 0 0 1 14 8A6 6 0 0 1 2 8z", circle},
		{"relative arcs", "m2 8a6 6 0 1 0 12 0a6 6 0 1 0-12 0z", circle},
		{"compact arc flags", "m2 8a6 6 0 1012 0a6 6 0 10-12 0z", circle},
		{"arc radius scaled up", "M2 8A1 1 0 0 1 14 8A1 1 0 0 1 2 8z", circle},
		{"degenerate arc is a line", "M2 2h12A0 0 0 0 1 14 14H2z", square},
		{"smooth quadratic reflects the control point", "M2 14Q2 2 8 2T14 14z", dome(t, "M2 14Q2 2 8 2Q14 2 14 14z")},
		{"relative quadratic", "m2 14q0-12 6-12t6 12z", dome(t, "M2 14Q2 2 8 2Q14 2 14 14z")},
		{"smooth cubic reflects the control point", "M2 14C2 6 4 2 8 2S14 6 14 14z", dome(t, "M2 14C2 6 4 2 8 2C12 2 14 6 14 14z")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := coverage(t, head+`<path d="`+tt.path+`"/></svg>`)
			if d := maxDiff(got, tt.want); d > 40 {
				t.Errorf("alpha differs by up to %d from the reference shape", d)
			}
		})
	}
}

// TestPathTkQuestionIcon renders Tk's ::tk::icons::question (icons.tcl),
// whose glyph uses relative arcs: without them the "?" is a broken polygon
// that floods most of the circle.
func TestPathTkQuestionIcon(t *testing.T) {
	alpha := coverage(t, `<svg width="32" height="32" xmlns="http://www.w3.org/2000/svg">`+
		`<path d="m17.6 27.2h-3.2v-3.2h3.2zm3.312-12.4-1.44 1.472c-1.152 1.168-1.872 2.128-1.872 4.528h-3.2v-0.8c0-1.76 0.72-3.36 1.872-4.528l1.984-2.016a3.128 3.128 0 0 0 0.944-2.256c0-1.76-1.44-3.2-3.2-3.2s-3.2 1.44-3.2 3.2h-3.2c0-3.536 2.864-6.4 6.4-6.4s6.4 2.864 6.4 6.4c0 1.408-0.576 2.688-1.488 3.6z"/></svg>`)
	sum := 0
	for _, a := range alpha {
		sum += int(a)
	}
	// The glyph covers about 60 px; the area inside its bounding box is ~120.
	if px := sum / 255; px < 40 || px > 90 {
		t.Errorf("question glyph covers %d px, want 40..90", px)
	}
	if a := alpha[12*32+16]; a != 0 {
		t.Errorf("hole of the hook at (16,12) has alpha %d, want 0", a)
	}
}
