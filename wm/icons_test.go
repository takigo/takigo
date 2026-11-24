package wm

import (
	"image"
	"image/color"
	"testing"

	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/window"
)

// iconServer adds the calls of the icon and map paths to fakeServer.
type iconServer struct {
	*fakeServer
	icons     map[platform.WindowID][]platform.IconImage
	pixmaps   int
	freed     []platform.PixmapID
	withdrawn []platform.WindowID
	maps      int
}

func (s *iconServer) SetWindowIcons(w platform.WindowID, icons []platform.IconImage) {
	s.icons[w] = icons
}
func (s *iconServer) CreateBitmapFromData(platform.DrawableID, []byte, uint, uint) platform.PixmapID {
	s.pixmaps++
	return platform.PixmapID(100 + s.pixmaps)
}
func (s *iconServer) FreePixmap(p platform.PixmapID) { s.freed = append(s.freed, p) }
func (s *iconServer) MapWindow(platform.WindowID)    { s.maps++ }
func (s *iconServer) MapRaised(platform.WindowID)    { s.maps++ }
func (s *iconServer) WithdrawWindow(w platform.WindowID, _ int) {
	s.withdrawn = append(s.withdrawn, w)
}

func newIconDisplay() (*window.Display, *iconServer) {
	s := &iconServer{fakeServer: newFake(), icons: map[platform.WindowID][]platform.IconImage{}}
	return &window.Display{Server: s, Windows: map[platform.WindowID]*window.Window{}}, s
}

func newTop(d *window.Display, id platform.WindowID) *WmInfo {
	return Init(&window.Window{Display: d, PlatformID: id, Name: "top", Width: 200, Height: 100})
}

func TestIconPhotoDefault(t *testing.T) {
	d, s := newIconDisplay()
	root := newTop(d, 1)
	before := newTop(d, 2)

	img := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	img.SetNRGBA(0, 0, color.NRGBA{10, 20, 30, 128})
	root.SetIconPhotoDefault(img)
	if got := s.icons[1]; len(got) != 1 || got[0].Width != 2 || got[0].Pix[3] != 128 || got[0].Pix[0] != 10 {
		t.Fatalf("root icon = %+v, want the straight-alpha picture", got)
	}
	if _, ok := s.icons[2]; ok {
		t.Error("the default changed a toplevel created before it")
	}
	newTop(d, 3)
	if len(s.icons[3]) != 1 {
		t.Error("a toplevel created after -default has no icon")
	}

	root.SetIconPhotoDefault()
	newTop(d, 4)
	if _, ok := s.icons[4]; ok {
		t.Error("an empty -default still gave later toplevels an icon")
	}
	before.SetIconPhoto(img)
	if len(s.icons[2]) != 1 {
		t.Error("SetIconPhoto did not set the icon")
	}
}

func TestStateBeforeFirstMap(t *testing.T) {
	d, s := newIconDisplay()

	iconic := newTop(d, 1)
	iconic.Iconify()
	if s.iconified || s.maps != 0 {
		t.Error("Iconify before the first map talked to the window manager")
	}
	if s.hints.InitialState != platform.IconicState || iconic.State() != StateIconic {
		t.Errorf("initial state = %d, State = %v; want iconic", s.hints.InitialState, iconic.State())
	}
	iconic.Map(false)
	if s.maps != 1 || iconic.State() != StateIconic {
		t.Errorf("Map: %d maps, State %v; want one map, still iconic", s.maps, iconic.State())
	}

	withdrawn := newTop(d, 2)
	withdrawn.Withdraw()
	withdrawn.Map(false)
	if s.maps != 1 || len(s.withdrawn) != 0 || withdrawn.State() != StateWithdrawn {
		t.Errorf("a window withdrawn before its first map was mapped (%d maps, %d withdraws)", s.maps, len(s.withdrawn))
	}
	withdrawn.Deiconify()
	if s.maps != 2 || withdrawn.State() != StateNormal || s.hints.InitialState != platform.NormalState {
		t.Error("Deiconify did not show the withdrawn window")
	}

	fresh := newTop(d, 3)
	if fresh.State() != StateNormal {
		t.Errorf("a new window is %v, want normal", fresh.State())
	}
}

func TestStateFollowsWindowManager(t *testing.T) {
	d, _ := newIconDisplay()
	info := newTop(d, 1)
	info.Map(false)
	info.MapNotify(true)

	info.MapNotify(false) // the user minimized it
	if info.State() != StateIconic || info.Win.IsMapped() {
		t.Errorf("after an UnmapNotify State = %v, mapped %v; want iconic, unmapped", info.State(), info.Win.IsMapped())
	}
	info.MapNotify(true) // and restored it
	if info.State() != StateNormal || !info.Win.IsMapped() {
		t.Errorf("after a MapNotify State = %v; want normal", info.State())
	}

	info.Withdraw()
	info.MapNotify(true) // queued before the withdraw
	if info.State() != StateWithdrawn {
		t.Errorf("a stale MapNotify made a withdrawn window %v", info.State())
	}
}

func TestIconHints(t *testing.T) {
	d, s := newIconDisplay()
	info := newTop(d, 1)

	if err := info.SetIconBitmap("nosuchbitmap"); err == nil {
		t.Error("an unknown bitmap was accepted")
	}
	if err := info.SetIconBitmap("questhead"); err != nil {
		t.Fatal(err)
	}
	if err := info.SetIconMask("gray50"); err != nil {
		t.Fatal(err)
	}
	info.SetIconPosition(5, 6, true)
	h := s.hints
	if h.Flags&(platform.IconPixmapHint|platform.IconMaskHint|platform.IconPositionHint) !=
		platform.IconPixmapHint|platform.IconMaskHint|platform.IconPositionHint ||
		h.IconPixmap == 0 || h.IconMask == 0 || h.IconX != 5 || h.IconY != 6 {
		t.Fatalf("WM_HINTS = %+v", h)
	}
	if !h.Input || h.Flags&platform.InputHint == 0 {
		t.Error("the icon hints dropped the input hint")
	}
	if info.IconBitmap() != "questhead" || info.IconMask() != "gray50" {
		t.Error("the bitmap names are not kept")
	}
	if x, y, ok := info.IconPosition(); !ok || x != 5 || y != 6 {
		t.Error("IconPosition does not report the hint")
	}

	// A state change rewrites WM_HINTS without losing the icon.
	info.Iconify()
	if s.hints.Flags&platform.IconPixmapHint == 0 || s.hints.IconPixmap != h.IconPixmap {
		t.Error("Iconify wiped the icon bitmap hint")
	}

	pix := h.IconPixmap
	if err := info.SetIconBitmap(""); err != nil {
		t.Fatal(err)
	}
	if s.hints.Flags&platform.IconPixmapHint != 0 || len(s.freed) != 1 || s.freed[0] != pix {
		t.Errorf("removing the bitmap: flags %b, freed %v", s.hints.Flags, s.freed)
	}
	info.SetIconPosition(0, 0, false)
	if _, _, ok := info.IconPosition(); ok {
		t.Error("the icon position was not removed")
	}
}

func TestIconWindow(t *testing.T) {
	d, s := newIconDisplay()
	main := newTop(d, 1)
	icon := newTop(d, 2)
	icon.Map(false)

	if err := main.SetIconWindow(&window.Window{Display: d, PlatformID: 9}); err == nil {
		t.Error("a window that is not a toplevel was accepted")
	}
	if err := main.SetIconWindow(icon.Win); err != nil {
		t.Fatal(err)
	}
	if s.hints.Flags&platform.IconWindowHint == 0 || s.hints.IconWindow != 2 {
		t.Errorf("WM_HINTS = %+v, want icon window 2", s.hints)
	}
	if len(s.withdrawn) != 1 || icon.State() != StateIcon || main.IconWindow() != icon.Win {
		t.Errorf("icon window: withdrawn %v, state %v", s.withdrawn, icon.State())
	}
	other := newTop(d, 3)
	if err := other.SetIconWindow(icon.Win); err == nil {
		t.Error("one window became the icon of two")
	}
	icon.Map(false)
	if icon.State() != StateIcon {
		t.Error("Map showed an icon window")
	}

	if err := main.SetIconWindow(nil); err != nil {
		t.Fatal(err)
	}
	if s.hints.Flags&platform.IconWindowHint != 0 || icon.State() != StateWithdrawn {
		t.Errorf("removing the icon window: flags %b, its state %v", s.hints.Flags, icon.State())
	}
}
