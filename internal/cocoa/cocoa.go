//go:build darwin

// Package cocoa provides low-level cgo bindings to macOS Cocoa/AppKit/CoreGraphics
// for the takigo GUI toolkit. It is the macOS equivalent of internal/xlib/.
package cocoa

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework CoreText -framework QuartzCore -framework CoreGraphics
#include "cocoa.h"
#include <stdlib.h>
*/
import "C"
import (
	"runtime"
	"unsafe"
)

func init() {
	// Cocoa requires the main thread. Lock the OS thread so that
	// Go's main goroutine is always on thread 0.
	runtime.LockOSThread()
}

// Opaque handle types — defined as uintptr for easy conversion with platform types.
type (
	Window   uintptr
	Drawable uintptr
	Pixmap   uintptr
	GC       uintptr
	Cursor   uintptr
	Atom     uint64
	Time     uint64
	FontID   uintptr
)

// RawEvent wraps the C-level event structure.
type RawEvent struct {
	ev C.CocoaRawEvent
}

func (e *RawEvent) Type() int           { return int(e.ev._type) }
func (e *RawEvent) Window() Window      { return Window(uintptr(e.ev.window)) }
func (e *RawEvent) State() uint         { return uint(e.ev.state) }
func (e *RawEvent) KeyCode() uint       { return uint(e.ev.keycode) }
func (e *RawEvent) KeySym() uint64      { return uint64(e.ev.keysym) }
func (e *RawEvent) Str() string         { return C.GoString(&e.ev.str[0]) }
func (e *RawEvent) X() int              { return int(e.ev.x) }
func (e *RawEvent) Y() int              { return int(e.ev.y) }
func (e *RawEvent) RootX() int          { return int(e.ev.rootX) }
func (e *RawEvent) RootY() int          { return int(e.ev.rootY) }
func (e *RawEvent) Button() uint        { return uint(e.ev.button) }
func (e *RawEvent) Width() int          { return int(e.ev.width) }
func (e *RawEvent) Height() int         { return int(e.ev.height) }
func (e *RawEvent) ExposeX() int        { return int(e.ev.exposeX) }
func (e *RawEvent) ExposeY() int        { return int(e.ev.exposeY) }
func (e *RawEvent) ExposeWidth() int    { return int(e.ev.exposeWidth) }
func (e *RawEvent) ExposeHeight() int   { return int(e.ev.exposeHeight) }
func (e *RawEvent) ExposeCount() int    { return int(e.ev.exposeCount) }
func (e *RawEvent) MessageType() uint64 { return uint64(e.ev.messageType) }
func (e *RawEvent) FocusMode() int      { return int(e.ev.focusMode) }
func (e *RawEvent) FocusDetail() int    { return int(e.ev.focusDetail) }
func (e *RawEvent) Time_() uint64       { return uint64(e.ev.time) }

func (e *RawEvent) MessageData() [5]int64 {
	var d [5]int64
	for i := 0; i < 5; i++ {
		d[i] = int64(e.ev.messageData[i])
	}
	return d
}

// ---- Application lifecycle ----

func Init()            { C.CocoaInit() }
func Run()             { C.CocoaRun() }
func Stop()            { C.CocoaStop() }
func Flush()           { C.CocoaFlush() }
func StartEventPump()  { C.CocoaStartEventPump() }
func PumpEvents()      { C.CocoaPumpEvents() }

// ---- Screen info ----

func ScreenWidth() int      { return int(C.CocoaScreenWidth()) }
func ScreenHeight() int     { return int(C.CocoaScreenHeight()) }
func ScreenWidthMM() int    { return int(C.CocoaScreenWidthMM()) }
func ScreenHeightMM() int   { return int(C.CocoaScreenHeightMM()) }
func ScreenDepth() int      { return int(C.CocoaScreenDepth()) }
func BackingScale() float64 { return float64(C.CocoaScreenBackingScale()) }

// ---- Window management ----

func CreateWindow(parent Window, x, y int, width, height, borderWidth uint,
	bgPixel uint64, eventMask int64, overrideRedirect bool) Window {
	return Window(C.CocoaCreateWindow(C.CocoaWindowID(parent),
		C.int(x), C.int(y), C.uint(width), C.uint(height), C.uint(borderWidth),
		C.uint64_t(bgPixel), C.int64_t(eventMask), C.bool(overrideRedirect)))
}

func CreateSimpleWindow(parent Window, x, y int, width, height, borderWidth uint,
	border, background uint64) Window {
	return Window(C.CocoaCreateSimpleWindow(C.CocoaWindowID(parent),
		C.int(x), C.int(y), C.uint(width), C.uint(height), C.uint(borderWidth),
		C.uint64_t(border), C.uint64_t(background)))
}

func DestroyWindow(w Window)    { C.CocoaDestroyWindow(C.CocoaWindowID(w)) }
func MapWindow(w Window)        { C.CocoaMapWindow(C.CocoaWindowID(w)) }
func MapRaised(w Window)        { C.CocoaMapRaised(C.CocoaWindowID(w)) }
func UnmapWindow(w Window)      { C.CocoaUnmapWindow(C.CocoaWindowID(w)) }
func RaiseWindow(w Window)      { C.CocoaRaiseWindow(C.CocoaWindowID(w)) }
func LowerWindow(w Window)      { C.CocoaLowerWindow(C.CocoaWindowID(w)) }
func MoveWindow(w Window, x, y int) {
	C.CocoaMoveWindow(C.CocoaWindowID(w), C.int(x), C.int(y))
}
func ResizeWindow(w Window, width, height uint) {
	C.CocoaResizeWindow(C.CocoaWindowID(w), C.uint(width), C.uint(height))
}
func MoveResizeWindow(w Window, x, y int, width, height uint) {
	C.CocoaMoveResizeWindow(C.CocoaWindowID(w), C.int(x), C.int(y), C.uint(width), C.uint(height))
}
func SelectInput(w Window, eventMask int64) {
	C.CocoaSelectInput(C.CocoaWindowID(w), C.int64_t(eventMask))
}
func StoreName(w Window, name string) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	C.CocoaStoreName(C.CocoaWindowID(w), cname)
}
func TranslateCoordinates(src, dst Window, srcX, srcY int) (dstX, dstY int) {
	var dx, dy C.int
	C.CocoaTranslateCoordinates(C.CocoaWindowID(src), C.CocoaWindowID(dst),
		C.int(srcX), C.int(srcY), &dx, &dy)
	return int(dx), int(dy)
}
func SetWindowBackground(w Window, pixel uint64) {
	C.CocoaSetWindowBackground(C.CocoaWindowID(w), C.uint64_t(pixel))
}
func ClearWindow(w Window)   { C.CocoaClearWindow(C.CocoaWindowID(w)) }
func SetInputFocus(w Window) { C.CocoaSetInputFocus(C.CocoaWindowID(w)) }

func ClearArea(w Window, x, y int, width, height uint, exposures bool) {
	C.CocoaClearArea(C.CocoaWindowID(w), C.int(x), C.int(y),
		C.uint(width), C.uint(height), C.bool(exposures))
}

// ---- Graphics context ----

func CreateGC(fg, bg uint64, lineWidth, function int) GC {
	return GC(C.CocoaCreateGC(C.uint64_t(fg), C.uint64_t(bg), C.int(lineWidth), C.int(function)))
}
func FreeGC(gc GC)                       { C.CocoaFreeGC(C.CocoaGCID(gc)) }
func SetForeground(gc GC, pixel uint64)  { C.CocoaSetForeground(C.CocoaGCID(gc), C.uint64_t(pixel)) }
func SetBackground(gc GC, pixel uint64)  { C.CocoaSetBackground(C.CocoaGCID(gc), C.uint64_t(pixel)) }
func SetFillStyle(gc GC, fillStyle int)  { C.CocoaSetFillStyle(C.CocoaGCID(gc), C.int(fillStyle)) }
func SetLineAttributes(gc GC, lineWidth uint, lineStyle, capStyle, joinStyle int) {
	C.CocoaSetLineAttributes(C.CocoaGCID(gc), C.uint(lineWidth),
		C.int(lineStyle), C.int(capStyle), C.int(joinStyle))
}
func SetDashes(gc GC, dashOffset int, dashList []byte) {
	C.CocoaSetDashes(C.CocoaGCID(gc), C.int(dashOffset),
		(*C.uchar)(unsafe.Pointer(&dashList[0])), C.int(len(dashList)))
}

// ---- Drawing ----

func FillRectangle(d Drawable, gc GC, x, y int, w, h uint) {
	C.CocoaFillRectangle(C.CocoaDrawableID(d), C.CocoaGCID(gc),
		C.int(x), C.int(y), C.uint(w), C.uint(h))
}
func DrawRectangle(d Drawable, gc GC, x, y int, w, h uint) {
	C.CocoaDrawRectangle(C.CocoaDrawableID(d), C.CocoaGCID(gc),
		C.int(x), C.int(y), C.uint(w), C.uint(h))
}
func DrawLine(d Drawable, gc GC, x1, y1, x2, y2 int) {
	C.CocoaDrawLine(C.CocoaDrawableID(d), C.CocoaGCID(gc),
		C.int(x1), C.int(y1), C.int(x2), C.int(y2))
}

// XPoint matches the platform.Point layout: two int16 fields.
type XPoint struct {
	X, Y int16
}

func DrawLines(d Drawable, gc GC, points []XPoint, mode int) {
	C.CocoaDrawLines(C.CocoaDrawableID(d), C.CocoaGCID(gc),
		(*C.int16_t)(unsafe.Pointer(&points[0])), C.int(len(points)), C.int(mode))
}
func FillPolygon(d Drawable, gc GC, points []XPoint, shape, mode int) {
	C.CocoaFillPolygon(C.CocoaDrawableID(d), C.CocoaGCID(gc),
		(*C.int16_t)(unsafe.Pointer(&points[0])), C.int(len(points)),
		C.int(shape), C.int(mode))
}
func FillArc(d Drawable, gc GC, x, y int, w, h uint, angle1, angle2 int) {
	C.CocoaFillArc(C.CocoaDrawableID(d), C.CocoaGCID(gc),
		C.int(x), C.int(y), C.uint(w), C.uint(h), C.int(angle1), C.int(angle2))
}
func DrawArc(d Drawable, gc GC, x, y int, w, h uint, angle1, angle2 int) {
	C.CocoaDrawArc(C.CocoaDrawableID(d), C.CocoaGCID(gc),
		C.int(x), C.int(y), C.uint(w), C.uint(h), C.int(angle1), C.int(angle2))
}
func CopyArea(src, dst Drawable, gc GC, srcX, srcY int, w, h uint, dstX, dstY int) {
	C.CocoaCopyArea(C.CocoaDrawableID(src), C.CocoaDrawableID(dst), C.CocoaGCID(gc),
		C.int(srcX), C.int(srcY), C.uint(w), C.uint(h), C.int(dstX), C.int(dstY))
}
func PutImageRGBA(d Drawable, gc GC, depth int,
	rgbaData []byte, stride int, imgW, imgH int,
	srcX, srcY, dstX, dstY, w, h int, bgPixel uint64) {
	C.CocoaPutImageRGBA(C.CocoaDrawableID(d), C.CocoaGCID(gc), C.int(depth),
		(*C.uchar)(unsafe.Pointer(&rgbaData[0])), C.int(stride),
		C.int(imgW), C.int(imgH),
		C.int(srcX), C.int(srcY), C.int(dstX), C.int(dstY),
		C.int(w), C.int(h), C.uint64_t(bgPixel))
}

// ---- Pixmap ----

func CreatePixmap(width, height, depth uint) Pixmap {
	return Pixmap(C.CocoaCreatePixmap(C.uint(width), C.uint(height), C.uint(depth)))
}
func FreePixmap(pm Pixmap) { C.CocoaFreePixmap(C.CocoaPixmapID(pm)) }
func CreateBitmapFromData(bits []byte, width, height uint) Pixmap {
	return Pixmap(C.CocoaCreateBitmapFromData((*C.uchar)(unsafe.Pointer(&bits[0])),
		C.uint(width), C.uint(height)))
}

// ---- Cursor ----

func CreateCursor(shape uint) Cursor {
	return Cursor(C.CocoaCreateCursor(C.uint(shape)))
}
func DefineCursor(w Window, c Cursor) {
	C.CocoaDefineCursor(C.CocoaWindowID(w), C.CocoaCursorID(c))
}
func SetCursorShape(w Window, shape uint) {
	C.CocoaSetCursorShape(C.CocoaWindowID(w), C.uint(shape))
}
func UndefineCursor(w Window) { C.CocoaUndefineCursor(C.CocoaWindowID(w)) }
func FreeCursor(c Cursor)    { C.CocoaFreeCursor(C.CocoaCursorID(c)) }

// ---- Grab ----

func GrabPointer(w Window) int    { return int(C.CocoaGrabPointer(C.CocoaWindowID(w))) }
func UngrabPointer()               { C.CocoaUngrabPointer() }
func GrabKeyboard(w Window) int   { return int(C.CocoaGrabKeyboard(C.CocoaWindowID(w))) }
func UngrabKeyboard()              { C.CocoaUngrabKeyboard() }

// ---- Clipboard ----

func SetClipboardText(text string) {
	ctext := C.CString(text)
	defer C.free(unsafe.Pointer(ctext))
	C.CocoaSetClipboardText(ctext)
}

func GetClipboardText() string {
	cstr := C.CocoaGetClipboardText()
	if cstr == nil {
		return ""
	}
	defer C.CocoaFreeString(cstr)
	return C.GoString(cstr)
}

// ---- Events ----

func NextEvent() *RawEvent {
	ev := &RawEvent{}
	for {
		if C.CocoaNextEvent(&ev.ev) != 0 {
			return ev
		}
		// No event yet — the pump is running on the main thread, and events
		// will arrive asynchronously. Loop and retry.
		runtime.Gosched()
	}
}

func Pending() int { return int(C.CocoaPending()) }

// ---- Atoms ----

func InternAtom(name string, onlyIfExists bool) Atom {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return Atom(C.CocoaInternAtom(cname, C.bool(onlyIfExists)))
}

func GetAtomName(atom Atom) string {
	return C.GoString(C.CocoaGetAtomName(C.uint64_t(atom)))
}

// ---- Font ----

func OpenFont(family string, size float64, weight, slant int) FontID {
	cfamily := C.CString(family)
	defer C.free(unsafe.Pointer(cfamily))
	return FontID(C.CocoaOpenFont(cfamily, C.double(size), C.int(weight), C.int(slant)))
}

func CloseFont(f FontID)       { C.CocoaCloseFont(C.CocoaFontID(f)) }
func FontAscent(f FontID) int  { return int(C.CocoaFontAscent(C.CocoaFontID(f))) }
func FontDescent(f FontID) int { return int(C.CocoaFontDescent(C.CocoaFontID(f))) }
func FontMaxWidth(f FontID) int { return int(C.CocoaFontMaxWidth(C.CocoaFontID(f))) }
func FontIsFixed(f FontID) bool { return bool(C.CocoaFontIsFixed(C.CocoaFontID(f))) }

func MeasureString(f FontID, s string) int {
	if len(s) == 0 {
		return 0
	}
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	return int(C.CocoaMeasureString(C.CocoaFontID(f), cs, C.int(len(s))))
}

func DrawString(d Drawable, f FontID, x, y int, s string, pixel uint64, r, g, b uint16) {
	if len(s) == 0 {
		return
	}
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	C.CocoaDrawString(C.CocoaDrawableID(d), C.CocoaFontID(f),
		C.int(x), C.int(y), cs, C.int(len(s)),
		C.uint64_t(pixel), C.uint16_t(r), C.uint16_t(g), C.uint16_t(b))
}

// ---- WM ----

func SetWMProtocols(w Window)        { C.CocoaSetWMProtocols(C.CocoaWindowID(w)) }
func IconifyWindow(w Window)         { C.CocoaIconifyWindow(C.CocoaWindowID(w)) }
func WithdrawWindow(w Window)        { C.CocoaWithdrawWindow(C.CocoaWindowID(w)) }

func SetWMHints(w Window, input bool, initialState int) {
	C.CocoaSetWMHints(C.CocoaWindowID(w), C.bool(input), C.int(initialState))
}

func SetWMNormalHints(w Window, minW, minH, maxW, maxH int) {
	C.CocoaSetWMNormalHints(C.CocoaWindowID(w), C.int(minW), C.int(minH),
		C.int(maxW), C.int(maxH))
}

func SetClassHint(w Window, name, class string) {
	cname := C.CString(name)
	cclass := C.CString(class)
	defer C.free(unsafe.Pointer(cname))
	defer C.free(unsafe.Pointer(cclass))
	C.CocoaSetClassHint(C.CocoaWindowID(w), cname, cclass)
}

func SetTransientFor(w, parent Window) {
	C.CocoaSetTransientFor(C.CocoaWindowID(w), C.CocoaWindowID(parent))
}

func SetIconName(w Window, name string) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	C.CocoaSetIconName(C.CocoaWindowID(w), cname)
}
