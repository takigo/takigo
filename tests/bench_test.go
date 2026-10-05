//go:build linux || freebsd || openbsd || netbsd

package takigo_test

import (
	"strconv"
	"testing"

	takigo "github.com/takigo/takigo"
	"github.com/takigo/takigo/geometry"
	"github.com/takigo/takigo/geometry/grid"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/ttk"
	_ "github.com/takigo/takigo/ttk/defaulttheme"
	"github.com/takigo/takigo/widget/frame"
	"github.com/takigo/takigo/widget/label"
)

func benchApp(b *testing.B) *takigo.App {
	testutil.RequireDisplay(b)
	a, err := takigo.NewApp(takigo.Size(400, 400))
	if err != nil {
		b.Skip(err)
	}
	b.Cleanup(a.Destroy)
	return a
}

// Relayout and redraw of 60 classic and ttk widgets after a text change.
func BenchmarkAppReconfigureRelayout(b *testing.B) {
	a := benchApp(b)
	f := frame.New(a, "f")
	pack.Pack(geometry.Group{f})
	var labels []*label.Label
	var buttons []*ttk.Button
	for i := range 30 {
		l := label.New(f, "l"+strconv.Itoa(i), label.Text("label"))
		t := ttk.NewButton(f, "b"+strconv.Itoa(i), ttk.ButtonText("button"))
		grid.Grid(geometry.Group{l, t}, grid.Row(i))
		labels = append(labels, l)
		buttons = append(buttons, t)
	}
	a.UpdateIdleTasks()
	i := 0
	for b.Loop() {
		i++
		s := strconv.Itoa(i)
		for j := range labels {
			labels[j].Configure(label.Text(s))
			buttons[j].Configure(ttk.ButtonText(s))
		}
		a.UpdateIdleTasks()
	}
	a.Server().Sync(false)
}

// Creating and destroying ttk widgets builds layouts from the shared theme.
func BenchmarkAppCreateDestroyTtk(b *testing.B) {
	a := benchApp(b)
	for b.Loop() {
		f := frame.New(a, "f")
		for i := range 20 {
			t := ttk.NewButton(f, "b"+strconv.Itoa(i), ttk.ButtonText("button"))
			pack.Pack(geometry.Group{t})
		}
		pack.Pack(geometry.Group{f})
		a.UpdateIdleTasks()
		f.Destroy()
	}
}

func BenchmarkGetImageRGBA(b *testing.B) {
	a := benchApp(b)
	srv := a.Server()
	pm := srv.CreatePixmap(platform.WindowDrawable(a.Root().PlatformID), 32, 32, uint(a.Root().Depth))
	defer srv.FreePixmap(pm)
	for b.Loop() {
		if srv.GetImageRGBA(platform.PixmapDrawable(pm), 0, 0, 32, 32) == nil {
			b.Fatal("GetImageRGBA failed")
		}
	}
}
