// Listbox event bindings, porting tk/library/listbox.tcl.

package listbox

import (
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
)

func bindListbox(lb *Listbox, app widget.AppContext) {
	w := lb.Win

	// Expose.
	app.Dispatcher().Bind(w.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		lb.Display()
	})

	// Configure.
	app.Dispatcher().Bind(w.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			w.Width = ev.ConfigWidth
			w.Height = ev.ConfigHeight
			lb.notifyYScrollbar()
			lb.Display()
		}
	})

	// Focus.
	app.Dispatcher().Bind(w.PlatformID, event.FocusChangeMask, func(ev *event.Event) {
		if ev.Type == event.FocusInType {
			lb.HasFocus = true
			lb.Display()
		} else if ev.Type == event.FocusOutType {
			lb.HasFocus = false
			lb.Display()
		}
	})

	// Button press.
	app.Dispatcher().Bind(w.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		// Take focus.
		widget.Focus(app, w)

		if ev.Button == 1 {
			idx := lb.indexAtY(ev.Y)
			if idx < 0 || idx >= len(lb.items) {
				return
			}
			handleSelect(lb, idx, ev.State)
			lb.activeIndex = idx
			lb.Display()
			notifySelect(lb)
		} else if ev.Button == 4 {
			// Scroll up.
			lb.YView(lb.topIndex - 3)
		} else if ev.Button == 5 {
			// Scroll down.
			lb.YView(lb.topIndex + 3)
		}
	})

	// Motion (drag select for browse/extended).
	app.Dispatcher().Bind(w.PlatformID, event.MotionMask, func(ev *event.Event) {
		if ev.State&platform.Button1Mask == 0 {
			return
		}
		idx := lb.indexAtY(ev.Y)
		if idx < 0 || idx >= len(lb.items) {
			return
		}

		switch lb.selectMode {
		case SelectBrowse:
			lb.selected = map[int]bool{idx: true}
			lb.activeIndex = idx
		case SelectExtended:
			if lb.selAnchor < 0 {
				lb.selAnchor = idx
			}
			lb.selected = make(map[int]bool)
			lo, hi := lb.selAnchor, idx
			if lo > hi {
				lo, hi = hi, lo
			}
			for i := lo; i <= hi; i++ {
				lb.selected[i] = true
			}
			lb.activeIndex = idx
		}
		lb.See(idx)
		lb.Display()
		notifySelect(lb)
	})

	// Keyboard.
	app.Dispatcher().Bind(w.PlatformID, event.KeyPressMask, func(ev *event.Event) {
		switch ev.KeySym {
		case platform.XK_Up:
			moveActive(lb, -1)
		case platform.XK_Down:
			moveActive(lb, 1)
		case platform.XK_Prior: // PageUp
			moveActive(lb, -lb.visibleLines())
		case platform.XK_Next: // PageDown
			moveActive(lb, lb.visibleLines())
		case platform.XK_Home:
			setActive(lb, 0)
		case platform.XK_End:
			setActive(lb, len(lb.items)-1)
		case platform.XK_space:
			// Toggle selection on active item.
			if lb.activeIndex >= 0 && lb.activeIndex < len(lb.items) {
				if lb.selectMode == SelectMultiple {
					if lb.selected[lb.activeIndex] {
						delete(lb.selected, lb.activeIndex)
					} else {
						lb.selected[lb.activeIndex] = true
					}
				} else {
					lb.selected = map[int]bool{lb.activeIndex: true}
				}
				lb.Display()
			}
		}
	})
}

func handleSelect(lb *Listbox, idx int, state uint) {
	switch lb.selectMode {
	case SelectSingle:
		lb.selected = map[int]bool{idx: true}
		lb.selAnchor = idx

	case SelectBrowse:
		lb.selected = map[int]bool{idx: true}
		lb.selAnchor = idx

	case SelectMultiple:
		if lb.selected[idx] {
			delete(lb.selected, idx)
		} else {
			lb.selected[idx] = true
		}

	case SelectExtended:
		shift := state&platform.ShiftMask != 0
		ctrl := state&platform.ControlMask != 0

		if shift && lb.selAnchor >= 0 {
			// Extend from anchor.
			lb.selected = make(map[int]bool)
			lo, hi := lb.selAnchor, idx
			if lo > hi {
				lo, hi = hi, lo
			}
			for i := lo; i <= hi; i++ {
				lb.selected[i] = true
			}
		} else if ctrl {
			// Toggle single item.
			if lb.selected[idx] {
				delete(lb.selected, idx)
			} else {
				lb.selected[idx] = true
			}
			lb.selAnchor = idx
		} else {
			// Normal click: select single.
			lb.selected = map[int]bool{idx: true}
			lb.selAnchor = idx
		}
	}
}

func moveActive(lb *Listbox, delta int) {
	newIdx := lb.activeIndex + delta
	setActive(lb, newIdx)
}

func setActive(lb *Listbox, idx int) {
	if len(lb.items) == 0 {
		return
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= len(lb.items) {
		idx = len(lb.items) - 1
	}
	lb.activeIndex = idx

	// In browse/single mode, also select.
	if lb.selectMode == SelectBrowse || lb.selectMode == SelectSingle {
		lb.selected = map[int]bool{idx: true}
	}

	lb.See(idx)
	lb.Display()
}

func notifySelect(lb *Listbox) {
	if lb.SelectCmd != nil {
		lb.SelectCmd()
	}
}
