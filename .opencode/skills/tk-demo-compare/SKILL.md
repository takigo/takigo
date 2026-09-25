---
name: tk-demo-compare
description: Use when comparing or fixing a takigo (Go port of Tk) demo against its Tcl/Tk original in tk/library/demos/. Triggers on phrases like "compare demo", "fix demo", "demo doesn't match", "visual diff", "behavioural check", or naming a demos/<name> directory alongside its tk/library/demos/<name>.tcl counterpart. Drives scripts/demo_compare.sh to screenshot both sides for visual diffs, scripts/demo_interact.sh to drive the demo through xdotool events for behavioural diffs, reads the PNGs to spot differences, reads both sources, and edits the Go source until the odiff diff % is acceptable or a takigo core bug is identified.
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

- `DISPLAY` is set (default `:0`), or `HEADLESS=1` to run under `xvfb-run`.
  If neither is available, ask the user to start `xvfb-run` or an X server.
- The following tools must be available; check with `command -v`:
  `xdotool`, `import` (ImageMagick), `magick`, `montage`, `odiff`,
  `go`, `bash`, and the project's own `./tk/unix/wish`. For headless runs
  also `xvfb-run` (from `xorg-server-xvfb`).
- Working directory is the repo root (`/home/msorc/projects/takigo` in this
  project). The skill assumes that.
- Network/image access is fine — the Read tool can read PNGs.

If any tool is missing, report it and stop; do **not** try to install.

### Step 0 — verify environment

Before doing anything else, run this single Bash call:

```bash
for t in xdotool magick montage odiff go bash; do
  command -v "$t" >/dev/null 2>&1 || { echo "missing tool: $t"; exit 1; }
done
if [[ -z "${DISPLAY:-}" && "${HEADLESS:-0}" != "1" ]]; then
  echo "DISPLAY not set and HEADLESS!=1"; exit 1
fi
test -x ./tk/unix/wish || { echo "build wish first (make -C tk/unix)"; exit 1; }
```

If any check fails, report the missing prerequisite to the user and stop.
Skipping this step is a common source of "the screenshot is just a black
PNG" failures.

---

## Existing infrastructure (always use it)

| Script | Purpose |
|---|---|
| `scripts/demo_map.sh`         | Map Go demo name → Tcl demo name (auto-derived from the filesystem; override: `systray → -`). Run with no args for the full list, or with one arg for the mapping of that demo. |
| `scripts/demo_screenshot.sh <go_demo> {go\|tcl} <out.png> [tcl_name]` | Build, launch, screenshot one side. Reads the window title from `demos/<go_demo>/main.go` via `cmd/demotitle` and waits for it via `xdotool search` (no window manager required). Honours `DEMO_GEOMETRY` and `XFT_DPI` env vars to align with the Go side. Re-execs under `xvfb-run` when `HEADLESS=1` or `DISPLAY` is unset. |
| `scripts/demo_compare.sh <demo> [tcl_demo]` | Screenshots both sides, normalises sizes, computes the **odiff diff %** (0–100, lower = more similar, anti-aliasing ignored), and produces four outputs: `<demo>_go.png`, `<demo>_tcl.png`, `<demo>_diff.png`, and a side-by-side montage `<demo>_side.png`. Per-demo compare failures are written to `tmp/logs/<demo>.compare.err`. |
| `scripts/demo_refine.sh <demo> [--retake]` | One-shot wrapper around `demo_compare.sh`; re-prints the paths so they can be `Read`. |
| `scripts/demo_batch.sh [--retake] [--stability N] [--update-baseline] [prefix]` | Screenshot and score every comparable demo headless (uses `demo_map.sh`). Timers are frozen on both sides, so animated demos are included. Produces `tmp/screenshots/scores_sorted.txt` and `scores.tsv`; `--update-baseline` writes the committed `demos/parity.tsv`. |
| `scripts/demo_wrapper.tcl`    | Run a Tk demo standalone (no widget launcher). Used internally by the screenshot scripts. Reads `DEMO_GEOMETRY` from env. |
| `scripts/demo_interact.sh`    | Drive a Go demo through `xdotool` events (key, type, click, wait), capture before/after PNGs, optionally diff (odiff). Use this for **behavioural** verification — does clicking this button do X? does typing into the entry update the variable? See step 5b below. |
| `scripts/_lib.sh`             | Shared helpers (`tcl_demo_for`, `run_compare`, `set_skip_if_exists`, `odiff_score`, `find_windows_exact`, `maybe_xvfb`). Source this from any new script that needs them. |
| `cmd/demotitle`               | Small Go CLI: `go run ./cmd/demotitle <path>` extracts the first `takigo.Title("...")`; `… -geometry <path>` extracts `takigo.Geometry("...")`. |
| `scripts/fix_demo.sh`         | **Deprecated.** Out-of-session script that invokes an LLM CLI (default `claude -p`, configurable via `LLM_TOOL`). Superseded by this skill — prefer the skill. Refuses to run `claude` when `CLAUDECODE` is set. |
| `scripts/fix_all.sh`          | **Deprecated.** Out-of-session batch wrapper. Same caveats as `fix_demo.sh`. |

Outputs live in `tmp/screenshots/`. Set `SKIP_IF_EXISTS=0` to force retakes,
`SKIP_IF_EXISTS=1` (default) to reuse. `SETTLE_SECS` (default `5`) is the maximum
wait for the window to stop changing before capturing. Fonts are pinned to
DejaVu (`PIN_FONTS=1`) and timers frozen (`TAKIGO_FREEZE_TIMERS=1`) on both
sides by default, so scores are reproducible; `demos/parity.tsv` holds the
committed baseline.

---

## Workflow

Run these steps **in order**. Stay in the repo root.

### 1. Identify the demo

Parse the user's input. Common forms:
- `demos/label/main.go` → go_demo = `label`
- `tk/library/demos/label.tcl` → go_demo = `label` (look up the Go side)
- `label` → go_demo = `label`
- names can differ or be Go-only (`systray → -`) — always consult
  `scripts/demo_map.sh`

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

### 5b. Behavioural verification (optional but recommended for interactive demos)

If the demo has interactive widgets (entries, buttons, checkbuttons,
listboxes, canvas items, ...), verify behaviour too — a demo can look
correct but be functionally broken.

```bash
bash scripts/demo_interact.sh <go_demo> --diff \
    [--wait MS] [--key "<seq>"] [--type "text"] \
    [--click X,Y | --click <widget_name>] \
    [--list-widgets]
```

This launches the demo, sends each event in order via `xdotool`,
captures before/after screenshots, and prints the odiff diff %
between them. A diff % > 0 means the event visibly changed the screen.

**Widget-aware clicks are the default.** Prefer `--click <name>` over
`--click X,Y` — takigo sets each widget's Go name as its X11 window
name (when the demo is launched with `TAKIGO_DEBUG_NAME_WIDGETS=1`,
which `demo_interact.sh` does automatically), and `xdotool` resolves
the name to the XID. The script then clicks the centre of the
resolved window. This survives layout changes — if a demo's
geometry shifts by 30 px, the widget-name click still hits the
target.

```bash
# Discover what widgets exist and where they are.
bash scripts/demo_interact.sh entry1 --list-widgets
# Output:
#   XID          NAME                 X           Y           WIDTH     HEIGHT
#   0x6000010    e1                    315         574         689       36
#   0x6000017    e2                    315         622         689       36
#   0x600000c    dismiss               874         726         139       40
#   ...
```

The legacy `--click X,Y` form (window-relative pixel coordinates) is
still supported for one-off clicks when you know exact coordinates.
Tab traversal order is depth-first pre-order: buttons from
`demohelper.AddSeeDismiss` come *before* entries, so for
`demos/entry1`, three `--key Tab` presses reach the first entry.

Examples:

```bash
# Click the widget named 'e1', wait, type "hello", diff.
bash scripts/demo_interact.sh entry1 \
    --click e1 --wait 500 --type "hello" --diff
# Expected: diff % > 0, "hello" appears in the entry.

# Same thing with raw pixel coordinates (legacy mode).
bash scripts/demo_interact.sh entry1 \
    --click 200,265 --wait 500 --type "hello" --diff

# Verify pressing Tab and Return navigates a form.
bash scripts/demo_interact.sh form \
    --key Tab --type "Alice" --key Tab --type "Bob" \
    --key Return --diff
```

When the Tcl side has analogous interaction, do the same on the Tcl
side by sourcing `tk/library/demos/<x>.tcl` from `tclsh` and driving
it with `event generate`, then compare the resulting screenshots.

### 6. Re-screenshot and verify

```bash
export SKIP_IF_EXISTS=0
bash scripts/demo_compare.sh <go_demo> <tcl_demo>
```

Read the new `_go.png` and `_side.png`. Repeat steps 3–6 until:
- Visual diff is acceptable (low odiff diff %, no obvious widget mismatches),
  **and**
- The Go source semantically matches the Tcl source,
  **and** if you ran step 5b, the behavioural diff % is acceptable.

**Stop iterating rule.** If the odiff diff % does not drop by more than 10%
across three consecutive iterations, stop editing the demo. The remaining
gap is almost always in takigo core (e.g. `widget/label/label.go` not
honouring `-wraplength` correctly) or in unavoidable sub-pixel rendering
differences between Xft (Go) and Tk's text drawing. Switch to escalation
mode (step 7).

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
- The final visual diff % (`demo_compare.sh`).
- The final behavioural diff % (`demo_interact.sh` — only if the demo
  has interactive widgets and you ran step 5b).
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
- Don't lower the diff % by making the window smaller than Tk's — that hides
  differences, doesn't fix them.
- Don't replace classic widgets with TTK to "get a closer match" if the Tcl
  original uses classic. Match the original.
- Tcl's `addSeeDismiss` produces a **TTK** separator + buttons inside a
  **TTK** frame. The Go equivalent is `demohelper.AddSeeDismiss` — use it.
- Tcl demos position the window with `positionWindow` at `+300+300` (or
  similar). Set `takigo.Geometry("+300+300")` (or whatever Tcl does) to
  match — the screenshot script reads this from the Go source via
  `cmd/demotitle -geometry` and passes it to the Tcl wrapper.
- The Tcl demos may use the launcher fonts (`mainFont`, `boldFont`, etc.).
  In Go, prefer `app.FontRegistry().Get(tkfont.TkDefaultFont)` or an
  explicit `-family/-size/-weight` matching the original. The screenshot
  script exports `XFT_DPI` so the Tcl wrapper picks the same dpi.
- Window title is taken from `takigo.Title("...")` — make sure it exactly
  matches Tk's `wm title` string, otherwise `demo_compare.sh` will time
  out waiting for the window.
- The diff score is the **odiff diff %** (0–100, lower = more similar),
  computed with anti-aliasing ignored. It is not comparable with the old
  normalized-MAE numbers from earlier runs — don't mix the two scales.
- Don't run `scripts/fix_demo.sh` or `scripts/fix_all.sh` from inside
  opencode with the default `claude` backend — they refuse to run `claude`
  when `CLAUDECODE` is set. Set `LLM_TOOL` to another tool (e.g.
  `opencode-deepseek-v4-pro`) if you want to drive them from here. Use this
  skill instead; it supersedes them.
- When in doubt, run `wish tk/library/demos/<tcl_demo>.tcl` by hand (via
  `scripts/demo_wrapper.tcl`) and inspect the window before editing.
- For behavioural checks, prefer `--click <widget_name>` over
  `--click X,Y` — widget-name clicks resolve via X11 window names set
  by takigo (when `TAKIGO_DEBUG_NAME_WIDGETS=1` is set, which
  `demo_interact.sh` does automatically). Use `--list-widgets` to
  discover widget names + positions before clicking. Legacy
  `--click X,Y` is window-relative (top-left = 0,0).
- Tab traversal is depth-first pre-order, so `demos/entry1`'s button
  bar (dismiss → code) comes before the entries — three Tab presses
  reach the first entry.
- `demo_interact.sh` requires a real X11 display (DISPLAY set) and
  `xdotool` installed. Step 0's environment check covers both.

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
naming is consistent within a package. The inverse lookup (Tcl option →
takigo source line) lives in
[`references/tcl-option-map.md`](references/tcl-option-map.md) — read it
when a Tcl option isn't doing what you expect in Go, since the bug is
usually in the option's constructor.