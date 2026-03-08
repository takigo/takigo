// Demo: Font chooser dialog.
// Ported from Tk's fontchoose.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/dialog"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Font Selection Dialog"),
		takigo.Geometry("+300+300"),
		takigo.IconName("fontchooser"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength("4i"),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("Press the button below to choose a new font for the text shown in this window."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Content frame (sunken border like the Tk original), packed into f.
	cf := ttk.NewFrame(f, "cf",
		ttk.FrameRelief(option.ReliefSunken),
		ttk.FramePadding(ttk.Padding{Top: 2, Right: 2, Bottom: 2, Left: 2}),
	)
	pack.Pack(cf, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true),
		pack.PadX(10), pack.PadY(5))

	// Text widget + scrollbar in row 0 of cf (grid layout inside cf).
	tw := text.New(cf, "msg",
		text.Width(40),
		text.Height(6),
		text.WrapModeOpt(text.WrapWord),
		text.BorderWidthOpt(0),
	)

	yscroll := ttk.NewScrollbar(cf, "vs",
		ttk.ScrollbarOrientOpt(ttk.Vertical),
		ttk.ScrollbarCommandOpt(func(args ...any) {
			if len(args) < 1 {
				return
			}
			switch args[0] {
			case "moveto":
				if len(args) >= 2 {
					if fv, ok := args[1].(float64); ok {
						tw.YViewMoveTo(fv)
					}
				}
			case "scroll":
				if len(args) >= 3 {
					n, _ := args[1].(int)
					unit, _ := args[2].(string)
					tw.YViewScroll(n, unit == "pages")
				}
			}
		}),
	)
	tw.YScrollCmd = func(first, last float64) {
		yscroll.Set(first, last)
	}

	grid.Grid(tw, grid.Row(0), grid.Column(0), grid.Sticky(grid.NSEW))
	grid.Grid(yscroll, grid.Row(0), grid.Column(1), grid.Sticky(grid.NS))
	grid.ColumnConfigure(cf, 0, grid.Weight(1))
	grid.RowConfigure(cf, 0, grid.Weight(1))

	tw.Insert("end", "Press the buttons below to choose a new font for the "+
		"text shown in this window.\n")

	// Current font descriptor for passing back into the dialog.
	currentFontDesc := ""

	// "Set font ..." button — row 1, spans both columns, sticky east.
	setFontBtn := ttk.NewButton(cf, "font",
		ttk.ButtonText("Set font ..."),
		ttk.ButtonCommand(func() {
			opts := []dialog.FontOption{
				dialog.FontTitle("Font Selection"),
			}
			if currentFontDesc != "" {
				opts = append(opts, dialog.FontInitial(currentFontDesc))
			}
			fontDesc, ok := dialog.ChooseFont(app, opts...)
			if ok {
				currentFontDesc = fontDesc
				fnt, err := app.FontRegistry().Get(fontDesc)
				if err == nil {
					tw.Font = fnt
					tw.Display()
				}
			}
		}),
	)
	grid.Grid(setFontBtn, grid.Row(1), grid.Column(0), grid.ColumnSpan(2), grid.Sticky(grid.StickE))

	_ = yscroll
	_ = setFontBtn
	app.Run()
}
