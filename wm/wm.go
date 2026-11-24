// Package wm provides window manager interaction for toplevel windows.
// It ports the essential subset of tk/unix/tkUnixWm.c: title, geometry,
// size constraints, resizable, transient, iconify, state, icons and
// protocols.
package wm

import (
	"fmt"
	"image"
	"image/draw"
	"slices"
	"strconv"
	"strings"

	"github.com/takigo/takigo/bitmap"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/window"
)

// State represents a toplevel window's WM state.
type State int

const (
	StateNormal    State = iota
	StateIconic          // minimized
	StateWithdrawn       // not visible, not in taskbar
	StateZoomed          // maximized (EWMH)
	StateIcon            // the icon window of another toplevel
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
	case StateIcon:
		return "icon"
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
	UserX, UserY int // user-requested position
	UserW, UserH int // user-requested size (-1 = not set)

	// pending holds sizes the application requested whose ConfigureNotify
	// has not arrived yet.
	pending     [][2]int
	NegativeX   bool // x from right edge
	NegativeY   bool // y from bottom edge
	PositionSet bool // user has set position

	// Size constraints.
	MinWidth, MinHeight int
	MaxWidth, MaxHeight int // 0 = unconstrained
	ResizableW          bool
	ResizableH          bool

	// State. CurrentState is the state last asked for; State reports
	// the one the window is in.
	CurrentState State
	Withdrawn    bool
	mapped       bool // shown, as last requested or reported by Map/Unmap events
	everMapped   bool // false until the first map (Tk's WM_NEVER_MAPPED)

	// hints is the whole WM_HINTS property, so that setting one field
	// keeps the others (Tk's wmPtr->hints).
	hints                platform.WMHints
	iconBitmap, iconMask string
	iconWindow, iconFor  *window.Window

	// Transient.
	TransientFor *window.Window

	// Gridding (Tk_SetGrid): the window whose grid the toplevel resizes
	// in, its request in grid units and the pixel size of a unit.
	gridWin                     *window.Window
	reqGridWidth, reqGridHeight int
	widthInc, heightInc         int

	// Protocol handlers.
	Protocols map[platform.AtomID]func()

	// Atoms cached per-display.
	atoms *wmAtoms
}

// wmAtoms caches interned atoms and the display's default icon.
type wmAtoms struct {
	NetWMName            platform.AtomID
	NetWMIconName        platform.AtomID
	NetWMState           platform.AtomID
	NetWMStateMaxVert    platform.AtomID
	NetWMStateMaxHorz    platform.AtomID
	NetWMStateFullscreen platform.AtomID
	NetWMStateAbove      platform.AtomID
	NetWMPing            platform.AtomID
	UTF8String           platform.AtomID
	WMDeleteWindow       platform.AtomID
	WMProtocols          platform.AtomID
	WMTransientFor       platform.AtomID

	// defaultIcons is the "wm iconphoto -default" picture every toplevel
	// created later starts with (Tk's per-display default icon).
	defaultIcons []platform.IconImage
}

// getAtoms returns the WM atoms of disp, interning them on first use.
func getAtoms(disp *window.Display) *wmAtoms {
	if a, ok := disp.WMAtoms.(*wmAtoms); ok {
		return a
	}
	d := disp.Server
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
	disp.WMAtoms = a
	return a
}

// Init initializes WM state for a toplevel window.
// Call after the X window is created.
func Init(w *window.Window) *WmInfo {
	d := w.Display.Server
	atoms := getAtoms(w.Display)

	info := &WmInfo{
		Win:          w,
		Title:        w.Name,
		Class:        "Takigo",
		UserW:        -1,
		UserH:        -1,
		MinWidth:     1,
		MinHeight:    1,
		ResizableW:   true,
		ResizableH:   true,
		CurrentState: StateNormal,
		Protocols:    make(map[platform.AtomID]func()),
		atoms:        atoms,
		hints: platform.WMHints{
			Flags:        platform.InputHint | platform.StateHint,
			Input:        true,
			InitialState: platform.NormalState,
		},
	}

	// Tk's TkWmProtocolEventProc destroys the window when a
	// WM_DELETE_WINDOW message arrives and no handler was registered.
	info.OffDeleteWindow()

	info.updateHints()
	if len(atoms.defaultIcons) > 0 && w.PlatformID != 0 {
		d.SetWindowIcons(w.PlatformID, atoms.defaultIcons)
	}
	w.OnDestroy(info.forget)

	// Set WM_CLASS.
	d.SetClassHint(w.PlatformID, w.Name, info.Class)

	// Set initial WM_PROTOCOLS.
	info.updateProtocols()

	// Set initial size hints.
	info.updateSizeHints()

	// Store on window for protocol dispatch (mirrors TkWindow.wmInfoPtr).
	w.WmData = info

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

// SetIconPhoto sets the window's icon from one or more sizes of the same
// picture, like Tk's "wm iconphoto"; the window system picks the size it
// needs. With no images the icon is removed. On macOS it sets the
// application's Dock icon, as Tk does.
func (info *WmInfo) SetIconPhoto(imgs ...image.Image) {
	if info.Win.PlatformID == 0 {
		return
	}
	info.Win.Display.Server.SetWindowIcons(info.Win.PlatformID, iconImages(imgs))
}

// SetIconPhotoDefault is "wm iconphoto -default": it sets the window's
// icon and makes it the icon of every toplevel created afterwards on the
// same display. With no images later toplevels start without one.
func (info *WmInfo) SetIconPhotoDefault(imgs ...image.Image) {
	icons := iconImages(imgs)
	info.atoms.defaultIcons = icons
	if info.Win.PlatformID != 0 {
		info.Win.Display.Server.SetWindowIcons(info.Win.PlatformID, icons)
	}
}

// iconImages converts icon pictures to straight-alpha RGBA.
func iconImages(imgs []image.Image) []platform.IconImage {
	icons := make([]platform.IconImage, 0, len(imgs))
	for _, img := range imgs {
		if img == nil {
			continue
		}
		b := img.Bounds()
		n := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
		draw.Draw(n, n.Rect, img, b.Min, draw.Src)
		icons = append(icons, platform.IconImage{Width: b.Dx(), Height: b.Dy(), Pix: n.Pix})
	}
	return icons
}

// SetIconBitmap is "wm iconbitmap": a Tk bitmap (a built-in name or
// "@file.xbm") the window manager shows as the icon, through WM_HINTS.
// An empty name removes it. Only X11 window managers read it; Windows
// and macOS take "wm iconphoto" instead.
func (info *WmInfo) SetIconBitmap(name string) error {
	pix, err := info.bitmapPixmap(name, info.hints.IconPixmap)
	if err != nil {
		return err
	}
	info.iconBitmap, info.hints.IconPixmap = name, pix
	info.setHintFlag(platform.IconPixmapHint, name != "")
	return nil
}

// IconBitmap returns the name given to SetIconBitmap.
func (info *WmInfo) IconBitmap() string { return info.iconBitmap }

// SetIconMask is "wm iconmask": the bitmap whose set bits are the part of
// the icon bitmap that is shown. An empty name removes it.
func (info *WmInfo) SetIconMask(name string) error {
	pix, err := info.bitmapPixmap(name, info.hints.IconMask)
	if err != nil {
		return err
	}
	info.iconMask, info.hints.IconMask = name, pix
	info.setHintFlag(platform.IconMaskHint, name != "")
	return nil
}

// IconMask returns the name given to SetIconMask.
func (info *WmInfo) IconMask() string { return info.iconMask }

// bitmapPixmap makes a depth-1 pixmap of the named bitmap and frees old,
// which it replaces. An empty name gives no pixmap.
func (info *WmInfo) bitmapPixmap(name string, old platform.PixmapID) (platform.PixmapID, error) {
	var pix platform.PixmapID
	if name != "" {
		w, h, bits, ok := bitmap.Data(name)
		if !ok {
			return 0, fmt.Errorf("bitmap %q not defined", name)
		}
		if info.Win.PlatformID != 0 {
			pix = info.Win.Display.Server.CreateBitmapFromData(
				platform.WindowDrawable(info.Win.PlatformID), bits, uint(w), uint(h))
		}
	}
	if old != 0 {
		info.Win.Display.Server.FreePixmap(old)
	}
	return pix, nil
}

// SetIconPosition is "wm iconposition": where the window manager should
// put the icon. ok false removes the hint.
func (info *WmInfo) SetIconPosition(x, y int, ok bool) {
	info.hints.IconX, info.hints.IconY = x, y
	info.setHintFlag(platform.IconPositionHint, ok)
}

// IconPosition returns the icon position hint and whether one is set.
func (info *WmInfo) IconPosition() (x, y int, ok bool) {
	return info.hints.IconX, info.hints.IconY, info.hints.Flags&platform.IconPositionHint != 0
}

// SetIconWindow is "wm iconwindow": the toplevel icon is withdrawn and
// shown by the window manager as this window's icon. nil removes it, and
// the former icon window stays withdrawn.
func (info *WmInfo) SetIconWindow(icon *window.Window) error {
	var iconInfo *WmInfo
	if icon != nil {
		var ok bool
		if iconInfo, ok = icon.WmData.(*WmInfo); !ok {
			return fmt.Errorf("can't use %s as icon window: not at top level", icon.PathName)
		}
		if iconInfo.iconFor != nil && iconInfo.iconFor != info.Win {
			return fmt.Errorf("%s is already an icon for %s", icon.PathName, iconInfo.iconFor.PathName)
		}
	}
	if info.iconWindow != nil {
		if old, ok := info.iconWindow.WmData.(*WmInfo); ok {
			old.iconFor = nil
			old.Withdrawn = true
			old.CurrentState = StateWithdrawn
			old.hints.InitialState = platform.WithdrawnState
		}
	}
	info.iconWindow = icon
	info.hints.IconWindow = 0
	if iconInfo != nil {
		iconInfo.iconFor = info.Win
		if iconInfo.everMapped && !iconInfo.Withdrawn && icon.PlatformID != 0 {
			icon.Display.Server.WithdrawWindow(icon.PlatformID, icon.Display.Screen)
			window.MarkUnmapped(icon)
		}
		iconInfo.mapped = false
		info.hints.IconWindow = icon.PlatformID
	}
	info.setHintFlag(platform.IconWindowHint, iconInfo != nil)
	return nil
}

// IconWindow returns the window given to SetIconWindow.
func (info *WmInfo) IconWindow() *window.Window { return info.iconWindow }

func (info *WmInfo) setHintFlag(flag int64, on bool) {
	if on {
		info.hints.Flags |= flag
	} else {
		info.hints.Flags &^= flag
	}
	info.updateHints()
}

// updateHints sends WM_HINTS (UpdateHints).
func (info *WmInfo) updateHints() {
	if info.Win.PlatformID == 0 {
		return
	}
	h := info.hints
	info.Win.Display.Server.SetWMHints(info.Win.PlatformID, &h)
}

// forget frees what the window holds when it is destroyed.
func (info *WmInfo) forget() {
	_ = info.SetIconBitmap("")
	_ = info.SetIconMask("")
	if info.iconFor != nil {
		if owner, ok := info.iconFor.WmData.(*WmInfo); ok && owner.iconWindow == info.Win {
			owner.iconWindow = nil
			owner.hints.IconWindow = 0
			owner.hints.Flags &^= platform.IconWindowHint
			if !owner.Win.IsDestroyed() {
				owner.updateHints()
			}
		}
	}
	if info.iconWindow != nil {
		if icon, ok := info.iconWindow.WmData.(*WmInfo); ok {
			icon.iconFor = nil
		}
	}
}

// Geometry returns the current geometry as "WxH+X+Y".
func (info *WmInfo) Geometry() string {
	w := info.Win
	xSign, ySign := "+", "+"
	x, y := w.X, w.Y
	// A negative position is the distance from the right/bottom screen
	// edge, as wm geometry reports it (tkUnixWm.c ConfigureEvent).
	if info.NegativeX {
		xSign = "-"
		x = w.Display.Server.ScreenWidth(w.Display.Screen) - (w.X + w.Width)
	}
	if info.NegativeY {
		ySign = "-"
		y = w.Display.Server.ScreenHeight(w.Display.Screen) - (w.Y + w.Height)
	}
	return fmt.Sprintf("%dx%d%s%d%s%d", w.Width, w.Height, xSign, x, ySign, y)
}

// SetGeometry parses and applies a geometry string like "800x600+100+50".
func (info *WmInfo) SetGeometry(geom string) error {
	w, h, x, y, hasSize, hasPos, negX, negY, err := parseGeometry(geom)
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

// GeometryRequest ports the size part of TkWmUpdateGeom: a size given with
// wm geometry wins over the content's request.
func (info *WmInfo) GeometryRequest(reqW, reqH int) (int, int) {
	w, h := reqW, reqH
	if info.UserW > 0 {
		w = info.UserW
	}
	if info.UserH > 0 {
		h = info.UserH
	}
	return w, h
}

// ExpectSize records a size the application requested; see
// ConfigureNotify. Like WM_SYNC_PENDING it tells our own resizes from the
// user's; only the last few requests are kept.
func (info *WmInfo) ExpectSize(w, h int) {
	const keep = 16
	info.pending = append(info.pending, [2]int{w, h})
	if n := len(info.pending); n > keep {
		info.pending = slices.Clone(info.pending[n-keep:])
	}
}

// ConfigureNotify ports the size part of ConfigureEvent (tkUnixWm.c): a
// reported size that is neither the current one nor one the application
// asked for was made by the user (dragging a border) and becomes the
// window's geometry, which later content requests then keep. A dimension
// still equal to the request stays unset.
func (info *WmInfo) ConfigureNotify(width, height int) {
	for i, p := range info.pending {
		if p == [2]int{width, height} {
			// The request is answered, and so are the earlier ones the
			// window manager skipped.
			info.pending = info.pending[i+1:]
			return
		}
	}
	w := info.Win
	if width == w.Width && height == w.Height {
		return
	}
	if info.UserW > 0 || width != w.ReqWidth {
		info.UserW = width
	}
	if info.UserH > 0 || height != w.ReqHeight {
		info.UserH = height
	}
}

// applyGeometry sends the geometry to the X server.
func (info *WmInfo) applyGeometry() {
	w := info.Win
	d := w.Display.Server
	if w.PlatformID == platform.WindowID(0) {
		return
	}
	info.ExpectSize(w.Width, w.Height)
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

// SetGrid ports Tk_SetGrid: the toplevel resizes in steps of widthInc x
// heightInc pixels, gridWin's request being reqWidth x reqHeight units
// (a text widget's -setgrid). Only one window grids a toplevel at a time.
func (info *WmInfo) SetGrid(gridWin *window.Window, reqWidth, reqHeight, widthInc, heightInc int) {
	widthInc = max(widthInc, 1)
	heightInc = max(heightInc, 1)
	if info.gridWin != nil && info.gridWin != gridWin {
		return
	}
	if info.gridWin == gridWin && info.reqGridWidth == reqWidth && info.reqGridHeight == reqHeight &&
		info.widthInc == widthInc && info.heightInc == heightInc {
		return
	}
	info.gridWin = gridWin
	info.reqGridWidth, info.reqGridHeight = reqWidth, reqHeight
	info.widthInc, info.heightInc = widthInc, heightInc
	info.updateSizeHints()
}

// UnsetGrid ports Tk_UnsetGrid: the toplevel resizes by the pixel again.
func (info *WmInfo) UnsetGrid(gridWin *window.Window) {
	if info.gridWin != gridWin {
		return
	}
	info.gridWin = nil
	info.updateSizeHints()
}

// updateSizeHints sends WM_NORMAL_HINTS to the X server (UpdateSizeHints in
// tkUnixWm.c). For a gridded toplevel the base size is the part of the
// request that is not grid units, and the minimum and the increments are
// in units.
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
	if info.gridWin != nil {
		hints.Flags |= platform.PBaseSize
		hints.BaseWidth = max(w.ReqWidth-info.reqGridWidth*info.widthInc, 0)
		hints.BaseHeight = max(w.ReqHeight-info.reqGridHeight*info.heightInc, 0)
		hints.MinWidth = hints.BaseWidth + info.MinWidth*info.widthInc
		hints.MinHeight = hints.BaseHeight + info.MinHeight*info.heightInc
		hints.WidthInc, hints.HeightInc = info.widthInc, info.heightInc
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

// Iconify minimizes the window (TkpWmSetState with IconicState). A
// window not mapped yet starts iconic when it is, a withdrawn one is
// mapped iconic.
func (info *WmInfo) Iconify() {
	if info.TransientFor != nil {
		return // transient windows cannot be independently iconified
	}
	w := info.Win
	if w.PlatformID == platform.WindowID(0) {
		return
	}
	info.CurrentState = StateIconic
	info.hints.InitialState = platform.IconicState
	switch {
	case !info.everMapped:
		info.updateHints()
	case info.Withdrawn:
		info.Withdrawn = false
		info.updateHints()
		info.mapWindow(false)
	default:
		info.mapped = false
		w.Display.Server.IconifyWindow(w.PlatformID, w.Display.Screen)
	}
}

// Deiconify restores the window from iconic/withdrawn state.
func (info *WmInfo) Deiconify() {
	w := info.Win
	if w.PlatformID == platform.WindowID(0) {
		return
	}
	info.CurrentState = StateNormal
	info.Withdrawn = false
	info.hints.InitialState = platform.NormalState
	info.updateHints()
	info.mapWindow(false)
}

// Withdraw hides the window completely (not in taskbar). A window not
// mapped yet stays unmapped: Map leaves it alone.
func (info *WmInfo) Withdraw() {
	w := info.Win
	if w.PlatformID == platform.WindowID(0) {
		return
	}
	info.CurrentState = StateWithdrawn
	info.Withdrawn = true
	info.hints.InitialState = platform.WithdrawnState
	info.mapped = false
	if info.everMapped {
		w.Display.Server.WithdrawWindow(w.PlatformID, w.Display.Screen)
	}
}

// Map shows the toplevel in the state it was given (TkWmMapWindow): not at
// all when withdrawn, minimized when Iconify came first. raise raises it
// above its siblings too.
func (info *WmInfo) Map(raise bool) {
	if info.Win.PlatformID == 0 || info.Withdrawn || info.iconFor != nil {
		return
	}
	info.mapWindow(raise)
}

func (info *WmInfo) mapWindow(raise bool) {
	w := info.Win
	info.everMapped = true
	info.mapped = info.hints.InitialState != platform.IconicState
	window.SyncBackground(w)
	if raise {
		w.Display.Server.MapRaised(w.PlatformID)
	} else {
		w.Display.Server.MapWindow(w.PlatformID)
	}
	window.MarkMapped(w)
}

// MapNotify records that the toplevel was mapped or unmapped, also by
// the window manager when the user minimized or restored it
// (WrapperEventProc). A MapNotify still queued when the window was
// withdrawn is stale: only the application maps a withdrawn window.
func (info *WmInfo) MapNotify(mapped bool) {
	if !mapped {
		info.mapped = false
		window.MarkUnmapped(info.Win)
		return
	}
	if info.Withdrawn || info.iconFor != nil {
		return
	}
	info.mapped, info.everMapped = true, true
	info.hints.InitialState = platform.NormalState
	if info.CurrentState != StateZoomed {
		info.CurrentState = StateNormal
	}
	window.MarkMapped(info.Win)
}

// State returns the window's state as "wm state" reports it.
func (info *WmInfo) State() State {
	switch {
	case info.iconFor != nil:
		return StateIcon
	case info.Withdrawn:
		return StateWithdrawn
	case info.mapped || (!info.everMapped && info.hints.InitialState == platform.NormalState):
		if info.CurrentState == StateZoomed {
			return StateZoomed
		}
		return StateNormal
	}
	return StateIconic
}

// OnDeleteWindow registers a callback for the WM_DELETE_WINDOW protocol.
func (info *WmInfo) OnDeleteWindow(fn func()) {
	info.Protocols[info.atoms.WMDeleteWindow] = fn
	info.updateProtocols()
}

// OffDeleteWindow restores the default WM_DELETE_WINDOW behavior:
// destroying the window. Mirrors Tk's behavior when no user handler
// is installed.
func (info *WmInfo) OffDeleteWindow() {
	w := info.Win
	info.Protocols[info.atoms.WMDeleteWindow] = func() {
		window.DestroyWindow(w)
	}
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

	// _NET_WM_PING: send the message back to the root window unchanged
	// but for its window field, as TkWmProtocolEventProc does.
	if protocolAtom == info.atoms.NetWMPing {
		d := info.Win.Display
		d.Server.SendClientMessage(d.RootWindow, d.RootWindow, messageType,
			data[0], data[1], data[2], data[3], data[4])
		d.Server.Flush()
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
		if !slices.Contains(atoms, atom) {
			atoms = append(atoms, atom)
		}
	}

	w.Display.Server.ChangePropertyAtoms(w.PlatformID, info.atoms.WMProtocols, atoms)
}

// Geometry is a parsed "wm geometry" string: a size, a position, both
// or neither, with the position's edges (NegX: from the right, NegY: from
// the bottom).
type Geometry struct {
	Width, Height int
	X, Y          int
	HasSize       bool
	HasPos        bool
	NegX, NegY    bool
}

// ParseGeometry parses Tk's geometry string, [=]WxH[+-X+-Y] or +-X+-Y.
func ParseGeometry(geom string) (Geometry, error) {
	var g Geometry
	w, h, x, y, hasSize, hasPos, negX, negY, err := parseGeometry(geom)
	if err != nil {
		return g, err
	}
	return Geometry{Width: w, Height: h, X: x, Y: y, HasSize: hasSize, HasPos: hasPos, NegX: negX, NegY: negY}, nil
}

func parseGeometry(geom string) (w, h, x, y int, hasSize, hasPos, negX, negY bool, err error) {
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
