// Package darktheme registers the "dark" TTK theme: clam's layouts and
// elements in a dark palette. It has no counterpart in Tk.
// Import with blank identifier to auto-register: _ "github.com/takigo/takigo/ttk/darktheme"
package darktheme

import (
	"github.com/takigo/takigo/ttk"
	"github.com/takigo/takigo/ttk/clamtheme"
)

// Palette is the dark theme's colours.
var Palette = clamtheme.Palette{
	Frame:                0x2e2e2e,
	Dark:                 0x262626,
	Darker:               0x1f1f1f,
	Darkest:              0x141414,
	Lighter:              0x3c3c3c,
	Light:                0x505050,
	Foreground:           0xe6e6e6,
	DisabledFg:           0x808080,
	Field:                0x1e1e1e,
	Select:               0x3d6f99,
	SelectForeground:     0xffffff,
	HeadingHover:         0x3c3c3c,
	AltIndicator:         0x5895bc,
	DisabledAltIndicator: 0x5a5a5a,
}

func init() {
	ttk.RegisterTheme(clamtheme.New("dark", Palette))
}
