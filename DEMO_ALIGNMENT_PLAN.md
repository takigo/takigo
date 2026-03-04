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

### S3. Description Labels ✅ RESOLVED
**Issue:** Tcl creates explicit `ttk::label` widgets with `-font $font -wraplength 4i -justify left`. Go delegates to `demohelper.Setup()`.
**Fix:** Added `WrapLength("4i")` to demohelper's description label. Also added `label.WrapLength(any)` option to classic label widget for per-demo customization (e.g., ttknote Tab 1).

### S4. Window Sizing
**Issue:** Go hardcodes pixel dimensions; Tcl uses `positionWindow` helper.
**Fix:** Keep Go's explicit sizing but ensure dimensions produce similar visual results. No action unless sizes are wildly wrong.

### S5. See Code / Dismiss Buttons
**Issue:** Tcl has `addSeeDismiss` buttons; Go uses demohelper.
**Fix:** Acceptable difference — demohelper handles this. No change needed.

---

## Per-Demo Changes

### 1. `anilabel`
- [x] **Layout**: Match Tcl's padding values — fixed frame PadX("7.5p") PadY("7.5p"), label Expand/Anchor/PadX/PadY, l3 fixedW=false
- [x] **Content**: Text animation timing 150ms matches Tcl's 150ms; GIF animation (100ms) is placeholder — animated GIF frames not implemented (known limitation)

### 2. `aniwave`
- [x] **Layout**: Match Tcl's padding values — fixed canvas 300x200, PadX("7.5p") PadY("7.5p"), button PadY("3p")
- [x] **Content**: Fixed wave coordinates for 300px width, 100px center Y

### 3. `arrow`
- **Layout**: Uses pack in both — OK
- [x] **Missing interactivity**: Rewrote as full interactive editor — 3 draggable control boxes (box1/box2/box3) change arrow parameters a/b/c/width in real time; full redraw on ButtonRelease; matches Tcl's arrowMove1/2/3 logic
- [x] **Missing scale widgets**: Not using scales (matches Tcl — Tcl uses draggable boxes, not scales); dimension annotations with arrows and parameter text at bottom
- [x] **Canvas size**: 500x350 — matches Tcl's 375p x 262.5p at 96DPI
- [x] **Scrollbar**: No scrollbar in arrow demo — N/A

### 4. `bind`
- [x] **Layout**: Removed wrapper frame, text+scrollbar packed directly into app; height changed to 24
- [x] **Missing features**: Set `text.ReadOnly(true)` after inserting content (equivalent to `-state disabled`); hover highlight uses background color (TagRelief not implemented)
- [x] **Text content**: 6 demo links matching Tcl, initial focus set via `app.After(0, SetInputFocus)`
- [x] **Scrollbar**: Replace classic with TTK

### 5. `bitmap`
- [x] **Layout**: Tcl uses pack (not grid) — Go matches: two row frames, columns packed left within each row
- [x] **Content**: All 10 bitmap names match Tcl (error/gray12/gray25/gray50/gray75/hourglass/info/question/questhead/warning)
- [x] **Padding**: Fixed column PadX(".25c") PadY(".25c"), removed label PadY(2)

### 6. `button`
- [x] **Widget types**: Tcl uses classic `button` — Go matches
- [x] **Layout**: Changed to PadY("1.5p"), removed padX, removed expand
- [x] **Content**: Changed colors to X11 named colors (PeachPuff1, LightBlue1, SeaGreen2, Yellow1)

### 7. `check`
- [ ] **Missing tri-state**: Tcl checkbutton uses onvalue="all"/offvalue="none"/tristatevalue="partial" on master button. Go only has bool — shows partial as unchecked. Known limitation.
- [x] **Widget types**: Tcl uses classic `checkbutton` (not ttk) — Go matches
- [x] **Layout**: Fixed PadY("1.5p"), PadX("12p")
- [x] **Variable linkage**: Variable[bool] works correctly for on/off; master tristate logic approximated (shows checked only when all three are on)

### 8. `clrpick`
- [x] **Layout**: Fixed — removed button PadX/PadY, changed pack PadY(5) → PadY("2m"), added Anchor(AnchorCenter)
- [x] **Button text**: "Set background color ..." and "Set foreground color ..." — matches Tcl exactly
- [ ] **Color display**: Tcl applies color recursively to all children via setColor_helper. Go applies to buttons only. Known limitation — no recursive widget color propagation.

### 9. `colors`
- [x] **Extra colors**: Added missing SlateGray1-4 and LightSteelBlue1-4 to match Tcl's list
- [x] **Layout**: Fixed — frame with BorderWidth(10) matching Tcl's `borderwidth 7.5p`; FillY not FillBoth
- [x] **Scrollbar**: Replace classic with TTK

### 10. `combo`
- [x] **Widget types**: Tcl uses `ttk::combobox` — Go matches (ttk.NewCombobox)
- [x] **Layout**: Fixed — added inner TTK frame, PadY("3p") PadX("7.5p")
- [x] **Content**: Removed status label, changed disabled text to "unchangable", fixed editCombo values, removed Return binding placeholder

### 11. `cscroll`
- [x] **Layout**: Grid for canvas+scrollbars; fixed — removed PadX/PadY from gridFrame pack
- [x] **Scrollbar**: Replace classic with TTK scrollbar
- [x] **Canvas**: Scroll region (-416,-416,1890,756) matches Tcl's {-11c -11c 50c 20c} at 37.8px/cm; rectangles in pixel coordinates match Tcl's cm-based grid

### 12. `ctext`
- [x] **Anchor selector**: 3×3 grid of colored boxes (LightSkyBlue1) click to set text anchor (SE/S/SW/E/Center/W/NE/N/NW) — matches Tcl's mkTextConfigBox
- [x] **Justification selector**: 3 boxes (SeaGreen2) for left/center/right — matches Tcl's justify section
- [x] **Config box hover**: Enter=black fill, Leave=restore original fill — matches Tcl's textEnter/Leave
- [ ] **Text editing**: Canvas text cursor/insert/delete not implemented — known limitation (no canvas text cursor API)
- [ ] **Angle selector**: Canvas text rotation not implemented — known limitation
- [x] **Scrollbar**: Not needed for current canvas size — N/A

### 13. `dialog1`
- [x] **Layout**: Tcl uses `tk_dialog` immediately; Go wraps in "Show Dialog" button — acceptable architectural difference for standalone demo
- [x] **Dialog behavior**: Uses local grab (dialog.ShowMessage with local grab)

### 14. `dialog2`
- [x] **Layout**: Same as dialog1 — acceptable architectural difference
- [x] **Dialog behavior**: Uses MsgWarning type matching Tcl's warning icon

### 15. `entry1`
- [x] **Widget types**: Tcl uses classic `entry` — Go matches
- [x] **Layout**: Changed to PadX("7.5p") PadY("3p")
- [x] **Content**: Fixed e2 long text to match Tcl; e3 placeholder matches

### 16. `entry2`
- [x] **Layout**: Frame BorderWidth(10) ≈ 7.5p, spacer Height(10) ≈ 7.5p, pack FillX — matches Tcl
- [x] **Widget types**: Tcl uses classic `entry` — Go matches; TTK scrollbar is S1 systematic issue
- [x] **Validation**: entry2.tcl has no validation (that's entry3) — N/A

### 17. `entry3`
- [ ] **Missing validation**: Tcl demo showcases entry validation (-validate, -validatecommand). Go entry widget lacks validation API — known limitation.
- [x] **Layout**: Fixed — mid frame no PadX/PadY, entry PadX("1m") PadY("1m"), grid PadX("3m") PadY("1m")
- [x] **Widget types**: Tcl uses classic entry — Go matches

### 18. `filebox`
- [x] **Layout**: Fixed — outer frame PadX("1c"), grid padding PadY("3p") PadX("3p")
- [x] **Button commands**: OpenFile/SaveFile dialogs with matching file types; open type toggles Open vs Save dialog
- [x] **Types list**: Updated file type filter list to match Tcl
- [x] **Motif checkbutton**: Added "Use Motif Style Dialog" ttk::checkbutton (X11 only, matches Tcl)

### 19. `floor`
- [ ] **MAJOR**: Go demo is significantly simplified. Tcl has multi-floor building with interactive room highlighting, tooltips, and color changes on hover.
- [ ] Add all floor canvas items matching Tcl's complex building layout — known limitation (hundreds of DEC WRL polygon coordinates, too complex to port)
- [x] Add room hover highlighting (enter/leave bindings) — implemented via transparent room overlays with BindItem per-room
- [x] Add room labels and tooltips — room labels + status label shows room name on hover
- [ ] Add multi-floor support (Tcl has floor1/floor2/floor3 tabs or buttons) — known limitation

### 20. `fontchoose`
- [x] **Layout**: Changed to grid inside inner frame — text+scrollbar row 0, button row 1 sticky-E
- [x] **Widget types**: Converted to ttk::frame (sunken relief + padding), ttk::scrollbar (S1 done), ttk::button — matches Tcl
- [x] **Font preview**: Text widget width 40, height 6, same initial text, font changes on Apply — matches Tcl

### 21. `form`
- [x] **Layout**: Already uses grid inside formFrame — labels col 0, entries col 1
- [x] **Entry fields**: Labels+entries arranged in grid rows with sticky EW
- [x] **Padding**: PadX(5) PadY(4) per row

### 22. `goldberg`
- [ ] **MAJOR (~95% missing)**: Tcl demo is a complex Rube Goldberg machine animation with multiple stages. Go demo is a bare skeleton.
- [ ] This is the most complex demo. Consider implementing incrementally or marking as "simplified version."
- [ ] Add all animation stages, canvas items, and timing logic

### 23. `hscale`
- [x] **Layout**: Added frame with BorderWidth(10) matching `borderwidth 7.5p`; removed outer padding
- [x] **Scale config**: Tcl has `-tickinterval 50` — implemented TickIntervalOpt(50) in scale widget
- [x] **Canvas**: height 50, BorderWidthOpt(0), HighlightWidthOpt(0)
- [ ] **DPI scaling**: Tcl applies `$tk::scalingPct` scaling to canvas items — not implemented, known limitation

### 24. `icon`
- [ ] **MAJOR**: Go demo is completely redesigned from Tcl original. Tcl shows icon bitmaps in a grid; Go shows something different.
- [ ] Rewrite to match Tcl's icon bitmap grid display
- [ ] Match Tcl's bitmap names and layout

### 25. `image1`
- [x] **Layout**: Fixed — 2 labels stacked vertically (Top), PadX(".5m") PadY(".5m"), Relief(Sunken) BorderWidth(1)
- [x] **Image loading**: Loads actual earth.gif and earthris.gif from demos/images/
- [x] **Widget types**: Using label for image display

### 26. `image2`
- [x] **Layout**: Rewrote — grid layout: dir labelframe (row 0 col 0-1), file listbox (row 1 col 0), image label (row 1 col 1); PadX("1m") PadY("1m")
- [x] **Image list**: Lists actual image files from demos/images/ directory; "Select Dir." button reloads
- [x] **Scrollbar**: Replace classic with TTK

### 27. `items`
- [x] **Scrollbar**: Added TTK x/y scrollbars with scroll region (0,0,660,520) — grid layout with canvas
- [x] **Images section**: Added 8th section showing ouster.png and plowed_field.png as canvas image items
- [x] **Smooth lines**: Already in Go (Smooth(true) used in section 3 lines) ✓
- [ ] **Missing bitmaps section**: Canvas bitmap items not implemented — known limitation
- [ ] **Missing window items section**: Canvas window items (embedded widgets) not implemented — known limitation
- [x] **Layout**: Tcl uses 3×3 grid with cm coordinates; Go uses fixed pixel layout with 2 new sections — functionally similar, acceptable difference

### 28. `knightstour`
- [x] **Board size**: 8x8 matching Tcl
- [x] **Missing features**: Added click-to-set-start (click canvas to choose starting square, shows green highlight); added Edgemost tiebreaker (prefer squares closer to edge on degree tie); added Repeat checkbutton for continuous random tours
- [x] **Canvas size**: 400x400 — larger than Tcl's 192p≈256px but functionally equivalent
- [x] **Algorithm**: Warnsdorff's heuristic + Edgemost tiebreaker — matches Tcl's enhanced algorithm

### 29. `label`
- [x] **Widget types**: Tcl uses classic `label` — Go matches
- [x] **Layout**: Fixed — frame PadX("7.5p") PadY("7.5p"), label PadY("1.5p")
- [x] **Content**: Three text labels (First/Second/Third with relief), Ouster image + caption

### 30. `labelframe`
- [x] **Layout**: Changed to grid — two labelframes side-by-side with ColumnConfigure weight 1
- [x] **Widget types**: Tcl uses classic `labelframe` — Go matches (widget/labelframe); inner widgets are classic checkbutton/radiobutton — Go matches
- [x] **Content**: Changed to "Value" radios 1-4 + "Options" checkbuttons matching Tcl
- [x] **Padding**: PadX("2m") PadY("2m") for grid, PadY("1.5p") for inner items

### 31. `mclist`
- [x] **Layout**: Changed to grid in container frame — tree(0,0), yscroll(0,1), xscroll(1,0); removed extra PadX/PadY; fixed title to "Multi-Column List"
- [x] **Treeview config**: 3 columns country/capital/currency; headings Country/Capital/Currency; sort by column; column widths 180/180/80 (Tcl auto-calculates from font, Go hardcodes — acceptable)
- [x] **Data**: Match Tcl's country data list exactly — fixed South Korea→South Africa, Brasilia→Brazilia
- [x] **Scrollbar**: Replace classic with TTK

### 32. `menu`
- [x] **Layout**: Status bar with PadX(2)/PadY(2) matches Tcl; menuBar frame at top; menubuttons packed left
- [ ] **Menu items**: Tcl has tearoff/cascade menus with Icon sets not in Go — acceptable X11 difference
- [x] **Widget types**: Tcl attaches menus to window menubar; Go uses frame+menubutton — acceptable X11 difference (no native menubar)

### 33. `menubu`
- [x] **Layout**: Changed to grid compass layout — Below(row 0 col 1), Right(row 1 col 0), Left(row 1 col 2), Above(row 2 col 1)
- [x] **Widget types**: Tcl uses classic `menubutton` with `-relief raised` — Go matches
- [x] **Menu content**: Menu items match Tcl's structure

### 34. `msgbox`
- [x] **Layout**: Fixed — no columns frame, pack left/right directly, separator frames with Ridge relief, PadX(".5c") PadY(".5c"), radio PadY("1.5p")
- [x] **Button labels**: "Message Box" button at bottom with PadY("1.5p")
- [x] **Message box options**: Added abortretryignore and retrycancel to types list

### 35. `paned1`
- [x] **Layout**: PadX("2m") PadY("1.5p") matches Tcl; FillBoth+Expand matches
- [x] **Pane sizes**: Two 150px initial panes with left=yellow, right=cyan
- [x] **Content**: Labels match Tcl ("This is the\nleft side" / "This is the\nright side")

### 36. `paned2`
- [x] **Layout**: Fixed — panedwindow PadX("2m") PadY("1.5p")
- [x] **Widget types**: Tcl uses ttk::scrollbar — now uses ttk.NewScrollbar
- [x] **listbox item 0 highlight**: Added ItemConfigure(idx, fg, bg) to listbox; paned2 uses it to invert colors of item 0 (matching Tcl's `itemconfigure 0 -background fg -foreground bg`)
- [x] **Content**: Listbox has same 18 Tk widget names; bottom text widget width 30 height 8 wrap none; initial text matches Tcl

### 37. `pendulum`
- [x] **Canvas**: Fixed — white background, 320x200px (240p x 150p), grey50 pivot/plate, black rod, yellow bob with black outline
- [x] **Animation**: Fixed — phase space grey0-grey90 trail levels, removed shadow, matching Tcl colors
- [x] **Layout**: Fixed — removed PadX/PadY from container and canvases

### 38. `plot`
- [x] **Data mismatch**: Updated to 7 Tcl data points, y-axis 0-250, x 0-100 with ticks every 10.
- [x] **Canvas size**: Tcl uses 337.5p x 225p ≈ 450x300px; Go uses 500x350 (data points scaled accordingly); added ReliefRaised; changed pack to FillX only (no PadX/PadY)
- [x] **Axis labels**: Helvetica 16 font, correct tick values
- [x] **Interactive dragging**: BindItem on "point" tag with ButtonPress/ButtonRelease/Motion — point dragging implemented

### 39. `print`
- [x] **Layout**: Fixed — btnFrame at bottom (Print Canvas left AnchorW, Print Text right AnchorE), PadX("3p"); content frame with canvas left + text right
- [x] **Button text**: "Print Canvas" and "Print Text" matching Tcl
- [x] **Print functionality**: `tk print` not available in Go — buttons show "not available" dialog (known limitation, documented)

### 40. `puzzle`
- [x] **Layout**: Fixed — puzzle frame 120x120px (90p), PadX("1c") PadY("1c"), removed button PadX/PadY
- [x] **Widget types**: Neither Tcl nor Go uses canvas — both use buttons with place geometry (RelX/RelY 0.25 grid); tile appearance matches
- [x] **Interaction**: Click-to-move verified matching Tcl (adjacent-only, no diagonal)

### 41. `radio`
- [x] **Widget types**: Tcl uses classic `radiobutton` (not ttk) — Go matches; tristatevalue="multi" not implemented (known limitation)
- [x] **Layout**: Rewrote — inner body frame, grid with 3 labelframes (size/color/align), tristate button, PadX(".5c") PadY(".5c"); compass grid inside alignFrame; PadY("1.5p") for inner items
- [x] **Variable linkage**: Variable[string] linkage working; tristate uses Variable[bool]

### 42. `ruler`
- [x] **Missing interactivity**: Fully interactive — drag from well to create new tabs; drag existing tabs to move; drag far out to delete (shown in gray, deleted on release); grid snap at 0.25c intervals
- [x] **Ruler markings**: cm/half-cm/quarter-cm tick marks with labels 0-11
- [x] **Canvas size**: 560×100px matching Tcl's 14.8c×2.5c at 38px/cm
- [x] **Well**: Gray rectangle + prototype tab at right edge; click+drag to create new tabs

### 43. `sayings`
- [x] **Layout**: Tcl uses grid — changed Go to grid; added PadX("1c") matching Tcl's `-padx 1c`
- [x] **Data mismatch**: Go has different sayings list — Updated to Tcl's 21 sayings in Tcl's order.
- [x] **Scrollbar**: Replace classic with TTK
- [x] **Listbox**: Width 20, Height 10 matches Tcl; default selectmode (browse); setgrid not supported — known limitation

### 44. `search`
- [x] **Layout**: Fixed — two rows (File name + Load File; Search string + Highlight), PadY("3p") PadX("7.5p"), scrollbar right then text
- [x] **Text widget**: Match text size and configuration — initial text matches Tcl, uses os.ReadFile for loading
- [x] **Search functionality**: textSearch finds all instances, tags with "search" (yellow bg), scrolls to first match
- [x] **Scrollbar**: Replace classic with TTK

### 45. `spin`
- [x] **Widget types**: Tcl uses classic `spinbox` (not ttk::spinbox) — Go matches (widget/spinbox)
- [x] **Layout**: Match padding values — screenunit.Px("7.5p") and screenunit.Px("3p")
- [ ] **Validation**: Tcl has `-validate key -validatecommand {string is integer %P}` on s1. No validation API in Go entry — known limitation.
- [x] **Width**: All Tcl spinboxes have `-width 10`. Added spinbox.WidthOpt(10) to all 3.
- [x] **Labels removed**: Removed extra labels not in Tcl original

### 46. `states`
- [x] **Layout**: Fixed — added justification labelframe with Left/Center/Right radiobuttons; lbFrame BorderWidth(19) (≈.5c); removed PadX/PadY; FillY
- [x] **Widget types**: Tcl uses ttk::scrollbar — now uses ttk.NewScrollbar
- [ ] **radiobuttons tristatevalue**: `-tristatevalue "multi"` for multi-selection state — not implemented, known limitation
- [x] **States**: All 50 US states in alphabetical order — matches Tcl exactly; justify command N/A (listbox has no justify option in Go)

### 47. `style`
- [x] **Layout**: Fixed — scrollbar right, text fills rest (no wrapper frame), matching Tcl's pack order
- [x] **Style examples**: Fixed — width 70, height 32, font "Courier 12", tags matching Tcl (bold/big/verybig/tiny/color1/color2/underline/overstrike)
- [x] **Missing tags**: Added TagJustify (left/center/right), TagOffset/TagOffsetStr (superscript/subscript), TagLMargin1/2Str, TagRMarginStr, TagSpacing1/2/3Str to text tag system; fixed overstrike rendering; updated style demo to match Tcl's full 8-section demo (Font/Color/Underline/Overstrike/Justification/Superscripts/Margins/Spacing). Also added TextWidget.EndIndex() helper.
- [ ] **Missing tags**: TagRelief/TagBorderWidth (3-D text effects) — known limitation (requires drawing borders around text runs)

### 48. `systray`
- [x] **Layout**: Fixed — labelframe "f" with Create/Modify/Destroy buttons (PadX("3p") PadY("3p")); "Display Notification" button outside frame; no status label
- [x] **Icon**: Auto-creates tray icon at startup (matching Tcl's `create` call at end); Modify toggles tooltip
- [x] **Notification**: Added "Display Notification" button using dialog.ShowMessage as fallback (tk sysnotify not implemented)
- [x] **Context menu**: Added right-click popup menu with Status/Exit commands; TrayRightClickHandler(func(x,y int)) added to systray package

### 49. `text`
- [x] **Text height**: 30 lines matching Tcl
- [x] **Text width**: No explicit width set
- [x] **Wrap mode**: Tcl uses default (no wrap) — Go keeps WrapWord for readability; acceptable difference
- [x] **Missing font chooser**: Added "Font Chooser..." TTK button (packed bottom) that opens dialog.ChooseFont modally and updates text widget font on selection
- [x] **Scrollbar**: Replace classic with TTK
- [x] **Text content**: Comprehensive numbered list (7 items) covering all editing features
- [x] **Focus**: Added `app.After(0, SetInputFocus)` for initial focus
- [ ] **setgrid**: Tcl uses `-setgrid 1` — not supported, known limitation

### 50. `textpeer`
- [ ] **MAJOR**: Go demo simulates peering with copy buttons; Tcl uses true text peering (`peer create`) — text widget peer feature not implemented, known limitation
- [x] **Layout**: Rewrote — inner frame `w` with grid, RowSpan(2) for text+scrollbar, ColumnConfigure weight 1
- [x] **Scrollbar**: Replace classic with TTK
- [x] **Text height**: Tcl uses 10 lines — changed Go to 10
- [x] **Text width**: Tcl uses default — removed explicit Width
- [x] **Buttons**: Changed to "Make Peer" / "Delete Peer" (simplified peering via copy)

### 51. `toolbar`
- [x] **Layout**: Changed to grid inside inner frame — toolbar(row 0), sep(row 1), text(row 2 expand)
- [ ] **Missing tearoff**: Tcl has tearoff grip mechanism — complex, known limitation
- [ ] **Missing Toolbutton style**: Tcl applies `-style Toolbutton` to buttons — no TTK Toolbutton style in Go yet, known limitation
- [x] **Checkbutton**: Replaced simulated button with ttk.NewCheckbutton + Variable[bool] — matches Tcl's ttk::checkbutton (without Toolbutton style indicator suppression)
- [x] **Font change**: Tcl combobox selection changes text widget font. Added font change on combobox select.
- [x] **Text scrollbar**: Tcl has no scrollbar on text widget. Removed extra scrollbar.
- [x] **Padding**: Fixed — PadX("1.5p") PadY("3p") matching Tcl's `padx 1.5p pady 3p`

### 52. `tree`
- [x] **Layout**: Tcl uses grid — **change Go from pack to grid** for treeview + scrollbar arrangement
- [x] **Heading text**: Change "#0" heading from "Name" to "Directory Structure"; change "size" heading from "Size" to "File Size"
- [x] **Column width**: Change size column from 100 to 70
- [x] **Missing horizontal scrollbar**: Added xscroll widget (X scrolling not yet fully implemented in treeview)
- [x] **Scrollbar**: Replace classic with TTK
- [x] **Root source**: Tcl uses file volumes; Go uses home dir — acceptable platform difference
- [ ] **Missing icons**: `tk fileicon` not available in Go — known limitation
- [x] **Size formatting**: Change "KB" to "kB" and "B" to "bytes" to match Tcl

### 53. `ttkbut`
- [ ] **Missing toggleswitch**: No toggleswitch widget in Go — known limitation.
- [x] **Layout grid**: Fixed — removed container PadX/PadY, fixed RowConfigure to row 0, PadX("3p") PadY("1.5p")
- [ ] **Missing `-uniform yes`**: grid.ColumnConfigure uniform not supported — known limitation
- [x] **Padding**: Fixed — buttons/checkbuttons PadY("1.5p"), radiobuttons PadX("3p") PadY("1.5p")
- [x] **Theme sorting**: Added `sort.Strings(themes)` matching Tcl's `lsort [ttk::themes]`
- [x] **Grid bug**: Fixed duplicate grid call
- [x] **TTK Radiobutton**: Converted radiobutton.New → ttk.NewRadiobutton (uses new ttk/radiobutton.go)
- [x] **TTK Checkbutton**: Already using ttk.NewCheckbutton (from previous session)

### 54. `ttkmenu`
- [ ] **m4 style**: Tcl applies `-style TMenubutton.Toolbutton` to m4 (toolbar-style button appearance). No custom style API in Go yet — known limitation.
- [x] **m5 direction**: Added `ttk.MenubuttonDirection(ttk.DirBelow)` to m5
- [x] **Padding**: Fixed grid PadX("2.25p") PadY("1.5p") for all five menubuttons
- [x] **Description text**: Go text is shorter — match Tcl's longer description

### 55. `ttknote`
- [x] **Tab frames**: Changed all tab panes from classic Frame to ttk.NewFrame
- [x] **Tab 1 layout**: Changed from pack to grid — descLabel row 0 span 2 sticky NEW, button+label row 1
- [x] **Tab 1 label**: Added WrapLength("4i") to descLabel — wraplength implemented in classic label widget; demohelper also uses WrapLength("4i") for all demo description labels
- [ ] **Tab underlines**: Tcl uses -underline 0 on tabs and button for keyboard shortcuts — not supported in Go, known limitation
- [x] **Tab padding**: Added PadX("1.5p") PadY("3p") on notebook, PadY("1.5p") on grid items
- [x] **Scrollbar**: Replace classic with TTK in Tab 3
- [x] **Scrollbar padding**: Added PadX("1.5p") PadY("1.5p") on scrollbar and text
- [ ] **Ctrl+Tab traversal**: Requires `ttk::notebook::enableTraversal` — known limitation
- [x] **"Neat!" label**: Classic label is used (textvariable implemented as manual Text update)

### 56. `ttkpane`
- [x] **Clocks pane**: Implemented live timezone clocks with `app.After(1000ms)`, updating HH:MM:SS per timezone via `time.LoadLocation`
- [x] **Separators**: Added `ttk.NewSeparator` between clock entries matching Tcl structure
- [x] **Button command**: `dialog.ShowMessage` with "Button Pressed" title and "That hurt..." message
- [ ] **TTK panedwindow**: Go uses classic panedwindow; Tcl uses `ttk::panedwindow` — no TTK panedwindow yet, known limitation
- [x] **Text content**: Text widget starts empty (matches Tcl)
- [x] **Padding**: Fixed — outer no padx/pady, button PadX("1.5p") PadY("3p"), text PadX("1.5p") PadY("1.5p")

### 57. `ttkprogress`
- [x] **Layout**: Changed from pack to grid — bars span 2 columns, Start sticky-E, Stop sticky-W
- [x] **Buttons**: Changed to 2 buttons (Start Progress / Stop Progress) controlling both bars
- [x] **Button alignment**: Start sticky E, Stop sticky W
- [x] **Remove labels**: Removed "Determinate:" / "Indeterminate:" labels
- [x] **Remove separator**: Removed separator between bars
- [x] **Description label**: Handled by demohelper — acceptable
- [x] **Padding**: Using screenunit.Px("3p") and screenunit.Px("7.5p")

### 58. `ttkscale`
- [x] **Scale value display**: Removed `ShowValueOpt(false)` so scale shows value
- [x] **Colors**: Changed to X11 named colors (Red, Orange, Yellow, Green, Blue, Violet)
- [x] **Frame structure**: Added frame wrapper with BorderWidth(10) matching `borderwidth 7.5p`
- [x] **Anchor**: Changed from AnchorCenter to AnchorW
- [x] **Order**: Label packed before scale (matching Tcl's `pack $w.frame.label $w.frame.scale`)

### 59. `ttkspin`
- [x] **Remove labels**: Removed extra labels (intLabel, floatLabel, valLabel)
- [x] **Width**: Added WidthOpt(10) to each spinbox
- [ ] **Validation**: Tcl s1 has `-validate key -validatecommand {string is integer %P}` — no validation API in Go, known limitation
- [x] **Padding**: Using screenunit.Px("7.5p") and screenunit.Px("3p")
- [x] **Description**: Updated to match Tcl's longer description

### 60. `twind`
- [ ] **MAJOR (~70% missing)**: Tcl demo showcases embedded windows in text (buttons, canvas, checkbutton, images). Go demo only shows text styling.
- [x] **Text dimensions**: Changed from 55x28 to 70x35
- [ ] **Missing embedded windows**: Buttons inside text, canvas plot, color buttons, image embedding — not implemented
- [x] **Missing tags**: Added center (justify center + spacing1/3=5m) and buttons (lmargin1/2=1c, rmargin=1c, spacing1=3m) tag configs; rewrote demo using EndIndex() helper; applied tags to appropriate content sections
- [x] **Scrollbar**: Replace classic with TTK; add horizontal scrollbar toggle
- [x] **Border settings**: Added `text.BorderWidthOpt(0)` and `tw.HighlightWidth = 0`

### 61. `unicodeout`
- [x] **Missing emoji sample**: Added emoji row (😀💩👍🇳🇱) — shown on X11+XFT (which takigo uses)
- [x] **Label width**: Added label.Width(240) to sample labels (~30 chars)
- [x] **Grid padding**: Fixed — PadX("1m") on language labels (col 0), no PadX on sample labels
- [x] **Frame pack side**: Changed from pack.Top to pack.Bottom
- [x] **Frame padding**: Changed PadX(10) PadY(5) → PadX("2m") PadY("1m") matching Tcl

### 62. `vscale`
- [x] **Canvas size**: Changed from 60 to 50 width
- [x] **Scale tick interval**: Implemented TickIntervalOpt(50) — tick marks drawn at 0/50/100/150/200/250
- [ ] **Scale length**: Tcl uses `213p` (~284px) — no TotalLength option, known limitation
- [x] **Frame borders**: Added BorderWidth(10) matching `borderwidth 7.5p`
- [x] **Canvas border**: Added BorderWidthOpt(0) and HighlightWidthOpt(0)
- [ ] **DPI scaling**: Tcl scales canvas items with `$tk::scalingPct` — not implemented, known limitation
- [x] **ShowValue**: Removed explicit `ShowValueOpt(true)`
- [x] **Scale pack**: Removed PadX(10) from scale pack — Tcl has no padding

### 63. `windowicons`
- [x] **Missing "Set Window Icon" button**: Added 4th button matching Tcl's 4 buttons
- [ ] **Badge support**: `wm iconbadge` not available on X11/Go — shows placeholder dialog, known limitation
- [ ] **Icon source**: Tcl uses embedded base64 PNG; Go generates procedurally — known limitation
- [ ] **DPI scaling**: Tcl applies zoom factor to icon — not implemented, known limitation
- [x] **Layout padding**: Fixed — changed to PadX("3p"), removed PadY; changed button text to "Set Window Icon to Globe"

### 64. `widget_demo`
- [x] Lists all 63 demos (matching count of all demos in demos/ excluding widget_demo itself)
- [x] Categories match Tcl's widget.tcl: Labels/Listboxes/Entries/Text/Canvases/Scales/Paned/Menus/Dialogs/Animation/Misc

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
1. `form` — grid→pack (change to grid) ✅ DONE
2. `fontchoose` — grid→pack (change to grid) ✅ DONE
3. `labelframe` — grid→pack (change to grid) ✅ DONE
4. `menubu` — grid→pack (change to grid) ✅ DONE
5. `sayings` — grid→pack (change to grid) ✅ DONE
6. `cscroll` — grid for scrollbar layout (change to grid) ✅ DONE
7. `textpeer` — grid→pack (change to grid) ✅ DONE
8. `toolbar` — grid→pack (change to grid) ✅ DONE
9. `tree` — grid→pack (change to grid) ✅ DONE
10. `ttkprogress` — grid→pack (change to grid) ✅ DONE
11. `ttknote` (tab 1) — grid→pack (change to grid) ✅ DONE

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
