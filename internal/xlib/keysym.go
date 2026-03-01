package xlib

/*
#include <X11/Xlib.h>
#include <X11/keysym.h>
*/
import "C"

// Common key symbols.
const (
	XK_q         = KeySym(C.XK_q)
	XK_Escape    = KeySym(C.XK_Escape)
	XK_Return    = KeySym(C.XK_Return)
	XK_BackSpace = KeySym(C.XK_BackSpace)
	XK_Tab       = KeySym(C.XK_Tab)
	XK_space     = KeySym(C.XK_space)
	XK_Left      = KeySym(C.XK_Left)
	XK_Right     = KeySym(C.XK_Right)
	XK_Up        = KeySym(C.XK_Up)
	XK_Down      = KeySym(C.XK_Down)
	XK_Home      = KeySym(C.XK_Home)
	XK_End       = KeySym(C.XK_End)
	XK_Delete    = KeySym(C.XK_Delete)
)

// Modifier masks.
const (
	ShiftMask   = C.ShiftMask
	LockMask    = C.LockMask
	ControlMask = C.ControlMask
	Mod1Mask    = C.Mod1Mask // typically Alt
	Mod2Mask    = C.Mod2Mask
	Mod3Mask    = C.Mod3Mask
	Mod4Mask    = C.Mod4Mask // typically Super/Win
	Mod5Mask    = C.Mod5Mask
)
