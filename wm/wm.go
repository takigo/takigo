// Package wm provides window manager interaction for toplevel windows.
// It ports the essential subset of tk/unix/tkUnixWm.c: title, geometry,
// size constraints, resizable, transient, iconify, state, and protocols.
package wm

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/window"
)

// State represents a toplevel window's WM state.
type State int

const (
	StateNormal    State = iota
	StateIconic          // minimized
	StateWithdrawn       // not visible, not in taskbar
	StateZoomed          // maximized (EWMH)
)

// String returns the state name.
func (s State) String() string {
	switch s {
	case StateNormal:
		return "normal"
	case StateIconic:
		return "iconic"
	case StateWithdrawn:
		return "withdrawn"
	case StateZoomed:
		return "zoomed"
	default:
		return "normal"
	}
}

// WmInfo holds per-toplevel WM state. Stored on Window.WmData.
type WmInfo struct {
	Win *window.Window

	Title    string
	IconName string
	Class    string // WM_CLASS res_class

	// Geometry.
	UserX, UserY     int  // user-requested position
	UserW, UserH     int  // user-requested size (-1 = not set)
	NegativeX        bool // x from right edge
	NegativeY        bool // y from bottom edge
	PositionSet      bool // user has set position

	// Size constraints.
	MinWidth, MinHeight int
	MaxWidth, MaxHeight int // 0 = unconstrained
	ResizableW         bool
	ResizableH         bool

	// State.
	CurrentState State
	Withdrawn    bool

	// Transient.
	TransientFor *window.Window

	// Protocol handlers.
	Protocols map[platform.AtomID]func()

	// Atoms cached per-display.
	atoms *wmAtoms
}

// wmAtoms caches interned atoms.
type wmAtoms struct {
	NetWMName             platform.AtomID
	NetWMIconName         platform.AtomID
	NetWMState            platform.AtomID
	NetWMStateMaxVert     platform.AtomID
	NetWMStateMaxHorz     platform.AtomID
	NetWMStateFullscreen  platform.AtomID
	NetWMStateAbove       platform.AtomID
	NetWMPing             platform.AtomID
	UTF8String            platform.AtomID
	WMDeleteWindow        platform.AtomID
	WMProtocols           platform.AtomID
	WMTransientFor        platform.AtomID
}

var atomCache = map[platform.DisplayServer]*wmAtoms{}

func getAtoms(d platform.DisplayServer) *wmAtoms {
	if a, ok := atomCache[d]; ok {
		return a
	}
	a := &wmAtoms{
		NetWMName:            d.InternAtom("_NET_WM_NAME", false),
		NetWMIconName:        d.InternAtom("_NET_WM_ICON_NAME", false),
		NetWMState:           d.InternAtom("_NET_WM_STATE", false),
		NetWMStateMaxVert:    d.InternAtom("_NET_WM_STATE_MAXIMIZED_VERT", false),
		NetWMStateMaxHorz:    d.InternAtom("_NET_WM_STATE_MAXIMIZED_HORZ", false),
		NetWMStateFullscreen: d.InternAtom("_NET_WM_STATE_FULLSCREEN", false),
		NetWMStateAbove:      d.InternAtom("_NET_WM_STATE_ABOVE", false),
		NetWMPing:            d.InternAtom("_NET_WM_PING", false),
		UTF8String:           d.InternAtom("UTF8_STRING", false),
		WMDeleteWindow:       d.InternAtom("WM_DELETE_WINDOW", false),
		WMProtocols:          d.InternAtom("WM_PROTOCOLS", false),
		WMTransientFor:       d.InternAtom("WM_TRANSIENT_FOR", false),
	}
	atomCache[d] = a
	return a
}

// Init initializes WM state for a toplevel window.
// Call after the X window is created.
func Init(w *window.Window) *WmInfo {
	d := w.Display.Server
	atoms := getAtoms(d)

	info := &WmInfo{
		Win:         w,
		Title:       w.Name,
		Class:       "Takigo",
		UserW:       -1,
		UserH:       -1,
		MinWidth:    1,
		MinHeight:   1,
		ResizableW:  true,
		ResizableH:  true,
		CurrentState: StateNormal,
		Protocols:   make(map[platform.AtomID]func()),
		atoms:       atoms,
	}

	// Set default WM hints.
	d.SetWMHints(w.PlatformID, &platform.WMHints{
		Flags:        platform.InputHint | platform.StateHint,
		Input:        true,
		InitialState: platform.NormalState,
	})

	// Set WM_CLASS.
	d.SetClassHint(w.PlatformID, w.Name, info.Class)

	// Set initial WM_PROTOCOLS.
	info.updateProtocols()

	// Set initial size hints.
	info.updateSizeHints()

	return info
}

// SetTitle sets the window title.
func (info *WmInfo) SetTitle(title string) {
	info.Title = title
	d := info.Win.Display.Server
	w := info.Win.PlatformID

	// ICCCM: WM_NAME.
	d.StoreName(w, title)

	// EWMH: _NET_WM_NAME as UTF-8.
	data := []byte(title)
	d.ChangeProperty(w, info.atoms.NetWMName, info.atoms.UTF8String,
		8, platform.PropModeReplace, data, len(data))
}

// SetIconName sets the icon name.
func (info *WmInfo) SetIconName(name string) {
	info.IconName = name
	d := info.Win.Display.Server
	w := info.Win.PlatformID
	d.SetIconName(w, name)
	data := []byte(name)
	d.ChangeProperty(w, info.atoms.NetWMIconName, info.atoms.UTF8String,
		8, platform.PropModeReplace, data, len(data))
}

// Geometry returns the current geometry as "WxH+X+Y".
func (info *WmInfo) Geometry() string {
	w := info.Win
	xSign, ySign := "+", "+"
	x, y := w.X, w.Y
	if info.NegativeX {
		xSign = "-"
		x = -x
	}
	if info.NegativeY {
		ySign = "-"
		y = -y
	}
	return fmt.Sprintf("%dx%d%s%d%s%d", w.Width, w.Height, xSign, x, ySign, y)
}

// SetGeometry parses and applies a geometry string like "800x600+100+50".
func (info *WmInfo) SetGeometry(geom string) error {
	w, h, x, y, hasSize, hasPos, negX, negY, err := ParseGeometry(geom)
	if err != nil {
		return err
	}

	if hasSize {
		info.UserW = w
		info.UserH = h
		info.Win.Width = w
		info.Win.Height = h
	}
	if hasPos {
		info.PositionSet = true
		info.NegativeX = negX
		info.NegativeY = negY

		if negX {
			screen := info.Win.Display.Screen
			sw := info.Win.Display.Server.ScreenWidth(screen)
			info.UserX = sw - info.Win.Width - x
		} else {
			info.UserX = x
		}
		if negY {
			screen := info.Win.Display.Screen
			sh := info.Win.Display.Server.ScreenHeight(screen)
			info.UserY = sh - info.Win.Height - y
		} else {
			info.UserY = y
		}
		info.Win.X = info.UserX
		info.Win.Y = info.UserY
	}

	info.applyGeometry()
	return nil
}

// applyGeometry sends the geometry to the X server.
func (info *WmInfo) applyGeometry() {
	w := info.Win
	d := w.Display.Server
	if w.PlatformID == platform.WindowID(0) {
		return
	}
	d.MoveResizeWindow(w.PlatformID, w.X, w.Y, uint(w.Width), uint(w.Height))
	info.updateSizeHints()
}

// SetMinSize sets the minimum window size.
func (info *WmInfo) SetMinSize(w, h int) {
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	info.MinWidth = w
	info.MinHeight = h
	info.updateSizeHints()
}

// SetMaxSize sets the maximum window size. Use 0 for unconstrained.
func (info *WmInfo) SetMaxSize(w, h int) {
	info.MaxWidth = w
	info.MaxHeight = h
	info.updateSizeHints()
}

// SetResizable sets whether the window can be resized in each axis.
func (info *WmInfo) SetResizable(width, height bool) {
	info.ResizableW = width
	info.ResizableH = height
	info.updateSizeHints()
}

// updateSizeHints sends WM_NORMAL_HINTS to the X server.
func (info *WmInfo) updateSizeHints() {
	w := info.Win
	if w.PlatformID == platform.WindowID(0) {
		return
	}

	hints := &platform.SizeHints{
		Flags:      platform.PMinSize | platform.PResizeInc | platform.PWinGravity,
		MinWidth:   info.MinWidth,
		MinHeight:  info.MinHeight,
		WidthInc:   1,
		HeightInc:  1,
		WinGravity: platform.NorthWestGravity,
	}

	if info.PositionSet {
		hints.Flags |= platform.USPosition
		hints.X = w.X
		hints.Y = w.Y
	}

	if info.MaxWidth > 0 || info.MaxHeight > 0 {
		hints.Flags |= platform.PMaxSize
		hints.MaxWidth = info.MaxWidth
		hints.MaxHeight = info.MaxHeight
	}

	// Non-resizable: set min == max for that axis.
	if !info.ResizableW {
		hints.Flags |= platform.PMaxSize
		hints.MinWidth = w.Width
		hints.MaxWidth = w.Width
	}
	if !info.ResizableH {
		hints.Flags |= platform.PMaxSize
		hints.MinHeight = w.Height
		hints.MaxHeight = w.Height
	}

	w.Display.Server.SetWMNormalHints(w.PlatformID, hints)
}

// SetTransientFor marks this window as a transient (dialog) for the given parent.
// Pass nil to clear the transient relationship.
func (info *WmInfo) SetTransientFor(parent *window.Window) {
	info.TransientFor = parent
	d := info.Win.Display.Server
	if parent != nil && parent.PlatformID != platform.WindowID(0) {
		d.SetTransientForHint(info.Win.PlatformID, parent.PlatformID)
	} else {
		d.DeleteProperty(info.Win.PlatformID, info.atoms.WMTransientFor)
	}
}

// Iconify minimizes the window.
func (info *WmInfo) Iconify() {
	if info.TransientFor != nil {
		return // transient windows cannot be independently iconified
	}
	w := info.Win
	if w.PlatformID == platform.WindowID(0) {
		return
	}
	info.CurrentState = StateIconic
	w.Display.Server.IconifyWindow(w.PlatformID, w.Display.Screen)
}

// Deiconify restores the window from iconic/withdrawn state.
func (info *WmInfo) Deiconify() {
	w := info.Win
	if w.PlatformID == platform.WindowID(0) {
		return
	}
	info.CurrentState = StateNormal
	info.Withdrawn = false
	w.Display.Server.SetWMHints(w.PlatformID, &platform.WMHints{
		Flags:        platform.InputHint | platform.StateHint,
		Input:        true,
		InitialState: platform.NormalState,
	})
	w.Display.Server.MapWindow(w.PlatformID)
}

// Withdraw hides the window completely (not in taskbar).
func (info *WmInfo) Withdraw() {
	w := info.Win
	if w.PlatformID == platform.WindowID(0) {
		return
	}
	info.CurrentState = StateWithdrawn
	info.Withdrawn = true
	w.Display.Server.WithdrawWindow(w.PlatformID, w.Display.Screen)
}

// GetState returns the current WM state.
func (info *WmInfo) GetState() State {
	return info.CurrentState
}

// OnDeleteWindow registers a callback for the WM_DELETE_WINDOW protocol.
func (info *WmInfo) OnDeleteWindow(fn func()) {
	info.Protocols[info.atoms.WMDeleteWindow] = fn
	info.updateProtocols()
}

// OnProtocol registers a callback for a named WM protocol.
func (info *WmInfo) OnProtocol(name string, fn func()) {
	atom := info.Win.Display.Server.InternAtom(name, false)
	info.Protocols[atom] = fn
	info.updateProtocols()
}

// HandleClientMessage processes a ClientMessage event for WM protocols.
// Returns true if the event was handled.
func (info *WmInfo) HandleClientMessage(messageType platform.AtomID, data [5]int64) bool {
	if messageType != info.atoms.WMProtocols {
		return false
	}
	protocolAtom := platform.AtomID(data[0])

	// Handle _NET_WM_PING: reflect back to root.
	if protocolAtom == info.atoms.NetWMPing {
		// Would need to send event back to root. Skip for now.
		return true
	}

	if fn, ok := info.Protocols[protocolAtom]; ok {
		fn()
		return true
	}
	return false
}

// updateProtocols sets the WM_PROTOCOLS property.
func (info *WmInfo) updateProtocols() {
	w := info.Win
	if w.PlatformID == platform.WindowID(0) {
		return
	}

	// Always include WM_DELETE_WINDOW and _NET_WM_PING.
	atoms := []platform.AtomID{info.atoms.WMDeleteWindow, info.atoms.NetWMPing}

	// Add registered protocols.
	for atom := range info.Protocols {
		found := false
		for _, a := range atoms {
			if a == atom {
				found = true
				break
			}
		}
		if !found {
			atoms = append(atoms, atom)
		}
	}

	w.Display.Server.ChangePropertyAtoms(w.PlatformID, info.atoms.WMProtocols, atoms)
}

// ParseGeometry parses a geometry string "WxH+X+Y" (each part optional).
// Returns size, position, flags for which parts were present, and negative flags.
func ParseGeometry(geom string) (w, h, x, y int, hasSize, hasPos, negX, negY bool, err error) {
	s := strings.TrimPrefix(geom, "=")

	// Try to find size part (WxH).
	xIdx := strings.IndexAny(s, "+-")
	sizePart := s
	posPart := ""
	if xIdx >= 0 {
		sizePart = s[:xIdx]
		posPart = s[xIdx:]
	}

	if sizePart != "" {
		parts := strings.SplitN(sizePart, "x", 2)
		if len(parts) == 2 {
			w, err = strconv.Atoi(parts[0])
			if err != nil {
				return 0, 0, 0, 0, false, false, false, false, fmt.Errorf("bad geometry %q", geom)
			}
			h, err = strconv.Atoi(parts[1])
			if err != nil {
				return 0, 0, 0, 0, false, false, false, false, fmt.Errorf("bad geometry %q", geom)
			}
			hasSize = true
		} else if xIdx < 0 {
			// No size, no position — invalid.
			return 0, 0, 0, 0, false, false, false, false, fmt.Errorf("bad geometry %q", geom)
		}
	}

	if posPart != "" {
		// Parse position: {+|-}X{+|-}Y
		// Find the second +/- (the Y separator).
		rest := posPart
		var xStr, yStr string
		var xSign, ySign byte

		if len(rest) > 0 && (rest[0] == '+' || rest[0] == '-') {
			xSign = rest[0]
			rest = rest[1:]
		}
		// Find next +/-.
		nextIdx := strings.IndexAny(rest, "+-")
		if nextIdx < 0 {
			return 0, 0, 0, 0, false, false, false, false, fmt.Errorf("bad geometry %q: missing Y", geom)
		}
		xStr = rest[:nextIdx]
		ySign = rest[nextIdx]
		yStr = rest[nextIdx+1:]

		x, err = strconv.Atoi(xStr)
		if err != nil {
			return 0, 0, 0, 0, false, false, false, false, fmt.Errorf("bad geometry %q", geom)
		}
		y, err = strconv.Atoi(yStr)
		if err != nil {
			return 0, 0, 0, 0, false, false, false, false, fmt.Errorf("bad geometry %q", geom)
		}
		negX = xSign == '-'
		negY = ySign == '-'
		hasPos = true
	}

	return w, h, x, y, hasSize, hasPos, negX, negY, nil
}
