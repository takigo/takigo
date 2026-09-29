//go:build linux || freebsd || openbsd || netbsd

package xlib

/*
#include <X11/Xlib.h>
#include <X11/Xatom.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"encoding/binary"
	"unsafe"
)

// Predefined atoms for selections.
const (
	XA_PRIMARY   = Atom(C.XA_PRIMARY)
	XA_SECONDARY = Atom(C.XA_SECONDARY)
)

// SetSelectionOwner sets the owner of a selection.
func (d *Display) SetSelectionOwner(selection Atom, owner Window, time Time) {
	C.XSetSelectionOwner(d.ptr, C.Atom(selection), C.Window(owner), C.Time(time))
}

// GetSelectionOwner returns the current owner of a selection.
func (d *Display) GetSelectionOwner(selection Atom) Window {
	return Window(C.XGetSelectionOwner(d.ptr, C.Atom(selection)))
}

// ConvertSelection requests conversion of a selection.
func (d *Display) ConvertSelection(selection, target, property Atom, requestor Window, time Time) {
	C.XConvertSelection(d.ptr, C.Atom(selection), C.Atom(target),
		C.Atom(property), C.Window(requestor), C.Time(time))
}

// GetWindowProperty reads a window property. Returns the data as bytes,
// the actual type, and the actual format.
func (d *Display) GetWindowProperty(w Window, property Atom, offset, length int64, delete bool) ([]byte, Atom, int) {
	var actualType C.Atom
	var actualFormat C.int
	var nItems, bytesAfter C.ulong
	var data *C.uchar
	var del C.int
	if delete {
		del = 1
	}

	C.XGetWindowProperty(d.ptr, C.Window(w), C.Atom(property),
		C.long(offset), C.long(length), del, C.AnyPropertyType,
		&actualType, &actualFormat, &nItems, &bytesAfter, &data)

	if data == nil {
		return nil, Atom(actualType), int(actualFormat)
	}
	defer C.XFree(unsafe.Pointer(data))

	n := int(nItems)
	var result []byte
	switch int(actualFormat) {
	case 32:
		// Xlib stores format-32 items as C longs; return them packed
		// as 4 bytes per item in native byte order.
		longs := unsafe.Slice((*C.ulong)(unsafe.Pointer(data)), n)
		result = make([]byte, n*4)
		for i, v := range longs {
			binary.NativeEndian.PutUint32(result[i*4:], uint32(v))
		}
	default:
		size := n
		if actualFormat == 16 {
			size *= 2
		}
		result = make([]byte, size)
		if size > 0 {
			C.memcpy(unsafe.Pointer(&result[0]), unsafe.Pointer(data), C.size_t(size))
		}
	}
	return result, Atom(actualType), int(actualFormat)
}

// SendSelectionNotify sends a SelectionNotify event to a requestor.
func (d *Display) SendSelectionNotify(requestor Window, selection, target, property Atom, time Time) {
	var ev C.XEvent
	notify := (*C.XSelectionEvent)(unsafe.Pointer(&ev))
	notify._type = C.SelectionNotify
	notify.requestor = C.Window(requestor)
	notify.selection = C.Atom(selection)
	notify.target = C.Atom(target)
	notify.property = C.Atom(property)
	notify.time = C.Time(time)
	C.XSendEvent(d.ptr, C.Window(requestor), 0, 0, &ev)
}

// ChangePropertyString sets a string property on a window.
func (d *Display) ChangePropertyString(w Window, property, typ Atom, data string) {
	cdata := C.CString(data)
	defer C.free(unsafe.Pointer(cdata))
	C.XChangeProperty(d.ptr, C.Window(w), C.Atom(property), C.Atom(typ),
		8, C.PropModeReplace, (*C.uchar)(unsafe.Pointer(cdata)), C.int(len(data)))
}
