package ttk

import (
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
)

// --- Hit Testing ---

type hitRegion int

const (
	hitNothing hitRegion = iota
	hitHeading
	hitSeparator
	hitIndicator
	hitTree
	hitCell
)

type hitResult struct {
	region  hitRegion
	itemID  string
	colIdx  int // -1 for tree column
	dispIdx int // display list index
}

func (tv *Treeview) hitTest(x, y int) hitResult {
	// Heading area.
	if tv.showHeadings && y < tv.headingHeight {
		colX := 0
		colIdx := -1
		if tv.showTree {
			if x >= tv.treeColumnWidth-3 && x <= tv.treeColumnWidth+1 {
				return hitResult{region: hitSeparator, colIdx: -1}
			}
			if x < tv.treeColumnWidth {
				return hitResult{region: hitHeading, colIdx: -1}
			}
			colX = tv.treeColumnWidth
		}
		for ci, col := range tv.columns {
			colIdx = ci
			nextX := colX + col.Width
			if x >= nextX-3 && x <= nextX+1 {
				return hitResult{region: hitSeparator, colIdx: colIdx}
			}
			if x >= colX && x < nextX {
				return hitResult{region: hitHeading, colIdx: colIdx}
			}
			colX = nextX
		}
		return hitResult{region: hitHeading, colIdx: colIdx}
	}

	// Item area.
	itemAreaY := tv.headerOffset()
	if y < itemAreaY {
		return hitResult{region: hitNothing}
	}

	row := (y - itemAreaY) / tv.rowHeight
	dispIdx := tv.topIndex + row
	if dispIdx >= len(tv.displayList) {
		return hitResult{region: hitNothing}
	}

	entry := tv.displayList[dispIdx]
	item := entry.item
	depth := entry.depth

	// Check tree column.
	if tv.showTree && x < tv.treeColumnWidth {
		indentX := depth * tv.indent
		indicatorEnd := indentX + tv.indent
		if x >= indentX && x < indicatorEnd && len(item.Children) > 0 {
			return hitResult{region: hitIndicator, itemID: item.ID, colIdx: -1, dispIdx: dispIdx}
		}
		return hitResult{region: hitTree, itemID: item.ID, colIdx: -1, dispIdx: dispIdx}
	}

	// Check data columns.
	colX := 0
	if tv.showTree {
		colX = tv.treeColumnWidth
	}
	for ci, col := range tv.columns {
		if x >= colX && x < colX+col.Width {
			return hitResult{region: hitCell, itemID: item.ID, colIdx: ci, dispIdx: dispIdx}
		}
		colX += col.Width
	}

	return hitResult{region: hitCell, itemID: item.ID, colIdx: -1, dispIdx: dispIdx}
}

// --- Event Bindings ---

func bindTreeview(tv *Treeview, app widget.AppContext) {
	win := tv.Win

	// Expose.
	app.Dispatcher().Bind(win.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		tv.Display()
	})

	// Configure (resize).
	app.Dispatcher().Bind(win.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			win.Width = ev.ConfigWidth
			win.Height = ev.ConfigHeight
			tv.notifyYScrollbar()
			tv.Display()
		}
	})

	// Focus.
	app.Dispatcher().Bind(win.PlatformID, event.FocusChangeMask, func(ev *event.Event) {
		if ev.Type == event.FocusInType {
			tv.hasFocus = true
			tv.Display()
		} else if ev.Type == event.FocusOutType {
			tv.hasFocus = false
			tv.Display()
		}
	})

	// Button press.
	app.Dispatcher().Bind(win.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		// Take focus.
		app.Server().SetInputFocus(win.PlatformID, platform.RevertToParent, platform.CurrentTime)

		if ev.Button == 1 {
			hit := tv.hitTest(ev.X, ev.Y)
			switch hit.region {
			case hitSeparator:
				// Start column resize.
				tv.resizeCol = hit.colIdx
				tv.resizeDragX = ev.X
			case hitHeading:
				// Heading click.
				if hit.colIdx >= 0 && hit.colIdx < len(tv.columns) {
					col := tv.columns[hit.colIdx]
					if col.HeadingCommand != nil {
						col.HeadingCommand()
					}
				}
			case hitIndicator:
				// Toggle open/close.
				if item := tv.items[hit.itemID]; item != nil {
					tv.SetItemOpen(hit.itemID, !item.Open)
				}
			case hitTree, hitCell:
				// Select item; detect double-click to toggle open.
				if hit.itemID != "" {
					isDouble := hit.itemID == tv.lastClickItem &&
						ev.Time-tv.lastClickTime < 500
					tv.lastClickTime = ev.Time
					tv.lastClickItem = hit.itemID
					tv.handleSelect(hit.itemID, hit.dispIdx, ev.State)
					tv.focus = hit.itemID
					tv.Display()
					if isDouble {
						if item := tv.items[hit.itemID]; item != nil && len(item.Children) > 0 {
							tv.SetItemOpen(hit.itemID, !item.Open)
						}
						if tv.OnDoubleClick != nil {
							tv.OnDoubleClick(hit.itemID)
						}
					}
				}
			}
		} else if ev.Button == 4 {
			tv.YView(tv.topIndex - 3)
		} else if ev.Button == 5 {
			tv.YView(tv.topIndex + 3)
		}
	})

	// Motion (column resize drag).
	app.Dispatcher().Bind(win.PlatformID, event.MotionMask, func(ev *event.Event) {
		if tv.resizeCol < 0 {
			return
		}
		delta := ev.X - tv.resizeDragX
		if tv.resizeCol == -1 {
			// Tree column resize.
			newW := tv.treeColumnWidth + delta
			if newW < 20 {
				newW = 20
			}
			tv.treeColumnWidth = newW
		} else if tv.resizeCol < len(tv.columns) {
			col := tv.columns[tv.resizeCol]
			newW := col.Width + delta
			if newW < col.MinWidth {
				newW = col.MinWidth
			}
			col.Width = newW
		}
		tv.resizeDragX = ev.X
		tv.Display()
	})

	// Button release.
	app.Dispatcher().Bind(win.PlatformID, event.ButtonReleaseMask, func(ev *event.Event) {
		if ev.Button == 1 {
			tv.resizeCol = -1
		}
	})

	// Keyboard.
	app.Dispatcher().Bind(win.PlatformID, event.KeyPressMask, func(ev *event.Event) {
		ks := ev.KeySym
		switch {
		case ks == platform.XK_Up:
			tv.moveFocus(-1)
		case ks == platform.XK_Down:
			tv.moveFocus(1)
		case ks == platform.XK_Left:
			// Collapse current or move to parent.
			if item := tv.items[tv.focus]; item != nil {
				if item.Open && len(item.Children) > 0 {
					tv.SetItemOpen(tv.focus, false)
				} else if item.Parent != nil && item.Parent.ID != "" {
					tv.focus = item.Parent.ID
					tv.SelectionSet(tv.focus)
					tv.See(tv.focus)
				}
			}
		case ks == platform.XK_Right:
			// Expand current or move to first child.
			if item := tv.items[tv.focus]; item != nil {
				if !item.Open && len(item.Children) > 0 {
					tv.SetItemOpen(tv.focus, true)
				} else if len(item.Children) > 0 {
					tv.focus = item.Children[0].ID
					tv.SelectionSet(tv.focus)
					tv.See(tv.focus)
				}
			}
		case ks == platform.XK_Return || ks == platform.XK_space:
			if item := tv.items[tv.focus]; item != nil && len(item.Children) > 0 {
				tv.SetItemOpen(tv.focus, !item.Open)
			}
		case ks == platform.XK_Home:
			if len(tv.displayList) > 0 {
				tv.focus = tv.displayList[0].item.ID
				tv.SelectionSet(tv.focus)
				tv.See(tv.focus)
			}
		case ks == platform.XK_End:
			if len(tv.displayList) > 0 {
				tv.focus = tv.displayList[len(tv.displayList)-1].item.ID
				tv.SelectionSet(tv.focus)
				tv.See(tv.focus)
			}
		}
	})

	// Enter/Leave for hover state.
	app.Dispatcher().Bind(win.PlatformID, event.EnterMask, func(ev *event.Event) {
		tv.ChangeState(StateHover|StateActive, 0)
	})
	app.Dispatcher().Bind(win.PlatformID, event.LeaveMask, func(ev *event.Event) {
		tv.ChangeState(0, StateHover|StateActive|StatePressed)
	})
}

func (tv *Treeview) handleSelect(id string, dispIdx int, state uint) {
	switch tv.selectMode {
	case TreeSelectNone:
		return
	case TreeSelectBrowse:
		tv.selection = map[string]bool{id: true}
		tv.selAnchor = dispIdx
	case TreeSelectExtended:
		shift := state&platform.ShiftMask != 0
		ctrl := state&platform.ControlMask != 0
		if shift && tv.selAnchor >= 0 {
			tv.selection = make(map[string]bool)
			lo, hi := tv.selAnchor, dispIdx
			if lo > hi {
				lo, hi = hi, lo
			}
			for i := lo; i <= hi; i++ {
				if i < len(tv.displayList) {
					tv.selection[tv.displayList[i].item.ID] = true
				}
			}
		} else if ctrl {
			if tv.selection[id] {
				delete(tv.selection, id)
			} else {
				tv.selection[id] = true
			}
			tv.selAnchor = dispIdx
		} else {
			tv.selection = map[string]bool{id: true}
			tv.selAnchor = dispIdx
		}
	}
	if tv.OnSelect != nil {
		tv.OnSelect()
	}
}

func (tv *Treeview) moveFocus(delta int) {
	if len(tv.displayList) == 0 {
		return
	}
	idx := tv.displayIndex(tv.focus)
	if idx < 0 {
		idx = 0
	} else {
		idx += delta
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= len(tv.displayList) {
		idx = len(tv.displayList) - 1
	}
	tv.focus = tv.displayList[idx].item.ID
	tv.SelectionSet(tv.focus)
	tv.See(tv.focus)
}
