package platform

// RawEvent is a platform-neutral opaque wrapper around a raw event.
// Each backend provides its own implementation.
type RawEvent struct {
	// Backend-specific data stored as interface{}.
	// The X11 backend stores *xlib.RawEvent here.
	Data any

	// Parsed fields available without backend-specific access.
	EventType   int
	EventWindow WindowID
}

// KeyEvent holds parsed key press/release data.
type KeyEvent struct {
	EventWindow  WindowID
	RootWindow   WindowID
	X, Y         int
	RootX, RootY int
	State        uint
	KeyCode      uint
	KeySym       KeySym
	Str          string
	Time         Timestamp
}

// ButtonEvent holds parsed button press/release data.
type ButtonEvent struct {
	EventWindow  WindowID
	X, Y         int
	RootX, RootY int
	State        uint
	Button       uint
	Time         Timestamp
}

// MotionEvent holds parsed motion event data.
type MotionEvent struct {
	EventWindow  WindowID
	X, Y         int
	RootX, RootY int
	State        uint
	Time         Timestamp
}

// ExposeEvent holds parsed expose event data.
type ExposeEvent struct {
	EventWindow   WindowID
	X, Y          int
	Width, Height int
	Count         int
}

// ConfigureEvent holds parsed configure event data.
type ConfigureEvent struct {
	EventWindow   WindowID
	X, Y          int
	Width, Height int
}

// ClientMessageEvent holds parsed client message data.
type ClientMessageEvent struct {
	EventWindow WindowID
	MessageType AtomID
	Format      int
	Data        [5]int64
}

// CrossingEvent holds parsed enter/leave event data.
type CrossingEvent struct {
	EventWindow WindowID
	X, Y        int
	State       uint
	Time        Timestamp
}

// DestroyEvent holds parsed destroy event data.
type DestroyEvent struct {
	EventWindow WindowID
}

// FocusEvent holds parsed focus event data.
type FocusEvent struct {
	EventWindow WindowID
}

// PropertyEvent holds parsed property event data.
type PropertyEvent struct {
	EventWindow WindowID
	Atom        AtomID
}

// X11 event type constants. Platform backends use these for EventType.
const (
	KeyPressEvent         = 2
	KeyReleaseEvent       = 3
	ButtonPressEvent      = 4
	ButtonReleaseEvent    = 5
	MotionNotifyEvent     = 6
	EnterNotifyEvent      = 7
	LeaveNotifyEvent      = 8
	FocusInEvent          = 9
	FocusOutEvent         = 10
	ExposeEvent_          = 12
	DestroyNotifyEvent    = 17
	UnmapNotifyEvent      = 18
	MapNotifyEvent        = 19
	MapRequestEvent       = 20
	ReparentNotifyEvent   = 21
	ConfigureNotifyEvent  = 22
	ConfigureRequestEvent = 23
	GravityNotifyEvent    = 24
	ResizeRequestEvent    = 25
	CirculateNotifyEvent  = 26
	PropertyNotifyEvent   = 28
	SelectionClearEvent   = 29
	SelectionRequestEvent = 30
	SelectionNotifyEvent  = 31
	ColormapNotifyEvent   = 32
	ClientMessageEvent_   = 33
	MappingNotifyEvent    = 34
)

// EventParser parses platform-specific raw events into typed events.
// Each backend implements this interface.
type EventParser interface {
	// ParseKeyEvent parses a raw event as a key event.
	ParseKeyEvent(ev *RawEvent) KeyEvent

	// ParseKeyEventIM parses a key event using input method.
	ParseKeyEventIM(ev *RawEvent) KeyEvent

	// ParseButtonEvent parses a raw event as a button event.
	ParseButtonEvent(ev *RawEvent) ButtonEvent

	// ParseMotionEvent parses a raw event as a motion event.
	ParseMotionEvent(ev *RawEvent) MotionEvent

	// ParseExposeEvent parses a raw event as an expose event.
	ParseExposeEvent(ev *RawEvent) ExposeEvent

	// ParseConfigureEvent parses a raw event as a configure event.
	ParseConfigureEvent(ev *RawEvent) ConfigureEvent

	// ParseClientMessageEvent parses a raw event as a client message.
	ParseClientMessageEvent(ev *RawEvent) ClientMessageEvent

	// ParseCrossingEvent parses a raw event as a crossing event.
	ParseCrossingEvent(ev *RawEvent) CrossingEvent

	// ParseDestroyEvent parses a raw event as a destroy event.
	ParseDestroyEvent(ev *RawEvent) DestroyEvent

	// ParseFocusEvent parses a raw event as a focus event.
	ParseFocusEvent(ev *RawEvent) FocusEvent

	// ParsePropertyEvent parses a raw event as a property event.
	ParsePropertyEvent(ev *RawEvent) PropertyEvent
}
