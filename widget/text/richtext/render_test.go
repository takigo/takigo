package richtext_test

import (
	"bytes"
	"encoding/base64"
	goimage "image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/widget/text"
	"github.com/takigo/takigo/widget/text/richtext"
)

// pngDataURI is a w×h red PNG as a data: URI.
func pngDataURI(t *testing.T, w, h int) string {
	t.Helper()
	img := goimage.NewRGBA(goimage.Rect(0, 0, w, h))
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i], img.Pix[i+3] = 255, 255
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

func TestRenderIntoWidget(t *testing.T) {
	app := testutil.NewTestApp(t)
	tw := text.New(app, "t", text.Width(60), text.Height(30))
	pack.Pack(tw, pack.FillOpt(pack.FillBoth), pack.Expand(true))

	html := `<h1>Title</h1><p>Some <b>bold</b> text and <a href="https://x.org/">a link</a>.</p>
		<pre>code line</pre><p><img src="` + pngDataURI(t, 40, 20) + `" width="20" alt="red"></p>`
	if err := richtext.Render(tw, html); err != nil {
		t.Fatal(err)
	}
	if got, want := tw.Get("1.0", "end"), "Title\nSome bold text and a link.\ncode line\n"; got != want {
		t.Errorf("text = %q, want %q", got, want)
	}
	if names := tw.ImageNames(); len(names) != 1 {
		t.Fatalf("images = %v, want one", names)
	}
	img, align, _, _, ok := tw.ImageAt("4.0")
	// width="20" is in CSS pixels, which scale with the screen's DPI.
	if !ok || img.Width() >= 40 || abs(img.Width()-2*img.Height()) > 1 || align != text.AlignBaseline {
		t.Errorf("image at 4.0: ok %v, %dx%d, align %v; want about 20x10 on the baseline", ok, img.Width(), img.Height(), align)
	}

	// The h1 text has a font twice the size of the body text.
	var h1Size, bodySize float64
	for _, tg := range tw.Doc().TagsAt(text.Index{Line: 1}) {
		if tg.Font != nil {
			h1Size = tg.Font.Attrs().Size
		}
	}
	for _, tg := range tw.Doc().TagsAt(text.Index{Line: 2}) {
		if tg.Font != nil {
			bodySize = tg.Font.Attrs().Size
		}
	}
	if h1Size == 0 || bodySize == 0 || h1Size/bodySize < 1.8 {
		t.Errorf("h1 font size %v, body %v: want the heading twice as big", h1Size, bodySize)
	}

	var box *text.Tag
	for _, tg := range tw.Doc().TagsAt(text.Index{Line: 3}) {
		if tg.BlockBackground != nil {
			box = tg
		}
	}
	if box == nil || box.BlockRadius == 0 {
		t.Fatalf("the pre line has no rounded block box")
	}

	// The link's characters carry a tag with the hand cursor.
	r, found := tw.Search("a link", "1.0")
	if !found {
		t.Fatal("link text not found")
	}
	var linkTag string
	for _, n := range tw.TagNamesAt(r.Start) {
		if tg := tw.Doc().Tags[n]; tg.CursorSet {
			linkTag = n
		}
	}
	if linkTag == "" {
		t.Fatal("the link has no tag with a cursor")
	}

	// The box background reaches the screen.
	grab := testutil.Grab(t, app, tw.Window())
	want := color.NRGBA{0xf6, 0xf8, 0xfa, 0xff}
	seen := false
	for y := 0; y < grab.Rect.Dy() && !seen; y++ {
		for x := 0; x < grab.Rect.Dx(); x++ {
			if grab.NRGBAAt(x, y) == want {
				seen = true
				break
			}
		}
	}
	if !seen {
		t.Error("the pre box's background is not drawn")
	}

	// A second Render leaves no tags of the first behind.
	before := len(tw.TagNames())
	if err := richtext.Render(tw, html); err != nil {
		t.Fatal(err)
	}
	if after := len(tw.TagNames()); after != before {
		t.Errorf("%d tags after the second render, %d after the first", after, before)
	}
	richtext.Clear(tw)
	for _, n := range tw.TagNames() {
		if strings.HasPrefix(n, "richtext:") {
			t.Errorf("Clear left tag %s", n)
		}
	}
}

func abs(n int) int { return max(n, -n) }

func TestRenderReflowsOnResize(t *testing.T) {
	app := testutil.NewTestApp(t)
	tw := text.New(app, "t", text.Width(30), text.Height(10))
	// No geometry manager: the test sets the width and announces it as the
	// text's ConfigureNotify handler does; the reflow runs at idle.
	resize := func(w int) {
		tw.Window().Width = w
		tw.Window().NotifyConfigure()
		app.UpdateIdleTasks()
	}
	resize(250)
	html := `<table><tr><td>a</td><td>` + strings.Repeat("word ", 40) + `</td></tr></table>`
	if err := richtext.Render(tw, html); err != nil {
		t.Fatal(err)
	}
	narrow := tw.Doc().LineCount()
	if narrow < 3 {
		t.Fatalf("the cell did not wrap: %d lines", narrow)
	}
	resize(1000)
	if wide := tw.Doc().LineCount(); wide >= narrow {
		t.Errorf("after widening %d lines, before %d: no reflow", wide, narrow)
	}
	if err := richtext.Render(tw, html, richtext.Reflow(false)); err != nil {
		t.Fatal(err)
	}
	before := tw.Doc().LineCount()
	resize(250)
	if after := tw.Doc().LineCount(); after != before {
		t.Errorf("Reflow(false) still reflowed: %d lines, then %d", before, after)
	}
}
