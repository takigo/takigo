//go:build linux || freebsd || openbsd || netbsd

package x11

import (
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/platform"
)

// X11EventParser implements platform.EventParser for X11.
type X11EventParser struct {
	dpy *xlib.Display
}

// NewEventParser creates a new X11 event parser.
func NewEventParser(dpy *xlib.Display) *X11EventParser {
	return &X11EventParser{dpy: dpy}
}

func (p *X11EventParser) ParseKeyEvent(ev *platform.RawEvent) platform.KeyEvent {
	raw := ev.Data.(*xlib.RawEvent)
	k := raw.ParseKeyEvent()
	return platform.KeyEvent{
		EventWindow: platform.WindowID(k.EventWindow),
		RootWindow:  platform.WindowID(k.RootWindow),
		X:           k.X,
		Y:           k.Y,
		RootX:       k.RootX,
		RootY:       k.RootY,
		State:       k.State,
		KeyCode:     k.KeyCode,
		KeySym:      platform.KeySym(k.KeySym),
		Str:         k.Str,
		Time:        platform.Timestamp(k.Time),
	}
}

func (p *X11EventParser) ParseKeyEventIM(ev *platform.RawEvent) platform.KeyEvent {
	raw := ev.Data.(*xlib.RawEvent)
	k := p.dpy.ParseKeyEventIM(raw)
	return platform.KeyEvent{
		EventWindow: platform.WindowID(k.EventWindow),
		RootWindow:  platform.WindowID(k.RootWindow),
		X:           k.X,
		Y:           k.Y,
		RootX:       k.RootX,
		RootY:       k.RootY,
		State:       k.State,
		KeyCode:     k.KeyCode,
		KeySym:      platform.KeySym(k.KeySym),
		Str:         k.Str,
		Time:        platform.Timestamp(k.Time),
	}
}

func (p *X11EventParser) ParseButtonEvent(ev *platform.RawEvent) platform.ButtonEvent {
	raw := ev.Data.(*xlib.RawEvent)
	b := raw.ParseButtonEvent()
	return platform.ButtonEvent{
		EventWindow: platform.WindowID(b.EventWindow),
		X:           b.X,
		Y:           b.Y,
		RootX:       b.RootX,
		RootY:       b.RootY,
		State:       b.State,
		Button:      b.Button,
		Time:        platform.Timestamp(b.Time),
	}
}

func (p *X11EventParser) ParseMotionEvent(ev *platform.RawEvent) platform.MotionEvent {
	raw := ev.Data.(*xlib.RawEvent)
	m := raw.ParseMotionEvent()
	return platform.MotionEvent{
		EventWindow: platform.WindowID(m.EventWindow),
		X:           m.X,
		Y:           m.Y,
		RootX:       m.RootX,
		RootY:       m.RootY,
		State:       m.State,
		Time:        platform.Timestamp(m.Time),
	}
}

func (p *X11EventParser) ParseExposeEvent(ev *platform.RawEvent) platform.ExposeEvent {
	raw := ev.Data.(*xlib.RawEvent)
	e := raw.ParseExposeEvent()
	return platform.ExposeEvent{
		EventWindow: platform.WindowID(e.EventWindow),
		X:           e.X,
		Y:           e.Y,
		Width:       e.Width,
		Height:      e.Height,
		Count:       e.Count,
	}
}

func (p *X11EventParser) ParseConfigureEvent(ev *platform.RawEvent) platform.ConfigureEvent {
	raw := ev.Data.(*xlib.RawEvent)
	c := raw.ParseConfigureEvent()
	return platform.ConfigureEvent{
		EventWindow: platform.WindowID(c.EventWindow),
		X:           c.X,
		Y:           c.Y,
		Width:       c.Width,
		Height:      c.Height,
	}
}

func (p *X11EventParser) ParseClientMessageEvent(ev *platform.RawEvent) platform.ClientMessageEvent {
	raw := ev.Data.(*xlib.RawEvent)
	cm := raw.ParseClientMessageEvent()
	return platform.ClientMessageEvent{
		EventWindow: platform.WindowID(cm.EventWindow),
		MessageType: platform.AtomID(cm.MessageType),
		Format:      cm.Format,
		Data:        cm.Data,
	}
}

func (p *X11EventParser) ParseCrossingEvent(ev *platform.RawEvent) platform.CrossingEvent {
	raw := ev.Data.(*xlib.RawEvent)
	c := raw.ParseCrossingEvent()
	return platform.CrossingEvent{
		EventWindow: platform.WindowID(c.EventWindow),
		X:           c.X,
		Y:           c.Y,
		State:       c.State,
		Time:        platform.Timestamp(c.Time),
	}
}

func (p *X11EventParser) ParseDestroyEvent(ev *platform.RawEvent) platform.DestroyEvent {
	raw := ev.Data.(*xlib.RawEvent)
	d := raw.ParseDestroyEvent()
	return platform.DestroyEvent{
		EventWindow: platform.WindowID(d.EventWindow),
	}
}

func (p *X11EventParser) ParseFocusEvent(ev *platform.RawEvent) platform.FocusEvent {
	raw := ev.Data.(*xlib.RawEvent)
	f := raw.ParseFocusEvent()
	return platform.FocusEvent{
		EventWindow: platform.WindowID(f.EventWindow),
		Mode:        f.Mode,
		Detail:      f.Detail,
	}
}

func (p *X11EventParser) ParsePropertyEvent(ev *platform.RawEvent) platform.PropertyEvent {
	raw := ev.Data.(*xlib.RawEvent)
	prop := raw.ParsePropertyEvent()
	return platform.PropertyEvent{
		EventWindow: platform.WindowID(prop.EventWindow),
		Atom:        platform.AtomID(prop.Atom),
		Deleted:     prop.Deleted,
	}
}

func (p *X11EventParser) ParseSelectionRequestEvent(ev *platform.RawEvent) platform.SelectionRequestParsed {
	raw := ev.Data.(*xlib.RawEvent)
	s := raw.ParseSelectionRequestEvent()
	return platform.SelectionRequestParsed{
		Owner:     platform.WindowID(s.Owner),
		Requestor: platform.WindowID(s.Requestor),
		Selection: platform.AtomID(s.Selection),
		Target:    platform.AtomID(s.Target),
		Property:  platform.AtomID(s.Property),
		Time:      platform.Timestamp(s.Time),
	}
}

func (p *X11EventParser) ParseSelectionClearEvent(ev *platform.RawEvent) platform.SelectionClearParsed {
	raw := ev.Data.(*xlib.RawEvent)
	s := raw.ParseSelectionClearEvent()
	return platform.SelectionClearParsed{
		Window:    platform.WindowID(s.Window),
		Selection: platform.AtomID(s.Selection),
	}
}

func (p *X11EventParser) ParseSelectionNotifyEvent(ev *platform.RawEvent) platform.SelectionNotifyParsed {
	raw := ev.Data.(*xlib.RawEvent)
	s := raw.ParseSelectionNotifyEvent()
	return platform.SelectionNotifyParsed{
		Requestor: platform.WindowID(s.Requestor),
		Selection: platform.AtomID(s.Selection),
		Target:    platform.AtomID(s.Target),
		Property:  platform.AtomID(s.Property),
		Time:      platform.Timestamp(s.Time),
	}
}

// Verify at compile time.
var _ platform.EventParser = (*X11EventParser)(nil)
