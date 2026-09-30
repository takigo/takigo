package text

import "github.com/msorc/takigo/screenunit"

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
	return func(t *TextWidget) { t.SetBackgroundName(name) }
}

// Foreground sets the text foreground color.
func Foreground(name string) TextOption {
	return func(t *TextWidget) { t.SetForegroundName(name) }
}

// FontOpt sets the font.
func FontOpt(name string) TextOption {
	return func(t *TextWidget) { t.SetFontName(name) }
}

// BorderWidthOpt sets the border width.
func BorderWidthOpt(w any) TextOption {
	return func(t *TextWidget) { t.BorderWidth = screenunit.PxOr(w, t.BorderWidth) }
}

// HighlightThickness sets -highlightthickness.
func HighlightThickness(n any) TextOption {
	return func(t *TextWidget) { t.HighlightWidth = screenunit.PxOr(n, t.HighlightWidth) }
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
func PadXOpt(n any) TextOption {
	return func(t *TextWidget) { t.PadX = screenunit.PxOr(n, t.PadX) }
}

// PadYOpt sets vertical padding between the border and the text content.
func PadYOpt(n any) TextOption {
	return func(t *TextWidget) { t.PadY = screenunit.PxOr(n, t.PadY) }
}

// --- Ttk-compatible aliases (prefix with Text) for consistent naming ---
// These aliases match the naming convention used by ttk widgets
// allowing consistent option naming when both classic and ttk widgets are used.

// TextWidth is an alias for Width.
var TextWidth = Width

// TextHeight is an alias for Height.
var TextHeight = Height

// TextWrapModeOpt is an alias for WrapModeOpt.
var TextWrapModeOpt = WrapModeOpt

// TextBackground is an alias for Background.
var TextBackground = Background

// TextForeground is an alias for Foreground.
var TextForeground = Foreground

// TextFontOpt is an alias for FontOpt.
var TextFontOpt = FontOpt

// TextBorderWidthOpt is an alias for BorderWidthOpt.
var TextBorderWidthOpt = BorderWidthOpt

// TextTabWidth is an alias for TabWidth.
var TextTabWidth = TabWidth

// TextUndoOpt is an alias for UndoOpt.
var TextUndoOpt = UndoOpt

// TextYScrollCommand is an alias for YScrollCommand.
var TextYScrollCommand = YScrollCommand

// TextXScrollCommand is an alias for XScrollCommand.
var TextXScrollCommand = XScrollCommand

// TextInsertWidth is an alias for InsertWidth.
var TextInsertWidth = InsertWidth

// TextReadOnly is an alias for ReadOnly.
var TextReadOnly = ReadOnly

// TextSetGridOpt is an alias for SetGridOpt.
var TextSetGridOpt = SetGridOpt

// TextPadXOpt is an alias for PadXOpt.
var TextPadXOpt = PadXOpt

// TextPadYOpt is an alias for PadYOpt.
var TextPadYOpt = PadYOpt
