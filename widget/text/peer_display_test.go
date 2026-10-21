package text_test

import (
	"image/color"
	"testing"
	"time"

	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/widget/text"
)

func peerLuma(c color.NRGBA) int {
	return (299*int(c.R) + 587*int(c.G) + 114*int(c.B)) / 1000
}

// TestPrimaryRepaintsWhenPeerEdits checks that the primary text widget of a
// shared document repaints when a peer edits it: the peers share the
// document, and Tk redraws every peer of a shared text when the text
// changes.
func TestPrimaryRepaintsWhenPeerEdits(t *testing.T) {
	app := testutil.NewTestApp(t)
	primary := text.New(app, "primary", text.Width(40), text.Height(6),
		text.Background("white"), text.Foreground("black"))
	peer := text.NewPeer(primary.Doc(), app, "peer", text.Width(40), text.Height(6))
	pack.Pack(primary, pack.FillOpt(pack.FillBoth), pack.Expand(true))
	pack.Pack(peer, pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// The peer edits the shared document after the first paint.
	app.After(50*time.Millisecond, func() {
		if err := peer.Insert("1.0", "MMMM"); err != nil {
			t.Errorf("peer.Insert: %v", err)
		}
	})

	img := testutil.Grab(t, app, primary.Win)
	dark := 0
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if peerLuma(img.NRGBAAt(x, y)) < 96 {
				dark++
			}
		}
	}
	if dark < 20 {
		t.Errorf("only %d dark pixels in a %dx%d primary: the peer's edit was not painted",
			dark, b.Dx(), b.Dy())
	}
}
