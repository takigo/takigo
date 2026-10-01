package labelframe_test

import (
	"testing"

	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/testutil"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/labelframe"
)

func TestLabelframeLabelMakesRoom(t *testing.T) {
	app := testutil.NewTestApp(t)
	plain := labelframe.New(app, "plain")
	titled := labelframe.New(app, "titled", labelframe.Text("Options"))

	// The label sits in the top border, so content starts lower.
	if titled.Win.InternalBorderTop <= plain.Win.InternalBorderTop {
		t.Errorf("top inset with a label %d, without %d", titled.Win.InternalBorderTop, plain.Win.InternalBorderTop)
	}
	if titled.Win.InternalBorderLeft != plain.Win.InternalBorderLeft {
		t.Errorf("the label changed the left inset: %d vs %d", titled.Win.InternalBorderLeft, plain.Win.InternalBorderLeft)
	}

	before := titled.Win.InternalBorderLeft
	if err := titled.Configure(labelframe.BorderWidth(before + 5)); err != nil {
		t.Fatal(err)
	}
	if titled.Win.InternalBorderLeft <= before {
		t.Errorf("a wider border did not grow the inset: %d vs %d", titled.Win.InternalBorderLeft, before)
	}
}

func TestLabelframeHoldsContent(t *testing.T) {
	app := testutil.NewTestApp(t)
	lf := labelframe.New(app, "lf", labelframe.Text("Group"))
	inner := label.New(lf, "inner", label.Text("some content in the frame"))
	pack.Pack(inner)
	app.UpdateIdleTasks()
	if lf.Win.ReqWidth < inner.Win.ReqWidth || lf.Win.ReqHeight <= inner.Win.ReqHeight {
		t.Errorf("frame request %dx%d does not enclose content %dx%d",
			lf.Win.ReqWidth, lf.Win.ReqHeight, inner.Win.ReqWidth, inner.Win.ReqHeight)
	}
}
