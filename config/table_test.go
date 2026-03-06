package config

import (
	"testing"
)

type testWidget struct {
	Name  string
	Width int
}

func newTestTable() *Table {
	t := NewTable()
	t.Register(Spec{Name: "name", Type: TypeString}, func(w any) any {
		return w.(*testWidget).Name
	}, func(w any, v any) error {
		w.(*testWidget).Name = v.(string)
		return nil
	}, 1)
	t.Register(Spec{Name: "width", Alias: "w", Type: TypeInt}, func(w any) any {
		return w.(*testWidget).Width
	}, func(w any, v any) error {
		w.(*testWidget).Width = v.(int)
		return nil
	}, 2)
	return t
}

func TestCget(t *testing.T) {
	tbl := newTestTable()
	w := &testWidget{Name: "hello", Width: 42}

	v, err := tbl.Cget(w, "name")
	if err != nil {
		t.Fatal(err)
	}
	if v != "hello" {
		t.Errorf("Cget(name) = %v, want hello", v)
	}

	v, err = tbl.Cget(w, "width")
	if err != nil {
		t.Fatal(err)
	}
	if v != 42 {
		t.Errorf("Cget(width) = %v, want 42", v)
	}
}

func TestCgetAlias(t *testing.T) {
	tbl := newTestTable()
	w := &testWidget{Width: 99}

	v, err := tbl.Cget(w, "w")
	if err != nil {
		t.Fatal(err)
	}
	if v != 99 {
		t.Errorf("Cget(w) = %v, want 99", v)
	}
}

func TestCgetUnknown(t *testing.T) {
	tbl := newTestTable()
	w := &testWidget{}

	_, err := tbl.Cget(w, "nonexistent")
	if err == nil {
		t.Error("expected error for unknown option")
	}
}

func TestConfigure(t *testing.T) {
	tbl := newTestTable()
	w := &testWidget{}

	changed, err := tbl.Configure(w, "name", "test", "width", 100)
	if err != nil {
		t.Fatal(err)
	}
	if changed != 3 { // 1|2
		t.Errorf("changed = %d, want 3", changed)
	}
	if w.Name != "test" {
		t.Errorf("Name = %q, want test", w.Name)
	}
	if w.Width != 100 {
		t.Errorf("Width = %d, want 100", w.Width)
	}
}

func TestConfigureAlias(t *testing.T) {
	tbl := newTestTable()
	w := &testWidget{}

	_, err := tbl.Configure(w, "w", 50)
	if err != nil {
		t.Fatal(err)
	}
	if w.Width != 50 {
		t.Errorf("Width = %d, want 50", w.Width)
	}
}

func TestConfigureUnknown(t *testing.T) {
	tbl := newTestTable()
	w := &testWidget{}

	_, err := tbl.Configure(w, "bad", "value")
	if err == nil {
		t.Error("expected error for unknown option")
	}
}

func TestConfigureOddArgs(t *testing.T) {
	tbl := newTestTable()
	w := &testWidget{}

	_, err := tbl.Configure(w, "name")
	if err == nil {
		t.Error("expected error for odd number of args")
	}
}

func TestSpecs(t *testing.T) {
	tbl := newTestTable()
	specs := tbl.Specs()
	if len(specs) != 2 {
		t.Fatalf("len(Specs) = %d, want 2", len(specs))
	}
	found := map[string]bool{}
	for _, s := range specs {
		found[s.Name] = true
	}
	if !found["name"] || !found["width"] {
		t.Errorf("expected name and width specs, got %v", found)
	}
}
