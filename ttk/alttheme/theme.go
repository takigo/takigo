// Package alttheme registers the "alt" (alternate) TTK theme.
// Import with blank identifier to auto-register: _ "github.com/msorc/takigo/ttk/alttheme"
package alttheme

import (
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/defaulttheme" // ensure default theme init runs first
)

// Alt theme colors (from Tk's altTheme.tcl).
const (
	frameColor   uint64 = 0xd9d9d9
	windowColor  uint64 = 0xffffff
	darkerColor  uint64 = 0xc3c3c3
	borderColor  uint64 = 0x414141
	activeBg     uint64 = 0xececec
	disabledFg   uint64 = 0xa3a3a3
	selectBg     uint64 = 0x4a6984
	selectFg     uint64 = 0xffffff
	altIndicator uint64 = 0xaaaaaa
)

func init() {
	defaultTheme := ttk.CurrentTheme()
	theme := ttk.NewTheme("alt", defaultTheme)

	// Root style ".".
	root := theme.GetStyle(".")
	root.Defaults["-background"] = frameColor
	root.Defaults["-foreground"] = uint64(0x000000)
	root.Defaults["-troughcolor"] = darkerColor
	root.Defaults["-bordercolor"] = borderColor
	root.Defaults["-selectbackground"] = selectBg
	root.Defaults["-selectforeground"] = selectFg

	root.Maps["-background"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: frameColor},
		{Spec: ttk.StateSpec{OnBits: ttk.StateActive}, Value: activeBg},
	}
	root.Maps["-foreground"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: disabledFg},
	}

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
	tbutton.Defaults["-anchor"] = option.AnchorCenter
	tbutton.Defaults["-padding"] = ttk.Padding{Left: 8, Top: 4, Right: 8, Bottom: 4}
	tbutton.Defaults["-relief"] = option.ReliefRaised
	tbutton.Defaults["-borderwidth"] = 2
	tbutton.Maps["-relief"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed, OffBits: ttk.StateDisabled}, Value: option.ReliefSunken},
		{Spec: ttk.StateSpec{OnBits: ttk.StateActive, OffBits: ttk.StateDisabled}, Value: option.ReliefRaised},
	}

	// Toolbutton style.
	toolbutton := theme.GetStyle("Toolbutton")
	toolbutton.Defaults["-padding"] = ttk.Padding{Left: 8, Top: 4, Right: 8, Bottom: 4}
	toolbutton.Defaults["-relief"] = option.ReliefFlat
	toolbutton.Defaults["-borderwidth"] = 2
	toolbutton.Maps["-relief"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: option.ReliefFlat},
		{Spec: ttk.StateSpec{OnBits: ttk.StateSelected}, Value: option.ReliefSunken},
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: option.ReliefSunken},
		{Spec: ttk.StateSpec{OnBits: ttk.StateActive}, Value: option.ReliefRaised},
	}
	toolbutton.Maps["-background"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: darkerColor},
		{Spec: ttk.StateSpec{OnBits: ttk.StateActive}, Value: activeBg},
	}

	// TMenubutton style.
	tmenubutton := theme.GetStyle("TMenubutton")
	tmenubutton.Defaults["-padding"] = ttk.Padding{Left: 8, Top: 4, Right: 8, Bottom: 4}
	tmenubutton.Defaults["-relief"] = option.ReliefRaised
	tmenubutton.Defaults["-borderwidth"] = 2

	// TCheckbutton style (overrides default theme).
	// In alt theme, indicators are white by default, with state maps
	// changing alternate/disabled/pressed to grays.
	tcheckbutton := theme.GetStyle("TCheckbutton")
	tcheckbutton.Defaults["-padding"] = "1.5p"
	tcheckbutton.Defaults["-indicatormargin"] = "0 1.5p 3p 1.5p"
	tcheckbutton.Defaults["-indicatorbackground"] = windowColor
	tcheckbutton.Defaults["-indicatorforeground"] = windowColor
	tcheckbutton.Maps["-indicatorbackground"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: frameColor},
		{Spec: ttk.StateSpec{OnBits: ttk.StateAlternate}, Value: altIndicator},
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: frameColor},
		{Spec: ttk.StateSpec{OnBits: ttk.StateSelected}, Value: selectBg},
	}

	// TRadiobutton style (overrides default theme).
	tradiobutton := theme.GetStyle("TRadiobutton")
	tradiobutton.Defaults["-padding"] = "1.5p"
	tradiobutton.Defaults["-indicatormargin"] = "0 1.5p 3p 1.5p"
	tradiobutton.Defaults["-indicatorbackground"] = windowColor
	tradiobutton.Defaults["-indicatorforeground"] = windowColor
	tradiobutton.Maps["-indicatorbackground"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: frameColor},
		{Spec: ttk.StateSpec{OnBits: ttk.StateAlternate}, Value: altIndicator},
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: frameColor},
		{Spec: ttk.StateSpec{OnBits: ttk.StateSelected}, Value: selectBg},
	}

	// TSeparator styles.
	tsepH := theme.GetStyle("TSeparator.Horizontal")
	tsepH.Defaults["-relief"] = option.ReliefFlat
	tsepV := theme.GetStyle("TSeparator.Vertical")
	tsepV.Defaults["-relief"] = option.ReliefFlat

	// TScrollbar style.
	tscrollbar := theme.GetStyle("Vertical.TScrollbar")
	tscrollbar.Defaults["-troughcolor"] = darkerColor
	hscrollbar := theme.GetStyle("Horizontal.TScrollbar")
	hscrollbar.Defaults["-troughcolor"] = darkerColor

	// TTreeview style.
	ttreeview := theme.GetStyle("TTreeview")
	ttreeview.Defaults["-fieldbackground"] = windowColor
	ttreeview.Defaults["-selectbackground"] = selectBg
	ttreeview.Defaults["-selectforeground"] = selectFg
	ttreeview.Maps["-background"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: frameColor},
		{Spec: ttk.StateSpec{OnBits: ttk.StateSelected}, Value: selectBg},
	}
	ttreeview.Maps["-foreground"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: disabledFg},
		{Spec: ttk.StateSpec{OnBits: ttk.StateSelected}, Value: selectFg},
	}

	// Treeview.Heading style.
	tvHeading := theme.GetStyle("Treeview.Heading")
	tvHeading.Defaults["-relief"] = option.ReliefRaised

	// Progressbar styles.
	hProgress := theme.GetStyle("Horizontal.TProgressbar")
	hProgress.Defaults["-barcolor"] = selectBg
	vProgress := theme.GetStyle("Vertical.TProgressbar")
	vProgress.Defaults["-barcolor"] = selectBg

	// TEntry style.
	tentry := theme.GetStyle("TEntry")
	tentry.Defaults["-background"] = frameColor
	tentry.Defaults["-foreground"] = uint64(0x000000)
	tentry.Defaults["-fieldbackground"] = windowColor
	tentry.Defaults["-selectbackground"] = selectBg
	tentry.Defaults["-selectforeground"] = selectFg
	tentry.Defaults["-insertwidth"] = 1
	tentry.Defaults["-padding"] = ttk.Padding{Left: 1, Top: 1, Right: 1, Bottom: 1}
	tentry.Defaults["-insertcolor"] = uint64(0x000000)

	// TEntry layout: background → highlight → border → padding
	theme.RegisterLayout("TEntry",
		ttk.L("background", ttk.Expand|ttk.FillF,
			ttk.L("highlight", ttk.Expand|ttk.FillF,
				ttk.L("border", ttk.Expand|ttk.FillF|ttk.Border,
					ttk.L("padding", ttk.Expand|ttk.FillF)))))

	ttk.RegisterTheme(theme)
}
