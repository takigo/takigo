package font

import "testing"

type countingOpener struct{ opened, closed int }

type countedFont struct {
	o     *countingOpener
	attrs Attributes
}

func (f *countedFont) Attrs() Attributes        { return f.attrs }
func (f *countedFont) Metrics() Metrics         { return Metrics{Ascent: 8, Descent: 2} }
func (f *countedFont) MeasureString(string) int { return 0 }
func (f *countedFont) Close()                   { f.o.closed++ }

func (o *countingOpener) OpenFont(a Attributes) (Font, error) {
	o.opened++
	return &countedFont{o, a}, nil
}
func (o *countingOpener) Families() []string { return nil }

func TestRegistrySharesFontsBySpelling(t *testing.T) {
	o := &countingOpener{}
	r := NewRegistry(o)
	a, _ := r.Get("Helvetica 12 bold")
	b, _ := r.Get("{Helvetica} 12 bold")
	c, _ := r.Get("-family Helvetica -size 12 -weight bold")
	if a != b || b != c || o.opened != 1 {
		t.Errorf("three spellings opened %d fonts (same: %v %v), want 1 shared", o.opened, a == b, b == c)
	}
}

func TestRegistryDefineKeepsOldFontOpen(t *testing.T) {
	o := &countingOpener{}
	r := NewRegistry(o)
	old, _ := r.Get("TkDefaultFont")
	r.Define("TkDefaultFont", Attributes{Family: "Courier", Size: 20})
	nf, _ := r.Get("TkDefaultFont")
	if o.closed != 0 {
		t.Error("Define closed a font a widget may still hold")
	}
	if nf == old || nf.Attrs().Family != "Courier" {
		t.Errorf("Get after Define = %+v, want the new Courier font", nf.Attrs())
	}
	r.Close()
	if o.closed != 2 {
		t.Errorf("Close closed %d fonts, want 2", o.closed)
	}
}
