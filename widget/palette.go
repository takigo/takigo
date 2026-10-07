package widget

import "github.com/takigo/takigo/window"

// Palette is the set of default colours the classic widgets of one App
// start with, as Tk colour strings. LightPalette holds Tk's own defaults
// (tkUnixDefault.h); DarkPalette is takigo's dark counterpart. A widget
// reads its palette when it is created, so choose the palette before
// creating widgets (takigo.UseAppearance, takigo.FollowSystemAppearance).
type Palette struct {
	// Dark tells themed widgets to pick a dark theme too.
	Dark bool

	Background, Foreground             string
	ActiveBackground, ActiveForeground string
	DisabledForeground                 string
	HighlightBackground                string // focus ring without the focus
	HighlightColor                     string // focus ring with the focus
	FieldBackground                    string // entry, text, listbox, spinbox
	SelectBackground, SelectForeground string // selected text and menu entries
	SelectColor                        string // check and radio indicator
	InsertBackground                   string // insertion cursor
	TroughColor                        string // scrollbar and scale trough
}

// LightPalette is Tk's default colours.
var LightPalette = Palette{
	Background:          DefBackground,
	Foreground:          DefForeground,
	ActiveBackground:    DefActiveBackground,
	ActiveForeground:    DefActiveForeground,
	DisabledForeground:  DefDisabledForeground,
	HighlightBackground: DefHighlightBg,
	HighlightColor:      DefHighlightColor,
	FieldBackground:     "#ffffff",
	SelectBackground:    "#3399ff",
	SelectForeground:    "#ffffff",
	SelectColor:         DefSelectColor,
	InsertBackground:    DefInsertBackground,
	TroughColor:         "#c3c3c3",
}

// DarkPalette matches the ttk "dark" theme (ttk/darktheme).
var DarkPalette = Palette{
	Dark:                true,
	Background:          "#2e2e2e",
	Foreground:          "#e6e6e6",
	ActiveBackground:    "#3c3c3c",
	ActiveForeground:    "#ffffff",
	DisabledForeground:  "#808080",
	HighlightBackground: "#2e2e2e",
	HighlightColor:      "#e6e6e6",
	FieldBackground:     "#1e1e1e",
	SelectBackground:    "#3d6f99",
	SelectForeground:    "#ffffff",
	SelectColor:         "#1e1e1e",
	InsertBackground:    "#e6e6e6",
	TroughColor:         "#1f1f1f",
}

var paletteKey = new(window.ValueKey)

// SetPalette gives the App whose root window is root its palette.
func SetPalette(root *window.Window, p Palette) {
	root.SetValue(paletteKey, p)
}

// PaletteFor returns app's palette: LightPalette unless another was set.
func PaletteFor(app AppContext) Palette {
	if app != nil {
		if root := app.Root(); root != nil {
			if p, ok := root.Value(paletteKey).(Palette); ok {
				return p
			}
		}
	}
	return LightPalette
}

// DisabledColor returns the palette's disabled foreground for app, or
// Tk's own if it cannot be resolved.
func DisabledColor(app AppContext) (pixel uint64, r, g, b uint16) {
	if app != nil {
		if c, err := app.ColorCache().Get(PaletteFor(app).DisabledForeground); err == nil {
			return c.Pixel, c.Red, c.Green, c.Blue
		}
	}
	return 0xa3a3a3, 0xa300, 0xa300, 0xa300
}
