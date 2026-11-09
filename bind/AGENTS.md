# bind/ — event bindings

Applies to `bind/`; the root `AGENTS.md` still applies. The package ports
`tk/generic/tkBind.c` (pattern parsing, the tag chain, match ordering,
promotion of multi-event sequences, Double/Triple); the default virtual
events stand in for the `event add` lines of `tk/library/tk.tcl`. Read the
matching tkBind.c function before changing a rule: most functions here
name the one they port.

## Files

- `bind/bind.go`: `Engine` API — `Bind`/`Unbind` on a tag, `BindWindow`
  on a widget's path, `SetBindTags`/`BindTags`, `AddVirtualEvent`,
  `GenerateEvent`, `Install`. `Bind` is a generic method over
  `SequenceSpec` (a pattern string or a `Sequence`).
- `bind/pattern.go`: `Modifier`, `Pattern`, `Sequence`, the builders
  (`Key`, `Button`, `On`, `Virtual`, `Then`), `Parse`, `matches`.
- `bind/table.go`: `BindingTable`, the per-tag binding lists.
- `bind/dispatch.go`: the chain walk, promotion, virtual matching, click
  counting, default tag chain.
- `bind/compare.go`: which of several matching bindings wins (Compare,
  CompareModMasks, IsBetterMatch).
- `bind/virtual.go`: default virtual events. `bind/keysym.go`: key names.

## Patterns

- `Parse` takes Tk syntax: `<Modifier-...-Type-Detail>`, several in a row
  for a multi-event sequence, `<<Name>>` for a virtual event. A one-byte
  pattern is a key press of that character, so `"a"` is `<Key-a>` and
  `"<"` alone is valid (the `<` key), as in Tk. `<1>` is a button press;
  `<Key-1>` is the digit. `<Motion-1>` means `<B1-Motion>`.
- Modifier names are matched case-insensitively (`parseModifier`):
  `Shift`, `Control`/`Ctrl`, `Alt`/`Mod1`/`M1`/`Option` (all `ModAlt`),
  `Meta`/`Mod4`/`Super` (`ModMeta`), `Mod2`/`M2`, `Mod3`/`M3`/`Num`,
  `Mod5`/`M5`, `Lock`, `B1`..`B5`/`Button1`..`Button5`, `Double`,
  `Triple`, `Any` (ignored). `Command` is per platform
  (`commandModifier`, tkBind.c modArray): `Mod2` on macOS (the Cocoa
  backend reports Command as Mod2 and Option as Mod1), `ModControl` on
  Windows, `ModMeta` elsewhere. A new modifier needs a `Modifier` bit, a
  case in `parseModifier`, its state bit in `eventModifiers` and its name
  in `patternString`, so `Sequence.String` round-trips.
- Key details resolve through `lookupKeySym`: a single printable
  character is its own keysym (case matters: `<Key-A>` is Shift+a); a
  longer name is looked up case-insensitively in `keysymNames`.
  `keySymName` is the inverse used by `String` (shortest name wins, else
  `0x..`). Add a key name to `keysymNames` only; the inverse map is built
  from it.
- `Sequence.String()` is the binding's identity: `BindingTable` replaces
  and removes by it, so two spellings of one pattern must print the same.

## Tables and the tag chain

- `BindingTable` maps a tag (widget path, class, toplevel path, `all`, or
  any string) to its bindings. Lists are copy-on-write so a dispatch keeps
  a stable snapshot while handlers rebind. Re-binding a sequence on a tag
  replaces the handler and keeps its definition number (tkBind.c
  FindSequence); `UnregisterWindow` drops the path's bindings, as
  Tk_DeleteAllBindings.
- Default chain (`buildTagChain`): path, class, toplevel path, `all`; the
  toplevel tag is left out for the toplevel itself. It is built lazily
  from `Window.Class` on the first event (`tagInfoFor`), so widgets need
  not call `RegisterWindow`. `SetBindTags` without the class tag disables
  the class behaviour.
- Per tag, at most one binding fires: the best match by `compare.go`
  (more specialised pattern, then more modifiers, then defined last). A
  completed multi-event sequence and a matching virtual event compete with
  the single-event match; a virtual event wins only when its physical
  definition is strictly more specific. A handler returning `true` is
  Tk's `break`; the walk also stops once a handler destroyed the window.
- Multi-event sequences wait per window in `Engine.prom`
  (`advancePromoted`); `promotionSurvives` ports Tk_BindEvent's expiry
  rules (modifier keys pass through, another key or button drops it).

## Engine and Dispatcher

`Engine.Install` sets the engine as the `event.Dispatcher`'s chain
(`Dispatcher.SetChain`); it is not a global handler. For key, button,
motion, enter and leave events the dispatcher runs the chain, which runs
the window's own `Dispatcher.Bind` handlers (the widget's class
behaviour) via `DispatchWindow` when it reaches the class tag, so a path
binding runs first and can break them. For every other type (including
`MouseWheel`, Expose, Configure, Focus) the window's handlers run first,
then the chain, like Tk_CreateEventHandler before Tk_BindEvent; global
handlers run last. `widget.AppContext` exposes only the dispatcher, not
the engine: widgets bind through `Dispatcher().Bind` (`widget/AGENTS.md`),
applications through `App.Bind()`.

## Virtual events, clicks

- `bind/virtual.go` installs a reduced, platform-independent set:
  `<<Copy>>` `<Control-c>`, `<<Cut>>` `<Control-x>`, `<<Paste>>`
  `<Control-v>`, `<<SelectAll>>` `<Control-a>`, `<<Undo>>` `<Control-z>`,
  `<<Redo>>` `<Control-y>` and `<Control-Shift-z>`. Tk's lists in
  `tk/library/tk.tcl` are per platform and longer; port from there when
  adding one. Only single-pattern definitions take part in matching
  (`indexVirtuals`); matches are tried in definition order so the choice
  never depends on map order.
- `GenerateEvent` runs only the tag bindings for `<<Name>>` with a
  `VirtualType` event; it does not reach `Dispatcher.Bind` handlers.
- `updateClickState` counts presses and releases separately (Tk keeps one
  curEvent per type): same window and button, within 500 ms and 5 pixels
  of the previous one (NEARBY_MS, NEARBY_PIXELS). Count 2 adds
  `ModDouble`, 3 `ModTriple`, more adds neither. A `Double` pattern needs
  the click modifier; a plain pattern still matches the second click.

## Tests

- Table-driven parser tests: `bind/pattern_test.go`; chains and break:
  `bind/chain_test.go`; dispatch and ordering: `bind/dispatch_test.go`,
  `bind/sequence_test.go`, `bind/table_test.go`; all display-free.
- Fuzz targets in `bind/pattern_fuzz_test.go`: `FuzzParse`,
  `FuzzParseSequenceString`,
  `FuzzPatternMatches`, `FuzzPatternSpecificity`
  (`go test ./bind/ -run '^$' -fuzz '^FuzzParse$' -fuzztime 30s`). Add a
  seed for every new syntax.
- Benchmarks: `bind/dispatch_bench_test.go` (engine key dispatch, with
  and without pending sequences); the dispatcher's own are in
  `event/dispatch_bench_test.go`.

## Adding a pattern feature or tag

- Event type: a case in `parseEventType` and `eventTypeName`, and in
  `event.isBindingEvent` if the widget's class handlers must run inside
  the chain. Detail syntax: `setDetail`. Ordering: `compare.go`, after
  the matching tkBind.c code.
- A new kind of tag needs no table change: bind the string and put it in
  the window's chain with `SetBindTags`. Changing the default chain
  changes every widget: edit `buildTagChain` only to follow Tk.
