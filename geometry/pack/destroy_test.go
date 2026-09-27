package pack

import (
	"testing"

	"github.com/msorc/takigo/geometry"
	"github.com/msorc/takigo/window"
)

func TestDestroyDropsPackState(t *testing.T) {
	root := &window.Window{PathName: ".", Display: &window.Display{}}
	frame := window.NewChildWindow(root, "f", 0, 0, 10, 10)
	b1 := window.NewChildWindow(frame, "b1", 0, 0, 10, 10)
	b2 := window.NewChildWindow(frame, "b2", 0, 0, 10, 10)
	outside := window.NewChildWindow(root, "o", 0, 0, 10, 10)
	Pack(geometry.Group{b1, b2})
	Pack(geometry.Group{outside}, In(frame))

	window.DestroyWindow(b1)
	if n := len(packers[frame].entries); n != 2 {
		t.Fatalf("frame has %d packed entries after destroying b1, want 2", n)
	}
	if _, ok := containerOf[b1]; ok {
		t.Error("containerOf still holds destroyed b1")
	}

	window.DestroyWindow(frame)
	if _, ok := packers[frame]; ok {
		t.Error("packers still holds destroyed container")
	}
	if _, ok := containerOf[outside]; ok || outside.GeomManager != nil {
		t.Error("content packed -in a destroyed container is still managed")
	}
	if _, ok := containerOf[b2]; ok {
		t.Error("containerOf still holds b2")
	}
}
