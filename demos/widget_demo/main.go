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
	{"Labels, buttons, checkbuttons, and radiobuttons", []demoEntry{
		{"label", "Labels (text and bitmaps)"},
		{"unicodeout", "Labels and UNICODE text"},
		{"button", "Buttons"},
		{"check", "Check-buttons (select any of a group)"},
		{"radio", "Radio-buttons (select one of a group)"},
		{"puzzle", "A 15-puzzle game made out of buttons"},
		{"icon", "Iconic buttons that use bitmaps"},
		{"image1", "Two labels displaying images"},
		{"image2", "A simple user interface for viewing images"},
		{"labelframe", "Labelled frames"},
		{"ttkbut", "The simple Themed Tk widgets"},
	}},
	{"Listboxes and Trees", []demoEntry{
		{"states", "The 50 states"},
		{"colors", "Colors: change the color scheme for the application"},
		{"sayings", "A collection of famous and infamous sayings"},
		{"mclist", "A multi-column list of countries"},
		{"tree", "A directory browser tree"},
	}},
	{"Entries, Spin-boxes and Combo-boxes", []demoEntry{
		{"entry1", "Entries without scrollbars"},
		{"entry2", "Entries with scrollbars"},
		{"entry3", "Validated entries and password fields"},
		{"spin", "Spin-boxes"},
		{"ttkspin", "Themed spin-boxes"},
		{"combo", "Combo-boxes"},
		{"form", "Simple Rolodex-like form"},
	}},
	{"Text", []demoEntry{
		{"text", "Basic editable text"},
		{"style", "Text display styles"},
		{"bind", "Hypertext (tag bindings)"},
		{"twind", "A text widget with embedded windows and other features"},
		{"search", "A search tool built with a text widget"},
		{"textpeer", "Peering text widgets"},
	}},
	{"Canvases", []demoEntry{
		{"items", "The canvas item types"},
		{"plot", "A simple 2-D plot"},
		{"ctext", "Text items in canvases"},
		{"arrow", "An editor for arrowheads on canvas lines"},
		{"ruler", "A ruler with adjustable tab stops"},
		{"floor", "A building floor plan"},
		{"cscroll", "A simple scrollable canvas"},
		{"knightstour", "A Knight's tour of the chess board"},
	}},
	{"Scales and Progress Bars", []demoEntry{
		{"hscale", "Horizontal scale"},
		{"vscale", "Vertical scale"},
		{"ttkscale", "Themed scale linked to a label with traces"},
		{"ttkprogress", "Progress bar"},
	}},
	{"Paned Windows and Notebooks", []demoEntry{
		{"paned1", "Horizontal paned window"},
		{"paned2", "Vertical paned window"},
		{"ttkpane", "Themed nested panes"},
		{"ttknote", "Notebook widget"},
	}},
	{"Menus and Toolbars", []demoEntry{
		{"menu", "Menus and cascades (sub-menus)"},
		{"menubu", "Menu-buttons"},
		{"ttkmenu", "Themed menu buttons"},
		{"toolbar", "Themed toolbar"},
	}},
	{"Common Dialogs", []demoEntry{
		{"msgbox", "Message boxes"},
		{"filebox", "File selection dialog"},
		{"clrpick", "Color picker"},
		{"fontchoose", "Font selection dialog"},
		{"systray", "System tray icon and notification"},
		{"print", "Printing from canvas and text widgets"},
	}},
	{"Animation", []demoEntry{
		{"anilabel", "Animated labels"},
		{"aniwave", "Animated wave"},
		{"pendulum", "Pendulum simulation"},
		{"goldberg", "A celebration of Rube Goldberg"},
	}},
	{"Miscellaneous", []demoEntry{
		{"bitmap", "The built-in bitmaps"},
		{"dialog1", "A dialog box with a local grab"},
		{"dialog2", "A dialog box with a global grab"},
		{"windowicons", "Window icons and badges"},
	}},
}

func main() {
	app, err := takigo.NewApp(takigo.Title("Widget Demonstration"), takigo.Size(1024, 600))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	// Find demo base directory.
	_, thisFile, _, _ := runtime.Caller(0)
	demoBase := filepath.Dir(filepath.Dir(thisFile))

	// Header.
	header := label.New(app, "header",
		label.Text("Widget Demonstration"),
		label.Anchor(option.AnchorCenter),
		label.PadX(10), label.PadY(8),
	)
	pack.Pack(header, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// Bottom button frame.
	btnFrame := frame.New(app, "btnframe")
	pack.Pack(btnFrame, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))

	dismissBtn := button.New(btnFrame, "dismiss",
		button.Text("Quit"), button.Command(func() { app.Quit() }),
		button.PadX(10), button.PadY(4),
	)
	pack.Pack(dismissBtn, pack.SideOpt(pack.Right), pack.PadX(10))

	// Status bar.
	statusLabel := label.New(app, "status",
		label.Text("Select a demo and click Run."),
		label.Anchor(option.AnchorW),
		label.Background("#e8e8e8"),
		label.PadX(5), label.PadY(2),
	)
	pack.Pack(statusLabel, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Main content: left listbox + right source viewer.
	contentFrame := frame.New(app, "content")
	pack.Pack(contentFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(5), pack.PadY(2))

	// Left panel: listbox of demos.
	leftFrame := frame.New(contentFrame, "left")
	pack.Pack(leftFrame, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth),
		pack.Expand(false))

	// Action buttons (pack bottom FIRST so listbox doesn't consume all space).
	actionFrame := frame.New(leftFrame, "actions")
	pack.Pack(actionFrame, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(3))

	sb := scrollbar.New(leftFrame, "sb")
	lb := listbox.New(leftFrame, "demos",
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
	pack.Pack(sb, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(lb, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Populate listbox with categories and demos.
	type listItem struct {
		isCategory bool
		demoDir    string
		demoDesc   string
	}
	var items []listItem

	for _, cat := range categories {
		lb.Insert(lb.ItemCount(), fmt.Sprintf("-- %s --", cat.title))
		items = append(items, listItem{isCategory: true})
		for _, d := range cat.demos {
			lb.Insert(lb.ItemCount(), fmt.Sprintf("  %s", d.desc))
			items = append(items, listItem{demoDir: d.name, demoDesc: d.desc})
		}
	}

	// Right panel: source viewer.
	rightFrame := frame.New(contentFrame, "right")
	pack.Pack(rightFrame, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillBoth),
		pack.Expand(true))

	rightLabel := label.New(rightFrame, "srclabel",
		label.Text("Source Code:"),
		label.Anchor(option.AnchorW),
		label.PadX(5),
	)
	pack.Pack(rightLabel, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	srcSb := scrollbar.New(rightFrame, "srcsb")
	srcText := text.New(rightFrame, "source",
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
	pack.Pack(srcSb, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(srcText, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth),
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

	runBtn := button.New(actionFrame, "run",
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
	pack.Pack(runBtn, pack.SideOpt(pack.Left), pack.PadX(3))

	codeBtn := button.New(actionFrame, "code",
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
	pack.Pack(codeBtn, pack.SideOpt(pack.Left), pack.PadX(3))

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
	app.Run()
}

func countDemos() int {
	total := 0
	for _, cat := range categories {
		total += len(cat.demos)
	}
	return total
}
