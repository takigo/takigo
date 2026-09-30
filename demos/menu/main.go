// Demo: Menus and cascaded menus using menubars.
// Ported from Tk's menu.tcl demo.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	tkimage "github.com/msorc/takigo/image"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/menu"
)

// demoImagesDir returns the path to demos/images/.
func demoImagesDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "images")
}

func main() {
	app, err := takigo.NewApp(takigo.Title("Menu Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("menu"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Status bar at the very bottom (packed first so it stays below btns).
	statusBar := frame.New(f, "statusBar")
	statusLabel := label.New(statusBar, "label",
		label.Text("    "),
		label.Anchor(option.AnchorW),
		label.Relief(option.ReliefSunken),
		label.BorderWidth(1),
		label.FontOpt("Helvetica 10"),
	)
	pack.Pack(statusLabel, pack.SideOpt(pack.Left), pack.Expand(true), pack.FillOpt(pack.FillBoth), pack.PadX(2))
	pack.Pack(statusBar, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(2))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// "toplevel $w -menu $w.menu": the menubar clone is named
	// "#menu#menu" after the Tcl menu path.
	menuBar := menu.NewMenubar(app, "#menu#menu")

	msg := label.New(f, "msg",
		label.WrapLength("4i"),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("This window contains a menubar with cascaded menus.  You can post a menu from the keyboard by typing Alt+x, where \"x\" is the character underlined on the menu.  You can then traverse among the menus using the arrow keys.  When a menu is posted, you can invoke the current entry by typing space, or you can invoke any entry by typing its underlined character.  If a menu entry has an accelerator, you can invoke the entry without posting the menu just by typing the accelerator. The rightmost menu can be torn off into a palette by selecting the first item in the menu."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	setStatus := func(s string) {
		statusLabel.Configure(label.Text(s))
	}

	// ── File menu ──
	fileMenu := menu.New(app, "filemenu")
	fileMenu.AddCommand("Open...", func() { setStatus("File > Open") })
	fileMenu.AddCommand("New", func() { setStatus("File > New") })
	fileMenu.AddCommand("Save", func() { setStatus("File > Save") })
	fileMenu.AddCommand("Save As...", func() { setStatus("File > Save As") })
	fileMenu.AddSeparator()
	fileMenu.AddCommand("Print Setup...", func() { setStatus("File > Print Setup") })
	fileMenu.AddCommand("Print...", func() { setStatus("File > Print") })
	fileMenu.AddSeparator()
	fileMenu.AddCommand("Dismiss Menus Demo", func() { app.Quit() })

	menuBar.AddCascade("File", 0, fileMenu)

	// ── Basic menu ──
	basicMenu := menu.New(app, "basicmenu")
	basicMenu.AddCommand("Long entry that does nothing", nil)
	for _, letter := range []string{"A", "B", "C", "D", "E", "F"} {
		l := letter
		// -underline 14: "Print letter \"X\"" → index 14 is the letter character.
		basicMenu.AddCommandAccelUL(fmt.Sprintf("Print letter \"%s\"", l), "Meta+"+l, 14, func() {
			setStatus(fmt.Sprintf("Basic > Print letter \"%s\"", l))
		})
	}

	menuBar.AddCascade("Basic", 0, basicMenu)

	// ── Cascades menu ──
	cascadeMenu := menu.New(app, "cascademenu")

	// -underline 6: "Print hello" → index 6 = 'h'; "Print goodbye" → index 6 = 'g'.
	cascadeMenu.AddCommandAccelUL("Print hello", "Meta+H", 6, func() {
		setStatus("Cascades > Print hello")
	})
	cascadeMenu.AddCommandAccelUL("Print goodbye", "Meta+G", 6, func() {
		setStatus("Cascades > Print goodbye")
	})

	// Check buttons submenu.
	checkMenu := menu.New(app, "checkmenu")
	checkMenu.AddCheckbutton("Oil checked", false, func() { setStatus("Cascades > Check > Oil") })
	checkMenu.AddCheckbutton("Transmission checked", true, func() { setStatus("Cascades > Check > Transmission") })
	checkMenu.AddCheckbutton("Brakes checked", false, func() { setStatus("Cascades > Check > Brakes") })
	checkMenu.AddCheckbutton("Lights checked", true, func() { setStatus("Cascades > Check > Lights") })
	checkMenu.AddSeparator()
	checkMenu.AddCommand("Show current values", func() {
		entries := checkMenu.Entries()
		var msg strings.Builder
		msg.WriteString("Check values:")
		for _, e := range entries {
			if e.Type == menu.Checkbutton {
				val := "off"
				if e.Checked {
					val = "on"
				}
				fmt.Fprintf(&msg, "  %s=%s", e.Label, val)
			}
		}
		setStatus(msg.String())
	})

	cascadeMenu.AddCascadeUL("Check buttons", 0, checkMenu)

	// Radio buttons submenu.
	radioMenu := menu.New(app, "radiomenu")
	radioMenu.AddRadiobutton("10 point", false, func() { setStatus("Cascades > Radio > 10 point") })
	radioMenu.AddRadiobutton("14 point", true, func() { setStatus("Cascades > Radio > 14 point") })
	radioMenu.AddRadiobutton("18 point", false, func() { setStatus("Cascades > Radio > 18 point") })
	radioMenu.AddRadiobutton("24 point", false, func() { setStatus("Cascades > Radio > 24 point") })
	radioMenu.AddRadiobutton("32 point", false, func() { setStatus("Cascades > Radio > 32 point") })
	radioMenu.AddSeparator()
	radioMenu.AddRadiobutton("Roman", false, func() { setStatus("Cascades > Radio > Roman") })
	radioMenu.AddRadiobutton("Bold", true, func() { setStatus("Cascades > Radio > Bold") })
	radioMenu.AddRadiobutton("Italic", false, func() { setStatus("Cascades > Radio > Italic") })
	radioMenu.AddSeparator()
	radioMenu.AddCommand("Show current values", func() {
		entries := radioMenu.Entries()
		var msg strings.Builder
		msg.WriteString("Radio values:")
		for _, e := range entries {
			if e.Type == menu.Radiobutton && e.Checked {
				fmt.Fprintf(&msg, "  selected=%s", e.Label)
			}
		}
		setStatus(msg.String())
	})

	cascadeMenu.AddCascadeUL("Radio buttons", 0, radioMenu)

	menuBar.AddCascade("Cascades", 0, cascadeMenu)

	// ── Icons menu ── (matches Tcl's menu.tcl order: before More)
	iconsMenu := menu.New(app, "iconsmenu")

	// Try to load earthmenu.png for image-only entry.
	earthImg, earthErr := tkimage.NewPhotoFromFile("earthmenu", filepath.Join(demoImagesDir(), "earthmenu.png"))
	if earthErr == nil {
		app.ImageRegistry().Register(earthImg)
		iconsMenu.AddCommandImage("", earthImg, widget.CompoundNone, func() {
			setStatus("Icons > Earth image (image-only entry)")
		})
	} else {
		iconsMenu.AddCommand("(earthmenu.png not found)", nil)
	}

	menuBar.AddCascade("Icons", 0, iconsMenu)

	// ── More menu ──
	moreMenu := menu.New(app, "moremenu")
	for _, item := range []string{
		"An entry",
		"Another entry",
		"Does nothing",
		"Does almost nothing",
		"Does almost nothing also",
		"Make life meaningful",
	} {
		lbl := item
		moreMenu.AddCommand(lbl, func() {
			setStatus(fmt.Sprintf("More > %s", lbl))
		})
	}
	moreMenu.AddCommand("\U0001f60d Make friends", func() {
		setStatus("More > Make friends")
	})
	// Compound entry: image + text (matching Tcl's entryconfigure -image lilearth -compound left).
	if earthErr == nil {
		moreMenu.AddCommandImage("Does almost nothing also (image)", earthImg, widget.CompoundLeft, func() {
			setStatus("More > Does almost nothing also (image)")
		})
	}

	menuBar.AddCascade("More", 0, moreMenu)

	// ── Colors menu ── (tearoff enabled; per-entry colored backgrounds)
	colorsMenu := menu.New(app, "colorsmenu", menu.TearOffOpt(true))
	for _, clr := range []string{"red", "orange", "yellow", "green", "blue"} {
		c := clr
		colorsMenu.AddCommandBg(c, "black", c, func() {
			setStatus(fmt.Sprintf("Colors > %s", c))
		})
	}

	menuBar.AddCascade("Colors", 1, colorsMenu)

	_ = statusLabel
	app.Run()
}
