package spinbox_test

import (
	"errors"
	"testing"

	"github.com/takigo/takigo/color"
	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/widget/spinbox"
)

func TestSpinRangeAndWrap(t *testing.T) {
	app := testutil.NewTestApp(t)
	s := spinbox.New(app, "s", spinbox.FromOpt(0), spinbox.ToOpt(2), spinbox.IncrementOpt(1))
	if got := s.GetText(); got != "0" {
		t.Fatalf("initial text %q, want 0", got)
	}
	s.SpinUp()
	s.SpinUp()
	s.SpinUp()
	if got := s.GetText(); got != "2" {
		t.Errorf("spinning past -to gave %q, want 2", got)
	}
	s.Configure(spinbox.WrapOpt(true))
	s.SpinUp()
	if got := s.GetText(); got != "0" {
		t.Errorf("wrapping past -to gave %q, want 0", got)
	}
	s.SpinDown()
	if got := s.GetText(); got != "2" {
		t.Errorf("wrapping below -from gave %q, want 2", got)
	}
}

func TestSpinValues(t *testing.T) {
	app := testutil.NewTestApp(t)
	s := spinbox.New(app, "s", spinbox.ValuesOpt([]string{"red", "green", "blue"}))
	if got := s.GetText(); got != "red" {
		t.Fatalf("initial value %q, want red", got)
	}
	s.SpinUp()
	s.SpinUp()
	s.SpinUp()
	if got := s.GetText(); got != "blue" {
		t.Errorf("spinning past the last value gave %q, want blue", got)
	}
	s.SetText("green")
	s.SpinDown()
	if got := s.GetText(); got != "red" {
		t.Errorf("SpinDown from a set value gave %q, want red", got)
	}
}

func TestSpinboxEditsAndConfigure(t *testing.T) {
	app := testutil.NewTestApp(t)
	s := spinbox.New(app, "s", spinbox.FromOpt(0), spinbox.ToOpt(100))
	s.SetText("hello")
	s.Insert(5, " world")
	s.Delete(0, 1)
	if got := s.GetText(); got != "ello world" {
		t.Errorf("after edits %q", got)
	}
	s.SelectAll()
	s.DeleteSelection()
	if got := s.GetText(); got != "" {
		t.Errorf("DeleteSelection left %q", got)
	}
	narrow := s.Win.ReqWidth
	if err := s.Configure(spinbox.WidthOpt(30)); err != nil || s.Win.ReqWidth <= narrow {
		t.Errorf("Configure(WidthOpt(30)): err %v, width %d (was %d)", err, s.Win.ReqWidth, narrow)
	}
	if err := s.Configure(spinbox.Background("no-such-colour")); !errors.Is(err, color.ErrUnknown) {
		t.Errorf("a bad colour gave %v", err)
	}
}
