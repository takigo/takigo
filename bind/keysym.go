// Package bind provides a Tk-compatible event binding engine with pattern
// matching, tag-based dispatch chains, virtual events, and double-click support.
package bind

import (
	"strconv"

	"github.com/takigo/takigo/platform"
)

// keysymNames maps lowercase keysym name strings to platform.KeySym values.
// Used by the pattern parser to resolve key names in patterns like <Key-Return>.
var keysymNames = map[string]platform.KeySym{
	// Letters (lowercase keysyms)
	"a": platform.KeySym(0x0061),
	"b": platform.KeySym(0x0062),
	"c": platform.KeySym(0x0063),
	"d": platform.KeySym(0x0064),
	"e": platform.KeySym(0x0065),
	"f": platform.KeySym(0x0066),
	"g": platform.KeySym(0x0067),
	"h": platform.KeySym(0x0068),
	"i": platform.KeySym(0x0069),
	"j": platform.KeySym(0x006a),
	"k": platform.KeySym(0x006b),
	"l": platform.KeySym(0x006c),
	"m": platform.KeySym(0x006d),
	"n": platform.KeySym(0x006e),
	"o": platform.KeySym(0x006f),
	"p": platform.KeySym(0x0070),
	"q": platform.KeySym(0x0071),
	"r": platform.KeySym(0x0072),
	"s": platform.KeySym(0x0073),
	"t": platform.KeySym(0x0074),
	"u": platform.KeySym(0x0075),
	"v": platform.KeySym(0x0076),
	"w": platform.KeySym(0x0077),
	"x": platform.KeySym(0x0078),
	"y": platform.KeySym(0x0079),
	"z": platform.KeySym(0x007a),

	// Digits
	"0": platform.KeySym(0x0030),
	"1": platform.KeySym(0x0031),
	"2": platform.KeySym(0x0032),
	"3": platform.KeySym(0x0033),
	"4": platform.KeySym(0x0034),
	"5": platform.KeySym(0x0035),
	"6": platform.KeySym(0x0036),
	"7": platform.KeySym(0x0037),
	"8": platform.KeySym(0x0038),
	"9": platform.KeySym(0x0039),

	// Named keys
	"return":    platform.XK_Return,
	"escape":    platform.XK_Escape,
	"backspace": platform.XK_BackSpace,
	"tab":       platform.XK_Tab,
	"space":     platform.XK_space,
	"delete":    platform.XK_Delete,

	// Arrow keys
	"left":  platform.XK_Left,
	"right": platform.XK_Right,
	"up":    platform.XK_Up,
	"down":  platform.XK_Down,

	// Navigation
	"home":     platform.XK_Home,
	"end":      platform.XK_End,
	"prior":    platform.XK_Prior,
	"next":     platform.XK_Next,
	"pageup":   platform.XK_Prior,
	"pagedown": platform.XK_Next,

	// Function keys
	"f1":  platform.KeySym(0xffbe),
	"f2":  platform.KeySym(0xffbf),
	"f3":  platform.KeySym(0xffc0),
	"f4":  platform.KeySym(0xffc1),
	"f5":  platform.KeySym(0xffc2),
	"f6":  platform.KeySym(0xffc3),
	"f7":  platform.KeySym(0xffc4),
	"f8":  platform.KeySym(0xffc5),
	"f9":  platform.KeySym(0xffc6),
	"f10": platform.KeySym(0xffc7),
	"f11": platform.KeySym(0xffc8),
	"f12": platform.KeySym(0xffc9),

	// Misc
	"insert":      platform.KeySym(0xff63),
	"pause":       platform.KeySym(0xff13),
	"scroll_lock": platform.KeySym(0xff14),
	"caps_lock":   platform.KeySym(0xffe5),
	"num_lock":    platform.KeySym(0xff7f),
	"print":       platform.KeySym(0xff61),

	// Punctuation / symbols
	"minus":        platform.KeySym(0x002d),
	"plus":         platform.KeySym(0x002b),
	"equal":        platform.KeySym(0x003d),
	"bracketleft":  platform.KeySym(0x005b),
	"bracketright": platform.KeySym(0x005d),
	"backslash":    platform.KeySym(0x005c),
	"semicolon":    platform.KeySym(0x003b),
	"apostrophe":   platform.KeySym(0x0027),
	"comma":        platform.KeySym(0x002c),
	"period":       platform.KeySym(0x002e),
	"slash":        platform.KeySym(0x002f),
	"grave":        platform.KeySym(0x0060),
	"asciitilde":   platform.KeySym(0x007e),
	"exclam":       platform.KeySym(0x0021),
	"at":           platform.KeySym(0x0040),
	"numbersign":   platform.KeySym(0x0023),
	"dollar":       platform.KeySym(0x0024),
	"percent":      platform.KeySym(0x0025),
	"asciicircum":  platform.KeySym(0x005e),
	"ampersand":    platform.KeySym(0x0026),
	"asterisk":     platform.KeySym(0x002a),
	"parenleft":    platform.KeySym(0x0028),
	"parenright":   platform.KeySym(0x0029),
	"underscore":   platform.KeySym(0x005f),
	"braceleft":    platform.KeySym(0x007b),
	"braceright":   platform.KeySym(0x007d),
	"bar":          platform.KeySym(0x007c),
	"colon":        platform.KeySym(0x003a),
	"quotedbl":     platform.KeySym(0x0022),
	"less":         platform.KeySym(0x003c),
	"greater":      platform.KeySym(0x003e),
	"question":     platform.KeySym(0x003f),
}

// lookupKeySym resolves a key name (case-insensitive) to a KeySym.
// Returns 0 if not found.
func lookupKeySym(name string) platform.KeySym {
	// Try exact lowercase match.
	lower := toLower(name)
	if ks, ok := keysymNames[lower]; ok {
		return ks
	}
	// Single character: use its ASCII value as keysym.
	if len(name) == 1 {
		ch := name[0]
		if ch >= 0x20 && ch <= 0x7e {
			return platform.KeySym(ch)
		}
	}
	return 0
}

// keysymNamesByValue is the inverse of keysymNames; where several names share
// a keysym (prior/pageup) the shortest, then alphabetically first, is kept.
var keysymNamesByValue = func() map[platform.KeySym]string {
	m := make(map[platform.KeySym]string, len(keysymNames))
	for name, ks := range keysymNames {
		if old, ok := m[ks]; !ok || len(name) < len(old) || len(name) == len(old) && name < old {
			m[ks] = name
		}
	}
	return m
}()

// keySymName returns a name for ks that lookupKeySym resolves back to ks.
func keySymName(ks platform.KeySym) string {
	if name, ok := keysymNamesByValue[ks]; ok {
		return name
	}
	if ks >= 0x20 && ks <= 0x7e {
		return string(rune(ks))
	}
	return "0x" + strconv.FormatUint(uint64(ks), 16)
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
