//go:build darwin

// Package cocoa provides the macOS/Cocoa backend for the platform abstraction layer.
//
// # Known gaps
//
// The backend currently satisfies platform.DisplayServer with a number of
// no-op or zero-returning stubs. They are intentionally left empty so the
// type checks, but they drop data — code that relies on them works on X11
// and silently fails on macOS. The list below is the source of truth for
// what is intentionally stubbed versus what is incomplete:
//
//   - ChangeProperty, DeleteProperty, GetWindowProperty — window
//     properties (other than WM_NAME) are not yet wired to NSWindow.
//   - SendEvent, SendClientMessage — synthetic event dispatch is X11-only.
//   - ConvertSelection, SendSelectionNotify — clipboard transfer is
//     partial: SetSelectionOwner only records the owner in-process and
//     does not advertise it via NSPasteboard.
//   - InitIM, SetICFocus, UnsetICFocus — no Cocoa input method integration;
//     HasIM returns false so the event loop skips IM routing.
//   - CreatePixmap stipple parameters — CGPatternRef support missing
//     (see TODO in this file).
//
// Before removing any of these stubs, port the behavior to NSWindow /
// NSPasteboard / NSTextInputClient as appropriate.
package cocoa

import (
	"github.com/msorc/takigo/font"
	clib "github.com/msorc/takigo/internal/cocoa"
	"github.com/msorc/takigo/platform"
)

// Type conversion helpers — platform types are uintptr, internal/cocoa types
// are C.uintptr_t based. We go through uintptr as an intermediary.
func toWin(w platform.WindowID) clib.Window          { return clib.Window(uintptr(w)) }
func toDrawable(d platform.DrawableID) clib.Drawable { return clib.Drawable(uintptr(d)) }
func toGC(gc platform.GCID) clib.GC                  { return clib.GC(uintptr(gc)) }
func toCursor(c platform.CursorID) clib.Cursor       { return clib.Cursor(uintptr(c)) }
func toPixmap(p platform.PixmapID) clib.Pixmap       { return clib.Pixmap(uintptr(p)) }
func fromWin(w clib.Window) platform.WindowID        { return platform.WindowID(uintptr(w)) }

// CocoaDisplay wraps the low-level Cocoa connection.
type CocoaDisplay struct {
	rootWindow platform.WindowID
	clipOwner  platform.WindowID
	atoms      *platform.Atoms
}

// NewDisplayServer initializes Cocoa and returns a composed DisplayServer
// along with a FontOpener for the default screen.
func NewDisplayServer(displayName string) (platform.DisplayServer, font.FontOpener, error) {
	clib.Init()

	rootWin := clib.CreateWindow(clib.Window(0), 0, 0, 1, 1, 0, 0x00D9D9D9, 0, false)

	core := &CocoaDisplay{
		rootWindow: fromWin(rootWin),
		atoms: &platform.Atoms{
			WMName:        platform.AtomID(clib.InternAtom("WM_NAME", false)),
			String:        platform.AtomID(clib.InternAtom("STRING", false)),
			WMNormalHints: platform.AtomID(clib.InternAtom("WM_NORMAL_HINTS", false)),
			Primary:       platform.AtomID(clib.InternAtom("PRIMARY", false)),
			Secondary:     platform.AtomID(clib.InternAtom("SECONDARY", false)),
			Atom:          platform.AtomID(clib.InternAtom("ATOM", false)),
			Cardinal:      platform.AtomID(clib.InternAtom("CARDINAL", false)),
			Window:        platform.AtomID(clib.InternAtom("WINDOW", false)),
		},
	}
	ds := platform.NewDisplayServer(
		core, // DisplayCore
		core, // WindowManager
		core, // Drawer
		core, // GCManager
		core, // PixmapManager
		core, // EventSource
		core, // GrabManager
		core, // SelectionManager
		core, // CursorManager
		core, // PropertyManager
		core, // InputMethodManager
	)
	return ds, core.FontOpener(core.DefaultScreen()), nil
}

// Atoms returns the resolved well-known atom table.
func (d *CocoaDisplay) Atoms() *platform.Atoms { return d.atoms }

// EventParser creates a CocoaEventParser.
func (d *CocoaDisplay) EventParser() platform.EventParser {
	return &EventParser{}
}

// FontOpener creates a FontOpener for font loading.
func (d *CocoaDisplay) FontOpener(screen int) font.FontOpener {
	return &FontOpener{}
}

// --- DisplayCore ---

func (d *CocoaDisplay) Close()                                  { clib.Stop() }
func (d *CocoaDisplay) DefaultScreen() int                      { return 0 }
func (d *CocoaDisplay) DefaultRootWindow() platform.WindowID    { return d.rootWindow }
func (d *CocoaDisplay) RootWindow(screen int) platform.WindowID { return d.rootWindow }
func (d *CocoaDisplay) DefaultDepth(screen int) int             { return clib.ScreenDepth() }
func (d *CocoaDisplay) ScreenWidth(screen int) int              { return clib.ScreenWidth() }
func (d *CocoaDisplay) ScreenHeight(screen int) int             { return clib.ScreenHeight() }
func (d *CocoaDisplay) ScreenWidthMM(screen int) int            { return clib.ScreenWidthMM() }
func (d *CocoaDisplay) ScreenHeightMM(screen int) int           { return clib.ScreenHeightMM() }
func (d *CocoaDisplay) WhitePixel(screen int) uint64            { return 0x00FFFFFF }
func (d *CocoaDisplay) BlackPixel(screen int) uint64            { return 0x00000000 }
func (d *CocoaDisplay) ConnectionNumber() int                   { return -1 }
func (d *CocoaDisplay) Sync(discard bool)                       { clib.Flush() }
func (d *CocoaDisplay) Flush()                                  { clib.Flush() }
func (d *CocoaDisplay) Pending() int                            { return clib.Pending() }
func (d *CocoaDisplay) ResourceManagerString() string           { return "" }

// --- WindowManager ---

func (d *CocoaDisplay) CreateWindow(parent platform.WindowID, x, y int, width, height, borderWidth uint,
	depth int, class uint, valueMask uint64, attrs *platform.WindowAttrs) platform.WindowID {
	w := clib.CreateWindow(toWin(parent), x, y, width, height, borderWidth,
		attrs.BackgroundPixel, attrs.EventMask, attrs.OverrideRedirect)
	return fromWin(w)
}

func (d *CocoaDisplay) CreateSimpleWindow(parent platform.WindowID, x, y int, width, height, borderWidth uint, border, background uint64) platform.WindowID {
	w := clib.CreateSimpleWindow(toWin(parent), x, y, width, height, borderWidth, border, background)
	return fromWin(w)
}

func (d *CocoaDisplay) DestroyWindow(w platform.WindowID) { clib.DestroyWindow(toWin(w)) }
func (d *CocoaDisplay) MapWindow(w platform.WindowID)     { clib.MapWindow(toWin(w)) }
func (d *CocoaDisplay) MapRaised(w platform.WindowID)     { clib.MapRaised(toWin(w)) }
func (d *CocoaDisplay) UnmapWindow(w platform.WindowID)   { clib.UnmapWindow(toWin(w)) }
func (d *CocoaDisplay) RaiseWindow(w platform.WindowID)   { clib.RaiseWindow(toWin(w)) }
func (d *CocoaDisplay) LowerWindow(w platform.WindowID)   { clib.LowerWindow(toWin(w)) }

func (d *CocoaDisplay) MoveWindow(w platform.WindowID, x, y int) {
	clib.MoveWindow(toWin(w), x, y)
}
func (d *CocoaDisplay) ResizeWindow(w platform.WindowID, width, height uint) {
	clib.ResizeWindow(toWin(w), width, height)
}
func (d *CocoaDisplay) MoveResizeWindow(w platform.WindowID, x, y int, width, height uint) {
	clib.MoveResizeWindow(toWin(w), x, y, width, height)
}
func (d *CocoaDisplay) SelectInput(w platform.WindowID, eventMask int64) {
	clib.SelectInput(toWin(w), eventMask)
}
func (d *CocoaDisplay) StoreName(w platform.WindowID, name string) {
	clib.StoreName(toWin(w), name)
}
func (d *CocoaDisplay) TranslateCoordinates(src, dst platform.WindowID, srcX, srcY int) (int, int) {
	return clib.TranslateCoordinates(toWin(src), toWin(dst), srcX, srcY)
}

// --- Drawer ---

func (d *CocoaDisplay) FillRectangle(drawable platform.DrawableID, gc platform.GCID, x, y int, width, height uint) {
	clib.FillRectangle(toDrawable(drawable), toGC(gc), x, y, width, height)
}
func (d *CocoaDisplay) DrawRectangle(drawable platform.DrawableID, gc platform.GCID, x, y int, width, height uint) {
	clib.DrawRectangle(toDrawable(drawable), toGC(gc), x, y, width, height)
}
func (d *CocoaDisplay) DrawLine(drawable platform.DrawableID, gc platform.GCID, x1, y1, x2, y2 int) {
	clib.DrawLine(toDrawable(drawable), toGC(gc), x1, y1, x2, y2)
}
func (d *CocoaDisplay) DrawLines(drawable platform.DrawableID, gc platform.GCID, points []platform.Point, mode int) {
	xpoints := make([]clib.XPoint, len(points))
	for i, p := range points {
		xpoints[i] = clib.XPoint{X: p.X, Y: p.Y}
	}
	clib.DrawLines(toDrawable(drawable), toGC(gc), xpoints, mode)
}
func (d *CocoaDisplay) FillPolygon(drawable platform.DrawableID, gc platform.GCID, points []platform.Point, shape, mode int) {
	xpoints := make([]clib.XPoint, len(points))
	for i, p := range points {
		xpoints[i] = clib.XPoint{X: p.X, Y: p.Y}
	}
	clib.FillPolygon(toDrawable(drawable), toGC(gc), xpoints, shape, mode)
}
func (d *CocoaDisplay) FillArc(drawable platform.DrawableID, gc platform.GCID, x, y int, width, height uint, angle1, angle2 int) {
	clib.FillArc(toDrawable(drawable), toGC(gc), x, y, width, height, angle1, angle2)
}
func (d *CocoaDisplay) DrawArc(drawable platform.DrawableID, gc platform.GCID, x, y int, width, height uint, angle1, angle2 int) {
	clib.DrawArc(toDrawable(drawable), toGC(gc), x, y, width, height, angle1, angle2)
}
func (d *CocoaDisplay) ClearWindow(w platform.WindowID) {
	clib.ClearWindow(toWin(w))
}
func (d *CocoaDisplay) SetWindowBackground(w platform.WindowID, pixel uint64) {
	clib.SetWindowBackground(toWin(w), pixel)
}
func (d *CocoaDisplay) ClearArea(w platform.WindowID, x, y int, width, height uint, exposures bool) {
	clib.ClearArea(toWin(w), x, y, width, height, exposures)
}
func (d *CocoaDisplay) CopyArea(src, dst platform.DrawableID, gc platform.GCID, srcX, srcY int, width, height uint, dstX, dstY int) {
	clib.CopyArea(toDrawable(src), toDrawable(dst), toGC(gc), srcX, srcY, width, height, dstX, dstY)
}
func (d *CocoaDisplay) PutImageRGBA(drawable platform.DrawableID, gc platform.GCID, depth int,
	rgbaData []byte, stride int, imgW, imgH int,
	srcX, srcY, dstX, dstY, w, h int, bgPixel uint64) {
	clib.PutImageRGBA(toDrawable(drawable), toGC(gc), depth,
		rgbaData, stride, imgW, imgH, srcX, srcY, dstX, dstY, w, h, bgPixel)
}
func (d *CocoaDisplay) GetImageRGBA(platform.DrawableID, int, int, int, int) []byte {
	return nil
}
func (d *CocoaDisplay) SetDashes(gc platform.GCID, dashOffset int, dashList []byte) {
	clib.SetDashes(toGC(gc), dashOffset, dashList)
}

// --- GCManager ---

func (d *CocoaDisplay) CreateGC(drawable platform.DrawableID, valueMask uint64, values *platform.GCValues) platform.GCID {
	gc := clib.CreateGC(values.Foreground, values.Background, values.LineWidth, values.Function)
	return platform.GCID(uintptr(gc))
}
func (d *CocoaDisplay) FreeGC(gc platform.GCID) { clib.FreeGC(toGC(gc)) }
func (d *CocoaDisplay) SetForeground(gc platform.GCID, pixel uint64) {
	clib.SetForeground(toGC(gc), pixel)
}
func (d *CocoaDisplay) SetBackground(gc platform.GCID, pixel uint64) {
	clib.SetBackground(toGC(gc), pixel)
}
func (d *CocoaDisplay) SetLineAttributes(gc platform.GCID, lineWidth uint, lineStyle, capStyle, joinStyle int) {
	clib.SetLineAttributes(toGC(gc), lineWidth, lineStyle, capStyle, joinStyle)
}
func (d *CocoaDisplay) SetFillStyle(gc platform.GCID, fillStyle int) {
	clib.SetFillStyle(toGC(gc), fillStyle)
}
func (d *CocoaDisplay) SetStipple(gc platform.GCID, stipple platform.PixmapID) {
	// TODO: implement stipple pattern support via CGPatternRef.
}
func (d *CocoaDisplay) SetTSOrigin(gc platform.GCID, x, y int) {}

// --- PixmapManager ---

func (d *CocoaDisplay) CreatePixmap(drawable platform.DrawableID, width, height, depth uint) platform.PixmapID {
	return platform.PixmapID(uintptr(clib.CreatePixmap(width, height, depth)))
}
func (d *CocoaDisplay) FreePixmap(pixmap platform.PixmapID) {
	clib.FreePixmap(toPixmap(pixmap))
}
func (d *CocoaDisplay) CreateBitmapFromData(drawable platform.DrawableID, bits []byte, width, height uint) platform.PixmapID {
	return platform.PixmapID(uintptr(clib.CreateBitmapFromData(bits, width, height)))
}

// --- EventSource ---

func (d *CocoaDisplay) NextEvent() *platform.RawEvent {
	raw := clib.NextEvent()
	return &platform.RawEvent{
		Data:        raw,
		EventType:   raw.Type(),
		EventWindow: platform.WindowID(uintptr(raw.Window())),
	}
}
func (d *CocoaDisplay) FilterEvent(ev *platform.RawEvent) bool {
	return false
}

// --- GrabManager ---

func (d *CocoaDisplay) GrabPointer(grabWindow platform.WindowID, ownerEvents bool, eventMask uint,
	pointerMode, keyboardMode int, confineTo platform.WindowID, cursor platform.CursorID, time platform.Timestamp) int {
	return clib.GrabPointer(toWin(grabWindow))
}
func (d *CocoaDisplay) UngrabPointer(time platform.Timestamp) { clib.UngrabPointer() }
func (d *CocoaDisplay) GrabKeyboard(grabWindow platform.WindowID, ownerEvents bool,
	pointerMode, keyboardMode int, time platform.Timestamp) int {
	return clib.GrabKeyboard(toWin(grabWindow))
}
func (d *CocoaDisplay) UngrabKeyboard(time platform.Timestamp) { clib.UngrabKeyboard() }

// --- SelectionManager ---

func (d *CocoaDisplay) SetSelectionOwner(selection platform.AtomID, owner platform.WindowID, time platform.Timestamp) {
	// Partial: stores the owner in-process; does not advertise via
	// NSPasteboard. See package doc ("Known gaps").
	d.clipOwner = owner
}
func (d *CocoaDisplay) GetSelectionOwner(selection platform.AtomID) platform.WindowID {
	return d.clipOwner
}
func (d *CocoaDisplay) ConvertSelection(selection, target, property platform.AtomID, requestor platform.WindowID, time platform.Timestamp) {
	// Stub. See package doc ("Known gaps").
}
func (d *CocoaDisplay) SendSelectionNotify(requestor platform.WindowID, selection, target, property platform.AtomID, time platform.Timestamp) {
	// Stub. See package doc ("Known gaps").
}

// --- CursorManager ---

func (d *CocoaDisplay) CreateFontCursor(shape uint) platform.CursorID {
	return platform.CursorID(uintptr(clib.CreateCursor(shape)))
}
func (d *CocoaDisplay) DefineCursor(w platform.WindowID, cursor platform.CursorID) {
	clib.DefineCursor(toWin(w), toCursor(cursor))
}
func (d *CocoaDisplay) SetCursorShape(w platform.WindowID, shape uint) {
	clib.SetCursorShape(toWin(w), shape)
}
func (d *CocoaDisplay) UndefineCursor(w platform.WindowID)  { clib.UndefineCursor(toWin(w)) }
func (d *CocoaDisplay) FreeCursor(cursor platform.CursorID) { clib.FreeCursor(toCursor(cursor)) }

// --- PropertyManager ---

func (d *CocoaDisplay) InternAtom(name string, onlyIfExists bool) platform.AtomID {
	return platform.AtomID(clib.InternAtom(name, onlyIfExists))
}
func (d *CocoaDisplay) GetAtomName(atom platform.AtomID) string {
	return clib.GetAtomName(clib.Atom(atom))
}
func (d *CocoaDisplay) SetWMProtocols(w platform.WindowID, protocols []platform.AtomID) int {
	clib.SetWMProtocols(toWin(w))
	return 1
}
func (d *CocoaDisplay) SetWMNormalHints(w platform.WindowID, hints *platform.SizeHints) {
	clib.SetWMNormalHints(toWin(w), hints.MinWidth, hints.MinHeight, hints.MaxWidth, hints.MaxHeight)
}
func (d *CocoaDisplay) SetWMHints(w platform.WindowID, hints *platform.WMHints) {
	clib.SetWMHints(toWin(w), hints.Input, hints.InitialState)
}
func (d *CocoaDisplay) SetClassHint(w platform.WindowID, name, class string) {
	clib.SetClassHint(toWin(w), name, class)
}
func (d *CocoaDisplay) SetTransientForHint(w platform.WindowID, propWindow platform.WindowID) {
	clib.SetTransientFor(toWin(w), toWin(propWindow))
}
func (d *CocoaDisplay) SetInputFocus(w platform.WindowID, revertTo int, time platform.Timestamp) {
	clib.SetInputFocus(toWin(w))
}
func (d *CocoaDisplay) GetInputFocus() (platform.WindowID, int) {
	return d.rootWindow, platform.RevertToParent
}
func (d *CocoaDisplay) ChangeProperty(w platform.WindowID, prop, propType platform.AtomID, format int, mode int, data []byte, nelements int) {
	// Stub. See package doc ("Known gaps").
}
func (d *CocoaDisplay) ChangePropertyString(w platform.WindowID, property, typ platform.AtomID, data string) {
	name := d.GetAtomName(property)
	if name == "_NET_WM_NAME" || name == "WM_NAME" {
		clib.StoreName(toWin(w), data)
	}
}
func (d *CocoaDisplay) ChangePropertyAtoms(w platform.WindowID, prop platform.AtomID, atoms []platform.AtomID) {
	// Stub. See package doc ("Known gaps").
}
func (d *CocoaDisplay) GetWindowProperty(w platform.WindowID, property platform.AtomID, offset, length int64, delete bool) ([]byte, platform.AtomID, int) {
	// Stub. See package doc ("Known gaps").
	return nil, 0, 0
}
func (d *CocoaDisplay) DeleteProperty(w platform.WindowID, prop platform.AtomID) {
	// Stub. See package doc ("Known gaps").
}
func (d *CocoaDisplay) SendEvent(w platform.WindowID, propagate bool, eventMask int64, ev *platform.RawEvent) {
	// Stub. See package doc ("Known gaps").
}
func (d *CocoaDisplay) SendClientMessage(w, target platform.WindowID, msgType platform.AtomID, d0, d1, d2, d3, d4 int64) {
	// Stub. See package doc ("Known gaps").
}
func (d *CocoaDisplay) IconifyWindow(w platform.WindowID, screen int) {
	clib.IconifyWindow(toWin(w))
}
func (d *CocoaDisplay) WithdrawWindow(w platform.WindowID, screen int) {
	clib.WithdrawWindow(toWin(w))
}
func (d *CocoaDisplay) SetIconName(w platform.WindowID, name string) {
	clib.SetIconName(toWin(w), name)
}

// --- EventPumper (for event.EventPumper interface) ---

// PumpEvents processes all pending NSEvents on the main thread.
// This is called by the event loop on each tick to drive the Cocoa event system.
func (d *CocoaDisplay) PumpEvents() {
	clib.PumpEvents()
}

// --- InputMethodManager ---

// InitIM, HasIM, SetICFocus and UnsetICFocus are stubs. HasIM reports
// false so the event loop skips IM routing. See package doc ("Known gaps").
func (d *CocoaDisplay) InitIM(root platform.WindowID)  {}
func (d *CocoaDisplay) HasIM() bool                    { return false }
func (d *CocoaDisplay) SetICFocus(w platform.WindowID) {}
func (d *CocoaDisplay) UnsetICFocus()                  {}

// Compile-time interface checks
var _ platform.DisplayCore = (*CocoaDisplay)(nil)
var _ platform.WindowManager = (*CocoaDisplay)(nil)
var _ platform.Drawer = (*CocoaDisplay)(nil)
var _ platform.GCManager = (*CocoaDisplay)(nil)
var _ platform.PixmapManager = (*CocoaDisplay)(nil)
var _ platform.EventSource = (*CocoaDisplay)(nil)
var _ platform.GrabManager = (*CocoaDisplay)(nil)
var _ platform.SelectionManager = (*CocoaDisplay)(nil)
var _ platform.CursorManager = (*CocoaDisplay)(nil)
var _ platform.PropertyManager = (*CocoaDisplay)(nil)
var _ platform.InputMethodManager = (*CocoaDisplay)(nil)
