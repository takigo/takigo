package dialog

import (
	"fmt"

	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/entry"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/scale"
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

	app := parent.AppContext()

	// Parse initial color.
	r, g, b := parseHexColor(cfg.initialColor)

	d := New(parent, cfg.title, 350, 280)

	var chosenColor string
	var rScale, gScale, bScale *scale.Scale
	var hexEntry *entry.Entry
	var previewFrame *frame.Frame

	updatePreview := func() {
		rv := int(rScale.Get())
		gv := int(gScale.Get())
		bv := int(bScale.Get())
		hex := fmt.Sprintf("#%02x%02x%02x", rv, gv, bv)
		chosenColor = hex

		// Update preview frame background.
		if col, err := app.ColorCache().Get(hex); err == nil {
			previewFrame.Background = col
			previewFrame.UpdateBorder()
			previewFrame.Window().BackgroundPixel = col.Pixel
			previewFrame.Display()
		}

		// Update hex entry.
		hexEntry.SetText(hex)
	}

	// Sliders frame.
	slidersFrame := newFrame(d.Content, "sliders")
	pack.Pack(slidersFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	// R slider.
	rFrame := newFrame(slidersFrame, "rframe")
	pack.Pack(rFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))
	rLabel := label.New(rFrame, "rlabel", label.Text("R:"))
	pack.Pack(rLabel, pack.SideOpt(pack.Left), pack.PadX(5))
	rScale = scale.New(rFrame, "rscale",
		scale.OrientOpt(scale.Horizontal),
		scale.FromOpt(0), scale.ToOpt(255),
		scale.ValueOpt(float64(r)),
		scale.ResolutionOpt(1),
		scale.ShowValueOpt(true),
		scale.CommandOpt(func(v float64) { updatePreview() }),
	)
	pack.Pack(rScale, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillX), pack.Expand(true))

	// G slider.
	gFrame := newFrame(slidersFrame, "gframe")
	pack.Pack(gFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))
	gLabel := label.New(gFrame, "glabel", label.Text("G:"))
	pack.Pack(gLabel, pack.SideOpt(pack.Left), pack.PadX(5))
	gScale = scale.New(gFrame, "gscale",
		scale.OrientOpt(scale.Horizontal),
		scale.FromOpt(0), scale.ToOpt(255),
		scale.ValueOpt(float64(g)),
		scale.ResolutionOpt(1),
		scale.ShowValueOpt(true),
		scale.CommandOpt(func(v float64) { updatePreview() }),
	)
	pack.Pack(gScale, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillX), pack.Expand(true))

	// B slider.
	bFrame := newFrame(slidersFrame, "bframe")
	pack.Pack(bFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))
	bLabel := label.New(bFrame, "blabel", label.Text("B:"))
	pack.Pack(bLabel, pack.SideOpt(pack.Left), pack.PadX(5))
	bScale = scale.New(bFrame, "bscale",
		scale.OrientOpt(scale.Horizontal),
		scale.FromOpt(0), scale.ToOpt(255),
		scale.ValueOpt(float64(b)),
		scale.ResolutionOpt(1),
		scale.ShowValueOpt(true),
		scale.CommandOpt(func(v float64) { updatePreview() }),
	)
	pack.Pack(bScale, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillX), pack.Expand(true))

	// Preview + hex entry frame.
	bottomFrame := newFrame(d.Content, "bottom")
	pack.Pack(bottomFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	// Color preview.
	previewFrame = frame.New(bottomFrame, "preview",
		frame.Width(60), frame.Height(40),
		frame.BorderWidth(2), frame.Relief(1), // ReliefSunken
	)
	pack.Pack(previewFrame, pack.SideOpt(pack.Left), pack.PadX(10))

	// Hex entry.
	hexLabel := label.New(bottomFrame, "hexlabel", label.Text("Hex:"))
	pack.Pack(hexLabel, pack.SideOpt(pack.Left), pack.PadX(5))
	hexEntry = entry.New(bottomFrame, "hexentry",
		entry.Width(10),
		entry.Text(cfg.initialColor),
	)
	pack.Pack(hexEntry, pack.SideOpt(pack.Left), pack.PadX(5))

	// Suppress unused variable warnings.
	_ = rLabel
	_ = gLabel
	_ = bLabel
	_ = hexLabel

	// Initial preview.
	chosenColor = cfg.initialColor
	updatePreview()

	// Buttons.
	addButtons(d, []dialogButton{
		{text: "OK", result: ResultOK, isDefault: true},
		{text: "Cancel", result: ResultCancel},
	})

	result := d.Run()
	if result == ResultOK {
		return chosenColor, true
	}
	return "", false
}

// parseHexColor parses "#RRGGBB" to r,g,b values (0-255).
func parseHexColor(s string) (int, int, int) {
	if len(s) == 7 && s[0] == '#' {
		var r, g, b int
		fmt.Sscanf(s[1:], "%02x%02x%02x", &r, &g, &b)
		return r, g, b
	}
	return 0, 0, 0
}
