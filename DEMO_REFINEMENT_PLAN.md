# Demo Refinement Plan

Align `demos/` with `tk/library/demos/` so that names, categories, descriptions, data, layout, and functionality match the original Tk demos as closely as possible.

---

## 1. Widget Demo Launcher (`widget_demo`)

The launcher should mirror `tk/library/demos/widget` — a text widget with clickable hyperlinks organized by category, not a listbox.

### Changes:
- **Layout**: Replace listbox+source-viewer split with a single scrollable Text widget (like the Tk original) that shows categorized demo links as clickable text
- **Title**: "Widget Demonstration" (match Tk)
- **Categories**: Reorganize to match the 12 Tk categories exactly:
  1. Labels, buttons, checkbuttons, and radiobuttons
  2. Listboxes and Trees
  3. Entries, Spin-boxes and Combo-boxes
  4. Text
  5. Canvases
  6. Scales and Progress Bars
  7. Paned Windows and Notebooks
  8. Menus and Toolbars
  9. Common Dialogs
  10. Animation
  11. Miscellaneous
- **Demo entries**: Match the exact list from Tk's widget file (see Section 12 below for the full mapping)
- **Buttons**: Bottom frame with "Quit" button (keep)
- **Status bar**: Keep, show demo name on hover
- **Behavior**: Click a numbered demo entry to launch it via `go run .`; show "See Code" popup with source viewer (separate toplevel or inline)
- **Text tags**: `title`, `demo`, `visited`, `hot` for hover/click styling
- Skip: Mac-specific demos (mac_styles, mac_tabs, mac_wm) — not applicable to X11

---

## 2. Demo Description Texts

Every demo uses `demohelper.Setup(title, w, h, description)`. The description should match the Tk original's description label text (the `wraplength 4i` label at the top of each demo).

### Per-demo title and description updates:

| Demo | New Title | Description (match Tk) |
|------|-----------|----------------------|
| label | "Label Demonstration" | "Five labels are displayed below, each showing a different relief..." *(update to match Tk's actual text about text+image labels)* |
| button | "Button Demonstration" | "If you click on any of the four buttons below, the background of the button area will change to the color indicated..." |
| check | "Checkbutton Demonstration" | "Four checkbuttons are displayed below. The first is a tri-state Safety Check..." |
| radio | "Radiobutton Demonstration" | "Three groups of radiobuttons are displayed below..." |
| entry1 | "Entry Demonstration (no scrollbars)" | Keep, but add Motif/Emacs binding mention |
| entry2 | "Entry Demonstration (with scrollbars)" | Keep, but add Motif/Emacs binding mention |
| entry3 | "Constrained Entry Demonstration" | Rename from "Validated Entries" |
| form | "Form Demonstration" | Rename from "Form Entry" |
| states | "Listbox Demonstration (50 states)" | Update to mention justification control |
| colors | "Listbox Demonstration (colors)" | Update to mention double-click palette change |
| sayings | "Listbox Demonstration (well-known sayings)" | Update to mention horizontal+vertical scrollbars |
| text | "Text Demonstration - Basic Facilities" | Match Tk |
| style | "Text Demonstration - Display Styles" | Match Tk |
| hscale | "Horizontal Scale Demonstration" | Match Tk |
| vscale | "Vertical Scale Demonstration" | Match Tk |
| paned1 | "Horizontal Paned Window Demonstration" | Match Tk |
| paned2 | "Vertical Paned Window Demonstration" | Match Tk |
| items | "Canvas Item Demonstration" | Match Tk's long description |
| plot | "Plot Demonstration" | Match Tk |
| arrow | "Arrowhead Editor Demonstration" | Match Tk — interactive editor |
| ruler | "Ruler Demonstration" | Match Tk — tab stop dragging |
| cscroll | "Scrollable Canvas Demonstration" | Match Tk |
| ctext | "Canvas Text Demonstration" | Match Tk — editable text |
| floor | "Floorplan Canvas Demonstration" | Match Tk |
| knightstour | "Knight's Tour" | Keep |
| menu | "Menu Demonstration" | Match Tk |
| menubu | "Menu Button Demonstration" | Match Tk |
| toolbar | "Toolbar Demonstration" | Match Tk |
| ttkbut | "Simple Ttk Widgets" | Match Tk |
| ttkpane | "Themed Nested Panes" | Match Tk |
| ttkscale | "Themed Scale Demonstration" | Match Tk |
| ttknote | "Ttk Notebook Widget" | Match Tk |
| ttkprogress | "Progress Bar Demonstration" | Match Tk |
| combo | "Combobox Demonstration" | Match Tk |
| ttkmenu | "Ttk Menu Buttons" | Match Tk |
| ttkspin | "Themed Spinbox Demonstration" | Match Tk |
| msgbox | "Message Box Demonstration" | Match Tk |
| filebox | "File Selection Dialogs" | Match Tk |
| clrpick | "Color Selection Dialog" | Match Tk |
| fontchoose | "Font Selection Dialog" | Match Tk |
| dialog1 | *(keep)* | Match Tk |
| dialog2 | *(keep)* | Match Tk |
| anilabel | "Animated Label Demonstration" | Match Tk |
| aniwave | "Animated Wave Demonstration" | Match Tk |
| pendulum | "Pendulum Animation Demonstration" | Match Tk |
| goldberg | "TkGoldberg" | Match Tk |
| bitmap | "Bitmap Demonstration" | Match Tk |
| image1 | "Image Demonstration #1" | Match Tk |
| image2 | "Image Demonstration #2" | Match Tk |
| puzzle | "15-Puzzle Demonstration" | Match Tk |
| unicodeout | "Unicode Label Demonstration" | Match Tk |
| systray | "System Tray Demonstration" | Match Tk |
| print | "Printing Demonstration" | Match Tk |
| windowicons | "Window Icon Demonstration" | Keep |
| icon | "Iconic Button Demonstration" | Match Tk |

---

## 3. Per-Demo Functional Fixes (by category)

### 3.1 Labels, Buttons, Checkbuttons, Radiobuttons

**label** — Restructure to match Tk:
- Left frame: 3 text labels (flat, raised, sunken relief)
- Right frame: 1 photo image label + 1 caption label below it
- Need an image file or generated image for the right-side photo
- Keep 2 frames side-by-side with `pack side=left`

**button** — Fix colors:
- Change button colors to: Peach Puff, Light Blue, Sea Green, Yellow (match Tk)
- Remove auto-reset behavior (not in Tk)
- Description should mention Tab/Space keyboard navigation

**check** — Add tri-state master:
- Add 4th checkbutton "Safety Check" at top as tri-state master
- Master reflects partial/all/none state based on sub-checkbuttons
- Remove status label (not in Tk)
- Add "See Variables" button equivalent or skip (Tk-specific)

**radio** — Add third group:
- Add "Alignment" group (3rd labelframe) with radiobuttons
- Use labelframes instead of plain frames for all 3 groups
- Add 6 colors (Red, Green, Blue, Yellow, Orange, Purple) instead of 5
- Remove status label

**puzzle** — Randomize:
- Use random shuffle for initial tile positions (not fixed order)

**icon** — Match Tk's `indicatoron 0` concept:
- Show buttons with images that toggle appearance (no indicator)
- Use flag-up/flag-down style images
- This requires `indicatoron` support — if not available, approximate with button image swapping

**image1** — Use real images:
- Load actual image files (earth.gif equivalent) instead of generated gradients
- Or bundle small PNG/JPEG files in the demos directory
- Show 2 labels with loaded photo images

**image2** — Add directory browsing:
- Add an entry for directory path + "Select Dir." button
- Listbox populated from directory listing
- Double-click loads and displays selected image file
- Requires file-based image loading support

**labelframe** — Match Tk:
- One labelframe with radiobuttons (gender or similar)
- One labelframe with a widget as the label (if supported) or just text label
- Current implementation seems reasonable, verify details match

**ttkbut** — Major expansion:
- Add 4 widget groups: Buttons (per-theme), Checkbuttons, Radiobuttons, Toggleswitch
- Add enable/disable toggle
- Use grid layout, not single-column pack
- Each theme button switches the active theme

---

### 3.2 Listboxes and Trees

**states** — Add justification control:
- Add "Justification" labelframe with Left/Center/Right radiobuttons above the listbox
- Clicking a radiobutton changes listbox text justification
- Use TTK scrollbar

**colors** — Fix color data:
- Add missing color families (SlateGray1-4, LightSteelBlue1-4)
- Double-click should change application palette (as much as possible), not just root bg

**sayings** — Add horizontal scrollbar:
- Add both X and Y scrollbars (the demo's primary point)
- Use grid layout instead of pack for proper 2-scrollbar arrangement
- Match Tk's sayings data set

**mclist** — Already good. Verify column headings and data match Tk (countries/capitals/currencies).

**tree** — Already good. Verify lazy-load behavior matches.

---

### 3.3 Entries, Spin-boxes, Combo-boxes

**entry1** — Minor fixes:
- Set placeholder text to "Enter text here" (no trailing `...`)
- Add `placeholderforeground` if supported

**entry2** — Minor fixes:
- Use TTK scrollbar if available
- Add placeholder on entry 3
- Add spacer frames between pairs

**entry3** — Rename + restructure:
- Rename to "Constrained Entry Demonstration"
- Use 2x2 grid of labelframes (not 4-row single column)
- Labels should be labelframe titles: "Integer Entry", "Length-Constrained Entry", "US Phone-Number Entry", "Password Entry"
- Add validation where possible (integer check, length limit, password max 8 chars)

**form** — Minor fixes:
- Rename to "Form Demonstration"
- Bind Return to close window
- Entry width 40 (not 30)
- Keep City/State labels (Go improvement over Tcl)

**spin** — Fix data:
- Use Tk's data: integer 1-10 with validation, float 0-3 step 0.5 with `%05.2f` format, Australian cities
- Remove wrap on cities spinbox

**ttkspin** — Fix widget type + data:
- Must use TTK spinbox widget (not classic)
- Data: integer 1-10, float 0-3 step 0.5, Australian cities (Canberra, Sydney, Melbourne, Perth, Adelaide, Brisbane, Hobart, Darwin, "Alice Springs")

**combo** — Fix data + layout:
- Use Australian cities data (not countries)
- Use labelframes instead of plain labels
- First combobox: "Fully Editable" with placeholder, Return-key binding adds new values
- Second: "Disabled"
- Third: "Defined List Only" (readonly)

---

### 3.4 Text

**text** — Add font chooser button:
- Add "Show Font Dialog" button integration if feasible
- Match Tk's sample text content about scrolling, editing, selection

**style** — Add missing tag styles:
- Add: raised/sunken relief tags, superscript/subscript, right/center justify, margins/spacing
- Match font choices to Tk (use Courier for code samples)

**bind** — Expand linked demos:
- Add all ~12 canvas/animation demo links (not just 6)
- Match Tk's link text and descriptions

**search** — Add file loading:
- Add "File" entry + "Load File" button
- Add blinking highlight toggle via `After`
- Match Tk's text search mechanism

**twind** — Note limitation:
- Embedded windows in text not supported; keep as rich-text demo
- Update description to match Tk's intent
- Consider if any embedded-widget features can be approximated

**textpeer** — Note limitation:
- True peer text (shared document) not implemented
- Keep copy buttons as workaround
- Update description honestly

---

### 3.5 Canvases

**items** — Add interactivity + scrollbars:
- Add x+y scrollbars with scroll region
- Add drag-to-move items (Button-1 press/motion)
- Add hover effects (Enter/Leave on items)
- Add more item types if possible (bitmap items, smooth curves)

**plot** — Implement point dragging:
- Add ButtonPress-1 on data points → B1-Motion drag
- Update connecting line as points move
- Remove sine overlay (not in Tk)

**arrow** — Make interactive:
- Replace static display with editable arrow
- Add 3 draggable control handles (a/b/c shape parameters)
- Show small example arrows that update in real time
- Display current arrowshape values as text

**ruler** — Make interactive:
- Add tab-stop "well" at right end
- Drag from well to create new tab stops
- Drag existing tab stops to reposition
- Drag far off ruler to delete

**cscroll** — Fix grid + add interaction:
- Expand to 20x20 grid (not 10x10)
- Add Button-2 drag to pan
- Add mouse wheel scroll
- Add Button-1 click prints cell index to stdout

**ctext** — Make interactive:
- Add anchor selector boxes (clickable radio-button-like)
- Add justification selector boxes
- Add angle pie sectors
- Support in-canvas text editing

**floor** — Expand to multi-floor:
- Consider loading actual room data (or simplified version)
- Add floor selection buttons
- Add mouse hover highlighting rooms
- Add room number entry field

**knightstour** — Add controls:
- Add Stop and Reset buttons
- Add speed/delay scale slider

---

### 3.6 Scales and Progress Bars

**hscale** — Fix visual:
- Change from angle rotation to arrow length control (match Tk)
- Draw polygon+line arrow shape that resizes
- Add tick interval 50 on scale
- Range 0-250

**vscale** — Fix visual:
- Change from bar chart to vertical arrow shape (match Tk)
- Draw polygon+line arrow shape that resizes
- Add tick interval 50 on scale
- Range 0-250

**ttkscale** — Fix to use TTK + rainbow colors:
- Use TTK scale widget (not classic)
- Range 0-5 (not 0-100)
- Display rainbow color names (Red, Orange, Yellow, Green, Blue, Violet)
- Change label foreground color to match

**ttkprogress** — Minor layout fix:
- Use grid layout (not pack)
- Single Start/Stop pair controls both bars

---

### 3.7 Paned Windows and Notebooks

**paned1** — Fix to 2 panes:
- Reduce to 2 panes (left yellow, right cyan) to match Tk
- Label text: "This is the\nleft side", "This is the\nright side"

**paned2** — Fix content:
- Top pane: listbox of Tk widget names (button, canvas, checkbutton, etc.)
- Bottom pane: text widget with x+y scrollbars in grid layout
- Highlight first list item

**ttkpane** — Major expansion:
- Implement nested panedwindows (outer horizontal, inner vertical on each side)
- Pane contents: Button pane, Clocks pane (live world clocks via After), Progress pane (running indeterminate bar), Text pane

**ttknote** — Fix tabs:
- Tab 1: Description + "Neat!" button + transient message
- Tab 2: Disabled tab (cannot be selected)
- Tab 3: Text editor with scrollbar
- Add Ctrl+Tab traversal

---

### 3.8 Menus and Toolbars

**menu** — Major rework:
- Replace File/Edit/Help with Tk's structure: File (with accelerators), Basic, Cascades (with check/radio submenus)
- Add keyboard accelerator display
- Add check/radiobutton menu items
- Add tearoff indicator if feasible

**menubu** — Fix layout:
- Use compass grid layout (3x3) instead of horizontal row
- Label text: "Below", "Right", "Left", "Above" with center-flush if applicable
- Menu items: "Below menu: first item", etc.

**ttkmenu** — Rethink concept:
- Change from File/Edit/Help to theme-switching menubuttons (match Tk)
- 5 menubuttons in compass directions, each listing available themes
- One styled as Toolbutton

**toolbar** — Add features:
- Add Toolbutton-styled checkbutton
- Add combobox with font families
- Add text widget below that shows output
- Consider tearoff handle if feasible

---

### 3.9 Common Dialogs

**msgbox** — Make interactive:
- Replace 5 fixed buttons with radiobutton selectors
- Left column: icon type radios (error, info, question, warning)
- Right column: button type radios (abortretryignore, ok, okcancel, retrycancel, yesno, yesnocancel)
- Single "Message Box" button triggers the selected combination
- Show second dialog confirming which button was pressed

**filebox** — Add entry widgets:
- Replace simple status label with per-operation entry widgets
- Add richer file type filters
- Grid layout with label+entry+browse-button rows

**clrpick** — Add second button:
- Two buttons: "Set background color" and "Set foreground color"
- Apply chosen colors to all widgets in the window

**fontchoose** — Apply font live:
- Replace static label with a text widget + scrollbar
- Actually apply the chosen font to the text widget
- Use named font object

**dialog1** — Keep current approach (button + modal), minor text fixes.

**dialog2** — Keep, note global grab limitation.

---

### 3.10 Animation

**anilabel** — Expand:
- Add 3 scrolling text labels (different messages) in a labelframe
- Add 1 animated GIF/image label in a separate labelframe (if animated images supported, otherwise note limitation)

**aniwave** — Add controls:
- Add Play/Pause button
- Add direction control
- Match background (black) and wave color (green)

**pendulum** — Add phase space graph:
- Add second canvas (right side) for angle vs angular velocity phase space plot
- Add click/drag to reposition pendulum bob

**goldberg** — Expand:
- Add more mechanisms beyond 3 ramps
- Closer approximation of the multi-stage chain reaction

---

### 3.11 Miscellaneous

**bitmap** — Use actual X11 bitmap patterns:
- Generate RGBA images that match the actual X11 bitmap pixel patterns (gray stipple densities, icon shapes)
- Or document limitation if pixel-accurate bitmaps aren't feasible

**unicodeout** — Expand + fix layout:
- Two-column grid: language name (left) + text sample (right)
- Add missing languages: Armenian, Georgian, Thai
- Match Tk's sample texts where possible

**systray** — Add lifecycle controls:
- Replace single "Add" button with Create/Modify/Destroy buttons in a labelframe
- Add "Display Notification" button (if sysnotify is supported)
- Remove auto-removal timer

**print** — Add print buttons:
- Add "Print Canvas" and "Print Text" buttons
- They can show "printing not available" dialog or attempt `lp`/`lpr` if feasible

**windowicons** — Add badge buttons:
- Add buttons: "Set Badge to 3", "Set Badge to 11", "Reset Badge"
- If badge not implementable, keep as interactive icon-setting demo with button

---

## 4. Priority Ordering

### Priority 1 — Structural alignment (do first)
1. Reorganize `widget_demo` categories and demo list to match Tk
2. Update all demo titles and descriptions to match Tk
3. Fix demo data sets (colors in button, Australian cities in combo/ttkspin/spin, sayings, states)

### Priority 2 — Layout and widget corrections
4. `label` — restructure to text+image layout
5. `paned1` — reduce to 2 panes, fix colors/text
6. `hscale`/`vscale` — fix canvas visuals to match Tk's arrow shape
7. `ttkscale` — rainbow colors, range 0-5
8. `entry3` — rename, restructure as 2x2 labelframes
9. `sayings` — add horizontal scrollbar
10. `states` — add justification radiobuttons
11. `unicodeout` — two-column grid, expand languages
12. `menubu` — compass grid layout
13. `msgbox` — interactive radiobutton selection
14. `clrpick` — two buttons (fg/bg)
15. `filebox` — entry widgets per operation
16. `fontchoose` — text widget with live font application

### Priority 3 — Feature additions (existing demos)
17. `check` — tri-state master checkbutton
18. `radio` — 3rd Alignment group + labelframes
19. `ttkbut` — expand to full TTK showcase
20. `ttkpane` — nested panes with live content
21. `ttknote` — disabled tab, text editor tab
22. `ttkmenu` — theme-switching menubuttons
23. `toolbar` — checkbutton, combobox, text output area
24. `menu` — accelerators, check/radio items, cascades
25. `combo` — Australian cities, labelframes, Return-key binding
26. `anilabel` — 3 labels + image label
27. `aniwave` — play/pause/direction controls
28. `pendulum` — phase space graph + drag interaction
29. `knightstour` — stop/reset/speed controls
30. `systray` — create/modify/destroy buttons + notification

### Priority 4 — Interactivity additions (canvas demos)
31. `items` — scrollbars, drag-to-move, hover effects
32. `plot` — point dragging
33. `arrow` — interactive editor with draggable handles
34. `ruler` — drag-to-create/move/delete tab stops
35. `cscroll` — expand grid, add pan/wheel/click
36. `ctext` — interactive controls + text editing
37. `floor` — multi-floor, hover, room entry
38. `search` — file loading, blink toggle

### Priority 5 — Limitations to document/skip
39. `twind` — embedded windows not supported (document)
40. `textpeer` — shared document not supported (document)
41. `image1`/`image2` — need real image files or close approximation
42. `goldberg` — expand incrementally
43. `print` — printing not available (add placeholder buttons)

---

## 5. Full Demo List Mapping

| # | Tk Demo | Go Demo Dir | Status |
|---|---------|-------------|--------|
| 1 | label | label | EXISTS — needs restructure |
| 2 | unicodeout | unicodeout | EXISTS — needs expansion |
| 3 | button | button | EXISTS — needs color fix |
| 4 | check | check | EXISTS — needs tri-state |
| 5 | radio | radio | EXISTS — needs 3rd group |
| 6 | puzzle | puzzle | EXISTS — needs random shuffle |
| 7 | icon | icon | EXISTS — needs rework |
| 8 | image1 | image1 | EXISTS — needs real images |
| 9 | image2 | image2 | EXISTS — needs dir browsing |
| 10 | labelframe | labelframe | EXISTS — verify |
| 11 | ttkbut | ttkbut | EXISTS — needs expansion |
| 12 | states | states | EXISTS — needs justification |
| 13 | colors | colors | EXISTS — needs color data fix |
| 14 | sayings | sayings | EXISTS — needs horiz scrollbar |
| 15 | mclist | mclist | EXISTS — verify |
| 16 | tree | tree | EXISTS — verify |
| 17 | entry1 | entry1 | EXISTS — minor fixes |
| 18 | entry2 | entry2 | EXISTS — minor fixes |
| 19 | entry3 | entry3 | EXISTS — needs restructure |
| 20 | spin | spin | EXISTS — needs data fix |
| 21 | ttkspin | ttkspin | EXISTS — needs TTK widget |
| 22 | combo | combo | EXISTS — needs data/layout fix |
| 23 | form | form | EXISTS — minor fixes |
| 24 | text | text | EXISTS — minor fixes |
| 25 | style | style | EXISTS — needs more tags |
| 26 | bind | bind | EXISTS — needs more links |
| 27 | twind | twind | EXISTS — document limitation |
| 28 | search | search | EXISTS — needs file loading |
| 29 | textpeer | textpeer | EXISTS — document limitation |
| 30 | items | items | EXISTS — needs interactivity |
| 31 | plot | plot | EXISTS — needs dragging |
| 32 | ctext | ctext | EXISTS — needs interactivity |
| 33 | arrow | arrow | EXISTS — needs interactivity |
| 34 | ruler | ruler | EXISTS — needs interactivity |
| 35 | floor | floor | EXISTS — needs expansion |
| 36 | cscroll | cscroll | EXISTS — needs fixes |
| 37 | knightstour | knightstour | EXISTS — needs controls |
| 38 | hscale | hscale | EXISTS — needs visual fix |
| 39 | vscale | vscale | EXISTS — needs visual fix |
| 40 | ttkscale | ttkscale | EXISTS — needs TTK + colors |
| 41 | ttkprogress | ttkprogress | EXISTS — minor layout fix |
| 42 | paned1 | paned1 | EXISTS — needs 2-pane fix |
| 43 | paned2 | paned2 | EXISTS — needs content fix |
| 44 | ttkpane | ttkpane | EXISTS — needs nested panes |
| 45 | ttknote | ttknote | EXISTS — needs fixes |
| 46 | menu | menu | EXISTS — needs rework |
| 47 | menubu | menubu | EXISTS — needs layout fix |
| 48 | ttkmenu | ttkmenu | EXISTS — needs rethink |
| 49 | toolbar | toolbar | EXISTS — needs features |
| 50 | msgbox | msgbox | EXISTS — needs interactive |
| 51 | filebox | filebox | EXISTS — needs entries |
| 52 | clrpick | clrpick | EXISTS — needs 2nd button |
| 53 | fontchoose | fontchoose | EXISTS — needs live apply |
| 54 | systray | systray | EXISTS — needs lifecycle |
| 55 | print | print | EXISTS — add buttons |
| 56 | dialog1 | dialog1 | EXISTS — minor fixes |
| 57 | dialog2 | dialog2 | EXISTS — minor fixes |
| 58 | anilabel | anilabel | EXISTS — needs expansion |
| 59 | aniwave | aniwave | EXISTS — needs controls |
| 60 | pendulum | pendulum | EXISTS — needs phase space |
| 61 | goldberg | goldberg | EXISTS — needs expansion |
| 62 | bitmap | bitmap | EXISTS — needs pattern fix |
| 63 | windowicons | windowicons | EXISTS — needs buttons |

**Tk demos NOT ported (skip)**:
- `mac_styles`, `mac_tabs`, `mac_wm` — macOS-only, not applicable
- `accessiblewidget` — screen reader support, skip for now

All 63 non-Mac demos exist in Go. No new demo directories need to be created.
