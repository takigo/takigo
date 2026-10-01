package scrollbar_test

import (
	"testing"

	"github.com/msorc/takigo/internal/testutil"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/scrollbar"
)

func TestScrollbarGeometry(t *testing.T) {
	app := testutil.NewTestApp(t)
	v := scrollbar.New(app, "v")
	h := scrollbar.New(app, "h", scrollbar.OrientOpt(scrollbar.Horizontal))
	if v.Orient != scrollbar.Vertical {
		t.Errorf("default orientation = %v, want vertical", v.Orient)
	}
	// A scrollbar is thin across its axis and asks for the same thickness
	// either way round.
	if v.Win.ReqWidth != h.Win.ReqHeight {
		t.Errorf("vertical width %d != horizontal height %d", v.Win.ReqWidth, h.Win.ReqHeight)
	}
	if v.Win.ReqHeight <= v.Win.ReqWidth || h.Win.ReqWidth <= h.Win.ReqHeight {
		t.Errorf("requests %dx%d (vertical), %dx%d (horizontal) are not elongated",
			v.Win.ReqWidth, v.Win.ReqHeight, h.Win.ReqWidth, h.Win.ReqHeight)
	}

	fat := scrollbar.New(app, "fat", scrollbar.WidthOpt(30))
	if fat.Win.ReqWidth <= v.Win.ReqWidth {
		t.Errorf("WidthOpt(30) is not wider: %d vs %d", fat.Win.ReqWidth, v.Win.ReqWidth)
	}
}

func TestScrollbarSetAndCommand(t *testing.T) {
	app := testutil.NewTestApp(t)
	var got []widget.ScrollRequest
	s := scrollbar.New(app, "s", scrollbar.CommandOpt(func(r widget.ScrollRequest) { got = append(got, r) }))
	s.Set(0.25, 0.75)
	s.Set(-1, 2) // clamped, not a panic
	s.Command(widget.ScrollPages(1))
	if len(got) != 1 || !got[0].Pages || got[0].Count != 1 {
		t.Errorf("command received %+v", got)
	}
}
