# Takigo Translation Analysis

## Scope

This review compares the Go port against the Tk/Tcl source mapping described in `CLAUDE.md`, while explicitly excluding the items marked not done in `MISSING_FUNCTIONALITY.md` (`tkUnixPrint`, `tkUnixEmbed`, `tkIcu`, `tkAccessibility`).

The codebase builds and tests successfully with `go test ./...`, so the focus here is translation fidelity rather than build breakage.

## High-level conclusion

The port covers a large, usable subset of Tk, but several subsystems are intentionally narrower than upstream Tk. The biggest patterns are:

- Tk command/interpreter behavior is replaced with simpler Go-native APIs.
- Some “implemented” subsystems only cover the common path and omit Tk edge cases.
- A few features are present in name, but the behavior is materially simplified compared with Tk 9.1.
- There are a small number of places where the current implementation looks misleading or incomplete enough to count as wrong, not just simplified.

## Implemented incorrectly or materially incomplete

### 1. Binding engine ignores multi-event sequences

Evidence:

- `bind/dispatch.go:71-73` skips any binding where `len(b.seq.Patterns) != 1` with the comment `only single-pattern sequences for now`.

Impact:

- Tk supports binding sequences such as multi-event patterns and more complex sequence matching in `tkBind.c`.
- The parser accepts `Sequence`, but the dispatcher only evaluates one-pattern bindings, so the surface API implies support that the runtime does not actually provide.

Assessment:

- This is more than a simplification of API shape; it is a partial implementation of Tk binding semantics.

### 2. Message widget is missing the documented variable linkage

Evidence:

- `widget/message/message.go:21-34` defines only static text/layout fields.
- `widget/message/message.go:39-121` exposes options for text, font, colors, aspect, width, padding, and highlight width.
- There is no `textvariable`-style linkage, variable field, or listener wiring.

Impact:

- `MISSING_FUNCTIONALITY.md` describes message variable linkage as a completed feature, but the actual widget does not implement it.
- This is a documentation/implementation mismatch and a missing piece relative to `tkMessage.c`.

Assessment:

- This should be treated as a real missing feature inside an otherwise implemented widget.

### 3. `dialog.ChooseFont` is only partially implemented

Evidence:

- `dialog/fontchooser.go:99-109` creates “Bold” and “Italic” labels.
- `dialog/fontchooser.go:111-121` creates a preview area.
- `dialog/fontchooser.go:141-146` explicitly discards those variables with `_ = boldLabel`, `_ = italicLabel`, `_ = previewLabel`.
- The returned descriptor at `dialog/fontchooser.go:166-172` uses `selectedBold` / `selectedItalic`, but those values are only initialized from the initial font and are never changed by user interaction.

Impact:

- The UI suggests togglable style controls and live preview, but neither is wired.
- Family and size selection work through the listboxes, but bold/italic are effectively read-only, and the preview does not reflect user changes.

Assessment:

- This is implemented in a misleading/incomplete way rather than being a clean simplification.

### 4. Canvas window items are not true Tk canvas window items

Evidence:

- `canvas/item_window.go:8-15` describes the item as an overlay and says only `"nw"` anchor is currently supported.
- `canvas/item_window.go:74-77` makes item display a no-op.
- `canvas/canvas.go:245-278` positions embedded windows with raw `MoveResizeWindow` calls after drawing.

Impact:

- Tk canvas window items participate in the canvas item model more deeply; here they are effectively external child windows laid on top of the canvas.
- Anchor handling is incomplete.
- Rendering/clipping/stacking behavior will diverge from Tk in edge cases because the window is not part of the canvas pixmap rendering path.

Assessment:

- This is a deliberate partial translation. It works for common demo cases but is not feature-equivalent.

### 5. Text embedded windows are overlay-positioned by line, not by full text layout position

Evidence:

- `widget/text/text.go:654-662` stores an embedded window at an index.
- `widget/text/text.go:664-689` positions the window at `t.inset + dl.leftMargin` / line Y for the first visible display line with the same logical line.
- The implementation ignores the character offset within the line and does not reserve inline layout space.

Impact:

- Embedded windows are not part of text flow the way Tk text windows are.
- Multiple embedded windows on the same line, embedding at mid-line positions, and wrapping interactions will diverge from `tkText` behavior.

Assessment:

- This is a substantial simplification of text window embedding.

## Simplified relative to Tk

### 6. Text index parsing is only a practical subset

Evidence:

- `widget/text/index.go:171-222` supports only `line.char`, `end`, `insert`, `sel.first`, `sel.last`, and mark names.
- The file header says it ports `a practical subset of tk/generic/tkText*.c`.

Impact:

- Tk text indices support many more expressions and modifiers (`+/- chars`, words, lines, display indices, tag-relative forms, etc.).
- Any higher-level behavior depending on full Tk index grammar will not translate directly.

### 7. Selection/clipboard support is minimal

Evidence:

- `selection/selection.go:40-45` explicitly says incremental transfers are not implemented.
- `selection/selection.go:95-103` supports only `TARGETS`, `UTF8_STRING`, and `XA_STRING`.

Impact:

- Large X11 selection transfers and richer target negotiation from `tkSelect.c` / Tk clipboard behavior are missing.
- Interop with other X11 clients will be more limited than Tk’s.

### 8. Font descriptor parsing is much narrower than Tk

Evidence:

- `font/font.go:66-90` supports only XLFD, simple option-value, and simple `"Family size style"` formats.
- `font/font.go:146-180` tokenizes option-value descriptors with `strings.Fields`.
- `font/font.go:183-190` explicitly says Tcl brace quoting is simplified.

Impact:

- Tk font descriptors allow richer list semantics, quoting, and command-level behavior.
- Family names with spaces and Tcl-style structured descriptors are not handled with Tk fidelity.

### 9. Dialogs are Go reimplementations of the Tcl scripts, but much narrower

Examples:

- `dialog/colorchooser.go:128-156` returns the slider-driven color only; the hex entry is updated from the sliders, but user edits to the entry are never parsed back into the chooser state.
- `dialog/filedialog.go` is a basic directory/listbox browser and does not aim for the full behavior of `tkfbox.tcl`.
- `dialog/messagebox.go` uses simple label-based icons and button sets, not Tk’s platform-specific script behavior.

Impact:

- The dialogs are functional, but they are best understood as practical replacements, not close translations of the Tcl dialog logic.

### 10. Image subsystem is reduced to a simplified photo model

Evidence:

- `image/image.go:1-4` explicitly says the port is simplified.
- `image/photo.go:43-52` supports file decode through Go stdlib decoders.
- `image/photo.go:43-45` documents PNG and GIF support only.
- `image/photo.go:54-66` adds XBM through custom parsing.

Impact:

- This is a much smaller image system than Tk’s `tkImage.c` + format handler architecture.
- There is no general image-type plug-in model comparable to Tk image handlers.
- Animated GIF behavior is not implemented as Tk-style image animation; stdlib decode gives a static decoded image.

### 11. Variable support is a lightweight observable, not Tk variable semantics

Evidence:

- `widget/variable.go:1-44` implements a generic in-memory observable value with `Get`, `Set`, and `OnChange`.
- The file comment calls it “a simple observable value”.

Impact:

- This is useful for checkbutton/radiobutton linkage, but it is not a port of Tcl variable behavior.
- No named variables, no trace API, no read/write/unset traces, no interpreter/global scoping behavior.

### 12. WM support is intentionally only the essential subset

Evidence:

- `wm/wm.go:1-3` says it ports the essential subset of `tkUnixWm.c`.

Impact:

- Core title/geometry/state handling is present, but Tk’s full WM surface is much larger.
- This is expected per project goals, but it should be understood as a subset, not a full WM port.

## Additional behavior divergences worth noting

### 13. Systray is functional but visually placeholder-grade

Evidence:

- `systray/systray.go:107-113` redraws a hard-coded placeholder.
- `systray/systray.go:139-156` paints a filled circle icon instead of using caller-provided image data.

Impact:

- Docking and click handling exist, but the actual tray icon rendering is not a real translation of Tk’s tray/icon capabilities.

### 14. Tearoff menus are detached copies, not live menu views

Evidence:

- `widget/menu/tearoff.go:36-38` says the tearoff window contains copies of the menu entries.
- `widget/menu/tearoff.go:116-124` stores `entries: append([]MenuEntry(nil), m.entries...)`.

Impact:

- Detached menu state can diverge from the original menu after tearoff.
- Tk tearoff behavior is more tightly integrated with the menu’s live state than this copied-entry model.

### 15. TTK spinbox and sizegrip are practical ports, not complete ones

Evidence:

- `ttk/spinbox.go:16-60` defines a focused subset around text, range/value modes, command callback, and validation callback.
- `ttk/sizegrip.go:15-25` defines only a bottom-right resize grip with drag state.
- `ttk/sizegrip.go:49` hardcodes `XC_bottom_right_corner`.

Impact:

- The main demo behavior is there, but these are not exhaustive translations of Tk’s TTK widget semantics and platform nuances.
- Sizegrip especially is X11-pragmatic rather than a full Tk parity implementation.

## Summary

The translation is strongest in the core widget/event/layout path and weakest where Tk depends on richer Tcl semantics, protocol edge cases, or script-driven UI logic.

The most important concrete gaps inside the “implemented” surface are:

1. Binding sequences are parsed but not dispatched.
2. Message widget variable linkage is missing despite being described as done.
3. Font chooser UI is partially wired and misleading.
4. Canvas/text embedded windows are overlay-based simplifications rather than full Tk embeddings.
5. Selection, images, dialogs, variables, and WM are all narrower than upstream Tk by design.

If you want, the next useful step is to convert this report into a prioritized fix list: “actual bugs/misleading implementations” first, then “high-value parity gaps” second.
