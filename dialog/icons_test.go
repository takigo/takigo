package dialog

import (
	"strings"
	"testing"

	"github.com/takigo/takigo/screenunit"
)

// TestBuiltinIcons checks that every built-in icon parses, has its nominal
// size at the screen's scaling and draws something: an SVG feature the
// rasterizer lacks would leave it blank or flood it.
func TestBuiltinIcons(t *testing.T) {
	scale := screenunit.ScalingPct()
	icons := map[string]struct {
		src  string
		size int
	}{
		"folder": {svgFolder, 16}, "file": {svgFile, 16}, "newfolder": {svgNewFolder, 16},
		"drive": {svgDrive, 16}, "back": {svgBack, 16}, "forward": {svgForward, 16},
		"up": {svgUp, 16}, "home": {svgHome, 16},
	}
	for name, set := range map[string]map[MessageType]string{"": msgIcons, "tk-": tkMsgIcons} {
		for typ, n := range map[MessageType]string{MsgError: "error", MsgWarning: "warning", MsgInfo: "info", MsgQuestion: "question"} {
			icons[name+n] = struct {
				src  string
				size int
			}{set[typ], 32}
		}
	}
	for name, ic := range icons {
		t.Run(name, func(t *testing.T) {
			p := svgIcon(ic.src, "#000000")
			defer p.Destroy()
			want := ic.size * scale / 100
			if p.Width() != want || p.Height() != want {
				t.Fatalf("size %dx%d, want %dx%d", p.Width(), p.Height(), want, want)
			}
			px := p.Pixels()
			covered := 0
			for i := 3; i < len(px); i += 4 {
				if px[i] > 0x80 {
					covered++
				}
			}
			if frac := float64(covered) / float64(want*want); frac < 0.05 || frac > 0.9 {
				t.Errorf("%.0f%% of the pixels are opaque, want an icon", frac*100)
			}
		})
	}
}

// TestGlyphInk checks that line art takes the ink it is given.
func TestGlyphInk(t *testing.T) {
	if !strings.Contains(svgBack, "currentColor") {
		t.Fatal("the back glyph is not drawn in currentColor")
	}
	p := svgIcon(svgBack, "#ff0000")
	defer p.Destroy()
	px := p.Pixels()
	for i := 0; i < len(px); i += 4 {
		if px[i+3] == 0xff && (px[i] != 0xff || px[i+1] != 0 || px[i+2] != 0) {
			t.Fatalf("opaque pixel %v, want red", px[i:i+4])
		}
	}
}
