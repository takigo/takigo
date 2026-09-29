//go:build linux || freebsd || openbsd || netbsd

package xlib

/*
#include <stdlib.h>
#include <X11/Xlib.h>
#include <X11/Xutil.h>
#include <X11/Xatom.h>
#include <string.h>

// Helper to send a ClientMessage event.
static void send_client_message(Display *dpy, Window w, Window target, Atom msgType,
	long d0, long d1, long d2, long d3, long d4) {
	XEvent ev;
	memset(&ev, 0, sizeof(ev));
	ev.xclient.type = ClientMessage;
	ev.xclient.window = w;
	ev.xclient.message_type = msgType;
	ev.xclient.format = 32;
	ev.xclient.data.l[0] = d0;
	ev.xclient.data.l[1] = d1;
	ev.xclient.data.l[2] = d2;
	ev.xclient.data.l[3] = d3;
	ev.xclient.data.l[4] = d4;
	// Client messages to the root window are for the window manager, which
	// selects SubstructureRedirect there (EWMH); elsewhere the owner gets it.
	long mask = NoEventMask;
	if (target == DefaultRootWindow(dpy)) {
		mask = SubstructureRedirectMask | SubstructureNotifyMask;
	}
	XSendEvent(dpy, target, False, mask, &ev);
}
*/
import "C"
import (
	"encoding/binary"
	"unsafe"
)

// SizeHints flags.
const (
	USPosition  = C.USPosition
	USSize      = C.USSize
	PPosition   = C.PPosition
	PSize       = C.PSize
	PMinSize    = C.PMinSize
	PMaxSize    = C.PMaxSize
	PResizeInc  = C.PResizeInc
	PWinGravity = C.PWinGravity
)

// WM state constants.
const (
	WithdrawnState = C.WithdrawnState
	NormalState    = C.NormalState
	IconicState    = C.IconicState
)

// Gravity constants.
const (
	NorthWestGravity = C.NorthWestGravity
	NorthGravity     = C.NorthGravity
	NorthEastGravity = C.NorthEastGravity
	WestGravity      = C.WestGravity
	CenterGravity    = C.CenterGravity
	EastGravity      = C.EastGravity
	SouthWestGravity = C.SouthWestGravity
	SouthGravity     = C.SouthGravity
	SouthEastGravity = C.SouthEastGravity
)

// SizeHints holds XSizeHints data.
type SizeHints struct {
	Flags               int64
	X, Y                int
	Width, Height       int
	MinWidth, MinHeight int
	MaxWidth, MaxHeight int
	WidthInc, HeightInc int
	WinGravity          int
}

// SetWMNormalHints sets the WM_NORMAL_HINTS property.
func (d *Display) SetWMNormalHints(w Window, hints *SizeHints) {
	var sh C.XSizeHints
	sh.flags = C.long(hints.Flags)
	sh.x = C.int(hints.X)
	sh.y = C.int(hints.Y)
	sh.width = C.int(hints.Width)
	sh.height = C.int(hints.Height)
	sh.min_width = C.int(hints.MinWidth)
	sh.min_height = C.int(hints.MinHeight)
	sh.max_width = C.int(hints.MaxWidth)
	sh.max_height = C.int(hints.MaxHeight)
	sh.width_inc = C.int(hints.WidthInc)
	sh.height_inc = C.int(hints.HeightInc)
	sh.win_gravity = C.int(hints.WinGravity)
	C.XSetWMNormalHints(d.ptr, C.Window(w), &sh)
}

// WMHints holds XWMHints data.
type WMHints struct {
	Flags        int64
	Input        bool
	InitialState int
}

// WMHints flag constants.
const (
	InputHint = C.InputHint
	StateHint = C.StateHint
)

// SetWMHints sets the WM_HINTS property.
func (d *Display) SetWMHints(w Window, hints *WMHints) {
	var h C.XWMHints
	h.flags = C.long(hints.Flags)
	if hints.Input {
		h.input = C.True
	} else {
		h.input = C.False
	}
	h.initial_state = C.int(hints.InitialState)
	C.XSetWMHints(d.ptr, C.Window(w), &h)
}

// SetClassHint sets the WM_CLASS property.
func (d *Display) SetClassHint(w Window, name, class string) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	cclass := C.CString(class)
	defer C.free(unsafe.Pointer(cclass))

	var ch C.XClassHint
	ch.res_name = cname
	ch.res_class = cclass
	C.XSetClassHint(d.ptr, C.Window(w), &ch)
}

// SetTransientForHint sets WM_TRANSIENT_FOR.
func (d *Display) SetTransientForHint(w Window, propWindow Window) {
	C.XSetTransientForHint(d.ptr, C.Window(w), C.Window(propWindow))
}

// DeleteProperty deletes a window property.
func (d *Display) DeleteProperty(w Window, prop Atom) {
	C.XDeleteProperty(d.ptr, C.Window(w), C.Atom(prop))
}

// ChangeProperty sets a window property. Format-32 data is packed as
// 4 bytes per item in native byte order; Xlib wants C longs, so it is
// widened here.
func (d *Display) ChangeProperty(w Window, prop, propType Atom, format int, mode int, data []byte, nelements int) {
	var dataPtr *C.uchar
	if format == 32 {
		nelements = min(nelements, len(data)/4)
		longs := make([]C.ulong, max(nelements, 1))
		for i := range nelements {
			longs[i] = C.ulong(binary.NativeEndian.Uint32(data[i*4:]))
		}
		dataPtr = (*C.uchar)(unsafe.Pointer(&longs[0]))
		C.XChangeProperty(d.ptr, C.Window(w), C.Atom(prop), C.Atom(propType),
			C.int(format), C.int(mode), dataPtr, C.int(nelements))
		return
	}
	if len(data) > 0 {
		dataPtr = (*C.uchar)(unsafe.Pointer(&data[0]))
	}
	C.XChangeProperty(d.ptr, C.Window(w), C.Atom(prop), C.Atom(propType),
		C.int(format), C.int(mode), dataPtr, C.int(nelements))
}

// ChangePropertyAtoms sets an atom-type window property.
func (d *Display) ChangePropertyAtoms(w Window, prop Atom, atoms []Atom) {
	if len(atoms) == 0 {
		return
	}
	longs := make([]C.ulong, len(atoms))
	for i, a := range atoms {
		longs[i] = C.ulong(a)
	}
	C.XChangeProperty(d.ptr, C.Window(w), C.Atom(prop), C.XA_ATOM,
		32, C.PropModeReplace,
		(*C.uchar)(unsafe.Pointer(&longs[0])), C.int(len(atoms)))
}

// PropertyReplace mode.
const PropModeReplace = C.PropModeReplace

// IconifyWindow sends a request to the WM to iconify the window.
func (d *Display) IconifyWindow(w Window, screen int) {
	C.XIconifyWindow(d.ptr, C.Window(w), C.int(screen))
}

// WithdrawWindow withdraws a window.
func (d *Display) WithdrawWindow(w Window, screen int) {
	C.XWithdrawWindow(d.ptr, C.Window(w), C.int(screen))
}

// SetInputFocus sets the input focus to a window.
func (d *Display) SetInputFocus(w Window, revertTo int, time Time) {
	C.XSetInputFocus(d.ptr, C.Window(w), C.int(revertTo), C.Time(time))
}

// GetInputFocus returns the current input focus window.
func (d *Display) GetInputFocus() (Window, int) {
	var focus C.Window
	var revert C.int
	C.XGetInputFocus(d.ptr, &focus, &revert)
	return Window(focus), int(revert)
}

// Revert-to constants for SetInputFocus.
const (
	RevertToNone        = C.RevertToNone
	RevertToPointerRoot = C.RevertToPointerRoot
	RevertToParent      = C.RevertToParent
)

// CurrentTime for X timestamps.
const CurrentTime = C.CurrentTime

// GrabPointer grabs the pointer.
func (d *Display) GrabPointer(grabWindow Window, ownerEvents bool, eventMask uint,
	pointerMode, keyboardMode int, confineTo Window, cursor Cursor, time Time) int {
	var owner C.int
	if ownerEvents {
		owner = C.True
	}
	return int(C.XGrabPointer(d.ptr, C.Window(grabWindow), owner,
		C.uint(eventMask), C.int(pointerMode), C.int(keyboardMode),
		C.Window(confineTo), C.Cursor(cursor), C.Time(time)))
}

// UngrabPointer releases the pointer grab.
func (d *Display) UngrabPointer(time Time) {
	C.XUngrabPointer(d.ptr, C.Time(time))
}

// GrabKeyboard grabs the keyboard.
func (d *Display) GrabKeyboard(grabWindow Window, ownerEvents bool,
	pointerMode, keyboardMode int, time Time) int {
	var owner C.int
	if ownerEvents {
		owner = C.True
	}
	return int(C.XGrabKeyboard(d.ptr, C.Window(grabWindow), owner,
		C.int(pointerMode), C.int(keyboardMode), C.Time(time)))
}

// UngrabKeyboard releases the keyboard grab.
func (d *Display) UngrabKeyboard(time Time) {
	C.XUngrabKeyboard(d.ptr, C.Time(time))
}

// Grab mode constants.
const (
	GrabModeSync  = C.GrabModeSync
	GrabModeAsync = C.GrabModeAsync
)

// Grab result constants.
const (
	GrabSuccess     = C.GrabSuccess
	AlreadyGrabbed  = C.AlreadyGrabbed
	GrabInvalidTime = C.GrabInvalidTime
	GrabNotViewable = C.GrabNotViewable
	GrabFrozen      = C.GrabFrozen
)

// SendEvent sends an event to a window.
func (d *Display) SendEvent(w Window, propagate bool, eventMask int64, ev *RawEvent) {
	var prop C.int
	if propagate {
		prop = C.True
	}
	C.XSendEvent(d.ptr, C.Window(w), prop, C.long(eventMask), &ev.ev)
}

// Predefined atoms.
var (
	XA_ATOM     = Atom(C.XA_ATOM)
	XA_CARDINAL = Atom(C.XA_CARDINAL)
	XA_WINDOW   = Atom(C.XA_WINDOW)
)

// SendClientMessage sends a ClientMessage event to a target window.
func (d *Display) SendClientMessage(w, target Window, msgType Atom, d0, d1, d2, d3, d4 int64) {
	C.send_client_message(d.ptr, C.Window(w), C.Window(target), C.Atom(msgType),
		C.long(d0), C.long(d1), C.long(d2), C.long(d3), C.long(d4))
}

// XSetIconName sets the icon name for a window.
func (d *Display) SetIconName(w Window, name string) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	C.XSetIconName(d.ptr, C.Window(w), cname)
}
