package takigo_test

import (
	"encoding/binary"
	"image"
	"image/color"
	"runtime"
	"testing"
	"time"

	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/widget/toplevel"
	"github.com/takigo/takigo/wm"
)

func skipUnlessX11(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		t.Skip("reads X11 window properties back")
	}
}

// wmHints reads WM_HINTS back as its 32-bit fields.
func wmHints(t *testing.T, srv platform.DisplayServer, w platform.WindowID) []uint32 {
	t.Helper()
	data, _, format := srv.GetWindowProperty(w, srv.InternAtom("WM_HINTS", false), 0, 64, false)
	if format != 32 || len(data) < 9*4 {
		t.Fatalf("WM_HINTS: format %d, %d bytes", format, len(data))
	}
	f := make([]uint32, len(data)/4)
	for i := range f {
		f[i] = binary.NativeEndian.Uint32(data[i*4:])
	}
	return f
}

// TestWithdrawBeforeMainLoop checks that "wm withdraw ." before the loop
// starts keeps the window off the screen until it is deiconified.
func TestWithdrawBeforeMainLoop(t *testing.T) {
	skipUnlessX11(t)
	testutil.Settle()
	app := testutil.NewTestApp(t)
	srv := app.Server()
	xroot := app.Display().RootWindow
	app.WmInfo().Withdraw()

	var hidden, shown bool
	var stateHidden, stateShown wm.State
	app.After(200*time.Millisecond, func() {
		srv.Sync(false)
		hidden = srv.ChildAt(xroot, 5, 5) != app.Root().PlatformID
		stateHidden = app.WmInfo().State()
		app.WmInfo().Deiconify()
	})
	app.After(400*time.Millisecond, func() {
		defer app.Quit()
		srv.Sync(false)
		shown = srv.ChildAt(xroot, 5, 5) == app.Root().PlatformID
		stateShown = app.WmInfo().State()
	})
	app.MainLoop()
	if !hidden || stateHidden != wm.StateWithdrawn {
		t.Errorf("withdrawn before MainLoop: off screen %v, state %v", hidden, stateHidden)
	}
	if !shown || stateShown != wm.StateNormal {
		t.Errorf("after Deiconify: on screen %v, state %v", shown, stateShown)
	}
}

// TestIconifyBeforeMainLoop checks that "wm iconify ." before the first
// map asks the window manager to start the window iconic.
func TestIconifyBeforeMainLoop(t *testing.T) {
	skipUnlessX11(t)
	app := testutil.NewTestApp(t)
	app.WmInfo().Iconify()
	app.After(100*time.Millisecond, app.Quit)
	app.MainLoop()
	h := wmHints(t, app.Server(), app.Root().PlatformID)
	if h[0]&platform.StateHint == 0 || h[2] != platform.IconicState {
		t.Errorf("WM_HINTS flags %#x, initial_state %d; want IconicState", h[0], h[2])
	}
}

func TestIconHintsOnServer(t *testing.T) {
	skipUnlessX11(t)
	app := testutil.NewTestApp(t)
	info := app.WmInfo()
	if err := info.SetIconBitmap("questhead"); err != nil {
		t.Fatal(err)
	}
	info.SetIconPosition(30, 40, true)
	h := wmHints(t, app.Server(), app.Root().PlatformID)
	if h[0]&platform.IconPixmapHint == 0 || h[3] == 0 {
		t.Errorf("WM_HINTS has no icon pixmap: %#x", h)
	}
	if h[0]&platform.IconPositionHint == 0 || h[5] != 30 || h[6] != 40 {
		t.Errorf("WM_HINTS icon position: %#x", h)
	}
	if h[0]&platform.InputHint == 0 || h[1] != 1 {
		t.Errorf("the icon hints lost the input hint: %#x", h)
	}
}

// TestIconPhotoDefault checks that "wm iconphoto -default" reaches a
// toplevel created afterwards.
func TestIconPhotoDefault(t *testing.T) {
	skipUnlessX11(t)
	app := testutil.NewTestApp(t)
	icon := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	icon.SetNRGBA(0, 0, color.NRGBA{R: 0xff, A: 0xff})
	app.WmInfo().SetIconPhotoDefault(icon)

	top := toplevel.New(app, "later")
	srv := app.Server()
	data, _, _ := srv.GetWindowProperty(top.Window().PlatformID, srv.InternAtom("_NET_WM_ICON", false), 0, 16, false)
	if len(data) != 12 || binary.NativeEndian.Uint32(data[8:]) != 0xffff0000 {
		t.Errorf("_NET_WM_ICON of a later toplevel = %x, want the default 1x1 red icon", data)
	}
}
