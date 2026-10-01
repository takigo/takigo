package canvas

import (
	"slices"
	"testing"

	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
)

// pickScene returns a canvas with filled rectangles "a" at (0,0)-(40,30) and
// "b" at (100,0)-(140,30), and a log of the item events each receives.
func pickScene(t *testing.T) (*Canvas, ItemID, ItemID, *[]string) {
	t.Helper()
	c := newBenchCanvas()
	a := c.createItem(newRectOvalItem("rectangle", 0, 0, 40, 30, c), []ItemOption{Tags("a"), fillPixel(1)})
	b := c.createItem(newRectOvalItem("rectangle", 100, 0, 140, 30, c), []ItemOption{Tags("b"), fillPixel(2)})
	log := new([]string)
	names := map[event.Type]string{
		event.EnterType: "enter", event.LeaveType: "leave", event.MotionType: "motion",
		event.ButtonPressType: "press", event.ButtonReleaseType: "release",
	}
	mask := event.EnterMask | event.LeaveMask | event.MotionMask | event.ButtonPressMask | event.ButtonReleaseMask
	for _, tag := range []string{"a", "b"} {
		c.BindItem(tag, mask, func(ev *event.Event) { *log = append(*log, names[ev.Type]+" "+tag) })
	}
	return c, a, b, log
}

func motion(x, y int, state uint) *event.Event {
	return &event.Event{Type: event.MotionType, X: x, Y: y, State: state}
}

func button(typ event.Type, x, y int, state uint) *event.Event {
	return &event.Event{Type: typ, X: x, Y: y, Button: 1, State: state}
}

// TestPickHoldsCurrentItemWhileButtonDown checks Tk's implicit item grab: once
// button 1 is pressed on an item, motion keeps going to it after the pointer
// leaves it, no other item is entered, and the release reaches it too;
// only then does the item under the pointer become current.
func TestPickHoldsCurrentItemWhileButtonDown(t *testing.T) {
	c, a, b, log := pickScene(t)
	b1 := platform.Button1Mask

	steps := []struct {
		ev      *event.Event
		current ItemID
		log     []string
	}{
		{motion(20, 15, 0), a, []string{"enter a", "motion a"}},
		{button(event.ButtonPressType, 20, 15, 0), a, []string{"press a"}},
		{motion(70, 15, b1), a, []string{"leave a", "motion a"}},
		{motion(120, 15, b1), a, []string{"motion a"}},
		// As in Tk, the release goes to the grabbed item first; the repick
		// after it (button up) sends <Leave> again and enters b.
		{button(event.ButtonReleaseType, 120, 15, b1), b, []string{"release a", "leave a", "enter b"}},
		{motion(125, 15, 0), b, []string{"motion b"}},
	}
	for i, s := range steps {
		*log = nil
		c.handlePointer(s.ev)
		if got := c.CurrentItem(); got != s.current {
			t.Errorf("step %d: current item = %d, want %d", i, got, s.current)
		}
		if !slices.Equal(*log, s.log) {
			t.Errorf("step %d: item events %q, want %q", i, *log, s.log)
		}
	}
}

// TestPickFollowsPointerWithoutButton checks ordinary hovering: each move
// onto another item leaves the old one and enters the new one.
func TestPickFollowsPointerWithoutButton(t *testing.T) {
	c, a, b, log := pickScene(t)
	c.handlePointer(motion(20, 15, 0))
	c.handlePointer(motion(120, 15, 0))
	c.handlePointer(motion(70, 15, 0))
	want := []string{"enter a", "motion a", "leave a", "enter b", "motion b", "leave b"}
	if !slices.Equal(*log, want) {
		t.Errorf("item events %q, want %q", *log, want)
	}
	if got := c.CurrentItem(); got != -1 {
		t.Errorf("current item = %d over empty space, want -1", got)
	}
	_, _ = a, b
}

// TestPickHoldsCurrentItemAcrossWindowLeave checks that dragging out of the
// canvas keeps the grabbed item current until the button is released.
func TestPickHoldsCurrentItemAcrossWindowLeave(t *testing.T) {
	c, a, _, log := pickScene(t)
	b1 := platform.Button1Mask
	c.handlePointer(motion(20, 15, 0))
	c.handlePointer(button(event.ButtonPressType, 20, 15, 0))
	c.handlePointer(&event.Event{Type: event.LeaveType, State: b1})
	if got := c.CurrentItem(); got != a {
		t.Fatalf("current item = %d after leaving the window mid-drag, want %d", got, a)
	}
	if want := []string{"enter a", "motion a", "press a", "leave a"}; !slices.Equal(*log, want) {
		t.Errorf("item events %q, want %q", *log, want)
	}
	*log = nil
	c.handlePointer(button(event.ButtonReleaseType, 500, 500, b1))
	if want := []string{"release a", "leave a"}; !slices.Equal(*log, want) {
		t.Errorf("item events %q, want %q", *log, want)
	}
	if got := c.CurrentItem(); got != -1 {
		t.Errorf("current item = %d after the release outside, want -1", got)
	}
}

// TestPickSkipsItemDeletedByLeaveHandler checks that an item deleted by the
// old item's <Leave> handler is not made current.
func TestPickSkipsItemDeletedByLeaveHandler(t *testing.T) {
	c, _, _, _ := pickScene(t)
	c.BindItem("a", event.LeaveMask, func(*event.Event) { c.Delete("b") })
	c.handlePointer(motion(20, 15, 0))
	c.handlePointer(motion(120, 15, 0))
	if got := c.CurrentItem(); got != -1 {
		t.Errorf("current item = %d, want -1 (b was deleted)", got)
	}
}
