package ttk

import (
	"testing"

	"github.com/takigo/takigo/window"
)

// The widget's own option wins over the style's state map, and the state
// map wins over the style's default (Tk's InitializeElementRecord).
func TestWidgetOptionBeatsStyleMap(t *testing.T) {
	th := NewTheme("widget-tier", nil)
	root := th.GetStyle(".")
	root.Defaults["-foreground"] = uint64(0x000000)
	root.Maps["-foreground"] = StateMap[any]{
		{Spec: StateSpec{OnBits: StateDisabled}, Value: uint64(0xa3a3a3)},
	}
	style := th.ResolveStyle("TLabel")

	// Without a widget option the style map wins over the style default.
	if got := LookupColor(style, "-foreground", StateDisabled, 0); got != 0xa3a3a3 {
		t.Errorf("disabled -foreground = %#x, want the style map 0xa3a3a3", got)
	}
	if got := LookupColor(style, "-foreground", 0, 0); got != 0x000000 {
		t.Errorf("normal -foreground = %#x, want the style default 0x000000", got)
	}

	// The widget's own option wins over the style map.
	w := &TtkWidget{Context: &DrawContext{Style: style}}
	w.SetWidgetOption("-foreground", uint64(0x123456))
	if got := LookupColor(w.Context.Style, "-foreground", StateDisabled, 0); got != 0x123456 {
		t.Errorf("disabled -foreground = %#x, want the widget option 0x123456", got)
	}
	// An option the widget does not set still comes from the style.
	if got := LookupColor(w.Context.Style, "-background", 0, 0xabcdef); got != 0xabcdef {
		t.Errorf("-background = %#x, want the fallback 0xabcdef", got)
	}
}

// A style of the parent theme that is registered after the child theme
// exists still becomes the fallback of the child theme's style.
func TestThemeFallbackRegisteredAfterChildTheme(t *testing.T) {
	parent := NewTheme("late-parent", nil)
	child := NewTheme("late-child", parent)
	s := child.ResolveStyle("TButton")
	if _, ok := s.Lookup("-padding", 0); ok {
		t.Fatal("found -padding before the parent theme registered the style")
	}

	// The parent theme registers the style afterwards.
	parent.GetStyle("TButton").Defaults["-padding"] = UniformPadding(4)

	v, ok := s.Lookup("-padding", 0)
	if !ok {
		t.Fatal("did not find the parent theme's -padding")
	}
	if p, ok := v.(Padding); !ok || p != UniformPadding(4) {
		t.Errorf("-padding = %v, want 4", v)
	}
}

// A derived style name falls back to the parent theme's class style.
func TestThemeFallbackDerivedName(t *testing.T) {
	parent := NewTheme("derived-parent", nil)
	parent.GetStyle("TButton").Defaults["-padding"] = UniformPadding(5)
	child := NewTheme("derived-child", parent)

	s := child.ResolveStyle("My.TButton")
	v, ok := s.Lookup("-padding", 0)
	if !ok {
		t.Fatal("did not find the parent theme's -padding")
	}
	if p, ok := v.(Padding); !ok || p != UniformPadding(5) {
		t.Errorf("-padding = %v, want 5", v)
	}
}

// The root style of a child theme inherits the root style of its parent
// theme, also when the parent theme creates its root style later.
func TestChildThemeRootInheritsParentRoot(t *testing.T) {
	parent := NewTheme("root-parent", nil)
	child := NewTheme("root-child", parent)

	// The parent theme creates its root style after the child theme.
	parent.GetStyle(".").Defaults["-background"] = uint64(0x123456)

	s := child.ResolveStyle(".")
	if got := LookupColor(s, "-background", 0, 0); got != 0x123456 {
		t.Errorf("-background = %#x, want the parent theme's root 0x123456", got)
	}
}

// initTestWidget prepares a widget that has a style but no display, so
// that Configure can run without a screen.
func initTestWidget(w *TtkWidget, style string) {
	theme := CurrentTheme()
	w.Win = &window.Window{}
	w.Theme = theme
	if theme != nil {
		w.Context = &DrawContext{Style: theme.ResolveStyle(style)}
	}
}

// A widget that derives its style from -orient keeps the style the user
// asked for, also when it is configured again (as UseTheme does).
func TestUserStyleIsKept(t *testing.T) {
	tests := []struct {
		name  string
		style string
		apply func() string
	}{
		{"scale", "My.TScale", func() string {
			w := &Scale{}
			initTestWidget(&w.TtkWidget, "TScale")
			if err := w.Configure(ScaleStyle("My.TScale")); err != nil {
				t.Fatalf("scale: %v", err)
			}
			return w.StyleName
		}},
		{"scrollbar", "My.TScrollbar", func() string {
			w := &Scrollbar{}
			initTestWidget(&w.TtkWidget, "TScrollbar")
			if err := w.Configure(ScrollbarStyle("My.TScrollbar")); err != nil {
				t.Fatalf("scrollbar: %v", err)
			}
			return w.StyleName
		}},
		{"separator", "My.TSeparator", func() string {
			w := &Separator{}
			initTestWidget(&w.TtkWidget, "TSeparator")
			if err := w.Configure(SeparatorStyle("My.TSeparator")); err != nil {
				t.Fatalf("separator: %v", err)
			}
			return w.StyleName
		}},
		{"progressbar", "My.TProgressbar", func() string {
			w := &Progressbar{}
			initTestWidget(&w.TtkWidget, "TProgressbar")
			if err := w.Configure(ProgressbarStyle("My.TProgressbar")); err != nil {
				t.Fatalf("progressbar: %v", err)
			}
			return w.StyleName
		}},
	}
	for _, tc := range tests {
		if got := tc.apply(); got != tc.style {
			t.Errorf("%s: style = %q, want %q", tc.name, got, tc.style)
		}
	}
}

// The style the user chose survives a Configure that does not mention it,
// which is what UseTheme does when it re-themes a widget.
func TestUserStyleSurvivesReconfigure(t *testing.T) {
	s := &Scale{}
	initTestWidget(&s.TtkWidget, "TScale")
	if err := s.Configure(ScaleStyle("My.TScale")); err != nil {
		t.Fatalf("scale: %v", err)
	}
	if err := s.Configure(); err != nil {
		t.Fatalf("scale: %v", err)
	}
	if s.StyleName != "My.TScale" {
		t.Errorf("scale: style = %q, want My.TScale", s.StyleName)
	}
}

// Without a user style the orientation still decides the style.
func TestOrientStyleIsDefault(t *testing.T) {
	s := &Scale{}
	initTestWidget(&s.TtkWidget, "TScale")
	if err := s.Configure(ScaleOrient(Vertical)); err != nil {
		t.Fatalf("scale: %v", err)
	}
	if s.StyleName != "Vertical.TScale" {
		t.Errorf("scale: style = %q, want Vertical.TScale", s.StyleName)
	}
	b := &Scrollbar{}
	initTestWidget(&b.TtkWidget, "TScrollbar")
	if err := b.Configure(ScrollbarOrient(Horizontal)); err != nil {
		t.Fatalf("scrollbar: %v", err)
	}
	if b.StyleName != "Horizontal.TScrollbar" {
		t.Errorf("scrollbar: style = %q, want Horizontal.TScrollbar", b.StyleName)
	}
}
