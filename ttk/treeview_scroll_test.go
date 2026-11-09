package ttk

import "testing"

func TestScrolled(t *testing.T) {
	tests := []struct {
		first, last, total int
		want               [3]int
	}{
		{0, 100, 400, [3]int{0, 100, 400}},
		{350, 450, 400, [3]int{300, 400, 400}}, // past the end slides back
		{50, 450, 400, [3]int{0, 400, 400}},    // wider than the columns
		{10, 20, 0, [3]int{0, 1, 1}},
	}
	for _, tt := range tests {
		f, l, n := scrolled(tt.first, tt.last, tt.total)
		if got := [3]int{f, l, n}; got != tt.want {
			t.Errorf("scrolled(%d, %d, %d) = %v, want %v", tt.first, tt.last, tt.total, got, tt.want)
		}
	}
}

func TestScrollTo(t *testing.T) {
	tests := []struct {
		newFirst, first, last, total, want int
	}{
		{50, 0, 100, 400, 50},
		{-5, 50, 150, 400, 0},
		{999, 0, 100, 400, 399},
		{310, 300, 400, 400, 300}, // at the end: no further
		{200, 300, 400, 400, 200}, // backwards from the end
	}
	for _, tt := range tests {
		if got := scrollTo(tt.newFirst, tt.first, tt.last, tt.total); got != tt.want {
			t.Errorf("scrollTo(%d, %d, %d, %d) = %d, want %d", tt.newFirst, tt.first, tt.last, tt.total, got, tt.want)
		}
	}
}

func TestItemLookTagPriority(t *testing.T) {
	px := func(v uint64) *uint64 { return &v }
	tv := &Treeview{tags: map[string]*treeTag{}}
	tv.tag("first").background = px(1)
	tv.tag("second").background = px(2)
	tv.tag("second").foreground = px(3)
	tv.tag("stripes").stripedBackground = px(4)

	// The tag created first wins, whatever the item's tag order.
	look := tv.itemLook(&TreeItem{Tags: []string{"second", "first"}}, false)
	if *look.background != 1 || *look.foreground != 3 {
		t.Errorf("look = bg %d fg %d, want bg 1 fg 3", *look.background, *look.foreground)
	}
	// A striped background replaces the background on striped rows only.
	item := &TreeItem{Tags: []string{"first", "stripes"}}
	if got := *tv.itemLook(item, true).background; got != 4 {
		t.Errorf("striped row background = %d, want 4", got)
	}
	if got := *tv.itemLook(item, false).background; got != 1 {
		t.Errorf("plain row background = %d, want 1", got)
	}
	if look := tv.itemLook(&TreeItem{Tags: []string{"unknown"}}, false); look.background != nil {
		t.Error("an unknown tag set a background")
	}
}
