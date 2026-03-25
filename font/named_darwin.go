package font

// macOS named font defaults matching Tk's tkMacOSXFont.c.
// Tk uses the system font (13pt) via CTFontCreateUIFontForLanguage.
var namedFontDefs = map[string]Attributes{
	TkDefaultFont:      {Family: "sans-serif", Size: 13, Weight: WeightNormal, Slant: SlantRoman},
	TkTextFont:         {Family: "sans-serif", Size: 13, Weight: WeightNormal, Slant: SlantRoman},
	TkFixedFont:        {Family: "monospace", Size: 11, Weight: WeightNormal, Slant: SlantRoman},
	TkMenuFont:         {Family: "sans-serif", Size: 13, Weight: WeightNormal, Slant: SlantRoman},
	TkHeadingFont:      {Family: "sans-serif", Size: 18, Weight: WeightBold, Slant: SlantRoman},
	TkCaptionFont:      {Family: "sans-serif", Size: 13, Weight: WeightBold, Slant: SlantRoman},
	TkSmallCaptionFont: {Family: "sans-serif", Size: 11, Weight: WeightNormal, Slant: SlantRoman},
	TkIconFont:         {Family: "sans-serif", Size: 13, Weight: WeightNormal, Slant: SlantRoman},
	TkTooltipFont:      {Family: "sans-serif", Size: 11, Weight: WeightNormal, Slant: SlantRoman},
}
