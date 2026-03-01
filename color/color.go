// Package color provides color parsing, X11 allocation, and caching.
// It ports tk/generic/tkColor.c and tk/unix/tkUnixColor.c.
package color

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/msorc/takigo/internal/xlib"
)

// Color represents an allocated color with its X11 pixel value and RGB components.
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

// Cache manages color allocations per display, caching by name and by value.
type Cache struct {
	mu      sync.RWMutex
	display *xlib.Display
	screen  int
	cmap    xlib.Colormap
	byName  map[string]*Color
	byValue map[colorKey]*Color
}

type colorKey struct {
	r, g, b uint16
}

// NewCache creates a new color cache for the given display.
func NewCache(display *xlib.Display, screen int, cmap xlib.Colormap) *Cache {
	return &Cache{
		display: display,
		screen:  screen,
		cmap:    cmap,
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

// parseHex parses hex color formats: RGB, RRGGBB, RRRRGGGGBBBB.
func parseHex(hex string) (r, g, b uint16, err error) {
	switch len(hex) {
	case 3: // #RGB
		rv, _ := strconv.ParseUint(string(hex[0])+string(hex[0]), 16, 16)
		gv, _ := strconv.ParseUint(string(hex[1])+string(hex[1]), 16, 16)
		bv, _ := strconv.ParseUint(string(hex[2])+string(hex[2]), 16, 16)
		return uint16(rv << 8), uint16(gv << 8), uint16(bv << 8), nil

	case 6: // #RRGGBB
		rv, _ := strconv.ParseUint(hex[0:2], 16, 16)
		gv, _ := strconv.ParseUint(hex[2:4], 16, 16)
		bv, _ := strconv.ParseUint(hex[4:6], 16, 16)
		return uint16(rv << 8), uint16(gv << 8), uint16(bv << 8), nil

	case 12: // #RRRRGGGGBBBB
		rv, _ := strconv.ParseUint(hex[0:4], 16, 16)
		gv, _ := strconv.ParseUint(hex[4:8], 16, 16)
		bv, _ := strconv.ParseUint(hex[8:12], 16, 16)
		return uint16(rv), uint16(gv), uint16(bv), nil

	default:
		return 0, 0, 0, fmt.Errorf("invalid hex color #%s", hex)
	}
}

// trueColorPixel constructs a 24-bit TrueColor pixel from 16-bit RGB.
func trueColorPixel(r, g, b uint16) uint64 {
	return uint64(r>>8)<<16 | uint64(g>>8)<<8 | uint64(b>>8)
}
