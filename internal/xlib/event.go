//go:build linux || freebsd || openbsd || netbsd

package xlib

/*
#include <X11/Xlib.h>
#include <X11/Xutil.h>
#include <string.h>
#include <locale.h>

// Helper to get event type from XEvent union.
static int xevent_type(XEvent *ev) { return ev->type; }

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
static int xevent_property_deleted(XEvent *ev) { return ev->xproperty.state == PropertyDelete; }

// Any event window accessor.
static Window xevent_any_window(XEvent *ev) { return ev->xany.window; }

// takigo_input carries the fields of a key, button, motion or crossing
// event, so parsing one costs a single cgo call instead of one per field.
typedef struct {
    Window window, root;
    Time time;
    int x, y, x_root, y_root;
    unsigned int state, detail;
} takigo_input;

static void xevent_input(XEvent *ev, takigo_input *o) {
    switch (ev->type) {
    case KeyPress: case KeyRelease: {
        XKeyEvent *k = &ev->xkey;
        o->window = k->window; o->root = k->root; o->time = k->time;
        o->x = k->x; o->y = k->y; o->x_root = k->x_root; o->y_root = k->y_root;
        o->state = k->state; o->detail = k->keycode;
        break;
    }
    case ButtonPress: case ButtonRelease: {
        XButtonEvent *b = &ev->xbutton;
        o->window = b->window; o->root = b->root; o->time = b->time;
        o->x = b->x; o->y = b->y; o->x_root = b->x_root; o->y_root = b->y_root;
        o->state = b->state; o->detail = b->button;
        break;
    }
    case MotionNotify: {
        XMotionEvent *m = &ev->xmotion;
        o->window = m->window; o->root = m->root; o->time = m->time;
        o->x = m->x; o->y = m->y; o->x_root = m->x_root; o->y_root = m->y_root;
        o->state = m->state; o->detail = 0;
        break;
    }
    case EnterNotify: case LeaveNotify: {
        XCrossingEvent *c = &ev->xcrossing;
        o->window = c->window; o->root = c->root; o->time = c->time;
        o->x = c->x; o->y = c->y; o->x_root = c->x_root; o->y_root = c->y_root;
        o->state = c->state; o->detail = c->detail;
        break;
    }
    }
}

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

// XFilterEvent wrapper. A MappingNotify is consumed here after refreshing
// Xlib's keyboard/modifier map, as Tk_HandleEvent does, so key events
// keep translating correctly after e.g. setxkbmap.
static int filter_event(XEvent *ev) {
    if (ev->type == MappingNotify) {
        XRefreshKeyboardMapping(&ev->xmapping);
        return 1;
    }
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

	return e.keyEvent(KeySym(ks), str)
}

// ParseKeyEventIM extracts key event data using Xutf8LookupString via XIM.
// This correctly handles non-Latin scripts (Cyrillic, etc.) with Caps Lock.
// Like TkpGetString it retries with a buffer of the reported size on
// XBufferOverflow and only trusts the string when the status says one was
// returned.
func (d *Display) ParseKeyEventIM(e *RawEvent) KeyEvent {
	buf := make([]C.char, 64)
	var ks C.KeySym
	var status C.int
	n := C.utf8_lookup_string(d.xic, &e.ev, &buf[0], C.int(len(buf)), &ks, &status)
	if status == C.XBufferOverflow && n > 0 {
		buf = make([]C.char, n+1)
		n = C.utf8_lookup_string(d.xic, &e.ev, &buf[0], C.int(len(buf)), &ks, &status)
	}

	str := ""
	if (status == C.XLookupChars || status == C.XLookupBoth) && n > 0 && int(n) <= len(buf) {
		str = C.GoStringN(&buf[0], n)
	}

	return e.keyEvent(KeySym(ks), str)
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
	in := e.input()
	return ButtonEvent{
		EventWindow: Window(in.window),
		X:           int(in.x),
		Y:           int(in.y),
		RootX:       int(in.x_root),
		RootY:       int(in.y_root),
		State:       uint(in.state),
		Button:      uint(in.detail),
		Time:        Time(in.time),
	}
}

// input reads a key, button, motion or crossing event in one cgo call.
func (e *RawEvent) input() C.takigo_input {
	var in C.takigo_input
	C.xevent_input(&e.ev, &in)
	return in
}

// keyEvent builds a KeyEvent from the event and its looked-up keysym.
func (e *RawEvent) keyEvent(ks KeySym, str string) KeyEvent {
	in := e.input()
	return KeyEvent{
		EventWindow: Window(in.window),
		RootWindow:  Window(in.root),
		X:           int(in.x),
		Y:           int(in.y),
		RootX:       int(in.x_root),
		RootY:       int(in.y_root),
		State:       uint(in.state),
		KeyCode:     uint(in.detail),
		KeySym:      ks,
		Str:         str,
		Time:        Time(in.time),
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
	in := e.input()
	return MotionEvent{
		EventWindow: Window(in.window),
		X:           int(in.x),
		Y:           int(in.y),
		RootX:       int(in.x_root),
		RootY:       int(in.y_root),
		State:       uint(in.state),
		Time:        Time(in.time),
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
	in := e.input()
	return CrossingEvent{
		EventWindow: Window(in.window),
		X:           int(in.x),
		Y:           int(in.y),
		State:       uint(in.state),
		Time:        Time(in.time),
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
	Deleted     bool // PropertyDelete rather than PropertyNewValue
}

// ParsePropertyEvent extracts PropertyNotify event data.
func (e *RawEvent) ParsePropertyEvent() PropertyEventData {
	return PropertyEventData{
		EventWindow: Window(C.xevent_property_window(&e.ev)),
		Atom:        Atom(C.xevent_property_atom(&e.ev)),
		Deleted:     C.xevent_property_deleted(&e.ev) != 0,
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
