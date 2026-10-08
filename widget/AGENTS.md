# widget/ — classic widgets

Applies to `widget/` and every `widget/<name>/` package; the root `AGENTS.md`
still applies. Each widget ports `tk/generic/tk<Name>.c` and
`tk/library/<name>.tcl` (option defaults, class bindings); check them before
changing behaviour. `demos/square/square` (Tk's `tkSquare.c`) is the example
of a custom widget built on the same pieces (`widget.InitBase`,
`widget.Configure`, `window.NewChildWindow`).

## Shape of a widget

- **Constructor signature.** Always:
  ```go
  func New(parent widget.Caregiver, name string, opts ...XxxOption) *Xxx
  ```
  Use `parent.AppContext()` to reach app singletons; never import the
  top-level `takigo` package from a widget (import cycle; widgets depend on
  `widget.AppContext`). Set `w.Class` (the Tk class name) in the constructor
  and mark the window focusable with `w.Flags |= window.FlagFocusable` when
  it takes keys.
- **Functional options.** Define `type XxxOption func(*Xxx)` and one
  constructor per settable field. Colour and font options are generic over
  `color.Spec` (a name or `color.RGB(...)`) and `font.Spec` (a descriptor
  or `font.Attributes`) and go through the shared setters
  (`SetBackgroundColor`, `SetForegroundColor`, `SetFont`, `LookupColor`,
  …) — see `Background` in `widget/button/button.go`. A setter that cannot
  resolve its value keeps the previous one and calls `OptionFailed(err)`:
  inside `Configure` the error is collected and returned, in a constructor
  (which returns no error) it goes to the App's `slog` logger
  (`takigo.WithLogger`). Never `log.Printf` from library code.
- **Runtime configuration.** Every widget has a typed
  `Configure(opts ...XxxOption)` taking the same options as its
  constructor (Tk's `configure`). Change options after creation through
  it — never by writing a field and calling `Display()`, or by calling an
  option func directly. It returns the joined errors of the options that
  failed (`errors.Is(err, color.ErrUnknown)`, `font.ErrNotFound`); the
  others still take effect. Implement it with
  `widget.Configure(w, opts, w.computeGeometry)`, which does what Tk's
  WorldChanged procs do: rebuilds the border, recomputes the geometry,
  syncs the window background (`Window.SetBackgroundPixel`), re-arranges
  the content when the margins changed, asks the geometry manager to
  re-lay out when the request changed, and schedules a redraw.
  `SetText`-style methods are thin wrappers over `Configure`.
- **Option naming.** Use the bare Tk option name where possible
  (`button.Text`, `label.Width`, `label.Background`, `label.Relief`). An
  `Opt` suffix is used when the bare name is unavailable or ambiguous in
  that package (`label.FontOpt`, `label.ImageOpt`, `label.JustifyOpt`,
  `scale.FromOpt`, `menu.TearOffOpt`, …) — check the package's existing
  setters before adding one. `docs/options.md` lists every option
  (regenerate with `python3 scripts/options_doc.py > docs/options.md`).
- **Distances.** A distance option is generic over `screenunit.Length`
  (`int | float64 | screenunit.Distance`):
  `func PadX[L screenunit.Length](p L) ButtonOption`, storing
  `screenunit.ToPixels(p)`. Callers pass pixels or `screenunit.Pt(3)`,
  `Mm(2)`, `Cm(1)`, `In(0.5)` (Tk's `3p`, `2m`, `1c`, `0.5i`); a string
  does not compile. In library code write constants as
  `screenunit.Pt(3).Pixels()`. `screenunit.Parse` reads Tk's string form
  and returns an error wrapping `ErrBadDistance`.
- **Adding a Tk option.** Mirror `tk/generic/<file>.c` and
  `library/<widget>.tcl`; add the option setter in `widget/<name>/<name>.go`
  and read the field in `computeGeometry`/`display` — `Configure` then
  handles it at runtime.
- **Defaults.** Colours and fonts come from `widget/palette.go`
  (`LightPalette` mirrors `tk/unix/tkUnixDefault.h`); read them with
  `widget.PaletteFor(app)`, never a literal.

## Events, focus, redraws

- **Event handling.** In the widget's `bindings.go` write `bindXxx(w, app)`
  that calls `app.Dispatcher().Bind(w.PlatformID, event.XxxMask, ...)` for
  each event. Always check `ev.Type == event.XxxType` inside the handler
  because the mask may catch multiple event types. A widget's input
  handlers (keys, buttons, motion, enter/leave) are its class bindings and
  run at its class tag in the `bind.Engine` chain (path, class, toplevel,
  `all`), so a binding on the widget's path runs first and returning
  `true` (break) from it, or `SetBindTags` without the class, suppresses
  them. Handlers for other events (Expose, Configure, Destroy, Focus…) run
  before any binding, like Tk's C event handlers. A widget that needs
  events for every window uses `Dispatcher().BindGlobalFor(w.PlatformID, …)`,
  which goes away with the widget (`BindGlobal` handlers live forever).
  The wheel arrives as `event.MouseWheelType` (bind with `MouseWheelMask`,
  pattern `<MouseWheel>`) with `ev.Delta` at 120 per notch, positive up,
  and `ShiftMask` for a horizontal wheel, as in Tk 9; buttons 4-7 never
  arrive as button events. Turn deltas into scroll units with an
  `event.WheelAccumulator` so high-resolution wheels add up. Test
  "accelerator" modifiers as `ControlMask|platform.CommandMask`: Mod2 is
  Command on macOS but NumLock on X11 and Windows.
- **Focus.** Take the keyboard focus with `widget.Focus(app, w)` (Tk's
  `focus`), never `SetInputFocus`: as in Tk, the X focus stays on the
  toplevel and `focus.Manager` redirects key events to the focus widget,
  so a key handler receives keys only while its widget has the focus. A
  class handler that consumes a key (Text's `<Tab>`) sets `ev.Handled` so
  Tab traversal leaves it alone. Popups (menus, dropdowns) take the focus
  on post and give it back on unpost.
- **Redraws.** Draw in a private `display()` registered with
  `Base.SetDisplayProc`; the exported `Display()` only calls
  `Base.EventuallyRedraw`, which coalesces requests into one idle-time
  redraw into a pixmap (Tk's `REDRAW_PENDING` + `Tk_GetPixmap`). Draw
  through `w.Drawable()`, never `platform.WindowDrawable(w.PlatformID)`,
  so the redirection applies.
- **Layout timing.** pack/grid relayout synchronously inside
  `GeometryRequest`; Tk does it at idle. The App's event filter rewrites
  child ConfigureNotify sizes to `Window.Width/Height` and
  `window.IsViewable()` stands in for `Tk_IsMapped`. State computed
  incrementally from sizes (a treeview's column remainder, a canvas
  `-confine` origin) must do its work at idle or gate it on `IsViewable`,
  never read `ev.ConfigWidth` for child windows. Anything that maps a
  `Window` calls `window.MarkMapped`, or its children never map.

## Shared internals

- `widget/internal/tkbutton`: the button family (label, button,
  checkbutton, radiobutton), a port of `tkUnixButton.c` told apart by type.
- `widget/internal/entryedit`: editing shared by the classic entry and
  spinbox (`tkEntry.c`); `internal/textedit` holds the text measuring and
  word helpers every entry-like widget uses.
- Top-of-file comment names the Tk source(s) the file ports:
  ```go
  // Package button implements the button widget with press/hover interaction.
  // It ports the button-specific parts of tk/generic/tkButton.c and
  // library/button.tcl.
  ```
