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
	theme.RegisterElement("field", ttk.NewFieldElementFactory)
	theme.RegisterElement("Menubutton.indicator", ttk.NewMenubuttonIndicatorElementFactory)
	// Note: "label" is registered per-widget via NewLabelElementFactory.

	// Root style "." defaults.
	root := theme.GetStyle(".")
	root.Defaults["-background"] = uint64(0xd9d9d9)
	root.Defaults["-foreground"] = uint64(0x000000)
	root.Defaults["-borderwidth"] = 1
	root.Defaults["-focuscolor"] = uint64(0x000000)

	// ttk::style map "." -background [list disabled $colors(-frame) active $colors(-activebg)]
	root.Maps["-background"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: uint64(0xd9d9d9)},
		{Spec: ttk.StateSpec{OnBits: ttk.StateActive}, Value: uint64(0xececec)},
	}

	// Disabled foreground for all widgets.
	root.Maps["-foreground"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: uint64(0xa3a3a3)},
	}

	// Indicator element colors (used by Checkbutton.indicator / Radiobutton.indicator).
	// Matches Tcl's default theme SVG palette (shade/border/light/hi on 4 sides).
	root.Defaults["-shadecolor"] = uint64(0x888888)
	root.Defaults["-lightcolor"] = uint64(0xdddddd)
	root.Defaults["-bordercolor"] = uint64(0x414141)

	// TFrame style.
	tframe := theme.GetStyle("TFrame")
	tframe.Defaults["-relief"] = option.ReliefFlat

	// TCheckbutton style.
	// Matches Tcl's default theme: indicatorbackground defaults to white and
	// turns dark navy when selected/alternate, lighter blue when pressed,
	// gray when disabled.
	tcheckbutton := theme.GetStyle("TCheckbutton")
	tcheckbutton.Defaults["-padding"] = "0.75p"
	tcheckbutton.Defaults["-indicatormargin"] = "0 1.5p 3p 1.5p"
	tcheckbutton.Defaults["-indicatorbackground"] = uint64(0xffffff)
	tcheckbutton.Defaults["-indicatorforeground"] = uint64(0xffffff)
	tcheckbutton.Maps["-indicatorbackground"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateAlternate | ttk.StateDisabled}, Value: uint64(0xa3a3a3)},
		{Spec: ttk.StateSpec{OnBits: ttk.StateAlternate | ttk.StatePressed}, Value: uint64(0x5895bc)},
		{Spec: ttk.StateSpec{OnBits: ttk.StateAlternate}, Value: uint64(0x4a6984)},
		{Spec: ttk.StateSpec{OnBits: ttk.StateSelected | ttk.StateDisabled}, Value: uint64(0xa3a3a3)},
		{Spec: ttk.StateSpec{OnBits: ttk.StateSelected | ttk.StatePressed}, Value: uint64(0x5895bc)},
		{Spec: ttk.StateSpec{OnBits: ttk.StateSelected}, Value: uint64(0x4a6984)},
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: uint64(0xd9d9d9)},
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: uint64(0xc3c3c3)},
	}

	// TRadiobutton style.
	tradiobutton := theme.GetStyle("TRadiobutton")
	tradiobutton.Defaults["-padding"] = "0.75p"
	tradiobutton.Defaults["-indicatormargin"] = "0 1.5p 3p 1.5p"
	tradiobutton.Defaults["-indicatorbackground"] = uint64(0xffffff)
	tradiobutton.Defaults["-indicatorforeground"] = uint64(0xffffff)
	tradiobutton.Maps["-indicatorbackground"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateAlternate | ttk.StateDisabled}, Value: uint64(0xa3a3a3)},
		{Spec: ttk.StateSpec{OnBits: ttk.StateAlternate | ttk.StatePressed}, Value: uint64(0x5895bc)},
		{Spec: ttk.StateSpec{OnBits: ttk.StateAlternate}, Value: uint64(0x4a6984)},
		{Spec: ttk.StateSpec{OnBits: ttk.StateSelected | ttk.StateDisabled}, Value: uint64(0xa3a3a3)},
		{Spec: ttk.StateSpec{OnBits: ttk.StateSelected | ttk.StatePressed}, Value: uint64(0x5895bc)},
		{Spec: ttk.StateSpec{OnBits: ttk.StateSelected}, Value: uint64(0x4a6984)},
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: uint64(0xd9d9d9)},
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: uint64(0xc3c3c3)},
	}

	// TLabel style: no padding, inherits "." -borderwidth 1 (defaults.tcl).
	tlabel := theme.GetStyle("TLabel")
	tlabel.Defaults["-relief"] = option.ReliefFlat

	// TButton style.
	tbutton := theme.GetStyle("TButton")
	tbutton.Defaults["-anchor"] = option.AnchorCenter
	tbutton.Defaults["-padding"] = "2.25p"
	tbutton.Defaults["-width"] = -9
	tbutton.Defaults["-relief"] = option.ReliefRaised
	tbutton.Defaults["-shiftrelief"] = 1

	// Relief: sunken when pressed and not disabled (Tcl: {!disabled pressed} sunken).
	tbutton.Maps["-relief"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed, OffBits: ttk.StateDisabled}, Value: option.ReliefSunken},
	}

	// Toolbutton style: flat by default, raised on hover, sunken on press/select.
	// Used by toolbar buttons and styled menubuttons.
	// Matches Tcl: disabled flat, selected sunken, pressed sunken, active raised.
	toolbutton := theme.GetStyle("Toolbutton")
	toolbutton.Defaults["-padding"] = "1.5p"
	toolbutton.Defaults["-relief"] = option.ReliefFlat
	toolbutton.Maps["-relief"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: option.ReliefFlat},
		{Spec: ttk.StateSpec{OnBits: ttk.StateSelected}, Value: option.ReliefSunken},
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: option.ReliefSunken},
		{Spec: ttk.StateSpec{OnBits: ttk.StateHover}, Value: option.ReliefRaised},
	}
	toolbutton.Maps["-background"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: uint64(0xc3c3c3)},
		{Spec: ttk.StateSpec{OnBits: ttk.StateHover}, Value: uint64(0xececec)},
	}

	// TMenubutton.Toolbutton: same visual behavior as Toolbutton but for menubuttons.
	tmbToolbutton := theme.GetStyle("TMenubutton.Toolbutton")
	tmbToolbutton.Defaults["-padding"] = "1.5p"
	tmbToolbutton.Defaults["-relief"] = option.ReliefFlat
	tmbToolbutton.Maps["-relief"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: option.ReliefFlat},
		{Spec: ttk.StateSpec{OnBits: ttk.StateSelected}, Value: option.ReliefSunken},
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: option.ReliefSunken},
		{Spec: ttk.StateSpec{OnBits: ttk.StateHover}, Value: option.ReliefRaised},
	}
	tmbToolbutton.Maps["-background"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: uint64(0xc3c3c3)},
		{Spec: ttk.StateSpec{OnBits: ttk.StateHover}, Value: uint64(0xececec)},
	}

	// TSeparator styles.
	tsepH := theme.GetStyle("TSeparator.Horizontal")
	tsepH.Parent = theme.GetStyle("TSeparator") // takigo-only name, not a Tk style
	tsepH.Defaults["-relief"] = option.ReliefFlat
	tsepV := theme.GetStyle("TSeparator.Vertical")
	tsepV.Parent = theme.GetStyle("TSeparator") // takigo-only name, not a Tk style
	tsepV.Defaults["-relief"] = option.ReliefFlat

	// Layout templates.

	// TFrame: border → padding
	theme.RegisterLayout("TFrame",
		ttk.L("background", ttk.Expand|ttk.FillF,
			ttk.L("border", ttk.Expand|ttk.FillF|ttk.Border,
				ttk.L("padding", ttk.Expand|ttk.FillF))))

	// TLabelframe (defaults.tcl) and FrameLayout-style Labelframe.border.
	tlabelframe := theme.GetStyle("TLabelframe")
	tlabelframe.Defaults["-relief"] = option.ReliefGroove
	tlabelframe.Defaults["-borderwidth"] = 2
	theme.RegisterLayout("TLabelframe", ttk.L("border", ttk.Expand|ttk.FillF))

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

	// Toolbutton: same layout as TButton.
	theme.RegisterLayout("Toolbutton",
		ttk.L("background", ttk.Expand|ttk.FillF,
			ttk.L("border", ttk.Expand|ttk.FillF|ttk.Border,
				ttk.L("focus", ttk.Expand|ttk.FillF,
					ttk.L("padding", ttk.Expand|ttk.FillF,
						ttk.L("label", ttk.Expand|ttk.FillF))))))

	// TMenubutton.Toolbutton: same layout as TMenubutton.
	theme.RegisterLayout("TMenubutton.Toolbutton",
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

	// TMenubutton style.
	tmenubutton := theme.GetStyle("TMenubutton")
	tmenubutton.Defaults["-padding"] = "7.5p 2.25p"
	tmenubutton.Defaults["-relief"] = option.ReliefRaised
	tmenubutton.Defaults["-arrowsize"] = "3.75p"
	tmenubutton.Defaults["-arrowpadding"] = "2.25p"
	tmenubutton.Defaults["-arrowcolor"] = uint64(0x000000)
	tmenubutton.Maps["-arrowcolor"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: uint64(0xa3a3a3)},
	}

	// TMenubutton layout: border → focus → [indicator(right) + padding → label]
	theme.RegisterLayout("TMenubutton",
		ttk.L("background", ttk.Expand|ttk.FillF,
			ttk.L("border", ttk.Expand|ttk.FillF|ttk.Border,
				ttk.L("focus", ttk.Expand|ttk.FillF,
					ttk.L("Menubutton.indicator", ttk.PackRight),
					ttk.L("padding", ttk.FillXF,
						ttk.L("label", 0))))))

	// TNotebook style.
	tnotebook := theme.GetStyle("TNotebook")
	tnotebook.Defaults["-padding"] = ttk.Padding{Left: 2, Top: 2, Right: 2, Bottom: 2}
	tnotebook.Maps["-background"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateBackground}, Value: uint64(0xd0d0d0)},
		{Spec: ttk.StateSpec{OnBits: ttk.StateHover}, Value: uint64(0xececec)},
	}

	// TTreeview style.
	ttreeview := theme.GetStyle("TTreeview")
	ttreeview.Defaults["-background"] = uint64(0xd9d9d9)
	ttreeview.Defaults["-foreground"] = uint64(0x000000)
	ttreeview.Defaults["-fieldbackground"] = uint64(0xffffff)
	ttreeview.Defaults["-selectbackground"] = uint64(0x4a6984)
	ttreeview.Defaults["-selectforeground"] = uint64(0xffffff)
	ttreeview.Defaults["-bordercolor"] = uint64(0xd9d9d9)
	ttreeview.Defaults["-borderwidth"] = 1
	ttreeview.Defaults["-relief"] = option.ReliefSunken
	ttreeview.Defaults["-indent"] = 20

	// Treeview.Heading style.
	tvHeading := theme.GetStyle("Treeview.Heading")
	tvHeading.Defaults["-background"] = uint64(0xd9d9d9)
	tvHeading.Defaults["-foreground"] = uint64(0x000000)
	tvHeading.Defaults["-relief"] = option.ReliefRaised
	tvHeading.Defaults["-borderwidth"] = 1
	tvHeading.Maps["-background"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: uint64(0xc0c0c0)},
		{Spec: ttk.StateSpec{OnBits: ttk.StateHover}, Value: uint64(0xececec)},
	}

	// TScrollbar styles.
	vScrollbar := theme.GetStyle("Vertical.TScrollbar")
	vScrollbar.Defaults["-troughcolor"] = uint64(0xc3c3c3)
	vScrollbar.Defaults["-borderwidth"] = 1

	hScrollbar := theme.GetStyle("Horizontal.TScrollbar")
	hScrollbar.Defaults["-troughcolor"] = uint64(0xc3c3c3)
	hScrollbar.Defaults["-borderwidth"] = 1

	// TSpinbox style.
	tspinbox := theme.GetStyle("TSpinbox")
	tspinbox.Defaults["-background"] = uint64(0xd9d9d9)
	tspinbox.Defaults["-foreground"] = uint64(0x000000)
	tspinbox.Defaults["-fieldbackground"] = uint64(0xffffff)
	tspinbox.Defaults["-arrowsize"] = "7.5p"
	tspinbox.Defaults["-arrowcolor"] = uint64(0x000000)
	tspinbox.Defaults["-padding"] = "1.5p 0 7.5p 0"
	tspinbox.Defaults["-focuswidth"] = 1
	tspinbox.Defaults["-focuscolor"] = uint64(0x4a6984)
	tspinbox.Maps["-fieldbackground"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateReadonly}, Value: uint64(0xd9d9d9)},
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: uint64(0xd9d9d9)},
	}
	tspinbox.Maps["-arrowcolor"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: uint64(0xa3a3a3)},
	}

	// TEntry style.
	tentry := theme.GetStyle("TEntry")
	tentry.Defaults["-background"] = uint64(0xd9d9d9)
	tentry.Defaults["-foreground"] = uint64(0x000000)
	tentry.Defaults["-fieldbackground"] = uint64(0xffffff)
	tentry.Defaults["-selectbackground"] = uint64(0x4a6984)
	tentry.Defaults["-selectforeground"] = uint64(0xffffff)
	tentry.Defaults["-insertwidth"] = 1
	tentry.Defaults["-padding"] = ttk.Padding{Left: 1, Top: 1, Right: 1, Bottom: 1}
	tentry.Defaults["-insertcolor"] = uint64(0x000000)

	tentry.Defaults["-focuswidth"] = 2
	tentry.Defaults["-focuscolor"] = uint64(0x4a6984)

	// EntryLayout (ttkEntry.c).
	theme.RegisterLayout("TEntry",
		ttk.L("field", ttk.FillF|ttk.Border,
			ttk.L("padding", ttk.FillF,
				ttk.L("textarea", ttk.FillF))))

	// TSizegrip style.
	tsizegrip := theme.GetStyle("TSizegrip")
	tsizegrip.Defaults["-background"] = uint64(0xd9d9d9)

	// TCombobox (defaults.tcl).
	tcombo := theme.GetStyle("TCombobox")
	tcombo.Defaults["-arrowsize"] = "9p"
	tcombo.Defaults["-arrowcolor"] = uint64(0x000000)
	tcombo.Defaults["-fieldbackground"] = uint64(0xffffff)
	tcombo.Defaults["-padding"] = ttk.UniformPadding(1)
	tcombo.Defaults["-focuswidth"] = 1
	tcombo.Defaults["-focuscolor"] = uint64(0x4a6984)
	tcombo.Maps["-fieldbackground"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateReadonly}, Value: uint64(0xd9d9d9)},
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: uint64(0xd9d9d9)},
	}
	tcombo.Maps["-arrowcolor"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: uint64(0xa3a3a3)},
	}

	// TScale (defaults.tcl).
	tscale := theme.GetStyle("TScale")
	tscale.Defaults["-innercolor"] = uint64(0x4a6984)
	tscale.Defaults["-outercolor"] = uint64(0xffffff)
	tscale.Defaults["-bordercolor"] = uint64(0xc3c3c3)
	tscale.Defaults["-troughcolor"] = uint64(0xc3c3c3)
	tscale.Defaults["-groovewidth"] = "3p"
	tscale.Maps["-outercolor"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateHover}, Value: uint64(0xececec)},
	}

	// Progressbar styles.
	hProgress := theme.GetStyle("Horizontal.TProgressbar")
	hProgress.Defaults["-troughcolor"] = uint64(0xc3c3c3)
	hProgress.Defaults["-barcolor"] = uint64(0x4a6984)
	hProgress.Defaults["-borderwidth"] = 1

	vProgress := theme.GetStyle("Vertical.TProgressbar")
	vProgress.Defaults["-troughcolor"] = uint64(0xc3c3c3)
	vProgress.Defaults["-barcolor"] = uint64(0x4a6984)
	vProgress.Defaults["-borderwidth"] = 1

	ttk.RegisterTheme(theme)
}
