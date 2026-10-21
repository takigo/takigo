package dialog

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestSplitPatterns(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"", []string{"*"}},
		{"*.go", []string{"*.go"}},
		{"*.c  *.h", []string{"*.c", "*.h"}},
	}
	for _, tt := range tests {
		got := splitPatterns(tt.in)
		if len(got) != len(tt.want) {
			t.Fatalf("splitPatterns(%q) = %v, want %v", tt.in, got, tt.want)
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("splitPatterns(%q) = %v, want %v", tt.in, got, tt.want)
			}
		}
	}
}

func TestMatchesPatterns(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		want     bool
	}{
		{"main.go", nil, true},
		{"main.go", []string{"*"}, true},
		{"main.go", []string{"*.go"}, true},
		{"main.c", []string{"*.go"}, false},
		{"main.h", []string{"*.c", "*.h"}, true},
		{"README", []string{"*.c", "*.h"}, false},
	}
	for _, tt := range tests {
		if got := matchesPatterns(tt.name, tt.patterns); got != tt.want {
			t.Errorf("matchesPatterns(%q, %v) = %v, want %v", tt.name, tt.patterns, got, tt.want)
		}
	}
}

func TestHumanSize(t *testing.T) {
	tests := []struct {
		n    int64
		want string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{10 * 1024, "10 KB"},
		{5 * 1024 * 1024, "5.0 MB"},
		{3 * 1024 * 1024 * 1024, "3.0 GB"},
	}
	for _, tt := range tests {
		if got := humanSize(tt.n); got != tt.want {
			t.Errorf("humanSize(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestExpandUser(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home directory")
	}
	tests := []struct{ in, want string }{
		{"", ""},
		{"plain", "plain"},
		{"~", home},
		{"~/x/y", filepath.Join(home, "x", "y")},
		{"~nosuchuser_takigo/x", "~nosuchuser_takigo/x"},
	}
	for _, tt := range tests {
		if got := expandUser(tt.in); got != tt.want {
			t.Errorf("expandUser(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestResolveFile(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"a.txt", "sub/b.txt"} {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("TAKIGO_TEST_DIR", sub)

	tests := []struct {
		text, ext string
		flag      resolveFlag
		dir, file string
	}{
		{"a.txt", "", resolveOK, dir, "a.txt"},
		{"a", ".txt", resolveOK, dir, "a.txt"},
		{"sub", "", resolveOK, sub, ""},
		{"sub", ".txt", resolveOK, sub, ""},
		{"sub/", ".txt", resolveOK, sub, ""},
		{"sub/b.txt", "", resolveOK, sub, "b.txt"},
		{sub, "", resolveOK, sub, ""},
		{"..", "", resolveOK, filepath.Dir(dir), ""},
		{"new.txt", "", resolveNewFile, dir, "new.txt"},
		{"new", ".txt", resolveNewFile, dir, "new.txt"},
		{"*.txt", "", resolvePattern, dir, "*.txt"},
		{"sub/*.go", ".txt", resolvePattern, sub, "*.go"},
		{"missing/x.txt", "", resolvePath, filepath.Join(dir, "missing"), "x.txt"},
		{"$TAKIGO_TEST_DIR", "", resolveOK, sub, ""},
	}
	for _, tt := range tests {
		flag, gotDir, gotFile := resolveFile(dir, tt.text, tt.ext)
		if flag != tt.flag || gotDir != tt.dir || gotFile != tt.file {
			t.Errorf("resolveFile(%q, %q) = (%d, %q, %q), want (%d, %q, %q)",
				tt.text, tt.ext, flag, gotDir, gotFile, tt.flag, tt.dir, tt.file)
		}
	}
}

func TestCompleteName(t *testing.T) {
	candidates := []string{"main.go", "main_test.go", "makefile", "docs/"}
	tests := []struct {
		prefix string
		want   string
		ok     bool
	}{
		{"d", "docs/", true},
		{"ma", "ma", false},
		{"mai", "main", true},
		{"main", "main", false},
		{"main_", "main_test.go", true},
		{"zzz", "zzz", false},
		{"docs/", "docs/", false},
	}
	for _, tt := range tests {
		got, ok := completeName(tt.prefix, candidates)
		if got != tt.want || ok != tt.ok {
			t.Errorf("completeName(%q) = (%q, %v), want (%q, %v)", tt.prefix, got, ok, tt.want, tt.ok)
		}
	}
}

func TestNameListRoundTrip(t *testing.T) {
	names := []string{"a.txt", "my file.txt", "c"}
	text := quoteNames(names)
	if text != `a.txt "my file.txt" c` {
		t.Errorf("quoteNames = %q", text)
	}
	got := parseNameList(text)
	if len(got) != len(names) {
		t.Fatalf("parseNameList(%q) = %v", text, got)
	}
	for i := range names {
		if got[i] != names[i] {
			t.Errorf("parseNameList(%q)[%d] = %q, want %q", text, i, got[i], names[i])
		}
	}
	if got := parseNameList("   "); len(got) != 0 {
		t.Errorf("parseNameList(blank) = %v", got)
	}
}

func TestNameListRoundTripSpecialNames(t *testing.T) {
	names := []string{`a"b`, `a\b`, "my file.txt", `c"d e`, "plain", `trailing\`}
	text := quoteNames(names)
	got := parseNameList(text)
	if len(got) != len(names) {
		t.Fatalf("parseNameList(%q) = %v, want %v", text, got, names)
	}
	for i := range names {
		if got[i] != names[i] {
			t.Errorf("parseNameList(%q)[%d] = %q, want %q", text, i, got[i], names[i])
		}
	}
	if want := `"a\"b"`; !strings.Contains(text, want) {
		t.Errorf("quoteNames(%q) = %q, want it to contain %q", names, text, want)
	}
}

func TestUnderlineAmpersand(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in    string
		text  string
		under int
	}{
		{"&Open", "Open", 0},
		{"File &name:", "File name:", 5},
		{"a&&b", "a&b", -1},
		{"a&&&b", "a&b", 2},
		{"none", "none", -1},
		{"trailing&", "trailing", -1},
		{"&Über", "Über", 0},
		{"Fichier &à ouvrir", "Fichier à ouvrir", 8},
	}
	for _, tt := range tests {
		text, under := underlineAmpersand(tt.in)
		if text != tt.text || under != tt.under {
			t.Errorf("underlineAmpersand(%q) = %q, %d; want %q, %d", tt.in, text, under, tt.text, tt.under)
		}
	}
}

type fakeUnderliner struct {
	text  string
	under int
}

func (f *fakeUnderliner) GetText() string   { return f.text }
func (f *fakeUnderliner) GetUnderline() int { return f.under }

func TestFindMnemonic(t *testing.T) {
	t.Parallel()
	var got string
	ok := &fakeUnderliner{"Save", 0}
	d := &Dialog{}
	d.addMnemonic(&fakeUnderliner{"File name:", 5}, func() { got = "name" })
	d.addMnemonic(ok, func() { got = "ok" })
	d.addMnemonic(&fakeUnderliner{"plain", -1}, func() { got = "plain" })

	for _, tt := range []struct {
		r    rune
		want string
	}{{'n', "name"}, {'N', "name"}, {'s', "ok"}, {'p', ""}, {'x', ""}} {
		got = ""
		if fn := d.findMnemonic(tt.r); fn != nil {
			fn()
		}
		if got != tt.want {
			t.Errorf("findMnemonic(%q) ran %q, want %q", tt.r, got, tt.want)
		}
	}
	// The match follows the widget's current text, as when the OK button
	// turns into "Open" on a selected directory.
	ok.text = "Open"
	if d.findMnemonic('s') != nil || d.findMnemonic('o') == nil {
		t.Error("relabelled button: mnemonic did not follow its text")
	}
}

func TestDirHistory(t *testing.T) {
	t.Parallel()
	var h dirHistory
	if h.canBack() || h.canForward() {
		t.Fatal("empty history can move")
	}
	h.visit("/a")
	h.visit("/a") // refresh: no new entry
	h.visit("/b")
	h.visit("/c")
	if !h.canBack() || h.canForward() || len(h.dirs) != 3 {
		t.Fatalf("after visits: %v pos %d", h.dirs, h.pos)
	}
	if d, ok := h.back(); !ok || d != "/b" {
		t.Errorf("back = %q, %v", d, ok)
	}
	if d, ok := h.back(); !ok || d != "/a" {
		t.Errorf("back = %q, %v", d, ok)
	}
	if _, ok := h.back(); ok {
		t.Error("back past the start")
	}
	if d, ok := h.forward(); !ok || d != "/b" {
		t.Errorf("forward = %q, %v", d, ok)
	}
	h.visit("/d") // drops /c
	if h.canForward() || !slices.Equal(h.dirs, []string{"/a", "/b", "/d"}) {
		t.Errorf("visit after back: %v pos %d", h.dirs, h.pos)
	}
}

func TestParseNameListTrailingBackslash(t *testing.T) {
	got := parseNameList(`"C:\My Files\" b`)
	if len(got) != 2 || got[0] != `C:\My Files\` || got[1] != "b" {
		t.Errorf("got %q", got)
	}
}

func TestTypeExt(t *testing.T) {
	tests := []struct{ pattern, want string }{
		{"*.png *.jpg", ".png"},
		{"*", ""},
		{"*.[ch] *.go", ".go"},
		{"*.tar.gz *.tgz", ".tar.gz"},
		{"Makefile *.mk", ".mk"},
		{"*. *.a..b", ""},
	}
	for _, tt := range tests {
		if got := typeExt(splitPatterns(tt.pattern)); got != tt.want {
			t.Errorf("typeExt(%q) = %q, want %q", tt.pattern, got, tt.want)
		}
	}
}

func TestSwapExt(t *testing.T) {
	tests := []struct {
		name, old, ext string
		want           string
		ok             bool
	}{
		{"report.png", "*.png *.jpg", ".txt", "report.txt", true},
		{"report.jpg", "*.png *.jpg", ".txt", "report.txt", true},
		{"REPORT.PNG", "*.png", ".jpg", "REPORT.jpg", true},
		{"report.txt", "*.png *.jpg", ".gif", "report.txt", false},
		{"report", "*.png", ".jpg", "report", false},
		{"report.png", "*.png", "", "report.png", false},
		{"a.tar.gz", "*.gz *.tar.gz", ".zip", "a.zip", true},
		{".png", "*.png", ".jpg", ".png", false},
		{"x.png", "*", ".jpg", "x.png", false},
		{"x.png", "*.png", ".png", "x.png", false},
	}
	for _, tt := range tests {
		got, ok := swapExt(tt.name, splitPatterns(tt.old), tt.ext)
		if got != tt.want || ok != tt.ok {
			t.Errorf("swapExt(%q, %q, %q) = %q, %v, want %q, %v", tt.name, tt.old, tt.ext, got, ok, tt.want, tt.ok)
		}
	}
}
