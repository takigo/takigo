//go:build windows

package windows

import (
	"encoding/binary"
	"sync"

	w32 "github.com/msorc/takigo/internal/win32"
	"github.com/msorc/takigo/platform"
)

// Per-window property storage for emulating X11 properties.
type windowProperties struct {
	mu         sync.Mutex
	properties map[platform.AtomID][]byte
	wmHints    *platform.WMHints
	sizeHints  *platform.SizeHints
	protocols  []platform.AtomID
}

var (
	propsMu sync.Mutex
	propsDB = make(map[platform.WindowID]*windowProperties)
)

func getProps(w platform.WindowID) *windowProperties {
	propsMu.Lock()
	defer propsMu.Unlock()
	p := propsDB[w]
	if p == nil {
		p = &windowProperties{
			properties: make(map[platform.AtomID][]byte),
		}
		propsDB[w] = p
	}
	return p
}

// --- PropertyManager implementation ---

func (d *WindowsDisplay) InternAtom(name string, onlyIfExists bool) platform.AtomID {
	return d.internAtom(name, onlyIfExists)
}

func (d *WindowsDisplay) GetAtomName(atom platform.AtomID) string {
	return d.getAtomName(atom)
}

func (d *WindowsDisplay) SetWMProtocols(w platform.WindowID, protocols []platform.AtomID) int {
	p := getProps(w)
	p.mu.Lock()
	p.protocols = make([]platform.AtomID, len(protocols))
	copy(p.protocols, protocols)
	p.mu.Unlock()
	return 0
}

func (d *WindowsDisplay) SetWMNormalHints(w platform.WindowID, hints *platform.SizeHints) {
	p := getProps(w)
	p.mu.Lock()
	p.sizeHints = hints
	p.mu.Unlock()

	// Apply min/max size constraints via SetWindowLong style and MINMAXINFO.
	// This is handled in WM_GETMINMAXINFO if we respond to it.
}

func (d *WindowsDisplay) SetWMHints(w platform.WindowID, hints *platform.WMHints) {
	p := getProps(w)
	p.mu.Lock()
	p.wmHints = hints
	p.mu.Unlock()
}

func (d *WindowsDisplay) SetClassHint(w platform.WindowID, name, class string) {
	// No direct equivalent on Windows; can be stored in properties.
}

func (d *WindowsDisplay) SetTransientForHint(w platform.WindowID, propWindow platform.WindowID) {
	// Set the owner window for transient windows (dialogs).
	// On Windows, this is done at creation time. For already-created windows,
	// we can use SetWindowLongPtr with GWL_HWNDPARENT (not well-supported).
}

func (d *WindowsDisplay) SetInputFocus(w platform.WindowID, revertTo int, time platform.Timestamp) {
	hwnd := toHWND(w)
	if hwnd == 0 {
		return
	}
	w32.SetFocus(hwnd)
	d.focusWindow = w
}

func (d *WindowsDisplay) GetInputFocus() (platform.WindowID, int) {
	hwnd := w32.GetFocus()
	return fromHWND(hwnd), platform.RevertToParent
}

func (d *WindowsDisplay) ChangeProperty(w platform.WindowID, prop, propType platform.AtomID,
	format int, mode int, data []byte, nelements int) {
	p := getProps(w)
	p.mu.Lock()
	p.properties[prop] = make([]byte, len(data))
	copy(p.properties[prop], data)
	p.mu.Unlock()
}

func (d *WindowsDisplay) ChangePropertyString(w platform.WindowID, property, typ platform.AtomID, data string) {
	d.ChangeProperty(w, property, typ, 8, platform.PropModeReplace, []byte(data), len(data))
}

func (d *WindowsDisplay) ChangePropertyAtoms(w platform.WindowID, prop platform.AtomID, atoms []platform.AtomID) {
	// Store atom list as raw bytes.
	data := make([]byte, len(atoms)*4)
	for i, a := range atoms {
		binary.NativeEndian.PutUint32(data[i*4:], uint32(a))
	}
	d.ChangeProperty(w, prop, d.atoms.Atom, 32, platform.PropModeReplace, data, len(atoms))
}

func (d *WindowsDisplay) GetWindowProperty(w platform.WindowID, property platform.AtomID,
	offset, length int64, shouldDelete bool) ([]byte, platform.AtomID, int) {
	p := getProps(w)
	p.mu.Lock()
	defer p.mu.Unlock()

	data, ok := p.properties[property]
	if !ok {
		return nil, 0, 0
	}

	if shouldDelete {
		delete(p.properties, property)
	}
	return data, d.atoms.String, 8
}

func (d *WindowsDisplay) DeleteProperty(w platform.WindowID, prop platform.AtomID) {
	p := getProps(w)
	p.mu.Lock()
	delete(p.properties, prop)
	p.mu.Unlock()
}

func (d *WindowsDisplay) SendEvent(w platform.WindowID, propagate bool, eventMask int64, ev *platform.RawEvent) {
	// Re-post the event to our channel.
	d.postEvent(ev)
}

func (d *WindowsDisplay) SendClientMessage(w, target platform.WindowID,
	msgType platform.AtomID, d0, d1, d2, d3, d4 int64) {
	raw := &WinRawEvent{
		Window:      w,
		MessageType: msgType,
		MessageData: [5]int64{d0, d1, d2, d3, d4},
	}
	d.postEvent(&platform.RawEvent{
		Data:        raw,
		EventType:   platform.ClientMessageEvent_,
		EventWindow: w,
	})
}

func (d *WindowsDisplay) IconifyWindow(w platform.WindowID, screen int) {
	w32.ShowWindow(toHWND(w), w32.SW_MINIMIZE)
}

func (d *WindowsDisplay) WithdrawWindow(w platform.WindowID, screen int) {
	w32.ShowWindow(toHWND(w), w32.SW_HIDE)
}

func (d *WindowsDisplay) SetIconName(w platform.WindowID, name string) {
	// Windows doesn't have a separate icon name; use SetWindowText.
	// The taskbar text is the same as the window title.
}
