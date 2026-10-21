package spinbox

import (
	"strconv"
	"testing"
	"time"

	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/widget"
)

// TestSpinboxAutoRepeat ports the auto-repeat of the spinner buttons:
// ::tk::spinbox::Invoke (tk/library/spinbox.tcl:301) re-invokes the
// element every -repeatinterval ms once -repeatdelay ms have passed, and
// the release stops it.
func TestSpinboxAutoRepeat(t *testing.T) {
	app := testutil.NewTestApp(t)
	s := New(app, "s",
		FromOpt(0), ToOpt(10000), IncrementOpt(1),
		RepeatDelayOpt(20), RepeatIntervalOpt(20))

	// A spinbox of a known size, so the hit test finds the buttons.
	s.Win.Width, s.Win.Height = 200, 30
	s.computeGeometry()

	startx, top, _ := s.buttonBox()
	app.Dispatcher().Dispatch(&event.Event{
		Type: event.ButtonPressType, Window: s.Win.PlatformID,
		X: startx + 2, Y: top + 2, Button: 1,
	})
	if got := s.GetText(); got != "1" {
		t.Fatalf("the press left the value at %q, want 1", got)
	}

	// The value keeps increasing while the button is held.
	app.After(300*time.Millisecond, func() { app.Quit() })
	app.MainLoop()
	value, err := strconv.Atoi(s.GetText())
	if err != nil {
		t.Fatalf("value %q is not a number", s.GetText())
	}
	if value < 3 {
		t.Fatalf("a held button left the value at %d, want the action repeated", value)
	}

	// Releasing the button stops the repeat.
	app.Dispatcher().Dispatch(&event.Event{
		Type: event.ButtonReleaseType, Window: s.Win.PlatformID, Button: 1,
	})
	app.After(200*time.Millisecond, func() { app.Quit() })
	app.MainLoop()
	if s.GetText() != strconv.Itoa(value) {
		t.Errorf("the value changed after the release: %d -> %q", value, s.GetText())
	}
}

// TestSpinboxAutoScan ports ::tk::spinbox::AutoScan (tk/library/spinbox.tcl:565):
// the text is scrolled towards the pointer while the pointer is held
// outside the spinbox.
func TestSpinboxAutoScan(t *testing.T) {
	app := testutil.NewTestApp(t)
	s := New(app, "s", FromOpt(0), ToOpt(10000))
	s.Win.Width, s.Win.Height = 200, 30
	s.computeGeometry()
	s.Text = []rune("0123456789012345678901234567890123456789")
	s.computeGeometry()

	// The pointer leaves the spinbox on the right with button 1 held.
	app.Dispatcher().Dispatch(&event.Event{
		Type: event.LeaveType, Window: s.Win.PlatformID,
		X: s.Win.Width + 5, Y: 10, State: platform.Button1Mask,
	})
	app.After(200*time.Millisecond, func() { app.Quit() })
	app.MainLoop()
	if s.LeftIndex == 0 {
		t.Error("the auto-scan did not scroll the text")
	}
}

// TestSpinboxOptions checks the options that were missing: the colours and
// the state of the spinbox could not be set before (Tk's
// -selectbackground, -selectforeground, -insertbackground and -state).
func TestSpinboxOptions(t *testing.T) {
	app := testutil.NewTestApp(t)
	s := New(app, "s",
		SelectBackground("red"),
		SelectForeground("blue"),
		InsertBackground("green"),
		State(widget.StateNormal),
	)
	if s.SelBg == nil {
		t.Error("SelectBackground did not set the selection background")
	}
	if s.SelFg == nil {
		t.Error("SelectForeground did not set the selection foreground")
	}
	if s.InsertBg == nil {
		t.Error("InsertBackground did not set the insert background")
	}
	if err := s.Configure(SelectBackground("yellow")); err != nil {
		t.Fatalf("Configure: %v", err)
	}
	if s.SelBg == nil {
		t.Error("Configure did not apply the selection background")
	}
}

// TestSpinboxState checks that a disabled or read-only spinbox cannot be
// edited, as Tk's insert and delete commands do not (tk/generic/tkEntry.c).
func TestSpinboxState(t *testing.T) {
	app := testutil.NewTestApp(t)

	s := New(app, "s", State(widget.StateDisabled))
	s.Insert(0, "x")
	if s.GetText() != "0" {
		t.Errorf("a disabled spinbox was edited: %q", s.GetText())
	}

	r := New(app, "r", ReadOnly(true))
	r.Insert(0, "x")
	if r.GetText() != "0" {
		t.Errorf("a read-only spinbox was edited: %q", r.GetText())
	}
	// A read-only spinbox can still be selected and copied.
	r.SelectRange(0, 1)
	if !r.HasSelection() {
		t.Error("a read-only spinbox cannot be selected")
	}
}
