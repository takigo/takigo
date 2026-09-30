//go:build linux || freebsd || openbsd || netbsd

// Phase 13 demo: dialogs, spinbox, busy window, system tray.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/busy"
	"github.com/msorc/takigo/dialog"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/systray"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/scale"
	"github.com/msorc/takigo/widget/scrollbar"
	"github.com/msorc/takigo/widget/spinbox"
)

func main() {
	app, err := takigo.NewApp(
		takigo.Title("Phase 13: Dialogs Demo"),
		takigo.Size(600, 500),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	root := app.Root()

	// Status label.
	statusLabel := label.New(app, "status",
		label.Text("Status: Ready"),
	)
	pack.Pack(statusLabel, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))

	setStatus := func(msg string) {
		statusLabel.Configure(label.Text(msg))
	}

	// --- Dialog Buttons ---
	dlgFrame := frame.New(app, "dlgframe")
	pack.Pack(dlgFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	dlgLabel := label.New(dlgFrame, "dlglabel", label.Text("Dialogs:"))
	pack.Pack(dlgLabel, pack.SideOpt(pack.Left), pack.PadX(5))

	// Message box buttons.
	msgInfoBtn := button.New(dlgFrame, "msginfo",
		button.Text("Info"),
		button.Command(func() {
			result := dialog.ShowMessage(app,
				dialog.MsgTitle("Information"),
				dialog.MsgMessage("This is an informational message."),
				dialog.MsgDetail("Phase 13 is working!"),
				dialog.MsgType(dialog.MsgInfo),
				dialog.MsgButtons(dialog.BtnOK),
			)
			setStatus(fmt.Sprintf("Message box result: %d", result))
		}),
	)
	pack.Pack(msgInfoBtn, pack.SideOpt(pack.Left), pack.PadX(3))

	msgQBtn := button.New(dlgFrame, "msgq",
		button.Text("Question"),
		button.Command(func() {
			result := dialog.ShowMessage(app,
				dialog.MsgTitle("Question"),
				dialog.MsgMessage("Do you want to continue?"),
				dialog.MsgType(dialog.MsgQuestion),
				dialog.MsgButtons(dialog.BtnYesNoCancel),
			)
			setStatus(fmt.Sprintf("Question result: %d", result))
		}),
	)
	pack.Pack(msgQBtn, pack.SideOpt(pack.Left), pack.PadX(3))

	// Color chooser button.
	colorBtn := button.New(dlgFrame, "colorbtn",
		button.Text("Color..."),
		button.Command(func() {
			color, ok := dialog.ChooseColor(app,
				dialog.ColorInitial("#3399ff"),
			)
			if ok {
				setStatus(fmt.Sprintf("Chosen color: %s", color))
			} else {
				setStatus("Color chooser cancelled")
			}
		}),
	)
	pack.Pack(colorBtn, pack.SideOpt(pack.Left), pack.PadX(3))

	// Font chooser button.
	fontBtn := button.New(dlgFrame, "fontbtn",
		button.Text("Font..."),
		button.Command(func() {
			fontDesc, ok := dialog.ChooseFont(app)
			if ok {
				setStatus(fmt.Sprintf("Chosen font: %s", fontDesc))
			} else {
				setStatus("Font chooser cancelled")
			}
		}),
	)
	pack.Pack(fontBtn, pack.SideOpt(pack.Left), pack.PadX(3))

	// File dialog button.
	fileBtn := button.New(dlgFrame, "filebtn",
		button.Text("Open File..."),
		button.Command(func() {
			path, ok := dialog.OpenFile(app,
				dialog.FileTitle("Select a file"),
				dialog.FileTypes(
					dialog.FileType{Name: "Go files", Pattern: "*.go"},
					dialog.FileType{Name: "All files", Pattern: "*"},
				),
			)
			if ok {
				setStatus(fmt.Sprintf("Selected: %s", path))
			} else {
				setStatus("File dialog cancelled")
			}
		}),
	)
	pack.Pack(fileBtn, pack.SideOpt(pack.Left), pack.PadX(3))

	// --- Spinbox ---
	spinFrame := frame.New(app, "spinframe")
	pack.Pack(spinFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(10))

	spinLabel := label.New(spinFrame, "spinlabel", label.Text("Spinbox (0-100):"))
	pack.Pack(spinLabel, pack.SideOpt(pack.Left), pack.PadX(5))

	spin := spinbox.New(spinFrame, "spin1",
		spinbox.FromOpt(0),
		spinbox.ToOpt(100),
		spinbox.IncrementOpt(1),
		spinbox.WrapOpt(true),
		spinbox.CommandOpt(func(v string) {
			setStatus(fmt.Sprintf("Spinbox value: %s", v))
		}),
	)
	pack.Pack(spin, pack.SideOpt(pack.Left), pack.PadX(5))

	// Values-mode spinbox.
	valSpinLabel := label.New(spinFrame, "valspinlabel", label.Text("Values:"))
	pack.Pack(valSpinLabel, pack.SideOpt(pack.Left), pack.PadX(5))

	valSpin := spinbox.New(spinFrame, "spin2",
		spinbox.ValuesOpt([]string{"Apple", "Banana", "Cherry", "Date", "Elderberry"}),
		spinbox.WrapOpt(true),
		spinbox.CommandOpt(func(v string) {
			setStatus(fmt.Sprintf("Values spinbox: %s", v))
		}),
	)
	pack.Pack(valSpin, pack.SideOpt(pack.Left), pack.PadX(5))

	// --- Mouse wheel test: Scale + Scrollbar ---
	wheelFrame := frame.New(app, "wheelframe")
	pack.Pack(wheelFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(10))

	wheelLabel := label.New(wheelFrame, "wheellabel",
		label.Text("Mouse wheel test (hover + scroll):"))
	pack.Pack(wheelLabel, pack.SideOpt(pack.Left), pack.PadX(5))

	testScale := scale.New(wheelFrame, "testscale",
		scale.OrientOpt(scale.Horizontal),
		scale.FromOpt(0), scale.ToOpt(100),
		scale.ValueOpt(50),
		scale.CommandOpt(func(v float64) {
			setStatus(fmt.Sprintf("Scale value: %.0f", v))
		}),
	)
	pack.Pack(testScale, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillX), pack.Expand(true))

	testScrollbar := scrollbar.New(wheelFrame, "testscroll",
		scrollbar.OrientOpt(scrollbar.Vertical),
		scrollbar.CommandOpt(func(args ...any) {
			setStatus(fmt.Sprintf("Scrollbar: %v", args))
		}),
	)
	testScrollbar.Set(0.3, 0.6)
	pack.Pack(testScrollbar, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))

	// --- Busy window ---
	busyFrame := frame.New(app, "busyframe")
	pack.Pack(busyFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	busyBtn := button.New(busyFrame, "busybtn",
		button.Text("Busy 2s"),
		button.Command(func() {
			setStatus("Busy...")
			bw := busy.Hold(app, root)
			app.After(2*time.Second, func() {
				bw.Release()
				setStatus("Busy released!")
			})
		}),
	)
	pack.Pack(busyBtn, pack.SideOpt(pack.Left), pack.PadX(5))

	// --- System tray ---
	trayBtn := button.New(busyFrame, "traybtn",
		button.Text("Add Tray Icon"),
		button.Command(func() {
			tray, err := systray.New(app, app.Display(),
				systray.TrayTooltip("Takigo Demo"),
				systray.TrayClickHandler(func() {
					setStatus("Tray icon clicked!")
				}),
			)
			if err != nil {
				setStatus(fmt.Sprintf("Tray error: %v", err))
			} else {
				setStatus("Tray icon added")
				// Destroy after 10 seconds.
				app.After(10*time.Second, func() {
					tray.Destroy()
					setStatus("Tray icon removed")
				})
			}
		}),
	)
	pack.Pack(trayBtn, pack.SideOpt(pack.Left), pack.PadX(5))

	// Suppress unused variable warnings.
	_ = dlgLabel
	_ = spinLabel
	_ = valSpinLabel
	_ = spin
	_ = valSpin
	_ = wheelLabel
	_ = testScale

	fmt.Println("Phase 13 demo running. Close the window to exit.")
	app.Run()
}
