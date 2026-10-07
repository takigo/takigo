# takigo architecture review

Date: 2026-10-08, after the review and clean-up for v0.3.0 (see
`CHANGELOG.md`). The previous review of 2026-08-31 and its follow-ups on
the `modernize` branch are closed; this one records the state of the code
and the decisions that stand.
Scope: package layout, platform abstraction, widget API, geometry and event
subsystems, what is shared and what is deliberately not.

## Strengths

- **Clean layered architecture.** `takigo` → `widget` / `ttk` / `canvas` /
  `dialog` → `window` / `event` / `geometry` / `bind` / `focus` / `grab` /
  `wm` → `platform` → `internal/{xlib,cocoa,win32}`. There are no upward
  imports; `go list` confirms it. The three backends sit behind
  `platform.DisplayServer` (`platform/display.go`, 101 methods in 11
  capability interfaces), composed from pure-Go code with a thin cgo layer.
- **One shape for every widget.** `New(parent widget.Caregiver, name string,
  opts ...XxxOption)` plus a typed `Configure(opts ...XxxOption) error`;
  colour, font and distance options are generic over `color.Spec`,
  `font.Spec` and `screenunit.Length`; a failed option keeps the old value
  and is reported (`errors.Is(err, color.ErrUnknown)`).
- **Shared implementations where Tk shares them.** The button family
  (label, button, checkbutton, radiobutton) draws through
  `widget/internal/tkbutton`, a port of `tkUnixButton.c` told apart by type;
  the classic entry and spinbox edit through `widget/internal/entryedit`
  (`tkEntry.c`); the ttk entry, combobox and spinbox through
  `ttk/internal/entrytext` and `drawFieldText`; the three geometry managers
  register with `geometry.Registry`, which holds the container plumbing of
  `Tk_MaintainGeometry` once.
- **Threading model is explicit** (`THREADING.md`, `event/loop.go`): the
  loop goroutine owns all UI state; the only other goroutines post to
  channels. Several Apps may run at once, each on its own goroutine; their
  per-window state lives on the windows (`window.Window.Value`,
  `geometry.Table`), not in package maps.
- **Rendering parity is measured.** The 68 demos are compared with Tk 9.1
  pixel by pixel and tree by tree (`scripts/demo_batch.sh`,
  `demos/parity.tsv`); every change to the widgets above kept all 62
  comparable demos at their baseline.
- **Redraws coalesce** as in Tk: `widget.Base.EventuallyRedraw` (a pixmap
  per redraw), `TtkWidget.redisplay` (a persistent back buffer), the canvas
  (damage rectangles) and the text widget each port their Tk widget's
  display procedure.

## Decisions that stand

- **`widget.AppContext` is what a widget needs and no more**: `Root()` (the
  per-App anchor for palettes, themes and the classic flag), `Dispatcher`,
  `FocusManager`, `Clipboard`, the `Resources` and the `Scheduler`.
  `*takigo.App` implements it; a widget reaches the display server through
  its window. `Caregiver` (a window plus its context) is what constructors
  take, so the App and every widget can be a parent.
- **`ttk` stays one package.** Its ~480 exports are what Tk's `ttk::`
  namespace has; splitting it into `ttk/<name>` would collide with the
  classic packages (`button` and `ttk/button`), force the shared helpers
  into the public API and churn every demo. Its options are named alike
  instead: `<Widget><Option>`, a `<Widget>Style` on every widget, one
  `FieldState` for the entry-like widgets.
- **`ttk.Panedwindow` wraps the classic widget.** Tk has a separate
  `ttkPanedwindow.c`; the wrapper serves until a themed one is ported. It
  has no `PanedwindowStyle` option for that reason.
- **Screen DPI is process-wide** (`screenunit`), set by the last `NewApp`.
  Two Apps on displays with different DPIs, or one App across monitors with
  different DPIs, get one scale. Tk has the same model per interpreter;
  changing it means threading a window through every distance conversion.
- **`platform.DisplayServer` is X-shaped** (atoms, properties, GC value
  masks, `uint64` pixels) because Tk emulates Xlib on Windows and macOS the
  same way. Methods a backend cannot provide are no-ops there and are
  listed in the backend's package doc (`platform/cocoa/display.go`); drag
  and drop and the system tray are X11-only.
- **`window.Add{Mapped,Unmapped,Moved}Hook` are exported** for the
  geometry managers, which register in `init()`; they are not safe to call
  at run time.
- **Custom widgets are supported**: `widget.InitBase`, `widget.Configure`,
  `window.NewChildWindow`, `MakeWindowExist` and `platform.DisplayServer`
  stay public. `demos/square/square` (Tk's `tkSquare.c`) is the example.

## Open

- **`ttk.Panedwindow`** pushes ttk-only flags into `widget/panedwindow`
  (`FlagSash`, `Weighted`, `GripSize`). A port of `ttkPanedwindow.c` would
  remove them.
- **Disabled images are not stippled.** Tk stipples a disabled widget's
  image (`TkpDisplayButton`); the button family draws it as is.
- **The ttk theme engine** (`theme.go`, `layout.go`, `element(s).go`,
  `state.go`, `box.go`, about 1,600 lines) could be its own package so that
  the theme packages stop importing the widgets; `UseTheme` re-themes live
  widgets, which keeps a hook in `ttk` either way.
- **Test depth** is uneven: the parsers, geometry managers, canvas and text
  are well covered; `internal/xlib` and `platform/cocoa` have no tests of
  their own, and the Windows backend only unit tests. The screenshot
  pipeline catches rendering regressions on X11 only.
- **`testutil.Settle`** is a fixed 200 ms sleep for real-desktop runs; the
  virtual display does not need it.
