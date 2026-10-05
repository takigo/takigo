package canvas

import (
	"strings"
	"testing"

	"github.com/takigo/takigo/color"
	"github.com/takigo/takigo/font"
)

// stubFont implements font.Font with whatever Attributes the test passes.
type stubFont struct {
	fam  string
	size float64
	wt   font.Weight
	slt  font.Slant
}

func (s *stubFont) Attrs() font.Attributes {
	return font.Attributes{Family: s.fam, Size: s.size, Weight: s.wt, Slant: s.slt}
}
func (s *stubFont) Metrics() font.Metrics    { return font.Metrics{} }
func (s *stubFont) MeasureString(string) int { return 0 }
func (s *stubFont) Close()                   {}

func TestPsY(t *testing.T) {
	ps := NewPSContext()
	ps.Y2 = 100
	if got := ps.PsY(0); got != 100 {
		t.Errorf("PsY(0) = %v, want 100", got)
	}
	if got := ps.PsY(100); got != 0 {
		t.Errorf("PsY(100) = %v, want 0", got)
	}
	if got := ps.PsY(50); got != 50 {
		t.Errorf("PsY(50) = %v, want 50", got)
	}
}

func TestPath(t *testing.T) {
	ps := NewPSContext()
	ps.Y2 = 100
	ps.Path([]float64{10, 20, 30, 40})
	out := ps.String()
	if !strings.Contains(out, "10 80 moveto") { // PsY(20) = 80
		t.Errorf("expected moveto with PsY conversion; got %q", out)
	}
	if !strings.Contains(out, "30 60 lineto") { // PsY(40) = 60
		t.Errorf("expected lineto; got %q", out)
	}
}

func TestColorRGB(t *testing.T) {
	ps := NewPSContext()
	ps.ColorLevel = 2
	c := &color.ColorRef{Red: 0x8000, Green: 0x4000, Blue: 0x0000}
	ps.Color(c)
	out := ps.String()
	want := "setrgbcolor AdjustColor"
	if !strings.Contains(out, want) {
		t.Errorf("missing %q in %q", want, out)
	}
}

func TestColorMono(t *testing.T) {
	ps := NewPSContext()
	ps.ColorLevel = 0
	ps.Color(&color.ColorRef{Red: 0xFFFF, Green: 0x0000, Blue: 0x0000})
	if !strings.Contains(ps.String(), "setgray") {
		t.Errorf("expected setgray with mono colormode; got %q", ps.String())
	}
}

func TestColorGray(t *testing.T) {
	ps := NewPSContext()
	ps.ColorLevel = 1
	// Pure red → 0.299
	ps.Color(&color.ColorRef{Red: 0xFFFF, Green: 0x0000, Blue: 0x0000})
	if !strings.Contains(ps.String(), "0.299 setgray") {
		t.Errorf("expected Rec.601 luma 0.299; got %q", ps.String())
	}
}

func TestColorPrepassSkips(t *testing.T) {
	ps := NewPSContext()
	ps.Prepass = true
	ps.ColorLevel = 2
	ps.Color(&color.ColorRef{Red: 0xFFFF})
	if ps.String() != "" {
		t.Errorf("prepass should not emit color; got %q", ps.String())
	}
}

func TestFontNameAliases(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Arial", "Helvetica"},
		{"arial", "Helvetica"},
		{"Geneva", "Helvetica"},
		{"Times New Roman", "Times"},
		{"New York", "Times"},
		{"Courier New", "Courier"},
		{"Monaco", "Courier"},
	}
	for _, c := range cases {
		got := familyAlias(c.in)
		if got != c.want {
			t.Errorf("familyAlias(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFontNameNewCenturySchoolbook(t *testing.T) {
	got := familyAlias("NewCenturySchoolbook")
	if got != "NewCenturySchoolbook" {
		t.Errorf("expected unchanged; got %q", got)
	}
	got = titleCase("new century schoolbook")
	if got != "NewCenturySchoolbook" {
		t.Errorf("titleCase result = %q, want NewCenturySchoolbook", got)
	}
}

func TestFontEmit(t *testing.T) {
	cases := []struct {
		fam          string
		size         float64
		wantContains string
	}{
		{"Helvetica", 12, "/Helvetica findfont 12 scalefont ISOEncode setfont"},
		{"Times", 14, "/Times-Roman findfont 14 scalefont ISOEncode setfont"},
		{"Courier", 10, "/Courier findfont 10 scalefont ISOEncode setfont"},
		{"Symbol", 12, "/Symbol findfont 12 scalefont setfont"}, // no ISOEncode for Symbol
		{"ZapfChancery", 12, "/Zapfchancery-Medium findfont 12 scalefont ISOEncode setfont"},
	}
	for _, c := range cases {
		f := &stubFont{fam: c.fam, size: c.size}
		got := psFontEmit(f)
		if !strings.Contains(got, c.wantContains) {
			t.Errorf("psFontEmit(%q, %v) = %q, want substring %q",
				c.fam, c.size, got, c.wantContains)
		}
	}
}

func TestFontRegisterOnPrepass(t *testing.T) {
	ps := NewPSContext()
	ps.Prepass = true
	f := &stubFont{fam: "Helvetica", size: 12}
	ps.Font(f)
	if len(ps.fonts) != 1 || !ps.fonts["Helvetica"] {
		t.Errorf("prepass should register font name; got %v", ps.fonts)
	}
	if ps.String() != "" {
		t.Errorf("prepass should not emit font output; got %q", ps.String())
	}
}

func TestBitmapHexEncoding(t *testing.T) {
	ps := NewPSContext()
	ps.Bitmap([]byte{0xff, 0x00, 0xaa}, 8, 1)
	out := ps.String()
	if !strings.Contains(out, "<ff00aa>") {
		t.Errorf("hex encoding wrong; got %q", out)
	}
}

func TestStippleEmitsClipAndFill(t *testing.T) {
	ps := NewPSContext()
	ps.Stipple([]byte{0xff}, 8, 1)
	out := ps.String()
	for _, want := range []string{"gsave", "StrokeClip", "imagemask", "grestore"} {
		if !strings.Contains(out, want) {
			t.Errorf("stipple output missing %q: %q", want, out)
		}
	}
}

func TestOutlineDash(t *testing.T) {
	ps := NewPSContext()
	ps.Outline(1, []int{2, 4}, 0, nil, nil)
	out := ps.String()
	if !strings.Contains(out, "[2 4] 0 setdash") {
		t.Errorf("expected dash pattern; got %q", out)
	}
}

func TestOutlineNoDash(t *testing.T) {
	ps := NewPSContext()
	ps.Outline(1, nil, 0, nil, nil)
	out := ps.String()
	if !strings.Contains(out, "] 0 setdash") {
		t.Errorf("expected empty dash; got %q", out)
	}
}

func TestEmitHeaderIncludesFonts(t *testing.T) {
	ps := NewPSContext()
	ps.fonts["Helvetica"] = true
	ps.fonts["Times"] = true
	ps.Title = "test"
	ps.Width, ps.Height = 200, 100
	ps.EmitHeader()
	out := ps.String()
	if !strings.Contains(out, "%!PS-Adobe-3.0 EPSF-3.0") {
		t.Errorf("missing EPSF header; got %q", out)
	}
	if !strings.Contains(out, "%%Title: test") {
		t.Errorf("missing title; got %q", out)
	}
	if !strings.Contains(out, "%%DocumentNeededResources: font Helvetica") {
		t.Errorf("missing Helvetica font resource; got %q", out)
	}
	if !strings.Contains(out, "%%EndComments") {
		t.Errorf("missing EndComments; got %q", out)
	}
}

func TestDashInts(t *testing.T) {
	if got := dashInts([]byte{2, 4, 6}); got[0] != 2 || got[1] != 4 || got[2] != 6 {
		t.Errorf("dashInts = %v", got)
	}
}

func TestPrologEmbed(t *testing.T) {
	if len(tkPsPreamble) < 5000 {
		t.Errorf("prolog suspiciously small: %d bytes", len(tkPsPreamble))
	}
	for _, want := range []string{"%%BeginProlog", "%%EndProlog", "/ISOEncode", "/StrokeClip", "/StippleFill", "/AdjustColor", "/CurrentEncoding"} {
		if !strings.Contains(tkPsPreamble, want) {
			t.Errorf("prolog missing %q", want)
		}
	}
}

func TestPsEscape(t *testing.T) {
	cases := map[string]string{
		"hello":       "(hello)",
		"a(b)c":       "(a\\(b\\)c)",
		"back\\slash": "(back\\\\slash)",
		"new\nline":   "(new\\nline)",
	}
	for in, want := range cases {
		got := "(" + psEscape(in) + ")"
		if got != want {
			t.Errorf("psEscape(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSampledSpline(t *testing.T) {
	pts := []float64{0, 0, 50, 50, 100, 0}
	out := sampleSpline(pts, 4)
	if len(out) < len(pts) {
		t.Errorf("sampleSpline returned fewer points than input: %v", out)
	}
}

func TestPostscriptEmitsEachItemOnce(t *testing.T) {
	c := newBenchCanvas()
	c.createItem(newRectOvalItem("rectangle", 10, 10, 50, 40, c), []ItemOption{fillPixel(0xff0000)})
	c.createItem(newRectOvalItem("rectangle", 60, 10, 90, 40, c), []ItemOption{fillPixel(0x00ff00)})
	out, err := c.Postscript()
	if err != nil {
		t.Fatal(err)
	}
	body := out[strings.Index(out, "%%EndSetup"):]
	if n := strings.Count(body, "stroke\n"); n != 2 {
		t.Errorf("%d outlines stroked, want 2 (one per item):\n%s", n, body)
	}
	if strings.Count(body, "gsave") != strings.Count(body, "grestore") {
		t.Errorf("unbalanced gsave/grestore:\n%s", body)
	}
}
