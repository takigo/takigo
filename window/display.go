// Package window provides the core window and display structures,
// porting TkWindow and TkDisplay from Tk's tkInt.h.
package window

import (
	"github.com/msorc/takigo/platform"
)

// Display holds per-display state shared across all windows on a single
// display connection. Ports TkDisplay from tk/generic/tkInt.h.
type Display struct {
	Server platform.DisplayServer
	Name   string
	Screen int
	Depth  int

	// Window lookup: platform window ID -> *Window.
	Windows map[platform.WindowID]*Window

	// Root window of the default screen.
	RootWindow platform.WindowID

	// Pixel values.
	WhitePixel uint64
	BlackPixel uint64

	// WM atoms.
	WMDeleteWindow platform.AtomID
	WMProtocols    platform.AtomID

	// destroyHooks run for every window destroyed on this display, before
	// the window's own hooks; see OnWindowDestroy.
	destroyHooks []func(*Window)
}

// OnWindowDestroy registers fn to run for every window destroyed on this
// display, after its descendants are gone and while it is still
// registered, so fn can deliver <Destroy> and drop per-window state
// (event handlers, bind tags, focus) as Tk_DestroyWindow does.
func (d *Display) OnWindowDestroy(fn func(*Window)) {
	d.destroyHooks = append(d.destroyHooks, fn)
}

// NewDisplay opens a display and initializes the Display struct.
func NewDisplay(server platform.DisplayServer) (*Display, error) {
	screen := server.DefaultScreen()

	d := &Display{
		Server:     server,
		Screen:     screen,
		Depth:      server.DefaultDepth(screen),
		Windows:    make(map[platform.WindowID]*Window),
		RootWindow: server.DefaultRootWindow(),
		WhitePixel: server.WhitePixel(screen),
		BlackPixel: server.BlackPixel(screen),
	}

	// Intern WM atoms.
	d.WMDeleteWindow = server.InternAtom("WM_DELETE_WINDOW", false)
	d.WMProtocols = server.InternAtom("WM_PROTOCOLS", false)

	return d, nil
}

// Close closes the display connection.
func (d *Display) Close() {
	if d.Server != nil {
		d.Server.Close()
		d.Server = nil
	}
}

// RegisterWindow adds a window to the lookup table.
func (d *Display) RegisterWindow(wid platform.WindowID, w *Window) {
	d.Windows[wid] = w
}

// UnregisterWindow removes a window from the lookup table.
func (d *Display) UnregisterWindow(wid platform.WindowID) {
	delete(d.Windows, wid)
}

// LookupWindow finds a Window by its platform window ID.
func (d *Display) LookupWindow(wid platform.WindowID) *Window {
	return d.Windows[wid]
}
