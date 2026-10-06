//go:build linux || freebsd || openbsd || netbsd

package xft

import (
	"os"
	"strings"
	"testing"

	"github.com/takigo/takigo/font"

	"github.com/takigo/takigo/internal/displaylock"
	"github.com/takigo/takigo/internal/xlib"
)

func openTestXft(t testing.TB) *XftFont {
	t.Helper()
	displaylock.UseVirtualDisplay()
	if os.Getenv("DISPLAY") == "" {
		t.Skip("no display available (set DISPLAY or use xvfb-run)")
	}
	dpy, err := xlib.OpenDisplay("")
	if err != nil {
		t.Skip(err)
	}
	t.Cleanup(dpy.Close)
	s := dpy.DefaultScreen()
	f, err := OpenXft(dpy, s, dpy.DefaultVisual(s), dpy.DefaultColormap(s),
		font.Attributes{Family: "DejaVu Sans", Size: 10})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(f.Close)
	return f
}

// The cached per-rune widths must add up to what Xft measures for the
// whole string, including runes drawn with fallback fonts.
func TestXftMeasureStringMatchesXft(t *testing.T) {
	f := openTestXft(t)
	for _, s := range []string{
		"", "a", "Hello, World!", "WAVAy To.", "tab\there", "naïve café",
		"Ελληνικά", "Русский текст", "日本語のテキスト", "mixed 中文 and ASCII",
		"emoji 😀 in text", strings.Repeat("x", 500),
	} {
		cached, direct := f.MeasureString(s), f.measureUncached(s)
		if cached != direct {
			t.Errorf("MeasureString(%q) = %d, Xft measures %d", s, cached, direct)
		}
		if again := f.MeasureString(s); again != cached {
			t.Errorf("MeasureString(%q) changed from %d to %d", s, cached, again)
		}
	}
}

func BenchmarkXftMeasureString(b *testing.B) {
	f := openTestXft(b)
	const line = "The quick brown fox jumps over the lazy dog 0123456789"
	b.Run("cached", func(b *testing.B) {
		for range b.N {
			f.MeasureString(line)
		}
	})
	b.Run("xft", func(b *testing.B) {
		for range b.N {
			f.measureUncached(line)
		}
	})
}
