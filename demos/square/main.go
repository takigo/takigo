// Demo: Square widget with drag-to-move interaction.
// Demonstrates the custom square widget from tkSquare.c.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/square"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Square Widget Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("square"),
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
		label.Text("This window contains a square widget. You can drag the "+
			"colored square around with the mouse (button 1). "+
			"The square stays within the widget bounds."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Default square.
	s1 := square.New(f, "s1",
		square.SizeOpt(40),
		square.PosXOpt(50),
		square.PosYOpt(50),
		square.BorderWidthOpt(3),
		square.Relief(option.ReliefRaised),
	)
	pack.Pack(s1, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX("7p"), pack.PadY("7p"))

	// Second square with different colors.
	s2 := square.New(f, "s2",
		square.SizeOpt(30),
		square.PosXOpt(20),
		square.PosYOpt(20),
		square.Foreground("#006400"),
		square.Background("#f0f0f0"),
		square.BorderWidthOpt(2),
		square.Relief(option.ReliefSunken),
	)
	pack.Pack(s2, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX("7p"), pack.PadY("7p"))

	// Bind drag interaction for both squares.
	for _, sq := range []*square.Square{s1, s2} {
		bindDrag(app, sq)
	}

	app.Run()
}

func bindDrag(app *takigo.App, s *square.Square) {
	var dragOffX, dragOffY int

	app.Dispatcher().Bind(s.Window().PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button == 1 {
			dragOffX = ev.X - s.PosX
			dragOffY = ev.Y - s.PosY
		}
	})

	app.Dispatcher().Bind(s.Window().PlatformID, event.MotionMask, func(ev *event.Event) {
		s.SetPosition(ev.X-dragOffX, ev.Y-dragOffY)
	})
}
