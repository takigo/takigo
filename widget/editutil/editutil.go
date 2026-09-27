// Package editutil provides shared editing utilities for text-like widgets
// (entry, spinbox, text). It encapsulates clipboard and key handling logic
// to reduce duplication across widget types.
package editutil

import (
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// ClipboardKeySym returns true if the keysym is a clipboard operation
// that should be handled by the shared logic.
func ClipboardKeySym(keysym platform.KeySym, ctrl, shift bool) (op ClipboardOp, handled bool) {
	switch {
	case ctrl && keysym == platform.XK_c:
		return ClipCopy, true
	case ctrl && keysym == platform.XK_x:
		return ClipCut, true
	case ctrl && keysym == platform.XK_v:
		return ClipPaste, true
	case ctrl && keysym == platform.XK_Insert:
		return ClipCopy, true
	case shift && keysym == platform.XK_Insert:
		return ClipPaste, true
	case shift && keysym == platform.XK_Delete:
		return ClipCut, true
	}
	return ClipNone, false
}

type ClipboardOp int

const (
	ClipNone ClipboardOp = iota
	ClipCopy
	ClipCut
	ClipPaste
)

// Clipboardable is the minimal interface for clipboard operations.
type Clipboardable interface {
	GetSelection() string
	ReadOnly() bool
	App() widget.AppContext
	Win() window.Windower
	DeleteSelection()
	SeeInsert()
	NotifyYScrollbar()
	Display()
}

// HandleClipboard performs the clipboard operation.
func HandleClipboard(c Clipboardable, op ClipboardOp, ev *event.Event) {
	switch op {
	case ClipCopy:
		if sel := c.GetSelection(); sel != "" {
			c.App().Clipboard().Set(c.Win().Window().PlatformID, sel, platform.Timestamp(ev.Time))
		}
	case ClipCut:
		if c.ReadOnly() {
			return
		}
		if sel := c.GetSelection(); sel != "" {
			c.App().Clipboard().Set(c.Win().Window().PlatformID, sel, platform.Timestamp(ev.Time))
			c.DeleteSelection()
			c.SeeInsert()
			c.NotifyYScrollbar()
			c.Display()
		}
	case ClipPaste:
		if c.ReadOnly() {
			return
		}
		c.App().Clipboard().Get(c.Win().Window().PlatformID, platform.Timestamp(ev.Time), func(text string) {
			if text == "" {
				return
			}
			// Widget-specific paste logic must be handled by caller
			// after this returns, or via a callback
		})
	}
}

// EmacsKeySym maps Emacs-style control keys to navigation actions.
func EmacsKeySym(keysym platform.KeySym, ctrl bool) (action EmacsAction, handled bool) {
	if !ctrl {
		return EmacsNone, false
	}
	switch keysym {
	case platform.XK_a:
		return EmacsStartOfLine, true
	case platform.XK_e:
		return EmacsEndOfLine, true
	case platform.XK_b:
		return EmacsBackwardChar, true
	case platform.XK_f:
		return EmacsForwardChar, true
	case platform.XK_p:
		return EmacsPreviousLine, true
	case platform.XK_n:
		return EmacsNextLine, true
	case platform.XK_d:
		return EmacsDeleteChar, true
	case platform.XK_k:
		return EmacsKillLine, true
	case platform.XK_w:
		return EmacsKillRegion, true
	case platform.XK_z:
		return EmacsUndo, true
	case platform.XK_y:
		return EmacsRedo, true // Note: XK_y is both kill-region and redo in original
	}
	return EmacsNone, false
}

type EmacsAction int

const (
	EmacsNone EmacsAction = iota
	EmacsStartOfLine
	EmacsEndOfLine
	EmacsBackwardChar
	EmacsForwardChar
	EmacsPreviousLine
	EmacsNextLine
	EmacsDeleteChar
	EmacsKillLine
	EmacsKillRegion
	EmacsUndo
	EmacsRedo
)

// Navigator interface for movement operations.
type Navigator interface {
	Get() string
	InsertPos() int
	SetInsertPos(int)
	MoveCursor(newPos, anchor int, shift bool)
	ReadOnly() bool
	DeleteRange(first, last int) bool
	SeeInsert()
	NotifyYScrollbar()
	Display()
	App() widget.AppContext
	Win() window.Windower
	GetSelection() string
	DeleteSelection()
}

// HandleEmacsKey performs the emacs-style key action for single-line widgets.
func HandleEmacsKey(n Navigator, action EmacsAction, shift bool) {
	switch action {
	case EmacsStartOfLine:
		n.MoveCursor(0, n.InsertPos(), shift)
	case EmacsEndOfLine:
		n.MoveCursor(len(n.Get()), n.InsertPos(), shift)
	case EmacsBackwardChar:
		n.MoveCursor(n.InsertPos()-1, n.InsertPos(), shift)
	case EmacsForwardChar:
		n.MoveCursor(n.InsertPos()+1, n.InsertPos(), shift)
	case EmacsDeleteChar:
		if n.ReadOnly() {
			return
		}
		pos := n.InsertPos()
		if pos < len(n.Get()) {
			n.DeleteRange(pos, pos+1)
			n.SeeInsert()
			n.NotifyYScrollbar()
			n.Display()
		}
	case EmacsKillLine:
		if n.ReadOnly() {
			return
		}
		pos := n.InsertPos()
		if pos < len(n.Get()) {
			n.DeleteRange(pos, len(n.Get()))
			n.SeeInsert()
			n.NotifyYScrollbar()
			n.Display()
		}
	case EmacsKillRegion:
		if n.ReadOnly() {
			return
		}
		if sel := n.GetSelection(); sel != "" {
			n.App().Clipboard().Set(n.Win().Window().PlatformID, sel, platform.Timestamp(0))
			n.DeleteSelection()
			n.SeeInsert()
			n.NotifyYScrollbar()
			n.Display()
		}
	case EmacsUndo:
		// Widget-specific
	case EmacsRedo:
		// Widget-specific
	}
}
