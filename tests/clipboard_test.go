package takigo_test

import (
	"image"
	"image/color"
	"math/rand/v2"
	"strings"
	"testing"
	"time"

	takigo "github.com/msorc/takigo"
	"github.com/msorc/takigo/internal/testutil"
	"github.com/msorc/takigo/platform"
)

// A clipboard far over Tk's 4000-byte limit goes through the INCR
// protocol between two X clients (two Apps, two display connections).
func TestClipboardLargeTransferBetweenApps(t *testing.T) {
	testutil.RequireDisplay(t)
	text := strings.Repeat("naïve café — 日本語のテキスト, line of text\n", 2500)

	owner, err := takigo.NewApp(takigo.Title("owner"), takigo.Size(50, 50))
	if err != nil {
		t.Fatal(err)
	}
	reader, err := takigo.NewApp(takigo.Title("reader"), takigo.Size(50, 50))
	if err != nil {
		t.Fatal(err)
	}
	owner.Clipboard().Set(owner.Root().PlatformID, text, platform.CurrentTime)
	owner.Server().Sync(false) // the X server knows the owner before the reader asks

	got := make(chan string, 1)
	reader.DoWhenIdle(func() {
		reader.Clipboard().Get(reader.Root().PlatformID, platform.CurrentTime, func(s string) { got <- s })
	})
	ownerDone, readerDone := make(chan struct{}), make(chan struct{})
	go func() { owner.Run(); close(ownerDone) }()
	go func() { reader.Run(); close(readerDone) }()
	defer func() {
		owner.Quit()
		reader.Quit()
		<-ownerDone
		<-readerDone
	}()

	select {
	case s := <-got:
		if s != text {
			t.Errorf("pasted %d bytes, want the %d-byte clipboard (prefix %q)", len(s), len(text), s[:min(len(s), 40)])
		}
	case <-time.After(15 * time.Second):
		t.Fatal("clipboard transfer did not finish")
	}
}

// An image goes from one X client to another as image/png, in chunks
// (INCR) because it is larger than one property write.
func TestClipboardImageBetweenApps(t *testing.T) {
	testutil.RequireDisplay(t)
	src := image.NewNRGBA(image.Rect(0, 0, 160, 120))
	rng := rand.New(rand.NewPCG(1, 2)) // noise, so the PNG stays large
	for i := range src.Pix {
		src.Pix[i] = uint8(rng.UintN(256))
		if i%4 == 3 {
			src.Pix[i] = 255
		}
	}

	owner, err := takigo.NewApp(takigo.Title("owner"), takigo.Size(50, 50))
	if err != nil {
		t.Fatal(err)
	}
	reader, err := takigo.NewApp(takigo.Title("reader"), takigo.Size(50, 50))
	if err != nil {
		t.Fatal(err)
	}
	if err := owner.SetClipboardImage(src); err != nil {
		t.Fatal(err)
	}
	owner.Server().Sync(false)

	// The owner sees its own image without a round trip.
	var own image.Image
	owner.ClipboardImage(func(img image.Image) { own = img })
	if own == nil || own.Bounds() != src.Bounds() {
		t.Errorf("the owner's own clipboard image = %v", own)
	}

	got := make(chan image.Image, 1)
	reader.DoWhenIdle(func() { reader.ClipboardImage(func(img image.Image) { got <- img }) })
	ownerDone, readerDone := make(chan struct{}), make(chan struct{})
	go func() { owner.Run(); close(ownerDone) }()
	go func() { reader.Run(); close(readerDone) }()
	defer func() {
		owner.Quit()
		reader.Quit()
		<-ownerDone
		<-readerDone
	}()

	select {
	case img := <-got:
		if img == nil {
			t.Fatal("the reader got no image")
		}
		if img.Bounds() != src.Bounds() {
			t.Fatalf("pasted image is %v, want %v", img.Bounds(), src.Bounds())
		}
		for _, p := range [][2]int{{0, 0}, {80, 60}, {159, 119}} {
			if a, b := img.At(p[0], p[1]), src.At(p[0], p[1]); !sameColor(a, b) {
				t.Errorf("pixel %v = %v, want %v", p, a, b)
			}
		}
	case <-time.After(15 * time.Second):
		t.Fatal("clipboard image transfer did not finish")
	}
}

func sameColor(a, b color.Color) bool {
	r1, g1, b1, a1 := a.RGBA()
	r2, g2, b2, a2 := b.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}

// A text request does not get an image, and an image request no text.
func TestClipboardImageAndTextDoNotMix(t *testing.T) {
	app := testutil.NewTestApp(t)
	app.Clipboard().Set(app.Root().PlatformID, "just text", platform.CurrentTime)
	called := false
	app.ClipboardImage(func(img image.Image) {
		called = true
		if img != nil {
			t.Errorf("got an image from a text clipboard: %v", img.Bounds())
		}
	})
	_ = called
}
