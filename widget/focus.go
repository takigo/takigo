package widget

import (
	"github.com/takigo/takigo/focus"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/window"
)

// focuser is the App's focus manager access (takigo.App.FocusManager).
type focuser interface {
	FocusManager() *focus.Manager
}

// Focus gives w the keyboard focus through the application's focus
// manager, as Tk's focus command does, so the manager's state, Tab
// traversal and the redirection of keys to w stay consistent.
func Focus(app AppContext, w *window.Window) {
	if f, ok := app.(focuser); ok {
		f.FocusManager().SetFocus(w)
		return
	}
	app.Server().SetInputFocus(w.PlatformID, platform.RevertToParent, platform.CurrentTime)
}

// FocusWindow returns the window with the application's focus, or nil.
func FocusWindow(app AppContext) *window.Window {
	if f, ok := app.(focuser); ok {
		return f.FocusManager().FocusWindow()
	}
	return nil
}
