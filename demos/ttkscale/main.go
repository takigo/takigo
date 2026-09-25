// Demo: Horizontal scale with label feedback.
// Ported from Tk's ttkscale.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Themed Scale Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("ttkscale"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := ttk.NewFrame(app, "f")
	pack.Pack(f, pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := ttk.NewLabel(f, "msg",
		ttk.LabelWrapLength("3.5i"),
		ttk.LabelJustify(option.JustifyLeft),
		ttk.LabelText("A label tied to a horizontal scale is displayed below.  If you click or drag mouse button 1 in the scale, you can change the contents of the label; a callback command is used to couple the slider to both the text and the coloring of the label."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top), pack.PadX(".5c"))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	colorList := []string{"Red", "Orange", "Yellow", "Green", "Blue", "Violet"}

	fr := ttk.NewFrame(f, "frame", ttk.FrameBorderWidth("7.5p"))
	pack.Pack(fr, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	valueLabel := ttk.NewLabel(fr, "label")
	sc := ttk.NewScale(fr, "scale",
		ttk.ScaleFrom(0),
		ttk.ScaleTo(5),
		ttk.ScaleCommand(func(v float64) {
			c := colorList[int(v)]
			if col, err := app.ColorCache().Get(c); err == nil {
				valueLabel.SetWidgetOption("-foreground", col.Pixel)
			}
			valueLabel.SetText("Color: " + c)
		}),
	)
	sc.Set(0)
	pack.Pack(valueLabel)
	pack.Pack(sc)

	app.Run()
}
