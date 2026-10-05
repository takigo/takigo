package ttk_test

import (
	"testing"

	"github.com/takigo/takigo/geometry/pack"
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
	styled := ttk.NewEntry(app, "styled", ttk.EntryStyleOpt("Custom.TEntry"))
	if got := styled.Context.Style.Name; got != "Custom.TEntry" {
		t.Errorf("EntryStyleOpt: style = %q", got)
	}

	def.Configure(ttk.EntryWidth(40), ttk.EntryState2("disabled"))
	if def.Win.ReqWidth != wide.Win.ReqWidth {
		t.Errorf("Configure(EntryWidth(40)): ReqWidth = %d, want %d", def.Win.ReqWidth, wide.Win.ReqWidth)
	}
	if def.State&ttk.StateDisabled == 0 {
		t.Error("Configure(EntryState2(disabled)) left the state enabled")
	}
	def.Configure(ttk.EntryState2("normal"))
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
