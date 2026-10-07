package scale_test

import (
	"testing"

	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/widget/scale"
)

func TestScaleSetClampsAndRounds(t *testing.T) {
	app := testutil.NewTestApp(t)
	var got []float64
	s := scale.New(app, "s", scale.FromOpt(0), scale.ToOpt(10), scale.ResolutionOpt(0.5))
	s.Command = func(v float64) { got = append(got, v) }
	s.Set(3.3)
	if v := s.Get(); v != 3.5 {
		t.Errorf("Set(3.3) with resolution 0.5 = %v, want 3.5", v)
	}
	s.Set(42)
	if v := s.Get(); v != 10 {
		t.Errorf("Set past -to = %v, want 10", v)
	}
	s.Set(-1)
	if v := s.Get(); v != 0 {
		t.Errorf("Set below -from = %v, want 0", v)
	}
	if len(got) == 0 || got[len(got)-1] != 0 {
		t.Errorf("Command saw %v", got)
	}
}

func TestScaleOrientationAndLength(t *testing.T) {
	app := testutil.NewTestApp(t)
	// Tk's scale is vertical unless told otherwise.
	h := scale.New(app, "h", scale.LengthOpt(200), scale.OrientOpt(scale.Horizontal))
	v := scale.New(app, "v", scale.LengthOpt(200))
	if h.Win.ReqWidth < 200 || v.Win.ReqHeight < 200 {
		t.Errorf("length not along the orientation: horizontal %dx%d, vertical %dx%d",
			h.Win.ReqWidth, h.Win.ReqHeight, v.Win.ReqWidth, v.Win.ReqHeight)
	}
	if h.Win.ReqHeight >= 200 || v.Win.ReqWidth >= 200 {
		t.Errorf("the other axis is as long as the scale: horizontal %dx%d, vertical %dx%d",
			h.Win.ReqWidth, h.Win.ReqHeight, v.Win.ReqWidth, v.Win.ReqHeight)
	}
	before := h.Win.ReqWidth
	if err := h.Configure(scale.LengthOpt(300)); err != nil || h.Win.ReqWidth != before+100 {
		t.Errorf("Configure(LengthOpt(300)): err %v, width %d (was %d)", err, h.Win.ReqWidth, before)
	}
	plain := h.Win.ReqHeight
	h.Configure(scale.LabelOpt("Volume"))
	if h.Win.ReqHeight <= plain {
		t.Errorf("a label did not make the scale taller: %d vs %d", h.Win.ReqHeight, plain)
	}
}
