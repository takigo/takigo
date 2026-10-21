package spinbox

import (
	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/widget"
)

// bindSpinbox registers all event handlers for the spinbox widget.
func bindSpinbox(s *Spinbox, app widget.AppContext) {
	w := s.Win

	// Expose.
	app.Dispatcher().Bind(w.PlatformID, event.ExposureMask, s.handleExpose)

	// Configure (resize).
	app.Dispatcher().Bind(w.PlatformID, event.StructureNotifyMask, s.handleConfigure)

	// Focus.
	app.Dispatcher().Bind(w.PlatformID, event.FocusChangeMask, s.handleFocus)

	// Mouse: click to position cursor or press buttons.
	var wheel event.WheelAccumulator
	app.Dispatcher().Bind(w.PlatformID, event.MouseWheelMask, func(ev *event.Event) {
		if ev.Type != event.MouseWheelType {
			return
		}
		for n := wheel.Units(ev.Delta, 1); n != 0; {
			if n < 0 {
				s.SpinUp()
				n++
			} else {
				s.SpinDown()
				n--
			}
		}
	})

	app.Dispatcher().Bind(w.PlatformID, event.ButtonPressMask, s.handleButtonPress)

	// Mouse release.
	app.Dispatcher().Bind(w.PlatformID, event.ButtonReleaseMask, s.handleButtonRelease)

	// Mouse drag to select.
	app.Dispatcher().Bind(w.PlatformID, event.MotionMask, s.handleMotion)

	// Auto-scan: the pointer leaves the spinbox with button 1 held.
	app.Dispatcher().Bind(w.PlatformID, event.LeaveMask, s.handleLeave)
	app.Dispatcher().Bind(w.PlatformID, event.EnterMask, s.handleEnter)

	// Keyboard.
	app.Dispatcher().Bind(w.PlatformID, event.KeyPressMask, s.handleKeyPress)
	app.Dispatcher().Bind(w.PlatformID, event.VirtualMask, s.edit.HandleVirtual)
}

// handleExpose handles Exposure events.
func (s *Spinbox) handleExpose(ev *event.Event) {
	if ev.ExposeCount > 0 {
		return
	}
	s.Display()
}

// handleConfigure handles ConfigureNotify (resize) events.
func (s *Spinbox) handleConfigure(ev *event.Event) {
	if ev.Type == event.ConfigureType {
		s.computeGeometry()
		s.Display()
	}
}

// handleFocus handles FocusIn/FocusOut events.
func (s *Spinbox) handleFocus(ev *event.Event) {
	switch ev.Type {
	case event.FocusInType:
		s.hasFocus = true
		s.cursorOn = true
		s.blink.Start()
		s.Display()
	case event.FocusOutType:
		s.hasFocus = false
		s.blink.Stop()
		s.Display()
	}
}

// handleButtonPress handles mouse button press events.
func (s *Spinbox) handleButtonPress(ev *event.Event) {
	if ev.Button == 1 {
		app := s.App
		// A disabled spinbox does not take the focus, as Tk's
		// ::tk::spinbox::ButtonDown does (tk/library/spinbox.tcl:380).
		if s.State != widget.StateDisabled {
			widget.Focus(app, s.Win)
		}

		btn := s.hitButton(ev.X, ev.Y)
		switch btn {
		case "up":
			s.pressedButton = "up"
			s.spin()
			s.startRepeat()
		case "down":
			s.pressedButton = "down"
			s.spin()
			s.startRepeat()
		default:
			if s.State == widget.StateDisabled {
				return
			}
			// Click in text area -- position cursor.
			s.ClearSelection()
			s.InsertPos = s.closestGap(ev.X)
			s.SelAnchor = s.InsertPos
			s.Display()
		}
	}
}

// handleButtonRelease handles mouse button release events.
func (s *Spinbox) handleButtonRelease(ev *event.Event) {
	if ev.Button == 1 {
		s.stopRepeat()
		s.stopAutoScan()
		s.pressedButton = ""
		s.Display()
	}
}

// handleMotion handles mouse motion events (drag to select).
func (s *Spinbox) handleMotion(ev *event.Event) {
	s.lastX = ev.X
	if ev.State&platform.Button1Mask != 0 {
		// The spinner buttons stop repeating when the pointer wanders
		// off them, as ::tk::spinbox::Motion does (tk/library/spinbox.tcl).
		if s.pressedButton != "" && s.hitButton(ev.X, ev.Y) != s.pressedButton {
			s.stopRepeat()
			return
		}
		// Only drag-select in text area.
		if s.hitButton(ev.X, ev.Y) != "" {
			return
		}
		s.ExtendTo(s.closestGap(ev.X))
		s.seeInsert()
		s.Display()
	}
}

// handleLeave handles the pointer leaving the spinbox with button 1 held:
// the text is auto-scanned towards the pointer, as ::tk::spinbox::AutoScan
// does (tk/library/spinbox.tcl:565).
func (s *Spinbox) handleLeave(ev *event.Event) {
	if ev.State&platform.Button1Mask == 0 || s.pressedButton != "" {
		return
	}
	dir := 0
	switch {
	case ev.X >= s.Win.Width:
		dir = 1
	case ev.X < 0:
		dir = -1
	case s.lastX >= s.Win.Width-1:
		dir = 1
	case s.lastX <= 0:
		dir = -1
	}
	s.lastX = ev.X
	if dir != 0 {
		s.startAutoScan(dir)
	}
}

// handleEnter handles the pointer entering the spinbox again: the
// auto-scan stops (tk::CancelRepeat in tk/library/spinbox.tcl).
func (s *Spinbox) handleEnter(ev *event.Event) {
	s.stopAutoScan()
}

// handleKeyPress spins on Up and Down (library/spinbox.tcl) and edits the
// text as an entry does otherwise.
func (s *Spinbox) handleKeyPress(ev *event.Event) {
	switch ev.KeySym {
	case platform.XK_Up:
		s.SpinUp()
	case platform.XK_Down:
		s.SpinDown()
	default:
		s.edit.HandleKey(ev)
	}
}
