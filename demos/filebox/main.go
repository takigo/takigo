// Demo: File open/save dialogs.
// Ported from Tk's filebox.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/dialog"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("File Dialogs"), takigo.Size(400, 250))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	msg := label.New(root, "msg", app,
		label.Text("Click a button to open a file dialog.\nThe selected path is shown in the status bar."),
		label.Anchor(option.AnchorW),
		label.PadX(10), label.PadY(5),
	)
	pack.Pack(msg.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	btnFrame := frame.New(root, "btnframe", app)
	pack.Pack(btnFrame.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))
	dismissBtn := button.New(btnFrame.Window(), "dismiss", app,
		button.Text("Dismiss"), button.Command(func() { app.Quit() }),
		button.PadX(10), button.PadY(4),
	)
	pack.Pack(dismissBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(10))

	statusLabel := label.New(root, "status", app,
		label.Text("Selected: —"),
		label.Anchor(option.AnchorW),
		label.Background("#e8e8e8"),
		label.PadX(5), label.PadY(2),
	)
	pack.Pack(statusLabel.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	setStatus := func(s string) {
		statusLabel.Text = s
		statusLabel.Display()
	}

	fileTypes := []dialog.FileType{
		{Name: "Go files", Pattern: "*.go"},
		{Name: "Text files", Pattern: "*.txt"},
		{Name: "All files", Pattern: "*"},
	}

	openBtn := button.New(root, "open", app,
		button.Text("Open File..."),
		button.Command(func() {
			path, ok := dialog.OpenFile(app,
				dialog.FileParent(root),
				dialog.FileTitle("Open File"),
				dialog.FileTypes(fileTypes...),
			)
			if ok {
				setStatus(fmt.Sprintf("Open: %s", path))
			} else {
				setStatus("Open cancelled")
			}
		}),
		button.PadX(10), button.PadY(6),
	)
	pack.Pack(openBtn.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(30), pack.PadY(10))

	saveBtn := button.New(root, "save", app,
		button.Text("Save File..."),
		button.Command(func() {
			path, ok := dialog.SaveFile(app,
				dialog.FileParent(root),
				dialog.FileTitle("Save File"),
				dialog.FileTypes(fileTypes...),
			)
			if ok {
				setStatus(fmt.Sprintf("Save: %s", path))
			} else {
				setStatus("Save cancelled")
			}
		}),
		button.PadX(10), button.PadY(6),
	)
	pack.Pack(saveBtn.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(30), pack.PadY(5))

	// Root events.
	app.Dispatcher().Bind(root.XWindow, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			root.Width = ev.ConfigWidth
			root.Height = ev.ConfigHeight
			pack.ArrangeContainer(root)
		}
	})
	app.Dispatcher().Bind(root.XWindow, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		d := root.Display.XDisplay
		d.SetForeground(root.GC, bgColor.Pixel)
		d.FillRectangle(root.Drawable(), root.GC, 0, 0, uint(root.Width), uint(root.Height))
		d.Flush()
	})
	app.Dispatcher().BindGlobal(event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == xlib.XK_Escape {
			app.Quit()
		}
	})

	_ = msg
	_ = dismissBtn
	_ = statusLabel
	_ = openBtn
	_ = saveBtn
	app.MainLoop()
}
