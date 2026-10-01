# takigo architecture review

Date: 2026-08-31 (status re-checked 2026-10-01)
Scope: package layout, platform abstraction, widget API, geometry/event subsystems, observability of dead/duplicate code.

## Strengths

- **Clean layered architecture.** `takigo` → `widget` / `ttk` → `window` / `event` / `geometry` → `platform` → `internal/<x11|cocoa|win32>`. The three backends sit behind ~330 lines of `platform.DisplayServer` interface in `platform/display.go`.
- **Functional options API is consistent** across all 50+ `New(parent, name, opts...)` constructors. The `Opt`-suffix convention for keyword collisions (`Width`, `Text`, `Command`, ...) reads cleanly.
- **Threading model is correct and well-documented** at `event/loop.go:21-27`: the loop goroutine owns `rawHandler`, `idleQueue`, and dispatch state; the reader goroutine only writes to channels.
- **Cocoa pump-mode is cleanly abstracted** via the optional `EventPumper` interface (`event/loop.go:12`). The loop unconditionally checks for it, so there is no `//go:build` switching in the loop body.
- **`bind.Engine` integrates via `BindGlobal(AllEventsMask, ...)`** so widget-level handlers fire first and Tk-style tag chains still work (`bind/bind.go:24`).
- **Geometry manager abstraction** (`window.GeomManager` / `geometry.Manager` alias) cleanly separates pack / grid / place without widget-level coupling.
- **`widget.AppContext` interface** breaks the takigo ↔ widget import cycle without leaking concrete types.

## High-impact

### 1. `Configure()` signature is inconsistent and effectively dead on classic widgets
- `widget.Widget` interface (`widget/widget.go:52`) requires `Configure(opts ...option.Option)` (type-erased). Classic widgets satisfy it.
- Their setters are typed `ButtonOption = func(*Button)`, **not** `option.Option`. So `b.Configure(button.Text("x"))` does not compile.
- `Canvas.Configure` (`canvas/canvas.go:323`) and `ttk.Entry.Configure` (`ttk/entry.go:402`) do it right with typed `CanvasOption` / `EntryOption`.
- No demo calls `Configure` on a classic widget; the feature is dead code.
- **Fix**: drop `widget.Widget.Configure` from the interface; let each widget declare its own typed option slice. Then `option.Option` / `option.Apply` / `option/option_test.go` (97 LOC) can be deleted.
- **Status**: done (2026-09-30). Every classic and ttk widget has a typed `Configure(opts ...XxxOption)` backed by `widget.Configure` / `ttk.configure`; `option.Option` and `option.Apply` are gone.

### 2. `BindEngine` interface is unused surface
- `widget.BindEngine` (`widget/widget.go:90`) defines `RegisterWindow` / `UnregisterWindow`.
- `App.BindEngine()` returns that narrow interface; `App.BindEng()` returns the concrete `*bind.Engine`.
- Only `App.BindEng()` is actually called (in `cmd/bind_demo`, `demos/colors`, `form`, `image2`, `search`).
- **Fix**: delete `widget.BindEngine` interface and the `App.BindEngine()` method; keep `App.BindEng()`.
- **Status**: done (2026-10-01). The interface and `BindEngine()` are gone; `App.BindEng()` is now `App.Bind()`.

### 3. Image draw calls use unnecessary inline type assertions
- `widget/checkbutton/checkbutton.go:441-446` and `widget/radiobutton/radiobutton.go:387`:
  ```go
  if photo, ok := img.(interface {
      Draw(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
           depth int, imgX, imgY, w, h, dstX, dstY int, bgPixel uint64)
  }); ok { photo.Draw(...) }
  ```
- `widget.WidgetImage` (`widget/widget.go:21`) already declares that exact `Draw` signature, and `*image.Photo` implements it.
- **Fix**: call `img.Draw(d, w.Drawable(), gc, w.Depth, 0, 0, imgW, imgH, ix, iy, bgPixel)` directly. The whole `if photo, ok := ...` block disappears.
- **Status**: open (`checkbutton.go:486`, `radiobutton.go:411`).

### 4. Duplicate WM atom cache
- `window.Display` interns `WMDeleteWindow` / `WMProtocols` (`window/display.go:47-48`) and stores them on the struct. They are used once in `window/create.go:64` and never read again.
- Meanwhile `wm.WmInfo` has its own `wmAtoms` cache (`wm/wm.go:77`) that re-interns the same atoms (plus more) per `DisplayServer`.
- **Fix**: remove `Display.WMDeleteWindow` / `Display.WMProtocols` fields and their `InternAtom` calls; keep `wmAtoms` as the single source of truth.
- **Status**: partly done (2026-10-01). `Display.WMProtocols` is removed; `WMDeleteWindow` is still used by `window/create.go:68`, so it needs to move to the `wm` cache first.

### 5. Widget test coverage is essentially zero
- 288 source files, 30 test files (4,579 LOC of tests). Tests concentrate in `bind`, `geometry/grid`, `geometry/pack`, `canvas/postscript`, `widget/text` (pure data only).
- No tests for Button, Label, Frame, Entry, Checkbutton, Radiobutton, Scrollbar, Scale, Spinbox, Listbox, PanedWindow, Menu, Menubutton, Message, Toplevel, Labelframe, Square, or Canvas widget construction. No tests for `color.Parse` / `Cache`, `screenunit.Px`, `wm.ParseGeometry`, or `geometry.UsableWidth`.
- Most of these are pure Go and need no X server. The screenshot pipeline catches rendering regressions but not algorithmic ones.
- **Fix**: table-driven tests for `color.Parse`, `screenunit.Px`, `wm.ParseGeometry`, `geometry.{UsableWidth, GeometryRequest}`, plus per-widget `computeGeometry` math via a stub font registry.
- **Status**: partly done. `frame`, `label`, `listbox`, `scale`, `spinbox` and `text` have tests, and `screenunit`, `wm`, `bind` and the geometry managers are covered. Still untested: button, checkbutton, radiobutton, menu, menubutton, message, panedwindow, scrollbar, square, toplevel, labelframe. The repository now has 105 test files.

## Medium-impact

### 6. `AppContext` interface forces a self-referential method
- `AppContext` requires `AppContext() AppContext` (`widget/widget.go:108`) purely to satisfy `Caregiver`. Every implementation hand-codes this trivial method; easy to forget.
- **Fix**: split `Caregiver` (Window + getter for services) from `AppContext`; pass `*App` via a build-tag shim.
- **Status**: open.

### 7. `widget.AppContext` is doing too much (16 methods)
- Bundles dispatcher + idle + timer + clipboard + quit + nested loop + close handlers + color / font / image registries + self-reference.
- **Fix**: split into `widget.Resources` (registries), `widget.EventHost` (dispatcher / idle / after), `widget.Clipboard`. Pass only what each widget needs.
- **Status**: open; `AppContext` has grown to about 16 methods.

### 8. Rendering duplication across interactive widgets
- `widget/button/button.go:222-317`, `widget/checkbutton/checkbutton.go:285-468`, and `widget/radiobutton/radiobutton.go` all reimplement: background fill, 3D border draw, active / disabled / pressed / focus color switching, anchor placement, text-draw-with-font, highlight ring.
- `widget.Base.DrawBackground` and `widget.Base.DrawHighlightBorder` already extract parts of this.
- **Fix**: introduce `widget.RenderState` and `widget.DrawShell(d, w, rs)`. Eliminates ~150 LOC of repetition and makes visual parity easier to maintain.
- **Status**: open.

### 9. `computeGeometry` duplicated per widget
- Button, Checkbutton, Label, Text all repeat: measure text, max(textH, imageH), pad + inset, set `ReqWidth` / `ReqHeight`. `widget.CompoundSize` exists (`widget/widget.go:282`) but no helper for the full layout.
- **Fix**: add `widget.MeasureRequest(textW, textH, img, compound, padX, padY, borderWidth, highlightWidth) (reqW, reqH int)`.
- **Status**: open.

### 10. `cursor.Shape` is not enforced as the parameter type
- `Window.SetCursor(shape uint)` (`window/window.go:132`) takes a raw `uint`. Typos like `w.SetCursor(uint(cursor.XTerm + 1))` compile silently.
- **Fix**: change to `SetCursor(shape cursor.Shape)` with a typed enum.
- **Status**: done (2026-10-01). `Window.SetCursor` and `platform.CursorManager.SetCursorShape` take `cursor.Shape`.

### 11. `AppContext.Window()` and `Caregiver.Window()` overlap
- `Base.Win` is the widget's own window; `AppContext.Window()` returns the root window. Two ways to get a window from the same object.
- **Fix**: drop `AppContext.Window()` and let widgets call `app.Root()`-style methods explicitly.
- **Status**: open.

## Low-impact / cleanup

### 12. Toplevel construction duplicates Window creation logic
- `widget/toplevel/toplevel.go:82-128` manually constructs a `Window` and calls `d.Server.CreateWindow` / `CreateGC` / `RegisterWindow` / `AddChild` directly, duplicating `window/create.go:CreateMainWindow`.
- **Fix**: add `window.NewToplevelWindow(d, name, transientFor)`.
- **Status**: open. The pattern is now repeated in `toplevel.go`, `menu.go` and `tearoff.go`.

### 13. `WindowID` and friends are `uintptr` aliases
- Works on every supported platform (all 64-bit). Worth a one-line comment in `platform/types.go` documenting that assumption, since `uintptr` size is technically platform-dependent.
- **Status**: open.

### 14. `event.TypeToMask` is a switch
- Trivial; could be a `[N]Mask` table. Micro-perf, no correctness impact.
- **Status**: open.

### 15. `Window.BackgroundPixel` mutated in three places
- `Base.InitBase`'s `BackgroundHook`, every widget's `Display()`, and `ApplyBackgroundRecursive`. The "mirrors `Base.Background.Pixel`" invariant is convention only — no enforcement.
- **Fix**: add `Window.SetBackground(c *color.Color)` that updates pixel + Background + X attribute atomically.
- **Status**: done (2026-09-30) as `Window.SetBackgroundPixel`, the only writer of `BackgroundPixel` outside `window/`.

### 16. `config.Table` is dead infrastructure
- `config/table.go` and `config/types.go` provide a typed cget / configure mechanism, have tests, and no production caller. Tk's cget / configure introspection is unimplemented.
- **Fix**: delete, or pilot with Button to validate the pattern.
- **Status**: done (2026-10-01). `config/` is deleted, with the equally unused `gc/` and `widget/editutil/`.

### 17. Package doc on `takigo` is one line
- `pkg.go.dev` will show "Package takigo is a pure Go port of the Tk GUI toolkit." The first paragraph of `AGENTS.md` would make a much better godoc.
- **Status**: done (2026-10-01). The package doc is in `doc.go`, with examples in `example_test.go`.

### 18. `option` package mixes value types and the generic option mechanism
- `option.Option`, `option.Apply`, `option.Relief`, `option.Anchor`, `option.Justify`. The value types belong with widgets; `Option` / `Apply` exist only because of issue #1.
- **Status**: done (2026-09-30). `Option` / `Apply` removed; the package holds only the value types.

### 19. `App.RegisterCloseHandler` / `App.UnregisterCloseHandler` are dead
- Toplevels have `OnClose` / `OnDeleteWindow`. None of the 66 demos call the App-level helpers.
- **Status**: obsolete. `dialog/dialog.go` now uses `RegisterCloseHandler` / `UnregisterCloseHandler`, so they are not dead.

### 20. No widget-by-name lookup
- `PathName` and `Name` are stored on `Window` but no `Display.LookupByName(path string) *Window` helper exists. `bind.BindingTable` works on string tags but cannot tell you which widget resolved.
- **Status**: open.

### 21. Direct-draw vs pixmap-draw inconsistency
- Button / Label / Frame draw straight to the window. Canvas / TtkWidget / TextWidget use offscreen pixmaps. Different behaviour during resize / expose. Document or unify.
- **Status**: resolved. Classic widgets redraw through `Base.EventuallyRedraw` into a pixmap at idle time (`widget/widget.go`), like Tk's `REDRAW_PENDING` + `Tk_GetPixmap`.

### 22. Package-level state in demo helpers
- `demos/demohelper.go:337` has `var codeWindow` / `varsWindow`. Fine for a single demo process; would break under tests or multi-window demos.
- **Status**: open (now `demos/demohelper/`).

## Suggested implementation order

1. ~~**#1** typed Configure~~ — done
2. **#3** drop inline type assertions — trivial
3. **#4** dedupe WM atom cache — trivial
4. ~~**#2** pick a side for BindEngine~~ — done
5. ~~**#10** typed `cursor.Shape`~~ — done
6. ~~**#16** decide `config` package fate~~ — deleted
7. **#5** remaining widget unit tests — biggest coverage gap
8. **#8 / #9** extract `RenderState` / `MeasureRequest` — biggest LOC reduction

Items #6, #7, #11, #12, #13, #17 are nice-to-haves that can be deferred or skipped.

## Open questions

- **Typed `Configure` direction.** Resolved (2026-09-30): per-widget typed options, `option.Option` dropped.
- **`config` package fate.** Resolved (2026-10-01): deleted.
