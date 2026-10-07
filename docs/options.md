# Option reference

Every widget option of Tk is a Go function in the widget's package that returns
that widget's option type, to pass to `New` or to `Configure`. This table lists
them with the Tk option each stands for. The classic widget packages use the
bare Tk name where it is free and an `Opt` suffix otherwise; the `ttk` package
prefixes every option with the widget name. A distance option takes pixels or
a `screenunit.Distance`; a colour option a name or `color.RGB`; a font option
a name or `font.Attributes`.

The Tk column is derived from the Go name, so where takigo chose another name
(`canvas.ArcStyleOpt` is Tk's `-style`, `BitmapBackground` its `-background`)
it is only a hint. Regenerate with `python3 scripts/options_doc.py > docs/options.md`.

## canvas

| Option type | Go option | Takes | Tk option |
|---|---|---|---|
| `CanvasOption` | `canvas.Antialias` | `on bool` | `-antialias` |
| `CanvasOption` | `canvas.Background` | `name C` | `-background` |
| `CanvasOption` | `canvas.BorderWidthOpt` | `bw L` | `-borderwidth` |
| `CanvasOption` | `canvas.CloseEnough` | `d float64` | `-closeenough` |
| `CanvasOption` | `canvas.Height` | `h L` | `-height` |
| `CanvasOption` | `canvas.HighlightWidthOpt` | `hw L` | `-highlightwidth` |
| `CanvasOption` | `canvas.ReliefOpt` | `r option.Relief` | `-relief` |
| `CanvasOption` | `canvas.ScrollRegion` | `x1, y1, x2, y2 int` | `-scrollregion` |
| `CanvasOption` | `canvas.Width` | `w L` | `-width` |
| `ItemOption` | `canvas.ActiveFill` | `name C` | `-activefill` |
| `ItemOption` | `canvas.AnchorOpt` | `a option.Anchor` | `-anchor` |
| `ItemOption` | `canvas.ArcStyleOpt` | `s ArcStyle` | `-arcstyle` |
| `ItemOption` | `canvas.Arrow` | `mode ArrowMode` | `-arrow` |
| `ItemOption` | `canvas.ArrowShape` | `a, b, c float64` | `-arrowshape` |
| `ItemOption` | `canvas.BitmapBackground` | `r, g, b, a uint8` | `-bitmapbackground` |
| `ItemOption` | `canvas.BitmapForeground` | `r, g, b uint8` | `-bitmapforeground` |
| `ItemOption` | `canvas.CapStyleOpt` | `capStyle int` | `-capstyle` |
| `ItemOption` | `canvas.Dash` | `pattern ...byte` | `-dash` |
| `ItemOption` | `canvas.DisabledFill` | `name C` | `-disabledfill` |
| `ItemOption` | `canvas.Extent` | `deg float64` | `-extent` |
| `ItemOption` | `canvas.FillColor` | `name C` | `-fillcolor` |
| `ItemOption` | `canvas.FillNone` | `` | `-fillnone` |
| `ItemOption` | `canvas.FontOpt` | `name F` | `-font` |
| `ItemOption` | `canvas.ImageOpt` | `img widget.WidgetImage` | `-image` |
| `ItemOption` | `canvas.JoinStyleOpt` | `join int` | `-joinstyle` |
| `ItemOption` | `canvas.JustifyOpt` | `j option.Justify` | `-justify` |
| `ItemOption` | `canvas.OutlineColor` | `name C` | `-outlinecolor` |
| `ItemOption` | `canvas.OutlineNone` | `` | `-outlinenone` |
| `ItemOption` | `canvas.OutlineStipple` | `spec string` | `-outlinestipple` |
| `ItemOption` | `canvas.OutlineWidth` | `width L` | `-outlinewidth` |
| `ItemOption` | `canvas.Smooth` | `on bool` | `-smooth` |
| `ItemOption` | `canvas.SplineSteps` | `n int` | `-splinesteps` |
| `ItemOption` | `canvas.StartAngle` | `deg float64` | `-startangle` |
| `ItemOption` | `canvas.StateOpt` | `s ItemState` | `-state` |
| `ItemOption` | `canvas.Stipple` | `spec string` | `-stipple` |
| `ItemOption` | `canvas.Tags` | `tags ...string` | `-tags` |
| `ItemOption` | `canvas.TextAngle` | `deg float64` | `-textangle` |
| `ItemOption` | `canvas.TextColor` | `name C` | `-textcolor` |
| `ItemOption` | `canvas.TextOpt` | `s string` | `-text` |
| `ItemOption` | `canvas.WidthOpt` | `width L` | `-width` |
| `PostscriptOption` | `canvas.PSColorMode` | `mode string` | `-pscolormode` |
| `PostscriptOption` | `canvas.PSColormapVar` | `m map[string]string` | `-pscolormapvar` |
| `PostscriptOption` | `canvas.PSFile` | `path string` | `-psfile` |
| `PostscriptOption` | `canvas.PSFontMapVar` | `m map[string][2]string` | `-psfontmapvar` |
| `PostscriptOption` | `canvas.PSPageAnchor` | `a option.Anchor` | `-pspageanchor` |
| `PostscriptOption` | `canvas.PSPageHeight` | `h float64` | `-pspageheight` |
| `PostscriptOption` | `canvas.PSPageWidth` | `w float64` | `-pspagewidth` |
| `PostscriptOption` | `canvas.PSPageX` | `x float64` | `-pspagex` |
| `PostscriptOption` | `canvas.PSPageY` | `y float64` | `-pspagey` |
| `PostscriptOption` | `canvas.PSProlog` | `b bool` | `-psprolog` |
| `PostscriptOption` | `canvas.PSRegion` | `x, y, w, h int` | `-psregion` |
| `PostscriptOption` | `canvas.PSRotate` | `b bool` | `-psrotate` |
| `PostscriptOption` | `canvas.PSTitle` | `t string` | `-pstitle` |
| `PostscriptOption` | `canvas.PSWriter` | `w io.Writer` | `-pswriter` |

## geometry/grid

| Option type | Go option | Takes | Tk option |
|---|---|---|---|
| `GridOption` | `grid.Column` | `col int` | `-column` |
| `GridOption` | `grid.ColumnSpan` | `n int` | `-columnspan` |
| `GridOption` | `grid.IPadX` | `p L` | `-ipadx` |
| `GridOption` | `grid.IPadY` | `p L` | `-ipady` |
| `GridOption` | `grid.In` | `container window.Windower` | `-in` |
| `GridOption` | `grid.PadX` | `p L` | `-padx` |
| `GridOption` | `grid.PadXPair` | `left A, right B` | `-padxpair` |
| `GridOption` | `grid.PadY` | `p L` | `-pady` |
| `GridOption` | `grid.PadYPair` | `top A, bottom B` | `-padypair` |
| `GridOption` | `grid.Row` | `r int` | `-row` |
| `GridOption` | `grid.RowSpan` | `n int` | `-rowspan` |
| `GridOption` | `grid.Sticky` | `s option.Sticky` | `-sticky` |
| `SlotOption` | `grid.MinSize` | `n int` | `-minsize` |
| `SlotOption` | `grid.Pad` | `n int` | `-pad` |
| `SlotOption` | `grid.Uniform` | `name string` | `-uniform` |
| `SlotOption` | `grid.Weight` | `n int` | `-weight` |

## geometry/pack

| Option type | Go option | Takes | Tk option |
|---|---|---|---|
| `PackOption` | `pack.After` | `sibling window.Windower` | `-after` |
| `PackOption` | `pack.Anchor` | `a option.Anchor` | `-anchor` |
| `PackOption` | `pack.Before` | `sibling window.Windower` | `-before` |
| `PackOption` | `pack.Expand` | `b bool` | `-expand` |
| `PackOption` | `pack.FillOpt` | `f Fill` | `-fill` |
| `PackOption` | `pack.IPadX` | `p L` | `-ipadx` |
| `PackOption` | `pack.IPadY` | `p L` | `-ipady` |
| `PackOption` | `pack.In` | `container window.Windower` | `-in` |
| `PackOption` | `pack.PadX` | `p L` | `-padx` |
| `PackOption` | `pack.PadXPair` | `left A, right B` | `-padxpair` |
| `PackOption` | `pack.PadY` | `p L` | `-pady` |
| `PackOption` | `pack.PadYPair` | `top A, bottom B` | `-padypair` |
| `PackOption` | `pack.SideOpt` | `s Side` | `-side` |

## geometry/place

| Option type | Go option | Takes | Tk option |
|---|---|---|---|
| `PlaceOption` | `place.Anchor` | `a option.Anchor` | `-anchor` |
| `PlaceOption` | `place.Height` | `v int` | `-height` |
| `PlaceOption` | `place.In` | `container window.Windower` | `-in` |
| `PlaceOption` | `place.RelHeight` | `v float64` | `-relheight` |
| `PlaceOption` | `place.RelWidth` | `v float64` | `-relwidth` |
| `PlaceOption` | `place.RelX` | `v float64` | `-relx` |
| `PlaceOption` | `place.RelY` | `v float64` | `-rely` |
| `PlaceOption` | `place.Width` | `v int` | `-width` |
| `PlaceOption` | `place.X` | `v int` | `-x` |
| `PlaceOption` | `place.Y` | `v int` | `-y` |

## ttk

| Option type | Go option | Takes | Tk option |
|---|---|---|---|
| `ButtonOption` | `ttk.ButtonCompound` | `c widget.Compound` | `-compound` |
| `ButtonOption` | `ttk.ButtonFont` | `name F` | `-font` |
| `ButtonOption` | `ttk.ButtonImage` | `img widget.WidgetImage` | `-image` |
| `ButtonOption` | `ttk.ButtonStyle` | `name string` | `-style` |
| `ButtonOption` | `ttk.ButtonText` | `s string` | `-text` |
| `ButtonOption` | `ttk.ButtonUnderline` | `i int` | `-underline` |
| `ButtonOption` | `ttk.ButtonWidth` | `n int` | `-width` |
| `CheckbuttonOption` | `ttk.CheckbuttonAlternate` | `` | `-alternate` |
| `CheckbuttonOption` | `ttk.CheckbuttonFont` | `name F` | `-font` |
| `CheckbuttonOption` | `ttk.CheckbuttonStyle` | `name string` | `-style` |
| `CheckbuttonOption` | `ttk.CheckbuttonText` | `s string` | `-text` |
| `CheckbuttonOption` | `ttk.CheckbuttonVar` | `v *widget.Variable[bool]` | `-var` |
| `ColumnOption` | `ttk.ColAnchor` | `a option.Anchor` | `-colanchor` |
| `ColumnOption` | `ttk.ColMinWidth` | `w L` | `-colminwidth` |
| `ColumnOption` | `ttk.ColSeparator` | `b bool` | `-colseparator` |
| `ColumnOption` | `ttk.ColStretch` | `b bool` | `-colstretch` |
| `ColumnOption` | `ttk.ColWidth` | `w L` | `-colwidth` |
| `ComboboxOption` | `ttk.ComboboxPlaceholder` | `s string` | `-placeholder` |
| `ComboboxOption` | `ttk.ComboboxState` | `s FieldState` | `-state` |
| `ComboboxOption` | `ttk.ComboboxStyle` | `name string` | `-style` |
| `ComboboxOption` | `ttk.ComboboxText` | `s string` | `-text` |
| `ComboboxOption` | `ttk.ComboboxValues` | `vals []string` | `-values` |
| `EntryOption` | `ttk.EntryExportSelection` | `on bool` | `-exportselection` |
| `EntryOption` | `ttk.EntryFont` | `name F` | `-font` |
| `EntryOption` | `ttk.EntryJustify` | `j option.Justify` | `-justify` |
| `EntryOption` | `ttk.EntryPlaceholder` | `s string` | `-placeholder` |
| `EntryOption` | `ttk.EntryShow` | `ch rune` | `-show` |
| `EntryOption` | `ttk.EntryState` | `s FieldState` | `-state` |
| `EntryOption` | `ttk.EntryStyle` | `name string` | `-style` |
| `EntryOption` | `ttk.EntryText` | `s string` | `-text` |
| `EntryOption` | `ttk.EntryTextVariable` | `v *widget.Variable[string]` | `-textvariable` |
| `EntryOption` | `ttk.EntryValidate` | `v ValidateMode` | `-validate` |
| `EntryOption` | `ttk.EntryWidth` | `n int` | `-width` |
| `FrameOption` | `ttk.FrameBackground` | `pixel uint64` | `-background` |
| `FrameOption` | `ttk.FrameBorderWidth` | `w L` | `-borderwidth` |
| `FrameOption` | `ttk.FrameHeight` | `v L` | `-height` |
| `FrameOption` | `ttk.FramePadding` | `p Padding` | `-padding` |
| `FrameOption` | `ttk.FrameRelief` | `r option.Relief` | `-relief` |
| `FrameOption` | `ttk.FrameStyle` | `name string` | `-style` |
| `FrameOption` | `ttk.FrameWidth` | `v L` | `-width` |
| `HeadingOption` | `ttk.HeadAnchor` | `a option.Anchor` | `-headanchor` |
| `HeadingOption` | `ttk.HeadText` | `text string` | `-headtext` |
| `ItemOption` | `ttk.ItemID` | `id string` | `-itemid` |
| `ItemOption` | `ttk.ItemImage` | `img widget.WidgetImage` | `-itemimage` |
| `ItemOption` | `ttk.ItemOpen` | `open bool` | `-itemopen` |
| `ItemOption` | `ttk.ItemTags` | `tags ...string` | `-itemtags` |
| `ItemOption` | `ttk.ItemText` | `text string` | `-itemtext` |
| `ItemOption` | `ttk.ItemValues` | `vals ...string` | `-itemvalues` |
| `LabelOption` | `ttk.LabelAnchor` | `a option.Anchor` | `-anchor` |
| `LabelOption` | `ttk.LabelCompound` | `c widget.Compound` | `-compound` |
| `LabelOption` | `ttk.LabelFont` | `name F` | `-font` |
| `LabelOption` | `ttk.LabelForeground` | `pixel uint64` | `-foreground` |
| `LabelOption` | `ttk.LabelImage` | `img widget.WidgetImage` | `-image` |
| `LabelOption` | `ttk.LabelJustify` | `j option.Justify` | `-justify` |
| `LabelOption` | `ttk.LabelPadding` | `spec string` | `-padding` |
| `LabelOption` | `ttk.LabelStyle` | `name string` | `-style` |
| `LabelOption` | `ttk.LabelText` | `s string` | `-text` |
| `LabelOption` | `ttk.LabelTextVariable` | `v *widget.Variable[string]` | `-textvariable` |
| `LabelOption` | `ttk.LabelWrapLength` | `v L` | `-wraplength` |
| `LabelframeOption` | `ttk.LabelframeBorderWidth` | `bw L` | `-frameborderwidth` |
| `LabelframeOption` | `ttk.LabelframeLabelWidget` | `w window.Windower` | `-framelabelwidget` |
| `LabelframeOption` | `ttk.LabelframePadding` | `spec string` | `-framepadding` |
| `LabelframeOption` | `ttk.LabelframeStyle` | `name string` | `-framestyle` |
| `LabelframeOption` | `ttk.LabelframeText` | `s string` | `-frametext` |
| `MenubuttonOption` | `ttk.MenubuttonCompound` | `c widget.Compound` | `-compound` |
| `MenubuttonOption` | `ttk.MenubuttonDirection` | `d Direction` | `-direction` |
| `MenubuttonOption` | `ttk.MenubuttonImage` | `img widget.WidgetImage` | `-image` |
| `MenubuttonOption` | `ttk.MenubuttonMenu` | `m *menu.Menu` | `-menu` |
| `MenubuttonOption` | `ttk.MenubuttonStyle` | `name string` | `-style` |
| `MenubuttonOption` | `ttk.MenubuttonText` | `s string` | `-text` |
| `NotebookOption` | `ttk.NotebookStyle` | `name string` | `-style` |
| `ProgressbarOption` | `ttk.ProgressbarLength` | `l L` | `-length` |
| `ProgressbarOption` | `ttk.ProgressbarMaximum` | `v float64` | `-maximum` |
| `ProgressbarOption` | `ttk.ProgressbarMode` | `m ProgressMode` | `-mode` |
| `ProgressbarOption` | `ttk.ProgressbarOrient` | `o Orientation` | `-orient` |
| `ProgressbarOption` | `ttk.ProgressbarStyle` | `name string` | `-style` |
| `ProgressbarOption` | `ttk.ProgressbarValue` | `v float64` | `-value` |
| `RadiobuttonOption` | `ttk.RadiobuttonAlternate` | `` | `-alternate` |
| `RadiobuttonOption` | `ttk.RadiobuttonStyle` | `name string` | `-style` |
| `RadiobuttonOption` | `ttk.RadiobuttonText` | `s string` | `-text` |
| `RadiobuttonOption` | `ttk.RadiobuttonValue` | `v string` | `-value` |
| `RadiobuttonOption` | `ttk.RadiobuttonVar` | `v *widget.Variable[string]` | `-var` |
| `ScaleOption` | `ttk.ScaleFrom` | `v float64` | `-from` |
| `ScaleOption` | `ttk.ScaleLength` | `v L` | `-length` |
| `ScaleOption` | `ttk.ScaleOrient` | `o Orientation` | `-orient` |
| `ScaleOption` | `ttk.ScaleStyle` | `name string` | `-style` |
| `ScaleOption` | `ttk.ScaleTo` | `v float64` | `-to` |
| `ScaleOption` | `ttk.ScaleValue` | `v float64` | `-value` |
| `ScaleOption` | `ttk.ScaleVariable` | `v *widget.Variable[float64]` | `-variable` |
| `ScrollbarOption` | `ttk.ScrollbarOrient` | `o Orient` | `-orient` |
| `ScrollbarOption` | `ttk.ScrollbarStyle` | `name string` | `-style` |
| `SeparatorOption` | `ttk.SeparatorOrient` | `o Orientation` | `-orient` |
| `SeparatorOption` | `ttk.SeparatorStyle` | `name string` | `-style` |
| `SizegripOption` | `ttk.SizegripStyle` | `name string` | `-style` |
| `SpinboxOption` | `ttk.SpinboxFormat` | `f string` | `-format` |
| `SpinboxOption` | `ttk.SpinboxFrom` | `v float64` | `-from` |
| `SpinboxOption` | `ttk.SpinboxIncrement` | `v float64` | `-increment` |
| `SpinboxOption` | `ttk.SpinboxStyle` | `name string` | `-style` |
| `SpinboxOption` | `ttk.SpinboxTo` | `v float64` | `-to` |
| `SpinboxOption` | `ttk.SpinboxValidate` | `v ValidateMode` | `-validate` |
| `SpinboxOption` | `ttk.SpinboxValues` | `v []string` | `-values` |
| `SpinboxOption` | `ttk.SpinboxWidth` | `w int` | `-width` |
| `SpinboxOption` | `ttk.SpinboxWrap` | `b bool` | `-wrap` |
| `TabOption` | `ttk.TabPadding` | `p Padding` | `-tabpadding` |
| `TabOption` | `ttk.TabState` | `s State` | `-tabstate` |
| `TabOption` | `ttk.TabText` | `s string` | `-tabtext` |
| `TabOption` | `ttk.TabUnderline` | `i int` | `-tabunderline` |
| `ToggleswitchOption` | `ttk.ToggleswitchStyle` | `name string` | `-style` |
| `ToggleswitchOption` | `ttk.ToggleswitchText` | `s string` | `-text` |
| `ToggleswitchOption` | `ttk.ToggleswitchVar` | `v *widget.Variable[bool]` | `-var` |
| `TreeviewOption` | `ttk.TreeviewColumns` | `ids ...string` | `-columns` |
| `TreeviewOption` | `ttk.TreeviewHeight` | `rows int` | `-height` |
| `TreeviewOption` | `ttk.TreeviewSelectMode` | `mode TreeSelectMode` | `-selectmode` |
| `TreeviewOption` | `ttk.TreeviewShow` | `parts ...string` | `-show` |
| `TreeviewOption` | `ttk.TreeviewStyle` | `name string` | `-style` |

## widget/button

| Option type | Go option | Takes | Tk option |
|---|---|---|---|
| `ButtonOption` | `button.Anchor` | `a option.Anchor` | `-anchor` |
| `ButtonOption` | `button.Background` | `name C` | `-background` |
| `ButtonOption` | `button.BorderWidth` | `w L` | `-borderwidth` |
| `ButtonOption` | `button.CompoundOpt` | `c widget.Compound` | `-compound` |
| `ButtonOption` | `button.Default` | `state DefaultState` | `-default` |
| `ButtonOption` | `button.FontOpt` | `name F` | `-font` |
| `ButtonOption` | `button.Foreground` | `name C` | `-foreground` |
| `ButtonOption` | `button.HighlightThickness` | `w L` | `-highlightthickness` |
| `ButtonOption` | `button.ImageOpt` | `img widget.WidgetImage` | `-image` |
| `ButtonOption` | `button.PadX` | `p L` | `-padx` |
| `ButtonOption` | `button.PadY` | `p L` | `-pady` |
| `ButtonOption` | `button.ReliefOpt` | `r option.Relief` | `-relief` |
| `ButtonOption` | `button.Text` | `s string` | `-text` |
| `ButtonOption` | `button.Width` | `n int` | `-width` |

## widget/checkbutton

| Option type | Go option | Takes | Tk option |
|---|---|---|---|
| `CheckbuttonOption` | `checkbutton.Anchor` | `a option.Anchor` | `-anchor` |
| `CheckbuttonOption` | `checkbutton.Background` | `name C` | `-background` |
| `CheckbuttonOption` | `checkbutton.BoolVar` | `v *widget.Variable[bool]` | `-boolvar` |
| `CheckbuttonOption` | `checkbutton.FontOpt` | `name F` | `-font` |
| `CheckbuttonOption` | `checkbutton.Foreground` | `name C` | `-foreground` |
| `CheckbuttonOption` | `checkbutton.ImageOpt` | `img widget.WidgetImage` | `-image` |
| `CheckbuttonOption` | `checkbutton.IndicatorOnOpt` | `on bool` | `-indicatoron` |
| `CheckbuttonOption` | `checkbutton.OffValueOpt` | `v string` | `-offvalue` |
| `CheckbuttonOption` | `checkbutton.OnValueOpt` | `v string` | `-onvalue` |
| `CheckbuttonOption` | `checkbutton.PadX` | `p L` | `-padx` |
| `CheckbuttonOption` | `checkbutton.PadY` | `p L` | `-pady` |
| `CheckbuttonOption` | `checkbutton.SelectColor` | `name C` | `-selectcolor` |
| `CheckbuttonOption` | `checkbutton.SelectImageOpt` | `img widget.WidgetImage` | `-selectimage` |
| `CheckbuttonOption` | `checkbutton.State` | `st widget.State` | `-state` |
| `CheckbuttonOption` | `checkbutton.Text` | `s string` | `-text` |
| `CheckbuttonOption` | `checkbutton.TristateValueOpt` | `v string` | `-tristatevalue` |
| `CheckbuttonOption` | `checkbutton.Var` | `v *widget.Variable[string]` | `-var` |

## widget/entry

| Option type | Go option | Takes | Tk option |
|---|---|---|---|
| `EntryOption` | `entry.Background` | `name C` | `-background` |
| `EntryOption` | `entry.BorderWidth` | `w L` | `-borderwidth` |
| `EntryOption` | `entry.FontOpt` | `name F` | `-font` |
| `EntryOption` | `entry.Foreground` | `name C` | `-foreground` |
| `EntryOption` | `entry.Placeholder` | `s string` | `-placeholder` |
| `EntryOption` | `entry.PlaceholderForeground` | `name C` | `-placeholderforeground` |
| `EntryOption` | `entry.Show` | `ch rune` | `-show` |
| `EntryOption` | `entry.Text` | `s string` | `-text` |
| `EntryOption` | `entry.ValidateOpt` | `v string` | `-validate` |
| `EntryOption` | `entry.Width` | `w int` | `-width` |

## widget/frame

| Option type | Go option | Takes | Tk option |
|---|---|---|---|
| `FrameOption` | `frame.Background` | `name C` | `-background` |
| `FrameOption` | `frame.BorderWidth` | `w L` | `-borderwidth` |
| `FrameOption` | `frame.Height` | `h int` | `-height` |
| `FrameOption` | `frame.HighlightThickness` | `n L` | `-highlightthickness` |
| `FrameOption` | `frame.Relief` | `r option.Relief` | `-relief` |
| `FrameOption` | `frame.Width` | `w int` | `-width` |

## widget/label

| Option type | Go option | Takes | Tk option |
|---|---|---|---|
| `LabelOption` | `label.Anchor` | `a option.Anchor` | `-anchor` |
| `LabelOption` | `label.Background` | `name C` | `-background` |
| `LabelOption` | `label.Bitmap` | `name string` | `-bitmap` |
| `LabelOption` | `label.BorderWidth` | `w L` | `-borderwidth` |
| `LabelOption` | `label.CompoundOpt` | `c widget.Compound` | `-compound` |
| `LabelOption` | `label.FontOpt` | `name F` | `-font` |
| `LabelOption` | `label.Foreground` | `name C` | `-foreground` |
| `LabelOption` | `label.Height` | `h int` | `-height` |
| `LabelOption` | `label.ImageOpt` | `img widget.WidgetImage` | `-image` |
| `LabelOption` | `label.JustifyOpt` | `j option.Justify` | `-justify` |
| `LabelOption` | `label.PadX` | `p L` | `-padx` |
| `LabelOption` | `label.PadY` | `p L` | `-pady` |
| `LabelOption` | `label.Relief` | `r option.Relief` | `-relief` |
| `LabelOption` | `label.Text` | `s string` | `-text` |
| `LabelOption` | `label.TextVariable` | `v *widget.Variable[string]` | `-textvariable` |
| `LabelOption` | `label.Width` | `w int` | `-width` |
| `LabelOption` | `label.WrapLength` | `w L` | `-wraplength` |

## widget/labelframe

| Option type | Go option | Takes | Tk option |
|---|---|---|---|
| `LabelframeOption` | `labelframe.Background` | `name C` | `-background` |
| `LabelframeOption` | `labelframe.BorderWidth` | `w L` | `-borderwidth` |
| `LabelframeOption` | `labelframe.FontOpt` | `name F` | `-font` |
| `LabelframeOption` | `labelframe.Foreground` | `name C` | `-foreground` |
| `LabelframeOption` | `labelframe.Height` | `h int` | `-height` |
| `LabelframeOption` | `labelframe.LabelAnchor` | `a option.Anchor` | `-labelanchor` |
| `LabelframeOption` | `labelframe.LabelWidgetOpt` | `w widget.Widget` | `-labelwidget` |
| `LabelframeOption` | `labelframe.PadX` | `p L` | `-padx` |
| `LabelframeOption` | `labelframe.PadY` | `p L` | `-pady` |
| `LabelframeOption` | `labelframe.Relief` | `r option.Relief` | `-relief` |
| `LabelframeOption` | `labelframe.Text` | `s string` | `-text` |
| `LabelframeOption` | `labelframe.Width` | `w int` | `-width` |

## widget/listbox

| Option type | Go option | Takes | Tk option |
|---|---|---|---|
| `ListboxOption` | `listbox.Background` | `name C` | `-background` |
| `ListboxOption` | `listbox.Foreground` | `name C` | `-foreground` |
| `ListboxOption` | `listbox.Height` | `h int` | `-height` |
| `ListboxOption` | `listbox.Items` | `items ...string` | `-items` |
| `ListboxOption` | `listbox.JustifyOpt` | `j option.Justify` | `-justify` |
| `ListboxOption` | `listbox.SelectModeOpt` | `m SelectMode` | `-selectmode` |

## widget/menu

| Option type | Go option | Takes | Tk option |
|---|---|---|---|
| `MenuOption` | `menu.Background` | `name C` | `-background` |
| `MenuOption` | `menu.FontOpt` | `name F` | `-font` |
| `MenuOption` | `menu.TearOffOpt` | `on bool` | `-tearoff` |

## widget/menubutton

| Option type | Go option | Takes | Tk option |
|---|---|---|---|
| `MenubuttonOption` | `menubutton.Background` | `name C` | `-background` |
| `MenubuttonOption` | `menubutton.DirectionOpt` | `d Direction` | `-direction` |
| `MenubuttonOption` | `menubutton.Foreground` | `name C` | `-foreground` |
| `MenubuttonOption` | `menubutton.IndicatorOnOpt` | `on bool` | `-indicatoron` |
| `MenubuttonOption` | `menubutton.PadX` | `p L` | `-padx` |
| `MenubuttonOption` | `menubutton.PadY` | `p L` | `-pady` |
| `MenubuttonOption` | `menubutton.Relief` | `r option.Relief` | `-relief` |

## widget/message

| Option type | Go option | Takes | Tk option |
|---|---|---|---|
| `MessageOption` | `message.Anchor` | `a option.Anchor` | `-anchor` |
| `MessageOption` | `message.Aspect` | `a int` | `-aspect` |
| `MessageOption` | `message.Background` | `name C` | `-background` |
| `MessageOption` | `message.BorderWidth` | `w L` | `-borderwidth` |
| `MessageOption` | `message.FontOpt` | `name F` | `-font` |
| `MessageOption` | `message.Foreground` | `name C` | `-foreground` |
| `MessageOption` | `message.HighlightWidth` | `w L` | `-highlightwidth` |
| `MessageOption` | `message.JustifyOpt` | `j option.Justify` | `-justify` |
| `MessageOption` | `message.PadX` | `p L` | `-padx` |
| `MessageOption` | `message.PadY` | `p L` | `-pady` |
| `MessageOption` | `message.Relief` | `r option.Relief` | `-relief` |
| `MessageOption` | `message.Text` | `s string` | `-text` |
| `MessageOption` | `message.WidthOpt` | `w L` | `-width` |

## widget/panedwindow

| Option type | Go option | Takes | Tk option |
|---|---|---|---|
| `PanedWindowOption` | `panedwindow.Background` | `name C` | `-background` |
| `PanedWindowOption` | `panedwindow.HandleSizeOpt` | `s L` | `-handlesize` |
| `PanedWindowOption` | `panedwindow.OrientOpt` | `o Orient` | `-orient` |
| `PanedWindowOption` | `panedwindow.SashWidthOpt` | `w L` | `-sashwidth` |

## widget/radiobutton

| Option type | Go option | Takes | Tk option |
|---|---|---|---|
| `RadiobuttonOption` | `radiobutton.Anchor` | `a option.Anchor` | `-anchor` |
| `RadiobuttonOption` | `radiobutton.Background` | `name C` | `-background` |
| `RadiobuttonOption` | `radiobutton.FontOpt` | `name F` | `-font` |
| `RadiobuttonOption` | `radiobutton.Foreground` | `name C` | `-foreground` |
| `RadiobuttonOption` | `radiobutton.ImageOpt` | `img widget.WidgetImage` | `-image` |
| `RadiobuttonOption` | `radiobutton.IndicatorOnOpt` | `on bool` | `-indicatoron` |
| `RadiobuttonOption` | `radiobutton.PadX` | `p L` | `-padx` |
| `RadiobuttonOption` | `radiobutton.PadY` | `p L` | `-pady` |
| `RadiobuttonOption` | `radiobutton.Text` | `s string` | `-text` |
| `RadiobuttonOption` | `radiobutton.TristateValueOpt` | `v string` | `-tristatevalue` |
| `RadiobuttonOption` | `radiobutton.Value` | `v string` | `-value` |
| `RadiobuttonOption` | `radiobutton.Var` | `v *widget.Variable[string]` | `-var` |
| `RadiobuttonOption` | `radiobutton.Width` | `n int` | `-width` |

## widget/scale

| Option type | Go option | Takes | Tk option |
|---|---|---|---|
| `ScaleOption` | `scale.Background` | `name C` | `-background` |
| `ScaleOption` | `scale.LengthOpt` | `n L` | `-length` |
| `ScaleOption` | `scale.ResolutionOpt` | `v float64` | `-resolution` |
| `ScaleOption` | `scale.SliderLengthOpt` | `n L` | `-sliderlength` |
| `ScaleOption` | `scale.TickIntervalOpt` | `v float64` | `-tickinterval` |
| `ScaleOption` | `scale.WidthOpt` | `w L` | `-width` |

## widget/scrollbar

| Option type | Go option | Takes | Tk option |
|---|---|---|---|
| `ScrollbarOption` | `scrollbar.OrientOpt` | `o Orient` | `-orient` |
| `ScrollbarOption` | `scrollbar.WidthOpt` | `w L` | `-width` |

## widget/spinbox

| Option type | Go option | Takes | Tk option |
|---|---|---|---|
| `SpinboxOption` | `spinbox.Background` | `name C` | `-background` |
| `SpinboxOption` | `spinbox.ButtonBackground` | `name C` | `-buttonbackground` |

## widget/text

| Option type | Go option | Takes | Tk option |
|---|---|---|---|
| `TagOption` | `text.TagBackground` | `name C` | `-tagbackground` |
| `TagOption` | `text.TagBgStipple` | `name string` | `-tagbgstipple` |
| `TagOption` | `text.TagBorderWidth` | `n L` | `-tagborderwidth` |
| `TagOption` | `text.TagFgStipple` | `name string` | `-tagfgstipple` |
| `TagOption` | `text.TagFont` | `name F` | `-tagfont` |
| `TagOption` | `text.TagForeground` | `name C` | `-tagforeground` |
| `TagOption` | `text.TagJustify` | `j option.Justify` | `-tagjustify` |
| `TagOption` | `text.TagLMargin1` | `pixels L` | `-taglmargin1` |
| `TagOption` | `text.TagLMargin2` | `pixels L` | `-taglmargin2` |
| `TagOption` | `text.TagOffset` | `pixels L` | `-tagoffset` |
| `TagOption` | `text.TagOverstrike` | `on bool` | `-tagoverstrike` |
| `TagOption` | `text.TagRMargin` | `pixels L` | `-tagrmargin` |
| `TagOption` | `text.TagRelief` | `r option.Relief` | `-tagrelief` |
| `TagOption` | `text.TagSpacing1` | `pixels L` | `-tagspacing1` |
| `TagOption` | `text.TagSpacing2` | `pixels L` | `-tagspacing2` |
| `TagOption` | `text.TagSpacing3` | `pixels L` | `-tagspacing3` |
| `TagOption` | `text.TagUnderline` | `on bool` | `-tagunderline` |
| `TextOption` | `text.Background` | `name C` | `-background` |
| `TextOption` | `text.BorderWidthOpt` | `w L` | `-borderwidth` |
| `TextOption` | `text.FontOpt` | `name F` | `-font` |
| `TextOption` | `text.Foreground` | `name C` | `-foreground` |
| `TextOption` | `text.Height` | `h int` | `-height` |
| `TextOption` | `text.HighlightThickness` | `n L` | `-highlightthickness` |
| `TextOption` | `text.InsertWidth` | `w L` | `-insertwidth` |
| `TextOption` | `text.PadXOpt` | `n L` | `-padx` |
| `TextOption` | `text.PadYOpt` | `n L` | `-pady` |
| `TextOption` | `text.ReadOnly` | `on bool` | `-readonly` |
| `TextOption` | `text.SetGridOpt` | `on bool` | `-setgrid` |
| `TextOption` | `text.TabWidth` | `w int` | `-tabwidth` |
| `TextOption` | `text.UndoOpt` | `enabled bool` | `-undo` |
| `TextOption` | `text.Width` | `w int` | `-width` |
| `TextOption` | `text.WrapModeOpt` | `mode WrapMode` | `-wrapmode` |

## widget/toplevel

| Option type | Go option | Takes | Tk option |
|---|---|---|---|
| `ToplevelOption` | `toplevel.Background` | `name C` | `-background` |
| `ToplevelOption` | `toplevel.Geometry` | `geom string` | `-geometry` |
| `ToplevelOption` | `toplevel.IconName` | `s string` | `-iconname` |
| `ToplevelOption` | `toplevel.IconPhoto` | `imgs ...image.Image` | `-iconphoto` |
| `ToplevelOption` | `toplevel.MinSize` | `w, h int` | `-minsize` |
| `ToplevelOption` | `toplevel.Resizable` | `w, h bool` | `-resizable` |
| `ToplevelOption` | `toplevel.Title` | `s string` | `-title` |
| `ToplevelOption` | `toplevel.TransientFor` | `parent window.Windower` | `-transientfor` |

