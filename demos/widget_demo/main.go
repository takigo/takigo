// Widget Demo Launcher — browse and run all takigo demos.
// Ported from Tk's widget demo launcher (widget.tcl).
// Uses a text widget with clickable hyperlinks for demo entries,
// matching the original Tk widget demo's look and behavior.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/dialog"
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/menu"
	"github.com/msorc/takigo/widget/menubutton"
	"github.com/msorc/takigo/widget/scrollbar"
	"github.com/msorc/takigo/widget/text"
)

type demoEntry struct {
	name string
	desc string
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
	app, err := takigo.NewApp(takigo.Title("Widget Demonstration"), takigo.Size(800, 600))
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

	// ── Menu bar ──
	menuBar := frame.New(app, "menubar",
		frame.Relief(option.ReliefRaised),
		frame.BorderWidth(1),
	)
	pack.Pack(menuBar, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	fileMenu := menu.New(app, "filemenu")
	if menuFont, err := app.FontRegistry().Get("Sans 10"); err == nil {
		fileMenu.Font = menuFont
	}
	fileMenu.AddCommandAccel("About...", "F1", func() {
		dialog.ShowMessage(app,
			dialog.MsgTitle("About Widget Demo"),
			dialog.MsgMessage("Tk widget demonstration application"),
			dialog.MsgDetail("A Go port of the Tk widget demo using the Takigo toolkit."),
			dialog.MsgType(dialog.MsgInfo),
			dialog.MsgButtons(dialog.BtnOK),
		)
	})
	fileMenu.AddSeparator()
	fileMenu.AddCommandAccel("Quit", "Meta+Q", func() { app.Quit() })

	fileMb := menubutton.New(menuBar, "filemb",
		menubutton.Text("File"),
		menubutton.MenuOpt(fileMenu),
		menubutton.UnderlineOpt(0),
	)
	pack.Pack(fileMb, pack.SideOpt(pack.Left))

	// ── Status bar with sizegrip ──
	statusBar := frame.New(app, "statusbar")
	pack.Pack(statusBar, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(1))

	statusLabel := label.New(statusBar, "status",
		label.Text("   "),
		label.Anchor(option.AnchorW),
		label.PadX(5), label.PadY(1),
	)
	pack.Pack(statusLabel, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth),
		pack.Expand(true))

	// Sizegrip: small frame that draws diagonal resize lines.
	grip := frame.New(statusBar, "grip",
		frame.Width(15), frame.Height(15),
	)
	pack.Pack(grip, pack.SideOpt(pack.Right), pack.PadX(1))
	drawSizegrip(app, grip, bgColor)

	// ── Text widget with scrollbar ──
	sb := scrollbar.New(app, "sb")
	t := text.New(app, "t",
		text.WrapModeOpt(text.WrapWord),
		text.Width(70), text.Height(30),
		text.FontOpt("Sans 10"),
		text.ReadOnly(true),
	)
	t.YScrollCmd = func(first, last float64) { sb.Set(first, last) }
	sb.Command = func(args ...any) {
		if len(args) >= 2 {
			action, _ := args[0].(string)
			number, _ := args[1].(float64)
			switch action {
			case "moveto":
				t.YViewMoveTo(number)
			case "scroll":
				unit := "units"
				if len(args) >= 3 {
					if u, ok := args[2].(string); ok {
						unit = u
					}
				}
				if unit == "pages" {
					t.YViewScroll(int(number), true)
				} else {
					t.YViewScroll(int(number), false)
				}
			}
		}
	}
	pack.Pack(sb, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(t, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(1))

	// ── Configure tags ──
	t.TagConfigure("title", text.TagFont("Sans Bold 14"))
	t.TagConfigure("subtitle", text.TagFont("Sans Bold 10"))
	t.TagConfigure("bold", text.TagFont("Sans Bold 10"))
	// Demo links: normal weight, blue, underlined.
	t.TagConfigure("demo",
		text.TagForeground("blue"),
		text.TagUnderline(true),
		text.TagFont("Sans 10"),
	)
	t.TagConfigure("hot",
		text.TagForeground("red"),
		text.TagUnderline(true),
	)

	line := 1

	// Title.
	t.Insert("1.0", "Tk Widget Demonstrations\n")
	t.TagAdd("title", "1.0", "1.end")
	line++

	// Blank line.
	t.Insert("end", "\n")
	line++

	// Intro paragraph.
	introLine := line
	intro1 := "This application provides a front end for several short scripts " +
		"that demonstrate what you can do with Tk widgets. Each of the numbered " +
		"lines below describes a demonstration; you can click on it to invoke " +
		"the demonstration. Once the demonstration window appears, you can click the "
	t.Insert("end", intro1)
	t.Insert("end", "See Code")
	intro2 := " button to see the Go code that created the demonstration. " +
		"If you wish, you can edit the code and click the "
	t.Insert("end", intro2)
	t.Insert("end", "Rerun Demo")
	t.Insert("end", " button in the code window to reinvoke the demonstration "+
		"with the modified code.\n")

	// Bold "See Code" and "Rerun Demo".
	seeCodeStart := len(intro1)
	seeCodeEnd := seeCodeStart + len("See Code")
	t.TagAdd("bold",
		fmt.Sprintf("%d.%d", introLine, seeCodeStart),
		fmt.Sprintf("%d.%d", introLine, seeCodeEnd))
	rerunStart := seeCodeEnd + len(intro2)
	rerunEnd := rerunStart + len("Rerun Demo")
	t.TagAdd("bold",
		fmt.Sprintf("%d.%d", introLine, rerunStart),
		fmt.Sprintf("%d.%d", introLine, rerunEnd))
	line++

	// ── Categories and demos ──
	// Indentation spaces before demo number (not part of the link).
	const indent = "      "

	for _, cat := range categories {
		t.Insert("end", "\n")
		line++

		subStart := fmt.Sprintf("%d.0", line)
		t.Insert("end", cat.title+"\n")
		subEnd := fmt.Sprintf("%d.0", line+1)
		t.TagAdd("subtitle", subStart, subEnd)
		line++

		for i, d := range cat.demos {
			demoLine := line
			numStr := fmt.Sprintf("%d. %s", i+1, d.desc)
			t.Insert("end", indent+numStr+"\n")
			line++

			tagName := "demo-" + d.name
			// Tag only the numbered text, not the leading spaces.
			linkStart := fmt.Sprintf("%d.%d", demoLine, len(indent))
			linkEnd := fmt.Sprintf("%d.%d", demoLine, len(indent)+len(numStr))
			t.TagAdd("demo", linkStart, linkEnd)
			t.TagAdd(tagName, linkStart, linkEnd)
			t.TagConfigure(tagName,
				text.TagForeground("blue"),
				text.TagUnderline(true),
				text.TagFont("Sans 10"),
			)

			si := linkStart
			ei := linkEnd
			demoDir := d.name
			demoDesc := d.desc

			// Click → run demo.
			t.TagBind(tagName, "<Button-1>", func() {
				dir := filepath.Join(demoBase, demoDir)
				statusLabel.Text = fmt.Sprintf("Running: %s...", demoDesc)
				statusLabel.Display()
				go func() {
					cmd := exec.Command("go", "run", ".")
					cmd.Dir = dir
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					if err := cmd.Run(); err != nil {
						fmt.Fprintf(os.Stderr, "Demo %s error: %v\n", demoDir, err)
					}
				}()
			})

			// Hover → red highlight + status message.
			t.TagBind(tagName, "<Enter>", func() {
				t.TagAdd("hot", si, ei)
				statusLabel.Text = fmt.Sprintf("Run the \"%s\" sample program", demoDir)
				statusLabel.Display()
				t.Display()
			})
			t.TagBind(tagName, "<Leave>", func() {
				t.TagRemove("hot", si, ei)
				statusLabel.Text = "   "
				statusLabel.Display()
				t.Display()
			})
		}
	}

	// Scroll to top after populating.
	t.See("1.0")

	// ── Root events ──
	app.Dispatcher().Bind(root.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			root.Width = ev.ConfigWidth
			root.Height = ev.ConfigHeight
			pack.ArrangeContainer(root)
		}
	})
	app.Dispatcher().Bind(root.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		d := root.Display.Server
		d.SetForeground(root.GC, bgColor.Pixel)
		d.FillRectangle(root.Drawable(), root.GC, 0, 0, uint(root.Width), uint(root.Height))
		d.Flush()
	})
	app.Dispatcher().BindGlobal(event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == platform.XK_Escape {
			app.Quit()
		}
	})

	app.Run()
}

// drawSizegrip sets up expose handling for a frame that draws diagonal
// resize grip lines, matching Tk's ttkElements.c SizegripDraw algorithm.
func drawSizegrip(app *takigo.App, grip *frame.Frame, bg *color.Color) {
	border := draw.NewBorder(bg.Red, bg.Green, bg.Blue)

	app.Dispatcher().Bind(grip.Window().PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		w := grip.Window()
		d := w.Display.Server
		gc := w.GC

		// Fill background.
		d.SetForeground(gc, bg.Pixel)
		d.FillRectangle(w.Drawable(), gc, 0, 0, uint(w.Width), uint(w.Height))

		// Tk algorithm: gripSize=15, gripCount=3,
		// gripThickness = gripSize*3/(gripCount*5) = 3,
		// gripSpace = gripSize/3 - gripThickness = 2.
		gripCount := 3
		gripSize := w.Height
		if w.Width < gripSize {
			gripSize = w.Width
		}
		gripThickness := gripSize * 3 / (gripCount * 5)
		gripSpace := gripSize/3 - gripThickness

		x1 := w.Width - 1
		y1 := w.Height - 1
		x2 := x1
		y2 := y1

		for g := 0; g < gripCount; g++ {
			x1 -= gripSpace
			y2 -= gripSpace
			for i := 1; i < gripThickness; i++ {
				d.SetForeground(gc, border.DarkPixel)
				d.DrawLine(w.Drawable(), gc, x1, y1, x2, y2)
				x1--
				y2--
			}
			d.SetForeground(gc, border.LightPixel)
			d.DrawLine(w.Drawable(), gc, x1, y1, x2, y2)
			x1--
			y2--
		}
		d.Flush()
	})
}
