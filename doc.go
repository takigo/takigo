// Package takigo is a Go port of the Tk 9.1 GUI toolkit: the classic and
// themed (ttk) widgets, the pack, grid and place geometry managers, the
// canvas and text widgets, bindings, dialogs and the window-manager
// interface, behind a Go API of typed constructors and functional options.
//
// It has no third-party dependencies. The Windows backend is pure Go; the
// X11 (Linux, BSD) and macOS backends reach Xlib/Xft and AppKit through cgo.
//
// # Applications
//
// An [App] owns the display connection, the root window and the event loop.
// Widgets are created under the App or under another widget, placed with a
// geometry manager, and changed later through their Configure method:
//
//	app, err := takigo.NewApp(takigo.Title("Hello"))
//	if err != nil {
//		log.Fatal(err)
//	}
//	b := button.New(app, "hello", button.Text("Hello, world"), button.Command(app.Quit))
//	pack.Pack(b, pack.PadX(20), pack.PadY(20))
//	app.Run()
//
// # Packages
//
// Classic widgets live in widget/<name> (button, label, entry, listbox, text,
// menu, ...), themed widgets in ttk, the canvas in canvas, and the geometry
// managers in geometry/pack, geometry/grid and geometry/place. Package bind
// holds Tk's binding system, dialog the file, colour, font and message
// dialogs, and wm the window-manager interface.
//
// # Threading
//
// Everything that touches a window runs on the goroutine that calls
// [App.Run]. Other goroutines schedule work on it with [App.RunOnMain],
// [App.DoWhenIdle] and [App.After]; THREADING.md in the repository lists
// which types are safe to use from where.
package takigo
