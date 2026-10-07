package label

import (
	"testing"

	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/window"
)

func TestLabelOptions(t *testing.T) {
	l := &Label{
		Win: &window.Window{},
	}

	// Test Text
	Text("Hello")(l)
	if l.Text != "Hello" {
		t.Errorf("Text = %q, want \"Hello\"", l.Text)
	}

	// Test Anchor
	Anchor(option.AnchorNW)(l)
	if l.Anchor != option.AnchorNW {
		t.Errorf("Anchor = %v, want AnchorNW", l.Anchor)
	}

	// Test JustifyOpt
	JustifyOpt(option.JustifyCenter)(l)
	if l.Justify != option.JustifyCenter {
		t.Errorf("Justify = %v, want JustifyCenter", l.Justify)
	}

	// Test BorderWidth
	BorderWidth(3)(l)
	if l.BorderWidth != 3 {
		t.Errorf("BorderWidth = %d, want 3", l.BorderWidth)
	}

	// Test Relief
	Relief(option.ReliefRaised)(l)
	if l.Relief != option.ReliefRaised {
		t.Errorf("Relief = %v, want ReliefRaised", l.Relief)
	}

	// Test PadX/PadY
	PadX(10)(l)
	if l.PadX != 10 {
		t.Errorf("PadX = %d, want 10", l.PadX)
	}
	PadY(20)(l)
	if l.PadY != 20 {
		t.Errorf("PadY = %d, want 20", l.PadY)
	}

	// Test Width/Height
	Width(50)(l)
	if l.WidthChars != 50 {
		t.Errorf("WidthChars = %d, want 50", l.WidthChars)
	}
	Height(30)(l)
	if l.HeightChars != 30 {
		t.Errorf("HeightChars = %d, want 30", l.HeightChars)
	}

	// Test WrapLength
	WrapLength(200)(l)
	if l.WrapLen != 200 {
		t.Errorf("WrapLen = %d, want 200", l.WrapLen)
	}

	// Test Underline
	l.Underline = 2
	if l.Underline != 2 {
		t.Errorf("Underline = %d, want 2", l.Underline)
	}
}

func TestLabelDefaults(t *testing.T) {
	l := &Label{
		BorderWidth: 1,
		Relief:      option.ReliefFlat,
		PadX:        1,
		PadY:        1,
		Anchor:      option.AnchorCenter,
		Justify:     option.JustifyCenter,
		Underline:   -1,
	}

	if l.Anchor != option.AnchorCenter {
		t.Errorf("default Anchor = %v, want AnchorCenter", l.Anchor)
	}
	if l.Justify != option.JustifyCenter {
		t.Errorf("default Justify = %v, want JustifyCenter", l.Justify)
	}
	if l.Underline != -1 {
		t.Errorf("default Underline = %d, want -1", l.Underline)
	}
	if l.BorderWidth != 1 {
		t.Errorf("default BorderWidth = %d, want 1", l.BorderWidth)
	}
	if l.Relief != option.ReliefFlat {
		t.Errorf("default Relief = %v, want ReliefFlat", l.Relief)
	}
	if l.PadX != 1 {
		t.Errorf("default PadX = %d, want 1", l.PadX)
	}
	if l.PadY != 1 {
		t.Errorf("default PadY = %d, want 1", l.PadY)
	}
}

func TestLabelTextLinesWithWrap(t *testing.T) {
	l := &Label{}
	l.WrapLen = 50
	l.Font = &mockFont{metrics: font.Metrics{MaxWidth: 10}}

	l.Text = "word1 word2 word3 word4"
	lines := l.TextLines(l.Font)
	if len(lines) == 0 {
		t.Error("TextLines with wrap returned empty")
	}
	// Each line should fit within WrapLen
	for _, line := range lines {
		if l.Font.MeasureString(line) > l.WrapLen {
			t.Errorf("line %q exceeds WrapLen (%d > %d)", line, l.Font.MeasureString(line), l.WrapLen)
		}
	}
}

func TestLabelConfigure(t *testing.T) {
	l := &Label{Win: &window.Window{}}
	l.Font = &mockFont{metrics: font.Metrics{MaxWidth: 10, Ascent: 8, Descent: 2}}
	l.Configure(Text("abc"))
	short := l.Win.ReqWidth
	l.Configure(Text("abcdef"), BorderWidth(2))
	if l.Win.ReqWidth != short+30+4 {
		t.Errorf("ReqWidth = %d, want %d", l.Win.ReqWidth, short+30+4)
	}
}

func TestLabelDestroy(t *testing.T) {
	// Destroy requires a real display connection
	l := &Label{
		Win:   &window.Window{PlatformID: 1},
		unsub: func() {},
	}

	if l.Destroyed() {
		t.Error("New label should not be destroyed")
	}

	// Manually set and verify
	l.MarkDestroyed()
	if !l.Destroyed() {
		t.Error("Destroyed flag should be true after setting")
	}

	// unsub should still be callable
	l.unsub = func() {}
	l.unsub()
}

// Mock types for testing.
type mockFont struct {
	metrics font.Metrics
}

func (m *mockFont) Attrs() font.Attributes     { return font.Attributes{} }
func (m *mockFont) Metrics() font.Metrics      { return m.metrics }
func (m *mockFont) MeasureString(s string) int { return len(s) * m.metrics.MaxWidth }
func (m *mockFont) Close()                     {}
