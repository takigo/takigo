package ttk

import "testing"

func benchThemes() *Theme {
	parent := NewTheme("benchparent", nil)
	parent.GetStyle(".").Defaults["-background"] = uint64(0xd9d9d9)
	parent.GetStyle("TButton").Defaults["-padding"] = "3"
	child := NewTheme("benchchild", parent)
	child.GetStyle(".").Defaults["-foreground"] = uint64(0)
	child.GetStyle("TButton")
	parent.RegisterLayout("TButton", &LayoutTemplate{})
	return child
}

// Per-draw path: an option that only the parent theme's style has.
func BenchmarkStyleLookupViaFallback(b *testing.B) {
	s := benchThemes().ResolveStyle("TButton")
	for b.Loop() {
		_, _ = s.Lookup("-padding", 0)
	}
}

// Per-layout-build path.
func BenchmarkResolveStyleAndLayout(b *testing.B) {
	t := benchThemes()
	for b.Loop() {
		_ = t.ResolveStyle("TButton")
		_ = t.GetLayout("Toolbutton.TButton")
		_ = t.GetElement("Button.border")
	}
}
