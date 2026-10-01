# The takigo Tutorial

takigo is a pure-Go port of the Tk GUI toolkit. If you have used Tk from Tcl,
Python (`tkinter`) or Perl, most ideas carry straight over: the same widgets,
the same geometry managers (`pack`, `grid`, `place`), the same binding
language (`<Control-s>`, `<Double-Button-1>`, `<<Copy>>`), the same themed
`ttk` widgets. What changes is the surface: instead of commands and strings,
you get Go packages, typed constructors and functional options.

This tutorial starts from an empty directory and ends with a complete,
persistent to-do application. Along the way it covers every major part of the
toolkit. Each chapter builds on the previous ones, but the reference-style
sections (widget tour, binding patterns, canvas items) can be read on their
own.

All complete programs in this tutorial compile against the current tree and
were run under Xvfb while it was written.

---

## Contents

1. [Setting up](#1-setting-up)
2. [Your first window](#2-your-first-window)
3. [Core concepts](#3-core-concepts)
4. [Geometry management](#4-geometry-management)
5. [A tour of the classic widgets](#5-a-tour-of-the-classic-widgets)
6. [Variables](#6-variables)
7. [Events and bindings](#7-events-and-bindings)
8. [Scrolling](#8-scrolling)
9. [Menus](#9-menus)
10. [Toplevel windows and the window manager](#10-toplevel-windows-and-the-window-manager)
11. [Dialogs](#11-dialogs)
12. [The text widget](#12-the-text-widget)
13. [The canvas](#13-the-canvas)
14. [Images](#14-images)
15. [Timers, idle work and goroutines](#15-timers-idle-work-and-goroutines)
16. [Themed widgets (ttk)](#16-themed-widgets-ttk)
17. [Putting it together: a to-do app](#17-putting-it-together-a-to-do-app)
18. [Testing GUI code](#18-testing-gui-code)
19. [Debugging tips](#19-debugging-tips)
20. [Tcl/Tk → takigo cheat sheet](#20-tcltk--takigo-cheat-sheet)
21. [Where to go next](#21-where-to-go-next)

---

## 1. Setting up

### Requirements

- **Go 1.27** or newer. The module's `go` directive is `1.27.0`; with an older
  local Go, `GOTOOLCHAIN=auto` downloads the right toolchain automatically.
- **Linux / BSD (X11):** a C compiler and the X11, Xft and fontconfig headers.
  On Debian/Ubuntu:

  ```bash
  sudo apt install build-essential libx11-dev libxft-dev libfontconfig1-dev
  ```

- **macOS:** Xcode command-line tools (the Cocoa backend uses cgo).
- **Windows:** nothing extra; the Win32 backend is pure Go.

takigo has **no third-party Go dependencies** — only the standard library.

### A new project

```bash
mkdir hello && cd hello
go mod init example.com/hello
go get github.com/msorc/takigo
```

### Running without a desktop

Everything in this tutorial also runs on a headless machine under a virtual X
server, which is how the examples were checked:

```bash
xvfb-run -a -s "-screen 0 1280x1024x24 -noreset" go run .
```

The `-noreset` flag matters when you start several programs in a row: without
it Xvfb resets whenever its last client disconnects and briefly refuses new
connections.

---

## 2. Your first window

Create `main.go`:

```go
package main

import (
	"fmt"
	"log"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Hello"))
	if err != nil {
		log.Fatal(err)
	}

	greeting := label.New(app, "greeting", label.Text("Hello, takigo!"))
	pack.Pack(greeting, pack.PadX(20), pack.PadY(10))

	count := 0
	btn := button.New(app, "btn",
		button.Text("Click me"),
		button.Command(func() {
			count++
			greeting.Configure(label.Text(fmt.Sprintf("Clicked %d times", count)))
		}),
	)
	pack.Pack(btn, pack.PadY(10))

	quit := button.New(app, "quit", button.Text("Quit"), button.Command(app.Quit))
	pack.Pack(quit, pack.PadY(10))

	app.Run()
}
```

Run it with `go run .`. A small window appears with a label and two buttons;
the window shrinks to fit its contents, exactly as a Tk window does.

Line by line:

- **`takigo.NewApp(opts...)`** opens the display connection, creates the root
  window (Tk's `.`), and sets up the event loop, colour cache, font registry,
  image registry, binding engine, focus manager and selection manager. Useful
  options: `takigo.Title`, `takigo.Geometry("800x600+100+50")`,
  `takigo.Size(w, h)`, `takigo.IconName`, `takigo.DisplayName(":1")`.
- **`label.New(app, "greeting", ...)`** creates a label whose parent is the
  root window. Every widget constructor has the same shape:
  `New(parent, name, options...)`.
- **`pack.Pack(widget, ...)`** hands the widget to the packer. A widget that
  no geometry manager manages is never shown — the most common beginner
  surprise, in Go as in Tcl.
- **`button.Command(func() {...})`** is the Go equivalent of `-command`.
- **`greeting.Configure(...)`** changes options after creation, like
  `.greeting configure -text ...`.
- **`app.Run()`** maps the window, runs the event loop until `app.Quit()` is
  called, then releases all resources. It blocks. If you need to do something
  between the loop ending and cleanup, call `app.MainLoop()` and then
  `app.Destroy()` yourself.

Closing the window with the window manager's close button also ends the loop.

---

## 3. Core concepts

### 3.1 The widget tree

Widgets form a tree rooted at the application window. Each has a Tk-style
*path name* built from its parent's path and its own name:

```go
f := frame.New(app, "toolbar")          // path ".toolbar"
b := button.New(f, "save", ...)         // path ".toolbar.save"
fmt.Println(b.Window().PathName)        // ".toolbar.save"
```

Names must be unique among siblings. The path is what you use when you bind
events to one specific widget (chapter 7).

The first argument of every constructor is a `widget.Caregiver` — anything
that can parent a widget: the `*takigo.App` itself or any widget. Every
widget exposes its underlying `*window.Window` through `Window()`, which holds
geometry (`Width`, `Height`, `ReqWidth`, `ReqHeight`), the path name, the
class name, and the hierarchy (`Parent`, `Children`).

To remove a widget and all its descendants, call its `Destroy()` method.

### 3.2 Functional options

Tk configures widgets with `-option value` pairs. takigo uses functional
options, one Go function per Tk option:

```go
b := button.New(parent, "ok",
	button.Text("OK"),          // -text
	button.Width(10),           // -width
	button.Background("gray85"),// -background
	button.Command(onOK),       // -command
)
```

Conventions that apply across the classic widget packages:

- The bare Tk option name is used when it is available: `label.Text`,
  `label.Width`, `label.Relief`, `frame.Background`.
- An `Opt` suffix appears when the bare name would clash with something else
  in the package: `label.FontOpt`, `label.ImageOpt`, `label.JustifyOpt`,
  `scale.FromOpt`, `menu.TearOffOpt`. When in doubt, check the package with
  `go doc github.com/msorc/takigo/widget/label`.
- Every classic package also exports **prefixed aliases**, so code reads the
  same whether you use classic or themed widgets: `button.Text`,
  `label.Text`, `entry.Text`, …
- In the `ttk` package, where all themed widgets live together, options are
  always prefixed: `ttk.ButtonText`, `ttk.EntryWidth`, `ttk.LabelFont`.

**Changing options later.** Every widget has a typed `Configure` method that
accepts the same options as its constructor:

```go
b.Configure(button.Text("Cancel"), button.Command(onCancel))
```

`Configure` does everything Tk does after a `configure`: recomputes the
requested size, asks the geometry manager for a new layout when that size
changed, updates the window background, and schedules a redraw. Always go
through it; don't assign to exported fields and hope for a redraw. Some
widgets add convenience wrappers such as `button.SetText` or
`ttk.Label.SetText`, which call `Configure` for you.

**Bad values are errors, not panics.** A distance is checked by the compiler.
A colour or font is looked up when the option runs; if the lookup fails the
widget keeps its previous value, the other options still apply, and
`Configure` returns the error:

```go
if err := ok.Configure(button.Background("grren")); errors.Is(err, color.ErrUnknown) {
	// color: unknown color "grren"
}
```

A constructor has no error to return, so there the failure goes to the App's
logger instead — `slog.Default()` unless you pass `takigo.WithLogger`:

```
level=WARN msg="option not applied" widget=.ok err="color: unknown color \"grren\""
```

Colours and fonts can also be given as values rather than names:
`button.Background(color.RGB(0, 128, 0))`,
`label.FontOpt(font.Attributes{Family: "Helvetica", Size: 12})`.

### 3.3 Colours, fonts and distances

**Colours** are strings, as in Tk: X11 colour names (`"steel blue"`,
`"SeaGreen2"`, `"gray85"`) or hex values (`"#3a6ea5"`).

**Fonts** are strings too, in any of Tk's forms:

```go
label.FontOpt("Helvetica 14 bold")           // family size style...
label.FontOpt("{DejaVu Sans Mono} 11")
label.FontOpt("-family Times -size 12 -slant italic")
label.FontOpt("TkFixedFont")                 // a named font
```

The standard Tk named fonts exist as constants in the `font` package:
`font.TkDefaultFont`, `font.TkTextFont`, `font.TkFixedFont`,
`font.TkMenuFont`, `font.TkHeadingFont`, `font.TkCaptionFont`,
`font.TkSmallCaptionFont`, `font.TkIconFont`, `font.TkTooltipFont`.

You can define your own named fonts and derive variants:

```go
reg := app.FontRegistry()
reg.Define("Heading", font.Attributes{Family: "Helvetica", Size: 16, Weight: font.WeightBold})
pack.Pack(label.New(app, "h", label.Text("Title"), label.FontOpt("Heading")))

bigger := reg.Derive(font.TkDefaultFont, 14, font.WeightBold) // returns a descriptor string
pack.Pack(label.New(app, "d", label.Text("Derived"), label.FontOpt(bigger)))
```

Redefining a named font with `Define` affects widgets that look it up
afterwards.

**Distances** (padding, border widths, lengths) take a number of pixels (an
`int` or `float64`) or a `screenunit.Distance` in Tk's other units:
`screenunit.Pt(3)` (points, Tk's `3p`), `screenunit.Mm(2)` (millimetres),
`screenunit.Cm(1)` (centimetres), `screenunit.In(0.5)` (inches). The
conversion uses the screen's real DPI, so `Cm(1)` really is about a
centimetre. Anything else, a string included, does not compile.

```go
pack.Pack(w, pack.PadX(screenunit.Mm(2)), pack.PadY(4))
label.WrapLength(screenunit.In(4))
```

Options typed `int` take pixels; convert a distance for them with
`screenunit.Cm(1).Pixels()` (or `.Float()` for canvas coordinates). A distance
that arrives as text, from a settings file say, goes through
`screenunit.Parse("1.5p")`, which returns an error wrapping
`screenunit.ErrBadDistance` when it is malformed.

### 3.4 Enumerated option values

Shared enumerations live in the `option` package:

| Kind | Values |
|------|--------|
| `option.Relief` | `ReliefFlat`, `ReliefRaised`, `ReliefSunken`, `ReliefGroove`, `ReliefRidge`, `ReliefSolid` |
| `option.Anchor` | `AnchorCenter`, `AnchorN`, `AnchorNE`, `AnchorE`, `AnchorSE`, `AnchorS`, `AnchorSW`, `AnchorW`, `AnchorNW` |
| `option.Justify` | `JustifyLeft`, `JustifyCenter`, `JustifyRight` |

Image/text combination is `widget.Compound`: `CompoundNone`, `CompoundLeft`,
`CompoundRight`, `CompoundTop`, `CompoundBottom`, `CompoundCenter`.

---

## 4. Geometry management

A geometry manager decides where each widget goes and how big it is. takigo
has Tk's three managers, each in its own package. Within one container, use
one manager — mixing `pack` and `grid` in the same parent is an error in Tk
and a recipe for fights here. Different containers can each use a different
manager, and nesting frames is how you build complex layouts.

### 4.1 pack

`pack` places widgets against the sides of the remaining space (the
*cavity*), in the order you pack them.

```go
import "github.com/msorc/takigo/geometry/pack"

pack.Pack(toolbar, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))
pack.Pack(status,  pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))
pack.Pack(sidebar, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillY))
pack.Pack(content, pack.FillOpt(pack.FillBoth), pack.Expand(true))
```

| Option | Meaning |
|--------|---------|
| `SideOpt(pack.Top \| Bottom \| Left \| Right)` | Which side of the cavity to take (default `Top`). |
| `FillOpt(pack.FillNone \| FillX \| FillY \| FillBoth)` | Stretch the widget to fill its parcel. |
| `Expand(true)` | Claim a share of any extra space in the container. |
| `Anchor(option.AnchorW)` | Where to put the widget inside a larger parcel. |
| `PadX(p)`, `PadY(p)` | External padding on both sides. |
| `PadXPair(l, r)`, `PadYPair(t, b)` | Asymmetric external padding. |
| `IPadX(p)`, `IPadY(p)` | Internal padding added to the widget's own size. |
| `In(container)` | Pack inside another container (must be the parent or a descendant of it), at the end of its order. |
| `Before(w)`, `After(w)` | Put the widget just before or after the packed widget `w` in the packing order. |

Rules of thumb that hold in Tk and here:

- **Order matters.** When the window shrinks, widgets packed last lose space
  first. Pack the things that must stay visible — status bars, button rows,
  scrollbars — *before* the stretchy content.
- `FillOpt` controls the widget inside its parcel; `Expand` controls whether
  the parcel grows. A text area that should grow with the window needs both:
  `pack.FillOpt(pack.FillBoth), pack.Expand(true)`.

To pack several widgets with the same options in one call, wrap them in a
`geometry.Group`:

```go
import "github.com/msorc/takigo/geometry"

pack.Pack(geometry.Group{ok, cancel, help}, pack.SideOpt(pack.Left), pack.PadX(4))
```

**Packing a widget again** works like Tcl's `pack configure`: only the options
you give change, and the widget keeps its place in the packing order unless
you pass `In`, `Before` or `After`. So adjusting one option at run time is
just:

```go
pack.Pack(sidebar, pack.FillOpt(pack.FillY)) // side, padding etc. stay as they were
```

`Before` and `After` also insert new widgets anywhere in the order — for
example a search bar shown on demand above existing content:

```go
pack.Pack(searchBar, pack.Before(content), pack.FillOpt(pack.FillX))
```

With a group, the first widget goes next to the sibling and the others follow
it in order.

`pack.Forget(w)` removes a widget from the layout without destroying it. Pack
it again to bring it back; it then starts from the default options and goes
at the end, as a newly packed widget does.

### 4.2 grid

`grid` arranges widgets in rows and columns. It is usually the best choice
for forms and dialogs.

```go
import "github.com/msorc/takigo/geometry/grid"

grid.Grid(nameLabel, grid.Row(0), grid.Column(0), grid.Sticky(grid.StickE))
grid.Grid(nameEntry, grid.Row(0), grid.Column(1), grid.Sticky(grid.EW))
grid.ColumnConfigure(parent, 1, grid.Weight(1))
```

| Option | Meaning |
|--------|---------|
| `Row(r)`, `Column(c)` | Cell. If `Row` is omitted, the widget goes in the next empty row. `Column` defaults to 0. |
| `RowSpan(n)`, `ColumnSpan(n)` | Occupy several cells. |
| `Sticky(s)` | Which cell edges to stick to: `grid.StickN`, `StickS`, `StickE`, `StickW`, combined with `\|`, or the shortcuts `grid.NS`, `grid.EW`, `grid.NSEW`. |
| `PadX`, `PadY`, `PadXPair`, `PadYPair`, `IPadX`, `IPadY`, `In` | As for `pack`. |

Rows and columns are configured on the container:

```go
grid.ColumnConfigure(f, 0, grid.Weight(1), grid.MinSize(80))
grid.RowConfigure(f, 2, grid.Weight(1), grid.Pad(4))
grid.ColumnConfigure(f, 1, grid.Weight(1), grid.Uniform("buttons"))
```

- `Weight(n)` — how extra space is shared. Columns with weight 0 (the default)
  never grow. **If nothing grows when you resize the window, you forgot a
  weight.**
- `MinSize(px)`, `Pad(px)` — minimum size and extra space for the slot.
- `Uniform(group)` — all slots in the same group get sizes proportional to
  their weights (equal-width button columns).

Other container-level calls: `grid.SetAnchor(container, option.AnchorN)`
positions the whole grid when the container is larger than it, and
`grid.SetPropagate(container, false)` stops the grid from resizing its
container.

**Gridding a widget again** works like Tcl's `grid configure`: only the
options you give change. A widget that is already gridded keeps its row,
column and container unless you pass `Row`, `Column` or `In`:

```go
grid.Grid(entry, grid.Sticky(grid.NSEW)) // same cell, new sticky
grid.Grid(entry, grid.Column(2))         // same row, new column
```

`grid.Forget(w)` clears a widget's grid options; gridding it again starts
from the defaults.

A **group** is laid out left to right across one row, which makes simple
forms very compact. Tk's relative-placement shortcuts (`x`, `-`, `^`) are
available as `grid.Relative`:

```go
grid.Grid(geometry.Group{a, b, c}, grid.Sticky(grid.NSEW))
grid.Grid(geometry.Group{
	d,
	grid.Relative(grid.RelLeft), // "-": d spans one more column
	grid.Relative(grid.RelUp),   // "^": c (above) spans one more row
}, grid.Sticky(grid.NSEW))
```

`grid.Relative(grid.RelEmpty)` is Tk's `x` (skip a column).

#### Example: a sign-up form

This form combines `grid`, entry validation, a password field, a status line
and a keyboard shortcut.

```go
package main

import (
	"log"
	"strconv"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/bind"
	"github.com/msorc/takigo/geometry"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/entry"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Sign up"))
	if err != nil {
		log.Fatal(err)
	}

	form := frame.New(app, "form", frame.BorderWidth(10))
	grid.Grid(form, grid.Sticky(grid.NSEW))
	grid.ColumnConfigure(app, 0, grid.Weight(1))
	grid.RowConfigure(app, 0, grid.Weight(1))

	nameLbl := label.New(form, "nameLbl", label.Text("Name:"), label.Anchor(option.AnchorE))
	name := entry.New(form, "name", entry.Width(30))

	ageLbl := label.New(form, "ageLbl", label.Text("Age:"), label.Anchor(option.AnchorE))
	age := entry.New(form, "age",
		entry.Width(4),
		entry.ValidateOpt("key"),
		entry.ValidateCmdOpt(func(proposed string) bool {
			if proposed == "" {
				return true
			}
			n, err := strconv.Atoi(proposed)
			return err == nil && n >= 0 && n < 150
		}),
	)

	pwLbl := label.New(form, "pwLbl", label.Text("Password:"), label.Anchor(option.AnchorE))
	pw := entry.New(form, "pw", entry.Width(30), entry.Show('•'))

	status := label.New(form, "status", label.Anchor(option.AnchorW))

	submit := func() {
		if name.GetText() == "" {
			status.Configure(label.Text("Please enter a name"), label.Foreground("red3"))
			widget.Focus(app, name.Window())
			return
		}
		status.Configure(
			label.Text("Welcome, "+name.GetText()+"!"),
			label.Foreground("dark green"),
		)
	}
	ok := button.New(form, "ok", button.Text("Submit"), button.Command(submit))

	grid.Grid(geometry.Group{nameLbl, name}, grid.Sticky(grid.EW), grid.PadY(2))
	grid.Grid(ageLbl, grid.Row(1), grid.Column(0), grid.Sticky(grid.EW), grid.PadY(2))
	grid.Grid(age, grid.Row(1), grid.Column(1), grid.Sticky(grid.StickW), grid.PadY(2))
	grid.Grid(geometry.Group{pwLbl, pw}, grid.Sticky(grid.EW), grid.PadY(2))
	grid.Grid(status, grid.ColumnSpan(2), grid.Sticky(grid.EW))
	grid.Grid(ok, grid.Column(1), grid.Sticky(grid.StickE), grid.PadY(6))
	grid.ColumnConfigure(form, 1, grid.Weight(1))

	app.Bind().Bind(app.Window().PathName, "<Return>", func(*bind.EventData) bool {
		submit()
		return true
	})

	widget.Focus(app, name.Window())
	app.Run()
}
```

Things to notice:

- The form frame is gridded into the root with `NSEW` and the root's row and
  column get weight 1, so the form follows the window; inside the form only
  column 1 (the entries) has weight, so labels keep their width and entries
  stretch.
- The labels stick `EW` and anchor their text `E`: they fill column 0 and
  right-align their text, so the colons line up.
- The age row uses explicit cells because its two widgets need different
  `Sticky` values — a group applies one set of options to all its members.
- Rows 0, 2, 3 and 4 are assigned automatically ("next empty row").
- `entry.ValidateOpt("key")` checks every keystroke; the callback receives the
  *proposed* new text and returns `false` to reject it (Tk's `%P`). Other
  modes: `"focus"`, `"focusin"`, `"focusout"`, `"all"`, `"none"`.
- The `<Return>` binding is on the toplevel's path `"."`, which is part of
  every widget's binding chain, so Return submits from any entry (chapter 7).

### 4.3 place

`place` puts a widget at an absolute or relative position. It is handy for
overlays, badges and custom layouts; for ordinary forms prefer `grid`.

```go
import "github.com/msorc/takigo/geometry/place"

f := frame.New(app, "stage", frame.Width(300), frame.Height(200))
pack.Pack(f)

// Centred in its container, whatever the container's size.
place.Place(badge, place.RelX(0.5), place.RelY(0.5), place.Anchor(option.AnchorCenter))

// 4 px from the bottom-right corner.
place.Place(corner, place.RelX(1), place.RelY(1), place.X(-4), place.Y(-4),
	place.Anchor(option.AnchorSE))

// A 6-px bar across the full width at the top.
place.Place(bar, place.RelWidth(1), place.Height(6))
```

Absolute (`X`, `Y`, `Width`, `Height`) and relative (`RelX`, `RelY`,
`RelWidth`, `RelHeight`, fractions of the container) values add up, exactly
as in Tk. `place.Forget(w)` un-places a widget. `place` never changes the
container's size, which is why the frame above was given an explicit size.

### 4.4 Reading sizes

Layout happens at idle time, as in Tk: right after you pack a widget, its
`Width` and `Height` are not final yet. If you need them before the event
loop gets a chance to run, flush pending layout and redraws first:

```go
app.UpdateIdleTasks()
w := lbl.Window().Width      // actual size
rw := lbl.Window().ReqWidth  // size the widget asked for
```

---

## 5. A tour of the classic widgets

The classic widgets live in `widget/<name>` packages. Their look matches Tk's
defaults on Unix. The themed equivalents are covered in chapter 16; you can
freely mix both kinds in one window.

### frame and labelframe

Containers. A `frame` is a plain rectangle; a `labelframe` draws a titled
border.

```go
f := frame.New(app, "f", frame.BorderWidth(2), frame.Relief(option.ReliefGroove),
	frame.Background("gray90"), frame.Width(200), frame.Height(100))

lf := labelframe.New(app, "opts", labelframe.Text("Options"),
	labelframe.PadX(8), labelframe.PadY(4))
```

`labelframe.LabelWidgetOpt(w)` uses any widget (say, a checkbutton) as the
title, and `labelframe.LabelAnchor` positions it.

### label and message

A `label` shows text and/or an image. `WrapLength` wraps long text;
`JustifyOpt` aligns wrapped lines; `Anchor` positions the content in a label
that is larger than its content.

```go
l := label.New(app, "l",
	label.Text("A long explanation that wraps at four inches."),
	label.WrapLength(screenunit.In(4)),
	label.JustifyOpt(option.JustifyLeft),
	label.Relief(option.ReliefSunken),
)
```

`label.Width(n)` and `label.Height(n)` are in characters and lines for text
labels. `label.TextVariable(v)` binds the text to a variable (chapter 6).

A `message` wraps text to a given aspect ratio rather than a fixed width:

```go
msg := message.New(app, "msg",
	message.Text("A message widget wraps long text to keep a given aspect ratio."),
	message.Aspect(300)) // width = 3 × height
```

### button

```go
b := button.New(app, "save",
	button.Text("Save"),
	button.Command(save),
	button.Width(8),
	button.Default(button.DefaultActive), // draw the "default button" ring
)
b.Invoke() // run the command programmatically, like "$b invoke"
```

Buttons can show images too (`button.ImageOpt`, `button.CompoundOpt`); see
chapter 14.

### entry

A one-line text field.

```go
e := entry.New(app, "e",
	entry.Width(30),
	entry.Placeholder("Search…"),
)
text := e.GetText()
e.SetText("replacement")
e.InsertChars(0, "prefix ")
e.SelectAll()
```

`entry.Show('*')` masks input; validation was shown in the form example.

### checkbutton and radiobutton

Both are driven by a shared `*widget.Variable[string]` (chapter 6). A
checkbutton writes its *on value* (`"1"` by default) or *off value* (`"0"`);
radiobuttons that share one variable write their own `Value`.

```go
cheese := widget.NewVariable("1")
checkbutton.New(f, "cheese", checkbutton.Text("Extra cheese"), checkbutton.Var(cheese))

size := widget.NewVariable("medium")
for _, s := range []string{"small", "medium", "large"} {
	radiobutton.New(f, s, radiobutton.Text(s), radiobutton.Value(s), radiobutton.Var(size))
}
```

`checkbutton.OnValueOpt`/`OffValueOpt` change the stored values, and
`Command` runs a callback on every click. `Selected()`, `Toggle()`,
`Select()` and `Invoke()` let you drive them from code.

### scale

A slider over a numeric range.

```go
s := scale.New(app, "vol",
	scale.OrientOpt(scale.Horizontal),
	scale.FromOpt(0), scale.ToOpt(100), scale.ResolutionOpt(5),
	scale.LabelOpt("Volume"),
	scale.TickIntervalOpt(25),
	scale.LengthOpt(200),
	scale.CommandOpt(func(v float64) { setVolume(v) }),
)
s.Set(50)
v := s.Get()
```

### spinbox

A numeric or list-valued entry with arrows.

```go
n := spinbox.New(app, "n", spinbox.FromOpt(1), spinbox.ToOpt(10), spinbox.IncrementOpt(1),
	spinbox.WidthOpt(5), spinbox.CommandOpt(func(v string) { log.Print(v) }))

c := spinbox.New(app, "c", spinbox.ValuesOpt([]string{"red", "green", "blue"}),
	spinbox.WrapOpt(true))

current := n.GetText()
```

### listbox

A scrollable list of strings.

```go
lb := listbox.New(app, "lb",
	listbox.Items("alpha", "beta", "gamma"),
	listbox.Height(10),
	listbox.SelectModeOpt(listbox.SelectExtended), // Single, Browse, Multiple, Extended
)
lb.Insert(lb.ItemCount(), "delta")  // append
lb.Delete(0, 0)                      // delete items first..last (inclusive)
sel := lb.Selection()                // []int of selected indices
lb.SelectionSet(2, 2)
lb.See(10)                           // scroll item 10 into view
lb.ItemConfigure(1, "gray50", "")    // per-item foreground/background
lb.SelectCmd = func() { /* selection changed by the user */ }
```

Chapter 8 connects a listbox to a scrollbar.

### panedwindow

Resizable panes separated by draggable sashes.

```go
pw := panedwindow.New(app, "pw", panedwindow.OrientOpt(panedwindow.Horizontal))
pack.Pack(pw, pack.FillOpt(pack.FillBoth), pack.Expand(true))

left := listbox.New(pw, "left")
right := text.New(pw, "right")
pw.Add(left.Window(), 100)   // second argument: minimum pane size in pixels
pw.Add(right.Window(), 200)
```

Children of a panedwindow are created with the panedwindow as their parent
and added with `Add` instead of `pack`/`grid`.

### menubutton

A button that posts a menu; see chapter 9.

### Example: linking widgets together

This small order form uses radiobuttons, checkbuttons, a scale and a label
that follows a variable.

```go
package main

import (
	"fmt"
	"log"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/checkbutton"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/labelframe"
	"github.com/msorc/takigo/widget/radiobutton"
	"github.com/msorc/takigo/widget/scale"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Pizza"))
	if err != nil {
		log.Fatal(err)
	}

	size := widget.NewVariable("medium")
	cheese := widget.NewVariable("1")
	olives := widget.NewVariable("0")

	summary := widget.NewVariable("")
	var slices float64 = 8

	update := func() {
		extras := ""
		if cheese.Get() == "1" {
			extras += " +cheese"
		}
		if olives.Get() == "1" {
			extras += " +olives"
		}
		summary.Set(fmt.Sprintf("%s pizza%s, %.0f slices", size.Get(), extras, slices))
	}

	// Pack the status line first so it keeps its space when the window shrinks.
	out := label.New(app, "summary", label.TextVariable(summary), label.Relief(option.ReliefSunken))
	pack.Pack(out, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	sizes := labelframe.New(app, "sizes", labelframe.Text("Size"), labelframe.PadX(8), labelframe.PadY(4))
	pack.Pack(sizes, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillY), pack.PadX(8), pack.PadY(8))
	for _, s := range []string{"small", "medium", "large"} {
		rb := radiobutton.New(sizes, s,
			radiobutton.Text(s),
			radiobutton.Value(s),
			radiobutton.Var(size),
			radiobutton.Anchor(option.AnchorW),
		)
		pack.Pack(rb, pack.FillOpt(pack.FillX))
	}

	toppings := labelframe.New(app, "toppings", labelframe.Text("Toppings"), labelframe.PadX(8), labelframe.PadY(4))
	pack.Pack(toppings, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillY), pack.PadY(8))
	pack.Pack(checkbutton.New(toppings, "cheese",
		checkbutton.Text("Extra cheese"), checkbutton.Var(cheese), checkbutton.Anchor(option.AnchorW)),
		pack.FillOpt(pack.FillX))
	pack.Pack(checkbutton.New(toppings, "olives",
		checkbutton.Text("Olives"), checkbutton.Var(olives), checkbutton.Anchor(option.AnchorW)),
		pack.FillOpt(pack.FillX))

	sl := scale.New(app, "slices",
		scale.OrientOpt(scale.Vertical),
		scale.FromOpt(4), scale.ToOpt(12), scale.ResolutionOpt(2),
		scale.ValueOpt(slices),
		scale.LabelOpt("Slices"),
		scale.CommandOpt(func(v float64) { slices = v; update() }),
	)
	pack.Pack(sl, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillY), pack.PadX(8), pack.PadY(8))

	for _, v := range []*widget.Variable[string]{size, cheese, olives} {
		v.OnChange(func(_, _ string) { update() })
	}
	update()

	app.Run()
}
```

Nothing here calls `Configure` on the summary label: it follows `summary`
on its own. That is the pattern to reach for whenever several widgets show
one piece of state.

---

## 6. Variables

Tk links widgets through global Tcl variables. takigo uses a small generic
observable, `widget.Variable[T]`:

```go
v := widget.NewVariable("initial") // Variable[string]
v.Set("new")                       // notifies listeners if the value changed
s := v.Get()

unsubscribe := v.OnChange(func(old, new string) {
	log.Printf("%q -> %q", old, new)
})
defer unsubscribe()
```

- `Set` with the current value does nothing, so listeners only see real
  changes.
- `widget.NewUnsetVariable[string]()` creates a variable with no value yet,
  like a Tcl variable that doesn't exist. Radiobuttons linked to it start
  with none selected, as in Tk. `IsSet()` tells the two states apart.
- Which `T` a widget wants is part of its option's signature: classic
  checkbuttons, radiobuttons, labels and entries use `Variable[string]`;
  `ttk.CheckbuttonVar` and `ttk.ToggleswitchVar` use `Variable[bool]`;
  `ttk.ScaleVariable` uses `Variable[float64]`.
- Variables are **not** goroutine-safe: `Set` runs listeners, which reconfigure
  widgets. Only touch them on the event loop (chapter 15).

---

## 7. Events and bindings

### 7.1 Commands

Buttons, checkbuttons, radiobuttons, scales, spinboxes and menu entries take
a callback (`Command`, `CommandOpt`). That's all most UI code needs.

### 7.2 The binding engine

For everything else — keys, mouse clicks on arbitrary widgets, hover effects —
there is Tk's `bind`. Get the engine from the app and bind a pattern to a
*tag*:

```go
eng := app.Bind()

eng.Bind(entry.Window().PathName, "<Return>", func(ed *bind.EventData) bool {
	search(entry.GetText())
	return true // "break": stop processing this event
})
```

**Tags and the binding chain.** Each window has a list of tags. By default it
is, in order:

1. the widget's own path (`".form.name"`),
2. its class (`"Entry"`, `"Button"`, …),
3. its toplevel's path (`"."`),
4. `"all"`.

When an event arrives, the engine walks the chain and, for each tag, runs the
single most specific binding that matches. A handler returning `true` stops
the walk, like a Tk script ending in `break`.

A widget's built-in behaviour — an entry inserting characters, a button
flashing when clicked — is attached at its **class** tag. So:

- A binding on the widget's **path** runs *before* the built-in behaviour and
  can suppress it by returning `true`.
- A binding on the **class** tag (e.g. `"Button"`) applies to every widget of
  that class.
- A binding on the toplevel path (`"."` for the main window) catches events
  from every widget inside that window — that is how the form example's
  `<Return>` works from any entry.
- A binding on `"all"` applies everywhere.

```go
eng.Bind("Button", "<Button-3>", func(*bind.EventData) bool {
	log.Print("right-click on some button")
	return false
})
```

You can inspect or replace a window's chain with `eng.BindTags(w.Window())`
and `eng.SetBindTags(w.Window(), tags)`. Leaving the class tag out disables
the widget's built-in behaviour. `eng.Unbind(tag, pattern)` removes a
binding.

### 7.3 Pattern syntax

The pattern language is Tk's.

| Pattern | Matches |
|---------|---------|
| `<Button-1>` / `<ButtonPress-1>` / `<1>` | Left-button press (buttons 1–9). `<Button>` matches any button. A lone digit is a button, as in Tk; the digit key is `<Key-1>`. |
| `<ButtonRelease-1>` | Left-button release. |
| `<Double-Button-1>`, `<Triple-Button-1>` | Double / triple click. |
| `<B1-Motion>` / `<Motion-1>` | Motion while button 1 is held (`B1`…`B5`, also spelled `Button1`…). |
| `<Motion>` | Pointer motion. |
| `<Enter>`, `<Leave>` | Pointer crosses into / out of the window. |
| `<Key-a>`, `<a>`, `a` | The `a` key. `<Key>` matches any key. |
| `<Return>`, `<Escape>`, `<Tab>`, `<space>`, `<BackSpace>`, `<Delete>` | Named keys. |
| `<Left>`, `<Right>`, `<Up>`, `<Down>`, `<Home>`, `<End>`, `<Prior>`, `<Next>` | Navigation keys. |
| `<F1>` … `<F12>` | Function keys. |
| `<KeyRelease-x>` | Key release. |
| `<Control-s>`, `<Control-Shift-z>`, `<Alt-F4>`, `<Meta-x>` | Modifier combinations. |
| `<FocusIn>`, `<FocusOut>`, `<Configure>`, `<Map>`, `<Unmap>`, `<Destroy>`, `<Expose>` | Window events. |
| `<<Copy>>` | A virtual event (below). |

### 7.4 Event details

The handler receives a `*bind.EventData`; the actual event is in `RawEvent`:

```go
eng.Bind("all", "<Motion>", func(ed *bind.EventData) bool {
	ev := ed.RawEvent.(*event.Event)
	status.Configure(label.Text(fmt.Sprintf("%d, %d", ev.X, ev.Y)))
	return false
})
```

Useful fields of `event.Event`:

| Field | Meaning |
|-------|---------|
| `Type` | `event.KeyPressType`, `ButtonPressType`, `MotionType`, `EnterType`, … |
| `X`, `Y` | Pointer position in the event window. |
| `RootX`, `RootY` | Pointer position on the screen (use for popup menus). |
| `Button` | Button number (1 left, 2 middle, 3 right; 4/5 wheel). |
| `State` | Modifier and button mask, e.g. `ev.State&platform.Button1Mask != 0` while dragging, `platform.ShiftMask`, `platform.ControlMask`. |
| `KeySym`, `Str` | Key symbol and the text it produced. |
| `ConfigWidth`, `ConfigHeight` | New size for `<Configure>`. |
| `Name` | Virtual event name, without `<< >>`. |

**Cross-platform shortcuts.** On macOS the Command key arrives as Mod2, which
on X11 and Windows is NumLock. When you check modifiers yourself, test the
"accelerator" modifier as `ev.State&(platform.ControlMask|platform.CommandMask)`.

### 7.5 Virtual events

A virtual event names an abstract action and maps it to one or more physical
patterns. The standard set — `<<Copy>>`, `<<Cut>>`, `<<Paste>>`,
`<<SelectAll>>`, `<<Undo>>` and `<<Redo>>` — is predefined, and you can add your own:

```go
eng.AddVirtualEvent("Save", "<Control-s>", "<F2>")
eng.Bind("all", "<<Save>>", func(*bind.EventData) bool {
	save()
	return true
})

// Fire it programmatically, e.g. from a toolbar button:
eng.GenerateEvent(someWidget.Window(), "Save")
```

`eng.RemoveVirtualEvent("Save")` removes the mapping.

### 7.6 Low-level event handlers

Under the binding engine sits the event dispatcher, which delivers raw events
per window and event mask. Widgets use it for their own behaviour; you can
use it too when you want every event of a kind on one window:

```go
id := app.Dispatcher().Bind(c.Window().PlatformID,
	event.ButtonPressMask|event.MotionMask|event.ButtonReleaseMask,
	func(ev *event.Event) {
		switch ev.Type { // a mask can deliver several event types
		case event.ButtonPressType:
			// ...
		case event.MotionType:
			// ...
		}
	})
// later:
app.Dispatcher().UnbindID(id)
```

Always switch on `ev.Type`: a mask such as `event.StructureNotifyMask` covers
several event types. For input events (keys, buttons, motion, crossing) these
handlers run at the window's class tag in the binding chain, like the
built-in behaviour; handlers for other events (expose, configure, focus) run
before any binding.

### 7.7 Keyboard focus

Key events go to the widget with the keyboard focus. Give it the focus with:

```go
widget.Focus(app, entry.Window())
```

(not by talking to the window system directly — takigo, like Tk, keeps the
system focus on the toplevel and redirects keys itself). `widget.FocusWindow(app)`
returns the focused window. Tab and Shift-Tab move the focus between
focusable widgets in stacking order.

---

## 8. Scrolling

Scrollable widgets (listbox, text, canvas, entry, treeview) and scrollbars talk
through Tk's two-way protocol:

- The widget reports its visible fraction to a **scroll command**:
  `func(first, last float64)`. Point it at the scrollbar's `Set`.
- The scrollbar reports user actions to its **command** as Tk-style
  arguments: `("moveto", fraction float64)` or
  `("scroll", n int, "units"|"pages")`. Translate those into the widget's
  `YViewMoveTo` / `YViewScroll` (or the `X` versions).

A small adapter makes this a one-liner for any vertically scrollable widget:

```go
type yScroller interface {
	YViewMoveTo(fraction float64)
	YViewScroll(count int, pages bool)
}

func yScrollCommand(w yScroller) func(args ...any) {
	return func(args ...any) {
		if len(args) == 0 {
			return
		}
		switch args[0] {
		case "moveto":
			if f, ok := args[1].(float64); ok {
				w.YViewMoveTo(f)
			}
		case "scroll":
			n, _ := args[1].(int)
			unit, _ := args[2].(string)
			w.YViewScroll(n, unit == "pages")
		}
	}
}
```

Because each side needs the other, create the scrollbar first, then the
widget, then connect the scrollbar with `Configure`:

```go
sb := scrollbar.New(body, "sb", scrollbar.OrientOpt(scrollbar.Vertical))
lb := listbox.New(body, "lb", listbox.YScrollCommand(sb.Set))
sb.Configure(scrollbar.CommandOpt(yScrollCommand(lb)))

pack.Pack(sb, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
pack.Pack(lb, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))
```

The same wiring works with `ttk.NewScrollbar` (`ttk.ScrollbarOrientOpt`,
`ttk.ScrollbarCommandOpt`). The mouse wheel scrolls these widgets without any
extra code.

### Example: a filterable colour list

```go
package main

import (
	"log"
	"strings"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/bind"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/widget/entry"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/listbox"
	"github.com/msorc/takigo/widget/scrollbar"
)

// yScroller and yScrollCommand as above.

func main() {
	app, err := takigo.NewApp(takigo.Title("Color browser"))
	if err != nil {
		log.Fatal(err)
	}

	all := []string{
		"alice blue", "antique white", "aquamarine", "azure", "beige", "bisque",
		"blanched almond", "blue violet", "burlywood", "cadet blue", "chartreuse",
		"chocolate", "coral", "cornflower blue", "cornsilk", "dark goldenrod",
		"dark khaki", "dark olive green", "dark orange", "dark orchid", "dark salmon",
		"dark sea green", "dark slate blue", "deep pink", "deep sky blue", "dodger blue",
		"firebrick", "forest green", "gainsboro", "gold", "goldenrod", "honeydew",
		// ... as many X11 colour names as you like
	}

	filter := entry.New(app, "filter", entry.Placeholder("type to filter…"))
	pack.Pack(filter, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadX(4), pack.PadY(4))

	swatch := label.New(app, "swatch", label.Text("select a color"), label.Height(2))
	pack.Pack(swatch, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	body := frame.New(app, "body")
	pack.Pack(body, pack.FillOpt(pack.FillBoth), pack.Expand(true))

	var lb *listbox.Listbox
	sb := scrollbar.New(body, "sb", scrollbar.OrientOpt(scrollbar.Vertical))
	lb = listbox.New(body, "lb",
		listbox.Items(all...),
		listbox.Height(15),
		listbox.SelectModeOpt(listbox.SelectBrowse),
		listbox.YScrollCommand(sb.Set),
	)
	sb.Configure(scrollbar.CommandOpt(yScrollCommand(lb)))
	pack.Pack(sb, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(lb, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	lb.SelectCmd = func() {
		sel := lb.Selection()
		if len(sel) == 0 {
			return
		}
		name := lb.GetItems()[sel[0]]
		swatch.Configure(label.Text(name), label.Background(name))
	}

	app.Bind().Bind(filter.Window().PathName, "<KeyRelease>", func(*bind.EventData) bool {
		q := strings.ToLower(filter.GetText())
		lb.Delete(0, lb.ItemCount()-1)
		for _, n := range all {
			if strings.Contains(strings.ToLower(n), q) {
				lb.Insert(lb.ItemCount(), n)
			}
		}
		return false
	})

	app.Run()
}
```

The `<KeyRelease>` binding runs after the entry has already processed the key
press, so `GetText` sees the updated text.

---

## 9. Menus

### A menubar

```go
bar := menu.NewMenubar(app, "menubar") // Tk: . configure -menu ...

file := menu.New(app, "file")
file.AddCommandAccelUL("Open…", "Ctrl+O", 0, open) // label, accelerator text, underline, callback
file.AddCommandAccelUL("Save", "Ctrl+S", 0, save)
file.AddSeparator()
file.AddCommandUL("Quit", 0, app.Quit)
bar.AddCascade("File", 0, file) // label, underlined character, menu
```

The underlined character makes `Alt+F` open the File menu, and inside a posted
menu the underlined letter picks an entry. Arrow keys, Return and Escape work
as in Tk.

Entry types:

| Method | Entry |
|--------|-------|
| `AddCommand(label, fn)` | A plain command. |
| `AddCommandUL(label, underline, fn)` | …with an underlined character. |
| `AddCommandAccel(label, accel, fn)`, `AddCommandAccelUL(...)` | …with accelerator text on the right. |
| `AddCommandImage(label, img, compound, fn)` | …with an image. |
| `AddCheckbutton(label, checked, fn)` | A check entry. |
| `AddRadiobutton(label, checked, fn)` | A radio entry. |
| `AddCascade(label, submenu)`, `AddCascadeUL(...)` | A submenu. |
| `AddSeparator()` | A separator line. |

As in Tk 9, menus have no tear-off entry unless you ask for one with
`menu.TearOffOpt(true)` (Tk 8 added one by default).

> **Accelerator text is only a label.** As in Tk, `"Ctrl+S"` next to a menu
> entry doesn't make Ctrl+S do anything. Bind the key yourself (chapter 7);
> the editor example below does.

### Popup (context) menus

Post a menu at the pointer's screen position:

```go
ctx := menu.New(app, "ctx")
ctx.AddCommand("Copy", doCopy)
ctx.AddCommand("Paste", doPaste)

app.Bind().Bind(target.Window().PathName, "<Button-3>", func(ed *bind.EventData) bool {
	ev := ed.RawEvent.(*event.Event)
	ctx.Post(ev.RootX, ev.RootY)
	return true
})
```

`Unpost()` closes it; clicking elsewhere or pressing Escape also does.

### Menubuttons

A menubutton shows a menu below itself when pressed. With
`IndicatorOnOpt(true)` it draws the little option-menu indicator.

```go
m := menu.New(app, "sizes")
m.AddRadiobutton("Small", false, func() { setSize("small") })
m.AddRadiobutton("Large", true, func() { setSize("large") })

mb := menubutton.New(app, "mb", menubutton.Text("Size"), menubutton.MenuOpt(m),
	menubutton.IndicatorOnOpt(true))
```

The themed version is `ttk.NewMenubutton(parent, name, ttk.MenubuttonText(...),
ttk.MenubuttonMenu(m))`.

### Example: a tiny text editor

This program brings together menus, keyboard shortcuts, the text widget
(chapter 12), scrolling and the standard dialogs (chapter 11).

```go
package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/bind"
	"github.com/msorc/takigo/dialog"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/menu"
	"github.com/msorc/takigo/widget/scrollbar"
	"github.com/msorc/takigo/widget/text"
)

type editor struct {
	app    *takigo.App
	tw     *text.TextWidget
	status *label.Label
	path   string
	saved  string
}

func (e *editor) contents() string { return e.tw.Get("1.0", "end") }
func (e *editor) dirty() bool      { return e.contents() != e.saved }

func (e *editor) setPath(p string) {
	e.path = p
	title := "Untitled"
	if p != "" {
		title = filepath.Base(p)
	}
	e.app.WmInfo().SetTitle(title + " — Tiny Editor")
}

// confirmDiscard asks before throwing away unsaved changes.
func (e *editor) confirmDiscard() bool {
	if !e.dirty() {
		return true
	}
	switch dialog.ShowMessage(e.app,
		dialog.MsgTitle("Unsaved changes"),
		dialog.MsgMessage("Save changes before continuing?"),
		dialog.MsgType(dialog.MsgQuestion),
		dialog.MsgButtons(dialog.BtnYesNoCancel),
	) {
	case dialog.ResultYes:
		return e.save()
	case dialog.ResultNo:
		return true
	default:
		return false
	}
}

func (e *editor) open() {
	if !e.confirmDiscard() {
		return
	}
	p, ok := dialog.OpenFile(e.app,
		dialog.FileTitle("Open"),
		dialog.FileTypes(
			dialog.FileType{Name: "Text files", Pattern: "*.txt *.md"},
			dialog.FileType{Name: "All files", Pattern: "*"},
		),
	)
	if !ok {
		return
	}
	data, err := os.ReadFile(p)
	if err != nil {
		dialog.ShowMessage(e.app, dialog.MsgType(dialog.MsgError),
			dialog.MsgMessage("Cannot open file"), dialog.MsgDetail(err.Error()))
		return
	}
	e.tw.Delete("1.0", "end")
	e.tw.Insert("1.0", string(data))
	e.tw.MarkSet("insert", "1.0")
	e.tw.Edit("reset")
	e.saved = e.contents()
	e.setPath(p)
}

func (e *editor) save() bool {
	if e.path == "" {
		p, ok := dialog.SaveFile(e.app, dialog.FileDefaultExtension(".txt"))
		if !ok {
			return false
		}
		e.setPath(p)
	}
	body := e.contents()
	if err := os.WriteFile(e.path, []byte(body), 0o644); err != nil {
		dialog.ShowMessage(e.app, dialog.MsgType(dialog.MsgError),
			dialog.MsgMessage("Cannot save file"), dialog.MsgDetail(err.Error()))
		return false
	}
	e.saved = body
	e.status.Configure(label.Text("Saved " + e.path))
	return true
}

func (e *editor) quit() {
	if e.confirmDiscard() {
		e.app.Quit()
	}
}

func main() {
	app, err := takigo.NewApp(takigo.Size(640, 480))
	if err != nil {
		log.Fatal(err)
	}
	e := &editor{app: app}
	e.setPath("")

	// Menubar.
	bar := menu.NewMenubar(app, "menubar")
	file := menu.New(app, "file")
	file.AddCommandAccelUL("Open…", "Ctrl+O", 0, e.open)
	file.AddCommandAccelUL("Save", "Ctrl+S", 0, func() { e.save() })
	file.AddSeparator()
	file.AddCommandUL("Quit", 0, e.quit)
	bar.AddCascade("File", 0, file)

	edit := menu.New(app, "edit")
	edit.AddCommandAccel("Undo", "Ctrl+Z", func() { e.tw.Edit("undo") })
	edit.AddCommandAccel("Redo", "Ctrl+Shift+Z", func() { e.tw.Edit("redo") })
	edit.AddSeparator()
	edit.AddCommand("Select All", func() { e.tw.SelectAll() })
	edit.AddCommand("Font…", func() {
		if f, ok := dialog.ChooseFont(app, dialog.FontTitle("Editor font")); ok {
			e.tw.Configure(text.FontOpt(f))
		}
	})
	bar.AddCascade("Edit", 0, edit)

	// Status line, scrollbar and text.
	e.status = label.New(app, "status", label.Anchor(option.AnchorW), label.Relief(option.ReliefSunken))
	pack.Pack(e.status, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	sb := scrollbar.New(app, "sb")
	e.tw = text.New(app, "text",
		text.UndoOpt(true),
		text.WrapModeOpt(text.WrapWord),
		text.FontOpt("TkFixedFont"),
		text.YScrollCommand(sb.Set),
	)
	sb.Configure(scrollbar.CommandOpt(func(args ...any) {
		switch args[0] {
		case "moveto":
			e.tw.YViewMoveTo(args[1].(float64))
		case "scroll":
			e.tw.YViewScroll(args[1].(int), args[2] == "pages")
		}
	}))
	pack.Pack(sb, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(e.tw, pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Keyboard shortcuts. Binding on the text's own path runs before its
	// class bindings; returning true (break) stops Text's own <Control-o>
	// (open line) from also running.
	eng := app.Bind()
	path := e.tw.Window().PathName
	eng.Bind(path, "<Control-o>", func(*bind.EventData) bool { e.open(); return true })
	eng.Bind(path, "<Control-s>", func(*bind.EventData) bool { e.save(); return true })
	eng.Bind(path, "<Control-q>", func(*bind.EventData) bool { e.quit(); return true })

	// The window manager's close button.
	app.WmInfo().OnDeleteWindow(e.quit)

	widget.Focus(app, e.tw.Window())
	app.Run()
}
```

Notes:

- Dirty tracking compares the text with the last saved contents — simple and
  always right, including after undo.
- `app.WmInfo().OnDeleteWindow(e.quit)` intercepts the window's close button,
  so the user gets the "unsaved changes" question instead of losing work.
- The dialogs are modal and *return their result*: `open` reads like
  straight-line code even though the event loop keeps running while the
  dialog is up (chapter 11 explains how).

---

## 10. Toplevel windows and the window manager

### The main window

The root window's window-manager state is `app.WmInfo()`:

```go
wm := app.WmInfo()
wm.SetTitle("My App")
wm.SetIconName("myapp")
_ = wm.SetGeometry("800x600+100+50") // "WxH", "+X+Y" or both
wm.SetMinSize(400, 300)
wm.SetMaxSize(1600, 1200)
wm.SetResizable(true, false)          // width, height
wm.Iconify()
wm.Deiconify()
wm.Withdraw()                         // hide completely
g := wm.Geometry()                    // current "WxH+X+Y"
wm.OnDeleteWindow(func() { ... })     // close button (the default quits the app)
```

### More windows

`toplevel.New` creates an additional top-level window. Toplevels start
hidden; call `Show()` when they are ready.

```go
about := toplevel.New(app, "about",
	toplevel.Title("About"),
	toplevel.Geometry("300x200+100+100"),
	toplevel.MinSize(200, 100),
	toplevel.Resizable(false, false),
	toplevel.TransientFor(app), // a dialog-like window that stays above its owner
)
body := label.New(about, "body", label.Text("Made with takigo"))
pack.Pack(body, pack.PadX(20), pack.PadY(20))

about.OnClose(about.Destroy) // or about.Hide to keep it for later
about.Show()
```

A toplevel is a `Caregiver`, so it parents widgets just like the app does, and
has its own `WmInfo` field with the same methods as `app.WmInfo()`. `Hide()`
withdraws it; `Destroy()` destroys it and everything inside. The capstone app
in chapter 17 shows a reusable "About" window.

---

## 11. Dialogs

The `dialog` package has Tk's standard dialogs. All of them are **modal**:
they run a nested event loop until the user answers, then return the answer.
Your code waits, but the application stays alive — other windows repaint,
timers fire.

### Message boxes

```go
res := dialog.ShowMessage(app,
	dialog.MsgTitle("Delete file"),
	dialog.MsgMessage("Delete report.pdf?"),
	dialog.MsgDetail("This cannot be undone."),
	dialog.MsgType(dialog.MsgWarning),     // MsgInfo, MsgWarning, MsgError, MsgQuestion
	dialog.MsgButtons(dialog.BtnYesNo),    // BtnOK, BtnOKCancel, BtnYesNo, BtnYesNoCancel, BtnAbortRetryIgnore
)
if res == dialog.ResultYes {
	// ...
}
```

Results: `ResultOK`, `ResultCancel`, `ResultYes`, `ResultNo`, `ResultAbort`,
`ResultRetry`, `ResultIgnore`. Closing the box counts as `ResultCancel`.

### Files and directories

```go
path, ok := dialog.OpenFile(app,
	dialog.FileTitle("Open image"),
	dialog.FileInitialDir("~/Pictures"),
	dialog.FileTypes(
		dialog.FileType{Name: "Images", Pattern: "*.png *.gif *.jpg"},
		dialog.FileType{Name: "All files", Pattern: "*"},
	),
)

paths, ok := dialog.OpenFiles(app)           // multiple selection

out, ok := dialog.SaveFile(app,
	dialog.FileInitialFile("untitled.txt"),
	dialog.FileDefaultExtension(".txt"),
	dialog.FileConfirmOverwrite(true),
)

dir, ok := dialog.ChooseDirectory(app, dialog.DirTitle("Export to"), dialog.DirMustExist(true))
```

`ok` is `false` when the user cancelled.

### Colours, fonts and short answers

```go
if c, ok := dialog.ChooseColor(app, dialog.ColorInitial("#ff8800")); ok {
	swatch.Configure(label.Background(c)) // c is "#rrggbb"
}

if f, ok := dialog.ChooseFont(app, dialog.FontInitial("Helvetica 12")); ok {
	tw.Configure(text.FontOpt(f)) // f is a font descriptor string
}

if name, ok := dialog.AskString(app, "Rename", "New name:", oldName); ok {
	rename(name)
}
```

### Writing your own modal dialog

The recipe the built-in dialogs follow: build a toplevel, grab input, run a
nested loop until a channel closes, then clean up.

```go
func askNumber(app *takigo.App, prompt string) (int, bool) {
	top := toplevel.New(app, "ask", toplevel.Title("Question"), toplevel.TransientFor(app))
	f := ttk.NewFrame(top, "f", ttk.FramePadding(ttk.UniformPadding(12)))
	pack.Pack(f, pack.FillOpt(pack.FillBoth), pack.Expand(true))
	pack.Pack(ttk.NewLabel(f, "l", ttk.LabelText(prompt)))
	sb := ttk.NewSpinbox(f, "sb", ttk.SpinboxFrom(0), ttk.SpinboxTo(100))
	pack.Pack(sb, pack.PadY(6))

	done := make(chan struct{})
	result, ok := 0, false
	finish := func(accept bool) {
		select {
		case <-done:
			return // already finished
		default:
		}
		if accept {
			if n, err := strconv.Atoi(sb.Get()); err == nil {
				result, ok = n, true
			}
		}
		close(done)
	}
	pack.Pack(ttk.NewButton(f, "ok", ttk.ButtonText("OK"), ttk.ButtonCommand(func() { finish(true) })),
		pack.SideOpt(pack.Left))
	pack.Pack(ttk.NewButton(f, "cancel", ttk.ButtonText("Cancel"), ttk.ButtonCommand(func() { finish(false) })),
		pack.SideOpt(pack.Right))
	top.OnClose(func() { finish(false) })

	top.Show()
	app.GrabManager().Set(top.Window(), false) // local grab: other windows ignore input
	widget.Focus(app, sb.Window())
	app.RunNestedLoop(done)                    // returns when done is closed
	app.GrabManager().Release()
	top.Destroy()
	return result, ok
}
```

`RunNestedLoop` must be called from the event loop (for example from a button
command), which is where dialogs are normally opened anyway.
`app.RunNestedLoopContext(ctx, done)` also returns when `ctx` is cancelled.

---

## 12. The text widget

`widget/text` is a full multi-line editor: tags for styling, marks, undo,
embedded widgets and images, and peers.

### Creating and filling

```go
tw := text.New(app, "t",
	text.Width(80), text.Height(24),        // characters, lines
	text.WrapModeOpt(text.WrapWord),        // WrapNone, WrapChar, WrapWord
	text.FontOpt("TkFixedFont"),
	text.UndoOpt(true),
	text.TabWidth(4),
	text.PadXOpt(screenunit.Mm(2)),
)
tw.Insert("end", "Hello\nworld\n")
tw.Insert("1.0", ">> ")
tw.Delete("1.0", "1.3")
first := tw.Get("1.0", "1.end")
all := tw.Get("1.0", "end")
```

### Indices

An index names a position between characters. It has the same grammar as in
Tk: a *base*, optionally followed by *modifiers*.

| Base | Meaning |
|------|---------|
| `"3.0"` | Line 3 (1-based), character 0 (0-based) — the start of line 3. |
| `"3.end"` | End of line 3 (before its newline). |
| `"end"` | End of the text. |
| `"insert"` | The insertion cursor. |
| any mark name | The mark's position. |
| `"tag.first"`, `"tag.last"` | Start of the first / end of the last range of a tag, e.g. `"sel.first"`. Invalid when the tag has no ranges. |
| `"@x,y"` | The character nearest the pixel `(x, y)` in the widget. |
| an embedded window's path | Its position, e.g. `".t.btn"`. |

| Modifier | Meaning |
|----------|---------|
| `+ n chars`, `- n chars` | Move `n` characters (a line break counts as one). Also `indices`; `any` and `display` variants are accepted. |
| `+ n lines`, `- n lines` | Same character offset, `n` lines down or up. |
| `+ n display lines` | Same x position, `n` wrapped display lines down (or up with `-`). |
| `linestart`, `lineend` | Start or end of the line; with `display`, of the wrapped display line. |
| `wordstart`, `wordend` | Start of the word at the index, or just past its end. |

Units can be abbreviated as in Tk (`"insert +1c"`, `"insert -2l"`), spaces are
optional, and modifiers chain left to right:

```go
tw.Delete("insert linestart", "insert lineend")    // clear the current line
word := tw.Get("insert wordstart", "insert wordend")
tw.TagAdd("hl", "sel.first", "sel.first +5c")
tw.See("end -1 lines")
body := tw.Get("1.0", "end -1c")                   // everything, as in Tcl
```

As in Tk, `"end"` sits after an implicit final newline, so `"end -1c"` is the
end of the content. Here, `Get("1.0", "end")` and `Get("1.0", "end -1c")`
return the same text: a Go document has no trailing newline to include.

The same arithmetic is available on `text.Index{Line, Char}` values through
the document (`tw.Doc()`): `text.ParseIndex(doc, spec)` resolves an index
string, and `text.Forward`, `text.Backward`, `text.LineStart` and
`text.LineEnd` move positions directly. `tw.EndIndex()` returns the end
position as an index string, which is convenient for "remember where this
insert started":

```go
start := tw.EndIndex()
tw.Insert("end", "important")
tw.TagAdd("bold", start, tw.EndIndex())
```

### Tags

A tag is a named set of character ranges with display options.

```go
tw.TagConfigure("bold", text.TagFont("TkFixedFont 12 bold"))
tw.TagConfigure("warn", text.TagForeground("red3"), text.TagBackground("#fff0f0"))
tw.TagConfigure("link", text.TagForeground("blue"), text.TagUnderline(true))
tw.TagConfigure("quote",
	text.TagLMargin1(screenunit.Mm(12)), text.TagLMargin2(screenunit.Mm(12)),
	text.TagRMargin(screenunit.Mm(10)),
	text.TagSpacing1(screenunit.Pt(4)),
)
tw.TagConfigure("center", text.TagJustify(option.JustifyCenter))
tw.TagConfigure("super", text.TagOffset(screenunit.Pt(4)))

tw.TagAdd("bold", "1.0", "1.5")
tw.TagRemove("bold", "1.0", "1.5")
```

The full set of tag options mirrors Tk's: font, colours, underline,
overstrike, justify, margins (`LMargin1`, `LMargin2`, `RMargin`), spacing
(`Spacing1/2/3`), offset, relief and border width, and stipples
(`TagBgStipple`, `TagFgStipple`). The `...Str` variants take Tk distances.
The `style` demo (`demos/style/main.go`) shows every one.

Tags can react to the mouse — the basis of hyperlinks:

```go
tw.TagBind("link", "<Button-1>", func() { openURL() })
tw.TagBind("link", "<Enter>", func() { /* change cursor / colour */ })
tw.TagBind("link", "<Leave>", func() { /* restore */ })
```

The selection is the special tag `"sel"`: `tw.HasSelection()`,
`tw.GetSelection()`, `tw.SelectAll()`.

### Marks

A mark is a named position that moves with the text as you edit around it.
`"insert"` is the cursor; you can make your own:

```go
tw.MarkSet("bookmark", "12.0")
tw.MarkSet("insert", "1.0") // move the cursor
tw.MarkGravity("bookmark", text.GravityLeft)
tw.See("bookmark")          // scroll it into view
```

### Undo and editing

With `text.UndoOpt(true)` the widget records edits. Ctrl+Z and Ctrl+Shift+Z
work out of the box; from code:

```go
tw.Edit("undo")
tw.Edit("redo")
tw.Edit("separator") // close the current undo group
tw.Edit("reset")     // forget history (e.g. after loading a file)
```

`text.ReadOnly(true)` turns the widget into a viewer (selection and copying
still work).

### Embedded widgets and images

```go
btn := button.New(tw, "more", button.Text("More…"), button.Command(showMore))
tw.WindowCreate("end", btn.Window())   // create the widget as a child of the text
tw.ImageCreate("end", img)             // any image, see chapter 14
```

As in Tk, each embedded window or image occupies one index position: text
flows around it and it moves with edits before it. Its path (for a window)
or its image name can be used as an index, e.g. `tw.See(".t.more")`.
Deleting its position removes it; for a window that destroys the widget,
while `tw.RemoveWindow(w)` only takes it out of the text. `Get` and the
selection return just the characters, and undoing a deletion brings back
the text but not the windows or images that were in it.

### Peers

Several text widgets can share one document, each with its own view, cursor
and scroll position — a split-view editor in one line:

```go
peer := text.NewPeer(tw.Doc(), app, "peer", text.Height(10))
```

The editor example in chapter 9 puts most of this together.

---

## 13. The canvas

`canvas` is a 2-D drawing surface of *items* — shapes, text, images, even
other widgets — that you can move, restyle and bind events to after creating
them.

### Items

```go
c := canvas.New(app, "c",
	canvas.Width(screenunit.Cm(12)), canvas.Height(screenunit.Cm(8)),
	canvas.Background("white"),
)
pack.Pack(c, pack.FillOpt(pack.FillBoth), pack.Expand(true))

rect := c.CreateRectangle(10, 10, 110, 60,
	canvas.FillColor("light blue"), canvas.OutlineColor("navy"), canvas.OutlineWidth(2))
c.CreateOval(150, 10, 250, 110, canvas.FillColor("gold"))
c.CreateLine([]float64{10, 150, 100, 120, 200, 170},
	canvas.WidthOpt(3), canvas.Smooth(true), canvas.Arrow(canvas.ArrowLast))
c.CreatePolygon([]float64{300, 10, 350, 90, 250, 90},
	canvas.FillColor("pale green"), canvas.OutlineColor("black"))
c.CreateArc(300, 120, 400, 220,
	canvas.StartAngle(0), canvas.Extent(270), canvas.ArcStyleOpt(canvas.ArcStylePieslice),
	canvas.FillColor("tomato"))
c.CreateText(10, 250, canvas.TextOpt("Hello"), canvas.AnchorOpt(option.AnchorNW),
	canvas.FontOpt("Helvetica 18 bold"), canvas.TextColor("gray20"))
c.CreateImage(450, 50, canvas.ImageOpt(img))

btn := button.New(c, "b", button.Text("I'm a widget"))
c.CreateWindow(450, 200, btn.Window())
```

Every `Create…` returns an `int64` item ID. Common item options:

| Option | Applies to |
|--------|-----------|
| `FillColor`, `FillNone`, `OutlineColor`, `OutlineNone`, `OutlineWidth`, `WidthOpt` | shapes and lines |
| `Dash(4, 2)`, `Stipple("gray50")`, `OutlineStipple` | patterns |
| `ActiveFill`, `DisabledFill` | colour while the pointer is over the item / when disabled |
| `Arrow(canvas.ArrowNone\|ArrowFirst\|ArrowLast\|ArrowBoth)`, `ArrowShape`, `Smooth`, `SplineSteps`, `CapStyleOpt`, `JoinStyleOpt` | lines |
| `StartAngle`, `Extent`, `ArcStyleOpt(ArcStylePieslice\|ArcStyleChord\|ArcStyleArc)` | arcs |
| `TextOpt`, `FontOpt`, `TextColor`, `TextAngle`, `JustifyOpt`, `AnchorOpt` | text |
| `ImageOpt`, `AnchorOpt` | images |
| `Tags("a", "b")` | all |
| `StateOpt(canvas.ItemStateNormal\|ItemStateDisabled\|ItemStateHidden)` | all |

### IDs and tags

Most methods take a `tagOrID` string: an item ID (`fmt.Sprint(id)`), a tag
name, `"all"`, or `"current"` (the item under the pointer).

```go
c.Move("enemy", 5, 0)                // every item tagged "enemy"
c.MoveTo(fmt.Sprint(rect), 50, 50)
c.Scale("all", 0, 0, 2, 2)           // origin x, y; factors x, y
c.Raise("selected")
c.Lower("background")
c.AddTag("selected", "current")
c.DeleteTag("selected", "all")
_ = c.ItemConfigure("selected", canvas.FillColor("orange"))
coords := c.ItemCoords(fmt.Sprint(rect))
_ = c.SetItemCoords(fmt.Sprint(rect), []float64{0, 0, 40, 40})
c.Delete("temp")
```

Finding items:

```go
ids := c.FindWithTag("enemy")
hits := c.Find("overlapping", x1, y1, x2, y2) // also "enclosed", "closest", "all"
nearest := c.FindClosest(x, y, 0, "")
x1, y1, x2, y2 := c.BBox("all")
tags := c.GetTags("current")
```

### Item events

`BindItem(tagOrID, mask, handler)` attaches a handler to items; the canvas
keeps track of which item is under the pointer and delivers `Enter`, `Leave`,
button and motion events to it.

```go
c.BindItem("shape", event.EnterMask, func(*event.Event) {
	_ = c.ItemConfigure("current", canvas.OutlineWidth(3))
})
c.BindItem("shape", event.LeaveMask, func(*event.Event) {
	_ = c.ItemConfigure("current", canvas.OutlineWidth(1))
})
```

Event coordinates are window coordinates; convert them with `c.CanvasX(ev.X)`
and `c.CanvasY(ev.Y)` whenever the canvas can scroll.

### Example: a sketch pad

Drag on empty space to draw a box, drag any shape to move it, right-click to
delete it.

```go
package main

import (
	"fmt"
	"log"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Sketch"))
	if err != nil {
		log.Fatal(err)
	}

	hint := label.New(app, "hint",
		label.Text("Drag on empty space to draw a box · drag a shape to move it · right-click deletes"))
	pack.Pack(hint, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	c := canvas.New(app, "c",
		canvas.Width(screenunit.Cm(12)), canvas.Height(screenunit.Cm(8)),
		canvas.Background("white"),
	)
	pack.Pack(c, pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// A few static items.
	c.CreateText(10, 10,
		canvas.TextOpt("Sketch pad"),
		canvas.AnchorOpt(option.AnchorNW),
		canvas.FontOpt("Helvetica 14 bold"),
	)
	c.CreateLine([]float64{20, 200, 120, 150, 220, 200, 320, 150},
		canvas.Smooth(true), canvas.WidthOpt(3), canvas.FillColor("steel blue"),
		canvas.Arrow(canvas.ArrowLast), canvas.Tags("shape"))
	c.CreateOval(250, 40, 330, 120,
		canvas.FillColor("gold"), canvas.OutlineColor("orange"), canvas.OutlineWidth(2),
		canvas.ActiveFill("yellow"), canvas.Tags("shape"))

	// Pressing on any item tagged "shape" picks it up (button 1) or
	// deletes it (button 3). The drag itself is handled below on the canvas
	// window, together with drawing, so remember which item was picked up.
	dragID := int64(-1)
	var lastX, lastY float64
	c.BindItem("shape", event.ButtonPressMask, func(ev *event.Event) {
		switch ev.Button {
		case 1:
			dragID = c.CurrentItem()
			lastX, lastY = c.CanvasX(ev.X), c.CanvasY(ev.Y)
		case 3:
			c.Delete("current")
		}
	})

	// Everything else is handled on the canvas window itself: dragging,
	// and drawing a new box when the press was on empty space.
	var startX, startY float64
	rubber := int64(-1)
	n := 0
	app.Dispatcher().Bind(c.Window().PlatformID,
		event.ButtonPressMask|event.MotionMask|event.ButtonReleaseMask,
		func(ev *event.Event) {
			x, y := c.CanvasX(ev.X), c.CanvasY(ev.Y)
			switch ev.Type {
			case event.ButtonPressType:
				if ev.Button != 1 || c.CurrentItem() != -1 {
					return
				}
				startX, startY = x, y
				rubber = c.CreateRectangle(x, y, x, y,
					canvas.OutlineColor("gray40"), canvas.Dash(4, 4))
			case event.MotionType:
				if ev.State&platform.Button1Mask == 0 {
					return
				}
				switch {
				case dragID != -1:
					c.Move(fmt.Sprint(dragID), x-lastX, y-lastY)
					lastX, lastY = x, y
				case rubber != -1:
					_ = c.SetItemCoords(fmt.Sprint(rubber), []float64{startX, startY, x, y})
				}
			case event.ButtonReleaseType:
				dragID = -1
				if rubber == -1 {
					return
				}
				coords := c.ItemCoords(fmt.Sprint(rubber))
				c.Delete(fmt.Sprint(rubber))
				rubber = -1
				n++
				c.CreateRectangle(coords[0], coords[1], coords[2], coords[3],
					canvas.FillColor([]string{"tomato", "pale green", "light blue"}[n%3]),
					canvas.ActiveFill("white"),
					canvas.Tags("shape", fmt.Sprintf("box%d", n)))
			}
		})

	app.Run()
}
```

How it works:

- The canvas registers its own window handlers when it is created, so they run
  before ours: by the time our `ButtonPress` handler runs, the item binding
  has already recorded `dragID`, and `CurrentItem()` tells us whether the
  press was on an item (`-1` means empty space).
- As in Tk, the item a button was pressed on stays "current" until the button
  is released, however fast the pointer moves: it keeps getting `<Motion>` and
  gets the `<ButtonRelease>`, and no other item is entered meanwhile. So a
  drag could equally be written as an item binding that moves `"current"`;
  here one window handler does both dragging and drawing.
- The rubber band is an ordinary item that is resized on every motion and
  replaced by a filled rectangle on release.

### Scrolling a large canvas

Give the canvas a `ScrollRegion` larger than its window and connect scrollbars
as in chapter 8 (`XScrollCommand`/`YScrollCommand` on the canvas;
`XViewMoveTo`/`XViewScroll`/`YViewMoveTo`/`YViewScroll` for the scrollbar
commands):

```go
c := canvas.New(f, "c", canvas.Width(400), canvas.Height(300),
	canvas.ScrollRegion(0, 0, 2000, 2000),
	canvas.XScrollCommand(hsb.Set), canvas.YScrollCommand(vsb.Set))
```

### PostScript output

```go
if _, err := c.Postscript(canvas.PSFile("drawing.ps"), canvas.PSColorMode("color")); err != nil {
	log.Print(err)
}
```

Other options set the region, page size and position, rotation and title
(`PSRegion`, `PSPageWidth`, `PSRotate`, `PSTitle`, …); `PSWriter(w)` writes to
any `io.Writer`, and with neither option the PostScript is returned as a
string.

---

## 14. Images

Images are `*image.Photo` values (package `github.com/msorc/takigo/image`;
import it under another name if you also use the standard `image` package).
Create one, register it with the app, and pass it to any widget that shows
images.

```go
import (
	goimage "image"
	gocolor "image/color"

	"github.com/msorc/takigo/image"
)

img, err := image.NewPhotoFromFile("earth", "demos/images/earth.gif")
if err != nil {
	log.Fatal(err)
}
app.ImageRegistry().Register(img)

pack.Pack(label.New(app, "pic",
	label.ImageOpt(img), label.Text("Earth"), label.CompoundOpt(widget.CompoundTop)))
```

`NewPhotoFromFile` picks the decoder from the extension: PNG, GIF and JPEG
(through the standard library), XBM, PPM/PGM/PBM, and SVG (rendered by a Go
port of Tk's own SVG rasteriser). There is also `NewPhotoFromReader` for
embedded data (`//go:embed`) and `NewPhotoFromXBM` with custom foreground and
background colours.

Registering hands ownership to the app: registered images are freed when the
app is destroyed, and `app.ImageRegistry().Get(name)` finds them by name.

**From Go images.** Anything you can draw into an `*image.RGBA` can be shown:

```go
rgba := goimage.NewRGBA(goimage.Rect(0, 0, 256, 32))
for x := range 256 {
	for y := range 32 {
		rgba.Set(x, y, gocolor.RGBA{uint8(x), 0, uint8(255 - x), 255})
	}
}
grad := image.NewPhoto("gradient", rgba)
app.ImageRegistry().Register(grad)
```

If you change the pixels of a photo that is already displayed (through
`p.RGBA()`), call `p.Invalidate()` and redraw the widgets that show it.

**Scaling.** `image.NewPhotoFromPhoto(src, "big", image.Zoom(2))` makes a
zoomed copy.

**Where images go:** `label.ImageOpt`, `button.ImageOpt`,
`checkbutton.ImageOpt`/`SelectImageOpt`, `ttk.ButtonImage`, `ttk.LabelImage`,
`ttk.ItemImage` (treeview rows), `menu.AddCommandImage`,
`canvas.ImageOpt`, and `text.ImageCreate`. Combine text and image with the
widget's `Compound` option.

---

## 15. Timers, idle work and goroutines

### One loop, one goroutine

All UI work happens on the **event loop goroutine** — the one that calls
`app.Run()`. Widgets, windows, variables, geometry managers and the binding
engine are not synchronised; touching them from another goroutine is a data
race. `go test -race` catches such races when a test exercises them.

A handful of `App` methods *are* safe from any goroutine because they only
queue work for the loop:

| Method | Does |
|--------|------|
| `app.RunOnMain(fn)` | Run `fn` on the loop as soon as possible. |
| `app.After(d, fn)` | Run `fn` on the loop after `d`; returns a cancel function. |
| `app.DoWhenIdle(fn)` | Run `fn` when the loop is idle (after pending events and redraws). |
| `app.Quit()` | Stop the loop. |
| `app.Done()` | A channel closed once `Quit` has been called. |

`THREADING.md` in the repository lists exactly which types and methods are
goroutine-safe.

### Timers

```go
cancel := app.After(3*time.Second, func() { status.Configure(label.Text("")) })
// ...
cancel() // like "after cancel"; reports whether it was still pending
```

A repeating timer re-schedules itself:

```go
var tick func()
tick = func() {
	clock.SetText(time.Now().Format("15:04:05"))
	app.After(time.Second, tick)
}
tick()
```

### Background work

Do slow work (network, disk, computation) in a goroutine and post results
back with `RunOnMain`:

```go
go func() {
	data, err := fetch(url) // no widget calls here
	app.RunOnMain(func() {
		if err != nil {
			status.Configure(label.Text(err.Error()))
			return
		}
		show(data)
	})
}()
```

To read UI state from a goroutine, send it back over a channel — and also
wait on `app.Done()`, because callbacks still queued when the app quits never
run:

```go
res := make(chan int, 1)
app.RunOnMain(func() { res <- lb.ItemCount() })
select {
case n := <-res:
	use(n)
case <-app.Done():
}
```

Never do this *on* the loop goroutine: the callback can't run until your code
returns, so you would wait forever.

### Example: a cancellable worker

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Worker"))
	if err != nil {
		log.Fatal(err)
	}

	f := ttk.NewFrame(app, "f", ttk.FramePadding(ttk.UniformPadding(12)))
	pack.Pack(f, pack.FillOpt(pack.FillBoth), pack.Expand(true))

	clock := ttk.NewLabel(f, "clock")
	pack.Pack(clock, pack.FillOpt(pack.FillX))

	var tick func()
	tick = func() {
		clock.SetText(time.Now().Format("15:04:05"))
		app.After(time.Second, tick)
	}
	tick()

	bar := ttk.NewProgressbar(f, "bar", ttk.ProgressbarLength(300), ttk.ProgressbarMaximum(100))
	pack.Pack(bar, pack.FillOpt(pack.FillX), pack.PadY(8))

	status := ttk.NewLabel(f, "status", ttk.LabelText("Idle"))
	pack.Pack(status, pack.FillOpt(pack.FillX))

	var cancel context.CancelFunc
	var start, stop *ttk.Button

	start = ttk.NewButton(f, "start", ttk.ButtonText("Start"), ttk.ButtonCommand(func() {
		if cancel != nil {
			return
		}
		var ctx context.Context
		ctx, cancel = context.WithCancel(context.Background())
		status.SetText("Working…")

		go func() {
			// Background goroutine: no widget calls here.
			for i := 1; i <= 100; i++ {
				select {
				case <-ctx.Done():
					app.RunOnMain(func() { status.SetText("Cancelled") })
					return
				case <-time.After(50 * time.Millisecond):
				}
				app.RunOnMain(func() { bar.SetValue(float64(i)) })
			}
			app.RunOnMain(func() {
				status.SetText(fmt.Sprintf("Done at %s", time.Now().Format("15:04:05")))
				cancel = nil
			})
		}()
	}))
	stop = ttk.NewButton(f, "stop", ttk.ButtonText("Stop"), ttk.ButtonCommand(func() {
		if cancel != nil {
			cancel()
			cancel = nil
		}
	}))
	pack.Pack(start, pack.SideOpt(pack.Left), pack.PadY(8))
	pack.Pack(stop, pack.SideOpt(pack.Left), pack.PadX(8), pack.PadY(8))

	app.Run()
}
```

`cancel` is read and written only from button commands and `RunOnMain`
callbacks — all on the loop goroutine — so it needs no mutex. The goroutine
only touches its own `ctx` and posts closures.

### Platform notes

- **macOS:** `NewApp` must be called on the main goroutine (AppKit requires
  the main thread), and `Run` runs there. Just call both from `main`.
- **Windows:** `NewApp` locks its goroutine to the OS thread that owns the
  windows; call `Run` from that same goroutine.
- **X11:** one process may run several Apps, each on its own goroutine with its
  own display connection. Each App's widgets belong to that App's loop.

---

## 16. Themed widgets (ttk)

The `ttk` package contains Tk's themed widgets. They take their look from a
*theme* and *styles* instead of per-widget colour options, which keeps an
application consistent and lets you restyle it in one place.

### Themes

Themes are separate packages that register themselves when imported. **Import
at least one** — without a theme, ttk widgets have no layout and draw
nothing:

```go
import (
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/defaulttheme" // "default"
	_ "github.com/msorc/takigo/ttk/clamtheme"    // "clam"
	_ "github.com/msorc/takigo/ttk/alttheme"     // "alt"
	_ "github.com/msorc/takigo/ttk/classictheme" // "classic"
)

func main() {
	ttk.SetCurrentTheme("clam") // before creating ttk widgets
	...
}
```

`ttk.ThemeNames()` lists the registered themes, `ttk.CurrentTheme()` returns
the active one. The first theme registered becomes current, so call
`SetCurrentTheme` explicitly when you import more than one.

Changing the theme while widgets exist is possible, but existing widgets have
to be told: call `w.RefreshTheme()` and then `w.Display()` on each (the
`ttkbut` demo does this for a theme-switcher row).

### The widgets

| Constructor | Tk | Notes |
|-------------|----|-------|
| `ttk.NewFrame` | `ttk::frame` | `FramePadding(ttk.UniformPadding(8))`, `FrameRelief`, `FrameBorderWidth` |
| `ttk.NewLabelframe` | `ttk::labelframe` | `LabelframeText`, `LabelframePadding("8 4")` |
| `ttk.NewLabel` | `ttk::label` | `LabelText`, `LabelTextVariable`, `LabelImage`, `LabelWrapLength`, `LabelPadding`; `SetText` |
| `ttk.NewButton` | `ttk::button` | `ButtonText`, `ButtonCommand`, `ButtonImage`, `ButtonWidth`, `ButtonStyleOpt`; `Invoke` |
| `ttk.NewCheckbutton` | `ttk::checkbutton` | `CheckbuttonVar(*Variable[bool])`, `CheckbuttonCommand` |
| `ttk.NewRadiobutton` | `ttk::radiobutton` | `RadiobuttonVar(*Variable[string])`, `RadiobuttonValue` |
| `ttk.NewToggleswitch` | `ttk::toggleswitch` | `ToggleswitchVar(*Variable[bool])` |
| `ttk.NewEntry` | `ttk::entry` | `EntryTextVariable`, `EntryPlaceholder`, `EntryShow`, `EntryValidate`/`EntryValidateCmd`/`EntryInvalidCmd`; `Get`, `Set` |
| `ttk.NewCombobox` | `ttk::combobox` | `ComboboxValues`, `ComboboxText`, `ComboboxCbState(ttk.ComboReadonly)`, `ComboboxCommand`; `Get`, `Set` |
| `ttk.NewSpinbox` | `ttk::spinbox` | `SpinboxFrom/To/Increment/Values/Wrap/Format/Command`; `Get` |
| `ttk.NewScale` | `ttk::scale` | `ScaleFrom`, `ScaleTo`, `ScaleVariable(*Variable[float64])`, `ScaleLength`, `ScaleCommand` |
| `ttk.NewProgressbar` | `ttk::progressbar` | `ProgressbarMaximum`, `ProgressbarMode(ttk.ProgressIndeterminate)`; `SetValue`, `Step`, `Start(interval)`, `Stop` |
| `ttk.NewScrollbar` | `ttk::scrollbar` | `ScrollbarOrientOpt(ttk.Vertical)`, `ScrollbarCommandOpt`; `Set` |
| `ttk.NewSeparator` | `ttk::separator` | `SeparatorOrient(ttk.Horizontal)` |
| `ttk.NewSizegrip` | `ttk::sizegrip` | bottom-right resize handle |
| `ttk.NewMenubutton` | `ttk::menubutton` | `MenubuttonText`, `MenubuttonMenu` |
| `ttk.NewNotebook` | `ttk::notebook` | `Add(pane.Window(), "Title")`, `Select`, `Selected`, `SetTabState` |
| `ttk.NewPanedwindow` | `ttk::panedwindow` | same API as the classic panedwindow |
| `ttk.NewTreeview` | `ttk::treeview` | trees and tables (below) |

There's no themed listbox or text widget in Tk either; use the classic ones
alongside ttk widgets.

### Notebook

```go
nb := ttk.NewNotebook(root, "nb")
pack.Pack(nb, pack.FillOpt(pack.FillBoth), pack.Expand(true))

general := ttk.NewFrame(nb, "general", ttk.FramePadding(ttk.UniformPadding(8)))
nb.Add(general.Window(), "General")
advanced := ttk.NewFrame(nb, "advanced")
nb.Add(advanced.Window(), "Advanced")

nb.SetTabUnderline(0, 0)             // Alt+G selects the first tab
nb.SetTabState(1, ttk.StateDisabled)
nb.Select(0)
```

Pages are created with the notebook as their parent and added with `Add`.

### Treeview

A treeview shows a hierarchy (`#0` is the tree column) with optional data
columns — or, with `TreeviewShow("headings")`, a plain multi-column table.

```go
tv := ttk.NewTreeview(page, "tv",
	ttk.TreeviewColumns("size", "modified"),
	ttk.TreeviewShow("tree", "headings"),
	ttk.TreeviewHeight(12),
	ttk.TreeviewSelectMode(ttk.TreeSelectBrowse),
)
tv.HeadingConfigure("#0", ttk.HeadText("Name"))
tv.HeadingConfigure("size", ttk.HeadText("Size"), ttk.HeadCommand(sortBySize))
tv.ColumnConfigure("size", ttk.ColWidth(80), ttk.ColAnchor(option.AnchorE))

dir := tv.Insert("", -1, ttk.ItemText("src"), ttk.ItemOpen(true)) // "" = root, -1 = append
tv.Insert(dir, -1, ttk.ItemID("src/main.go"), ttk.ItemText("main.go"),
	ttk.ItemValues("2.1 KB", "today"), ttk.ItemImage(fileIcon))

tv.OnSelect = func() { log.Print(tv.Selection()) }
tv.OnOpen = func(id string) { loadChildrenLazily(id) }
tv.OnDoubleClick = func(id string) { openItem(id) }
```

Items are addressed by string IDs (generated unless you pass `ItemID`).
Other operations: `Delete`, `Move`, `Children`, `Parent`, `Item(id)` (returns
the `*ttk.TreeItem` with `Text`, `Values`, `Open`, …), `SetItemText`,
`SetItemValues`, `SetItemOpen`, `SelectionSet/Add/Remove`, `See`, `SortChildren`,
`SetSortIndicator`, `SetStripe`. Scroll it like any other widget
(`TreeviewYScrollCommand`, `YViewMoveTo`, `YViewScroll`).

### Styles

A style is a named set of option values, looked up by the widget at draw
time, with per-state overrides. Style names follow Tk's inheritance rule:
`"Accent.TButton"` inherits everything from `"TButton"`, which inherits from
the root style `"."`.

```go
accent := ttk.CurrentTheme().GetStyle("Accent.TButton") // created on first use

if c, err := app.ColorCache().Get("royal blue"); err == nil {
	accent.Defaults["-background"] = c.Pixel
}
if c, err := app.ColorCache().Get("white"); err == nil {
	accent.Defaults["-foreground"] = c.Pixel
}
if c, err := app.ColorCache().Get("dodger blue"); err == nil {
	accent.Maps["-background"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateActive}, Value: c.Pixel},
	}
}

ok := ttk.NewButton(f, "ok", ttk.ButtonText("OK"), ttk.ButtonStyleOpt("Accent.TButton"))
```

Details worth knowing:

- Option names keep Tk's leading dash: `"-background"`, `"-foreground"`,
  `"-padding"`, `"-relief"`, `"-borderwidth"`, `"-anchor"`, …
- Colour values are **pixel values** (`uint64`), not names. Get them from the
  app's colour cache as above.
- Other values use Go types: `option.Relief`, `option.Anchor`, `int` for
  widths, a padding string like `"4 2"` or a `ttk.Padding` for padding.
- `Maps` are consulted before `Defaults` along the whole inheritance chain,
  and the first matching entry wins, as with `ttk::style map`. If a parent
  style maps an option for a state (the root maps `-background` for `active`),
  add your own map entry for that state or the parent's will show through.
- State flags: `StateActive`, `StateDisabled`, `StateFocus`, `StatePressed`,
  `StateSelected`, `StateBackground`, `StateAlternate`, `StateInvalid`,
  `StateReadonly`, `StateHover`. `StateSpec{OnBits: a, OffBits: b}` matches
  "all of `a` set, none of `b` set" (Tk's `{a !b}`).
- Fill in styles before creating widgets that use them. A widget's own state
  can be changed with `w.ChangeState(set, clear)`.
- For a single widget, `w.SetWidgetOption("-padding", "8 4")` overrides the
  style (like passing `-padding` to that widget in Tk).

### Example: a themed settings window

```go
package main

import (
	"log"
	"strconv"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget"
)

type planet struct {
	name   string
	moons  int
	radius int // km
}

func main() {
	ttk.SetCurrentTheme("clam") // before creating any ttk widget

	app, err := takigo.NewApp(takigo.Title("Themed widgets"))
	if err != nil {
		log.Fatal(err)
	}

	// A custom style: "Accent.TButton" inherits everything from "TButton".
	accent := ttk.CurrentTheme().GetStyle("Accent.TButton")
	if c, err := app.ColorCache().Get("royal blue"); err == nil {
		accent.Defaults["-background"] = c.Pixel
	}
	if c, err := app.ColorCache().Get("white"); err == nil {
		accent.Defaults["-foreground"] = c.Pixel
	}
	if c, err := app.ColorCache().Get("dodger blue"); err == nil {
		accent.Maps["-background"] = ttk.StateMap[any]{
			{Spec: ttk.StateSpec{OnBits: ttk.StateActive}, Value: c.Pixel},
		}
	}

	root := ttk.NewFrame(app, "root", ttk.FramePadding(ttk.UniformPadding(8)))
	pack.Pack(root, pack.FillOpt(pack.FillBoth), pack.Expand(true))

	nb := ttk.NewNotebook(root, "nb")
	pack.Pack(nb, pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// --- Tab 1: a treeview with columns ---
	page1 := ttk.NewFrame(nb, "planets", ttk.FramePadding(ttk.UniformPadding(6)))
	nb.Add(page1.Window(), "Planets")

	var tv *ttk.Treeview
	vsb := ttk.NewScrollbar(page1, "vsb", ttk.ScrollbarOrientOpt(ttk.Vertical))
	tv = ttk.NewTreeview(page1, "tv",
		ttk.TreeviewColumns("moons", "radius"),
		ttk.TreeviewShow("tree", "headings"),
		ttk.TreeviewHeight(8),
		ttk.TreeviewYScrollCommand(vsb.Set),
	)
	vsb.Configure(ttk.ScrollbarCommandOpt(func(args ...any) {
		switch args[0] {
		case "moveto":
			tv.YViewMoveTo(args[1].(float64))
		case "scroll":
			tv.YViewScroll(args[1].(int), args[2] == "pages")
		}
	}))
	tv.HeadingConfigure("#0", ttk.HeadText("Body"))
	tv.HeadingConfigure("moons", ttk.HeadText("Moons"))
	tv.HeadingConfigure("radius", ttk.HeadText("Radius (km)"))
	tv.ColumnConfigure("moons", ttk.ColWidth(70), ttk.ColAnchor(option.AnchorE))
	tv.ColumnConfigure("radius", ttk.ColWidth(100), ttk.ColAnchor(option.AnchorE))

	groups := map[string][]planet{
		"Inner planets": {{"Mercury", 0, 2440}, {"Venus", 0, 6052}, {"Earth", 1, 6371}, {"Mars", 2, 3390}},
		"Gas giants":    {{"Jupiter", 95, 69911}, {"Saturn", 146, 58232}},
		"Ice giants":    {{"Uranus", 28, 25362}, {"Neptune", 16, 24622}},
	}
	for _, g := range []string{"Inner planets", "Gas giants", "Ice giants"} {
		parent := tv.Insert("", -1, ttk.ItemText(g), ttk.ItemOpen(true))
		for _, p := range groups[g] {
			tv.Insert(parent, -1,
				ttk.ItemID(p.name),
				ttk.ItemText(p.name),
				ttk.ItemValues(strconv.Itoa(p.moons), strconv.Itoa(p.radius)),
			)
		}
	}

	grid.Grid(tv, grid.Row(0), grid.Column(0), grid.Sticky(grid.NSEW))
	grid.Grid(vsb, grid.Row(0), grid.Column(1), grid.Sticky(grid.NS))
	grid.ColumnConfigure(page1, 0, grid.Weight(1))
	grid.RowConfigure(page1, 0, grid.Weight(1))

	selected := ttk.NewLabel(page1, "sel", ttk.LabelText("Select a planet"))
	grid.Grid(selected, grid.Row(1), grid.Column(0), grid.ColumnSpan(2), grid.Sticky(grid.EW))
	tv.OnSelect = func() {
		if sel := tv.Selection(); len(sel) == 1 && tv.Parent(sel[0]) != "" {
			item := tv.Item(sel[0])
			selected.SetText(item.Text + " has " + item.Values[0] + " moons")
		}
	}

	// --- Tab 2: settings ---
	page2 := ttk.NewFrame(nb, "settings", ttk.FramePadding(ttk.UniformPadding(12)))
	nb.Add(page2.Window(), "Settings")

	notify := widget.NewVariable(true)
	pack.Pack(ttk.NewCheckbutton(page2, "notify",
		ttk.CheckbuttonText("Enable notifications"), ttk.CheckbuttonVar(notify)),
		pack.Anchor(option.AnchorW))

	dark := widget.NewVariable(false)
	pack.Pack(ttk.NewToggleswitch(page2, "dark",
		ttk.ToggleswitchText("Pretend dark mode"), ttk.ToggleswitchVar(dark)),
		pack.Anchor(option.AnchorW), pack.PadY(6))

	units := widget.NewVariable("metric")
	for _, u := range []string{"metric", "imperial"} {
		pack.Pack(ttk.NewRadiobutton(page2, u,
			ttk.RadiobuttonText(u), ttk.RadiobuttonValue(u), ttk.RadiobuttonVar(units)),
			pack.Anchor(option.AnchorW))
	}

	volume := widget.NewVariable(50.0)
	pack.Pack(ttk.NewScale(page2, "volume",
		ttk.ScaleFrom(0), ttk.ScaleTo(100), ttk.ScaleVariable(volume), ttk.ScaleLength(screenunit.Cm(5))),
		pack.Anchor(option.AnchorW), pack.PadY(6))

	pack.Pack(ttk.NewSeparator(page2, "sep", ttk.SeparatorOrient(ttk.Horizontal)),
		pack.FillOpt(pack.FillX), pack.PadY(6))

	lang := ttk.NewCombobox(page2, "lang",
		ttk.ComboboxValues([]string{"English", "Deutsch", "Français", "日本語"}),
		ttk.ComboboxText("English"),
		ttk.ComboboxCbState(ttk.ComboReadonly),
	)
	pack.Pack(lang, pack.Anchor(option.AnchorW))

	// --- Bottom row ---
	bottom := ttk.NewFrame(root, "bottom")
	pack.Pack(bottom, pack.FillOpt(pack.FillX), pack.PadY(4))
	pack.Pack(ttk.NewButton(bottom, "ok",
		ttk.ButtonText("OK"), ttk.ButtonStyleOpt("Accent.TButton"), ttk.ButtonCommand(app.Quit)),
		pack.SideOpt(pack.Right))
	pack.Pack(ttk.NewButton(bottom, "cancel",
		ttk.ButtonText("Cancel"), ttk.ButtonCommand(app.Quit)),
		pack.SideOpt(pack.Right), pack.PadX(4))

	app.Run()
}
```

---

## 17. Putting it together: a to-do app

The last program is a small but complete application: a persistent task list
with a menubar, keyboard shortcuts, dialogs, an About window, and a clean
split between the data model and the UI.

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/bind"
	"github.com/msorc/takigo/dialog"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/listbox"
	"github.com/msorc/takigo/widget/menu"
	"github.com/msorc/takigo/widget/toplevel"
)

// Task is one to-do entry; the model knows nothing about widgets.
type Task struct {
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type Store struct {
	path  string
	Tasks []Task
}

func (s *Store) Load() error {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.Tasks)
}

func (s *Store) Save() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.Tasks, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

// UI holds the widgets and the glue between them and the Store.
type UI struct {
	app   *takigo.App
	store *Store
	input *ttk.Entry
	list  *listbox.Listbox
	count *ttk.Label
	about *toplevel.Toplevel
}

func (u *UI) refresh() {
	sel := u.list.Selection()
	u.list.Delete(0, u.list.ItemCount()-1)
	left := 0
	for i, t := range u.store.Tasks {
		mark := "☐ "
		if t.Done {
			mark = "☑ "
		}
		u.list.Insert(i, mark+t.Title)
		if t.Done {
			u.list.ItemConfigure(i, "gray50", "")
		} else {
			left++
		}
	}
	if len(sel) > 0 && sel[0] < u.list.ItemCount() {
		u.list.SelectionSet(sel[0], sel[0])
	}
	u.count.SetText(fmt.Sprintf("%d of %d left", left, len(u.store.Tasks)))
	if err := u.store.Save(); err != nil {
		log.Printf("save: %v", err)
	}
}

func (u *UI) selected() (int, bool) {
	sel := u.list.Selection()
	if len(sel) == 0 {
		return 0, false
	}
	return sel[0], true
}

func (u *UI) add() {
	title := u.input.Get()
	if title == "" {
		return
	}
	u.store.Tasks = append(u.store.Tasks, Task{Title: title})
	u.input.Set("")
	u.refresh()
	u.list.See(u.list.ItemCount() - 1)
}

func (u *UI) toggle() {
	if i, ok := u.selected(); ok {
		u.store.Tasks[i].Done = !u.store.Tasks[i].Done
		u.refresh()
	}
}

func (u *UI) rename() {
	i, ok := u.selected()
	if !ok {
		return
	}
	if title, ok := dialog.AskString(u.app, "Rename", "New title:", u.store.Tasks[i].Title); ok && title != "" {
		u.store.Tasks[i].Title = title
		u.refresh()
	}
}

func (u *UI) remove() {
	i, ok := u.selected()
	if !ok {
		return
	}
	u.store.Tasks = append(u.store.Tasks[:i], u.store.Tasks[i+1:]...)
	u.refresh()
}

func (u *UI) clearDone() {
	if dialog.ShowMessage(u.app,
		dialog.MsgTitle("Clear completed"),
		dialog.MsgMessage("Remove all completed tasks?"),
		dialog.MsgType(dialog.MsgQuestion),
		dialog.MsgButtons(dialog.BtnOKCancel),
	) != dialog.ResultOK {
		return
	}
	kept := u.store.Tasks[:0]
	for _, t := range u.store.Tasks {
		if !t.Done {
			kept = append(kept, t)
		}
	}
	u.store.Tasks = kept
	u.refresh()
}

func (u *UI) showAbout() {
	if u.about != nil && !u.about.Destroyed {
		u.about.Show()
		return
	}
	u.about = toplevel.New(u.app, "about",
		toplevel.Title("About"),
		toplevel.Resizable(false, false),
		toplevel.TransientFor(u.app),
	)
	f := ttk.NewFrame(u.about, "f", ttk.FramePadding(ttk.UniformPadding(16)))
	pack.Pack(f, pack.FillOpt(pack.FillBoth), pack.Expand(true))
	pack.Pack(ttk.NewLabel(f, "msg",
		ttk.LabelText("Tasks\nA takigo tutorial app"),
		ttk.LabelJustify(option.JustifyCenter)))
	pack.Pack(ttk.NewButton(f, "close",
		ttk.ButtonText("Close"), ttk.ButtonCommand(u.about.Destroy)), pack.PadY(8))
	u.about.OnClose(u.about.Destroy)
	u.about.Show()
}

func (u *UI) build() {
	app := u.app

	bar := menu.NewMenubar(app, "menubar")
	tasks := menu.New(app, "tasks")
	tasks.AddCommandAccel("Toggle done", "Space", u.toggle)
	tasks.AddCommandAccel("Rename…", "F2", u.rename)
	tasks.AddCommandAccel("Delete", "Del", u.remove)
	tasks.AddSeparator()
	tasks.AddCommand("Clear completed…", u.clearDone)
	tasks.AddSeparator()
	tasks.AddCommand("Quit", app.Quit)
	bar.AddCascade("Tasks", 0, tasks)
	help := menu.New(app, "help")
	help.AddCommand("About", u.showAbout)
	bar.AddCascade("Help", 0, help)

	f := ttk.NewFrame(app, "main", ttk.FramePadding(ttk.UniformPadding(8)))
	pack.Pack(f, pack.FillOpt(pack.FillBoth), pack.Expand(true))

	u.input = ttk.NewEntry(f, "input", ttk.EntryPlaceholder("What needs doing?"))
	addBtn := ttk.NewButton(f, "add", ttk.ButtonText("Add"), ttk.ButtonCommand(u.add))

	vsb := ttk.NewScrollbar(f, "vsb", ttk.ScrollbarOrientOpt(ttk.Vertical))
	u.list = listbox.New(f, "list",
		listbox.Height(12), listbox.Width(40),
		listbox.SelectModeOpt(listbox.SelectBrowse),
		listbox.YScrollCommand(vsb.Set),
	)
	vsb.Configure(ttk.ScrollbarCommandOpt(func(args ...any) {
		switch args[0] {
		case "moveto":
			u.list.YViewMoveTo(args[1].(float64))
		case "scroll":
			u.list.YViewScroll(args[1].(int), args[2] == "pages")
		}
	}))
	u.count = ttk.NewLabel(f, "count")

	grid.Grid(u.input, grid.Row(0), grid.Column(0), grid.Sticky(grid.EW), grid.PadY(4))
	grid.Grid(addBtn, grid.Row(0), grid.Column(1), grid.ColumnSpan(2), grid.PadX(4))
	grid.Grid(u.list, grid.Row(1), grid.Column(0), grid.ColumnSpan(2), grid.Sticky(grid.NSEW))
	grid.Grid(vsb, grid.Row(1), grid.Column(2), grid.Sticky(grid.NS))
	grid.Grid(u.count, grid.Row(2), grid.Column(0), grid.ColumnSpan(3), grid.Sticky(grid.StickW))
	grid.ColumnConfigure(f, 0, grid.Weight(1))
	grid.RowConfigure(f, 1, grid.Weight(1))

	eng := app.Bind()
	in := u.input.Window().PathName
	lb := u.list.Window().PathName
	on := func(tag, pattern string, fn func()) {
		if err := eng.Bind(tag, pattern, func(*bind.EventData) bool { fn(); return true }); err != nil {
			log.Fatal(err)
		}
	}
	on(in, "<Return>", u.add)
	on(lb, "<space>", u.toggle)
	on(lb, "<Double-Button-1>", u.toggle)
	on(lb, "<F2>", u.rename)
	on(lb, "<Delete>", u.remove)

	widget.Focus(app, u.input.Window())
}

func main() {
	app, err := takigo.NewApp(takigo.Title("Tasks"))
	if err != nil {
		log.Fatal(err)
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	store := &Store{path: filepath.Join(dir, "takigo-tasks", "tasks.json")}
	if err := store.Load(); err != nil {
		dialog.ShowMessage(app, dialog.MsgType(dialog.MsgWarning),
			dialog.MsgMessage("Could not read saved tasks"), dialog.MsgDetail(err.Error()))
	}

	ui := &UI{app: app, store: store}
	ui.build()
	ui.refresh()

	app.Run()
}
```

### Design notes

- **Model/UI split.** `Task` and `Store` are plain Go and could be unit-tested
  without a display. The `UI` type owns the widgets and translates between
  user actions and model changes. Every action edits the model and calls
  `refresh`, which rebuilds the list from the model — no state lives only in
  a widget.
- **One layout per container.** The main frame uses `grid` (the entry and the
  list stretch, the button and scrollbar don't); the frame itself is packed
  into the root; the About window uses `pack`.
- **Bindings in one place.** The `on` helper binds and breaks. Path bindings
  on the listbox run before the Listbox class bindings, so `<space>` toggles a
  task instead of doing whatever the class would do with it.
- **The About window is created lazily** and reused while it exists;
  `Destroyed` tells us when the user closed it.
- **Errors reach the user.** A corrupt save file shows a warning at startup
  instead of silently losing data.

Ideas for extending it: due dates in a treeview with columns, drag-to-reorder
on a canvas, a `<<Save>>` virtual event, or a background goroutine syncing to
a server with `RunOnMain` for the updates.

---

## 18. Testing GUI code

Most of an application can — and should — be tested without a display: keep
logic in plain Go types (like `Store` above) and test them as usual.

For tests that do need widgets, use a real App under Xvfb and skip when no
display is available:

```go
package main

import (
	"os"
	"testing"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/label"
)

func newTestApp(t *testing.T) *takigo.App {
	t.Helper()
	if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
		t.Skip("no display (run under xvfb-run)")
	}
	app, err := takigo.NewApp(takigo.Title("test"), takigo.Size(200, 150))
	if err != nil {
		t.Fatalf("NewApp: %v", err)
	}
	t.Cleanup(app.Destroy)
	return app
}

func TestButtonUpdatesLabel(t *testing.T) {
	app := newTestApp(t)
	l := label.New(app, "l", label.Text("before"))
	b := button.New(app, "b", button.Command(func() { l.Configure(label.Text("after")) }))

	b.Invoke() // drive the widget from code
	app.UpdateIdleTasks()

	if got := l.Text; got != "after" {
		t.Fatalf("label text = %q, want %q", got, "after")
	}
}
```

Run with:

```bash
xvfb-run -a -s "-screen 0 1280x1024x24 -noreset" go test -race ./...
```

Tips:

- Drive widgets through their methods (`Invoke`, `Set`, `SetText`,
  `SelectionSet`, `Toggle`) rather than synthesising mouse clicks.
- Call `app.UpdateIdleTasks()` before asserting on sizes or anything that
  happens at idle time.
- `go test` runs packages in parallel. If several packages open windows on the
  same display, they can overlap and steal each other's focus; either give
  each its own Xvfb or run with `-p 1`. (Inside the takigo repository,
  `internal/testutil` handles this with a display lock.)
- For end-to-end checks, `xdotool` can type and click into a program running
  under Xvfb, and ImageMagick's `import -window root` takes screenshots — this
  is how the examples in this tutorial were exercised.

---

## 19. Debugging tips

- **Nothing appears?** The widget was never given to a geometry manager, or its
  parent wasn't. Every widget, all the way up, must be packed, gridded, placed
  or added to a paned window or notebook.
- **Nothing resizes?** Missing `pack.Expand(true)` / `pack.FillOpt(...)`, or a
  missing `grid.Weight` on the row or column.
- **Option ignored?** Check the error `Configure` returns, and the log for
  options given to a constructor: a bad colour or font is reported with the
  widget's path, and the old value is kept.
- **Data races.** Run with `-race`. A report that mentions a widget or window
  touched from a goroutine you started means a missing `RunOnMain`.
- **Name your windows.** `TAKIGO_DEBUG_NAME_WIDGETS=1` gives every X window
  its widget name, so `xdotool search --name .form.ok` finds it.
- **Dump the widget tree.** `TAKIGO_DUMP_TREE=/tmp/tree.json` makes the app
  write its widget tree (paths, classes, geometry, options) as JSON whenever it
  changes — useful to see what the layout really is.
- **Compare with Tk.** If something behaves differently from Tk, the
  corresponding Tcl demo under `tk/library/demos/` and the Go port under
  `demos/` are the quickest way to check — see `AGENTS.md` for the comparison
  scripts.

---

## 20. Tcl/Tk → takigo cheat sheet

| Tcl/Tk | takigo |
|--------|--------|
| `package require Tk` | `app, err := takigo.NewApp(...)` |
| `wm title . "X"` | `takigo.Title("X")` or `app.WmInfo().SetTitle("X")` |
| `button .b -text Hi -command cb` | `b := button.New(app, "b", button.Text("Hi"), button.Command(cb))` |
| `.b configure -text Bye` | `b.Configure(button.Text("Bye"))` |
| `.b invoke` | `b.Invoke()` |
| `destroy .b` | `b.Destroy()` |
| `pack .b -side left -fill x -expand 1` | `pack.Pack(b, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillX), pack.Expand(true))` |
| `pack forget .b` | `pack.Forget(b)` |
| `pack .b -before .x` | `pack.Pack(b, pack.Before(x))` |
| `grid .l .e -sticky ew` | `grid.Grid(geometry.Group{l, e}, grid.Sticky(grid.EW))` |
| `grid .x -row 1 -column 2` | `grid.Grid(x, grid.Row(1), grid.Column(2))` |
| `grid columnconfigure . 0 -weight 1` | `grid.ColumnConfigure(app, 0, grid.Weight(1))` |
| `place .b -relx 0.5 -rely 0.5 -anchor center` | `place.Place(b, place.RelX(.5), place.RelY(.5), place.Anchor(option.AnchorCenter))` |
| `set v 1; checkbutton .c -variable v` | `v := widget.NewVariable("1"); checkbutton.New(app, "c", checkbutton.Var(v))` |
| `trace add variable v write cb` | `v.OnChange(func(old, new string) {...})` |
| `bind .e <Return> {cb; break}` | `app.Bind().Bind(e.Window().PathName, "<Return>", func(*bind.EventData) bool { cb(); return true })` |
| `bind all <Control-q> exit` | `app.Bind().Bind("all", "<Control-q>", ...)` |
| `bindtags .e` | `app.Bind().BindTags(e.Window())` |
| `event add <<Save>> <Control-s>` | `app.Bind().AddVirtualEvent("Save", "<Control-s>")` |
| `event generate .w <<Save>>` | `app.Bind().GenerateEvent(w.Window(), "Save")` |
| `focus .e` | `widget.Focus(app, e.Window())` |
| `after 1000 cb` | `app.After(time.Second, cb)` |
| `after cancel $id` | `cancel()` (returned by `After`) |
| `after idle cb` | `app.DoWhenIdle(cb)` |
| `update idletasks` | `app.UpdateIdleTasks()` |
| `winfo width .w` | `w.Window().Width` (after `UpdateIdleTasks`) |
| `toplevel .t` | `t := toplevel.New(app, "t", ...); t.Show()` |
| `wm protocol . WM_DELETE_WINDOW cb` | `app.WmInfo().OnDeleteWindow(cb)` |
| `menu .m; . configure -menu .m` | `bar := menu.NewMenubar(app, "menubar")` |
| `.m add command -label X -command cb` | `m.AddCommand("X", cb)` |
| `tk_popup .m $x $y` | `m.Post(ev.RootX, ev.RootY)` |
| `tk_messageBox -type yesno -message M` | `dialog.ShowMessage(app, dialog.MsgMessage(M), dialog.MsgButtons(dialog.BtnYesNo))` |
| `tk_getOpenFile` / `tk_getSaveFile` | `dialog.OpenFile(app, ...)` / `dialog.SaveFile(app, ...)` |
| `tk_chooseDirectory` / `tk_chooseColor` | `dialog.ChooseDirectory(app)` / `dialog.ChooseColor(app)` |
| `pack .b -before .x` | `pack.Pack(b, pack.Before(x))` |
| `image create photo -file f.png` | `img, _ := image.NewPhotoFromFile("name", "f.png"); app.ImageRegistry().Register(img)` |
| `font create H -size 16 -weight bold` | `app.FontRegistry().Define("H", font.Attributes{Size: 16, Weight: font.WeightBold})` |
| `.t insert end "x" tag` | `start := t.EndIndex(); t.Insert("end", "x"); t.TagAdd("tag", start, t.EndIndex())` |
| `.t tag configure b -font {...}` | `t.TagConfigure("b", text.TagFont("..."))` |
| `.t get "insert linestart" "insert lineend"` | `t.Get("insert linestart", "insert lineend")` |
| `.c create rectangle 0 0 10 10 -fill red -tags box` | `c.CreateRectangle(0, 0, 10, 10, canvas.FillColor("red"), canvas.Tags("box"))` |
| `.c bind box <1> cb` | `c.BindItem("box", event.ButtonPressMask, cb)` (check `ev.Button`) |
| `ttk::style configure Accent.TButton -background X` | `ttk.CurrentTheme().GetStyle("Accent.TButton").Defaults["-background"] = pixel` |
| `ttk::style theme use clam` | `ttk.SetCurrentTheme("clam")` (import `ttk/clamtheme`) |
| `vwait done` | `app.RunNestedLoop(doneCh)` |

---

## 21. Where to go next

- **The demos.** `demos/` holds 67 programs, most of them line-by-line ports of
  Tk's own widget demos (`tk/library/demos/*.tcl`). They are the best
  reference for any widget's full API. `demos/widget_demo` is the launcher
  that runs them all:

  ```bash
  go run ./demos/widget_demo
  ```

- **API documentation.** `go doc` works on every package, e.g.
  `go doc github.com/msorc/takigo/widget/text` or
  `go doc github.com/msorc/takigo/ttk.Treeview`.
- **`THREADING.md`** — the precise concurrency contract.
- **`AGENTS.md`** — the repository layout, coding conventions, and the tools
  for comparing takigo with Tk pixel by pixel. Read it before contributing.
- **The Tk manual pages.** Semantics follow Tk 9.1, so the official Tk
  documentation answers most "what exactly does this option do" questions.
