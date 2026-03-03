// Demo: Unicode text display.
// Ported from Tk's unicodeout.tcl demo.
package main

import (
	"fmt"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app := demohelper.Setup("Unicode Label Demonstration", 550, 500,
		"This is a sample of Tk's support for languages that use "+
			"non-Western character sets. However, what you will actually see "+
			"below depends largely on what character sets you have installed, "+
			"and what you see for characters that are not present varies greatly "+
			"between platforms as well.")

	// Frame to hold the two-column grid of language samples.
	f := frame.New(app, "samples")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Unicode samples matching Tk's unicodeout.tcl.
	samples := []struct {
		lang string
		text string
	}{
		{"Arabic", "\uFE94\uFEF4\uFE91\uFEAE\uFECC\uFEDF\uFE8D \uFE94\uFEE4\uFEE0\uFEDC\uFEDF\uFE8D"},
		{"Trad. Chinese", "\u4E2D\u570B\u7684\u6F22\u5B57"},
		{"Simpl. Chinese", "\u6C49\u8BED"},
		{"French", "Langue fran\u00E7aise"},
		{"Greek", "\u0395\u03BB\u03BB\u03B7\u03BD\u03B9\u03BA\u03AE \u03B3\u03BB\u03CE\u03C3\u03C3\u03B1"},
		{"Hebrew", "\u05EA\u05D9\u05E8\u05D1\u05E2 \u05D1\u05EA\u05DB"},
		{"Hindi", "\u0939\u093F\u0928\u094D\u0926\u0940 \u092D\u093E\u0937\u093E"},
		{"Icelandic", "\u00CDslenska"},
		{"Japanese", "\u65E5\u672C\u8A9E\u306E\u3072\u3089\u304C\u306A, \u6F22\u5B57\u3068\u30AB\u30BF\u30AB\u30CA"},
		{"Korean", "\uB300\uD55C\uBBFC\uAD6D\uC758 \uD55C\uAE00"},
		{"Russian", "\u0420\u0443\u0441\u0441\u043A\u0438\u0439 \u044F\u0437\u044B\u043A"},
	}

	for i, s := range samples {
		langLabel := label.New(f, fmt.Sprintf("l%d", i),
			label.Text(s.lang+":"),
			label.Anchor(option.AnchorNW),
			label.PadY(0),
		)
		sampleLabel := label.New(f, fmt.Sprintf("s%d", i),
			label.Text(s.text),
			label.Anchor(option.AnchorNW),
			label.PadY(0),
		)
		grid.Grid(langLabel, grid.Row(i), grid.Column(0),
			grid.Sticky(grid.EW), grid.PadX(5), grid.PadY(0))
		grid.Grid(sampleLabel, grid.Row(i), grid.Column(1),
			grid.Sticky(grid.EW), grid.PadX(5), grid.PadY(0))
	}

	grid.ColumnConfigure(f.Window(), 1, grid.SlotConfig{Weight: 1})

	app.Run()
}
