# AGENTS.md — takigo

A pure-Go port of the Tk GUI toolkit. The Go source mirrors the structure of
the original Tcl/Tk codebase; the corresponding Tk 9.1 sources live in the
gitignored `tk/` and `tcl/` directories and are the reference: when changing
widget behaviour, read the Tk source it ports first. Never edit `tk/` or
`tcl/` (re-vendoring wipes them).

Module path: `github.com/takigo/takigo`. Go `1.27.0` (`go` directive; with
an older local Go, `GOTOOLCHAIN=auto` fetches it). Stdlib only: there is no
`go.sum` and no module may be added.

## Per-area guides

Coding agents load these when they touch the directory; read the one for
the area you change.

- `widget/AGENTS.md` — classic widgets: constructor and option shape,
  `Configure`, naming, distances, events, focus, redraws, layout timing.
- `ttk/AGENTS.md` — themed widgets and themes, which Tk file is which theme,
  style inheritance.
- `platform/AGENTS.md` — the `DisplayServer` boundary, the three backends,
  build tags, cgo gotchas.
- `demos/AGENTS.md` — the demos and the screenshot parity pipeline against
  Tk (`scripts/README.md` has every script option).
- `THREADING.md` — loop-only vs goroutine-safe APIs; `docs/architecture-review.md`
  — the design decisions that stand and the open points (2026-10-08);
  `CONTRIBUTING.md` — the pre-change checklist and releases.

## Build, test, lint

```bash
go build ./...
go test -short ./...                  # GUI tests start a private Xvfb when installed
go test -short ./canvas/ -run TestParseXBM
go build -o /tmp/button ./demos/button && DISPLAY=:0 /tmp/button
go test ./bind/ -run '^$' -fuzz '^FuzzParse$' -fuzztime 30s
go test ./event/ -run '^$' -bench .
go vet ./...
go fix ./...                          # modernizers; CI runs `go fix -diff ./...`
golangci-lint run ./...               # .golangci.yml (v2, goimports local-prefixes); CI gates changed lines only
make check                            # vet, gofmt, go fix, Windows vet, check-docs
make apidiff                          # exported API changes since the last v* tag
make release VERSION=v0.2.0 [PUSH=1]  # changelog section -> commit + annotated tag
```

`make help` lists every goal. Run `make check` and `go test -short ./...`
before you finish. CI (`.github/workflows/ci.yml`) runs vet, gofmt,
`go fix -diff`, `scripts/check_tutorial.sh`, `scripts/check_docs.sh` and
`go test -race` under Xvfb on Linux, vets the Windows backend from Linux
(`CGO_ENABLED=0 GOOS=windows go vet ./...`), runs the display-free tests on
Windows and builds/tests macOS. Linux builds need `libx11-dev libxft-dev
libfontconfig1-dev`; widgets, ttk and the pure-logic packages also build and
test with `CGO_ENABLED=0` (the `no-cgo` job). Run display tests under
`xvfb-run -a -s "-screen 0 1280x1024x24 -noreset"`: without `-noreset` Xvfb
resets when its last client disconnects and tests that open an App right
after destroying one fail with "cannot open display". Existing code is not
lint-clean; don't mass-fix unrelated warnings.

Versions exist only as git tags (`takigo.Version()` reads the build info).
Releases stay at v0.x; a breaking change bumps the minor version. A
user-visible change gets a `CHANGELOG.md` entry under `## Unreleased`.

## Repository layout

```
tk.go, tk_{x11,darwin,windows}.go   top-level App + per-platform platformInit
tests/                               black-box App tests and benchmarks (need a display)
widget/, widget/<name>/              Widget interface, Base, AppContext; the classic widgets
ttk/, ttk/<name>theme/               themed widgets (embed TtkWidget); theme registrations
internal/textedit/                   text measuring and word helpers of every entry-like widget
window/  event/  platform/           Window + hierarchy; event loop and dispatcher; DisplayServer
platform/{x11,cocoa,windows}/        the backends, over internal/{xlib,cocoa,win32}
geometry/{pack,grid,place}/          the three geometry managers (shared infra in geometry/)
bind/  canvas/  dialog/  draw/       bindings; canvas (item_<shape>.go); dialogs; 3D drawing
font/  color/  image/  bitmap/       resources (font backends: font/xft, font/gdi, font/coretext)
wm/  focus/  grab/                   window manager and input subsystems
internal/selection/  internal/xdnd/  selections + clipboard; XDND drag and drop
cursor/ busy/ systray/ screenunit/ option/ appearance/ internal/dbus/ internal/nanosvg/
internal/treedump/  internal/cmd/    widget-tree JSON dump; demotitle, demodiff tools
internal/testutil/  internal/displaylock/   display-test helpers; cross-package display lock
demos/<name>/main.go  demos/demohelper/     one demo per directory; shared demo boilerplate
scripts/                             screenshot/compare pipeline, release and check scripts
docs/                                tutorial, option reference, architecture review
tk/  tcl/  tmp/                      vendored Tk/Tcl (reference only); scratch output — all gitignored
```

## Architecture

The `App` (`tk.go`) owns the display connection (`window.Display`), the
event loop (`event.Loop`) and the shared singletons (`color.Cache`,
`font.Registry`, `image.Registry`, `bind.Engine`, `focus.Manager`,
`selection.Manager`); it implements `widget.AppContext`, which is all a
widget may depend on. Every UI element is built on `window.Window`
(`platform.WindowID`, geometry, `Parent`/`Children`/`PathName` such as
`.frame1.button1`). Widgets embed `widget.Base` (classic) or
`ttk.TtkWidget` (themed) and expose
`New(parent widget.Caregiver, name string, opts ...XxxOption)` plus a typed
`Configure`. `platform.DisplayServer` is the only windowing boundary.
`geometry/{pack,grid,place}` register as `window.GeomManager` singletons; a
widget's `geometry.GeometryRequest` makes the parent's manager re-run its
layout. Per-container state lives on the windows (`geometry.Table`,
`Window.Value`), not in package maps, because several Apps may run at once
on different goroutines.

## Rules that apply everywhere

- **Tk is the reference.** Mirror `tk/generic/<file>.c`,
  `tk/generic/ttk/<file>.c` and `tk/library/<file>.tcl`; a top-of-file
  comment names the Tk source(s) a file ports.
- **Comments are sparse.** Add one only for a non-obvious Tk semantic or a
  porting decision.
- **Modern Go** (1.27): `for i := range n`, `min`/`max`, `slices`/`maps`,
  `strings.Cut`/`SplitSeq`, `errors.AsType`, `new(expr)`, promoted fields in
  composite literals, generic methods where a method needs its own type
  parameter. `go fix -diff ./...` is a CI gate (Linux and Windows).
- **Don't block the main goroutine.** Everything UI-related runs on the
  event loop: `app.DoWhenIdle`, `app.After`, `app.RunOnMain` from other
  goroutines; `app.RunNestedLoop(doneCh)` for modal dialogs. The loop
  goroutine owns `rawHandler`/`idleQueue`; other goroutines only post to
  channels (`event/loop.go:25-33`, `THREADING.md`).
- **Never import the top-level `takigo` package** from `widget/` or any
  widget package (import cycle); use `widget.AppContext`. Never
  `log.Printf` from library code: errors go to the App's `slog` logger.
- **Departures from Tk** (anti-aliased canvas items, `canvas/smooth.go`)
  keep the Tk-exact behaviour available: `widget.Classic(app)` is true in
  an App created with `takigo.Classic()` or run with `TAKIGO_CLASSIC=1`,
  which the demo comparison scripts set. Gate a new departure on it and
  keep axis-aligned geometry pixel-identical to Tk.
- **Core changes shift demo screenshots.** After touching widgets, fonts,
  geometry or drawing run `bash scripts/demo_batch.sh --retake` and
  `bash scripts/demo_gate.sh` (`demos/AGENTS.md`).
- **Don't edit `AGENTS.md` paths carelessly:** `scripts/check_docs.sh`
  fails when a guide names a path that is gone.

## Tests

- Unit tests live next to the code and are table-driven where the logic is
  pure (`geometry/grid/grid_test.go`, `bind/pattern_test.go`,
  `screenunit/screenunit_test.go`). Fuzz targets:
  `bind/pattern_fuzz_test.go`, `geometry/grid/grid_fuzz_test.go`.
  Benchmarks: `event/dispatch_bench_test.go`, `event/loop_bench_test.go`,
  `geometry/{pack,grid,place}/bench_test.go`, `ttk/theme_bench_test.go`,
  `canvas/bench_test.go` (display-free), `tests/bench_test.go` (display).
- A test that needs a display calls `testutil.RequireDisplay(t)` (or the
  package's `requireDisplay`) first; an integration test uses
  `testutil.NewTestApp(t)`. `RequireDisplay` skips without `DISPLAY` or
  `WAYLAND_DISPLAY`, starts a private Xvfb through
  `displaylock.UseVirtualDisplay()` when Xvfb is installed
  (`TAKIGO_TEST_DISPLAY=real` uses your screen instead) and holds
  `internal/displaylock`, because `go test ./...` runs package binaries in
  parallel on one screen where their windows overlap. Any new check of
  `DISPLAY` in a test must call `UseVirtualDisplay` first.
- On a real desktop a window manager places and stacks windows as it
  likes: display tests that count X errors, open several generations of
  Apps or aim at window positions call `testutil.Settle()` between
  generations and measure positions instead of assuming them
  (`rootOrigin`/`uncoveredPoint` in `dnd_test.go`); a test that would send
  events over foreign windows skips.

## Common tasks

| Task | Where |
|---|---|
| Add a Tk widget option | `widget/AGENTS.md`; mirror the Tk C/Tcl source, setter + `computeGeometry`/`display` |
| Add a ttk widget option or theme | `ttk/AGENTS.md` |
| Add a platform capability | `platform/AGENTS.md`; extend `platform/display.go`, implement in all three backends |
| Fix a wrong-looking demo | the `tk-demo-compare` skill (`.agents/skills/`), `demos/AGENTS.md` |
| Add a binding tag or pattern | `bind/table.go` (`BindingTable`) + `bind/pattern.go` (Tk pattern syntax) |
| Adjust default widget colour/font | `widget/palette.go`; read with `widget.PaletteFor(app)`, never a literal |
| Adjust DPI / unit conversion | `screenunit/screenunit.go` (Tk's `tkCmds.c` ScalingCmd); `screenunit.SetScreenDPI` is called once in `NewApp` from the screen metrics + `Xft.dpi` |
| Understand a Tk semantic | grep `tk/library/<file>.tcl` and `tk/generic/<file>.c` |
| Cut a release | `make release VERSION=vX.Y.Z` (`scripts/release.sh`); check `make apidiff` first |
| Regenerate the option reference | `python3 scripts/options_doc.py > docs/options.md` |

## Environment variables

| Variable | Effect |
|---|---|
| `DISPLAY` / `WAYLAND_DISPLAY` | When unset, display-dependent tests skip |
| `TAKIGO_TEST_DISPLAY=real` | Display tests use your screen instead of a private Xvfb |
| `TAKIGO_CLASSIC=1` | Tk-exact drawing (no anti-aliased canvas), like `takigo.Classic()` |
| `TAKIGO_APPEARANCE` | `light` or `dark`: overrides the desktop's appearance for `appearance.System` |
| `TAKIGO_FREEZE_TIMERS=1` | `event.Loop.After` drops positive-delay timers (screenshot scripts set it) |
| `TAKIGO_DUMP_TREE=<file>` | Rewrite the widget tree as JSON every 250ms when it changes (`internal/treedump`) |
| `TAKIGO_DEBUG_NAME_WIDGETS=1` | `XStoreName` every child window with its Go name, for `xdotool` |
| `HEADLESS=1`, `PIN_FONTS`, `SKIP_IF_EXISTS`, `WISH`, `LLM_TOOL`, … | screenshot pipeline: `scripts/README.md` |
