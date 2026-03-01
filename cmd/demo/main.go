// Phase 6 demo: Entry widget, scrollbar, and text editing.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/focus"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/entry"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/scrollbar"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Takigo Phase 6 — Entry & Scrollbar"), takigo.Size(550, 400))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	// Focus manager.
	focusMgr := focus.NewManager(app.Dispatcher(), app.DisplayPtr())
	focusMgr.BindTraversal(root)

	// Title.
	titleLabel := label.New(root, "title", app,
		label.Text("Phase 6: Entry Widget & Scrollbar"),
		label.PadX(10),
		label.PadY(5),
	)
	pack.Pack(titleLabel.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	// --- Entry fields ---

	// Name entry.
	nameFrame := frame.New(root, "nameFrame", app)
	pack.Pack(nameFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadX(10), pack.PadY(5))

	nameLabel := label.New(nameFrame.Window(), "nameLabel", app,
		label.Text("Name:"),
		label.Anchor(option.AnchorW),
	)
	pack.Pack(nameLabel.Window(), pack.SideOpt(pack.Left), pack.PadX(5))

	nameEntry := entry.New(nameFrame.Window(), "nameEntry", app,
		entry.Width(30),
		entry.Placeholder("Enter your name"),
	)
	pack.Pack(nameEntry.Window(), pack.SideOpt(pack.Left), pack.FillOpt(pack.FillX), pack.Expand(true), pack.PadX(5))

	// Password entry.
	passFrame := frame.New(root, "passFrame", app)
	pack.Pack(passFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadX(10), pack.PadY(5))

	passLabel := label.New(passFrame.Window(), "passLabel", app,
		label.Text("Password:"),
		label.Anchor(option.AnchorW),
	)
	pack.Pack(passLabel.Window(), pack.SideOpt(pack.Left), pack.PadX(5))

	passEntry := entry.New(passFrame.Window(), "passEntry", app,
		entry.Width(30),
		entry.Show('*'),
	)
	pack.Pack(passEntry.Window(), pack.SideOpt(pack.Left), pack.FillOpt(pack.FillX), pack.Expand(true), pack.PadX(5))

	// Long text entry with scrollbar.
	scrollFrame := frame.New(root, "scrollFrame", app)
	pack.Pack(scrollFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadX(10), pack.PadY(5))

	scrollLabel := label.New(scrollFrame.Window(), "scrollLabel", app,
		label.Text("Long text:"),
		label.Anchor(option.AnchorW),
	)
	pack.Pack(scrollLabel.Window(), pack.SideOpt(pack.Left), pack.PadX(5))

	longEntry := entry.New(scrollFrame.Window(), "longEntry", app,
		entry.Width(25),
		entry.Text("This is a long text entry that can be scrolled horizontally with the scrollbar below."),
	)
	pack.Pack(longEntry.Window(), pack.SideOpt(pack.Left), pack.FillOpt(pack.FillX), pack.Expand(true), pack.PadX(5))

	// Horizontal scrollbar connected to the long entry.
	hScrollFrame := frame.New(root, "hScrollFrame", app)
	pack.Pack(hScrollFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadX(10))

	hScroll := scrollbar.New(hScrollFrame.Window(), "hscroll", app,
		scrollbar.OrientOpt(scrollbar.Horizontal),
		scrollbar.WidthOpt(12),
		scrollbar.CommandOpt(func(args ...interface{}) {
			if len(args) < 1 {
				return
			}
			switch args[0] {
			case "moveto":
				if len(args) >= 2 {
					if f, ok := args[1].(float64); ok {
						longEntry.XViewMoveTo(f)
					}
				}
			case "scroll":
				if len(args) >= 3 {
					n, _ := args[1].(int)
					unit, _ := args[2].(string)
					longEntry.XViewScroll(n, unit == "pages")
				}
			}
		}),
	)
	pack.Pack(hScroll.Window(), pack.SideOpt(pack.Left), pack.FillOpt(pack.FillX), pack.Expand(true), pack.PadX(80))

	// Connect entry scroll notification to scrollbar.
	longEntry.ScrollCmd = func(first, last float64) {
		hScroll.Set(first, last)
	}

	// Initialize scrollbar position.
	first, last := longEntry.VisibleRange()
	hScroll.Set(first, last)

	// --- Buttons ---

	btnFrame := frame.New(root, "btnFrame", app)
	pack.Pack(btnFrame.Window(), pack.SideOpt(pack.Top), pack.PadY(15))

	getBtn := button.New(btnFrame.Window(), "get", app,
		button.Text("Get Values"),
		button.PadX(10),
		button.PadY(3),
		button.Command(func() {
			fmt.Printf("Name: %q\n", nameEntry.GetText())
			fmt.Printf("Password: %q\n", passEntry.GetText())
			fmt.Printf("Long text: %q\n", longEntry.GetText())
			if sel := nameEntry.SelectedText(); sel != "" {
				fmt.Printf("Name selection: %q\n", sel)
			}
		}),
	)
	pack.Pack(getBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(5))

	clearBtn := button.New(btnFrame.Window(), "clear", app,
		button.Text("Clear All"),
		button.PadX(10),
		button.PadY(3),
		button.Command(func() {
			nameEntry.SetText("")
			passEntry.SetText("")
			longEntry.SetText("")
			fmt.Println("All entries cleared")
		}),
	)
	pack.Pack(clearBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(5))

	selectAllBtn := button.New(btnFrame.Window(), "selall", app,
		button.Text("Select All Name"),
		button.PadX(10),
		button.PadY(3),
		button.Command(func() {
			nameEntry.SelectAll()
			nameEntry.Display()
		}),
	)
	pack.Pack(selectAllBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(5))

	// Status.
	statusLabel := label.New(root, "status", app,
		label.Text("Type in entries. Tab to switch focus. Esc to quit."),
		label.Background("#e8e8e8"),
		label.Anchor(option.AnchorW),
		label.PadX(5),
		label.PadY(2),
	)
	pack.Pack(statusLabel.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Root event handlers.
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
		gc := root.GC
		d.SetForeground(gc, bgColor.Pixel)
		d.FillRectangle(root.Drawable(), gc, 0, 0, uint(root.Width), uint(root.Height))
		d.Flush()
	})

	// Global key handler — works regardless of which widget has focus.
	app.Dispatcher().BindGlobal(event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == xlib.XK_Escape {
			app.Quit()
		}
	})

	// Set initial focus.
	focusMgr.SetFocus(nameEntry.Window())

	fmt.Println("Takigo Phase 6 Demo — Entry & Scrollbar")
	fmt.Println("Type in entry fields. Tab to navigate. Esc to quit.")
	app.MainLoop()
	fmt.Println("Goodbye!")

	_ = titleLabel
	_ = nameLabel
	_ = passLabel
	_ = scrollLabel
	_ = statusLabel
	_ = focusMgr
}
