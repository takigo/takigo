package wm_test

import (
	"fmt"
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/dialog"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/widget/label"
	"github.com/takigo/takigo/wm"
)

// ParseGeometry reads the "wm geometry" syntax, a size, a position or both.
func ExampleParseGeometry() {
	for _, s := range []string{"640x480", "+100+50", "300x200-0-0"} {
		g, err := wm.ParseGeometry(s)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("size %v %dx%d, position %v %d,%d (from right %v, from bottom %v)\n",
			g.HasSize, g.Width, g.Height, g.HasPos, g.X, g.Y, g.NegX, g.NegY)
	}
	// Output:
	// size true 640x480, position false 0,0 (from right false, from bottom false)
	// size false 0x0, position true 100,50 (from right false, from bottom false)
	// size true 300x200, position true 0,0 (from right true, from bottom true)
}

// The App's WmInfo is the main window's "wm" command: its title, size
// limits and the handler for the window manager's close button.
func ExampleWmInfo() {
	app, err := takigo.NewApp()
	if err != nil {
		log.Fatal(err)
	}
	pack.Pack(label.New(app, "l", label.Text("Close me")), pack.PadX(40), pack.PadY(40))
	info := app.WmInfo()
	info.SetTitle("Window manager")
	info.SetMinSize(200, 100)
	if err := info.SetGeometry("400x200+100+100"); err != nil {
		log.Fatal(err)
	}
	info.OnDeleteWindow(func() {
		if dialog.ShowMessage(app, dialog.MsgMessage("Quit?"), dialog.MsgButtons(dialog.BtnYesNo)) == dialog.ResultYes {
			app.Quit()
		}
	})
	app.Run()
}
