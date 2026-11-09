package button_test

import (
	"errors"
	goimage "image"
	gocolor "image/color"
	"image/draw"
	"testing"

	"github.com/takigo/takigo/color"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/image"
	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/button"
)

func TestButtonGeometryFollowsOptions(t *testing.T) {
	app := testutil.NewTestApp(t)
	short := button.New(app, "short", button.Text("OK"))
	long := button.New(app, "long", button.Text("A much longer label"))
	if long.Win.ReqWidth <= short.Win.ReqWidth {
		t.Errorf("longer text is not wider: %d vs %d", long.Win.ReqWidth, short.Win.ReqWidth)
	}
	if long.Win.ReqHeight != short.Win.ReqHeight {
		t.Errorf("heights differ for one-line labels: %d vs %d", long.Win.ReqHeight, short.Win.ReqHeight)
	}

	two := button.New(app, "two", button.Text("one\ntwo"))
	if two.Win.ReqHeight <= short.Win.ReqHeight {
		t.Errorf("two lines are not taller: %d vs %d", two.Win.ReqHeight, short.Win.ReqHeight)
	}

	w10 := button.New(app, "w10", button.Text("x"), button.Width(10))
	w20 := button.New(app, "w20", button.Text("x"), button.Width(20))
	if w20.Win.ReqWidth <= w10.Win.ReqWidth {
		t.Errorf("Width(20) is not wider than Width(10): %d vs %d", w20.Win.ReqWidth, w10.Win.ReqWidth)
	}

	// Padding adds exactly twice its size on each axis.
	base := button.New(app, "base", button.Text("pad"), button.PadX(0), button.PadY(0))
	padded := button.New(app, "padded", button.Text("pad"), button.PadX(7), button.PadY(screenunit.Px(3)))
	if dw, dh := padded.Win.ReqWidth-base.Win.ReqWidth, padded.Win.ReqHeight-base.Win.ReqHeight; dw != 14 || dh != 6 {
		t.Errorf("padding 7x3 grew the request by %dx%d, want 14x6", dw, dh)
	}
	// So does the border.
	thick := button.New(app, "thick", button.Text("pad"), button.PadX(0), button.PadY(0), button.BorderWidth(base.BorderWidth+4))
	if dw := thick.Win.ReqWidth - base.Win.ReqWidth; dw != 8 {
		t.Errorf("4 more border pixels grew the width by %d, want 8", dw)
	}
}

func TestButtonInvokeAndConfigure(t *testing.T) {
	app := testutil.NewTestApp(t)
	calls := 0
	b := button.New(app, "b", button.Text("go"), button.Command(func() { calls++ }))
	b.Invoke()
	b.Invoke()
	if calls != 2 {
		t.Errorf("command ran %d times, want 2", calls)
	}

	before := b.Win.ReqWidth
	b.SetText("a considerably longer caption")
	if b.Text != "a considerably longer caption" || b.Win.ReqWidth <= before {
		t.Errorf("SetText: text %q, width %d (was %d)", b.Text, b.Win.ReqWidth, before)
	}
	if err := b.Configure(button.ReliefOpt(option.ReliefSunken), button.Anchor(option.AnchorW)); err != nil {
		t.Errorf("Configure = %v", err)
	}
	if b.Relief != option.ReliefSunken || b.Anchor != option.AnchorW {
		t.Errorf("relief %v anchor %v", b.Relief, b.Anchor)
	}
	if err := b.Configure(button.Foreground("definitely-not-a-colour")); !errors.Is(err, color.ErrUnknown) {
		t.Errorf("Configure(bad colour) = %v, want color.ErrUnknown", err)
	}

	b.Destroy()
	b.Destroy() // idempotent
	if !b.Destroyed() {
		t.Error("Destroyed not set")
	}
}

func TestDisabledButtonDoesNotInvoke(t *testing.T) {
	app := testutil.NewTestApp(t)
	calls := 0
	b := button.New(app, "b", button.Text("go"), button.Command(func() { calls++ }))
	b.State = widget.StateDisabled
	b.Invoke()
	if calls != 0 {
		t.Errorf("a disabled button ran its command %d times", calls)
	}
}

// TestDisabledImageIsStippled checks TkpDisplayButton's stippleGC: a
// disabled button's image gets a gray50 stipple in the background colour.
func TestDisabledImageIsStippled(t *testing.T) {
	app := testutil.NewTestApp(t)
	red := goimage.NewRGBA(goimage.Rect(0, 0, 16, 16))
	draw.Draw(red, red.Bounds(), &goimage.Uniform{gocolor.RGBA{255, 0, 0, 255}}, goimage.Point{}, draw.Src)
	img := image.NewPhoto("red", red)

	on := button.New(app, "on", button.ImageOpt(img), button.Background("white"))
	off := button.New(app, "off", button.ImageOpt(img), button.Background("white"),
		button.State(widget.StateDisabled))
	pack.Pack(on)
	pack.Pack(off)
	shot := testutil.Grab(t, app, app.Root())
	count := func(b *button.Button) int {
		n := 0
		for y := b.Win.Y; y < b.Win.Y+b.Win.Height; y++ {
			for x := b.Win.X; x < b.Win.X+b.Win.Width; x++ {
				if c := shot.NRGBAAt(x, y); c.R > 200 && c.G < 60 && c.B < 60 {
					n++
				}
			}
		}
		return n
	}
	normal, disabled := count(on), count(off)
	if normal != 256 {
		t.Fatalf("normal button shows %d red pixels, want the whole 16x16 image", normal)
	}
	if disabled < 100 || disabled > 156 {
		t.Errorf("disabled button shows %d of 256 image pixels, want about half (gray50)", disabled)
	}
}
