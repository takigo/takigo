// Package widget default values ported from tk/unix/tkUnixDefault.h.
package widget

// Default color strings used by InitBase.
const (
	DefBackground        = "#d9d9d9"
	DefForeground        = "#000000"
	DefActiveBackground  = "#ececec"
	DefActiveForeground  = "#000000"
	DefDisabledForeground = "#a3a3a3"
	DefSelectColor       = "#ffffff"
	DefHighlightColor    = "#000000"
	DefHighlightBg       = "#d9d9d9"
	DefInsertBackground  = "#000000"
)

// Default dimensions (in pixels).
const (
	DefBorderWidth     = 1
	DefHighlightWidth  = 1
	DefPadX            = 1
	DefPadY            = 1
)

// State represents a widget's interaction state.
type State int

const (
	StateNormal   State = iota
	StateActive         // mouse is over the widget
	StateDisabled       // widget is not interactive
)
