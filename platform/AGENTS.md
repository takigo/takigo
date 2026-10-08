# platform/ — the display servers

Applies to `platform/`, `platform/<x11|cocoa|windows>/` and the cgo bridges
in `internal/xlib`, `internal/cocoa`, `internal/win32`; the root `AGENTS.md`
still applies.

- **The boundary.** `platform.DisplayServer` in `platform/display.go` is
  the single abstraction over windowing: `DisplayCore` plus the capability
  interfaces (`WindowManager`, `Drawer`, `GCManager`, `PixmapManager`,
  `EventSource`, `GrabManager`, `SelectionManager`, `CursorManager`,
  `PropertyManager`, `InputMethodManager`). It is X-shaped (atoms,
  properties, GC value masks, `uint64` pixels) because Tk emulates Xlib on
  Windows and macOS the same way; a method a backend cannot provide is a
  no-op there and is listed in the backend's package doc
  (`platform/cocoa/display.go`). Drag and drop and the system tray are
  X11-only.
- **Backends.** Each `platform/<os>/` exposes
  `NewDisplayServer(name) (platform.DisplayServer, font.FontOpener, error)`,
  selected at compile time by the `platformInit` shim in `tk_<os>.go`. The
  returned `DisplayServer` is a composed wrapper, so backend methods that
  are not part of a capability interface are **not** reachable by type
  assertion on it — return them explicitly from `NewDisplayServer`, as
  `FontOpener` is. Each backend has `var _ platform.Xxx = (*XxxDisplay)(nil)`
  compile-time checks.
- **Adding a capability.** Add the method to the matching capability
  interface in `platform/display.go` (or a new one composed into
  `DisplayServer`) and implement it in all three backends. For optional,
  backend-specific behaviour use an optional interface checked by type
  assertion, as `event.EventPumper` is in `event/loop.go`.
- **Build tags** use `//go:build <goos>` at the very top of the file,
  before the `package` line (`tk_x11.go`, `font/named_*.go`,
  `font/{xft,gdi,coretext}/`, `systray/systray.go`). Keep the cgo surface
  minimal: the bulk of each backend is pure Go in `platform/<os>/`, the
  cgo lives under `internal/<os>/`. The Windows backend is pure Go
  (syscall) and is vetted from Linux with
  `CGO_ENABLED=0 GOOS=windows go vet ./...`; the cgo-only darwin files are
  not type-checked by the Linux modernizer gate, so keep them modern by
  hand.
- **Threading.** The loop goroutine owns all UI state; the reader goroutine
  only posts raw events to a channel (`event/loop.go`, `THREADING.md`).
  Several Apps may run at once on different goroutines, so per-window state
  lives on the windows (`window.Window.Value`), never in package maps.
- **cgo gotchas (X11).** `xlib.GC` is a C pointer: test it with
  `xlib.IsZeroGC()`/`xlib.ZeroGC()`. `xlib.Window` is `C.Window`: compare
  with `xlib.Window(0)`, not a bare `0`. The `FC_*` fontconfig macros are Go
  string constants, so the Fc functions go through C helper wrappers. X11
  `#define` constants (`LineSolid`, `CapButt`, …) and the keysyms
  `XK_Prior`/`XK_Next` (`0xff55`/`0xff56`) are numeric constants in Go
  because cgo cannot see macros. XEvent union members are read through C
  helper functions. RGBA images go to X through `put_rgba_image`, which
  converts to BGRA and pre-composites alpha against the background pixel.
- **Tests.** `internal/xlib` and `platform/cocoa` have no tests of their
  own; `platform/x11/x11_test.go` needs a display (`requireDisplay`), the
  Windows backend has unit tests only. See the root `AGENTS.md`, "Tests".
