// Package xlib provides low-level cgo bindings to the X11 Xlib library.
package xlib

/*
#cgo LDFLAGS: -lX11
#include <X11/Xlib.h>
#include <X11/Xutil.h>
#include <X11/Xatom.h>
*/
import "C"
import "unsafe"

// Type aliases for X11 types.
type (
	Window   C.Window
	Drawable C.Drawable
	Pixmap   C.Pixmap
	Colormap C.Colormap
	Atom     C.Atom
	GC       C.GC
	Cursor   C.Cursor
	KeySym   C.KeySym
	KeyCode  C.KeyCode
	Time     C.Time
	VisualID C.VisualID
)

// Display wraps an X11 Display connection.
type Display struct {
	ptr *C.Display
}

// Ptr returns the underlying C Display pointer for use in other cgo calls.
func (d *Display) Ptr() unsafe.Pointer {
	return unsafe.Pointer(d.ptr)
}

// Screen wraps an X11 Screen struct.
type Screen struct {
	ptr *C.Screen
}

// Visual wraps an X11 Visual struct.
type Visual struct {
	ptr *C.Visual
}

// Ptr returns the underlying C Visual pointer for use in other cgo calls.
func (v *Visual) Ptr() unsafe.Pointer {
	return unsafe.Pointer(v.ptr)
}

// XColor wraps an X11 XColor struct.
type XColor struct {
	Pixel uint64
	Red   uint16
	Green uint16
	Blue  uint16
	Flags byte
}

// Event mask constants.
const (
	NoEventMask             = C.NoEventMask
	KeyPressMask            = C.KeyPressMask
	KeyReleaseMask          = C.KeyReleaseMask
	ButtonPressMask         = C.ButtonPressMask
	ButtonReleaseMask       = C.ButtonReleaseMask
	EnterWindowMask         = C.EnterWindowMask
	LeaveWindowMask         = C.LeaveWindowMask
	PointerMotionMask       = C.PointerMotionMask
	ButtonMotionMask        = C.ButtonMotionMask
	ExposureMask            = C.ExposureMask
	StructureNotifyMask     = C.StructureNotifyMask
	SubstructureNotifyMask  = C.SubstructureNotifyMask
	SubstructureRedirectMask = C.SubstructureRedirectMask
	FocusChangeMask         = C.FocusChangeMask
	PropertyChangeMask      = C.PropertyChangeMask
	VisibilityChangeMask    = C.VisibilityChangeMask
)

// Event type constants.
const (
	KeyPress         = C.KeyPress
	KeyRelease       = C.KeyRelease
	ButtonPress      = C.ButtonPress
	ButtonRelease    = C.ButtonRelease
	MotionNotify     = C.MotionNotify
	EnterNotify      = C.EnterNotify
	LeaveNotify      = C.LeaveNotify
	FocusIn          = C.FocusIn
	FocusOut         = C.FocusOut
	Expose           = C.Expose
	DestroyNotify    = C.DestroyNotify
	UnmapNotify      = C.UnmapNotify
	MapNotify        = C.MapNotify
	MapRequest       = C.MapRequest
	ReparentNotify   = C.ReparentNotify
	ConfigureNotify  = C.ConfigureNotify
	ConfigureRequest = C.ConfigureRequest
	GravityNotify    = C.GravityNotify
	ResizeRequest    = C.ResizeRequest
	CirculateNotify  = C.CirculateNotify
	PropertyNotify   = C.PropertyNotify
	SelectionClear   = C.SelectionClear
	SelectionRequest = C.SelectionRequest
	SelectionNotify  = C.SelectionNotify
	ColormapNotify   = C.ColormapNotify
	ClientMessage    = C.ClientMessage
	MappingNotify    = C.MappingNotify
)

// Window attribute constants.
const (
	CWBackPixel       = C.CWBackPixel
	CWBorderPixel     = C.CWBorderPixel
	CWBitGravity      = C.CWBitGravity
	CWEventMask       = C.CWEventMask
	CWColormap        = C.CWColormap
	CWOverrideRedirect = C.CWOverrideRedirect

	InputOutput = C.InputOutput
	InputOnly   = C.InputOnly

	CopyFromParent = 0
)

// GC value mask constants.
const (
	GCForeground = C.GCForeground
	GCBackground = C.GCBackground
	GCLineWidth  = C.GCLineWidth
	GCLineStyle  = C.GCLineStyle
	GCFont       = C.GCFont
	GCFunction   = C.GCFunction
)

// GC function constants.
const (
	GXcopy = C.GXcopy
)

// Atom predefined constants.
var (
	XA_WM_NAME          = Atom(C.XA_WM_NAME)
	XA_STRING           = Atom(C.XA_STRING)
	XA_WM_NORMAL_HINTS  = Atom(C.XA_WM_NORMAL_HINTS)
)

// None is the X11 None constant.
const None = C.None
