package canvas

import (
	"strconv"
	"unicode/utf8"

	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/geometry/place"
	"github.com/takigo/takigo/platform"
)

// itemHandler stores a per-item event binding.
type itemHandler struct {
	mask    event.Mask
	handler func(*event.Event)
}

// bindCanvas sets up the canvas event handlers for expose, configure,
// destroy, mouse motion (pick), and mouse button (dispatch).
func bindCanvas(c *Canvas) {
	disp := c.App.Dispatcher()
	w := c.Win

	// Expose → redraw.
	disp.Bind(w.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		c.Display()
	})

	// Configure → resize.
	disp.Bind(w.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		switch ev.Type {
		case event.ConfigureType:
			w.Width = ev.ConfigWidth
			w.Height = ev.ConfigHeight
			c.setOrigin(c.xOrigin, c.yOrigin)
			c.scheduleRedraw()
			c.notifyScrollbars()
			place.ArrangeContainer(w)
			w.NotifyConfigure()
		case event.DestroyType:
			c.Destroy()
		}
	})

	var wheel event.WheelAccumulator
	disp.Bind(w.PlatformID, event.MouseWheelMask, func(ev *event.Event) {
		if ev.Type != event.MouseWheelType {
			return
		}
		n := wheel.Units(ev.Delta, 5)
		if n == 0 {
			return
		}
		if ev.State&platform.ShiftMask != 0 {
			c.XViewScroll(n, false)
		} else {
			c.YViewScroll(n, false)
		}
	})

	// Pointer motion, crossing and buttons → pick the current item and
	// dispatch to its bindings.
	disp.Bind(w.PlatformID, event.MotionMask|event.EnterMask|event.LeaveMask|
		event.ButtonPressMask|event.ButtonReleaseMask, func(ev *event.Event) {
		c.handlePointer(ev)
	})

	// Key press → dispatch to focused text item.
	disp.Bind(w.PlatformID, event.KeyPressMask, func(ev *event.Event) {
		if ev.Type != event.KeyPressType || c.focusItemID == 0 {
			return
		}
		e, ok := c.idMap[c.focusItemID]
		if !ok {
			return
		}
		ti, ok := e.item.(*TextItem)
		if !ok {
			return
		}
		c.redrawItems(e)
		switch ev.KeySym {
		case platform.XK_BackSpace:
			if ti.cursorPos > 0 {
				// Find previous rune boundary.
				prev := ti.cursorPos
				_, sz := utf8.DecodeLastRuneInString(ti.text[:prev])
				ti.DeleteChars(prev-sz, prev)
			}
		case platform.XK_Delete:
			if ti.cursorPos < len(ti.text) {
				_, sz := utf8.DecodeRuneInString(ti.text[ti.cursorPos:])
				ti.DeleteChars(ti.cursorPos, ti.cursorPos+sz)
			}
		case platform.XK_Left:
			if ti.cursorPos > 0 {
				_, sz := utf8.DecodeLastRuneInString(ti.text[:ti.cursorPos])
				ti.cursorPos -= sz
			}
		case platform.XK_Right:
			if ti.cursorPos < len(ti.text) {
				_, sz := utf8.DecodeRuneInString(ti.text[ti.cursorPos:])
				ti.cursorPos += sz
			}
		case platform.XK_Home:
			ti.cursorPos = 0
		case platform.XK_End:
			ti.cursorPos = len(ti.text)
		case platform.XK_Return:
			ti.InsertText(ti.cursorPos, "\n")
		default:
			// Insert what %A would give: the composed string, or the
			// keysym's character; modifiers, function keys and Control
			// combinations have none.
			if ev.State&platform.ControlMask != 0 {
				break
			}
			s := ev.Str
			if s == "" {
				if r := platform.KeySymToRune(ev.KeySym); r != 0 {
					s = string(r)
				}
			}
			if s != "" && s[0] >= 0x20 && s[0] != 0x7f {
				ti.InsertText(ti.cursorPos, s)
			}
		}
		c.redrawItems(e)
	})
}

// allButtons is the state of any mouse button being down.
const allButtons = platform.Button1Mask | platform.Button2Mask | platform.Button3Mask

// buttonMask returns the state bit of a mouse button (Tk_GetButtonMask).
func buttonMask(button uint) uint {
	if button < 1 || button > 5 {
		return 0
	}
	return platform.Button1Mask << (button - 1)
}

// handlePointer is CanvasBindProc: it keeps pointerState, repicks the current
// item and dispatches ev to it. On a press it repicks with the state before
// the press; on a release it dispatches first, with the button still down,
// and repicks after, so a drag's release reaches the item it started on.
func (c *Canvas) handlePointer(ev *event.Event) {
	x, y := float64(ev.X), float64(ev.Y)
	switch ev.Type {
	case event.ButtonPressType:
		c.pointerState = ev.State
		c.pickCurrentItem(x, y)
		c.pointerState ^= buttonMask(ev.Button)
		c.dispatchItemEvent(ev)
	case event.ButtonReleaseType:
		c.pointerState = ev.State
		c.dispatchItemEvent(ev)
		c.pointerState = ev.State &^ buttonMask(ev.Button)
		c.pickCurrentItem(x, y)
	case event.EnterType:
		c.pointerState = ev.State
		c.pickCurrentItem(x, y)
	case event.LeaveType:
		c.pointerState = ev.State
		c.pick(nil)
	case event.MotionType:
		c.pointerState = ev.State
		c.pickCurrentItem(x, y)
		c.dispatchItemEvent(ev)
	}
}

// pickCurrentItem makes the item at window position (winX, winY) current.
func (c *Canvas) pickCurrentItem(winX, winY float64) {
	canvasX := winX + float64(c.xOrigin) - float64(c.inset)
	canvasY := winY + float64(c.yOrigin) - float64(c.inset)
	c.pick(c.findClosest(canvasX, canvasY, c.closeEnough))
}

// pick is Tk's PickCurrentItem: it makes entry the current item, sending
// <Leave> to the old one and <Enter> to the new one. While a button is down
// the old item stays current: it gets its <Leave> when the pointer moves off
// it, but no other item is entered until the buttons are released.
func (c *Canvas) pick(entry *itemEntry) {
	// A <Leave> handler below may move the pointer state on; the pending
	// call finishes the pick.
	if c.repicking {
		return
	}
	if entry == c.currentItem && !c.leftGrabbed {
		return
	}
	buttonDown := c.pointerState&allButtons != 0
	if !buttonDown {
		c.leftGrabbed = false
	}
	if entry != c.currentItem && c.currentItem != nil && !c.leftGrabbed {
		c.repicking = true
		c.dispatchToItem(c.currentItem, &event.Event{Type: event.LeaveType})
		c.repicking = false
		// The <Leave> handler may have deleted the item being entered.
		if entry != nil && entry.dead {
			entry = nil
		}
	}
	if entry != c.currentItem && buttonDown {
		c.leftGrabbed = true
		return
	}
	old := c.currentItem
	c.leftGrabbed = false
	c.currentItem = entry
	// -activefill and friends depend on which item is current.
	if old != entry {
		if hasActive(old) {
			c.redrawItems(old)
		}
		if hasActive(entry) {
			c.redrawItems(entry)
		}
	}
	if entry != nil {
		c.dispatchToItem(entry, &event.Event{Type: event.EnterType})
	}
}

func hasActive(e *itemEntry) bool {
	if e == nil {
		return false
	}
	b := itemBase(e.item)
	return b != nil && b.activeFill != nil
}

// dispatchItemEvent dispatches an event to the current item's bindings.
func (c *Canvas) dispatchItemEvent(ev *event.Event) {
	if c.currentItem == nil {
		return
	}
	c.dispatchToItem(c.currentItem, ev)
}

// dispatchToItem dispatches an event to all matching bindings for an item:
// those on its ID, then those on each of its tags.
func (c *Canvas) dispatchToItem(entry *itemEntry, ev *event.Event) {
	if len(c.idBindings) == 0 && len(c.itemBindings) == 0 {
		return
	}
	evMask := event.TypeToMask(ev.Type)
	if evMask == 0 {
		return
	}
	runItemHandlers(c.idBindings[entry.id], evMask, ev)
	if base := itemBase(entry.item); base != nil {
		for _, tag := range base.Tags {
			runItemHandlers(c.itemBindings[tag], evMask, ev)
		}
	}
}

func runItemHandlers(handlers []itemHandler, evMask event.Mask, ev *event.Event) {
	for _, h := range handlers {
		if h.mask&evMask != 0 {
			h.handler(ev)
		}
	}
}

// BindItem binds an event handler to items matching tagOrID.
func (c *Canvas) BindItem[S Selector](sel S, mask event.Mask, handler func(*event.Event)) {
	tagOrID := selectorString(sel)
	h := itemHandler{mask: mask, handler: handler}
	if id, err := strconv.ParseInt(tagOrID, 10, 64); err == nil {
		c.idBindings[ItemID(id)] = append(c.idBindings[ItemID(id)], h)
		return
	}
	c.itemBindings[tagOrID] = append(c.itemBindings[tagOrID], h)
}
