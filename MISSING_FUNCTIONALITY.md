# Missing Functionality Summary

This document lists the Tk functionalities that have not yet been translated from C to Go in takigo.

---

## 1. tkSquare

**Location:** `tk/generic/tkSquare.c`

**Description:** A demo/example widget that displays a movable/resizable square. Used primarily as an example of widget implementation. Not included in standard wish, but available in tktest.

**Status:** Implemented in `widget/square/square.go`. Demo in `demos/square/`.

**Key Features:**
- Movable/resizable square within a widget
- Options: `-background`, `-borderwidth`, `-doublebuffer`, `-foreground`, `-posx`, `-posy`, `-relief`, `-size`
- Standard widget command pattern (configure, cget, etc.)
- Double-buffered redisplay

---

## 2. tkMessage

**Location:** `tk/generic/tkMessage.c`

**Description:** A message widget that displays multi-line text with automatic line wrapping based on aspect ratio.

**Status:** Implemented in `widget/message/message.go`. Demo in `demos/msgwidget/`.

**Key Features:**
- Multi-line text display with word wrapping
- Options: `-aspect`, `-font`, `-foreground`, `-padx`, `-pady`, `-width`, `-anchor`, `-justify`
- Text layout caching
- Variable linkage (displays contents of a Tcl variable)
- Focus highlight support

**Note:** The message dialog (`dialog.MessageBox`) is already implemented in takigo—this is a separate widget.

---

## 3. tkUnixPrint (CUPS Printing)

**Location:** `tk/unix/tkUnixPrint.c`

**Description:** Unix platform implementation of printing via CUPS (Common Unix Printing System).

**Status:** Not implemented in takigo. Demo shows "not available" message.

**Key Features:**
- `::tk::print::cups defaultprinter` - Get default printer
- `::tk::print::cups getprinters` - List available printers
- `::tk::print::cups print` - Print data with options (-copies, -collate, -dpi, etc.)
- Integration with canvas PostScript output

**Dependencies:** Requires libcups development libraries.

---

## 4. tkUnixEmbed (Window Embedding)

**Location:** `tk/unix/tkUnixEmbed.c`

**Description:** X11 window embedding (Xembed protocol) - allows embedding a Tk window into another application or embedding a foreign window into Tk.

**Status:** Not implemented in takigo.

**Key Features:**
- `Tk_UseWindow` - Embed Tk window into foreign window
- Container/embedded window relationship management
- Event propagation between container and embedded app
- Focus handling for embedded windows
- Geometry requests from embedded window to container
- Xembed protocol support (XEMBED_MAPPED, XEMBED_FOCUS_IN, etc.)

**Related:** Canvas window items (`canvas.CreateWindow`) and text widget embedded windows are partially implemented.

---

## 5. tkIcu (Unicode Text Boundaries)

**Location:** `tk/generic/tkIcu.c`

**Description:** Unicode text boundary detection using ICU library - provides grapheme cluster and word boundary detection.

**Status:** Not implemented in takigo.

**Key Features:**
- `tk_wordBreakBefore` - Find word break before position
- `tk_wordBreakAfter` - Find word break after position
- `tk_graphemeBreakBefore` - Find grapheme break before position
- `tk_graphemeBreakAfter` - Find grapheme break after position
- Locale-aware text boundary detection
- Runtime ICU library loading

**Dependencies:** Requires libicu development libraries.

---

## 6. tkAccessibility (AT-SPI)

**Location:** `tk/unix/tkUnixAccessibility.c`

**Description:** Accessibility support for screen readers via AT-SPI (Assistive Technology Service Provider Interface) / Gnome Accessibility Toolkit.

**Status:** Not implemented in takigo.

**Key Features:**
- ATK object creation for each Tk widget
- Role mapping (Button, Checkbox, Entry, Label, etc.)
- Focus tracking and notification
- Screen reader communication
- Accessible name/description for widgets

**Dependencies:** Requires ATK and ATK-bridge libraries.

---

## 7. TTK Spinbox

**Tk Location:** `tk/generic/ttk/ttkEntry.c` (lines 1970-2189), `tk/library/ttk/spinbox.tcl`

**Description:** Themed spinbox widget (ttk::spinbox).

**Status:** Implemented in `ttk/spinbox.go`. Demo updated in `demos/ttkspin/`.

**Key Differences (TTK vs Classic):**

| Feature | Classic (takigo) | TTK (Tk) |
|---------|-----------------|----------|
| Package | `widget/spinbox` | `ttk` |
| Styling | Custom drawing | Theme-based elements |
| Elements | N/A | `Spinbox.field`, `Spinbox.uparrow`, `Spinbox.downarrow`, `Spinbox.padding`, `Spinbox.textarea` |
| Layout | Pack-based | TTK layout engine |

**TTK Spinbox Layout:**
```
TSpinbox
├── Spinbox.field (pack top, fill x)
│   ├── Spinbox.uparrow (pack top, stick e)
│   └── Spinbox.downarrow (pack bottom, stick e)
└── Spinbox.padding (fill both)
    └── Spinbox.textarea (fill both)
```

---

## 8. TTK Sizegrip

**Tk Location:** `tk/generic/ttk/ttkSeparator.c` (lines 83-134), `tk/generic/ttk/ttkElements.c` (lines 638-705), `tk/library/ttk/sizegrip.tcl`

**Description:** A grip handle in the lower-right corner of a window for interactive resizing.

**Status:** Implemented in `ttk/sizegrip.go`. Widget demo updated to use `ttk.NewSizegrip`.

**Key Features:**
- Widget class: `TSizegrip`
- Element: `Sizegrip.sizegrip` with configurable grip size
- Drag handling: `<Button-1>`, `<B1-Motion>`, `<ButtonRelease-1>`
- Resizes parent toplevel window on drag
- Automatic positioning (pack bottom-right corner)

**Element Implementation (ttkElements.c):**
- `SizegripElement` struct with background, gripSize options
- `SizegripSize` - Calculates grip handle size
- `SizegripDraw` - Draws diagonal grip lines

**takigo Current State:** Widget demo now uses `ttk.NewSizegrip` widget with interactive resize support.

---

## Summary Table

| Feature | File(s) | Priority | Dependencies |
|---------|---------|----------|--------------|
| ~~tkSquare~~ | ~~tkSquare.c~~ | ~~Low~~ | ~~Done~~ |
| ~~tkMessage~~ | ~~tkMessage.c~~ | ~~Medium~~ | ~~Done~~ |
| tkUnixPrint | tkUnixPrint.c | Medium | libcups |
| tkUnixEmbed | tkUnixEmbed.c | Medium | None |
| tkIcu | tkIcu.c | Low | libicu |
| tkAccessibility | tkUnixAccessibility.c | Low | libatk, libatk-bridge |
| ~~TTK Spinbox~~ | ~~ttkEntry.c, spinbox.tcl~~ | ~~Medium~~ | ~~Done~~ |
| ~~TTK Sizegrip~~ | ~~ttkSeparator.c, ttkElements.c, sizegrip.tcl~~ | ~~Medium~~ | ~~Done~~ |
