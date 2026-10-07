package ttk_test

import (
	"testing"
	"time"

	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/geometry/place"
	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/ttk"
	_ "github.com/takigo/takigo/ttk/defaulttheme"
)

func TestEntryWidthAndStyleOptions(t *testing.T) {
	app := testutil.NewTestApp(t)
	def := ttk.NewEntry(app, "def")
	wide := ttk.NewEntry(app, "wide", ttk.EntryWidth(40))
	if wide.Win.ReqWidth <= def.Win.ReqWidth {
		t.Errorf("EntryWidth(40): ReqWidth = %d, default = %d", wide.Win.ReqWidth, def.Win.ReqWidth)
	}
	styled := ttk.NewEntry(app, "styled", ttk.EntryStyle("Custom.TEntry"))
	if got := styled.Context.Style.Name; got != "Custom.TEntry" {
		t.Errorf("EntryStyle: style = %q", got)
	}

	def.Configure(ttk.EntryWidth(40), ttk.EntryState(ttk.FieldDisabled))
	if def.Win.ReqWidth != wide.Win.ReqWidth {
		t.Errorf("Configure(EntryWidth(40)): ReqWidth = %d, want %d", def.Win.ReqWidth, wide.Win.ReqWidth)
	}
	if def.State&ttk.StateDisabled == 0 {
		t.Error("Configure(EntryState2(disabled)) left the state enabled")
	}
	def.Configure(ttk.EntryState(ttk.FieldNormal))
	if def.State&ttk.StateDisabled != 0 {
		t.Error("Configure(EntryState2(normal)) left the state disabled")
	}
}

func TestConfigureRelaysOutParent(t *testing.T) {
	app := testutil.NewTestApp(t)
	f := ttk.NewFrame(app, "f")
	pack.Pack(f)
	l := ttk.NewLabel(f, "l", ttk.LabelText("x"))
	b := ttk.NewButton(f, "b", ttk.ButtonText("x"))
	pack.Pack(l)
	pack.Pack(b)
	app.UpdateIdleTasks()
	labelW, buttonW, frameW := l.Win.ReqWidth, b.Win.ReqWidth, f.Win.ReqWidth

	l.Configure(ttk.LabelText("a considerably longer label text"))
	b.Configure(ttk.ButtonText("a considerably longer button text"))
	if l.Win.ReqWidth <= labelW || b.Win.ReqWidth <= buttonW {
		t.Fatalf("requests did not grow: label %d -> %d, button %d -> %d", labelW, l.Win.ReqWidth, buttonW, b.Win.ReqWidth)
	}
	app.UpdateIdleTasks()
	if f.Win.ReqWidth <= frameW {
		t.Errorf("frame request %d did not follow its content (was %d)", f.Win.ReqWidth, frameW)
	}

	frameW = f.Win.ReqWidth
	f.Configure(ttk.FramePadding(ttk.Padding{Left: 20, Right: 20}))
	app.UpdateIdleTasks()
	if f.Win.ReqWidth != frameW+40 {
		t.Errorf("FramePadding: frame request = %d, want %d", f.Win.ReqWidth, frameW+40)
	}
}

// Hiding a notebook pane runs the unmap hooks of the geometry managers, so
// content placed -in the pane from elsewhere is unmapped with it
// (Tk_MaintainGeometry).
func TestNotebookUnmapsContentPlacedInHiddenPane(t *testing.T) {
	app := testutil.NewTestApp(t)
	nb := ttk.NewNotebook(app, "nb")
	pack.Pack(nb)
	first := ttk.NewFrame(nb, "first")
	second := ttk.NewFrame(nb, "second")
	nb.Add(first.Win, "First")
	nb.Add(second.Win, "Second")
	floating := ttk.NewLabel(app, "floating", ttk.LabelText("in the second pane"))
	place.Place(floating, place.In(second), place.X(5), place.Y(5))

	// The root is mapped when the loop starts, so check from inside it.
	var shown, hidden, reshown bool
	app.After(50*time.Millisecond, func() {
		defer app.Quit()
		nb.Select(1)
		shown = floating.Win.IsMapped()
		nb.Select(0)
		hidden = !floating.Win.IsMapped()
		nb.Select(1)
		reshown = floating.Win.IsMapped()
	})
	app.MainLoop()
	if !shown {
		t.Fatal("content placed in the selected pane is not mapped")
	}
	if !hidden {
		t.Error("content placed in a hidden pane stayed mapped")
	}
	if !reshown {
		t.Error("content not remapped with its pane")
	}
}
