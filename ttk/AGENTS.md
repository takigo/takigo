# ttk/ — themed widgets and themes

Applies to `ttk/`, `ttk/<name>theme/` and `ttk/internal/`; the root and
`widget/AGENTS.md` still apply (events, focus, distances and the option
setters are the same). The widgets port `tk/generic/ttk/ttk<Name>.c` and
`tk/library/ttk/<name>.tcl`; they embed `TtkWidget`.

- **One package.** `ttk` hosts every themed widget (what Tk's `ttk::`
  namespace has); it is not split into `ttk/<name>` because the names
  would collide with the classic packages. Every option is prefixed with
  the widget name: `ttk.ButtonText`, `ttk.ButtonCommand`,
  `ttk.ButtonStyle`; every widget has a `<Widget>Style`, the entry-like
  widgets share one `FieldState`.
- **Configure.** Widgets implement their typed `Configure` with
  `configure(&w.TtkWidget, w, opts, sync, size)` in `ttk/widget.go`, which
  also rebuilds the layout when the style name changed. Redraws go through
  `TtkWidget.redisplay` (a persistent back buffer).
- **Shared editing.** The entry, combobox and spinbox edit through
  `ttk/internal/entrytext` and draw through `drawFieldText`.
- **Themes.** A theme is a `ttk/<name>theme/theme.go` that registers with
  `ttk.RegisterTheme(...)` at init; demos require one with
  `_ "github.com/takigo/takigo/ttk/<name>theme"`. Themes: default, clam,
  alt, classic, dark. Theme parent chain: clam → default; style parent
  chain: `TButton` → `.`. Style lookups with their own type parameter use
  the generic method form,
  `style.LookupAs[option.Relief]("-indicatorrelief", state)`.
- **Which Tk file is the default theme.** `tk/generic/ttk/ttkDefaultTheme.c`
  creates the *alt* theme (`Ttk_CreateTheme(interp, "alt", NULL)`) despite
  its name; its border and field elements live here as
  `AltBorderElement`/`AltFieldElement` in `ttk/alttheme`. The ttk default
  theme is the generic elements of `tk/generic/ttk/ttkElements.c` plus
  `tk/library/ttk/defaults.tcl`. Porting default-theme drawing from
  ttkDefaultTheme.c moves every ttk button by a corner pixel.
- **Style inheritance.** As in `Ttk_GetStyle`/`Ttk_CreateLayout`, style
  `a.b.c` inherits from `b.c`, then `c`: `Horizontal.TScale` inherits
  `TScale`, `TMenubutton.Toolbutton` resolves to the Toolbutton style and
  layout. Keep Tk's `Variant.Class` order for new names; a takigo-only
  name such as `TSeparator.Horizontal` needs an explicit `Parent` in each
  theme file.
- **`ttk.Panedwindow` wraps the classic widget** until `ttkPanedwindow.c`
  is ported, which is why it has no `PanedwindowStyle` option and why
  `widget/panedwindow` carries `FlagSash`, `Weighted` and `GripSize`.
- Benchmarks for style lookup and layout build: `ttk/theme_bench_test.go`.
