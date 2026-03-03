// Demo: Font chooser dialog.
// Ported from Tk's fontchoose.tcl demo.
package main

import (
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/dialog"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/scrollbar"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app := demohelper.Setup("Font Selection Dialog", 450, 300,
		"Press the button below to choose a new font for the text shown in this window.")

	// Content frame (sunken border like the Tk original).
	contentFrame := frame.New(app, "content",
		frame.BorderWidth(2),
		frame.Relief(1), // sunken
	)
	pack.Pack(contentFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Text widget with scrollbar showing sample text.
	tw := text.New(contentFrame, "msg",
		text.Width(40),
		text.Height(6),
		text.WrapModeOpt(text.WrapWord),
	)

	yscroll := scrollbar.New(contentFrame, "vs",
		scrollbar.OrientOpt(scrollbar.Vertical),
		scrollbar.CommandOpt(func(args ...any) {
			if len(args) < 1 {
				return
			}
			switch args[0] {
			case "moveto":
				if len(args) >= 2 {
					if f, ok := args[1].(float64); ok {
						tw.YViewMoveTo(f)
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

	pack.Pack(yscroll, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(tw, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	tw.Insert("end", "Press the button below to choose a new font for the "+
		"text shown in this window.\n")

	// Current font descriptor for passing back into the dialog.
	currentFontDesc := ""

	// "Set font ..." button.
	setFontBtn := button.New(app, "font",
		button.Text("Set font ..."),
		button.Command(func() {
			opts := []dialog.FontOption{
				dialog.FontTitle("Font Selection"),
			}
			if currentFontDesc != "" {
				opts = append(opts, dialog.FontInitial(currentFontDesc))
			}
			fontDesc, ok := dialog.ChooseFont(app, opts...)
			if ok {
				currentFontDesc = fontDesc
				f, err := app.FontRegistry().Get(fontDesc)
				if err == nil {
					tw.Font = f
					tw.Display()
				}
			}
		}),
		button.PadX(15), button.PadY(8),
	)
	pack.Pack(setFontBtn, pack.SideOpt(pack.Top), pack.PadX(10), pack.PadY(10))

	_ = yscroll
	_ = setFontBtn
	app.Run()
}
