# Code Review: Takigo

**Date:** 2026-03-26
**Reviewer:** Claude (senior engineer review)

## Overall Assessment

Well-structured, ambitious port of Tk 9.1 to Go. Clean package separation, good use of interfaces to break circular imports, and consistent patterns across widgets. Several issues worth addressing, organized by severity.

---

## Critical / High Priority

### 1. XImage buffer leak in C helper (`internal/xlib/pixmap.go`)

`put_rgba_image` allocates a buffer, calls `XCreateImage()`, but never checks if `XCreateImage` returns NULL. If it fails, the buffer leaks silently.

**Fix:** Add `if (!img) { free(buf); return; }` after `XCreateImage()`.

### 2. XIM resources never freed (`internal/xlib/`)

`xim` and `xic` (input method) are initialized but never closed. `Display.Close()` should call `XCloseIM(xim)` / `XDestroyIC(xic)`.

### 3. Treeview O(n²) insert performance (`ttk/treeview.go:286-296`)

Every insert at a non-end index copies the entire children slice. Inserting 1000 items = ~500k item moves. `rebuildDisplayList()` is also called on every insert/delete/move/open/close, walking the entire tree each time — and sometimes multiple times per operation.

**Fix:** Batch `rebuildDisplayList()` calls via idle callback, or at minimum deduplicate consecutive rebuilds.

---

## Medium Priority

### 4. `SetRawEventHandler` race condition (`event/loop.go:180`)

`l.rawHandler` is written by `SetRawEventHandler()` and read by `handleRaw()` without synchronization. Safe today because both run on the main goroutine, but undocumented and fragile.

**Fix:** Either protect with a mutex, or document that it must be called before `Run()`.

### 5. TTK widget nil layout dereference (`ttk/widget.go`)

`RefreshTheme()` can set `w.Layout = nil` if the new theme has no layout template. `Display()` doesn't guard against this — will panic.

**Fix:** Guard `Display()` with `if w.Layout == nil { return }`.

### 6. Treeview displayList/displayDepth sync (`ttk/treeview.go`)

These two slices are rebuilt together in `rebuildDisplayList()` but accessed independently in display code. If they ever get out of sync (e.g., concurrent event during display), it's a panic.

**Fix:** Use a single `[]displayEntry` struct slice instead of two parallel slices.

### 7. Dispatcher unbind granularity (`event/dispatch.go:47`)

`Unbind()` removes **all** handlers for a window. If multiple subsystems bind to the same window (focus manager + bind engine + widget), one unbind tears everything down.

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
| P0 | Add XCreateImage NULL check in `put_rgba_image` | 5 min |
| P0 | Free XIM/XIC in `Display.Close()` | 15 min |
| P1 | Batch/deduplicate `rebuildDisplayList()` in treeview | 1-2 hr |
| P1 | Guard TTK `Display()` against nil layout | 5 min |
| P2 | Document single-goroutine requirement for event loop | 15 min |
| P2 | Merge `displayList`/`displayDepth` into single struct slice | 30 min |
| P3 | Add tag index map to canvas for large item counts | 2-3 hr |
| P3 | Add per-handler unbind to dispatcher | 1 hr |

The codebase is in solid shape for its scope. The main areas to harden are resource cleanup edge cases in the C layer and treeview performance for large datasets.
