// Package bind provides a Tk-compatible event binding engine with pattern
// matching, tag-based dispatch chains, virtual events, and double-click support.
package bind

import "github.com/msorc/takigo/internal/xlib"

// keysymNames maps lowercase keysym name strings to xlib.KeySym values.
// Used by the pattern parser to resolve key names in patterns like <Key-Return>.
var keysymNames = map[string]xlib.KeySym{
	// Letters (lowercase keysyms)
	"a": xlib.KeySym(0x0061),
	"b": xlib.KeySym(0x0062),
	"c": xlib.KeySym(0x0063),
	"d": xlib.KeySym(0x0064),
	"e": xlib.KeySym(0x0065),
	"f": xlib.KeySym(0x0066),
	"g": xlib.KeySym(0x0067),
	"h": xlib.KeySym(0x0068),
	"i": xlib.KeySym(0x0069),
	"j": xlib.KeySym(0x006a),
	"k": xlib.KeySym(0x006b),
	"l": xlib.KeySym(0x006c),
	"m": xlib.KeySym(0x006d),
	"n": xlib.KeySym(0x006e),
	"o": xlib.KeySym(0x006f),
	"p": xlib.KeySym(0x0070),
	"q": xlib.KeySym(0x0071),
	"r": xlib.KeySym(0x0072),
	"s": xlib.KeySym(0x0073),
	"t": xlib.KeySym(0x0074),
	"u": xlib.KeySym(0x0075),
	"v": xlib.KeySym(0x0076),
	"w": xlib.KeySym(0x0077),
	"x": xlib.KeySym(0x0078),
	"y": xlib.KeySym(0x0079),
	"z": xlib.KeySym(0x007a),

	// Digits
	"0": xlib.KeySym(0x0030),
	"1": xlib.KeySym(0x0031),
	"2": xlib.KeySym(0x0032),
	"3": xlib.KeySym(0x0033),
	"4": xlib.KeySym(0x0034),
	"5": xlib.KeySym(0x0035),
	"6": xlib.KeySym(0x0036),
	"7": xlib.KeySym(0x0037),
	"8": xlib.KeySym(0x0038),
	"9": xlib.KeySym(0x0039),

	// Named keys
	"return":    xlib.XK_Return,
	"escape":    xlib.XK_Escape,
	"backspace": xlib.XK_BackSpace,
	"tab":       xlib.XK_Tab,
	"space":     xlib.XK_space,
	"delete":    xlib.XK_Delete,

	// Arrow keys
	"left":  xlib.XK_Left,
	"right": xlib.XK_Right,
	"up":    xlib.XK_Up,
	"down":  xlib.XK_Down,

	// Navigation
	"home":   xlib.XK_Home,
	"end":    xlib.XK_End,
	"prior":  xlib.XK_Prior,
	"next":   xlib.XK_Next,
	"pageup": xlib.XK_Prior,
	"pagedown": xlib.XK_Next,

	// Function keys
	"f1":  xlib.KeySym(0xffbe),
	"f2":  xlib.KeySym(0xffbf),
	"f3":  xlib.KeySym(0xffc0),
	"f4":  xlib.KeySym(0xffc1),
	"f5":  xlib.KeySym(0xffc2),
	"f6":  xlib.KeySym(0xffc3),
	"f7":  xlib.KeySym(0xffc4),
	"f8":  xlib.KeySym(0xffc5),
	"f9":  xlib.KeySym(0xffc6),
	"f10": xlib.KeySym(0xffc7),
	"f11": xlib.KeySym(0xffc8),
	"f12": xlib.KeySym(0xffc9),

	// Misc
	"insert":      xlib.KeySym(0xff63),
	"pause":       xlib.KeySym(0xff13),
	"scroll_lock": xlib.KeySym(0xff14),
	"caps_lock":   xlib.KeySym(0xffe5),
	"num_lock":    xlib.KeySym(0xff7f),
	"print":       xlib.KeySym(0xff61),

	// Punctuation / symbols
	"minus":        xlib.KeySym(0x002d),
	"plus":         xlib.KeySym(0x002b),
	"equal":        xlib.KeySym(0x003d),
	"bracketleft":  xlib.KeySym(0x005b),
	"bracketright": xlib.KeySym(0x005d),
	"backslash":    xlib.KeySym(0x005c),
	"semicolon":    xlib.KeySym(0x003b),
	"apostrophe":   xlib.KeySym(0x0027),
	"comma":        xlib.KeySym(0x002c),
	"period":       xlib.KeySym(0x002e),
	"slash":        xlib.KeySym(0x002f),
	"grave":        xlib.KeySym(0x0060),
	"asciitilde":   xlib.KeySym(0x007e),
	"exclam":       xlib.KeySym(0x0021),
	"at":           xlib.KeySym(0x0040),
	"numbersign":   xlib.KeySym(0x0023),
	"dollar":       xlib.KeySym(0x0024),
	"percent":      xlib.KeySym(0x0025),
	"asciicircum":  xlib.KeySym(0x005e),
	"ampersand":    xlib.KeySym(0x0026),
	"asterisk":     xlib.KeySym(0x002a),
	"parenleft":    xlib.KeySym(0x0028),
	"parenright":   xlib.KeySym(0x0029),
	"underscore":   xlib.KeySym(0x005f),
	"braceleft":    xlib.KeySym(0x007b),
	"braceright":   xlib.KeySym(0x007d),
	"bar":          xlib.KeySym(0x007c),
	"colon":        xlib.KeySym(0x003a),
	"quotedbl":     xlib.KeySym(0x0022),
	"less":         xlib.KeySym(0x003c),
	"greater":      xlib.KeySym(0x003e),
	"question":     xlib.KeySym(0x003f),
}

// lookupKeySym resolves a key name (case-insensitive) to a KeySym.
// Returns 0 if not found.
func lookupKeySym(name string) xlib.KeySym {
	// Try exact lowercase match.
	lower := toLower(name)
	if ks, ok := keysymNames[lower]; ok {
		return ks
	}
	// Single character: use its ASCII value as keysym.
	if len(name) == 1 {
		ch := name[0]
		if ch >= 0x20 && ch <= 0x7e {
			return xlib.KeySym(ch)
		}
	}
	return 0
}

// toLower is a simple ASCII-only lowercase conversion.
func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}
