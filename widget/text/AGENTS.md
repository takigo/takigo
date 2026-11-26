# widget/text/ — the text widget

Applies to `widget/text/`; the root `AGENTS.md` and `widget/AGENTS.md`
still apply. Read the Tk file a part ports before changing it:

| Go | Tk source |
|---|---|
| `text.go` (Insert/Delete/See, scroll API, embedded windows) | `tk/generic/tkText.c`, `tkTextWind.c` |
| `line.go` (`Document`: lines, marks, tag ranges, images) | `tkTextBTree.c`, `tkTextMark.c`, `tkTextImage.c` |
| `index.go`, `indexexpr.go` | `tkTextIndex.c` (`GetIndex`, `ForwBack`, `StartEnd`) |
| `tag.go`, `tags.go` | `tkTextTag.c` |
| `display.go`, `layout.go` | `tkTextDisp.c` (`LayoutDLine`, `DisplayDLine`, `GetYView`) |
| `bbox.go` | `TkTextIndexBbox`, `TkTextDLineInfo` (`tkTextDisp.c`) |
| `tabs.go` | `TkTextGetTabs`, `AdjustForTab`, `SizeOfTab` (`tkTextDisp.c`) |
| `image.go` | `tkTextImage.c` (image create/configure/names, `-align`) |
| `search.go` | `TextSearchCmd` (`tkText.c`) |
| `dump.go` | `TextDumpCmd`, `DumpLine` (`tkText.c`) |
| `blocks.go` | margin colours of `DisplayDLine`; block boxes, rules and markers are takigo's |
| `floats.go` | none: floating images and `TagClear` are takigo's |
| `undo.go` | the undo part of `tkText.c` |
| `bindings.go`, `selection.go` | `library/text.tcl` |

## Document model

- There is no B-tree. `Document` holds `Lines []*Line`, each a `[]rune`
  with no newline, and no final newline; `Marks map[string]*Mark`;
  `Tags map[string]*Tag` plus `tagsByPriority`, each `Tag` owning its
  sorted, disjoint, merged `[]TagRange`.
- `Index{Line, Char}`: `Line` 1-based, `Char` a 0-based rune offset.
  `Clamp` pulls an index into the document; `Document.Insert`/`Delete`
  clamp their arguments. Methods take an `IndexSpec` (`Index` or a Tk
  index string) and resolve it with `indexOf`; a bad string is
  `ErrBadIndex`. `indexexpr.go` documents the grammar and how `"end"`
  maps onto a document without the final newline (`"end"` and
  `"end -1c"` give the same position).
- Embedded windows and images are a placeholder rune (`U+FFFC`) in the
  text plus a right-gravity mark just before it, registered with
  `addObject`; `Delete` drops the objects whose placeholder it removes and
  tells `subscribeObjects` subscribers. `Get` and the selection strip
  placeholders, and undo does not restore them, as in Tk.

## Marks and tags

- `insert` has right gravity, `current` left; marks created by `MarkSet`
  get right gravity (Tk's default); `MarkGravity` changes it.
  `Document.Insert`/`Delete` adjust every mark and tag range before
  changing the lines.
- A tag's priority is its creation order. `sel` is created by `New` with
  `selPriority` (`1 << 30`), above every other tag, standing in for Tk's
  `tag raise sel`; `putTag` never hands that priority out. `TagRaise` and
  `TagLower` reorder `tagsByPriority` and renumber: tags below `sel` from
  0, those raised above it from `selPriority`, so a new tag still goes
  below the selection. Attributes resolve lowest priority first, so the
  highest-priority tag that sets an attribute wins.
- `TagConfigure` bumps `layoutGen` (part of the layout cache key);
  `TagAdd`/`TagRemove` notify listeners only when the tag
  `affectsLayout`. A new layout-affecting tag option must be counted in
  `affectsLayout`.
- `TagBind` keeps per-widget handlers by event string
  (`normalizeTagEvent`): `<Enter>`, `<Leave>` (`updateTagHover`),
  `<Motion>`, `<Button-N>`, `<Double-Button-N>`, `<Triple-Button-N>` and
  `<ButtonRelease-N>`. The widget counts clicks itself (`clickCounter`,
  the binding engine's thresholds) and `fireTagClick` runs each tag's
  most specific binding. Pointer motion and clicks move the `current`
  mark first, as `TkTextPickCurrent` does.
  `TagCursor` (not in Tk) is shown by `updateTagCursor`, falling back to
  the widget's `Cursor`.

## Edits and undo

- `Document.Insert`/`Delete` change the model only. Undo is recorded by
  the `TextWidget` methods and the key bindings (`RecordInsert`,
  `RecordDelete`); an edit made through the `Document` directly is not
  undoable.
- `UndoStack` grouping: a one-byte, non-newline insert adjacent to a
  one-insert group extends it; one-byte deletes extend a one-delete group
  forwards or backwards (Delete, BackSpace); otherwise actions of the same
  type join the open group and a type change closes it. A few key
  bindings in `bindings.go` (Return among them) close the group with
  `Separator`; `Edit("separator")` is Tk's
  `edit separator`. Depth is `undoMaxDepth`.
- The history belongs to the document (`Document.UndoStack`), so every
  peer undoes the same history, as with Tk's shared text.

## Change notification and peers

- After every edit and layout-affecting tag change the document calls
  `Listeners`, then the `Subscribe` subscribers, with a `Change`
  (`From`..`To` changed, `Delta` lines inserted or removed).
- Every widget, the first one included, joins through `shareDocument`:
  it subscribes `layout.apply` first, then the callback that runs
  `keepViewInDocument`, updates the scrollbar and schedules a redraw. The
  order matters: the view clamp reads the already-updated layout.
- `NewPeer(doc, ...)` shares a `Document`. `New` and `NewPeer` each set
  the widget defaults; a new default goes into both. Images live in the
  document, so every peer shows them; embedded windows live on the widget
  that created them (`embeddedWindows`), tracked by a mark in the shared
  document.

## Display

- `display` is registered with `SetDisplayProc`, so it draws into the
  base redraw's pixmap (`widget/AGENTS.md`, Redraws); `renderTo` clears,
  lays out the visible lines and draws them, then
  `positionEmbeddedWindows` moves and maps the embedded windows.
- `layoutCache` (`layout.go`) holds each logical line's wrapped
  `displayLine`s and heights per widget, with Fenwick trees for the
  pixel and display-line totals. It is keyed on width, font, wrap mode,
  `layoutGen` and the embedded objects; a `Change` invalidates the lines
  it names. Read layout through `t.layout.line`, never by re-wrapping.
- `segmentsForRange` splits a display line into `textSegment`s of uniform
  style from the tag ranges on the line; per-line properties (margins,
  spacing, justify, wrap, tabs, block box, marker, rule) come from
  `resolveLineProps`, from the tags on the line's first character.
- `measureRange` must agree with the segments: it takes the fast path
  (the widget font) only when `needsSegments` finds no tag font, elide or
  tab stops on the line, and otherwise sums the segments' widths.
- Elided characters are zero-width segments; a line whose characters and
  newline are all elided gets zero-height display lines (`validate`).
  Eliding a newline alone does not join two lines, unlike Tk.
- A block box (`Tag.hasBlock`) runs over consecutive lines that start
  with the same tag while it covers the newlines between them; the
  opening and closing lines get the border and padding added to their
  spacing (`outerSp1`/`outerSp3` keep the spacing outside the box).
  Whether a line opens or closes a box depends on its neighbours, so
  `layoutCache.apply` invalidates the lines either side of a change (and
  the line after inserted ones). Boxes nest: `resolveBlocks` orders them
  outermost first (margins, then the lower priority outside) and
  `drawBlocks` draws them in that order; rounded corners are blended over
  the box around them or the widget background, and drawn square in a
  Classic App.
- Floating images (`ImageFloatOpt`) take no width in their line; the
  display lines beside them get wider margins. Once the document has a
  float (`layoutKey.floats`), a line's layout depends on the floats
  carried from the line above (`lineLayout.carryIn`/`carryOut`): `line`
  first `settle`s the stale lines above it in order, and a line whose
  carry-out changes invalidates the next one. `wrapLineFloats` breaks one
  display line at a time because the margins depend on the heights above.
  `TestLayoutCacheMatchesRecomputeFloats` checks the cache against a fresh
  layout under random edits.

## The view

`topLine` (1-based), `topCharOffset` (display line within it) and
`topPixelOffset` (pixels of that display line above the view, Tk's
`topPixelOffset`) are the view; `viewTop()` is its first document pixel
and `setViewTop` moves it to a pixel. `computeVisibleLines` starts the
first line at `y = -topPixelOffset`, so drawing, `indexFromPixel`, `BBox`
and `DLineInfo` all shift together; `renderTo` clears the `-pady` strips
afterwards because the display server has no clip. Pixel scrolling
(`YViewMoveTo`, `YViewScrollPixels`, pages, the wheel) keeps an offset;
scrolling by units, `YView(line)` and `See` of a line above the view
reset it, as Tk's `YScrollByLines` and `TkTextSetYView` do.
`keepViewInDocument` runs inside the change notification: lines removed
above the view move `topLine` up, a view whose line went away falls back
to the line the removal joined it to, and `clampView` pulls it into the
document. `clampScrollPosition` keeps the last page full to the pixel:
the highest top is `total - view`, part way into its line. Scroll and
yview code (`yviewFractions`, `computeVisibleLines`,
`computeDisplayLinesBefore`) must tolerate a stale `topLine`; code that
moves the view sets all three fields.

## Tests

- Display-free: `newTestPeer` in `peer_test.go` and `newBenchWidget` in
  `bench_test.go` build a widget on a bare `window.Window` with a fixed
  7px-wide font and no display server; the index, tag, undo, layout,
  embed, peer, view-clamp (`view_clamp_test.go`) and rich-tag
  (`richtag_test.go`) tests use them. They have no App, so set colour and
  font fields on `doc.tag(name)` instead of calling `TagConfigure`.
- `richtext/` turns HTML into tags (parser, CSS cascade, builder); its
  tests check the builder's plan without a display.
- Needs a display: `peer_display_test.go`; `tests/ime_test.go` drives
  the input method through an App.
- Benchmarks (display-free): `go test ./widget/text/ -run '^$' -bench .`.
