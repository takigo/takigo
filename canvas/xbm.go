package canvas

import (
	"fmt"
	"image/color"
	"os"
	"strconv"
	"strings"
)

// XBMData holds parsed XBM (X BitMap) data.
type XBMData struct {
	Width, Height int
	Bits          []byte // row-major, LSB-first, packed 8 pixels per byte
}

// ParseXBM parses an XBM file from its text content.
// XBM is an ASCII C header that defines a 1-bit bitmap.
func ParseXBM(src string) (*XBMData, error) {
	xbm := &XBMData{}

	for _, line := range strings.Split(src, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#define") {
			fields := strings.Fields(line)
			if len(fields) < 3 {
				continue
			}
			name := fields[1]
			val, err := strconv.Atoi(fields[2])
			if err != nil {
				continue
			}
			if strings.HasSuffix(name, "_width") {
				xbm.Width = val
			} else if strings.HasSuffix(name, "_height") {
				xbm.Height = val
			}
		}
	}

	if xbm.Width == 0 || xbm.Height == 0 {
		return nil, fmt.Errorf("xbm: missing width or height")
	}

	// Extract hex bytes from the static array.
	start := strings.Index(src, "{")
	end := strings.LastIndex(src, "}")
	if start < 0 || end < 0 || end <= start {
		return nil, fmt.Errorf("xbm: could not find data array")
	}
	body := src[start+1 : end]

	xbm.Bits = nil
	for _, tok := range strings.FieldsFunc(body, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n' || r == '\r'
	}) {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		tok = strings.TrimPrefix(tok, "0x")
		tok = strings.TrimPrefix(tok, "0X")
		b, err := strconv.ParseUint(tok, 16, 8)
		if err != nil {
			continue
		}
		xbm.Bits = append(xbm.Bits, byte(b))
	}

	rowBytes := (xbm.Width + 7) / 8
	expected := rowBytes * xbm.Height
	if len(xbm.Bits) < expected {
		return nil, fmt.Errorf("xbm: expected %d bytes, got %d", expected, len(xbm.Bits))
	}

	return xbm, nil
}

// ParseXBMFile parses an XBM file from the filesystem.
func ParseXBMFile(path string) (*XBMData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseXBM(string(data))
}

// ToRGBA converts XBM data to an RGBA image using the given foreground and
// background colors. Pixels with bit=1 get fg; pixels with bit=0 get bg.
// bg.A=0 makes the background transparent.
func (x *XBMData) ToRGBA(fg, bg color.RGBA) []byte {
	w, h := x.Width, x.Height
	rowBytes := (w + 7) / 8
	rgba := make([]byte, w*h*4)
	for row := range h {
		for col := range w {
			byteIdx := row*rowBytes + col/8
			bitIdx := uint(col % 8)
			bit := (x.Bits[byteIdx] >> bitIdx) & 1
			off := (row*w + col) * 4
			if bit == 1 {
				rgba[off+0] = fg.R
				rgba[off+1] = fg.G
				rgba[off+2] = fg.B
				rgba[off+3] = fg.A
			} else {
				rgba[off+0] = bg.R
				rgba[off+1] = bg.G
				rgba[off+2] = bg.B
				rgba[off+3] = bg.A
			}
		}
	}
	return rgba
}
