# demos/ — visual parity with Tk

Applies to `demos/`; the root `AGENTS.md` still applies. One Go demo per
directory, `demos/<name>/main.go`, most a port of
`tk/library/demos/<name>.tcl` (`bash scripts/demo_map.sh` prints the
Go ↔ Tcl mapping; `msgwidget`, `square`, `ttkentry`, `widget_demo` have no
same-named `.tcl`). `demos/widget_demo` is the launcher, the counterpart of
`tk/library/demos/widget`; `demos/images/` holds assets only;
`demos/square/square/` is the custom-widget example (`tkSquare.c`).

## Boilerplate

Use `demos/demohelper` for the common Tk demo patterns (see
`demos/button/main.go`): `AddSeeDismiss(parent)` for the standard bottom-row
"See Code / Dismiss" buttons, `AddBottomButtons(parent, ...)` for other rows,
and the `image`, `font` and `vars` globals for resource lookup. The window
title must equal Tk's `wm title` string and the geometry Tk's
`positionWindow` (`takigo.Geometry("+300+300")`): the screenshot scripts read
both from the source (`internal/cmd/demotitle`).

Tcl demos express many sizes in points (`4i`, `7.5p`, goldberg's
`Dims(...)p`): convert them with `screenunit.In`, `screenunit.Pt`, …, and
read `screenunit.ScalingPct()` where the Tcl uses `$tk::scalingPct`; never
hardcode a scale factor. Under frozen timers `demo_wrapper.tcl` runs
`expr {srand(1)}`, and Tcl's `srand` also consumes the first random number
(`knightstour` mirrors this with a port of Tcl's `rand()`).

## Comparing a demo with Tk

- **Skill:** `.agents/skills/tk-demo-compare/SKILL.md` (also reachable as
  `.claude/skills/` and `.opencode/skills/`) drives the loop end to end:
  screenshot both sides, read the tree diff, read the images, edit, repeat.
  Trigger phrases: "compare demo", "fix demo", "demo doesn't match",
  "visual diff", "behavioural check". Outside an agent session the same loop
  is `scripts/fix_demo.sh`/`scripts/fix_all.sh`.
- **Scripts** (`scripts/README.md` has every option and environment variable):
  ```bash
  bash scripts/demo_compare.sh <demo> [tcl]     # screenshot Go+Tk, odiff diff %
  bash scripts/demo_refine.sh <demo> [--retake] # screenshot + show paths
  bash scripts/demo_batch.sh  [--retake] [--stability N] [--update-baseline] [pfx]
                                                # all demos headless, sorted by score
  bash scripts/demo_gate.sh                     # latest batch vs demos/parity.tsv; exit 1 on regression
  bash scripts/demo_interact.sh <demo> ...      # xdotool-driven behavioural test
  ```
- **Diff score:** odiff diff % in `[0, 100]` (anti-aliasing ignored); lower
  is closer to Tk. Older notes quoting normalized-MAE values in `[0, 1]` use
  a different scale.
- **Structural diff first:** every compare dumps both widget trees and runs
  `internal/cmd/demodiff`, writing `tmp/screenshots/<demo>_tree.txt` (Go
  path ⇄ Tcl path, root causes first). Read it before the images;
  REQSIZE/FONT/RENDER entries that repeat across demos for one widget class
  are core bugs, not demo bugs. `window.Window.Class` carries the Tk class
  name for this.
- **Determinism:** screenshots pin fonts (`PIN_FONTS`), freeze timers
  (`TAKIGO_FREEZE_TIMERS`), set `TAKIGO_CLASSIC=1` (Tk-exact drawing) and
  wait until two consecutive grabs match. The committed baseline
  `demos/parity.tsv` is regenerated with
  `bash scripts/demo_batch.sh --retake --stability 3 --update-baseline`;
  only compare scores produced under the same settings (batch runs
  headless). For core changes: `demo_batch.sh --retake`, then
  `demo_gate.sh`; accept with `demo_gate.sh --accept` once the regressions
  are understood. Pixel % can rise while tree diffs fall (a partly fixed
  layout shifts); judge by the tree diff first.
- **Headless:** `HEADLESS=1` (or an unset/unreachable `DISPLAY`) re-execs
  the screenshot scripts under `xvfb-run`.
- **Targeting widgets:** `TAKIGO_DEBUG_NAME_WIDGETS=1` calls `XStoreName`
  on each child window so `xdotool` and `demo_interact.sh --click <name>`
  can target widgets by name (`--list-widgets` prints them); see
  `window/create.go`.

## The Tk side

- **Wish:** `./tk/unix/wish`, built from the vendored Tk 9.1
  (`make -C tk/unix`; override with `WISH`). System `wish` is 8.6 and
  incompatible. `demos/parity.tsv` was recorded against Tk 9.1b1
  `tcltk/tk@22f94052cdf0` and Tcl `tcltk/tcl@bb4e6e8795cb` (other revisions
  drift: `mclist`/`tree`/`ttkbut` changed between snapshots), with Tk
  configured `--disable-bidi` (the HarfBuzz layout wraps some labels
  differently). From scratch: clone `tcltk/tcl` and `tcltk/tk` into `tcl/`
  and `tk/`, `./configure --disable-shared` in `tcl/unix`, then
  `./configure --with-tcl=$PWD/../../tcl/unix --disable-shared --disable-bidi`
  and `make` in `tk/unix`.
- **Tools on PATH:** `xdotool`, ImageMagick 7 (`import`, `magick`,
  `montage`; with ImageMagick 6 a `magick` wrapper that runs `identify` for
  `magick identify` and `convert` otherwise works), `odiff`
  (`npm i -g odiff-bin`), `xrdb`, `go`, `bash`; plus `xvfb-run` for
  headless runs. The skill verifies these and stops if any are missing.
