// Package color provides color parsing, allocation, and caching.
// It ports tk/generic/tkColor.c and tk/unix/tkUnixColor.c.
package color

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// Color represents an allocated color with its pixel value and RGB components.
type Color struct {
	Pixel uint64
	Red   uint16
	Green uint16
	Blue  uint16
	Name  string // original name used to allocate
}

// RGBA returns the color as 8-bit RGBA values.
func (c *Color) RGBA() (r, g, b, a uint8) {
	return uint8(c.Red >> 8), uint8(c.Green >> 8), uint8(c.Blue >> 8), 255
}

// Ref returns a lightweight copy of the color's pixel and RGB data.
func (c *Color) Ref() *ColorRef {
	return &ColorRef{Pixel: c.Pixel, Red: c.Red, Green: c.Green, Blue: c.Blue}
}

// ColorRef is a lightweight color reference storing pixel value and RGB components.
// Used by widgets to store resolved color data without referencing the cache.
type ColorRef struct {
	Pixel uint64
	Red   uint16
	Green uint16
	Blue  uint16
}

// Cache manages color allocations per display, caching by name and by value.
type Cache struct {
	mu      sync.RWMutex
	screen  int
	byName  map[string]*Color
	byValue map[colorKey]*Color
}

type colorKey struct {
	r, g, b uint16
}

// NewCache creates a new color cache.
func NewCache(screen int) *Cache {
	return &Cache{
		screen:  screen,
		byName:  make(map[string]*Color),
		byValue: make(map[colorKey]*Color),
	}
}

// Get allocates or retrieves a cached color by name.
// Supports: "#RGB", "#RRGGBB", "#RRRRGGGGBBBB", named X11 colors.
func (c *Cache) Get(name string) (*Color, error) {
	c.mu.RLock()
	if col, ok := c.byName[name]; ok {
		c.mu.RUnlock()
		return col, nil
	}
	c.mu.RUnlock()

	// Parse color.
	r, g, b, err := Parse(name)
	if err != nil {
		return nil, err
	}

	// Construct pixel value for TrueColor displays.
	pixel := trueColorPixel(r, g, b)

	col := &Color{
		Pixel: pixel,
		Red:   r,
		Green: g,
		Blue:  b,
		Name:  name,
	}

	c.mu.Lock()
	c.byName[name] = col
	c.byValue[colorKey{r, g, b}] = col
	c.mu.Unlock()

	return col, nil
}

// GetByValue allocates or retrieves a cached color by RGB values (16-bit).
func (c *Cache) GetByValue(r, g, b uint16) (*Color, error) {
	key := colorKey{r, g, b}

	c.mu.RLock()
	if col, ok := c.byValue[key]; ok {
		c.mu.RUnlock()
		return col, nil
	}
	c.mu.RUnlock()

	pixel := trueColorPixel(r, g, b)
	col := &Color{
		Pixel: pixel,
		Red:   r,
		Green: g,
		Blue:  b,
		Name:  fmt.Sprintf("#%04x%04x%04x", r, g, b),
	}

	c.mu.Lock()
	c.byValue[key] = col
	c.mu.Unlock()

	return col, nil
}

// Parse parses a color string and returns 16-bit RGB components.
func Parse(name string) (r, g, b uint16, err error) {
	if len(name) > 0 && name[0] == '#' {
		return parseHex(name[1:])
	}

	// Try named color.
	if col, ok := NamedColors[strings.ToLower(name)]; ok {
		return col.R, col.G, col.B, nil
	}

	return 0, 0, 0, fmt.Errorf("unknown color %q", name)
}

// parseHex parses #RGB, #RRGGBB, #RRRGGGBBB and #RRRRGGGGBBBB, widening
// each component to 16 bits by repeating its digits as XParseColor
// (tk/xlib/xcolors.c) does, and rejects any non-hex digit.
func parseHex(hex string) (r, g, b uint16, err error) {
	n := len(hex) / 3
	if len(hex)%3 != 0 || n < 1 || n > 4 {
		return 0, 0, 0, fmt.Errorf("invalid hex color #%s", hex)
	}
	var c [3]uint16
	for i := range c {
		v, perr := strconv.ParseUint(hex[i*n:(i+1)*n], 16, 16)
		if perr != nil {
			return 0, 0, 0, fmt.Errorf("invalid hex color #%s", hex)
		}
		switch n {
		case 1:
			v *= 0x1111
		case 2:
			v *= 0x101
		case 3:
			v = v<<4 | v>>8
		}
		c[i] = uint16(v)
	}
	return c[0], c[1], c[2], nil
}

// trueColorPixel constructs a 24-bit TrueColor pixel from 16-bit RGB.
func trueColorPixel(r, g, b uint16) uint64 {
	return uint64(r>>8)<<16 | uint64(g>>8)<<8 | uint64(b>>8)
}
