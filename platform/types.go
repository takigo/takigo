// Package platform defines the cross-platform abstraction layer for takigo.
// It provides interfaces and opaque handle types that allow the widget layer
// to be independent of any specific windowing system (X11, Win32, Cocoa).
package platform

// WindowID is an opaque handle to a platform window.
// On X11 this wraps an X Window, on Win32 an HWND, on macOS an NSView*.
type WindowID uintptr

// DrawableID is an opaque handle to a drawable surface.
// A drawable can be either a window or a pixmap.
type DrawableID uintptr

// PixmapID is an opaque handle to an offscreen buffer.
type PixmapID uintptr

// GCID is an opaque handle to a graphics context.
type GCID uintptr

// CursorID is an opaque handle to a cursor.
type CursorID uintptr

// AtomID is an opaque interned string/property identifier.
// On X11 this is an Atom; other platforms map their own concepts.
type AtomID uint64

// KeySym is an abstract key symbol (keysym on X11, virtual key on Win32).
type KeySym uint64

// Timestamp is an event timestamp in platform-specific units.
type Timestamp uint64

// IsZeroGC returns true if the GC handle is the zero value.
func IsZeroGC(gc GCID) bool {
	return gc == 0
}

// ZeroGC returns the zero-value GC handle.
func ZeroGC() GCID {
	return 0
}

// WindowDrawable converts a WindowID to a DrawableID.
func WindowDrawable(w WindowID) DrawableID {
	return DrawableID(w)
}

// PixmapDrawable converts a PixmapID to a DrawableID.
func PixmapDrawable(p PixmapID) DrawableID {
	return DrawableID(p)
}
