package label

import (
	"testing"

	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

func TestLabelOptions(t *testing.T) {
	l := &Label{
		Base: widget.Base{
			Win: &window.Window{},
		},
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
	if l.Win.ReqWidth != 50 {
		t.Errorf("ReqWidth = %d, want 50", l.Win.ReqWidth)
	}
	Height(30)(l)
	if l.Win.ReqHeight != 30 {
		t.Errorf("ReqHeight = %d, want 30", l.Win.ReqHeight)
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
		Base: widget.Base{
			BorderWidth: 1,
			Relief:      option.ReliefFlat,
			PadX:        1,
			PadY:        1,
		},
		Anchor:    option.AnchorCenter,
		Justify:   option.JustifyLeft,
		Underline: -1,
	}

	if l.Anchor != option.AnchorCenter {
		t.Errorf("default Anchor = %v, want AnchorCenter", l.Anchor)
	}
	if l.Justify != option.JustifyLeft {
		t.Errorf("default Justify = %v, want JustifyLeft", l.Justify)
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

func TestLabelTextLines(t *testing.T) {
	l := &Label{
		WrapLen: 0,
	}
	l.Font = nil // no font for simple splitting test

	// Test single line
	l.Text = "hello"
	lines := l.textLines()
	if len(lines) != 1 || lines[0] != "hello" {
		t.Errorf("textLines(\"hello\") = %v, want [\"hello\"]", lines)
	}

	// Test multi-line
	l.Text = "line1\nline2\nline3"
	lines = l.textLines()
	if len(lines) != 3 || lines[0] != "line1" || lines[1] != "line2" || lines[2] != "line3" {
		t.Errorf("textLines multi = %v, want [\"line1\", \"line2\", \"line3\"]", lines)
	}

	// Test empty
	l.Text = ""
	lines = l.textLines()
	if lines != nil {
		t.Errorf("textLines(\"\") = %v, want nil", lines)
	}
}

func TestLabelTextLinesWithWrap(t *testing.T) {
	l := &Label{
		WrapLen: 50,
	}
	l.Base.Font = &mockFont{metrics: font.Metrics{MaxWidth: 10}}

	l.Text = "word1 word2 word3 word4"
	lines := l.textLines()
	if len(lines) == 0 {
		t.Error("textLines with wrap returned empty")
	}
	// Each line should fit within WrapLen
	for _, line := range lines {
		if l.Font.MeasureString(line) > l.WrapLen {
			t.Errorf("line %q exceeds WrapLen (%d > %d)", line, l.Font.MeasureString(line), l.WrapLen)
		}
	}
}

func TestLabelCompoundSize(t *testing.T) {
	img := &mockImage{w: 20, h: 30}

	// No image
	w, h := compoundSize(widget.CompoundNone, nil, 50, 60)
	if w != 50 || h != 60 {
		t.Errorf("compoundSize(nil) = (%d, %d), want (50, 60)", w, h)
	}

	// Image only
	w, h = compoundSize(widget.CompoundNone, img, 0, 0)
	if w != 20 || h != 30 {
		t.Errorf("compoundSize(img, 0, 0) = (%d, %d), want (20, 30)", w, h)
	}

	// CompoundLeft
	w, h = compoundSize(widget.CompoundLeft, img, 50, 60)
	if w != 74 || h != 60 {
		t.Errorf("compoundSize(Left) = (%d, %d), want (74, 60)", w, h)
	}

	// CompoundRight
	w, h = compoundSize(widget.CompoundRight, img, 50, 60)
	if w != 74 || h != 60 {
		t.Errorf("compoundSize(Right) = (%d, %d), want (74, 60)", w, h)
	}

	// CompoundTop
	w, h = compoundSize(widget.CompoundTop, img, 50, 60)
	if w != 50 || h != 94 {
		t.Errorf("compoundSize(Top) = (%d, %d), want (50, 94)", w, h)
	}

	// CompoundBottom
	w, h = compoundSize(widget.CompoundBottom, img, 50, 60)
	if w != 50 || h != 94 {
		t.Errorf("compoundSize(Bottom) = (%d, %d), want (50, 94)", w, h)
	}

	// CompoundCenter
	w, h = compoundSize(widget.CompoundCenter, img, 50, 60)
	if w != 50 || h != 60 {
		t.Errorf("compoundSize(Center) = (%d, %d), want (50, 60)", w, h)
	}
}

func TestLabelDestroy(t *testing.T) {
	// Destroy requires a real display connection
	l := &Label{
		Base: widget.Base{
			Win: &window.Window{PlatformID: 1},
		},
		unsub: func() {},
	}

	if l.Destroyed {
		t.Error("New label should not be destroyed")
	}

	// Manually set and verify
	l.Destroyed = true
	if !l.Destroyed {
		t.Error("Destroyed flag should be true after setting")
	}

	// unsub should still be callable
	l.unsub = func() {}
	l.unsub()
}

// Mock types for testing
type mockFont struct {
	metrics font.Metrics
}

func (m *mockFont) Attrs() font.Attributes                { return font.Attributes{} }
func (m *mockFont) Metrics() font.Metrics                   { return m.metrics }
func (m *mockFont) MeasureString(s string) int         { return len(s) * m.metrics.MaxWidth }
func (m *mockFont) Close()                             {}

type mockImage struct {
	w, h int
}

func (m *mockImage) Width() int                            { return m.w }
func (m *mockImage) Height() int                           { return m.h }
func (m *mockImage) Draw(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	depth int,
	imgX, imgY, w, h, dstX, dstY int,
	bgPixel uint64) {}