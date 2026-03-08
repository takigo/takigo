// Demo: Message widget with aspect-ratio-based text wrapping.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/message"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Message Widget Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("message"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Default aspect ratio (150).
	m1 := message.New(f, "m1",
		message.Text("This is a message widget. It displays multi-line text "+
			"with automatic word wrapping based on an aspect ratio. "+
			"The default aspect ratio is 150 (width = 1.5 * height)."),
		message.BorderWidth(2),
		message.Relief(option.ReliefGroove),
	)
	pack.Pack(m1, pack.SideOpt(pack.Top), pack.PadY("7p"))

	// Wide aspect ratio (300).
	m2 := message.New(f, "m2",
		message.Text("This message has a wider aspect ratio of 300, "+
			"so it prefers to be wider and shorter. "+
			"The widget adjusts its line breaking to achieve the target ratio."),
		message.Aspect(300),
		message.Foreground("#00008b"),
		message.BorderWidth(2),
		message.Relief(option.ReliefSunken),
	)
	pack.Pack(m2, pack.SideOpt(pack.Top), pack.PadY("7p"))

	// Narrow aspect ratio (75).
	m3 := message.New(f, "m3",
		message.Text("This message has a narrow aspect ratio of 75, "+
			"meaning it prefers a taller, narrower shape. "+
			"Good for sidebar-style messages."),
		message.Aspect(75),
		message.JustifyOpt(option.JustifyCenter),
		message.Foreground("#8b0000"),
		message.BorderWidth(2),
		message.Relief(option.ReliefRaised),
	)
	pack.Pack(m3, pack.SideOpt(pack.Top), pack.PadY("7p"))

	// Explicit width.
	m4 := message.New(f, "m4",
		message.Text("This message uses an explicit width of 200 pixels "+
			"instead of an aspect ratio. The text wraps at this fixed width."),
		message.WidthOpt(200),
		message.JustifyOpt(option.JustifyRight),
		message.Anchor(option.AnchorE),
	)
	pack.Pack(m4, pack.SideOpt(pack.Top), pack.PadY("7p"))

	app.Run()
}
