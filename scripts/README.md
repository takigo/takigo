# Demo Comparison & Auto-Fix Scripts

Automated visual comparison between Go and Tcl/Tk demos, with LLM-powered fixes
(Claude by default, configurable via the `LLM_TOOL` env var).

> **Note:** `fix_demo.sh` and `fix_all.sh` invoke an LLM CLI (default `claude`)
> to edit demos. When using `claude`, run from a **regular terminal**, not from
> inside a Claude Code session. Set `LLM_TOOL` to switch tools, e.g.
> `LLM_TOOL=opencode-deepseek-v4-pro bash scripts/fix_demo.sh button`.

## Quick Start

```bash
# Fix one demo
bash scripts/fix_demo.sh button

# Fix all demos (resumable)
bash scripts/fix_all.sh
```

---

## Scripts

### `fix_demo.sh` — Fix a single demo

```bash
bash scripts/fix_demo.sh <demoname> [options]

Options:
  --no-retake      Reuse existing screenshots (faster when iterating)
  --iterations N   Run N fix+compare cycles (default 1)
```

Examples:
```bash
bash scripts/fix_demo.sh button
bash scripts/fix_demo.sh button --no-retake
bash scripts/fix_demo.sh button --iterations 2
```

Output: updated screenshots in `tmp/screenshots/`, LLM log in `tmp/logs/<demo>_iter1.log`.

---

### `fix_all.sh` — Fix all demos with progress tracking

```bash
bash scripts/fix_all.sh [options]

Options:
  --status           Print current progress and exit
  --reset <demo>     Reset a demo to pending (re-run it next time)
  --skip  <demo>     Mark a demo as skipped
  --only  <demo>     Process only this one demo
  --filter <prefix>  Process only demos whose name starts with prefix
  --iterations N     Fix cycles per demo (default 1)
  --no-retake        Reuse existing screenshots between demos
```

Examples:
```bash
bash scripts/fix_all.sh                  # start or resume
bash scripts/fix_all.sh --status         # check progress
bash scripts/fix_all.sh --reset button   # retry a demo
bash scripts/fix_all.sh --skip goldberg  # skip a demo
bash scripts/fix_all.sh --filter ttk     # only ttk* demos
```

Progress is saved in `tmp/fix_all_progress.tsv` after each demo. Interrupt with
Ctrl+C at any time — the next run resumes from the first pending demo.

---

### Lower-level scripts (used internally)

| Script | Purpose |
|--------|---------|
| `demo_compare.sh <demo> [tcl_name]` | Screenshot both sides, compute odiff diff % (0–100, lower = more similar), generate montage |
| `demo_screenshot.sh <demo> go\|tcl <out.png> [tcl_name]` | Screenshot one side |
| `demo_refine.sh <demo> [--retake]` | Compare + print image paths for manual Claude analysis |
| `demo_batch.sh [--retake] [--stability N] [--update-baseline] [prefix]` | Score all demos headless (timers frozen, so animated demos are included); `--stability N` recaptures each side N times and flags self-diff > 0; `--update-baseline` writes `demos/parity.tsv` |
| `demo_map.sh [demo]` | Print Go↔Tcl name mapping (or Tcl name for one demo) |
| `demo_interact.sh <demo> [events...]` | Drive a Go demo with xdotool (`--key`, `--type`, `--click X,Y\|<name>`, `--wait`), capture before/after PNGs, `--diff` for odiff %, `--list-widgets` to discover names (needs `TAKIGO_DEBUG_NAME_WIDGETS=1`) |
| `tk_dump_tree.tcl` | Sourced by `demo_wrapper.tcl` when `TAKIGO_DUMP_TREE=<file>`: writes the Tk widget tree as JSON (same schema as `internal/treedump`, which the Go side writes via the same env var) |
| `go run ./cmd/demodiff [-json] [-go-png F -tcl-png F] go.tree.json tcl.tree.json` | Structural diff of two tree dumps: TOPLEVEL/FONT/CLASS/MISSING/EXTRA/REQSIZE/SIZE/POS/RENDER, root causes first; run automatically by `demo_compare.sh` |
| `_lib.sh` | Shared helpers (`tcl_demo_for`, `run_compare`, `odiff_score`, `maybe_xvfb`, `LLM_TOOL` launch configs); source it from new scripts |
| `demo_wrapper.tcl <demo>` | Run a Tk 9.1 demo standalone (used by screenshot scripts) |

---

## Output files

```
tmp/screenshots/<demo>_go.png      Go screenshot
tmp/screenshots/<demo>_tcl.png     Tcl/Tk screenshot
tmp/screenshots/<demo>_side.png    Side-by-side montage with diff score
tmp/screenshots/<demo>_diff.png    Pixel-level diff heatmap
tmp/screenshots/<demo>_{go,tcl}.tree.json  Widget-tree dumps (geometry, class, Tk options, fonts)
tmp/screenshots/<demo>_tree.txt    Structural diff (also _tree.json)
tmp/screenshots/scores_sorted.txt  Batch scores ranked worst-first
tmp/screenshots/scores.tsv         Batch scores with window sizes and self-diff
tmp/fontconfig/                    Generated pinned fontconfig (see PIN_FONTS)
demos/parity.tsv                   Committed per-demo baseline (score, sizes, status, note)
tmp/fix_all_progress.tsv           Per-demo status: pending/done/failed/skipped
tmp/logs/<demo>_iter1.log          the LLM's output for each fix attempt
```

## Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DISPLAY` | `:0` | X display |
| `HEADLESS` | `0` (`1` in `demo_batch.sh`) | `1` = run screenshots under `xvfb-run` (also auto-falls back when `DISPLAY` is unset) |
| `SETTLE_SECS` | `5` | Max wait for the window content to stop changing; capture happens as soon as two consecutive grabs are identical |
| `DUMP_TREE` | `1` | `1` = both sides write a widget-tree dump next to the screenshot and `demo_compare.sh` runs `cmd/demodiff` |
| `PIN_FONTS` | `1` | `1` = both sides use a private fontconfig with only DejaVu Sans/Serif/Mono, all other families aliased onto them, fixed antialias/hinting |
| `TAKIGO_FREEZE_TIMERS` | `1` | `1` = drop every timer with a positive delay on both sides (Go `event.Loop.After`, Tcl `after`) and disable cursor blink, so captures are deterministic |
| `TIMEOUT_SECS` | `15` | Max wait for demo window to appear |
| `SKIP_IF_EXISTS` | `0` | `1` = reuse an existing screenshot (`demo_refine.sh`/`demo_batch.sh` set it to `1` unless `--retake`) |
| `WISH` | `./tk/unix/wish` | Path to Tk 9.1 wish binary |
| `LLM_TOOL` | `claude` | LLM id used by `fix_demo.sh`/`fix_all.sh`. Maps to a launch command in `scripts/_lib.sh` (prompt on stdin, response on stdout). Known ids: `claude`, `opencode-deepseek-v4-pro`, `opencode-minimax-m3` |

### Required tools

`xdotool`, `import`/`magick`/`montage` (ImageMagick), `odiff`, `go`, `bash`.
Headless runs also need `xvfb-run` (`xorg-server-xvfb`). `wmctrl` and ImageMagick
`compare` are no longer used.
