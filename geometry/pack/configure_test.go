package pack

import (
	"slices"
	"testing"

	"github.com/msorc/takigo/geometry"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/window"
)

// g wraps one window for Pack.
func g(w *window.Window) geometry.Group { return geometry.Group{w} }

func packScene() (root, f *window.Window, kids []*window.Window) {
	root = &window.Window{PathName: ".", Display: &window.Display{}}
	f = window.NewChildWindow(root, "f", 0, 0, 10, 10)
	for _, n := range []string{"a", "b", "c", "d", "e"} {
		kids = append(kids, window.NewChildWindow(f, n, 0, 0, 10, 10))
	}
	return root, f, kids
}

// order returns the names of container's content in packing order.
func order(container *window.Window) []string {
	var names []string
	if p, ok := packers.Get(container); ok {
		for _, e := range p.entries {
			names = append(names, e.window.Name)
		}
	}
	return names
}

func configOf(w *window.Window) packConfig {
	return packers.Of(containerFor(w)).entry(w).config
}

// TestPackMergesOptionsOfPackedContent checks that re-packing keeps the
// options not given, as "pack configure" does in Tk, while content packed
// anew (e.g. after Forget) starts from the defaults.
func TestPackMergesOptionsOfPackedContent(t *testing.T) {
	_, f, k := packScene()
	a := k[0]
	Pack(g(a), SideOpt(Left), Expand(true), PadX(5), Anchor(option.AnchorN))
	Pack(g(a), FillOpt(FillX))

	got := configOf(a)
	want := packConfig{side: Left, fill: FillX, expand: true, anchor: option.AnchorN, padLeft: 5, padX: 10}
	if got != want {
		t.Errorf("config after re-pack = %+v, want %+v", got, want)
	}

	Forget(a)
	Pack(g(a), FillOpt(FillY))
	if got, want := configOf(a), (packConfig{side: Top, fill: FillY, anchor: option.AnchorCenter}); got != want {
		t.Errorf("config after forget and pack = %+v, want %+v", got, want)
	}
	if got := order(f); !slices.Equal(got, []string{"a"}) {
		t.Errorf("order = %v", got)
	}
}

func TestPackPosition(t *testing.T) {
	_, f, k := packScene()
	a, b, c, d, e := k[0], k[1], k[2], k[3], k[4]
	Pack(geometry.Group{a, b, c})

	steps := []struct {
		name string
		do   func()
		want []string
	}{
		{"re-pack keeps place", func() { Pack(g(a), SideOpt(Left)) }, []string{"a", "b", "c"}},
		{"before", func() { Pack(g(c), Before(a)) }, []string{"c", "a", "b"}},
		{"after", func() { Pack(g(c), After(b)) }, []string{"a", "b", "c"}},
		{"after last", func() { Pack(g(a), After(c)) }, []string{"b", "c", "a"}},
		{"new group after", func() { Pack(geometry.Group{d, e}, After(b)) }, []string{"b", "d", "e", "c", "a"}},
		{"group chains after the first", func() { Pack(geometry.Group{a, b}, Before(d)) }, []string{"a", "b", "d", "e", "c"}},
		{"in moves to the end", func() { Pack(g(a), In(f)) }, []string{"b", "d", "e", "c", "a"}},
		{"next to itself stays", func() { Pack(g(c), Before(c), SideOpt(Right)) }, []string{"b", "d", "e", "c", "a"}},
		{"sibling not packed", func() { Forget(e); Pack(g(a), Before(e)) }, []string{"b", "d", "c", "a"}},
	}
	for _, s := range steps {
		s.do()
		if got := order(f); !slices.Equal(got, s.want) {
			t.Fatalf("%s: order = %v, want %v", s.name, got, s.want)
		}
	}
	if got := configOf(c).side; got != Right {
		t.Errorf("Before(itself) did not apply the other options: side = %v", got)
	}
}

// TestPackAfterMovesToSiblingsContainer checks that After puts content in
// the sibling's container, which may differ from the content's parent.
func TestPackAfterMovesToSiblingsContainer(t *testing.T) {
	root, f, k := packScene()
	inner := window.NewChildWindow(f, "inner", 0, 0, 10, 10)
	x := window.NewChildWindow(inner, "x", 0, 0, 10, 10)
	a := k[0]
	Pack(geometry.Group{inner, a})
	Pack(g(x), In(f))
	if got := order(f); !slices.Equal(got, []string{"inner", "a", "x"}) {
		t.Fatalf("order = %v", got)
	}
	// x's parent is inner; After(inner) keeps it in f, right after inner.
	Pack(g(x), After(inner))
	if got := order(f); !slices.Equal(got, []string{"inner", "x", "a"}) {
		t.Errorf("order = %v, want [inner x a]", got)
	}
	if got := containerFor(x); got != f {
		t.Errorf("x is managed in %v, want f", got.PathName)
	}
	// Re-packing without a position keeps it in f rather than its parent.
	Pack(g(x), SideOpt(Left))
	if got := containerFor(x); got != f {
		t.Errorf("after re-pack x is managed in %v, want f", got.PathName)
	}
	_ = root
}
