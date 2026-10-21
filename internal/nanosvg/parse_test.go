package nanosvg

import "testing"

// TestParseMalformed checks that a document that is not valid XML is
// reported as an error instead of being silently truncated.
func TestParseMalformed(t *testing.T) {
	tests := []struct {
		name string
		src  string
	}{
		{"unclosed element", `<svg width="10" height="10"><rect x="0" y="0" width="5" height="5"`},
		{"bad attribute", `<svg width="10" height="10"><rect x="0" y="0" width=5 height="5"/></svg>`},
		{"stray end tag", `</svg>`},
		{"not xml", `this is not an svg`},
		{"empty", ``},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, err := Parse(tt.src)
			if err == nil {
				t.Fatalf("Parse(%q) = %v, want an error", tt.src, img)
			}
			if img == nil {
				t.Error("Parse returned no image")
			}
		})
	}
}

// TestParseValid checks that a well-formed document still parses.
func TestParseValid(t *testing.T) {
	img, err := Parse(`<svg width="20" height="10"><rect x="0" y="0" width="5" height="5" fill="#ff0000"/></svg>`)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if img.Width != 20 || img.Height != 10 {
		t.Errorf("size = %vx%v, want 20x10", img.Width, img.Height)
	}
	if len(img.shapes) != 1 {
		t.Errorf("shapes = %d, want 1", len(img.shapes))
	}
}
