//go:build linux || freebsd || openbsd || netbsd

// Package x11 provides the X11/Xlib backend for the platform abstraction layer.
package x11

import (
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/platform"
)

// X11Display implements platform.DisplayServer by wrapping *xlib.Display.
type X11Display struct {
	dpy *xlib.Display
}

// NewDisplayServer opens an X11 display connection and returns a DisplayServer.
func NewDisplayServer(name string) (*X11Display, error) {
	dpy, err := xlib.OpenDisplay(name)
	if err != nil {
		return nil, err
	}
	return &X11Display{dpy: dpy}, nil
}

// EventParser creates an X11EventParser for this display.
func (s *X11Display) EventParser() *X11EventParser {
	return NewEventParser(s.dpy)
}

// FontOpener creates an X11FontOpener for this display's default screen.
func (s *X11Display) FontOpener(screen int) *X11FontOpener {
	return NewFontOpener(s.dpy, screen, s.dpy.DefaultVisual(screen), s.dpy.DefaultColormap(screen))
}

// --- DisplayServer core methods ---

func (s *X11Display) Close()             { s.dpy.Close() }
func (s *X11Display) DefaultScreen() int { return s.dpy.DefaultScreen() }
func (s *X11Display) DefaultRootWindow() platform.WindowID {
	return platform.WindowID(s.dpy.DefaultRootWindow())
}
func (s *X11Display) RootWindow(screen int) platform.WindowID {
	return platform.WindowID(s.dpy.RootWindow(screen))
}
func (s *X11Display) DefaultDepth(screen int) int   { return s.dpy.DefaultDepth(screen) }
func (s *X11Display) ScreenWidth(screen int) int    { return s.dpy.ScreenWidth(screen) }
func (s *X11Display) ScreenHeight(screen int) int   { return s.dpy.ScreenHeight(screen) }
func (s *X11Display) ScreenWidthMM(screen int) int  { return s.dpy.ScreenWidthMM(screen) }
func (s *X11Display) ScreenHeightMM(screen int) int { return s.dpy.ScreenHeightMM(screen) }
func (s *X11Display) WhitePixel(screen int) uint64  { return s.dpy.WhitePixel(screen) }
func (s *X11Display) BlackPixel(screen int) uint64  { return s.dpy.BlackPixel(screen) }
func (s *X11Display) ConnectionNumber() int         { return s.dpy.ConnectionNumber() }
func (s *X11Display) Sync(discard bool)             { s.dpy.Sync(discard) }
func (s *X11Display) Flush()                        { s.dpy.Flush() }
func (s *X11Display) Pending() int                  { return s.dpy.Pending() }
func (s *X11Display) ResourceManagerString() string { return s.dpy.ResourceManagerString() }

// --- WindowManager ---

func (s *X11Display) CreateWindow(parent platform.WindowID, x, y int, width, height, borderWidth uint,
	depth int, class uint, valueMask uint64, attrs *platform.WindowAttrs) platform.WindowID {

	xattrs := &xlib.WindowAttributes{
		BackgroundPixel:  attrs.BackgroundPixel,
		BorderPixel:      attrs.BorderPixel,
		BitGravity:       attrs.BitGravity,
		EventMask:        attrs.EventMask,
		OverrideRedirect: attrs.OverrideRedirect,
	}

	// Get visual for the display.
	screen := s.dpy.DefaultScreen()
	visual := s.dpy.DefaultVisual(screen)
	colormap := s.dpy.DefaultColormap(screen)

	// Add colormap if requested via CW bits (the platform interface doesn't
	// expose colormap directly; X11 backend handles it).
	if valueMask&xlib.CWColormap != 0 {
		xattrs.Colormap = colormap
	}

	w := s.dpy.CreateWindow(xlib.Window(parent), x, y, width, height, borderWidth,
		depth, class, visual, valueMask, xattrs)
	return platform.WindowID(w)
}

func (s *X11Display) CreateSimpleWindow(parent platform.WindowID, x, y int, width, height, borderWidth uint, border, background uint64) platform.WindowID {
	return platform.WindowID(s.dpy.CreateSimpleWindow(xlib.Window(parent), x, y, width, height, borderWidth, border, background))
}

func (s *X11Display) DestroyWindow(w platform.WindowID) { s.dpy.DestroyWindow(xlib.Window(w)) }
func (s *X11Display) MapWindow(w platform.WindowID)     { s.dpy.MapWindow(xlib.Window(w)) }
func (s *X11Display) MapRaised(w platform.WindowID)     { s.dpy.MapRaised(xlib.Window(w)) }
func (s *X11Display) UnmapWindow(w platform.WindowID)   { s.dpy.UnmapWindow(xlib.Window(w)) }
func (s *X11Display) RaiseWindow(w platform.WindowID)   { s.dpy.RaiseWindow(xlib.Window(w)) }
func (s *X11Display) LowerWindow(w platform.WindowID)   { s.dpy.LowerWindow(xlib.Window(w)) }
func (s *X11Display) MoveWindow(w platform.WindowID, x, y int) {
	s.dpy.MoveWindow(xlib.Window(w), x, y)
}
func (s *X11Display) ResizeWindow(w platform.WindowID, width, height uint) {
	s.dpy.ResizeWindow(xlib.Window(w), width, height)
}
func (s *X11Display) MoveResizeWindow(w platform.WindowID, x, y int, width, height uint) {
	s.dpy.MoveResizeWindow(xlib.Window(w), x, y, width, height)
}
func (s *X11Display) SelectInput(w platform.WindowID, eventMask int64) {
	s.dpy.SelectInput(xlib.Window(w), eventMask)
}
func (s *X11Display) StoreName(w platform.WindowID, name string) {
	s.dpy.StoreName(xlib.Window(w), name)
}

func (s *X11Display) TranslateCoordinates(src, dst platform.WindowID, srcX, srcY int) (int, int) {
	return s.dpy.TranslateCoordinates(xlib.Window(src), xlib.Window(dst), srcX, srcY)
}

// --- Drawer ---

func (s *X11Display) FillRectangle(drawable platform.DrawableID, gc platform.GCID, x, y int, width, height uint) {
	s.dpy.FillRectangle(xlib.Drawable(drawable), toXGC(gc), x, y, width, height)
}

func (s *X11Display) DrawRectangle(drawable platform.DrawableID, gc platform.GCID, x, y int, width, height uint) {
	s.dpy.DrawRectangle(xlib.Drawable(drawable), toXGC(gc), x, y, width, height)
}

func (s *X11Display) DrawLine(drawable platform.DrawableID, gc platform.GCID, x1, y1, x2, y2 int) {
	s.dpy.DrawLine(xlib.Drawable(drawable), toXGC(gc), x1, y1, x2, y2)
}

func (s *X11Display) DrawLines(drawable platform.DrawableID, gc platform.GCID, points []platform.Point, mode int) {
	xpoints := make([]xlib.XPoint, len(points))
	for i, p := range points {
		xpoints[i] = xlib.XPoint{X: p.X, Y: p.Y}
	}
	s.dpy.DrawLines(xlib.Drawable(drawable), toXGC(gc), xpoints, mode)
}

func (s *X11Display) FillPolygon(drawable platform.DrawableID, gc platform.GCID, points []platform.Point, shape, mode int) {
	xpoints := make([]xlib.XPoint, len(points))
	for i, p := range points {
		xpoints[i] = xlib.XPoint{X: p.X, Y: p.Y}
	}
	s.dpy.FillPolygon(xlib.Drawable(drawable), toXGC(gc), xpoints, shape, mode)
}

func (s *X11Display) FillArc(drawable platform.DrawableID, gc platform.GCID, x, y int, width, height uint, angle1, angle2 int) {
	s.dpy.FillArc(xlib.Drawable(drawable), toXGC(gc), x, y, width, height, angle1, angle2)
}

func (s *X11Display) DrawArc(drawable platform.DrawableID, gc platform.GCID, x, y int, width, height uint, angle1, angle2 int) {
	s.dpy.DrawArc(xlib.Drawable(drawable), toXGC(gc), x, y, width, height, angle1, angle2)
}

func (s *X11Display) ClearWindow(w platform.WindowID) { s.dpy.ClearWindow(xlib.Window(w)) }

func (s *X11Display) SetWindowBackground(w platform.WindowID, pixel uint64) {
	s.dpy.SetWindowBackground(xlib.Window(w), pixel)
}

func (s *X11Display) ClearArea(w platform.WindowID, x, y int, width, height uint, exposures bool) {
	s.dpy.ClearArea(xlib.Window(w), x, y, width, height, exposures)
}

func (s *X11Display) CopyArea(src, dst platform.DrawableID, gc platform.GCID, srcX, srcY int, width, height uint, dstX, dstY int) {
	s.dpy.CopyArea(xlib.Drawable(src), xlib.Drawable(dst), toXGC(gc), srcX, srcY, width, height, dstX, dstY)
}

func (s *X11Display) PutImageRGBA(drawable platform.DrawableID, gc platform.GCID, depth int,
	rgbaData []byte, stride int, imgW, imgH int,
	srcX, srcY, dstX, dstY, w, h int, bgPixel uint64) {
	screen := s.dpy.DefaultScreen()
	visual := s.dpy.DefaultVisual(screen)
	s.dpy.PutImageRGBA(xlib.Drawable(drawable), toXGC(gc), visual, depth,
		rgbaData, stride, imgW, imgH, srcX, srcY, dstX, dstY, w, h, bgPixel)
}

func (s *X11Display) SetDashes(gc platform.GCID, dashOffset int, dashList []byte) {
	s.dpy.SetDashes(toXGC(gc), dashOffset, dashList)
}

// --- GCManager ---

func (s *X11Display) CreateGC(drawable platform.DrawableID, valueMask uint64, values *platform.GCValues) platform.GCID {
	xv := &xlib.GCValues{
		Foreground: values.Foreground,
		Background: values.Background,
		LineWidth:  values.LineWidth,
		Function:   values.Function,
	}
	return fromXGC(s.dpy.CreateGC(xlib.Drawable(drawable), valueMask, xv))
}

func (s *X11Display) FreeGC(gc platform.GCID) { s.dpy.FreeGC(toXGC(gc)) }
func (s *X11Display) SetForeground(gc platform.GCID, pixel uint64) {
	s.dpy.SetForeground(toXGC(gc), pixel)
}
func (s *X11Display) SetBackground(gc platform.GCID, pixel uint64) {
	s.dpy.SetBackground(toXGC(gc), pixel)
}

func (s *X11Display) SetLineAttributes(gc platform.GCID, lineWidth uint, lineStyle, capStyle, joinStyle int) {
	s.dpy.SetLineAttributes(toXGC(gc), lineWidth, lineStyle, capStyle, joinStyle)
}

func (s *X11Display) SetFillStyle(gc platform.GCID, fillStyle int) {
	s.dpy.SetFillStyle(toXGC(gc), fillStyle)
}

func (s *X11Display) SetStipple(gc platform.GCID, stipple platform.PixmapID) {
	s.dpy.SetStipple(toXGC(gc), xlib.Pixmap(stipple))
}

// --- PixmapManager ---

func (s *X11Display) CreatePixmap(drawable platform.DrawableID, width, height, depth uint) platform.PixmapID {
	return platform.PixmapID(s.dpy.CreatePixmap(xlib.Drawable(drawable), width, height, depth))
}

func (s *X11Display) FreePixmap(pixmap platform.PixmapID) { s.dpy.FreePixmap(xlib.Pixmap(pixmap)) }

func (s *X11Display) CreateBitmapFromData(drawable platform.DrawableID, bits []byte, width, height uint) platform.PixmapID {
	return platform.PixmapID(s.dpy.CreateBitmapFromData(xlib.Drawable(drawable), bits, width, height))
}

// --- EventSource ---

func (s *X11Display) NextEvent() *platform.RawEvent {
	raw := s.dpy.NextEvent()
	return &platform.RawEvent{
		Data:        raw,
		EventType:   raw.Type(),
		EventWindow: platform.WindowID(raw.Window()),
	}
}

func (s *X11Display) PeekEvent() *platform.RawEvent {
	raw := s.dpy.PeekEvent()
	return &platform.RawEvent{
		Data:        raw,
		EventType:   raw.Type(),
		EventWindow: platform.WindowID(raw.Window()),
	}
}

func (s *X11Display) FilterEvent(ev *platform.RawEvent) bool {
	raw := ev.Data.(*xlib.RawEvent)
	return raw.FilterEvent()
}

// --- GrabManager ---

func (s *X11Display) GrabPointer(grabWindow platform.WindowID, ownerEvents bool, eventMask uint,
	pointerMode, keyboardMode int, confineTo platform.WindowID, cursor platform.CursorID, time platform.Timestamp) int {
	return s.dpy.GrabPointer(xlib.Window(grabWindow), ownerEvents, eventMask,
		pointerMode, keyboardMode, xlib.Window(confineTo), xlib.Cursor(cursor), xlib.Time(time))
}

func (s *X11Display) UngrabPointer(time platform.Timestamp) { s.dpy.UngrabPointer(xlib.Time(time)) }

func (s *X11Display) GrabKeyboard(grabWindow platform.WindowID, ownerEvents bool,
	pointerMode, keyboardMode int, time platform.Timestamp) int {
	return s.dpy.GrabKeyboard(xlib.Window(grabWindow), ownerEvents, pointerMode, keyboardMode, xlib.Time(time))
}

func (s *X11Display) UngrabKeyboard(time platform.Timestamp) { s.dpy.UngrabKeyboard(xlib.Time(time)) }

// --- SelectionManager ---

func (s *X11Display) SetSelectionOwner(selection platform.AtomID, owner platform.WindowID, time platform.Timestamp) {
	s.dpy.SetSelectionOwner(xlib.Atom(selection), xlib.Window(owner), xlib.Time(time))
}

func (s *X11Display) GetSelectionOwner(selection platform.AtomID) platform.WindowID {
	return platform.WindowID(s.dpy.GetSelectionOwner(xlib.Atom(selection)))
}

func (s *X11Display) ConvertSelection(selection, target, property platform.AtomID, requestor platform.WindowID, time platform.Timestamp) {
	s.dpy.ConvertSelection(xlib.Atom(selection), xlib.Atom(target), xlib.Atom(property), xlib.Window(requestor), xlib.Time(time))
}

func (s *X11Display) SendSelectionNotify(requestor platform.WindowID, selection, target, property platform.AtomID, time platform.Timestamp) {
	s.dpy.SendSelectionNotify(xlib.Window(requestor), xlib.Atom(selection), xlib.Atom(target), xlib.Atom(property), xlib.Time(time))
}

// --- CursorManager ---

func (s *X11Display) CreateFontCursor(shape uint) platform.CursorID {
	return platform.CursorID(s.dpy.CreateFontCursor(shape))
}

func (s *X11Display) DefineCursor(w platform.WindowID, cursor platform.CursorID) {
	s.dpy.DefineCursor(xlib.Window(w), xlib.Cursor(cursor))
}

// shapeToX11Cursor maps abstract cursor.Shape values to X11 cursorfont.h indices.
var shapeToX11Cursor = [...]uint{
	0:  2,   // Arrow → XC_arrow
	1:  34,  // Crosshair → XC_crosshair
	2:  52,  // Fleur → XC_fleur
	3:  58,  // Hand1 → XC_hand1
	4:  60,  // Hand2 → XC_hand2
	5:  68,  // LeftPtr → XC_left_ptr
	6:  90,  // Plus → XC_plus
	7:  92,  // QuestionArrow → XC_question_arrow
	8:  108, // SBHDoubleArrow → XC_sb_h_double_arrow
	9:  116, // SBVDoubleArrow → XC_sb_v_double_arrow
	10: 120, // SizingAngle → XC_sizing
	11: 132, // TopLeftArrow → XC_top_left_arrow
	12: 150, // Watch → XC_watch
	13: 152, // XTerm → XC_xterm
	14: 14,  // BottomRightCorner → XC_bottom_right_corner
}

func (s *X11Display) SetCursorShape(w platform.WindowID, shape uint) {
	x11Shape := shape // default: pass through
	if shape < uint(len(shapeToX11Cursor)) {
		x11Shape = shapeToX11Cursor[shape]
	}
	s.dpy.DefineCursorFromFont(xlib.Window(w), x11Shape)
}

func (s *X11Display) UndefineCursor(w platform.WindowID) { s.dpy.UndefineCursor(xlib.Window(w)) }

func (s *X11Display) FreeCursor(cursor platform.CursorID) { s.dpy.FreeCursor(xlib.Cursor(cursor)) }

// --- PropertyManager ---

func (s *X11Display) InternAtom(name string, onlyIfExists bool) platform.AtomID {
	return platform.AtomID(s.dpy.InternAtom(name, onlyIfExists))
}

func (s *X11Display) GetAtomName(atom platform.AtomID) string {
	return s.dpy.GetAtomName(xlib.Atom(atom))
}

func (s *X11Display) SetWMProtocols(w platform.WindowID, protocols []platform.AtomID) int {
	xp := make([]xlib.Atom, len(protocols))
	for i, p := range protocols {
		xp[i] = xlib.Atom(p)
	}
	return s.dpy.SetWMProtocols(xlib.Window(w), xp)
}

func (s *X11Display) SetWMNormalHints(w platform.WindowID, hints *platform.SizeHints) {
	s.dpy.SetWMNormalHints(xlib.Window(w), &xlib.SizeHints{
		Flags:      hints.Flags,
		X:          hints.X,
		Y:          hints.Y,
		Width:      hints.Width,
		Height:     hints.Height,
		MinWidth:   hints.MinWidth,
		MinHeight:  hints.MinHeight,
		MaxWidth:   hints.MaxWidth,
		MaxHeight:  hints.MaxHeight,
		WidthInc:   hints.WidthInc,
		HeightInc:  hints.HeightInc,
		WinGravity: hints.WinGravity,
	})
}

func (s *X11Display) SetWMHints(w platform.WindowID, hints *platform.WMHints) {
	s.dpy.SetWMHints(xlib.Window(w), &xlib.WMHints{
		Flags:        hints.Flags,
		Input:        hints.Input,
		InitialState: hints.InitialState,
	})
}

func (s *X11Display) SetClassHint(w platform.WindowID, name, class string) {
	s.dpy.SetClassHint(xlib.Window(w), name, class)
}

func (s *X11Display) SetTransientForHint(w platform.WindowID, propWindow platform.WindowID) {
	s.dpy.SetTransientForHint(xlib.Window(w), xlib.Window(propWindow))
}

func (s *X11Display) SetInputFocus(w platform.WindowID, revertTo int, time platform.Timestamp) {
	s.dpy.SetInputFocus(xlib.Window(w), revertTo, xlib.Time(time))
}

func (s *X11Display) GetInputFocus() (platform.WindowID, int) {
	w, revert := s.dpy.GetInputFocus()
	return platform.WindowID(w), revert
}

func (s *X11Display) ChangeProperty(w platform.WindowID, prop, propType platform.AtomID, format int, mode int, data []byte, nelements int) {
	s.dpy.ChangeProperty(xlib.Window(w), xlib.Atom(prop), xlib.Atom(propType), format, mode, data, nelements)
}

func (s *X11Display) ChangePropertyString(w platform.WindowID, property, typ platform.AtomID, data string) {
	s.dpy.ChangePropertyString(xlib.Window(w), xlib.Atom(property), xlib.Atom(typ), data)
}

func (s *X11Display) ChangePropertyAtoms(w platform.WindowID, prop platform.AtomID, atoms []platform.AtomID) {
	xa := make([]xlib.Atom, len(atoms))
	for i, a := range atoms {
		xa[i] = xlib.Atom(a)
	}
	s.dpy.ChangePropertyAtoms(xlib.Window(w), xlib.Atom(prop), xa)
}

func (s *X11Display) GetWindowProperty(w platform.WindowID, property platform.AtomID, offset, length int64, delete bool) ([]byte, platform.AtomID, int) {
	data, atype, format := s.dpy.GetWindowProperty(xlib.Window(w), xlib.Atom(property), offset, length, delete)
	return data, platform.AtomID(atype), format
}

func (s *X11Display) DeleteProperty(w platform.WindowID, prop platform.AtomID) {
	s.dpy.DeleteProperty(xlib.Window(w), xlib.Atom(prop))
}

func (s *X11Display) SendEvent(w platform.WindowID, propagate bool, eventMask int64, ev *platform.RawEvent) {
	raw := ev.Data.(*xlib.RawEvent)
	s.dpy.SendEvent(xlib.Window(w), propagate, eventMask, raw)
}

func (s *X11Display) SendClientMessage(w, target platform.WindowID, msgType platform.AtomID, d0, d1, d2, d3, d4 int64) {
	s.dpy.SendClientMessage(xlib.Window(w), xlib.Window(target), xlib.Atom(msgType), d0, d1, d2, d3, d4)
}

func (s *X11Display) IconifyWindow(w platform.WindowID, screen int) {
	s.dpy.IconifyWindow(xlib.Window(w), screen)
}

func (s *X11Display) WithdrawWindow(w platform.WindowID, screen int) {
	s.dpy.WithdrawWindow(xlib.Window(w), screen)
}

func (s *X11Display) SetIconName(w platform.WindowID, name string) {
	s.dpy.SetIconName(xlib.Window(w), name)
}

// --- InputMethodManager ---

func (s *X11Display) InitIM(root platform.WindowID)  { s.dpy.InitIM(xlib.Window(root)) }
func (s *X11Display) HasIM() bool                    { return s.dpy.HasIM() }
func (s *X11Display) SetICFocus(w platform.WindowID) { s.dpy.SetICFocus(xlib.Window(w)) }
func (s *X11Display) UnsetICFocus()                  { s.dpy.UnsetICFocus() }

// Verify that X11Display implements platform.DisplayServer at compile time.
var _ platform.DisplayServer = (*X11Display)(nil)
