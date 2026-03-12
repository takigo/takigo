# Demo Comparison & Auto-Fix Scripts

Automated visual comparison between Go and Tcl/Tk demos, with Claude-powered fixes.

> **Note:** `fix_demo.sh` and `fix_all.sh` invoke `claude` and must be run from a
> **regular terminal**, not from inside a Claude Code session.

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

Output: updated screenshots in `tmp/screenshots/`, Claude log in `tmp/logs/<demo>_iter1.log`.

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
| `demo_compare.sh <demo> [tcl_name]` | Screenshot both sides, compute MAE diff, generate montage |
| `demo_screenshot.sh <demo> go\|tcl <out.png> [tcl_name]` | Screenshot one side |
| `demo_refine.sh <demo> [--retake]` | Compare + print image paths for manual Claude analysis |
| `demo_batch.sh [--retake] [prefix]` | Screenshot all demos, output `tmp/screenshots/scores_sorted.txt` |
| `demo_map.sh [demo]` | Print Go↔Tcl name mapping (or Tcl name for one demo) |
| `demo_wrapper.tcl <demo>` | Run a Tk 9.1 demo standalone (used by screenshot scripts) |

---

## Output files

```
tmp/screenshots/<demo>_go.png      Go screenshot
tmp/screenshots/<demo>_tcl.png     Tcl/Tk screenshot
tmp/screenshots/<demo>_side.png    Side-by-side montage with diff score
tmp/screenshots/<demo>_diff.png    Pixel-level diff heatmap
tmp/screenshots/scores_sorted.txt  Batch scores ranked worst-first
tmp/fix_all_progress.tsv           Per-demo status: pending/done/failed/skipped
tmp/logs/<demo>_iter1.log          Claude's output for each fix attempt
```

## Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DISPLAY` | `:0` | X display |
| `SETTLE_SECS` | `1.5` | Wait after window appears before screenshot |
| `TIMEOUT_SECS` | `15` | Max wait for demo window to appear |
| `SKIP_IF_EXISTS` | `1` | Skip screenshot if file already exists |
| `WISH` | `./tk/unix/wish` | Path to Tk 9.1 wish binary |
