// Widget Demo Launcher — browse and run all takigo demos.
// Ported from Tk's widget demo launcher (widget.tcl).
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/bind"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/listbox"
	"github.com/msorc/takigo/widget/scrollbar"
	"github.com/msorc/takigo/widget/text"
)

type demoEntry struct {
	name string // directory name under demo/
	desc string // human-readable description
}

type category struct {
	title string
	demos []demoEntry
}

var categories = []category{
	{"Labels, Buttons & Entries", []demoEntry{
		{"label", "Labels with relief styles"},
		{"button", "Buttons that change background"},
		{"puzzle", "15-puzzle game"},
		{"entry1", "Basic entry widgets"},
		{"entry2", "Entries with scrollbars"},
		{"entry3", "Validated entry fields"},
		{"form", "Form with grid layout"},
		{"spin", "Spinbox widgets"},
		{"check", "Checkbutton toggles"},
		{"radio", "Radiobutton groups"},
		{"labelframe", "Labelframe with border label"},
	}},
	{"Listboxes", []demoEntry{
		{"states", "US states listbox"},
		{"colors", "Color names with double-click"},
		{"sayings", "Famous sayings listbox"},
	}},
	{"Text Widget", []demoEntry{
		{"text", "Basic text editing"},
		{"style", "Text display styles (tags)"},
		{"bind_text", "Hypertext-like tag display"},
		{"search", "Text search and highlight"},
		{"textpeer", "Two text widgets with copy"},
		{"twind", "Text tags, styles, and undo"},
	}},
	{"Canvas", []demoEntry{
		{"items", "All canvas item types"},
		{"plot", "2D data plot"},
		{"arrow", "Arrow shapes on lines"},
		{"ruler", "Ruler with tab stops"},
		{"cscroll", "Scrollable canvas grid"},
		{"floor", "Building floorplan"},
		{"ctext", "Canvas text items"},
	}},
	{"Scales", []demoEntry{
		{"hscale", "Horizontal scale (arrow angle)"},
		{"vscale", "Vertical scale (bar height)"},
	}},
	{"Paned Windows", []demoEntry{
		{"paned1", "Horizontal colored panes"},
		{"paned2", "Vertical panes (listbox + text)"},
	}},
	{"Menus", []demoEntry{
		{"menu", "Menu bar (File/Edit/Help)"},
		{"menubu", "Menu buttons (4 directions)"},
		{"toolbar", "Toolbar with TTK menubutton"},
	}},
	{"Dialogs", []demoEntry{
		{"msgbox", "Message box varieties"},
		{"filebox", "File open/save dialogs"},
		{"clrpick", "Color picker dialog"},
		{"fontchoose", "Font chooser dialog"},
		{"dialog_modal", "Modal dialog examples"},
	}},
	{"Images", []demoEntry{
		{"image1", "Generated gradient/checkerboard"},
		{"image2", "Image viewer with listbox"},
		{"bitmap", "Built-in bitmap patterns"},
		{"icon", "Iconic buttons with images"},
	}},
	{"TTK Themed Widgets", []demoEntry{
		{"ttkbut", "TTK buttons and labels"},
		{"ttkpane", "TTK frames in paned window"},
		{"ttkscale", "Scale with label feedback"},
		{"ttknote", "TTK notebook (tabs)"},
		{"ttkprogress", "TTK progressbar"},
		{"combo", "TTK combobox"},
		{"ttkmenu", "TTK menubutton"},
		{"ttkspin", "TTK spinbox"},
		{"mclist", "Multi-column sortable list"},
		{"tree", "Directory tree browser"},
		{"windowicons", "Window icon setting"},
	}},
	{"Special Features", []demoEntry{
		{"systray", "System tray icon"},
		{"unicodeout", "Unicode text display"},
		{"print", "Canvas & text display"},
	}},
	{"Animation", []demoEntry{
		{"anilabel", "Animated scrolling label"},
		{"aniwave", "Animated sine waveform"},
		{"pendulum", "Pendulum physics simulation"},
		{"knightstour", "Knight's tour visualization"},
		{"goldberg", "Rube Goldberg machine"},
	}},
}

func main() {
	app, err := takigo.NewApp(takigo.Title("Takigo Widget Demos"), takigo.Size(700, 550))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	// Find demo base directory.
	_, thisFile, _, _ := runtime.Caller(0)
	demoBase := filepath.Dir(filepath.Dir(thisFile))

	// Header.
	header := label.New(root, "header", app,
		label.Text("Takigo Widget Demonstrations"),
		label.Anchor(option.AnchorCenter),
		label.PadX(10), label.PadY(8),
	)
	pack.Pack(header.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// Bottom button frame.
	btnFrame := frame.New(root, "btnframe", app)
	pack.Pack(btnFrame.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))

	dismissBtn := button.New(btnFrame.Window(), "dismiss", app,
		button.Text("Quit"), button.Command(func() { app.Quit() }),
		button.PadX(10), button.PadY(4),
	)
	pack.Pack(dismissBtn.Window(), pack.SideOpt(pack.Right), pack.PadX(10))

	// Status bar.
	statusLabel := label.New(root, "status", app,
		label.Text("Select a demo and click Run."),
		label.Anchor(option.AnchorW),
		label.Background("#e8e8e8"),
		label.PadX(5), label.PadY(2),
	)
	pack.Pack(statusLabel.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Main content: left listbox + right source viewer.
	contentFrame := frame.New(root, "content", app)
	pack.Pack(contentFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(5), pack.PadY(2))

	// Left panel: listbox of demos.
	leftFrame := frame.New(contentFrame.Window(), "left", app)
	pack.Pack(leftFrame.Window(), pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth),
		pack.Expand(false))

	// Action buttons (pack bottom FIRST so listbox doesn't consume all space).
	actionFrame := frame.New(leftFrame.Window(), "actions", app)
	pack.Pack(actionFrame.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(3))

	sb := scrollbar.New(leftFrame.Window(), "sb", app)
	lb := listbox.New(leftFrame.Window(), "demos", app,
		listbox.Height(20), listbox.Width(35),
	)
	lb.YScrollCmd = func(first, last float64) { sb.Set(first, last) }
	sb.Command = func(args ...interface{}) {
		if len(args) >= 2 {
			action, _ := args[0].(string)
			number, _ := args[1].(float64)
			switch action {
			case "moveto":
				lb.YViewMoveTo(number)
			case "scroll":
				unit := "units"
				if len(args) >= 3 {
					if u, ok := args[2].(string); ok {
						unit = u
					}
				}
				if unit == "pages" {
					lb.YViewScroll(int(number), true)
				} else {
					lb.YViewScroll(int(number), false)
				}
			}
		}
	}
	pack.Pack(sb.Window(), pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(lb.Window(), pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Populate listbox with categories and demos.
	type listItem struct {
		isCategory bool
		demoDir    string
		demoDesc   string
	}
	var items []listItem

	for _, cat := range categories {
		lb.Insert(lb.ItemCount(), fmt.Sprintf("── %s ──", cat.title))
		items = append(items, listItem{isCategory: true})
		for _, d := range cat.demos {
			lb.Insert(lb.ItemCount(), fmt.Sprintf("  %s", d.desc))
			items = append(items, listItem{demoDir: d.name, demoDesc: d.desc})
		}
	}

	// Right panel: source viewer.
	rightFrame := frame.New(contentFrame.Window(), "right", app)
	pack.Pack(rightFrame.Window(), pack.SideOpt(pack.Right), pack.FillOpt(pack.FillBoth),
		pack.Expand(true))

	rightLabel := label.New(rightFrame.Window(), "srclabel", app,
		label.Text("Source Code:"),
		label.Anchor(option.AnchorW),
		label.PadX(5),
	)
	pack.Pack(rightLabel.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	srcSb := scrollbar.New(rightFrame.Window(), "srcsb", app)
	srcText := text.New(rightFrame.Window(), "source", app,
		text.WrapModeOpt(text.WrapNone),
	)
	srcText.YScrollCmd = func(first, last float64) { srcSb.Set(first, last) }
	srcSb.Command = func(args ...interface{}) {
		if len(args) >= 2 {
			action, _ := args[0].(string)
			number, _ := args[1].(float64)
			switch action {
			case "moveto":
				srcText.YViewMoveTo(number)
			case "scroll":
				unit := "units"
				if len(args) >= 3 {
					if u, ok := args[2].(string); ok {
						unit = u
					}
				}
				if unit == "pages" {
					srcText.YViewScroll(int(number), true)
				} else {
					srcText.YViewScroll(int(number), false)
				}
			}
		}
	}
	pack.Pack(srcSb.Window(), pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(srcText.Window(), pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth),
		pack.Expand(true))

	getSelectedDemo := func() (string, string, bool) {
		sel := lb.Selection()
		if len(sel) == 0 {
			return "", "", false
		}
		idx := sel[0]
		if idx < 0 || idx >= len(items) || items[idx].isCategory {
			return "", "", false
		}
		return items[idx].demoDir, items[idx].demoDesc, true
	}

	runBtn := button.New(actionFrame.Window(), "run", app,
		button.Text("Run Demo"),
		button.Command(func() {
			dir, desc, ok := getSelectedDemo()
			if !ok {
				statusLabel.Text = "Select a demo first."
				statusLabel.Display()
				return
			}
			demoDir := filepath.Join(demoBase, dir)
			statusLabel.Text = fmt.Sprintf("Running: %s...", desc)
			statusLabel.Display()

			go func() {
				cmd := exec.Command("go", "run", ".")
				cmd.Dir = demoDir
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					fmt.Fprintf(os.Stderr, "Demo %s error: %v\n", dir, err)
				}
			}()
		}),
		button.PadX(8), button.PadY(3),
	)
	pack.Pack(runBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(3))

	codeBtn := button.New(actionFrame.Window(), "code", app,
		button.Text("See Code"),
		button.Command(func() {
			dir, desc, ok := getSelectedDemo()
			if !ok {
				statusLabel.Text = "Select a demo first."
				statusLabel.Display()
				return
			}
			srcPath := filepath.Join(demoBase, dir, "main.go")
			data, err := os.ReadFile(srcPath)
			if err != nil {
				statusLabel.Text = fmt.Sprintf("Error: %v", err)
				statusLabel.Display()
				return
			}
			srcText.Delete("1.0", "end")
			srcText.Insert("1.0", string(data))
			rightLabel.Text = fmt.Sprintf("Source: %s/main.go", dir)
			rightLabel.Display()
			statusLabel.Text = fmt.Sprintf("Showing source for: %s", desc)
			statusLabel.Display()
		}),
		button.PadX(8), button.PadY(3),
	)
	pack.Pack(codeBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(3))

	// Show source on selection change via double-click.
	eng := app.BindEng()
	eng.Bind(lb.Window().PathName, "<Double-Button-1>", func(ed *bind.EventData) bool {
		dir, _, ok := getSelectedDemo()
		if !ok {
			return false
		}
		srcPath := filepath.Join(demoBase, dir, "main.go")
		data, err := os.ReadFile(srcPath)
		if err != nil {
			return false
		}
		srcText.Delete("1.0", "end")
		srcText.Insert("1.0", string(data))
		rightLabel.Text = fmt.Sprintf("Source: %s/main.go", dir)
		rightLabel.Display()
		return false
	})

	// Initial source display.
	srcText.Insert("1.0", strings.Join([]string{
		"Welcome to the Takigo Widget Demonstrations!",
		"",
		"Select a demo from the list on the left, then:",
		"  - Click 'Run Demo' to launch it",
		"  - Click 'See Code' to view its source code",
		"  - Double-click a demo to view its source",
		"",
		fmt.Sprintf("Total demos: %d", countDemos()),
	}, "\n"))

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

	_ = header
	_ = dismissBtn
	_ = statusLabel
	_ = runBtn
	_ = codeBtn
	app.MainLoop()
}

func countDemos() int {
	total := 0
	for _, cat := range categories {
		total += len(cat.demos)
	}
	return total
}
