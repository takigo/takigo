// Demo: Ttk pane with some content.
// Ported from Tk's ttkpane.tcl demo.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/dialog"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/panedwindow"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Themed Nested Panes"),
		takigo.Geometry("+300+300"),
		takigo.IconName("ttkpane"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := ttk.NewLabel(f, "msg",
		ttk.LabelWrapLength("4i"),
		ttk.LabelJustify(option.JustifyLeft),
		ttk.LabelText("This demonstration shows off a nested set of themed paned windows. Their sizes can be changed by grabbing the area between each contained pane and dragging the divider."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	msgSep := ttk.NewSeparator(f, "msgSep")
	pack.Pack(msgSep, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Outer horizontal panedwindow (matches Tcl's ttk::panedwindow -orient horizontal).
	inner := ttk.NewFrame(f, "f")
	pack.Pack(inner, pack.FillOpt(pack.FillBoth), pack.Expand(true))
	outer := ttk.NewPanedwindow(inner, "outer",
		panedwindow.OrientOpt(panedwindow.Horizontal),
	)

	// Left inner vertical panedwindow.
	inLeft := ttk.NewPanedwindow(outer, "inLeft",
		panedwindow.OrientOpt(panedwindow.Vertical),
	)
	outer.Add(inLeft.Window(), 0)

	// Right inner vertical panedwindow.
	inRight := ttk.NewPanedwindow(outer, "inRight",
		panedwindow.OrientOpt(panedwindow.Vertical),
	)
	outer.Add(inRight.Window(), 0)

	// --- Left top pane: Button ---
	topLF := ttk.NewLabelframe(inLeft, "top",
		ttk.LabelframeText("Button"),
	)
	inLeft.Add(topLF.Window(), 0)

	pressBtn := ttk.NewButton(topLF, "b",
		ttk.ButtonText("Press Me"),
		ttk.ButtonCommand(func() {
			dialog.ShowMessage(app,
				dialog.MsgTitle("Button Pressed"),
				dialog.MsgMessage("Ouch!"),
				dialog.MsgDetail("That hurt..."),
				dialog.MsgType(dialog.MsgInfo),
			)
		}),
	)
	pack.Pack(pressBtn, pack.PadX("1.5p"), pack.PadY("3p"))

	// --- Left bottom pane: Clocks ---
	botLF := ttk.NewLabelframe(inLeft, "bot",
		ttk.LabelframeText("Clocks"),
	)
	inLeft.Add(botLF.Window(), 0)

	// Timezone data matching Tcl's testzones list.
	type zoneInfo struct {
		zone string
		city string
	}
	testZones := []zoneInfo{
		{"Europe/Berlin", "Berlin"},
		{"America/Argentina/Buenos_Aires", "Buenos Aires"},
		{"Africa/Johannesburg", "Johannesburg"},
		{"Europe/London", "London"},
		{"America/Los_Angeles", "Los Angeles"},
		{"Europe/Moscow", "Moscow"},
		{"America/New_York", "New York"},
		{"Asia/Singapore", "Singapore"},
		{"Australia/Sydney", "Sydney"},
		{"Asia/Tokyo", "Tokyo"},
	}

	type clockEntry struct {
		loc     *time.Location
		timeLbl *ttk.Label
	}
	var clocks []clockEntry

	for i, z := range testZones {
		loc, err := time.LoadLocation(z.zone)
		if err != nil {
			continue
		}

		// Separator between entries (matches Tcl's ttk::separator s$i for i > 0).
		if i > 0 {
			sep := ttk.NewSeparator(botLF, fmt.Sprintf("s%d", i))
			pack.Pack(sep, pack.FillOpt(pack.FillX))
		}

		// City name label.
		cityLbl := ttk.NewLabel(botLF, fmt.Sprintf("l%d", i),
			ttk.LabelText(z.city),
		)
		pack.Pack(cityLbl, pack.FillOpt(pack.FillX))

		// Time label (updated every second).
		timeLbl := ttk.NewLabel(botLF, fmt.Sprintf("t%d", i),
			ttk.LabelText("--:--:--"),
		)
		pack.Pack(timeLbl, pack.FillOpt(pack.FillX))

		clocks = append(clocks, clockEntry{loc: loc, timeLbl: timeLbl})
	}

	// Update all clocks every second (matches Tcl's every 1000).
	var updateClocks func()
	updateClocks = func() {
		now := time.Now()
		for _, c := range clocks {
			c.timeLbl.SetText(now.In(c.loc).Format("15:04:05"))
		}
		app.After(1000*time.Millisecond, updateClocks)
	}
	app.After(0, updateClocks)

	// --- Right top pane: Progress ---
	rightTopLF := ttk.NewLabelframe(inRight, "top",
		ttk.LabelframeText("Progress"),
	)
	inRight.Add(rightTopLF.Window(), 0)

	progress := ttk.NewProgressbar(rightTopLF, "progress",
		ttk.ProgressbarMode(ttk.ProgressIndeterminate),
	)
	pack.Pack(progress, pack.FillOpt(pack.FillBoth), pack.Expand(true))
	progress.Start(50 * time.Millisecond)

	// --- Right bottom pane: Text ---
	rightBotLF := ttk.NewLabelframe(inRight, "bot",
		ttk.LabelframeText("Text"),
	)
	inRight.Add(rightBotLF.Window(), 0)

	// A TEntry-styled ttk::frame gives the classic text a themed border.
	entryFrame := ttk.NewFrame(rightBotLF, "f", ttk.FrameStyleOpt("TEntry"))
	txt := text.New(entryFrame, "txt",
		text.Width(30),
		text.WrapModeOpt(text.WrapWord),
		text.BorderWidthOpt(0),
	)

	sb := ttk.NewScrollbar(rightBotLF, "sb",
		ttk.ScrollbarOrientOpt(ttk.Vertical),
		ttk.ScrollbarCommandOpt(func(args ...any) {
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

	// Pack scrollbar right, then text fills rest (matches Tcl structure).
	pack.Pack(sb, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(txt, pack.FillOpt(pack.FillBoth), pack.Expand(true),
		pack.PadX("1.5p"), pack.PadY("1.5p"))
	pack.Pack(entryFrame, pack.FillOpt(pack.FillBoth), pack.Expand(true))

	pack.Pack(outer, pack.FillOpt(pack.FillBoth), pack.Expand(true))

	_ = pressBtn
	_ = progress
	_ = txt
	_ = sb
	app.Run()
}
