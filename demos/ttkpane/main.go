// Demo: TTK frame with nested paned windows.
// Ported from Tk's ttkpane.tcl demo.
package main

import (
	"fmt"
	"time"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget/labelframe"
	"github.com/msorc/takigo/widget/panedwindow"
	"github.com/msorc/takigo/widget/scrollbar"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app := demohelper.Setup("Themed Nested Panes", 600, 400,
		"This demonstration shows off a nested set of themed paned windows. Their sizes can be changed by grabbing the area between each contained pane and dragging the divider.")

	ttk.SetCurrentTheme("clam")

	// Outer horizontal panedwindow.
	outer := panedwindow.New(app, "outer",
		panedwindow.OrientOpt(panedwindow.Horizontal),
	)
	pack.Pack(outer, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Left inner vertical panedwindow.
	inLeft := panedwindow.New(outer, "inleft",
		panedwindow.OrientOpt(panedwindow.Vertical),
	)
	outer.Add(inLeft.Window(), 200)

	// Right inner vertical panedwindow.
	inRight := panedwindow.New(outer, "inright",
		panedwindow.OrientOpt(panedwindow.Vertical),
	)
	outer.Add(inRight.Window(), 200)

	// --- Left top pane: Button ---
	buttonLF := labelframe.New(inLeft, "buttonlf",
		labelframe.Text("Button"),
	)
	inLeft.Add(buttonLF.Window(), 80)

	pressBtn := ttk.NewButton(buttonLF, "pressbtn",
		ttk.ButtonText("Press Me"),
		ttk.ButtonCommand(func() {
			fmt.Println("Ouch! That hurt...")
		}),
	)
	pack.Pack(pressBtn, pack.PadX(5), pack.PadY(5))

	// --- Left bottom pane: Clocks ---
	clocksLF := labelframe.New(inLeft, "clockslf",
		labelframe.Text("Clocks"),
	)
	inLeft.Add(clocksLF.Window(), 120)

	// Show city labels with static timezone names (no live clock since
	// we don't have a timer-label widget, but matching the Tk structure).
	cities := []string{
		"Berlin", "Buenos Aires", "Johannesburg", "London",
		"Los Angeles", "Moscow", "New York", "Singapore",
		"Sydney", "Tokyo",
	}
	for i, city := range cities {
		name := fmt.Sprintf("city%d", i)
		lbl := ttk.NewLabel(clocksLF, name,
			ttk.LabelText(city),
		)
		pack.Pack(lbl, pack.FillOpt(pack.FillX))
		_ = lbl
	}

	// --- Right top pane: Progress ---
	progressLF := labelframe.New(inRight, "progresslf",
		labelframe.Text("Progress"),
	)
	inRight.Add(progressLF.Window(), 80)

	progress := ttk.NewProgressbar(progressLF, "progress",
		ttk.ProgressbarMode(ttk.ProgressIndeterminate),
	)
	pack.Pack(progress, pack.FillOpt(pack.FillBoth), pack.Expand(true),
		pack.PadX(5), pack.PadY(5))
	progress.Start(50 * time.Millisecond)

	// --- Right bottom pane: Text ---
	textLF := labelframe.New(inRight, "textlf",
		labelframe.Text("Text"),
	)
	inRight.Add(textLF.Window(), 120)

	txt := text.New(textLF, "txt",
		text.Width(30),
		text.WrapModeOpt(text.WrapWord),
		text.Background("white"),
		text.BorderWidthOpt(0),
	)

	sb := scrollbar.New(textLF, "sb",
		scrollbar.OrientOpt(scrollbar.Vertical),
		scrollbar.WidthOpt(14),
		scrollbar.CommandOpt(func(args ...interface{}) {
			if len(args) < 1 {
				return
			}
			switch args[0] {
			case "moveto":
				if len(args) >= 2 {
					if f, ok := args[1].(float64); ok {
						txt.YViewMoveTo(f)
					}
				}
			case "scroll":
				if len(args) >= 3 {
					n, _ := args[1].(int)
					unit, _ := args[2].(string)
					txt.YViewScroll(n, unit == "pages")
				}
			}
		}),
	)
	txt.YScrollCmd = func(first, last float64) {
		sb.Set(first, last)
	}

	pack.Pack(sb, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(txt, pack.FillOpt(pack.FillBoth), pack.Expand(true))

	txt.Insert("1.0", `This is a text widget embedded in a themed paned window. You can edit this text, and resize the panes by dragging the sash between them.

The Tk ttkpane demo shows nested panedwindows with four panes:
  - Button: a simple press-me button
  - Clocks: timezone labels
  - Progress: an indeterminate progressbar
  - Text: this scrollable text widget

Try dragging the dividers to resize each section.`)

	_ = pressBtn
	_ = progress
	_ = txt
	_ = sb
	app.Run()
}
