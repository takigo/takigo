package widget

import (
	"bytes"
	"errors"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/window"
)

type stubFont struct{ attrs font.Attributes }

func (f stubFont) Attrs() font.Attributes   { return f.attrs }
func (stubFont) Metrics() font.Metrics      { return font.Metrics{Ascent: 10, Descent: 3, MaxWidth: 7} }
func (stubFont) MeasureString(s string) int { return 7 * len(s) }
func (stubFont) Close()                     {}
func (stubFont) DrawString(platform.DrawableID, int, int, string, uint64, uint16, uint16, uint16) {
}

type stubOpener struct{}

func (stubOpener) OpenFont(a font.Attributes) (font.Font, error) {
	if a.Family == "nofont" {
		return nil, errors.New("no such font")
	}
	return stubFont{a}, nil
}
func (stubOpener) Families() []string { return nil }

// fakeApp is a display-free AppContext with only the resource lookups.
type appContext = AppContext

type fakeApp struct {
	appContext
	colors *color.Cache
	fonts  *font.Registry
}

func newFakeApp() fakeApp {
	return fakeApp{colors: color.NewCache(0), fonts: font.NewRegistry(stubOpener{})}
}

func (a fakeApp) ColorCache() *color.Cache     { return a.colors }
func (a fakeApp) FontRegistry() *font.Registry { return a.fonts }
func (fakeApp) DoWhenIdle(func())              {}

func newTestBase(class string) *Base {
	return &Base{Win: &window.Window{Class: class}, App: newFakeApp()}
}

func TestSetBackgroundName(t *testing.T) {
	b := newTestBase("Button")
	if !b.SetBackgroundName("red") {
		t.Fatal("SetBackgroundName(red) = false")
	}
	if b.Background == nil || b.Background.Red != 0xffff || b.Border == nil {
		t.Fatalf("Background = %+v, Border = %v", b.Background, b.Border)
	}
	old := b.Background

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)
	if b.SetBackgroundName("no-such-colour") {
		t.Fatal("SetBackgroundName(bad) = true")
	}
	if b.Background != old {
		t.Fatal("bad name replaced the background")
	}
	if !strings.Contains(buf.String(), `button: failed to get color "no-such-colour"`) {
		t.Fatalf("log = %q", buf.String())
	}
}

func TestSetFontName(t *testing.T) {
	b := newTestBase("")
	if !b.SetFontName("Helvetica 12") {
		t.Fatal("SetFontName = false")
	}
	if b.Font.Attrs().Family != "Helvetica" {
		t.Fatalf("Font = %+v", b.Font.Attrs())
	}
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)
	old := b.Font
	if b.SetFontName("nofont 12") {
		t.Fatal("SetFontName(nofont) = true")
	}
	if b.Font != old {
		t.Fatal("bad name replaced the font")
	}
	if !strings.Contains(buf.String(), `widget: failed to get font "nofont 12"`) {
		t.Fatalf("log = %q", buf.String())
	}
}
