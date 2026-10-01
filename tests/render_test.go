package takigo_test

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/msorc/takigo/canvas"
	tkcolor "github.com/msorc/takigo/color"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/testutil"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

// These check what widgets actually paint, without golden images: only
// properties that hold for any font and any screen.

func luma(c color.NRGBA) int { return (299*int(c.R) + 587*int(c.G) + 114*int(c.B)) / 1000 }

func TestFramePaintsItsBackground(t *testing.T) {
	app := testutil.NewTestApp(t)
	f := frame.New(app, "f", frame.Width(60), frame.Height(40), frame.Background("red"))
	pack.Pack(f)
	img := testutil.Grab(t, app, f.Win)
	for _, p := range [][2]int{{1, 1}, {30, 20}, {58, 38}} {
		if got := img.NRGBAAt(p[0], p[1]); got != (color.NRGBA{255, 0, 0, 255}) {
			t.Errorf("pixel %v = %v, want red", p, got)
		}
	}
}

func TestLabelPaintsTextOnBackground(t *testing.T) {
	app := testutil.NewTestApp(t)
	l := label.New(app, "l", label.Text("MMMM"), label.Background("white"), label.Foreground("black"),
		label.PadX(10), label.PadY(10))
	pack.Pack(l)
	img := testutil.Grab(t, app, l.Win)

	white := color.NRGBA{255, 255, 255, 255}
	if got := img.NRGBAAt(2, 2); got != white {
		t.Errorf("corner pixel = %v, want the white background", got)
	}
	dark := 0
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if luma(img.NRGBAAt(x, y)) < 96 {
				dark++
			}
		}
	}
	if dark < 20 {
		t.Errorf("only %d dark pixels in a %dx%d label: the text was not drawn", dark, b.Dx(), b.Dy())
	}
}

func TestButtonReliefShading(t *testing.T) {
	app := testutil.NewTestApp(t)
	b := button.New(app, "b", button.Text("relief"), button.Background("#808080"),
		button.BorderWidth(3), button.HighlightThickness(0), button.ReliefOpt(option.ReliefRaised))
	pack.Pack(b)
	img := testutil.Grab(t, app, b.Win)
	w, h := img.Bounds().Dx(), img.Bounds().Dy()

	topLeft, bottomRight := luma(img.NRGBAAt(1, h/2)), luma(img.NRGBAAt(w-2, h/2))
	if topLeft <= 128 || bottomRight >= 128 {
		t.Errorf("raised border: left edge luma %d, right edge %d; want lighter and darker than the 128 face", topLeft, bottomRight)
	}
}

func TestSunkenReliefInvertsShading(t *testing.T) {
	app := testutil.NewTestApp(t)
	b := button.New(app, "b", button.Text("relief"), button.Background("#808080"),
		button.BorderWidth(3), button.HighlightThickness(0), button.ReliefOpt(option.ReliefSunken))
	pack.Pack(b)
	img := testutil.Grab(t, app, b.Win)
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	if left, right := luma(img.NRGBAAt(1, h/2)), luma(img.NRGBAAt(w-2, h/2)); left >= right {
		t.Errorf("sunken border: left edge luma %d is not darker than the right edge %d", left, right)
	}
}

func TestCanvasPaintsItems(t *testing.T) {
	app := testutil.NewTestApp(t)
	c := canvas.New(app, "c", canvas.Width(100), canvas.Height(80), canvas.Background("white"),
		canvas.BorderWidthOpt(0), canvas.HighlightWidthOpt(0))
	pack.Pack(c)
	id := c.CreateRectangle(20, 20, 60, 50, canvas.FillColor("blue"), canvas.OutlineNone())
	img := testutil.Grab(t, app, c.Win)

	if got := img.NRGBAAt(40, 35); got != (color.NRGBA{0, 0, 255, 255}) {
		t.Errorf("inside the rectangle = %v, want blue", got)
	}
	if got := img.NRGBAAt(80, 65); got != (color.NRGBA{255, 255, 255, 255}) {
		t.Errorf("outside the rectangle = %v, want white", got)
	}
	_ = id
}

// partial counts pixels that are neither the background nor the ink: the
// soft edge of an anti-aliased shape.
func partial(img interface {
	Bounds() image.Rectangle
	NRGBAAt(x, y int) color.NRGBA
}) int {
	n := 0
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if l := luma(img.NRGBAAt(x, y)); l > 16 && l < 239 {
				n++
			}
		}
	}
	return n
}

func diagonalCanvas(t *testing.T, opts ...canvas.CanvasOption) (*canvas.Canvas, func() int) {
	t.Helper()
	app := testutil.NewTestApp(t)
	base := []canvas.CanvasOption{canvas.Width(100), canvas.Height(80), canvas.Background("white"),
		canvas.BorderWidthOpt(0), canvas.HighlightWidthOpt(0)}
	c := canvas.New(app, "c", append(base, opts...)...)
	pack.Pack(c)
	c.CreateLine([]float64{5, 7, 93, 61}, canvas.OutlineColor("black"), canvas.OutlineWidth(2))
	c.CreateOval(20, 20, 70, 60, canvas.OutlineColor("black"))
	return c, func() int { return partial(testutil.Grab(t, app, c.Win)) }
}

func TestCanvasIsAntialiasedByDefault(t *testing.T) {
	_, soft := diagonalCanvas(t)
	if n := soft(); n < 40 {
		t.Errorf("%d soft-edge pixels on a diagonal and an oval: they are not anti-aliased", n)
	}
}

func TestCanvasAntialiasOff(t *testing.T) {
	_, soft := diagonalCanvas(t, canvas.Antialias(false))
	if n := soft(); n != 0 {
		t.Errorf("%d soft-edge pixels with Antialias(false), want the display server's hard edges", n)
	}
}

func TestClassicAppDrawsAsTk(t *testing.T) {
	t.Setenv("TAKIGO_CLASSIC", "1")
	_, soft := diagonalCanvas(t)
	if n := soft(); n != 0 {
		t.Errorf("%d soft-edge pixels in a Classic App, want none", n)
	}
}

// Axis-aligned shapes must stay sharp and where Tk puts them.
func TestAntialiasedRectangleIsCrisp(t *testing.T) {
	app := testutil.NewTestApp(t)
	c := canvas.New(app, "c", canvas.Width(100), canvas.Height(80), canvas.Background("white"),
		canvas.BorderWidthOpt(0), canvas.HighlightWidthOpt(0))
	pack.Pack(c)
	c.CreateRectangle(20, 20, 60, 50, canvas.FillColor("blue"), canvas.OutlineColor("black"))
	c.CreateRectangle(70, 20, 90, 50, canvas.OutlineColor("red"), canvas.OutlineWidth(2))
	img := testutil.Grab(t, app, c.Win)

	black, blue, white := color.NRGBA{0, 0, 0, 255}, color.NRGBA{0, 0, 255, 255}, color.NRGBA{255, 255, 255, 255}
	checks := []struct {
		x, y int
		want color.NRGBA
	}{
		{19, 35, white}, {20, 35, black}, {21, 35, blue}, {59, 35, blue}, {60, 35, black}, {61, 35, white},
		{40, 19, white}, {40, 20, black}, {40, 21, blue}, {40, 50, black}, {40, 51, white},
	}
	for _, ck := range checks {
		if got := img.NRGBAAt(ck.x, ck.y); got != ck.want {
			t.Errorf("pixel (%d,%d) = %v, want %v", ck.x, ck.y, got, ck.want)
		}
	}
	if n := partial(img); n != 0 {
		// The red outline is pure red (luma 76), counted as "partial" by
		// luma alone, so count real blends instead.
		blends := 0
		b := img.Bounds()
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				p := img.NRGBAAt(x, y)
				if p != black && p != blue && p != white && p != (color.NRGBA{255, 0, 0, 255}) {
					blends++
				}
			}
		}
		if blends != 0 {
			t.Errorf("%d blended pixels around axis-aligned rectangles, want sharp edges", blends)
		}
	}
}

func TestCanvasTranslucentFill(t *testing.T) {
	app := testutil.NewTestApp(t)
	c := canvas.New(app, "c", canvas.Width(100), canvas.Height(80), canvas.Background("white"),
		canvas.BorderWidthOpt(0), canvas.HighlightWidthOpt(0))
	pack.Pack(c)
	c.CreateRectangle(10, 10, 60, 60, canvas.FillColor("blue"), canvas.OutlineNone())
	// Half-transparent red over white and over the blue square.
	c.CreateRectangle(40, 20, 90, 50, canvas.FillColor(tkcolor.RGBA(255, 0, 0, 128)), canvas.OutlineNone())
	img := testutil.Grab(t, app, c.Win)

	near := func(got color.NRGBA, r, g, b int) bool {
		d := func(a uint8, b int) bool { return int(a)-b <= 2 && b-int(a) <= 2 }
		return d(got.R, r) && d(got.G, g) && d(got.B, b)
	}
	if got := img.NRGBAAt(75, 35); !near(got, 255, 127, 127) {
		t.Errorf("translucent red over white = %v, want about (255,127,127)", got)
	}
	if got := img.NRGBAAt(50, 35); !near(got, 128, 0, 127) {
		t.Errorf("translucent red over blue = %v, want about (128,0,127)", got)
	}
	if got := img.NRGBAAt(20, 35); got != (color.NRGBA{0, 0, 255, 255}) {
		t.Errorf("blue outside the overlap = %v", got)
	}
}

func TestCanvasDashedLineHasGaps(t *testing.T) {
	app := testutil.NewTestApp(t)
	c := canvas.New(app, "c", canvas.Width(100), canvas.Height(40), canvas.Background("white"),
		canvas.BorderWidthOpt(0), canvas.HighlightWidthOpt(0))
	pack.Pack(c)
	c.CreateLine([]float64{10, 20, 90, 20}, canvas.OutlineColor("black"), canvas.Dash(6, 4))
	img := testutil.Grab(t, app, c.Win)

	black, white := color.NRGBA{0, 0, 0, 255}, color.NRGBA{255, 255, 255, 255}
	for x := 10; x < 90; x++ {
		want := white
		if (x-10)%10 < 6 {
			want = black
		}
		if got := img.NRGBAAt(x, 20); got != want {
			t.Errorf("pixel (%d,20) = %v, want %v", x, got, want)
		}
	}
}

func TestCanvasDashedArcIsAntialiased(t *testing.T) {
	app := testutil.NewTestApp(t)
	c := canvas.New(app, "c", canvas.Width(120), canvas.Height(100), canvas.Background("white"),
		canvas.BorderWidthOpt(0), canvas.HighlightWidthOpt(0))
	pack.Pack(c)
	c.CreateArc(10, 10, 110, 90, canvas.StartAngle(20), canvas.Extent(300),
		canvas.FillColor("gold"), canvas.OutlineColor("black"), canvas.OutlineWidth(2), canvas.Dash(8, 6))
	img := testutil.Grab(t, app, c.Win)

	// Walk the ellipse: a dashed outline alternates between ink and the
	// fill, and its slanted dashes have soft edges.
	ink, gaps := 0, 0
	for deg := 30; deg < 310; deg++ {
		a := float64(deg) * math.Pi / 180
		x, y := 60+50*math.Cos(a), 50-40*math.Sin(a)
		if luma(img.NRGBAAt(int(x), int(y))) < 100 {
			ink++
		} else {
			gaps++
		}
	}
	if ink < 60 || gaps < 40 {
		t.Errorf("%d inked and %d open samples along the arc: it is not dashed", ink, gaps)
	}
	if got := img.NRGBAAt(35, 50); got != (color.NRGBA{255, 215, 0, 255}) {
		t.Errorf("inside the wedge = %v, want gold", got)
	}
	if n := partial(img); n < 30 {
		t.Errorf("%d soft-edge pixels on a dashed arc: it was drawn by the display server", n)
	}
}
