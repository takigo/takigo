package xlib

/*
#include <X11/Xlib.h>
#include <X11/Xutil.h>
#include <string.h>

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

// Destroy event accessor.
static Window xevent_destroy_window(XEvent *ev) { return ev->xdestroywindow.window; }

// Map/Unmap event accessors.
static Window xevent_map_window(XEvent *ev) { return ev->xmap.window; }
static Window xevent_unmap_window(XEvent *ev) { return ev->xunmap.window; }

// Property event accessors.
static Window xevent_property_window(XEvent *ev) { return ev->xproperty.window; }
static Atom xevent_property_atom(XEvent *ev) { return ev->xproperty.atom; }

// Any event window accessor.
static Window xevent_any_window(XEvent *ev) { return ev->xany.window; }

// XLookupString wrapper.
static int lookup_string(XEvent *ev, char *buf, int buflen, KeySym *ks) {
    return XLookupString(&ev->xkey, buf, buflen, ks, NULL);
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
	EventWindow Window
	RootWindow  Window
	X, Y        int
	RootX, RootY int
	State       uint
	KeyCode     uint
	KeySym      KeySym
	Str         string
	Time        Time
}

// ParseKeyEvent extracts key event data.
func (e *RawEvent) ParseKeyEvent() KeyEvent {
	var buf [32]C.char
	var ks C.KeySym
	n := C.lookup_string(&e.ev, &buf[0], 32, &ks)

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
	EventWindow    Window
	X, Y           int
	Width, Height  int
	Count          int
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
}

// ParseFocusEvent extracts focus event data.
func (e *RawEvent) ParseFocusEvent() FocusEvent {
	return FocusEvent{
		EventWindow: Window(C.xevent_focus_window(&e.ev)),
	}
}
