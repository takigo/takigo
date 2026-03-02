// Package defaulttheme registers the default TTK theme.
// Import with blank identifier to auto-register: _ "github.com/msorc/takigo/ttk/defaulttheme"
package defaulttheme

import (
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
)

func init() {
	theme := ttk.NewTheme("default", nil)

	// Register built-in elements.
	theme.RegisterElement("background", ttk.NewBackgroundElementFactory)
	theme.RegisterElement("border", ttk.NewBorderElementFactory)
	theme.RegisterElement("padding", ttk.NewPaddingElementFactory)
	theme.RegisterElement("focus", ttk.NewFocusElementFactory)
	theme.RegisterElement("separator.h", ttk.NewSeparatorElementFactory(ttk.Horizontal))
	theme.RegisterElement("separator.v", ttk.NewSeparatorElementFactory(ttk.Vertical))
	// Note: "label" is registered per-widget via NewLabelElementFactory.

	// Root style "." defaults.
	root := theme.GetStyle(".")
	root.Defaults["-background"] = uint64(0xd9d9d9)
	root.Defaults["-foreground"] = uint64(0x000000)
	root.Defaults["-borderwidth"] = 2
	root.Defaults["-focuscolor"] = uint64(0x000000)

	// TFrame style.
	tframe := theme.GetStyle("TFrame")
	tframe.Defaults["-relief"] = option.ReliefFlat
	tframe.Defaults["-borderwidth"] = 0

	// TLabel style.
	tlabel := theme.GetStyle("TLabel")
	tlabel.Defaults["-padding"] = ttk.Padding{Left: 4, Top: 2, Right: 4, Bottom: 2}
	tlabel.Defaults["-relief"] = option.ReliefFlat
	tlabel.Defaults["-borderwidth"] = 0

	// TButton style.
	tbutton := theme.GetStyle("TButton")
	tbutton.Defaults["-padding"] = ttk.Padding{Left: 8, Top: 4, Right: 8, Bottom: 4}
	tbutton.Defaults["-relief"] = option.ReliefRaised
	tbutton.Defaults["-borderwidth"] = 2

	// State maps for button relief.
	tbutton.Maps["-relief"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: option.ReliefSunken},
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: option.ReliefFlat},
	}

	// State maps for button background.
	tbutton.Maps["-background"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateHover}, Value: uint64(0xececec)},
	}

	// TSeparator styles.
	tsepH := theme.GetStyle("TSeparator.Horizontal")
	tsepH.Defaults["-relief"] = option.ReliefFlat
	tsepV := theme.GetStyle("TSeparator.Vertical")
	tsepV.Defaults["-relief"] = option.ReliefFlat

	// Layout templates.

	// TFrame: border → padding
	theme.RegisterLayout("TFrame",
		ttk.L("background", ttk.Expand|ttk.FillF,
			ttk.L("border", ttk.Expand|ttk.FillF|ttk.Border,
				ttk.L("padding", ttk.Expand|ttk.FillF))))

	// TLabel: border → padding → label
	theme.RegisterLayout("TLabel",
		ttk.L("background", ttk.Expand|ttk.FillF,
			ttk.L("border", ttk.Expand|ttk.FillF|ttk.Border,
				ttk.L("padding", ttk.Expand|ttk.FillF,
					ttk.L("label", ttk.Expand|ttk.FillF)))))

	// TButton: border → focus → padding → label
	theme.RegisterLayout("TButton",
		ttk.L("background", ttk.Expand|ttk.FillF,
			ttk.L("border", ttk.Expand|ttk.FillF|ttk.Border,
				ttk.L("focus", ttk.Expand|ttk.FillF,
					ttk.L("padding", ttk.Expand|ttk.FillF,
						ttk.L("label", ttk.Expand|ttk.FillF))))))

	// TSeparator.Horizontal: single separator element.
	theme.RegisterLayout("TSeparator.Horizontal",
		ttk.L("background", ttk.Expand|ttk.FillF,
			ttk.L("separator.h", ttk.Expand|ttk.FillF)))

	// TSeparator.Vertical: single separator element.
	theme.RegisterLayout("TSeparator.Vertical",
		ttk.L("background", ttk.Expand|ttk.FillF,
			ttk.L("separator.v", ttk.Expand|ttk.FillF)))

	ttk.RegisterTheme(theme)
}
