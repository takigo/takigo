// File dialogs: a directory browser shared by OpenFile, SaveFile,
// OpenFiles and ChooseDirectory. It ports the behaviour of
// tk/library/tkfbox.tcl and choosedir.tcl (path resolution, filters,
// hidden files, completion, overwrite confirmation, Alt mnemonics through
// tk.tcl's AmpWidget and AltKeyInDialog) on a themed layout with a places
// sidebar, a details list and back/forward history.
package dialog

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/geometry/grid"
	"github.com/takigo/takigo/geometry/pack"
	tkimage "github.com/takigo/takigo/image"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/ttk"
	_ "github.com/takigo/takigo/ttk/defaulttheme" // registers the default theme
	"github.com/takigo/takigo/widget"
)

// FileType describes a file type filter. Pattern holds one or more glob
// patterns separated by spaces, e.g. "*.c *.h".
type FileType struct {
	Name    string // display name, e.g. "Go files"
	Pattern string // glob pattern(s), e.g. "*.go"
}

// fileConfig holds file dialog options.
type fileConfig struct {
	title            string
	initialDir       string
	initialFile      string
	defaultExt       string
	fileTypes        []FileType
	typeVar          *widget.Variable[string]
	isSave           bool
	multiple         bool
	confirmOverwrite bool
}

// FileOption configures file dialogs.
type FileOption func(*fileConfig)

func FileTitle(s string) FileOption          { return func(c *fileConfig) { c.title = s } }
func FileInitialDir(s string) FileOption     { return func(c *fileConfig) { c.initialDir = s } }
func FileTypes(types ...FileType) FileOption { return func(c *fileConfig) { c.fileTypes = types } }

// FileInitialFile presets the file name entry (Tk's -initialfile).
func FileInitialFile(s string) FileOption { return func(c *fileConfig) { c.initialFile = s } }

// FileDefaultExtension is appended to a typed name without an extension
// (Tk's -defaultextension), e.g. ".txt". Without it the extension is the
// first *.ext pattern of the selected file type.
func FileDefaultExtension(s string) FileOption {
	return func(c *fileConfig) { c.defaultExt = s }
}

// FileTypeVariable selects the initial file type by name (Tk's
// -typevariable): the first type whose name starts with v's value. On OK
// v receives the name of the type selected then.
func FileTypeVariable(v *widget.Variable[string]) FileOption {
	return func(c *fileConfig) { c.typeVar = v }
}

// FileConfirmOverwrite controls whether SaveFile asks before returning an
// existing file (Tk's -confirmoverwrite, on by default).
func FileConfirmOverwrite(b bool) FileOption {
	return func(c *fileConfig) { c.confirmOverwrite = b }
}

// OpenFile displays a file open dialog. Returns the selected path and true,
// or "" and false if cancelled.
func OpenFile(parent widget.Caregiver, opts ...FileOption) (string, bool) {
	cfg := fileConfig{title: "Open File"}
	for _, opt := range opts {
		opt(&cfg)
	}
	paths, ok := newFileBrowser(parent, browseOpen, cfg).run()
	if !ok || len(paths) == 0 {
		return "", false
	}
	return paths[0], true
}

// OpenFiles displays a file open dialog that accepts several files (Tk's
// -multiple). Returns the selected paths and true, or nil and false if
// cancelled.
func OpenFiles(parent widget.Caregiver, opts ...FileOption) ([]string, bool) {
	cfg := fileConfig{title: "Open Files", multiple: true}
	for _, opt := range opts {
		opt(&cfg)
	}
	return newFileBrowser(parent, browseOpen, cfg).run()
}

// SaveFile displays a file save dialog. Returns the selected path and true,
// or "" and false if cancelled.
func SaveFile(parent widget.Caregiver, opts ...FileOption) (string, bool) {
	cfg := fileConfig{title: "Save File", isSave: true, confirmOverwrite: true}
	for _, opt := range opts {
		opt(&cfg)
	}
	paths, ok := newFileBrowser(parent, browseSave, cfg).run()
	if !ok || len(paths) == 0 {
		return "", false
	}
	return paths[0], true
}

type browserMode int

const (
	browseOpen browserMode = iota
	browseSave
	browseDir
)

type dirEntry struct {
	name    string
	isDir   bool
	size    int64
	modTime time.Time
	id      string
}

type fileBrowser struct {
	d     *Dialog
	app   widget.AppContext
	mode  browserMode
	cfg   fileConfig
	icons *fileIcons

	dir        string
	filter     []string
	filterType int    // index into cfg.fileTypes; -1 for a typed pattern
	defaultExt string // cfg.defaultExt, else the selected type's extension
	showHidden *widget.Variable[bool]
	sortCol    string
	sortRev    bool
	entries    []*dirEntry
	byID       map[string]*dirEntry
	mustExist  bool

	hist       dirHistory
	histMoving bool

	backBtn   *ttk.Button
	fwdBtn    *ttk.Button
	pathEntry *ttk.Entry
	places    *ttk.Treeview
	placeDirs map[string]string
	list      *ttk.Treeview
	vscroll   *ttk.Scrollbar
	nameEntry *ttk.Entry
	typeBox   *ttk.Combobox
	hiddenBtn *ttk.Checkbutton
	okBtn     *ttk.Button
	status    *ttk.Label

	updating bool
	result   []string
	warnHook func(msg string) // replaces the warning box in tests

	// Type-ahead: printable keys typed in the list select the first
	// name with that prefix, as file managers do.
	typed     string
	typedTime platform.Timestamp
}

const typeAheadTimeout = 1000 // ms between keys of one prefix

const (
	placesWidth   = 150
	nameColWidth  = 270
	sizeColWidth  = 80
	dateColWidth  = 130
	listRows      = 14
	dialogPadding = 8
)

func newFileBrowser(parent widget.Caregiver, mode browserMode, cfg fileConfig) *fileBrowser {
	b := &fileBrowser{
		app:        parent.AppContext(),
		mode:       mode,
		cfg:        cfg,
		icons:      newFileIcons(),
		showHidden: widget.NewVariable(false),
		sortCol:    "#0",
		filter:     []string{"*"},
		filterType: -1,
		defaultExt: cfg.defaultExt,
		placeDirs:  map[string]string{},
	}
	if len(cfg.fileTypes) > 0 {
		b.applyType(b.initialType(), false)
	}

	b.d = New(parent, cfg.title, 680, 460)
	b.d.Resizable = true
	b.d.Toplevel.Window().OnDestroy(b.icons.destroy)

	b.buildToolbar()
	b.buildBody()
	b.buildForm()
	b.buildButtons()
	b.bindKeys()

	b.load(b.startDir())
	if cfg.initialFile != "" {
		b.nameEntry.Set(cfg.initialFile)
		b.selectName(cfg.initialFile)
	}
	return b
}

// startDir is the first directory shown: -initialdir, else the working
// directory, else the home directory.
func (b *fileBrowser) startDir() string {
	for _, dir := range []string{b.cfg.initialDir, ".", "~"} {
		if dir == "" {
			continue
		}
		dir = expandUser(dir)
		if abs, err := filepath.Abs(dir); err == nil {
			dir = abs
		}
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}
	return string(filepath.Separator)
}

func (b *fileBrowser) buildToolbar() {
	bar := ttk.NewFrame(b.d.Content, "toolbar",
		ttk.FramePadding(ttk.Padding{Left: dialogPadding, Top: dialogPadding, Right: dialogPadding, Bottom: 4}))
	pack.Pack(bar, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	b.backBtn = ttk.NewButton(bar, "back", ttk.ButtonImage(b.icons.back), ttk.ButtonStyle("Toolbutton"),
		ttk.ButtonCommand(b.goBack))
	pack.Pack(b.backBtn, pack.SideOpt(pack.Left))
	b.fwdBtn = ttk.NewButton(bar, "forward", ttk.ButtonImage(b.icons.forward), ttk.ButtonStyle("Toolbutton"),
		ttk.ButtonCommand(b.goForward))
	pack.Pack(b.fwdBtn, pack.SideOpt(pack.Left), pack.PadXPair(2, 6))
	up := ttk.NewButton(bar, "up", ttk.ButtonImage(b.icons.up), ttk.ButtonStyle("Toolbutton"),
		ttk.ButtonCommand(b.upDir))
	pack.Pack(up, pack.SideOpt(pack.Left))
	home := ttk.NewButton(bar, "home", ttk.ButtonImage(b.icons.home), ttk.ButtonStyle("Toolbutton"),
		ttk.ButtonCommand(func() {
			if h, err := os.UserHomeDir(); err == nil {
				b.load(h)
			}
		}))
	pack.Pack(home, pack.SideOpt(pack.Left), pack.PadX(2))

	newDir := ttk.NewButton(bar, "newfolder", ttk.ButtonText("New Folder"),
		ttk.ButtonImage(b.icons.newFolder), ttk.ButtonCompound(widget.CompoundLeft),
		ttk.ButtonStyle("Toolbutton"), ttk.ButtonCommand(b.newFolder))
	pack.Pack(newDir, pack.SideOpt(pack.Right))

	b.pathEntry = ttk.NewEntry(bar, "path", ttk.EntryWidth(40))
	pack.Pack(b.pathEntry, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillX), pack.Expand(true), pack.PadX(6))
	b.app.Dispatcher().Bind(b.pathEntry.Window().PlatformID, event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == platform.XK_Return {
			ev.Handled = true
			b.activatePath()
		}
	})
}

func (b *fileBrowser) buildBody() {
	body := ttk.NewFrame(b.d.Content, "body",
		ttk.FramePadding(ttk.Padding{Left: dialogPadding, Top: 2, Right: dialogPadding, Bottom: 2}))
	pack.Pack(body, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Places sidebar.
	b.places = ttk.NewTreeview(body, "places",
		ttk.TreeviewShow("tree"),
		ttk.TreeviewSelectMode(ttk.TreeSelectBrowse),
		ttk.TreeviewHeight(listRows))
	b.places.ColumnConfigure("#0", ttk.ColWidth(placesWidth))
	b.fillPlaces()
	b.places.OnSelect = func() {
		if b.updating {
			return
		}
		if sel := b.places.Selection(); len(sel) == 1 {
			if dir, ok := b.placeDirs[sel[0]]; ok {
				b.load(dir)
			}
		}
	}
	pack.Pack(b.places, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillY), pack.PadXPair(0, 6))

	// File list with its scrollbar.
	listFrame := ttk.NewFrame(body, "listframe")
	pack.Pack(listFrame, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	selMode := ttk.TreeSelectBrowse
	if b.cfg.multiple {
		selMode = ttk.TreeSelectExtended
	}
	cols := []string{"size", "modified"}
	if b.mode == browseDir {
		cols = []string{"modified"}
	}
	b.list = ttk.NewTreeview(listFrame, "files",
		ttk.TreeviewColumns(cols...),
		ttk.TreeviewShow("tree", "headings"),
		ttk.TreeviewSelectMode(selMode),
		ttk.TreeviewHeight(listRows))
	b.list.ColumnConfigure("#0", ttk.ColWidth(nameColWidth))
	b.list.HeadingConfigure("#0", ttk.HeadText("Name"), ttk.HeadAnchor(option.AnchorW),
		ttk.HeadCommand(func() { b.sortBy("#0") }))
	if b.mode != browseDir {
		b.list.ColumnConfigure("size", ttk.ColWidth(sizeColWidth), ttk.ColAnchor(option.AnchorE), ttk.ColStretch(false))
		b.list.HeadingConfigure("size", ttk.HeadText("Size"), ttk.HeadAnchor(option.AnchorE),
			ttk.HeadCommand(func() { b.sortBy("size") }))
	}
	b.list.ColumnConfigure("modified", ttk.ColWidth(dateColWidth), ttk.ColStretch(false))
	b.list.HeadingConfigure("modified", ttk.HeadText("Modified"), ttk.HeadAnchor(option.AnchorW),
		ttk.HeadCommand(func() { b.sortBy("modified") }))
	b.list.SetSortIndicator("#0", false)
	b.list.OnSelect = b.listBrowse
	b.list.OnDoubleClick = b.invoke
	b.app.Dispatcher().Bind(b.list.Window().PlatformID, event.KeyPressMask, func(ev *event.Event) {
		switch ev.KeySym {
		case platform.XK_Return:
			// Return on the selection invokes it; otherwise it is the OK button.
			if b.invokeSelection() {
				ev.Handled = true
			}
		case platform.XK_BackSpace:
			b.upDir()
		default:
			b.typeAhead(ev)
		}
	})

	b.vscroll = ttk.NewScrollbar(listFrame, "vsb",
		ttk.ScrollbarOrient(ttk.Vertical),
		ttk.ScrollbarCommand(widget.ScrollY(b.list)))
	b.list.YScrollCmd = b.vscroll.Set

	grid.Grid(b.list, grid.Row(0), grid.Column(0), grid.Sticky(grid.NSEW))
	grid.Grid(b.vscroll, grid.Row(0), grid.Column(1), grid.Sticky(grid.NS))
	grid.ColumnConfigure(listFrame, 0, grid.Weight(1))
	grid.RowConfigure(listFrame, 0, grid.Weight(1))
}

func (b *fileBrowser) fillPlaces() {
	add := func(label, dir string, img widget.WidgetImage) {
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			return
		}
		id := b.places.Insert("", -1, ttk.ItemText(label), ttk.ItemImage(img))
		b.placeDirs[id] = dir
	}
	if home, err := os.UserHomeDir(); err == nil {
		add("Home", home, b.icons.home)
		for _, sub := range []string{"Desktop", "Documents", "Downloads", "Pictures", "Music", "Videos"} {
			add(sub, filepath.Join(home, sub), b.icons.folder)
		}
	}
	add("File System", string(filepath.Separator), b.icons.drive)
	if wd, err := os.Getwd(); err == nil {
		add("Working Dir", wd, b.icons.folder)
	}
}

func (b *fileBrowser) buildForm() {
	form := ttk.NewFrame(b.d.Content, "form",
		ttk.FramePadding(ttk.Padding{Left: dialogPadding, Top: 4, Right: dialogPadding, Bottom: 0}))
	pack.Pack(form, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	nameCaption := "File &name:"
	switch {
	case b.mode == browseDir:
		nameCaption = "&Folder:"
	case b.cfg.multiple:
		nameCaption = "File &names:"
	}
	row := 0
	nameLab := b.ampLabel(form, "namelab", nameCaption)
	b.nameEntry = ttk.NewEntry(form, "name", ttk.EntryWidth(40))
	b.d.addMnemonic(nameLab, b.selectEntry)
	grid.Grid(nameLab, grid.Row(row), grid.Column(0), grid.Sticky(grid.EW), grid.PadY(3))
	grid.Grid(b.nameEntry, grid.Row(row), grid.Column(1), grid.Sticky(grid.EW), grid.PadX(6), grid.PadY(3))
	b.app.Dispatcher().Bind(b.nameEntry.Window().PlatformID, event.KeyPressMask, func(ev *event.Event) {
		switch ev.KeySym {
		case platform.XK_Return:
			ev.Handled = true
			b.activateEntry()
		case platform.XK_Tab:
			if ev.State&platform.ShiftMask == 0 {
				ev.Handled = true
				b.completeEntry()
			}
		}
	})
	row++

	if b.mode != browseDir && len(b.cfg.fileTypes) > 0 {
		typeLab := b.ampLabel(form, "typelab", "Files of &type:")
		values := make([]string, len(b.cfg.fileTypes))
		for i, ft := range b.cfg.fileTypes {
			values[i] = fileTypeLabel(ft)
		}
		b.typeBox = ttk.NewCombobox(form, "types",
			ttk.ComboboxValues(values),
			ttk.ComboboxText(values[b.filterType]),
			ttk.ComboboxState(ttk.FieldReadonly),
			ttk.ComboboxCommand(b.setFilter))
		b.d.addMnemonic(typeLab, func() { widget.Focus(b.app, b.typeBox.Window()) })
		grid.Grid(typeLab, grid.Row(row), grid.Column(0), grid.Sticky(grid.EW), grid.PadY(3))
		grid.Grid(b.typeBox, grid.Row(row), grid.Column(1), grid.Sticky(grid.EW), grid.PadX(6), grid.PadY(3))
		row++
	}

	extra := ttk.NewFrame(form, "extra")
	grid.Grid(extra, grid.Row(row), grid.Column(1), grid.Sticky(grid.EW), grid.PadX(6), grid.PadY(3))
	hiddenText, hiddenUnder := underlineAmpersand("Show &hidden files")
	if b.mode == browseDir {
		hiddenText, hiddenUnder = underlineAmpersand("Show &hidden folders")
	}
	b.hiddenBtn = ttk.NewCheckbutton(extra, "hidden",
		ttk.CheckbuttonText(hiddenText),
		ttk.CheckbuttonUnderline(hiddenUnder),
		ttk.CheckbuttonVar(b.showHidden),
		ttk.CheckbuttonCommand(func() { b.load(b.dir) }))
	pack.Pack(b.hiddenBtn, pack.SideOpt(pack.Left))
	b.d.addMnemonic(b.hiddenBtn, b.hiddenBtn.Toggle)
	b.status = ttk.NewLabel(extra, "status", ttk.LabelText(""), ttk.LabelAnchor(option.AnchorE))
	pack.Pack(b.status, pack.SideOpt(pack.Right))

	grid.ColumnConfigure(form, 1, grid.Weight(1))
}

// ampLabel ports ::tk::AmpWidget for a caption label.
func (b *fileBrowser) ampLabel(parent widget.Caregiver, name, caption string) *ttk.Label {
	text, under := underlineAmpersand(caption)
	return ttk.NewLabel(parent, name, ttk.LabelText(text), ttk.LabelUnderline(under),
		ttk.LabelAnchor(option.AnchorE))
}

func fileTypeLabel(ft FileType) string {
	pat := strings.Join(splitPatterns(ft.Pattern), ", ")
	if ft.Name == "" {
		return pat
	}
	return ft.Name + " (" + pat + ")"
}

func (b *fileBrowser) buildButtons() {
	cancelText, cancelUnder := underlineAmpersand("&Cancel")
	cancel := ttk.NewButton(b.d.BtnFrame, "cancel", ttk.ButtonText(cancelText), ttk.ButtonUnderline(cancelUnder),
		ttk.ButtonWidth(8), ttk.ButtonCommand(func() { b.d.Close(ResultCancel) }))
	pack.Pack(cancel, pack.SideOpt(pack.Right), pack.PadX(dialogPadding), pack.PadY(2))
	b.okBtn = ttk.NewButton(b.d.BtnFrame, "ok", ttk.ButtonWidth(8), ttk.ButtonCommand(b.okCmd))
	b.restoreOKText()
	pack.Pack(b.okBtn, pack.SideOpt(pack.Right), pack.PadY(2))
	b.d.addMnemonic(b.okBtn, b.okBtn.Invoke)
	b.d.addMnemonic(cancel, cancel.Invoke)
	b.d.defaultButton = b.nameEntry.Window()
}

func (b *fileBrowser) bindKeys() {
	// Return anywhere else in the dialog is the OK button.
	b.d.bindKeyEvent(platform.XK_Return, func(ev *event.Event) { b.okCmd() })
	b.d.bindKeyEvent(platform.XK_Up, func(ev *event.Event) {
		if ev.State&platform.Mod1Mask != 0 {
			b.upDir()
		}
	})
	b.d.bindKeyEvent(platform.XK_Left, func(ev *event.Event) {
		if ev.State&platform.Mod1Mask != 0 {
			b.goBack()
		}
	})
	b.d.bindKeyEvent(platform.XK_Right, func(ev *event.Event) {
		if ev.State&platform.Mod1Mask != 0 {
			b.goForward()
		}
	})
	b.d.bindKeyEvent(platform.XK_h, func(ev *event.Event) {
		if ev.State&platform.ControlMask != 0 {
			b.hiddenBtn.Toggle()
		}
	})
	b.d.bindKeyEvent(platform.XK_l, func(ev *event.Event) {
		if ev.State&platform.ControlMask != 0 {
			b.focusPath()
		}
	})
	// Tk labels the directory field "&Directory:"; the path entry has no
	// caption here, so Alt+D is bound directly.
	b.d.bindKeyEvent(platform.XK_d, func(ev *event.Event) {
		if ev.State&platform.Mod1Mask != 0 && ev.State&platform.ControlMask == 0 {
			ev.Handled = true
			b.focusPath()
		}
	})
}

func (b *fileBrowser) focusPath() {
	widget.Focus(b.app, b.pathEntry.Window())
	_ = b.pathEntry.Selection().Range("0", "end")
}

func (b *fileBrowser) run() ([]string, bool) {
	if b.d.Run() == ResultOK && len(b.result) > 0 {
		return b.result, true
	}
	return nil, false
}

// --- Directory listing ---

func (b *fileBrowser) load(dir string) {
	if dir == "" {
		return
	}
	dir = filepath.Clean(dir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		b.warn(fmt.Sprintf("Cannot change to the directory %q.\n%s", dir, errText(err)))
		if b.dir != "" {
			b.pathEntry.Set(b.dir)
		}
		return
	}
	b.dir = dir
	if !b.histMoving {
		b.hist.visit(dir)
	}
	b.updateHistoryButtons()
	b.pathEntry.Set(dir)
	if b.mode == browseDir {
		b.nameEntry.Set(dir)
	}

	b.entries = b.entries[:0]
	b.byID = map[string]*dirEntry{}
	showHidden := b.showHidden.Get()
	for _, e := range entries {
		name := e.Name()
		if !showHidden && strings.HasPrefix(name, ".") {
			continue
		}
		de := &dirEntry{name: name, isDir: e.IsDir()}
		info, ierr := e.Info()
		if e.Type()&os.ModeSymlink != 0 {
			if target, serr := os.Stat(filepath.Join(dir, name)); serr == nil {
				info, ierr, de.isDir = target, nil, target.IsDir()
			}
		}
		if ierr == nil {
			de.size, de.modTime = info.Size(), info.ModTime()
		}
		if !de.isDir {
			if b.mode == browseDir || !matchesPatterns(name, b.filter) {
				continue
			}
		}
		b.entries = append(b.entries, de)
	}
	b.sortEntries()

	b.updating = true
	b.list.Delete(b.list.Children("")...)
	nDirs, nFiles := 0, 0
	for _, de := range b.entries {
		img := b.icons.file
		var values []string
		mod := ""
		if !de.modTime.IsZero() {
			mod = de.modTime.Format("2006-01-02 15:04")
		}
		if de.isDir {
			img = b.icons.folder
			nDirs++
		} else {
			nFiles++
		}
		if b.mode == browseDir {
			values = []string{mod}
		} else {
			size := ""
			if !de.isDir {
				size = humanSize(de.size)
			}
			values = []string{size, mod}
		}
		de.id = b.list.Insert("", -1, ttk.ItemText(de.name), ttk.ItemImage(img), ttk.ItemValues(values...))
		b.byID[de.id] = de
	}
	b.list.YViewMoveTo(0)
	b.syncPlaces()
	b.updating = false

	switch b.mode {
	case browseDir:
		b.status.SetText(fmt.Sprintf("%d folders", nDirs))
	default:
		b.status.SetText(fmt.Sprintf("%d folders, %d files", nDirs, nFiles))
	}
}

// syncPlaces highlights the place that is the current directory.
func (b *fileBrowser) syncPlaces() {
	for id, dir := range b.placeDirs {
		if dir == b.dir {
			b.places.SelectionSet(id)
			return
		}
	}
	b.places.SelectionSet()
}

func (b *fileBrowser) sortEntries() {
	less := func(a, c *dirEntry) bool {
		switch b.sortCol {
		case "size":
			if a.size != c.size && !a.isDir {
				return a.size < c.size
			}
		case "modified":
			if !a.modTime.Equal(c.modTime) {
				return a.modTime.Before(c.modTime)
			}
		}
		return strings.ToLower(a.name) < strings.ToLower(c.name)
	}
	sort.SliceStable(b.entries, func(i, j int) bool {
		a, c := b.entries[i], b.entries[j]
		if a.isDir != c.isDir {
			return a.isDir
		}
		if b.sortRev {
			return less(c, a)
		}
		return less(a, c)
	})
}

func (b *fileBrowser) sortBy(col string) {
	if b.sortCol == col {
		b.sortRev = !b.sortRev
	} else {
		b.sortCol, b.sortRev = col, false
	}
	b.list.SetSortIndicator(col, b.sortRev)
	b.load(b.dir)
}

// initialType is the index of the first type whose name starts with the
// type variable's value, else 0.
func (b *fileBrowser) initialType() int {
	if v := b.cfg.typeVar; v != nil && v.IsSet() {
		for i, ft := range b.cfg.fileTypes {
			if strings.HasPrefix(ft.Name, v.Get()) {
				return i
			}
		}
	}
	return 0
}

// setFilter is the type box command; a value that is no type's label is a
// pattern the user typed, which keeps the default extension as Tk does.
func (b *fileBrowser) setFilter(value string) {
	for i, ft := range b.cfg.fileTypes {
		if fileTypeLabel(ft) == value {
			b.applyType(i, true)
			return
		}
	}
	b.filter, b.filterType = splitPatterns(value), -1
	if b.dir != "" {
		b.load(b.dir)
	}
}

// applyType ports SetFilter. With swap, a save dialog outside Classic mode
// also changes the typed name's extension from the old type to the new
// one, as GTK and Windows dialogs do.
func (b *fileBrowser) applyType(i int, swap bool) {
	old := b.filter
	b.filter, b.filterType = splitPatterns(b.cfg.fileTypes[i].Pattern), i
	if b.cfg.defaultExt == "" {
		b.defaultExt = typeExt(b.filter)
	}
	if swap && b.mode == browseSave && !b.cfg.multiple && !widget.Classic(b.app) {
		if name, ok := swapExt(b.nameEntry.Get(), old, typeExt(b.filter)); ok {
			b.nameEntry.Set(name)
		}
	}
	if b.dir != "" {
		b.load(b.dir)
	}
}

func (b *fileBrowser) selectName(name string) {
	for _, de := range b.entries {
		if de.name == name {
			b.list.SelectionSet(de.id)
			b.list.SetFocus(de.id)
			b.list.See(de.id)
			return
		}
	}
}

// --- Navigation ---

func (b *fileBrowser) goBack()    { b.moveInHistory(b.hist.back) }
func (b *fileBrowser) goForward() { b.moveInHistory(b.hist.forward) }

// moveInHistory loads the directory step returns; when it cannot be
// read the history stays where it was.
func (b *fileBrowser) moveInHistory(step func() (string, bool)) {
	pos := b.hist.pos
	dir, ok := step()
	if !ok {
		return
	}
	b.histMoving = true
	b.load(dir)
	b.histMoving = false
	if b.dir != dir {
		b.hist.pos = pos
		b.updateHistoryButtons()
	}
}

// updateHistoryButtons enables the back and forward buttons; a disabled
// one shows its greyed icon, since a ttk image does not dim by itself.
func (b *fileBrowser) updateHistoryButtons() {
	for _, bs := range []struct {
		btn      *ttk.Button
		on       bool
		img, off *tkimage.Photo
	}{
		{b.backBtn, b.hist.canBack(), b.icons.back, b.icons.backOff},
		{b.fwdBtn, b.hist.canForward(), b.icons.forward, b.icons.forwardOff},
	} {
		if bs.on {
			bs.btn.ChangeState(0, ttk.StateDisabled)
			_ = bs.btn.Configure(ttk.ButtonImage(bs.img))
		} else {
			bs.btn.ChangeState(ttk.StateDisabled, 0)
			_ = bs.btn.Configure(ttk.ButtonImage(bs.off))
		}
	}
}

func (b *fileBrowser) upDir() {
	if parent := filepath.Dir(b.dir); parent != b.dir {
		b.load(parent)
	}
}

func (b *fileBrowser) activatePath() {
	text := expandUser(strings.TrimSpace(b.pathEntry.Get()))
	if !filepath.IsAbs(text) {
		text = filepath.Join(b.dir, text)
	}
	if info, err := os.Stat(text); err != nil || !info.IsDir() {
		b.warn(fmt.Sprintf("Directory %q does not exist.", text))
		b.pathEntry.Set(b.dir)
		return
	}
	b.load(text)
}

func (b *fileBrowser) newFolder() {
	name, ok := AskString(b.d.Toplevel, "New Folder", "Name of the new folder:", "New Folder")
	if !ok || strings.TrimSpace(name) == "" {
		return
	}
	name = strings.TrimSpace(name)
	if err := os.Mkdir(filepath.Join(b.dir, name), 0o777); err != nil {
		b.warn(fmt.Sprintf("Cannot create folder %q.\n%s", name, errText(err)))
		return
	}
	b.load(b.dir)
	b.selectName(name)
	if b.mode == browseDir {
		b.nameEntry.Set(filepath.Join(b.dir, name))
	}
}

// listBrowse ports ListBrowse: a selected file goes to the entry; a
// selected directory turns the button into "Open".
func (b *fileBrowser) listBrowse() {
	if b.updating {
		return
	}
	sel := b.list.Selection()
	if len(sel) == 0 {
		return
	}
	if b.mode == browseDir {
		if de := b.byID[sel[0]]; de != nil {
			b.nameEntry.Set(filepath.Join(b.dir, de.name))
		}
		return
	}
	if b.cfg.multiple {
		var names []string
		for _, id := range sel {
			if de := b.byID[id]; de != nil && !de.isDir {
				names = append(names, de.name)
			}
		}
		if len(names) > 0 {
			b.nameEntry.Set(quoteNames(names))
		}
		return
	}
	de := b.byID[sel[0]]
	if de == nil {
		return
	}
	if de.isDir {
		b.setOKText("&Open")
		return
	}
	b.nameEntry.Set(de.name)
	b.restoreOKText()
}

// setOKText ports ::tk::SetAmpText for the OK button.
func (b *fileBrowser) setOKText(caption string) {
	text, under := underlineAmpersand(caption)
	if b.okBtn.Text != text || b.okBtn.Underline != under {
		_ = b.okBtn.Configure(ttk.ButtonText(text), ttk.ButtonUnderline(under))
	}
}

func (b *fileBrowser) restoreOKText() {
	switch b.mode {
	case browseSave:
		b.setOKText("&Save")
	case browseDir:
		b.setOKText("&Select")
	default:
		b.setOKText("&Open")
	}
}

// invoke ports ListInvoke: enter a directory or accept a file.
func (b *fileBrowser) invoke(id string) {
	de := b.byID[id]
	if de == nil {
		return
	}
	path := filepath.Join(b.dir, de.name)
	if de.isDir {
		b.load(path)
		return
	}
	if b.mode != browseDir {
		b.done([]string{path})
	}
}

// invokeSelection ports ListInvoke for the keyboard: a single selected
// directory is entered, selected files are returned (all of them in a
// multiple-selection dialog). It reports whether there was a selection.
func (b *fileBrowser) invokeSelection() bool {
	sel := b.list.Selection()
	if len(sel) == 0 {
		if id := b.list.Focus(); id != "" {
			sel = []string{id}
		}
	}
	if len(sel) == 0 {
		return false
	}
	if !b.cfg.multiple || len(sel) == 1 {
		b.invoke(sel[0])
		return true
	}
	var paths []string
	for _, id := range sel {
		if de := b.byID[id]; de != nil && !de.isDir {
			paths = append(paths, filepath.Join(b.dir, de.name))
		}
	}
	if len(paths) > 0 {
		b.done(paths)
	}
	return true
}

// okCmd ports OkCmd: a selected directory is entered, otherwise the
// entry text is resolved.
func (b *fileBrowser) okCmd() {
	sel := b.list.Selection()
	if len(sel) == 1 && b.mode != browseDir {
		if de := b.byID[sel[0]]; de != nil && de.isDir {
			b.load(filepath.Join(b.dir, de.name))
			return
		}
	}
	b.activateEntry()
}

// activateEntry ports ActivateEnt and chooseDir::OkCmd.
func (b *fileBrowser) activateEntry() {
	text := b.nameEntry.Get()
	if b.mode == browseDir {
		b.acceptDir(text)
		return
	}
	if strings.TrimSpace(text) == "" {
		return
	}
	if !b.cfg.multiple {
		if path, ok := b.verifyName(text); ok && path != "" {
			b.done([]string{path})
		}
		return
	}
	var paths []string
	for _, name := range parseNameList(text) {
		path, ok := b.verifyName(name)
		if !ok {
			return
		}
		if path != "" {
			paths = append(paths, path)
		}
	}
	if len(paths) > 0 {
		b.done(paths)
	}
}

// verifyName ports VerifyFileName. It returns the resolved path when the
// name denotes a file to return, "" and true when the dialog navigated
// instead, and false after reporting an error.
func (b *fileBrowser) verifyName(text string) (string, bool) {
	flag, dir, file := resolveFile(b.dir, text, b.defaultExt)
	path := filepath.Join(dir, file)
	switch flag {
	case resolveOK:
		if file == "" {
			b.load(dir)
			b.nameEntry.Set("")
			return "", true
		}
		return path, true
	case resolvePattern:
		b.filter = splitPatterns(file)
		if b.typeBox != nil {
			b.typeBox.Set(file)
		} else {
			b.load(dir)
		}
		return "", true
	case resolveNewFile:
		if b.mode == browseOpen {
			b.warn(fmt.Sprintf("File %q does not exist.", path))
			b.selectEntry()
			return "", false
		}
		return path, true
	default:
		b.warn(fmt.Sprintf("Directory %q does not exist.", dir))
		b.selectEntry()
		return "", false
	}
}

func (b *fileBrowser) acceptDir(text string) {
	if strings.TrimSpace(text) == "" {
		text = b.dir
	}
	flag, dir, file := resolveFile(b.dir, text, "")
	path := filepath.Join(dir, file)
	switch {
	case flag == resolveOK && file == "":
		b.finish([]string{dir})
	case flag == resolveOK:
		b.warn(fmt.Sprintf("%q is not a directory.", path))
		b.selectEntry()
	case b.mustExist:
		b.warn(fmt.Sprintf("Directory %q does not exist.", path))
		b.selectEntry()
	default:
		b.finish([]string{path})
	}
}

func (b *fileBrowser) selectEntry() {
	widget.Focus(b.app, b.nameEntry.Window())
	_ = b.nameEntry.Selection().Range("0", "end")
	_ = b.nameEntry.ICursor("end")
}

// completeEntry ports CompleteEnt: Tab completes the typed name against
// the listed files and directories.
func (b *fileBrowser) completeEntry() {
	text := b.nameEntry.Get()
	if b.cfg.multiple {
		names := parseNameList(text)
		if len(names) != 1 {
			return
		}
		text = names[0]
	}
	prefix := text
	dir := b.dir
	if i := strings.LastIndex(text, string(filepath.Separator)); i >= 0 {
		_, dir, _ = resolveFile(b.dir, text[:i+1], "")
		prefix = text[i+1:]
	}
	var candidates []string
	if dir == b.dir {
		for _, de := range b.entries {
			candidates = append(candidates, entryCompletion(de.name, de.isDir))
		}
	} else if entries, err := os.ReadDir(dir); err == nil {
		for _, e := range entries {
			if !b.showHidden.Get() && strings.HasPrefix(e.Name(), ".") {
				continue
			}
			if !e.IsDir() && !matchesPatterns(e.Name(), b.filter) {
				continue
			}
			candidates = append(candidates, entryCompletion(e.Name(), e.IsDir()))
		}
	}
	completed, ok := completeName(prefix, candidates)
	if !ok {
		return
	}
	newText := text[:len(text)-len(prefix)] + completed
	if b.cfg.multiple {
		newText = quoteNames([]string{newText})
	}
	b.nameEntry.Set(newText)
	_ = b.nameEntry.ICursor("end")
}

func (b *fileBrowser) typeAhead(ev *event.Event) {
	if ev.State&(platform.ControlMask|platform.Mod1Mask) != 0 {
		return
	}
	s := ev.Str
	if s == "" {
		if r := platform.KeySymToRune(ev.KeySym); r > 0 {
			s = string(r)
		}
	}
	if s == "" || s[0] < 32 {
		return
	}
	if ev.Time-b.typedTime > typeAheadTimeout {
		b.typed = ""
	}
	b.typedTime = ev.Time
	b.typed += strings.ToLower(s)
	for _, de := range b.entries {
		if strings.HasPrefix(strings.ToLower(de.name), b.typed) {
			b.list.SelectionSet(de.id)
			b.list.SetFocus(de.id)
			b.list.See(de.id)
			b.listBrowse()
			return
		}
	}
}

func entryCompletion(name string, isDir bool) string {
	if isDir {
		return name + string(filepath.Separator)
	}
	return name
}

// done ports Done: confirm overwriting on save, then return the paths.
func (b *fileBrowser) done(paths []string) {
	if b.mode == browseSave && b.cfg.confirmOverwrite {
		if _, err := os.Stat(paths[0]); err == nil {
			reply := ShowMessage(b.d.Toplevel,
				MsgTitle("Confirm Save"),
				MsgType(MsgWarning),
				MsgButtons(BtnYesNo),
				MsgMessage(fmt.Sprintf("File %q already exists.\nDo you want to overwrite it?", paths[0])))
			if reply != ResultYes {
				return
			}
		}
	}
	if b.cfg.typeVar != nil && b.filterType >= 0 {
		b.cfg.typeVar.Set(b.cfg.fileTypes[b.filterType].Name)
	}
	b.finish(paths)
}

func (b *fileBrowser) finish(paths []string) {
	b.result = paths
	b.d.Close(ResultOK)
}

func (b *fileBrowser) warn(msg string) {
	if b.warnHook != nil {
		b.warnHook(msg)
		return
	}
	ShowMessage(b.d.Toplevel, MsgTitle(b.cfg.title), MsgType(MsgWarning), MsgMessage(msg))
}

func errText(err error) string {
	if pe, ok := errors.AsType[*os.PathError](err); ok {
		err = pe.Err
	}
	s := err.Error()
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:] + "."
}
