# AGENTS.md — takigo

A pure-Go port of the Tk GUI toolkit. The Go source mirrors the structure of
the original Tcl/Tk codebase; the corresponding Tk sources live in the
gitignored `tk/` and `tcl/` directories and are used as the reference.

Module path: `github.com/msorc/takigo`
Go version: `1.25.0` (toolchain `go1.26.6+` on this host)

---

## Build, test, lint

```bash
# Build everything for the host platform.
go build ./...

# Run all unit tests. Most tests don't need an X display; GUI tests
# (tk_test.go, platform/x11/x11_test.go, anything using internal/testutil)
# call t.Skip when DISPLAY is unset.
go test -short ./...

# Run only a package.
go test -short ./canvas/ -run TestParseXBM

# Run a single demo binary (Linux/X11).
go build -o /tmp/button ./demos/button && DISPLAY=:0 /tmp/button

# Fuzz / benchmark targets.
go test ./bind/ -run '^$' -fuzz '^FuzzParse$' -fuzztime 30s
go test ./event/ -run '^$' -bench .

# Static analysis.
go vet ./...
golangci-lint run ./...   # config in .golangci.yml (v2 format)
```

There is **no `Makefile` and no `go.sum`** — dependencies are stdlib only.
`go vet` produces one pre-existing false-positive on
`platform/x11/convert.go:15` (unsafe.Pointer in cgo shim); ignore it.
`.golangci.yml` enables a broad linter set plus `gofmt`/`goimports` with
`local-prefixes: github.com/msorc/takigo`; it is not a CI gate, so existing
code is not lint-clean — don't mass-fix unrelated warnings.

---

## Repository layout

```
tk.go / tk_x11.go / tk_darwin.go / tk_windows.go  — top-level App + per-platform platformInit
tk_test.go                                         — display-skipping smoke tests

widget/                                            — Widget interface, Base struct, defaults, AppContext
widget/<name>/                                     — classic widgets: button, label, frame, entry,
                                                    listbox, menu, menubutton, message, panedwindow,
                                                    radiobutton, scale, scrollbar, spinbox, square,
                                                    text, toplevel, labelframe, checkbutton, entryutil
                                                    — each widget has a `<name>.go` and usually a
                                                      `bindings.go` for event handlers
ttk/                                               — themed widget package; widgets embed TtkWidget
ttk/<name>theme/                                   — theme registrations (default, clam, alt, classic)
ttk/entrytext/                                     — shared editing logic for ttk entry/combobox/spinbox

window/                                            — Window struct, display, hierarchy, creation
event/                                             — event loop, dispatcher, raw event parsing
platform/                                          — DisplayServer interface (the platform abstraction)
platform/x11/                                      — Linux/BSD backend (uses internal/xlib)
platform/cocoa/                                    — macOS backend (uses internal/cocoa cgo)
platform/windows/                                  — Windows backend (uses internal/win32 cgo)
internal/xlib/, internal/cocoa/, internal/win32/   — low-level bindings (cgo where needed)
internal/treedump/                                 — widget-tree JSON dump (TAKIGO_DUMP_TREE) shared with scripts/tk_dump_tree.tcl
internal/testutil/                                 — helpers for tests that need a real display

geometry/                                          — shared manager infrastructure, Elementer, Group
geometry/pack/  geometry/grid/  geometry/place/    — the three Tk geometry managers

bind/                                              — Tk binding system (tag chain, pattern parsing)
canvas/                                            — canvas widget (item_<shape>.go per shape)
dialog/                                            — file/color/message/font/dir chooser dialogs
draw/                                              — high-level drawing primitives (3D borders etc.)
font/  color/  image/  bitmap/                     — resource subsystems
wm/  focus/  grab/  selection/                     — window manager + input subsystems
config/  cursor/  busy/  systray/  screenunit/  option/  gc/
                                                    — supporting subsystems

demos/<name>/main.go                               — one Go demo per directory, mirrors tk/library/demos/*.tcl
demos/demohelper/                                  — shared demo boilerplate (AddSeeDismiss, images, vars)
demos/images/                                      — image assets only (gif/png/xbm), no main.go
demos/widget_demo/                                 — launcher, Go counterpart of tk/library/demos/widget

cmd/                                               — top-level test programs (demo, bind_demo,
                                                    canvas_demo, dialog_demo, text_demo, dialog_test_debug, demotitle, demodiff)

scripts/                                           — bash + tcl screenshot/compare pipeline (see below)
docs/architecture-review.md                        — architecture review snapshot (2026-08-31)
nemotron_review.md                                 — code review + log of implemented fixes
THREADING.md                                       — threading contract: loop-only vs goroutine-safe APIs
tk/  tcl/                                          — vendored Tk 9.1 + Tcl 9.1 source (gitignored) — REFERENCE ONLY
tmp/                                               — gitignored; screenshot output and progress files
```

### Reference sources (gitignored)

- `tk/library/demos/<name>.tcl` — original Tcl demo for `demos/<name>/main.go`.
- `tk/generic/<file>.c` and `tk/generic/ttk/<file>.c` — original C widget
  implementations ported to Go.
- `tk/library/<file>.tcl` — original Tcl-side widget code (option defaults,
  class bindings, configure validation).

When changing widget behaviour, always check the corresponding Tk source to
confirm the intended semantics; the Go port tracks Tk 9.1.

---

## Architecture overview

The `App` (in `tk.go`) owns the X/Cocoa/Win32 display connection (`window.Display`),
the event loop (`event.Loop`), and shared singletons: `color.Cache`,
`font.Registry`, `image.Registry`, `bind.Engine`, `focus.Manager`,
`selection.Manager`. It implements `widget.AppContext`.

Every UI element (widget, dialog, toplevel) is built on `window.Window`,
which holds a `platform.WindowID`, geometry state, and the standard
hierarchy (`Parent` / `Children` / `PathName` like `.frame1.button1`).

Widgets embed `widget.Base` (classic) or `ttk.TtkWidget` (themed).
They expose a constructor `func New(parent widget.Caregiver, name string, opts ...XxxOption) *Xxx`
and a matching `XxxOption` functional-option type.

The `platform.DisplayServer` interface in `platform/display.go` is the single
abstraction boundary for windowing. It is composed of `DisplayCore` plus
capability interfaces (`WindowManager`, `Drawer`, `GCManager`,
`PixmapManager`, `EventSource`, `GrabManager`, `SelectionManager`,
`CursorManager`, `PropertyManager`, `InputMethodManager`). Per-platform
implementations live in `platform/<x11|cocoa|windows>/`; each exposes
`NewDisplayServer(name) (platform.DisplayServer, font.FontOpener, error)`,
selected at compile time via the `platformInit` shim in `tk_<os>.go`. The
returned `DisplayServer` is a composed wrapper, so backend methods that are
not part of any capability interface are **not** reachable by type assertion
on it — return them explicitly from `NewDisplayServer` (as `FontOpener` is).

`geometry/{pack,grid,place}/` each register themselves as `window.GeomManager`
singletons. When a widget calls `geometry.GeometryRequest(w, w, h)`, the
parent's manager's `RequestProc` re-runs the layout.

---

## Coding conventions

- **Widget constructor signature.** Always:
  ```go
  func New(parent widget.Caregiver, name string, opts ...XxxOption) *Xxx
  ```
  Use `parent.AppContext()` to reach app singletons; never import the
  top-level `takigo` package from a widget (use `widget.AppContext`).

- **Functional options.** Define `type XxxOption func(*Xxx)` and one
  constructor per settable field. Option setters that look up a resource
  (color, font) `log.Printf` a warning on lookup failure and keep the
  previous value — see `Background` in `widget/button/button.go`.

- **Option naming.** Classic widget packages use the bare Tk option name
  where possible (`button.Text`, `label.Width`, `label.Background`,
  `label.Relief`). An `Opt` suffix is used when the bare name is unavailable
  or ambiguous in that package (`label.FontOpt`, `label.ImageOpt`,
  `label.JustifyOpt`, `scale.FromOpt`, `menu.TearOffOpt`, …) — check the
  package's existing setters before adding one. Every classic package also
  exports ttk-style prefixed aliases as package-level vars
  (`var ButtonText = Text`, `label.LabelText`, `entry.EntryText`, …) so
  classic and ttk code read the same; add the alias when adding a setter.
  The `ttk` package, which hosts every themed widget, always prefixes with
  the widget name (`ttk.ButtonText`, `ttk.ButtonCommand`,
  `ttk.ButtonStyleOpt`).

- **Distances.** Tk accepts `"3p"`, `"2m"`, `"1c"`, `"0.5i"`, or a bare
  number. Use `screenunit.Px(value)` (accepts `int`, `float64`, `string`)
  in every PadX/PadY/IpadX/IpadY option setter. `screenunit.SetScreenDPI`
  is called once in `NewApp` from X11 screen metrics + `Xft.dpi`.

- **Event handling.** In a widget's `bindings.go` write `bindXxx(w, app)`
  that calls `app.Dispatcher().Bind(w.PlatformID, event.XxxMask, ...)` for
  each event. Always check `ev.Type == event.XxxType` inside the handler
  because the mask may catch multiple event types. Mark the parent window
  focusable with `w.Flags |= window.FlagFocusable` in the constructor.

- **Document the source port.** Top-of-file comment should reference the
  Tk source(s) it ports, e.g.:
  ```go
  // Package button implements the button widget with press/hover interaction.
  // It ports the button-specific parts of tk/generic/tkButton.c and
  // library/button.tcl.
  ```

- **Comments are intentionally sparse.** Do not add narrative comments
  unless they explain a non-obvious Tk semantic or a porting decision.

- **No third-party deps.** Stdlib only — do not add modules.

---

## Platform-specific code

- Build tags use `//go:build <goos>` (Go 1.17+ style) at the very top of
  the file, before the `package` line. See `tk_x11.go`, `tk_darwin.go`,
  `tk_windows.go`, `font/{xft,coretext,gdi,named_*}.go`, `systray/systray.go`.
- The cgo bridges live under `internal/<platform>/`. Keep cgo surface
  minimal; the bulk of each backend is pure Go in `platform/<platform>/`.
- When adding a platform capability, add the method to the matching
  capability interface in `platform/display.go` (or a new one composed into
  `DisplayServer`) and implement it in all three backends; each backend has
  `var _ platform.Xxx = (*XxxDisplay)(nil)` compile-time checks. For
  optional, backend-specific behaviour use an optional interface checked by
  type assertion, as `event.EventPumper` is at `event/loop.go:67`.

---

## Tests

- Unit tests live next to the code: `canvas/xbm_test.go`, etc.
- Tests that need a real display must call `requireDisplay(t)` /
  `testutil.RequireDisplay(t)` first, which `t.Skip`s when neither
  `DISPLAY` nor `WAYLAND_DISPLAY` is set.
- Prefer table-driven tests for pure logic (`geometry/grid/grid_test.go`,
  `geometry/pack/pack_test.go`, `canvas/geometry_test.go`, `screenunit/screenunit_test.go`,
  `wm/wm_test.go`, `bind/{table,pattern}_test.go`).
- Fuzz targets: `bind/pattern_fuzz_test.go`, `geometry/grid/grid_fuzz_test.go`.
  Benchmarks: `event/dispatch_bench_test.go` (dispatcher hot path).
- No CI and no race-detector gate today.

---

## Demos — visual parity with Tk

The project ships **67** demos under `demos/<name>/main.go`, most a Go port
of the corresponding `tk/library/demos/<name>.tcl` (`bash scripts/demo_map.sh`
prints the Go↔Tcl mapping; `msgwidget`, `square`, `ttkentry`, `widget_demo`
have no same-named `.tcl`). There is a dedicated
skill and script pipeline for comparing and fixing them.

- **Skill** (auto-loaded by description): `.opencode/skills/tk-demo-compare/SKILL.md`.
  Trigger phrases: "compare demo", "fix demo", "demo doesn't match",
  "visual diff", "behavioural check", or naming a `demos/<name>` alongside
  its `tk/library/demos/<name>.tcl` counterpart.
- **Scripts** (entry points, see `scripts/README.md`):
  ```bash
  bash scripts/demo_compare.sh <demo> [tcl]    # screenshot Go+Tk, compute odiff diff %
  bash scripts/demo_refine.sh <demo> [--retake] # screenshot + show paths
  bash scripts/demo_batch.sh  [--retake] [--stability N] [--update-baseline] [pfx]
                                                # all demos headless, sorted by score
  bash scripts/demo_gate.sh                     # latest batch vs demos/parity.tsv; exit 1 on regression
  bash scripts/demo_interact.sh <demo> ...      # xdotool-driven behavioural test
  bash scripts/fix_demo.sh  <demo>             # single demo, LLM-driven fix
  bash scripts/fix_all.sh   [--status]          # resumable batch fix
  ```
  `fix_demo.sh`/`fix_all.sh` shell out to an LLM CLI chosen by `LLM_TOOL`
  (default `claude`; ids defined in `scripts/_lib.sh`). Don't run them with
  `claude` from inside a Claude Code session.
- **Wish binary used for Tcl side:** `./tk/unix/wish` (built from the
  vendored Tk 9.1; `make -C tk/unix`). System `wish` is 8.6 and incompatible.
- **Required tools on PATH:** `xdotool`, ImageMagick (`import`, `magick`,
  `montage`), `odiff`, `xrdb`, `go`, `bash`; plus `xvfb-run` for headless
  runs. `wmctrl` and ImageMagick `compare` are no longer used. The skill
  verifies these and stops if any are missing.
- **Diff score:** odiff diff % in `[0, 100]` (anti-aliasing ignored) from
  `scripts/demo_compare.sh`. Lower = closer to Tk. Older notes quoting
  normalized-MAE values in `[0, 1]` use a different scale — don't compare them.
- **Determinism:** screenshots pin fonts (`PIN_FONTS`), freeze timers
  (`TAKIGO_FREEZE_TIMERS`) and wait until two consecutive grabs match. The
  committed baseline `demos/parity.tsv` is regenerated with
  `bash scripts/demo_batch.sh --retake --stability 3 --update-baseline`; only
  compare scores produced under the same settings (batch runs headless).
  For core changes: `demo_batch.sh --retake`, then `demo_gate.sh`; accept
  with `demo_gate.sh --accept` once the regressions are understood. Pixel %
  can rise while tree diffs fall (a partly fixed layout shifts); judge by the
  tree diff first.
- **Structural diff:** every compare also dumps both widget trees and runs
  `cmd/demodiff`, writing `tmp/screenshots/<demo>_tree.txt` (Go path ⇄ Tcl
  path, root causes first). Read it before the images; REQSIZE/FONT/RENDER
  entries that repeat across demos for one widget class are core bugs.
  `window.Window.Class` carries the Tk class name for this.
- **Headless:** `HEADLESS=1` (or an unset/unreachable `DISPLAY`) re-execs
  the screenshot scripts under `xvfb-run`.
- **Debug aid:** `TAKIGO_DEBUG_NAME_WIDGETS=1` calls `XStoreName` on each
  child window so `xdotool` and `demo_interact.sh --click <name>` can target
  widgets by name instead of pixel coordinates (`--list-widgets` prints
  them). See `window/create.go:55`.

### Demo boilerplate

Use `demos/demohelper` for common Tk demo patterns (see
`demos/button/main.go`): `AddSeeDismiss(parent)` for the standard
bottom-row "See Code / Dismiss" buttons, `AddBottomButtons(parent, ...)`
for other rows, and the `image`, `font`, and `vars` globals for resource
lookup.

---

## Common tasks — where to look

| Task | Location |
|---|---|
| Add a Tk widget option | mirror `tk/generic/<file>.c` + `library/<widget>.tcl`; update `widget/<name>/<name>.go` option setter, then `bindings.go` if it triggers a redraw |
| Add a new platform capability | extend `platform.DisplayServer` (`platform/display.go`) and implement in all three `platform/<x11|cocoa|windows>/` |
| Add a new ttk theme | drop a `ttk/<name>theme/theme.go` that registers with `ttk.RegisterTheme(...)` at init; require it from demos with `_ "github.com/msorc/takigo/ttk/<name>theme"` |
| Fix a wrong-looking demo | use the `tk-demo-compare` skill — it drives the comparison and edit loop end-to-end |
| Add a unit test for pure logic | table-driven `*_test.go` next to the source, no display required |
| Add an integration test needing X | import `internal/testutil` and call `testutil.NewTestApp(t)` (registers `t.Cleanup`) |
| Understand a Tk semantic | grep `tk/library/<file>.tcl` and the relevant `tk/generic/<file>.c` (vendored, gitignored) |
| Adjust default widget colour/font | `widget/defaults.go` (mirrors `tk/unix/tkUnixDefault.h`) |
| Add a binding tag | `bind/table.go` (`BindingTable`) + `bind/pattern.go` (Tk pattern syntax) |
| Adjust DPI / unit conversion | `screenunit/screenunit.go` (Tk's `tkCmds.c:1316` ScalingCmd) |

---

## Things to avoid

- **Don't modify `tk/` or `tcl/`.** They are gitignored vendored reference
  sources. They exist for reading and for `./tk/unix/wish` during demo
  comparison; changes here are wiped on re-vendoring.
- **Don't import the top-level `takigo` package from `widget/` or any
  widget package.** It creates an import cycle; widgets depend on
  `widget.AppContext` instead.
- **Don't block the main goroutine.** Everything UI-related runs on the
  event loop. Use `app.DoWhenIdle(fn)`, `app.After(d, fn)`, or
  `app.RunOnMain(fn)` from other goroutines; for modal dialogs use
  `app.RunNestedLoop(doneCh)` or `app.RunNestedLoopContext(ctx, doneCh)`.
  `THREADING.md` lists which types are loop-only vs goroutine-safe.
- **Don't add third-party dependencies.** Keep go.mod / no-go.sum
  structure intact.
- **Don't add comments** unless they document a Tk porting decision or a
  non-obvious invariant.
- **Don't introduce goroutine-shared state in event handlers** — see
  `event/loop.go:24-27` for the threading contract (loop goroutine owns
  `rawHandler`/`idleQueue`; the reader goroutine only posts to channels).

---

## Useful environment variables

| Variable | Effect |
|---|---|
| `DISPLAY` / `WAYLAND_DISPLAY` | When unset, display-dependent tests skip |
| `TAKIGO_DEBUG_NAME_WIDGETS=1` | `XStoreName`s every child window with its Go name (for `xdotool`, `demo_interact.sh`) |
| `XFT_DPI` | Pushed into Tk resources by `demo_wrapper.tcl` so Tk-side fonts match the Go side |
| `SKIP_IF_EXISTS=1` | Reuse existing screenshots in `demo_compare.sh` |
| `SETTLE_SECS` / `TIMEOUT_SECS` | Demo screenshot wait tuning (see `scripts/README.md`) |
| `PIN_FONTS` | `1` (default in screenshot scripts) = private DejaVu-only fontconfig for both sides |
| `TAKIGO_DUMP_TREE=<file>` | Go apps (and `demo_wrapper.tcl`) rewrite the widget tree as JSON every 250ms when it changes (`internal/treedump`); compared by `cmd/demodiff` |
| `TAKIGO_FREEZE_TIMERS=1` | `event.Loop.After` drops positive-delay timers (and `demo_wrapper.tcl` does the same to `after`); set by the screenshot scripts |
| `HEADLESS=1` | Run screenshot/interact scripts under `xvfb-run` |
| `LLM_TOOL` | LLM CLI used by `fix_demo.sh` / `fix_all.sh` (default `claude`) |
| `WISH` | Override path to Tk 9.1 wish binary (default `./tk/unix/wish`) |
