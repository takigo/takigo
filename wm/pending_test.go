package wm

import (
	"testing"

	"github.com/takigo/takigo/window"
)

// A ConfigureNotify that matches a pending size in one dimension only is
// the user's resize: a user dragging one edge keeps the other dimension.
func TestConfigureNotifyOneDimensionIsUser(t *testing.T) {
	info, _ := newToplevel(t)
	w := info.Win
	w.ReqWidth, w.ReqHeight = 120, 80
	window.ResizeToplevel(w, 120, 80)

	info.ConfigureNotify(120, 95)
	if info.UserH != 95 {
		t.Fatalf("user height = %d, want 95", info.UserH)
	}
}

// A size that matches no pending size at all is the user's.
func TestConfigureNotifyUserResizeStillRecorded(t *testing.T) {
	info, _ := newToplevel(t)
	w := info.Win
	w.ReqWidth, w.ReqHeight = 120, 80
	window.ResizeToplevel(w, 120, 80)

	info.ConfigureNotify(400, 300)
	if info.UserW != 400 || info.UserH != 300 {
		t.Errorf("user size = %dx%d, want 400x300", info.UserW, info.UserH)
	}
}
