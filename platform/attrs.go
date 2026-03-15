package platform

// WindowAttrs holds attributes for window creation.
type WindowAttrs struct {
	BackgroundPixel  uint64
	BorderPixel      uint64
	BitGravity       int
	EventMask        int64
	OverrideRedirect bool
}

// GCValues describes the desired attributes for a graphics context.
type GCValues struct {
	Foreground uint64
	Background uint64
	LineWidth  int
	Function   int
}

// Point represents an x,y coordinate.
type Point struct {
	X, Y int16
}

// SizeHints holds window sizing hints.
type SizeHints struct {
	Flags                int64
	X, Y                 int
	Width, Height        int
	MinWidth, MinHeight  int
	MaxWidth, MaxHeight  int
	WidthInc, HeightInc  int
	WinGravity           int
}

// WMHints holds window manager hints.
type WMHints struct {
	Flags        int64
	Input        bool
	InitialState int
}

// Event mask constants (matching X11 values).
const (
	NoEventMask              = 0
	KeyPressMask             = 1 << 0
	KeyReleaseMask           = 1 << 1
	ButtonPressMask          = 1 << 2
	ButtonReleaseMask        = 1 << 3
	EnterWindowMask          = 1 << 4
	LeaveWindowMask          = 1 << 5
	PointerMotionMask        = 1 << 6
	ButtonMotionMask         = 1 << 13
	ExposureMask             = 1 << 15
	StructureNotifyMask      = 1 << 17
	SubstructureNotifyMask   = 1 << 19
	SubstructureRedirectMask = 1 << 20
	FocusChangeMask          = 1 << 21
	PropertyChangeMask       = 1 << 22
	VisibilityChangeMask     = 1 << 16
)

// Window class constants.
const (
	InputOutput = 1
	InputOnly   = 2
)

// CW attribute mask constants.
const (
	CWBackPixel        = 1 << 1
	CWBorderPixel      = 1 << 3
	CWBitGravity       = 1 << 4
	CWEventMask        = 1 << 11
	CWOverrideRedirect = 1 << 9
)

// GC value mask constants.
const (
	GCForeground = 1 << 2
	GCBackground = 1 << 3
	GCLineWidth  = 1 << 4
	GCLineStyle  = 1 << 5
	GCFont       = 1 << 14
	GCFunction   = 1 << 0
)

// GC function constants.
const (
	GXcopy = 3
)

// Line style constants.
const (
	LineSolid      = 0
	LineOnOffDash  = 1
	LineDoubleDash = 2
)

// Cap style constants.
const (
	CapNotLast    = 0
	CapButt       = 1
	CapRound      = 2
	CapProjecting = 3
)

// Join style constants.
const (
	JoinMiter = 0
	JoinRound = 1
	JoinBevel = 2
)

// Fill style constants.
const (
	FillSolid          = 0
	FillTiled          = 1
	FillStippled       = 2
	FillOpaqueStippled = 3
)

// Polygon shape constants.
const (
	PolygonComplex   = 0
	PolygonNonconvex = 1
	PolygonConvex    = 2
)

// Coordinate mode constants.
const (
	CoordModeOrigin   = 0
	CoordModePrevious = 1
)

// SizeHints flags.
const (
	USPosition  = 1 << 0
	USSize      = 1 << 1
	PPosition   = 1 << 2
	PSize       = 1 << 3
	PMinSize    = 1 << 4
	PMaxSize    = 1 << 5
	PResizeInc  = 1 << 6
	PWinGravity = 1 << 9
)

// WM state constants.
const (
	WithdrawnState = 0
	NormalState    = 1
	IconicState    = 3
)

// WM hints flags.
const (
	InputHint = 1 << 0
	StateHint = 1 << 1
)

// Gravity constants.
const (
	NorthWestGravity = 1
	NorthGravity     = 2
	NorthEastGravity = 3
	WestGravity      = 4
	CenterGravity    = 5
	EastGravity      = 6
	SouthWestGravity = 7
	SouthGravity     = 8
	SouthEastGravity = 9
)

// Focus revert-to constants.
const (
	RevertToNone        = 0
	RevertToPointerRoot = 1
	RevertToParent      = 2
)

// CurrentTime for timestamps.
const CurrentTime = Timestamp(0)

// Grab mode constants.
const (
	GrabModeSync  = 0
	GrabModeAsync = 1
)

// Grab result constants.
const (
	GrabSuccess     = 0
	AlreadyGrabbed  = 1
	GrabInvalidTime = 2
	GrabNotViewable = 3
	GrabFrozen      = 4
)

// Property mode constants.
const PropModeReplace = 0

// None is the zero/null value for handles.
const None = 0

// CopyFromParent for window creation.
const CopyFromParent = 0

// Modifier masks (matching X11 values).
const (
	ShiftMask   = uint(1 << 0)
	LockMask    = uint(1 << 1)
	ControlMask = uint(1 << 2)
	Mod1Mask    = uint(1 << 3) // typically Alt
	Mod2Mask    = uint(1 << 4)
	Mod3Mask    = uint(1 << 5)
	Mod4Mask    = uint(1 << 6) // typically Super/Win
	Mod5Mask    = uint(1 << 7)
	Button1Mask = uint(1 << 8)
	Button2Mask = uint(1 << 9)
	Button3Mask = uint(1 << 10)
)

// Common key symbols (keysyms).
const (
	XK_BackSpace    = KeySym(0xff08)
	XK_Tab          = KeySym(0xff09)
	XK_ISO_Left_Tab = KeySym(0xfe20)
	XK_Return       = KeySym(0xff0d)
	XK_Escape       = KeySym(0xff1b)
	XK_Insert       = KeySym(0xff63)
	XK_Delete       = KeySym(0xffff)
	XK_Home         = KeySym(0xff50)
	XK_Left         = KeySym(0xff51)
	XK_Up           = KeySym(0xff52)
	XK_Right        = KeySym(0xff53)
	XK_Down         = KeySym(0xff54)
	XK_Prior        = KeySym(0xff55) // PageUp
	XK_Next         = KeySym(0xff56) // PageDown
	XK_End          = KeySym(0xff57)
	XK_space        = KeySym(0x0020)

	// Latin letter keysyms.
	XK_a = KeySym(0x0061)
	XK_b = KeySym(0x0062)
	XK_c = KeySym(0x0063)
	XK_d = KeySym(0x0064)
	XK_e = KeySym(0x0065)
	XK_f = KeySym(0x0066)
	XK_k = KeySym(0x006b)
	XK_n = KeySym(0x006e)
	XK_p = KeySym(0x0070)
	XK_q = KeySym(0x0071)
	XK_v = KeySym(0x0076)
	XK_w = KeySym(0x0077)
	XK_x = KeySym(0x0078)
	XK_y = KeySym(0x0079)
	XK_z = KeySym(0x007a)
)
