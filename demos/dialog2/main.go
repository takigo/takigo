// Demo: A dialog box with a global grab.
// Ported from Tk's dialog2.tcl demo: the demo window is the tk_dialog itself.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/dialog"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Dialog with global grab"),
		takigo.IconName("Dialog"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var d *dialog.TkDialog
	d = dialog.BuildTkDialog(app, "This dialog box uses a global grab. If you are using an X11 window manager you will be prevented from interacting with anything on your display until you invoke one of the buttons below.  This is almost always a bad idea; don't use global grabs with X11 unless you're truly desperate.  On macOS systems you will not be able to interact with any window belonging to this process, but interaction with other macOS Applications will still be possible.",
		"warning", 0, []string{"OK", "Cancel", "Show Code"},
		func(i int) {
			switch i {
			case 0:
				fmt.Println("You pressed OK")
			case 1:
				fmt.Println("You pressed Cancel")
			case 2:
				demohelper.ShowCode(app)
				return
			}
			app.Quit()
		})
	// after idle {.dialog2.msg configure -wraplength 4i}
	app.DoWhenIdle(func() { d.Msg.Apply(label.WrapLength("4i")) })
	// tk_dialog gives the default button the focus (tk::SetFocusGrab).
	app.After(0, func() { app.FocusManager().SetFocus(d.Buttons[0].Window()) })

	app.Run()
}
