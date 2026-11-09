# canvas/ — the canvas widget

Applies to `canvas/`; the root `AGENTS.md` and `widget/AGENTS.md` still
apply. Read the Tk file a part ports before changing it:

| Go | Tk source |
|---|---|
| `canvas.go`, `tags.go`, `events.go`, `scroll.go`, `findmove.go`, `item.go` | `tk/generic/tkCanvas.c` (widget, display list, find, pick, bindings, origin) |
| `item_line.go`, `item_polygon.go`, `item_rectoval.go`, `item_arc.go` | `tkCanvLine.c`, `tkCanvPoly.c`, `tkRectOval.c`, `tkCanvArc.c` |
| `item_text.go`, `item_image.go`, `item_bitmap.go`, `item_window.go` | `tkCanvText.c`, `tkCanvImg.c`, `tkCanvBmap.c`, `tkCanvWind.c` |
| `stipple.go`, dashes in `itemopts.go` | `tkCanvUtil.c` (`Tk_CanvasSetStippleOrigin`, `Tk_GetDash`) |
| `spline.go` | `tkTrig.c` (`TkMakeBezierCurve`) |
| `canvas_postscript.go`, `postscript.go`, `postscript_options.go` | `tkCanvPs.c` |
| `postscript_fonts.go`, `prolog.go` | `tkFont.c` (`Tk_PostscriptFontName`), `library/mkpsenc.tcl` |
| `smooth.go` | none: takigo's anti-aliased drawing (see below) |

## How the canvas differs from other widgets

- It keeps its own redraw path instead of `Base.SetDisplayProc`: a damage
  rectangle in canvas coordinates (`eventuallyRedraw`, Tk's
  `Tk_CanvasEventuallyRedraw`), repainted at idle by `redrawDamage` into a
  private pixmap with Tk's 30px overdraw, then `CopyArea` to the window.
  `scheduleRedraw` damages the whole window, border included.
- The exported `Display()` paints synchronously; library code asks for a
  redraw with `scheduleRedraw` or `redrawItems`, never `Display()`.
- `Configure` is hand-written, not `widget.Configure`: it fixes the origin
  when the inset changes, confines the view and calls
  `geometry.GeometryRequest` itself. A new canvas option that changes the
  size or inset goes through it.

## Items

- An item type is a struct embedding `ItemBase` (ID, tags, integer bbox,
  state, stipples, `-activefill`/`-disabledfill`) and implementing `Item`.
  `ItemBase.canvas` is the back-pointer to the owning `Canvas`; it gives
  the item the colour cache, font registry, display server, scratch
  buffers and the current item (`fillFor`).
- To add one: a `newXxxItem(..., c *Canvas)` that sets defaults (the
  canvas's `defaultInk()`, never a literal black), sets `canvas` and calls
  `updateBBox`; a `base()` method (`itemBase` finds `ItemBase` through it);
  `updateBBox` porting the Tk `ComputeXxxBbox` (Tk's rounding: first
  point truncated, the rest rounded); `Display` drawing with
  `drawablePoint`/`drawableCoord` (Tk_CanvasDrawableCoords) and
  `c.scratchPoints`; `PointDistance`, `AreaOverlap`, `Scale`,
  `Translate`, `Delete`, `Postscript`; a `Canvas.CreateXxx` that calls
  `createItem`. `SetCoords`, `Configure`, `Scale` and `Translate` must
  recompute the bbox: culling, picking and damage all trust it.
- `ItemOption` is `func(*Canvas, Item) error`: it needs the canvas to
  resolve colours (`c.ColorCache().Resolve`) and fonts. One option serves
  every item type through a type switch, so a new item type adds a `case`
  to each option it accepts (`FillColor`, `OutlineColor`, `Dash`, ...),
  and a new option lists every item type it applies to. Mind Tk's
  per-type meaning: `-fill` is the line colour of a line and the text
  colour of a text item. A failing option returns its error:
  `ItemConfigure` returns it, `createItem` reports it through
  `OptionFailed` and skips the remaining options.
- An item change is bracketed by `c.redrawItems(entry)` before and after,
  so both the old and the new box get repainted (`ItemConfigure`,
  `SetItemCoords`, `Move` do this; copy the pattern).

## Display list, tags, IDs

- `c.items` is the display list, bottom to top; `paint` walks it in order,
  skipping hidden items and culling by bbox against the damaged area.
  `idMap` maps IDs to entries; `tagIndex` maps a tag to its entries, kept
  up to date by `ItemBase.AddTag`/`RemoveTag`, `addItem` and `Delete`.
- `Delete` only marks entries `dead`; they leave `c.items` at the next
  `compact`, which every reader of `c.items` calls first. New code that
  iterates `c.items` must too.
- Item methods take a `Selector` (`ItemID` or string). `resolve` handles
  decimal IDs, `"all"`, `"current"` and plain tags, in display order. Tk's
  tag expressions (`&&`, `||`, `!`, `^`) are not supported. `"all"` and
  `"current"` are never stored in `ItemBase.Tags`, so `BindItem` on them
  matches nothing.
- `BindItem` keeps per-ID and per-tag handlers on the canvas, dispatched by
  `dispatchToItem` (ID first, then each tag in tag order); they do not go
  through `bind.Engine`.

## Picking and events

`events.go` ports `CanvasBindProc` and `PickCurrentItem`: `findClosest`
returns the topmost non-hidden item within `-closeenough`, with a bbox
prefilter (`bboxNear`). While a button is down the current item keeps the
pointer (`pointerState`, `leftGrabbed`), so a drag's release reaches the
item it started on; `pick` redraws items with `-activefill` when the
current item changes. Keep that ordering when touching it;
`canvas/pick_test.go` covers it.

## Scrolling

`xOrigin`/`yOrigin` are the canvas coordinates at the inner (inset) edge,
starting at the inset so canvas 0 lies under the border as in Tk.
`setOrigin` ports `CanvasSetOrigin` with `-confine` always on and no
scroll increments. Confinement depends on the window size, which the
geometry manager sets at idle: `Configure` confines against the current
size and the ConfigureNotify handler confines again once the new size
arrives (see "Layout timing" in `widget/AGENTS.md`).

## Embedded windows

A `WindowItem`'s `Display` is a no-op: `paint` calls
`positionWindowItems` after the `CopyArea`, which moves, maps and
`window.MarkMapped`s each visible window and unmaps the rest.
`nWindowItems` counts them, so update it when adding or deleting one.
PostScript emits only a placeholder outline for them.

## PostScript

`Canvas.Postscript` builds a `PSContext` and calls each visible item's
`Postscript` twice: first with `ps.Prepass` set, when items only register
their fonts and must emit nothing, then for real. The prolog is Tk's verbatim, so output stays
byte-comparable with Tk's: keep new item output in Tk's operator order.

## Anti-aliasing (`smooth.go`)

A departure from Tk (see the root `AGENTS.md`): line, polygon,
rectangle/oval and arc implement `smoothItem.paintSmooth` and are
rasterized with `internal/nanosvg` into a `layer`, blended item by item
so a damaged-area repaint gives the same pixels as a full one. Text,
images, bitmaps, windows and stippled items go to the display server
between layers (`paintSmooth` returns false to fall back). `c.antialias`
defaults to `!widget.Classic(app)` and is set by `Antialias`; with it off,
every item draws through its Tk-exact `Display`. A new departure must be
gated on `c.antialias`, leave `Display` Tk-exact and keep axis-aligned
shapes pixel-identical.

## Tests and benchmarks

- Display-free (the default): `newBenchCanvas()` in `canvas/bench_test.go`
  builds a canvas over `benchServer` (no-op drawing calls) with a fixed
  7px-wide font; `pick_test.go`, `lifecycle_test.go`, `geometry_test.go`,
  `postscript_test.go`, `xbm_test.go` and the other unit tests use it or
  pure functions.
- Need a display: `damage_test.go` (a damaged-area repaint must equal a
  full repaint) and `bitmap_display_test.go`. Run them after
  touching `paint`, `smooth.go` or an item's `Display`.
- Benchmarks (display-free): `go test ./canvas/ -run '^$' -bench .` —
  pick, find, resolve, create/delete, raise/lower, move, coords, paint.
- Drawing changes shift demo screenshots: `make demo-batch RETAKE=1` and
  `make demo-gate` (`demos/AGENTS.md`).
