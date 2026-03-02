package dialog

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/entry"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/listbox"
	"github.com/msorc/takigo/window"
)

// FileType describes a file type filter.
type FileType struct {
	Name    string // display name, e.g. "Go files"
	Pattern string // glob pattern, e.g. "*.go"
}

// fileConfig holds file dialog options.
type fileConfig struct {
	parent     *window.Window
	title      string
	initialDir string
	fileTypes  []FileType
	isSave     bool
}

// FileOption configures file dialogs.
type FileOption func(*fileConfig)

func FileParent(w *window.Window) FileOption   { return func(c *fileConfig) { c.parent = w } }
func FileTitle(s string) FileOption            { return func(c *fileConfig) { c.title = s } }
func FileInitialDir(s string) FileOption       { return func(c *fileConfig) { c.initialDir = s } }
func FileTypes(types ...FileType) FileOption   { return func(c *fileConfig) { c.fileTypes = types } }

// OpenFile displays a file open dialog. Returns the selected path and true,
// or "" and false if cancelled.
func OpenFile(app widget.AppContext, opts ...FileOption) (string, bool) {
	cfg := fileConfig{title: "Open File"}
	for _, opt := range opts {
		opt(&cfg)
	}
	return showFileDialog(app, cfg)
}

// SaveFile displays a file save dialog. Returns the selected path and true,
// or "" and false if cancelled.
func SaveFile(app widget.AppContext, opts ...FileOption) (string, bool) {
	cfg := fileConfig{title: "Save File", isSave: true}
	for _, opt := range opts {
		opt(&cfg)
	}
	return showFileDialog(app, cfg)
}

func showFileDialog(app widget.AppContext, cfg fileConfig) (string, bool) {
	if cfg.initialDir == "" {
		cfg.initialDir, _ = os.Getwd()
	}

	currentDir := cfg.initialDir
	d := New(app, cfg.parent, cfg.title, 450, 400)

	// Current directory label + Up button.
	navFrame := newFrame(d.Content.Window(), "nav", app)
	pack.Pack(navFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(2))

	dirLabel := label.New(navFrame.Window(), "dirlabel", app,
		label.Text(currentDir),
		label.Anchor(3), // AnchorW
	)
	pack.Pack(dirLabel.Window(), pack.SideOpt(pack.Left), pack.FillOpt(pack.FillX),
		pack.Expand(true), pack.PadX(5))

	// File listbox.
	fileList := listbox.New(d.Content.Window(), "filelist", app,
		listbox.Width(50),
		listbox.Height(15),
	)
	pack.Pack(fileList.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(5), pack.PadY(2))

	// Filename entry.
	entryFrame := newFrame(d.Content.Window(), "entryframe", app)
	pack.Pack(entryFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(2))

	fnLabel := label.New(entryFrame.Window(), "fnlabel", app, label.Text("File:"))
	pack.Pack(fnLabel.Window(), pack.SideOpt(pack.Left), pack.PadX(5))

	fnEntry := entry.New(entryFrame.Window(), "fnentry", app, entry.Width(40))
	pack.Pack(fnEntry.Window(), pack.SideOpt(pack.Left), pack.FillOpt(pack.FillX),
		pack.Expand(true), pack.PadX(5))

	_ = fnLabel

	// Populate file list.
	var populateList func()
	populateList = func() {
		entries, err := os.ReadDir(currentDir)
		if err != nil {
			return
		}

		// Clear current items.
		fileList.Delete(0, fileList.ItemCount()-1)

		var items []string
		// Parent directory.
		items = append(items, "../")

		// Directories first.
		for _, e := range entries {
			if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
				items = append(items, e.Name()+"/")
			}
		}

		// Then files, filtered by file type patterns.
		for _, e := range entries {
			if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
				continue
			}
			if matchesFileTypes(e.Name(), cfg.fileTypes) {
				items = append(items, e.Name())
			}
		}

		fileList.Insert(0, items...)
		dirLabel.Text = currentDir
		dirLabel.Display()
	}

	populateList()

	// Double-click on listbox to navigate/select.
	app.Dispatcher().Bind(fileList.Window().XWindow, event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button != 1 {
			return
		}
		// Check for double-click by looking at selection.
		sel := fileList.Selection()
		if len(sel) == 0 {
			return
		}
		items := fileList.GetItems()
		if sel[0] >= len(items) {
			return
		}
		item := items[sel[0]]
		if strings.HasSuffix(item, "/") {
			// Directory — navigate.
			dirName := strings.TrimSuffix(item, "/")
			if dirName == ".." {
				currentDir = filepath.Dir(currentDir)
			} else {
				currentDir = filepath.Join(currentDir, dirName)
			}
			populateList()
		} else {
			// File — put in entry.
			fnEntry.SetText(item)
		}
	})

	// Keyboard: Enter on listbox to navigate/select.
	app.Dispatcher().Bind(fileList.Window().XWindow, event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym != xlib.XK_Return {
			return
		}
		sel := fileList.Selection()
		if len(sel) == 0 {
			return
		}
		items := fileList.GetItems()
		if sel[0] >= len(items) {
			return
		}
		item := items[sel[0]]
		if strings.HasSuffix(item, "/") {
			dirName := strings.TrimSuffix(item, "/")
			if dirName == ".." {
				currentDir = filepath.Dir(currentDir)
			} else {
				currentDir = filepath.Join(currentDir, dirName)
			}
			populateList()
		} else {
			fnEntry.SetText(item)
		}
	})

	// Buttons.
	okText := "Open"
	if cfg.isSave {
		okText = "Save"
	}
	addButtons(d, []dialogButton{
		{text: okText, result: ResultOK, isDefault: true},
		{text: "Cancel", result: ResultCancel},
	})

	result := d.Run()
	if result == ResultOK {
		filename := fnEntry.GetText()
		if filename != "" {
			fullPath := filepath.Join(currentDir, filename)
			return fullPath, true
		}
	}
	return "", false
}

// matchesFileTypes checks if a filename matches any of the given file types.
// An empty list means accept all files.
func matchesFileTypes(name string, types []FileType) bool {
	if len(types) == 0 {
		return true
	}
	for _, ft := range types {
		if matched, _ := filepath.Match(ft.Pattern, name); matched {
			return true
		}
	}
	return false
}
