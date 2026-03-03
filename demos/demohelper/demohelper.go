// Package demohelper provides common boilerplate for Tk demo applications.
package demohelper

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

// Setup creates a standard demo window with a description label and dismiss
// button. Calls os.Exit(1) on failure.
func Setup(title string, width, height int, description string) *takigo.App {
	app, err := takigo.NewApp(takigo.Title(title), takigo.Size(width, height))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	// Description label.
	msg := label.New(app, "msg",
		label.Text(description),
		label.Anchor(option.AnchorW),
		label.PadX(10),
		label.PadY(5),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// Dismiss button at bottom.
	btnFrame := frame.New(app, "btnframe")
	pack.Pack(btnFrame, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))

	dismissBtn := button.New(btnFrame, "dismiss",
		button.Text("Dismiss"),
		button.Command(func() { app.Quit() }),
		button.PadX(10),
		button.PadY(4),
	)
	pack.Pack(dismissBtn, pack.SideOpt(pack.Left), pack.PadX(10))

	// Configure handler.
	app.Dispatcher().Bind(root.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			root.Width = ev.ConfigWidth
			root.Height = ev.ConfigHeight
			pack.ArrangeContainer(root)
		}
	})

	// Expose handler — reads root.BackgroundPixel so dynamic bg changes work.
	app.Dispatcher().Bind(root.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		d := root.Display.Server
		gc := root.GC
		d.SetForeground(gc, root.BackgroundPixel)
		d.FillRectangle(root.Drawable(), gc, 0, 0, uint(root.Width), uint(root.Height))
		d.Flush()
	})

	// Escape to quit.
	app.Dispatcher().BindGlobal(event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == platform.XK_Escape {
			app.Quit()
		}
	})

	return app
}

// NewFrame creates a plain frame as a child of the given parent.
func NewFrame(parent widget.Caregiver, name string) *frame.Frame {
	return frame.New(parent, name)
}

// DemoDir returns the absolute path to a demo directory by name,
// relative to the demos/ root found via the caller's source file location.
func DemoDir(name string) string {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		return name
	}
	// Walk up from callers' source file to find the demos/ root.
	// demos/demohelper/demohelper.go → demos/ is one level up.
	demosRoot := filepath.Dir(filepath.Dir(file))
	return filepath.Join(demosRoot, name)
}
