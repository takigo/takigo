//go:build darwin

package cocoa

import (
	clib "github.com/msorc/takigo/internal/cocoa"
	"github.com/msorc/takigo/platform"
)

// EventParser implements platform.EventParser for macOS.
type EventParser struct{}

func raw(ev *platform.RawEvent) *clib.RawEvent {
	return ev.Data.(*clib.RawEvent)
}

func (p *EventParser) ParseKeyEvent(ev *platform.RawEvent) platform.KeyEvent {
	r := raw(ev)
	return platform.KeyEvent{
		EventWindow: platform.WindowID(uintptr(r.Window())),
		X:           r.X(),
		Y:           r.Y(),
		RootX:       r.RootX(),
		RootY:       r.RootY(),
		State:       r.State(),
		KeyCode:     r.KeyCode(),
		KeySym:      platform.KeySym(r.KeySym()),
		Str:         r.Str(),
		Time:        platform.Timestamp(r.Time_()),
	}
}

func (p *EventParser) ParseKeyEventIM(ev *platform.RawEvent) platform.KeyEvent {
	// On macOS, input method is handled natively — same as ParseKeyEvent.
	return p.ParseKeyEvent(ev)
}

func (p *EventParser) ParseButtonEvent(ev *platform.RawEvent) platform.ButtonEvent {
	r := raw(ev)
	return platform.ButtonEvent{
		EventWindow: platform.WindowID(uintptr(r.Window())),
		X:           r.X(),
		Y:           r.Y(),
		RootX:       r.RootX(),
		RootY:       r.RootY(),
		State:       r.State(),
		Button:      r.Button(),
		Time:        platform.Timestamp(r.Time_()),
	}
}

func (p *EventParser) ParseMotionEvent(ev *platform.RawEvent) platform.MotionEvent {
	r := raw(ev)
	return platform.MotionEvent{
		EventWindow: platform.WindowID(uintptr(r.Window())),
		X:           r.X(),
		Y:           r.Y(),
		RootX:       r.RootX(),
		RootY:       r.RootY(),
		State:       r.State(),
		Time:        platform.Timestamp(r.Time_()),
	}
}

func (p *EventParser) ParseExposeEvent(ev *platform.RawEvent) platform.ExposeEvent {
	r := raw(ev)
	return platform.ExposeEvent{
		EventWindow: platform.WindowID(uintptr(r.Window())),
		X:           r.ExposeX(),
		Y:           r.ExposeY(),
		Width:       r.ExposeWidth(),
		Height:      r.ExposeHeight(),
		Count:       r.ExposeCount(),
	}
}

func (p *EventParser) ParseConfigureEvent(ev *platform.RawEvent) platform.ConfigureEvent {
	r := raw(ev)
	return platform.ConfigureEvent{
		EventWindow: platform.WindowID(uintptr(r.Window())),
		X:           r.X(),
		Y:           r.Y(),
		Width:       r.Width(),
		Height:      r.Height(),
	}
}

func (p *EventParser) ParseClientMessageEvent(ev *platform.RawEvent) platform.ClientMessageEvent {
	r := raw(ev)
	return platform.ClientMessageEvent{
		EventWindow: platform.WindowID(uintptr(r.Window())),
		MessageType: platform.AtomID(r.MessageType()),
		Data:        r.MessageData(),
	}
}

func (p *EventParser) ParseCrossingEvent(ev *platform.RawEvent) platform.CrossingEvent {
	r := raw(ev)
	return platform.CrossingEvent{
		EventWindow: platform.WindowID(uintptr(r.Window())),
		X:           r.X(),
		Y:           r.Y(),
		State:       r.State(),
		Time:        platform.Timestamp(r.Time_()),
	}
}

func (p *EventParser) ParseDestroyEvent(ev *platform.RawEvent) platform.DestroyEvent {
	r := raw(ev)
	return platform.DestroyEvent{
		EventWindow: platform.WindowID(uintptr(r.Window())),
	}
}

func (p *EventParser) ParseFocusEvent(ev *platform.RawEvent) platform.FocusEvent {
	r := raw(ev)
	return platform.FocusEvent{
		EventWindow: platform.WindowID(uintptr(r.Window())),
		Mode:        r.FocusMode(),
		Detail:      r.FocusDetail(),
	}
}

func (p *EventParser) ParsePropertyEvent(ev *platform.RawEvent) platform.PropertyEvent {
	r := raw(ev)
	return platform.PropertyEvent{
		EventWindow: platform.WindowID(uintptr(r.Window())),
	}
}

func (p *EventParser) ParseSelectionRequestEvent(ev *platform.RawEvent) platform.SelectionRequestParsed {
	return platform.SelectionRequestParsed{}
}

func (p *EventParser) ParseSelectionClearEvent(ev *platform.RawEvent) platform.SelectionClearParsed {
	return platform.SelectionClearParsed{}
}

func (p *EventParser) ParseSelectionNotifyEvent(ev *platform.RawEvent) platform.SelectionNotifyParsed {
	return platform.SelectionNotifyParsed{}
}

// Verify at compile time.
var _ platform.EventParser = (*EventParser)(nil)
