package takigo_test

import (
	"fmt"
	"testing"

	"github.com/msorc/takigo/bind"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/testutil"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/listbox"
	"github.com/msorc/takigo/widget/scrollbar"
)

func TestMouseWheelScrollsListbox(t *testing.T) {
	app := testutil.NewTestApp(t)
	items := make([]string, 100)
	for i := range items {
		items[i] = fmt.Sprintf("item %d", i)
	}
	first := -1.0
	lb := listbox.New(app, "lb", listbox.Items(items...), listbox.Height(5),
		listbox.YScrollCommand(func(f, _ float64) { first = f }))
	pack.Pack(lb)
	app.UpdateIdleTasks()

	wheel := func(delta int, state uint) {
		app.Dispatcher().Dispatch(&event.Event{Type: event.MouseWheelType, Window: lb.Win.PlatformID, Delta: delta, State: state})
		app.UpdateIdleTasks()
	}
	wheel(-120, 0)
	if first != 0.03 {
		t.Errorf("one notch down: view starts at %v, want 0.03 (3 of 100 items)", first)
	}
	wheel(120, 0)
	if first != 0 {
		t.Errorf("one notch back up: view starts at %v, want 0", first)
	}
	// A high-resolution wheel: four quarter steps make one notch.
	for range 4 {
		wheel(-30, 0)
	}
	if first != 0.03 {
		t.Errorf("four quarter notches: view starts at %v, want 0.03", first)
	}
	// A horizontal wheel does not move the vertical view.
	wheel(-120, platform.ShiftMask)
	if first != 0.03 {
		t.Errorf("a horizontal wheel moved the vertical view to %v", first)
	}
}

func TestMouseWheelOnScrollbarAndBinding(t *testing.T) {
	app := testutil.NewTestApp(t)
	var got []widget.ScrollRequest
	sb := scrollbar.New(app, "sb", scrollbar.CommandOpt(func(r widget.ScrollRequest) { got = append(got, r) }))
	pack.Pack(sb)
	app.UpdateIdleTasks()

	var bound []int
	if err := app.Bind().BindWindow(sb, "<MouseWheel>", func(ed *bind.EventData) bool {
		bound = append(bound, ed.Event.Delta)
		return false
	}); err != nil {
		t.Fatal(err)
	}
	app.Dispatcher().Dispatch(&event.Event{Type: event.MouseWheelType, Window: sb.Win.PlatformID, Delta: -120})
	if len(got) != 1 || got[0] != widget.ScrollUnits(3) {
		t.Errorf("scrollbar command got %+v, want one scroll of 3 units", got)
	}
	if len(bound) != 1 || bound[0] != -120 {
		t.Errorf("<MouseWheel> binding got deltas %v, want [-120]", bound)
	}
}
