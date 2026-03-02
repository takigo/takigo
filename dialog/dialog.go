// Package dialog provides common dialog infrastructure and standard dialogs
// (message box, color chooser, font chooser, file dialog).
package dialog

import (
	"fmt"

	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/xlib"
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
	ResultNone   DialogResult = iota
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
	escBound    bool
	returnBound bool
}

// New creates a new dialog as a transient window over parent.
// The minWidth/minHeight set minimum dimensions; the dialog will grow
// to fit its content if needed.
func New(app widget.AppContext, parent *window.Window, title string, minWidth, minHeight int) *Dialog {
	d := &Dialog{
		App:      app,
		done:     make(chan struct{}),
		parent:   parent,
		minWidth: minWidth,
		minHeight: minHeight,
	}

	// Create transient toplevel at an initial size.
	d.Toplevel = toplevel.New(parent, "dialog", app,
		toplevel.Title(title),
		toplevel.TransientFor(parent),
	)

	// Set an initial geometry so pack has room to work.
	d.Toplevel.WmInfo.SetGeometry(fmt.Sprintf("%dx%d", minWidth, minHeight))

	// Button frame at bottom — pack first so it claims space before content.
	d.BtnFrame = newFrame(d.Toplevel.Window(), "buttons", app)
	pack.Pack(d.BtnFrame.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))

	// Content frame fills the remaining area.
	d.Content = newFrame(d.Toplevel.Window(), "content", app)
	pack.Pack(d.Content.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Register close handler for WM_DELETE_WINDOW routing.
	closeFn := func() {
		d.Close(ResultCancel)
	}
	app.RegisterCloseHandler(d.Toplevel.Window().XWindow, closeFn)

	// Bind Escape to cancel.
	app.Dispatcher().Bind(d.Toplevel.Window().XWindow, event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == xlib.XK_Escape {
			d.Close(ResultCancel)
		}
	})

	return d
}

// Run shows the dialog, grabs input, and blocks until Close is called.
// It runs a nested event loop so events continue to be processed.
func (d *Dialog) Run() DialogResult {
	tw := d.Toplevel.Window()

	// Auto-size: compute required size from content, use min dimensions as floor.
	// The content and buttons have been packed, so their ReqWidth/ReqHeight
	// reflect the minimum space needed.
	contentReq := d.Content.Window().ReqWidth
	btnReq := d.BtnFrame.Window().ReqWidth
	reqW := contentReq
	if btnReq > reqW {
		reqW = btnReq
	}
	reqH := d.Content.Window().ReqHeight + d.BtnFrame.Window().ReqHeight + 10 // 10 for padY

	width := d.minWidth
	if reqW+20 > width { // 20 for some horizontal margin
		width = reqW + 20
	}
	height := d.minHeight
	if reqH+20 > height { // 20 for some vertical margin
		height = reqH + 20
	}

	// Apply the computed size.
	tw.Width = width
	tw.Height = height
	tw.Display.XDisplay.MoveResizeWindow(tw.XWindow, tw.X, tw.Y, uint(width), uint(height))
	d.Toplevel.WmInfo.SetResizable(false, false)

	// Center over parent.
	centerOverParent(d.Toplevel.WmInfo, d.parent, width, height)

	// Re-arrange all containers with the final sizes.
	// ArrangeContainer only handles direct children, so we need ArrangeAll
	// to also re-layout children nested inside Content and BtnFrame.
	pack.ArrangeAll()

	d.Toplevel.Show()
	tw.Display.XDisplay.Flush()

	// Run a nested event loop until the dialog is closed.
	d.App.RunNestedLoop(d.done)

	return d.result
}

// Close sets the result, hides the dialog, and signals done.
func (d *Dialog) Close(result DialogResult) {
	d.result = result

	// Unregister the close handler.
	d.App.UnregisterCloseHandler(d.Toplevel.Window().XWindow)

	d.Toplevel.Hide()
	d.Toplevel.Destroy()

	select {
	case <-d.done:
		// Already closed.
	default:
		close(d.done)
	}
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
	sw := parent.Display.XDisplay.ScreenWidth(screen)
	sh := parent.Display.XDisplay.ScreenHeight(screen)
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
	info.Win.Display.XDisplay.MoveWindow(info.Win.XWindow, x, y)
}

// newFrame creates a frame with minimal requested size, suitable for dialog layout.
// Unlike frame.New() which defaults to 200x200, this creates frames that let
// pack's cavity algorithm work correctly with fixed-size dialog windows.
func newFrame(parent *window.Window, name string, app widget.AppContext, opts ...frame.FrameOption) *frame.Frame {
	f := frame.New(parent, name, app, opts...)
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
		btn := button.New(d.BtnFrame.Window(), "btn_"+b.text, d.App,
			button.Text(b.text),
			button.Command(func() {
				d.Close(res)
			}),
		)
		pack.Pack(btn.Window(), pack.SideOpt(pack.Left), pack.PadX(5))

		if b.isDefault && !d.returnBound {
			d.returnBound = true
			defResult := res
			d.App.Dispatcher().Bind(d.Toplevel.Window().XWindow, event.KeyPressMask, func(ev *event.Event) {
				if ev.KeySym == xlib.XK_Return {
					d.Close(defResult)
				}
			})
		}
	}
}
