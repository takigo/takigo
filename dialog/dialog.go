// Package dialog provides common dialog infrastructure and standard dialogs
// (message box, color chooser, font chooser, file dialog).
package dialog

import (
	"fmt"

	"github.com/msorc/takigo/event"
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

	parent      *window.Window
	minWidth    int
	minHeight   int
	result      DialogResult
	done        chan struct{}
	closed      bool
	returnBound bool
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
	app.Dispatcher().Bind(d.Toplevel.Window().PlatformID, event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == platform.XK_Escape {
			d.Close(ResultCancel)
		}
	})

	// Destroyed from elsewhere (its parent went away): end Run.
	d.Toplevel.Window().OnDestroy(func() { d.finish(ResultCancel, false) })

	return d
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
	// The content and buttons have been packed, so their ReqWidth/ReqHeight
	// reflect the minimum space needed.
	contentReq := d.Content.Window().ReqWidth
	btnReq := d.BtnFrame.Window().ReqWidth
	reqW := max(contentReq, btnReq)
	reqH := d.Content.Window().ReqHeight + d.BtnFrame.Window().ReqHeight + 10 // 10 for padY

	width := max(d.minWidth, reqW+20)   // 20 for some horizontal margin
	height := max(d.minHeight, reqH+20) // 20 for some vertical margin

	// Apply the computed size.
	tw.Width = width
	tw.Height = height
	tw.Display.Server.MoveResizeWindow(tw.PlatformID, tw.X, tw.Y, uint(width), uint(height))
	d.Toplevel.WmInfo.SetResizable(false, false)

	// Center over parent.
	centerOverParent(d.Toplevel.WmInfo, d.parent, width, height)

	// Re-arrange all containers with the final sizes.
	// ArrangeContainer only handles direct children, so we need ArrangeAll
	// to also re-layout children nested inside Content and BtnFrame.
	pack.ArrangeAll()

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
			d.App.Dispatcher().Bind(d.Toplevel.Window().PlatformID, event.KeyPressMask, func(ev *event.Event) {
				if ev.KeySym == platform.XK_Return {
					d.Close(defResult)
				}
			})
		}
	}
}
