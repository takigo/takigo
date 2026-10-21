package dialog

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/entry"
	"github.com/takigo/takigo/widget/frame"
	"github.com/takigo/takigo/widget/label"
	"github.com/takigo/takigo/widget/scale"
)

// colorConfig holds ChooseColor options.
type colorConfig struct {
	title        string
	initialColor string // "#RRGGBB"
}

// ColorOption configures ChooseColor.
type ColorOption func(*colorConfig)

func ColorTitle(s string) ColorOption   { return func(c *colorConfig) { c.title = s } }
func ColorInitial(s string) ColorOption { return func(c *colorConfig) { c.initialColor = s } }

// colorChooser is the colour chooser: the sliders, the preview and the
// hex entry, which all show the same colour.
type colorChooser struct {
	dialog   *Dialog
	rScale   *scale.Scale
	gScale   *scale.Scale
	bScale   *scale.Scale
	hexEntry *entry.Entry
	preview  *frame.Frame

	chosen string // the colour as "#RRGGBB"
	// fromEntry tells the preview that the sliders move because of the hex
	// entry, so that the entry keeps the text the user typed.
	fromEntry bool
}

// updatePreview shows the sliders' colour in the preview and in the hex
// entry.
func (c *colorChooser) updatePreview() {
	rv := int(c.rScale.Get())
	gv := int(c.gScale.Get())
	bv := int(c.bScale.Get())
	hex := fmt.Sprintf("#%02x%02x%02x", rv, gv, bv)
	c.chosen = hex

	// Update preview frame background.
	c.preview.Configure(frame.Background(hex))

	// Update hex entry.
	if !c.fromEntry {
		c.hexEntry.SetText(hex)
	}
}

// applyHex takes a "#RRGGBB" colour and moves the sliders to it. It does
// nothing when the text is not a colour.
func (c *colorChooser) applyHex(s string) {
	r, g, b, ok := parseHexColorStrict(s)
	if !ok {
		return
	}
	c.chosen = fmt.Sprintf("#%02x%02x%02x", r, g, b)
	c.fromEntry = true
	c.rScale.Set(float64(r))
	c.gScale.Set(float64(g))
	c.bScale.Set(float64(b))
	c.fromEntry = false
}

// color returns the chosen colour; the hex entry has the last word.
func (c *colorChooser) color() string {
	if r, g, b, ok := parseHexColorStrict(c.hexEntry.GetText()); ok {
		return fmt.Sprintf("#%02x%02x%02x", r, g, b)
	}
	return c.chosen
}

// newColorChooser builds the colour chooser dialog over parent.
func newColorChooser(parent widget.Caregiver, title, initial string) *colorChooser {
	r, g, b := parseHexColor(initial)
	d := New(parent, title, 350, 280)
	c := &colorChooser{dialog: d}

	// Sliders frame.
	slidersFrame := newFrame(d.Content, "sliders")
	pack.Pack(slidersFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	// R slider.
	rFrame := newFrame(slidersFrame, "rframe")
	pack.Pack(rFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))
	rLabel := label.New(rFrame, "rlabel", label.Text("R:"))
	pack.Pack(rLabel, pack.SideOpt(pack.Left), pack.PadX(5))
	c.rScale = scale.New(rFrame, "rscale",
		scale.OrientOpt(scale.Horizontal),
		scale.FromOpt(0), scale.ToOpt(255),
		scale.ValueOpt(float64(r)),
		scale.ResolutionOpt(1),
		scale.ShowValueOpt(true),
		scale.CommandOpt(func(v float64) { c.updatePreview() }),
	)
	pack.Pack(c.rScale, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillX), pack.Expand(true))

	// G slider.
	gFrame := newFrame(slidersFrame, "gframe")
	pack.Pack(gFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))
	gLabel := label.New(gFrame, "glabel", label.Text("G:"))
	pack.Pack(gLabel, pack.SideOpt(pack.Left), pack.PadX(5))
	c.gScale = scale.New(gFrame, "gscale",
		scale.OrientOpt(scale.Horizontal),
		scale.FromOpt(0), scale.ToOpt(255),
		scale.ValueOpt(float64(g)),
		scale.ResolutionOpt(1),
		scale.ShowValueOpt(true),
		scale.CommandOpt(func(v float64) { c.updatePreview() }),
	)
	pack.Pack(c.gScale, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillX), pack.Expand(true))

	// B slider.
	bFrame := newFrame(slidersFrame, "bframe")
	pack.Pack(bFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))
	bLabel := label.New(bFrame, "blabel", label.Text("B:"))
	pack.Pack(bLabel, pack.SideOpt(pack.Left), pack.PadX(5))
	c.bScale = scale.New(bFrame, "bscale",
		scale.OrientOpt(scale.Horizontal),
		scale.FromOpt(0), scale.ToOpt(255),
		scale.ValueOpt(float64(b)),
		scale.ResolutionOpt(1),
		scale.ShowValueOpt(true),
		scale.CommandOpt(func(v float64) { c.updatePreview() }),
	)
	pack.Pack(c.bScale, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillX), pack.Expand(true))

	// Preview + hex entry frame.
	bottomFrame := newFrame(d.Content, "bottom")
	pack.Pack(bottomFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	// Color preview.
	c.preview = frame.New(bottomFrame, "preview",
		frame.Width(60), frame.Height(40),
		frame.BorderWidth(2), frame.Relief(option.ReliefSunken),
	)
	pack.Pack(c.preview, pack.SideOpt(pack.Left), pack.PadX(10))

	// Hex entry. A colour typed in it moves the sliders as soon as it is
	// a valid one (Tk's color entry validates the text).
	hexLabel := label.New(bottomFrame, "hexlabel", label.Text("Hex:"))
	pack.Pack(hexLabel, pack.SideOpt(pack.Left), pack.PadX(5))
	c.hexEntry = entry.New(bottomFrame, "hexentry",
		entry.Width(10),
		entry.Text(initial),
		entry.ValidateOpt("key"),
		entry.ValidateCmdOpt(func(s string) bool {
			c.applyHex(s)
			return true
		}),
	)
	pack.Pack(c.hexEntry, pack.SideOpt(pack.Left), pack.PadX(5))

	// Return in the hex entry applies the colour; the dialog's default
	// Return still accepts it.
	d.App.Dispatcher().Bind(c.hexEntry.Window().PlatformID, event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == platform.XK_Return {
			c.applyHex(c.hexEntry.GetText())
		}
	})

	// Initial preview.
	c.chosen = initial
	c.updatePreview()

	return c
}

// ChooseColor displays a modal color chooser dialog.
// Returns the chosen color as "#RRGGBB" and true, or "" and false if cancelled.
func ChooseColor(parent widget.Caregiver, opts ...ColorOption) (string, bool) {
	cfg := colorConfig{
		title:        "Choose Color",
		initialColor: "#000000",
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	c := newColorChooser(parent, cfg.title, cfg.initialColor)

	// Buttons.
	addButtons(c.dialog, []dialogButton{
		{text: "OK", result: ResultOK, isDefault: true},
		{text: "Cancel", result: ResultCancel},
	})

	result := c.dialog.Run()
	if result == ResultOK {
		return c.color(), true
	}
	return "", false
}

// parseHexColor parses "#RRGGBB" to r,g,b values (0-255).
func parseHexColor(s string) (int, int, int) {
	r, g, b, _ := parseHexColorStrict(s)
	return r, g, b
}

// parseHexColorStrict parses "#RRGGBB" to r,g,b values (0-255) and reports
// whether the text is a colour.
func parseHexColorStrict(s string) (int, int, int, bool) {
	s = strings.TrimSpace(s)
	if len(s) != 7 || s[0] != '#' {
		return 0, 0, 0, false
	}
	v, err := strconv.ParseUint(s[1:], 16, 32)
	if err != nil {
		return 0, 0, 0, false
	}
	return int(v >> 16), int(v >> 8 & 0xff), int(v & 0xff), true
}
