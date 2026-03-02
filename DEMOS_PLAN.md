# Plan: Convert Tk Library Demos to Go

## Overview

Port all 66+ Tcl demo scripts from `tk/library/demos/` to Go using the takigo toolkit. Each demo becomes a standalone `demos/<name>/main.go` program. A launcher program (`demos/widget_demo/`) provides a GUI index of all demos.

---

## Approach

- Each demo → `demos/<demo_name>/main.go`
- Demos grouped into phases by widget category (matching Tk demo categories)
- Prioritize demos that exercise already-implemented features
- Skip/defer demos requiring unimplemented features (noted below)

---

## Feature Gap Analysis

### Available in takigo (can port now)
- Basic widgets: Frame, Label, Button, Entry, Scrollbar, Scale, Listbox, PanedWindow, Menu, Menubutton, Spinbox
- TTK: Frame, Label, Button, Separator
- Canvas: Rectangle, Oval, Line, Polygon, Arc, Text, Image items
- Text widget: Insert, Delete, Tags, Marks, Undo/Redo, Wrap modes
- Geometry: Pack, Grid, Place
- Dialogs: Message box, Color chooser, Font chooser, File dialog
- Images: Photo (RGBA, from file/reader)
- Bind engine: Pattern matching, virtual events, tag chains
- System tray, Busy window, Toplevel

### NOT available (must skip or stub)
- **Checkbutton widget** — not implemented (needed by: check.tcl, menu.tcl)
- **Radiobutton widget** — not implemented (needed by: radio.tcl, menu.tcl)
- **Combobox (TTK)** — not implemented (needed by: combo.tcl)
- **Treeview (TTK)** — not implemented (needed by: tree.tcl, mclist.tcl)
- **Progressbar (TTK)** — not implemented (needed by: ttkprogress.tcl)
- **Notebook (TTK)** — not implemented (needed by: ttknote.tcl)
- **Canvas embedded windows** — not implemented (needed by: twind.tcl)
- **Text embedded windows** — not implemented (needed by: twind.tcl)
- **Text peering** — not implemented (needed by: textpeer.tcl)
- **Bitmap images (XBM)** — not implemented (needed by: bitmap.tcl, icon.tcl)
- **Print subsystem** — not implemented (needed by: print.tcl)
- **Labelframe** — not implemented (needed by: labelframe.tcl)
- **Toolbar/tearoff** — not implemented (needed by: toolbar.tcl)
- **After/timer animation** — `App.After()` exists but may need testing for animation loops
- **Variable traces** — not implemented Tcl-style (needed by: ttkscale.tcl, check.tcl, radio.tcl)
- **Window icons** — not implemented (needed by: windowicons.tcl)
- **macOS-specific** — N/A on Linux (mac_styles.tcl, mac_tabs.tcl, mac_wm.tcl)

---

## Phase Structure

### Phase A: Demo Launcher (1 demo)

**`demo/widget_demo/`** — Main demo launcher with clickable list of all demos.
- Text widget with styled category headings and clickable demo names
- Each click launches a demo subprocess (`exec.Command`)
- "See Code" button opens source in text widget
- Categories match Tk's grouping

---

### Phase B: Labels, Buttons & Basic Widgets (6 demos)

| Demo | Tcl Source | Description | Notes |
|------|-----------|-------------|-------|
| `demo/demo_label/` | label.tcl | Text labels with various options | Straightforward |
| `demo/demo_button/` | button.tcl | Buttons that change background color | Uses `App.After()` for delayed reset |
| `demo/demo_puzzle/` | puzzle.tcl | 15-puzzle game with buttons | Grid layout + button commands |
| `demo/demo_icon/` | icon.tcl | Buttons with images | Needs PNG images (skip XBM) |
| `demo/demo_entry1/` | entry1.tcl | Basic entry widgets | Entry + pack layout |
| `demo/demo_entry2/` | entry2.tcl | Entries with scrollbars | Entry + Scrollbar integration |

**Skipped:** check.tcl (no Checkbutton), radio.tcl (no Radiobutton), bitmap.tcl (no XBM), labelframe.tcl (no Labelframe)

---

### Phase C: Listboxes & Scrolling (3 demos)

| Demo | Tcl Source | Description | Notes |
|------|-----------|-------------|-------|
| `demo/demo_states/` | states.tcl | US states listbox | Listbox + Scrollbar |
| `demo/demo_colors/` | colors.tcl | Color names listbox | Listbox + color display |
| `demo/demo_sayings/` | sayings.tcl | Scrollable sayings | Listbox + X/Y scrollbars |

**Skipped:** mclist.tcl (no Treeview), tree.tcl (no Treeview)

---

### Phase D: Entry & Form Widgets (3 demos)

| Demo | Tcl Source | Description | Notes |
|------|-----------|-------------|-------|
| `demo/demo_entry3/` | entry3.tcl | Validated entries & password | Entry with Show(rune) |
| `demo/demo_spin/` | spin.tcl | Spinbox widgets | Spinbox range + values |
| `demo/demo_form/` | form.tcl | Simple form layout | Multiple entries + grid |

**Skipped:** combo.tcl (no Combobox)

---

### Phase E: Text Widget (5 demos)

| Demo | Tcl Source | Description | Notes |
|------|-----------|-------------|-------|
| `demo/demo_text/` | text.tcl | Basic text editing | Text + Scrollbar + undo |
| `demo/demo_style/` | style.tcl | Text display styles | Tags: bold, italic, colors |
| `demo/demo_bind_text/` | bind.tcl | Hypertext-like tag bindings | Tag bindings for hover/click |
| `demo/demo_search/` | search.tcl | Text search tool | Entry + Text + search logic |
| `demo/demo_ctext/` | ctext.tcl | Canvas text editing | Canvas text item + key bindings |

**Skipped:** twind.tcl (embedded windows), textpeer.tcl (text peering)

---

### Phase F: Canvas Demos (6 demos)

| Demo | Tcl Source | Description | Notes |
|------|-----------|-------------|-------|
| `demo/demo_items/` | items.tcl | All canvas item types | Comprehensive canvas showcase |
| `demo/demo_plot/` | plot.tcl | 2D plot with draggable points | Canvas + mouse bindings |
| `demo/demo_arrow/` | arrow.tcl | Editable arrowheads | Canvas lines + arrow shapes |
| `demo/demo_ruler/` | ruler.tcl | Ruler with tab stops | Canvas polygons + drag |
| `demo/demo_cscroll/` | cscroll.tcl | Scrollable canvas grid | Canvas + 2D scrollbars |
| `demo/demo_floor/` | floor.tcl | Building floorplan | Large canvas drawing (complex) |

---

### Phase G: Scales & Sliders (2 demos)

| Demo | Tcl Source | Description | Notes |
|------|-----------|-------------|-------|
| `demo/demo_hscale/` | hscale.tcl | Horizontal scale → canvas arrow | Scale + Canvas drawing |
| `demo/demo_vscale/` | vscale.tcl | Vertical scale → canvas arrow | Scale + Canvas drawing |

---

### Phase H: Paned Windows (2 demos)

| Demo | Tcl Source | Description | Notes |
|------|-----------|-------------|-------|
| `demo/demo_paned1/` | paned1.tcl | Horizontal panes with colors | PanedWindow horizontal |
| `demo/demo_paned2/` | paned2.tcl | Vertical panes with content | PanedWindow vertical + text/listbox |

---

### Phase I: Menus (2 demos)

| Demo | Tcl Source | Description | Notes |
|------|-----------|-------------|-------|
| `demo/demo_menu/` | menu.tcl | Menu bar with cascades | Menu + keyboard shortcuts |
| `demo/demo_menubu/` | menubu.tcl | Menu buttons (4 directions) | Menubutton directional |

**Skipped:** ttkmenu.tcl (TTK menubutton not implemented)

---

### Phase J: Dialogs (4 demos)

| Demo | Tcl Source | Description | Notes |
|------|-----------|-------------|-------|
| `demo/demo_msgbox/` | msgbox.tcl | Message box varieties | dialog.ShowMessage |
| `demo/demo_filebox/` | filebox.tcl | File open/save dialogs | dialog.OpenFile/SaveFile |
| `demo/demo_clrpick/` | clrpick.tcl | Color picker | dialog.ChooseColor |
| `demo/demo_fontchoose/` | fontchoose.tcl | Font chooser | dialog.ChooseFont |

---

### Phase K: Images (2 demos)

| Demo | Tcl Source | Description | Notes |
|------|-----------|-------------|-------|
| `demo/demo_image1/` | image1.tcl | Photo images in labels | Load PNG/GIF files |
| `demo/demo_image2/` | image2.tcl | Image directory viewer | Image loading + selection |

---

### Phase L: TTK Themed Widgets (3 demos)

| Demo | Tcl Source | Description | Notes |
|------|-----------|-------------|-------|
| `demo/demo_ttkbut/` | ttkbut.tcl | TTK buttons/labels | TTK Button/Label/Separator |
| `demo/demo_ttkpane/` | ttkpane.tcl | TTK paned windows | TTK + PanedWindow |
| `demo/demo_ttkscale/` | ttkscale.tcl | TTK scale + label | Scale feedback display |

**Skipped:** ttkprogress.tcl (no Progressbar), ttknote.tcl (no Notebook), ttkspin.tcl (same as spin.tcl)

---

### Phase M: Special Features (3 demos)

| Demo | Tcl Source | Description | Notes |
|------|-----------|-------------|-------|
| `demo/demo_systray/` | systray.tcl | System tray icon | systray package |
| `demo/demo_unicodeout/` | unicodeout.tcl | Unicode text display | Label/Text with Unicode |
| `demo/demo_dialog_modal/` | dialog1.tcl + dialog2.tcl | Modal dialog examples | grab local/global |

---

### Phase N: Animation & Simulation (4 demos)

| Demo | Tcl Source | Description | Notes |
|------|-----------|-------------|-------|
| `demo/demo_anilabel/` | anilabel.tcl | Animated scrolling label | App.After() timer loop |
| `demo/demo_aniwave/` | aniwave.tcl | Animated canvas waveform | Canvas line coords + timer |
| `demo/demo_pendulum/` | pendulum.tcl | Pendulum physics simulation | Canvas + math + timer |
| `demo/demo_knightstour/` | knightstour.tcl | Knight's tour visualization | Canvas + algorithm + timer |

---

### Phase O: Complex Showcase (1 demo)

| Demo | Tcl Source | Description | Notes |
|------|-----------|-------------|-------|
| `demo/demo_goldberg/` | goldberg.tcl | Rube Goldberg machine | Very complex canvas animation |

---

## Summary

| Phase | Count | Description |
|-------|-------|-------------|
| A | 1 | Demo launcher |
| B | 6 | Labels, buttons, entries |
| C | 3 | Listboxes |
| D | 3 | Entry/form/spinbox |
| E | 5 | Text widget |
| F | 6 | Canvas |
| G | 2 | Scales |
| H | 2 | Paned windows |
| I | 2 | Menus |
| J | 4 | Dialogs |
| K | 2 | Images |
| L | 3 | TTK widgets |
| M | 3 | Special features |
| N | 4 | Animation |
| O | 1 | Goldberg showcase |
| **Total** | **47** | |

**Skipped demos (19):** check.tcl, radio.tcl, bitmap.tcl, labelframe.tcl, combo.tcl, mclist.tcl, tree.tcl, twind.tcl, textpeer.tcl, ttkprogress.tcl, ttknote.tcl, ttkspin.tcl, ttkmenu.tcl, toolbar.tcl, print.tcl, windowicons.tcl, mac_styles.tcl, mac_tabs.tcl, mac_wm.tcl

---

## Implementation Notes

### Naming Convention
- Demo directories: `demo/demo_<name>/main.go`

### Common Patterns
```go
// Standard demo skeleton
func main() {
    app, err := takigo.NewApp(takigo.Title("Demo Name"), takigo.Size(400, 300))
    if err != nil { log.Fatal(err) }
    defer app.Destroy()

    root := app.Root()
    bgColor, _ := app.ColorCache().Get("#d9d9d9")
    root.BackgroundPixel = bgColor.Pixel

    // Create widgets...
    // Pack/Grid/Place...
    // Bind events...

    app.MainLoop()
}
```

### Dismiss Button
Each demo should include a "Dismiss" button that calls `app.Quit()`.

### Assets
Image demos should reference `tk/library/demos/images/` for test images (PNG, GIF, PPM).

### Execution Order
Start with Phase B (simple widgets) to establish patterns, then proceed alphabetically. Phase A (launcher) can be built last once all demos exist.
