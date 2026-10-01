package event

import "github.com/msorc/takigo/platform"

// Event is the unified event struct delivered to handlers.
// Only fields relevant to the event type are populated.
type Event struct {
	Type   Type
	Window platform.WindowID // target window

	// Key events
	KeySym  platform.KeySym
	KeyCode uint
	Str     string // text from key event

	// Pointer events (button, motion, crossing)
	X, Y         int // position relative to event window
	RootX, RootY int // position relative to root window
	Button       uint
	State        uint // modifier mask
	// Delta is a MouseWheel event's distance: 120 per notch, positive up
	// (or left, with ShiftMask in State). High-resolution wheels send
	// smaller steps; see WheelAccumulator.
	Delta int

	// Expose
	ExposeX, ExposeY          int
	ExposeWidth, ExposeHeight int
	ExposeCount               int

	// Configure
	ConfigX, ConfigY          int
	ConfigWidth, ConfigHeight int

	// Client message
	MessageType platform.AtomID
	MessageData [5]int64

	// Focus events
	FocusMode   int // platform.FocusModeNormal etc.
	FocusDetail int // platform.FocusDetailInferior etc.; the crossing detail for Enter/Leave
	// Handled is set by a class handler that consumes the event, like a
	// Tk binding script ending in break (Text's <Tab>), so the global
	// handlers after it (focus traversal) leave it alone.
	Handled bool

	// Property events
	Atom platform.AtomID

	// Virtual events: the name, without the << >>.
	Name string

	// Timestamp (when available)
	Time platform.Timestamp
}

// FromRawEvent converts a platform.RawEvent to a typed Event.
func FromRawEvent(raw *platform.RawEvent, parser platform.EventParser) Event {
	return FromRawEventIM(raw, parser, false)
}

// FromRawEventIM converts a platform.RawEvent to a typed Event, using XIM for
// key events when hasIM is true.
func FromRawEventIM(raw *platform.RawEvent, parser platform.EventParser, hasIM bool) Event {
	ev := Event{
		Window: raw.EventWindow,
	}

	switch raw.EventType {
	case platform.KeyPressEvent, platform.KeyReleaseEvent:
		var k platform.KeyEvent
		if hasIM {
			k = parser.ParseKeyEventIM(raw)
		} else {
			k = parser.ParseKeyEvent(raw)
		}
		if raw.EventType == platform.KeyPressEvent {
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

	case platform.ButtonPressEvent, platform.ButtonReleaseEvent:
		b := parser.ParseButtonEvent(raw)
		if raw.EventType == platform.ButtonPressEvent {
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
		// TkQueueWindowEvent (tkEvent.c): buttons 4-7 are the wheels.
		if b.Button >= 4 && b.Button <= 7 {
			if ev.Type == ButtonReleaseType {
				ev.Type = 0
				break
			}
			ev.Type = MouseWheelType
			ev.Delta = b.WheelDelta
			if ev.Delta == 0 {
				ev.Delta = 120
				if b.Button&1 != 0 {
					ev.Delta = -120
				}
			}
			if b.Button > 5 {
				ev.State |= platform.ShiftMask
			}
			ev.Button = 0
		}

	case platform.MotionNotifyEvent:
		m := parser.ParseMotionEvent(raw)
		ev.Type = MotionType
		ev.Window = m.EventWindow
		ev.X = m.X
		ev.Y = m.Y
		ev.RootX = m.RootX
		ev.RootY = m.RootY
		ev.State = m.State
		ev.Time = m.Time

	case platform.EnterNotifyEvent, platform.LeaveNotifyEvent:
		c := parser.ParseCrossingEvent(raw)
		if raw.EventType == platform.EnterNotifyEvent {
			ev.Type = EnterType
		} else {
			ev.Type = LeaveType
		}
		ev.Window = c.EventWindow
		ev.X = c.X
		ev.Y = c.Y
		ev.State = c.State
		ev.Time = c.Time
		ev.FocusDetail = c.Detail

	case platform.FocusInEvent:
		f := parser.ParseFocusEvent(raw)
		ev.Type = FocusInType
		ev.Window = f.EventWindow
		ev.FocusMode = f.Mode
		ev.FocusDetail = f.Detail

	case platform.FocusOutEvent:
		f := parser.ParseFocusEvent(raw)
		ev.Type = FocusOutType
		ev.Window = f.EventWindow
		ev.FocusMode = f.Mode
		ev.FocusDetail = f.Detail

	case platform.ExposeEvent_:
		e := parser.ParseExposeEvent(raw)
		ev.Type = ExposeType
		ev.Window = e.EventWindow
		ev.ExposeX = e.X
		ev.ExposeY = e.Y
		ev.ExposeWidth = e.Width
		ev.ExposeHeight = e.Height
		ev.ExposeCount = e.Count

	case platform.ConfigureNotifyEvent:
		cfg := parser.ParseConfigureEvent(raw)
		ev.Type = ConfigureType
		ev.Window = cfg.EventWindow
		ev.ConfigX = cfg.X
		ev.ConfigY = cfg.Y
		ev.ConfigWidth = cfg.Width
		ev.ConfigHeight = cfg.Height

	case platform.DestroyNotifyEvent:
		d := parser.ParseDestroyEvent(raw)
		ev.Type = DestroyType
		ev.Window = d.EventWindow

	case platform.MapNotifyEvent:
		ev.Type = MapType

	case platform.UnmapNotifyEvent:
		ev.Type = UnmapType

	case platform.PropertyNotifyEvent:
		prop := parser.ParsePropertyEvent(raw)
		ev.Type = PropertyType
		ev.Window = prop.EventWindow
		ev.Atom = prop.Atom

	case platform.ClientMessageEvent_:
		cm := parser.ParseClientMessageEvent(raw)
		ev.Type = ClientMessageType
		ev.Window = cm.EventWindow
		ev.MessageType = cm.MessageType
		ev.MessageData = cm.Data

	case platform.VirtualEvent:
		ev.Type = VirtualType
		ev.Name, _ = raw.Data.(string)
	}

	return ev
}

// WheelAccumulator turns MouseWheel deltas into whole scroll units, carrying
// the remainder so that the small steps of a high-resolution wheel add up
// instead of rounding to nothing (tk::MouseWheel in tk.tcl).
type WheelAccumulator struct {
	rem int
}

// Units returns how many units to scroll for a delta at unitsPerNotch per
// 120. The sign is that of a view command: positive scrolls down or right.
func (a *WheelAccumulator) Units(delta, unitsPerNotch int) int {
	total := a.rem - delta*unitsPerNotch
	n := total / 120
	a.rem = total - n*120
	return n
}
