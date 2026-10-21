package ttk

import (
	"testing"

	"github.com/takigo/takigo/platform"
)

// fixedElement is a test element with a fixed size and no padding.
type fixedElement struct {
	w, h int
}

func (e fixedElement) Size(State) (int, int, Padding) { return e.w, e.h, Padding{} }

func (e fixedElement) Draw(platform.DisplayServer, platform.DrawableID, platform.GCID, Box, State) {}

// layoutTestTheme returns a theme whose elements all have the given size,
// and lays out tmpl in a 100x40 box.
func layoutTestTheme(t *testing.T, tmpl *LayoutTemplate, ew, eh int) *Layout {
	t.Helper()
	theme := NewTheme("layout-test", nil)
	factory := func(*DrawContext) Element { return fixedElement{w: ew, h: eh} }
	for _, name := range []string{"outer", "inner", "main", "top"} {
		theme.RegisterElement(name, factory)
	}
	lay := NewLayout(tmpl, theme, &DrawContext{}, theme.GetStyle("."))
	lay.Place(0, Box{X: 0, Y: 0, Width: 100, Height: 40})
	return lay
}

func nodeParcel(t *testing.T, lay *Layout, name string) Box {
	t.Helper()
	node := findNode(lay.Root, name)
	if node == nil {
		t.Fatalf("layout has no element %q", name)
	}
	return node.Parcel
}

// TestLayoutFillAxesAreIndependent checks that a fill flag only stretches
// its own axis, as TTK_FILL_X and TTK_FILL_Y do in Tk
// (tk/generic/ttk/ttkTheme.h).
func TestLayoutFillAxesAreIndependent(t *testing.T) {
	tests := []struct {
		name  string
		flags uint
		want  Box
	}{
		{"fill x", FillXF, Box{X: 0, Y: 15, Width: 100, Height: 10}},
		{"fill y", FillYF, Box{X: 45, Y: 0, Width: 10, Height: 40}},
		{"fill both", FillF, Box{X: 0, Y: 0, Width: 100, Height: 40}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lay := layoutTestTheme(t, L("outer", FillF, L("inner", tc.flags)), 10, 10)
			if got := nodeParcel(t, lay, "inner"); got != tc.want {
				t.Errorf("inner parcel = %+v, want %+v", got, tc.want)
			}
		})
	}
}

// TestLayoutExpandUsesWholeCavity checks that an expanding element gets the
// whole cavity (Ttk_PositionBox: TTK_EXPAND has priority over the pack
// side) and fills it.
func TestLayoutExpandUsesWholeCavity(t *testing.T) {
	lay := layoutTestTheme(t, L("outer", FillF, L("main", Expand|FillF)), 10, 10)
	want := Box{X: 0, Y: 0, Width: 100, Height: 40}
	if got := nodeParcel(t, lay, "main"); got != want {
		t.Errorf("main parcel = %+v, want %+v", got, want)
	}
}

// TestLayoutExpandBeatsPackSide checks that an expanding element ignores its
// pack side, as Ttk_PositionBox does.
func TestLayoutExpandBeatsPackSide(t *testing.T) {
	lay := layoutTestTheme(t, L("outer", FillF, L("top", PackTop|Expand|FillF)), 10, 10)
	want := Box{X: 0, Y: 0, Width: 100, Height: 40}
	if got := nodeParcel(t, lay, "top"); got != want {
		t.Errorf("top parcel = %+v, want %+v", got, want)
	}
}

// TestLayoutPackSideTakesParcel checks that a packed element without expand
// takes its parcel from the cavity.
func TestLayoutPackSideTakesParcel(t *testing.T) {
	lay := layoutTestTheme(t, L("outer", FillF, L("top", PackTop|FillF)), 10, 10)
	want := Box{X: 0, Y: 0, Width: 100, Height: 10}
	if got := nodeParcel(t, lay, "top"); got != want {
		t.Errorf("top parcel = %+v, want %+v", got, want)
	}
}
