# Demo Refinement Instructions: Tcl → Go 1:1 Conversion

This document describes how to refine Go demos in `demos/` to be 1:1 conversions of Tcl originals in `tk/library/demos/`. The rules are derived from five manually verified demos: label, unicodeout, button, check, radio.

## Reference Files

- **Confirmed Go demos**: `demos/{label,unicodeout,button,check,radio}/main.go`
- **Confirmed Tcl originals**: `tk/library/demos/{label,unicodeout,button,check,radio}.tcl`

---

## 1. Structural Template

Every Go demo follows this exact order:

```go
// Demo: <one-line description matching Tcl's opening comment>.
// Ported from Tk's <name>.tcl demo.
package main

import (
    // standard library imports
    // takigo imports
)

func main() {
    // 1. App creation
    // 2. Outer frame
    // 3. Description label (msg)
    // 4. Bottom buttons (See Dismiss / See Variables)
    // 5. Main content widgets
    // 6. app.Run()
}
```

### 1.1 Comment Header

First two lines are always:
```go
// Demo: <Short description>.
// Ported from Tk's <name>.tcl demo.
```

The description should match the spirit of the Tcl file's opening comment block.

### 1.2 App Creation

Maps from Tcl's:
```tcl
set w .label
toplevel $w
wm title $w "Label Demonstration"
wm iconname $w "label"
positionWindow $w
```

To Go's:
```go
app, err := takigo.NewApp(takigo.Title("Label Demonstration"),
    takigo.Geometry("+300+300"),
    takigo.IconName("label"),
)
if err != nil {
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(1)
}
```

Rules:
- `wm title` → `takigo.Title()`
- `wm iconname` → `takigo.IconName()`
- `positionWindow` → `takigo.Geometry("+300+300")` (always `+300+300`)
- Always include the `if err != nil` error block with `fmt.Fprintf(os.Stderr, ...)` + `os.Exit(1)`

### 1.3 Outer Frame

Always present immediately after app creation:
```go
f := frame.New(app, "f")
pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))
```

This frame is the parent for all demo content. Variable name is always `f`.

### 1.4 Description Label

Maps from Tcl's:
```tcl
label $w.msg -font $font -wraplength 4i -justify left -text "..."
pack $w.msg -side top
```

To Go's:
```go
msg := label.New(f, "msg",
    label.WrapLength("4i"),
    label.JustifyOpt(option.JustifyLeft),
    label.Text("..."),
)
pack.Pack(msg, pack.SideOpt(pack.Top))
```

Rules:
- Widget name is always `"msg"`, variable is always `msg`
- Parent is always `f` (the outer frame)
- `-font $font` is **omitted** in Go (uses default font)
- `-wraplength` value comes directly from Tcl (e.g., `"4i"`, `"5i"`)
- `-justify left` → `label.JustifyOpt(option.JustifyLeft)`
- Text content must match Tcl's text exactly (same wording)
- If Tcl uses `-anchor w`, add `label.Anchor(option.AnchorW)` in Go

### 1.5 Bottom Buttons

Maps from Tcl's:
```tcl
# Without variables:
set btns [addSeeDismiss $w.buttons $w]
pack $btns -side bottom -fill x

# With variables:
set btns [addSeeDismiss $w.buttons $w [list var1 var2 var3]]
pack $btns -side bottom -fill x
```

To Go's:
```go
// Without variables:
btns := demohelper.AddSeeDismiss(f)
pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

// With variables:
vars := make(demohelper.DemoVars[string]) // or DemoVars[bool] for checkbuttons
btns := demohelper.AddVarsSeeDismiss(f, &vars)
pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))
```

Rules:
- If Tcl's `addSeeDismiss` has a variable list → use `AddVarsSeeDismiss` with appropriate type
- If no variable list → use `AddSeeDismiss`
- Variable type: `DemoVars[bool]` for checkbutton demos, `DemoVars[string]` for radiobutton demos
- Buttons are always packed **after** the msg label and **before** main content
- When using grid layout (like radio demo), use `grid.Grid(btns, ...)` instead of `pack.Pack`

### 1.6 Main Content

Placed between the buttons and `app.Run()`. Follows the Tcl structure 1:1 — same widgets, same geometry manager, same option values.

### 1.7 App Run

Always the last line in `main()`:
```go
app.Run()
```

---

## 2. Widget Conversion Rules

### 2.1 Widget Names

Tcl path `$w.left.l1` → Go `label.New(left, "l1", ...)` where `left` is the parent widget.

The Go widget name (second argument) matches the last component of the Tcl path.

### 2.2 Widget Options

Direct 1:1 mapping. Tcl `-option value` → Go `widget.OptionName(value)`.

Common mappings:
| Tcl | Go |
|---|---|
| `-text "..."` | `label.Text("...")` / `button.Text("...")` etc. |
| `-relief raised` | `label.Relief(option.ReliefRaised)` |
| `-relief sunken` | `label.Relief(option.ReliefSunken)` |
| `-wraplength 4i` | `label.WrapLength("4i")` |
| `-justify left` | `label.JustifyOpt(option.JustifyLeft)` |
| `-anchor w` | `label.Anchor(option.AnchorW)` |
| `-anchor nw` | `label.Anchor(option.AnchorNW)` |
| `-borderwidth 2` | `label.BorderWidth(2)` |
| `-variable varName` | `checkbutton.Var(varRef)` / `radiobutton.Var(varRef)` |
| `-value val` | `radiobutton.Value(val)` |
| `-command {...}` | `button.Command(func() {...})` |
| `-indicatoron 0` | `radiobutton.IndicatorOnOpt(false)` |
| `-tristatevalue "multi"` | `radiobutton.TristateValueOpt("multi")` |
| `-compound left` | `label.CompoundOpt(widget.CompoundLeft)` |
| `-bitmap questhead` | `label.Bitmap("questhead")` |
| `-image imgRef` | `label.ImageOpt(imgRef)` |
| `-width N` | `label.Width(N)` / `button.Width(N)` |
| `-pady 0` | `label.PadY(0)` |
| `-padx 1.5p` | `labelframe.PadX("1.5p")` |

### 2.3 Options to Omit

- `-font $font` — omitted in Go (default font used)
- `-relief flat` on checkbuttons/radiobuttons — omitted (default in Go)
- Platform-specific branches (`[tk windowingsystem] eq "aqua"`) — omitted (X11 only)

---

## 3. Geometry Manager Conversion

### 3.1 Pack

Tcl:
```tcl
pack $w.left $w.right -side left -expand yes -padx 7.5p -pady 7.5p -fill both
```

Go (each widget packed separately, or use `geometry.Group` for identical options):
```go
pack.Pack(left, pack.SideOpt(pack.Left), pack.Expand(true),
    pack.PadX("7.5p"), pack.PadY("7.5p"), pack.FillOpt(pack.FillBoth))
pack.Pack(right, pack.SideOpt(pack.Left), pack.Expand(true),
    pack.PadX("7.5p"), pack.PadY("7.5p"), pack.FillOpt(pack.FillBoth))
```

When Tcl packs multiple widgets with identical options in one line (`pack $a $b $c -side top ...`), Go can use `geometry.Group`:
```go
pack.Pack(geometry.Group{cb1, cb2, cb3}, pack.SideOpt(pack.Top), ...)
```

Common pack mappings:
| Tcl | Go |
|---|---|
| `-side top` | `pack.SideOpt(pack.Top)` |
| `-side bottom` | `pack.SideOpt(pack.Bottom)` |
| `-side left` | `pack.SideOpt(pack.Left)` |
| `-side right` | `pack.SideOpt(pack.Right)` |
| `-fill x` | `pack.FillOpt(pack.FillX)` |
| `-fill y` | `pack.FillOpt(pack.FillY)` |
| `-fill both` | `pack.FillOpt(pack.FillBoth)` |
| `-expand yes` | `pack.Expand(true)` |
| `-anchor w` | `pack.Anchor(option.AnchorW)` |
| `-padx "7.5p"` | `pack.PadX("7.5p")` |
| `-pady "1.5p"` | `pack.PadY("1.5p")` |

### 3.2 Grid

Tcl:
```tcl
grid $w.msg -row 0 -column 0 -columnspan 3 -sticky nsew
grid $w.f.l$j $w.f.s$j -sticky ew -pady 0
grid configure $w.f.l$j -padx 1m
grid columnconfigure $w.f 1 -weight 1
```

Go:
```go
grid.Grid(msg, grid.Row(0), grid.Column(0), grid.ColumnSpan(3), grid.Sticky(grid.NSEW))
grid.Grid(langLabel, grid.Row(i), grid.Column(0),
    grid.Sticky(grid.EW), grid.PadX("1m"), grid.PadY(0))
grid.Grid(sampleLabel, grid.Row(i), grid.Column(1),
    grid.Sticky(grid.EW), grid.PadY(0))
grid.ColumnConfigure(samples_f, 1, grid.Weight(1))
```

When Tcl uses `grid configure` to modify a single option after initial grid placement, fold it into the initial `grid.Grid()` call in Go.

For relative grid placement (Tcl's `grid x $w.right.top`), use:
```go
grid.Grid(geometry.Group{grid.Relative(grid.RelEmpty), rightButtons["top"]})
```

---

## 4. Variable System

### 4.1 Tcl Variables → Go Variables

Tcl:
```tcl
# Implicit creation via -variable option
checkbutton $w.b1 -variable wipers
radiobutton $w.left.b10 -variable size -value "10"
```

Go:
```go
wipers := widget.NewVariable(false)           // bool for checkbuttons
sizeVar := widget.NewVariable("12")           // string for radiobuttons
```

Rules:
- Checkbutton variables are `widget.NewVariable(false)` (bool type)
- Radiobutton variables are `widget.NewVariable("defaultValue")` (string type)
- Default value matches the initially selected option in Tcl
- Register variables with `vars["name"] = varRef` when using `AddVarsSeeDismiss`

### 4.2 Variable Traces → OnChange

Tcl:
```tcl
trace add variable align write {apply {args {
    $w.right.l configure -compound $align
}}}
```

Go:
```go
alignVar.OnChange(func(_, v string) {
    switch v {
    case "top":
        l.Compound = widget.CompoundTop
    // ...
    }
    l.Display()
})
```

---

## 5. Command Callbacks

Tcl:
```tcl
button $w.b1 -text "Peach Puff" -command [list colorrefresh $w PeachPuff1]
```

Go:
```go
btn := button.New(f, "btn_Peach Puff",
    button.Text("Peach Puff"),
    button.Command(func() { changeColor("PeachPuff1") }),
)
```

Rules:
- Tcl procs defined before widget creation become Go closures/functions defined before widget creation
- Closure captures work naturally in Go; the pattern mirrors Tcl's variable scoping
- Platform-specific branches in Tcl procs (aqua checks) are omitted

---

## 6. Tcl Constructs to Skip

These Tcl patterns have no Go equivalent and should be omitted:

1. **`if {![info exists widgetDemo]}`** — boilerplate guard, not needed
2. **`package require tk`** — handled by Go imports
3. **`catch {destroy $w}`** — no pre-existing window to destroy
4. **`positionWindow $w`** — replaced by `takigo.Geometry("+300+300")`
5. **`[tk windowingsystem] eq "aqua"` branches** — X11 only, skip aqua/win32 code
6. **`update`** — Tk event processing hint, not needed
7. **`$w cget -cursor` / `$w conf -cursor watch`** — cursor manipulation during loading
8. **`destroy $w.wait`** — loading indicator cleanup
9. **`-font $font`** — uses default demo font

---

## 7. Text Content Fidelity

The description label text (`msg`) must match Tcl's text **with these adjustments**:

1. Replace "Tcl variable" with "variable" (as in check.tcl → check demo)
2. Remove references to "Click the See Variables button" if present (the button is self-explanatory)
3. Keep all other wording identical to Tcl

---

## 8. Import Organization

Standard Go import grouping:
```go
import (
    // Standard library (fmt, os, strings, etc.)

    // Takigo packages
    "github.com/msorc/takigo"
    "github.com/msorc/takigo/demos/demohelper"
    "github.com/msorc/takigo/geometry/pack"
    "github.com/msorc/takigo/option"
    "github.com/msorc/takigo/widget/frame"
    "github.com/msorc/takigo/widget/label"
    // ... only import packages actually used
)
```

Only import packages that are used. Order: standard library first, then takigo packages alphabetically.

---

## 9. Refinement Checklist

For each demo, compare `demos/<name>/main.go` against `tk/library/demos/<name>.tcl`:

- [ ] Comment header matches Tcl's description
- [ ] App title matches Tcl's `wm title`
- [ ] App icon name matches Tcl's `wm iconname`
- [ ] Description label text matches Tcl (with adjustments per §7)
- [ ] WrapLength matches Tcl's `-wraplength` value
- [ ] Same geometry manager used (pack vs grid) as Tcl
- [ ] Same widget hierarchy (parent-child relationships match Tcl paths)
- [ ] Same widget names (last path component matches)
- [ ] Same option values (relief, anchor, padx, pady, etc.)
- [ ] Same pack/grid options (side, fill, expand, sticky, row, column, span)
- [ ] Variables registered with demohelper when Tcl uses `addSeeDismiss` with var list
- [ ] Commands/callbacks implement same logic as Tcl procs
- [ ] No extra widgets or options not in Tcl (unless required by Go API)
- [ ] No missing widgets or options that are in Tcl
- [ ] Platform-specific Tcl code (aqua/win32) is omitted
- [ ] Tcl boilerplate (widgetDemo check, package require, catch destroy) is omitted
