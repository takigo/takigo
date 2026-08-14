---
name: tk-demo-compare
description: Use when comparing or fixing a takigo (Go port of Tk) demo against its Tcl/Tk original in tk/library/demos/. Triggers on phrases like "compare demo", "fix demo", "demo doesn't match", "visual diff", or any task that names a specific takigo demo directory (e.g. demos/label) alongside a tk demo (e.g. tk/library/demos/label.tcl). Takes screenshots of both sides via the project's scripts/demo_compare.sh, reads the resulting images, reads both sources, and iteratively edits the Go source until it visually and behaviourally matches the Tcl/Tk original.
---

# Compare & fix a takigo demo vs its Tk/Tcl original

This skill drives an autonomous **visual + behavioural** comparison between a
**takigo** (Go) demo under `demos/<name>/main.go` and its **Tcl/Tk**
counterpart `tk/library/demos/<name>.tcl`, then fixes the Go side so the two
match.

The repo already ships a working screenshot + diff pipeline (see *Existing
infrastructure* below). Always use it instead of reinventing.

---

## When to invoke this skill

- The user names a demo and wants to know whether the Go version matches Tk
  (`compare demos/label with tk/library/demos/label.tcl`).
- The user wants a single demo fixed (`fix the label demo`, `demos/button is
  wrong`, `make goldberg match tk`).
- The user wants a batch pass (run for many/all demos).
- The user reports a visual or behavioural discrepancy and points at a demo
  file.

Do **not** invoke for purely cosmetic fixes the user has already identified,
or for non-demo code (`widget/`, `canvas/`, etc.) — those are regular code-edit
tasks.

---

## Preconditions

- `DISPLAY` is set (default `:0`). If not, ask the user to start `xvfb-run`
  or an X server.
- The following tools must be available; check with `command -v`:
  `wmctrl`, `xdotool`, `import` (ImageMagick), `magick`, `compare`,
  `go`, `bash`, and the project's own `./tk/unix/wish`.
- Working directory is the repo root (`/home/msorc/projects/takigo` in this
  project). The skill assumes that.
- Network/image access is fine — the Read tool can read PNGs.

If any tool is missing, report it and stop; do **not** try to install.

---

## Existing infrastructure (always use it)

| Script | Purpose |
|---|---|
| `scripts/demo_map.sh`         | Map Go demo name → Tcl demo name (same name unless `windowicons → icon`). Run with no args for the full list, or with one arg for the mapping of that demo. |
| `scripts/demo_screenshot.sh <go_demo> {go\|tcl} <out.png> [tcl_name]` | Build, launch, screenshot one side. Reads the window title from `demos/<go_demo>/main.go` (`takigo.Title("...")`) and waits for it via `wmctrl`. |
| `scripts/demo_compare.sh <demo> [tcl_demo]` | Screenshots both sides, normalises sizes, computes an MAE diff with ImageMagick `compare`, and produces three images: `<demo>_go.png`, `<demo>_tcl.png`, `<demo>_diff.png`, plus a side-by-side montage `<demo>_side.png`. Prints the MAE score. |
| `scripts/demo_refine.sh <demo> [--retake]` | One-shot wrapper around `demo_compare.sh`; re-prints the paths so they can be `Read`. |
| `scripts/demo_batch.sh [--retake] [prefix]` | Screenshot and score every comparable demo (uses `demo_map.sh`). Produces `tmp/screenshots/scores_sorted.txt`. |
| `scripts/demo_wrapper.tcl`    | Run a Tk demo standalone (no widget launcher). Used internally by the screenshot scripts. |
| `scripts/fix_demo.sh`         | *(Out-of-session only — invokes `claude -p`, must not be run from inside an opencode session.)* |
| `scripts/fix_all.sh`          | *(Same — out-of-session batch wrapper.)* |

Outputs live in `tmp/screenshots/`. Set `SKIP_IF_EXISTS=0` to force retakes,
`SKIP_IF_EXISTS=1` (default) to reuse. `SETTLE_SECS` (default `1.5`) controls
how long to wait after the window appears before capturing.

---

## Workflow

Run these steps **in order**. Stay in the repo root.

### 1. Identify the demo

Parse the user's input. Common forms:
- `demos/label/main.go` → go_demo = `label`
- `tk/library/demos/label.tcl` → go_demo = `label` (look up the Go side)
- `label` → go_demo = `label`
- `windowicons` → go_demo = `windowicons`, tcl_demo = `icon` (different
  names — always consult `scripts/demo_map.sh`)

Resolve the Tcl counterpart with `bash scripts/demo_map.sh <go_demo>`. If it
returns `-`, the demo is Go-only and there is nothing to compare — stop and
tell the user.

Verify both source files exist:
- `demos/<go_demo>/main.go`
- `tk/library/demos/<tcl_demo>.tcl`

### 2. Take fresh screenshots

```bash
export SKIP_IF_EXISTS=0   # always retake for a fix session
bash scripts/demo_compare.sh <go_demo> <tcl_demo>
```

Wait for it to finish. The script prints:
- `Diff score: <number>` — mean absolute error, 0 = identical, higher = worse
- Paths to all four PNGs

If the score is `0` (or empty/very small), the demo already matches —
report success and stop unless the user asked for behavioural fixes too.

### 3. Visual analysis (Read the images)

Use the Read tool on **all three** of:
- `tmp/screenshots/<go_demo>_go.png`
- `tmp/screenshots/<go_demo>_tcl.png`
- `tmp/screenshots/<go_demo>_side.png` (montage — best for spotting diffs)

Describe to yourself what you see in each image: window title, widget
hierarchy, sizes, paddings, fonts, colours, relief/borders, label text
content, image presence. List concrete differences as a checklist.

Common issues to look for, in priority order:
1. **Missing or extra widgets** vs the Tcl side (e.g. image label, button).
2. **Wrong padding/margin** (`pack.PadX`/`pack.PadY`, `grid.PadX`/`grid.PadY`,
   `-ipadx`/`-ipady` in Tcl → `pack.IpadX`/`pack.IpadY` in Go).
3. **Wrong font** (size, family, weight) — the Tcl demos use named fonts
   (`$font`, `$boldFont`, `$fixedFont`). Match them with `tkfont.TkDefaultFont`,
   `tkfont.TkTextFont`, `ttk.LabelFont`, or explicit `-family/-size/-weight`
   options. See `widget/label`, `ttk/label`, etc. for option names.
4. **Wrong widget dimensions** (`-width/-height` in Tcl → `Width/Height` in
   Go; for ttk use `ttk.LabelWidth` etc.).
5. **Wrong colours or relief** (`option.ReliefRaised/Sunken/Flat/Groove/Ridge`,
   `color.RGBA(...)`, hex strings through `app.ColorCache().Get("#...")`).
6. **Wrong geometry string** (`takigo.Geometry("+300+300")` matches Tcl's
   `positionWindow` which sets `+300+300`).
7. **Wrong parent/grid placement** (`grid.Row`/`grid.Column` vs Tcl's
   `-row/-column`, `grid.Sticky` vs `-sticky`, `grid.RowSpan`/`grid.ColumnSpan`).
8. **Wrong demo scaffolding** — the demo must call `demohelper.AddSeeDismiss`
   to match Tcl's `addSeeDismiss $w.buttons $w` (See Code / Dismiss buttons).
9. **TTK vs classic mismatch** — Tcl's widget launcher uses TTK for the
   buttons bar but classic `label` for the demo body. Match that split:
   `demohelper.AddSeeDismiss` uses TTK; the demo's widgets themselves are
   usually classic `widget/label`, `widget/frame`, etc.
10. **Image scaling** — Tcl demos use `[expr {$tk::scalingPct / 100}]` for
    zoom; takigo's equivalent is `screenunit.ScalingPct()/100`.

### 4. Read both source files

Use Read on:
- `demos/<go_demo>/main.go` (the file to edit)
- `tk/library/demos/<tcl_demo>.tcl` (the reference)

Read related files as needed (`demohelper/demohelper.go` for the button bar,
`tcl/library/demos/widget` for the launcher equivalent — this lives at
`tk/library/demos/widget`, search the actual path).

Cross-check every Tcl widget/property against the Go side. Examples:
- Tcl: `label $w.msg -font $font -wraplength 4i -justify left -text "..."`
- Go:  `label.New(f, "msg", label.WrapLength("4i"), label.JustifyOpt(...), label.Text(...))`
  (note: `label.Font(...)` may be needed to set the font — check
  `widget/label/label.go` for the option name).

### 5. Edit the Go source

Use the Edit tool. One focused edit per visual difference. Keep the diff
minimal — only change what's needed to match Tk.

Conventions in this codebase (see `CLAUDE.md`):
- **No comments unless required** — don't add explanatory comments to the
  edited file.
- Functional options: `widget.NewX(parent, name, opt1, opt2, ...)`.
- Names: usually a single string, no dots (Tcl uses `.a.b.c`; Go uses
  `app`/`f`/`left` etc.).
- Geometry manager: `pack.Pack`/`grid.Grid` with options
  (`pack.SideOpt(pack.Top)`, `grid.Sticky(grid.EW)`, etc.).
- Distance strings: `"7.5p"` (points), `"4i"` (inches), `"3m"` (mm) — same
  syntax as Tcl.
- Relies on `demohelper.AddSeeDismiss(f)` for the bottom button bar.

After editing, **always build** to catch option-name typos:
```bash
go build ./demos/<go_demo>/
```
If it fails, fix the compile error and try again. Don't proceed with a broken
build.

### 6. Re-screenshot and verify

```bash
export SKIP_IF_EXISTS=0
bash scripts/demo_compare.sh <go_demo> <tcl_demo>
```

Read the new `_go.png` and `_side.png`. Repeat steps 3–6 until:
- Visual diff is acceptable (low MAE score, no obvious widget mismatches),
  **and**
- The Go source semantically matches the Tcl source.

### 7. Iterate or escalate

- If you can't reach a low diff after 3 iterations, the issue is likely in
  takigo core (e.g. `widget/label/label.go` not honouring `-wraplength`
  correctly), not in the demo. Stop editing the demo, write a short report,
  and point at the takigo file that needs a fix.
- If the demo uses a feature takigo doesn't yet implement (e.g. some TTK
  layout feature), add a `// TODO: <feature>` marker instead of inventing
  behaviour and tell the user.

### 8. Report

End with:
- The final MAE score.
- A short list of concrete changes made (file paths + brief descriptions).
- Any takigo bugs discovered that need separate fixes.

---

## Batch mode (optional)

If the user asks for many/all demos:

1. Run `bash scripts/demo_batch.sh --retake` once to populate
   `tmp/screenshots/scores_sorted.txt`.
2. Sort demos worst-first from the scores file.
3. For each demo in order, run steps 1–7 above.
4. Track progress in a TodoWrite list so the user can see where you are.
5. Use the `general` subagent to work on independent demos in parallel **only
   if** the user explicitly asks for parallelism; otherwise do them serially
   to keep screenshots and edits tidy.

---

## Pitfalls / things to avoid

- Don't edit `scripts/` or any takigo library file unless the demo *can't*
  be made to match without it. The goal is to fix demos, not core.
- Don't lower the MAE by making the window smaller than Tk's — that hides
  differences, doesn't fix them.
- Don't replace classic widgets with TTK to "get a closer match" if the Tcl
  original uses classic. Match the original.
- Tcl's `addSeeDismiss` produces a **TTK** separator + buttons inside a
  **TTK** frame. The Go equivalent is `demohelper.AddSeeDismiss` — use it.
- Tcl demos position the window with `positionWindow` at `+300+300` (or
  similar). Set `takigo.Geometry("+300+300")` (or whatever Tcl does) to match.
- The Tcl demos may use the launcher fonts (`mainFont`, `boldFont`, etc.).
  In Go, prefer `app.FontRegistry().Get(tkfont.TkDefaultFont)` or an
  explicit `-family/-size/-weight` matching the original.
- Window title is taken from `takigo.Title("...")` — make sure it exactly
  matches Tk's `wm title` string, otherwise `demo_compare.sh` will time out
  waiting for the window.
- When in doubt, run `wish tk/library/demos/<tcl_demo>.tcl` by hand (via
  `scripts/demo_wrapper.tcl`) and inspect the window before editing.

---

## Reference: option-name mapping cheatsheet

| Tcl option | Go (classic) | Go (TTK) |
|---|---|---|
| `-text "..."` | `label.Text("...")` | `ttk.LabelText("...")` / `ttk.ButtonText(...)` |
| `-image NAME` | `label.ImageOpt(img)` | `ttk.LabelImage(img)` / `ttk.ButtonImage(img)` |
| `-compound left` | `widget.CompoundLeft` (compound) | `ttk.ButtonCompound(widget.CompoundLeft)` |
| `-font NAME` | `label.Font(font)` | `ttk.LabelFont(font)` / `ttk.ButtonFont(font)` |
| `-width N` | `label.Width(N)` | `ttk.LabelWidth(N)` / `ttk.ButtonWidth(N)` |
| `-height N` | `label.Height(N)` | `ttk.LabelHeight(N)` |
| `-wraplength 4i` | `label.WrapLength("4i")` | n/a |
| `-justify left` | `label.JustifyOpt(option.JustifyLeft)` | `ttk.LabelJustify(option.JustifyLeft)` (verify) |
| `-anchor w` | `label.Anchor(option.AnchorW)` | `ttk.LabelAnchor(option.AnchorW)` |
| `-relief raised` | `label.Relief(option.ReliefRaised)` | n/a (use style) |
| `-borderwidth 2` | `label.BorderWidth(2)` | n/a |
| `-side top` | `pack.SideOpt(pack.Top)` | n/a (use `grid` for TTK) |
| `-expand yes` | `pack.Expand(true)` | `grid.Sticky(grid.NSEW)` + `grid.Row/ColumnWeight` |
| `-fill both` | `pack.FillOpt(pack.FillBoth)` | (sticky covers most of it) |
| `-padx 7.5p` | `pack.PadX("7.5p")` | `grid.PadX("7.5p")` |
| `-pady 7.5p` | `pack.PadY("7.5p")` | `grid.PadY("7.5p")` |
| `-ipadx N` | `pack.IpadX(N)` | n/a |
| `-ipady N` | `pack.IpadY(N)` | n/a |
| `-sticky ew` | n/a | `grid.Sticky(grid.EW)` |
| `-row N -column N` | n/a | `grid.Row(N)`, `grid.Column(N)` |
| `-columnspan N` | n/a | `grid.ColumnSpan(N)` |
| `-rowspan N` | n/a | `grid.RowSpan(N)` |
| `-command CMD` | `button.Command(cmd)` | `ttk.ButtonCommand(cmd)` |
| `-textvariable VAR` | `entry.TextVar(v)` | `ttk.LabelTextVar(v)` (verify) |

When an option you need isn't in this table, **grep** the relevant package
(`widget/<x>/<x>.go`) for similar option names before guessing — the option
naming is consistent within a package.