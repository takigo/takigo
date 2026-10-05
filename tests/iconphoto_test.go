package takigo_test

import (
	"encoding/binary"
	"image"
	"image/color"
	"runtime"
	"testing"

	"github.com/takigo/takigo/internal/testutil"
)

func TestSetIconPhoto(t *testing.T) {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		t.Skip("reads the X11 _NET_WM_ICON property back")
	}
	app := testutil.NewTestApp(t)
	srv := app.Server()
	win := app.Root().PlatformID
	prop := srv.InternAtom("_NET_WM_ICON", false)

	icon := image.NewNRGBA(image.Rect(0, 0, 2, 3))
	icon.SetNRGBA(1, 0, color.NRGBA{R: 0x11, G: 0x22, B: 0x33, A: 0x80})
	small := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	small.SetNRGBA(0, 0, color.NRGBA{R: 0xff, A: 0xff})

	app.WmInfo().SetIconPhoto(icon, small)
	srv.Flush()

	data, _, format := srv.GetWindowProperty(win, prop, 0, 1024, false)
	if format != 32 {
		t.Fatalf("_NET_WM_ICON format = %d, want 32", format)
	}
	var got []uint32
	for i := 0; i+4 <= len(data); i += 4 {
		got = append(got, binary.NativeEndian.Uint32(data[i:]))
	}
	// width, height, 6 ARGB pixels; then the 1x1 icon.
	want := []uint32{2, 3, 0, 0x80112233, 0, 0, 0, 0, 1, 1, 0xffff0000}
	if len(got) != len(want) {
		t.Fatalf("property has %d cardinals, want %d: %x", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("cardinal %d = %#x, want %#x", i, got[i], want[i])
		}
	}

	app.WmInfo().SetIconPhoto()
	srv.Flush()
	if data, _, _ := srv.GetWindowProperty(win, prop, 0, 1024, false); len(data) != 0 {
		t.Errorf("the icon property still has %d bytes after removing the icon", len(data))
	}
}
