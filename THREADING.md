# Threading Contract

This document specifies which types are safe for concurrent access from multiple goroutines and which must only be used from the **event loop goroutine**.

## Threading Model

takigo uses a **single-threaded event loop** (running on the main goroutine) with channel-based communication for cross-goroutine operations:

```
┌─────────────────────────────────────────────────────────────┐
│                      Main Goroutine                          │
│  ┌────────────────────────────────────────────────────────┐  │
│  │                    Event Loop                             │  │
│  │  • Dispatches events to handlers                         │  │
│  │  • Runs idle callbacks (DoWhenIdle)                      │  │
│  │  • Runs timer callbacks (After)                          │  │
│  │  • Runs cross-goroutine callbacks (RunOnMain)            │  │
│  └────────────────────────────────────────────────────────┘  │
│                    ▲            ▲             ▲               │
│                    │            │             │               │
│              eventCh      idle queue    main queue (+ wake)  │
│                    │            │             │               │
└────────────────────┼────────────┼─────────────┼───────────────┘
                     │            │             │
        ┌────────────┘            │             └────────────┐
        ▼                         ▼                          ▼
┌──────────────┐          ┌─────────────┐            ┌──────────────┐
│ Reader       │          │ Other       │            │ time.AfterFunc│
│ Goroutine    │          │ Goroutines  │            │ (stdlib)      │
│ (X11:        │          │ (app code)  │            │              │
│  NextEvent)  │          │             │            │              │
└──────────────┘          └─────────────┘            └──────────────┘
```

- **Event loop goroutine**: Owns all UI state (windows, widgets, geometry, focus, selection, bindings). All handlers, idle callbacks, and timer callbacks execute here.
- **Reader goroutine**: Exactly one per `Loop`, started by the first `Run`/`RunNested` and shared by nested loops so events stay in order. Blocks on `NextEvent()` (X11) or the C ring buffer (macOS), posts raw events to `eventCh`.
- **Application goroutines**: Can safely call the posting APIs (`DoWhenIdle`, `After`, `RunOnMain`, `Quit`) to schedule work on the event loop. These never block — they append to mutex-guarded queues and wake the loop — so they are also safe from handlers on the loop goroutine and before `Run`.

---

## Goroutine-Safe Types

These types use internal synchronization (`sync.Mutex`/`sync.RWMutex` or channel-based APIs) and can be called from **any goroutine**.

| Type | Package | Safe Methods | Notes |
|------|---------|--------------|-------|
| `*event.Dispatcher` | `event` | `Bind`, `BindGlobal`, `Unbind`, `UnbindID` | Uses `sync.RWMutex`. `Dispatch()` is loop-only. |
| `*event.Loop` | `event` | `DoWhenIdle`, `After`, `RunOnMain`, `Quit`, `SetRawEventHandler` (before `Run`) | Append to mutex-guarded queues; never block. `Quit` is idempotent. `Run()`/`RunNested()` are loop-only. |
| `*takigo.App` | `takigo` | `DoWhenIdle`, `After`, `RunOnMain`, `Quit`, `Dispatcher()`, `ColorCache()`, `FontRegistry()`, `ImageRegistry()`, `Server()`, `BindEngine()`, `Clipboard()`, `FocusManager()`, `WmInfo()`, `RegisterCloseHandler`, `UnregisterCloseHandler` | Delegates to `Loop` or returns synchronized registries. `RunNestedLoop` is loop-only. |
| `*color.Cache` | `color` | `Get`, `GetByValue` | Uses `sync.RWMutex`. |
| `*font.Registry` | `font` | `Define`, `Get`, `GetAttrs`, `Derive`, `Close` | Uses `sync.Mutex`. |
| `*image.Registry` | `image` | `Register`, `Get`, `Unregister`, `DestroyAll` | Uses `sync.RWMutex`. |
| `platform.DisplayServer` | `platform` | All methods | Backend-dependent. X11 uses thread-safe Xlib calls; Cocoa/Win32 require main-thread calls (enforced by event loop). |

---

## Loop-Only Types

These types **must only be accessed from the event loop goroutine**. They have no internal synchronization and concurrent access causes data races.

| Type | Package | Reason |
|------|---------|--------|
| `*window.Window` | `window` | Hierarchy (`Parent`, `Children`), geometry (`X`, `Y`, `Width`, `Height`, `ReqWidth`, `ReqHeight`), flags, `GC`, `WmData`, `BackgroundHook`. |
| `*window.Display` | `window` | Window lookup map, screen info. |
| `*widget.Base` / all widgets | `widget` | All fields: visual options, `NeedRedraw`, `Destroyed`, `Border`. |
| `*draw.Border` | `draw` | Precomputed 3D border colors. |
| Geometry managers (`*pack.Packer`, `*grid.Gridder`, `*placer.Placer`) | `geometry/pack`, `geometry/grid`, `geometry/place` | Package-global `packers` map, per-container layout state. |
| `*wm.WmInfo` | `wm` | Per-toplevel WM state (title, geometry, protocols, handlers). |
| `*focus.Manager` | `focus` | Focus traversal ring, focused window tracking. |
| `*bind.Engine` | `bind` | Binding tables, tag chains, class bindings. |
| `*event.Loop` (internal fields) | `event` | `rawHandler`, `idleQueue`, `dispatcher`, `pumper` — only `Run()` goroutine touches these. |

---

## Cross-Goroutine API

Use these methods to safely schedule work on the event loop from any goroutine:

```go
// Schedule fn to run during next idle phase (coalesced).
app.DoWhenIdle(fn func())

// Schedule fn to run after duration d on the event loop.
app.After(d time.Duration, fn func())

// Schedule fn to run on the event loop goroutine (for UI updates from background goroutines).
app.RunOnMain(fn func())

// Stop the event loop (safe from any goroutine).
app.Quit()

// Nested event loop for modal dialogs (must be called from a handler on the event loop).
app.RunNestedLoop(done <-chan struct{})
```

**Do not** call widget methods, window methods, or geometry managers from background goroutines. Instead, capture needed data and use `RunOnMain`:

```go
// ❌ Wrong: called from background goroutine
button.Configure(widget.TextOpt("new text"))

// ✅ Correct: schedule on event loop
app.RunOnMain(func() {
    button.Configure(widget.TextOpt("new text"))
})
```

---

## Platform Backend Considerations

- **X11**: `platform.DisplayServer` methods are generally thread-safe (`xlib.OpenDisplay` calls `XInitThreads` before the first connection). The reader goroutine calls `NextEvent()` concurrently with the event loop calling `Flush()`.
- **macOS/Cocoa**: All AppKit calls **must** run on the main thread. The event loop runs on the main thread; `EventPumper.PumpEvents()` is called from the event loop. Background goroutines must use `RunOnMain`.
- **Windows**: Similar to macOS — window messages must be processed on the thread that created the window. The event loop runs on that thread.

---

## Summary Table

| Category | Types | Access Pattern |
|----------|-------|----------------|
| **Goroutine-safe (sync)** | `Dispatcher`, `Loop` (channel methods), `App` (channel methods), `color.Cache`, `font.Registry`, `image.Registry`, `selection.Manager` | Any goroutine |
| **Loop-only (no sync)** | `Window`, `Display`, `Base`/widgets, `Border`, geometry managers, `WmInfo`, `focus.Manager`, `selection.Manager`, `bind.Engine` | Event loop goroutine only |
| **Platform-dependent** | `DisplayServer` | X11: any; Cocoa/Win32: main thread only (enforced by loop) |

---

## Enforcement

- `go vet` and `-race` detector will catch most violations.
- The `event.Loop` struct documents its threading contract in its type comment (`event/loop.go:22-26`).
- `widget.Base` and `window.Window` have no synchronization — treat them as **not thread-safe**.