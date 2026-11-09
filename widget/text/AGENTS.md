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
  `tag raise sel`; `putTag` never hands that priority out. There is no
  `tag raise`/`tag lower` yet. Attributes resolve lowest priority first,
  so the highest-priority tag that sets an attribute wins.
- `TagConfigure` bumps `layoutGen` (part of the layout cache key);
  `TagAdd`/`TagRemove` notify listeners only when the tag
  `affectsLayout`. A new layout-affecting tag option must be counted in
  `affectsLayout`.
- `TagBind` keeps per-widget handlers by event string; only `<Enter>`,
  `<Leave>` (`updateTagHover`) and `<Button-1>` are fired.

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
  spacing, justify) come from `resolveLineProps`.

## The view

`topLine` (1-based) and `topCharOffset` (display line within it) are the
view. `keepViewInDocument` runs inside the change notification: lines
removed above the view move `topLine` up, a view whose line went away
falls back to the line the removal joined it to, and `clampView` pulls it
into the document. `clampScrollPosition` keeps the last page full.
Scroll and yview code (`yviewFractions`, `computeVisibleLines`,
`computeDisplayLinesBefore`) must tolerate a stale `topLine`; code that
moves the view sets both fields.

## Tests

- Display-free: `newTestPeer` in `peer_test.go` and `newBenchWidget` in
  `bench_test.go` build a widget on a bare `window.Window` with a fixed
  7px-wide font and no display server; the index, tag, undo, layout,
  embed, peer and view-clamp tests (`view_clamp_test.go`) use them.
- Needs a display: `peer_display_test.go`; `tests/ime_test.go` drives
  the input method through an App.
- Benchmarks (display-free): `go test ./widget/text/ -run '^$' -bench .`.
