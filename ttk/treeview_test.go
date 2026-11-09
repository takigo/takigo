package ttk_test

import (
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/takigo/takigo/bind"
	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/geometry/place"
	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/ttk"
)

func TestTreeviewStructure(t *testing.T) {
	app := testutil.NewTestApp(t)
	tv := ttk.NewTreeview(app, "tv", ttk.TreeviewColumns("a", "b"))
	p := tv.Insert("", -1, ttk.ItemID("p"))
	c1 := tv.Insert(p, -1, ttk.ItemID("c1"))
	c2 := tv.Insert(p, -1, ttk.ItemID("c2"), ttk.ItemValues("x"))
	q := tv.Insert("", -1, ttk.ItemID("q"))

	if tv.Next(c1) != c2 || tv.Prev(c2) != c1 || tv.Next(c2) != "" || tv.Prev(p) != "" {
		t.Errorf("next/prev: next(c1)=%q prev(c2)=%q next(c2)=%q prev(p)=%q",
			tv.Next(c1), tv.Prev(c2), tv.Next(c2), tv.Prev(p))
	}
	if tv.Index(c2) != 1 || tv.Index(q) != 1 {
		t.Errorf("index(c2)=%d index(q)=%d, want 1 1", tv.Index(c2), tv.Index(q))
	}

	// Cell values: by ID, "#n" display number and index; padded on set.
	if tv.CellValue(c2, "a") != "x" || tv.CellValue(c2, "#1") != "x" || tv.CellValue(c2, "b") != "" {
		t.Errorf("CellValue: a=%q #1=%q b=%q", tv.CellValue(c2, "a"), tv.CellValue(c2, "#1"), tv.CellValue(c2, "b"))
	}
	tv.SetCellValue(c1, "1", "y")
	if got := tv.Item(c1).Values; !slices.Equal(got, []string{"", "y"}) {
		t.Errorf("SetCellValue(c1, 1): values %q", got)
	}
	if got := tv.CellValues(c1); got["a"] != "" || got["b"] != "y" {
		t.Errorf("CellValues(c1) = %v", got)
	}

	// Moving an item under its own descendant is refused.
	tv.Move(p, c1, 0)
	if tv.Parent(p) != "" {
		t.Errorf("p moved under its child: parent %q", tv.Parent(p))
	}

	// Detach keeps the item and its children; Move puts it back.
	tv.Detach(p)
	if !tv.IsDetached(p) || tv.IsDetached(c1) || !slices.Equal(tv.Detached(), []string{p}) {
		t.Errorf("after Detach: IsDetached(p)=%v IsDetached(c1)=%v Detached=%v", tv.IsDetached(p), tv.IsDetached(c1), tv.Detached())
	}
	if got := tv.Children(""); !slices.Equal(got, []string{q}) {
		t.Errorf("root children after Detach = %v", got)
	}
	tv.Move(p, "", 0)
	if tv.IsDetached(p) || !slices.Equal(tv.Children(""), []string{p, q}) || !slices.Equal(tv.Children(p), []string{c1, c2}) {
		t.Errorf("after reattach: root %v, p %v", tv.Children(""), tv.Children(p))
	}

	// The selection lists items in tree order, closed subtrees included.
	tv.SelectionSet(q, c2)
	if got := tv.Selection(); !slices.Equal(got, []string{c2, q}) {
		t.Errorf("Selection = %v, want [c2 q]", got)
	}
	tv.SelectionToggle(c2, c1)
	if got := tv.Selection(); !slices.Equal(got, []string{c1, q}) {
		t.Errorf("Selection after toggle = %v, want [c1 q]", got)
	}

	// See opens the ancestors.
	tv.See(c1)
	if !tv.Item(p).Open {
		t.Error("See did not open the parent")
	}
}

func TestTreeviewTags(t *testing.T) {
	app := testutil.NewTestApp(t)
	tv := ttk.NewTreeview(app, "tv")
	a := tv.Insert("", -1, ttk.ItemTags("red"))
	b := tv.Insert("", -1)
	tv.TagAdd("blue", a, b)
	tv.TagConfigure("green", ttk.TagForeground(0x00ff00))

	if got := tv.TagNames(); !slices.Equal(got, []string{"red", "blue", "green"}) {
		t.Errorf("TagNames = %v", got)
	}
	if got := tv.TagHas("blue"); !slices.Equal(got, []string{a, b}) {
		t.Errorf("TagHas(blue) = %v", got)
	}
	tv.TagRemove("blue", a)
	if tv.HasTag(a, "blue") || !tv.HasTag(b, "blue") {
		t.Errorf("after TagRemove(blue, a): a %v, b %v", tv.HasTag(a, "blue"), tv.HasTag(b, "blue"))
	}
	tv.TagRemove("blue")
	if len(tv.TagHas("blue")) != 0 {
		t.Error("TagRemove without items left the tag on an item")
	}
	tv.TagDelete("red")
	if tv.HasTag(a, "red") || slices.Contains(tv.TagNames(), "red") {
		t.Error("TagDelete(red) kept the tag")
	}
}

func TestTreeviewIdentifyAndBBox(t *testing.T) {
	app := testutil.NewTestApp(t)
	tv := ttk.NewTreeview(app, "tv", ttk.TreeviewColumns("a", "b", "c"),
		ttk.TreeviewDisplayColumns("c", "a"))
	tv.ColumnConfigure("#0", ttk.ColWidth(100), ttk.ColStretch(false))
	for _, id := range []string{"a", "b", "c"} {
		tv.ColumnConfigure(id, ttk.ColWidth(50), ttk.ColStretch(false))
	}
	p := tv.Insert("", -1, ttk.ItemText("parent"), ttk.ItemOpen(true))
	c := tv.Insert(p, -1, ttk.ItemText("child"), ttk.ItemValues("va", "vb", "vc"))
	place.Place(tv, place.X(0), place.Y(0), place.Width(202), place.Height(200))

	type result struct {
		bboxRow, bboxCell, bboxTree [4]int
		okHidden                    bool
		regions                     []ttk.TreeRegion
		cols                        []string
		item                        string
		element                     string
		cellC                       string
	}
	var r result
	app.After(50*time.Millisecond, func() {
		defer app.Quit()
		bb := func(id, col string) [4]int {
			x, y, w, h, _ := tv.BBox(id, col)
			return [4]int{x, y, w, h}
		}
		r.bboxRow = bb(c, "")
		r.bboxCell = bb(c, "a") // second shown data column
		r.bboxTree = bb(c, "#0")
		_, _, _, _, r.okHidden = tv.BBox(c, "b")
		_, rowY, _, rowH, _ := tv.BBox(c, "")
		y := rowY + rowH/2
		head := 5
		r.regions = []ttk.TreeRegion{
			tv.IdentifyRegion(50, head),  // tree heading
			tv.IdentifyRegion(101, head), // edge between #0 and #1
			tv.IdentifyRegion(50, y),     // tree column of the child
			tv.IdentifyRegion(130, y),    // cell c
			tv.IdentifyRegion(50, 190),   // below the items
		}
		r.cols = []string{tv.IdentifyColumn(50), tv.IdentifyColumn(130), tv.IdentifyColumn(170), tv.IdentifyColumn(250)}
		r.item = tv.IdentifyItem(y)
		_, py, _, ph, _ := tv.BBox(p, "")
		r.element = tv.IdentifyElement(3, py+ph/2)
		r.cellC = tv.CellValue(c, "#1")
	})
	app.MainLoop()

	fb := 1
	rowH := r.bboxRow[3]
	if r.bboxRow[0] != fb || r.bboxRow[2] != 200 || r.bboxRow[1] <= fb {
		t.Errorf("row bbox = %v, want x %d width 200 below the heading", r.bboxRow, fb)
	}
	if r.bboxCell != [4]int{fb + 150, r.bboxRow[1], 50, rowH} {
		t.Errorf("cell a bbox = %v, want x %d (after #0 and c), width 50", r.bboxCell, fb+150)
	}
	indent := r.bboxTree[0] - fb
	if indent <= 0 || r.bboxTree[2] != 100-indent {
		t.Errorf("tree bbox = %v: the child's indent is not taken off", r.bboxTree)
	}
	if r.okHidden {
		t.Error("BBox of a column -displaycolumns hides is ok")
	}
	want := []ttk.TreeRegion{ttk.RegionHeading, ttk.RegionSeparator, ttk.RegionTree, ttk.RegionCell, ttk.RegionNothing}
	if !slices.Equal(r.regions, want) {
		t.Errorf("regions = %v, want %v", r.regions, want)
	}
	if !slices.Equal(r.cols, []string{"#0", "#1", "#2", ""}) {
		t.Errorf("columns = %v, want [#0 #1 #2 \"\"]", r.cols)
	}
	if r.item != c {
		t.Errorf("IdentifyItem = %q, want %q", r.item, c)
	}
	if r.element != "Treeitem.indicator" {
		t.Errorf("element at the parent's indicator = %q", r.element)
	}
	if r.cellC != "vc" {
		t.Errorf("CellValue(#1) = %q, want vc (the first displayed column)", r.cellC)
	}
}

func TestTreeviewEvents(t *testing.T) {
	app := testutil.NewTestApp(t)
	tv := ttk.NewTreeview(app, "tv")
	p := tv.Insert("", -1, ttk.ItemTags("hot"))
	tv.Insert(p, -1)
	place.Place(tv, place.X(0), place.Y(0), place.Width(200), place.Height(200))

	var events []string
	for _, name := range []string{ttk.TreeviewSelectEvent, ttk.TreeviewOpenEvent, ttk.TreeviewCloseEvent} {
		if err := app.Bind().BindWindow(tv, "<<"+name+">>", func(*bind.EventData) bool {
			events = append(events, name)
			return false
		}); err != nil {
			t.Fatal(err)
		}
	}
	var tagged []string
	tv.TagBind("hot", event.ButtonPressMask|event.KeyPressMask, func(ev *event.Event) {
		tagged = append(tagged, strconv.Itoa(int(ev.Type)))
	})

	app.After(50*time.Millisecond, func() {
		tv.SelectionSet(p)
		tv.SetFocus(p)
		key := func(ks platform.KeySym) {
			app.Dispatcher().Dispatch(&event.Event{Type: event.KeyPressType, Window: tv.Win.PlatformID, KeySym: ks})
		}
		key(platform.XK_Right) // OpenItem
		key(platform.XK_Left)  // CloseItem
		_, y, _, h, _ := tv.BBox(p, "")
		app.Dispatcher().Dispatch(&event.Event{Type: event.ButtonPressType, Window: tv.Win.PlatformID,
			Button: 1, X: 50, Y: y + h/2})
		app.Dispatcher().Dispatch(&event.Event{Type: event.ButtonReleaseType, Window: tv.Win.PlatformID,
			Button: 1, X: 50, Y: y + h/2})
	})
	app.After(150*time.Millisecond, app.Quit)
	app.MainLoop()

	for _, name := range []string{ttk.TreeviewSelectEvent, ttk.TreeviewOpenEvent, ttk.TreeviewCloseEvent} {
		if !slices.Contains(events, name) {
			t.Errorf("no <<%s>> in %v", name, events)
		}
	}
	if oi, ci := slices.Index(events, ttk.TreeviewOpenEvent), slices.Index(events, ttk.TreeviewCloseEvent); oi > ci {
		t.Errorf("events out of order: %v", events)
	}
	if len(tagged) < 3 {
		t.Errorf("tag bindings saw %v, want two key presses and a button press", tagged)
	}
}
