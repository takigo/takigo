package entry

import (
	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/widget"
)

// bindEntry registers all event handlers for the entry widget.
func bindEntry(e *Entry, app widget.AppContext) {
	w := e.Win

	// Expose.
	app.Dispatcher().Bind(w.PlatformID, event.ExposureMask, e.handleExpose)

	// Configure (resize).
	app.Dispatcher().Bind(w.PlatformID, event.StructureNotifyMask, e.handleConfigure)

	// Focus.
	app.Dispatcher().Bind(w.PlatformID, event.FocusChangeMask, e.handleFocus)

	// Mouse: click to position cursor and take focus.
	app.Dispatcher().Bind(w.PlatformID, event.ButtonPressMask, e.handleButtonPress)

	// Mouse: drag to select.
	app.Dispatcher().Bind(w.PlatformID, event.MotionMask, e.handleMotion)

	// Keyboard.
	app.Dispatcher().Bind(w.PlatformID, event.KeyPressMask, func(ev *event.Event) { e.edit.HandleKey(ev) })
	app.Dispatcher().Bind(w.PlatformID, event.VirtualMask, e.edit.HandleVirtual)
}

// handleExpose handles Exposure events.
func (e *Entry) handleExpose(ev *event.Event) {
	if ev.ExposeCount > 0 {
		return
	}
	e.Display()
}

// handleConfigure handles ConfigureNotify (resize) events.
func (e *Entry) handleConfigure(ev *event.Event) {
	if ev.Type == event.ConfigureType {
		e.computeGeometry()
		e.Display()
		// Tk's EntryUpdateScrollbar runs on every redisplay, so the
		// scrollbar learns the real visible fraction once laid out.
		e.notifyScrollbar()
	}
}

// handleFocus handles FocusIn/FocusOut events.
func (e *Entry) handleFocus(ev *event.Event) {
	switch ev.Type {
	case event.FocusInType:
		e.HasFocus = true
		e.CursorOn = true
		e.tryFocusValidate("focusin")
		e.Display()
	case event.FocusOutType:
		e.HasFocus = false
		e.tryFocusValidate("focusout")
		e.Display()
	}
}

// handleButtonPress handles mouse button press events.
func (e *Entry) handleButtonPress(ev *event.Event) {
	if ev.Button == 1 {
		// Request X11 input focus so key events come to this window.
		widget.Focus(e.App, e.Win)
		e.ClearSelection()
		e.InsertPos = e.closestGap(ev.X)
		e.SelAnchor = e.InsertPos
		e.Display()
	}
}

// handleMotion handles mouse motion events (drag to select).
func (e *Entry) handleMotion(ev *event.Event) {
	if ev.State&platform.Button1Mask != 0 {
		e.ExtendTo(e.closestGap(ev.X))
		e.seeInsert()
		e.Display()
	}
}
