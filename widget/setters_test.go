package widget

import (
	"bytes"
	"errors"
	"log/slog"
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
	log    *bytes.Buffer
}

func newFakeApp() fakeApp {
	return fakeApp{colors: color.NewCache(0), fonts: font.NewRegistry(stubOpener{}), log: &bytes.Buffer{}}
}

func (a fakeApp) ColorCache() *color.Cache     { return a.colors }
func (a fakeApp) FontRegistry() *font.Registry { return a.fonts }
func (fakeApp) DoWhenIdle(func())              {}
func (a fakeApp) Logger() *slog.Logger {
	return slog.New(slog.NewTextHandler(a.log, nil))
}

func newTestBase(class string) *Base {
	return &Base{Win: &window.Window{Class: class}, App: newFakeApp()}
}

func TestSetBackgroundColor(t *testing.T) {
	b := newTestBase("Button")
	if !b.SetBackgroundColor("red") {
		t.Fatal("SetBackgroundColor(red) = false")
	}
	if b.Background == nil || b.Background.Red != 0xffff || b.Border == nil {
		t.Fatalf("Background = %+v, Border = %v", b.Background, b.Border)
	}
	old := b.Background

	// Outside Configure (a constructor) the failure is logged.
	buf := b.App.(fakeApp).log
	if b.SetBackgroundColor("no-such-colour") {
		t.Fatal("SetBackgroundColor(bad) = true")
	}
	if b.Background != old {
		t.Fatal("bad name replaced the background")
	}
	if !strings.Contains(buf.String(), "option not applied") || !strings.Contains(buf.String(), "no-such-colour") {
		t.Fatalf("log = %q", buf.String())
	}

	// Inside Configure it is returned instead.
	buf.Reset()
	b.BeginOptions()
	b.SetBackgroundColor("no-such-colour")
	b.SetForegroundColor(color.RGB(0, 0, 255))
	err := b.EndOptions()
	if !errors.Is(err, color.ErrUnknown) {
		t.Fatalf("EndOptions() = %v, want color.ErrUnknown", err)
	}
	if buf.Len() != 0 {
		t.Fatalf("collected error was also logged: %q", buf.String())
	}
	if b.Foreground == nil || b.Foreground.Blue != 0xffff || b.Foreground.Red != 0 {
		t.Fatalf("Foreground = %+v, want blue", b.Foreground)
	}
	if b.EndOptions() != nil {
		t.Fatal("EndOptions did not reset the collected errors")
	}
}

func TestSetFont(t *testing.T) {
	b := newTestBase("")
	if !b.SetFont("Helvetica 12") {
		t.Fatal("SetFontName = false")
	}
	if b.Font.Attrs().Family != "Helvetica" {
		t.Fatalf("Font = %+v", b.Font.Attrs())
	}
	old := b.Font
	b.BeginOptions()
	if b.SetFont("nofont 12") {
		t.Fatal("SetFont(nofont) = true")
	}
	if err := b.EndOptions(); !errors.Is(err, font.ErrNotFound) {
		t.Fatalf("EndOptions() = %v, want font.ErrNotFound", err)
	}
	if b.Font != old {
		t.Fatal("bad name replaced the font")
	}
	if !b.SetFont(font.Attributes{Family: "Courier", Size: 10}) || b.Font.Attrs().Family != "Courier" {
		t.Fatalf("SetFont(Attributes): Font = %+v", b.Font.Attrs())
	}
}
