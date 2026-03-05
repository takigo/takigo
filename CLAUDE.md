# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Takigo** is a pure Go port of the Tk 9.1 GUI toolkit, targeting X11/Linux via cgo Xlib bindings. The original C source lives under `tk/` (Tk 9.1a1) and `tcl/` (Tcl 9.0) for reference. The Go code reimplements Tk's functionality using a functional-options API instead of Tcl command strings.

## Build & Test

```bash
go build ./...                    # build all packages
go build ./demos/widget_demo      # build the main demo launcher
go build ./demos/<name>           # build individual demo

go test ./color/ ./option/ ./ttk/ ./bind/ ./widget/text/ ./event/ -v   # unit tests (pure Go + cgo)
go test . -v                      # integration tests (needs DISPLAY or xvfb-run)
```

Tests requiring X11 use `internal/testutil.RequireDisplay(t)` and `NewTestApp(t)` helpers.

## Go Package Structure → C Source Mapping

### Root Package

| Go | C equivalent | Description |
|---|---|---|
| `tk.go` | `tk/generic/tkWindow.c`, `tk/generic/tkMain.c` | App struct, NewApp, MainLoop, Quit; display connection, root window, event loop |

### Core Infrastructure

| Go package | C equivalent | Description |
|---|---|---|
| `internal/xlib/` | `tk/unix/tkUnix*.c`, Xlib headers | Low-level cgo bindings: display, window, event, draw, GC, atom, keysym, pixmap, selection, WM |
| `platform/` | — | Platform-neutral types/interfaces for display, events, fonts, keysym-to-Unicode |
| `platform/x11/` | `tk/unix/tkUnixEvent.c`, `tkUnixKey.c` | X11 platform implementation: event conversion, display init, font backend |
| `event/` | `tk/generic/tkEvent.c` | Event types/masks, Dispatcher (per-window handlers), Loop (select-based main loop) |
| `window/` | `tk/generic/tkWindow.c` | Window/Display structs, parent-child hierarchy, create/destroy, geometry data |
| `option/` | `tk/generic/tkOption.c`, `tk/generic/tkGet.c` | Functional option type, Relief/Anchor/Justify/Wrap enums |
| `config/` | `tk/generic/tkConfig.c`, `tk/generic/tkOldConfig.c` | Runtime cget/configure table with specs, getters, setters |
| `color/` | `tk/generic/tkColor.c`, `tk/unix/tkUnixColor.c` | Color parsing (#hex, named), caching, X11 allocation |
| `gc/` | `tk/generic/tkGC.c` | GC value-based caching pool (ref-counted) |
| `draw/` | `tk/generic/tk3d.c`, `tk/unix/tkUnix3d.c` | 3D relief borders (raised/sunken/groove/ridge), drawing primitives |
| `font/` | `tk/generic/tkFont.c`, `tk/unix/tkUnixRFont.c`, `tk/unix/tkUnixFont.c` | Font interface, Xft/fontconfig cgo bindings, named fonts registry |
| `cursor/` | `tk/generic/tkCursor.c`, `tk/unix/tkUnixCursor.c` | Cursor creation/management |
| `screenunit/` | `tk/generic/tkGet.c` (`Tk_GetPixels`) | Tk-style distance→pixel conversion (px, cm, mm, in, pt) |

### Geometry Managers

| Go package | C equivalent | Description |
|---|---|---|
| `geometry/` | `tk/generic/tkGeometry.c` | Manager interface (defined in `window/` to avoid circular imports) |
| `geometry/pack/` | `tk/generic/tkPack.c` | Pack geometry manager (cavity-based, side/fill/expand) |
| `geometry/grid/` | `tk/generic/tkGrid.c` | Grid geometry manager (rows/cols, weight, sticky, span, uniform) |
| `geometry/place/` | `tk/generic/tkPlace.c` | Place geometry manager (absolute/relative positioning) |

### Classic Widgets

| Go package | C equivalent | Description |
|---|---|---|
| `widget/` | `tk/generic/tkCmds.c` | Base widget type, defaults, Variable[T] observable type |
| `widget/frame/` | `tk/generic/tkFrame.c` | Frame container widget |
| `widget/toplevel/` | `tk/generic/tkFrame.c` (toplevel portion) | Toplevel window widget |
| `widget/label/` | `tk/generic/tkButton.c` (label portion) | Static text/image label |
| `widget/button/` | `tk/generic/tkButton.c`, `tk/unix/tkUnixButton.c` | Push button with command callback |
| `widget/checkbutton/` | `tk/generic/tkButton.c` (checkbutton portion) | Toggle checkbox with Variable[bool] |
| `widget/radiobutton/` | `tk/generic/tkButton.c` (radiobutton portion) | Mutual-exclusion radio with Variable[string] |
| `widget/entry/` | `tk/generic/tkEntry.c` | Single-line text entry with selection, validation |
| `widget/spinbox/` | `tk/generic/tkEntry.c` (spinbox portion) | Entry + up/down increment buttons, range/values modes |
| `widget/scrollbar/` | `tk/generic/tkScrollbar.c`, `tk/unix/tkUnixScrlbr.c` | Classic scrollbar widget |
| `widget/scale/` | `tk/generic/tkScale.c`, `tk/unix/tkUnixScale.c` | Slider widget |
| `widget/listbox/` | `tk/generic/tkListbox.c` | Scrollable list with selection modes |
| `widget/panedwindow/` | `tk/generic/tkPanedWindow.c` | Resizable sash-separated panes |
| `widget/menu/` | `tk/generic/tkMenu.c`, `tk/generic/tkMenuDraw.c`, `tk/unix/tkUnixMenu.c` | Popup/tearoff menu (override-redirect + grab) |
| `widget/menubutton/` | `tk/generic/tkMenubutton.c`, `tk/unix/tkUnixMenubu.c` | Button that posts associated menu |
| `widget/labelframe/` | `tk/generic/tkFrame.c` (labelframe portion) | Frame with text label in border gap |
| `widget/text/` | `tk/generic/tkText*.c` (6 files) | Multi-line text editor: Document model, marks, tags, undo/redo, wrapping, selection, display engine |
| `canvas/` | `tk/generic/tkCanvas.c`, `tk/generic/tkCanv*.c`, `tk/generic/tkRectOval.c` | 2D drawing surface: rect/oval/line/polygon/arc/text/image/bitmap/window items, scrolling, tags, per-item events |

### TTK Themed Widgets

| Go package | C equivalent | Description |
|---|---|---|
| `ttk/` | `tk/generic/ttk/ttkWidget.c`, `ttkState.c`, `ttkTheme.c`, `ttkLayout.c`, `ttkElements.c` | Theme engine, layout engine, state system, element interface, base widget |
| `ttk/defaulttheme/` | `tk/generic/ttk/ttkDefaultTheme.c` | Default theme (blank-import registration via init()) |
| `ttk/clamtheme/` | `tk/generic/ttk/ttkClamTheme.c` | Clam theme with custom border elements |
| `ttk.Frame` | `tk/generic/ttk/ttkFrame.c` | Themed frame |
| `ttk.Label` | `tk/generic/ttk/ttkLabel.c` | Themed label |
| `ttk.Button` | `tk/generic/ttk/ttkButton.c` | Themed button |
| `ttk.Checkbutton` | `tk/generic/ttk/ttkButton.c` | Themed checkbutton |
| `ttk.Radiobutton` | `tk/generic/ttk/ttkButton.c` | Themed radiobutton |
| `ttk.Separator` | `tk/generic/ttk/ttkSeparator.c` | Themed separator |
| `ttk.Scrollbar` | `tk/generic/ttk/ttkScrollbar.c` | Themed scrollbar |
| `ttk.Notebook` | `tk/generic/ttk/ttkNotebook.c` | Tabbed container |
| `ttk.Combobox` | `tk/generic/ttk/ttkEntry.c` | Entry + dropdown list |
| `ttk.Menubutton` | `tk/generic/ttk/ttkButton.c` | Themed menubutton |
| `ttk.Progressbar` | `tk/generic/ttk/ttkProgress.c` | Determinate/indeterminate progress bar |
| `ttk.Panedwindow` | `tk/generic/ttk/ttkPanedwindow.c` | Themed paned window (wraps classic with flat sash) |
| `ttk.Treeview` | `tk/generic/ttk/ttkTreeview.c` | Hierarchical multicolumn list/tree |
| `ttk.Toggleswitch` | `tk/generic/ttk/ttkToggleswitch.c` | iOS-style toggle switch |

### Supporting Subsystems

| Go package | C equivalent | Description |
|---|---|---|
| `bind/` | `tk/generic/tkBind.c` | Binding engine: pattern parser, binding tables, tag-chain dispatch, virtual events |
| `focus/` | `tk/generic/tkFocus.c`, `tk/unix/tkUnixFocus.c` | Focus management |
| `grab/` | `tk/generic/tkGrab.c` | Grab management (local/global) |
| `selection/` | `tk/generic/tkSelect.c`, `tk/unix/tkUnixSelect.c` | X11 selection (clipboard) |
| `image/` | `tk/generic/tkImage.c`, `tk/generic/tkImgPhoto.c` | Image interface, Photo (RGBA + pixmap cache), registry |
| `dialog/` | `tk/library/msgbox.tcl`, `tk/library/clrpick.tcl`, `tk/library/fontchooser.tcl`, `tk/library/tkfbox.tcl` | Modal dialogs: message box, color/font/file choosers (Go reimpl of Tcl scripts) |
| `busy/` | `tk/generic/tkBusy.c` | Busy overlay (InputOnly window blocks input) |
| `systray/` | `tk/unix/tkUnixSysTray.c` | X11 system tray (_NET_SYSTEM_TRAY) |
| `wm/` | `tk/unix/tkUnixWm.c` | Window manager interaction (title, geometry, state, icons, hints) |

### Demos & Commands

| Path | C/Tcl equivalent | Description |
|---|---|---|
| `demos/widget_demo/` | `tk/library/demos/widget` | Main demo launcher |
| `demos/<name>/` | `tk/library/demos/<name>.tcl` | Individual demo programs (~60 demos) |
| `demos/demohelper/` | — | Shared demo infrastructure (window setup, description labels) |
| `cmd/demo/` | — | Phase 9 demo (TTK + classic widgets) |
| `cmd/canvas_demo/` | — | Phase 10 demo (canvas items) |
| `cmd/text_demo/` | — | Phase 11 demo (text widget) |
| `cmd/bind_demo/` | — | Phase 12 demo (binding engine) |
| `cmd/dialog_demo/` | — | Phase 13 demo (dialogs, spinbox, busy, tray) |

### Internal & Test Support

| Path | Description |
|---|---|
| `internal/xlib/` | Low-level cgo Xlib bindings with C helper functions |
| `internal/testutil/` | Test helpers: RequireDisplay(t), NewTestApp(t) |

## Architecture Notes

- **No Tcl interpreter** — Tk's Tcl command layer is replaced by Go's type system and functional options
- **X11-only** — currently targets X11/Linux via cgo; `platform/` interface exists for future backends
- **Event loop** — dedicated goroutine reads XNextEvent, posts to channel; main goroutine runs select loop
- **Double buffering** — canvas, text, and TTK widgets render to pixmap, then CopyArea to window
- **Circular import avoidance** — interfaces defined in `window/` (GeomManager) and `widget/` (AppContext, BindEngine, TextProvider, WidgetImage) to break dependency cycles
- **TTK themes** — registered via blank-import `init()` (e.g., `_ "github.com/msorc/takigo/ttk/defaulttheme"`)
- **C files not ported** — `tkStubInit.c`/`tkStubLib.c` (stubs mechanism), `tkConsole.c`, `tkTest.c`, `tkSquare.c`, PostScript output (`tkCanvPs.c`), `tkMessage.c`, `tkOldConfig.c`
