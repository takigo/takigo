package dialog

import (
	"github.com/takigo/takigo/widget"
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
	b := newFileBrowser(parent, browseDir, fileConfig{title: cfg.title, initialDir: cfg.initialDir})
	b.mustExist = cfg.mustExist
	paths, ok := b.run()
	if !ok || len(paths) == 0 {
		return "", false
	}
	return paths[0], true
}
