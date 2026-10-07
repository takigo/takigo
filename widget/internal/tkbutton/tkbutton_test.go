package tkbutton

import (
	"testing"

	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/window"
)

// fixedFont is 10px per character, 15px per line.
type fixedFont struct{}

func (fixedFont) Attrs() font.Attributes { return font.Attributes{} }
func (fixedFont) Metrics() font.Metrics  { return font.Metrics{MaxWidth: 10, Ascent: 12, Descent: 3} }
func (fixedFont) MeasureString(s string) int {
	return 10 * len([]rune(s))
}
func (fixedFont) Close() {}

type fakeImage struct{ w, h int }

func (i fakeImage) Width() int  { return i.w }
func (i fakeImage) Height() int { return i.h }
func (fakeImage) Draw(platform.DisplayServer, platform.DrawableID, platform.GCID, int, int, int, int, int, int, int, uint64) {
}

func newWidget(t Type) (*widget.Base, *Shared) {
	b := &widget.Base{Win: &window.Window{}, Font: fixedFont{}}
	s := &Shared{}
	Init(s, t)
	return b, s
}

func TestTextLines(t *testing.T) {
	s := &Shared{}
	for _, tt := range []struct {
		text string
		wrap int
		want []string
	}{
		{"hello", 0, []string{"hello"}},
		{"line1\nline2\nline3", 0, []string{"line1", "line2", "line3"}},
		{"", 0, []string{""}},
		{"word1 word2 word3", 60, []string{"word1", "word2", "word3"}},
	} {
		s.Text, s.WrapLen = tt.text, tt.wrap
		got := s.TextLines(fixedFont{})
		if len(got) != len(tt.want) {
			t.Errorf("TextLines(%q, wrap %d) = %q, want %q", tt.text, tt.wrap, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("TextLines(%q, wrap %d) = %q, want %q", tt.text, tt.wrap, got, tt.want)
				break
			}
		}
	}
}

// The request of a compound widget is the image and text laid out by
// -compound, plus padding and border (TkpComputeButtonGeometry).
func TestComputeGeometryCompound(t *testing.T) {
	for _, tt := range []struct {
		compound widget.Compound
		w, h     int
	}{
		{widget.CompoundNone, 20, 30},
		{widget.CompoundLeft, 73, 60},
		{widget.CompoundRight, 73, 60},
		{widget.CompoundTop, 50, 92},
		{widget.CompoundBottom, 50, 92},
		{widget.CompoundCenter, 50, 60},
	} {
		b, s := newWidget(TypeLabel)
		b.PadX, b.PadY, b.BorderWidth = 3, 2, 1
		s.Img = fakeImage{20, 30}
		s.Text = "abcde\nabcd\nabc\nab" // 50 x 60
		s.Compound = tt.compound
		ComputeGeometry(b, s)
		wantW, wantH := tt.w+2*b.BorderWidth, tt.h+2*b.BorderWidth
		if tt.compound != widget.CompoundNone {
			wantW += 2 * b.PadX
			wantH += 2 * b.PadY
		}
		if b.Win.ReqWidth != wantW || b.Win.ReqHeight != wantH {
			t.Errorf("compound %v: request %dx%d, want %dx%d", tt.compound, b.Win.ReqWidth, b.Win.ReqHeight, wantW, wantH)
		}
	}
}

func TestComputeGeometryByType(t *testing.T) {
	// Text "ab": 20 x 15; padding 1; border 1; highlight 0.
	for _, tt := range []struct {
		typ          Type
		widthChars   int
		indicator    bool
		wantW, wantH int
	}{
		{TypeLabel, 0, false, 20 + 2 + 2, 15 + 2 + 2},
		{TypeButton, 0, false, 20 + 2 + 2 + 2, 15 + 2 + 2 + 2},
		{TypeLabel, 5, false, 50 + 2 + 2, 15 + 2 + 2},
		{TypeCheck, 0, true, 20 + 2 + 2 + (15 + 10), 15 + 2 + 2},
		{TypeRadio, 4, true, 40 + 2 + 2 + (15 + 10), 15 + 2 + 2},
		{TypeCheck, 0, false, 20 + 2 + 2, 15 + 2 + 2},
	} {
		b, s := newWidget(tt.typ)
		b.PadX, b.PadY, b.BorderWidth = 1, 1, 1
		s.Text, s.WidthChars, s.IndicatorOn = "ab", tt.widthChars, tt.indicator
		ComputeGeometry(b, s)
		if b.Win.ReqWidth != tt.wantW || b.Win.ReqHeight != tt.wantH {
			t.Errorf("type %d width %d indicator %v: request %dx%d, want %dx%d",
				tt.typ, tt.widthChars, tt.indicator, b.Win.ReqWidth, b.Win.ReqHeight, tt.wantW, tt.wantH)
		}
		if w, h := s.TextSize(); w != 20 || h != 15 {
			t.Errorf("TextSize = %dx%d, want 20x15", w, h)
		}
	}
}

// A button's -default leaves 5 pixels around it; an image is not padded.
func TestComputeGeometryDefaultRingAndImage(t *testing.T) {
	b, s := newWidget(TypeButton)
	b.PadX, b.PadY, b.BorderWidth, b.HighlightWidth = 3, 2, 1, 1
	s.Img = fakeImage{20, 30}
	ComputeGeometry(b, s)
	if b.Win.ReqWidth != 20+2+4 || b.Win.ReqHeight != 30+2+4 {
		t.Errorf("image button request %dx%d, want 26x36", b.Win.ReqWidth, b.Win.ReqHeight)
	}
	s.Default = DefaultNormal
	ComputeGeometry(b, s)
	if b.Win.ReqWidth != 20+2+4+10 || Inset(b, s) != 7 {
		t.Errorf("-default normal: request width %d inset %d, want 36 and 7", b.Win.ReqWidth, Inset(b, s))
	}
}

func TestShift(t *testing.T) {
	_, s := newWidget(TypeButton)
	for _, tt := range []struct {
		relief option.Relief
		winW   int
		wantX  int
	}{
		{option.ReliefRaised, 30, 0},
		{option.ReliefFlat, 30, 0},   // 30-20 even: one back
		{option.ReliefFlat, 31, 1},   // odd
		{option.ReliefSunken, 30, 1}, // 2, one back
		{option.ReliefSunken, 31, 2},
		{option.ReliefRidge, 30, 1}, // ridge never takes one back
	} {
		x, _ := s.shift(tt.winW, 31, tt.relief, 0, 0, 20, 20)
		if x != tt.wantX {
			t.Errorf("shift(%v, width %d) = %d, want %d", tt.relief, tt.winW, x, tt.wantX)
		}
	}
	_, l := newWidget(TypeLabel)
	if x, _ := l.shift(31, 31, option.ReliefSunken, 0, 0, 0, 0); x != 0 {
		t.Errorf("a label shifted its content by %d", x)
	}
}
