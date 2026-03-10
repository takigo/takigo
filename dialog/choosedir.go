package dialog

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
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

// ChooseDirectory displays a directory chooser dialog. Returns the selected
// directory path and true, or "" and false if cancelled.
func ChooseDirectory(parent widget.Caregiver, opts ...DirOption) (string, bool) {
	cfg := dirConfig{title: "Choose Directory", mustExist: true}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.initialDir == "" {
		cfg.initialDir, _ = os.Getwd()
	}

	app := parent.AppContext()
	currentDir := cfg.initialDir
	d := New(parent, cfg.title, 400, 350)

	// Current directory label.
	dirLabel := label.New(d.Content, "dirlabel",
		label.Text(currentDir),
		label.Anchor(3), // AnchorW
	)
	pack.Pack(dirLabel, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(5), pack.PadY(2))

	// Directory listbox.
	dirList := listbox.New(d.Content, "dirlist",
		listbox.Width(50),
		listbox.Height(15),
	)
	pack.Pack(dirList, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(5), pack.PadY(2))

	// Populate directory list.
	var populateList func()
	populateList = func() {
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
	}

	populateList()

	// navigateDir handles navigation into a selected subdirectory.
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

	// Double-click to navigate into a subdirectory.
	app.Dispatcher().Bind(dirList.Window().PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button != 1 {
			return
		}
		navigateDir()
	})

	// Enter key navigates into selected directory.
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
		return currentDir, true
	}
	return "", false
}
