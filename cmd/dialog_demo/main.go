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
	statusLabel := label.New(root, "status", app,
		label.Text("Status: Ready"),
	)
	pack.Pack(statusLabel.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))

	setStatus := func(msg string) {
		statusLabel.Text = msg
		statusLabel.Display()
	}

	// --- Dialog Buttons ---
	dlgFrame := frame.New(root, "dlgframe", app)
	pack.Pack(dlgFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	dlgLabel := label.New(dlgFrame.Window(), "dlglabel", app, label.Text("Dialogs:"))
	pack.Pack(dlgLabel.Window(), pack.SideOpt(pack.Left), pack.PadX(5))

	// Message box buttons.
	msgInfoBtn := button.New(dlgFrame.Window(), "msginfo", app,
		button.Text("Info"),
		button.Command(func() {
			result := dialog.ShowMessage(app,
				dialog.MsgParent(root),
				dialog.MsgTitle("Information"),
				dialog.MsgMessage("This is an informational message."),
				dialog.MsgDetail("Phase 13 is working!"),
				dialog.MsgType(dialog.MsgInfo),
				dialog.MsgButtons(dialog.BtnOK),
			)
			setStatus(fmt.Sprintf("Message box result: %d", result))
		}),
	)
	pack.Pack(msgInfoBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(3))

	msgQBtn := button.New(dlgFrame.Window(), "msgq", app,
		button.Text("Question"),
		button.Command(func() {
			result := dialog.ShowMessage(app,
				dialog.MsgParent(root),
				dialog.MsgTitle("Question"),
				dialog.MsgMessage("Do you want to continue?"),
				dialog.MsgType(dialog.MsgQuestion),
				dialog.MsgButtons(dialog.BtnYesNoCancel),
			)
			setStatus(fmt.Sprintf("Question result: %d", result))
		}),
	)
	pack.Pack(msgQBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(3))

	// Color chooser button.
	colorBtn := button.New(dlgFrame.Window(), "colorbtn", app,
		button.Text("Color..."),
		button.Command(func() {
			color, ok := dialog.ChooseColor(app,
				dialog.ColorParent(root),
				dialog.ColorInitial("#3399ff"),
			)
			if ok {
				setStatus(fmt.Sprintf("Chosen color: %s", color))
			} else {
				setStatus("Color chooser cancelled")
			}
		}),
	)
	pack.Pack(colorBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(3))

	// Font chooser button.
	fontBtn := button.New(dlgFrame.Window(), "fontbtn", app,
		button.Text("Font..."),
		button.Command(func() {
			fontDesc, ok := dialog.ChooseFont(app,
				dialog.FontParent(root),
			)
			if ok {
				setStatus(fmt.Sprintf("Chosen font: %s", fontDesc))
			} else {
				setStatus("Font chooser cancelled")
			}
		}),
	)
	pack.Pack(fontBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(3))

	// File dialog button.
	fileBtn := button.New(dlgFrame.Window(), "filebtn", app,
		button.Text("Open File..."),
		button.Command(func() {
			path, ok := dialog.OpenFile(app,
				dialog.FileParent(root),
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
	pack.Pack(fileBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(3))

	// --- Spinbox ---
	spinFrame := frame.New(root, "spinframe", app)
	pack.Pack(spinFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(10))

	spinLabel := label.New(spinFrame.Window(), "spinlabel", app, label.Text("Spinbox (0-100):"))
	pack.Pack(spinLabel.Window(), pack.SideOpt(pack.Left), pack.PadX(5))

	spin := spinbox.New(spinFrame.Window(), "spin1", app,
		spinbox.FromOpt(0),
		spinbox.ToOpt(100),
		spinbox.IncrementOpt(1),
		spinbox.WrapOpt(true),
		spinbox.CommandOpt(func(v string) {
			setStatus(fmt.Sprintf("Spinbox value: %s", v))
		}),
	)
	pack.Pack(spin.Window(), pack.SideOpt(pack.Left), pack.PadX(5))

	// Values-mode spinbox.
	valSpinLabel := label.New(spinFrame.Window(), "valspinlabel", app, label.Text("Values:"))
	pack.Pack(valSpinLabel.Window(), pack.SideOpt(pack.Left), pack.PadX(5))

	valSpin := spinbox.New(spinFrame.Window(), "spin2", app,
		spinbox.ValuesOpt([]string{"Apple", "Banana", "Cherry", "Date", "Elderberry"}),
		spinbox.WrapOpt(true),
		spinbox.CommandOpt(func(v string) {
			setStatus(fmt.Sprintf("Values spinbox: %s", v))
		}),
	)
	pack.Pack(valSpin.Window(), pack.SideOpt(pack.Left), pack.PadX(5))

	// --- Mouse wheel test: Scale + Scrollbar ---
	wheelFrame := frame.New(root, "wheelframe", app)
	pack.Pack(wheelFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(10))

	wheelLabel := label.New(wheelFrame.Window(), "wheellabel", app,
		label.Text("Mouse wheel test (hover + scroll):"))
	pack.Pack(wheelLabel.Window(), pack.SideOpt(pack.Left), pack.PadX(5))

	testScale := scale.New(wheelFrame.Window(), "testscale", app,
		scale.OrientOpt(scale.Horizontal),
		scale.FromOpt(0), scale.ToOpt(100),
		scale.ValueOpt(50),
		scale.CommandOpt(func(v float64) {
			setStatus(fmt.Sprintf("Scale value: %.0f", v))
		}),
	)
	pack.Pack(testScale.Window(), pack.SideOpt(pack.Left), pack.FillOpt(pack.FillX), pack.Expand(true))

	testScrollbar := scrollbar.New(wheelFrame.Window(), "testscroll", app,
		scrollbar.OrientOpt(scrollbar.Vertical),
		scrollbar.CommandOpt(func(args ...interface{}) {
			setStatus(fmt.Sprintf("Scrollbar: %v", args))
		}),
	)
	testScrollbar.Set(0.3, 0.6)
	pack.Pack(testScrollbar.Window(), pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))

	// --- Busy window ---
	busyFrame := frame.New(root, "busyframe", app)
	pack.Pack(busyFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	busyBtn := button.New(busyFrame.Window(), "busybtn", app,
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
	pack.Pack(busyBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(5))

	// --- System tray ---
	trayBtn := button.New(busyFrame.Window(), "traybtn", app,
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
	pack.Pack(trayBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(5))

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
