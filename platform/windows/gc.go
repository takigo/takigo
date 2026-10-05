//go:build windows

package windows

import (
	w32 "github.com/takigo/takigo/internal/win32"
	"github.com/takigo/takigo/platform"
)

// gcState emulates an X11 graphics context as a Go struct.
// When drawing, GDI objects are created from this state temporarily.
type gcState struct {
	foreground uint64
	background uint64
	lineWidth  int
	lineStyle  int
	capStyle   int
	joinStyle  int
	fillStyle  int
	function   int // GX raster op
	stipple    platform.PixmapID
	dashOffset int
	dashList   []byte

	// The GDI pen and brush last made for this GC, reused while the
	// state they depend on is unchanged (X keeps them in the GC too),
	// instead of a CreatePen/DeleteObject pair per primitive.
	pen      w32.HPEN
	penKey   penKey
	brush    w32.HBRUSH
	brushKey w32.COLORREF
}

// penKey is the GC state a pen is made from.
type penKey struct {
	color        w32.COLORREF
	width, style int32
}

// --- GCManager implementation ---

func (d *WindowsDisplay) CreateGC(drawable platform.DrawableID, valueMask uint64, values *platform.GCValues) platform.GCID {
	gc := &gcState{
		foreground: 0,          // black
		background: 0x00FFFFFF, // white
		lineWidth:  0,
		lineStyle:  platform.LineSolid,
		capStyle:   platform.CapButt,
		joinStyle:  platform.JoinMiter,
		fillStyle:  platform.FillSolid,
		function:   platform.GXcopy,
	}

	if values != nil {
		if valueMask&platform.GCForeground != 0 {
			gc.foreground = values.Foreground
		}
		if valueMask&platform.GCBackground != 0 {
			gc.background = values.Background
		}
		if valueMask&platform.GCLineWidth != 0 {
			gc.lineWidth = values.LineWidth
		}
		if valueMask&platform.GCFunction != 0 {
			gc.function = values.Function
		}
	}

	d.gcMu.Lock()
	id := platform.GCID(d.gcNext)
	d.gcNext++
	d.gcs[id] = gc
	d.gcMu.Unlock()

	return id
}

func (d *WindowsDisplay) FreeGC(gc platform.GCID) {
	d.gcMu.Lock()
	if g, ok := d.gcs[gc]; ok {
		g.release()
	}
	delete(d.gcs, gc)
	d.gcMu.Unlock()
}

func (d *WindowsDisplay) SetForeground(gc platform.GCID, pixel uint64) {
	d.gcMu.Lock()
	if g, ok := d.gcs[gc]; ok {
		g.foreground = pixel
	}
	d.gcMu.Unlock()
}

func (d *WindowsDisplay) SetBackground(gc platform.GCID, pixel uint64) {
	d.gcMu.Lock()
	if g, ok := d.gcs[gc]; ok {
		g.background = pixel
	}
	d.gcMu.Unlock()
}

func (d *WindowsDisplay) SetLineAttributes(gc platform.GCID, lineWidth uint, lineStyle, capStyle, joinStyle int) {
	d.gcMu.Lock()
	if g, ok := d.gcs[gc]; ok {
		g.lineWidth = int(lineWidth)
		g.lineStyle = lineStyle
		g.capStyle = capStyle
		g.joinStyle = joinStyle
	}
	d.gcMu.Unlock()
}

func (d *WindowsDisplay) SetArcMode(platform.GCID, int) {}

func (d *WindowsDisplay) SetFillStyle(gc platform.GCID, fillStyle int) {
	d.gcMu.Lock()
	if g, ok := d.gcs[gc]; ok {
		g.fillStyle = fillStyle
	}
	d.gcMu.Unlock()
}

func (d *WindowsDisplay) SetTSOrigin(gc platform.GCID, x, y int) {}

func (d *WindowsDisplay) SetStipple(gc platform.GCID, stipple platform.PixmapID) {
	d.gcMu.Lock()
	if g, ok := d.gcs[gc]; ok {
		g.stipple = stipple
	}
	d.gcMu.Unlock()
}

func (d *WindowsDisplay) SetDashes(gc platform.GCID, dashOffset int, dashList []byte) {
	d.gcMu.Lock()
	if g, ok := d.gcs[gc]; ok {
		g.dashOffset = dashOffset
		g.dashList = make([]byte, len(dashList))
		copy(g.dashList, dashList)
	}
	d.gcMu.Unlock()
}

// getGC returns a snapshot of the GC state.
func (d *WindowsDisplay) getGC(gc platform.GCID) *gcState {
	d.gcMu.Lock()
	defer d.gcMu.Unlock()
	return d.gcs[gc]
}

// gcROP2 maps GX function constants to Win32 R2_* modes.
func gcROP2(fn int) int32 {
	switch fn {
	case 0: // GXclear
		return w32.R2_BLACK
	case 1: // GXand
		return w32.R2_MASKPEN
	case 2: // GXandReverse
		return w32.R2_MASKPENNOT
	case 3: // GXcopy (default)
		return w32.R2_COPYPEN
	case 4: // GXandInverted
		return w32.R2_MASKNOTPEN
	case 5: // GXnoop
		return w32.R2_NOP
	case 6: // GXxor
		return w32.R2_XORPEN
	case 7: // GXor
		return w32.R2_MERGEPEN
	case 8: // GXnor
		return w32.R2_NOTMERGEPEN
	case 9: // GXequiv
		return w32.R2_NOTXORPEN
	case 10: // GXinvert
		return w32.R2_NOT
	case 11: // GXorReverse
		return w32.R2_MERGEPENNOT
	case 12: // GXcopyInverted
		return w32.R2_NOTCOPYPEN
	case 13: // GXorInverted
		return w32.R2_MERGENOTPEN
	case 14: // GXnand
		return w32.R2_NOTMASKPEN
	case 15: // GXset
		return w32.R2_WHITE
	default:
		return w32.R2_COPYPEN
	}
}

// gcBitBltROP maps GX function constants to BitBlt raster ops.
func gcBitBltROP(fn int) uint32 {
	switch fn {
	case 0:
		return w32.BLACKNESS
	case 3:
		return w32.SRCCOPY
	case 6:
		return w32.SRCINVERT
	case 7:
		return w32.SRCPAINT
	case 8:
		return w32.NOTSRCERASE
	case 10:
		return w32.DSTINVERT
	case 12:
		return w32.NOTSRCCOPY
	case 15:
		return w32.WHITENESS
	default:
		return w32.SRCCOPY
	}
}

// getPen returns the GC's pen for its current state, creating it (and
// deleting the previous one) when the colour, width or style changed.
// The pen belongs to the GC; callers must not delete it.
func (g *gcState) getPen() w32.HPEN {
	width := max(int32(g.lineWidth), 1)
	style := int32(w32.PS_SOLID)
	switch g.lineStyle {
	case platform.LineOnOffDash, platform.LineDoubleDash:
		style = w32.PS_DASH
	}
	k := penKey{pixelToCOLORREF(g.foreground), width, style}
	if g.pen == 0 || g.penKey != k {
		if g.pen != 0 {
			w32.DeleteObject(w32.HGDIOBJ(g.pen))
		}
		g.pen, g.penKey = w32.CreatePen(style, width, k.color), k
	}
	return g.pen
}

// getBrush returns the GC's solid brush in its foreground colour; like
// getPen it is cached and owned by the GC.
func (g *gcState) getBrush() w32.HBRUSH {
	c := pixelToCOLORREF(g.foreground)
	if g.brush == 0 || g.brushKey != c {
		if g.brush != 0 {
			w32.DeleteObject(w32.HGDIOBJ(g.brush))
		}
		g.brush, g.brushKey = w32.CreateSolidBrush(c), c
	}
	return g.brush
}

// release deletes the GC's cached GDI objects.
func (g *gcState) release() {
	if g.pen != 0 {
		w32.DeleteObject(w32.HGDIOBJ(g.pen))
		g.pen = 0
	}
	if g.brush != 0 {
		w32.DeleteObject(w32.HGDIOBJ(g.brush))
		g.brush = 0
	}
}
