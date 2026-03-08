// Widget Demo Launcher — browse and run all takigo demos.
// Ported from Tk's widget demo launcher (tk/library/demos/widget).
// Uses a text widget with clickable hyperlinks for demo entries,
// matching the original Tk widget demo's look and behavior.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/cursor"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/menu"
	"github.com/msorc/takigo/widget/menubutton"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	// wm title . "Widget Demonstration"
	app, err := takigo.NewApp(takigo.Title("Widget Demonstration"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	// TODO: wm iconwindow . [toplevel ._iconWindow]
	// TODO: wm iconname . "tkWidgetDemo"

	// Find demo base directory.
	// set tk_demoDirectory [file join [pwd] [file dirname [info script]]]
	_, thisFile, _, _ := runtime.Caller(0)
	demoBase := filepath.Dir(filepath.Dir(thisFile))

	// ── Fonts ──
	// In Tk, named fonts (mainFont, fixedFont, boldFont, titleFont, statusFont,
	// varsFont) are created from TkDefaultFont/TkFixedFont. In Go, we use
	// font specification strings directly where needed.
	// set font mainFont

	// ── Menu bar ──
	// TODO: proper menu bar (menu .menuBar / . configure -menu .menuBar)
	// Currently using frame + menubutton as a workaround.
	// menu .menuBar -tearoff 0
	menuBar := frame.New(app, "menuBar",
		frame.Relief(option.ReliefRaised),
		frame.BorderWidth(1),
	)
	pack.Pack(menuBar, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// menu .menuBar.file -tearoff 0
	fileMenu := menu.New(app, "file")
	if menuFont, err := app.FontRegistry().Get("Sans 10"); err == nil {
		fileMenu.Font = menuFont
	}
	// .menuBar.file add command -label "About..." -accelerator "<F1>"
	fileMenu.AddCommandAccel("About...", "F1", func() {
		// TODO: tkAboutDialog — use dialog.ShowMessage
		fmt.Println("About: Tk widget demonstration application")
	})
	// .menuBar.file add sep
	fileMenu.AddSeparator()
	// .menuBar.file add command -label "Quit" -accelerator "Meta-Q"
	fileMenu.AddCommandAccel("Quit", "Meta+Q", func() { app.Quit() })

	fileMb := menubutton.New(menuBar, "filemb",
		menubutton.Text("File"),
		menubutton.MenuOpt(fileMenu),
		menubutton.UnderlineOpt(0),
	)
	pack.Pack(fileMb, pack.SideOpt(pack.Left))

	// ── Status bar ──
	// ttk::frame .statusBar
	statusBar := frame.New(app, "statusBar")
	// ttk::label .statusBar.lab -text "   " -anchor w
	statusLabel := label.New(statusBar, "lab",
		label.Text("   "),
		label.Anchor(option.AnchorW),
	)
	// pack .statusBar.lab -side left -padx 1.5p -expand yes -fill both
	pack.Pack(statusLabel, pack.SideOpt(pack.Left), pack.PadX("1.5p"),
		pack.Expand(true), pack.FillOpt(pack.FillBoth))

	// ttk::sizegrip .statusBar.foo
	grip := ttk.NewSizegrip(statusBar, "foo")
	// pack .statusBar.foo -side right -padx 1.5p
	pack.Pack(grip, pack.SideOpt(pack.Right), pack.PadX("1.5p"))

	// pack .statusBar -side bottom -fill x -pady 1.5p
	pack.Pack(statusBar, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX),
		pack.PadY("1.5p"))

	// ── Text widget with scrollbar ──
	// ttk::frame .textFrame
	textFrame := frame.New(app, "textFrame")

	// ttk::scrollbar .s -orient vertical -command {.t yview} -takefocus 1
	s := ttk.NewScrollbar(textFrame, "s",
		ttk.ScrollbarOrientOpt(ttk.Vertical),
	)
	// pack .s -in .textFrame -side right -fill y
	pack.Pack(s, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))

	// set textheight 30
	// catch { set textheight [expr {
	//     ([winfo screenheight .] * 0.7) /
	//     [font metrics mainFont -displayof . -linespace]
	// }]}
	textHeight := 30
	mainFont, err := app.FontRegistry().Get("Sans 10")
	if err == nil {
		screen := app.Root().Display.Screen
		screenH := app.Server().ScreenHeight(screen)
		linespace := mainFont.Metrics().Linespace()
		if linespace > 0 {
			textHeight = (screenH * 7 / 10) / linespace
		}
	}

	// text .t -yscrollcommand {.s set} -wrap word -width 70 -height $textheight
	//     -font mainFont -setgrid 1 -highlightthickness 0
	//     -padx 3p -pady 1.5p -takefocus 0
	t := text.New(textFrame, "t",
		text.WrapModeOpt(text.WrapWord),
		text.Width(70), text.Height(textHeight),
		text.FontOpt("Sans 10"),
		// TODO: text.SetGrid(true), text.HighlightThickness(0),
		// TODO: text.TakeFocus(false)
	)

	// Wire scrollbar ↔ text.
	t.YScrollCmd = func(first, last float64) { s.Set(first, last) }
	s.Command = func(args ...any) {
		if len(args) < 2 {
			return
		}
		action, _ := args[0].(string)
		switch action {
		case "moveto":
			if f, ok := args[1].(float64); ok {
				t.YViewMoveTo(f)
			}
		case "scroll":
			n, _ := args[1].(int)
			unit := "units"
			if len(args) >= 3 {
				if u, ok := args[2].(string); ok {
					unit = u
				}
			}
			t.YViewScroll(n, unit == "pages")
		}
	}

	// pack .t -in .textFrame -expand y -fill both -padx 1
	pack.Pack(t, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(1))
	// pack .textFrame -expand yes -fill both
	pack.Pack(textFrame, pack.Expand(true), pack.FillOpt(pack.FillBoth))

	// ── Configure tags ──
	// .t tag configure title -font titleFont
	t.TagConfigure("title", text.TagFont("Sans Bold 14"))
	// .t tag configure subtitle -font titleFont
	t.TagConfigure("subtitle", text.TagFont("Sans Bold 10"))
	// .t tag configure bold -font boldFont
	t.TagConfigure("bold", text.TagFont("Sans Bold 10"))

	// .t tag configure demospace -lmargin1 1c -lmargin2 1c
	t.TagConfigure("demospace",
		text.TagLMargin1Str("1c"),
		text.TagLMargin2Str("1c"),
	)

	// .t tag configure demo -lmargin1 1c -lmargin2 1c -foreground blue -underline 1
	t.TagConfigure("demo",
		text.TagLMargin1Str("1c"),
		text.TagLMargin2Str("1c"),
		text.TagForeground("blue"),
		text.TagUnderline(true),
	)
	// .t tag configure visited -lmargin1 1c -lmargin2 1c -foreground #303080 -underline 1
	t.TagConfigure("visited",
		text.TagLMargin1Str("1c"),
		text.TagLMargin2Str("1c"),
		text.TagForeground("#303080"),
		text.TagUnderline(true),
	)
	// .t tag configure hot -foreground red -underline 1
	t.TagConfigure("hot",
		text.TagForeground("red"),
		text.TagUnderline(true),
	)

	// ── Populate content ──
	// Matches Tcl's addFormattedText content structure.
	addFormattedText(t, formattedContent)

	// ── Tag bindings ──
	// In Tcl, bindings are on the generic "demo" tag with @%x,%y index lookup
	// to find which demo-* tag is at the cursor. We don't have @%x,%y support,
	// so we bind per demo-* tag individually.
	for _, cat := range categories {
		for _, d := range cat.demos {
			demoDir := d.name
			demoDesc := d.desc
			tagName := "demo-" + d.name

			// .t tag bind demo <ButtonRelease-1> { invoke [.t index {@%x,%y}] }
			t.TagBind(tagName, "<Button-1>", func() {
				dir := filepath.Join(demoBase, demoDir)
				// showStatus
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

			// .t tag bind demo <Enter> { ... .t config -cursor hand2 ... showStatus }
			t.TagBind(tagName, "<Enter>", func() {
				t.Window().SetCursor(cursor.Hand2)
				statusLabel.Text = fmt.Sprintf("Run the \"%s\" sample program", demoDir)
				statusLabel.Display()
			})
			// .t tag bind demo <Leave> { ... .t config -cursor xterm ... }
			t.TagBind(tagName, "<Leave>", func() {
				t.Window().SetCursor(cursor.XTerm)
				statusLabel.Text = "   "
				statusLabel.Display()
			})
		}
	}

	// .t configure -state disabled
	text.ReadOnly(true)(t)

	// Scroll to top.
	t.See("1.0")

	// focus .s
	// TODO: focus scrollbar

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

// ── Content data ──
// Matches the Tcl addFormattedText format:
//
//	@@title    Heading text
//	@@subtitle Section heading
//	@@demo name  Description text
//	@@bold / @@normal  Style switches
//	@@newline  Insert newline
//
// Plain text lines are concatenated with spaces.

type demoEntry struct {
	name string
	desc string
}

type category struct {
	title string
	demos []demoEntry
}

var categories = []category{
	// @@subtitle Labels, buttons, checkbuttons, and radiobuttons
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
	// @@subtitle Listboxes and Trees
	{"Listboxes and Trees", []demoEntry{
		{"states", "The 50 states"},
		{"colors", "Colors: change the color scheme for the application"},
		{"sayings", "A collection of famous and infamous sayings"},
		{"mclist", "A multi-column list of countries"},
		{"tree", "A directory browser tree"},
	}},
	// @@subtitle Entries, Spin-boxes and Combo-boxes
	{"Entries, Spin-boxes and Combo-boxes", []demoEntry{
		{"entry1", "Entries without scrollbars"},
		{"entry2", "Entries with scrollbars"},
		{"entry3", "Validated entries and password fields"},
		{"spin", "Spin-boxes"},
		{"ttkspin", "Themed spin-boxes"},
		{"combo", "Combo-boxes"},
		{"form", "Simple Rolodex-like form"},
	}},
	// @@subtitle Text
	{"Text", []demoEntry{
		{"text", "Basic editable text"},
		{"style", "Text display styles"},
		{"bind", "Hypertext (tag bindings)"},
		{"twind", "A text widget with embedded windows and other features"},
		{"search", "A search tool built with a text widget"},
		{"textpeer", "Peering text widgets"},
	}},
	// @@subtitle Canvases
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
	// @@subtitle Scales and Progress Bars
	{"Scales and Progress Bars", []demoEntry{
		{"hscale", "Horizontal scale"},
		{"vscale", "Vertical scale"},
		{"ttkscale", "Themed scale linked to a label with traces"},
		{"ttkprogress", "Progress bar"},
	}},
	// @@subtitle Paned Windows and Notebooks
	{"Paned Windows and Notebooks", []demoEntry{
		{"paned1", "Horizontal paned window"},
		{"paned2", "Vertical paned window"},
		{"ttkpane", "Themed nested panes"},
		{"ttknote", "Notebook widget"},
	}},
	// @@subtitle Menus and Toolbars
	{"Menus and Toolbars", []demoEntry{
		{"menu", "Menus and cascades (sub-menus)"},
		{"menubu", "Menu-buttons"},
		{"ttkmenu", "Themed menu buttons"},
		{"toolbar", "Themed toolbar"},
	}},
	// @@subtitle Common Dialogs
	{"Common Dialogs", []demoEntry{
		{"msgbox", "Message boxes"},
		{"filebox", "File selection dialog"},
		{"clrpick", "Color picker"},
		{"fontchoose", "Font selection dialog"},
		{"systray", "System tray icon and notification"},
		// TODO: {"print", "Printing from canvas and text widgets"},
	}},
	// @@subtitle Animation
	{"Animation", []demoEntry{
		{"anilabel", "Animated labels"},
		{"aniwave", "Animated wave"},
		{"pendulum", "Pendulum simulation"},
		{"goldberg", "A celebration of Rube Goldberg"},
	}},
	// @@subtitle Miscellaneous
	{"Miscellaneous", []demoEntry{
		{"bitmap", "The built-in bitmaps"},
		{"dialog1", "A dialog box with a local grab"},
		{"dialog2", "A dialog box with a global grab"},
		{"windowicons", "Window icons and badges"},
		{"msgwidget", "Message widget with aspect-ratio wrapping"},
		{"square", "A draggable square widget"},
	}},
}

// formattedContent matches Tcl's addFormattedText { ... } block.
// The format is: @@directive args, or plain text lines.
const formattedContent = `
@@title	Tk Widget Demonstrations

This application provides a front end for several short scripts
that demonstrate what you can do with Tk widgets.  Each of the
numbered lines below describes a demonstration; you can click on
it to invoke the demonstration.  Once the demonstration window
appears, you can click the
@@bold
See Code
@@normal
button to see the Go code that created the demonstration.
@@newline

@@subtitle	Labels, buttons, checkbuttons, and radiobuttons
@@demo label	Labels (text and bitmaps)
@@demo unicodeout	Labels and UNICODE text
@@demo button	Buttons
@@demo check	Check-buttons (select any of a group)
@@demo radio	Radio-buttons (select one of a group)
@@demo puzzle	A 15-puzzle game made out of buttons
@@demo icon	Iconic buttons that use bitmaps
@@demo image1	Two labels displaying images
@@demo image2	A simple user interface for viewing images
@@demo labelframe	Labelled frames
@@demo ttkbut	The simple Themed Tk widgets

@@subtitle	Listboxes and Trees
@@demo states	The 50 states
@@demo colors	Colors: change the color scheme for the application
@@demo sayings	A collection of famous and infamous sayings
@@demo mclist	A multi-column list of countries
@@demo tree	A directory browser tree

@@subtitle	Entries, Spin-boxes and Combo-boxes
@@demo entry1	Entries without scrollbars
@@demo entry2	Entries with scrollbars
@@demo entry3	Validated entries and password fields
@@demo spin	Spin-boxes
@@demo ttkspin	Themed spin-boxes
@@demo combo	Combo-boxes
@@demo form	Simple Rolodex-like form

@@subtitle	Text
@@demo text	Basic editable text
@@demo style	Text display styles
@@demo bind	Hypertext (tag bindings)
@@demo twind	A text widget with embedded windows and other features
@@demo search	A search tool built with a text widget
@@demo textpeer	Peering text widgets

@@subtitle	Canvases
@@demo items	The canvas item types
@@demo plot	A simple 2-D plot
@@demo ctext	Text items in canvases
@@demo arrow	An editor for arrowheads on canvas lines
@@demo ruler	A ruler with adjustable tab stops
@@demo floor	A building floor plan
@@demo cscroll	A simple scrollable canvas
@@demo knightstour	A Knight's tour of the chess board

@@subtitle	Scales and Progress Bars
@@demo hscale	Horizontal scale
@@demo vscale	Vertical scale
@@demo ttkscale	Themed scale linked to a label with traces
@@demo ttkprogress	Progress bar

@@subtitle	Paned Windows and Notebooks
@@demo paned1	Horizontal paned window
@@demo paned2	Vertical paned window
@@demo ttkpane	Themed nested panes
@@demo ttknote	Notebook widget

@@subtitle	Menus and Toolbars
@@demo menu	Menus and cascades (sub-menus)
@@demo menubu	Menu-buttons
@@demo ttkmenu	Themed menu buttons
@@demo toolbar	Themed toolbar

@@subtitle	Common Dialogs
@@demo msgbox	Message boxes
@@demo filebox	File selection dialog
@@demo clrpick	Color picker
@@demo fontchoose	Font selection dialog
@@demo systray	System tray icon and notification

@@subtitle	Animation
@@demo anilabel	Animated labels
@@demo aniwave	Animated wave
@@demo pendulum	Pendulum simulation
@@demo goldberg	A celebration of Rube Goldberg

@@subtitle	Miscellaneous
@@demo bitmap	The built-in bitmaps
@@demo dialog1	A dialog box with a local grab
@@demo dialog2	A dialog box with a global grab
@@demo windowicons	Window icons and badges
@@demo msgwidget	Message widget with aspect-ratio wrapping
@@demo square	A draggable square widget
`

// addFormattedText populates a text widget from the formatted content string,
// matching Tcl's addFormattedText proc. Directives:
//
//	@@title text     → insert title with "title" tag
//	@@subtitle text  → insert subtitle with "subtitle" tag + demospace
//	@@demo name desc → insert numbered demo link with "demo" + "demo-name" tags
//	@@bold           → switch style to "bold"
//	@@normal         → switch style to "normal" (no tag)
//	@@newline        → insert newline
//	plain text       → insert with current style
func addFormattedText(t *text.TextWidget, content string) {
	style := "normal"
	isNL := true
	demoCount := 0

	// insertTagged inserts text at "end" and applies the given tags.
	// Matches Tcl's `.t insert end "text" tagList` pattern.
	insertTagged := func(s string, tags ...string) {
		start := t.EndIndex()
		t.Insert("end", s)
		if len(tags) > 0 {
			end := t.EndIndex()
			for _, tag := range tags {
				t.TagAdd(tag, start, end)
			}
		}
	}

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "@@") {
			data := line[2:]
			parts := strings.SplitN(data, "\t", 2)
			key := strings.TrimSpace(parts[0])
			value := ""
			if len(parts) > 1 {
				value = strings.TrimSpace(parts[1])
			}

			switch {
			case key == "title":
				// .t insert end $title\n title \n normal
				insertTagged(value+"\n", "title")
				insertTagged("\n")

			case key == "subtitle":
				// .t insert end "\n" {} $subtitle subtitle " \n " demospace
				insertTagged("\n")
				insertTagged(value, "subtitle")
				insertTagged(" \n ", "demospace")
				demoCount = 0

			case strings.HasPrefix(key, "demo"):
				// @@demo name	description
				nameParts := strings.Fields(key)
				var demoName, description string
				if len(nameParts) >= 2 {
					demoName = nameParts[1]
					description = value
				} else {
					vp := strings.SplitN(value, "\t", 2)
					demoName = strings.TrimSpace(vp[0])
					if len(vp) > 1 {
						description = strings.TrimSpace(vp[1])
					}
				}
				demoCount++
				demoText := fmt.Sprintf("%d. %s", demoCount, description)

				// .t insert end "N. desc" {demo demo-name}
				insertTagged(demoText, "demo", "demo-"+demoName)
				// .t insert end " \n " demospace
				insertTagged(" \n ", "demospace")

			case key == "newline":
				// .t insert end \n $style
				if style != "normal" {
					insertTagged("\n", style)
				} else {
					insertTagged("\n")
				}
				isNL = true

			case key == "bold":
				style = "bold"
			case key == "normal":
				style = "normal"
			}
			continue
		}

		// Plain text line — insert with current style tag.
		// In Tcl: .t insert end " " $style (space) then .t insert end $line $style
		if !isNL {
			if style != "normal" {
				insertTagged(" ", style)
			} else {
				t.Insert("end", " ")
			}
		}
		isNL = false
		if style != "normal" {
			insertTagged(line, style)
		} else {
			t.Insert("end", line)
		}
	}
}

