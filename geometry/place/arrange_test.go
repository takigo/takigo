package place

import (
	"testing"

	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/window"
)

type fakeServer struct {
	platform.DisplayServer
	moves  map[platform.WindowID][4]int
	mapped map[platform.WindowID]bool
}

func (s *fakeServer) MoveResizeWindow(w platform.WindowID, x, y int, width, height uint) {
	s.moves[w] = [4]int{x, y, int(width), int(height)}
}
func (s *fakeServer) MapWindow(w platform.WindowID)     { s.mapped[w] = true }
func (s *fakeServer) UnmapWindow(w platform.WindowID)   { s.mapped[w] = false }
func (s *fakeServer) DestroyWindow(w platform.WindowID) {}

func newDisplay() (*window.Display, *fakeServer) {
	s := &fakeServer{moves: map[platform.WindowID][4]int{}, mapped: map[platform.WindowID]bool{}}
	return &window.Display{Server: s, Windows: map[platform.WindowID]*window.Window{}}, s
}

func newContainer(d *window.Display, w, h int) *window.Window {
	c := &window.Window{Display: d, PlatformID: 1, Width: w, Height: h, PathName: ".c"}
	c.Flags |= window.FlagMapped
	return c
}

var nextID platform.WindowID = 100

func newChild(parent *window.Window, reqW, reqH int) *window.Window {
	nextID++
	c := &window.Window{Display: parent.Display, Parent: parent, PlatformID: nextID, ReqWidth: reqW, ReqHeight: reqH}
	parent.Children = append(parent.Children, c)
	return c
}

func geom(w *window.Window) [4]int { return [4]int{w.X, w.Y, w.Width, w.Height} }

func TestArrangeGeometry(t *testing.T) {
	tests := []struct {
		name string
		bw   int
		opts []PlaceOption
		want [4]int
	}{
		{"requested size", 0, nil, [4]int{0, 0, 30, 20}},
		{"absolute", 0, []PlaceOption{X(5), Y(7), Width(40), Height(10)}, [4]int{5, 7, 40, 10}},
		{"relative", 0, []PlaceOption{RelX(0.5), RelY(0.25), RelWidth(0.5), RelHeight(0.5)}, [4]int{100, 25, 100, 50}},
		{"relative plus absolute", 0, []PlaceOption{RelWidth(0.5), Width(10), RelHeight(0), Height(3)}, [4]int{0, 0, 110, 3}},
		{"center", 0, []PlaceOption{RelX(0.5), RelY(0.5), Anchor(option.AnchorCenter)}, [4]int{85, 40, 30, 20}},
		{"n", 0, []PlaceOption{X(50), Anchor(option.AnchorN)}, [4]int{35, 0, 30, 20}},
		{"ne", 0, []PlaceOption{X(50), Anchor(option.AnchorNE)}, [4]int{20, 0, 30, 20}},
		{"e", 0, []PlaceOption{X(50), Y(50), Anchor(option.AnchorE)}, [4]int{20, 40, 30, 20}},
		{"se", 0, []PlaceOption{RelX(1), RelY(1), Anchor(option.AnchorSE)}, [4]int{170, 80, 30, 20}},
		{"s", 0, []PlaceOption{X(50), Y(50), Anchor(option.AnchorS)}, [4]int{35, 30, 30, 20}},
		{"sw", 0, []PlaceOption{X(50), Y(50), Anchor(option.AnchorSW)}, [4]int{50, 30, 30, 20}},
		{"w", 0, []PlaceOption{X(50), Y(50), Anchor(option.AnchorW)}, [4]int{50, 40, 30, 20}},
		{"negative rounds away from zero", 0, []PlaceOption{RelX(-0.0025)}, [4]int{-1, 0, 30, 20}},
		{"border width", 2, nil, [4]int{0, 0, 30, 20}},
		{"border width absolute", 2, []PlaceOption{Width(40), Height(10)}, [4]int{0, 0, 36, 6}},
		{"clamped", 0, []PlaceOption{Width(0), Height(0)}, [4]int{0, 0, 1, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, s := newDisplay()
			c := newContainer(d, 200, 100)
			ch := newChild(c, 30, 20)
			ch.BorderWidth = tt.bw
			Place(ch, tt.opts...)
			t.Cleanup(func() { Forget(ch) })
			if got := geom(ch); got != tt.want {
				t.Errorf("geometry = %v, want %v", got, tt.want)
			}
			if got := s.moves[ch.PlatformID]; got != tt.want {
				t.Errorf("MoveResizeWindow = %v, want %v", got, tt.want)
			}
			if !s.mapped[ch.PlatformID] || !ch.IsMapped() {
				t.Error("content of a mapped container was not mapped")
			}
		})
	}
}

func TestArrangeInternalBorders(t *testing.T) {
	d, _ := newDisplay()
	c := newContainer(d, 200, 100)
	c.InternalBorderLeft, c.InternalBorderTop = 10, 5
	c.InternalBorderRight, c.InternalBorderBottom = 10, 5
	ch := newChild(c, 30, 20)
	Place(ch, RelX(0), RelY(0), RelWidth(1), RelHeight(1))
	defer Forget(ch)
	if got, want := geom(ch), [4]int{10, 5, 180, 90}; got != want {
		t.Errorf("geometry = %v, want %v", got, want)
	}
}

func TestPlaceReconfigureAndForget(t *testing.T) {
	d, s := newDisplay()
	c := newContainer(d, 200, 100)
	ch := newChild(c, 30, 20)
	Place(ch, X(10))
	Place(ch, X(20))
	p := placers[c]
	if p == nil || len(p.entries) != 1 {
		t.Fatalf("re-placing added an entry: %+v", p)
	}
	if ch.X != 20 {
		t.Errorf("X = %d after re-place, want 20", ch.X)
	}
	if ch.GeomManager != mgr {
		t.Error("content not managed by place")
	}

	Forget(ch)
	if _, ok := placers[c]; ok {
		t.Error("container state kept after its last content was forgotten")
	}
	if _, ok := containerOf[ch]; ok {
		t.Error("containerOf kept a forgotten window")
	}
	if ch.GeomManager != nil || ch.IsMapped() || s.mapped[ch.PlatformID] {
		t.Error("forgotten content still managed or mapped")
	}
	Forget(&window.Window{})
}

func TestUnmappedContainerDefersMapping(t *testing.T) {
	d, s := newDisplay()
	c := newContainer(d, 200, 100)
	c.Flags &^= window.FlagMapped
	ch := newChild(c, 30, 20)
	Place(ch)
	defer Forget(ch)
	if s.mapped[ch.PlatformID] {
		t.Fatal("content mapped inside an unmapped container")
	}
	window.MarkMapped(c)
	if !s.mapped[ch.PlatformID] {
		t.Error("content not mapped once its container was")
	}
}

func TestArrangeSkipsUnrealizedContainer(t *testing.T) {
	d, s := newDisplay()
	c := newContainer(d, 200, 100)
	c.PlatformID = 0
	ch := newChild(c, 30, 20)
	Place(ch, X(10))
	defer Forget(ch)
	if ch.X != 0 || len(s.moves) != 0 {
		t.Error("arranged content of a container with no platform window")
	}
}

func TestManagerProcs(t *testing.T) {
	d, s := newDisplay()
	c := newContainer(d, 200, 100)
	a, b := newChild(c, 30, 20), newChild(c, 30, 20)
	Place(a)
	Place(b, X(50))
	if mgr.Name() != "place" {
		t.Errorf("Name = %q", mgr.Name())
	}

	a.ReqWidth = 60
	mgr.RequestProc(a)
	if a.Width != 60 {
		t.Errorf("width = %d after RequestProc, want 60", a.Width)
	}

	mgr.LostContentProc(a)
	if len(placers[c].entries) != 1 || placers[c].entries[0].window != b {
		t.Error("LostContentProc did not drop the content")
	}

	b.ReqHeight = 44
	delete(s.moves, b.PlatformID)
	ArrangeAll()
	if s.moves[b.PlatformID][3] != 44 {
		t.Error("ArrangeAll did not re-arrange")
	}
	Forget(b)
}

func TestPlaceIn(t *testing.T) {
	d, s := newDisplay()
	top := newContainer(d, 300, 300)
	sib := newChild(top, 0, 0)
	sib.X, sib.Y, sib.Width, sib.Height = 40, 30, 100, 100
	sib.Flags |= window.FlagMapped
	ch := newChild(top, 20, 10)

	Place(ch, In(sib), RelX(0.5), RelY(0.5), Anchor(option.AnchorCenter))
	if got, want := geom(ch), [4]int{80, 75, 20, 10}; got != want {
		t.Errorf("geometry = %v, want %v", got, want)
	}
	if !placers[sib].hasForeign() {
		t.Error("hasForeign = false for -in content")
	}

	sib.X = 60
	window.NotifyMoved(sib)
	if ch.X != 100 {
		t.Errorf("X = %d after the container moved, want 100", ch.X)
	}

	window.MarkUnmapped(sib)
	if s.mapped[ch.PlatformID] || ch.IsMapped() {
		t.Error("-in content stayed mapped when its container was unmapped")
	}
	window.MarkMapped(sib)
	if !ch.IsMapped() {
		t.Error("-in content not remapped with its container")
	}

	Place(ch, X(1))
	if _, ok := placers[sib]; ok {
		t.Error("moving content back to its parent left it in the -in container")
	}
	if containerFor(ch) != top || geom(ch)[0] != 1 {
		t.Errorf("content not placed in its parent: container %v, geometry %v", containerFor(ch).PathName, geom(ch))
	}

	Place(ch, In(sib))
	window.DestroyWindow(sib)
	if _, ok := placers[sib]; ok {
		t.Error("state kept for a destroyed container")
	}
	if ch.GeomManager != nil || ch.IsMapped() {
		t.Error("-in content still managed or mapped after its container was destroyed")
	}
	if _, ok := containerOf[ch]; ok {
		t.Error("containerOf kept content of a destroyed container")
	}
	delete(placers, top)
}

func TestDestroyContentDropsIt(t *testing.T) {
	d, _ := newDisplay()
	c := newContainer(d, 200, 100)
	ch := newChild(c, 30, 20)
	Place(ch)
	window.DestroyWindow(ch)
	if _, ok := placers[c]; ok {
		t.Error("container state kept after its content was destroyed")
	}
	if _, ok := containerOf[ch]; ok {
		t.Error("containerOf kept a destroyed window")
	}
}

func TestPlaceWithoutParent(t *testing.T) {
	w := &window.Window{}
	Place(w, X(1))
	if w.GeomManager != nil {
		t.Error("placed a window with no parent")
	}
}
