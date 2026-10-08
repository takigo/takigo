# Tcl option → takigo source map

When a Tcl option doesn't behave the way the Tcl original does in the Go
port, the first place to look is the option's constructor in the takigo
source. Use this map to jump straight to the right file.

The convention is consistent within each widget package: a Tcl `-foo`
option is a Go `Foo` constructor taking a typed value (generic over
`screenunit.Length`, `color.Spec` or `font.Spec` for distances, colours and
fonts), placed in the same file as the widget's `New`; an `Opt` suffix marks
the names the bare Tk name could not take. `docs/options.md` lists them all.

## widget/label

| Tcl option | Go constructor | Source |
|---|---|---|
| `-text` | `label.Text` | `widget/label/label.go` |
| `-textvariable` | `label.TextVariable` | `widget/label/label.go` |
| `-image` | `label.ImageOpt` | `widget/label/label.go` |
| `-compound` | `label.CompoundOpt` | `widget/label/label.go` |
| `-bitmap` | `label.Bitmap` | `widget/label/label.go` |
| `-width` (chars) | `label.Width` | `widget/label/label.go` |
| `-height` (lines) | `label.Height` | `widget/label/label.go` |
| `-wraplength` | `label.WrapLength` | `widget/label/label.go` |
| `-justify` | `label.JustifyOpt` | `widget/label/label.go` |
| `-anchor` | `label.Anchor` | `widget/label/label.go` |
| `-padx` | `label.PadX` | `widget/label/label.go` |
| `-pady` | `label.PadY` | `widget/label/label.go` |
| `-relief` | `label.Relief` | `widget/label/label.go` |
| `-borderwidth` | `label.BorderWidth` | `widget/label/label.go` |
| `-background` | `label.Background` | `widget/label/label.go` |
| `-foreground` | `label.Foreground` | `widget/label/label.go` |
| `-font` | `label.FontOpt` | `widget/label/label.go` |
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
| `-onvalue` / `-offvalue` | `checkbutton.OnValueOpt` / `OffValueOpt` | `widget/checkbutton/` |
| `-value` | `radiobutton.Value` | `widget/radiobutton/` |
| `-relief` | `button.ReliefOpt` | `widget/button/` |

When the option isn't where you'd expect, grep the widget's `New` function
— it lists every constructor in argument order.

## geometry/pack

| Tcl option | Go constructor | Source |
|---|---|---|
| `-side` | `pack.SideOpt` | `geometry/pack/pack.go` |
| `-expand` | `pack.Expand` | `geometry/pack/pack.go` |
| `-fill` | `pack.FillOpt` | `geometry/pack/pack.go` |
| `-anchor` | `pack.Anchor` | `geometry/pack/pack.go` |
| `-padx` | `pack.PadX` | `geometry/pack/pack.go` |
| `-pady` | `pack.PadY` | `geometry/pack/pack.go` |
| `-ipadx` | `pack.IPadX` | `geometry/pack/pack.go` |
| `-ipady` | `pack.IPadY` | `geometry/pack/pack.go` |
| `-in` | `pack.In` | `geometry/pack/pack.go` |
| `-after` / `-before` | (not yet exposed) | — |

## geometry/grid

| Tcl option | Go constructor | Source |
|---|---|---|
| `-row` / `-column` | `grid.Row` / `grid.Column` | `geometry/grid/grid.go` |
| `-rowspan` / `-columnspan` | `grid.RowSpan` / `grid.ColumnSpan` | same |
| `-sticky` | `grid.Sticky` (constants `grid.NSEW`, `grid.EW`, etc.) | same |
| `-padx` / `-pady` | `grid.PadX` / `grid.PadY` | same |
| `-ipadx` / `-ipady` | `grid.IPadX` / `grid.IPadY` | same |
| column weight | `grid.ColumnConfigure(..., grid.Weight(n))` | same |
| row weight | `grid.RowConfigure(..., grid.Weight(n))` | same |
| Tk's `x` shortcut | `grid.Relative(grid.RelEmpty)` | `geometry/grid/grid.go` |

## geometry/place

| Tcl option | Go constructor | Source |
|---|---|---|
| `-x` / `-y` | `place.X` / `place.Y` | `geometry/place/place.go` |
| `-relx` / `-rely` | `place.RelX` / `place.RelY` | same |
| `-width` / `-height` | `place.Width` / `place.Height` | same |
| `-relwidth` / `-relheight` | `place.RelWidth` / `place.RelHeight` | same |
| `-anchor` | `place.Anchor` | same |
| `-bordermode` | not ported (inside is assumed) | same |
| `-in` | `place.In` | same |

## ttk widgets

TTK widgets use the same constructor name pattern but live under
`ttk/`. Example:

| Tcl option | Go constructor | Source |
|---|---|---|
| `-text` | `ttk.LabelText` / `ttk.ButtonText` | `ttk/ttklabel.go`, `ttk/button.go` |
| `-image` | `ttk.LabelImage` / `ttk.ButtonImage` | same |
| `-compound` | `ttk.LabelCompound` / `ttk.ButtonCompound` | same (use `widget.CompoundLeft` etc.) |
| `-command` | `ttk.ButtonCommand` | `ttk/button.go` |
| `-variable` | `ttk.CheckbuttonVar` / `ttk.RadiobuttonVar` | `ttk/checkbutton.go`, `ttk/radiobutton.go` |
| `-width` (chars) | `ttk.ButtonWidth` (the ttk label has no width option yet) | per-widget file |
| `-style` | (style engine — see `ttk.Style(...)`) | `ttk/theme.go` |

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
