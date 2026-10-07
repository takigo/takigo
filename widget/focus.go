package widget

import "github.com/takigo/takigo/window"

// Focus gives w the keyboard focus through the application's focus
// manager, as Tk's focus command does, so the manager's state, Tab
// traversal and the redirection of keys to w stay consistent.
func Focus(app AppContext, w *window.Window) {
	app.FocusManager().SetFocus(w)
}

// FocusWindow returns the window with the application's focus, or nil.
func FocusWindow(app AppContext) *window.Window {
	return app.FocusManager().FocusWindow()
}
