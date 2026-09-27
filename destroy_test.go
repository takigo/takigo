package takigo_test

import (
	"testing"

	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/testutil"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
)

func TestDestroyingParentCleansUpChildWidget(t *testing.T) {
	app := testutil.NewTestApp(t)
	f := frame.New(app, "f")
	b := button.New(f, "b", button.Text("x"))
	pack.Pack(geometry.Group{f})
	pack.Pack(geometry.Group{b})

	id := b.Win.PlatformID
	disp := app.Dispatcher()
	destroyEvents := 0
	disp.Bind(id, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.DestroyType {
			destroyEvents++
		}
	})
	exposed := false
	disp.Bind(id, event.ExposureMask, func(*event.Event) { exposed = true })

	f.Destroy()

	if destroyEvents != 1 {
		t.Errorf("<Destroy> delivered %d times to the child, want 1", destroyEvents)
	}
	if !b.Destroyed {
		t.Error("child widget not marked destroyed with its parent")
	}
	if app.Display().LookupWindow(id) != nil {
		t.Error("child window still registered")
	}
	disp.Dispatch(&event.Event{Type: event.ExposeType, Window: id})
	if exposed {
		t.Error("handler of a destroyed window still bound")
	}
}
