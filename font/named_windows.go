//go:build windows

package font

// Windows named font defaults matching Tk's tkWinDefault.h.
var namedFontDefs = map[string]Attributes{
	TkDefaultFont:      {Family: "sans-serif", Size: 10, Weight: WeightNormal, Slant: SlantRoman},
	TkTextFont:         {Family: "sans-serif", Size: 10, Weight: WeightNormal, Slant: SlantRoman},
	TkFixedFont:        {Family: "monospace", Size: 10, Weight: WeightNormal, Slant: SlantRoman},
	TkMenuFont:         {Family: "sans-serif", Size: 10, Weight: WeightNormal, Slant: SlantRoman},
	TkHeadingFont:      {Family: "sans-serif", Size: 12, Weight: WeightBold, Slant: SlantRoman},
	TkCaptionFont:      {Family: "sans-serif", Size: 10, Weight: WeightBold, Slant: SlantRoman},
	TkSmallCaptionFont: {Family: "sans-serif", Size: 9, Weight: WeightNormal, Slant: SlantRoman},
	TkIconFont:         {Family: "sans-serif", Size: 9, Weight: WeightNormal, Slant: SlantRoman},
	TkTooltipFont:      {Family: "sans-serif", Size: 9, Weight: WeightNormal, Slant: SlantRoman},
}
