//go:build linux || freebsd || openbsd || netbsd

package xlib

/*
#include <stdlib.h>
#include <X11/Xlib.h>
*/
import "C"
import "unsafe"

// InternAtom returns the atom for the given name.
// If onlyIfExists is true, it returns None if the atom doesn't already exist.
func (d *Display) InternAtom(name string, onlyIfExists bool) Atom {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	var exists C.int
	if onlyIfExists {
		exists = 1
	}
	return Atom(C.XInternAtom(d.ptr, cname, exists))
}

// GetAtomName returns the name of an atom.
func (d *Display) GetAtomName(atom Atom) string {
	cname := C.XGetAtomName(d.ptr, C.Atom(atom))
	if cname == nil {
		return ""
	}
	name := C.GoString(cname)
	C.XFree(unsafe.Pointer(cname))
	return name
}
