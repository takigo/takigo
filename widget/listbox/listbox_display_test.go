package listbox_test

import (
	"slices"
	"testing"

	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/widget/listbox"
)

func TestListboxItems(t *testing.T) {
	app := testutil.NewTestApp(t)
	lb := listbox.New(app, "lb", listbox.Items("a", "b", "c"))
	lb.Insert(1, "x", "y")
	lb.Insert(99, "z")
	if got := lb.Items(); !slices.Equal(got, []string{"a", "x", "y", "b", "c", "z"}) {
		t.Fatalf("items = %q", got)
	}
	lb.Delete(1, 2)
	if got, n := lb.Items(), lb.ItemCount(); !slices.Equal(got, []string{"a", "b", "c", "z"}) || n != 4 {
		t.Errorf("after Delete(1, 2): %q, count %d", got, n)
	}
	lb.Delete(3, 99)
	if n := lb.ItemCount(); n != 3 {
		t.Errorf("Delete past the end left %d items, want 3", n)
	}
}

func TestListboxSelection(t *testing.T) {
	app := testutil.NewTestApp(t)
	lb := listbox.New(app, "lb", listbox.Items("a", "b", "c", "d"), listbox.SelectModeOpt(listbox.SelectExtended))
	lb.SelectionSet(1, 2)
	if got := lb.Selection(); !slices.Equal(got, []int{1, 2}) {
		t.Errorf("Selection = %v, want [1 2]", got)
	}
	lb.SelectionClear(2, 2)
	if got := lb.Selection(); !slices.Equal(got, []int{1}) {
		t.Errorf("after SelectionClear(2, 2): %v, want [1]", got)
	}
	lb.Delete(0, 0)
	if got := lb.Selection(); !slices.Equal(got, []int{0}) {
		t.Errorf("selection did not follow the deleted item: %v, want [0]", got)
	}
}

func TestListboxGeometryAndConfigure(t *testing.T) {
	app := testutil.NewTestApp(t)
	short := listbox.New(app, "short", listbox.Items("a"), listbox.Width(10), listbox.Height(2))
	tall := listbox.New(app, "tall", listbox.Items("a"), listbox.Width(10), listbox.Height(6))
	if tall.Win.ReqHeight <= short.Win.ReqHeight {
		t.Errorf("Height(6) is not taller than Height(2): %d vs %d", tall.Win.ReqHeight, short.Win.ReqHeight)
	}
	if err := short.Configure(listbox.Width(30)); err != nil || short.Win.ReqWidth <= tall.Win.ReqWidth {
		t.Errorf("Configure(Width(30)): err %v, width %d vs %d", err, short.Win.ReqWidth, tall.Win.ReqWidth)
	}
	if err := short.Configure(listbox.Background("no-such-colour")); err == nil {
		t.Error("a bad colour gave no error")
	}
}
