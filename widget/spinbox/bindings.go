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
		s.Display()
	case event.FocusOutType:
		s.hasFocus = false
		s.Display()
	}
}

// handleButtonPress handles mouse button press events.
func (s *Spinbox) handleButtonPress(ev *event.Event) {
	if ev.Button == 1 {
		app := s.App
		widget.Focus(app, s.Win)

		btn := s.hitButton(ev.X, ev.Y)
		switch btn {
		case "up":
			s.pressedButton = "up"
			s.SpinUp()
			s.Display()
		case "down":
			s.pressedButton = "down"
			s.SpinDown()
			s.Display()
		default:
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
		s.pressedButton = ""
		s.Display()
	}
}

// handleMotion handles mouse motion events (drag to select).
func (s *Spinbox) handleMotion(ev *event.Event) {
	if ev.State&platform.Button1Mask != 0 {
		// Only drag-select in text area.
		if s.hitButton(ev.X, ev.Y) != "" {
			return
		}
		s.ExtendTo(s.closestGap(ev.X))
		s.seeInsert()
		s.Display()
	}
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
