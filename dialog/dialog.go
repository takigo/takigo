// Package dialog provides common dialog infrastructure and standard dialogs
// (message box, color chooser, font chooser, file dialog).
package dialog

import (
	"fmt"

	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/grab"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/toplevel"
	"github.com/msorc/takigo/window"
	"github.com/msorc/takigo/wm"
)

// DialogResult identifies which button was pressed to close a dialog.
type DialogResult int

const (
	ResultNone DialogResult = iota
	ResultOK
	ResultCancel
	ResultYes
	ResultNo
	ResultAbort
	ResultRetry
	ResultIgnore
)

// Dialog is the base infrastructure for modal dialogs.
type Dialog struct {
	Toplevel *toplevel.Toplevel
	App      widget.AppContext
	Content  *frame.Frame // content area
	BtnFrame *frame.Frame // button row at bottom

	// Resizable lets the user resize the dialog; the size computed by
	// Run becomes its minimum.
	Resizable bool

	parent      *window.Window
	minWidth    int
	minHeight   int
	result      DialogResult
	done        chan struct{}
	closed      bool
	returnBound bool
	// defaultButton gets the focus while the dialog runs.
	defaultButton *window.Window
}

// New creates a new dialog as a transient window over parent.
// The minWidth/minHeight set minimum dimensions; the dialog will grow
// to fit its content if needed.
func New(parent widget.Caregiver, title string, minWidth, minHeight int) *Dialog {
	app := parent.AppContext()
	d := &Dialog{
		App:       app,
		done:      make(chan struct{}),
		parent:    parent.Window(),
		minWidth:  minWidth,
		minHeight: minHeight,
	}

	// Create transient toplevel at an initial size.
	d.Toplevel = toplevel.New(parent, "dialog",
		toplevel.Title(title),
		toplevel.TransientFor(parent),
	)

	// Set an initial geometry so pack has room to work.
	_ = d.Toplevel.WmInfo.SetGeometry(fmt.Sprintf("%dx%d", minWidth, minHeight))

	// Button frame at bottom — pack first so it claims space before content.
	d.BtnFrame = newFrame(d.Toplevel, "buttons")
	pack.Pack(d.BtnFrame, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))

	// Content frame fills the remaining area.
	d.Content = newFrame(d.Toplevel, "content")
	pack.Pack(d.Content, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Register close handler for WM_DELETE_WINDOW routing.
	closeFn := func() {
		d.Close(ResultCancel)
	}
	app.RegisterCloseHandler(d.Toplevel.Window().PlatformID, closeFn)

	// Bind Escape to cancel.
	d.bindKey(platform.XK_Escape, func() { d.Close(ResultCancel) })

	// Destroyed from elsewhere (its parent went away): end Run.
	d.Toplevel.Window().OnDestroy(func() { d.finish(ResultCancel, false) })

	return d
}

// bindKey runs fn for keysym pressed anywhere in the dialog, like Tk's
// bind $w <Key> on the dialog toplevel: keys go to the focus widget, so
// the handler matches on the toplevel of the event's window.
func (d *Dialog) bindKey(keysym platform.KeySym, fn func()) {
	d.bindKeyEvent(keysym, func(*event.Event) { fn() })
}

// bindKeyEvent is bindKey with the event, for handlers that depend on
// the focus window or the modifiers.
func (d *Dialog) bindKeyEvent(keysym platform.KeySym, fn func(ev *event.Event)) {
	tw := d.Toplevel.Window()
	d.App.Dispatcher().BindGlobalFor(tw.PlatformID, event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym != keysym || ev.Handled {
			return
		}
		for w := tw.Display.LookupWindow(ev.Window); w != nil; w = w.Parent {
			if w == tw {
				fn(ev)
				return
			}
			if w.IsTopLevel() {
				return
			}
		}
	})
}

// grabber is the App's optional grab support (takigo.App.GrabManager).
type grabber interface {
	GrabManager() *grab.Manager
}

// Run shows the dialog, takes a local grab, and blocks until Close is
// called or the dialog is destroyed.
// It runs a nested event loop so events continue to be processed.
func (d *Dialog) Run() DialogResult {
	tw := d.Toplevel.Window()

	// Auto-size: compute required size from content, use min dimensions as floor.
	// Pack and grid propagate requested sizes at idle time; arrange the
	// tree now, children first, so the nested containers' ReqWidth/ReqHeight
	// reflect the space their content needs.
	arrangeNow(tw)
	contentReq := d.Content.Window().ReqWidth
	btnReq := d.BtnFrame.Window().ReqWidth
	reqW := max(contentReq, btnReq)
	reqH := d.Content.Window().ReqHeight + d.BtnFrame.Window().ReqHeight + 10 // 10 for padY

	width := max(d.minWidth, reqW+20)   // 20 for some horizontal margin
	height := max(d.minHeight, reqH+20) // 20 for some vertical margin

	// Apply the computed size as the dialog's geometry, so later content
	// requests keep it (as with wm geometry in Tk).
	_ = d.Toplevel.WmInfo.SetGeometry(fmt.Sprintf("%dx%d", width, height))
	if d.Resizable {
		d.Toplevel.WmInfo.SetMinSize(width, height)
		d.Toplevel.WmInfo.SetResizable(true, true)
	} else {
		d.Toplevel.WmInfo.SetResizable(false, false)
	}

	// Center over parent.
	centerOverParent(d.Toplevel.WmInfo, d.parent, width, height)

	// Re-arrange the dialog's containers with the final sizes.
	// ArrangeContainer only handles direct children, so we need ArrangeAll
	// to also re-layout children nested inside Content and BtnFrame.
	pack.ArrangeAll(tw)
	grid.ArrangeAll(tw)

	d.Toplevel.Show()
	tw.Display.Server.Flush()

	// Like tk_dialog and tk_messageBox, take a local grab so the rest of
	// the application ignores input until the dialog closes.
	if g, ok := d.App.(grabber); ok {
		gm := g.GrabManager()
		prev := gm.Current()
		gm.Set(tw, false)
		defer func() {
			gm.Release()
			if prev != nil && !prev.IsDestroyed() {
				gm.Set(prev, false)
			}
		}()
	}

	// tk::SetFocusGrab: the focus goes to the default button (or the
	// dialog) while it runs, and back afterwards (RestoreFocusGrab).
	oldFocus := widget.FocusWindow(d.App)
	if d.defaultButton != nil {
		widget.Focus(d.App, d.defaultButton)
	} else {
		widget.Focus(d.App, tw)
	}
	defer func() {
		if oldFocus != nil && !oldFocus.IsDestroyed() {
			widget.Focus(d.App, oldFocus)
		}
	}()

	// Run a nested event loop until the dialog is closed.
	d.App.RunNestedLoop(d.done)

	return d.result
}

// Close sets the result, hides the dialog, and signals done.
func (d *Dialog) Close(result DialogResult) { d.finish(result, true) }

// finish ends the dialog once; destroy is false when its toplevel is
// already being destroyed.
func (d *Dialog) finish(result DialogResult, destroy bool) {
	// A flag rather than sync.Once: destroying the toplevel re-enters
	// finish through its destroy hook.
	if d.closed {
		return
	}
	d.closed = true
	d.result = result

	// Unregister the close handler.
	d.App.UnregisterCloseHandler(d.Toplevel.Window().PlatformID)

	if destroy {
		d.Toplevel.Hide()
		d.Toplevel.Destroy()
	}

	close(d.done)
}

// arrangeNow lays out every pack and grid container under w, children
// first, so each container's requested size is up to date.
func arrangeNow(w *window.Window) {
	for _, c := range w.Children {
		arrangeNow(c)
	}
	pack.ArrangeContainer(w)
	grid.ArrangeContainer(w)
}

// centerOverParent positions the dialog centered over the parent window.
func centerOverParent(info *wm.WmInfo, parent *window.Window, width, height int) {
	if parent == nil {
		return
	}

	px, py := parent.X, parent.Y
	pw, ph := parent.Width, parent.Height

	x := px + (pw-width)/2
	y := py + (ph-height)/2

	// Clamp to screen bounds.
	screen := parent.Display.Screen
	sw := parent.Display.Server.ScreenWidth(screen)
	sh := parent.Display.Server.ScreenHeight(screen)
	if x+width > sw {
		x = sw - width
	}
	if y+height > sh {
		y = sh - height
	}
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	info.Win.X = x
	info.Win.Y = y
	info.Win.Display.Server.MoveWindow(info.Win.PlatformID, x, y)
}

// newFrame creates a frame with minimal requested size, suitable for dialog layout.
// Unlike frame.New() which defaults to 200x200, this creates frames that let
// pack's cavity algorithm work correctly with fixed-size dialog windows.
func newFrame(parent widget.Caregiver, name string, opts ...frame.FrameOption) *frame.Frame {
	f := frame.New(parent, name, opts...)
	// Set small requested AND actual sizes so pack's cavity algorithm
	// works correctly. The actual size will be set by the parent's pack.
	f.Window().ReqWidth = 1
	f.Window().ReqHeight = 1
	f.Window().Width = 1
	f.Window().Height = 1
	return f
}

// dialogButton describes a button to add to a dialog.
type dialogButton struct {
	text      string
	result    DialogResult
	isDefault bool
}

// addButtons creates buttons in the dialog's button frame.
func addButtons(d *Dialog, buttons []dialogButton) {
	for _, b := range buttons {
		res := b.result // capture for closure
		btn := button.New(d.BtnFrame, "btn_"+b.text,
			button.Text(b.text),
			button.Command(func() {
				d.Close(res)
			}),
		)
		pack.Pack(btn, pack.SideOpt(pack.Left), pack.PadX(5))

		if b.isDefault && !d.returnBound {
			d.returnBound = true
			defResult := res
			d.defaultButton = btn.Window()
			d.bindKey(platform.XK_Return, func() { d.Close(defResult) })
		}
	}
}
