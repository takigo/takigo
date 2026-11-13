// Treeview bindings: tk/library/ttk/treeview.tcl.
package ttk

import (
	"strconv"

	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/widget"
)

// treePressMode is State(pressMode) in treeview.tcl.
type treePressMode int

const (
	pressNone treePressMode = iota
	pressResize
	pressHeading
)

// treeviewPress is the per-widget press state of treeview.tcl.
type treeviewPress struct {
	mode    treePressMode
	heading *TreeColumn // pressHeading: the pressed heading
	pressed bool        // the pointer is still on it
}

// bindTreeview registers all event handlers for the treeview widget.
func bindTreeview(tv *Treeview, app widget.AppContext) {
	win := tv.Win
	d := app.Dispatcher()

	// Item tag bindings run before the class bindings, like Tk's event
	// handler that runs before Tk_BindEvent.
	d.Bind(win.PlatformID, event.KeyPressMask|event.KeyReleaseMask|event.ButtonPressMask|
		event.ButtonReleaseMask|event.MotionMask|event.VirtualMask, tv.dispatchTagEvent)

	d.Bind(win.PlatformID, event.StructureNotifyMask, tv.handleConfigure)
	d.Bind(win.PlatformID, event.FocusChangeMask, tv.handleFocus)

	// TtkScrollable <MouseWheel> and <Shift-MouseWheel> (utils.tcl).
	var wheel event.WheelAccumulator
	d.Bind(win.PlatformID, event.MouseWheelMask, func(ev *event.Event) {
		if ev.Type != event.MouseWheelType {
			return
		}
		n := wheel.Units(ev.Delta, 3)
		switch {
		case n == 0:
		case ev.State&platform.ShiftMask != 0:
			// x units are pixels.
			tv.XViewScroll(n, false)
		default:
			tv.YView(tv.topIndex + n)
		}
	})

	d.Bind(win.PlatformID, event.ButtonPressMask, tv.handleButtonPress)
	d.Bind(win.PlatformID, event.MotionMask, tv.handleMotion)
	d.Bind(win.PlatformID, event.ButtonReleaseMask, tv.handleButtonRelease)
	d.Bind(win.PlatformID, event.KeyPressMask, tv.handleKeyPress)
	d.Bind(win.PlatformID, event.EnterMask, tv.handleEnter)
	d.Bind(win.PlatformID, event.LeaveMask, tv.handleLeave)
}

// handleConfigure handles ConfigureNotify (resize) events.
func (tv *Treeview) handleConfigure(ev *event.Event) {
	if ev.Type == event.ConfigureType {
		// The geometry manager already set Width/Height; a queued
		// ConfigureNotify may carry a stale size. Like Tk, redisplay
		// (and so ResizeColumns) at idle, once the layout has settled.
		tv.scheduleRedisplay()
	}
}

// handleFocus handles FocusIn/FocusOut events.
func (tv *Treeview) handleFocus(ev *event.Event) {
	switch ev.Type {
	case event.FocusInType:
		tv.hasFocus = true
		tv.scheduleRedisplay()
	case event.FocusOutType:
		tv.hasFocus = false
		tv.scheduleRedisplay()
	}
}

// handleEnter handles Enter events (hover state).
func (tv *Treeview) handleEnter(ev *event.Event) {
	tv.ChangeState(StateHover|StateActive, 0)
}

// handleLeave handles Leave events (hover state).
func (tv *Treeview) handleLeave(ev *event.Event) {
	tv.ChangeState(0, StateHover|StateActive|StatePressed)
}

// handleButtonPress ports Press, DoubleClick, <Shift-Button-1> (extend)
// and <Control-Button-1> (<<ToggleSelection>>, toggle).
func (tv *Treeview) handleButtonPress(ev *event.Event) {
	if ev.Button != 1 {
		return
	}
	widget.Focus(tv.App, tv.Win)
	x, y := ev.X, ev.Y
	id := tv.IdentifyItem(y)

	isDouble := id != "" && id == tv.lastClickItem && ev.Time-tv.lastClickTime < 500
	tv.lastClickTime, tv.lastClickItem = ev.Time, id
	if isDouble {
		tv.lastClickItem = ""
		tv.toggle(id)
		if tv.OnDoubleClick != nil {
			tv.OnDoubleClick(id)
		}
		return
	}

	switch {
	case ev.State&platform.ShiftMask != 0:
		if id != "" {
			tv.selectOp(id, selectExtend)
		}
		return
	case ev.State&platform.ControlMask != 0:
		if id != "" {
			tv.selectOp(id, selectToggle)
		}
		return
	}

	switch tv.IdentifyRegion(x, y) {
	case RegionHeading:
		col, _ := tv.columnAt(x)
		tv.press = treeviewPress{mode: pressHeading, heading: col, pressed: true}
	case RegionSeparator:
		col, _ := tv.columnAt(x)
		tv.press = treeviewPress{mode: pressResize, heading: col}
	case RegionTree, RegionCell:
		tv.selectOp(id, selectChoose)
		if tv.IdentifyElement(x, y) == "Treeitem.indicator" {
			tv.toggle(id)
		}
	default:
	}
}

// handleMotion ports Drag: resize.drag and heading.drag.
func (tv *Treeview) handleMotion(ev *event.Event) {
	switch tv.press.mode {
	case pressResize:
		tv.DragColumn("#"+strconv.Itoa(tv.displayNumber(tv.press.heading)), ev.X)
	case pressHeading:
		col, _ := tv.columnAt(ev.X)
		tv.press.pressed = tv.IdentifyRegion(ev.X, ev.Y) == RegionHeading && col == tv.press.heading
	default:
	}
}

// handleButtonRelease ports Release: a heading's command runs at idle
// when the pointer is still on it.
func (tv *Treeview) handleButtonRelease(ev *event.Event) {
	if ev.Button != 1 {
		return
	}
	p := tv.press
	tv.press = treeviewPress{}
	switch p.mode {
	case pressResize:
		tv.Drop()
	case pressHeading:
		if p.pressed && p.heading != nil && p.heading.HeadingCommand != nil {
			tv.App.DoWhenIdle(p.heading.HeadingCommand)
		}
	default:
	}
}

// handleKeyPress ports the key bindings: Keynav, <Prior>/<Next>,
// <Return>/<space> (ToggleFocus).
func (tv *Treeview) handleKeyPress(ev *event.Event) {
	switch ev.KeySym {
	case platform.XK_Up, platform.XK_Down, platform.XK_Left, platform.XK_Right:
		tv.keynav(ev.KeySym)
	case platform.XK_Prior:
		tv.YViewScroll(-1, true)
	case platform.XK_Next:
		tv.YViewScroll(1, true)
	case platform.XK_Return, platform.XK_space:
		if tv.focus != "" {
			tv.toggle(tv.focus)
		}
	case platform.XK_Home:
		if tv.syncDisplayList(); len(tv.displayList) > 0 {
			tv.selectOp(tv.displayList[0].item.ID, selectChoose)
		}
	case platform.XK_End:
		if tv.syncDisplayList(); len(tv.displayList) > 0 {
			tv.selectOp(tv.displayList[len(tv.displayList)-1].item.ID, selectChoose)
		}
	}
}

// keynav ports Keynav for -selecttype item.
func (tv *Treeview) keynav(dir platform.KeySym) {
	focus := tv.focus
	if focus == "" {
		return
	}
	isOpen := func(id string) bool {
		item := tv.items[id]
		return item != nil && item.Open && len(item.Children) > 0
	}
	switch dir {
	case platform.XK_Up:
		if up := tv.Prev(focus); up == "" {
			focus = tv.Parent(focus)
		} else {
			for isOpen(up) {
				c := tv.items[up].Children
				up = c[len(c)-1].ID
			}
			focus = up
		}
	case platform.XK_Down:
		if isOpen(focus) {
			focus = tv.items[focus].Children[0].ID
		} else {
			up, down := focus, ""
			for up != "" {
				if down = tv.Next(up); down != "" {
					break
				}
				up = tv.Parent(up)
			}
			focus = down
		}
	case platform.XK_Left:
		if isOpen(focus) {
			tv.closeItem(focus)
		} else {
			focus = tv.Parent(focus)
		}
	case platform.XK_Right:
		tv.openItem(focus)
	}
	if focus != "" {
		tv.selectOp(focus, selectChoose)
	}
}

// openItem ports OpenItem: focus the item, send <<TreeviewOpen>>, open it.
func (tv *Treeview) openItem(id string) {
	tv.SetFocus(id)
	tv.generate(TreeviewOpenEvent)
	tv.SetItemOpen(id, true)
}

// closeItem ports CloseItem: close the item, focus it, send <<TreeviewClose>>.
func (tv *Treeview) closeItem(id string) {
	tv.SetItemOpen(id, false)
	tv.SetFocus(id)
	tv.generate(TreeviewCloseEvent)
}

// toggle ports Toggle: open or close an item that has children.
func (tv *Treeview) toggle(id string) {
	item := tv.items[id]
	if item == nil || len(item.Children) == 0 {
		return
	}
	if item.Open {
		tv.closeItem(id)
	} else {
		tv.openItem(id)
	}
}

type selectOperation int

const (
	selectChoose selectOperation = iota
	selectExtend
	selectToggle
)

// selectOp ports SelectOp and the select.<op>.<selectmode> procedures.
func (tv *Treeview) selectOp(id string, op selectOperation) {
	switch {
	case tv.selectMode == TreeSelectNone:
		tv.SetFocus(id)
		tv.See(id)
	case tv.selectMode == TreeSelectBrowse || op == selectChoose:
		tv.browseTo(id)
	case op == selectToggle:
		tv.SelectionToggle(id)
	default: // extend
		if anchor := tv.focus; anchor != "" && tv.items[anchor] != nil {
			tv.SelectionSet(tv.between(anchor, id)...)
		} else {
			tv.browseTo(id)
		}
	}
}

// browseTo ports BrowseTo: see, focus and select the item.
func (tv *Treeview) browseTo(id string) {
	tv.See(id)
	tv.SetFocus(id)
	tv.SelectionSet(id)
}

// between ports ttk::treeview::between: the items from a to b inclusive,
// in preorder over the whole tree, closed items' descendants included.
func (tv *Treeview) between(a, b string) []string {
	var out []string
	in := false
	tv.preorder(func(item *TreeItem) {
		switch {
		case item.ID == a || item.ID == b:
			out = append(out, item.ID)
			if a != b {
				in = !in
			}
		case in:
			out = append(out, item.ID)
		}
	})
	return out
}
