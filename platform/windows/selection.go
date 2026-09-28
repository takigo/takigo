//go:build windows

package windows

import (
	"unsafe"

	w32 "github.com/msorc/takigo/internal/win32"
	"github.com/msorc/takigo/platform"
)

// --- SelectionManager implementation ---

func (d *WindowsDisplay) SetSelectionOwner(selection platform.AtomID, owner platform.WindowID, time platform.Timestamp) {
	d.clipMu.Lock()
	d.clipOwner[selection] = owner
	d.clipMu.Unlock()
}

func (d *WindowsDisplay) GetSelectionOwner(selection platform.AtomID) platform.WindowID {
	d.clipMu.Lock()
	defer d.clipMu.Unlock()
	return d.clipOwner[selection]
}

func (d *WindowsDisplay) ConvertSelection(selection, target, property platform.AtomID,
	requestor platform.WindowID, time platform.Timestamp) {
	// Check if we own the selection first.
	d.clipMu.Lock()
	data, hasData := d.clipData[selection]
	d.clipMu.Unlock()

	if hasData {
		// We own it — simulate the selection notify.
		raw := &WinRawEvent{
			Window:      requestor,
			MessageType: selection,
		}
		d.postEvent(&platform.RawEvent{
			Data:        raw,
			EventType:   platform.SelectionNotifyEvent,
			EventWindow: requestor,
		})
		_ = data
		return
	}

	// Read from Windows clipboard.
	text := d.readClipboard()
	if text != "" {
		d.clipMu.Lock()
		d.clipData[selection] = text
		d.clipMu.Unlock()
	}

	raw := &WinRawEvent{
		Window:      requestor,
		MessageType: selection,
	}
	d.postEvent(&platform.RawEvent{
		Data:        raw,
		EventType:   platform.SelectionNotifyEvent,
		EventWindow: requestor,
	})
}

func (d *WindowsDisplay) SendSelectionNotify(requestor platform.WindowID,
	selection, target, property platform.AtomID, time platform.Timestamp) {
	// No-op on Windows; clipboard is handled synchronously.
}

// SetClipboardText writes text to the Windows clipboard.
func (d *WindowsDisplay) SetClipboardText(text string) bool {
	d.writeClipboard(text)
	return true
}

// ClipboardText reads CF_UNICODETEXT from the Windows clipboard.
func (d *WindowsDisplay) ClipboardText() (string, bool) {
	return d.readClipboard(), true
}

// writeClipboard writes text to the Windows clipboard.
func (d *WindowsDisplay) writeClipboard(text string) {
	if !w32.OpenClipboard(d.rootHWND) {
		return
	}
	defer w32.CloseClipboard()

	w32.EmptyClipboard()

	utf16 := utf16FromString(text)
	size := uintptr(len(utf16) * 2)
	hMem := w32.GlobalAlloc(w32.GMEM_MOVEABLE, size)
	if hMem == 0 {
		return
	}

	ptr := w32.GlobalLockPtr(hMem)
	if ptr == 0 {
		w32.GlobalFree(hMem)
		return
	}

	dst := unsafe.Slice((*uint16)(ptrToUnsafe(ptr)), len(utf16))
	copy(dst, utf16)
	w32.GlobalUnlock(hMem)

	w32.SetClipboardData(w32.CF_UNICODETEXT, w32.HANDLE(hMem))
}

// readClipboard reads text from the Windows clipboard.
func (d *WindowsDisplay) readClipboard() string {
	if !w32.OpenClipboard(d.rootHWND) {
		return ""
	}
	defer w32.CloseClipboard()

	hData := w32.GetClipboardData(w32.CF_UNICODETEXT)
	if hData == 0 {
		return ""
	}

	ptr := w32.GlobalLockPtr(w32.HGLOBAL(hData))
	if ptr == 0 {
		return ""
	}
	defer w32.GlobalUnlock(w32.HGLOBAL(hData))

	// Read UTF-16 string. Limit to a reasonable max length.
	size := w32.GlobalSize(w32.HGLOBAL(hData))
	nchars := min(int(size/2), 1<<20)
	data := unsafe.Slice((*uint16)(ptrToUnsafe(ptr)), nchars)
	return utf16ToString(data)
}

// utf16FromString converts a Go string to null-terminated UTF-16.
func utf16FromString(s string) []uint16 {
	runes := []rune(s)
	result := make([]uint16, 0, len(runes)+1)
	for _, r := range runes {
		if r <= 0xFFFF {
			result = append(result, uint16(r))
		} else {
			// Surrogate pair.
			r -= 0x10000
			result = append(result, uint16(0xD800+(r>>10)))
			result = append(result, uint16(0xDC00+(r&0x3FF)))
		}
	}
	result = append(result, 0) // null terminator
	return result
}

// utf16ToString converts a null-terminated UTF-16 slice to Go string.
func utf16ToString(s []uint16) string {
	var runes []rune
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			break
		}
		if s[i] >= 0xD800 && s[i] <= 0xDBFF && i+1 < len(s) {
			hi := rune(s[i]) - 0xD800
			lo := rune(s[i+1]) - 0xDC00
			runes = append(runes, 0x10000+hi<<10+lo)
			i++
		} else {
			runes = append(runes, rune(s[i]))
		}
	}
	return string(runes)
}

// ptrToUnsafe converts a uintptr (from a Win32 API call) to unsafe.Pointer.
// go vet flags this as "possible misuse of unsafe.Pointer" but the conversion
// is correct: the uintptr comes from a Win32 syscall (GlobalLock) that returns
// a valid memory pointer. This pattern is standard in Go Win32 bindings.
func ptrToUnsafe(p uintptr) unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(&p))
}
