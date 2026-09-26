package platform

// WindowManager manages window creation, destruction, and manipulation.
type WindowManager interface {
	// CreateWindow creates a new window.
	CreateWindow(parent WindowID, x, y int, width, height, borderWidth uint,
		depth int, class uint, valueMask uint64, attrs *WindowAttrs) WindowID

	// CreateSimpleWindow creates a simple window.
	CreateSimpleWindow(parent WindowID, x, y int, width, height, borderWidth uint, border, background uint64) WindowID

	// DestroyWindow destroys a window.
	DestroyWindow(w WindowID)

	// MapWindow maps a window to the screen.
	MapWindow(w WindowID)

	// MapRaised maps and raises a window.
	MapRaised(w WindowID)

	// UnmapWindow unmaps a window.
	UnmapWindow(w WindowID)

	// RaiseWindow raises a window.
	RaiseWindow(w WindowID)

	// LowerWindow lowers a window.
	LowerWindow(w WindowID)

	// MoveWindow moves a window.
	MoveWindow(w WindowID, x, y int)

	// ResizeWindow resizes a window.
	ResizeWindow(w WindowID, width, height uint)

	// MoveResizeWindow moves and resizes a window.
	MoveResizeWindow(w WindowID, x, y int, width, height uint)

	// SelectInput sets the event mask for a window.
	SelectInput(w WindowID, eventMask int64)

	// StoreName sets the window title.
	StoreName(w WindowID, name string)

	// TranslateCoordinates translates coordinates from src to dst window.
	TranslateCoordinates(src, dst WindowID, srcX, srcY int) (dstX, dstY int)
}

// Drawer provides drawing primitives.
type Drawer interface {
	// FillRectangle fills a rectangle.
	FillRectangle(drawable DrawableID, gc GCID, x, y int, width, height uint)

	// DrawRectangle draws a rectangle outline.
	DrawRectangle(drawable DrawableID, gc GCID, x, y int, width, height uint)

	// DrawLine draws a line.
	DrawLine(drawable DrawableID, gc GCID, x1, y1, x2, y2 int)

	// DrawLines draws connected line segments.
	DrawLines(drawable DrawableID, gc GCID, points []Point, mode int)

	// FillPolygon fills a polygon.
	FillPolygon(drawable DrawableID, gc GCID, points []Point, shape, mode int)

	// FillArc fills an arc.
	FillArc(drawable DrawableID, gc GCID, x, y int, width, height uint, angle1, angle2 int)

	// DrawArc draws an arc outline.
	DrawArc(drawable DrawableID, gc GCID, x, y int, width, height uint, angle1, angle2 int)

	// ClearWindow clears the entire window.
	ClearWindow(w WindowID)

	// SetWindowBackground sets the background pixel attribute of a window.
	SetWindowBackground(w WindowID, pixel uint64)

	// ClearArea clears a rectangular area.
	ClearArea(w WindowID, x, y int, width, height uint, exposures bool)

	// CopyArea copies a rectangular area between drawables.
	CopyArea(src, dst DrawableID, gc GCID, srcX, srcY int, width, height uint, dstX, dstY int)

	// PutImageRGBA puts an RGBA image onto a drawable.
	PutImageRGBA(drawable DrawableID, gc GCID, depth int,
		rgbaData []byte, stride int, imgW, imgH int,
		srcX, srcY, dstX, dstY, w, h int, bgPixel uint64)

	// GetImageRGBA reads a w x h area of a drawable as opaque RGBA, for
	// blending partly transparent images over what is already drawn (Tk's
	// BlendComplexAlpha does XGetImage). It returns nil when unsupported.
	GetImageRGBA(drawable DrawableID, x, y, w, h int) []byte

	// SetDashes sets the dash pattern for a GC.
	SetDashes(gc GCID, dashOffset int, dashList []byte)
}

// GCManager manages graphics contexts.
type GCManager interface {
	// CreateGC creates a new graphics context.
	CreateGC(drawable DrawableID, valueMask uint64, values *GCValues) GCID

	// FreeGC frees a graphics context.
	FreeGC(gc GCID)

	// SetForeground sets the foreground color of a GC.
	SetForeground(gc GCID, pixel uint64)

	// SetBackground sets the background color of a GC.
	SetBackground(gc GCID, pixel uint64)

	// SetLineAttributes sets line drawing attributes.
	SetLineAttributes(gc GCID, lineWidth uint, lineStyle, capStyle, joinStyle int)

	// SetFillStyle sets the fill style of a GC (FillSolid, FillStippled, etc.).
	SetFillStyle(gc GCID, fillStyle int)

	// SetStipple sets the stipple pixmap (depth-1 bitmap) for a GC.
	SetStipple(gc GCID, stipple PixmapID)

	// SetTSOrigin sets the tile/stipple origin of a GC (XSetTSOrigin).
	SetTSOrigin(gc GCID, x, y int)
}

// PixmapManager manages offscreen pixmaps.
type PixmapManager interface {
	// CreatePixmap creates a pixmap.
	CreatePixmap(drawable DrawableID, width, height, depth uint) PixmapID

	// FreePixmap frees a pixmap.
	FreePixmap(pixmap PixmapID)

	// CreateBitmapFromData creates a depth-1 pixmap from XBM-format bit data.
	CreateBitmapFromData(drawable DrawableID, bits []byte, width, height uint) PixmapID
}

// EventSource provides raw event access.
type EventSource interface {
	// NextEvent blocks until the next event and returns it.
	NextEvent() *RawEvent

	// FilterEvent returns true if the event was consumed by input method.
	FilterEvent(ev *RawEvent) bool
}

// GrabManager manages pointer and keyboard grabs.
type GrabManager interface {
	// GrabPointer grabs the pointer.
	GrabPointer(grabWindow WindowID, ownerEvents bool, eventMask uint,
		pointerMode, keyboardMode int, confineTo WindowID, cursor CursorID, time Timestamp) int

	// UngrabPointer releases the pointer grab.
	UngrabPointer(time Timestamp)

	// GrabKeyboard grabs the keyboard.
	GrabKeyboard(grabWindow WindowID, ownerEvents bool,
		pointerMode, keyboardMode int, time Timestamp) int

	// UngrabKeyboard releases the keyboard grab.
	UngrabKeyboard(time Timestamp)
}

// SelectionManager manages clipboard/selection operations.
type SelectionManager interface {
	// SetSelectionOwner sets the selection owner.
	SetSelectionOwner(selection AtomID, owner WindowID, time Timestamp)

	// GetSelectionOwner returns the current selection owner.
	GetSelectionOwner(selection AtomID) WindowID

	// ConvertSelection requests selection conversion.
	ConvertSelection(selection, target, property AtomID, requestor WindowID, time Timestamp)

	// SendSelectionNotify sends a selection notify event.
	SendSelectionNotify(requestor WindowID, selection, target, property AtomID, time Timestamp)
}

// CursorManager manages cursors.
type CursorManager interface {
	// CreateFontCursor creates a cursor from the standard cursor font.
	CreateFontCursor(shape uint) CursorID

	// DefineCursor sets the cursor for a window.
	DefineCursor(w WindowID, cursor CursorID)

	// SetCursorShape creates and sets a cursor from an abstract shape ID
	// (cursor.Shape values). Each backend maps these to native cursors.
	SetCursorShape(w WindowID, shape uint)

	// UndefineCursor reverts a window to its parent's cursor.
	UndefineCursor(w WindowID)

	// FreeCursor frees a cursor.
	FreeCursor(cursor CursorID)
}

// PropertyManager manages window properties and atoms.
type PropertyManager interface {
	// InternAtom returns the atom for the given name.
	InternAtom(name string, onlyIfExists bool) AtomID

	// GetAtomName returns the name of an atom.
	GetAtomName(atom AtomID) string

	// SetWMProtocols sets the WM_PROTOCOLS property.
	SetWMProtocols(w WindowID, protocols []AtomID) int

	// SetWMNormalHints sets window sizing hints.
	SetWMNormalHints(w WindowID, hints *SizeHints)

	// SetWMHints sets window manager hints.
	SetWMHints(w WindowID, hints *WMHints)

	// SetClassHint sets the WM_CLASS property.
	SetClassHint(w WindowID, name, class string)

	// SetTransientForHint sets WM_TRANSIENT_FOR.
	SetTransientForHint(w WindowID, propWindow WindowID)

	// SetInputFocus sets the input focus.
	SetInputFocus(w WindowID, revertTo int, time Timestamp)

	// GetInputFocus returns the current input focus window.
	GetInputFocus() (WindowID, int)

	// ChangeProperty sets a window property.
	ChangeProperty(w WindowID, prop, propType AtomID, format int, mode int, data []byte, nelements int)

	// ChangePropertyString sets a string property.
	ChangePropertyString(w WindowID, property, typ AtomID, data string)

	// ChangePropertyAtoms sets an atom-type property.
	ChangePropertyAtoms(w WindowID, prop AtomID, atoms []AtomID)

	// GetWindowProperty reads a window property.
	GetWindowProperty(w WindowID, property AtomID, offset, length int64, delete bool) ([]byte, AtomID, int)

	// DeleteProperty deletes a window property.
	DeleteProperty(w WindowID, prop AtomID)

	// SendEvent sends an event to a window.
	SendEvent(w WindowID, propagate bool, eventMask int64, ev *RawEvent)

	// SendClientMessage sends a ClientMessage event.
	SendClientMessage(w, target WindowID, msgType AtomID, d0, d1, d2, d3, d4 int64)

	// IconifyWindow iconifies a window.
	IconifyWindow(w WindowID, screen int)

	// WithdrawWindow withdraws a window.
	WithdrawWindow(w WindowID, screen int)

	// SetIconName sets the icon name for a window.
	SetIconName(w WindowID, name string)
}

// InputMethodManager manages input method state.
type InputMethodManager interface {
	// InitIM initializes the input method.
	InitIM(root WindowID)

	// HasIM returns true if input method was initialized.
	HasIM() bool

	// SetICFocus notifies the input method that the given window gained focus.
	SetICFocus(w WindowID)

	// UnsetICFocus notifies the input method that a window lost focus.
	UnsetICFocus()
}

// DisplayCore provides core display connection methods.
type DisplayCore interface {
	// Close closes the display connection.
	Close()

	// DefaultScreen returns the default screen number.
	DefaultScreen() int

	// DefaultRootWindow returns the root window ID.
	DefaultRootWindow() WindowID

	// RootWindow returns the root window for a given screen.
	RootWindow(screen int) WindowID

	// DefaultDepth returns the default depth for a given screen.
	DefaultDepth(screen int) int

	// ScreenWidth returns the screen width in pixels.
	ScreenWidth(screen int) int

	// ScreenHeight returns the screen height in pixels.
	ScreenHeight(screen int) int

	// ScreenWidthMM returns the screen width in millimeters.
	ScreenWidthMM(screen int) int

	// ScreenHeightMM returns the screen height in millimeters.
	ScreenHeightMM(screen int) int

	// WhitePixel returns the white pixel value.
	WhitePixel(screen int) uint64

	// BlackPixel returns the black pixel value.
	BlackPixel(screen int) uint64

	// ConnectionNumber returns the file descriptor for the connection.
	ConnectionNumber() int

	// Sync flushes and waits for all requests to complete.
	Sync(discard bool)

	// Flush flushes the output buffer.
	Flush()

	// Pending returns the number of events in the queue.
	Pending() int

	// ResourceManagerString returns the X resource manager string
	// (e.g., Xft.dpi settings). Returns "" if not available.
	ResourceManagerString() string

	// Atoms returns the backend's resolved well-known atom table.
	// Guaranteed non-nil after NewDisplayServer returns.
	Atoms() *Atoms

	// EventParser returns the platform-specific event parser.
	EventParser() EventParser
}

// DisplayServer is the interface for a platform display server.
// It aggregates all capability interfaces for windowing functionality.
type DisplayServer interface {
	DisplayCore
	WindowManager
	Drawer
	GCManager
	PixmapManager
	EventSource
	GrabManager
	SelectionManager
	CursorManager
	PropertyManager
	InputMethodManager
}

// displayServer is the concrete implementation of DisplayServer.
// It composes all capability interfaces.
type displayServer struct {
	DisplayCore
	WindowManager
	Drawer
	GCManager
	PixmapManager
	EventSource
	GrabManager
	SelectionManager
	CursorManager
	PropertyManager
	InputMethodManager
}

// NewDisplayServer creates a new DisplayServer from its capability implementations.
func NewDisplayServer(
	core DisplayCore,
	wm WindowManager,
	drawer Drawer,
	gc GCManager,
	pixmap PixmapManager,
	eventSrc EventSource,
	grab GrabManager,
	sel SelectionManager,
	cursor CursorManager,
	prop PropertyManager,
	im InputMethodManager,
) DisplayServer {
	return &displayServer{
		DisplayCore:        core,
		WindowManager:      wm,
		Drawer:             drawer,
		GCManager:          gc,
		PixmapManager:      pixmap,
		EventSource:        eventSrc,
		GrabManager:        grab,
		SelectionManager:   sel,
		CursorManager:      cursor,
		PropertyManager:    prop,
		InputMethodManager: im,
	}
}
