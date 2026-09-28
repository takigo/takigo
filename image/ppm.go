package image

import (
	"bufio"
	"bytes"
	"fmt"
	goimage "image"
	"image/color"
	"io"
	"os"
	"strconv"
	"strings"
)

// NewPhotoFromPPMFile loads a PPM/PGM/PBM (Netpbm) file and creates a Photo.
// Supports P1-P6 (ASCII and binary bitmap, greymap, pixmap).
func NewPhotoFromPPMFile(name, path string) (*Photo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("image: open ppm %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()
	return NewPhotoFromPPMReader(name, f)
}

// NewPhotoFromPPMReader decodes a PPM/PGM/PBM image from a reader.
func NewPhotoFromPPMReader(name string, r io.Reader) (*Photo, error) {
	img, err := decodePPM(r)
	if err != nil {
		return nil, fmt.Errorf("image: decode ppm: %w", err)
	}
	return NewPhoto(name, img), nil
}

// decodePPM parses the Netpbm format (P1..P6) and returns an *image.RGBA.
// Reference: http://netpbm.sourceforge.net/doc/ppm.html
func decodePPM(r io.Reader) (*goimage.RGBA, error) {
	br := bufio.NewReader(r)

	magic, err := readPPMToken(br)
	if err != nil {
		return nil, err
	}
	if len(magic) != 2 || magic[0] != 'P' || magic[1] < '1' || magic[1] > '6' {
		return nil, fmt.Errorf("not a Netpbm file (magic=%q)", magic)
	}
	format := magic[1] - '0'

	w, err := readPPMInt(br)
	if err != nil {
		return nil, fmt.Errorf("missing width: %w", err)
	}
	h, err := readPPMInt(br)
	if err != nil {
		return nil, fmt.Errorf("missing height: %w", err)
	}
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("invalid dimensions %dx%d", w, h)
	}

	maxVal := 1
	if format != 1 && format != 4 {
		maxVal, err = readPPMInt(br)
		if err != nil {
			return nil, fmt.Errorf("missing maxval: %w", err)
		}
		if maxVal < 1 || maxVal > 65535 {
			return nil, fmt.Errorf("invalid maxval %d", maxVal)
		}
	}

	// readPPMToken already consumes the single whitespace byte that
	// terminates the header (per the Netpbm spec) for ASCII formats; the
	// call to readPPMInt for maxVal also consumes its trailing whitespace.
	// For binary formats the same applies: the newline after the maxval
	// has been consumed and the data stream follows immediately.

	rgba := goimage.NewRGBA(goimage.Rect(0, 0, w, h))

	// P1 / P4: 1-bit bitmap (PBM).
	if format == 1 || format == 4 {
		return decodePBM(rgba, br, format == 4)
	}

	// P2 / P5: 8/16-bit greyscale (PGM).
	if format == 2 || format == 5 {
		return decodePGM(rgba, br, format == 5, maxVal)
	}

	// P3 / P6: 8/16-bit RGB (PPM).
	return decodePPMColor(rgba, br, format == 6, maxVal)
}

func decodePBM(rgba *goimage.RGBA, br *bufio.Reader, binary bool) (*goimage.RGBA, error) {
	w := rgba.Rect.Dx()
	h := rgba.Rect.Dy()
	black := color.RGBA{0, 0, 0, 255}
	white := color.RGBA{255, 255, 255, 255}

	setPx := func(x, y int, on bool) {
		if on {
			rgba.SetRGBA(x, y, black)
		} else {
			rgba.SetRGBA(x, y, white)
		}
	}

	if binary {
		rowBytes := (w + 7) / 8
		buf := make([]byte, rowBytes)
		for y := 0; y < h; y++ {
			if _, err := io.ReadFull(br, buf); err != nil {
				return nil, fmt.Errorf("pbm: read row %d: %w", y, err)
			}
			for x := 0; x < w; x++ {
				bit := (buf[x/8] >> uint(7-x%8)) & 1
				setPx(x, y, bit == 1)
			}
		}
		return rgba, nil
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			v, err := readPPMInt(br)
			if err != nil {
				return nil, fmt.Errorf("pbm: read pixel (%d,%d): %w", x, y, err)
			}
			setPx(x, y, v != 0)
		}
	}
	return rgba, nil
}

func decodePGM(rgba *goimage.RGBA, br *bufio.Reader, binary bool, maxVal int) (*goimage.RGBA, error) {
	w := rgba.Rect.Dx()
	h := rgba.Rect.Dy()
	n := w * h

	if binary {
		data, err := readGrayBinary(br, n, maxVal)
		if err != nil {
			return nil, err
		}
		writeGray(rgba, data, maxVal)
		return rgba, nil
	}

	data := make([]uint32, n)
	for i := 0; i < n; i++ {
		v, err := readPPMInt(br)
		if err != nil {
			return nil, fmt.Errorf("pgm: read pixel %d: %w", i, err)
		}
		if v < 0 {
			v = 0
		} else if v > maxVal {
			v = maxVal
		}
		data[i] = uint32(v)
	}
	writeGray(rgba, data, maxVal)
	return rgba, nil
}

func decodePPMColor(rgba *goimage.RGBA, br *bufio.Reader, binary bool, maxVal int) (*goimage.RGBA, error) {
	w := rgba.Rect.Dx()
	h := rgba.Rect.Dy()
	n := w * h

	if binary {
		data, err := readRGBBinary(br, n, maxVal)
		if err != nil {
			return nil, err
		}
		writeRGB(rgba, data, maxVal)
		return rgba, nil
	}

	data := make([]uint32, n*3)
	for i := 0; i < n*3; i++ {
		v, err := readPPMInt(br)
		if err != nil {
			return nil, fmt.Errorf("ppm: read sample %d: %w", i, err)
		}
		if v < 0 {
			v = 0
		} else if v > maxVal {
			v = maxVal
		}
		data[i] = uint32(v)
	}
	writeRGB(rgba, data, maxVal)
	return rgba, nil
}

func readGrayBinary(br *bufio.Reader, n, maxVal int) ([]uint32, error) {
	out := make([]uint32, n)
	if maxVal < 256 {
		buf := make([]byte, n)
		if _, err := io.ReadFull(br, buf); err != nil {
			return nil, fmt.Errorf("pgm: read data: %w", err)
		}
		for i, b := range buf {
			out[i] = uint32(b)
		}
		return out, nil
	}
	for i := 0; i < n; i++ {
		v, err := readUint16BE(br)
		if err != nil {
			return nil, fmt.Errorf("pgm: read sample %d: %w", i, err)
		}
		out[i] = uint32(v)
	}
	return out, nil
}

func readRGBBinary(br *bufio.Reader, n, maxVal int) ([]uint32, error) {
	out := make([]uint32, n*3)
	if maxVal < 256 {
		buf := make([]byte, n*3)
		if _, err := io.ReadFull(br, buf); err != nil {
			return nil, fmt.Errorf("ppm: read data: %w", err)
		}
		for i, b := range buf {
			out[i] = uint32(b)
		}
		return out, nil
	}
	for i := 0; i < n*3; i++ {
		v, err := readUint16BE(br)
		if err != nil {
			return nil, fmt.Errorf("ppm: read sample %d: %w", i, err)
		}
		out[i] = uint32(v)
	}
	return out, nil
}

func writeGray(rgba *goimage.RGBA, data []uint32, maxVal int) {
	w := rgba.Rect.Dx()
	h := rgba.Rect.Dy()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			v := byte((uint64(data[y*w+x]) * 255) / uint64(maxVal))
			rgba.SetRGBA(x, y, color.RGBA{v, v, v, 255})
		}
	}
}

func writeRGB(rgba *goimage.RGBA, data []uint32, maxVal int) {
	w := rgba.Rect.Dx()
	h := rgba.Rect.Dy()
	mv := uint64(maxVal)
	for y := 0; y < h; y++ {
		row := y * w
		for x := 0; x < w; x++ {
			i := (row + x) * 3
			r := byte((uint64(data[i]) * 255) / mv)
			g := byte((uint64(data[i+1]) * 255) / mv)
			b := byte((uint64(data[i+2]) * 255) / mv)
			rgba.SetRGBA(x, y, color.RGBA{r, g, b, 255})
		}
	}
}

func readUint16BE(br *bufio.Reader) (uint16, error) {
	var buf [2]byte
	if _, err := io.ReadFull(br, buf[:]); err != nil {
		return 0, err
	}
	return uint16(buf[0])<<8 | uint16(buf[1]), nil
}

// readPPMToken reads the next whitespace-separated token from a PPM stream,
// skipping comments (# to end of line) and whitespace.
func readPPMToken(br *bufio.Reader) (string, error) {
	var b bytes.Buffer
	for {
		c, err := br.ReadByte()
		if err != nil {
			if b.Len() == 0 {
				return "", err
			}
			return b.String(), io.EOF
		}
		if c == '#' {
			// skip to end of line
			for {
				cc, err := br.ReadByte()
				if err != nil {
					break
				}
				if cc == '\n' {
					break
				}
			}
			continue
		}
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\v' || c == '\f' {
			if b.Len() > 0 {
				return b.String(), nil
			}
			continue
		}
		b.WriteByte(c)
	}
}

// readPPMInt reads the next integer from the PPM stream.
func readPPMInt(br *bufio.Reader) (int, error) {
	tok, err := readPPMToken(br)
	if err != nil && tok == "" {
		return 0, err
	}
	v, perr := strconv.Atoi(strings.TrimSpace(tok))
	if perr != nil {
		return 0, fmt.Errorf("invalid integer %q: %w", tok, perr)
	}
	return v, nil
}
