// Package window provides the core window and display structures,
// porting TkWindow and TkDisplay from Tk's tkInt.h.
package window

import (
	"github.com/msorc/takigo/internal/xlib"
)

// Display holds per-display state shared across all windows on a single
// X11 display connection. Ports TkDisplay from tk/generic/tkInt.h.
type Display struct {
	XDisplay *xlib.Display
	Name     string
	Screen   int
	Depth    int
	Visual   *xlib.Visual
	Colormap xlib.Colormap

	// Window lookup: X window ID -> *Window.
	Windows map[xlib.Window]*Window

	// Root window of the default screen.
	RootXWindow xlib.Window

	// Pixel values.
	WhitePixel uint64
	BlackPixel uint64

	// WM atoms.
	WMDeleteWindow xlib.Atom
	WMProtocols    xlib.Atom
}

// NewDisplay opens an X11 display and initializes the Display struct.
func NewDisplay(name string) (*Display, error) {
	xdpy, err := xlib.OpenDisplay(name)
	if err != nil {
		return nil, err
	}

	screen := xdpy.DefaultScreen()

	d := &Display{
		XDisplay:    xdpy,
		Name:        name,
		Screen:      screen,
		Depth:       xdpy.DefaultDepth(screen),
		Visual:      xdpy.DefaultVisual(screen),
		Colormap:    xdpy.DefaultColormap(screen),
		Windows:     make(map[xlib.Window]*Window),
		RootXWindow: xdpy.DefaultRootWindow(),
		WhitePixel:  xdpy.WhitePixel(screen),
		BlackPixel:  xdpy.BlackPixel(screen),
	}

	// Intern WM atoms.
	d.WMDeleteWindow = xdpy.InternAtom("WM_DELETE_WINDOW", false)
	d.WMProtocols = xdpy.InternAtom("WM_PROTOCOLS", false)

	return d, nil
}

// Close closes the display connection.
func (d *Display) Close() {
	if d.XDisplay != nil {
		d.XDisplay.Close()
		d.XDisplay = nil
	}
}

// RegisterWindow adds a window to the lookup table.
func (d *Display) RegisterWindow(xwin xlib.Window, w *Window) {
	d.Windows[xwin] = w
}

// UnregisterWindow removes a window from the lookup table.
func (d *Display) UnregisterWindow(xwin xlib.Window) {
	delete(d.Windows, xwin)
}

// LookupWindow finds a Window by its X11 window ID.
func (d *Display) LookupWindow(xwin xlib.Window) *Window {
	return d.Windows[xwin]
}
