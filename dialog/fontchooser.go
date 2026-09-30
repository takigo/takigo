package dialog

import (
	"strconv"

	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/checkbutton"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/listbox"
)

// fontConfig holds ChooseFont options.
type fontConfig struct {
	title       string
	initialFont string // font descriptor
}

// FontOption configures ChooseFont.
type FontOption func(*fontConfig)

func FontTitle(s string) FontOption   { return func(c *fontConfig) { c.title = s } }
func FontInitial(s string) FontOption { return func(c *fontConfig) { c.initialFont = s } }

// ChooseFont displays a modal font chooser dialog.
// Returns a font descriptor string and true, or "" and false if cancelled.
func ChooseFont(parent widget.Caregiver, opts ...FontOption) (string, bool) {
	cfg := fontConfig{
		title: "Choose Font",
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	d := New(parent, cfg.title, 450, 350)

	// Get available font families.
	families := parent.AppContext().FontRegistry().Families()
	if len(families) == 0 {
		families = []string{"sans-serif", "serif", "monospace"}
	}

	// Size options.
	sizes := []string{"8", "9", "10", "11", "12", "14", "16", "18", "20", "24", "28", "32", "36", "48", "72"}

	selectedFamily := "sans-serif"
	selectedSize := "12"
	selectedBold := false
	selectedItalic := false

	// Parse initial font if provided.
	if cfg.initialFont != "" {
		if attrs, err := font.ParseDescriptor(cfg.initialFont); err == nil {
			if attrs.Family != "" {
				selectedFamily = attrs.Family
			}
			if attrs.Size > 0 {
				selectedSize = strconv.Itoa(int(attrs.Size))
			}
			selectedBold = attrs.Weight == font.WeightBold
			selectedItalic = attrs.Slant == font.SlantItalic
		}
	}

	// Top area: family list + size list.
	listsFrame := newFrame(d.Content, "lists")
	pack.Pack(listsFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Family listbox with label.
	familyFrame := newFrame(listsFrame, "famframe")
	pack.Pack(familyFrame, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true), pack.PadX(5))

	familyLabel := label.New(familyFrame, "famlabel", label.Text("Family:"))
	pack.Pack(familyLabel, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	familyList := listbox.New(familyFrame, "famlist",
		listbox.Items(families...),
		listbox.Width(25),
		listbox.Height(10),
	)
	pack.Pack(familyList, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Size listbox with label.
	sizeFrame := newFrame(listsFrame, "sizeframe")
	pack.Pack(sizeFrame, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillY), pack.PadX(5))

	sizeLabel := label.New(sizeFrame, "sizelabel", label.Text("Size:"))
	pack.Pack(sizeLabel, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	sizeList := listbox.New(sizeFrame, "sizelist",
		listbox.Items(sizes...),
		listbox.Width(6),
		listbox.Height(10),
	)
	pack.Pack(sizeList, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Style toggles, the checkbuttons of library/fontchooser.tcl.
	styleFrame := newFrame(d.Content, "styleframe")
	pack.Pack(styleFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	boolVar := func(on bool) *widget.Variable[string] {
		if on {
			return widget.NewVariable("1")
		}
		return widget.NewVariable("0")
	}
	boldVar, italicVar := boolVar(selectedBold), boolVar(selectedItalic)
	var updatePreview func()
	boldCheck := checkbutton.New(styleFrame, "bold", checkbutton.Text("Bold"),
		checkbutton.Var(boldVar), checkbutton.Command(func() { updatePreview() }))
	pack.Pack(boldCheck, pack.SideOpt(pack.Left), pack.PadX(5))
	italicCheck := checkbutton.New(styleFrame, "italic", checkbutton.Text("Italic"),
		checkbutton.Var(italicVar), checkbutton.Command(func() { updatePreview() }))
	pack.Pack(italicCheck, pack.SideOpt(pack.Left), pack.PadX(5))

	// Preview label.
	previewFrame := newFrame(d.Content, "previewframe",
		frame.BorderWidth(1), frame.Relief(1), // sunken
	)
	pack.Pack(previewFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadX(5), pack.PadY(5))

	previewLabel := label.New(previewFrame, "preview",
		label.Text("AaBbCcDd 123"),
		label.PadX(5), label.PadY(10),
	)
	pack.Pack(previewLabel, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// Select initial family in list.
	for i, f := range families {
		if f == selectedFamily {
			familyList.SelectionSet(i, i)
			familyList.See(i)
			break
		}
	}

	// Select initial size.
	for i, s := range sizes {
		if s == selectedSize {
			sizeList.SelectionSet(i, i)
			sizeList.See(i)
			break
		}
	}

	// descriptor builds the font from the current selections.
	descriptor := func() string {
		if sel := familyList.Selection(); len(sel) > 0 && sel[0] < len(families) {
			selectedFamily = families[sel[0]]
		}
		if sel := sizeList.Selection(); len(sel) > 0 && sel[0] < len(sizes) {
			selectedSize = sizes[sel[0]]
		}
		desc := "{" + selectedFamily + "} " + selectedSize
		if boldVar.Get() == "1" {
			desc += " bold"
		}
		if italicVar.Get() == "1" {
			desc += " italic"
		}
		return desc
	}
	// The sample shows the chosen font, as the chooser's preview does.
	updatePreview = func() { previewLabel.Configure(label.FontOpt(descriptor())) }
	familyList.SelectCmd = updatePreview
	sizeList.SelectCmd = updatePreview
	updatePreview()

	// Buttons.
	addButtons(d, []dialogButton{
		{text: "OK", result: ResultOK, isDefault: true},
		{text: "Cancel", result: ResultCancel},
	})

	result := d.Run()
	if result == ResultOK {
		return descriptor(), true
	}
	return "", false
}
