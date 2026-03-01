// Phase 7 demo: Scale, Listbox, PanedWindow, Menu, Menubutton.
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
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/listbox"
	"github.com/msorc/takigo/widget/menu"
	"github.com/msorc/takigo/widget/menubutton"
	"github.com/msorc/takigo/widget/panedwindow"
	"github.com/msorc/takigo/widget/scale"
	"github.com/msorc/takigo/widget/scrollbar"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Takigo Phase 7 — Complete Classic Widgets"), takigo.Size(700, 500))
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
		label.Text("Phase 7: Scale, Listbox, PanedWindow, Menu"),
		label.PadX(10),
		label.PadY(5),
	)
	pack.Pack(titleLabel.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	// --- Status label (at bottom) ---
	statusLabel := label.New(root, "status", app,
		label.Text("Drag sliders, select items, resize panes. Right-click for menu. Esc to quit."),
		label.Background("#e8e8e8"),
		label.Anchor(option.AnchorW),
		label.PadX(5),
		label.PadY(2),
	)
	pack.Pack(statusLabel.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// --- Top row: Menu button + Scale ---
	topFrame := frame.New(root, "topFrame", app)
	pack.Pack(topFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadX(10), pack.PadY(5))

	// Create popup menu.
	popupMenu := menu.New(root, "popup", app)
	popupMenu.AddCommand("New", func() { fmt.Println("Menu: New") })
	popupMenu.AddCommand("Open", func() { fmt.Println("Menu: Open") })
	popupMenu.AddSeparator()
	popupMenu.AddCheckbutton("Auto-save", true, func() { fmt.Println("Menu: Auto-save toggled") })
	popupMenu.AddSeparator()
	popupMenu.AddCommand("Quit", func() { app.Quit() })

	// Menubutton.
	mb := menubutton.New(topFrame.Window(), "filemenu", app,
		menubutton.Text("File"),
		menubutton.MenuOpt(popupMenu),
		menubutton.PadX(8),
		menubutton.PadY(2),
	)
	pack.Pack(mb.Window(), pack.SideOpt(pack.Left), pack.PadX(5))

	// Scale (horizontal).
	scaleLabel := label.New(topFrame.Window(), "scaleLabel", app,
		label.Text("Value: 0"),
		label.Anchor(option.AnchorW),
	)

	sc := scale.New(topFrame.Window(), "scale1", app,
		scale.OrientOpt(scale.Horizontal),
		scale.FromOpt(0),
		scale.ToOpt(100),
		scale.ValueOpt(50),
		scale.ShowValueOpt(true),
		scale.CommandOpt(func(v float64) {
			scaleLabel.Text = fmt.Sprintf("Value: %.0f", v)
			scaleLabel.Display()
		}),
	)
	pack.Pack(sc.Window(), pack.SideOpt(pack.Left), pack.FillOpt(pack.FillX), pack.Expand(true), pack.PadX(5))
	pack.Pack(scaleLabel.Window(), pack.SideOpt(pack.Left), pack.PadX(5))

	// --- PanedWindow with Listbox + Info label ---
	pw := panedwindow.New(root, "paned", app,
		panedwindow.OrientOpt(panedwindow.Horizontal),
	)
	pack.Pack(pw.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Left pane: frame with listbox + scrollbar.
	leftFrame := frame.New(pw.Window(), "leftFrame", app)

	items := make([]string, 30)
	for i := range items {
		items[i] = fmt.Sprintf("Item %d — example list entry", i+1)
	}

	lb := listbox.New(leftFrame.Window(), "listbox", app,
		listbox.Items(items...),
		listbox.Height(10),
		listbox.Width(25),
		listbox.SelectModeOpt(listbox.SelectExtended),
	)

	yscroll := scrollbar.New(leftFrame.Window(), "yscroll", app,
		scrollbar.OrientOpt(scrollbar.Vertical),
		scrollbar.WidthOpt(14),
		scrollbar.CommandOpt(func(args ...interface{}) {
			if len(args) < 1 {
				return
			}
			switch args[0] {
			case "moveto":
				if len(args) >= 2 {
					if f, ok := args[1].(float64); ok {
						lb.YViewMoveTo(f)
					}
				}
			case "scroll":
				if len(args) >= 3 {
					n, _ := args[1].(int)
					unit, _ := args[2].(string)
					lb.YViewScroll(n, unit == "pages")
				}
			}
		}),
	)

	// Connect listbox to scrollbar.
	lb.YScrollCmd = func(first, last float64) {
		yscroll.Set(first, last)
	}

	pack.Pack(yscroll.Window(), pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(lb.Window(), pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Right pane: info label.
	rightFrame := frame.New(pw.Window(), "rightFrame", app)
	infoLabel := label.New(rightFrame.Window(), "info", app,
		label.Text("Select items in the listbox.\nDrag the sash to resize panes."),
		label.Anchor(option.AnchorNW),
		label.PadX(10),
		label.PadY(10),
	)
	pack.Pack(infoLabel.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Add panes.
	pw.Add(leftFrame.Window(), 100)
	pw.Add(rightFrame.Window(), 100)

	// Initialize scrollbar.
	first, last := lb.YVisibleRange()
	yscroll.Set(first, last)

	// --- Button row ---
	btnFrame := frame.New(root, "btnFrame", app)
	pack.Pack(btnFrame.Window(), pack.SideOpt(pack.Top), pack.PadY(5))

	getSelBtn := button.New(btnFrame.Window(), "getSel", app,
		button.Text("Get Selection"),
		button.PadX(10),
		button.PadY(3),
		button.Command(func() {
			sel := lb.Selection()
			fmt.Printf("Selected indices: %v\n", sel)
			for _, i := range sel {
				items := lb.GetItems()
				if i < len(items) {
					fmt.Printf("  %s\n", items[i])
				}
			}
		}),
	)
	pack.Pack(getSelBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(5))

	addBtn := button.New(btnFrame.Window(), "addItem", app,
		button.Text("Add Item"),
		button.PadX(10),
		button.PadY(3),
		button.Command(func() {
			n := lb.ItemCount() + 1
			lb.Insert(lb.ItemCount(), fmt.Sprintf("New item %d", n))
		}),
	)
	pack.Pack(addBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(5))

	// --- Root event handlers ---
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

	// Right-click popup menu.
	app.Dispatcher().Bind(root.XWindow, event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button == 3 {
			popupMenu.Post(ev.RootX, ev.RootY)
		}
	})

	// Global key handler.
	app.Dispatcher().BindGlobal(event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == xlib.XK_Escape {
			app.Quit()
		}
	})

	focusMgr.SetFocus(lb.Window())

	fmt.Println("Takigo Phase 7 Demo — Complete Classic Widgets")
	fmt.Println("Right-click for popup menu. Esc to quit.")
	app.MainLoop()
	fmt.Println("Goodbye!")

	_ = titleLabel
	_ = statusLabel
	_ = focusMgr
	_ = infoLabel
	_ = scaleLabel
}
