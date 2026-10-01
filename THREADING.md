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
| `*event.Dispatcher` | `event` | `Bind`, `BindGlobal`, `Unbind`, `UnbindID` | Copy-on-write handler lists under a `sync.RWMutex`; a handler unbound during a dispatch is not called. `Dispatch()` is loop-only. |
| `*event.Loop` | `event` | `DoWhenIdle`, `After`, `RunOnMain`, `Quit`, `SetRawEventHandler` (before `Run`) | Append to mutex-guarded queues; never block. `Quit` is idempotent. `Run()`/`RunNested()` are loop-only. |
| `*takigo.App` | `takigo` | `DoWhenIdle`, `After`, `RunOnMain`, `Quit`, `Done`, `Dispatcher()`, `ColorCache()`, `FontRegistry()`, `ImageRegistry()`, `Server()`, `Bind()`, `Clipboard()`, `FocusManager()`, `GrabManager()`, `WmInfo()` | Delegates to `Loop` or returns the (shared) singletons; what you may call on those depends on their own row. `RunNestedLoop`, `RunNestedLoopContext`, `UpdateIdleTasks`, `RegisterCloseHandler` and `UnregisterCloseHandler` are loop-only. |
| `*color.Cache` | `color` | `Get`, `GetByValue` | Uses `sync.RWMutex`. |
| `*font.Registry` | `font` | `Define`, `Get`, `GetAttrs`, `Derive`, `Close` | Uses `sync.Mutex`. |
| `*image.Registry` | `image` | `Register`, `Get`, `Unregister`, `DestroyAll` | Uses `sync.RWMutex`. |

---

## Loop-Only Types

These types **must only be accessed from the event loop goroutine**. They have no internal synchronization and concurrent access causes data races.

| Type | Package | Reason |
|------|---------|--------|
| `*window.Window` | `window` | Hierarchy (`Parent`, `Children`), geometry (`X`, `Y`, `Width`, `Height`, `ReqWidth`, `ReqHeight`), flags, `GC`, `WmData`, `BackgroundHook`. |
| `*window.Display` | `window` | Window lookup map, screen info. |
| `*widget.Base` / all widgets, canvas, text | `widget`, `ttk`, `canvas` | All fields: visual options, `NeedRedraw`, `Destroyed`, `Border`; canvas items, text documents. |
| `*widget.Variable[T]` | `widget` | `Set` runs the listeners, which configure widgets. |
| `*image.Photo` | `image` | Pixels and the pixmap cache. Decoding a file into a new Photo is fine anywhere; once a widget shows it, change it on the loop. |
| `*draw.Border` | `draw` | Precomputed 3D border colors. |
| Geometry managers (`pack`, `grid`, `place`) | `geometry/pack`, `geometry/grid`, `geometry/place` | Per-container layout state, kept on the windows (`Window.Value`). |
| `*wm.WmInfo` | `wm` | Per-toplevel WM state (title, geometry, protocols, handlers). |
| `*focus.Manager` | `focus` | Focus traversal ring, focused window tracking. |
| `*grab.Manager` | `grab` | The current grab window. |
| `*bind.Engine` | `bind` | Binding tables, tag chains, class bindings. |
| `platform.DisplayServer` | `platform` | On X11 concurrent calls do not corrupt memory (Xlib is thread-safe after `XInitThreads`), but the state they act on is the loop's: shared GCs, the input context, sequences of requests that must not interleave. Cocoa/Win32 calls must come from the loop's thread. |
| `*event.Loop` (internal fields) | `event` | `rawHandler`, `idleQueue`, `mainQueue`, `dispatcher`, `pumper` — only `Run()` goroutine touches these. |

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

// Closed once Quit has been called.
app.Done() <-chan struct{}

// Run that also quits when ctx is done (safe to cancel from any goroutine).
app.RunContext(ctx) error
```

A modal dialog opened over `dialog.WithContext(ctx, parent)` closes, reporting
"cancelled", when `ctx` is done.

A nested loop runs everything the outer loop would: events, timers, and the
idle and `RunOnMain` callbacks that were queued behind the callback that
started it, in the order they were posted. Widgets behind a modal dialog keep
repainting.

**Do not** call widget methods, window methods, or geometry managers from background goroutines. Instead, capture needed data and use `RunOnMain`:

```go
// ❌ Wrong: called from background goroutine
b.Configure(button.Text("new text"))

// ✅ Correct: schedule on event loop
app.RunOnMain(func() {
    b.Configure(button.Text("new text"))
})
```

To get a value back, send it on a channel and also wait for `app.Done()`:
callbacks still queued when the loop quits never run. Never wait like this on
the loop goroutine itself — the callback cannot run until you return.

```go
res := make(chan int, 1)
app.RunOnMain(func() { res <- b.Window().Width })
select {
case width := <-res:
    use(width)
case <-app.Done():
}
```

---

## Multiple Apps

On X11 a process may run several Apps at once, each with its own display
connection and its own loop goroutine (the goroutine that calls `Run`). Every
rule above then holds per App: an App's windows, widgets and managers belong
to its loop goroutine, and another App's goroutine is a "background goroutine"
to it. `multiapp_test.go` runs this under the race detector.

Some state is process-wide and shared by all Apps:

| State | Rule |
|-------|------|
| ttk themes (`ttk.RegisterTheme`, `Theme` elements, styles, layouts) | The `Theme` methods are goroutine-safe. A style's `Defaults` and `Maps` are plain maps: fill them in before widgets use the style, or while no other App is running. |
| Default theme (`ttk.SetCurrentTheme`) | The process default, used by an App until `ttk.UseTheme(app, name)` gives it its own. `UseTheme` is per App and re-themes that App's widgets; call it on the App's loop goroutine. |
| `screenunit` DPI | Set by every `NewApp`; the last one wins. Reads and writes are atomic. |
| libXft | Serialized by `xlib.XftMu`; nothing to do. |
| X error handler | One for the process; `GetImageRGBA`'s error trap is serialized. |

Windows and macOS support one App at a time: Win32 has one window procedure
for the process, AppKit one `NSApp`. Apps one after another are fine.

---

## Platform Backend Considerations

- **X11**: `xlib.OpenDisplay` calls `XInitThreads` before the first connection, because the reader goroutine calls `NextEvent()` while the event loop draws and flushes. That makes Xlib calls memory-safe from any goroutine; it does not make the display server a goroutine-safe API (see the loop-only table).
- **macOS/Cocoa**: All AppKit calls **must** run on the main thread. `NewApp` must be called from the main goroutine (it fails elsewhere) and the event loop runs there; `EventPumper.PumpEvents()` is called from the event loop. Background goroutines must use `RunOnMain`.
- **Windows**: Window messages are processed on the thread that created the window. `NewApp` locks its goroutine to that thread; call `Run` from the same goroutine.

---

## Summary Table

| Category | Types | Access Pattern |
|----------|-------|----------------|
| **Goroutine-safe (sync)** | `Dispatcher` (bind/unbind), `Loop` (posting methods), `App` (posting methods), `color.Cache`, `font.Registry`, `image.Registry`, `ttk.Theme` methods | Any goroutine |
| **Loop-only** | `Window`, `Display`, `Base`/widgets, `Variable`, `Photo`, `Border`, geometry managers, `WmInfo`, `focus.Manager`, `grab.Manager`, `selection.Manager` (its mutex guards only its own maps; it calls the display server and callbacks), `bind.Engine`, `DisplayServer` | The App's event loop goroutine only |

---

## Enforcement

- Nothing checks at run time which goroutine calls a loop-only API; the `-race` detector reports violations that a test exercises.
- The `event.Loop` struct documents its threading contract in its type comment.
- `widget.Base` and `window.Window` have no synchronization — treat them as **not thread-safe**.