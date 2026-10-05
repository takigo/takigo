package grid_test

import (
	"testing"

	"github.com/takigo/takigo/geometry"
	"github.com/takigo/takigo/geometry/grid"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/window"
)

type benchServer struct{ platform.DisplayServer }

func (benchServer) MoveResizeWindow(platform.WindowID, int, int, uint, uint) {}
func (benchServer) MoveWindow(platform.WindowID, int, int)                   {}
func (benchServer) ResizeWindow(platform.WindowID, uint, uint)               {}
func (benchServer) MapWindow(platform.WindowID)                              {}
func (benchServer) UnmapWindow(platform.WindowID)                            {}
func (benchServer) DestroyWindow(platform.WindowID)                          {}

func benchTree(n int) (*window.Window, []*window.Window) {
	d := &window.Display{Server: benchServer{}, Windows: map[platform.WindowID]*window.Window{}}
	c := &window.Window{Display: d, PlatformID: 1, Width: 800, Height: 600, PathName: ".c"}
	c.Flags |= window.FlagMapped
	kids := make([]*window.Window, n)
	for i := range kids {
		k := &window.Window{Display: d, Parent: c, PlatformID: platform.WindowID(100 + i), ReqWidth: 40, ReqHeight: 20}
		c.Children = append(c.Children, k)
		kids[i] = k
	}
	return c, kids
}

func BenchmarkGridArrange50(b *testing.B) {
	c, kids := benchTree(50)
	for i, k := range kids {
		grid.Grid(geometry.Group{k}, grid.Row(i/5), grid.Column(i%5))
	}
	for b.Loop() {
		grid.ArrangeContainer(c)
	}
}

func BenchmarkGridForgetRegrid50(b *testing.B) {
	_, kids := benchTree(50)
	for b.Loop() {
		for i, k := range kids {
			grid.Grid(geometry.Group{k}, grid.Row(i/5), grid.Column(i%5))
		}
		for _, k := range kids {
			grid.Forget(k)
		}
	}
}
