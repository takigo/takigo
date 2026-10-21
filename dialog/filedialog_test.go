package dialog

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/ttk"
	"github.com/takigo/takigo/widget"
)

func newTestBrowser(t *testing.T, mode browserMode, cfg fileConfig) (*fileBrowser, string) {
	t.Helper()
	app := testutil.NewTestApp(t)
	root := t.TempDir()
	for _, d := range []string{"a", "b", ".hidden"} {
		if err := os.Mkdir(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	cfg.initialDir = root
	b := newFileBrowser(app, mode, cfg)
	t.Cleanup(func() { b.d.Close(ResultCancel) })
	return b, root
}

func altKey(b *fileBrowser, target *ttk.Entry, ks platform.KeySym) {
	b.app.Dispatcher().Dispatch(&event.Event{
		Type:   event.KeyPressType,
		Window: target.Window().PlatformID,
		KeySym: ks,
		Str:    string(platform.KeySymToRune(ks)),
		State:  platform.Mod1Mask,
	})
}

func TestFileBrowserMnemonics(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("no Alt mnemonics on macOS")
	}
	b, _ := newTestBrowser(t, browseSave, fileConfig{title: "Save", confirmOverwrite: true})

	if b.okBtn.Text != "Save" || b.okBtn.Underline != 0 {
		t.Errorf("OK button = %q underline %d", b.okBtn.Text, b.okBtn.Underline)
	}

	widget.Focus(b.app, b.pathEntry.Window())
	altKey(b, b.pathEntry, platform.XK_n)
	if got := widget.FocusWindow(b.app); got != b.nameEntry.Window() {
		t.Errorf("Alt+N focused %v, want the name entry", got)
	}
	if b.nameEntry.Get() != "" {
		t.Errorf("Alt+N typed into the entry: %q", b.nameEntry.Get())
	}

	altKey(b, b.nameEntry, platform.XK_d)
	if got := widget.FocusWindow(b.app); got != b.pathEntry.Window() {
		t.Errorf("Alt+D focused %v, want the path entry", got)
	}

	n := len(b.entries)
	altKey(b, b.nameEntry, platform.XK_h)
	if !b.showHidden.Get() || len(b.entries) != n+1 {
		t.Errorf("Alt+H: showHidden %v, %d entries (was %d)", b.showHidden.Get(), len(b.entries), n)
	}

	// A selected directory turns Save into Open, and Alt+O follows.
	b.selectName("a")
	b.listBrowse()
	if b.okBtn.Text != "Open" || b.okBtn.Underline != 0 {
		t.Fatalf("OK button on a directory = %q underline %d", b.okBtn.Text, b.okBtn.Underline)
	}
	dir := b.dir
	altKey(b, b.nameEntry, platform.KeySym('o'))
	if b.dir != filepath.Join(dir, "a") {
		t.Errorf("Alt+O: dir %q, want %q", b.dir, filepath.Join(dir, "a"))
	}
}

func TestFileBrowserHistory(t *testing.T) {
	b, root := newTestBrowser(t, browseOpen, fileConfig{title: "Open"})
	disabled := func(btn *ttk.Button) bool { return btn.State&ttk.StateDisabled != 0 }
	if !disabled(b.backBtn) || !disabled(b.fwdBtn) {
		t.Fatal("history buttons enabled at the start")
	}

	a, ab := filepath.Join(root, "a"), filepath.Join(root, "b")
	b.load(a)
	b.load(ab)
	b.load(ab) // a refresh is not a new entry
	b.goBack()
	if b.dir != a || disabled(b.backBtn) || disabled(b.fwdBtn) {
		t.Errorf("back: dir %q back disabled %v forward disabled %v", b.dir, disabled(b.backBtn), disabled(b.fwdBtn))
	}
	b.goBack()
	if b.dir != root || !disabled(b.backBtn) {
		t.Errorf("back to start: dir %q back disabled %v", b.dir, disabled(b.backBtn))
	}
	b.goForward()
	if runtime.GOOS != "darwin" {
		altKey(b, b.nameEntry, platform.XK_Right)
	} else {
		b.goForward()
	}
	if b.dir != ab || !disabled(b.fwdBtn) {
		t.Errorf("forward: dir %q forward disabled %v", b.dir, disabled(b.fwdBtn))
	}

	// A directory that went away is reported and the history stays put.
	if err := os.Remove(a); err != nil {
		t.Fatal(err)
	}
	warned := false
	b.warnHook = func(string) { warned = true }
	b.goBack()
	if !warned || b.dir != ab || b.hist.pos != 2 {
		t.Errorf("back to a removed dir: warned %v dir %q pos %d", warned, b.dir, b.hist.pos)
	}
}

func TestFileBrowserSaveExtension(t *testing.T) {
	types := []FileType{{"Images", "*.png *.jpg"}, {"Text", "*.txt"}, {"All files", "*"}}
	b, root := newTestBrowser(t, browseSave, fileConfig{title: "Save", fileTypes: types})
	if b.defaultExt != ".png" {
		t.Fatalf("default extension %q, want .png from the first type", b.defaultExt)
	}
	if !widget.Classic(b.app) {
		for _, step := range []struct{ name, label, want string }{
			{"report.png", fileTypeLabel(types[1]), "report.txt"},
			{"notes.md", fileTypeLabel(types[0]), "notes.md"},
			{"report.txt", fileTypeLabel(types[2]), "report.txt"},
		} {
			b.nameEntry.Set(step.name)
			b.typeBox.Set(step.label)
			if got := b.nameEntry.Get(); got != step.want {
				t.Errorf("%q after choosing %q: entry %q, want %q", step.name, step.label, got, step.want)
			}
		}
	}
	b.typeBox.Set(fileTypeLabel(types[1]))
	b.nameEntry.Set("report")
	b.activateEntry()
	if want := filepath.Join(root, "report.txt"); len(b.result) != 1 || b.result[0] != want {
		t.Errorf("result %q, want %q", b.result, want)
	}
}

func TestFileBrowserExplicitExtension(t *testing.T) {
	b, root := newTestBrowser(t, browseSave, fileConfig{title: "Save", defaultExt: ".dat",
		fileTypes: []FileType{{"Images", "*.png"}, {"Text", "*.txt"}}})
	b.typeBox.Set("Text (*.txt)")
	b.nameEntry.Set("report")
	b.activateEntry()
	if want := filepath.Join(root, "report.dat"); len(b.result) != 1 || b.result[0] != want {
		t.Errorf("result %q, want %q", b.result, want)
	}
}

func TestFileBrowserTypeVariable(t *testing.T) {
	v := widget.NewVariable("")
	v.Set("Te")
	b, root := newTestBrowser(t, browseSave, fileConfig{title: "Save", typeVar: v,
		fileTypes: []FileType{{"Images", "*.png"}, {"Text", "*.txt"}}})
	if b.filterType != 1 || b.typeBox.Get() != "Text (*.txt)" || b.defaultExt != ".txt" {
		t.Fatalf("initial type %d %q ext %q, want Text", b.filterType, b.typeBox.Get(), b.defaultExt)
	}
	b.typeBox.Set("Images (*.png)")
	b.nameEntry.Set("pic")
	b.activateEntry()
	if want := filepath.Join(root, "pic.png"); len(b.result) != 1 || b.result[0] != want || v.Get() != "Images" {
		t.Errorf("result %q type %q, want %q Images", b.result, v.Get(), want)
	}
}
