// Demo: keyboard and menu-mnemonic diagnostics.
//
// Not a port of a Tk demo. It logs every key press and release the
// dispatcher sees (keysym, keycode, text, modifier bits, whether a class
// handler already consumed it) next to a menubar with Alt+<letter>
// mnemonics, so a platform backend's key delivery can be checked by hand:
// plain letters, Alt/Option combinations, Shift, Control, Command/Super,
// dead keys, non-Latin layouts, and typing in an entry.
package main

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/canvas"
	"github.com/takigo/takigo/demos/demohelper"
	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/entry"
	"github.com/takigo/takigo/widget/frame"
	"github.com/takigo/takigo/widget/label"
	"github.com/takigo/takigo/widget/menu"
	"github.com/takigo/takigo/widget/text"
	"github.com/takigo/takigo/window"
)

// maxLogLines bounds the log so a long session stays cheap to redraw.
const maxLogLines = 400

func main() {
	app, err := takigo.NewApp(takigo.Title("Keyboard and Menu Key Debugging"),
		takigo.Geometry("+300+300"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	menuBar := menu.NewMenubar(app, "menubar")

	msg := label.New(f, "msg",
		label.WrapLength(560),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("Press keys anywhere in this window; every event is logged below "+
			"(and to stdout). Try Alt+F, Alt+E, Alt+V, Alt+T (Option on macOS) to post "+
			"the menus, plain letters such as D and S, Shift/Control/Command "+
			"combinations, and typing in the entry. Menu commands log a "+
			"\"menu:\" line, so a mnemonic that worked is easy to tell from one "+
			"that did not."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadX(6), pack.PadY(4))

	entryRow := frame.New(f, "entryrow")
	entryLabel := label.New(entryRow, "l", label.Text("Entry (also gets focus):"))
	ent := entry.New(entryRow, "e", entry.Width(30))
	pack.Pack(entryLabel, pack.SideOpt(pack.Left), pack.PadX(6))
	pack.Pack(ent, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillX), pack.Expand(true), pack.PadX(6))
	pack.Pack(entryRow, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(4))

	// A focusable canvas: the sketch view in TeleSAPR is one, and plain
	// letters (D, S) must reach it. Click it or Tab to it for focus; the
	// highlight ring shows when it has it.
	cv := canvas.New(f, "canvas",
		canvas.Width(560),
		canvas.Height(110),
		canvas.Background("white"),
		canvas.HighlightWidthOpt(2),
	)
	cv.Win.Flags |= window.FlagFocusable
	cv.CreateRectangle(30, 25, 130, 85, canvas.FillColor("#9ecbff"), canvas.OutlineColor("black"))
	cv.CreateOval(170, 25, 270, 85, canvas.FillColor("#ffd27f"), canvas.OutlineColor("black"))
	cv.CreateLine([]float64{310, 85, 360, 25, 410, 85}, canvas.OutlineColor("#c0392b"), canvas.OutlineWidth(3))
	cv.CreatePolygon([]float64{450, 85, 500, 25, 540, 85}, canvas.FillColor("#a8e6a1"), canvas.OutlineColor("black"))
	cv.CreateText(280, 100, canvas.TextOpt("canvas: click or Tab here, then press D, S, Alt+letter"))
	pack.Pack(cv, pack.SideOpt(pack.Top), pack.PadX(6), pack.PadY(4))

	counts := label.New(f, "counts",
		label.Anchor(option.AnchorW),
		label.Relief(option.ReliefSunken),
		label.BorderWidth(1),
	)
	pack.Pack(counts, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadX(6), pack.PadY(2))

	logw := text.New(f, "log",
		text.Width(80),
		text.Height(18),
		text.WrapModeOpt(text.WrapNone),
	)
	pack.Pack(logw, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true), pack.PadX(6), pack.PadY(4))

	var presses, releases, menuHits, lines int
	updateCounts := func() {
		counts.Configure(label.Text(fmt.Sprintf(
			"key presses: %d   releases: %d   menu commands: %d", presses, releases, menuHits)))
	}
	updateCounts()

	logf := func(format string, args ...any) {
		line := fmt.Sprintf(format, args...)
		fmt.Println(line)
		if lines >= maxLogLines {
			_ = logw.Delete("1.0", "2.0")
			lines--
		}
		_ = logw.Insert("end", line+"\n")
		_ = logw.See("end")
		lines++
	}

	// Menus. The mnemonic letters F, E, V and T include E, which is a
	// dead key (Option+E) on a US macOS layout.
	menuHit := func(path string) func() {
		return func() {
			menuHits++
			updateCounts()
			logf("menu: %s", path)
		}
	}
	fileMenu := menu.New(app, "filemenu")
	fileMenu.AddCommand("Open...", menuHit("File > Open"))
	fileMenu.AddCommand("Save", menuHit("File > Save"))
	fileMenu.AddSeparator()
	fileMenu.AddCommand("Quit", func() { app.Quit() })
	menuBar.AddCascade("File", 0, fileMenu)

	editMenu := menu.New(app, "editmenu")
	editMenu.AddCommand("Undo", menuHit("Edit > Undo"))
	editMenu.AddCommand("Redo", menuHit("Edit > Redo"))
	menuBar.AddCascade("Edit", 0, editMenu)

	viewMenu := menu.New(app, "viewmenu")
	viewMenu.AddCommand("Zoom in", menuHit("View > Zoom in"))
	viewMenu.AddCommand("Zoom out", menuHit("View > Zoom out"))
	menuBar.AddCascade("View", 0, viewMenu)

	testMenu := menu.New(app, "testmenu")
	for _, l := range []string{"A", "B", "C", "D", "S"} {
		testMenu.AddCommandAccelUL(fmt.Sprintf("Letter %s", l), "Meta+"+l, 7, menuHit("Test > Letter "+l))
	}
	menuBar.AddCascade("Test", 0, testMenu)

	clearLog := func() {
		_ = logw.Delete("1.0", "end")
		lines = 0
		presses, releases, menuHits = 0, 0, 0
		updateCounts()
	}
	app.Dispatcher().BindGlobal(event.KeyPressMask|event.KeyReleaseMask, func(ev *event.Event) {
		kind := "press  "
		if ev.Type == event.KeyReleaseType {
			kind = "release"
			releases++
		} else {
			presses++
		}
		updateCounts()
		logf("%s keysym=%#06x %-10s code=%-4d str=%-8q mods=%-24s win=%v handled=%v",
			kind, uint(ev.KeySym), keysymRune(ev.KeySym), ev.KeyCode, ev.Str,
			modNames(ev.State), ev.Window, ev.Handled)
		// Escape clears nothing: it is a key worth seeing. Ctrl+L clears.
		if ev.Type == event.KeyPressType && ev.State&platform.ControlMask != 0 && ev.KeySym == platform.KeySym('l') {
			clearLog()
		}
	})

	widget.Focus(app, ent.Window())
	app.Dispatcher().Bind(cv.Win.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		widget.Focus(app, cv.Win)
		logf("canvas: click at %d,%d, focus requested", ev.X, ev.Y)
	})
	app.Dispatcher().Bind(cv.Win.PlatformID, event.FocusChangeMask, func(ev *event.Event) {
		logf("canvas: focus %v", ev.Type == event.FocusInType)
		cv.Display()
	})
	logf("ready. Ctrl+L clears the log. Press keys.")
	app.Run()
}

// modNames lists the modifier bits set in a key event's State. Which of
// Mod1..Mod5 is Alt, Option or Command depends on the backend, so the raw
// names are printed.
func modNames(state uint) string {
	var names []string
	for _, m := range []struct {
		mask uint
		name string
	}{
		{platform.ShiftMask, "Shift"},
		{platform.LockMask, "Lock"},
		{platform.ControlMask, "Control"},
		{platform.Mod1Mask, "Mod1"},
		{platform.Mod2Mask, "Mod2"},
		{platform.Mod3Mask, "Mod3"},
		{platform.Mod4Mask, "Mod4"},
		{platform.Mod5Mask, "Mod5"},
	} {
		if state&m.mask != 0 {
			names = append(names, m.name)
		}
	}
	if len(names) == 0 {
		return "-"
	}
	return strings.Join(names, "+")
}

// keysymRune is the printable rune a keysym stands for, or "" when it is
// not one. Latin-1 keysyms equal their code point and 0x01000000|cp is
// Unicode, as in X11.
func keysymRune(k platform.KeySym) string {
	cp := rune(k)
	if k&0x01000000 != 0 {
		cp = rune(k &^ 0x01000000)
	}
	if k < 0x20 || (k >= 0x7f && k < 0xa0) || (k > 0xff && k&0x01000000 == 0) || !unicode.IsPrint(cp) {
		return ""
	}
	return string(cp)
}
