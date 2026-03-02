// Demo: TTK Combobox with editable, readonly, and disabled states.
// Ported from Tk's combo.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Combobox Demonstration"), takigo.Size(450, 400))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	ttk.SetCurrentTheme("clam")

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	// Description.
	msg := label.New(root, "msg", app,
		label.Text("Three comboboxes are shown below: editable,\nreadonly, and disabled. Click the arrow to see the\ndropdown list."),
		label.Anchor(option.AnchorW),
		label.PadX(10), label.PadY(5),
	)
	pack.Pack(msg.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// Dismiss button.
	btnFrame := frame.New(root, "btnframe", app)
	pack.Pack(btnFrame.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))
	dismissBtn := button.New(btnFrame.Window(), "dismiss", app,
		button.Text("Dismiss"), button.Command(func() { app.Quit() }),
		button.PadX(10), button.PadY(4),
	)
	pack.Pack(dismissBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(10))

	// Status label.
	statusLabel := label.New(root, "status", app,
		label.Text("Selection: (none)"),
		label.Anchor(option.AnchorW),
		label.PadX(10),
	)
	pack.Pack(statusLabel.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))

	countries := []string{
		"Australia", "Canada", "France", "Germany",
		"Japan", "United Kingdom", "United States",
	}

	// Editable combobox.
	editLabel := label.New(root, "editlabel", app,
		label.Text("Editable:"),
		label.Anchor(option.AnchorW),
		label.PadX(20),
	)
	pack.Pack(editLabel.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	editCombo := ttk.NewCombobox(root, "editcombo", app,
		ttk.ComboboxValues(countries),
		ttk.ComboboxText("Australia"),
		ttk.ComboboxCommand(func(v string) {
			statusLabel.Text = "Selection: " + v
			statusLabel.Display()
		}),
	)
	pack.Pack(editCombo.Window(), pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(5))

	// Readonly combobox.
	roLabel := label.New(root, "rolabel", app,
		label.Text("Readonly:"),
		label.Anchor(option.AnchorW),
		label.PadX(20),
	)
	pack.Pack(roLabel.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	roCombo := ttk.NewCombobox(root, "rocombo", app,
		ttk.ComboboxValues(countries),
		ttk.ComboboxText("Canada"),
		ttk.ComboboxCbState(ttk.ComboReadonly),
		ttk.ComboboxCommand(func(v string) {
			statusLabel.Text = "Selection: " + v
			statusLabel.Display()
		}),
	)
	pack.Pack(roCombo.Window(), pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(5))

	// Disabled combobox.
	disLabel := label.New(root, "dislabel", app,
		label.Text("Disabled:"),
		label.Anchor(option.AnchorW),
		label.PadX(20),
	)
	pack.Pack(disLabel.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	disCombo := ttk.NewCombobox(root, "discombo", app,
		ttk.ComboboxValues(countries),
		ttk.ComboboxText("France"),
		ttk.ComboboxCbState(ttk.ComboDisabled),
	)
	pack.Pack(disCombo.Window(), pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(5))

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
	_ = editLabel
	_ = editCombo
	_ = roLabel
	_ = roCombo
	_ = disLabel
	_ = disCombo
	app.MainLoop()
}
