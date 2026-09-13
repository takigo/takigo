// Package button implements the button widget with press/hover interaction.
// It ports the button-specific parts of tk/generic/tkButton.c and
// library/button.tcl.
package button

import (
	"log"

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
}

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
	return func(b *Button) { b.PadX = screenunit.Px(p) }
}

// PadY sets vertical padding.
// Accepts int (pixels), float64 (rounded pixels), or string with unit suffix ("3p", "2m", "1c", "0.5i").
func PadY(p any) ButtonOption {
	return func(b *Button) { b.PadY = screenunit.Px(p) }
}

// Width sets the requested button width in characters (like Tk's -width option).
// When > 0, the button is at least this many "0" characters wide.
func Width(n int) ButtonOption {
	return func(b *Button) { b.WidthChars = n }
}

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
func (b *Button) computeGeometry() {
	if b.Font != nil && b.Text != "" {
		b.textWidth = b.Font.MeasureString(b.Text)
		m := b.Font.Metrics()
		b.textHeight = m.Linespace()
	} else {
		b.textWidth = 0
		b.textHeight = 0
	}

	contentW, contentH := widget.CompoundSize(b.Compound, b.Img, b.textWidth, b.textHeight)

	if b.WidthChars > 0 && b.Font != nil {
		if b.zeroCharWidth == 0 {
			b.zeroCharWidth = b.Font.MeasureString("0")
		}
		if minW := b.WidthChars * b.zeroCharWidth; minW > contentW {
			contentW = minW
		}
	}

	inset := b.BorderWidth + b.HighlightWidth
	w := b.Win
	w.ReqWidth = contentW + 2*b.PadX + 2*inset
	w.ReqHeight = contentH + 2*b.PadY + 2*inset
}

// Display draws the button.
func (b *Button) Display() {
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
	if border != nil && b.BorderWidth > 0 {
		hlw := b.HighlightWidth
		draw.Draw3DRectangle(d, w.Drawable(), gc, border,
			hlw, hlw, w.Width-2*hlw, w.Height-2*hlw, b.BorderWidth, relief)
	}

	// Draw content (image and/or text).
	inset := b.BorderWidth + b.HighlightWidth
	availW := max(0, w.Width-2*inset-2*b.PadX)
	availH := max(0, w.Height-2*inset-2*b.PadY)
	frameX := inset + b.PadX
	frameY := inset + b.PadY

	// Shift content 1px down-right when pressed (Tk behavior).
	pressOff := 0
	if b.pressed {
		pressOff = 1
	}

	hasImg := b.Img != nil
	hasText := b.Font != nil && b.Text != "" && fgCol != nil

	if hasImg && hasText && b.Compound != widget.CompoundNone {
		drawCompoundButton(b, w, frameX, frameY, availW, availH, bgPixel, fgCol, pressOff)
	} else if hasImg {
		imgW := b.Img.Width()
		imgH := b.Img.Height()
		ix, iy := widget.AnchorText(b.Anchor, frameX, frameY, availW, availH, imgW, imgH)
		b.Img.Draw(w.Display.Server, w.Drawable(), gc,
			w.Depth, 0, 0, imgW, imgH, ix+pressOff, iy+pressOff, bgPixel)
	} else if hasText {
		textX, textY := widget.AnchorText(b.Anchor, frameX, frameY,
			availW, availH, b.textWidth, b.textHeight)
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
	b.DrawHighlightBorder(b.HasFocus, 0)

	d.Flush()
}

// drawCompoundButton draws image and text in compound mode for a button.
func drawCompoundButton(b *Button, w *window.Window,
	frameX, frameY, availW, availH int, bgPixel uint64,
	fgCol *color.ColorRef, pressOff int) {

	imgW := b.Img.Width()
	imgH := b.Img.Height()
	contentW, contentH := widget.CompoundSize(b.Compound, b.Img, b.textWidth, b.textHeight)

	cx, cy := widget.AnchorText(b.Anchor, frameX, frameY, availW, availH, contentW, contentH)
	cx += pressOff
	cy += pressOff

	var imgX, imgY, textX, textY int
	switch b.Compound {
	case widget.CompoundLeft:
		imgX = cx
		imgY = cy + (contentH-imgH)/2
		textX = cx + imgW + 4
		textY = cy + (contentH-b.textHeight)/2
	case widget.CompoundRight:
		textX = cx
		textY = cy + (contentH-b.textHeight)/2
		imgX = cx + b.textWidth + 4
		imgY = cy + (contentH-imgH)/2
	case widget.CompoundTop:
		imgX = cx + (contentW-imgW)/2
		imgY = cy
		textX = cx + (contentW-b.textWidth)/2
		textY = cy + imgH + 4
	case widget.CompoundBottom:
		textX = cx + (contentW-b.textWidth)/2
		textY = cy
		imgX = cx + (contentW-imgW)/2
		imgY = cy + b.textHeight + 4
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
