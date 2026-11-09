# dialog/ — standard dialogs

Applies to `dialog/`; the root and `widget/AGENTS.md` still apply. The
dialogs port the Tcl megawidgets of `tk/library`: `dialog.tcl`
(`dialog/tkdialog.go`), `msgbox.tcl` (`dialog/messagebox.go`), `clrpick.tcl`
(`dialog/colorchooser.go`), `fontchooser.tcl` (`dialog/fontchooser.go`),
`tkfbox.tcl` and `choosedir.tcl` (`dialog/filedialog.go`,
`dialog/choosedir.go`), plus the helpers of `tk.tcl` (`::tk::FindAltKeyTarget`,
`::tk::UnderlineAmpersand`, `AmpWidget`) in `dialog/dialog.go`.

- **Modal loop.** `Dialog.Run` takes a local grab and blocks in
  `App.RunNestedLoop` (`RunNestedLoopContext` when a context was given)
  until `Close` or the toplevel is destroyed. Never block the loop
  goroutine any other way (root `AGENTS.md`, `THREADING.md`).
- **Sizing.** pack and grid propagate requested sizes at idle, so `Run`
  calls `arrangeNow` (children first, `pack.ArrangeContainer` +
  `grid.ArrangeContainer`) before it reads `Content`/`BtnFrame`
  `ReqWidth`/`ReqHeight`, then `pack.ArrangeAll`/`grid.ArrangeAll` with the
  final size. A new dialog that skips this renders clipped.
  `Dialog.Resizable` makes the computed size the WM minimum instead of a
  fixed size.
- **Buttons and keys.** Every `Dialog` binds Escape to `ResultCancel`;
  `addButtons` builds the button row and binds Return to the default
  button. Alt mnemonics come from `&` in captions (`underlineAmpersand`,
  found at key time by `findMnemonic`). A popup inside a dialog (combobox dropdown) sets `ev.Handled` on
  the keys it consumes, or the dialog acts on them too.
- **Icon-only ttk buttons** need `ttk.ButtonImage`: the label element sizes
  from the image only when the text is empty.

## Checking a dialog by hand

The demo pipeline does not click through dialogs. Use a scratch module
outside the repo (`replace github.com/takigo/takigo => <checkout>` in its
`go.mod`) whose `main` opens the dialog from `app.After`, and run it under
`xvfb-run -a -s "-screen 0 1280x1024x24 -noreset"`. Drive it with
`xdotool mousemove X Y click 1`, `xdotool type`, `xdotool key` and capture
with `import -window root shot.png`. Without a window manager the dialog
sits at 0,0, so the coordinates in a `TAKIGO_DUMP_TREE` dump are xdotool
coordinates. Compare `reqw`/`reqh` with `w`/`h` in the dump before tuning
padding.
