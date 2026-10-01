//go:build windows

package windows

import (
	"sync"

	w32 "github.com/msorc/takigo/internal/win32"
	"github.com/msorc/takigo/platform"
)

const (
	iconSmall = 0 // ICON_SMALL: title bar
	iconBig   = 1 // ICON_BIG: Alt-Tab and the taskbar
)

// windowIcons holds the icons created for each window, so they can be
// destroyed when replaced.
var (
	windowIconsMu sync.Mutex
	windowIcons   = map[w32.HWND][2]w32.HICON{}
)

// SetWindowIcons sets the window's small and big icons (WM_SETICON) from
// the sizes closest to 16 and 32 pixels, as tkWinWm.c's WmIconphotoCmd
// does with its title-bar and Alt-Tab icons.
func (d *WindowsDisplay) SetWindowIcons(w platform.WindowID, icons []platform.IconImage) {
	hwnd := toHWND(w)
	var made [2]w32.HICON
	if small := closestIcon(icons, 16); small != nil {
		made[iconSmall] = createIcon(small)
	}
	if big := closestIcon(icons, 32); big != nil {
		made[iconBig] = createIcon(big)
	}
	w32.SendMessage(hwnd, w32.WM_SETICON, iconSmall, w32.LPARAM(made[iconSmall]))
	w32.SendMessage(hwnd, w32.WM_SETICON, iconBig, w32.LPARAM(made[iconBig]))

	windowIconsMu.Lock()
	old := windowIcons[hwnd]
	if made == ([2]w32.HICON{}) {
		delete(windowIcons, hwnd)
	} else {
		windowIcons[hwnd] = made
	}
	windowIconsMu.Unlock()
	for _, h := range old {
		if h != 0 {
			w32.DestroyIcon(h)
		}
	}
}

// closestIcon returns the valid icon whose width is nearest to size,
// preferring the larger of two equally near.
func closestIcon(icons []platform.IconImage, size int) *platform.IconImage {
	var best *platform.IconImage
	for i := range icons {
		ic := &icons[i]
		if ic.Width <= 0 || ic.Height <= 0 || len(ic.Pix) < 4*ic.Width*ic.Height {
			continue
		}
		if best == nil {
			best = ic
			continue
		}
		db, di := abs(best.Width-size), abs(ic.Width-size)
		if di < db || (di == db && ic.Width > best.Width) {
			best = ic
		}
	}
	return best
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// iconBits converts RGBA pixels to what CreateIcon takes for a 32-bit
// icon: top-down BGRA colour rows, and a 1-bit AND mask with rows padded
// to 16 bits. The alpha channel carries the transparency, so the mask is
// all zero (draw every pixel).
func iconBits(ic *platform.IconImage) (and, xor []byte) {
	xor = make([]byte, 4*ic.Width*ic.Height)
	for i := 0; i < len(xor); i += 4 {
		xor[i], xor[i+1], xor[i+2], xor[i+3] = ic.Pix[i+2], ic.Pix[i+1], ic.Pix[i], ic.Pix[i+3]
	}
	maskRow := (ic.Width + 15) / 16 * 2
	return make([]byte, maskRow*ic.Height), xor
}

func createIcon(ic *platform.IconImage) w32.HICON {
	and, xor := iconBits(ic)
	return w32.CreateIcon(w32.GetModuleHandle(nil), int32(ic.Width), int32(ic.Height), 1, 32, &and[0], &xor[0])
}
