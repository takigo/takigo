package takigo

import (
	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/window"
	"github.com/takigo/takigo/wm"
)

// The App's event policy: what Tk_HandleEvent (tkEvent.c), the focus and
// grab filters and the window-manager protocol handling do before an
// event reaches a widget.

// filterEvent runs on every converted event before it is dispatched.
// Tk never reads a child window's size back from X: the geometry managers
// own it. A queued ConfigureNotify can describe a size that has since
// been replaced, so the current one is reported to every handler.
func (a *App) filterEvent(ev *event.Event) {
	switch ev.Type {
	case event.ConfigureType:
		w := a.display.LookupWindow(ev.Window)
		switch {
		case w == nil:
		case !w.IsTopLevel():
			ev.ConfigWidth, ev.ConfigHeight = w.Width, w.Height
		case w.WmData != nil:
			// A size the user dragged to becomes the toplevel's geometry
			// (ConfigureEvent in tkUnixWm.c).
			w.WmData.ConfigureNotify(ev.ConfigWidth, ev.ConfigHeight)
		}
	case event.FocusInType, event.FocusOutType:
		// Tk's focus model: the manager turns toplevel focus changes into
		// FocusIn/FocusOut on its focus windows and drops the rest
		// (TkFocusFilterEvent).
		if a.focusMgr != nil && !a.focusMgr.FilterEvent(ev) {
			ev.Type = 0
		}
	case event.KeyPressType, event.KeyReleaseType, event.ButtonPressType, event.ButtonReleaseType,
		event.MotionType, event.EnterType, event.LeaveType, event.MouseWheelType:
		// Keys go to the focus window (TkFocusKeyEvent), and leaving a
		// toplevel can end an implicit focus.
		if a.focusMgr != nil && (ev.Type == event.KeyPressType || ev.Type == event.KeyReleaseType || ev.Type == event.LeaveType) {
			a.focusMgr.FilterEvent(ev)
		}
		// A local grab (tkGrab.c's TkPointerEvent) discards input for
		// this application's windows outside the grab tree.
		if a.grabMgr.Current() == nil {
			return
		}
		if w := a.display.LookupWindow(ev.Window); w != nil && a.grabMgr.ShouldRedirect(w) {
			ev.Type = 0
		}
	}
}

// handleRawEvent routes the selection events, which the dispatcher does
// not carry, to the selection manager: serving the clipboard and the
// answers to its own requests.
func (a *App) handleRawEvent(raw *platform.RawEvent) {
	switch raw.EventType {
	case platform.SelectionRequestEvent:
		req := a.parser.ParseSelectionRequestEvent(raw)
		a.selMgr.HandleSelectionRequest(req.Requestor, req.Selection, req.Target, req.Property, req.Time)
	case platform.SelectionClearEvent:
		clr := a.parser.ParseSelectionClearEvent(raw)
		a.selMgr.HandleSelectionClear(clr.Selection)
	case platform.SelectionNotifyEvent:
		ntf := a.parser.ParseSelectionNotifyEvent(raw)
		a.selMgr.HandleSelectionNotify(ntf.Requestor, ntf.Property)
	case platform.PropertyNotifyEvent:
		prop := a.parser.ParsePropertyEvent(raw)
		a.selMgr.HandlePropertyNotify(prop.EventWindow, prop.Atom, prop.Deleted)
	}
}

// initWM gives the root window its window-manager state, as Tk does for
// ".": WM_CLASS, WM_HINTS, the size hints and WM_PROTOCOLS, then the
// title, geometry and icon name the App was created with. Closing the
// window quits the App until OnDeleteWindow is told otherwise.
func (a *App) initWM(cfg appConfig) {
	a.wmInfo = wm.Init(a.root)
	a.wmInfo.SetTitle(cfg.title)
	if cfg.geometry != "" {
		if err := a.wmInfo.SetGeometry(cfg.geometry); err != nil {
			cfg.logger.Warn("bad geometry", "geometry", cfg.geometry, "err", err)
		}
	}
	if cfg.iconName != "" {
		a.wmInfo.SetIconName(cfg.iconName)
	}
	a.wmInfo.OnDeleteWindow(a.Quit)
}

// forgetWindow runs when a window is destroyed. Tk_DestroyWindow delivers
// <Destroy> and then forgets the window: its event handlers, bind tags and
// focus state go with it, so their closures do not leak and a reused
// window ID starts clean.
func (a *App) forgetWindow(w *window.Window) {
	if w.PlatformID == 0 {
		return
	}
	a.dispatcher.Dispatch(&event.Event{Type: event.DestroyType, Window: w.PlatformID})
	a.dispatcher.Unbind(w.PlatformID)
	a.bindEng.UnregisterWindow(w)
	a.focusMgr.HandleDestroyWindow(w)
}

// routeClientMessage gives a client message to the drag-and-drop
// manager, else to the toplevel's window-manager state, as
// TkWmProtocolEventProc (tkUnixWm.c) does for WM_DELETE_WINDOW,
// _NET_WM_PING and the other protocols.
func (a *App) routeClientMessage(ev *event.Event) {
	if ev.Type != event.ClientMessageType {
		return
	}
	if a.dnd != nil && a.dnd.Handle(ev) {
		return
	}
	w := a.display.LookupWindow(ev.Window)
	if w == nil {
		return
	}
	if info := w.WmData; info != nil {
		info.HandleClientMessage(ev.MessageType, ev.MessageData)
	}
}
