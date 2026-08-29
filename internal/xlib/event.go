//go:build linux || freebsd || openbsd || netbsd

package xlib

/*
#include <X11/Xlib.h>
#include <X11/Xutil.h>
#include <string.h>
#include <locale.h>

// Helper to get event type from XEvent union.
static int xevent_type(XEvent *ev) { return ev->type; }

// Key event accessors.
static Window xevent_key_window(XEvent *ev) { return ev->xkey.window; }
static Window xevent_key_root(XEvent *ev) { return ev->xkey.root; }
static unsigned int xevent_key_state(XEvent *ev) { return ev->xkey.state; }
static unsigned int xevent_key_keycode(XEvent *ev) { return ev->xkey.keycode; }
static int xevent_key_x(XEvent *ev) { return ev->xkey.x; }
static int xevent_key_y(XEvent *ev) { return ev->xkey.y; }
static int xevent_key_x_root(XEvent *ev) { return ev->xkey.x_root; }
static int xevent_key_y_root(XEvent *ev) { return ev->xkey.y_root; }
static Time xevent_key_time(XEvent *ev) { return ev->xkey.time; }

// Button event accessors.
static Window xevent_button_window(XEvent *ev) { return ev->xbutton.window; }
static unsigned int xevent_button_button(XEvent *ev) { return ev->xbutton.button; }
static unsigned int xevent_button_state(XEvent *ev) { return ev->xbutton.state; }
static int xevent_button_x(XEvent *ev) { return ev->xbutton.x; }
static int xevent_button_y(XEvent *ev) { return ev->xbutton.y; }
static int xevent_button_x_root(XEvent *ev) { return ev->xbutton.x_root; }
static int xevent_button_y_root(XEvent *ev) { return ev->xbutton.y_root; }
static Time xevent_button_time(XEvent *ev) { return ev->xbutton.time; }

// Motion event accessors.
static Window xevent_motion_window(XEvent *ev) { return ev->xmotion.window; }
static unsigned int xevent_motion_state(XEvent *ev) { return ev->xmotion.state; }
static int xevent_motion_x(XEvent *ev) { return ev->xmotion.x; }
static int xevent_motion_y(XEvent *ev) { return ev->xmotion.y; }
static int xevent_motion_x_root(XEvent *ev) { return ev->xmotion.x_root; }
static int xevent_motion_y_root(XEvent *ev) { return ev->xmotion.y_root; }
static Time xevent_motion_time(XEvent *ev) { return ev->xmotion.time; }

// Expose event accessors.
static Window xevent_expose_window(XEvent *ev) { return ev->xexpose.window; }
static int xevent_expose_x(XEvent *ev) { return ev->xexpose.x; }
static int xevent_expose_y(XEvent *ev) { return ev->xexpose.y; }
static int xevent_expose_width(XEvent *ev) { return ev->xexpose.width; }
static int xevent_expose_height(XEvent *ev) { return ev->xexpose.height; }
static int xevent_expose_count(XEvent *ev) { return ev->xexpose.count; }

// Configure event accessors.
static Window xevent_configure_window(XEvent *ev) { return ev->xconfigure.window; }
static int xevent_configure_x(XEvent *ev) { return ev->xconfigure.x; }
static int xevent_configure_y(XEvent *ev) { return ev->xconfigure.y; }
static int xevent_configure_width(XEvent *ev) { return ev->xconfigure.width; }
static int xevent_configure_height(XEvent *ev) { return ev->xconfigure.height; }

// Client message accessors.
static Window xevent_client_window(XEvent *ev) { return ev->xclient.window; }
static Atom xevent_client_message_type(XEvent *ev) { return ev->xclient.message_type; }
static int xevent_client_format(XEvent *ev) { return ev->xclient.format; }
static long xevent_client_data_l(XEvent *ev, int i) { return ev->xclient.data.l[i]; }

// Crossing event accessors.
static Window xevent_crossing_window(XEvent *ev) { return ev->xcrossing.window; }
static int xevent_crossing_x(XEvent *ev) { return ev->xcrossing.x; }
static int xevent_crossing_y(XEvent *ev) { return ev->xcrossing.y; }
static unsigned int xevent_crossing_state(XEvent *ev) { return ev->xcrossing.state; }
static Time xevent_crossing_time(XEvent *ev) { return ev->xcrossing.time; }

// Focus event accessors.
static Window xevent_focus_window(XEvent *ev) { return ev->xfocus.window; }
static int    xevent_focus_mode(XEvent *ev)   { return ev->xfocus.mode; }
static int    xevent_focus_detail(XEvent *ev) { return ev->xfocus.detail; }

// Destroy event accessor.
static Window xevent_destroy_window(XEvent *ev) { return ev->xdestroywindow.window; }

// Map/Unmap event accessors.
static Window xevent_map_window(XEvent *ev) { return ev->xmap.window; }
static Window xevent_unmap_window(XEvent *ev) { return ev->xunmap.window; }

// Selection request event accessors.
static Window xevent_selreq_owner(XEvent *ev)    { return ev->xselectionrequest.owner; }
static Window xevent_selreq_requestor(XEvent *ev){ return ev->xselectionrequest.requestor; }
static Atom   xevent_selreq_selection(XEvent *ev){ return ev->xselectionrequest.selection; }
static Atom   xevent_selreq_target(XEvent *ev)   { return ev->xselectionrequest.target; }
static Atom   xevent_selreq_property(XEvent *ev) { return ev->xselectionrequest.property; }
static Time   xevent_selreq_time(XEvent *ev)     { return ev->xselectionrequest.time; }

// Selection clear event accessors.
static Window xevent_selclr_window(XEvent *ev)   { return ev->xselectionclear.window; }
static Atom   xevent_selclr_selection(XEvent *ev){ return ev->xselectionclear.selection; }

// Selection notify event accessors (sent to the requestor).
static Window xevent_selnot_requestor(XEvent *ev){ return ev->xselection.requestor; }
static Atom   xevent_selnot_selection(XEvent *ev){ return ev->xselection.selection; }
static Atom   xevent_selnot_target(XEvent *ev)   { return ev->xselection.target; }
static Atom   xevent_selnot_property(XEvent *ev) { return ev->xselection.property; }
static Time   xevent_selnot_time(XEvent *ev)     { return ev->xselection.time; }

// Property event accessors.
static Window xevent_property_window(XEvent *ev) { return ev->xproperty.window; }
static Atom xevent_property_atom(XEvent *ev) { return ev->xproperty.atom; }

// Any event window accessor.
static Window xevent_any_window(XEvent *ev) { return ev->xany.window; }

// XLookupString wrapper.
static int lookup_string(XEvent *ev, char *buf, int buflen, KeySym *ks) {
    return XLookupString(&ev->xkey, buf, buflen, ks, NULL);
}

// Xutf8LookupString wrapper.
static int utf8_lookup_string(XIC ic, XEvent *ev, char *buf, int buflen, KeySym *ks, int *status_out) {
    Status status;
    int n = Xutf8LookupString(ic, &ev->xkey, buf, buflen, ks, &status);
    *status_out = (int)status;
    return n;
}

// XOpenIM wrapper. Sets locale first (required for Xutf8LookupString).
static XIM open_im(Display *dpy) {
    setlocale(LC_ALL, "");
    if (!XSupportsLocale()) {
        return NULL;
    }
    XSetLocaleModifiers("");
    return XOpenIM(dpy, NULL, NULL, NULL);
}

// XCreateIC wrapper.
static XIC create_ic(XIM im, Window w) {
    return XCreateIC(im,
        XNInputStyle, XIMPreeditNothing | XIMStatusNothing,
        XNClientWindow, w,
        XNFocusWindow, w,
        NULL);
}

// XSetICFocus / XUnsetICFocus wrappers.
static void set_ic_focus(XIC ic) {
    if (ic) XSetICFocus(ic);
}
static void unset_ic_focus(XIC ic) {
    if (ic) XUnsetICFocus(ic);
}
// set_ic_focus_window updates the XIC's XNFocusWindow and calls XSetICFocus.
// This tells the IM which actual window has focus (important for Xutf8LookupString
// to correctly process events from the focused widget window, not the root window).
static void set_ic_focus_window(XIC ic, Window w) {
    if (!ic) return;
    XSetICValues(ic, XNFocusWindow, w, NULL);
    XSetICFocus(ic);
}

// XFilterEvent wrapper.
static int filter_event(XEvent *ev) {
    return XFilterEvent(ev, None);
}
*/
import "C"

// RawEvent wraps a C XEvent. It is opaque to Go code;
// use the Parse* methods to extract typed event data.
type RawEvent struct {
	ev C.XEvent
}

// Type returns the X11 event type code.
func (e *RawEvent) Type() int {
	return int(C.xevent_type(&e.ev))
}

// Window returns the window associated with the event.
func (e *RawEvent) Window() Window {
	return Window(C.xevent_any_window(&e.ev))
}

// InitIM initializes the X Input Method for the display. Call after OpenDisplay.
// It's safe to call even if XIM is not available (silently fails).
func (d *Display) InitIM(root Window) {
	d.xim = C.open_im(d.ptr)
	if d.xim != nil {
		d.xic = C.create_ic(d.xim, C.Window(root))
	}
}

// HasIM returns true if XIM/XIC was successfully initialized.
func (d *Display) HasIM() bool {
	return d.xic != nil
}

// SetICFocus notifies the input method that input focus has moved to the
// given window. Must be called on FocusIn events.
// Passing the actual focused widget window (not just the root) allows
// Xutf8LookupString to correctly process key events from that window.
func (d *Display) SetICFocus(w Window) {
	if d.xic != nil {
		C.set_ic_focus_window(d.xic, C.Window(w))
	}
}

// UnsetICFocus notifies the input method that the client window has lost
// focus. Must be called on FocusOut events.
func (d *Display) UnsetICFocus() {
	if d.xic != nil {
		C.unset_ic_focus(d.xic)
	}
}

// FilterEvent returns true if the event was consumed by the input method.
func (e *RawEvent) FilterEvent() bool {
	return C.filter_event(&e.ev) != 0
}

// NextEvent blocks until the next event and returns it.
func (d *Display) NextEvent() *RawEvent {
	var ev RawEvent
	C.XNextEvent(d.ptr, &ev.ev)
	return &ev
}

// PeekEvent returns the next event without removing it from the queue.
func (d *Display) PeekEvent() *RawEvent {
	var ev RawEvent
	C.XPeekEvent(d.ptr, &ev.ev)
	return &ev
}

// KeyEvent holds parsed key press/release data.
type KeyEvent struct {
	EventWindow  Window
	RootWindow   Window
	X, Y         int
	RootX, RootY int
	State        uint
	KeyCode      uint
	KeySym       KeySym
	Str          string
	Time         Time
}

// ParseKeyEvent extracts key event data using XLookupString (legacy fallback).
func (e *RawEvent) ParseKeyEvent() KeyEvent {
	var buf [64]C.char
	var ks C.KeySym
	n := C.lookup_string(&e.ev, &buf[0], 64, &ks)

	str := ""
	if n > 0 {
		str = C.GoStringN(&buf[0], n)
	}

	return KeyEvent{
		EventWindow: Window(C.xevent_key_window(&e.ev)),
		RootWindow:  Window(C.xevent_key_root(&e.ev)),
		X:           int(C.xevent_key_x(&e.ev)),
		Y:           int(C.xevent_key_y(&e.ev)),
		RootX:       int(C.xevent_key_x_root(&e.ev)),
		RootY:       int(C.xevent_key_y_root(&e.ev)),
		State:       uint(C.xevent_key_state(&e.ev)),
		KeyCode:     uint(C.xevent_key_keycode(&e.ev)),
		KeySym:      KeySym(ks),
		Str:         str,
		Time:        Time(C.xevent_key_time(&e.ev)),
	}
}

// ParseKeyEventIM extracts key event data using Xutf8LookupString via XIM.
// This correctly handles non-Latin scripts (Cyrillic, etc.) with Caps Lock.
func (d *Display) ParseKeyEventIM(e *RawEvent) KeyEvent {
	var buf [64]C.char
	var ks C.KeySym
	var status C.int
	n := C.utf8_lookup_string(d.xic, &e.ev, &buf[0], 64, &ks, &status)

	str := ""
	if n > 0 {
		str = C.GoStringN(&buf[0], n)
	}

	return KeyEvent{
		EventWindow: Window(C.xevent_key_window(&e.ev)),
		RootWindow:  Window(C.xevent_key_root(&e.ev)),
		X:           int(C.xevent_key_x(&e.ev)),
		Y:           int(C.xevent_key_y(&e.ev)),
		RootX:       int(C.xevent_key_x_root(&e.ev)),
		RootY:       int(C.xevent_key_y_root(&e.ev)),
		State:       uint(C.xevent_key_state(&e.ev)),
		KeyCode:     uint(C.xevent_key_keycode(&e.ev)),
		KeySym:      KeySym(ks),
		Str:         str,
		Time:        Time(C.xevent_key_time(&e.ev)),
	}
}

// ButtonEvent holds parsed button press/release data.
type ButtonEvent struct {
	EventWindow  Window
	X, Y         int
	RootX, RootY int
	State        uint
	Button       uint
	Time         Time
}

// ParseButtonEvent extracts button event data.
func (e *RawEvent) ParseButtonEvent() ButtonEvent {
	return ButtonEvent{
		EventWindow: Window(C.xevent_button_window(&e.ev)),
		X:           int(C.xevent_button_x(&e.ev)),
		Y:           int(C.xevent_button_y(&e.ev)),
		RootX:       int(C.xevent_button_x_root(&e.ev)),
		RootY:       int(C.xevent_button_y_root(&e.ev)),
		State:       uint(C.xevent_button_state(&e.ev)),
		Button:      uint(C.xevent_button_button(&e.ev)),
		Time:        Time(C.xevent_button_time(&e.ev)),
	}
}

// MotionEvent holds parsed motion (pointer move) data.
type MotionEvent struct {
	EventWindow  Window
	X, Y         int
	RootX, RootY int
	State        uint
	Time         Time
}

// ParseMotionEvent extracts motion event data.
func (e *RawEvent) ParseMotionEvent() MotionEvent {
	return MotionEvent{
		EventWindow: Window(C.xevent_motion_window(&e.ev)),
		X:           int(C.xevent_motion_x(&e.ev)),
		Y:           int(C.xevent_motion_y(&e.ev)),
		RootX:       int(C.xevent_motion_x_root(&e.ev)),
		RootY:       int(C.xevent_motion_y_root(&e.ev)),
		State:       uint(C.xevent_motion_state(&e.ev)),
		Time:        Time(C.xevent_motion_time(&e.ev)),
	}
}

// ExposeEvent holds parsed expose event data.
type ExposeEvent struct {
	EventWindow   Window
	X, Y          int
	Width, Height int
	Count         int
}

// ParseExposeEvent extracts expose event data.
func (e *RawEvent) ParseExposeEvent() ExposeEvent {
	return ExposeEvent{
		EventWindow: Window(C.xevent_expose_window(&e.ev)),
		X:           int(C.xevent_expose_x(&e.ev)),
		Y:           int(C.xevent_expose_y(&e.ev)),
		Width:       int(C.xevent_expose_width(&e.ev)),
		Height:      int(C.xevent_expose_height(&e.ev)),
		Count:       int(C.xevent_expose_count(&e.ev)),
	}
}

// ConfigureEvent holds parsed configure (resize/move) event data.
type ConfigureEvent struct {
	EventWindow   Window
	X, Y          int
	Width, Height int
}

// ParseConfigureEvent extracts configure event data.
func (e *RawEvent) ParseConfigureEvent() ConfigureEvent {
	return ConfigureEvent{
		EventWindow: Window(C.xevent_configure_window(&e.ev)),
		X:           int(C.xevent_configure_x(&e.ev)),
		Y:           int(C.xevent_configure_y(&e.ev)),
		Width:       int(C.xevent_configure_width(&e.ev)),
		Height:      int(C.xevent_configure_height(&e.ev)),
	}
}

// ClientMessageEvent holds parsed client message data.
type ClientMessageEvent struct {
	EventWindow Window
	MessageType Atom
	Format      int
	Data        [5]int64
}

// ParseClientMessageEvent extracts client message event data.
func (e *RawEvent) ParseClientMessageEvent() ClientMessageEvent {
	cm := ClientMessageEvent{
		EventWindow: Window(C.xevent_client_window(&e.ev)),
		MessageType: Atom(C.xevent_client_message_type(&e.ev)),
		Format:      int(C.xevent_client_format(&e.ev)),
	}
	for i := 0; i < 5; i++ {
		cm.Data[i] = int64(C.xevent_client_data_l(&e.ev, C.int(i)))
	}
	return cm
}

// CrossingEvent holds parsed enter/leave event data.
type CrossingEvent struct {
	EventWindow Window
	X, Y        int
	State       uint
	Time        Time
}

// ParseCrossingEvent extracts crossing event data.
func (e *RawEvent) ParseCrossingEvent() CrossingEvent {
	return CrossingEvent{
		EventWindow: Window(C.xevent_crossing_window(&e.ev)),
		X:           int(C.xevent_crossing_x(&e.ev)),
		Y:           int(C.xevent_crossing_y(&e.ev)),
		State:       uint(C.xevent_crossing_state(&e.ev)),
		Time:        Time(C.xevent_crossing_time(&e.ev)),
	}
}

// DestroyEvent holds the window ID from a DestroyNotify.
type DestroyEvent struct {
	EventWindow Window
}

// ParseDestroyEvent extracts destroy event data.
func (e *RawEvent) ParseDestroyEvent() DestroyEvent {
	return DestroyEvent{
		EventWindow: Window(C.xevent_destroy_window(&e.ev)),
	}
}

// FocusEvent holds focus in/out event data.
type FocusEvent struct {
	EventWindow Window
	Mode        int // NotifyNormal=0, NotifyGrab=1, NotifyUngrab=2, NotifyWhileGrabbed=3
	Detail      int // NotifyAncestor=0, NotifyVirtual=1, NotifyInferior=2, NotifyNonlinear=3, ...
}

// ParseFocusEvent extracts focus event data.
func (e *RawEvent) ParseFocusEvent() FocusEvent {
	return FocusEvent{
		EventWindow: Window(C.xevent_focus_window(&e.ev)),
		Mode:        int(C.xevent_focus_mode(&e.ev)),
		Detail:      int(C.xevent_focus_detail(&e.ev)),
	}
}

// PropertyEventData holds data from a PropertyNotify event.
type PropertyEventData struct {
	EventWindow Window
	Atom        Atom
}

// ParsePropertyEvent extracts PropertyNotify event data.
func (e *RawEvent) ParsePropertyEvent() PropertyEventData {
	return PropertyEventData{
		EventWindow: Window(C.xevent_property_window(&e.ev)),
		Atom:        Atom(C.xevent_property_atom(&e.ev)),
	}
}

// SelectionRequestEvent holds data from a SelectionRequest event.
type SelectionRequestEvent struct {
	Owner     Window
	Requestor Window
	Selection Atom
	Target    Atom
	Property  Atom
	Time      Time
}

// ParseSelectionRequestEvent extracts SelectionRequest event data.
func (e *RawEvent) ParseSelectionRequestEvent() SelectionRequestEvent {
	return SelectionRequestEvent{
		Owner:     Window(C.xevent_selreq_owner(&e.ev)),
		Requestor: Window(C.xevent_selreq_requestor(&e.ev)),
		Selection: Atom(C.xevent_selreq_selection(&e.ev)),
		Target:    Atom(C.xevent_selreq_target(&e.ev)),
		Property:  Atom(C.xevent_selreq_property(&e.ev)),
		Time:      Time(C.xevent_selreq_time(&e.ev)),
	}
}

// SelectionClearEvent holds data from a SelectionClear event.
type SelectionClearEvent struct {
	Window    Window
	Selection Atom
}

// ParseSelectionClearEvent extracts SelectionClear event data.
func (e *RawEvent) ParseSelectionClearEvent() SelectionClearEvent {
	return SelectionClearEvent{
		Window:    Window(C.xevent_selclr_window(&e.ev)),
		Selection: Atom(C.xevent_selclr_selection(&e.ev)),
	}
}

// SelectionNotifyEvent holds data from a SelectionNotify event.
type SelectionNotifyEvent struct {
	Requestor Window
	Selection Atom
	Target    Atom
	Property  Atom
	Time      Time
}

// ParseSelectionNotifyEvent extracts SelectionNotify event data.
func (e *RawEvent) ParseSelectionNotifyEvent() SelectionNotifyEvent {
	return SelectionNotifyEvent{
		Requestor: Window(C.xevent_selnot_requestor(&e.ev)),
		Selection: Atom(C.xevent_selnot_selection(&e.ev)),
		Target:    Atom(C.xevent_selnot_target(&e.ev)),
		Property:  Atom(C.xevent_selnot_property(&e.ev)),
		Time:      Time(C.xevent_selnot_time(&e.ev)),
	}
}
