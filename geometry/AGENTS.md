# geometry/ — geometry managers

Applies to `geometry/` and `geometry/{pack,grid,place}`; the root
`AGENTS.md` still applies. `geometry/pack/pack.go` ports
`tk/generic/tkPack.c`, `geometry/grid/grid.go` `tk/generic/tkGrid.c`,
`geometry/place/place.go` `tk/generic/tkPlace.c`; the shared package
ports `tk/generic/tkGeometry.c` (Tk_ManageGeometry, Tk_GeometryRequest,
Tk_SetInternalBorder, Tk_MaintainGeometry) and the structure procs the
three managers share. Read the Tk arrange procedure (ArrangePacking,
ArrangeGrid, RecomputePlacement) before changing a layout rule.

## Shared infrastructure

- `geometry/manager.go`: `ManageGeometry` (tells the old manager via
  `LostContentProc` when another takes the window over),
  `GeometryRequest` (clamps to 1 and to `MinReqWidth/Height`, returns
  early when the request did not change, else calls the manager's
  `RequestProc`), `SetInternalBorder`. `Manager` aliases
  `window.GeomManager`.
- `window.GeomManager` (`Name`, `RequestProc`, `LostContentProc`) lives
  in `window/window.go`, not here: `Window.GeomManager` needs the type
  and `geometry` imports `window`.
- `geometry/registry.go`: `Registry[C Container]`, the
  `window.GeomManager` each of pack, grid and place registers as a
  package singleton (`var reg = geometry.NewRegistry(...)`). It holds the
  per-container state, the `-in` container of each content
  (`ContainerFor`), the destroy/resize hooks of a container (`For`),
  `Forget`, `ArrangeContainer`, `ArrangeAll`, and the
  Tk_MaintainGeometry part: map/move/unmap hooks keep content managed
  `-in` a non-parent container with it. `PlaceContent` is the shared tail
  of every arrange: position, MoveResize, map when viewable,
  `NotifyMoved`.
- `geometry/table.go`: `Table[V]`, a typed `map[*window.Window]V` whose
  entries live on the windows (`Window.Value`). Use it for any
  per-window manager state, never a package map: several Apps run at once
  on different goroutines. Loop-only.
- `geometry/idle.go`: `WhenIdle` coalesces arrange requests into one idle
  callback per container (Tk's REQUESTED_RELAYOUT + Tcl_DoWhenIdle)
  through `Display.DoWhenIdle`, which `NewApp` sets; without it (unit
  tests on `internal/stubs`) the arrange runs at once.
- `geometry/group.go`: `Elementer` and `Group`. `pack.Pack` and
  `grid.Grid` take an `Elementer`, so one widget or a
  `geometry.Group{a, b}` is managed with the same options in one call.

## Layout timing and mapping

- `Pack`/`Grid`/`Place` calls, content requests (`RequestProc`) and
  container resizes call `ScheduleArrange`, which lays out at idle
  (`WhenIdle`), as Tk does. `MainLoop` runs `UpdateIdleTasks` before it
  maps `.`; `App.UpdateIdleTasks` is Tk's `update idletasks`.
- `ArrangeContainer` lays out now. Code that reads `ReqWidth/ReqHeight`
  of a freshly built tree before the loop runs arranges first, children
  before parents: `arrangeNow` in `dialog/dialog.go` does this before
  sizing the dialog.
- Content maps only once its container (and every window up to the
  content's parent) is mapped (`window.ContainerViewable`).
  `window.MarkMapped` runs the registries' mapped hook, which arranges
  and so maps the container's content; anything that maps a window
  without it leaves the subtree unmapped. `window.MarkUnmapped` lets the
  registries unmap foreign `-in` content.
- Consequences for widgets (the App's ConfigureNotify filter,
  `IsViewable` in place of Tk_IsMapped, size-dependent state at idle):
  see "Layout timing" in `widget/AGENTS.md`.

## Other geometry managers

Widgets that manage their own content implement `window.GeomManager`
and claim it with `geometry.ManageGeometry`: `widget/panedwindow`
(panes), `widget/labelframe` (the `-labelwidget`), `ttk/notebook.go`
(tabs), `ttk/labelframe.go` (the label widget). Panedwindow schedules
its re-arrange at idle like the registries; the others re-lay out at
once.

## Grid specifics

- Row/column options are `SlotOption`s (`MinSize`, `Weight`, `Pad`,
  `Uniform`). `Uniform` slots share a minimum proportional to their
  weight (weight 0 counts as 1), as tkGrid.c's ResolveConstraints:
  `step2bUniformGroups`.
- `SetPropagate`/`GetPropagate` port `grid propagate`.

## Tests

- Display-free, on `internal/stubs`: `geometry/manager_test.go`,
  `geometry/grid/grid_test.go` (table-driven), `geometry/pack/pack_test.go`,
  `geometry/place/place_test.go`.
- Fuzz targets in `geometry/grid/grid_fuzz_test.go`: option parsing
  (`FuzzGridOption`, `FuzzSlotOption`) and the layout steps
  (`FuzzResolveConstraints`, `FuzzAdjustOffsets`, `FuzzApplySticky`,
  `FuzzComputeAnchor`).
- Benchmarks (`geometry/{pack,grid,place}/bench_test.go`): arranging 50
  contents, and forget/re-manage for pack and grid
  (`go test ./geometry/... -run '^$' -bench .`).
- Layout changes shift demo screenshots: `make demo-batch RETAKE=1` and
  `make demo-gate` (`demos/AGENTS.md`).
