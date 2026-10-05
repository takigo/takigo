// Demo: Show off the stock font selector dialog.
// Ported from Tk's fontchoose.tcl demo.
package main

import (
	"fmt"
	"github.com/takigo/takigo/font"
	"os"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/demos/demohelper"
	"github.com/takigo/takigo/dialog"
	"github.com/takigo/takigo/geometry/grid"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/ttk"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/frame"
	"github.com/takigo/takigo/widget/text"
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

	// Content frame (sunken border like the Tk original).
	cf := ttk.NewFrame(f, "f",
		ttk.FrameRelief(option.ReliefSunken),
		ttk.FramePadding(ttk.Padding{Left: 2, Top: 2, Right: 2, Bottom: 2}),
	)

	// Text widget + scrollbar in row 0 of cf (grid layout inside cf).
	tw := text.New(cf, "msg",
		text.FontOpt(font.TkDefaultFont),
		text.Width(40),
		text.Height(6),
		text.WrapModeOpt(text.WrapWord),
		text.BorderWidthOpt(0),
	)

	yscroll := ttk.NewScrollbar(cf, "vs",
		ttk.ScrollbarOrientOpt(ttk.Vertical),
		ttk.ScrollbarCommandOpt(widget.ScrollY(tw)),
	)
	tw.YScrollCmd = func(first, last float64) {
		yscroll.Set(first, last)
	}

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
				tw.Configure(text.FontOpt(fontDesc))
			}
		}),
	)

	grid.Grid(tw, grid.Row(0), grid.Column(0), grid.Sticky(grid.NSEW))
	grid.Grid(yscroll, grid.Row(0), grid.Column(1), grid.Sticky(grid.NSEW))
	grid.Grid(setFontBtn, grid.Row(1), grid.Column(0), grid.ColumnSpan(2), grid.Sticky(grid.StickE))
	grid.ColumnConfigure(cf, 0, grid.Weight(1))
	grid.RowConfigure(cf, 0, grid.Weight(1))

	// See Code / Dismiss buttons.
	btns := demohelper.AddSeeDismiss(f)

	// Outer grid layout matching Tcl.
	grid.Grid(cf, grid.Sticky(grid.NSEW))
	grid.Grid(btns, grid.Sticky(grid.EW))
	grid.ColumnConfigure(f, 0, grid.Weight(1))
	grid.RowConfigure(f, 0, grid.Weight(1))
	app.UpdateIdleTasks()
	grid.SetPropagate(cf, false)

	_ = yscroll
	_ = setFontBtn
	app.Run()
}
