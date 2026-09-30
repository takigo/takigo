// Demo: A dialog box with a local grab.
// Ported from Tk's dialog1.tcl demo: the demo window is the tk_dialog itself.
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
	app, err := takigo.NewApp(takigo.Title("Dialog with local grab"),
		takigo.IconName("Dialog"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	d := dialog.BuildTkDialog(app, "This is a modal dialog box.  It uses Tk's \"grab\" command to create a \"local grab\" on the dialog box.  The grab prevents any mouse or keyboard events from getting to any other windows in the application until you have answered the dialog by invoking one of the buttons below.  However, you can still interact with other applications.  For example, you should be able to edit text in the window named \"child\" which was created by a child interpreter.",
		"info", 0, []string{"OK", "Cancel", "Show Code"},
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
	// after idle {.dialog1.msg configure -wraplength 4i}
	app.DoWhenIdle(func() { d.Msg.Configure(label.WrapLength("4i")) })
	// tk_dialog gives the default button the focus (tk::SetFocusGrab).
	app.After(0, func() { app.FocusManager().SetFocus(d.Buttons[0].Window()) })

	app.Run()
}
