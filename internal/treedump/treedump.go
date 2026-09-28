// Package treedump captures the widget hierarchy of a running app as JSON so
// it can be compared with the same dump from Tk (scripts/tk_dump_tree.tcl).
// The schema is shared by both sides and by cmd/demodiff.
package treedump

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/window"
	"github.com/msorc/takigo/wm"
)

// SampleText is measured with every dumped font on both sides.
const SampleText = "The quick brown fox jumps over the lazy dog 0123456789"

// Dump is the whole snapshot.
type Dump struct {
	Side      string          `json:"side"`
	Toplevels []Toplevel      `json:"toplevels"`
	Widgets   []Widget        `json:"widgets"`
	Fonts     map[string]Font `json:"fonts"`
}

// Toplevel is a mapped toplevel window.
type Toplevel struct {
	Path  string `json:"path"`
	Title string `json:"title"`
	W     int    `json:"w"`
	H     int    `json:"h"`
}

// Widget is one window. X/Y are relative to its toplevel's origin.
type Widget struct {
	Path     string            `json:"path"`
	Class    string            `json:"class"`
	Toplevel string            `json:"toplevel"`
	Parent   string            `json:"parent"`
	Index    int               `json:"index"`
	X        int               `json:"x"`
	Y        int               `json:"y"`
	W        int               `json:"w"`
	H        int               `json:"h"`
	ReqW     int               `json:"reqw"`
	ReqH     int               `json:"reqh"`
	Manager  string            `json:"manager"`
	Mapped   bool              `json:"mapped"`
	Opts     map[string]string `json:"opts,omitempty"`
}

// Font describes a named font as actually resolved.
type Font struct {
	Family    string  `json:"family"`
	Size      float64 `json:"size"`
	Weight    string  `json:"weight"`
	Slant     string  `json:"slant"`
	Ascent    int     `json:"ascent"`
	Descent   int     `json:"descent"`
	Linespace int     `json:"linespace"`
	Fixed     bool    `json:"fixed"`
	Sample    int     `json:"sample"`
}

// Named fonts dumped on both sides.
var namedFonts = []string{
	font.TkDefaultFont, font.TkTextFont, font.TkFixedFont, font.TkMenuFont,
	font.TkHeadingFont, font.TkCaptionFont, font.TkSmallCaptionFont,
	font.TkIconFont, font.TkTooltipFont,
}

// Collect snapshots every mapped toplevel of d and its descendants.
func Collect(d *window.Display, fonts *font.Registry) *Dump {
	out := &Dump{Side: "go", Fonts: map[string]Font{}}
	var tops []*window.Window
	for _, w := range d.Windows {
		if w.IsTopLevel() && w.IsMapped() {
			tops = append(tops, w)
		}
	}
	sort.Slice(tops, func(i, j int) bool { return tops[i].PathName < tops[j].PathName })
	for _, top := range tops {
		t := Toplevel{Path: top.PathName, W: top.Width, H: top.Height - menubarHeight(top)}
		if info, ok := top.WmData.(*wm.WmInfo); ok {
			t.Title = info.Title
		}
		out.Toplevels = append(out.Toplevels, t)
		walk(out, top, top, 0, 0, 0)
	}
	for _, name := range namedFonts {
		f, err := fonts.Get(name)
		if err != nil {
			continue
		}
		out.Fonts[name] = describeFont(f)
	}
	return out
}

func walk(out *Dump, top, w *window.Window, index, x, y int) {
	parent := ""
	if w != top && w.Parent != nil {
		parent = w.Parent.PathName
	}
	manager := ""
	if w.GeomManager != nil {
		manager = w.GeomManager.Name()
	}
	h, reqH := w.Height, w.ReqHeight
	if w == top {
		// Like Tk, measure the toplevel without its menubar, which sits in
		// the wrapper above it: children are reported relative to the area
		// below the menubar, so the menubar itself gets a negative y.
		mb := menubarHeight(top)
		x, y, h, reqH = 0, 0, h-mb, reqH-mb
	}
	if w == top.Menubar {
		manager = "menubar"
	}
	out.Widgets = append(out.Widgets, Widget{
		Path: w.PathName, Class: w.Class, Toplevel: top.PathName, Parent: parent,
		Index: index, X: x, Y: y, W: w.Width, H: h,
		ReqW: w.ReqWidth, ReqH: reqH, Manager: manager, Mapped: w.IsMapped(),
	})
	base := y
	if w == top {
		base = -menubarHeight(top)
	}
	i := 0
	for _, c := range w.Children {
		if c.IsTopLevel() {
			continue
		}
		walk(out, top, c, i, x+c.X+w.BorderWidth, base+c.Y+w.BorderWidth)
		i++
	}
}

func describeFont(f font.Font) Font {
	a, m := f.Attrs(), f.Metrics()
	weight, slant := "normal", "roman"
	if a.Weight == font.WeightBold {
		weight = "bold"
	}
	if a.Slant == font.SlantItalic {
		slant = "italic"
	}
	return Font{
		Family: a.Family, Size: a.Size, Weight: weight, Slant: slant,
		Ascent: m.Ascent, Descent: m.Descent, Linespace: m.Linespace(),
		Fixed: m.Fixed, Sample: f.MeasureString(SampleText),
	}
}

// Marshal encodes a dump as indented JSON.
func Marshal(d *Dump) ([]byte, error) {
	return json.MarshalIndent(d, "", " ")
}

// Load reads a dump written by either side.
func Load(path string) (*Dump, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var d Dump
	if err := json.Unmarshal(b, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// WriteFileAtomic writes b to path via a temp file + rename so readers never
// see a partial dump.
func WriteFileAtomic(path string, b []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".treedump-*")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(b); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Chmod(0o644); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	return os.Rename(tmp.Name(), path)
}

func menubarHeight(top *window.Window) int {
	if top.Menubar == nil || !top.Menubar.IsMapped() {
		return 0
	}
	return top.Menubar.Height
}
