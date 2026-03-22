# Takigo Code Review — Second Pass (2026-03-22)

## Critical — Destroy Cleanup Gaps

### ~~1. Combobox missing Destroy() override~~ ✅ Fixed (e80a3dc)
**File:** `ttk/combobox.go`
Added `Destroy()` that calls `closeDropdown()` then `TtkWidget.Destroy()`.

### ~~2. TTK Checkbutton Variable leak~~ ✅ Fixed (e80a3dc)
**File:** `ttk/checkbutton.go`
Added `Destroy()` that unsubscribes Variable OnChange callback.

### ~~3. TTK Toggleswitch Variable leak~~ ✅ Fixed (e80a3dc)
**File:** `ttk/toggleswitch.go`
Added `Destroy()` that unsubscribes Variable OnChange callback.

### ~~4. Label TextVariable leak~~ ✅ Fixed (e80a3dc)
**File:** `widget/label/label.go`
`Destroy()` now calls `unsub()`.

### ~~5. Systray dispatcher leak~~ ✅ Fixed (e80a3dc)
**File:** `systray/systray.go`
`Destroy()` now calls `Unbind(t.win)` before `DestroyWindow`.

## Important — Integer Underflow

### ~~Layout underflow~~ ✅ Fixed (e80a3dc)
All `availW`/`availH` calculations now use `max(0, ...)`:
- `widget/label/label.go`
- `widget/button/button.go`
- `widget/checkbutton/checkbutton.go`
- `widget/message/message.go`
- `widget/text/text.go`

## Medium — Platform Layer

### ~~6. PropertyNotifyEvent not dispatched~~ ✅ Fixed (f4abd6d)
**File:** `event/event.go`
Added `case platform.PropertyNotifyEvent` to `FromRawEventIM`; added `Atom` field to `Event`.

### ~~7. PropertyEvent.Atom not populated~~ ✅ Fixed (f4abd6d)
**File:** `platform/x11/event.go`, `internal/xlib/event.go`
Added `xlib.ParsePropertyEvent()` using C accessors; X11 parser now populates `Atom`.

### ~~8. Windows missing compile-time interface check~~ ✅ Fixed (f4abd6d)
**File:** `platform/windows/display.go`
Added `var _ platform.DisplayServer = (*WindowsDisplay)(nil)`.

## Minor — Test Coverage Gaps

### ~~Pure-Go unit tests~~ ✅ Added
- `wm/wm_test.go` — `ParseGeometry()` table-driven tests (14 cases)
- `image/image_test.go` — Registry operations (register, get, overwrite, unregister, destroyAll)
- `grab/grab_test.go` — `State()` tree walk, `ShouldRedirect`, `RedirectTarget` (8 tests)
- `focus/focus_test.go` — `flattenTree`, `findToplevel`, `nextFocusable` forward/backward/wrap (9 tests)

### Remaining (X11-dependent, need display)
`geometry/place/`, `selection/`, `cursor/`

## Minor — Consistency

Option function naming: `Opt` suffix is used for enum/complex types (`FontOpt`, `ImageOpt`,
`JustifyOpt`, `ReliefOpt`, `CompoundOpt`, `WrapModeOpt`), while simple value setters omit it
(`Text`, `Background`, `Width`, `PadX`). Pattern is semi-intentional and consistent within
categories. Not worth a bulk rename — ~100+ call sites across codebase and demos.
