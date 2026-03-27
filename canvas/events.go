package canvas

import (
	"unicode/utf8"

	"github.com/msorc/takigo/event"
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
			c.scheduleRedraw()
			c.notifyScrollbars()
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
			// Insert printable character.
			if ev.KeySym >= 0x20 && ev.KeySym < 0x7f {
				ti.InsertText(ti.cursorPos, string(rune(ev.KeySym)))
			} else if ev.KeySym > 0x7f && ev.KeySym < 0x10ffff {
				ti.InsertText(ti.cursorPos, string(rune(ev.KeySym)))
			}
		}
		c.scheduleRedraw()
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
}

// dispatchItemEvent dispatches an event to the current item's bindings.
func (c *Canvas) dispatchItemEvent(ev *event.Event) {
	if c.currentItem == nil {
		return
	}
	c.dispatchToItem(c.currentItem, ev)
}

// dispatchToItem dispatches an event to all matching bindings for an item.
func (c *Canvas) dispatchToItem(entry *itemEntry, ev *event.Event) {
	evMask := event.TypeToMask(ev.Type)
	if evMask == 0 {
		return
	}

	// Check bindings by item ID.
	idKey := itemBindKey(entry.id)
	if handlers, ok := c.itemBindings[idKey]; ok {
		for _, h := range handlers {
			if h.mask&evMask != 0 {
				h.handler(ev)
			}
		}
	}

	// Check bindings by tag.
	if base := itemBase(entry.item); base != nil {
		for _, tag := range base.Tags {
			if handlers, ok := c.itemBindings[tag]; ok {
				for _, h := range handlers {
					if h.mask&evMask != 0 {
						h.handler(ev)
					}
				}
			}
		}
	}
}

// itemBindKey returns the binding map key for an item ID.
func itemBindKey(id int64) string {
	return "#" + itoa(id)
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 20)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		buf = append(buf, byte('0'+n%10))
		n /= 10
	}
	if neg {
		buf = append(buf, '-')
	}
	// Reverse.
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}

// BindItem binds an event handler to items matching tagOrID.
func (c *Canvas) BindItem(tagOrID string, mask event.Mask, handler func(*event.Event)) {
	// For numeric IDs, use the "#ID" key format.
	key := tagOrID
	if _, err := parseInt64(tagOrID); err == nil {
		key = "#" + tagOrID
	}
	c.itemBindings[key] = append(c.itemBindings[key], itemHandler{
		mask:    mask,
		handler: handler,
	})
}

func parseInt64(s string) (int64, error) {
	var n int64
	neg := false
	i := 0
	if len(s) > 0 && s[0] == '-' {
		neg = true
		i = 1
	}
	if i >= len(s) {
		return 0, errNotInt
	}
	for ; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, errNotInt
		}
		n = n*10 + int64(s[i]-'0')
	}
	if neg {
		n = -n
	}
	return n, nil
}

type parseError struct{}

func (parseError) Error() string { return "not an integer" }

var errNotInt = parseError{}
