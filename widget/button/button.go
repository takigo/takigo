// Package button implements the button widget with press/hover interaction.
// It ports the button-specific parts of tk/generic/tkButton.c and
// library/button.tcl.
package button

import (
	"log"
	"strings"

	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// Button is an interactive widget that invokes a command when clicked.
type Button struct {
	widget.Base

	Text      string
	Command   func() // invoked on click
	Anchor    option.Anchor
	Justify   option.Justify
	Underline int // index of char to underline (-1 = none)

	// Image support.
	Img      widget.WidgetImage
	Compound widget.Compound

	// State.
	State      widget.State
	OverRelief option.Relief // relief when mouse is over button
	OffRelief  option.Relief // relief when not pressed

	// Active colors (used on hover).
	ActiveBackground *color.ColorRef
	ActiveForeground *color.ColorRef

	WidthChars int // requested width in characters (0 = auto)

	textWidth     int
	textHeight    int
	zeroCharWidth int  // cached MeasureString("0") for WidthChars
	pressed       bool // button1 is held down
	HasFocus      bool // whether button currently has keyboard focus

	// Default is Tk's -default: room for (and, when active, a drawn)
	// default ring around the button.
	Default DefaultState
}

// DefaultState is the value of Tk's -default option.
type DefaultState int

const (
	DefaultDisabled DefaultState = iota
	DefaultNormal
	DefaultActive
)

// ButtonOption configures a Button.
type ButtonOption func(*Button)

// Text sets the button text.
func Text(s string) ButtonOption {
	return func(b *Button) { b.Text = s }
}

// Command sets the callback invoked when the button is clicked.
func Command(fn func()) ButtonOption {
	return func(b *Button) { b.Command = fn }
}

// Background sets the background color.
func Background(name string) ButtonOption {
	return func(b *Button) {
		col, err := b.App.ColorCache().Get(name)
		if err != nil {
			log.Printf("button: failed to get color %q: %v", name, err)
			return
		}
		b.Background = col
		b.UpdateBorder()
	}
}

// Foreground sets the text color.
func Foreground(name string) ButtonOption {
	return func(b *Button) {
		col, err := b.App.ColorCache().Get(name)
		if err != nil {
			log.Printf("button: failed to get color %q: %v", name, err)
			return
		}
		b.Foreground = col
	}
}

// FontOpt sets the font.
func FontOpt(name string) ButtonOption {
	return func(b *Button) {
		f, err := b.App.FontRegistry().Get(name)
		if err != nil {
			log.Printf("button: failed to get font %q: %v", name, err)
			return
		}
		b.Font = f
		b.zeroCharWidth = f.MeasureString("0")
	}
}

// ImageOpt sets the image to display.
func ImageOpt(img widget.WidgetImage) ButtonOption {
	return func(b *Button) { b.Img = img }
}

// CompoundOpt sets how text and image are combined.
func CompoundOpt(c widget.Compound) ButtonOption {
	return func(b *Button) { b.Compound = c }
}

// BorderWidth sets the border width.
func BorderWidth(w int) ButtonOption {
	return func(b *Button) { b.BorderWidth = w }
}

// Default sets -default: DefaultNormal leaves 5px for a default ring,
// DefaultActive also draws it (TkpDisplayButton).
func Default(state DefaultState) ButtonOption {
	return func(b *Button) { b.Default = state }
}

// HighlightThickness sets -highlightthickness (width of the focus ring).
func HighlightThickness(w any) ButtonOption {
	return func(b *Button) { b.HighlightWidth = screenunit.PxOr(w, b.HighlightWidth) }
}

// ReliefOpt sets the border relief.
func ReliefOpt(r option.Relief) ButtonOption {
	return func(b *Button) { b.Relief = r }
}

// Anchor sets the text anchor.
func Anchor(a option.Anchor) ButtonOption {
	return func(b *Button) { b.Anchor = a }
}

// PadX sets horizontal padding.
// Accepts int (pixels), float64 (rounded pixels), or string with unit suffix ("3p", "2m", "1c", "0.5i").
func PadX(p any) ButtonOption {
	return func(b *Button) { b.PadX = screenunit.PxOr(p, b.PadX) }
}

// PadY sets vertical padding.
// Accepts int (pixels), float64 (rounded pixels), or string with unit suffix ("3p", "2m", "1c", "0.5i").
func PadY(p any) ButtonOption {
	return func(b *Button) { b.PadY = screenunit.PxOr(p, b.PadY) }
}

// Width sets the requested button width in characters (like Tk's -width option).
// When > 0, the button is at least this many "0" characters wide.
func Width(n int) ButtonOption {
	return func(b *Button) { b.WidthChars = n }
}

// --- Ttk-compatible aliases (prefix with Button) for consistent naming ---
// These aliases match the naming convention used by ttk widgets (ttk.ButtonText, etc.)
// allowing consistent option naming when both classic and ttk widgets are used.

// ButtonText is an alias for Text.
var ButtonText = Text

// ButtonCommand is an alias for Command.
var ButtonCommand = Command

// ButtonBackground is an alias for Background.
var ButtonBackground = Background

// ButtonForeground is an alias for Foreground.
var ButtonForeground = Foreground

// ButtonFontOpt is an alias for FontOpt.
var ButtonFontOpt = FontOpt

// ButtonImageOpt is an alias for ImageOpt.
var ButtonImageOpt = ImageOpt

// ButtonCompoundOpt is an alias for CompoundOpt.
var ButtonCompoundOpt = CompoundOpt

// ButtonBorderWidth is an alias for BorderWidth.
var ButtonBorderWidth = BorderWidth

// ButtonDefault is an alias for Default.
var ButtonDefault = Default

// ButtonHighlightThickness is an alias for HighlightThickness.
var ButtonHighlightThickness = HighlightThickness

// ButtonReliefOpt is an alias for ReliefOpt.
var ButtonReliefOpt = ReliefOpt

// ButtonAnchor is an alias for Anchor.
var ButtonAnchor = Anchor

// ButtonPadX is an alias for PadX.
var ButtonPadX = PadX

// ButtonPadY is an alias for PadY.
var ButtonPadY = PadY

// ButtonWidth is an alias for Width.
var ButtonWidth = Width

// New creates a new Button widget as a child of parent.
func New(parent widget.Caregiver, name string, opts ...ButtonOption) *Button {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(w)

	b := &Button{
		Anchor:     option.AnchorCenter,
		Justify:    option.JustifyCenter,
		Underline:  -1,
		OverRelief: option.ReliefRaised,
		OffRelief:  option.ReliefFlat,
	}
	widget.InitBase(&b.Base, w, app)
	b.SetDisplayProc(b.display)
	w.Class = "Button"

	// Button-specific defaults (Tk: padx=3m, pady=1m, borderwidth=1, highlightthickness=1).
	b.BorderWidth = widget.DefBorderWidth
	b.Relief = option.ReliefRaised
	b.PadX = screenunit.Px("3m")
	b.PadY = screenunit.Px("1m")
	b.HighlightWidth = 1

	// Active colors.
	if ac, err := app.ColorCache().Get(widget.DefActiveBackground); err == nil {
		b.ActiveBackground = ac.Ref()
	}
	if af, err := app.ColorCache().Get(widget.DefActiveForeground); err == nil {
		b.ActiveForeground = af.Ref()
	}

	// Buttons are focusable via Tab traversal.
	w.Flags |= window.FlagFocusable

	for _, opt := range opts {
		opt(b)
	}

	// Compute geometry.
	b.computeGeometry()

	// Update X window background.
	if b.Background != nil {
		w.BackgroundPixel = b.Background.Pixel
	}

	// Bind events.
	bindButton(b, app)

	// FocusIn/FocusOut — track focus state and redraw highlight.
	app.Dispatcher().Bind(w.PlatformID, event.FocusChangeMask, func(ev *event.Event) {
		if ev.Type == event.FocusInType {
			b.HasFocus = true
		} else {
			b.HasFocus = false
		}
		b.Display()
	})

	return b
}

// computeGeometry computes text/image size and sets requested window size.
// inset is butPtr->inset: border, highlight and the default ring's room.
func (b *Button) inset() int {
	inset := b.BorderWidth + b.HighlightWidth
	if b.Default != DefaultDisabled {
		inset += 5
	}
	return inset
}

// computeGeometry ports TkpComputeButtonGeometry (tk/unix/tkUnixButton.c)
// for TYPE_BUTTON: -width is in characters for text and in pixels when an
// image is shown, the compound gap is padX/padY, and non-Motif push buttons
// get 2 extra pixels each way.
func (b *Button) computeGeometry() {
	b.textWidth, b.textHeight = 0, 0
	if b.Font != nil {
		// Tk_ComputeTextLayout lays out "" as one empty line.
		for _, line := range strings.Split(b.Text, "\n") {
			b.textWidth = max(b.textWidth, b.Font.MeasureString(line))
			b.textHeight += b.Font.Metrics().Linespace()
		}
	}
	haveText := b.textWidth != 0 && b.textHeight != 0

	var width, height int
	switch {
	case b.Img != nil && b.Compound != widget.CompoundNone && haveText:
		width, height = b.Img.Width(), b.Img.Height()
		switch b.Compound {
		case widget.CompoundTop, widget.CompoundBottom:
			height += b.textHeight + b.PadY
			width = max(width, b.textWidth)
		case widget.CompoundLeft, widget.CompoundRight:
			width += b.textWidth + b.PadX
			height = max(height, b.textHeight)
		default:
			width = max(width, b.textWidth)
			height = max(height, b.textHeight)
		}
		if b.WidthChars > 0 {
			width = b.WidthChars
		}
		width += 2 * b.PadX
		height += 2 * b.PadY
	case b.Img != nil:
		width, height = b.Img.Width(), b.Img.Height()
		if b.WidthChars > 0 {
			width = b.WidthChars
		}
	default:
		width, height = b.textWidth, b.textHeight
		if b.WidthChars > 0 && b.Font != nil {
			if b.zeroCharWidth == 0 {
				b.zeroCharWidth = b.Font.MeasureString("0")
			}
			width = b.WidthChars * b.zeroCharWidth
		}
		width += 2 * b.PadX
		height += 2 * b.PadY
	}
	width += 2
	height += 2

	inset := b.inset()
	w := b.Win
	w.ReqWidth = width + 2*inset
	w.ReqHeight = height + 2*inset
}

// Display schedules a redraw at idle time; see widget.Base.EventuallyRedraw.
func (b *Button) Display() {
	b.EventuallyRedraw()
}

// display draws the button.
func (b *Button) display() {
	if b.Destroyed {
		return
	}
	w := b.Win
	if w.PlatformID == platform.WindowID(0) {
		return
	}

	d := w.Display.Server
	gc := w.GC

	// Choose colors based on state.
	bgPixel := uint64(0)
	var fgCol *color.ColorRef
	if b.Background != nil {
		bgPixel = b.Background.Pixel
	}
	if b.Foreground != nil {
		fgCol = b.Foreground.Ref()
	}

	if b.State == widget.StateActive && b.ActiveBackground != nil {
		bgPixel = b.ActiveBackground.Pixel
	}
	if b.State == widget.StateActive && b.ActiveForeground != nil {
		fgCol = b.ActiveForeground
	}

	// Fill background.
	d.SetForeground(gc, bgPixel)
	d.FillRectangle(w.Drawable(), gc, 0, 0, uint(w.Width), uint(w.Height))

	// Determine relief.
	relief := b.Relief
	if b.pressed {
		relief = option.ReliefSunken
	}

	// Draw border.
	border := b.Border
	if b.State == widget.StateActive && b.Background != nil {
		// Recompute border from active background for proper shading.
		if b.ActiveBackground != nil {
			border = draw.NewBorder(b.ActiveBackground.Red, b.ActiveBackground.Green, b.ActiveBackground.Blue)
		}
	}
	if border != nil && relief != option.ReliefFlat {
		ringInset := b.HighlightWidth
		ring := b.Border
		if b.HighlightBackground != nil {
			ring = draw.NewBorder(b.HighlightBackground.Red, b.HighlightBackground.Green, b.HighlightBackground.Blue)
		}
		rect := func(in, bw int, rel option.Relief) {
			if rel == option.ReliefFlat {
				d.SetForeground(gc, ring.BgPixel)
				for i := range bw {
					d.DrawRectangle(w.Drawable(), gc, in+i, in+i,
						uint(w.Width-2*(in+i)-1), uint(w.Height-2*(in+i)-1))
				}
				return
			}
			draw.Draw3DRectangle(d, w.Drawable(), gc, ring, in, in, w.Width-2*in, w.Height-2*in, bw, rel)
		}
		switch b.Default {
		case DefaultActive:
			// 2px space, 1px sunken ring, 2px space (TkpDisplayButton).
			rect(ringInset, 2, option.ReliefFlat)
			rect(ringInset+2, 1, option.ReliefSunken)
			rect(ringInset+3, 2, option.ReliefFlat)
			ringInset += 5
		case DefaultNormal:
			rect(0, 5, option.ReliefFlat)
			ringInset += 5
		}
		draw.Draw3DRectangle(d, w.Drawable(), gc, border,
			ringInset, ringInset, w.Width-2*ringInset, w.Height-2*ringInset, b.BorderWidth, relief)
	}

	// Draw content (image and/or text).
	inset := b.inset()

	// Shift content 1px down-right when pressed (Tk behavior).
	pressOff := 0
	if b.pressed {
		pressOff = 1
	}

	hasImg := b.Img != nil
	hasText := b.Font != nil && b.Text != "" && fgCol != nil

	if hasImg && hasText && b.Compound != widget.CompoundNone {
		drawCompoundButton(b, w, bgPixel, fgCol, pressOff)
	} else if hasImg {
		imgW := b.Img.Width()
		imgH := b.Img.Height()
		ix, iy := widget.ComputeAnchor(b.Anchor, w.Width, w.Height, inset, 0, 0, imgW, imgH)
		b.Img.Draw(w.Display.Server, w.Drawable(), gc,
			w.Depth, 0, 0, imgW, imgH, ix+pressOff, iy+pressOff, bgPixel)
	} else if hasText {
		textX, textY := widget.ComputeAnchor(b.Anchor, w.Width, w.Height, inset, b.PadX, b.PadY, b.textWidth, b.textHeight)
		textX += pressOff
		textY += pressOff
		m := b.Font.Metrics()
		baseline := textY + m.Ascent
		if df, ok := b.Font.(platform.DrawableFont); ok {
			df.DrawString(w.Drawable(), textX, baseline, b.Text,
				fgCol.Pixel, fgCol.Red, fgCol.Green, fgCol.Blue)
		}
	}

	// Draw focus highlight ring.
	// The focus ring shrink-wraps the button, not the default ring's room.
	focusPad := 0
	if b.Default == DefaultNormal {
		focusPad = 5
	}
	b.DrawHighlightBorder(b.HasFocus, focusPad)

	d.Flush()
}

// drawCompoundButton draws image and text in compound mode for a button.
func drawCompoundButton(b *Button, w *window.Window,
	bgPixel uint64,
	fgCol *color.ColorRef, pressOff int) {

	imgW := b.Img.Width()
	imgH := b.Img.Height()
	contentW, contentH := imgW, imgH
	switch b.Compound {
	case widget.CompoundLeft, widget.CompoundRight:
		contentW, contentH = imgW+b.PadX+b.textWidth, max(imgH, b.textHeight)
	case widget.CompoundTop, widget.CompoundBottom:
		contentW, contentH = max(imgW, b.textWidth), imgH+b.PadY+b.textHeight
	case widget.CompoundCenter:
		contentW, contentH = max(imgW, b.textWidth), max(imgH, b.textHeight)
	}

	inset := b.inset()
	cx, cy := widget.ComputeAnchor(b.Anchor, w.Width, w.Height, inset, b.PadX, b.PadY, contentW, contentH)
	cx += pressOff
	cy += pressOff

	var imgX, imgY, textX, textY int
	switch b.Compound {
	case widget.CompoundLeft:
		imgX = cx
		imgY = cy + (contentH-imgH)/2
		textX = cx + imgW + b.PadX
		textY = cy + (contentH-b.textHeight)/2
	case widget.CompoundRight:
		textX = cx
		textY = cy + (contentH-b.textHeight)/2
		imgX = cx + b.textWidth + b.PadX
		imgY = cy + (contentH-imgH)/2
	case widget.CompoundTop:
		imgX = cx + (contentW-imgW)/2
		imgY = cy
		textX = cx + (contentW-b.textWidth)/2
		textY = cy + imgH + b.PadY
	case widget.CompoundBottom:
		textX = cx + (contentW-b.textWidth)/2
		textY = cy
		imgX = cx + (contentW-imgW)/2
		imgY = cy + b.textHeight + b.PadY
	case widget.CompoundCenter:
		imgX = cx + (contentW-imgW)/2
		imgY = cy + (contentH-imgH)/2
		textX = cx + (contentW-b.textWidth)/2
		textY = cy + (contentH-b.textHeight)/2
	}

	// Draw image.
	b.Img.Draw(w.Display.Server, w.Drawable(), w.GC,
		w.Depth, 0, 0, imgW, imgH, imgX, imgY, bgPixel)

	// Draw text.
	if b.Font != nil && fgCol != nil {
		m := b.Font.Metrics()
		baseline := textY + m.Ascent
		if df, ok := b.Font.(platform.DrawableFont); ok {
			df.DrawString(w.Drawable(), textX, baseline, b.Text,
				fgCol.Pixel, fgCol.Red, fgCol.Green, fgCol.Blue)
		}
	}
}

// Invoke executes the button's command.
func (b *Button) Invoke() {
	if b.State == widget.StateDisabled {
		return
	}
	if b.Command != nil {
		b.Command()
	}
}

// SetText updates the button's label text and refreshes its size and display.
func (b *Button) SetText(s string) {
	b.Text = s
	b.computeGeometry()
	b.Display()
}

// Configure applies options to the button.
func (b *Button) Configure(opts ...option.Option) {
	option.Apply(b, opts)
	b.UpdateBorder()
	b.computeGeometry()
	if b.Background != nil {
		b.Win.BackgroundPixel = b.Background.Pixel
	}
	b.Display()
}

// Destroy cleans up the button.
func (b *Button) Destroy() {
	if b.Destroyed {
		return
	}
	b.Destroyed = true
	window.DestroyWindow(b.Win)
}
