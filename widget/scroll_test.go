package widget

import (
	"reflect"
	"testing"
)

type recordingView struct{ calls []any }

func (v *recordingView) YViewMoveTo(f float64)     { v.calls = append(v.calls, f) }
func (v *recordingView) YViewScroll(n int, p bool) { v.calls = append(v.calls, n, p) }
func (v *recordingView) XViewMoveTo(f float64)     { v.calls = append(v.calls, "x", f) }
func (v *recordingView) XViewScroll(n int, p bool) { v.calls = append(v.calls, "x", n, p) }

func TestScrollCommands(t *testing.T) {
	v := &recordingView{}
	y, x := ScrollY(v), ScrollX(v)
	y(ScrollTo(0.25))
	y(ScrollUnits(-3))
	y(ScrollPages(1))
	x(ScrollTo(0.5))
	x(ScrollPages(-1))
	want := []any{0.25, -3, false, 1, true, "x", 0.5, "x", -1, true}
	if !reflect.DeepEqual(v.calls, want) {
		t.Errorf("calls = %v, want %v", v.calls, want)
	}
}
