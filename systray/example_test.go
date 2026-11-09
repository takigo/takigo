package systray_test

import (
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/systray"
)

// A tray icon docks in the desktop's system tray; New fails when no tray is
// running.
func ExampleNew() {
	app, err := takigo.NewApp(takigo.Title("Tray"))
	if err != nil {
		log.Fatal(err)
	}
	icon, err := systray.New(app, app.Display(),
		systray.TrayTooltip("takigo example"),
		systray.TrayClickHandler(func() {
			if err := systray.Notify("takigo", "The tray icon was clicked"); err != nil {
				log.Println(err)
			}
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer icon.Destroy()
	app.Run()
}
