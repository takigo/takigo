# Demo Alignment Plan: Go Demos → Tcl Originals

This plan documents all changes needed to align each Go demo in `demos/` with its
Tcl original in `tk/library/demos/`. Changes are organized per-demo.

---

## Systematic Issues (Apply to ALL demos)

These patterns recur across nearly every demo and should be addressed globally:

### S1. TTK Scrollbar
**Issue:** Almost every demo uses classic `scrollbar.New()` where Tcl uses `ttk::scrollbar`.
**Fix:** Replace classic scrollbar with TTK scrollbar in all demos that have one.
**Affected demos:** arrow, bind, colors, cscroll, ctext, items, knightstour, plot, sayings, search, text, textpeer, toolbar, tree, ttknote, ttkpane, twind, vscale, canvas demos.

### S2. Padding Unit Conversion ✅ RESOLVED
**Issue:** Tcl uses points (`p`), millimeters (`m`), centimeters (`c`), inches (`i`). Go uses pixels only.
**Fix:** Added `screenunit` package with `Px()` function that accepts Tk-style distance strings (`"3p"`, `"2m"`, `"1c"`, `"0.5i"`) and converts to pixels using actual screen DPI. All padding option functions (`PadX`, `PadY`, `IPadX`, `IPadY`) in pack, grid, label, button, checkbutton, and radiobutton now accept `any` type (int, float64, or string with unit suffix). TTK `PaddingFromAny()` helper added. Demos can now use `pack.PadX("3p")` to match Tcl originals.

### S3. Description Labels
**Issue:** Tcl creates explicit `ttk::label` widgets with `-font $font -wraplength -justify left`. Go delegates to `demohelper.Setup()`.
**Fix:** This is acceptable as an architectural difference — demohelper handles this. No change needed unless specific demos have wrong description text.

### S4. Window Sizing
**Issue:** Go hardcodes pixel dimensions; Tcl uses `positionWindow` helper.
**Fix:** Keep Go's explicit sizing but ensure dimensions produce similar visual results. No action unless sizes are wildly wrong.

### S5. See Code / Dismiss Buttons
**Issue:** Tcl has `addSeeDismiss` buttons; Go uses demohelper.
**Fix:** Acceptable difference — demohelper handles this. No change needed.

---

## Per-Demo Changes

### 1. `anilabel`
- [ ] **Layout**: Match Tcl's padding values (convert from points to pixels)
- [ ] **Content**: Verify animation timing matches Tcl original

### 2. `aniwave`
- [ ] **Layout**: Match Tcl's padding values
- [ ] **Content**: Verify wave animation parameters match

### 3. `arrow`
- **Layout**: Uses pack in both — OK
- [ ] **Missing interactivity**: Tcl has interactive canvas bindings to drag arrow endpoints. Go is static display only. Add mouse drag bindings for arrow manipulation.
- [ ] **Missing scale widgets**: Tcl has 3 scales (line width, arrow shape A/B/C) that dynamically update the canvas arrow. Add these.
- [ ] **Canvas size**: Tcl uses `5i x 5i` (480x480 px); verify Go matches.
- [ ] **Scrollbar**: Replace classic with TTK scrollbar (if present).

### 4. `bind`
- [ ] **Layout**: Tcl uses pack — verify Go matches
- [ ] **Missing features**: Add read-only text widget state, header text with font styling, raised relief on hover over items
- [ ] **Text content**: Match Tcl's directory listing format and content
- [ ] **Scrollbar**: Replace classic with TTK

### 5. `bitmap`
- [ ] **Layout**: Tcl uses grid for bitmap display — verify Go uses same
- [ ] **Content**: Verify all Tk built-in bitmap names match
- [ ] **Padding**: Convert Tcl padding values to pixels

### 6. `button`
- [ ] **Widget types**: Tcl uses `ttk::button` — verify Go uses TTK buttons
- [ ] **Layout**: Match Tcl's pack padding values
- [ ] **Content**: Match button text and commands

### 7. `check`
- [ ] **Missing tri-state**: Tcl checkbuttons support tri-state (alternate); Go only has on/off. Document as known limitation.
- [ ] **Widget types**: Tcl uses `ttk::checkbutton` — verify Go matches
- [ ] **Layout**: Match padding values
- [ ] **Variable linkage**: Verify Variable[bool] linkage matches Tcl's -variable behavior

### 8. `clrpick`
- [ ] **Layout**: Match Tcl's padding
- [ ] **Button text**: Match Tcl button labels exactly
- [ ] **Color display**: Verify color result display matches

### 9. `colors`
- [ ] **Extra colors**: Go has additional colors not in Tcl original. Remove extras to match Tcl's color list exactly.
- [ ] **Layout**: Tcl uses grid — verify Go uses same layout manager
- [ ] **Scrollbar**: Replace classic with TTK

### 10. `combo`
- [ ] **Widget types**: Verify TTK combobox usage matches
- [ ] **Layout**: Match padding values
- [ ] **Content**: Match combobox values list

### 11. `cscroll`
- [ ] **Layout**: Tcl uses grid for scrollbar layout — **change Go from pack to grid** for the scrollbar arrangement
- [ ] **Scrollbar**: Replace classic with TTK scrollbar
- [ ] **Canvas**: Verify canvas scroll region and item creation match

### 12. `ctext`
- [ ] **MAJOR**: Go demo is completely non-interactive (static text display). Tcl demo has interactive canvas text with editing capabilities.
- [ ] Add canvas text item creation matching Tcl
- [ ] Add text editing bindings (click to position cursor, type to insert, backspace/delete)
- [ ] Add selection bindings
- [ ] **Scrollbar**: Replace classic with TTK

### 13. `dialog1`
- [ ] **Layout**: Match padding values
- [ ] **Button text**: Match Tcl dialog button labels
- [ ] **Dialog behavior**: Verify modal behavior matches

### 14. `dialog2`
- [ ] **Layout**: Match padding values
- [ ] **Button text**: Match Tcl dialog button labels
- [ ] **Dialog behavior**: Verify matches Tcl (different dialog type from dialog1)

### 15. `entry1`
- [ ] **Widget types**: Tcl uses `ttk::entry` — verify Go uses TTK entry or classic entry as appropriate
- [ ] **Layout**: Match padding
- [ ] **Content**: Match sample entry text

### 16. `entry2`
- [ ] **Layout**: Match padding values
- [ ] **Widget types**: Match Tcl widget types
- [ ] **Validation**: Verify entry validation behavior if present

### 17. `entry3`
- [ ] **Missing validation**: Tcl demo showcases entry validation (-validate, -validatecommand). Go demo likely lacks validation. Add validation demonstrations.
- [ ] **Layout**: Match Tcl layout
- [ ] **Widget types**: Match Tcl entry widget type (ttk::entry vs classic)

### 18. `filebox`
- [ ] **Layout**: Match padding
- [ ] **Button commands**: Verify file dialog invocation matches
- [ ] **Types list**: Match file type filter list from Tcl

### 19. `floor`
- [ ] **MAJOR**: Go demo is significantly simplified. Tcl has multi-floor building with interactive room highlighting, tooltips, and color changes on hover.
- [ ] Add all floor canvas items matching Tcl's complex building layout
- [ ] Add room hover highlighting (enter/leave bindings)
- [ ] Add room labels and tooltips
- [ ] Add multi-floor support (Tcl has floor1/floor2/floor3 tabs or buttons)

### 20. `fontchoose`
- [ ] **Layout**: Tcl uses grid — **change Go from pack to grid**
- [ ] **Widget types**: Match Tcl widget types (ttk widgets where used)
- [ ] **Font preview**: Verify font preview text widget matches

### 21. `form`
- [ ] **Layout**: Tcl uses grid — **change Go from pack to grid**
- [ ] **Entry fields**: Match the form field labels and grid arrangement
- [ ] **Padding**: Convert Tcl grid padding to pixels

### 22. `goldberg`
- [ ] **MAJOR (~95% missing)**: Tcl demo is a complex Rube Goldberg machine animation with multiple stages. Go demo is a bare skeleton.
- [ ] This is the most complex demo. Consider implementing incrementally or marking as "simplified version."
- [ ] Add all animation stages, canvas items, and timing logic

### 23. `hscale`
- [ ] **Layout**: Match Tcl's pack layout and padding
- [ ] **Scale config**: Match from/to/tickinterval values
- [ ] **Canvas**: Match canvas size and polygon drawing
- [ ] **DPI scaling**: Tcl applies `$tk::scalingPct` scaling to canvas items — document as known limitation

### 24. `icon`
- [ ] **MAJOR**: Go demo is completely redesigned from Tcl original. Tcl shows icon bitmaps in a grid; Go shows something different.
- [ ] Rewrite to match Tcl's icon bitmap grid display
- [ ] Match Tcl's bitmap names and layout

### 25. `image1`
- [ ] **Layout**: Match Tcl padding and widget arrangement
- [ ] **Image loading**: Verify image file paths match
- [ ] **Widget types**: Match (label for image display)

### 26. `image2`
- [ ] **Layout**: Match Tcl layout (grid for image grid display)
- [ ] **Image list**: Match Tcl's image file listing behavior
- [ ] **Scrollbar**: Replace classic with TTK

### 27. `items`
- [ ] **Missing sections**: Tcl demo has multiple canvas item demonstration sections. Go demo is missing some sections.
- [ ] Add all missing canvas item types/sections to match Tcl
- [ ] **Scrollbar**: Replace classic with TTK
- [ ] **Layout**: Match Tcl's canvas size and scroll region

### 28. `knightstour`
- [ ] **Board size**: Tcl uses 8x8 board; Go uses 6x6. **Change to 8x8**.
- [ ] **Missing features**: Add animation speed control, step-by-step mode, solution display
- [ ] **Canvas**: Match Tcl's canvas size and square dimensions
- [ ] **Algorithm**: Verify knight's tour algorithm matches Tcl's implementation

### 29. `label`
- [ ] **Widget types**: Verify label widget type matches Tcl (classic vs TTK)
- [ ] **Layout**: Match padding values
- [ ] **Content**: Match label text examples

### 30. `labelframe`
- [ ] **Layout**: Tcl uses grid — **change Go from pack to grid**
- [ ] **Widget types**: Verify TTK labelframe usage matches
- [ ] **Content**: Match radiobutton/checkbutton content inside frames
- [ ] **Padding**: Convert Tcl grid padding to pixels

### 31. `mclist`
- [ ] **Layout**: Match Tcl's layout
- [ ] **Treeview config**: Verify column definitions, headings, and sort behavior match
- [ ] **Data**: Match Tcl's country data list exactly
- [ ] **Scrollbar**: Replace classic with TTK

### 32. `menu`
- [ ] **Layout**: Match Tcl's menu structure
- [ ] **Menu items**: Verify all menu items, accelerators, and submenus match Tcl
- [ ] **Widget types**: Match Tcl's menu widget behavior

### 33. `menubu`
- [ ] **Layout**: Tcl uses grid — **change Go from pack to grid**
- [ ] **Widget types**: Match Tcl's menubutton widgets
- [ ] **Menu content**: Match menu items in each menubutton

### 34. `msgbox`
- [ ] **Layout**: Match Tcl padding
- [ ] **Button labels**: Match Tcl button text
- [ ] **Message box options**: Verify icon/type options match

### 35. `paned1`
- [ ] **Layout**: Match Tcl's panedwindow configuration
- [ ] **Pane sizes**: Match initial pane sizes
- [ ] **Content**: Match pane content (labels, text, etc.)

### 36. `paned2`
- [ ] **Layout**: Match Tcl's nested panedwindow configuration
- [ ] **Widget types**: Match Tcl widgets in each pane
- [ ] **Content**: Match content in each pane

### 37. `pendulum`
- [ ] **Canvas**: Match Tcl's canvas size and pendulum drawing
- [ ] **Animation**: Verify pendulum physics/timing match
- [ ] **Layout**: Match padding values

### 38. `plot`
- [ ] **Data mismatch**: Go has different data points/scale from Tcl. **Match Tcl's exact data points.**
- [ ] **Canvas size**: Match Tcl's canvas dimensions
- [ ] **Axis labels**: Match Tcl's axis label formatting
- [ ] **Interactive dragging**: Verify point dragging matches Tcl

### 39. `print`
- [ ] **Layout**: Match Tcl's widget arrangement
- [ ] **Button text**: Match Tcl button labels
- [ ] **Print functionality**: Match Tcl's print dialog behavior

### 40. `puzzle`
- [ ] **Layout**: Match Tcl's grid/canvas layout for puzzle pieces
- [ ] **Canvas items**: Match puzzle tile appearance
- [ ] **Interaction**: Verify click-to-move matches Tcl

### 41. `radio`
- [ ] **Widget types**: Verify TTK radiobutton usage matches Tcl
- [ ] **Layout**: Match padding values
- [ ] **Variable linkage**: Verify Variable[string] linkage matches

### 42. `ruler`
- [ ] **Missing interactivity**: Tcl has interactive ruler with draggable tab stops. Go is static.
- [ ] Add tab stop dragging on canvas
- [ ] Add ruler markings matching Tcl's measurements
- [ ] **Canvas**: Match canvas size and ruler drawing

### 43. `sayings`
- [ ] **Layout**: Tcl uses grid — **change Go from pack to grid** for scrollbar arrangement
- [ ] **Data mismatch**: Go has different sayings list. **Match Tcl's exact sayings list.**
- [ ] **Scrollbar**: Replace classic with TTK
- [ ] **Listbox**: Match Tcl's listbox configuration (height, selectmode, etc.)

### 44. `search`
- [ ] **Layout**: Match Tcl's layout
- [ ] **Text widget**: Match text size and configuration
- [ ] **Search functionality**: Verify search/highlight behavior matches
- [ ] **Scrollbar**: Replace classic with TTK

### 45. `spin`
- [ ] **Widget types**: Tcl uses `ttk::spinbox` — verify Go matches
- [ ] **Layout**: Match padding values
- [ ] **Validation**: Tcl has `-validate key -validatecommand` on integer spinbox. Add if missing.
- [ ] **Width**: All Tcl spinboxes have `-width 10`. Match this.

### 46. `states`
- [ ] **Layout**: Match Tcl's widget arrangement
- [ ] **Widget types**: Match Tcl's TTK state demonstration widgets
- [ ] **States**: Verify all TTK states demonstrated match

### 47. `style`
- [ ] **Layout**: Match Tcl's layout
- [ ] **Style examples**: Match Tcl's style configuration examples
- [ ] **Widget types**: Match Tcl widgets used

### 48. `systray`
- [ ] **Layout**: Match Tcl's widget arrangement
- [ ] **Icon**: Match Tcl's tray icon behavior
- [ ] **Tooltip**: Verify tooltip matches

### 49. `text`
- [ ] **Text height**: Tcl uses 30 lines — **change Go from 20 to 30**
- [ ] **Text width**: Tcl doesn't specify width — **remove Go's explicit Width(60)**
- [ ] **Wrap mode**: Tcl uses default (no explicit wrap) — **remove Go's explicit WrapWord**
- [ ] **Missing font chooser**: Tcl has "Font Chooser" toggle button with full integration. Add font chooser button.
- [ ] **Scrollbar**: Replace classic with TTK
- [ ] **Text content**: Match Tcl's comprehensive numbered list (7 items) covering all editing features
- [ ] **Focus**: Add explicit `focus` on text widget
- [ ] **setgrid**: Tcl uses `-setgrid 1` — document as known limitation

### 50. `textpeer`
- [ ] **MAJOR**: Go demo simulates peering with copy buttons; Tcl uses true text peering (`peer create`).
- [ ] **Layout**: Tcl uses grid — **change Go from pack to grid**
- [ ] **Scrollbar**: Replace classic with TTK
- [ ] **Text height**: Tcl uses 10 lines — **change Go from 20 to 10**
- [ ] **Text width**: Tcl uses default — **remove Go's explicit Width(30)**
- [ ] **Buttons**: Change from "Copy -->" / "<-- Copy" to "Make Peer" / "Delete Peer" (requires text peering support)
- [ ] If text peering not implementable, keep simplified version but fix layout to grid and sizes to match

### 51. `toolbar`
- [ ] **Layout**: Tcl uses grid — **change Go from pack to grid** for toolbar and top-level layout
- [ ] **Missing tearoff**: Tcl has tearoff grip mechanism with drag-to-remove. Complex feature — document as known limitation or implement.
- [ ] **Missing Toolbutton style**: Tcl applies `-style Toolbutton` to buttons. Add this style.
- [ ] **Checkbutton**: Go uses regular ttk.Button with state toggle. **Change to ttk::checkbutton**.
- [ ] **Font change**: Tcl combobox selection changes text widget font. Add this functionality.
- [ ] **Text scrollbar**: Tcl has no scrollbar on text widget. **Remove Go's extra scrollbar.**
- [ ] **Padding**: Tcl uses `1.5p` and `3p` — convert to pixel equivalents

### 52. `tree`
- [ ] **Layout**: Tcl uses grid — **change Go from pack to grid** for treeview + scrollbar arrangement
- [ ] **Heading text**: Change "#0" heading from "Name" to "Directory Structure"; change "size" heading from "Size" to "File Size"
- [ ] **Column width**: Change size column from 100 to 70
- [ ] **Missing horizontal scrollbar**: Tcl has both H and V scrollbars. Add horizontal scrollbar.
- [ ] **Scrollbar**: Replace classic with TTK
- [ ] **Root source**: Tcl uses file volumes; Go uses home dir — acceptable platform difference
- [ ] **Missing icons**: Tcl uses `tk fileicon` for file/directory icons — document as known limitation
- [ ] **Size formatting**: Change "KB" to "kB" and "B" to "bytes" to match Tcl

### 53. `ttkbut`
- [ ] **Missing toggleswitch**: Tcl has 4th labelframe group with toggleswitch widget + setState procedure. Add if toggleswitch is available, otherwise document as known limitation.
- [ ] **Layout grid**: Verify row spanning matches Tcl (buttons rowspan 2, toggle spans columns 1-2)
- [ ] **Missing `-uniform yes`**: Add uniform column sizing to grid configuration
- [ ] **Padding**: Button padding is 5px in Go vs 1.5p (~2px) in Tcl — **reduce to 2px**
- [ ] **Theme sorting**: Tcl sorts themes with `lsort` — add sorting to Go's ThemeNames()
- [ ] **Grid bug**: Go grids buttons twice (lines 115-120, 123-124) — **fix duplicate grid call**

### 54. `ttkmenu`
- [ ] **m4 style**: Tcl applies `-style TMenubutton.Toolbutton` to m4. **Add this style.**
- [ ] **m5 direction**: Tcl uses `-direction below`; Go defaults. **Set direction to Below.**
- [ ] **Padding**: Tcl uses `2.25p` and `1.5p` — convert to pixel equivalents (3px and 2px matches)
- [ ] **Description text**: Go text is shorter — match Tcl's longer description

### 55. `ttknote`
- [ ] **Tab frames**: Tcl uses `ttk::frame` — **change Go from classic Frame to TTK Frame** for all tab panes
- [ ] **Tab 1 layout**: Tcl uses grid with row/column weights — **change Go from pack to grid**
- [ ] **Tab 1 label**: Change to TTK label with wraplength 4i and justify left
- [ ] **Tab underlines**: Add `-underline` indices for keyboard accessibility
- [ ] **Tab padding**: Add tab padding matching Tcl's `1.5p`
- [ ] **Scrollbar**: Replace classic with TTK in Tab 3
- [ ] **Scrollbar padding**: Add asymmetric padding matching Tcl (`padx {0 1.5p} pady 1.5p`)
- [ ] **Ctrl+Tab traversal**: Document as known limitation (requires `ttk::notebook::enableTraversal`)
- [ ] **"Neat!" label**: Change to TTK label; use `-textvariable` pattern if available

### 56. `ttkpane`
- [ ] **Clocks pane MAJOR**: Tcl has live timezone clocks updating every 1000ms. Go has static city name labels.
- [ ] Add live clock display with timezone support (use Go's `time` package)
- [ ] Add separators between clock entries
- [ ] **Button command**: Tcl shows `tk_messageBox` on press. **Change Go from `fmt.Println` to dialog.**
- [ ] **Tab frames**: Use TTK frames instead of classic frames where Tcl does
- [ ] **Text content**: Match Tcl's text widget content (or leave empty as Tcl does)
- [ ] **Padding**: Match Tcl's padding values

### 57. `ttkprogress`
- [ ] **Layout**: Tcl uses grid (buttons side-by-side) — **change Go from pack to grid**
- [ ] **Buttons**: Tcl has 2 buttons (Start/Stop) controlling both bars. Go has toggle buttons per bar. **Match Tcl's 2-button approach.**
- [ ] **Button alignment**: Tcl aligns Start right (`sticky e`) and Stop left (`sticky w`). Match this.
- [ ] **Remove labels**: Tcl has no "Determinate:" / "Indeterminate:" labels above bars. **Remove them.**
- [ ] **Remove separator**: Tcl has no separator between bars. **Remove it.**
- [ ] **Description label**: Should be TTK label with font and wraplength 4i
- [ ] **Padding**: Convert Tcl's `3p` and `7.5p` to pixel equivalents

### 58. `ttkscale`
- [ ] **Scale value display**: Tcl shows value (default); Go hides it. **Remove `ShowValueOpt(false)`.**
- [ ] **Colors**: Tcl uses X11 named colors. Go uses hex codes. **Change to named colors** (or keep hex if named colors resolve correctly).
- [ ] **Frame structure**: Tcl uses nested frames with borderwidth 7.5p. **Add frame wrapper.**
- [ ] **Anchor**: Tcl label uses default (left); Go uses center. **Change to default.**
- [ ] **Padding**: Match Tcl's frame borderwidth and label padding

### 59. `ttkspin`
- [ ] **Remove labels**: Tcl has no descriptive labels above each spinbox. **Remove Go's extra labels.**
- [ ] **Width**: All Tcl spinboxes have `-width 10`. **Add width option.**
- [ ] **Validation**: Tcl s1 has `-validate key -validatecommand {string is integer %P}`. Add if validation is available.
- [ ] **Padding**: Tcl uses unified `pady 3p padx 7.5p`. **Match these values** (~4px and ~10px).
- [ ] **Pack all at once**: Tcl packs all 3 spinboxes in one command. Structure doesn't matter, but padding should match.

### 60. `twind`
- [ ] **MAJOR (~70% missing)**: Tcl demo showcases embedded windows in text (buttons, canvas, checkbutton, images). Go demo only shows text styling.
- [ ] **Text dimensions**: Tcl uses width 70, height 35. **Change Go's 55x28 to 70x35.**
- [ ] **Missing embedded windows**: Buttons inside text, canvas plot, color buttons, image embedding
- [ ] **Missing tags**: center (justify center), buttons (margins), spacing tags
- [ ] **Scrollbar**: Replace classic with TTK; add horizontal scrollbar toggle
- [ ] **Border settings**: Add `-highlightthickness 0 -borderwidth 0`
- [ ] If embedded windows not implementable, document limitations but fix dimensions and tags

### 61. `unicodeout`
- [ ] **Missing emoji sample**: Tcl conditionally includes emoji row. Add emoji sample.
- [ ] **Label width**: Tcl sample labels have `-width 30`. **Add width constraint.**
- [ ] **Grid padding**: Tcl applies padx only to language labels; Go applies to both. **Fix asymmetry.**
- [ ] **Frame pack side**: Tcl packs frame `:bottom`; Go packs `:top`. **Change to bottom.**

### 62. `vscale`
- [ ] **Canvas size**: Tcl uses `37.5p x 37.5p` (~50x50px); Go uses `60x300px`. **Fix to match Tcl.**
- [ ] **Scale tick interval**: Tcl has `-tickinterval 50`. **Add tick interval.**
- [ ] **Scale length**: Tcl uses `213p` (~284px). Set scale length.
- [ ] **Frame borders**: Tcl has `borderwidth 7.5p`. Add frame borders.
- [ ] **Canvas border**: Tcl disables border (`bd 0 -highlightthickness 0`). Match this.
- [ ] **DPI scaling**: Tcl scales canvas items with `$tk::scalingPct` — document as known limitation
- [ ] **ShowValue**: Go has `ShowValueOpt(true)` which Tcl doesn't. **Remove explicit ShowValue.**

### 63. `windowicons`
- [ ] **Missing "Set Window Icon" button**: Tcl has 4 buttons; Go has 3. **Add icon button with embedded PNG.**
- [ ] **Badge support**: Tcl has working `wm iconbadge`. Go shows placeholder dialogs. Acceptable platform limitation.
- [ ] **Icon source**: Tcl uses embedded base64 PNG; Go generates procedurally. **Use embedded image matching Tcl.**
- [ ] **DPI scaling**: Tcl applies zoom factor. Document as known limitation.
- [ ] **Layout padding**: Tcl uses `3p`; Go uses `PadX(3) PadY(2)`. Adjust to match.

### 64. `widget_demo`
- [ ] This is the main demo launcher. Verify it lists all demos and matches Tcl's widget.tcl categories.
- [ ] Match demo categories and ordering from Tcl's widget.tcl

---

## Media Files

All image files from `tk/library/demos/images/` are already present in `demos/images/`:
- earth.gif, earthmenu.png, earthris.gif
- flagdown.xbm, flagup.xbm, gray25.xbm
- letters.xbm, noletter.xbm
- ouster.png, pattern.xbm
- plowed_field.png, starry_night.png
- tcllogo.gif, Tcl.svg, teapot.ppm, Tk_feather.png

No additional media copying needed.

---

## Priority Classification

### P0 — Layout Manager Mismatches (change pack→grid or vice versa)
These demos use the wrong layout manager and need structural changes:
1. `form` — grid→pack (change to grid)
2. `fontchoose` — grid→pack (change to grid)
3. `labelframe` — grid→pack (change to grid)
4. `menubu` — grid→pack (change to grid)
5. `sayings` — grid→pack (change to grid)
6. `cscroll` — grid for scrollbar layout (change to grid)
7. `textpeer` — grid→pack (change to grid)
8. `toolbar` — grid→pack (change to grid)
9. `tree` — grid→pack (change to grid)
10. `ttkprogress` — grid→pack (change to grid)
11. `ttknote` (tab 1) — grid→pack (change to grid)

### P1 — Widget Type Mismatches
These demos use wrong widget types:
1. **All scrollbars**: classic→TTK (systematic fix S1)
2. `toolbar` checkbutton: ttk.Button→ttk::checkbutton
3. `ttknote` tab frames: classic Frame→TTK Frame
4. `ttknote` labels: classic Label→TTK Label
5. `ttkbut` missing toggleswitch (if available)

### P2 — Missing Functionality
1. `arrow` — missing interactive scale/drag controls
2. `ctext` — missing interactive canvas text editing
3. `floor` — missing multi-floor/hover interactivity
4. `goldberg` — ~95% missing Rube Goldberg animation
5. `icon` — completely redesigned, needs rewrite
6. `knightstour` — 6x6→8x8, missing features
7. `ruler` — missing interactive tab dragging
8. `twind` — missing embedded windows/images
9. `ttkpane` — missing live clocks

### P3 — Data/Content Mismatches
1. `sayings` — different sayings list
2. `plot` — different data points
3. `colors` — extra colors in Go
4. `text` — different/shorter content
5. `tree` — different heading text
6. `ttkspin` — extra labels not in Tcl

### P4 — Padding/Sizing Adjustments
All demos need padding unit conversion (systematic fix S2).
Specific size fixes documented per-demo above.

---

## Estimated Scope

- **Total demos**: 64 (excluding demohelper and images)
- **Need layout manager change**: ~11 demos
- **Need TTK scrollbar swap**: ~19 demos
- **Need major rewrite/feature additions**: ~9 demos (arrow, ctext, floor, goldberg, icon, knightstour, ruler, twind, ttkpane clocks)
- **Need minor fixes only**: ~44 demos
- **No changes needed**: 0 (all have at least padding differences)
