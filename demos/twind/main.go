// Demo: Text widget with embedded windows.
// Ported from Tk's twind.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Text Demonstration - Embedded Windows and Other Features"),
		takigo.Geometry("+300+300"),
		takigo.IconName("Embedded Windows"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Frame with sunken border to hold the text widget (matches $w.f).
	tf := frame.New(f, "tf",
		frame.BorderWidth(1),
		frame.Relief(option.ReliefSunken),
	)

	tw := text.New(tf, "text",
		text.Width(70), text.Height(35),
		text.WrapModeOpt(text.WrapWord),
		text.BorderWidthOpt(0),
	)
	tw.HighlightWidth = 0
	pack.Pack(tw, pack.Expand(true), pack.FillOpt(pack.FillBoth))

	scroll := ttk.NewScrollbar(f, "scroll")
	tw.YScrollCmd = func(first, last float64) { scroll.Set(first, last) }
	scroll.Command = func(args ...any) {
		if len(args) >= 2 {
			action, _ := args[0].(string)
			number, _ := args[1].(float64)
			switch action {
			case "moveto":
				tw.YViewMoveTo(number)
			case "scroll":
				unit := "units"
				if len(args) >= 3 {
					if u, ok := args[2].(string); ok {
						unit = u
					}
				}
				tw.YViewScroll(int(number), unit == "pages")
			}
		}
	}
	pack.Pack(scroll, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(tf, pack.Expand(true), pack.FillOpt(pack.FillBoth))

	// Configure tags matching twind.tcl.
	tw.TagConfigure("center",
		text.TagJustify(option.JustifyCenter),
		text.TagSpacing1Str("5m"),
		text.TagSpacing3Str("5m"),
	)
	tw.TagConfigure("buttons",
		text.TagLMargin1Str("1c"),
		text.TagLMargin2Str("1c"),
		text.TagRMarginStr("1c"),
		text.TagSpacing1Str("3m"),
		text.TagSpacing2(0),
		text.TagSpacing3(0),
	)

	// Create Turn On / Turn Off buttons (matching $t.on / $t.off).
	onBtn := button.New(tw, "on",
		button.Text("Turn On"),
		button.Command(func() {
			// Create horizontal scrollbar and set wrap=none.
			tw.SetWrapMode(text.WrapNone)
		}),
	)
	offBtn := button.New(tw, "off",
		button.Text("Turn Off"),
		button.Command(func() {
			tw.SetWrapMode(text.WrapWord)
		}),
	)

	tw.Insert("end", "A text widget can contain many different kinds of items, ")
	tw.Insert("end", "both active and passive.  It can lay these out in various ")
	tw.Insert("end", "ways, with wrapping, tabs, centering, etc.  In addition, ")
	tw.Insert("end", "when the contents are too big for the window, smooth ")
	tw.Insert("end", "scrolling in all directions is provided.\n\n")

	tw.Insert("end", "A text widget can contain other widgets embedded ")
	tw.Insert("end", "it.  These are called \"embedded windows\", ")
	tw.Insert("end", "and they can consist of arbitrary widgets.  ")
	tw.Insert("end", "For example, here are two embedded button ")
	tw.Insert("end", "widgets.  You can click on the first button to ")
	tw.WindowCreate(tw.EndIndex(), onBtn.Window())
	tw.Insert("end", " horizontal scrolling, which also turns off ")
	tw.Insert("end", "word wrapping.  Or, you can click on the second ")
	tw.Insert("end", "button to\n")
	tw.WindowCreate(tw.EndIndex(), offBtn.Window())
	tw.Insert("end", " horizontal scrolling and turn back on word wrapping.\n\n")

	tw.Insert("end", "You may also find it useful to put embedded windows in ")
	tw.Insert("end", "a text without any actual text.  In this case the ")
	tw.Insert("end", "text widget acts like a geometry manager.  For ")
	tw.Insert("end", "example, here is a collection of buttons laid out ")
	tw.Insert("end", "neatly into rows by the text widget.  These buttons ")
	tw.Insert("end", "can be used to change the background color of the ")
	tw.Insert("end", "text widget (\"Default\" restores the color to ")
	tw.Insert("end", "its default).  If you click on the button labeled ")
	tw.Insert("end", "\"Short\", it changes to a longer string so that ")
	tw.Insert("end", "you can see how the text widget automatically ")
	tw.Insert("end", "changes the layout.  Click on the button again ")
	tw.Insert("end", "to restore the short string.\n")

	// Default button (matches $t.default).
	defaultBtnLine := tw.EndIndex()
	tw.Insert("end", "\n")
	defaultBtn := button.New(tw, "default",
		button.Text("Default"),
		button.Command(func() {
			text.Background("#ffffff")(tw)
			tw.Display()
		}),
	)
	tw.WindowCreate(defaultBtnLine, defaultBtn.Window())

	// Color buttons (matches Tcl's color list).
	colors := []string{
		"AntiqueWhite3", "Bisque1", "Bisque2", "Bisque3", "Bisque4",
		"SlateBlue3", "RoyalBlue1", "SteelBlue2", "DeepSkyBlue3", "LightBlue1",
		"DarkSlateGray1", "Aquamarine2", "DarkSeaGreen2", "SeaGreen1",
		"Yellow1", "IndianRed1", "IndianRed2", "Tan1", "Tan4",
	}
	for i, c := range colors {
		name := fmt.Sprintf("color%d", i+1)
		colorName := c
		btnLine := tw.EndIndex()
		tw.Insert("end", "\n")
		clrBtn := button.New(tw, name,
			button.Text(colorName),
			button.Command(func() {
				text.Background(colorName)(tw)
				tw.Display()
			}),
		)
		tw.WindowCreate(btnLine, clrBtn.Window())
	}

	// Tag the color buttons section.
	tagStart := defaultBtnLine
	tagEnd := tw.EndIndex()
	tw.TagAdd("buttons", tagStart, tagEnd)

	// Border, highlight, and padding buttons (matches $t.bigB etc.).
	tw.Insert("end", "\nYou can also change the usual border width and ")
	tw.Insert("end", "highlightthickness and padding.\n")

	normalBorder := tw.BorderWidth
	normalHighlight := tw.HighlightWidth

	bigBBtn := button.New(tw, "bigB",
		button.Text("Big borders"),
		button.Command(func() {
			tw.BorderWidth = 12
			tw.UpdateBorder()
			tw.Display()
		}),
	)
	tw.WindowCreate(tw.EndIndex(), bigBBtn.Window())

	smallBBtn := button.New(tw, "smallB",
		button.Text("Small borders"),
		button.Command(func() {
			tw.BorderWidth = normalBorder
			tw.UpdateBorder()
			tw.Display()
		}),
	)
	tw.WindowCreate(tw.EndIndex(), smallBBtn.Window())

	bigHBtn := button.New(tw, "bigH",
		button.Text("Big highlight"),
		button.Command(func() {
			tw.HighlightWidth = 12
			tw.Display()
		}),
	)
	tw.WindowCreate(tw.EndIndex(), bigHBtn.Window())

	smallHBtn := button.New(tw, "smallH",
		button.Text("Small highlight"),
		button.Command(func() {
			tw.HighlightWidth = normalHighlight
			tw.Display()
		}),
	)
	tw.WindowCreate(tw.EndIndex(), smallHBtn.Window())

	_ = bigBBtn
	_ = smallBBtn
	_ = bigHBtn
	_ = smallHBtn

	app.Run()
}
