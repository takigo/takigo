# AGENTS.md — takigo

A pure-Go port of the Tk GUI toolkit. The Go source mirrors the structure of
the original Tcl/Tk codebase; the corresponding Tk sources live in the
gitignored `tk/` and `tcl/` directories and are used as the reference.

Module path: `github.com/takigo/takigo`
Go version: `1.27.0` (`go` directive; with an older local Go, `GOTOOLCHAIN=auto`
fetches go1.27 from proxy.golang.org)

---

## Build, test, lint

```bash
# Build everything for the host platform.
go build ./...

# Run all unit tests. Most tests don't need an X display; GUI tests
# (tests/, platform/x11/x11_test.go, anything using internal/testutil)
# start a private Xvfb for their test binary when Xvfb is installed, so they
# never open windows on your screen (TAKIGO_TEST_DISPLAY=real uses your
# DISPLAY instead; without Xvfb they use it, or skip when it is unset).
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
go fix ./...              # apply modernizers; CI runs `go fix -diff ./...`
golangci-lint run ./...   # config in .golangci.yml (v2 format)

# Releases (CONTRIBUTING.md, "Releases").
make apidiff              # exported API changes since the last v* tag
make release VERSION=v0.2.0 [PUSH=1]   # changelog section -> commit + annotated tag
```

Versions exist only as git tags: no version constant, `takigo.Version()`
reads the build info. Releases stay at v0.x; a breaking change bumps the
minor version and gets a `CHANGELOG.md` entry under `## Unreleased`.

There is **no `go.sum`** — dependencies are stdlib only. The `Makefile` only
wraps the commands above (`make help` lists the goals).
`go vet ./...`, `gofmt -l` and `go fix -diff ./...` are clean and gated by
`.github/workflows/ci.yml`, which also vets the Windows backend
(`CGO_ENABLED=0 GOOS=windows go vet ./...`, pure Go, runs on Linux) and
builds on macOS. Linux builds need `libx11-dev libxft-dev
libfontconfig1-dev`, except that widgets, ttk and the pure-logic packages
also build and test with `CGO_ENABLED=0` (the `no-cgo` CI job); run display tests under
`xvfb-run -a -s "-screen 0 1280x1024x24 -noreset"` — without `-noreset`
Xvfb resets when its last client disconnects and refuses connections
meanwhile, so tests that open an App right after destroying one fail
with "cannot open display".
`.golangci.yml` enables a broad linter set plus `gofmt`/`goimports` with
`local-prefixes: github.com/takigo/takigo`; CI gates only the lines a change
touches (`only-new-issues`), so existing code is not lint-clean — don't
mass-fix unrelated warnings.

---

## Repository layout

```
tk.go / tk_x11.go / tk_darwin.go / tk_windows.go  — top-level App + per-platform platformInit
tests/                                              — black-box App tests and benchmarks (need a display)

widget/                                            — Widget interface, Base struct, defaults, AppContext
widget/<name>/                                     — classic widgets: button, label, frame, entry,
                                                    listbox, menu, menubutton, message, panedwindow,
                                                    radiobutton, scale, scrollbar, spinbox, square,
                                                    text, toplevel, labelframe, checkbutton, entryutil
                                                    — each widget has a `<name>.go` and usually a
                                                      `bindings.go` for event handlers
ttk/                                               — themed widget package; widgets embed TtkWidget
ttk/<name>theme/                                   — theme registrations (default, clam, alt, classic, dark)
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
internal/displaylock/                              — per-display lock file that serializes display tests across packages

geometry/                                          — shared manager infrastructure, Elementer, Group
geometry/pack/  geometry/grid/  geometry/place/    — the three Tk geometry managers

bind/                                              — Tk binding system (tag chain, pattern parsing)
canvas/                                            — canvas widget (item_<shape>.go per shape)
dialog/                                            — file/color/message/font/dir chooser dialogs
draw/                                              — high-level drawing primitives (3D borders etc.)
font/  color/  image/  bitmap/                     — resource subsystems; font/ itself is pure Go,
                                                    the backends are font/xft (X11), font/gdi
                                                    (Windows), font/coretext + platform/cocoa (macOS)
wm/  focus/  grab/  selection/                     — window manager + input subsystems
cursor/  busy/  systray/  screenunit/  option/     — supporting subsystems
appearance/                                        — desktop light/dark preference (not in Tk)
internal/dbus/                                     — minimal D-Bus client (settings portal, notifications)
internal/nanosvg/                                  — SVG rasterizer port; also anti-aliases canvas shapes

demos/<name>/main.go                               — one Go demo per directory, mirrors tk/library/demos/*.tcl
demos/demohelper/                                  — shared demo boilerplate (AddSeeDismiss, images, vars)
demos/images/                                      — image assets only (gif/png/xbm), no main.go
demos/widget_demo/                                 — launcher, Go counterpart of tk/library/demos/widget

internal/cmd/                                      — tools of the demo comparison pipeline (demotitle, demodiff)

scripts/                                           — bash + tcl screenshot/compare pipeline (see below)
docs/architecture-review.md                        — architecture review snapshot (2026-08-31)
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
parent's manager's `RequestProc` re-runs the layout. Their per-container
state lives on the windows (`geometry.Table`, `Window.Value`), not in
package maps: several Apps may run at once on different goroutines
(`THREADING.md`, "Multiple Apps"), so keep per-window state there too.

---

## Coding conventions

- **Widget constructor signature.** Always:
  ```go
  func New(parent widget.Caregiver, name string, opts ...XxxOption) *Xxx
  ```
  Use `parent.AppContext()` to reach app singletons; never import the
  top-level `takigo` package from a widget (use `widget.AppContext`).

- **Functional options.** Define `type XxxOption func(*Xxx)` and one
  constructor per settable field. Colour and font options are generic over
  `color.Spec` (a name or `color.RGB(...)`) and `font.Spec` (a descriptor
  or `font.Attributes`) and go through the shared setters
  (`SetBackgroundColor`, `SetForegroundColor`, `SetFont`, `LookupColor`,
  …) — see `Background` in `widget/button/button.go`. A setter that cannot
  resolve its value keeps the previous one and calls `OptionFailed(err)`:
  inside `Configure` the error is collected and returned, in a constructor
  (which returns no error) it goes to the App's `slog` logger
  (`takigo.WithLogger`). Never `log.Printf` from library code.

- **Runtime configuration.** Every widget has a typed
  `Configure(opts ...XxxOption)` taking the same options as its
  constructor (Tk's `configure`). Change options after creation through
  it — never by writing a field and calling `Display()`, or by calling an
  option func directly. It returns the joined errors of the options that
  failed (`errors.Is(err, color.ErrUnknown)`, `font.ErrNotFound`); the
  others still take effect. Classic widgets implement it with
  `widget.Configure(w, opts, w.computeGeometry)`, which does what Tk's
  WorldChanged procs do: rebuilds the border, recomputes the geometry,
  syncs the window background (`Window.SetBackgroundPixel`), re-arranges
  the content when the margins changed, asks the geometry manager to
  re-lay out when the request changed, and schedules a redraw. Ttk
  widgets use `configure(&w.TtkWidget, w, opts, sync, size)` in
  `ttk/widget.go`, which also rebuilds the layout when the style name
  changed. `SetText`-style methods are thin wrappers over `Configure`.

- **Option naming.** Classic widget packages use the bare Tk option name
  where possible (`button.Text`, `label.Width`, `label.Background`,
  `label.Relief`). An `Opt` suffix is used when the bare name is unavailable
  or ambiguous in that package (`label.FontOpt`, `label.ImageOpt`,
  `label.JustifyOpt`, `scale.FromOpt`, `menu.TearOffOpt`, …) — check the
  package's existing setters before adding one.
  The `ttk` package, which hosts every themed widget, always prefixes with
  the widget name (`ttk.ButtonText`, `ttk.ButtonCommand`,
  `ttk.ButtonStyleOpt`).

- **Distances.** A distance option is generic over `screenunit.Length`
  (`int | float64 | screenunit.Distance`):
  `func PadX[L screenunit.Length](p L) ButtonOption`, storing
  `screenunit.ToPixels(p)`. Callers pass pixels or `screenunit.Pt(3)`,
  `Mm(2)`, `Cm(1)`, `In(0.5)` (Tk's `3p`, `2m`, `1c`, `0.5i`); a string
  does not compile. In library code write constants as
  `screenunit.Pt(3).Pixels()`. `screenunit.Parse` reads Tk's string form
  and returns an error wrapping `ErrBadDistance`. `screenunit.SetScreenDPI`
  is called once in `NewApp` from X11 screen metrics + `Xft.dpi`.

- **Event handling.** In a widget's `bindings.go` write `bindXxx(w, app)`
  that calls `app.Dispatcher().Bind(w.PlatformID, event.XxxMask, ...)` for
  each event. Always check `ev.Type == event.XxxType` inside the handler
  because the mask may catch multiple event types. Mark the parent window
  focusable with `w.Flags |= window.FlagFocusable` in the constructor.
  Set `w.Class` in the constructor: a widget's input handlers (keys,
  buttons, motion, enter/leave) are its class bindings and run at its
  class tag in the `bind.Engine` chain (path, class, toplevel, `all`), so
  a binding on the widget's path runs first and returning `true` (break)
  from it, or `SetBindTags` without the class, suppresses them. Handlers
  for other events (Expose, Configure, Destroy, Focus…) run before any
  binding, like Tk's C event handlers. A widget that needs events for
  every window uses `Dispatcher().BindGlobalFor(w.PlatformID, …)`, which
  goes away with the widget (`BindGlobal` handlers live forever). The
  wheel arrives as `event.MouseWheelType` (bind with `MouseWheelMask`,
  pattern `<MouseWheel>`) with `ev.Delta` at 120 per notch, positive up, and
  `ShiftMask` for a horizontal wheel, as in Tk 9; buttons 4-7 never arrive
  as button events. Turn deltas into scroll units with an
  `event.WheelAccumulator` so high-resolution wheels add up. Test
  "accelerator" modifiers as `ControlMask|platform.CommandMask`: Mod2 is
  Command on macOS but NumLock on X11 and Windows.

- **Focus.** Take the keyboard focus with `widget.Focus(app, w)` (Tk's
  `focus`), never `SetInputFocus`: as in Tk, the X focus stays on the
  toplevel and `focus.Manager` redirects key events to the focus widget,
  so a key handler receives keys only while its widget has the focus. A
  class handler that consumes a key (Text's `<Tab>`) sets `ev.Handled` so
  Tab traversal leaves it alone. Popups (menus, dropdowns) take the focus
  on post and give it back on unpost.

- **Redraws.** Classic widgets draw in a private `display()` registered
  with `Base.SetDisplayProc`; the exported `Display()` only calls
  `Base.EventuallyRedraw`, which coalesces requests into one idle-time
  redraw into a pixmap (Tk's `REDRAW_PENDING` + `Tk_GetPixmap`). Draw
  through `w.Drawable()`, never `platform.WindowDrawable(w.PlatformID)`,
  so the redirection applies.

- **Departures from Tk.** Where takigo deliberately does better than Tk
  (anti-aliased canvas items, `canvas/smooth.go`), the Tk-exact behaviour
  stays available: `widget.Classic(app)` is true in an App created with
  `takigo.Classic()` or run with `TAKIGO_CLASSIC=1`, which the demo
  comparison scripts set. Gate a new departure on it, and keep axis-aligned
  geometry pixel-identical to Tk's so only the edges differ.

- **Document the source port.** Top-of-file comment should reference the
  Tk source(s) it ports, e.g.:
  ```go
  // Package button implements the button widget with press/hover interaction.
  // It ports the button-specific parts of tk/generic/tkButton.c and
  // library/button.tcl.
  ```

- **Modern Go.** Code targets Go 1.27 and the `go fix` modernizers are a
  CI gate (Linux and Windows; the cgo-only darwin files are not
  type-checked there, so keep them modern by hand): `for i := range n`,
  `min`/`max`, `slices`/`maps`, `strings.Cut`/`CutLast`/`SplitSeq`,
  `errors.AsType`, `new(expr)` instead of `v := x; &v`, promoted fields
  directly in composite literals. Generic methods (Go 1.27) are fine
  where a method needs its own type parameter, e.g.
  `style.LookupAs[option.Relief]("-indicatorrelief", state)`; the type
  argument must be explicit when it only appears in the result.

- **Comments are intentionally sparse.** Do not add narrative comments
  unless they explain a non-obvious Tk semantic or a porting decision.

- **No third-party deps.** Stdlib only — do not add modules.

---

## Platform-specific code

- Build tags use `//go:build <goos>` (Go 1.17+ style) at the very top of
  the file, before the `package` line. See `tk_x11.go`, `tk_darwin.go`,
  `tk_windows.go`, `font/named_*.go`, `font/{xft,gdi,coretext}/`, `systray/systray.go`.
- The cgo bridges live under `internal/<platform>/`. Keep cgo surface
  minimal; the bulk of each backend is pure Go in `platform/<platform>/`.
- When adding a platform capability, add the method to the matching
  capability interface in `platform/display.go` (or a new one composed into
  `DisplayServer`) and implement it in all three backends; each backend has
  `var _ platform.Xxx = (*XxxDisplay)(nil)` compile-time checks. For
  optional, backend-specific behaviour use an optional interface checked by
  type assertion, as `event.EventPumper` is at `event/loop.go:87`.

---

## Tests

- Unit tests live next to the code: `canvas/xbm_test.go`, etc.
- Tests that need a real display must call `requireDisplay(t)` /
  `testutil.RequireDisplay(t)` first, which `t.Skip`s when neither
  `DISPLAY` nor `WAYLAND_DISPLAY` is set. It also holds
  `internal/displaylock` for the test: `go test ./...` runs the packages'
  binaries in parallel on one screen, where their windows overlap.
- Display tests get their own Xvfb from `displaylock.UseVirtualDisplay()`,
  called by `testutil.RequireDisplay` and the other `requireDisplay`
  helpers; any new check of `DISPLAY` in a test must call it first.
  The rest of this item applies to `TAKIGO_TEST_DISPLAY=real`: on a real
  desktop a window manager places windows where it
  likes and may map them behind the user's own. Display tests that count X
  errors, open several generations of Apps, or aim at window positions call
  `testutil.Settle()` between generations (a real server hands the next
  connection the same resource IDs while the window manager still acts on
  the old windows) and measure positions instead of assuming them
  (`rootOrigin`/`uncoveredPoint` in `dnd_test.go`); a test that would have to
  send events over foreign windows skips instead.
- Prefer table-driven tests for pure logic (`geometry/grid/grid_test.go`,
  `geometry/pack/pack_test.go`, `canvas/geometry_test.go`, `screenunit/screenunit_test.go`,
  `wm/wm_test.go`, `bind/{table,pattern}_test.go`).
- Fuzz targets: `bind/pattern_fuzz_test.go`, `geometry/grid/grid_fuzz_test.go`.
  Benchmarks: `event/dispatch_bench_test.go` (dispatcher hot path),
  `event/loop_bench_test.go` (idle and RunOnMain queues),
  `geometry/{pack,grid,place}/bench_test.go` (arrange, forget/re-manage on a
  display-free fake), `ttk/theme_bench_test.go` (style lookup, layout build),
  `tests/bench_test.go` (relayout, ttk create/destroy, pixel read-back; needs a
  display),
  `canvas/bench_test.go` (pick, find, tag resolution, display-list edits and
  item redraw on a display-free canvas: `go test ./canvas/ -run '^$' -bench .`).
- CI (`.github/workflows/ci.yml`) runs vet, gofmt and `go test -race`
  (with a coverage profile) on Linux under Xvfb, vets Windows from Linux,
  runs the display-free tests on a Windows runner and builds/tests macOS
  with `-race`. The `lint` job runs golangci-lint on changed lines only and
  `vulncheck` runs govulncheck.

---

## Demos — visual parity with Tk

The project ships **68** demos under `demos/<name>/main.go`, most a Go port
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
  `demos/parity.tsv` was recorded against Tk 9.1b1 `tcltk/tk@22f94052cdf0` and
  Tcl `tcltk/tcl@bb4e6e8795cb` (check out those commits to reproduce it; other
  Tk revisions drift, e.g. `mclist`/`tree`/`ttkbut` changed between
  snapshots). Configure Tk with `--disable-bidi`: the baseline is recorded
  without the bidi/HarfBuzz text layout (on by default since 9.1), which
  wraps some labels differently (e.g. `button` gives 17.8% instead of 0).
  From scratch: clone `tcltk/tcl` and `tcltk/tk` into `tcl/` and `tk/`,
  `./configure --disable-shared` in `tcl/unix`, then
  `./configure --with-tcl=$PWD/../../tcl/unix --disable-shared --disable-bidi`
  and `make` in `tk/unix`.
- **Required tools on PATH:** `xdotool`, ImageMagick 7 (`import`, `magick`,
  `montage`; with ImageMagick 6 a `magick` wrapper that runs `identify`
  for `magick identify` and `convert` otherwise works), `odiff`
  (`npm i -g odiff-bin`), `xrdb`, `go`, `bash`; plus `xvfb-run` for headless
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
  `internal/cmd/demodiff`, writing `tmp/screenshots/<demo>_tree.txt` (Go path ⇄ Tcl
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
| Add a Tk widget option | mirror `tk/generic/<file>.c` + `library/<widget>.tcl`; add the option setter in `widget/<name>/<name>.go` and read the field in `computeGeometry`/`display` — `Configure` then handles it at runtime |
| Add a new platform capability | extend `platform.DisplayServer` (`platform/display.go`) and implement in all three `platform/<x11|cocoa|windows>/` |
| Add a new ttk theme | drop a `ttk/<name>theme/theme.go` that registers with `ttk.RegisterTheme(...)` at init; require it from demos with `_ "github.com/takigo/takigo/ttk/<name>theme"` |
| Fix a wrong-looking demo | use the `tk-demo-compare` skill — it drives the comparison and edit loop end-to-end |
| Add a unit test for pure logic | table-driven `*_test.go` next to the source, no display required |
| Add an integration test needing X | import `internal/testutil` and call `testutil.NewTestApp(t)` (registers `t.Cleanup`) |
| Understand a Tk semantic | grep `tk/library/<file>.tcl` and the relevant `tk/generic/<file>.c` (vendored, gitignored) |
| Adjust default widget colour/font | `widget/palette.go` (`LightPalette` mirrors `tk/unix/tkUnixDefault.h`; read defaults with `widget.PaletteFor(app)`, never a literal) |
| Add a binding tag | `bind/table.go` (`BindingTable`) + `bind/pattern.go` (Tk pattern syntax) |
| Cut a release | `make release VERSION=vX.Y.Z` (`scripts/release.sh`); check `make apidiff` first |
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
  `event/loop.go:25-33` for the threading contract (loop goroutine owns
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
| `TAKIGO_DUMP_TREE=<file>` | Go apps (and `demo_wrapper.tcl`) rewrite the widget tree as JSON every 250ms when it changes (`internal/treedump`); compared by `internal/cmd/demodiff` |
| `TAKIGO_CLASSIC=1` | Tk-exact drawing (no anti-aliased canvas), like `takigo.Classic()`; set by the screenshot scripts |
| `TAKIGO_APPEARANCE` | `light` or `dark`: overrides the desktop's appearance for `appearance.System` |
| `TAKIGO_FREEZE_TIMERS=1` | `event.Loop.After` drops positive-delay timers (and `demo_wrapper.tcl` does the same to `after`); set by the screenshot scripts |
| `HEADLESS=1` | Run screenshot/interact scripts under `xvfb-run` |
| `LLM_TOOL` | LLM CLI used by `fix_demo.sh` / `fix_all.sh` (default `claude`) |
| `WISH` | Override path to Tk 9.1 wish binary (default `./tk/unix/wish`) |
