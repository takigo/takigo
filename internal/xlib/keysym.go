package xlib

/*
#include <X11/Xlib.h>
#include <X11/keysym.h>
*/
import "C"

// Common key symbols.
const (
	XK_a         = KeySym(0x0061)
	XK_c         = KeySym(0x0063)
	XK_d         = KeySym(0x0064)
	XK_k         = KeySym(0x006b)
	XK_q         = KeySym(0x0071)
	XK_v         = KeySym(0x0076)
	XK_x         = KeySym(0x0078)
	XK_Escape    = KeySym(C.XK_Escape)
	XK_Return    = KeySym(C.XK_Return)
	XK_BackSpace = KeySym(C.XK_BackSpace)
	XK_Tab          = KeySym(C.XK_Tab)
	XK_ISO_Left_Tab = KeySym(0xfe20)
	XK_space     = KeySym(C.XK_space)
	XK_Left      = KeySym(C.XK_Left)
	XK_Right     = KeySym(C.XK_Right)
	XK_Up        = KeySym(C.XK_Up)
	XK_Down      = KeySym(C.XK_Down)
	XK_Home      = KeySym(C.XK_Home)
	XK_End       = KeySym(C.XK_End)
	XK_Delete    = KeySym(C.XK_Delete)
	XK_Prior     = KeySym(0xff55) // PageUp
	XK_Next      = KeySym(0xff56) // PageDown
)

// Modifier masks.
const (
	ShiftMask   = uint(C.ShiftMask)
	LockMask    = uint(C.LockMask)
	ControlMask = uint(C.ControlMask)
	Mod1Mask    = uint(C.Mod1Mask) // typically Alt
	Mod2Mask    = uint(C.Mod2Mask)
	Mod3Mask    = uint(C.Mod3Mask)
	Mod4Mask    = uint(C.Mod4Mask) // typically Super/Win
	Mod5Mask    = uint(C.Mod5Mask)
	Button1Mask = uint(1 << 8)
	Button2Mask = uint(1 << 9)
	Button3Mask = uint(1 << 10)
)
