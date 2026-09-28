//go:build windows

package windows

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"

	"github.com/msorc/takigo/font"
	w32 "github.com/msorc/takigo/internal/win32"
	"github.com/msorc/takigo/platform"
)

func init() {
	// Windows requires the main thread for window creation and message pumping.
	// Lock the OS thread early so Go's main goroutine stays on thread 0.
	runtime.LockOSThread()
}

const windowClassName = "TakigoWindowClass"

// WindowsDisplay implements all platform capability interfaces using Win32 API.
type WindowsDisplay struct {
	hInstance w32.HINSTANCE
	rootHWND  w32.HWND
	screenDC  w32.HDC // device context for the screen
	atoms     *platform.Atoms

	// Screen metrics (cached at init).
	screenWidth    int
	screenHeight   int
	screenWidthMM  int
	screenHeightMM int
	screenDepth    int
	dpiX           int

	// Atom emulation: bidirectional maps for string interning.
	atomMu     sync.Mutex
	atomNext   uint64
	nameToAtom map[string]platform.AtomID
	atomToName map[platform.AtomID]string

	// events delivers events from the window procedure to the event loop.
	events *platform.EventQueue

	// Per-window data.
	windowMu   sync.RWMutex
	windowData map[w32.HWND]*windowInfo

	// GC emulation.
	gcMu   sync.Mutex
	gcNext uint64
	gcs    map[platform.GCID]*gcState

	// Pixmap emulation.
	pixmapMu   sync.Mutex
	pixmapNext uint64
	pixmaps    map[platform.PixmapID]*pixmapInfo

	// Selection/clipboard state.
	clipMu    sync.Mutex
	clipOwner map[platform.AtomID]platform.WindowID
	clipData  map[platform.AtomID]string

	// Input focus tracking.
	focusWindow platform.WindowID

	// First half of a standalone WM_CHAR surrogate pair.
	highSurrogate uint16

	// Cursor cache.
	cursorMu    sync.Mutex
	cursorCache map[uint]w32.HCURSOR
}

// Compile-time interface checks.
var _ platform.DisplayCore = (*WindowsDisplay)(nil)
var _ platform.WindowManager = (*WindowsDisplay)(nil)
var _ platform.Drawer = (*WindowsDisplay)(nil)
var _ platform.GCManager = (*WindowsDisplay)(nil)
var _ platform.PixmapManager = (*WindowsDisplay)(nil)
var _ platform.EventSource = (*WindowsDisplay)(nil)
var _ platform.GrabManager = (*WindowsDisplay)(nil)
var _ platform.SelectionManager = (*WindowsDisplay)(nil)
var _ platform.CursorManager = (*WindowsDisplay)(nil)
var _ platform.PropertyManager = (*WindowsDisplay)(nil)
var _ platform.InputMethodManager = (*WindowsDisplay)(nil)

// windowInfo holds per-window state.
type windowInfo struct {
	hwnd       w32.HWND
	parent     w32.HWND
	eventMask  int64
	bgPixel    uint64
	isOverride bool // override-redirect (popup)
	isTopLevel bool // top-level window (parent was display root)
}

// NewDisplayServer initializes the Windows display and returns a composed
// DisplayServer along with a FontOpener for the default screen.
func NewDisplayServer(displayName string) (platform.DisplayServer, font.FontOpener, error) {

	// Set DPI awareness (best effort).
	_ = w32.SetProcessDpiAwareness(w32.PROCESS_PER_MONITOR_DPI_AWARE)

	hInstance := w32.GetModuleHandle(nil)
	if hInstance == 0 {
		return nil, nil, fmt.Errorf("GetModuleHandle failed")
	}

	d := &WindowsDisplay{
		hInstance:   hInstance,
		events:      platform.NewEventQueue(),
		windowData:  make(map[w32.HWND]*windowInfo),
		nameToAtom:  make(map[string]platform.AtomID),
		atomToName:  make(map[platform.AtomID]string),
		atomNext:    100, // start after predefined atoms
		gcs:         make(map[platform.GCID]*gcState),
		gcNext:      1,
		pixmaps:     make(map[platform.PixmapID]*pixmapInfo),
		pixmapNext:  1,
		clipOwner:   make(map[platform.AtomID]platform.WindowID),
		clipData:    make(map[platform.AtomID]string),
		cursorCache: make(map[uint]w32.HCURSOR),
		atoms: &platform.Atoms{
			// Windows has no X11 atoms; the Windows backend emulates
			// them with fixed IDs that match the X11 predefined values
			// for source compatibility.
			WMName:        platform.AtomID(1),
			String:        platform.AtomID(2),
			WMNormalHints: platform.AtomID(3),
			Primary:       platform.AtomID(4),
			Secondary:     platform.AtomID(5),
			Atom:          platform.AtomID(6),
			Cardinal:      platform.AtomID(7),
			Window:        platform.AtomID(8),
		},
	}

	// Install our WndProc.
	w32.SetGlobalWndProc(d.wndProc)

	// Register window class.
	className := syscall.StringToUTF16Ptr(windowClassName)
	wc := w32.WNDCLASSEXW{
		CbSize:        uint32(unsafe.Sizeof(w32.WNDCLASSEXW{})),
		Style:         w32.CS_HREDRAW | w32.CS_VREDRAW | w32.CS_DBLCLKS,
		LpfnWndProc:   w32.GetWndProcPtr(),
		HInstance:     hInstance,
		HCursor:       w32.LoadCursor(0, w32.MAKEINTRESOURCE(w32.IDC_ARROW)),
		HbrBackground: w32.HBRUSH(w32.GetStockObject(w32.WHITE_BRUSH)),
		LpszClassName: className,
	}
	atom := w32.RegisterClassEx(&wc)
	if atom == 0 {
		return nil, nil, fmt.Errorf("RegisterClassEx failed")
	}

	// Query screen metrics.
	d.screenDC = w32.GetDC(0)
	d.screenWidth = int(w32.GetSystemMetrics(w32.SM_CXSCREEN))
	d.screenHeight = int(w32.GetSystemMetrics(w32.SM_CYSCREEN))
	d.screenWidthMM = int(w32.GetDeviceCaps(d.screenDC, w32.HORZSIZE))
	d.screenHeightMM = int(w32.GetDeviceCaps(d.screenDC, w32.VERTSIZE))
	d.screenDepth = int(w32.GetDeviceCaps(d.screenDC, w32.BITSPIXEL))
	d.dpiX = int(w32.GetDeviceCaps(d.screenDC, w32.LOGPIXELSX))

	// Create a hidden root window (message-only parent).
	rootHWND := w32.CreateWindowEx(0,
		className,
		syscall.StringToUTF16Ptr(""),
		0, // no style — hidden
		0, 0, 1, 1,
		0, 0, hInstance, nil)
	if rootHWND == 0 {
		return nil, nil, fmt.Errorf("CreateWindowEx for root failed")
	}
	d.rootHWND = rootHWND
	d.windowData[rootHWND] = &windowInfo{
		hwnd:    rootHWND,
		bgPixel: 0x00FFFFFF, // white
	}

	ds := platform.NewDisplayServer(
		d, // DisplayCore
		d, // WindowManager
		d, // Drawer
		d, // GCManager
		d, // PixmapManager
		d, // EventSource
		d, // GrabManager
		d, // SelectionManager
		d, // CursorManager
		d, // PropertyManager
		d, // InputMethodManager
	)
	return ds, d.FontOpener(d.DefaultScreen()), nil
}

// EventParser creates a WindowsEventParser.
func (d *WindowsDisplay) EventParser() platform.EventParser {
	return &EventParser{}
}

// FontOpener creates a WindowsFontOpener for font loading.
func (d *WindowsDisplay) FontOpener(screen int) font.FontOpener {
	return &FontOpener{screenDC: d.screenDC, resolveDC: d.getDrawableDC}
}

// --- DisplayCore ---

func (d *WindowsDisplay) Close() {
	if d.screenDC != 0 {
		w32.ReleaseDC(0, d.screenDC)
		d.screenDC = 0
	}
	if d.rootHWND != 0 {
		w32.DestroyWindow(d.rootHWND)
		d.rootHWND = 0
	}
}

func (d *WindowsDisplay) DefaultScreen() int                      { return 0 }
func (d *WindowsDisplay) DefaultRootWindow() platform.WindowID    { return fromHWND(d.rootHWND) }
func (d *WindowsDisplay) RootWindow(screen int) platform.WindowID { return fromHWND(d.rootHWND) }
func (d *WindowsDisplay) DefaultDepth(screen int) int             { return d.screenDepth }
func (d *WindowsDisplay) ScreenWidth(screen int) int              { return d.screenWidth }
func (d *WindowsDisplay) ScreenHeight(screen int) int             { return d.screenHeight }
func (d *WindowsDisplay) ScreenWidthMM(screen int) int            { return d.screenWidthMM }
func (d *WindowsDisplay) ScreenHeightMM(screen int) int           { return d.screenHeightMM }
func (d *WindowsDisplay) WhitePixel(screen int) uint64            { return 0x00FFFFFF }
func (d *WindowsDisplay) BlackPixel(screen int) uint64            { return 0x00000000 }
func (d *WindowsDisplay) ConnectionNumber() int                   { return -1 }
func (d *WindowsDisplay) Sync(discard bool)                       {} // noop on Windows
func (d *WindowsDisplay) Flush()                                  {} // noop on Windows
func (d *WindowsDisplay) Pending() int                            { return d.events.Len() }
func (d *WindowsDisplay) ResourceManagerString() string           { return "" }
func (d *WindowsDisplay) Atoms() *platform.Atoms                  { return d.atoms }

// PumpEvents implements event.EventPumper for main-thread message pumping.
// It reports WM_QUIT (PostQuitMessage) so the event loop quits.
func (d *WindowsDisplay) PumpEvents() bool {
	var msg w32.MSG
	for w32.PeekMessage(&msg, 0, 0, 0, w32.PM_REMOVE) {
		if msg.Message == w32.WM_QUIT {
			return true
		}
		w32.TranslateMessage(&msg)
		w32.DispatchMessage(&msg)
	}
	return false
}

// internAtom returns the atom ID for the given name, creating one if needed.
func (d *WindowsDisplay) internAtom(name string, onlyIfExists bool) platform.AtomID {
	d.atomMu.Lock()
	defer d.atomMu.Unlock()

	if id, ok := d.nameToAtom[name]; ok {
		return id
	}
	if onlyIfExists {
		return 0
	}
	id := platform.AtomID(atomic.AddUint64(&d.atomNext, 1))
	d.nameToAtom[name] = id
	d.atomToName[id] = name
	return id
}

// getAtomName returns the name for an atom ID.
func (d *WindowsDisplay) getAtomName(atom platform.AtomID) string {
	d.atomMu.Lock()
	defer d.atomMu.Unlock()
	return d.atomToName[atom]
}

// getWindowInfo returns the info for a window, or nil.
func (d *WindowsDisplay) getWindowInfo(hwnd w32.HWND) *windowInfo {
	d.windowMu.RLock()
	defer d.windowMu.RUnlock()
	return d.windowData[hwnd]
}

// postEvent queues a raw event for the event loop.
func (d *WindowsDisplay) postEvent(ev *platform.RawEvent) {
	d.events.Push(ev)
}
