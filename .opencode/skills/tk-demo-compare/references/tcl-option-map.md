# Tcl option → takigo source map

When a Tcl option doesn't behave the way the Tcl original does in the Go
port, the first place to look is the option's constructor in the takigo
source. Use this map to jump straight to the right file.

The convention is consistent within each widget package: a Tcl `-foo`
option is a Go `Foo` constructor taking a typed value (or `any` for
distance / font specs), placed in the same file as the widget's `New`.

## widget/label

| Tcl option | Go constructor | Source |
|---|---|---|
| `-text` | `label.Text` | `widget/label/label.go:48` |
| `-textvariable` | `label.TextVariable` | `widget/label/label.go:54` |
| `-image` | `label.ImageOpt` | `widget/label/label.go:133` |
| `-compound` | `label.CompoundOpt` | `widget/label/label.go:138` |
| `-bitmap` | `label.Bitmap` | `widget/label/label.go:144` |
| `-width` (chars) | `label.Width` | `widget/label/label.go:164` |
| `-height` (lines) | `label.Height` | `widget/label/label.go:169` |
| `-wraplength` | `label.WrapLength` | `widget/label/label.go:176` |
| `-justify` | `label.JustifyOpt` | `widget/label/label.go:116` |
| `-anchor` | `label.Anchor` | `widget/label/label.go:111` |
| `-padx` | `label.PadX` | `widget/label/label.go:122` |
| `-pady` | `label.PadY` | `widget/label/label.go:128` |
| `-relief` | `label.Relief` | `widget/label/label.go:106` |
| `-borderwidth` | `label.BorderWidth` | `widget/label/label.go:101` |
| `-background` | `label.Background` | `widget/label/label.go:70` |
| `-foreground` | `label.Foreground` | `widget/label/label.go:81` |
| `-font` | `label.FontOpt` | `widget/label/label.go:91` |
| `-underline` | (not yet exposed) | — |

## widget/button / checkbutton / radiobutton

These share a base in `widget/button/button.go`. Checkbutton and
radiobutton override individual methods.

| Tcl option | Go constructor | Source |
|---|---|---|
| `-text` | `button.Text` | `widget/button/button.go` (search) |
| `-image` | `button.ImageOpt` | same |
| `-command` | `button.Command` | same |
| `-variable` | `checkbutton.Var` / `radiobutton.Var` | per-widget file |
| `-onvalue` / `-offvalue` | `checkbutton.OnValue` / `OffValue` | `widget/checkbutton/` |
| `-value` | `radiobutton.Value` | `widget/radiobutton/` |
| `-relief` | `button.Relief` | `widget/button/` |

When the option isn't where you'd expect, grep the widget's `New` function
— it lists every constructor in argument order.

## geometry/pack

| Tcl option | Go constructor | Source |
|---|---|---|
| `-side` | `pack.SideOpt` | `geometry/pack/pack.go` |
| `-expand` | `pack.Expand` | `geometry/pack/pack.go` |
| `-fill` | `pack.FillOpt` | `geometry/pack/pack.go` |
| `-anchor` | `pack.Anchor` | `geometry/pack/pack.go:57` |
| `-padx` | `pack.PadX` | `geometry/pack/pack.go:61` |
| `-pady` | `pack.PadY` | `geometry/pack/pack.go:65` |
| `-ipadx` | `pack.IpadX` | `geometry/pack/pack.go` |
| `-ipady` | `pack.IpadY` | `geometry/pack/pack.go` |
| `-in` | `pack.In` | `geometry/pack/pack.go` |
| `-after` / `-before` | (not yet exposed) | — |

## geometry/grid

| Tcl option | Go constructor | Source |
|---|---|---|
| `-row` / `-column` | `grid.Row` / `grid.Column` | `geometry/grid/grid.go` |
| `-rowspan` / `-columnspan` | `grid.RowSpan` / `grid.ColumnSpan` | same |
| `-sticky` | `grid.Sticky` (constants `grid.NSEW`, `grid.EW`, etc.) | same |
| `-padx` / `-pady` | `grid.PadX` / `grid.PadY` | same |
| `-ipadx` / `-ipady` | `grid.IpadX` / `grid.IpadY` | same |
| column weight | `grid.ColumnConfigure(..., grid.Weight(n))` | same |
| row weight | `grid.RowConfigure(..., grid.Weight(n))` | same |
| Tk's `x` shortcut | `grid.Relative(grid.RelEmpty)` | `geometry/grid/grid.go:227` |

## geometry/place

| Tcl option | Go constructor | Source |
|---|---|---|
| `-x` / `-y` | `place.X` / `place.Y` | `geometry/place/place.go` |
| `-relx` / `-rely` | `place.RelX` / `place.RelY` | same |
| `-width` / `-height` | `place.Width` / `place.Height` | same |
| `-relwidth` / `-relheight` | `place.RelWidth` / `place.RelHeight` | same |
| `-anchor` | `place.Anchor` | same |
| `-bordermode` | `place.BorderMode` | same |
| `-in` | `place.In` | same |

## ttk widgets

TTK widgets use the same constructor name pattern but live under
`ttk/`. Example:

| Tcl option | Go constructor | Source |
|---|---|---|
| `-text` | `ttk.LabelText` / `ttk.ButtonText` | `ttk/label.go`, `ttk/button.go` |
| `-image` | `ttk.LabelImage` / `ttk.ButtonImage` | same |
| `-compound` | `ttk.LabelCompound` / `ttk.ButtonCompound` | same (use `widget.CompoundLeft` etc.) |
| `-command` | `ttk.ButtonCommand` | `ttk/button.go` |
| `-variable` | `ttk.CheckbuttonVariable` / `ttk.RadiobuttonVariable` | `ttk/checkbutton.go`, `ttk/radiobutton.go` |
| `-width` (chars) | `ttk.LabelWidth` / `ttk.ButtonWidth` | per-widget file |
| `-style` | (style engine — see `ttk.Style(...)`) | `ttk/style.go` |

## Quick discovery cheat

If the option isn't in this table, find it the same way the takigo
authors do:

```bash
# 1. Find the widget's New function (one per package).
grep -n "^func New(" widget/<name>/<name>.go

# 2. Look at every option constructor referenced in its body.
#    They all live in the same file or a sibling.

# 3. If still missing, check the corresponding tk/generic/<widget>.c
#    for the Tcl option name and search the Go source for it.
```

This is faster than guessing — naming is consistent within each package,
and the `New` function's parameter list is the canonical option table.
