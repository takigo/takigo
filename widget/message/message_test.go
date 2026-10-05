package message_test

import (
	"strings"
	"testing"

	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/widget/message"
)

func TestMessageWraps(t *testing.T) {
	app := testutil.NewTestApp(t)
	text := strings.Repeat("several words that need wrapping ", 6)
	narrow := message.New(app, "n", message.Text(text), message.WidthOpt(120))
	wide := message.New(app, "w", message.Text(text), message.WidthOpt(600))

	if narrow.Win.ReqWidth >= wide.Win.ReqWidth {
		t.Errorf("WidthOpt(120) is not narrower than WidthOpt(600): %d vs %d", narrow.Win.ReqWidth, wide.Win.ReqWidth)
	}
	if narrow.Win.ReqHeight <= wide.Win.ReqHeight {
		t.Errorf("the narrower message is not taller: %d vs %d", narrow.Win.ReqHeight, wide.Win.ReqHeight)
	}

	// Without a width, the aspect ratio (100*width/height) shapes it.
	tall := message.New(app, "t", message.Text(text), message.Aspect(50))
	flat := message.New(app, "f", message.Text(text), message.Aspect(400))
	if tall.Win.ReqWidth >= flat.Win.ReqWidth {
		t.Errorf("Aspect(50) is not narrower than Aspect(400): %d vs %d", tall.Win.ReqWidth, flat.Win.ReqWidth)
	}

	before := narrow.Win.ReqHeight
	if err := narrow.Configure(message.Text("short")); err != nil {
		t.Fatal(err)
	}
	if narrow.Win.ReqHeight >= before {
		t.Errorf("a shorter text did not shrink the request: %d vs %d", narrow.Win.ReqHeight, before)
	}
}
