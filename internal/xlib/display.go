//go:build linux || freebsd || openbsd || netbsd

package xlib

/*
#include <stdlib.h>
#include <X11/Xlib.h>
*/
import "C"
import "unsafe"

// OpenDisplay opens a connection to the X server.
// If name is empty, it uses the DISPLAY environment variable.
func OpenDisplay(name string) (*Display, error) {
	var cname *C.char
	if name != "" {
		cname = C.CString(name)
		defer C.free(unsafe.Pointer(cname))
	}
	dpy := C.XOpenDisplay(cname)
	if dpy == nil {
		return nil, ErrNoDisplay
	}
	return &Display{ptr: dpy}, nil
}

// Close closes the display connection, freeing XIM/XIC resources first.
func (d *Display) Close() {
	if d.ptr != nil {
		if d.xic != nil {
			C.XDestroyIC(d.xic)
			d.xic = nil
		}
		if d.xim != nil {
			C.XCloseIM(d.xim)
			d.xim = nil
		}
		C.XCloseDisplay(d.ptr)
		d.ptr = nil
	}
}

// DefaultScreen returns the default screen number.
func (d *Display) DefaultScreen() int {
	return int(C.XDefaultScreen(d.ptr))
}

// DefaultRootWindow returns the root window of the default screen.
func (d *Display) DefaultRootWindow() Window {
	return Window(C.XDefaultRootWindow(d.ptr))
}

// RootWindow returns the root window of the given screen.
func (d *Display) RootWindow(screen int) Window {
	return Window(C.XRootWindow(d.ptr, C.int(screen)))
}

// DefaultDepth returns the default depth of the given screen.
func (d *Display) DefaultDepth(screen int) int {
	return int(C.XDefaultDepth(d.ptr, C.int(screen)))
}

// DefaultVisual returns the default visual of the given screen.
func (d *Display) DefaultVisual(screen int) *Visual {
	v := C.XDefaultVisual(d.ptr, C.int(screen))
	return &Visual{ptr: v}
}

// DefaultColormap returns the default colormap of the given screen.
func (d *Display) DefaultColormap(screen int) Colormap {
	return Colormap(C.XDefaultColormap(d.ptr, C.int(screen)))
}

// ScreenWidth returns the width in pixels of the given screen.
func (d *Display) ScreenWidth(screen int) int {
	return int(C.XDisplayWidth(d.ptr, C.int(screen)))
}

// ScreenHeight returns the height in pixels of the given screen.
func (d *Display) ScreenHeight(screen int) int {
	return int(C.XDisplayHeight(d.ptr, C.int(screen)))
}

// ScreenWidthMM returns the width in millimeters of the given screen.
func (d *Display) ScreenWidthMM(screen int) int {
	return int(C.XDisplayWidthMM(d.ptr, C.int(screen)))
}

// ScreenHeightMM returns the height in millimeters of the given screen.
func (d *Display) ScreenHeightMM(screen int) int {
	return int(C.XDisplayHeightMM(d.ptr, C.int(screen)))
}

// WhitePixel returns the white pixel value for the given screen.
func (d *Display) WhitePixel(screen int) uint64 {
	return uint64(C.XWhitePixel(d.ptr, C.int(screen)))
}

// BlackPixel returns the black pixel value for the given screen.
func (d *Display) BlackPixel(screen int) uint64 {
	return uint64(C.XBlackPixel(d.ptr, C.int(screen)))
}

// ConnectionNumber returns the file descriptor of the display connection.
func (d *Display) ConnectionNumber() int {
	return int(C.XConnectionNumber(d.ptr))
}

// Sync flushes the output buffer and waits for all requests to complete.
func (d *Display) Sync(discard bool) {
	var disc C.int
	if discard {
		disc = 1
	}
	C.XSync(d.ptr, disc)
}

// Flush flushes the output buffer.
func (d *Display) Flush() {
	C.XFlush(d.ptr)
}

// Pending returns the number of events waiting in the event queue.
func (d *Display) Pending() int {
	return int(C.XPending(d.ptr))
}

// ResourceManagerString returns the RESOURCE_MANAGER property from
// the root window, where xrdb stores settings like Xft.dpi.
// Returns empty string if no property is set.
func (d *Display) ResourceManagerString() string {
	cs := C.XResourceManagerString(d.ptr)
	if cs == nil {
		return ""
	}
	return C.GoString(cs)
}
