package canvas

import (
	"strings"
	"unicode"

	"github.com/takigo/takigo/font"
)

// psFontName converts a takigo font (family + weight + slant + size) to the
// PostScript font name and points size that the printer should select.
//
// Ports tk/generic/tkFont.c:1703-1843 (Tk_PostscriptFontName) so the canvas
// PostScript output references fonts the PostScript interpreter can resolve.
func psFontName(f font.Font) (psName string, points int) {
	a := f.Attrs()
	family := familyAlias(a.Family)
	family = titleCase(family)

	if family == "NewCenturySchoolbook" {
		family = "NewCenturySchlbk"
	}

	weightSuffix, slantSuffix := fontSuffixes(family, a.Weight, a.Slant)

	switch {
	case slantSuffix == "" && weightSuffix == "":
		if family == "Times" || family == "NewCenturySchlbk" || family == "Palatino" {
			psName = family + "-Roman"
		} else {
			psName = family
		}
	default:
		psName = family + "-" + weightSuffix + slantSuffix
	}

	points = max(int(a.Size+0.5), 1)
	return psName, points
}

// familyAlias maps Tk's well-known family aliases to canonical PostScript
// names. Mirrors tk/generic/tkFont.c:1726-1740.
func familyAlias(family string) string {
	if strings.HasPrefix(strings.ToLower(family), "itc ") {
		family = family[4:]
	}
	switch strings.ToLower(family) {
	case "arial", "geneva":
		return "Helvetica"
	case "times new roman", "new york":
		return "Times"
	case "courier new", "monaco":
		return "Courier"
	}
	return family
}

// titleCase applies Tk's "first letter of each word uppercase, rest lower,
// drop spaces" rule. Mirrors tk/generic/tkFont.c:1754-1771.
func titleCase(family string) string {
	var b strings.Builder
	upper := true
	for _, r := range family {
		if unicode.IsSpace(r) {
			upper = true
			continue
		}
		if upper {
			b.WriteRune(unicode.ToUpper(r))
			upper = false
		} else {
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}

// fontSuffixes returns the weight/slant suffix strings to append after the family
// name. tk'sFont.c:1788-1819.
func fontSuffixes(family string, w font.Weight, s font.Slant) (weight, slant string) {
	if w == font.WeightNormal {
		switch family {
		case "Bookman":
			weight = "Light"
		case "Avantgarde":
			weight = "Book"
		case "Zapfchancery":
			weight = "Medium"
		}
	} else {
		switch family {
		case "Bookman", "Avantgarde":
			weight = "Demi"
		default:
			weight = "Bold"
		}
	}

	switch s {
	case font.SlantRoman:
		// no slant suffix
	case font.SlantItalic, font.SlantOblique:
		switch family {
		case "Helvetica", "Courier", "Avantgarde":
			slant = "Oblique"
		default:
			slant = "Italic"
		}
	}
	return weight, slant
}

// psFontEmit produces the "/<name> findfont N scalefont ISOEncode setfont\n"
// fragment matching tk/generic/tkCanvPs.c:802-805.
func psFontEmit(f font.Font) string {
	name, points := psFontName(f)
	encode := " ISOEncode"
	if strings.HasPrefix(strings.ToLower(name), "symbol") {
		encode = ""
	}
	return "/" + name + " findfont " + psItoa(points) + " scalefont" + encode + " setfont\n"
}

func psItoa(n int) string {
	// small helper so we don't need strconv in the hot path
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
