package ttk

import (
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// Labelframe is ttk::labelframe (tk/generic/ttk/ttkFrame.c): a frame with a
// border element and a text label placed on its top-left edge.
type Labelframe struct {
	TtkWidget
	Text        string
	Font        font.Font
	LabelWidget *window.Window // -labelwidget, drawn instead of -text
}

// LabelframeOption configures a Labelframe.
type LabelframeOption func(*Labelframe)

// LabelframeText sets -text.
func LabelframeText(s string) LabelframeOption {
	return func(lf *Labelframe) { lf.Text = s }
}

// LabelframePadding sets -padding as a Tk padding spec.
func LabelframePadding(spec string) LabelframeOption {
	return func(lf *Labelframe) { lf.SetWidgetOption("-padding", spec) }
}

// LabelframeLabelWidget sets -labelwidget: a widget (usually a child of the
// labelframe) shown in the label's place instead of -text.
func LabelframeLabelWidget(w window.Windower) LabelframeOption {
	return func(lf *Labelframe) { lf.LabelWidget = w.Window() }
}

// SetLabelWidget is "configure -labelwidget" after creation.
func (lf *Labelframe) SetLabelWidget(w window.Windower) {
	lf.LabelWidget = w.Window()
	lf.updateMargins()
	if lf.Win.GeomManager != nil {
		lf.Win.GeomManager.RequestProc(lf.Win)
	}
	lf.Display()
}

// LabelframeBorderWidth sets -borderwidth.
func LabelframeBorderWidth(bw int) LabelframeOption {
	return func(lf *Labelframe) { lf.SetWidgetOption("-borderwidth", bw) }
}

// NewLabelframe creates a themed labelframe.
func NewLabelframe(parent widget.Caregiver, name string, opts ...LabelframeOption) *Labelframe {
	app := parent.AppContext()
	win := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(win)

	lf := &Labelframe{}
	lf.Font, _ = app.FontRegistry().Get(font.TkDefaultFont)
	lf.DisplayFunc = lf.Display
	InitTtkWidget(&lf.TtkWidget, win, app, "TLabelframe")
	for _, opt := range opts {
		opt(lf)
	}
	lf.updateMargins()
	win.ReqWidth, win.ReqHeight = win.MinReqWidth, win.MinReqHeight
	return lf
}

// Tk's DEFAULT_LABELINSET: the label's left/right margin for -labelanchor nw.
const labelframeInset = 8

func (lf *Labelframe) borderWidth() int {
	if lf.Context == nil {
		return 2
	}
	return LookupInt(lf.Context.Style, "-borderwidth", 0, 2)
}

func (lf *Labelframe) labelSize() (int, int) {
	if lf.LabelWidget != nil {
		return lf.LabelWidget.ReqWidth, lf.LabelWidget.ReqHeight
	}
	if lf.Text == "" || lf.Font == nil {
		return 0, 0
	}
	return lf.Font.MeasureString(lf.Text), lf.Font.Metrics().Linespace()
}

// updateMargins ports LabelframeSize: the content margins are the padding
// plus the border, with the label (and its margins) added on top, and the
// minimum request is the label plus the border.
func (lf *Labelframe) updateMargins() {
	bw := lf.borderWidth()
	var pad Padding
	if lf.Context != nil {
		pad = LookupPadding(lf.Context.Style, "-padding", 0, Padding{})
	}
	lw, lh := lf.labelSize()
	lw += 2 * labelframeInset
	win := lf.Win
	win.InternalBorderLeft = pad.Left + bw
	win.InternalBorderRight = pad.Right + bw
	win.InternalBorderTop = pad.Top + bw + lh
	win.InternalBorderBottom = pad.Bottom + bw
	win.MinReqWidth = lw + 2*bw
	win.MinReqHeight = lh + 2*bw
}

// SetText changes -text and re-requests the geometry.
func (lf *Labelframe) SetText(s string) {
	lf.Text = s
	lf.updateMargins()
	if lf.Win.GeomManager != nil {
		lf.Win.GeomManager.RequestProc(lf.Win)
	}
	lf.Display()
}

// Display ports LabelframeDoLayout/LabelframeDisplay: the border is moved
// down by half the label height so its top edge runs through the label.
func (lf *Labelframe) Display() {
	win := lf.Win
	if lf.Destroyed || lf.Layout == nil || win.PlatformID == 0 {
		return
	}
	width, height := win.Width, win.Height
	if width <= 0 || height <= 0 {
		return
	}
	d := win.Display.Server
	gc := win.GC
	if lf.pixmap == 0 || lf.pixmapW != width || lf.pixmapH != height {
		if lf.pixmap != 0 {
			d.FreePixmap(lf.pixmap)
		}
		lf.pixmap = d.CreatePixmap(win.Drawable(), uint(width), uint(height), uint(win.Depth))
		lf.pixmapW, lf.pixmapH = width, height
	}
	if lf.pixmap == 0 {
		return
	}
	pix := platform.PixmapDrawable(lf.pixmap)

	style := lf.Context.Style
	bg := LookupColor(style, "-background", lf.State, 0xd9d9d9)
	d.SetForeground(gc, bg)
	d.FillRectangle(pix, gc, 0, 0, uint(width), uint(height))

	lw, lh := lf.labelSize()
	borderY := lh - lh/2
	lf.Layout.Place(lf.State, Box{0, borderY, width, height - borderY})
	lf.Layout.Draw(lf.State, DrawArgs{Display: d, Drawable: pix, GC: gc})

	if lwin := lf.LabelWidget; lwin != nil {
		// LabelframePlaceContent: the label widget sits in the label parcel.
		moved := lwin.X != labelframeInset || lwin.Y != 0
		lwin.X, lwin.Y = labelframeInset, 0
		lwin.Width, lwin.Height = max(lwin.ReqWidth, 1), max(lh, 1)
		if lwin.PlatformID != 0 {
			d.MoveResizeWindow(lwin.PlatformID, lwin.X, lwin.Y, uint(lwin.Width), uint(lwin.Height))
			d.RaiseWindow(lwin.PlatformID)
			if !lwin.IsMapped() && win.IsMapped() {
				d.MapWindow(lwin.PlatformID)
				window.MarkMapped(lwin)
			}
		}
		if moved {
			window.NotifyMoved(lwin)
		}
	} else if lw > 0 {
		lx := labelframeInset
		d.SetForeground(gc, bg)
		d.FillRectangle(pix, gc, lx, 0, uint(lw), uint(lh))
		fg := LookupColor(lf.labelStyle(), "-foreground", lf.State, 0x000000)
		if df, ok := lf.Font.(platform.DrawableFont); ok {
			r := uint16((fg>>16)&0xFF) * 257
			g := uint16((fg>>8)&0xFF) * 257
			b := uint16(fg&0xFF) * 257
			df.DrawString(pix, lx, lf.Font.Metrics().Ascent, lf.Text, fg, r, g, b)
		}
	}

	d.CopyArea(pix, win.Drawable(), gc, 0, 0, uint(width), uint(height), 0, 0)
	d.Flush()
}

func (lf *Labelframe) labelStyle() *Style {
	if lf.Theme != nil {
		if s := lf.Theme.ResolveStyle("TLabelframe.Label"); s != nil {
			return s
		}
	}
	return lf.Context.Style
}
