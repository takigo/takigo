// Demo: Unicode text display.
// Ported from Tk's unicodeout.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Unicode Label Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("unicodeout"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength(screenunit.In(4)),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("This is a sample of Tk's support for languages that use non-Western character sets.  However, what you will actually see below depends largely on what character sets you have installed, and what you see for characters that are not present varies greatly between platforms as well."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Frame to hold the two-column grid of language samples.
	samples_f := frame.New(f, "samples")
	pack.Pack(samples_f, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(screenunit.Mm(2)), pack.PadY(screenunit.Mm(1)))

	// Unicode samples matching Tk's unicodeout.tcl.
	samples := []struct {
		lang string
		text string
	}{
		{"Arabic", "ﺔﻴﺑﺮﻌﻟﺍ ﺔﻤﻠﻜﻟﺍ"},
		{"Trad. Chinese", "中國的漢字"},
		{"Simpl. Chinese", "汉语"},
		{"French", "Langue française"},
		{"Greek", "Ελληνική γλώσσα"},
		{"Hebrew", "תירבע בתכ"}, // X11 presentation order, as usePresentationFormsFor
		{"Hindi", "हिन्दी भाषा"},
		{"Icelandic", "Íslenska"},
		{"Japanese", "日本語のひらがな, 漢字とカタカナ"},
		{"Korean", "대한민국의 한글"},
		{"Russian", "Русский язык"},
		// Emoji sample — shown on X11+XFT (which takigo uses).
		{"Emoji", "😀💩👍🇳🇱"},
	}

	for i, s := range samples {
		langLabel := label.New(samples_f, fmt.Sprintf("l%d", i+1),
			label.Text(s.lang+":"),
			label.Anchor(option.AnchorNW),
			label.PadY(0),
		)
		sampleLabel := label.New(samples_f, fmt.Sprintf("s%d", i+1),
			label.Text(s.text),
			label.Anchor(option.AnchorNW),
			label.PadY(0),
			label.Width(30),
		)
		// padx "1m" only on language label (column 0), not on sample label.
		grid.Grid(langLabel, grid.Row(i), grid.Column(0),
			grid.Sticky(grid.EW), grid.PadX(screenunit.Mm(1)), grid.PadY(0))
		grid.Grid(sampleLabel, grid.Row(i), grid.Column(1),
			grid.Sticky(grid.EW), grid.PadY(0))
	}

	grid.ColumnConfigure(samples_f, 1, grid.Weight(1))

	app.Run()
}
