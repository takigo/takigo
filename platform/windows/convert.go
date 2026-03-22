//go:build windows

// Package windows provides the Windows/Win32 backend for the platform abstraction layer.
package windows

import (
	w32 "github.com/msorc/takigo/internal/win32"
	"github.com/msorc/takigo/platform"
)

// Type conversion helpers — platform types are uintptr, win32 types are uintptr-based.
func toHWND(w platform.WindowID) w32.HWND       { return w32.HWND(uintptr(w)) }
func toDrawableHWND(d platform.DrawableID) w32.HWND { return w32.HWND(uintptr(d)) }
func fromHWND(h w32.HWND) platform.WindowID     { return platform.WindowID(uintptr(h)) }
func fromPixmap(p platform.PixmapID) uintptr     { return uintptr(p) }

// pixelToCOLORREF converts a platform pixel value (0x00RRGGBB) to Win32 COLORREF (0x00BBGGRR).
func pixelToCOLORREF(pixel uint64) w32.COLORREF {
	r := byte(pixel >> 16)
	g := byte(pixel >> 8)
	b := byte(pixel)
	return w32.RGB(r, g, b)
}

// colorrefToPixel converts a Win32 COLORREF to platform pixel value.
func colorrefToPixel(c w32.COLORREF) uint64 {
	r := w32.GetRValue(c)
	g := w32.GetGValue(c)
	b := w32.GetBValue(c)
	return uint64(r)<<16 | uint64(g)<<8 | uint64(b)
}
