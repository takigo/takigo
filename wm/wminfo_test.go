package wm

import (
	"slices"
	"testing"

	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/window"
)

type fakeServer struct {
	platform.DisplayServer
	atoms       map[string]platform.AtomID
	props       map[platform.AtomID][]byte
	protocols   []platform.AtomID
	name        string
	iconName    string
	class       [2]string
	hints       *platform.WMHints
	sizeHints   *platform.SizeHints
	moves       [][4]int
	transient   platform.WindowID
	deleted     []platform.AtomID
	iconified   bool
	withdrawn   bool
	mapped      bool
	destroyed   bool
	screenW     int
	screenH     int
	internCalls int
	pings       [][2]platform.WindowID
}

func newFake() *fakeServer {
	return &fakeServer{atoms: map[string]platform.AtomID{}, props: map[platform.AtomID][]byte{},
		screenW: 1920, screenH: 1080}
}

func (s *fakeServer) InternAtom(name string, _ bool) platform.AtomID {
	s.internCalls++
	if a, ok := s.atoms[name]; ok {
		return a
	}
	a := platform.AtomID(len(s.atoms) + 1)
	s.atoms[name] = a
	return a
}
func (s *fakeServer) SendClientMessage(w, target platform.WindowID, _ platform.AtomID, _, _, _, _, _ int64) {
	s.pings = append(s.pings, [2]platform.WindowID{w, target})
}
func (s *fakeServer) Flush()                                              {}
func (s *fakeServer) SetWMHints(_ platform.WindowID, h *platform.WMHints) { s.hints = h }
func (s *fakeServer) SetClassHint(_ platform.WindowID, name, class string) {
	s.class = [2]string{name, class}
}
func (s *fakeServer) StoreName(_ platform.WindowID, name string)   { s.name = name }
func (s *fakeServer) SetIconName(_ platform.WindowID, name string) { s.iconName = name }
func (s *fakeServer) ChangeProperty(_ platform.WindowID, prop, _ platform.AtomID, _, _ int, data []byte, _ int) {
	s.props[prop] = data
}
func (s *fakeServer) ChangePropertyAtoms(_ platform.WindowID, _ platform.AtomID, atoms []platform.AtomID) {
	s.protocols = atoms
}
func (s *fakeServer) DeleteProperty(_ platform.WindowID, prop platform.AtomID) {
	s.deleted = append(s.deleted, prop)
}
func (s *fakeServer) SetWMNormalHints(_ platform.WindowID, h *platform.SizeHints) { s.sizeHints = h }
func (s *fakeServer) MoveResizeWindow(_ platform.WindowID, x, y int, w, h uint) {
	s.moves = append(s.moves, [4]int{x, y, int(w), int(h)})
}
func (s *fakeServer) SetTransientForHint(_, parent platform.WindowID) { s.transient = parent }
func (s *fakeServer) IconifyWindow(platform.WindowID, int)            { s.iconified = true }
func (s *fakeServer) WithdrawWindow(platform.WindowID, int)           { s.withdrawn = true }
func (s *fakeServer) MapWindow(platform.WindowID)                     { s.mapped = true }
func (s *fakeServer) DestroyWindow(platform.WindowID)                 { s.destroyed = true }
func (s *fakeServer) ScreenWidth(int) int                             { return s.screenW }
func (s *fakeServer) ScreenHeight(int) int                            { return s.screenH }

func newToplevel(t *testing.T) (*WmInfo, *fakeServer) {
	t.Helper()
	s := newFake()
	d := &window.Display{Server: s, Windows: map[platform.WindowID]*window.Window{}}
	w := &window.Window{Display: d, PlatformID: 7, Name: "top", Width: 200, Height: 100}
	return Init(w), s
}

func TestStateString(t *testing.T) {
	for s, want := range map[State]string{
		StateNormal: "normal", StateIconic: "iconic", StateWithdrawn: "withdrawn",
		StateZoomed: "zoomed", State(99): "normal",
	} {
		if got := s.String(); got != want {
			t.Errorf("State(%d).String() = %q, want %q", s, got, want)
		}
	}
}

func TestInit(t *testing.T) {
	info, s := newToplevel(t)
	if info.Win.WmData != info {
		t.Error("WmData not set")
	}
	if info.Title != "top" || info.UserW != -1 || info.UserH != -1 || !info.ResizableW || !info.ResizableH {
		t.Errorf("unexpected defaults: %+v", info)
	}
	if s.class != [2]string{"top", "Takigo"} {
		t.Errorf("class hint = %v", s.class)
	}
	if s.hints == nil || !s.hints.Input || s.hints.InitialState != platform.NormalState {
		t.Errorf("WM hints = %+v", s.hints)
	}
	want := []platform.AtomID{s.atoms["WM_DELETE_WINDOW"], s.atoms["_NET_WM_PING"]}
	if !slices.Equal(s.protocols, want) {
		t.Errorf("WM_PROTOCOLS = %v, want %v", s.protocols, want)
	}
	if h := s.sizeHints; h == nil || h.MinWidth != 1 || h.MinHeight != 1 || h.Flags&platform.PMaxSize != 0 {
		t.Errorf("size hints = %+v", h)
	}

	n := s.internCalls
	Init(&window.Window{Display: info.Win.Display, PlatformID: 8})
	if s.internCalls != n {
		t.Error("atoms were interned again for the same display")
	}
}

func TestTitleAndIconName(t *testing.T) {
	info, s := newToplevel(t)
	info.SetTitle("héllo")
	if info.Title != "héllo" || s.name != "héllo" || string(s.props[s.atoms["_NET_WM_NAME"]]) != "héllo" {
		t.Errorf("title not set: name %q, _NET_WM_NAME %q", s.name, s.props[s.atoms["_NET_WM_NAME"]])
	}
	info.SetIconName("ico")
	if info.IconName != "ico" || s.iconName != "ico" || string(s.props[s.atoms["_NET_WM_ICON_NAME"]]) != "ico" {
		t.Error("icon name not set")
	}
}

func TestSetGeometry(t *testing.T) {
	tests := []struct {
		geom    string
		want    [4]int
		wantStr string
		posSet  bool
	}{
		{"300x150", [4]int{0, 0, 300, 150}, "300x150+0+0", false},
		{"+10+20", [4]int{10, 20, 200, 100}, "200x100+10+20", true},
		{"300x150-10-20", [4]int{1920 - 300 - 10, 1080 - 150 - 20, 300, 150}, "300x150-10-20", true},
	}
	for _, tt := range tests {
		t.Run(tt.geom, func(t *testing.T) {
			info, s := newToplevel(t)
			if err := info.SetGeometry(tt.geom); err != nil {
				t.Fatal(err)
			}
			w := info.Win
			if got := [4]int{w.X, w.Y, w.Width, w.Height}; got != tt.want {
				t.Errorf("window geometry = %v, want %v", got, tt.want)
			}
			if len(s.moves) != 1 || s.moves[0] != tt.want {
				t.Errorf("MoveResizeWindow calls = %v, want [%v]", s.moves, tt.want)
			}
			if info.PositionSet != tt.posSet {
				t.Errorf("PositionSet = %v, want %v", info.PositionSet, tt.posSet)
			}
			if got := s.sizeHints.Flags&platform.USPosition != 0; got != tt.posSet {
				t.Errorf("USPosition flag = %v, want %v", got, tt.posSet)
			}
			if got := info.Geometry(); got != tt.wantStr {
				t.Errorf("Geometry() = %q, want %q", got, tt.wantStr)
			}
		})
	}

	info, s := newToplevel(t)
	if err := info.SetGeometry("bogus"); err == nil {
		t.Error("SetGeometry accepted a bad geometry")
	}
	if len(s.moves) != 0 {
		t.Error("bad geometry was applied")
	}
}

func TestSizeHints(t *testing.T) {
	info, s := newToplevel(t)
	info.SetMinSize(0, -5)
	if info.MinWidth != 1 || info.MinHeight != 1 {
		t.Errorf("min size = %dx%d, want clamped to 1x1", info.MinWidth, info.MinHeight)
	}
	info.SetMinSize(50, 40)
	if h := s.sizeHints; h.MinWidth != 50 || h.MinHeight != 40 {
		t.Errorf("min hints = %dx%d", h.MinWidth, h.MinHeight)
	}

	info.SetMaxSize(800, 600)
	if h := s.sizeHints; h.Flags&platform.PMaxSize == 0 || h.MaxWidth != 800 || h.MaxHeight != 600 {
		t.Errorf("max hints = %+v", h)
	}
	info.SetMaxSize(0, 0)
	if s.sizeHints.Flags&platform.PMaxSize != 0 {
		t.Error("PMaxSize set with no maximum")
	}

	info.SetResizable(false, true)
	h := s.sizeHints
	if h.Flags&platform.PMaxSize == 0 || h.MinWidth != 200 || h.MaxWidth != 200 || h.MinHeight != 40 {
		t.Errorf("non-resizable width hints = %+v", h)
	}
	info.SetResizable(true, false)
	h = s.sizeHints
	if h.MinHeight != 100 || h.MaxHeight != 100 || h.MinWidth != 50 {
		t.Errorf("non-resizable height hints = %+v", h)
	}
}

func TestTransientAndStates(t *testing.T) {
	info, s := newToplevel(t)
	parent := &window.Window{PlatformID: 3}
	info.SetTransientFor(parent)
	if s.transient != 3 || info.TransientFor != parent {
		t.Error("transient hint not set")
	}
	info.Iconify()
	if s.iconified || info.GetState() != StateNormal {
		t.Error("a transient window was iconified")
	}
	info.SetTransientFor(nil)
	if !slices.Contains(s.deleted, s.atoms["WM_TRANSIENT_FOR"]) {
		t.Error("WM_TRANSIENT_FOR not deleted")
	}

	info.Iconify()
	if !s.iconified || info.GetState() != StateIconic {
		t.Error("Iconify did not iconify")
	}
	info.Withdraw()
	if !s.withdrawn || !info.Withdrawn || info.GetState() != StateWithdrawn {
		t.Error("Withdraw did not withdraw")
	}
	info.Deiconify()
	if !s.mapped || info.Withdrawn || info.GetState() != StateNormal || !info.Win.IsMapped() {
		t.Error("Deiconify did not restore the window")
	}
}

func TestUnrealizedWindowSkipsServer(t *testing.T) {
	info, s := newToplevel(t)
	info.Win.PlatformID = 0
	s.sizeHints, s.protocols = nil, nil
	info.SetMinSize(10, 10)
	info.OnDeleteWindow(func() {})
	if err := info.SetGeometry("10x10+1+1"); err != nil {
		t.Fatal(err)
	}
	info.Iconify()
	info.Withdraw()
	info.Deiconify()
	if s.sizeHints != nil || s.protocols != nil || len(s.moves) != 0 || s.iconified || s.withdrawn || s.mapped {
		t.Error("server called for a window with no platform window")
	}
	if info.GetState() != StateNormal {
		t.Errorf("state = %v, want normal", info.GetState())
	}
}

func TestProtocols(t *testing.T) {
	info, s := newToplevel(t)
	msg := func(proto string) bool {
		return info.HandleClientMessage(s.atoms["WM_PROTOCOLS"], [5]int64{int64(s.atoms[proto])})
	}

	if msg("WM_DELETE_WINDOW") {
		t.Error("WM_DELETE_WINDOW handled with no handler")
	}
	if !msg("_NET_WM_PING") {
		t.Error("_NET_WM_PING not handled")
	}
	if root := info.Win.Display.RootWindow; len(s.pings) != 1 || s.pings[0] != [2]platform.WindowID{root, root} {
		t.Errorf("_NET_WM_PING reply = %v, want one message to the root window %d", s.pings, root)
	}
	if info.HandleClientMessage(s.atoms["WM_DELETE_WINDOW"], [5]int64{}) {
		t.Error("handled a message that is not WM_PROTOCOLS")
	}

	deleted := 0
	info.OnDeleteWindow(func() { deleted++ })
	if !msg("WM_DELETE_WINDOW") || deleted != 1 {
		t.Error("WM_DELETE_WINDOW handler not called")
	}
	if len(s.protocols) != 2 {
		t.Errorf("WM_PROTOCOLS = %v, want WM_DELETE_WINDOW listed once", s.protocols)
	}

	taken := false
	info.OnProtocol("WM_TAKE_FOCUS", func() { taken = true })
	if !msg("WM_TAKE_FOCUS") || !taken {
		t.Error("WM_TAKE_FOCUS handler not called")
	}
	if !slices.Contains(s.protocols, s.atoms["WM_TAKE_FOCUS"]) {
		t.Error("WM_TAKE_FOCUS not advertised in WM_PROTOCOLS")
	}

	info.OffDeleteWindow()
	if !msg("WM_DELETE_WINDOW") || !info.Win.IsDestroyed() || !s.destroyed {
		t.Error("default WM_DELETE_WINDOW did not destroy the window")
	}
}

func (s *fakeServer) ResizeWindow(_ platform.WindowID, width, height uint) {
	s.moves = append(s.moves, [4]int{-1, -1, int(width), int(height)})
}

func TestResizeToplevelKeepsUserGeometry(t *testing.T) {
	info, _ := newToplevel(t)
	w := info.Win
	window.ResizeToplevel(w, 120, 80)
	if w.Width != 120 || w.Height != 80 {
		t.Fatalf("size without wm geometry = %dx%d, want the request 120x80", w.Width, w.Height)
	}
	if err := info.SetGeometry("300x200"); err != nil {
		t.Fatal(err)
	}
	window.ResizeToplevel(w, 150, 90)
	if w.Width != 300 || w.Height != 200 {
		t.Errorf("size after wm geometry 300x200 = %dx%d, want it kept", w.Width, w.Height)
	}
}

func TestConfigureNotifyRecordsUserResize(t *testing.T) {
	info, _ := newToplevel(t)
	w := info.Win
	w.ReqWidth, w.ReqHeight = 120, 80
	window.ResizeToplevel(w, 120, 80)
	window.ResizeToplevel(w, 130, 90)

	// Our own requests coming back, even late, are not user resizes.
	info.ConfigureNotify(120, 80)
	info.ConfigureNotify(130, 90)
	if info.UserW > 0 || info.UserH > 0 {
		t.Fatalf("own resize taken as the user's: %dx%d", info.UserW, info.UserH)
	}

	// The user drags the window to 400x300: the size sticks.
	info.ConfigureNotify(400, 300)
	w.Width, w.Height = 400, 300
	window.ResizeToplevel(w, 140, 95)
	if w.Width != 400 || w.Height != 300 {
		t.Errorf("size after a user resize and a new request = %dx%d, want 400x300", w.Width, w.Height)
	}
}

// A gridded toplevel (a text widget's -setgrid) resizes in whole cells:
// the hints carry the cell size as increments and the part of the request
// that is not cells as the base size, and UnsetGrid puts them back.
func TestSetGrid(t *testing.T) {
	info, s := newToplevel(t)
	w := info.Win
	w.ReqWidth, w.ReqHeight = 80*7+10, 24*13+6
	grid := &window.Window{Display: w.Display, Parent: w}

	info.SetGrid(grid, 80, 24, 7, 13)
	h := s.sizeHints
	if h.Flags&platform.PBaseSize == 0 || h.WidthInc != 7 || h.HeightInc != 13 {
		t.Fatalf("size hints after SetGrid = %+v", h)
	}
	if h.BaseWidth != 10 || h.BaseHeight != 6 || h.MinWidth != 10+7 || h.MinHeight != 6+13 {
		t.Errorf("base %dx%d min %dx%d, want base 10x6 min 17x19", h.BaseWidth, h.BaseHeight, h.MinWidth, h.MinHeight)
	}

	other := &window.Window{Display: w.Display, Parent: w}
	info.SetGrid(other, 1, 1, 2, 2)
	if s.sizeHints.WidthInc != 7 {
		t.Error("a second window took over the grid")
	}
	info.UnsetGrid(other)
	if s.sizeHints.WidthInc != 7 {
		t.Error("UnsetGrid by a window that does not grid cleared the grid")
	}

	info.UnsetGrid(grid)
	h = s.sizeHints
	if h.Flags&platform.PBaseSize != 0 || h.WidthInc != 1 || h.HeightInc != 1 || h.MinWidth != 1 {
		t.Errorf("size hints after UnsetGrid = %+v", h)
	}
}
