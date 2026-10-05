package widget_test

import (
	"testing"

	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/window"
)

func TestEventuallyRedrawCoalescesOffscreen(t *testing.T) {
	app := testutil.NewTestApp(t)
	w := window.NewChildWindow(app.Window(), "redraw", 0, 0, 20, 10)
	window.MakeWindowExist(w)
	var b widget.Base
	widget.InitBase(&b, w, app)

	onscreen := platform.WindowDrawable(w.PlatformID)
	calls := 0
	var target platform.DrawableID
	b.SetDisplayProc(func() {
		calls++
		target = w.Drawable()
	})

	for range 10 {
		b.EventuallyRedraw()
	}
	if calls != 0 {
		t.Fatalf("display ran %d times before idle, want deferred", calls)
	}
	app.UpdateIdleTasks()
	if calls != 1 {
		t.Fatalf("display ran %d times for 10 requests, want 1", calls)
	}
	if target == onscreen {
		t.Error("display drew straight to the window, want an off-screen pixmap")
	}
	if got := w.Drawable(); got != onscreen {
		t.Errorf("Drawable() = %v after redraw, want the window %v", got, onscreen)
	}

	b.EventuallyRedraw()
	window.DestroyWindow(w)
	app.UpdateIdleTasks()
	if calls != 1 {
		t.Errorf("display ran after the window was destroyed")
	}
}
