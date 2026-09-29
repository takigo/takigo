package canvas

import (
	"strconv"
	"unicode/utf8"

	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/place"
	"github.com/msorc/takigo/platform"
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
		if ev.Type == event.ConfigureType {
			w.Width = ev.ConfigWidth
			w.Height = ev.ConfigHeight
			c.setOrigin(c.xOrigin, c.yOrigin)
			c.scheduleRedraw()
			c.notifyScrollbars()
			place.ArrangeContainer(w)
			w.NotifyConfigure()
		} else if ev.Type == event.DestroyType {
			c.Destroy()
		}
	})

	// Motion → pick current item + dispatch Enter/Leave.
	disp.Bind(w.PlatformID, event.MotionMask, func(ev *event.Event) {
		c.pickCurrentItem(float64(ev.X), float64(ev.Y))
		c.dispatchItemEvent(ev)
	})

	// Enter/Leave window → update current item.
	disp.Bind(w.PlatformID, event.EnterMask|event.LeaveMask, func(ev *event.Event) {
		if ev.Type == event.LeaveType {
			c.setCurrentItem(nil, ev)
		} else {
			c.pickCurrentItem(float64(ev.X), float64(ev.Y))
		}
	})

	// Button press/release → mouse wheel scroll or dispatch to current item.
	disp.Bind(w.PlatformID, event.ButtonPressMask|event.ButtonReleaseMask, func(ev *event.Event) {
		if ev.Type == event.ButtonPressType {
			switch ev.Button {
			case 4: // scroll up
				c.YViewScroll(-5, false)
				return
			case 5: // scroll down
				c.YViewScroll(5, false)
				return
			case 6: // scroll left (horizontal wheel)
				c.XViewScroll(-5, false)
				return
			case 7: // scroll right (horizontal wheel)
				c.XViewScroll(5, false)
				return
			}
		}
		c.pickCurrentItem(float64(ev.X), float64(ev.Y))
		c.dispatchItemEvent(ev)
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

// pickCurrentItem updates c.currentItem based on the mouse position.
func (c *Canvas) pickCurrentItem(winX, winY float64) {
	// Convert window coords to canvas coords.
	canvasX := winX + float64(c.xOrigin) - float64(c.inset)
	canvasY := winY + float64(c.yOrigin) - float64(c.inset)

	entry := c.findClosest(canvasX, canvasY, c.closeEnough)
	if entry != c.currentItem {
		c.setCurrentItem(entry, nil)
	}
}

// setCurrentItem updates the current item, generating Enter/Leave events
// as needed.
func (c *Canvas) setCurrentItem(entry *itemEntry, triggerEvent *event.Event) {
	old := c.currentItem

	// Dispatch Leave to old item while "current" still points to it.
	if old != nil {
		c.dispatchToItem(old, &event.Event{Type: event.LeaveType})
	}

	// Update current item and dispatch Enter to new item.
	c.currentItem = entry
	if entry != nil {
		c.dispatchToItem(entry, &event.Event{Type: event.EnterType})
	}
	// -activefill and friends depend on which item is current.
	if hasActive(old) {
		c.redrawItems(old)
	}
	if hasActive(entry) {
		c.redrawItems(entry)
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
func (c *Canvas) BindItem(tagOrID string, mask event.Mask, handler func(*event.Event)) {
	h := itemHandler{mask: mask, handler: handler}
	if id, err := strconv.ParseInt(tagOrID, 10, 64); err == nil {
		c.idBindings[id] = append(c.idBindings[id], h)
		return
	}
	c.itemBindings[tagOrID] = append(c.itemBindings[tagOrID], h)
}
