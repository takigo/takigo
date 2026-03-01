// Phase 5 demo: Demonstrates toplevel windows, focus traversal,
// and modal dialog via grab.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/focus"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/grab"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/toplevel"
	"github.com/msorc/takigo/wm"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Takigo Phase 5 — WM, Focus, Grab"), takigo.Size(500, 400))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()

	// Initialize WM for root window.
	rootWm := wm.Init(root)
	rootWm.SetTitle("Takigo Phase 5 — WM, Focus, Grab")
	rootWm.SetMinSize(300, 200)
	rootWm.OnDeleteWindow(func() {
		app.Quit()
	})

	// Set root background.
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	// Focus manager.
	focusMgr := focus.NewManager(app.Dispatcher(), app.DisplayPtr())
	focusMgr.BindTraversal(root)

	// Grab manager.
	grabMgr := grab.NewManager(app.DisplayPtr(), app.Dispatcher())

	// Title label.
	titleLabel := label.New(root, "title", app,
		label.Text("Phase 5: Window Manager, Focus, Grab"),
		label.PadX(10),
		label.PadY(5),
	)
	pack.Pack(titleLabel.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	// Info label.
	infoLabel := label.New(root, "info", app,
		label.Text("Tab to traverse focus. Click buttons to open windows."),
		label.PadX(5),
		label.PadY(2),
	)
	pack.Pack(infoLabel.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// Button frame.
	btnFrame := frame.New(root, "buttons", app)
	pack.Pack(btnFrame.Window(), pack.SideOpt(pack.Top), pack.PadY(10))

	// Counter for new windows.
	windowCount := 0

	// "New Window" button.
	newWinBtn := button.New(btnFrame.Window(), "newwin", app,
		button.Text("New Window"),
		button.PadX(10),
		button.PadY(3),
		button.Command(func() {
			windowCount++
			name := fmt.Sprintf("win%d", windowCount)
			title := fmt.Sprintf("Window #%d", windowCount)

			tl := toplevel.New(root, name, app,
				toplevel.Title(title),
				toplevel.Geometry(fmt.Sprintf("300x200+%d+%d", 100+windowCount*30, 100+windowCount*30)),
			)

			// Add content.
			lbl := label.New(tl.Window(), "lbl", app,
				label.Text(fmt.Sprintf("This is %s", title)),
				label.PadX(10),
				label.PadY(10),
			)
			pack.Pack(lbl.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(10))

			closeBtn := button.New(tl.Window(), "close", app,
				button.Text("Close"),
				button.PadX(10),
				button.PadY(3),
			)
			closeBtn.Command = func() {
				tl.Destroy()
			}
			pack.Pack(closeBtn.Window(), pack.SideOpt(pack.Bottom), pack.PadY(10))

			tl.OnClose(func() {
				tl.Destroy()
			})

			tl.Show()
			fmt.Printf("Opened %s\n", title)
		}),
	)
	pack.Pack(newWinBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(5))

	// "Modal Dialog" button — demonstrates grab.
	modalBtn := button.New(btnFrame.Window(), "modal", app,
		button.Text("Modal Dialog"),
		button.PadX(10),
		button.PadY(3),
		button.Command(func() {
			// Create a dialog toplevel.
			dlg := toplevel.New(root, "dialog", app,
				toplevel.Title("Modal Dialog"),
				toplevel.Geometry("300x150+200+200"),
				toplevel.TransientFor(root),
				toplevel.Resizable(false, false),
			)

			dlgLabel := label.New(dlg.Window(), "msg", app,
				label.Text("This is a modal dialog."),
				label.PadX(10),
				label.PadY(10),
			)
			pack.Pack(dlgLabel.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(15))

			okBtn := button.New(dlg.Window(), "ok", app,
				button.Text("OK"),
				button.PadX(20),
				button.PadY(3),
			)
			okBtn.Command = func() {
				grabMgr.Release()
				dlg.Destroy()
				fmt.Println("Dialog closed")
			}
			pack.Pack(okBtn.Window(), pack.SideOpt(pack.Bottom), pack.PadY(10))

			dlg.OnClose(func() {
				grabMgr.Release()
				dlg.Destroy()
			})

			dlg.Show()

			// Set local grab for modal behavior.
			grabMgr.Set(dlg.Window(), false)
			fmt.Println("Modal dialog opened (local grab active)")
		}),
	)
	pack.Pack(modalBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(5))

	// "Iconify" button.
	iconBtn := button.New(btnFrame.Window(), "iconify", app,
		button.Text("Iconify"),
		button.PadX(10),
		button.PadY(3),
		button.Command(func() {
			rootWm.Iconify()
			fmt.Println("Window iconified")
		}),
	)
	pack.Pack(iconBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(5))

	// Status label.
	statusLabel := label.New(root, "status", app,
		label.Text("Ready. Press 'q' to quit."),
		label.Background("#e8e8e8"),
		label.Anchor(option.AnchorW),
		label.PadX(5),
		label.PadY(2),
	)
	pack.Pack(statusLabel.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Handle root resize.
	app.Dispatcher().Bind(root.XWindow, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			root.Width = ev.ConfigWidth
			root.Height = ev.ConfigHeight
			pack.ArrangeContainer(root)
		}
	})

	// Root expose.
	app.Dispatcher().Bind(root.XWindow, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		d := root.Display.XDisplay
		gc := root.GC
		d.SetForeground(gc, bgColor.Pixel)
		d.FillRectangle(root.Drawable(), gc, 0, 0, uint(root.Width), uint(root.Height))
		d.Flush()
	})

	// Key handler.
	app.Dispatcher().Bind(root.XWindow, event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == xlib.XK_q || ev.KeySym == xlib.XK_Escape {
			app.Quit()
		}
	})

	// Set initial focus.
	focusMgr.SetFocus(newWinBtn.Window())

	fmt.Println("Takigo Phase 5 Demo — WM, Focus, Grab")
	fmt.Println("Tab to traverse focus. Click buttons. Press 'q' to quit.")
	app.MainLoop()
	fmt.Println("Goodbye!")

	// Prevent unused warnings.
	_ = titleLabel
	_ = infoLabel
	_ = statusLabel
	_ = focusMgr
	_ = grabMgr
}
