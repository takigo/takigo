// Demo: This demonstration showcases the tk print commands.
// Ported from Tk's print.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/dialog"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Printing Demonstration"),
		takigo.Geometry("+300+300"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "l",
		label.Text("This demonstration showcases\nthe tk print command. Clicking the buttons below\nprints the data from the canvas and text widgets\nusing platform-native dialogs."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	// Content area: canvas left, text right. Created before the button row
	// so the canvas is in scope when the print button's closure runs.
	m := frame.New(f, "m")
	pack.Pack(m, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true))

	// Canvas with shapes.
	c := canvas.New(m, "c", canvas.Background("white"))
	pack.Pack(c, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth),
		pack.Expand(true))

	c.CreateRectangle(20, 20, 220, 80, canvas.FillColor("blue"), canvas.OutlineColor("black"))
	c.CreateOval(20, 100, 220, 160, canvas.FillColor("green"), canvas.OutlineColor("black"))
	c.CreateText(20, 180, canvas.AnchorOpt(option.AnchorNW),
		canvas.TextColor("black"),
		canvas.TextOpt("A short demo of simple canvas elements."))

	// Text widget with Tcl/Tk description.
	tw := text.New(m, "t", text.WrapModeOpt(text.WrapWord))
	pack.Pack(tw, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillBoth),
		pack.Expand(true))

	tw.Insert("end", "Tcl, or Tool Command Language, is an open-source multi-purpose C library which includes a powerful dynamic scripting language. Together they provide ideal cross-platform development environment for any programming project. It has served for decades as an essential system component in organizations ranging from NASA to Cisco Systems, is a must-know language in the fields of EDA, and powers companies such as FlightAware and F5 Networks.\n\nTcl is fit for both the smallest and largest programming tasks, obviating the need to decide whether it is overkill for a given job or whether a system written in Tcl will scale up as needed. Wherever a shell script might be used Tcl is a better choice, and entire web ecosystems and mission-critical control and testing systems have also been written in Tcl. Tcl excels in all these roles due to the minimal syntax of the language, the unique programming paradigm exposed at the script level, and the careful engineering that has gone into the design of the Tcl internals.")

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Button frame at the bottom.
	btnFrame := frame.New(f, "f")
	pack.Pack(btnFrame, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	printCanvasBtn := button.New(btnFrame, "c",
		button.Text("Print Canvas"),
		button.Command(func() {
			path := "/tmp/takigo-canvas.ps"
			_, err := c.Postscript(canvas.PSFile(path))
			if err != nil {
				dialog.ShowMessage(app,
					dialog.MsgTitle("Print"),
					dialog.MsgMessage("Failed: "+err.Error()),
					dialog.MsgType(dialog.MsgError),
				)
				return
			}
			dialog.ShowMessage(app,
				dialog.MsgTitle("Print"),
				dialog.MsgMessage("Saved to "+path),
				dialog.MsgType(dialog.MsgInfo),
			)
		}),
	)
	pack.Pack(printCanvasBtn, pack.SideOpt(pack.Left), pack.Anchor(option.AnchorW),
		pack.PadX("3p"))

	printTextBtn := button.New(btnFrame, "t",
		button.Text("Print Text"),
		button.Command(func() {
			// Text-widget printing is out of scope for the first port;
			// tk's full print.tcl driver layer builds on top of canvas
			// postscript + a platform print spooler. For now we just show
			// a "not yet" notice.
			dialog.ShowMessage(app,
				dialog.MsgTitle("Print"),
				dialog.MsgMessage("Text-widget printing is not yet implemented in takigo."),
				dialog.MsgType(dialog.MsgInfo),
			)
		}),
	)
	pack.Pack(printTextBtn, pack.SideOpt(pack.Right), pack.Anchor(option.AnchorE),
		pack.PadX("3p"))

	app.Run()
}
