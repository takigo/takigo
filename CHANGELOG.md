# Changelog

Versions follow [semantic versioning](https://semver.org); while the major
version is 0, a minor release may break the API and says so here.

## Unreleased

## v0.2.0 (2026-10-07)

v0.1.0 is the last version with the original, Tk-shaped API. The changes
below are breaking unless marked otherwise; each lists what to change.

- **The module path is `github.com/takigo/takigo`** (it was
  `github.com/msorc/takigo`): rewrite the imports.

### Options and errors

- **Distances are typed.** A distance option takes pixels (`int`,
  `float64`) or a `screenunit.Distance`; a string no longer compiles.
  `pack.PadY("1.5p")` becomes `pack.PadY(screenunit.Pt(1.5))`; `"2m"`,
  `"1c"`, `"0.5i"` become `Mm(2)`, `Cm(1)`, `In(0.5)`.
  `screenunit.Px("3p")` becomes `screenunit.Pt(3).Pixels()`, and
  `screenunit.Float("2c")` becomes `screenunit.Cm(2).Float()`.
  `screenunit.Parse` reads Tk's string form and returns an error.
  The text tag options `TagOffsetStr`, `TagLMargin1Str`, ... are merged into
  `TagOffset`, `TagLMargin1`, ...
- **`Configure` returns an error.** An option that cannot be applied (an
  unknown colour or font) keeps the previous value and is reported:
  `errors.Is(err, color.ErrUnknown)`, `font.ErrNotFound`. In a constructor
  the failure goes to the App's logger (`takigo.WithLogger`, default
  `slog.Default()`), no longer to package `log`.
- **Colours and fonts can be values.** Colour options also take
  `color.RGB(r, g, b)` and `color.RGBA(r, g, b, a)`; font options also take
  `font.Attributes`. Names and descriptors still work.
- **`pack.Pack` and `grid.Grid` return an error** (`pack.ErrNotPacked`,
  `grid.ErrBadIndex`) instead of logging.
- **The `ButtonText`-style aliases are gone** from the classic widget
  packages: use `button.Text`, `label.Text`, ...
- **One set of enums.** `option.Orient`, `option.Side`, `option.Direction`
  and `option.Sticky` replace the per-package types; the package names
  (`scale.Vertical`, `pack.Top`, `grid.StickE`) remain as aliases.
  `grid.Sticky` takes an `option.Sticky`, not an `int`.
  `Window.SetCursor` takes a `cursor.Shape`: drop the `uint(...)` cast.

### Widgets

- **Canvas item IDs are `canvas.ItemID`**, and item methods take an ID or a
  tag string directly: `c.Move(id, dx, dy)` instead of
  `c.Move(fmt.Sprintf("%d", id), dx, dy)`.
- **Text positions** are a `text.Index` or an index expression string.
  `Insert`, `Delete`, `See`, `MarkSet`, `TagAdd`, `TagRemove`,
  `WindowCreate` and `ImageCreate` return `text.ErrBadIndex` for a bad
  index; `EndIndex` returns an `Index`; `Index(spec)` resolves an expression.
- **Scrollbar commands** are `func(widget.ScrollRequest)`. Replace the
  closure decoding `("scroll", n, "units")` with `widget.ScrollY(w)` or
  `widget.ScrollX(w)`.
- **Mouse wheel.** The wheel arrives as `event.MouseWheelType` with
  `ev.Delta` (120 per notch, positive up; `ShiftMask` for a horizontal
  wheel), as in Tk 9. Buttons 4-7 no longer arrive as button presses; bind
  `<MouseWheel>` instead of `<Button-4>`.
- `checkbutton.BoolVar` links a `Variable[bool]` (additive).
- A widget name may be empty (one is generated), and a name a sibling
  already has is renamed with a warning.

### Bindings

- `App.BindEng()` is `App.Bind()`.
- `bind.EventData` has `Event *event.Event` and `Window`; the `RawEvent any`
  field and its type assertion are gone.
- `Engine.BindWindow(w, seq, handler)` binds a widget without its path name.
  Sequences can be built instead of parsed: `bind.Key`, `bind.Button`,
  `bind.On`, `bind.Virtual` (additive).

### Rendering

- **Canvas shapes are anti-aliased by default.** Axis-aligned geometry is
  unchanged to the pixel; slanted and curved edges are smooth.
  `canvas.Antialias(false)` restores the old drawing for one canvas;
  `takigo.Classic()` or `TAKIGO_CLASSIC=1` for a whole App.
- Canvas fills and outlines can be translucent (`color.RGBA`).

### New

- `ttk.UseTheme(app, name)` gives one App a theme and re-themes its widgets;
  `ttk/darktheme` adds a dark theme; `appearance.System()` /
  `App.Appearance()` report the desktop's light/dark preference and
  `ttk.UseSystemTheme` follows it.
- `takigo.UseAppearance(appearance.Dark)` and
  `takigo.FollowSystemAppearance()` give an App a dark look: classic widgets
  start from `widget.DarkPalette` and themed widgets use the `dark` theme
  when it is imported. A widget reads its palette when created, so this is
  chosen at start-up, not switched at run time.
- `WmInfo.SetIconPhoto` / `toplevel.IconPhoto` set the window icon.
- `systray.Notify` shows a desktop notification.
- Drag and drop with other applications (XDND, on X11 desktops):
  `App.OnDrop(w, handler)` receives dropped files and text, and
  `App.StartDrag(w, data, done)` drags them out.
- `App.SetClipboardImage` / `App.ClipboardImage` copy and paste images
  (between applications on X11).
- `App.RunContext(ctx)`; `dialog.WithContext(ctx, parent)` cancels a modal
  dialog.
- `App.Lookup(path)`, `Window.Lookup`, `Window.Descendants`,
  `Canvas.Items` (iterators).
- JPEG photos decode; `image.NewPhotoFramesFromGIF` loads animated GIFs.
- `widget.Resources` and `widget.Scheduler`, the two halves of `AppContext`.
- `takigo.Version()` reports the takigo version a program was built with,
  from the build info.

### Removed

- The unused packages `config`, `gc` and `widget/editutil`, and the
  `widget.BindEngine` interface.

## v0.1.0 (2026-10-01)

Baseline with the original, Tk-shaped API, under the module path
`github.com/msorc/takigo`. Not published to the module proxy.
