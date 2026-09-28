package dialog

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/entry"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/listbox"
)

// dirConfig holds directory dialog options.
type dirConfig struct {
	title      string
	initialDir string
	mustExist  bool
}

// DirOption configures directory dialogs.
type DirOption func(*dirConfig)

func DirTitle(s string) DirOption      { return func(c *dirConfig) { c.title = s } }
func DirInitialDir(s string) DirOption { return func(c *dirConfig) { c.initialDir = s } }
func DirMustExist(b bool) DirOption    { return func(c *dirConfig) { c.mustExist = b } }

// ChooseDirectory displays a directory chooser dialog matching Tk's
// tk::dialog::file::chooseDir. Returns the selected directory path and true,
// or "" and false if cancelled.
func ChooseDirectory(parent widget.Caregiver, opts ...DirOption) (string, bool) {
	cfg := dirConfig{title: "Choose Directory"}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.initialDir == "" {
		cfg.initialDir, _ = os.Getwd()
	}

	app := parent.AppContext()
	currentDir := cfg.initialDir
	d := New(parent, cfg.title, 400, 350)

	// Current directory label (like Tk's path menubutton area).
	dirLabel := label.New(d.Content, "dirlabel",
		label.Text(currentDir),
		label.Anchor(option.AnchorW),
	)
	pack.Pack(dirLabel, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(5), pack.PadY(2))

	// Directory listbox (shows only directories, like Tk's IconList).
	dirList := listbox.New(d.Content, "dirlist",
		listbox.Width(50),
		listbox.Height(15),
	)
	pack.Pack(dirList, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(5), pack.PadY(2))

	// Selection entry (like Tk's f2.ent — "Selection:" entry at bottom).
	entryFrame := newFrame(d.Content, "entryframe")
	pack.Pack(entryFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(2))

	selLabel := label.New(entryFrame, "sellabel", label.Text("Selection:"))
	pack.Pack(selLabel, pack.SideOpt(pack.Left), pack.PadX(5))

	selEntry := entry.New(entryFrame, "selentry", entry.Width(40))
	selEntry.SetText(currentDir)
	pack.Pack(selEntry, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillX),
		pack.Expand(true), pack.PadX(5))

	// Populate directory list with subdirectories of currentDir.
	populateList := func() {
		entries, err := os.ReadDir(currentDir)
		if err != nil {
			return
		}
		n := dirList.ItemCount()
		if n > 0 {
			dirList.Delete(0, n-1)
		}
		var items []string
		items = append(items, "../")
		for _, e := range entries {
			if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
				items = append(items, e.Name()+"/")
			}
		}
		dirList.Insert(0, items...)
		dirLabel.Text = currentDir
		dirLabel.Display()
		// Update entry to reflect current directory.
		selEntry.SetText(currentDir)
	}

	populateList()

	// navigateDir enters the selected subdirectory.
	navigateDir := func() {
		sel := dirList.Selection()
		if len(sel) == 0 {
			return
		}
		items := dirList.GetItems()
		if sel[0] >= len(items) {
			return
		}
		item := items[sel[0]]
		if dirName, ok := strings.CutSuffix(item, "/"); ok {
			if dirName == ".." {
				currentDir = filepath.Dir(currentDir)
			} else {
				currentDir = filepath.Join(currentDir, dirName)
			}
			populateList()
		}
	}

	// Track double-click timing (like Tk's IconList DblClick).
	var lastClickTime time.Time
	var lastClickIdx int = -1

	// Single-click: update entry with selected directory path (Tk's ListBrowse).
	// Double-click: navigate into directory (Tk's DblClick → ListInvoke).
	app.Dispatcher().Bind(dirList.Window().PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button != 1 {
			return
		}
		sel := dirList.Selection()
		if len(sel) == 0 {
			return
		}
		items := dirList.GetItems()
		if sel[0] >= len(items) {
			return
		}

		now := time.Now()
		idx := sel[0]

		// Check for double-click: same item within 500ms.
		if idx == lastClickIdx && now.Sub(lastClickTime) < 500*time.Millisecond {
			lastClickIdx = -1
			navigateDir()
			return
		}

		lastClickTime = now
		lastClickIdx = idx

		// Single click: update entry with full path of selected directory.
		item := items[idx]
		if dirName, ok := strings.CutSuffix(item, "/"); ok {
			var path string
			if dirName == ".." {
				path = filepath.Dir(currentDir)
			} else {
				path = filepath.Join(currentDir, dirName)
			}
			selEntry.SetText(path)
		}
	})

	// Enter key on listbox navigates into selected directory.
	app.Dispatcher().Bind(dirList.Window().PlatformID, event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym != platform.XK_Return {
			return
		}
		navigateDir()
	})

	// Buttons.
	addButtons(d, []dialogButton{
		{text: "OK", result: ResultOK, isDefault: true},
		{text: "Cancel", result: ResultCancel},
	})

	result := d.Run()
	if result == ResultOK {
		// Use the entry value (like Tk's OkCmd).
		selectedDir := selEntry.GetText()
		if selectedDir == "" {
			selectedDir = currentDir
		}
		// Validate mustExist if set.
		if cfg.mustExist {
			info, err := os.Stat(selectedDir)
			if err != nil || !info.IsDir() {
				return "", false
			}
		}
		return filepath.Clean(selectedDir), true
	}
	return "", false
}
