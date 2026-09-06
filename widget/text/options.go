package text

import "log"

// WrapMode specifies how lines are wrapped at widget boundaries.
type WrapMode int

const (
	WrapNone WrapMode = iota // no wrapping, horizontal scrolling
	WrapChar                 // wrap at any character
	WrapWord                 // wrap at word boundaries
)

// TextOption configures a TextWidget.
type TextOption func(*TextWidget)

// Width sets the preferred width in characters.
func Width(w int) TextOption {
	return func(t *TextWidget) { t.prefWidth = w }
}

// Height sets the preferred height in lines.
func Height(h int) TextOption {
	return func(t *TextWidget) { t.prefHeight = h }
}

// WrapModeOpt sets the wrap mode.
func WrapModeOpt(mode WrapMode) TextOption {
	return func(t *TextWidget) { t.wrapMode = mode }
}

// Background sets the background color.
func Background(name string) TextOption {
	return func(t *TextWidget) {
		col, err := t.App.ColorCache().Get(name)
		if err != nil {
			log.Printf("text: failed to get color %q: %v", name, err)
			return
		}
		t.Background = col
		t.UpdateBorder()
	}
}

// Foreground sets the text foreground color.
func Foreground(name string) TextOption {
	return func(t *TextWidget) {
		col, err := t.App.ColorCache().Get(name)
		if err != nil {
			log.Printf("text: failed to get color %q: %v", name, err)
			return
		}
		t.Foreground = col
	}
}

// FontOpt sets the font.
func FontOpt(name string) TextOption {
	return func(t *TextWidget) {
		f, err := t.App.FontRegistry().Get(name)
		if err != nil {
			log.Printf("text: failed to get font %q: %v", name, err)
			return
		}
		t.Font = f
	}
}

// BorderWidthOpt sets the border width.
func BorderWidthOpt(w int) TextOption {
	return func(t *TextWidget) { t.BorderWidth = w }
}

// TabWidth sets the tab width in characters.
func TabWidth(w int) TextOption {
	return func(t *TextWidget) { t.tabWidth = w }
}

// UndoOpt enables or disables undo/redo.
func UndoOpt(enabled bool) TextOption {
	return func(t *TextWidget) { t.undoEnabled = enabled }
}

// YScrollCommand sets the vertical scroll callback.
func YScrollCommand(fn func(first, last float64)) TextOption {
	return func(t *TextWidget) { t.YScrollCmd = fn }
}

// XScrollCommand sets the horizontal scroll callback.
func XScrollCommand(fn func(first, last float64)) TextOption {
	return func(t *TextWidget) { t.XScrollCmd = fn }
}

// InsertWidth sets the cursor width in pixels.
func InsertWidth(w int) TextOption {
	return func(t *TextWidget) { t.insertWidth = w }
}

// ReadOnly sets the text widget to read-only mode. Navigation and selection
// still work, but text insertion and deletion are blocked.
func ReadOnly(on bool) TextOption {
	return func(t *TextWidget) { t.readOnly = on }
}

// SetGridOpt enables or disables setgrid mode. When enabled, the nearest
// toplevel window's resize increments are set to the character cell size so
// that the window resizes in whole character increments (matching Tk's -setgrid).
func SetGridOpt(on bool) TextOption {
	return func(t *TextWidget) { t.setGrid = on }
}

// PadXOpt sets horizontal padding between the border and the text content.
func PadXOpt(n int) TextOption {
	return func(t *TextWidget) {
		t.PadX = n
		t.insetX = t.inset + t.PadX
	}
}

// PadYOpt sets vertical padding between the border and the text content.
func PadYOpt(n int) TextOption {
	return func(t *TextWidget) {
		t.PadY = n
		t.insetY = t.inset + t.PadY
	}
}
