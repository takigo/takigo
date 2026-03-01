// Phase 4 demo: Demonstrates frame, label, and button widgets
// with interactive behavior and geometry management.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Takigo Phase 4 — Widgets"), takigo.Size(500, 400))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()

	// Set root background.
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	// Create a frame at the top.
	topFrame := frame.New(root, "top", app,
		frame.Background("#c0c0c0"),
		frame.BorderWidth(2),
		frame.Relief(option.ReliefGroove),
	)
	pack.Pack(topFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5), pack.PadX(5))

	// Title label.
	titleLabel := label.New(topFrame.Window(), "title", app,
		label.Text("Takigo Widget Demo"),
		label.Background("#c0c0c0"),
		label.PadX(5),
		label.PadY(5),
	)
	pack.Pack(titleLabel.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// Counter display.
	counter := 0
	counterLabel := label.New(root, "counter", app,
		label.Text("Count: 0"),
		label.BorderWidth(2),
		label.Relief(option.ReliefSunken),
		label.PadX(10),
		label.PadY(5),
	)
	pack.Pack(counterLabel.Window(), pack.SideOpt(pack.Top), pack.PadY(10))

	// Button frame.
	btnFrame := frame.New(root, "buttons", app)
	pack.Pack(btnFrame.Window(), pack.SideOpt(pack.Top), pack.PadY(5))

	// Increment button.
	_ = button.New(btnFrame.Window(), "inc", app,
		button.Text("Increment"),
		button.Command(func() {
			counter++
			counterLabel.Text = fmt.Sprintf("Count: %d", counter)
			counterLabel.Display()
		}),
		button.PadX(10),
		button.PadY(3),
	)

	// Decrement button.
	_ = button.New(btnFrame.Window(), "dec", app,
		button.Text("Decrement"),
		button.Command(func() {
			counter--
			counterLabel.Text = fmt.Sprintf("Count: %d", counter)
			counterLabel.Display()
		}),
		button.PadX(10),
		button.PadY(3),
	)

	// Pack buttons side by side.
	for _, child := range btnFrame.Window().Children {
		pack.Pack(child, pack.SideOpt(pack.Left), pack.PadX(5))
	}

	// Reset button.
	_ = button.New(root, "reset", app,
		button.Text("Reset"),
		button.Command(func() {
			counter = 0
			counterLabel.Text = "Count: 0"
			counterLabel.Display()
		}),
		button.PadX(10),
		button.PadY(3),
	)
	for _, child := range root.Children {
		if child.Name == "reset" {
			pack.Pack(child, pack.SideOpt(pack.Top), pack.PadY(5))
			break
		}
	}

	// Status label at bottom.
	statusLabel := label.New(root, "status", app,
		label.Text("Ready. Click buttons to change the counter."),
		label.Background("#e8e8e8"),
		label.Anchor(option.AnchorW),
		label.PadX(5),
		label.PadY(2),
	)
	pack.Pack(statusLabel.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Handle ConfigureNotify for resize.
	app.Dispatcher().Bind(root.XWindow, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			root.Width = ev.ConfigWidth
			root.Height = ev.ConfigHeight
			pack.ArrangeContainer(root)
		}
	})

	// Root expose handler.
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

	fmt.Println("Takigo Phase 4 Demo — Widgets")
	fmt.Println("Click buttons to change counter. Press 'q' to quit.")
	app.MainLoop()
	fmt.Println("Goodbye!")

	// Prevent unused warnings.
	_ = topFrame
	_ = titleLabel
	_ = counterLabel
	_ = statusLabel
}
