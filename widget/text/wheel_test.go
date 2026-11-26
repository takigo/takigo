package text

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/takigo/takigo/color"
	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/internal/stubs"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/window"
)

// pictureWidget is a 400x200 text of 40 lines with a 400px image alone on
// line 10, as a help page with a picture.
func pictureWidget(t *testing.T) *TextWidget {
	t.Helper()
	var sb strings.Builder
	for i := range 40 {
		if i > 0 {
			sb.WriteByte('\n')
		}
		if i != 9 {
			sb.WriteString("some words on a line")
		}
	}
	tw, _ := newTestPeer(docWithText(sb.String()), 400, 200)
	if err := tw.ImageCreate("10.0", testImage{"pic", 300, 400}); err != nil {
		t.Fatal(err)
	}
	if h := tw.displayLinePixels(10)[0]; h != 400 {
		t.Fatalf("the image line is %d px, want 400", h)
	}
	return tw
}

func notch(tw *TextWidget, delta int) {
	tw.handleWheel(&event.Event{Type: event.MouseWheelType, Delta: delta})
}

// TestWheelScrollsSteadyPixels checks that every notch of the wheel moves
// the view by the same number of pixels (text.tcl's MouseWheel scrolls
// pixels), through text and through the tall image alike, until the end.
func TestWheelScrollsSteadyPixels(t *testing.T) {
	tw := pictureWidget(t)
	step := tw.wheelPixels()
	if step != 30 {
		t.Fatalf("a notch at 96 DPI is %d px, want Tk's 30", step)
	}
	maxTop := tw.layout.totalPixels() - 200
	for i := 1; tw.viewTop() < maxTop; i++ {
		before := tw.viewTop()
		notch(tw, -120)
		if got, want := tw.viewTop()-before, min(step, maxTop-before); got != want {
			t.Fatalf("notch %d (top %d, line %d): moved %d px, want %d", i, before, tw.topLine, got, want)
		}
	}
	for tw.viewTop() > 0 {
		before := tw.viewTop()
		notch(tw, 120)
		if got := before - tw.viewTop(); got != min(step, before) {
			t.Fatalf("notch up from %d moved %d px, want %d", before, got, min(step, before))
		}
	}
}

// TestWheelLeavesImagePartlyAbove checks that the view can stop with the
// image partly above it, and that bbox, dlineinfo and @x,y see it there.
func TestWheelLeavesImagePartlyAbove(t *testing.T) {
	tw := pictureWidget(t)
	imgTop := tw.displayLineTop(10, 0)
	for tw.viewTop() < imgTop+250 {
		notch(tw, -120)
	}
	off := tw.viewTop() - imgTop
	if tw.topLine != 10 || tw.topPixelOffset != off || off <= 0 || off >= 400 {
		t.Fatalf("view at line %d offset %d, want line 10 with %d px of it above", tw.topLine, tw.topPixelOffset, off)
	}
	x, y, w, h, ok := tw.BBox("10.0")
	if !ok || y != -off || w != 300 || h != 400 || x != 0 {
		t.Errorf("bbox 10.0 = %d,%d %dx%d %v, want 0,%d 300x400", x, y, w, h, ok, -off)
	}
	if _, y, _, h, _, ok := tw.DLineInfo("10.0"); !ok || y != -off || h != 400 {
		t.Errorf("dlineinfo 10.0 y=%d h=%d %v, want y=%d h=400", y, h, ok, -off)
	}
	// The line under the image is cut at the bottom of the view.
	dls := tw.computeVisibleLines()
	last := dls[len(dls)-1].logicalLine
	if _, y, _, h, _, ok := tw.DLineInfo(Index{last, 0}); !ok || y+h != 200 || h >= 13 {
		t.Errorf("dlineinfo %d.0 y=%d h=%d %v, want it cut at 200", last, y, h, ok)
	}
	if got := tw.indexFromPixel(5, 0); got != (Index{10, 0}) {
		t.Errorf("@5,0 = %v, want 10.0, the image", got)
	}
	if got := tw.indexFromPixel(1, 400-off); got != (Index{11, 0}) {
		t.Errorf("@1,%d = %v, want 11.0, the line under the image", 400-off, got)
	}
	// The scrollbar reports the same pixel back.
	first, _ := tw.yviewFractions()
	tw2 := pictureWidget(t)
	tw2.YViewMoveTo(first)
	if tw2.viewTop() != tw.viewTop() {
		t.Errorf("yview moveto %v gives top %d, want %d", first, tw2.viewTop(), tw.viewTop())
	}
}

// TestWheelEndClampsExactly checks that scrolling past the end stops with
// the last pixel at the bottom of the view, the top line partly hidden.
func TestWheelEndClampsExactly(t *testing.T) {
	tw := pictureWidget(t)
	for range 200 {
		notch(tw, -120)
	}
	total := tw.layout.totalPixels()
	if got := tw.viewTop(); got != total-200 {
		t.Fatalf("at the end the top is %d, want %d", got, total-200)
	}
	if _, last := tw.yviewFractions(); last != 1 {
		t.Errorf("at the end the view shows up to %v, want 1", last)
	}
	if (total-200)%13 != 0 && tw.topPixelOffset == 0 {
		t.Errorf("the end top is on a line boundary; want the top line partly hidden")
	}
	_, y, _, h, ok := tw.BBox("end")
	if !ok || y+h > 200 {
		t.Errorf("bbox end = y %d h %d %v, want the last line whole in the view", y, h, ok)
	}
}

// TestWheelHighResolutionAccumulates checks that the small deltas of a
// high-resolution wheel add up to a notch's pixels.
func TestWheelHighResolutionAccumulates(t *testing.T) {
	tw := pictureWidget(t)
	for range 4 {
		notch(tw, -30)
	}
	if got := tw.viewTop(); got != 30 {
		t.Errorf("four quarter notches scrolled %d px, want 30", got)
	}
	tw.Configure(WheelStep(52))
	notch(tw, -120)
	if got := tw.viewTop(); got != 82 {
		t.Errorf("a notch with WheelStep(52) went to %d, want 82", got)
	}
}

// drawRecorder records the strings, images and fills renderTo draws.
type drawRecorder struct {
	stubs.Server
	strings []drawnString
	fills   []fill
	fg      uint64
}

type drawnString struct {
	x, y int
	s    string
}

type fill struct {
	x, y, w, h int
	pixel      uint64
}

func (r *drawRecorder) SetForeground(_ platform.GCID, pixel uint64) { r.fg = pixel }
func (r *drawRecorder) FillRectangle(_ platform.DrawableID, _ platform.GCID, x, y int, w, h uint) {
	r.fills = append(r.fills, fill{x, y, int(w), int(h), r.fg})
}
func (*drawRecorder) SetTSOrigin(platform.GCID, int, int) {}

// recFont is fixedFont that draws into a drawRecorder.
type recFont struct {
	fixedFont
	r *drawRecorder
}

func (f recFont) DrawString(_ platform.DrawableID, x, y int, s string, _ uint64, _, _, _ uint16) {
	f.r.strings = append(f.r.strings, drawnString{x, y, s})
}

// TestPartialTopHitTestMatchesDrawing checks that at a partial top offset
// the text is drawn where bbox says and @x,y finds the character drawn at
// a point, and that the -pady strip above the text area is cleared.
func TestPartialTopHitTestMatchesDrawing(t *testing.T) {
	var sb strings.Builder
	for i := range 30 {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString("abcdefghij")
	}
	tw, _ := newTestPeer(docWithText(sb.String()), 200, 100)
	rec := &drawRecorder{}
	tw.Win.Display = &window.Display{Server: rec}
	tw.Font = recFont{r: rec}
	tw.Foreground = &color.Color{}
	tw.PadY = 4
	tw.insetY = 4
	tw.YViewScrollPixels(13*5 + 6) // line 6 has 6 of its 13 px above the view
	if tw.topLine != 6 || tw.topPixelOffset != 6 {
		t.Fatalf("view at line %d offset %d, want 6/6", tw.topLine, tw.topPixelOffset)
	}
	tw.renderTo(0)
	if len(rec.strings) == 0 {
		t.Fatal("nothing drawn")
	}
	for i, ds := range rec.strings {
		line := tw.topLine + i
		x, y, _, h, ok := tw.BBox(Index{line, 3})
		if !ok {
			t.Errorf("line %d drawn at y %d but bbox says it is off screen", line, ds.y)
			continue
		}
		// The baseline is the font's ascent (10) below the box's top.
		if y+10 != ds.y || x != ds.x+3*7 {
			t.Errorf("line %d: drawn at %d,%d, bbox of char 3 at %d,%d", line, ds.x, ds.y, x, y)
		}
		mid := max(y+h/2, tw.insetY)
		if got := tw.indexFromPixel(x+2, mid); got != (Index{line, 3}) {
			t.Errorf("@%d,%d = %v, want %d.3 drawn there", x+2, mid, got, line)
		}
		if utf8.RuneCountInString(ds.s) != 10 {
			t.Errorf("line %d drew %q", line, ds.s)
		}
	}
	if first := rec.strings[0]; first.y != 4-6+10 {
		t.Errorf("the top line's baseline is at %d, want 8 (6 px above the area)", first.y)
	}
	// The last fills clear the -pady strips over what the lines drew.
	n := len(rec.fills)
	if n < 2 || rec.fills[n-2] != (fill{0, 0, 200, 4, 0xFFFFFF}) {
		t.Errorf("the top -pady strip is not cleared last: %v", rec.fills[max(n-2, 0):])
	}
}
