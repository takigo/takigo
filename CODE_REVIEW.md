# Code Review: Takigo

**Date:** 2026-03-26
**Reviewer:** Claude (senior engineer review)

## Overall Assessment

Well-structured, ambitious port of Tk 9.1 to Go. Clean package separation, good use of interfaces to break circular imports, and consistent patterns across widgets. Several issues worth addressing, organized by severity.

---

## Critical / High Priority

### 1. ~~XImage buffer leak in C helper (`internal/xlib/pixmap.go`)~~ ALREADY FIXED

Code already has `if (img)` guard at line 63. No action needed.

### 2. ~~XIM resources never freed (`internal/xlib/`)~~ FIXED

`Display.Close()` now calls `XDestroyIC(xic)` and `XCloseIM(xim)` before `XCloseDisplay()`.

### 3. ~~Treeview O(n²) insert performance (`ttk/treeview.go`)~~ FIXED

Added `scheduleRedisplay()` method that batches `rebuildDisplayList()` + `Display()` via `DoWhenIdle`. Multiple structural changes (insert/delete/move/open/sort) before the next idle phase are coalesced into a single rebuild.

---

## Medium Priority

### 4. ~~`SetRawEventHandler` race condition (`event/loop.go`)~~ FIXED

Documented single-goroutine threading model on `Loop` struct and `SetRawEventHandler`. All handlers, idle callbacks, and timer callbacks run on the main goroutine; only `readEvents` runs concurrently.

### 5. ~~TTK widget nil layout dereference (`ttk/widget.go`)~~ ALREADY FIXED

`TtkWidget.Display()` already has `if w.Destroyed || w.Layout == nil { return }` guard at line 84.

### 6. ~~Treeview displayList/displayDepth sync (`ttk/treeview.go`)~~ FIXED

Merged parallel `displayList []*TreeItem` and `displayDepth []int` slices into a single `displayList []displayEntry` where `displayEntry` holds both `item` and `depth`. Cannot get out of sync.

### 7. ~~Dispatcher unbind granularity (`event/dispatch.go`)~~ FIXED

`Bind()` and `BindGlobal()` now return a `BindingID`. Added `UnbindID(id)` to remove a specific handler without affecting others on the same window. `Unbind(w)` still available for bulk cleanup.

---

## Low Priority / Code Quality

### 8. Event channel overflow (`event/loop.go:50`)

Event channel buffer is 64. If event generation outpaces processing, the `readEvents` goroutine blocks on send with no backpressure or overflow detection. Idle queue also grows unbounded.

### 9. `goto` for flow control (`event/loop.go:145, 285`)

`goto pumpDone` / `nestedPumpDone` is non-idiomatic Go. A `for/select` with explicit breaks would be clearer.

### 10. Canvas linear tag scan (`canvas/tags.go:29-36`)

Tag resolution iterates all items. Fine for small canvases, but O(n) per operation hurts with 10k+ items. A tag-to-items index map would fix this.

### 11. Canvas Delete/Raise/Lower rebuild entire slice (`canvas/canvas.go:428-487`)

Creates new slices via append loops. In-place compaction would be more efficient for bulk operations.

### 12. Theme style parent chain has no cycle detection (`ttk/theme.go:21-33`)

`Style.Lookup()` walks the parent chain without cycle guard. Unlikely in practice, but an infinite loop if it happens.

---

## What's Done Well

- **Package architecture** — Interfaces in `window/` and `widget/` cleanly break circular imports. The `GeomManager`, `BindEngine`, `TextProvider`, `WidgetImage` interfaces are well-designed.
- **Resource cleanup** — `DestroyWindow()` cascades correctly (children first, then GC, then unregister, then platform). `App.Destroy()` tears down in proper order (images -> fonts -> windows -> display).
- **Thread safety where it matters** — Theme registry uses `sync.RWMutex`. Color cache is properly locked.
- **Panic discipline** — Only 2 panic calls in the whole codebase, both in `Must*`/`Px()` functions with non-panicking alternatives. Follows Go conventions.
- **unsafe.Pointer usage** — All 67 instances are in FFI code where they're unavoidable, with proper `defer C.free()` patterns.
- **Double-buffering** — Consistent pixmap-based pattern across canvas, text, and TTK widgets. Correct allocate/free lifecycle.
- **Test quality** — Table-driven tests, proper error condition checks, resource cleanup with `defer app.Destroy()`.
- **TODOs are minimal** — 13 total, mostly in demo code. No critical features left as TODOs.

---

## Recommendations

| Priority | Action | Effort |
|----------|--------|--------|
| ~~P0~~ | ~~Add XCreateImage NULL check in `put_rgba_image`~~ | ALREADY FIXED |
| ~~P0~~ | ~~Free XIM/XIC in `Display.Close()`~~ | FIXED |
| ~~P1~~ | ~~Batch/deduplicate `rebuildDisplayList()` in treeview~~ | FIXED |
| ~~P1~~ | ~~Guard TTK `Display()` against nil layout~~ | ALREADY FIXED |
| ~~P2~~ | ~~Document single-goroutine requirement for event loop~~ | FIXED |
| ~~P2~~ | ~~Merge `displayList`/`displayDepth` into single struct slice~~ | FIXED |
| P3 | Add tag index map to canvas for large item counts | 2-3 hr |
| ~~P3~~ | ~~Add per-handler unbind to dispatcher~~ | FIXED |

The codebase is in solid shape for its scope. The main areas to harden are resource cleanup edge cases in the C layer and treeview performance for large datasets.
