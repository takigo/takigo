package ttk

import (
	"fmt"
	"math"

	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/internal/nanosvg"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// Toggleswitch is a TTK sliding on/off toggle switch.
type Toggleswitch struct {
	TtkWidget

	Text     string
	Font     font.Font
	Command  func()
	Variable *widget.Variable[bool]

	selected bool
	unsub    func()
	linked   *widget.Variable[bool]
}

// ToggleswitchOption configures a Toggleswitch.
type ToggleswitchOption func(*Toggleswitch)

// ToggleswitchText sets the label text.
func ToggleswitchText(s string) ToggleswitchOption {
	return func(t *Toggleswitch) { t.Text = s }
}

// ToggleswitchVar links the switch to a bool variable.
func ToggleswitchVar(v *widget.Variable[bool]) ToggleswitchOption {
	return func(t *Toggleswitch) { t.Variable = v }
}

// ToggleswitchCommand sets the callback invoked on toggle.
func ToggleswitchCommand(fn func()) ToggleswitchOption {
	return func(t *Toggleswitch) { t.Command = fn }
}

// NewToggleswitch creates a TTK toggle switch widget.
func NewToggleswitch(parent widget.Caregiver, name string, opts ...ToggleswitchOption) *Toggleswitch {
	app := parent.AppContext()
	win := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(win)

	win.Class = "Toggleswitch"
	ts := &Toggleswitch{}
	ts.Font, _ = app.FontRegistry().Get(font.TkDefaultFont)

	InitTtkWidget(&ts.TtkWidget, win, app, "TCheckbutton")
	win.OnDestroy(ts.Destroy)
	ts.DisplayFunc = ts.Display

	for _, opt := range opts {
		opt(ts)
	}

	ts.linkVariable()

	ts.computeSize()
	bindTtkHover(&ts.TtkWidget, app)
	bindToggleswitch(ts, app)
	return ts
}

// linkVariable syncs the state from -variable and follows its changes.
func (ts *Toggleswitch) linkVariable() {
	if ts.Variable == ts.linked {
		return
	}
	if ts.unsub != nil {
		ts.unsub()
		ts.unsub = nil
	}
	ts.linked = ts.Variable
	if ts.Variable == nil {
		return
	}
	sync := func() {
		ts.selected = ts.Variable.Get()
		if ts.selected {
			ts.State |= StateSelected
		} else {
			ts.State &^= StateSelected
		}
	}
	sync()
	ts.unsub = ts.Variable.OnChange(func(_, _ bool) {
		sync()
		ts.Display()
	})
}

// Configure sets options after creation.
func (ts *Toggleswitch) Configure(opts ...ToggleswitchOption) error {
	return configure(&ts.TtkWidget, ts, opts, ts.linkVariable, ts.computeSize)
}

// Toggleswitch2 geometry (tk/library/ttk/elements.tcl, troughData(2) and
// sliderData(2)) at 100% scaling; the images are SVGs scaled by
// ::tk::scalingPct.
const (
	tglTroughW, tglTroughH = 40, 20
	tglSliderW, tglSliderH = 20, 16
)

func tglScale() float64 { return float64(screenunit.ScalingPct()) / 100 }

// tglInset is the Tglswitch.focus ring plus the style's -padding 0.75p.
func tglInset() int { return 1 + screenunit.Pt(0.75).Pixels() }

func (ts *Toggleswitch) computeSize() {
	sc := tglScale()
	w := int(tglTroughW*sc) + 2*tglInset()
	h := int(tglTroughH*sc) + 2*tglInset()
	if ts.Font != nil && ts.Text != "" {
		w += 6 + ts.Font.MeasureString(ts.Text)
		h = max(h, ts.Font.Metrics().Linespace()+2*tglInset())
	}
	ts.Win.ReqWidth, ts.Win.ReqHeight = w, h
}

// troughColor ports CreateElements_genericLight's state map for the trough
// image on a background no lighter than #d9d9d9.
func (ts *Toggleswitch) troughColor() uint64 {
	st := ts.State
	sel := st&StateSelected != 0
	selBg := uint64(0x4a6984)
	if ts.Context != nil && ts.Context.Style != nil {
		if c := LookupColor(ts.Context.Style, "-selectbackground", 0, selBg); !colorIsLight(c) {
			selBg = c
		}
	}
	h, s, v := rgbToHsv(selBg)
	dv := -10.0
	if v < 80 {
		dv = 10
	}
	switch {
	case sel && st&StateDisabled != 0:
		return hsvToRgb(h, 33.3, 100)
	case sel && st&StatePressed != 0:
		return hsvToRgb(h, s, v+2*dv)
	case sel && st&StateHover != 0:
		return hsvToRgb(h, s, v+dv)
	case sel:
		return selBg
	case st&StateDisabled != 0:
		return 0xd1d1d1
	case st&StatePressed != 0:
		return 0xa3a3a3
	case st&StateHover != 0:
		return 0xb3b3b3
	}
	return 0xc3c3c3
}

// Display draws the Toggleswitch2 layout: the trough centred in the padding
// box and the slider at its left (or right when selected) end.
func (ts *Toggleswitch) Display() {
	if ts.Destroyed {
		return
	}
	win := ts.Win
	if win.PlatformID == 0 {
		return
	}
	d := win.Display.Server
	gc := win.GC
	width, height := win.Width, win.Height
	if width <= 0 || height <= 0 {
		return
	}
	if ts.pixmap == 0 || ts.pixmapW != width || ts.pixmapH != height {
		if ts.pixmap != 0 {
			d.FreePixmap(ts.pixmap)
		}
		ts.pixmap = d.CreatePixmap(win.Drawable(), uint(width), uint(height), uint(win.Depth))
		ts.pixmapW, ts.pixmapH = width, height
	}
	if ts.pixmap == 0 {
		return
	}
	pix := platform.PixmapDrawable(ts.pixmap)

	bg := uint64(0xd9d9d9)
	if ts.Context != nil && ts.Context.Style != nil {
		bg = LookupColor(ts.Context.Style, "-background", ts.State, bg)
	}
	d.SetForeground(gc, bg)
	d.FillRectangle(pix, gc, 0, 0, uint(width), uint(height))

	sc := tglScale()
	tw, th := int(tglTroughW*sc), int(tglTroughH*sc)
	sw := int(tglSliderW * sc)
	inset := tglInset()
	boxW := width - 2*inset
	if ts.Font != nil && ts.Text != "" {
		boxW = tw
	}
	tx := inset + (boxW-tw)/2
	ty := inset + (height-2*inset-th)/2
	sliderX := 0
	if ts.State&StateSelected != 0 {
		sliderX = tw - sw
	}
	px := toggleswitchPixels(tw, th, float32(sc), ts.troughColor(), sliderX, bg)
	d.PutImageRGBA(pix, gc, win.Depth, px, tw*4, tw, th, 0, 0, tx, ty, tw, th, bg)

	if ts.Font != nil && ts.Text != "" {
		if df, ok := ts.Font.(platform.DrawableFont); ok {
			fg := LookupColor(ts.Context.Style, "-foreground", ts.State, 0x000000)
			m := ts.Font.Metrics()
			r := uint16((fg>>16)&0xFF) * 257
			g := uint16((fg>>8)&0xFF) * 257
			b := uint16(fg&0xFF) * 257
			df.DrawString(pix, tx+tw+6, (height-m.Linespace())/2+m.Ascent, ts.Text, fg, r, g, b)
		}
	}

	if ts.State&StateFocus != 0 {
		d.SetForeground(gc, LookupColor(ts.Context.Style, "-focuscolor", ts.State, 0x000000))
		d.DrawRectangle(pix, gc, 0, 0, uint(width-1), uint(height-1))
	}

	d.CopyArea(pix, win.Drawable(), gc, 0, 0, uint(width), uint(height), 0, 0)
	d.Flush()
}

// toggleswitchPixels draws the Tglswitch2.trough image (troughData(2) of
// library/ttk/elements.tcl filled with trough) over bg, then the white
// Tglswitch2.slider image at sliderX, centred vertically, each blended like a
// Tk photo over what lies beneath.
func toggleswitchPixels(tw, th int, sc float32, trough uint64, sliderX int, bg uint64) []uint8 {
	troughSVG := fmt.Sprintf(`<svg width="40" height="20" version="1.1" xmlns="http://www.w3.org/2000/svg">
 <rect x="0" y="0" width="40" height="20" rx="10" fill='#%06x'/>
</svg>`, trough&0xffffff)
	const sliderSVG = `<svg width="20" height="16" version="1.1" xmlns="http://www.w3.org/2000/svg">
 <circle cx="10" cy="8" r="8" fill='#ffffff'/>
</svg>`
	out := make([]uint8, tw*th*4)
	for i := 0; i < len(out); i += 4 {
		out[i], out[i+1], out[i+2], out[i+3] = uint8(bg>>16), uint8(bg>>8), uint8(bg), 255
	}
	tp, w, h := draw.SVGImage(troughSVG, sc)
	nanosvg.Blend(out, tw, (tw-w)/2, (th-h)/2, tp, w, h)
	sp, w, h := draw.SVGImage(sliderSVG, sc)
	nanosvg.Blend(out, tw, sliderX, (th-h)/2, sp, w, h)
	return out
}

// colorIsLight ports ttk::toggleswitch::IsColorLight.
func colorIsLight(c uint64) bool {
	r, g, b := c>>16&0xff, c>>8&0xff, c&0xff
	return 5*g+2*r+b > 8*192
}

// rgbToHsv ports ttk::toggleswitch::Rgb2Hsv (s and v in percent).
func rgbToHsv(c uint64) (h, s, v float64) {
	r, g, b := float64(c>>16&0xff)/255, float64(c>>8&0xff)/255, float64(c&0xff)/255
	mn, mx := math.Min(r, math.Min(g, b)), math.Max(r, math.Max(g, b))
	d := mx - mn
	if mx != 0 {
		s = 100 * d / mx
	}
	v = 100 * mx
	switch {
	case d == 0:
	case mx == r:
		f := math.Mod((g-b)/d, 6)
		if f < 0 {
			f += 6
		}
		h = 60 * f
	case mx == g:
		h = 60 * ((b-r)/d + 2)
	default:
		h = 60 * ((r-g)/d + 4)
	}
	return
}

// hsvToRgb ports ttk::toggleswitch::Hsv2Rgb.
func hsvToRgb(h, s, v float64) uint64 {
	s, v = s/100, v/100
	c := s * v
	h /= 60
	x := c * (1 - math.Abs(math.Mod(h, 2)-1))
	var r, g, b float64
	switch int(h) {
	case 0:
		r, g = c, x
	case 1:
		r, g = x, c
	case 2:
		g, b = c, x
	case 3:
		g, b = x, c
	case 4:
		r, b = x, c
	default:
		r, b = c, x
	}
	m := v - c
	ch := func(f float64) uint64 { return uint64(math.Max(0, math.Min(255, math.Round(255*(f+m))))) }
	return ch(r)<<16 | ch(g)<<8 | ch(b)
}

// Toggle flips the switch state.
func (ts *Toggleswitch) Toggle() {
	if ts.State&StateDisabled != 0 {
		return
	}
	ts.selected = !ts.selected
	if ts.selected {
		ts.State |= StateSelected
	} else {
		ts.State &^= StateSelected
	}
	if ts.Variable != nil {
		ts.Variable.Set(ts.selected)
	}
	ts.Display()
	if ts.Command != nil {
		ts.Command()
	}
}

// Get returns the current on/off state.
func (ts *Toggleswitch) Get() bool {
	return ts.selected
}

func bindToggleswitch(ts *Toggleswitch, app widget.AppContext) {
	win := ts.Win

	app.Dispatcher().Bind(win.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button == 1 {
			ts.ChangeState(StatePressed, 0)
		}
	})

	app.Dispatcher().Bind(win.PlatformID, event.ButtonReleaseMask, func(ev *event.Event) {
		if ev.Button == 1 {
			wasPressed := ts.State&StatePressed != 0
			ts.ChangeState(0, StatePressed)
			if wasPressed && ev.X >= 0 && ev.X < win.Width && ev.Y >= 0 && ev.Y < win.Height {
				ts.Toggle()
			}
		}
	})
}

// Destroy cleans up the toggleswitch, unsubscribing from any linked Variable.
func (ts *Toggleswitch) Destroy() {
	if ts.unsub != nil {
		ts.unsub()
		ts.unsub = nil
	}
	ts.TtkWidget.Destroy()
}
