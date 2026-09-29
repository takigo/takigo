package ttk

import (
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
)

// bindEntry wires the TEntry class bindings: press/drag/double/triple,
// keyboard (incl. Emacs), clipboard, focus in/out, and global click.
//
// Mirrors tk/library/ttk/entry.tcl class bindings, simplified to a single
// dispatch callback per event type.
func bindEntry(e *Entry, app widget.AppContext) {
	win := e.Win

	// Button1: position cursor. Shift extends selection. Double/triple select word/line.
	app.Dispatcher().Bind(win.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		if e.StateMode == EntryDisabled {
			return
		}
		if ev.Button == 1 {
			// Force focus so FocusIn fires even when the previous focus was
			// on a non-focusable widget.
			widget.Focus(app, win)
			e.ChangeState(StateFocus, 0)

			pos := e.edit.ClosestGap(ev.X)
			e.edit.SelAnchor = pos
			e.edit.InsertPos = pos
			e.edit.ClearSelection()
			e.Display()
		}
	})

	// Motion-drag selection (when Button1 still held).
	app.Dispatcher().Bind(win.PlatformID, event.MotionMask, func(ev *event.Event) {
		if ev.State&platform.Button1Mask != 0 && e.StateMode != EntryDisabled {
			e.edit.MoveCursor(e.edit.ClosestGap(ev.X), e.edit.SelAnchor, true)
		}
	})

	// Key handling.
	app.Dispatcher().Bind(win.PlatformID, event.KeyPressMask, func(ev *event.Event) {
		if e.StateMode == EntryDisabled {
			return
		}
		switch ev.KeySym {
		case platform.XK_Left, platform.XK_Right, platform.XK_Home, platform.XK_End:
			e.edit.HandleNavKey(ev)
		case platform.XK_BackSpace, platform.XK_Delete, platform.XK_Insert:
			e.edit.HandleEditKey(ev)
		default:
			if ev.State&platform.ControlMask != 0 {
				e.edit.HandleCtrlKey(ev)
				return
			}
			e.edit.HandleKey(ev)
		}
		e.notifyTextVar()
	})
	app.Dispatcher().Bind(win.PlatformID, event.VirtualMask, func(ev *event.Event) {
		if e.StateMode != EntryDisabled {
			e.edit.HandleVirtual(ev)
			e.notifyTextVar()
		}
	})

	// Focus events.
	app.Dispatcher().Bind(win.PlatformID, event.FocusChangeMask, func(ev *event.Event) {
		if ev.Type == event.FocusInType {
			e.ChangeState(StateFocus, 0)
			if e.ValidateMode == ValidateFocus || e.ValidateMode == ValidateAll || e.ValidateMode == ValidateFocusIn {
				e.Validate()
			}
			e.Display()
		} else if ev.Type == event.FocusOutType {
			if e.ValidateMode == ValidateFocus || e.ValidateMode == ValidateAll || e.ValidateMode == ValidateFocusOut {
				e.Validate()
			}
			e.ChangeState(0, StateFocus)
			e.edit.ClearSelection()
			e.Display()
		}
	})

	// Hide cursor when the user clicks any other window.
	app.Dispatcher().BindGlobalFor(win.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		if e.State&StateFocus == 0 || e.StateMode == EntryDisabled {
			return
		}
		if ev.Window == win.PlatformID {
			return
		}
		e.ChangeState(0, StateFocus)
		e.edit.ClearSelection()
		e.Display()
	})
}
