# takigo code review — 2026-09-29

Scope: general code review plus a performance survey of the whole tree at
commit `85ee10d`. Six focused passes (event/bind hot path; canvas, draw,
font, color, image; text-family widgets; geometry and ttk core; platform
backends; classic widgets, dialogs and App) each read their area in full and
backed claims with throwaway tests or benchmarks. Nothing marked fixed in
`nemotron_review.md` or `docs/architecture-review.md` is repeated here.

Environment notes for reproducing: `tk/` is not vendored in this checkout,
so Tk semantics are cited by upstream file name. `libxft` headers are
absent, so cgo packages (`font/xft`, `platform/x11`, the root package,
`internal/testutil` and therefore the `widget` and `ttk` test binaries) were
reviewed by reading only. `gofmt`, `go vet` and `go fix -diff` are clean on
the no-cgo package set; all display-free tests pass. The bundled
`golangci-lint` 2.5.0 is built with Go 1.25 and refuses the Go 1.27 target,
so lint was not run.

Items marked **[verified]** were re-checked directly against the source by
the consolidating reviewer in addition to the pass that found them.

---

## 1. Fix-first list

These are the defects a user hits immediately or that hang, crash or leak.

| # | Sev | Area | Finding | Location |
|---|-----|------|---------|----------|
| 1 | high | input | NumLock (Mod2) is treated as Ctrl on X11: with NumLock on, `a` moves to line start, `k` kills the line in text, entry and spinbox. **[verified]** | `widget/text/bindings.go:154`, `widget/entry/bindings.go:145`, `widget/spinbox/bindings.go:175` |
| 2 | high | scale | `Scale.ticks()` loops forever when `TickInterval < Resolution/2` (e.g. `TickIntervalOpt(0.25)` with default resolution 1); first redraw hangs the loop. **[verified]** | `widget/scale/scale.go:394-401` |
| 3 | high | menu | Torn-off menus never draw entries: the `Metrics() fontMetrics` assertion can never match `font.Font.Metrics() font.Metrics`, so `display()` returns after the border. **[verified]** | `widget/menu/tearoff.go:271-278` |
| 4 | high | ttk | `Entry` in `readonly` state is fully editable (`Editable` only rejects disabled; key handler only returns for disabled). **[verified]** | `ttk/entry.go:180`, `ttk/entry_bindings.go:63` |
| 5 | high | ttk | `Entry` never scrolls horizontally; `leftIndex` is only used for the xscroll fraction, so text past the right edge is unreachable and the cursor disappears. | `ttk/entry.go:518-564` |
| 6 | high | ttk | `Treeview.Delete` skips every other child: `removeItem` ranges over `item.Children` while splicing the same slice; orphans stay in `tv.items`. **[verified]** | `ttk/treeview.go:314-326` |
| 7 | high | ttk | Style lookup interleaves map-then-default per chain level; Tk resolves maps over the whole chain first. Child `Defaults["-foreground"]` shadows the root disabled map, so disabled entries, spinboxes and treeviews draw black text in every shipped theme. **[verified]** | `ttk/theme.go:23-36` |
| 8 | high | ttk | `ResolveStyle` returns the parent theme's style object when the current theme lacks one, so under clam/alt `TNotebook`, `TLabelframe`, `TCombobox`, `TMenubutton`, `TScale` are painted in default-theme colours. **[verified]** | `ttk/theme.go:163-170` |
| 9 | high | grid | Negative `-row`/`-column` are not rejected; `arrange` indexes `rowOffsets[-1]` and panics. | `geometry/grid/grid.go:1081` |
| 10 | high | canvas | PostScript output emits every item twice: `TakeItemBuf` returns a substring without truncating the builder, then the body is appended again inside gsave/grestore. **[verified]** | `canvas/canvas_postscript.go:103-117`, `canvas/postscript.go:84-93` |
| 11 | high | canvas | Focused text item inserts any keysym below 0x10ffff, including Shift (0xffe1), arrows and F-keys, as characters. **[verified]** | `canvas/events.go:129-133` |
| 12 | high | x11 | `ParseKeyEventIM` ignores `status` and calls `GoStringN` with the returned length even on `XBufferOverflow`, reading past a 64-byte stack buffer on long IME commits. **[verified]** | `internal/xlib/event.go:267-291` |
| 13 | high | cocoa | `CocoaMeasureString`/`CocoaDrawString` alloc an `NSString` and an `NSAttributedString` per call and never release them (file is compiled without ARC). Also `TKContentView`, `NSTrackingArea` and the window delegate are over-retained, so destroyed toplevels leak their backing bitmaps. **[verified]** | `internal/cocoa/cocoa.m:2074-2111`, `~970-1060` |
| 14 | high | windows | Enter/Leave are never delivered: `WM_MOUSELEAVE` is handled but `TrackMouseEvent` is never called and nothing posts `EnterNotifyEvent`; hover states never engage. **[verified]** | `platform/windows/event.go:350,481`, `internal/win32/user32.go:63` |
| 15 | high | leak | Widgets that call `Dispatcher().BindGlobal` discard the `BindingID`; `UnbindID` has no callers outside `event/`, and the destroy hook only clears per-window handlers. Every destroyed ttk Entry, Combobox, Menu and mnemonic Menubutton stays reachable and runs on every ButtonPress/KeyPress for the process lifetime (measured: 500 leaked handlers turn a 287 ns click dispatch into 1235 ns). **[verified]** | `ttk/entry_bindings.go:106`, `ttk/combobox.go:678`, `widget/menu/bindings.go:109`, `widget/menubutton/menubutton.go:401`, `tk.go:190` |

---

## 2. Performance

### 2.1 Structural (largest wins)

**P1. Idle callbacks run between every queued event.** `event/loop.go:118`
calls `processIdleQueue()` at the top of every iteration, before the
`select` on `eventCh`. Tcl's `Tcl_DoOneEvent` services idle handlers only
when no event source is pending. A burst of N Motion/Configure events
therefore produces N full pixmap redraws (`widget.Base.redraw`:
CreatePixmap + display + CopyArea + FreePixmap) and N synchronous
relayouts (panedwindow sash drag, menu hover) instead of one. Reproduced
with 50 queued events and one coalesced `DoWhenIdle` each: 50 idle runs.
Fix: skip `processIdleQueue` while `len(l.eventCh) > 0` (and, in pump
mode, when the last pump produced events). **[verified]**

**P2. Text wrap is superlinear.** Three independent implementations share
the same shape:
- `widget/text/display.go:215-230` `wrapLine` measures the whole remainder
  then binary-searches with prefix measurements from `start`; a 100k-char
  paragraph measures 120M runes per relayout (0.59 s with a trivial font),
  and relayout happens on every keystroke in that line.
- `font/layout.go:16-61` `WrapLines`/`fitLine` re-measures `s[:i]` at every
  word boundary and allocates a string per character on the long-word path;
  1,992 chars are measured 25×, 19,992 chars 192×. Used by canvas text
  `-width`, label `-wraplength` and ttk labels on every geometry pass.
- `widget/message/message.go:247` measures `current+" "+word` from scratch
  per word, and `computeGeometry` bisects width calling it up to ~10×.
Fix: one pass that accumulates per-word (or per-rune advance) widths and
breaks at the last fitting boundary, as `Tk_ComputeTextLayout` does; expose
a `MeasureChars(maxX)`-style API on `font.Font` so xft can stop early.

**P3. Quadratic entry/listbox/treeview measurement.**
- `widget/entry/entry.go:560` `VisibleRange` measures every prefix
  (10k-rune entry: 50M runes per call), called from `notifyScrollbar` on
  every insert/delete. Use `entryutil.RuneIndexAtPixel` (already a binary
  search) from `LeftIndex`.
- `widget/listbox/listbox.go:458` `maxWidth` re-measures all items per call
  and is reached per `Insert` (O(N²) fill with an xscroll command) and per
  visible row inside `display()` for center/right justify. Cache and update
  incrementally as Tk's `listPtr->maxWidth`.
- `ttk/treeview_display.go:297` ellipsis truncation measures every byte
  prefix from the full length down, per cell per frame, and slices bytes so
  it can split a rune. Binary search on rune boundaries; cache per cell.

**P4. ~10 cgo transitions per X event.** `internal/xlib/event.go:12-63`
exposes one C getter per field; a Motion event costs `XNextEvent` + type +
window + 7 accessors + `XFilterEvent`, a KeyPress 13. A scratch benchmark
shows 7 accessor calls at 276 ns vs one struct-filling C call at 73 ns, so
parsing (~300-400 ns) outweighs the whole Go dispatch (222 ns, 0 allocs).
Two heap allocs per event (`xlib.RawEvent`, `platform.RawEvent`). Fix: one
C helper per event class filling a flat struct, or read the fixed-layout
union fields in Go via `unsafe`; pool the raw events.

**P5. Xft creates and destroys an XftDraw per string.**
`font/xft/xft.go:415-419,459-463` pay a RENDER CreatePicture/FreePicture
pair (plus the solid-source Picture) for every `DrawString`, in the redraw
path of every text-bearing widget. Tk keeps one XftDraw per font and calls
`XftDrawChange` when the drawable differs. Cache per `XftFont` and clear it
from the X11 backend's `FreePixmap`/`DestroyWindow` so the stale-Picture
errors the current comment mentions cannot recur.

**P6. ttk draws two to three times per event.** `Dispatcher.Bind` appends,
it does not replace, so the "override bindTtkCommon" comments in
`ttk/checkbutton.go:344-379`, `notebook.go:411-425`, `scrollbar.go:283-298`,
`sizegrip.go:133-146` (and radiobutton, toggleswitch, spinbox, combobox) are
wrong: each Expose and ConfigureNotify draws once via the widget's own
handler and once via `bindTtkCommon`'s `redisplay()`. Checkbutton's Enter
handler calls `ChangeState` (which draws) and then `Display()` again.
Nothing coalesces: `TtkWidget.NeedRedraw` is declared and unused. Remove
the duplicate bindings and route draws through `DoWhenIdle` + `NeedRedraw`
as canvas does.

**P7. ttk layout sizes every element once per ancestor level per draw.**
`ttk/layout.go:145-195` `placeNodes` calls `childrenSize()` per node,
re-sizing the subtree recursively (a TButton label is sized 4× in `Place`
then again in `Draw`). `LabelElement.Size` re-runs `WrapLines` and
`MeasureString` per line every call; `PaddingElement.Size` re-parses string
option values such as `"2.25p"` via `ParsePadding` each time (measured:
`LookupPadding` 118 ns / 2 allocs, `Layout.Size+Place` TMenubutton 3.6 µs /
40 allocs before any text measurement). Parse theme values once at
registration; size children bottom-up once per `Place`.

**P8. Per-frame bitmap upload.** `canvas/item_bitmap.go:98-106` converts
the XBM to a fresh w×h×4 RGBA and `PutImageRGBA`s it (which allocates
another conversion buffer) on every repaint. Cache a pixmap per (fg, bg,
depth), or use a depth-1 bitmap as clip mask for a `FillRectangle` as
Tk's `DisplayBitmap` does.

**P9. Flat containers allocate a full-size pixmap per Expose/Configure.**
`widget/frame/frame.go:138-143` and `widget/toplevel/toplevel.go:205-211`
go through `Base.redraw` to paint a flat background; an interactive resize
of a 1280×1024 toplevel allocates ~5 MB on the server per ConfigureNotify
(and per event, given P1). Draw the border directly and let the X window
background provide the fill, which also needs finding B7 below.

### 2.2 Per-call costs worth trimming

- `widget/label/label.go:303-311,441-447`: `textLines()` (WrapLines) runs
  in both `computeGeometry` and `display()`, plus `TextWidth` per line per
  Expose/hover. `widget/menu/menu.go:801-810` measures underline and
  accelerator widths per entry on every hover redraw. Cache in geometry.
- Every classic `display()` ends with `d.Flush()` (`button.go:477`,
  `label.go:472`, `frame.go:204`, `menu.go:711`, …) although the idle queue
  and `handleRaw` already flush; N redrawn widgets → N+1 XFlush syscalls.
- `widget/text/text.go:813-820` `seeIndex` scrolls one display line at a
  time (up to 1000 iterations, each computing the visible-line set); the
  Fenwick layout cache can answer the target top line in O(log n).
- `canvas/canvas.go:629` `Delete` runs `slices.DeleteFunc` over the whole
  display list, so item-by-item teardown is O(n²) (1.64 ms for 1000 items,
  1.6 µs per delete vs 0.48 µs for `SetItemCoords`).
- `bind/dispatch.go:166,124`: `Sequence.String()` on the hot path for
  multi-event sequences (718 ns / 11 allocs vs 222 ns / 0 without); key
  promotions by `*binding` instead. `bind/table.go:74` takes an RWMutex
  4-5 times per event in a loop-only engine.
- `font/xft/xft.go:333-358`: each rune missing from the primary font runs
  `FcFontMatch` + `XftFontOpenPattern`; a CJK paragraph with N distinct
  runes costs N fontconfig matches on first measure. Check already-opened
  fallbacks with `XftCharExists` first.
- `internal/xlib/draw.go:100-126`: polyline points are copied three times
  (`[]C.short`, then a C `malloc`ed `XPoint` array); `xlib.XPoint` already
  has XPoint's layout, so pass `unsafe.Pointer(&points[0])` directly.
- `platform/windows/draw.go` creates and deletes a pen or brush and
  acquires a DC per primitive; `font/gdi/gdi.go:133-150` creates a
  compatible DC per `MeasureString`. Cache HPEN/HBRUSH in `gcState` and a
  memDC per font.
- `internal/cocoa/cocoa.m:123-181,1488,1517,2131`: every primitive boxes
  IDs into `NSNumber` for dictionary lookups and calls
  `setNeedsDisplay:YES` on the whole view; `CocoaCopyArea` re-images the
  entire source bitmap per call.
- Uncached `InternAtom` round trips: `selection/selection.go:182` per
  paste, `systray/tray.go:43-44` per tooltip, `platform/x11/display.go:305`
  per `WakeEventReader`. A mutex-guarded map in `X11Display.InternAtom`.
- `window/create.go:148,188` sends one `XDestroyWindow` per descendant; X
  destroys subwindows itself.
- `tk.go:366` returns a fresh `&appClipboard{}` per `Clipboard()` call.
- `geometry/grid/grid.go:569,590,896`: `arrange` allocates layout/offset
  slices per pass and does 3-5 map lookups per slot per pass (Tk does the
  same; only worth it if profiling shows grid-heavy resizes).

### 2.3 Unbounded caches

- `color/color.go:63-94`: keyed by raw name, never evicts; "Red"/"red"/
  "RED" are three entries and `dialog/colorchooser.go:53-57` inserts a new
  `#rrggbb` per slider tick. Normalize keys and bound or refcount.
- `font/named.go:57-67,131-137`: `Registry.Get` caches per raw descriptor
  string forever (`Derive` results included); `Define` closes a font that
  widgets still hold, which nils the XftFont and segfaults on the next draw
  (latent, no callers today).
- `wm/wm.go:93-115` `atomCache` is a package-level map keyed by
  `DisplayServer`, unsynchronized and never pruned.

---

## 3. Correctness

### 3.1 Event loop, bindings, focus

- **B1.** `bind.Engine.UnregisterWindow` only deletes the tag entry; path
  bindings stay in the table (`BindingTable.RemoveAll` has no callers) and
  fire for any later window created at the same path. `bind/bind.go:70`.
- **B2.** `Engine.Bind` with an existing tag+pattern appends, and
  `findBinding` uses strict `>`, so a re-bind is dead and the old closure
  is retained. Tk replaces. `bind/table.go:45`, `bind/dispatch.go:115`.
- **B3.** Focus events are delivered twice: `focus.Manager.SetFocus`
  synthesizes FocusOut/FocusIn, then X delivers the real pair (all children
  select `FocusChangeMask`), so `-validate focusin/focusout` runs twice.
  Tk's `TkFocusFilterEvent` swallows the real ones on non-toplevels.
  `focus/focus.go:84-89`, `tk.go:200`.
- **B4.** Fifteen widgets call `SetInputFocus` directly and never inform
  `focus.Manager`, so its state goes stale; there is no `TkFocusKeyEvent`
  equivalent redirecting keys to the focus window, and no KeyPress handler
  checks that its widget has focus, so after Alt-Tab typing goes to the
  widget under the pointer. `HandleFocusOut` is dead code.
- **B5.** `MappingNotify` is dropped; nothing calls
  `XRefreshKeyboardMapping`, so keysyms go stale after `setxkbmap`.
- **B6.** Double/triple-click counts only apply to ButtonPress
  (`<Double-ButtonRelease-1>` never matches) and there is no proximity
  reset. `bind/dispatch.go:323`.
- **B7.** The X window background is baked in as `WhitePixel` at creation
  and widgets only set the Go field; `SetWindowBackground` is called only
  by `ApplyBackgroundRecursive`, so the server paints white between
  map/resize and the idle redraw. `window/create.go:207`.
- **B8.** `grab.Manager` has no callers and `Dispatcher` has no redirect
  hook, so modal dialogs do not grab; `dialog/dialog.go:96` says they do.
  `Dialog.Run` also never returns if the toplevel is destroyed externally
  (no `OnDestroy` closes `done`).
- **B9.** `App.Destroy` frees images and fonts before destroying windows,
  so destroy handlers run against closed fonts. `tk.go:319-325`.

### 3.2 Widgets

- **W1.** `widget/variable.go:463-470` unsubscribes by captured index, so
  removing listeners out of order deletes the wrong one or none; callers
  are the label/checkbutton/radiobutton Destroy paths. Remove by identity.
- **W2.** `ttk/ttklabel.go:24,56`: `LabelTextVariable` subscribes but
  `Label` has no Destroy, so the variable keeps dead labels alive.
- **W3.** `widget/panedwindow/panedwindow.go:47`: `LostContentProc` is a
  no-op, so a destroyed pane stays in `pw.panes` and every arrange calls
  `MoveResizeWindow`/`MapWindow` on window 0.
- **W4.** Notebook panes and the labelframe `-labelwidget` are never
  claimed via `geometry.ManageGeometry`: content size changes don't
  re-request, and destroying a pane leaves a dangling `tab.Window`.
  `ttk/notebook.go:66-139`, `ttk/labelframe.go:35-40`.
- **W5.** `place.Place` registers no `OnConfigure` hook, so relative
  placement inside anything other than a classic Frame or Canvas never
  follows a resize. `geometry/place/place.go:167`.
- **W6.** Pack and grid resize toplevels directly on every re-request,
  ignoring `wm geometry` and interactive resizes. `pack.go:307`,
  `grid.go:1029`.
- **W7.** Grid and pack drop the container state (row/column weights,
  anchor, propagate) when the last content is forgotten, and re-register
  `OnDestroy`/`OnConfigure` closures each time it is recreated (5 cycles →
  5 stale hooks). `grid.go:528`, `pack.go:285`.
- **W8.** `grid.Grid(w, Row(0), Column(0))` is treated as "unspecified"
  and lands on the next row. `grid.go:349`.
- **W9.** ttk scrollbar: `size` goes negative on bars shorter than ~40 px
  (including the initial 1×1 window), producing a negative thumb length
  that wraps to a huge `uint` width. `ttk/scrollbar.go:147-149`.
- **W10.** `widget/spinbox/spinbox.go:224,462`: `SetText` never clamps
  `LeftIndex`; `computeGeometry` then slices out of range (panic reproduced
  after typing 50 chars and calling `SpinUp`).
- **W11.** Text: `TabWidth` is dead (`const tabWidth = 4` is used) and tabs
  are rewritten to spaces on insert, so `Get` differs from what was
  inserted. `sel.last` returns the first range's end and only the first
  selection range is painted. Peers append listeners to `doc.Listeners`
  that are never removed; stipple pixmaps are never freed.
- **W12.** `entrytext.InsertAt` deletes the selection before validation
  can reject the insert; `DeleteSelection` and Ctrl+K bypass `allowEdit`.
  Classic entry Ctrl+K/Ctrl+D likewise bypass `tryEdit`.
- **W13.** Classic `Listbox.Insert/Delete` remap `selected` only, not
  `itemFg/itemBg`, `selAnchor`, `activeIndex`, `topIndex`; page-scroll on
  a ≤2-row list loses its sign (same in treeview).
- **W14.** Spinbox `-format` is ignored whenever `Increment >= 1`; the
  classic `valuesIndex` is never resynced after `SetText`.
- **W15.** Menu: pointer grab is lost after a cascade closes and
  `postedCascade` stays stale (`menu.go:952`, `:407`); motion-driven
  menubar switching calls `PostFromButton`, arming
  `skipGlobalButtonPress` so the next click anywhere is swallowed
  (`menubar.go:253`). The tearoff toplevel skips `wm.Init` and has no
  Destroy.
- **W16.** Checkbutton geometry ignores `-selectimage`; Checkbutton and
  Radiobutton option sets have drifted (`WidthChars` vs `SelectImg`).
- **W17.** File dialog navigates on a single click (no double-click
  timing; `choosedir.go` has the right logic). Font chooser Bold/Italic
  toggles and preview are inert.

### 3.3 Canvas, images, colors

- **C1.** `ItemConfigure(id, Tags(...))` replaces `base.Tags` without
  updating `tagIndex`; `FindWithTag` then returns stale membership.
  `canvas/itemopts.go:80`.
- **C2.** `Delete` never removes `idBindings[id]` or resets `focusItemID`;
  redraw-from-scratch apps accumulate closures. `canvas/canvas.go:608`.
- **C3.** `Destroy` frees only `c.pixmap`; stipple pixmaps in `c.stipples`
  leak on the server. `canvas/canvas.go:444`.
- **C4.** Text item cursor is drawn at `MeasureString(text[:cursorPos])`
  on the first line, wrong for any multi-line or wrapped text.
  `canvas/item_text.go:259`.
- **C5.** `ArcItem.PointDistance`/`AreaOverlap` are bbox stubs; a thin
  arc is picked anywhere in its bounding rectangle. `canvas/item_arc.go:189`.
- **C6.** `color.Parse` discards `ParseUint` errors (`#zzzzzz` → black,
  nil error) and rejects the 9-digit form. `color/color.go:138-162`.
- **C7.** XBM parsers accept negative/huge dimensions (panic in `ToRGBA`);
  the PPM decoder allocates from the header before reading a pixel.
  `canvas/xbm.go:42`, `image/photo.go:117`, `bitmap/bitmap.go`,
  `image/ppm.go:57`.
- **C8.** `nanosvg.Blend` divides by `dstW`; an SVG without width/height
  reaches `BlendOver` with a zero-width buffer and panics.
  `internal/nanosvg/rast.go:747`.

### 3.4 Platform backends

- **X1.** `_NET_WM_PING` is advertised but never answered, so EWMH window
  managers flag the app as not responding. `wm/wm.go:415-419,436`.
- **X2.** `PutImageRGBA` assumes a 0xRRGGBB 32-bpp visual; wrong colours or
  BadMatch on 16/30-bit visuals. `get_rgba_image` swaps the process-wide
  error handler around `XGetImage` (currently unreachable).
  `internal/xlib/pixmap.go:279-336`.
- **X3.** systray dock request puts the icon in `xclient.window` instead of
  the manager. `systray/systray.go:68-74`.
- **WN1.** Windows cursor is set on the window *class*, so one widget's
  I-beam changes the whole app. `platform/windows/cursor.go:40-67`.
- **WN2.** `propsDB` is never freed on `DestroyWindow`; recycled HWNDs
  inherit dead windows' protocols and hints. `GetWindowProperty` always
  reports format 8 even for 32-bit atom data. `platform/windows/property.go`.
- **WN3.** `font/gdi/gdi.go:211` calls `syscall.NewCallback` per
  `ListFamilies`; Go never frees callbacks and caps them, so repeated font
  dialogs eventually panic.
- **WN4.** Windows `PutImageRGBA` discards partial alpha (darkened AA
  edges). `platform/windows/draw.go:352-369`.
- **M1.** `CocoaGetAtomName` returns `[name UTF8String]` from inside a
  drained `@autoreleasepool`. `internal/cocoa/cocoa.m:~1951`.

---

## 4. Cleanup and API

- `gc/` has no non-test importers; its key ignores the `mask` argument.
- `canvas.displayFunc` is assigned and never read.
- `grab/` is dead (see B8): wire it in or delete it.
- `widget.Widget.Configure(opts ...option.Option)` is still dead on all
  classic widgets (architecture-review #1); `widget.BindEngine` and
  `App.BindEngine()` still have no callers (#2); the inline `Draw` type
  assertion is still at `checkbutton.go:487` and radiobutton (#3);
  `config.Table` still has no callers (#16); toplevel creation is now
  duplicated three times (`toplevel.go`, `menu.go`, `tearoff.go`) (#12).
- `Square.Display` draws synchronously with no coalescing; `TearoffWindow`
  duplicates ~150 lines of `Menu.displaySingleColumn`.
- `checkbutton.go:298` still measures `"0"` on every `computeGeometry`
  (the button cache from the earlier review was not carried over).
- `ttk` and `widget` test binaries do not build under `CGO_ENABLED=0`
  (they import `internal/testutil` → root package → `font/xft`), so the
  no-cgo CI job only compiles them; AGENTS.md says they test. Either give
  `testutil` a no-cgo stub or split the display-free tests out.
- `NewButton` builds its ttk layout up to three times per widget.
- Widget test coverage is still thin: button, checkbutton, radiobutton,
  menu, menubutton, message, panedwindow, scale, scrollbar, square,
  toplevel and labelframe have no test files. Several findings above
  (scale hang, tearoff, treeview delete, grid negative index, spinbox
  panic, variable unsubscribe) are one table-driven test each.

---

## 5. What is in good shape

- The pure-Go dispatch path is allocation-free and fast (Motion: 222 ns,
  0 allocs; canvas pick 5 µs, paint 15 µs, 0 allocs for a 1000-item scene).
  The dispatcher's copy-on-write handler slices are safe with off-loop
  `Bind` during dispatch; `After` cancel/ran CAS, `RunOnMain` after `Quit`
  and nested-loop channel handling are correct.
- The text widget's document/tag/layout core (Fenwick-tree layout cache,
  merged disjoint tag ranges, fuzz and bench coverage) is solid; a warm
  redraw of 31 lines costs 31 measurements.
- Geometry managers are a faithful, readable port with proper idle
  coalescing and clean `-in` handling.
- The X11 backend's threading (XInitThreads, single reader goroutine, Sync
  only where Tk syncs, format-32 property packing) is sound.
- Formatting, vet and the `go fix` modernizers are clean; the codebase is
  stdlib-only and the port-source comments make cross-checking with Tk easy.

## 6. Suggested order of work

1. The fix-first table (section 1): 15 items, most are one-line to
   one-function changes; each deserves a regression test.
2. P1 (idle ordering) and finding 15 (BindGlobal leak): both are in
   `event/` + a handful of call sites and improve every widget at once.
3. P2/P3 wrap and measurement passes, then P5 (XftDraw cache) and P4
   (event parsing): these are the remaining measurable CPU costs.
4. Focus rework (B3/B4) and the ttk redraw coalescing (P6/P7), which are
   larger design changes.

---

## 7. Status (branch `review-2026-09-fixes`)

Each fix below has a regression test where the behaviour is testable
without a window manager; the demo parity batch (`demo_batch.sh --retake`,
headless) scores every demo exactly as `master` does.

| Item | Status | Notes |
|------|--------|-------|
| Fix-first 1–12, 15 | done | `platform.CommandMask`; `Dispatcher.BindGlobalFor` ties widget global handlers to their window |
| Fix-first 13, 14, M1 | done, untested | Cocoa retain/release and Windows `TrackMouseEvent`: no macOS SDK / Windows host here (Windows builds and vets) |
| P1 | done | idle queue waits while events are queued (bounded to 64) |
| P2, P3 | done | `font.WrapLines` 2.96 → 0.21 ms, text `wrapLine` 26 → 1.1 ms on 20k chars; listbox keeps `maxWidth` |
| P4 | done | one cgo call per input event; raw events are not pooled yet |
| P5, P6, P8 | done | one XftDraw per display; ttk redraws queued at idle; canvas bitmaps via a depth-1 pixmap |
| P7 | done | elements sized once per `Place`; theme values are still parsed per lookup |
| P9 | unchanged | Tk's `DisplayFrame` also double-buffers; P1 makes it once per event burst |
| 2.2 | mostly done | display-proc `Flush`, `seeIndex`, `Sequence.String`, Xft fallbacks, polyline copies, `InternAtom`, `Clipboard`, label line breaks, canvas `Delete` (1000 one by one: 1.80 → 0.22 ms), GDI pens/brushes/measuring DC, Cocoa GC/pixmap lookups without `NSNumber` boxing (Windows and Cocoa untested on their platforms); not done: menu measurement caching, per-descendant `XDestroyWindow` (Cocoa relies on it), grid allocations |
| 2.3 | done | colours and fonts shared per value, name indexes bounded; `Define` no longer closes live fonts; wm atoms live on `window.Display` |
| B1–B3, B5–B9 | done | local grab via `App.GrabManager` in the event filter; focus echoes swallowed |
| B4 | open | the focus rework (Tk-style internal focus with key redirection) remains |
| W1–W17 | done | W6 honours `wm geometry` and records user resizes (ConfigureEvent); W11 keeps tab characters and lays them out to Tk's default tab stops |
| C1–C8 | done | |
| X1, X3, WN1–WN4 | done | WN* untested on Windows |
| X2 | open | the colour cache assumes 0xRRGGBB TrueColor throughout; fixing `PutImageRGBA` alone would not help |
| Section 4 | partly | dead `canvas.displayFunc` removed, `NewButton` builds its layout once, widget/ttk tests run without cgo; the dead-API decisions (`gc/`, `Widget.Configure`, `BindEngine`, `config.Table`, toplevel/tearoff duplication) are left to the maintainers |

`TestDamageRedrawMatchesFullRedraw` (canvas) fails intermittently when
`go test ./...` runs all packages against one Xvfb server (other test
windows overlap it at 0,0 and take the focus highlight); it fails the
same way on `master` and passes alone.
