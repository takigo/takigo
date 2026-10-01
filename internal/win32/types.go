//go:build windows

// Package win32 provides pure Go bindings to the Windows API via syscall.
// It is the Windows equivalent of internal/xlib/ (X11) and internal/cocoa/ (macOS).
package win32

// Handle types.
type (
	HWND      uintptr
	HDC       uintptr
	HBITMAP   uintptr
	HBRUSH    uintptr
	HPEN      uintptr
	HFONT     uintptr
	HCURSOR   uintptr
	HICON     uintptr
	HRGN      uintptr
	HINSTANCE uintptr
	HGLOBAL   uintptr
	HMENU     uintptr
	HGDIOBJ   uintptr
	HANDLE    uintptr
	WPARAM    uintptr
	LPARAM    uintptr
	LRESULT   uintptr
	ATOM      uint16
	COLORREF  uint32
	DWORD     uint32
)

// RGB creates a COLORREF from red, green, blue components (0-255).
func RGB(r, g, b byte) COLORREF {
	return COLORREF(uint32(r) | uint32(g)<<8 | uint32(b)<<16)
}

// GetRValue extracts the red component from a COLORREF.
func GetRValue(c COLORREF) byte { return byte(c) }

// GetGValue extracts the green component from a COLORREF.
func GetGValue(c COLORREF) byte { return byte(c >> 8) }

// GetBValue extracts the blue component from a COLORREF.
func GetBValue(c COLORREF) byte { return byte(c >> 16) }

// RECT structure.
type RECT struct {
	Left, Top, Right, Bottom int32
}

// POINT structure.
type POINT struct {
	X, Y int32
}

// SIZE structure.
type SIZE struct {
	CX, CY int32
}

// MSG structure for message loop.
type MSG struct {
	Hwnd    HWND
	Message uint32
	WParam  WPARAM
	LParam  LPARAM
	Time    uint32
	Pt      POINT
}

// WNDCLASSEXW structure.
type WNDCLASSEXW struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     HINSTANCE
	HIcon         HICON
	HCursor       HCURSOR
	HbrBackground HBRUSH
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       HICON
}

// PAINTSTRUCT structure.
type PAINTSTRUCT struct {
	Hdc         HDC
	FErase      int32
	RcPaint     RECT
	FRestore    int32
	FIncUpdate  int32
	RgbReserved [32]byte
}

// LOGFONTW structure for font creation.
type LOGFONTW struct {
	LfHeight         int32
	LfWidth          int32
	LfEscapement     int32
	LfOrientation    int32
	LfWeight         int32
	LfItalic         byte
	LfUnderline      byte
	LfStrikeOut      byte
	LfCharSet        byte
	LfOutPrecision   byte
	LfClipPrecision  byte
	LfQuality        byte
	LfPitchAndFamily byte
	LfFaceName       [32]uint16
}

// TEXTMETRICW structure.
type TEXTMETRICW struct {
	TmHeight           int32
	TmAscent           int32
	TmDescent          int32
	TmInternalLeading  int32
	TmExternalLeading  int32
	TmAveCharWidth     int32
	TmMaxCharWidth     int32
	TmWeight           int32
	TmOverhang         int32
	TmDigitizedAspectX int32
	TmDigitizedAspectY int32
	TmFirstChar        uint16
	TmLastChar         uint16
	TmDefaultChar      uint16
	TmBreakChar        uint16
	TmItalic           byte
	TmUnderlined       byte
	TmStruckOut        byte
	TmPitchAndFamily   byte
	TmCharSet          byte
}

// BITMAPINFOHEADER structure.
type BITMAPINFOHEADER struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

// BITMAPINFO structure (header + single color table entry for BI_RGB).
type BITMAPINFO struct {
	BmiHeader BITMAPINFOHEADER
	BmiColors [1]COLORREF
}

// LOGBRUSH structure.
type LOGBRUSH struct {
	LbStyle uint32
	LbColor COLORREF
	LbHatch uintptr
}

// LOGPEN structure.
type LOGPEN struct {
	LopnStyle uint32
	LopnWidth POINT
	LopnColor COLORREF
}

// Window styles.
const (
	WS_OVERLAPPED       = 0x00000000
	WS_POPUP            = 0x80000000
	WS_CHILD            = 0x40000000
	WS_MINIMIZE         = 0x20000000
	WS_VISIBLE          = 0x10000000
	WS_DISABLED         = 0x08000000
	WS_CLIPSIBLINGS     = 0x04000000
	WS_CLIPCHILDREN     = 0x02000000
	WS_MAXIMIZE         = 0x01000000
	WS_CAPTION          = 0x00C00000
	WS_BORDER           = 0x00800000
	WS_DLGFRAME         = 0x00400000
	WS_VSCROLL          = 0x00200000
	WS_HSCROLL          = 0x00100000
	WS_SYSMENU          = 0x00080000
	WS_THICKFRAME       = 0x00040000
	WS_MINIMIZEBOX      = 0x00020000
	WS_MAXIMIZEBOX      = 0x00010000
	WS_OVERLAPPEDWINDOW = WS_OVERLAPPED | WS_CAPTION | WS_SYSMENU | WS_THICKFRAME | WS_MINIMIZEBOX | WS_MAXIMIZEBOX
)

// Extended window styles.
const (
	WS_EX_TOPMOST          = 0x00000008
	WS_EX_TOOLWINDOW       = 0x00000080
	WS_EX_WINDOWEDGE       = 0x00000100
	WS_EX_CLIENTEDGE       = 0x00000200
	WS_EX_APPWINDOW        = 0x00040000
	WS_EX_LAYERED          = 0x00080000
	WS_EX_NOACTIVATE       = 0x08000000
	WS_EX_OVERLAPPEDWINDOW = WS_EX_WINDOWEDGE | WS_EX_CLIENTEDGE
)

// Window messages.
const (
	WM_NULL       = 0x0000
	WM_CREATE     = 0x0001
	WM_DESTROY    = 0x0002
	WM_MOVE       = 0x0003
	WM_SIZE       = 0x0005
	WM_ACTIVATE   = 0x0006
	WM_SETFOCUS   = 0x0007
	WM_KILLFOCUS  = 0x0008
	WM_ENABLE     = 0x000A
	WM_PAINT      = 0x000F
	WM_CLOSE      = 0x0010
	WM_QUIT       = 0x0012
	WM_ERASEBKGND = 0x0014
	WM_SHOWWINDOW = 0x0018
	WM_SETCURSOR  = 0x0020

	// HTCLIENT is WM_SETCURSOR's hit-test code for the client area.
	HTCLIENT                = 1
	WM_GETMINMAXINFO        = 0x0024
	WM_WINDOWPOSCHANGING    = 0x0046
	WM_WINDOWPOSCHANGED     = 0x0047
	WM_NCCALCSIZE           = 0x0083
	WM_NCHITTEST            = 0x0084
	WM_KEYDOWN              = 0x0100
	WM_KEYUP                = 0x0101
	WM_CHAR                 = 0x0102
	WM_DEADCHAR             = 0x0103
	WM_SYSKEYDOWN           = 0x0104
	WM_SYSKEYUP             = 0x0105
	WM_SYSCHAR              = 0x0106
	WM_UNICHAR              = 0x0109
	WM_IME_STARTCOMPOSITION = 0x010D
	WM_IME_ENDCOMPOSITION   = 0x010E
	WM_IME_COMPOSITION      = 0x010F
	WM_COMMAND              = 0x0111
	WM_TIMER                = 0x0113
	WM_MOUSEMOVE            = 0x0200
	WM_LBUTTONDOWN          = 0x0201
	WM_LBUTTONUP            = 0x0202
	WM_LBUTTONDBLCLK        = 0x0203
	WM_RBUTTONDOWN          = 0x0204
	WM_RBUTTONUP            = 0x0205
	WM_RBUTTONDBLCLK        = 0x0206
	WM_MBUTTONDOWN          = 0x0207
	WM_MBUTTONUP            = 0x0208
	WM_MBUTTONDBLCLK        = 0x0209
	WM_MOUSEWHEEL           = 0x020A
	WM_SETICON              = 0x0080
	WM_MOUSEHWHEEL          = 0x020E
	WM_ENTERSIZEMOVE        = 0x0231
	WM_EXITSIZEMOVE         = 0x0232
	WM_MOUSELEAVE           = 0x02A3
	WM_DPICHANGED           = 0x02E0
	WM_USER                 = 0x0400
)

// Class styles.
const (
	CS_VREDRAW = 0x0001
	CS_HREDRAW = 0x0002
	CS_DBLCLKS = 0x0008
	CS_OWNDC   = 0x0020
)

// ShowWindow commands.
const (
	SW_HIDE            = 0
	SW_SHOWNORMAL      = 1
	SW_SHOWMINIMIZED   = 2
	SW_SHOWMAXIMIZED   = 3
	SW_SHOWNOACTIVATE  = 4
	SW_SHOW            = 5
	SW_MINIMIZE        = 6
	SW_SHOWMINNOACTIVE = 7
	SW_SHOWNA          = 8
	SW_RESTORE         = 9
)

// SetWindowPos flags.
const (
	SWP_NOSIZE       = 0x0001
	SWP_NOMOVE       = 0x0002
	SWP_NOZORDER     = 0x0004
	SWP_NOACTIVATE   = 0x0010
	SWP_SHOWWINDOW   = 0x0040
	SWP_HIDEWINDOW   = 0x0080
	SWP_FRAMECHANGED = 0x0020
)

// SetWindowPos HWND constants.
var (
	HWND_TOP       = HWND(0)
	HWND_BOTTOM    = HWND(1)
	HWND_TOPMOST   = HWND(^uintptr(1)) // (HWND)-2
	HWND_NOTOPMOST = HWND(^uintptr(0)) // (HWND)-1 ... actually -2 for TOPMOST, let me fix
)

func init() {
	// Correct values: HWND_TOPMOST = -1, HWND_NOTOPMOST = -2
	HWND_TOPMOST = HWND(^uintptr(0))   // -1
	HWND_NOTOPMOST = HWND(^uintptr(1)) // -2
}

// PeekMessage flags.
const (
	PM_NOREMOVE = 0x0000
	PM_REMOVE   = 0x0001
)

// Raster operation codes for BitBlt.
const (
	SRCCOPY     = 0x00CC0020
	SRCPAINT    = 0x00EE0086
	SRCAND      = 0x008800C6
	SRCINVERT   = 0x00660046
	SRCERASE    = 0x00440328
	NOTSRCCOPY  = 0x00330008
	NOTSRCERASE = 0x001100A6
	MERGECOPY   = 0x00C000CA
	MERGEPAINT  = 0x00BB0226
	PATCOPY     = 0x00F00021
	PATPAINT    = 0x00FB0A09
	PATINVERT   = 0x005A0049
	DSTINVERT   = 0x00550009
	BLACKNESS   = 0x00000042
	WHITENESS   = 0x00FF0062
)

// SetROP2 modes.
const (
	R2_BLACK       = 1
	R2_NOTMERGEPEN = 2
	R2_MASKNOTPEN  = 3
	R2_NOTCOPYPEN  = 4
	R2_MASKPENNOT  = 5
	R2_NOT         = 6
	R2_XORPEN      = 7
	R2_NOTMASKPEN  = 8
	R2_MASKPEN     = 9
	R2_NOTXORPEN   = 10
	R2_NOP         = 11
	R2_MERGENOTPEN = 12
	R2_COPYPEN     = 13
	R2_MERGEPENNOT = 14
	R2_MERGEPEN    = 15
	R2_WHITE       = 16
)

// GDI object types / stock objects.
const (
	WHITE_BRUSH      = 0
	LTGRAY_BRUSH     = 1
	GRAY_BRUSH       = 2
	DKGRAY_BRUSH     = 3
	BLACK_BRUSH      = 4
	NULL_BRUSH       = 5
	WHITE_PEN        = 6
	BLACK_PEN        = 7
	NULL_PEN         = 8
	SYSTEM_FONT      = 13
	DEFAULT_GUI_FONT = 17
	DC_BRUSH         = 18
	DC_PEN           = 19
)

// Pen styles.
const (
	PS_SOLID         = 0
	PS_DASH          = 1
	PS_DOT           = 2
	PS_DASHDOT       = 3
	PS_DASHDOTDOT    = 4
	PS_NULL          = 5
	PS_INSIDEFRAME   = 6
	PS_USERSTYLE     = 7
	PS_GEOMETRIC     = 0x00010000
	PS_ENDCAP_ROUND  = 0x00000000
	PS_ENDCAP_SQUARE = 0x00000100
	PS_ENDCAP_FLAT   = 0x00000200
	PS_JOIN_ROUND    = 0x00000000
	PS_JOIN_BEVEL    = 0x00001000
	PS_JOIN_MITER    = 0x00002000
)

// Brush styles.
const (
	BS_SOLID   = 0
	BS_NULL    = 1
	BS_HOLLOW  = BS_NULL
	BS_HATCHED = 2
)

// Background modes.
const (
	TRANSPARENT = 1
	OPAQUE      = 2
)

// BI compression constants.
const (
	BI_RGB = 0
)

// DIB color constants.
const (
	DIB_RGB_COLORS = 0
)

// GetSystemMetrics indices.
const (
	SM_CXSCREEN     = 0
	SM_CYSCREEN     = 1
	SM_CXVSCROLL    = 2
	SM_CYHSCROLL    = 3
	SM_CYCAPTION    = 4
	SM_CXBORDER     = 5
	SM_CYBORDER     = 6
	SM_CXFIXEDFRAME = 7
	SM_CYFIXEDFRAME = 8
)

// GetDeviceCaps indices.
const (
	HORZSIZE   = 4  // horizontal size in mm
	VERTSIZE   = 6  // vertical size in mm
	HORZRES    = 8  // horizontal width in pixels
	VERTRES    = 10 // vertical height in pixels
	BITSPIXEL  = 12
	LOGPIXELSX = 88
	LOGPIXELSY = 90
)

// Virtual key codes.
const (
	VK_LBUTTON   = 0x01
	VK_RBUTTON   = 0x02
	VK_CANCEL    = 0x03
	VK_MBUTTON   = 0x04
	VK_BACK      = 0x08
	VK_TAB       = 0x09
	VK_CLEAR     = 0x0C
	VK_RETURN    = 0x0D
	VK_SHIFT     = 0x10
	VK_CONTROL   = 0x11
	VK_MENU      = 0x12 // Alt
	VK_PAUSE     = 0x13
	VK_CAPITAL   = 0x14
	VK_ESCAPE    = 0x1B
	VK_SPACE     = 0x20
	VK_PRIOR     = 0x21 // Page Up
	VK_NEXT      = 0x22 // Page Down
	VK_END       = 0x23
	VK_HOME      = 0x24
	VK_LEFT      = 0x25
	VK_UP        = 0x26
	VK_RIGHT     = 0x27
	VK_DOWN      = 0x28
	VK_SELECT    = 0x29
	VK_PRINT     = 0x2A
	VK_EXECUTE   = 0x2B
	VK_SNAPSHOT  = 0x2C
	VK_INSERT    = 0x2D
	VK_DELETE    = 0x2E
	VK_HELP      = 0x2F
	VK_LWIN      = 0x5B
	VK_RWIN      = 0x5C
	VK_APPS      = 0x5D
	VK_NUMPAD0   = 0x60
	VK_NUMPAD9   = 0x69
	VK_MULTIPLY  = 0x6A
	VK_ADD       = 0x6B
	VK_SEPARATOR = 0x6C
	VK_SUBTRACT  = 0x6D
	VK_DECIMAL   = 0x6E
	VK_DIVIDE    = 0x6F
	VK_F1        = 0x70
	VK_F2        = 0x71
	VK_F3        = 0x72
	VK_F4        = 0x73
	VK_F5        = 0x74
	VK_F6        = 0x75
	VK_F7        = 0x76
	VK_F8        = 0x77
	VK_F9        = 0x78
	VK_F10       = 0x79
	VK_F11       = 0x7A
	VK_F12       = 0x7B
	VK_F13       = 0x7C
	VK_F14       = 0x7D
	VK_F15       = 0x7E
	VK_F16       = 0x7F
	VK_F17       = 0x80
	VK_F18       = 0x81
	VK_F19       = 0x82
	VK_F20       = 0x83
	VK_F21       = 0x84
	VK_F22       = 0x85
	VK_F23       = 0x86
	VK_F24       = 0x87
	VK_NUMLOCK   = 0x90
	VK_SCROLL    = 0x91
	VK_LSHIFT    = 0xA0
	VK_RSHIFT    = 0xA1
	VK_LCONTROL  = 0xA2
	VK_RCONTROL  = 0xA3
	VK_LMENU     = 0xA4
	VK_RMENU     = 0xA5
)

// Cursor constants.
const (
	IDC_ARROW       = 32512
	IDC_IBEAM       = 32513
	IDC_WAIT        = 32514
	IDC_CROSS       = 32515
	IDC_UPARROW     = 32516
	IDC_SIZE        = 32640
	IDC_ICON        = 32641
	IDC_SIZENWSE    = 32642
	IDC_SIZENESW    = 32643
	IDC_SIZEWE      = 32644
	IDC_SIZENS      = 32645
	IDC_SIZEALL     = 32646
	IDC_NO          = 32648
	IDC_HAND        = 32649
	IDC_APPSTARTING = 32650
	IDC_HELP        = 32651
)

// Clipboard formats.
const (
	CF_TEXT        = 1
	CF_UNICODETEXT = 13
)

// GlobalAlloc flags.
const (
	GMEM_MOVEABLE = 0x0002
	GMEM_ZEROINIT = 0x0040
	GHND          = GMEM_MOVEABLE | GMEM_ZEROINIT
)

// SetWindowLongPtr indices.
const (
	GWL_STYLE     = -16
	GWL_EXSTYLE   = -20
	GWLP_USERDATA = -21
	GCLP_HCURSOR  = -12
)

// Mouse key constants (from wParam).
const (
	MK_LBUTTON = 0x0001
	MK_RBUTTON = 0x0002
	MK_SHIFT   = 0x0004
	MK_CONTROL = 0x0008
	MK_MBUTTON = 0x0010
)

// WHEEL_DELTA for mouse wheel messages.
const WHEEL_DELTA = 120

// GDI function return values.
const (
	GDI_ERROR   = 0xFFFFFFFF
	CLR_INVALID = 0xFFFFFFFF
	HGDI_ERROR  = HGDIOBJ(0xFFFFFFFF)
)

// CW_USEDEFAULT for CreateWindow.
const CW_USEDEFAULT = ^int32(0x7FFFFFFF) // 0x80000000

// Color constants for GetSysColor.
const (
	COLOR_SCROLLBAR       = 0
	COLOR_BACKGROUND      = 1
	COLOR_ACTIVECAPTION   = 2
	COLOR_INACTIVECAPTION = 3
	COLOR_MENU            = 4
	COLOR_WINDOW          = 5
	COLOR_WINDOWFRAME     = 6
	COLOR_MENUTEXT        = 7
	COLOR_WINDOWTEXT      = 8
	COLOR_CAPTIONTEXT     = 9
	COLOR_ACTIVEBORDER    = 10
	COLOR_INACTIVEBORDER  = 11
	COLOR_APPWORKSPACE    = 12
	COLOR_HIGHLIGHT       = 13
	COLOR_HIGHLIGHTTEXT   = 14
	COLOR_BTNFACE         = 15
	COLOR_BTNSHADOW       = 16
	COLOR_GRAYTEXT        = 17
	COLOR_BTNTEXT         = 18
	COLOR_BTNHIGHLIGHT    = 20
	COLOR_3DFACE          = COLOR_BTNFACE
	COLOR_3DSHADOW        = COLOR_BTNSHADOW
	COLOR_3DHIGHLIGHT     = COLOR_BTNHIGHLIGHT
)

// Font weight constants.
const (
	FW_DONTCARE = 0
	FW_NORMAL   = 400
	FW_BOLD     = 700
)

// Font charset constants.
const (
	DEFAULT_CHARSET = 1
)

// Font output precision.
const (
	OUT_DEFAULT_PRECIS = 0
	OUT_TT_PRECIS      = 4
)

// Font clip precision.
const (
	CLIP_DEFAULT_PRECIS = 0
)

// Font quality.
const (
	DEFAULT_QUALITY   = 0
	CLEARTYPE_QUALITY = 5
)

// Font pitch and family.
const (
	DEFAULT_PITCH  = 0
	FIXED_PITCH    = 1
	VARIABLE_PITCH = 2
	FF_DONTCARE    = 0
)

// TMPF flags in TEXTMETRICW.TmPitchAndFamily.
const (
	TMPF_FIXED_PITCH = 0x01 // Note: confusingly, if set the font is NOT fixed pitch
)

// Arc direction.
const (
	AD_COUNTERCLOCKWISE = 1
	AD_CLOCKWISE        = 2
)

// LOWORD/HIWORD helpers.
func LOWORD(l uintptr) uint16 { return uint16(l) }
func HIWORD(l uintptr) uint16 { return uint16(l >> 16) }

// GET_X_LPARAM / GET_Y_LPARAM for mouse messages.
func GET_X_LPARAM(lp LPARAM) int32 { return int32(int16(LOWORD(uintptr(lp)))) }
func GET_Y_LPARAM(lp LPARAM) int32 { return int32(int16(HIWORD(uintptr(lp)))) }

// GET_WHEEL_DELTA_WPARAM for mouse wheel messages.
func GET_WHEEL_DELTA_WPARAM(wp WPARAM) int16 { return int16(HIWORD(uintptr(wp))) }

// MAKEINTRESOURCE converts an integer resource identifier to a pointer.
// This matches the Win32 MAKEINTRESOURCE macro which casts an integer to a pointer.
func MAKEINTRESOURCE(id uint16) uintptr {
	return uintptr(id)
}
