package event

import "github.com/msorc/takigo/internal/xlib"

// Event is the unified event struct delivered to handlers.
// Only fields relevant to the event type are populated.
type Event struct {
	Type    Type
	Window  xlib.Window // target window

	// Key events
	KeySym  xlib.KeySym
	KeyCode uint
	Str     string // text from key event

	// Pointer events (button, motion, crossing)
	X, Y         int // position relative to event window
	RootX, RootY int // position relative to root window
	Button       uint
	State        uint // modifier mask

	// Expose
	ExposeX, ExposeY          int
	ExposeWidth, ExposeHeight int
	ExposeCount               int

	// Configure
	ConfigX, ConfigY          int
	ConfigWidth, ConfigHeight int

	// Client message
	MessageType xlib.Atom
	MessageData [5]int64

	// Timestamp (when available)
	Time xlib.Time
}

// FromRawEvent converts an xlib.RawEvent to a typed Event.
func FromRawEvent(raw *xlib.RawEvent) Event {
	ev := Event{
		Window: raw.Window(),
	}

	switch raw.Type() {
	case xlib.KeyPress, xlib.KeyRelease:
		k := raw.ParseKeyEvent()
		if raw.Type() == xlib.KeyPress {
			ev.Type = KeyPressType
		} else {
			ev.Type = KeyReleaseType
		}
		ev.Window = k.EventWindow
		ev.KeySym = k.KeySym
		ev.KeyCode = k.KeyCode
		ev.Str = k.Str
		ev.X = k.X
		ev.Y = k.Y
		ev.RootX = k.RootX
		ev.RootY = k.RootY
		ev.State = k.State
		ev.Time = k.Time

	case xlib.ButtonPress, xlib.ButtonRelease:
		b := raw.ParseButtonEvent()
		if raw.Type() == xlib.ButtonPress {
			ev.Type = ButtonPressType
		} else {
			ev.Type = ButtonReleaseType
		}
		ev.Window = b.EventWindow
		ev.X = b.X
		ev.Y = b.Y
		ev.RootX = b.RootX
		ev.RootY = b.RootY
		ev.Button = b.Button
		ev.State = b.State
		ev.Time = b.Time

	case xlib.MotionNotify:
		m := raw.ParseMotionEvent()
		ev.Type = MotionType
		ev.Window = m.EventWindow
		ev.X = m.X
		ev.Y = m.Y
		ev.RootX = m.RootX
		ev.RootY = m.RootY
		ev.State = m.State
		ev.Time = m.Time

	case xlib.EnterNotify, xlib.LeaveNotify:
		c := raw.ParseCrossingEvent()
		if raw.Type() == xlib.EnterNotify {
			ev.Type = EnterType
		} else {
			ev.Type = LeaveType
		}
		ev.Window = c.EventWindow
		ev.X = c.X
		ev.Y = c.Y
		ev.State = c.State
		ev.Time = c.Time

	case xlib.FocusIn:
		f := raw.ParseFocusEvent()
		ev.Type = FocusInType
		ev.Window = f.EventWindow

	case xlib.FocusOut:
		f := raw.ParseFocusEvent()
		ev.Type = FocusOutType
		ev.Window = f.EventWindow

	case xlib.Expose:
		e := raw.ParseExposeEvent()
		ev.Type = ExposeType
		ev.Window = e.EventWindow
		ev.ExposeX = e.X
		ev.ExposeY = e.Y
		ev.ExposeWidth = e.Width
		ev.ExposeHeight = e.Height
		ev.ExposeCount = e.Count

	case xlib.ConfigureNotify:
		cfg := raw.ParseConfigureEvent()
		ev.Type = ConfigureType
		ev.Window = cfg.EventWindow
		ev.ConfigX = cfg.X
		ev.ConfigY = cfg.Y
		ev.ConfigWidth = cfg.Width
		ev.ConfigHeight = cfg.Height

	case xlib.DestroyNotify:
		d := raw.ParseDestroyEvent()
		ev.Type = DestroyType
		ev.Window = d.EventWindow

	case xlib.MapNotify:
		ev.Type = MapType

	case xlib.UnmapNotify:
		ev.Type = UnmapType

	case xlib.ClientMessage:
		cm := raw.ParseClientMessageEvent()
		ev.Type = ClientMessageType
		ev.Window = cm.EventWindow
		ev.MessageType = cm.MessageType
		ev.MessageData = cm.Data
	}

	return ev
}
