# Takigo Code Review — Second Pass (2026-03-22)

## Critical — Destroy Cleanup Gaps

### 1. Combobox missing Destroy() override
**File:** `ttk/combobox.go`
No Destroy() method — dropdown window, dispatcher bindings, and pixmap leak when widget destroyed.
TtkWidget.Destroy() doesn't call closeDropdown().

### 2. TTK Checkbutton Variable leak
**File:** `ttk/checkbutton.go`
No Destroy() override — Variable OnChange callback fires after destroy, accesses dead widget state.

### 3. TTK Toggleswitch Variable leak
**File:** `ttk/toggleswitch.go`
No Destroy() override — same issue as checkbutton.

### 4. Label TextVariable leak
**File:** `widget/label/label.go:488-494`
Destroy() doesn't call `unsub()` — TextVariable callback fires on dead widget.

### 5. Systray dispatcher leak
**File:** `systray/systray.go:131-137`
Destroy() calls DestroyWindow but doesn't unbind from event dispatcher.

## Important — Integer Underflow

`availW = width - 2*inset - 2*padX` can go negative in small windows:
- `widget/label/label.go:318-319`
- `widget/button/button.go:278-279`
- `widget/checkbutton/checkbutton.go:303-304`
- `widget/message/message.go:319-320`
- `widget/text/text.go:576`

Should clamp to 0 minimum before use.

## Medium — Platform Layer

### 6. PropertyNotifyEvent not dispatched
**File:** `event/event.go:169`
Event type defined but not converted — property change handlers never fire.

### 7. PropertyEvent.Atom not populated
**File:** `platform/x11/event.go:154`
Atom field missing from parsed event.

### 8. Windows missing compile-time interface check
**File:** `platform/windows/display.go`
No `var _ platform.DisplayServer = (*WindowsDisplay)(nil)` assertion.

## Minor — Test Coverage Gaps

Still untested: `geometry/place/`, `focus/`, `grab/`, `selection/`, `wm/`, `image/`, `cursor/`

## Minor — Consistency

Option function naming: some use `FontOpt()`, others `Font()` — inconsistent `Opt` suffix.
