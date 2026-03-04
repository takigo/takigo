// Demo: Canvas and text display with print buttons.
// Ported from Tk's print.tcl demo (tk print not available in Go).
package main

import (
	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/dialog"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app := demohelper.Setup("Printing Demonstration", 700, 500,
		"This demonstration showcases the print command. Clicking the buttons below prints the data from the canvas and text widgets using platform-native dialogs.")

	// Button frame at the bottom.
	btnFrame := frame.New(app, "f")
	pack.Pack(btnFrame, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	printCanvasBtn := button.New(btnFrame, "c",
		button.Text("Print Canvas"),
		button.Command(func() {
			dialog.ShowMessage(app,
				dialog.MsgTitle("Print"),
				dialog.MsgMessage("Printing is not available on this platform."),
				dialog.MsgType(dialog.MsgInfo),
			)
		}),
	)
	pack.Pack(printCanvasBtn, pack.SideOpt(pack.Left), pack.Anchor(7), // AnchorW
		pack.PadX("3p"))

	printTextBtn := button.New(btnFrame, "t",
		button.Text("Print Text"),
		button.Command(func() {
			dialog.ShowMessage(app,
				dialog.MsgTitle("Print"),
				dialog.MsgMessage("Printing is not available on this platform."),
				dialog.MsgType(dialog.MsgInfo),
			)
		}),
	)
	pack.Pack(printTextBtn, pack.SideOpt(pack.Right), pack.Anchor(3), // AnchorE
		pack.PadX("3p"))

	// Content area: canvas left, text right.
	m := frame.New(app, "m")
	pack.Pack(m, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true))

	// Canvas with shapes.
	c := canvas.New(m, "c", canvas.Background("white"))
	pack.Pack(c, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth),
		pack.Expand(true))

	c.CreateRectangle(20, 20, 220, 80, canvas.FillColor("blue"), canvas.OutlineColor("black"))
	c.CreateOval(20, 100, 220, 160, canvas.FillColor("green"), canvas.OutlineColor("black"))
	c.CreateText(20, 180, canvas.TextOpt("A short demo of simple canvas elements."),
		canvas.TextColor("black"))

	// Text widget with Tcl/Tk description.
	tw := text.New(m, "t", text.WrapModeOpt(text.WrapWord))
	pack.Pack(tw, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillBoth),
		pack.Expand(true))

	tw.Insert("1.0", "Tcl, or Tool Command Language, is an open-source multi-purpose C library which includes a powerful dynamic scripting language. Together they provide ideal cross-platform development environment for any programming project. It has served for decades as an essential system component in organizations ranging from NASA to Cisco Systems, is a must-know language in the fields of EDA, and powers companies such as FlightAware and F5 Networks.\n\nTcl is fit for both the smallest and largest programming tasks, obviating the need to decide whether it is overkill for a given job or whether a system written in Tcl will scale up as needed. Wherever a shell script might be used Tcl is a better choice, and entire web ecosystems and mission-critical control and testing systems have also been written in Tcl. Tcl excels in all these roles due to the minimal syntax of the language, the unique programming paradigm exposed at the script level, and the careful engineering that has gone into the design of the Tcl internals.")

	_ = printCanvasBtn
	_ = printTextBtn
	app.Run()
}
