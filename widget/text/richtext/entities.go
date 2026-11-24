package richtext

import (
	"strconv"
	"strings"
)

// entities are the named character references in common use.
var entities = map[string]string{
	"amp": "&", "lt": "<", "gt": ">", "quot": "\"", "apos": "'", "nbsp": "\u00a0",
	"copy": "©", "reg": "®", "trade": "™", "mdash": "—", "ndash": "–", "hellip": "…",
	"laquo": "«", "raquo": "»", "lsquo": "‘", "rsquo": "’", "ldquo": "“", "rdquo": "”",
	"sbquo": "‚", "bdquo": "„", "bull": "•", "middot": "·", "times": "×", "divide": "÷",
	"deg": "°", "plusmn": "±", "para": "¶", "sect": "§", "euro": "€", "pound": "£",
	"yen": "¥", "cent": "¢", "larr": "←", "rarr": "→", "uarr": "↑", "darr": "↓",
	"harr": "↔", "lArr": "⇐", "rArr": "⇒", "hearts": "♥", "check": "✓", "minus": "−",
	"le": "≤", "ge": "≥", "ne": "≠", "asymp": "≈", "infin": "∞", "frac12": "½",
	"frac14": "¼", "frac34": "¾", "sup2": "²", "sup3": "³", "micro": "µ", "shy": "\u00ad",
	"ensp": "\u2002", "emsp": "\u2003", "thinsp": "\u2009", "zwj": "\u200d", "zwnj": "\u200c",
	"iexcl": "¡", "iquest": "¿", "auml": "ä", "ouml": "ö", "uuml": "ü", "Auml": "Ä",
	"Ouml": "Ö", "Uuml": "Ü", "szlig": "ß", "eacute": "é", "egrave": "è", "aacute": "á",
	"agrave": "à", "ccedil": "ç", "ntilde": "ñ", "oslash": "ø", "aring": "å", "alpha": "α",
	"beta": "β", "gamma": "γ", "delta": "δ", "pi": "π", "sigma": "σ", "omega": "ω",
	"lambda": "λ", "mu": "μ", "Delta": "Δ", "Omega": "Ω", "Sigma": "Σ",
}

// decodeEntities replaces character references; an unknown one is kept.
func decodeEntities(s string) string {
	if !strings.Contains(s, "&") {
		return s
	}
	var b strings.Builder
	for {
		amp := strings.IndexByte(s, '&')
		if amp < 0 {
			b.WriteString(s)
			return b.String()
		}
		b.WriteString(s[:amp])
		s = s[amp:]
		semi := strings.IndexByte(s, ';')
		if semi < 2 || semi > 32 {
			b.WriteByte('&')
			s = s[1:]
			continue
		}
		ref := s[1:semi]
		if r, ok := numericRef(ref); ok {
			b.WriteString(r)
		} else if r, ok := entities[ref]; ok {
			b.WriteString(r)
		} else {
			b.WriteString(s[:semi+1])
		}
		s = s[semi+1:]
	}
}

func numericRef(ref string) (string, bool) {
	num, ok := strings.CutPrefix(ref, "#")
	if !ok {
		return "", false
	}
	base := 10
	if h, ok := strings.CutPrefix(strings.ToLower(num), "x"); ok {
		num, base = h, 16
	}
	n, err := strconv.ParseUint(num, base, 32)
	if err != nil || n == 0 || n > 0x10ffff || n >= 0xd800 && n <= 0xdfff {
		return "\ufffd", err == nil
	}
	return string(rune(n)), true
}
