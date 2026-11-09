---
name: port-tk-option
description: Use when adding or fixing a Tk widget option in takigo (a Go port of Tk 9.1) — "add -wraplength to label", "port the -overrelief option", "ttk widget is missing -padding", "option X behaves differently than in Tk". Walks the Tk C/Tcl source, the typed option setter, geometry and display wiring, Configure, tests, docs/options.md, the changelog and the demo gate.
---

# Port a Tk option to a takigo widget

Read the root `AGENTS.md` and the area guide first: `widget/AGENTS.md` for
classic widgets, `ttk/AGENTS.md` for themed ones, `canvas/AGENTS.md` for
canvas items. The steps below are the procedure; the guides hold the rules.

## 1. Read what Tk does

- The option table and its default: `tk/generic/tk<Name>.c` (the
  `Tk_OptionSpec` array, `DEF_<WIDGET>_<OPTION>` in
  `tk/unix/tkUnixDefault.h`) or, for ttk, `tk/generic/ttk/ttk<Name>.c`
  plus the theme's `tk/library/ttk/<theme>.tcl` and `defaults.tcl`.
- Where Tk reads the field: `ComputeGeometry`/`WorldChanged` (size),
  `Display` (drawing), the class bindings in `tk/library/<name>.tcl`
  (behaviour). Note every place; a port that misses one drifts.
- Units: a `TK_OPTION_PIXELS` option is a distance (`screenunit.Length`),
  `TK_OPTION_COLOR` a `color.Spec`, `TK_OPTION_FONT` a `font.Spec`,
  `TK_OPTION_RELIEF`/`ANCHOR`/`JUSTIFY` the `option` package enums.

## 2. Add the setter

- Look at the package's existing setters for naming: the bare Tk name
  (`Width`), or the `Opt` suffix when the bare name is taken
  (`FontOpt`). ttk options are prefixed with the widget (`ButtonText`, `ButtonStyle`).
- Shape (`widget/button/button.go`):
  ```go
  // PadX sets horizontal padding.
  func PadX[L screenunit.Length](p L) ButtonOption {
      return func(b *Button) { b.PadX = screenunit.ToPixels(p) }
  }
  ```
  Colours and fonts go through the shared setters (`SetBackgroundColor`,
  `SetFont`, …) so failures reach `OptionFailed`. Defaults come from
  `widget.PaletteFor(app)`, never a literal.

## 3. Wire it

- Read the field in `computeGeometry` and/or `display` exactly where Tk
  does. `Configure` needs no change: `widget.Configure` (classic) or
  `configure` in `ttk/widget.go` already recomputes geometry and redraws.
- Behaviour from the Tcl bindings goes into the widget's `bindings.go`.
- A visible departure from Tk is gated on `widget.Classic(app)`.

## 4. Test

- A table test next to the widget for the setter and the computed
  geometry (`widget/button/button_test.go` style); display tests call
  `testutil.RequireDisplay(t)` first.
- Runtime change: `w.Configure(pkg.Opt(v))` returns nil and the request
  size changes as Tk's would.
- If the option is the package's main feature, extend its
  `example_test.go`.

## 5. Finish

```bash
python3 scripts/options_doc.py > docs/options.md
make check
go test -short ./widget/<name>/ ./ttk/
make demo-batch RETAKE=1 && make demo-gate   # when drawing or size changed
```

Add a `CHANGELOG.md` entry under `## Unreleased`. If a demo that uses the
option exists, run the `tk-demo-compare` skill on it.
