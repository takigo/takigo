package takigo_test

import (
	"testing"
	"time"

	"github.com/msorc/takigo/dialog"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/testutil"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/window"
)

// A modal dialog opened from a callback runs a nested loop. Callbacks
// queued behind that callback, and the redraws they ask for, run while the
// dialog is open, and the dialog's own content is laid out.
func TestModalDialogKeepsLoopWorking(t *testing.T) {
	app := testutil.NewTestApp(t)
	lbl := label.New(app, "l", label.Text("before"))
	pack.Pack(geometry.Group{lbl})

	result := dialog.ResultNone
	ranBehind, redrawn := false, false
	app.RunOnMain(func() {
		result = dialog.ShowMessage(app, dialog.MsgMessage("modal"), dialog.MsgButtons(dialog.BtnOKCancel))
		app.Quit()
	})
	app.RunOnMain(func() {
		ranBehind = true
		var top *window.Window
		for _, c := range app.Root().Children {
			if c.Name == "dialog" {
				top = c
			}
		}
		if top == nil {
			t.Error("the dialog was not open when the callback queued behind it ran")
			app.Quit()
			return
		}
		app.UpdateIdleTasks()
		laidOut := 0
		var check func(w *window.Window)
		check = func(w *window.Window) {
			if w.Class == "Button" || w.Class == "Label" {
				laidOut++
				if w.Width <= 1 || w.Height <= 1 || !w.IsMapped() {
					t.Errorf("%s is %dx%d, mapped=%v: the dialog was not laid out", w.PathName, w.Width, w.Height, w.IsMapped())
				}
			}
			for _, c := range w.Children {
				check(c)
			}
		}
		check(top)
		if laidOut < 4 {
			t.Errorf("found %d buttons and labels in the dialog, want its icon, message and two buttons", laidOut)
		}

		// A redraw requested behind the dialog happens while it is open.
		lbl.Configure(label.Text("behind the dialog"))
		app.DoWhenIdle(func() {
			redrawn = true
			app.Dispatcher().Dispatch(&event.Event{Type: event.KeyPressType, Window: top.PlatformID, KeySym: platform.XK_Return})
		})
	})
	app.After(10*time.Second, func() {
		t.Error("the dialog did not close")
		app.Quit()
	})
	app.MainLoop()

	if !ranBehind || !redrawn {
		t.Errorf("while the dialog was open: callback ran=%v, idle work ran=%v; want both", ranBehind, redrawn)
	}
	if result != dialog.ResultOK {
		t.Errorf("dialog result = %v, want ResultOK", result)
	}
}
